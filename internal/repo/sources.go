package repo

import (
	"context"
	"database/sql"
)

// FoodSource is the /api/v1/sources row shape consumed by attribution pages.
type FoodSource struct {
	Code            string
	Name            string
	License         string
	LicenseURL      sql.NullString
	AttributionText string
	SourceURL       sql.NullString
	Scope           sql.NullString
	Contact         sql.NullString
	Notes           sql.NullString
}

// ListFoodSources returns the food_sources actually in use, ordered by code:
// those backing at least one food, plus image providers (food_images has no
// source FK). Registered-but-empty sources (e.g. BDA before its import) are
// not attributed publicly.
func ListFoodSources(ctx context.Context, db *sql.DB) ([]FoodSource, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT code, name, license, license_url, attribution_text, source_url,
		       scope, contact, notes
		FROM food_sources fs
		WHERE fs.scope LIKE '%images%'
		   OR EXISTS (SELECT 1 FROM foods f WHERE f.source_id = fs.id)
		ORDER BY code ASC`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []FoodSource
	for rows.Next() {
		var s FoodSource
		if err := rows.Scan(&s.Code, &s.Name, &s.License, &s.LicenseURL,
			&s.AttributionText, &s.SourceURL, &s.Scope, &s.Contact, &s.Notes); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
