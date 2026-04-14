package importer

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tealeg/xlsx/v3"
)

// Format selects a BDA input parser.
type Format string

const (
	FormatAuto Format = ""
	FormatCSV  Format = "csv"
	FormatXLSX Format = "xlsx"
)

// ParseReport accumulates per-row parse outcomes.
type ParseReport struct {
	RowsParsed int
	RowsFailed int
	Errors     []string
}

// ParseBDA reads a BDA IEO file into FoodRow slice.
// Auto-detects format from extension when format == FormatAuto.
func ParseBDA(path string, format Format) ([]FoodRow, *ParseReport, error) {
	if format == FormatAuto {
		switch strings.ToLower(filepath.Ext(path)) {
		case ".xlsx":
			format = FormatXLSX
		default:
			format = FormatCSV
		}
	}
	switch format {
	case FormatCSV:
		return parseBDACSV(path)
	case FormatXLSX:
		return parseBDAXLSX(path)
	}
	return nil, nil, fmt.Errorf("unknown format: %s", format)
}

// BDA header synonyms (lowercased, accent-stripped) → FoodRow field.
var bdaHeaderMap = map[string]string{
	"id":                    "external_id",
	"codice":                "external_id",
	"nome alimento":         "name_it",
	"nome":                  "name_it",
	"alimento":              "name_it",
	"categoria":             "category",
	"parte edibile":         "edible_pct",
	"parte edibile (%)":     "edible_pct",
	"energia (kcal)":        "kcal",
	"kcal":                  "kcal",
	"proteine":              "protein_g",
	"proteine (g)":          "protein_g",
	"lipidi":                "fat_g",
	"lipidi (g)":            "fat_g",
	"grassi":                "fat_g",
	"carboidrati":           "carb_g",
	"carboidrati (g)":       "carb_g",
	"zuccheri":              "sugar_g",
	"fibra":                 "fiber_g",
	"fibra (g)":             "fiber_g",
	"acqua":                 "water_g",
	"ceneri":                "ash_g",
	"alcol":                 "alcohol_g",
	"sale":                  "salt_g",
	"sodio":                 "salt_g",
}

func mapHeaders(header []string) (map[string]int, error) {
	out := map[string]int{}
	for i, h := range header {
		key := NormalizeName(h)
		if field, ok := bdaHeaderMap[key]; ok {
			out[field] = i
		} else if field, ok := bdaHeaderMap[strings.ToLower(strings.TrimSpace(h))]; ok {
			out[field] = i
		}
	}
	if _, ok := out["name_it"]; !ok {
		return nil, fmt.Errorf("bda header missing name column; got %v", header)
	}
	return out, nil
}

func parseBDACSV(path string) ([]FoodRow, *ParseReport, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = f.Close() }()

	// Auto-detect comma vs tab by peeking.
	r := csv.NewReader(f)
	r.Comma = ','
	r.FieldsPerRecord = -1

	header, err := r.Read()
	if err != nil {
		return nil, nil, err
	}
	if len(header) == 1 && strings.Contains(header[0], "\t") {
		// Retry as TSV.
		_, _ = f.Seek(0, io.SeekStart)
		r = csv.NewReader(f)
		r.Comma = '\t'
		r.FieldsPerRecord = -1
		header, err = r.Read()
		if err != nil {
			return nil, nil, err
		}
	}

	idx, err := mapHeaders(header)
	if err != nil {
		return nil, nil, err
	}
	report := &ParseReport{}
	rows := []FoodRow{}
	lineNo := 1
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		lineNo++
		if err != nil {
			report.RowsFailed++
			report.Errors = append(report.Errors, fmt.Sprintf("line %d: %v", lineNo, err))
			continue
		}
		row := rowFromRecord(rec, idx)
		if row.NameIT == "" {
			report.RowsFailed++
			continue
		}
		rows = append(rows, row)
		report.RowsParsed++
	}
	return rows, report, nil
}

func parseBDAXLSX(path string) ([]FoodRow, *ParseReport, error) {
	wb, err := xlsx.OpenFile(path)
	if err != nil {
		return nil, nil, err
	}
	if len(wb.Sheets) == 0 {
		return nil, nil, fmt.Errorf("xlsx: no sheets")
	}
	sheet := wb.Sheets[0]

	var header []string
	rows := []FoodRow{}
	report := &ParseReport{}

	var idx map[string]int
	err = sheet.ForEachRow(func(row *xlsx.Row) error {
		rec := []string{}
		_ = row.ForEachCell(func(c *xlsx.Cell) error {
			rec = append(rec, c.String())
			return nil
		})
		if header == nil {
			header = rec
			m, herr := mapHeaders(header)
			if herr != nil {
				return herr
			}
			idx = m
			return nil
		}
		fr := rowFromRecord(rec, idx)
		if fr.NameIT == "" {
			report.RowsFailed++
			return nil
		}
		rows = append(rows, fr)
		report.RowsParsed++
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return rows, report, nil
}

func pick(rec []string, idx map[string]int, field string) string {
	i, ok := idx[field]
	if !ok || i >= len(rec) {
		return ""
	}
	return strings.TrimSpace(rec[i])
}

func rowFromRecord(rec []string, idx map[string]int) FoodRow {
	fr := FoodRow{
		ExternalID:  pick(rec, idx, "external_id"),
		NameIT:      pick(rec, idx, "name_it"),
		Category:    pick(rec, idx, "category"),
		EdiblePct:   parseFloatOr(pick(rec, idx, "edible_pct"), 100.0),
		Kcal:        parseFloatPtr(pick(rec, idx, "kcal")),
		ProteinG:    parseFloatPtr(pick(rec, idx, "protein_g")),
		FatG:        parseFloatPtr(pick(rec, idx, "fat_g")),
		CarbG:       parseFloatPtr(pick(rec, idx, "carb_g")),
		SugarG:      parseFloatPtr(pick(rec, idx, "sugar_g")),
		FiberG:      parseFloatPtr(pick(rec, idx, "fiber_g")),
		WaterG:      parseFloatPtr(pick(rec, idx, "water_g")),
		AshG:        parseFloatPtr(pick(rec, idx, "ash_g")),
		AlcoholG:    parseFloatPtr(pick(rec, idx, "alcohol_g")),
		SaltG:       parseFloatPtr(pick(rec, idx, "salt_g")),
		FetchedAt:   time.Now().UTC(),
	}
	if fr.ExternalID == "" {
		fr.ExternalID = "bda-" + NormalizeName(fr.NameIT)
	}
	raw, _ := json.Marshal(fr)
	fr.RawJSON = raw
	sum := sha256.Sum256(raw)
	fr.ContentHash = hex.EncodeToString(sum[:])
	return fr
}
