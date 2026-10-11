package handlers

import (
	"testing"

	"github.com/thecomputer07/pesa/internal/repo"
)

// BuildBooleanQuery (Postgres era): lowercase, stopwords dropped, tokens
// joined by spaces for ILIKE/pg_trgm; firstToken drives the prefix boost.

func TestBuildBooleanQuery_PastaDiSemola(t *testing.T) {
	q, first := repo.BuildBooleanQuery("pasta di semola")
	if q != "pasta semola" {
		t.Errorf("got %q, want %q", q, "pasta semola")
	}
	if first != "pasta" {
		t.Errorf("firstToken = %q", first)
	}
}

func TestBuildBooleanQuery_Empty(t *testing.T) {
	q, _ := repo.BuildBooleanQuery("")
	if q != "" {
		t.Errorf("empty input should produce empty q, got %q", q)
	}
}

func TestBuildBooleanQuery_StopwordOnly(t *testing.T) {
	q, _ := repo.BuildBooleanQuery("di la")
	if q != "" {
		t.Errorf("stopword-only input should produce empty q, got %q", q)
	}
}

func TestBuildBooleanQuery_Accents(t *testing.T) {
	// Accents are kept: catalog names are accented ("Caffè") and ILIKE
	// compares them literally.
	q, _ := repo.BuildBooleanQuery("Caffè")
	if q != "caffè" {
		t.Errorf("accent handling wrong: got %q", q)
	}
}

func TestBuildBooleanQuery_PunctuationAndCase(t *testing.T) {
	q, first := repo.BuildBooleanQuery("Parmigiano  Reggiano!")
	if q != "parmigiano reggiano" || first != "parmigiano" {
		t.Errorf("got %q / %q", q, first)
	}
}

func TestSearchTokenPatterns_AccentsAndTeaSpellings(t *testing.T) {
	q, _ := repo.BuildBooleanQuery("Thè alla pesca")
	got := repo.SearchTokenPatterns(q)
	want := []string{`\m(te|the|tea)`, `\m(pesca)`}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("patterns = %q, want %q", got, want)
	}
	q, _ = repo.BuildBooleanQuery("caffè")
	if got := repo.SearchTokenPatterns(q); len(got) != 1 || got[0] != `\m(caffe)` {
		t.Errorf("caffè -> %q", got)
	}
}
