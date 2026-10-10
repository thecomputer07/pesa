package importer

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ScannedOFF is a product read via barcode that is not yet in the catalog.
type ScannedOFF struct {
	Code     string // EAN as scanned
	Name     string
	Brand    string
	Category string
	Kcal     *float64
	ProteinG *float64
	FatG     *float64
	CarbG    *float64
	SugarG   *float64
	FiberG   *float64
	ImageURL string // local /static/images/... when already cached
}

// FindOFFFoodByCode returns the catalog food for an OFF barcode, or 0.
// Leading zeros are ignored: bulk imports store OFF's code ("80177173"),
// the scanner pads to 13 digits ("0000080177173").
func FindOFFFoodByCode(ctx context.Context, db *sql.DB, code string) (int64, error) {
	var id int64
	err := db.QueryRowContext(ctx, `
		SELECT f.id
		  FROM foods f
		  JOIN food_source_records r ON r.id = f.source_record_id
		  JOIN food_sources s ON s.id = r.source_id
		 WHERE s.code = 'off'
		   AND ltrim(r.external_id, '0') = ltrim($1, '0')
		   AND f.superseded_by IS NULL
		 ORDER BY f.id
		 LIMIT 1`, code).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return id, err
}

// AddScannedOFF adds a scanned product to the shared catalog as an Open Food
// Facts food (with its attribution) and returns its id. Products without
// carbohydrates are refused: a carb counter must never default them to 0.
func AddScannedOFF(ctx context.Context, db *sql.DB, p ScannedOFF) (int64, error) {
	if p.CarbG == nil {
		return 0, errors.New("scanned product has no carbohydrates")
	}
	if id, err := FindOFFFoodByCode(ctx, db, p.Code); err != nil || id > 0 {
		return id, err
	}
	sourceID, err := lookupSourceID(ctx, db, "off")
	if err != nil {
		return 0, err
	}
	name := strings.TrimSpace(p.Name)
	if p.Brand != "" {
		// same shape as the bulk import: "Brand — Product"
		name = p.Brand + " — " + name
	}
	if len(name) > 255 {
		name = name[:255]
	}
	cat := p.Category
	if len(cat) > 64 {
		cat = cat[:64]
	}
	raw, _ := json.Marshal(p)
	hash := sha256.Sum256(raw)
	row := FoodRow{
		ExternalID:  p.Code,
		URL:         "https://world.openfoodfacts.org/product/" + p.Code,
		NameIT:      name,
		Category:    cat,
		EdiblePct:   100,
		Kcal:        p.Kcal,
		ProteinG:    p.ProteinG,
		FatG:        p.FatG,
		CarbG:       p.CarbG,
		SugarG:      p.SugarG,
		FiberG:      p.FiberG,
		ContentHash: hex.EncodeToString(hash[:]),
		FetchedAt:   time.Now().UTC(),
		RawJSON:     raw,
	}
	report := &ImportReport{}
	if err := importOneRow(ctx, db, sourceID, row, ImportOpts{SourceCode: "off", Basis: BasisEdible}, report, false); err != nil {
		return 0, fmt.Errorf("add scanned off: %w", err)
	}
	if report.Quarantined > 0 {
		return 0, errors.New("scanned product failed sanity checks")
	}
	id, err := FindOFFFoodByCode(ctx, db, p.Code)
	if err != nil || id == 0 {
		return 0, fmt.Errorf("add scanned off: not found after insert: %v", err)
	}
	if p.ImageURL != "" {
		_, err = db.ExecContext(ctx, `
			INSERT INTO food_images (food_id, url, attribution, license, is_external)
			VALUES ($1, $2, 'Open Food Facts contributors', 'ODbL-1.0', $3)`,
			id, p.ImageURL, !strings.HasPrefix(p.ImageURL, "/static/"))
		if err != nil {
			return id, nil // the food is in; a missing photo is not fatal
		}
	}
	return id, nil
}
