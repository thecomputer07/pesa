package importer

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"net/url"
	"strings"
	"time"
)

// OFFOpts controls the Open Food Facts bulk import.
type OFFOpts struct {
	ImportOpts
	Country   string        // es. "italy"
	Limit     int           // max prodotti totali (0 = no cap)
	PageSize  int           // default 100
	RateDelay time.Duration // delay tra richieste (default 1s)
	UserAgent string
	MinCarb   bool // se true, scarta prodotti senza carbohydrates_100g
	// ByScans orders by unique_scans_n (how many people scanned the product:
	// the best proxy for "most eaten and drunk") via the classic v2 API.
	// search-a-licious' popularity_key put Herbalife and nameless products
	// first and caps at 10 000 results.
	ByScans    bool
	StartPage  int // resume a stopped run (default 1)
	HTTPClient *http.Client
}

type offNutriments struct {
	Energy100g *float64 `json:"energy-kcal_100g"`
	ProteinG   *float64 `json:"proteins_100g"`
	FatG       *float64 `json:"fat_100g"`
	CarbG      *float64 `json:"carbohydrates_100g"`
	// fallback when OFF only has the "prepared" value (see off.FromOFF)
	CarbPreparedG *float64 `json:"carbohydrates_prepared_100g"`
	SugarG        *float64 `json:"sugars_100g"`
	FiberG        *float64 `json:"fiber_100g"`
	SaltG         *float64 `json:"salt_100g"`
}

type offProduct struct {
	Code            string         `json:"code"`
	ProductName     string         `json:"product_name"`
	ProductNameIT   string         `json:"product_name_it"`
	Categories      string         `json:"categories"`
	Nutriments      *offNutriments `json:"nutriments"`
	ImageFrontURL   string         `json:"image_front_url"`
	ImageFrontSmall string         `json:"image_front_small_url"`
	ImageURL        string         `json:"image_url"`
	// search-a-licious restituisce brands come array, l'API v1 come stringa: accettiamo entrambi.
	Brands json.RawMessage `json:"brands"`
}

func (p offProduct) bestImageURL() string {
	for _, u := range []string{p.ImageFrontURL, p.ImageURL, p.ImageFrontSmall} {
		if strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "http://") {
			return u
		}
	}
	return ""
}

func (p offProduct) firstBrand() string {
	if len(p.Brands) == 0 {
		return ""
	}
	var list []string
	if err := json.Unmarshal(p.Brands, &list); err != nil {
		var s string
		if err := json.Unmarshal(p.Brands, &s); err != nil {
			return ""
		}
		list = strings.Split(s, ",")
	}
	return pickBrand(list)
}

// companyRe matches legal entities ("COCA-COLA SERVICES SA/NV", "Barilla G.
// e R. Fratelli SpA"): the API v2 lists them next to the consumer brand.
var companyRe = regexp.MustCompile(`(?i)(^|[\s,])(s\.?p\.?a\.?|s\.?r\.?l\.?|sas|sa/nv|s\.?a\.?|n\.?v\.?|gmbh|ag|ab|bv|ltd|limited|inc|llc|services|company|division|commerciale|produits|fratelli|group)([\s,.]|$)`)

// pickBrand returns the first consumer brand, skipping company names.
func pickBrand(list []string) string {
	first := ""
	for _, b := range list {
		b = strings.TrimSpace(b)
		if b == "" {
			continue
		}
		if first == "" {
			first = b
		}
		if !companyRe.MatchString(b) {
			return b
		}
	}
	return first
}

// Risposta di search-a-licious (https://search.openfoodfacts.org/docs).
type offSearchResponse struct {
	Count     int          `json:"count"`
	Page      int          `json:"page"`
	PageSize  int          `json:"page_size"`
	PageCount int          `json:"page_count"`
	Hits      []offProduct `json:"hits"`
}

