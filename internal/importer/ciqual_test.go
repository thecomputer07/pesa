package importer

import "testing"

func TestReadCiqualXML_Sushi(t *testing.T) {
	names := map[string]string{"25456": "Sushi o maki ai prodotti del mare"}
	rows, err := ReadCiqualXML("testdata/ciqual", names)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	// Pastis (1000) is in the fixture but not translated: skipped.
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(rows))
	}
	r := rows[0]
	if r.NameIT != "Sushi o maki ai prodotti del mare" || r.Category != "Piatti composti" {
		t.Errorf("name/category = %q / %q", r.NameIT, r.Category)
	}
	if r.CarbG == nil || *r.CarbG != 26.9 {
		t.Errorf("CarbG = %v, want 26.9", r.CarbG)
	}
	if r.Kcal == nil || *r.Kcal != 171 {
		t.Errorf("Kcal = %v, want 171 (kcal, not kJ)", r.Kcal)
	}
	if r.SaltG == nil || *r.SaltG != 0.55 {
		t.Errorf("SaltG = %v", r.SaltG)
	}
	if r.ContentHash == "" || len(r.RawJSON) == 0 {
		t.Error("missing content hash / raw json")
	}
	if res := SanityCheck(r); !res.OK {
		t.Errorf("sanity: %v", res.Reasons)
	}
}

func TestReadCiqualXML_UnknownCode(t *testing.T) {
	_, err := ReadCiqualXML("testdata/ciqual", map[string]string{"99999": "Inesistente"})
	if err == nil {
		t.Fatal("want error for a translated code missing from the release")
	}
}

func TestParseCiqualValue(t *testing.T) {
	cases := map[string]*float64{"12,5": ptr(12.5), "traces": ptr(0), "< 0,5": ptr(0), "-": nil, "": nil}
	for in, want := range cases {
		got := parseCiqualValue(in)
		if (got == nil) != (want == nil) || (got != nil && *got != *want) {
			t.Errorf("parseCiqualValue(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestCiqualNamesIT_Embedded(t *testing.T) {
	names, err := ParseCiqualNamesIT(ciqualNamesIT)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) < 800 {
		t.Errorf("embedded translations = %d, want the full subset (~859)", len(names))
	}
}

func ptr(v float64) *float64 { return &v }
