package importer

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ciqualNamesIT maps Ciqual alim_code -> Italian name. It also decides what
// gets imported: only translated codes, i.e. what CREA lacks (composite
// dishes, sweets, ice creams, sauces). Basic foods stay CREA-only.
//
//go:embed data/ciqual_it.tsv
var ciqualNamesIT []byte

// ciqualCategories gives an Italian category per Ciqual sub-group.
var ciqualCategories = map[string]string{
	"0101": "Insalate composte",
	"0102": "Zuppe",
	"0103": "Piatti composti",
	"0104": "Pizze, torte salate e crêpe",
	"0105": "Panini e sandwich",
	"0106": "Sfoglie e antipasti",
	"0702": "Cioccolato e prodotti al cioccolato",
	"0703": "Caramelle e confetteria",
	"0704": "Confetture e marmellate",
	"0705": "Cornetti e brioche",
	"0706": "Biscotti",
	"0707": "Cereali da colazione",
	"0708": "Barrette di cereali",
	"0709": "Torte e pasticcini",
	"0801": "Gelati",
	"0802": "Sorbetti",
	"0803": "Dessert gelati",
	"1001": "Salse",
}

// Ciqual const_code -> FoodRow field. Glucides (31000) are available
// carbohydrates, fibre excluded, like CREA.
var ciqualConstSetters = map[string]func(*FoodRow, float64){
	"328":   func(r *FoodRow, v float64) { r.Kcal = &v },
	"25000": func(r *FoodRow, v float64) { r.ProteinG = &v },
	"40000": func(r *FoodRow, v float64) { r.FatG = &v },
	"31000": func(r *FoodRow, v float64) { r.CarbG = &v },
	"32000": func(r *FoodRow, v float64) { r.SugarG = &v },
	"34100": func(r *FoodRow, v float64) { r.FiberG = &v },
	"400":   func(r *FoodRow, v float64) { r.WaterG = &v },
	"10000": func(r *FoodRow, v float64) { r.AshG = &v },
	"60000": func(r *FoodRow, v float64) { r.AlcoholG = &v },
	"10004": func(r *FoodRow, v float64) { r.SaltG = &v },
}

// ParseCiqualNamesIT reads the embedded "code<TAB>name" translation table.
func ParseCiqualNamesIT(data []byte) (map[string]string, error) {
	out := map[string]string{}
	sc := bufio.NewScanner(bytes.NewReader(data))
	line := 0
	for sc.Scan() {
		line++
		s := strings.TrimRight(sc.Text(), "\r")
		if strings.TrimSpace(s) == "" {
			continue
		}
		code, name, ok := strings.Cut(s, "\t")
		code, name = strings.TrimSpace(code), strings.TrimSpace(name)
		if !ok || code == "" || name == "" {
			return nil, fmt.Errorf("ciqual_it.tsv line %d: want code<TAB>name", line)
		}
		out[code] = name
	}
	return out, sc.Err()
}

// parseCiqualValue reads a Ciqual "teneur": "12,5" -> 12.5; "traces" and
// "< 0,5" -> 0 (below quantification); "-" or empty -> missing.
func parseCiqualValue(s string) *float64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "-" {
		return nil
	}
	if s == "traces" || strings.HasPrefix(s, "<") {
		z := 0.0
		return &z
	}
	v, err := strconv.ParseFloat(strings.ReplaceAll(s, ",", "."), 64)
	if err != nil {
		return nil
	}
	return &v
}

type ciqualAlim struct {
	Code     string `xml:"alim_code"`
	NameFR   string `xml:"alim_nom_fr"`
	NameEN   string `xml:"alim_nom_eng"`
	SubGroup string `xml:"alim_ssgrp_code"`
}

type ciqualCompo struct {
	AlimCode  string `xml:"alim_code"`
	ConstCode string `xml:"const_code"`
	Teneur    string `xml:"teneur"`
}

// eachXMLElement streams the <name> children of a Ciqual XML file.
func eachXMLElement(path, name string, fn func(d *xml.Decoder, se xml.StartElement) error) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	d := xml.NewDecoder(bufio.NewReader(f))
	for {
		tok, err := d.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("%s: %w", filepath.Base(path), err)
		}
		if se, ok := tok.(xml.StartElement); ok && se.Name.Local == name {
			if err := fn(d, se); err != nil {
				return err
			}
		}
	}
}