// ImportOFF itera le pagine dell'API OFF filtrate per country e carica
// ogni prodotto nella tabella foods con source_code='off'.
func ImportOFF(ctx context.Context, db *sql.DB, opts OFFOpts) (*ImportReport, error) {
	if opts.SourceCode == "" {
		opts.SourceCode = "off"
	}
	if opts.Basis == "" {
		opts.Basis = BasisEdible
	}
	if opts.Country == "" {
		opts.Country = "italy"
	}
	if opts.PageSize <= 0 {
		opts.PageSize = 100
	}
	if opts.RateDelay <= 0 {
		opts.RateDelay = time.Second
		if opts.ByScans {
			opts.RateDelay = 7 * time.Second // v2 search: 10 req/min
		}
	}
	if opts.UserAgent == "" {
		opts.UserAgent = "Pesa/0.1.0 (+https://pesadia.com)"
	}
	if opts.HTTPClient == nil {
		opts.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}

	sourceID, err := lookupSourceID(ctx, db, opts.SourceCode)
	if err != nil {
		return nil, err
	}

	report := &ImportReport{}

	page := 1
	if opts.StartPage > 1 {
		page = opts.StartPage
	}
	fetched := 0
	for {
		if ctx.Err() != nil {
			return report, ctx.Err()
		}
		if opts.Limit > 0 && fetched >= opts.Limit {
			break
		}
		resp, err := fetchOFFPage(ctx, opts, page)
		if err != nil {
			return report, fmt.Errorf("page %d: %w", page, err)
		}
		if len(resp.Hits) == 0 {
			break
		}
		for _, p := range resp.Hits {
			if opts.Limit > 0 && fetched >= opts.Limit {
				break
			}
			fetched++
			report.TotalRows++

			importOFFProduct(ctx, db, sourceID, p, opts, report)
		}
		slog.Info("off page done", "page", page, "hits", len(resp.Hits),
			"fetched", fetched, "inserted", report.InsertedFoods, "updated", report.UpdatedFoods,
			"unchanged", report.Unchanged, "quarantined", report.Quarantined)

		if resp.PageCount > 0 && page >= resp.PageCount {
			break
		}
		// The v2 API can return a few less than page_size mid-stream (99 of
		// 100): only an empty page or the last declared page ends it there.
		if !opts.ByScans && len(resp.Hits) < opts.PageSize {
			break
		}
		page++
		select {
		case <-ctx.Done():
			return report, ctx.Err()
		case <-time.After(opts.RateDelay):
		}
	}
	return report, nil
}

// offFirstCategory returns the first category, cut to the 64-char column.
func offFirstCategory(cats string) string {
	for _, c := range strings.Split(cats, ",") {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		if len(c) > 64 {
			c = c[:64]
		}
		return c
	}
	return ""
}

// importOFFProduct turns one OFF product into a food (+ its image). Shared by
// the API import and the dump import. Skips are silent, errors go in report.
func importOFFProduct(ctx context.Context, db *sql.DB, sourceID int64, p offProduct, opts OFFOpts, report *ImportReport) {
	if p.Code == "" {
		return
	}
	name := p.ProductNameIT
	if name == "" {
		name = p.ProductName
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	nutr := p.Nutriments
	if nutr == nil {
		nutr = &offNutriments{}
	}
	if nutr.CarbG == nil {
		nutr.CarbG = nutr.CarbPreparedG
	}
	if opts.MinCarb && nutr.CarbG == nil {
		return
	}

	raw, _ := json.Marshal(p)
	hash := sha256.Sum256(raw)
	displayName := name
	if b := p.firstBrand(); b != "" {
		displayName = b + " — " + name
	}
	row := FoodRow{
		ExternalID:  p.Code,
		URL:         "https://world.openfoodfacts.org/product/" + p.Code,
		NameIT:      displayName,
		Category:    offFirstCategory(p.Categories),
		EdiblePct:   100.0,
		Kcal:        nutr.Energy100g,
		ProteinG:    nutr.ProteinG,
		FatG:        nutr.FatG,
		CarbG:       nutr.CarbG,
		SugarG:      nutr.SugarG,
		FiberG:      nutr.FiberG,
		SaltG:       nutr.SaltG,
		ContentHash: hex.EncodeToString(hash[:]),
		FetchedAt:   time.Now().UTC(),
		RawJSON:     raw,
	}
	if err := importOneRow(ctx, db, sourceID, row, opts.ImportOpts, report, false); err != nil {
		report.Errors = append(report.Errors, fmt.Sprintf("%s: %v", row.ExternalID, err))
		return
	}
	if !opts.DryRun {
		if imgURL := p.bestImageURL(); imgURL != "" {
			if err := upsertFoodImage(ctx, db, sourceID, row.ExternalID, imgURL); err != nil {
				slog.Warn("off image upsert", "ean", row.ExternalID, "err", err)
			}
		}
	}
}

// upsertFoodImage inserisce (o sostituisce) l'URL immagine principale per il food
// collegato al record (source_id, external_id). Mantiene UN SOLO record per food
// per semplicità — la tabella food_images supporta più immagini ma il frontend
// mostra solo la prima.
func upsertFoodImage(ctx context.Context, db *sql.DB, sourceID int64, externalID, url string) error {
	var foodID int64
	err := db.QueryRowContext(ctx, `
		SELECT f.id FROM foods f
		JOIN food_source_records fsr ON fsr.id = f.source_record_id
		WHERE fsr.source_id = $1 AND fsr.external_id = $2`, sourceID, externalID).Scan(&foodID)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `DELETE FROM food_images WHERE food_id = $1`, foodID)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `
		INSERT INTO food_images (food_id, url, attribution, license, is_external)
		VALUES ($1, $2, 'Open Food Facts contributors', 'ODbL-1.0', TRUE)`, foodID, url)
	return err
}

