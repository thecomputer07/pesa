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

// Fibre is 2 kcal/g: CREA "Carciofi, crudi" (33 kcal, 2.7 p, 0.2 f, 2.5 c,
// 5.5 fibre) must pass — it was quarantined when fibre was ignored.
func TestSanityCheck_FibreCountsTowardsKcal(t *testing.T) {
	f := func(v float64) *float64 { return &v }
	r := FoodRow{Kcal: f(33), ProteinG: f(2.7), FatG: f(0.2), CarbG: f(2.5), FiberG: f(5.5)}
	if res := SanityCheck(r); !res.OK {
		t.Fatalf("artichoke rejected: %v", res.Reasons)
	}
}

func TestPickBrand_SkipsCompanies(t *testing.T) {
	cases := map[string][]string{
		"Coca-Cola":     {"COCA-COLA SERVICES SA/NV", " Coca-Cola"},
		"Nutella":       {"Nutella", " FERRERO FRANCE COMMERCIALE"},
		"wasa":          {"wasa", " Barilla Sverige AB"},
		"Mulino Bianco": {"Barilla G. e R. Fratelli SpA", " Mulino Bianco", " Mulino Bianco - Barilla"},
		"Barilla":       {"Barilla", " Barilla G. e R. Fratelli SpA"},
		"Lindt":         {"Lindt"},
	}
	for want, in := range cases {
		if got := pickBrand(in); got != want {
			t.Errorf("pickBrand(%q) = %q, want %q", in, got, want)
		}
	}
}

// CREA "Fecola di patate": 1.4 p + 0 f + 90.7 c (as monosaccharides) +
// 16.1 water ~ 108 g. Legitimate, must pass.
func TestSanityCheck_StarchAsMonosaccharides(t *testing.T) {
	f := func(v float64) *float64 { return &v }
	r := FoodRow{Kcal: f(346), ProteinG: f(1.4), FatG: f(0), CarbG: f(90.7), WaterG: f(16.1)}
	if res := SanityCheck(r); !res.OK {
		t.Fatalf("potato starch rejected: %v", res.Reasons)
	}
}
