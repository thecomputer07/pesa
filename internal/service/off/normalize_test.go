package off

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalize_Golden(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "off_barancini.json"))
	if err != nil {
		t.Fatal(err)
	}
	dto, err := FromOFF(data, "8001505005707")
	if err != nil {
		t.Fatalf("FromOFF: %v", err)
	}
	if dto.NameIT == "" {
		t.Errorf("NameIT empty")
	}
	if dto.NameEN == nil || *dto.NameEN == "" {
		t.Errorf("NameEN empty")
	}
	if dto.License != "ODbL-1.0" {
		t.Errorf("License: %q", dto.License)
	}
	if dto.SourceURL == nil || *dto.SourceURL == "" {
		t.Errorf("SourceURL missing")
	}
	if len(dto.Images) != 1 {
		t.Errorf("expected 1 image")
	}
}
