package export

import (
	"archive/zip"
	"bytes"
	"testing"
	"time"
)

// Worker logic is exercised end-to-end against a live DB in integration tests
// (not available in this package-level unit sweep). Here we assert static
// invariants: default TTL matches PRIV-01, HashSecret is stable, and bundle
// construction produces a valid zip from empty inputs.
func TestDefaultTTLSevenDays(t *testing.T) {
	want := 7 * 24 * time.Hour
	if DefaultTTL != want {
		t.Fatalf("TTL drift: want %v got %v", want, DefaultTTL)
	}
}

func TestHashSecretStable(t *testing.T) {
	a := HashSecret([]byte("k"))
	b := HashSecret([]byte("k"))
	if a != b {
		t.Fatalf("HashSecret non-deterministic")
	}
}

func TestEmptyBundleValidZip(t *testing.T) {
	parts, err := PartsFromTables([]byte(`{}`), nil, ReadmeMarkdown())
	if err != nil {
		t.Fatal(err)
	}
	data, err := BundleToBytes(parts)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := zip.NewReader(bytes.NewReader(data), int64(len(data))); err != nil {
		t.Fatalf("zip unreadable: %v", err)
	}
}
