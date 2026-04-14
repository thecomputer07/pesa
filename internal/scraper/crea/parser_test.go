package crea

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestParseDetailHTML_Pasta — representative subset (plan notes: 3-5 cases).
// TODO(phase-2-verifier): expand to full CREA schema coverage once real fixtures
// land after CREA outreach completes.
func TestParseDetailHTML_Pasta(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "sample.html"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	rec, err := ParseDetailHTML(b, "https://www.alimentinutrizione.it/valori-nutrizionali/1234")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if rec.NameIT != "Pasta di semola" {
		t.Errorf("NameIT = %q, want %q", rec.NameIT, "Pasta di semola")
	}
	if rec.CarbG == nil || *rec.CarbG != 71.2 {
		t.Errorf("CarbG = %v, want 71.2", rec.CarbG)
	}
	if rec.EdiblePct == nil || *rec.EdiblePct != 100.0 {
		t.Errorf("EdiblePct = %v, want 100.0", rec.EdiblePct)
	}
	if rec.ExternalID == "" {
		t.Errorf("ExternalID empty")
	}
	if len(rec.ContentHash) != 64 {
		t.Errorf("ContentHash len = %d, want 64", len(rec.ContentHash))
	}
}

func TestParseDetailHTML_GoldenJSON(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "sample.html"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	rec, err := ParseDetailHTML(b, "https://www.alimentinutrizione.it/valori-nutrizionali/1234")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	golden, err := os.ReadFile(filepath.Join("testdata", "sample.golden.json"))
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	var g map[string]interface{}
	if err := json.Unmarshal(golden, &g); err != nil {
		t.Fatalf("golden unmarshal: %v", err)
	}
	if g["name_it"].(string) != rec.NameIT {
		t.Errorf("golden name mismatch")
	}
	if g["carb_g"].(float64) != *rec.CarbG {
		t.Errorf("golden carb_g mismatch")
	}
}

// TestParseIndexHTML — subset; skips non-matching hrefs.
func TestParseIndexHTML(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "index.html"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	urls, err := ParseIndexHTML(b, "https://www.alimentinutrizione.it/")
	if err != nil {
		t.Fatalf("parse index: %v", err)
	}
	if len(urls) != 5 {
		t.Fatalf("got %d urls, want 5", len(urls))
	}
	if urls[0] != "https://www.alimentinutrizione.it/valori-nutrizionali/1001" {
		t.Errorf("unexpected first url: %s", urls[0])
	}
}
