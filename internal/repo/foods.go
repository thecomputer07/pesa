package repo

import (
	"context"
	"database/sql"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// Food is the repository-layer DTO for /api/v1/foods endpoints.
type Food struct {
	ID              int64
	SourceID        int64
	SourceCode      string
	SourceName      string
	License         string
	LicenseURL      sql.NullString
	AttributionText string
	SourceURL       sql.NullString
	NameIT          string
	NameEN          sql.NullString
	Synonyms        sql.NullString
	Category        sql.NullString
	State           string
	EdiblePct       float64
	Kcal            sql.NullFloat64
	ProteinG        sql.NullFloat64
	FatG            sql.NullFloat64
	CarbG           sql.NullFloat64
	SugarG          sql.NullFloat64
	FiberG          sql.NullFloat64
	WaterG          sql.NullFloat64
	AshG            sql.NullFloat64
	AlcoholG        sql.NullFloat64
	SaltG           sql.NullFloat64
	LastImportedAt  sql.NullTime
	ParentFoodID    sql.NullInt64
	ThumbURL        sql.NullString
	Score           float64
}

// FoodImage is the external-image row shape.
type FoodImage struct {
	URL         string
	Attribution sql.NullString
	License     sql.NullString
	IsExternal  bool
}

// FoodSibling è una voce minimale cotto/crudo collegata via parent_food_id.
type FoodSibling struct {
	ID     int64
	NameIT string
	State  string
}

// LoadSiblings restituisce i food collegati via parent_food_id:
//   - se foodID è un "child" (ha parent_food_id), include il parent e gli altri child
//   - se foodID è un "parent", include tutti i suoi child
// Esclude sempre il food stesso e le righe superseded_by non nulle.
func LoadSiblings(ctx context.Context, db *sql.DB, foodID int64, parentFoodID sql.NullInt64) ([]FoodSibling, error) {
	var rootID int64
	if parentFoodID.Valid {
		rootID = parentFoodID.Int64
	} else {
		rootID = foodID
	}
	rows, err := db.QueryContext(ctx, `
		SELECT id, name_it, state FROM foods
		 WHERE (id = $1 OR parent_food_id = $1)
		   AND id <> $2
		   AND superseded_by IS NULL
		 ORDER BY state, id`, rootID, foodID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []FoodSibling{}
	for rows.Next() {
		var s FoodSibling
		if err := rows.Scan(&s.ID, &s.NameIT, &s.State); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// IT stopwords stripped before search lookup.
var itSearchStopwords = map[string]struct{}{
	"di": {}, "del": {}, "della": {}, "dello": {}, "dei": {}, "degli": {},
	"delle": {}, "al": {}, "alla": {}, "allo": {}, "agli": {}, "alle": {},
	"con": {}, "e": {}, "o": {}, "la": {}, "il": {}, "lo": {}, "gli": {},
	"le": {}, "un": {}, "una": {}, "uno": {}, "in": {}, "da": {}, "a": {},
	"per": {},
}

// BuildBooleanQuery normalises a user query for pg_trgm / ILIKE search:
//  1. lowercase + unicode NFC
//  2. drop Italian stopwords
//  3. return the plain concatenated token string (lowercased, trimmed)
//
// Returns the normalised query string plus the first non-stopword token
// (used downstream for prefix boost).
func BuildBooleanQuery(raw string) (normQ, firstToken string) {
	s := strings.ToLower(strings.TrimSpace(raw))
	s = norm.NFC.String(s)
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			b.WriteRune(' ')
		}
	}
	toks := strings.Fields(b.String())
	kept := toks[:0]
	for _, t := range toks {
		if _, skip := itSearchStopwords[t]; skip {
			continue
		}
		kept = append(kept, t)
	}
	if len(kept) == 0 {
		return "", ""
	}
	return strings.Join(kept, " "), kept[0]
}

// SearchFoods runs pg_trgm ILIKE + ranking, excluding superseded rows.
// Falls back to ILIKE-based search when the primary query returns no results
// (small datasets or very short queries).
//
// userID=0 means anonymous: solo alimenti condivisi (user_id IS NULL).
// userID>0 include anche i custom food dell'utente.
func SearchFoods(ctx context.Context, db *sql.DB, query string, limit, page int) ([]Food, int, error) {
	return SearchFoodsForUser(ctx, db, 0, query, limit, page)
}

// SearchFoodsForUser è come SearchFoods ma include gli alimenti privati dell'utente.
func SearchFoodsForUser(ctx context.Context, db *sql.DB, userID int64, query string, limit, page int) ([]Food, int, error) {
	normQ, firstTok := BuildBooleanQuery(query)
	if normQ == "" {
		return nil, 0, nil
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	prefixLike := firstTok + "%"
	wordLike := "% " + firstTok + " %"
	ilikePat := "%" + normQ + "%"

	selectSQL := `
	SELECT f.id, f.source_id, s.code, s.name, s.license, s.license_url,
	       s.attribution_text, s.source_url,
	       f.name_it, f.name_en, f.synonyms, f.category, f.state,
	       f.edible_pct, f.kcal, f.protein_g, f.fat_g, f.carb_g, f.sugar_g,
	       f.fiber_g, f.water_g, f.ash_g, f.alcohol_g, f.salt_g,
	       f.last_imported_at,
	       (SELECT url FROM food_images WHERE food_id = f.id ORDER BY id LIMIT 1) AS thumb_url,
	       similarity(f.name_it, $1) AS relevance,
	       (CASE WHEN f.name_it ILIKE $2 THEN 5 ELSE 0 END) AS prefix_boost,
	       (CASE WHEN f.name_it = $3 OR f.name_it ILIKE $4 THEN 2 ELSE 0 END) AS exact_boost
	  FROM foods f
	  JOIN food_sources s ON s.id = f.source_id
	 WHERE (coalesce(f.name_it,'') || ' ' || coalesce(f.name_en,'') || ' ' || coalesce(f.synonyms,'')) ILIKE $5
	   AND f.superseded_by IS NULL
	   AND (f.user_id IS NULL OR f.user_id = $8)
	 ORDER BY (similarity(f.name_it, $1) + (CASE WHEN f.name_it ILIKE $2 THEN 5 ELSE 0 END) + (CASE WHEN f.name_it = $3 OR f.name_it ILIKE $4 THEN 2 ELSE 0 END)) DESC, f.name_it ASC
	 LIMIT $6 OFFSET $7`

	rows, err := db.QueryContext(ctx, selectSQL, normQ, prefixLike, firstTok, wordLike, ilikePat, limit, offset, userID)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	out := []Food{}
	for rows.Next() {
		var f Food
		var prefixB, exactB int
		if err := rows.Scan(
			&f.ID, &f.SourceID, &f.SourceCode, &f.SourceName, &f.License, &f.LicenseURL,
			&f.AttributionText, &f.SourceURL,
			&f.NameIT, &f.NameEN, &f.Synonyms, &f.Category, &f.State,
			&f.EdiblePct, &f.Kcal, &f.ProteinG, &f.FatG, &f.CarbG, &f.SugarG,
			&f.FiberG, &f.WaterG, &f.AshG, &f.AlcoholG, &f.SaltG,
			&f.LastImportedAt, &f.ThumbURL, &f.Score, &prefixB, &exactB,
		); err != nil {
			return nil, 0, err
		}
		f.Score += float64(prefixB + exactB)
		out = append(out, f)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	countSQL := `
	SELECT COUNT(*) FROM foods f
	 WHERE (coalesce(f.name_it,'') || ' ' || coalesce(f.name_en,'') || ' ' || coalesce(f.synonyms,'')) ILIKE $1
	   AND f.superseded_by IS NULL
	   AND (f.user_id IS NULL OR f.user_id = $2)`
	if err := db.QueryRowContext(ctx, countSQL, ilikePat, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Fallback: if primary search returned nothing, retry with ILIKE on first token.
	// Helps with small datasets and very short queries.
	if total == 0 {
		return searchFoodsLike(ctx, db, userID, firstTok, query, limit, offset)
	}
	return out, total, nil
}

// searchFoodsLike is a forgiving fallback: matches tokens via ILIKE %tok%
// against name_it / name_en / synonyms. No relevance ranking, alphabetic order.
func searchFoodsLike(ctx context.Context, db *sql.DB, userID int64, firstTok, raw string, limit, offset int) ([]Food, int, error) {
	// Use the lowercased+normalised first token; if empty, use the raw query trimmed.
	needle := firstTok
	if needle == "" {
		needle = strings.ToLower(strings.TrimSpace(raw))
	}
	if needle == "" {
		return nil, 0, nil
	}
	pattern := "%" + needle + "%"

	selectSQL := `
	SELECT f.id, f.source_id, s.code, s.name, s.license, s.license_url,
	       s.attribution_text, s.source_url,
	       f.name_it, f.name_en, f.synonyms, f.category, f.state,
	       f.edible_pct, f.kcal, f.protein_g, f.fat_g, f.carb_g, f.sugar_g,
	       f.fiber_g, f.water_g, f.ash_g, f.alcohol_g, f.salt_g,
	       f.last_imported_at
	  FROM foods f
	  JOIN food_sources s ON s.id = f.source_id
	 WHERE (f.name_it ILIKE $1 OR f.name_en ILIKE $1 OR f.synonyms ILIKE $1)
	   AND f.superseded_by IS NULL
	   AND (f.user_id IS NULL OR f.user_id = $4)
	 ORDER BY f.name_it ASC
	 LIMIT $2 OFFSET $3`

	rows, err := db.QueryContext(ctx, selectSQL, pattern, limit, offset, userID)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	out := []Food{}
	for rows.Next() {
		var f Food
		if err := rows.Scan(
			&f.ID, &f.SourceID, &f.SourceCode, &f.SourceName, &f.License, &f.LicenseURL,
			&f.AttributionText, &f.SourceURL,
			&f.NameIT, &f.NameEN, &f.Synonyms, &f.Category, &f.State,
			&f.EdiblePct, &f.Kcal, &f.ProteinG, &f.FatG, &f.CarbG, &f.SugarG,
			&f.FiberG, &f.WaterG, &f.AshG, &f.AlcoholG, &f.SaltG,
			&f.LastImportedAt,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, f)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM foods
		  WHERE (name_it ILIKE $1 OR name_en ILIKE $1 OR synonyms ILIKE $1)
		    AND superseded_by IS NULL
		    AND (user_id IS NULL OR user_id = $2)`,
		pattern, userID).Scan(&total); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// GetFoodByID returns a single food row (with source denorm) + its images,
// treating superseded rows as non-existent (returns (nil, nil, nil)).
func GetFoodByID(ctx context.Context, db *sql.DB, id int64) (*Food, []FoodImage, error) {
	row := db.QueryRowContext(ctx, `
		SELECT f.id, f.source_id, s.code, s.name, s.license, s.license_url,
		       s.attribution_text, s.source_url,
		       f.name_it, f.name_en, f.synonyms, f.category, f.state,
		       f.edible_pct, f.kcal, f.protein_g, f.fat_g, f.carb_g, f.sugar_g,
		       f.fiber_g, f.water_g, f.ash_g, f.alcohol_g, f.salt_g,
		       f.last_imported_at, f.superseded_by, f.parent_food_id
		  FROM foods f
		  JOIN food_sources s ON s.id = f.source_id
		 WHERE f.id = $1`, id)

	var f Food
	var superseded sql.NullInt64
	err := row.Scan(&f.ID, &f.SourceID, &f.SourceCode, &f.SourceName, &f.License, &f.LicenseURL,
		&f.AttributionText, &f.SourceURL,
		&f.NameIT, &f.NameEN, &f.Synonyms, &f.Category, &f.State,
		&f.EdiblePct, &f.Kcal, &f.ProteinG, &f.FatG, &f.CarbG, &f.SugarG,
		&f.FiberG, &f.WaterG, &f.AshG, &f.AlcoholG, &f.SaltG,
		&f.LastImportedAt, &superseded, &f.ParentFoodID)
	if err == sql.ErrNoRows {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	if superseded.Valid {
		return nil, nil, nil
	}

	imgs, err := loadImages(ctx, db, id)
	if err != nil {
		return nil, nil, err
	}
	return &f, imgs, nil
}

func loadImages(ctx context.Context, db *sql.DB, foodID int64) ([]FoodImage, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT url, attribution, license, is_external
		  FROM food_images
		 WHERE food_id = $1
		 ORDER BY id ASC`, foodID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []FoodImage{}
	for rows.Next() {
		var im FoodImage
		if err := rows.Scan(&im.URL, &im.Attribution, &im.License, &im.IsExternal); err != nil {
			return nil, err
		}
		out = append(out, im)
	}
	return out, rows.Err()
}
