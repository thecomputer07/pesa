package importer

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// wikiSkipCategoriesSQL sono le categorie CREA escluse dall'arricchimento
// Wikipedia (stesso elenco della migrazione 00029).
const wikiSkipCategoriesSQL = `'Carni fresche', 'Frattaglie', 'Prodotti della pesca'`

// WikiImageOpts controlla l'arricchimento immagini via Wikipedia REST API.
type WikiImageOpts struct {
	Limit      int           // 0 = tutti i food senza immagine
	Lang       string        // default "it"
	RateDelay  time.Duration // default 250ms (Wikipedia limit ~10 req/s)
	UserAgent  string
	HTTPClient *http.Client
	OnlySource string // se non vuoto, filtra per food_sources.code
}

// WikiImageReport è il sommario di un run.
type WikiImageReport struct {
	Total   int
	Found   int
	Missing int
	Errors  []string
}

// EnrichImagesFromWikipedia cerca su it.wikipedia.org una thumbnail per ogni
// food senza immagine e la inserisce in food_images.
func EnrichImagesFromWikipedia(ctx context.Context, db *sql.DB, opts WikiImageOpts) (*WikiImageReport, error) {
	if opts.Lang == "" {
		opts.Lang = "it"
	}
	if opts.RateDelay <= 0 {
		opts.RateDelay = 250 * time.Millisecond
	}
	if opts.UserAgent == "" {
		opts.UserAgent = "Pesa/0.1.0 (+https://pesadia.com; lorenzo.dimare@icloud.com)"
	}
	if opts.HTTPClient == nil {
		opts.HTTPClient = &http.Client{Timeout: 20 * time.Second}
	}

	// Query: food senza alcuna immagine, escludendo i superseded. Mai i
	// prodotti OFF: il nome contiene la marca e Wikipedia restituisce loghi,
	// stemmi o quadri omonimi ("Angelo Parodi" -> un angelo dipinto).
	// Nemmeno carne fresca, frattaglie e pesce: la voce Wikipedia illustra
	// l'animale vivo o uno schema anatomico, non il cibo nel piatto.
	q := `
		SELECT f.id, f.name_it
		  FROM foods f
		  LEFT JOIN food_images fi ON fi.food_id = f.id
		  JOIN food_sources fs ON fs.id = f.source_id
		 WHERE fi.id IS NULL
		   AND f.superseded_by IS NULL
		   AND fs.code <> 'off'
		   AND COALESCE(f.category, '') NOT IN (` + wikiSkipCategoriesSQL + `)`
	args := []interface{}{}
	if opts.OnlySource != "" {
		q += ` AND fs.code = $1`
		args = append(args, opts.OnlySource)
	}
	q += ` ORDER BY f.id`
	if opts.Limit > 0 {
		q += fmt.Sprintf(` LIMIT %d`, opts.Limit)
	}

	rows, err := db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	type item struct {
		id   int64
		name string
	}
	items := []item{}
	for rows.Next() {
		var it item
		if err := rows.Scan(&it.id, &it.name); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	report := &WikiImageReport{Total: len(items)}
	slog.Info("wiki enrich start", "items", len(items), "lang", opts.Lang)

	for i, it := range items {
		if ctx.Err() != nil {
			return report, ctx.Err()
		}
		thumb, err := wikiThumbnail(ctx, opts, it.name)
		if err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("%d %s: %v", it.id, it.name, err))
		} else if thumb != "" {
			_, ierr := db.ExecContext(ctx, `
				INSERT INTO food_images (food_id, url, attribution, license, is_external)
				VALUES ($1, $2, 'Wikimedia Commons (via Wikipedia '||$3||')', 'per-file (CC0/CC-BY/CC-BY-SA)', TRUE)`,
				it.id, thumb, opts.Lang)
			if ierr != nil {
				report.Errors = append(report.Errors, fmt.Sprintf("%d insert: %v", it.id, ierr))
			} else {
				report.Found++
			}
		} else {
			report.Missing++
		}
		if (i+1)%50 == 0 {
			slog.Info("wiki enrich progress", "done", i+1, "found", report.Found, "missing", report.Missing)
		}
		select {
		case <-ctx.Done():
			return report, ctx.Err()
		case <-time.After(opts.RateDelay):
		}
	}
	return report, nil
}

// wikiThumbnail prova varianti del nome per trovare una thumbnail su Wikipedia.
// Restituisce "" se nulla trovato (no errore).
func wikiThumbnail(ctx context.Context, opts WikiImageOpts, name string) (string, error) {
	for _, candidate := range nameCandidates(name) {
		thumb, err := wikiSummaryThumbnail(ctx, opts, candidate)
		if err != nil {
			continue
		}
		if thumb != "" {
			return thumb, nil
		}
	}
	return "", nil
}

// nameCandidates genera fino a 3 varianti del nome da provare su Wikipedia,
// dalla più specifica alla più generica.
func nameCandidates(name string) []string {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	out := []string{}
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		for _, e := range out {
			if strings.EqualFold(e, s) {
				return
			}
		}
		out = append(out, s)
	}
	add(name)
	// strip everything after comma (e.g. "Mela, fresca" → "Mela")
	if idx := strings.Index(name, ","); idx > 0 {
		add(name[:idx])
	}
	// brand prefix "Brand — Nome" → "Nome"
	if idx := strings.Index(name, " — "); idx > 0 {
		add(name[idx+len(" — "):])
	}
	// first 2 words
	parts := strings.Fields(name)
	if len(parts) >= 2 {
		add(parts[0] + " " + parts[1])
	}
	if len(parts) >= 1 {
		add(parts[0])
	}
	return out
}

type wikiSummary struct {
	Type      string `json:"type"`
	Thumbnail *struct {
		Source string `json:"source"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"thumbnail"`
}

func wikiSummaryThumbnail(ctx context.Context, opts WikiImageOpts, title string) (string, error) {
	u := fmt.Sprintf("https://%s.wikipedia.org/api/rest_v1/page/summary/%s",
		opts.Lang, url.PathEscape(strings.ReplaceAll(title, " ", "_")))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", opts.UserAgent)
	req.Header.Set("Accept", "application/json")
	resp, err := opts.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == 404 {
		return "", nil
	}
	if resp.StatusCode != 200 {
		_, _ = io.Copy(io.Discard, resp.Body)
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	var s wikiSummary
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return "", err
	}
	// Skip disambiguation and standard pages without image.
	if s.Type == "disambiguation" || s.Thumbnail == nil {
		return "", nil
	}
	// SVG thumbnails are logos, crests or diagrams, never food photos.
	if strings.Contains(strings.ToLower(s.Thumbnail.Source), ".svg") {
		return "", nil
	}
	return s.Thumbnail.Source, nil
}
