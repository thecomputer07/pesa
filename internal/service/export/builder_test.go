package export

import (
	"archive/zip"
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestWriteCSV_Deterministic(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteCSV(&buf, []string{"id", "name"}, [][]string{
		{"1", "alpha"}, {"2", "beta,with,comma"},
	}); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	want := "id,name\r\n1,alpha\r\n2,\"beta,with,comma\"\r\n"
	if got != want {
		t.Fatalf("csv mismatch:\n got: %q\nwant: %q", got, want)
	}
}

func TestBundleRoundTrip(t *testing.T) {
	parts := []BundleFile{
		{Name: "account.json", Body: []byte(`{"email":"a@b.it"}`)},
		{Name: "meals.csv", Body: []byte("id\r\n1\r\n")},
		{Name: "README.md", Body: ReadmeMarkdown()},
	}
	data, err := BundleToBytes(parts)
	if err != nil {
		t.Fatal(err)
	}
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, f := range r.File {
		names[f.Name] = true
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatal(err)
		}
		if len(body) == 0 {
			t.Fatalf("empty file in zip: %s", f.Name)
		}
	}
	for _, want := range []string{"account.json", "meals.csv", "README.md"} {
		if !names[want] {
			t.Fatalf("zip missing entry %s", want)
		}
	}
}

func TestReadmeMentionsBundleContents(t *testing.T) {
	body := string(ReadmeMarkdown())
	for _, entry := range []string{"account.json", "meals.csv", "recipes.csv", "favorites.csv"} {
		if !strings.Contains(body, entry) {
			t.Fatalf("README missing reference to %s", entry)
		}
	}
}

func TestPseudonymStable(t *testing.T) {
	a := Pseudonym("user@example.it", []byte("salt"))
	b := Pseudonym("user@example.it", []byte("salt"))
	if a != b {
		t.Fatalf("pseudonym not stable: %s vs %s", a, b)
	}
	if len(a) != 16 {
		t.Fatalf("pseudonym length: want 16, got %d", len(a))
	}
	c := Pseudonym("other@example.it", []byte("salt"))
	if a == c {
		t.Fatalf("pseudonym collision across emails")
	}
}