// findCiqualFile returns the single file in dir matching prefix_*.xml.
func findCiqualFile(dir, prefix string) (string, error) {
	m, err := filepath.Glob(filepath.Join(dir, prefix+"_*.xml"))
	if err != nil {
		return "", err
	}
	if len(m) != 1 {
		return "", fmt.Errorf("%s: want exactly one %s_*.xml, found %d", dir, prefix, len(m))
	}
	return m[0], nil
}

// ReadCiqualXML parses the Ciqual XML release in dir (alim_*.xml +
// compo_*.xml) into FoodRows for the codes present in names.
func ReadCiqualXML(dir string, names map[string]string) ([]FoodRow, error) {
	alimPath, err := findCiqualFile(dir, "alim")
	if err != nil {
		return nil, err
	}
	compoPath, err := findCiqualFile(dir, "compo")
	if err != nil {
		return nil, err
	}

	alims := map[string]ciqualAlim{}
	var order []string
	err = eachXMLElement(alimPath, "ALIM", func(d *xml.Decoder, se xml.StartElement) error {
		var a ciqualAlim
		if err := d.DecodeElement(&a, &se); err != nil {
			return err
		}
		a.Code = strings.TrimSpace(a.Code)
		if _, ok := names[a.Code]; !ok {
			return nil
		}
		a.NameFR, a.NameEN = strings.TrimSpace(a.NameFR), strings.TrimSpace(a.NameEN)
		a.SubGroup = strings.TrimSpace(a.SubGroup)
		alims[a.Code] = a
		order = append(order, a.Code)
		return nil
	})
	if err != nil {
		return nil, err
	}
	for code := range names {
		if _, ok := alims[code]; !ok {
			return nil, fmt.Errorf("ciqual_it.tsv: code %s not in %s", code, filepath.Base(alimPath))
		}
	}

	rows := map[string]*FoodRow{}
	raw := map[string]map[string]string{}
	for _, code := range order {
		a := alims[code]
		rows[code] = &FoodRow{
			ExternalID: code,
			URL:        "https://ciqual.anses.fr/#/aliments/" + code,
			NameIT:     names[code],
			Category:   ciqualCategories[a.SubGroup],
			EdiblePct:  100,
		}
		raw[code] = map[string]string{}
	}
	err = eachXMLElement(compoPath, "COMPO", func(d *xml.Decoder, se xml.StartElement) error {
		var c ciqualCompo
		if err := d.DecodeElement(&c, &se); err != nil {
			return err
		}
		code, cc := strings.TrimSpace(c.AlimCode), strings.TrimSpace(c.ConstCode)
		row, ok := rows[code]
		if !ok {
			return nil
		}
		set, ok := ciqualConstSetters[cc]
		if !ok {
			return nil
		}
		raw[code][cc] = strings.TrimSpace(c.Teneur)
		if v := parseCiqualValue(c.Teneur); v != nil {
			set(row, *v)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	out := make([]FoodRow, 0, len(order))
	for _, code := range order {
		a, row := alims[code], rows[code]
		b, _ := json.Marshal(map[string]interface{}{
			"code": code, "url": row.URL, "name_it": row.NameIT,
			"name_fr": a.NameFR, "name_en": a.NameEN, "ssgrp": a.SubGroup,
			"teneur": raw[code],
		})
		sum := sha256.Sum256(b)
		row.RawJSON = b
		row.ContentHash = hex.EncodeToString(sum[:])
		row.FetchedAt = now
		out = append(out, *row)
	}
	return out, nil
}

// ImportCiqual loads the Ciqual subset listed in data/ciqual_it.tsv from the
// XML release in dir. Idempotent like ImportCSV (content_hash short-circuit).
func ImportCiqual(ctx context.Context, db *sql.DB, dir string, opts ImportOpts) (*ImportReport, error) {
	if opts.SourceCode == "" {
		opts.SourceCode = "ciqual"
	}
	if opts.Basis == "" {
		opts.Basis = BasisEdible
	}
	names, err := ParseCiqualNamesIT(ciqualNamesIT)
	if err != nil {
		return nil, err
	}
	rows, err := ReadCiqualXML(dir, names)
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
