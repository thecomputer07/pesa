package importer

import (
	"testing"
)

// TestReadCREACSV exercises the parser on the fixture.
// Representative subset (3 assertions) — full DB integration is covered by
// the live mysql verification in plan-03 §Verification.
// TODO(phase-2-verifier): add sqlmock-based round-trip test of ImportCSV.
func TestReadCREACSV_Count(t *testing.T) {
	rows, err := ReadCREACSV("testdata/crea-fixture.csv")
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(rows) != 10 {
		t.Fatalf("rows = %d, want 10", len(rows))
	}
}

func TestReadCREACSV_Pasta(t *testing.T) {
	rows, err := ReadCREACSV("testdata/crea-fixture.csv")
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	r := rows[0]
	if r.NameIT != "Pasta di semola" {
		t.Errorf("NameIT = %q", r.NameIT)
	}
	if r.CarbG == nil || *r.CarbG != 71.2 {
		t.Errorf("CarbG = %v", r.CarbG)
	}
	if r.Kcal == nil || *r.Kcal != 353 {
		t.Errorf("Kcal = %v", r.Kcal)
	}
}

func TestSanityCheck_MacroOverflow(t *testing.T) {
	rows, err := ReadCREACSV("testdata/crea-fixture.csv")
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	// row 9001 = macro overflow (50+40+40+20=150 > 105)
	var overflow FoodRow
	for _, r := range rows {
		if r.ExternalID == "9001" {
			overflow = r
		}
	}
	res := SanityCheck(overflow)
	if res.OK {
		t.Errorf("expected macro overflow to fail sanity, got OK")
	}
}

func TestSanityCheck_KcalMismatch(t *testing.T) {
	rows, err := ReadCREACSV("testdata/crea-fixture.csv")
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	// row 9002: declared=100, computed = 4*10+9*5+4*10 = 125, drift 25% > 20%
	var m FoodRow
	for _, r := range rows {
		if r.ExternalID == "9002" {
			m = r
		}
	}
	res := SanityCheck(m)
	if res.OK {
		t.Errorf("expected kcal mismatch to fail sanity, got OK")
	}
}

func TestSanityCheck_ValidPasta(t *testing.T) {
	rows, _ := ReadCREACSV("testdata/crea-fixture.csv")
	res := SanityCheck(rows[0])
	if !res.OK {
		t.Errorf("pasta row should pass sanity: %v", res.Reasons)
	}
}

func TestNormalize_EdiblePassthrough(t *testing.T) {
	kcal := 353.0
	r := FoodRow{EdiblePct: 100, Kcal: &kcal}
	out := Normalize(r, BasisEdible)
	if *out.Kcal != 353.0 {
		t.Errorf("passthrough broken: %v", *out.Kcal)
	}
}
