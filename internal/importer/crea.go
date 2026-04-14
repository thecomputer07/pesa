package importer

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

// crea CSV column order (must mirror cmd/pesa/scrape.go recordToCSV).
const (
	colExternalID = iota
	colURL
	colNameIT
	colCategory
	colEdiblePct
	colKcal
	colProteinG
	colFatG
	colCarbG
	colSugarG
	colFiberG
	colWaterG
	colAshG
	colAlcoholG
	colSaltG
	colContentHash
	colFetchedAt
	colCount
)

// ReadCREACSV parses a plan-02 CSV into FoodRow slice.
func ReadCREACSV(path string) ([]FoodRow, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	header, err := r.Read()
	if err != nil {
		return nil, err
	}
	if len(header) < colCount {
		return nil, fmt.Errorf("csv header: got %d cols, want %d", len(header), colCount)
	}

	rows := []FoodRow{}
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		for len(rec) < colCount {
			rec = append(rec, "")
		}
		fr := FoodRow{
			ExternalID:  rec[colExternalID],
			URL:         rec[colURL],
			NameIT:      rec[colNameIT],
			Category:    rec[colCategory],
			EdiblePct:   parseFloatOr(rec[colEdiblePct], 100.0),
			Kcal:        parseFloatPtr(rec[colKcal]),
			ProteinG:    parseFloatPtr(rec[colProteinG]),
			FatG:        parseFloatPtr(rec[colFatG]),
			CarbG:       parseFloatPtr(rec[colCarbG]),
			SugarG:      parseFloatPtr(rec[colSugarG]),
			FiberG:      parseFloatPtr(rec[colFiberG]),
			WaterG:      parseFloatPtr(rec[colWaterG]),
			AshG:        parseFloatPtr(rec[colAshG]),
			AlcoholG:    parseFloatPtr(rec[colAlcoholG]),
			SaltG:       parseFloatPtr(rec[colSaltG]),
			ContentHash: rec[colContentHash],
		}
		if t, terr := time.Parse(time.RFC3339, rec[colFetchedAt]); terr == nil {
			fr.FetchedAt = t
		} else {
			fr.FetchedAt = time.Now().UTC()
		}
		raw, _ := json.Marshal(fr)
		fr.RawJSON = raw
		rows = append(rows, fr)
	}
	return rows, nil
}

func parseFloatPtr(s string) *float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	v, err := strconv.ParseFloat(strings.ReplaceAll(s, ",", "."), 64)
	if err != nil {
		return nil
	}
	return &v
}

func parseFloatOr(s string, def float64) float64 {
	if p := parseFloatPtr(s); p != nil {
		return *p
	}
	return def
}

// lookupSourceID returns the food_sources.id for a code ("crea","bda-ieo","manual").
func lookupSourceID(ctx context.Context, db *sql.DB, code string) (int64, error) {
	var id int64
	err := db.QueryRowContext(ctx, `SELECT id FROM food_sources WHERE code = $1`, code).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("source %q: %w", code, err)
	}
	return id, nil
}

// ImportCSV loads a CREA-format CSV into the DB with sanity check + quarantine.
//
// Idempotent: content_hash is used to short-circuit unchanged rows.
func ImportCSV(ctx context.Context, db *sql.DB, csvPath string, opts ImportOpts) (*ImportReport, error) {
	if opts.SourceCode == "" {
		opts.SourceCode = "crea"
	}
	if opts.Basis == "" {
		opts.Basis = BasisEdible
	}
	rows, err := ReadCREACSV(csvPath)
	if err != nil {
		return nil, err
	}
	report := &ImportReport{TotalRows: len(rows)}

	sourceID, err := lookupSourceID(ctx, db, opts.SourceCode)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		if err := importOneRow(ctx, db, sourceID, row, opts, report, false); err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("%s: %v", row.ExternalID, err))
		}
	}
	return report, nil
}

