package repo

import (
	"context"
	"database/sql"
	"time"
)

// CatalogColumns is the order of the values in every CatalogRow, sent once in
// the response so the payload stays compact (an array per food, no keys).
var CatalogColumns = []string{
	"id", "name_it", "name_en", "synonyms", "category", "state", "source",
	"parent_id", "edible_pct", "kcal", "protein_g", "fat_g", "carb_g",
	"sugar_g", "fiber_g", "water_g", "ash_g", "alcohol_g", "salt_g",
	"has_thumb", "ean", "last_imported_date",
}

// CatalogRow is one public food, values in CatalogColumns order.
type CatalogRow []any

// CatalogSource is the attribution shared by all foods of a source.
type CatalogSource struct {
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	License     string  `json:"license"`
	LicenseURL  *string `json:"license_url,omitempty"`
	Attribution string  `json:"attribution"`
	SourceURL   *string `json:"source_url,omitempty"`
}

// CatalogState identifies a version of the public catalog: the latest change
// (ms) and the number of live foods. Equal state = nothing to download.
type CatalogState struct {
	CursorMs int64
	Count    int
}

// publicFood: shared catalog rows only (never a user's custom food or recipe).
const publicFood = `f.user_id IS NULL AND f.recipe_id IS NULL`

// LoadCatalogState returns the current CatalogState (two index-friendly aggregates).
func LoadCatalogState(ctx context.Context, db *sql.DB) (CatalogState, error) {
	var st CatalogState
	var maxUpd sql.NullTime
	err := db.QueryRowContext(ctx, `
		SELECT max(f.updated_at),
		       count(*) FILTER (WHERE f.superseded_by IS NULL)
		  FROM foods f
		 WHERE `+publicFood).Scan(&maxUpd, &st.Count)
	if err != nil {
		return st, err
	}
	if maxUpd.Valid {
		st.CursorMs = maxUpd.Time.UnixMilli()
	}
	return st, nil
}

// LoadCatalog returns the live public foods changed after sinceMs (all of them
// when sinceMs is 0) and the ids that left the catalog after sinceMs
// (superseded). Hard-deleted rows are not reported: the client compares its
// row count with CatalogState.Count and downloads everything again on mismatch.
func LoadCatalog(ctx context.Context, db *sql.DB, sinceMs int64) ([]CatalogRow, []int64, error) {
	since := time.UnixMilli(sinceMs).UTC()
	rows, err := db.QueryContext(ctx, `
		SELECT f.id, f.name_it, f.name_en, f.synonyms, f.category, f.state, s.code,
		       f.parent_food_id, f.edible_pct, f.kcal, f.protein_g, f.fat_g, f.carb_g,
		       f.sugar_g, f.fiber_g, f.water_g, f.ash_g, f.alcohol_g, f.salt_g,
		       im.url IS NOT NULL,
		       CASE WHEN s.code = 'off' THEN r.external_id END,
		       f.last_imported_at,
		       f.superseded_by IS NOT NULL
		  FROM foods f
		  JOIN food_sources s ON s.id = f.source_id
		  LEFT JOIN food_source_records r ON r.id = f.source_record_id
		  LEFT JOIN LATERAL (SELECT url FROM food_images
		                      WHERE food_id = f.id ORDER BY id LIMIT 1) im ON true
		 WHERE `+publicFood+`
		   AND ($1::bigint = 0 OR f.updated_at > $2)
		   AND ($1::bigint > 0 OR f.superseded_by IS NULL)
		 ORDER BY f.id`, sinceMs, since)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()

	out := []CatalogRow{}
	gone := []int64{}
	for rows.Next() {
		var (
			id                                  int64
			nameIT, state, src                  string
			nameEN, syn, cat, ean               sql.NullString
			hasThumb                            bool
			parent                              sql.NullInt64
			edible                              float64
			kcal, prot, fat, carb, sugar, fiber sql.NullFloat64
			water, ash, alcohol, salt           sql.NullFloat64
			imported                            sql.NullTime
			superseded                          bool
		)
		if err := rows.Scan(&id, &nameIT, &nameEN, &syn, &cat, &state, &src,
			&parent, &edible, &kcal, &prot, &fat, &carb,
			&sugar, &fiber, &water, &ash, &alcohol, &salt,
			&hasThumb, &ean, &imported, &superseded); err != nil {
			return nil, nil, err
		}
		if superseded {
			gone = append(gone, id)
			continue
		}
		var imp any
		if imported.Valid {
			imp = imported.Time.UTC().Format("2006-01-02")
		}
		out = append(out, CatalogRow{
			id, nameIT, nullStr(nameEN), nullStr(syn), nullStr(cat), state, src,
			nullInt(parent), edible, nullF(kcal), nullF(prot), nullF(fat), nullF(carb),
			nullF(sugar), nullF(fiber), nullF(water), nullF(ash), nullF(alcohol), nullF(salt),
			hasThumb, nullStr(ean), imp,
		})
	}
	return out, gone, rows.Err()
}

// LoadCatalogSources returns the attribution of every food source.
func LoadCatalogSources(ctx context.Context, db *sql.DB) ([]CatalogSource, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT code, name, license, license_url, attribution_text, source_url
		  FROM food_sources ORDER BY code`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []CatalogSource{}
	for rows.Next() {
		var s CatalogSource
		var lu, su sql.NullString
		if err := rows.Scan(&s.Code, &s.Name, &s.License, &lu, &s.Attribution, &su); err != nil {
			return nil, err
		}
		if lu.Valid {
			s.LicenseURL = &lu.String
		}
		if su.Valid {
			s.SourceURL = &su.String
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func nullStr(v sql.NullString) any {
	if v.Valid {
		return v.String
	}
	return nil
}

func nullInt(v sql.NullInt64) any {
	if v.Valid {
		return v.Int64
	}
	return nil
}

func nullF(v sql.NullFloat64) any {
	if v.Valid {
		return v.Float64
	}
	return nil
}
