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
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// CREALiveOpts controlla l'import live dal sito alimentinutrizione.it (CREA).
type CREALiveOpts struct {
	ImportOpts
	IndexURL string // default: ricerca-per-alimento
	BaseURL  string // default: https://www.alimentinutrizione.it
	Limit    int    // 0 = tutti
	// OnlyMissing skips index entries that are already foods (fills gaps
	// without re-fetching ~830 pages).
	OnlyMissing bool
	RateDelay   time.Duration // default 2s (rispetto del server)
	UserAgent   string
	HTTPClient  *http.Client
}

var creaNutrientMap = map[string]func(*FoodRow, float64){
	"Acqua (g)":                   func(r *FoodRow, v float64) { r.WaterG = &v },
	"Energia (kcal)":              func(r *FoodRow, v float64) { r.Kcal = &v },
	"Proteine (g)":                func(r *FoodRow, v float64) { r.ProteinG = &v },
	"Lipidi (g)":                  func(r *FoodRow, v float64) { r.FatG = &v },
	"Carboidrati disponibili (g)": func(r *FoodRow, v float64) { r.CarbG = &v },
	"Zuccheri solubili (g)":       func(r *FoodRow, v float64) { r.SugarG = &v },
	"Fibra totale (g)":            func(r *FoodRow, v float64) { r.FiberG = &v },
	"Alcool (g)":                  func(r *FoodRow, v float64) { r.AlcoholG = &v },
	"Sodio (mg)":                  func(r *FoodRow, v float64) { sg := v * 2.5 / 1000.0; r.SaltG = &sg },
}

var crearNumRe = regexp.MustCompile(`-?\d+(?:[.,]\d+)?`)

// ImportCREALive scrape & importa alimenti da alimentinutrizione.it.
func ImportCREALive(ctx context.Context, db *sql.DB, opts CREALiveOpts) (*ImportReport, error) {
	if opts.SourceCode == "" {
		opts.SourceCode = "crea"
	}
	if opts.Basis == "" {
		opts.Basis = BasisEdible
	}
	if opts.BaseURL == "" {
		opts.BaseURL = "https://www.alimentinutrizione.it"
	}
	if opts.IndexURL == "" {
		opts.IndexURL = opts.BaseURL + "/tabelle-nutrizionali/ricerca-per-alimento"
	}
	if opts.RateDelay <= 0 {
		opts.RateDelay = 2 * time.Second
	}
	if opts.UserAgent == "" {
		opts.UserAgent = "Pesa/0.1.0 (+https://pesadia.com; lorenzo.dimare@icloud.com)"
	}
	if opts.HTTPClient == nil {
		opts.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}

	sourceID, err := lookupSourceID(ctx, db, opts.SourceCode)
	if err != nil {
		return nil, err
	}

	// Step 1: parse index → list of (id, url)
	urls, err := fetchCREAIndex(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("index: %w", err)
	}
	if opts.OnlyMissing {
		urls, err = dropCREAInCatalog(ctx, db, sourceID, urls)
		if err != nil {
			return nil, err
		}
	}
	if opts.Limit > 0 && len(urls) > opts.Limit {
		urls = urls[:opts.Limit]
	}
	slog.Info("crea live: index parsed", "count", len(urls))

	report := &ImportReport{TotalRows: len(urls)}

	// Step 2: per ciascun URL, fetch + parse + import
	for i, codeURL := range urls {
		if ctx.Err() != nil {
			return report, ctx.Err()
		}
		row, err := fetchCREADetail(ctx, opts, codeURL)
		if err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("%s: %v", codeURL.code, err))
		} else if row != nil {
			if err := importOneRow(ctx, db, sourceID, *row, opts.ImportOpts, report, false); err != nil {
				report.Errors = append(report.Errors, fmt.Sprintf("%s: %v", row.ExternalID, err))
			}
		}

		if (i+1)%25 == 0 {
			slog.Info("crea live progress", "done", i+1, "total", len(urls),
				"inserted", report.InsertedFoods, "updated", report.UpdatedFoods,
				"unchanged", report.Unchanged, "quarantined", report.Quarantined)
		}
		select {
		case <-ctx.Done():
			return report, ctx.Err()
		case <-time.After(opts.RateDelay):
		}
	}
	if err := RefreshCREAStates(ctx, db); err != nil {
		return report, err
	}
	return report, nil
}