// importOneRow is the shared per-row path used by CREA and BDA.
// supersedeLookup=true will consult the CREA match table to set superseded_by.
func importOneRow(ctx context.Context, db *sql.DB, sourceID int64, row FoodRow, opts ImportOpts, report *ImportReport, supersedeLookup bool) error {
	// Idempotency: if existing fsr with same content_hash, skip.
	var (
		existingRecID int64
		existingHash  string
	)
	err := db.QueryRowContext(ctx,
		`SELECT id, content_hash FROM food_source_records WHERE source_id=$1 AND external_id=$2`,
		sourceID, row.ExternalID).Scan(&existingRecID, &existingHash)
	switch {
	case err == sql.ErrNoRows:
		// insert new fsr below
	case err != nil:
		return err
	default:
		if existingHash == row.ContentHash {
			report.Unchanged++
			return nil
		}
	}

	if opts.DryRun {
		return nil
	}

	// UPSERT food_source_records.
	var recordID int64
	if existingRecID == 0 {
		err := db.QueryRowContext(ctx,
			`INSERT INTO food_source_records (source_id, external_id, raw_json, content_hash, fetched_at)
			 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			sourceID, row.ExternalID, row.RawJSON, row.ContentHash, row.FetchedAt).Scan(&recordID)
		if err != nil {
			return err
		}
	} else {
		_, err := db.ExecContext(ctx,
			`UPDATE food_source_records SET raw_json=$1, content_hash=$2, fetched_at=$3 WHERE id=$4`,
			row.RawJSON, row.ContentHash, row.FetchedAt, existingRecID)
		if err != nil {
			return err
		}
		recordID = existingRecID
	}

	norm := Normalize(row, opts.Basis)
	check := SanityCheck(norm)
	if !check.OK {
		_, err := db.ExecContext(ctx,
			`INSERT INTO imports_review (source_id, external_id, raw_json, reason, severity)
			 VALUES ($1, $2, $3, $4, $5)`,
			sourceID, row.ExternalID, row.RawJSON,
			strings.Join(check.Reasons, "; "), check.Severity)
		if err != nil {
			return err
		}
		report.Quarantined++
		return nil
	}

	var supersededBy sql.NullInt64
	if supersedeLookup {
		creaID, lerr := FindCREAMatch(ctx, db, NormalizeName(norm.NameIT), "raw")
		if lerr == nil && creaID > 0 {
			supersededBy = sql.NullInt64{Int64: creaID, Valid: true}
			report.SupersededByCREA++
		}
	}

	// UPSERT foods on (source_id, source_record_id).
	var existingFoodID int64
	err = db.QueryRowContext(ctx,
		`SELECT id FROM foods WHERE source_id=$1 AND source_record_id=$2`,
		sourceID, recordID).Scan(&existingFoodID)

	switch err {
	case sql.ErrNoRows:
		_, ierr := db.ExecContext(ctx, `
			INSERT INTO foods (source_id, source_record_id, name_it, category, state,
				edible_pct, kcal, protein_g, fat_g, carb_g, sugar_g, fiber_g,
				water_g, ash_g, alcohol_g, salt_g, last_imported_at, superseded_by)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)`,
			sourceID, recordID, norm.NameIT, nullString(norm.Category), "raw",
			norm.EdiblePct, nullFloat(norm.Kcal), nullFloat(norm.ProteinG),
			nullFloat(norm.FatG), nullFloat(norm.CarbG), nullFloat(norm.SugarG),
			nullFloat(norm.FiberG), nullFloat(norm.WaterG), nullFloat(norm.AshG),
			nullFloat(norm.AlcoholG), nullFloat(norm.SaltG),
			time.Now().UTC(), supersededBy)
		if ierr != nil {
			return ierr
		}
		report.InsertedFoods++
	case nil:
		_, uerr := db.ExecContext(ctx, `
			UPDATE foods SET name_it=$1, category=$2, edible_pct=$3, kcal=$4,
				protein_g=$5, fat_g=$6, carb_g=$7, sugar_g=$8, fiber_g=$9,
				water_g=$10, ash_g=$11, alcohol_g=$12, salt_g=$13, last_imported_at=$14,
				superseded_by=$15
			 WHERE id=$16`,
			norm.NameIT, nullString(norm.Category), norm.EdiblePct,
			nullFloat(norm.Kcal), nullFloat(norm.ProteinG), nullFloat(norm.FatG),
			nullFloat(norm.CarbG), nullFloat(norm.SugarG), nullFloat(norm.FiberG),
			nullFloat(norm.WaterG), nullFloat(norm.AshG), nullFloat(norm.AlcoholG),
			nullFloat(norm.SaltG), time.Now().UTC(), supersededBy, existingFoodID)
		if uerr != nil {
			return uerr
		}
		report.UpdatedFoods++
	default:
		return err
	}
	return nil
}

func nullFloat(p *float64) interface{} {
	if p == nil {
		return nil
	}
	return *p
}
func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
