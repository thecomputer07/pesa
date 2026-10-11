package importer

import (
	"math"
	"testing"
)

func TestReadUSDASurveyCSV_SalmonRoll(t *testing.T) {
	names := map[string]usdaName{"2708963": {"Maki al salmone", "Sushi"}}
	rows, err := ReadUSDASurveyCSV("testdata/usda", names)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	// Milk (2705384) is in the fixture but not listed: skipped.
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(rows))
	}
	r := rows[0]
	if r.NameIT != "Maki al salmone" || r.Category != "Sushi" {
		t.Errorf("name/category = %q / %q", r.NameIT, r.Category)
	}
	// 15.3 g by difference minus 0.7 g fibre = 14.6 g available carbohydrate.
	if r.CarbG == nil || math.Abs(*r.CarbG-14.6) > 1e-9 {
		t.Errorf("CarbG = %v, want 14.6 (available, fibre excluded)", r.CarbG)
	}
	if r.Kcal == nil || *r.Kcal != 103 {
		t.Errorf("Kcal = %v, want 103", r.Kcal)
	}
	// 350 mg sodium x 2.5 = 0.875 g salt.
	if r.SaltG == nil || math.Abs(*r.SaltG-0.875) > 1e-9 {
		t.Errorf("SaltG = %v, want 0.875", r.SaltG)
	}
	if res := SanityCheck(r); !res.OK {
		t.Errorf("sanity: %v", res.Reasons)
	}
}

func TestReadUSDASurveyCSV_UnknownID(t *testing.T) {
	_, err := ReadUSDASurveyCSV("testdata/usda", map[string]usdaName{"1": {"X", "Y"}})
	if err == nil {
		t.Fatal("want error for an id missing from the release")
	}
}

func TestUSDANamesIT_Embedded(t *testing.T) {
	names, err := ParseUSDANamesIT(usdaNamesIT)
	if err != nil {
		t.Fatal(err)
	}
	sushi := 0
	for _, n := range names {
		if n.Category == "Sushi" {
			sushi++
		}
	}
	if sushi != 14 {
		t.Errorf("sushi entries = %d, want 14", sushi)
	}
}
