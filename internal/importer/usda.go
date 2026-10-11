package importer

import (
	"context"
	"crypto/sha256"
	"database/sql"
	_ "embed"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// usdaNamesIT maps USDA FNDDS fdc_id -> Italian name and category, and like
// ciqual_it.tsv decides what is imported: only foods CREA and Ciqual lack
// (sushi by type, Japanese sides and sauces).
//
//go:embed data/usda_it.tsv
var usdaNamesIT []byte

type usdaName struct{ Name, Category string }

// ParseUSDANamesIT reads "fdc_id<TAB>name<TAB>category".
func ParseUSDANamesIT(data []byte) (map[string]usdaName, error) {
	out := map[string]usdaName{}
	for i, line := range strings.Split(string(data), "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		f := strings.Split(line, "\t")
		if len(f) != 3 || strings.TrimSpace(f[0]) == "" || strings.TrimSpace(f[1]) == "" {
			return nil, fmt.Errorf("usda_it.tsv line %d: want id<TAB>name<TAB>category", i+1)
		}
		out[strings.TrimSpace(f[0])] = usdaName{strings.TrimSpace(f[1]), strings.TrimSpace(f[2])}
	}
	return out, nil
}

// FNDDS nutrient numbers (food_nutrient.nutrient_id in the survey CSV).
const (
	usdaCarbByDiff = "205"
	usdaFiber      = "291"
	usdaSodiumMg   = "307"
)

var usdaSetters = map[string]func(*FoodRow, float64){
	"208": func(r *FoodRow, v float64) { r.Kcal = &v },
	"203": func(r *FoodRow, v float64) { r.ProteinG = &v },
	"204": func(r *FoodRow, v float64) { r.FatG = &v },
	"269": func(r *FoodRow, v float64) { r.SugarG = &v },
	"291": func(r *FoodRow, v float64) { r.FiberG = &v },
	"255": func(r *FoodRow, v float64) { r.WaterG = &v },
	"221": func(r *FoodRow, v float64) { r.AlcoholG = &v },
}

// eachCSVRow streams a CSV with a header row as maps.
func eachCSVRow(path string, fn func(map[string]string) error) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	r := csv.NewReader(f)
	r.ReuseRecord = true
	head, err := r.Read()
	if err != nil {
		return fmt.Errorf("%s: %w", filepath.Base(path), err)
	}
	cols := append([]string(nil), head...)
	for {
		rec, err := r.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("%s: %w", filepath.Base(path), err)
		}
		m := make(map[string]string, len(cols))
		for i, c := range cols {
			if i < len(rec) {
				m[c] = rec[i]
			}
		}
		if err := fn(m); err != nil {
			return err
		}
	}
}

// ReadUSDASurveyCSV parses the FoodData Central "survey food" (FNDDS) CSV
// release in dir (food.csv + food_nutrient.csv) for the ids in names.
//
// USDA reports carbohydrate *by difference*, fibre included; CREA and Ciqual
// report available carbohydrate. We store carb = by-difference - fibre so the
// three sources are comparable.
func ReadUSDASurveyCSV(dir string, names map[string]usdaName) ([]FoodRow, error) {
	desc := map[string]string{}
	err := eachCSVRow(filepath.Join(dir, "food.csv"), func(m map[string]string) error {
		if _, ok := names[m["fdc_id"]]; ok {
			desc[m["fdc_id"]] = m["description"]
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	for id := range names {
		if _, ok := desc[id]; !ok {
			return nil, fmt.Errorf("usda_it.tsv: fdc_id %s not in food.csv", id)
		}
	}

	rows := map[string]*FoodRow{}
	raw := map[string]map[string]float64{}
	for id, n := range names {
		rows[id] = &FoodRow{
			ExternalID: id,
			URL:        "https://fdc.nal.usda.gov/food-details/" + id + "/nutrients",
			NameIT:     n.Name,
			Category:   n.Category,
			EdiblePct:  100,
		}
		raw[id] = map[string]float64{}
	}
	err = eachCSVRow(filepath.Join(dir, "food_nutrient.csv"), func(m map[string]string) error {
		row, ok := rows[m["fdc_id"]]
		if !ok {
			return nil
		}
		v, err := strconv.ParseFloat(m["amount"], 64)
		if err != nil {
			return nil
		}
		nid := m["nutrient_id"]
		raw[m["fdc_id"]][nid] = v
		if set, ok := usdaSetters[nid]; ok {
			set(row, v)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	out := make([]FoodRow, 0, len(rows))
	for _, id := range sortedKeys(rows) {
		row, nut := rows[id], raw[id]
		if c, ok := nut[usdaCarbByDiff]; ok {
			avail := c - nut[usdaFiber]
			if avail < 0 {
				avail = 0
			}
			row.CarbG = &avail
		}
		if na, ok := nut[usdaSodiumMg]; ok {
			salt := na * 2.5 / 1000 // EU rule: salt = sodium x 2.5
			row.SaltG = &salt
		}
		b, _ := json.Marshal(map[string]interface{}{
			"fdc_id": id, "url": row.URL, "name_it": row.NameIT, "name_en": desc[id],
			"category": row.Category, "nutrients": nut,
		})
		sum := sha256.Sum256(b)
		row.RawJSON = b
		row.ContentHash = hex.EncodeToString(sum[:])
		row.FetchedAt = now
		out = append(out, *row)
	}
	return out, nil
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// ImportUSDA loads the subset listed in data/usda_it.tsv from the FNDDS CSV
// release in dir. Idempotent like ImportCSV (content_hash short-circuit).
func ImportUSDA(ctx context.Context, db *sql.DB, dir string, opts ImportOpts) (*ImportReport, error) {
	if opts.SourceCode == "" {
		opts.SourceCode = "usda"
	}
	if opts.Basis == "" {
		opts.Basis = BasisEdible
	}
	names, err := ParseUSDANamesIT(usdaNamesIT)
	if err != nil {
		return nil, err
	}
	rows, err := ReadUSDASurveyCSV(dir, names)
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
	if !opts.DryRun {
		// "Maki al salmone" does not contain the word: searching "sushi" must
		// still list every type.
		if _, err := db.ExecContext(ctx, `
			UPDATE foods f SET synonyms = 'sushi'
			  FROM food_sources s
			 WHERE s.id = f.source_id AND s.code = $1 AND f.category = 'Sushi'
			   AND f.synonyms IS DISTINCT FROM 'sushi'`, opts.SourceCode); err != nil {
			return report, fmt.Errorf("usda sushi synonyms: %w", err)
		}
	}
	return report, nil
}
