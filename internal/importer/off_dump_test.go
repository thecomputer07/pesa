package importer

import (
	"os"
	"testing"
)

func TestReadOFFDump_TopItalianWithCarbs(t *testing.T) {
	f, err := os.Open("testdata/off-dump.tsv")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	got, matched, err := ReadOFFDump(f, "Italy", 2)
	if err != nil {
		t.Fatal(err)
	}
	// France-only and no-carbs rows are excluded; 3 Italian rows with carbs.
	if matched != 3 {
		t.Errorf("matched = %d, want 3", matched)
	}
	if len(got) != 2 || got[0].Code != "3017620425035" || got[1].Code != "5449000000996" {
		t.Fatalf("top 2 by scans = %+v", got)
	}
	if b := got[1].firstBrand(); b != "Coca-Cola" {
		t.Errorf("brand = %q, want Coca-Cola", b)
	}
	if c := got[1].Nutriments.CarbG; c == nil || *c != 10.6 {
		t.Errorf("carbs = %v", c)
	}
}