// offV2Response is the classic /api/v2/search shape.
type offV2Response struct {
	Count    int          `json:"count"`
	Products []offProduct `json:"products"`
}

func fetchOFFPage(ctx context.Context, opts OFFOpts, page int) (*offSearchResponse, error) {
	const fields = "code,product_name,product_name_it,categories,nutriments,brands,image_url,image_front_url,image_front_small_url"
	q := url.Values{}
	var u string
	if opts.ByScans {
		q.Set("countries_tags_en", opts.Country)
		q.Set("sort_by", "unique_scans_n")
		q.Set("fields", fields)
		q.Set("page_size", fmt.Sprintf("%d", opts.PageSize))
		q.Set("page", fmt.Sprintf("%d", page))
		u = "https://world.openfoodfacts.org/api/v2/search?" + q.Encode()
	} else {
		// search-a-licious è l'endpoint OFF consigliato per bulk search (ElasticSearch-backed).
		// Doc: https://search.openfoodfacts.org/docs
		q.Set("q", fmt.Sprintf(`countries_tags:"en:%s"`, opts.Country))
		q.Set("sort_by", "popularity_key")
		q.Set("fields", fields)
		q.Set("page_size", fmt.Sprintf("%d", opts.PageSize))
		q.Set("page", fmt.Sprintf("%d", page))
		u = "https://search.openfoodfacts.org/search?" + q.Encode()
	}

	// OFF answers 503/429 when busy, sometimes for minutes: retry patiently
	// (5 s doubling to a 2 min cap, ~10 min overall) before giving up.
	var lastErr error
	backoff := 5 * time.Second
	for attempt := 1; attempt <= 8; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", opts.UserAgent)
		req.Header.Set("Accept", "application/json")

		resp, err := opts.HTTPClient.Do(req)
		if err != nil {
			lastErr = err
		} else if resp.StatusCode == 200 {
			var out offSearchResponse
			var derr error
			if opts.ByScans {
				var v2 offV2Response
				derr = json.NewDecoder(resp.Body).Decode(&v2)
				out = offSearchResponse{Count: v2.Count, Page: page, PageSize: opts.PageSize, Hits: v2.Products}
				if opts.PageSize > 0 {
					out.PageCount = (v2.Count + opts.PageSize - 1) / opts.PageSize
				}
			} else {
				derr = json.NewDecoder(resp.Body).Decode(&out)
			}
			_ = resp.Body.Close()
			if derr != nil {
				return nil, derr
			}
			return &out, nil
		} else {
			b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
			_ = resp.Body.Close()
			lastErr = fmt.Errorf("OFF %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
			if resp.StatusCode != 429 && resp.StatusCode < 500 {
				return nil, lastErr
			}
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
		if backoff *= 2; backoff > 2*time.Minute {
			backoff = 2 * time.Minute
		}
	}
	return nil, lastErr
}
