package importer

import "testing"

// Representative subset (3-5) of parsing + matching tests.
// TODO(phase-2-verifier): add sqlmock-driven test of ImportBDA merge logic.

func TestParseBDA_CSV_Count(t *testing.T) {
	rows, report, err := ParseBDA("testdata/bda-fixture.csv", FormatAuto)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(rows) != 6 {
		t.Fatalf("rows = %d, want 6", len(rows))
	}
	if report.RowsFailed != 0 {
		t.Errorf("unexpected failures: %v", report.Errors)
	}
}

func TestParseBDA_HeaderMapping(t *testing.T) {
	rows, _, err := ParseBDA("testdata/bda-fixture.csv", FormatAuto)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if rows[0].NameIT != "pasta di semola" {
		t.Errorf("NameIT = %q", rows[0].NameIT)
	}
	if rows[0].CarbG == nil || *rows[0].CarbG != 71.0 {
		t.Errorf("CarbG = %v", rows[0].CarbG)
	}
}

func TestParseBDA_OverflowRowDetected(t *testing.T) {
	rows, _, _ := ParseBDA("testdata/bda-fixture.csv", FormatAuto)
	var overflow FoodRow
	for _, r := range rows {
		if r.ExternalID == "B006" {
			overflow = r
		}
	}
	res := SanityCheck(overflow)
	if res.OK {
		t.Errorf("B006 should fail sanity check")
	}
}

func TestNormalizeName_StopwordsAndAccents(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Pasta di semola", "pasta semola"},
		{"Parmigiano Reggiano", "parmigiano reggiano"},
		{"Caffè d'orzo", "caffe dorzo"},
		{"Olio di oliva", "olio oliva"},
	}
	for _, c := range cases {
		got := NormalizeName(c.in)
		if got != c.want {
			t.Errorf("NormalizeName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNormalizeName_Empty(t *testing.T) {
	if NormalizeName("  ") != "" {
		t.Errorf("empty input should produce empty output")
	}
}