// dropCREAInCatalog keeps only index entries that are not foods yet.
func dropCREAInCatalog(ctx context.Context, db *sql.DB, sourceID int64, in []creaIndexEntry) ([]creaIndexEntry, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT r.external_id FROM foods f
		  JOIN food_source_records r ON r.id = f.source_record_id
		 WHERE f.source_id = $1`, sourceID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	have := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		have[id] = true
	}
	out := in[:0]
	for _, e := range in {
		if !have[e.code] {
			out = append(out, e)
		}
	}
	return out, rows.Err()
}

type creaIndexEntry struct {
	code string
	url  string
}

func fetchCREAIndex(ctx context.Context, opts CREALiveOpts) ([]creaIndexEntry, error) {
	body, err := httpGetBytes(ctx, opts.HTTPClient, opts.IndexURL, opts.UserAgent)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	out := []creaIndexEntry{}
	seen := map[string]bool{}
	doc.Find(`a[href^="/tabelle-nutrizionali/"]`).Each(func(_ int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		// target solo NNNNNN (6 cifre)
		base := strings.TrimPrefix(href, "/tabelle-nutrizionali/")
		if len(base) != 6 {
			return
		}
		if _, err := strconv.Atoi(base); err != nil {
			return
		}
		if seen[base] {
			return
		}
		seen[base] = true
		out = append(out, creaIndexEntry{code: base, url: opts.BaseURL + href})
	})
	return out, nil
}

func fetchCREADetail(ctx context.Context, opts CREALiveOpts, e creaIndexEntry) (*FoodRow, error) {
	body, err := httpGetBytes(ctx, opts.HTTPClient, e.url, opts.UserAgent)
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	row := &FoodRow{
		ExternalID: e.code,
		URL:        e.url,
		EdiblePct:  100.0,
		FetchedAt:  time.Now().UTC(),
	}

	// Nome: <title>AlimentiNUTrizione - {Nome}</title>
	title := strings.TrimSpace(doc.Find("title").First().Text())
	if idx := strings.Index(title, " - "); idx >= 0 {
		row.NameIT = strings.TrimSpace(title[idx+3:])
	} else {
		row.NameIT = title
	}
	if row.NameIT == "" {
		return nil, fmt.Errorf("nome assente")
	}

	// Header table: Categoria, Parte Edibile
	doc.Find("table tr").Each(func(_ int, tr *goquery.Selection) {
		tds := tr.Find("td")
		if tds.Length() < 2 {
			return
		}
		label := strings.TrimSpace(tds.Eq(0).Text())
		value := strings.TrimSpace(tds.Eq(1).Text())
		switch label {
		case "Categoria":
			row.Category = value
		case "Parte Edibile":
			if m := crearNumRe.FindString(value); m != "" {
				if v, err := strconv.ParseFloat(strings.ReplaceAll(m, ",", "."), 64); err == nil {
					row.EdiblePct = v
				}
			}
		}
	})

	// Nutrient rows: <tr class="corponutriente|corpominerali|...">
	doc.Find("tr.corponutriente").Each(func(_ int, tr *goquery.Selection) {
		tds := tr.Find("td")
		if tds.Length() < 3 {
			return
		}
		label := strings.TrimSpace(tds.Eq(0).Text())
		valueText := strings.TrimSpace(tds.Eq(2).Text())
		setter, ok := creaNutrientMap[label]
		if !ok {
			return
		}
		m := crearNumRe.FindString(valueText)
		if m == "" {
			return
		}
		v, err := strconv.ParseFloat(strings.ReplaceAll(m, ",", "."), 64)
		if err != nil {
			return
		}
		setter(row, v)
	})
	// Sodio (per salt_g) sta in tr.corpominerali
	doc.Find("tr.corpominerali").Each(func(_ int, tr *goquery.Selection) {
		tds := tr.Find("td")
		if tds.Length() < 3 {
			return
		}
		label := strings.TrimSpace(tds.Eq(0).Text())
		if label != "Sodio (mg)" {
			return
		}
		valueText := strings.TrimSpace(tds.Eq(2).Text())
		m := crearNumRe.FindString(valueText)
		if m == "" {
			return
		}
		v, err := strconv.ParseFloat(strings.ReplaceAll(m, ",", "."), 64)
		if err != nil {
			return
		}
		setter := creaNutrientMap["Sodio (mg)"]
		if setter != nil {
			setter(row, v)
		}
	})

	raw, _ := json.Marshal(map[string]interface{}{
		"code": e.code, "url": e.url, "name": row.NameIT, "category": row.Category,
		"edible_pct": row.EdiblePct, "kcal": row.Kcal, "carb_g": row.CarbG,
		"protein_g": row.ProteinG, "fat_g": row.FatG, "fiber_g": row.FiberG,
		"sugar_g": row.SugarG, "water_g": row.WaterG, "salt_g": row.SaltG,
		"alcohol_g": row.AlcoholG,
	})
	hash := sha256.Sum256(raw)
	row.RawJSON = raw
	row.ContentHash = hex.EncodeToString(hash[:])
	return row, nil
}

func httpGetBytes(ctx context.Context, client *http.Client, url, ua string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
