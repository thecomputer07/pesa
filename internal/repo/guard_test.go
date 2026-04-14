package repo

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// GUARD-03: schema must not contain any Art. 9 health data columns or tables.
// The scan is over SQL files in ../../migrations/ relative to this test.
func TestNoHealthDataColumns(t *testing.T) {
	forbidden := regexp.MustCompile(`(?i)(blood[_ ]?glucose|glicem|insulin|diagnosis|bolus|dosage|i_cho|\bfsi\b|\bcgm\b|insulin[_ ]?pump)`)
	root := filepath.Join("..", "..", "migrations")
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read migrations dir: %v", err)
	}
	found := false
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		found = true
		b, err := os.ReadFile(filepath.Join(root, e.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		if forbidden.MatchString(string(b)) {
			t.Errorf("GUARD-03 violation in %s", e.Name())
		}
	}
	if !found {
		t.Fatal("no migration files found — suite misconfigured")
	}
}
