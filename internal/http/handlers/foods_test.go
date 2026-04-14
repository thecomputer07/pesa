package handlers

import (
	"testing"

	"github.com/thecomputer07/pesa/internal/repo"
)

// Representative subset (5 cases) of BuildBooleanQuery behaviour.
// TODO(phase-2-verifier): add HTTP-level integration tests once the test
// harness supports ephemeral MariaDB (see plan-05 §Task 5).

func TestBuildBooleanQuery_PastaDiSemola(t *testing.T) {
	q, first := repo.BuildBooleanQuery("pasta di semola")
	if q != "+pasta +semola*" {
		t.Errorf("got %q, want %q", q, "+pasta +semola*")
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
	q, _ := repo.BuildBooleanQuery("caffè")
	if q != "+caffè*" {
		t.Errorf("accent handling wrong: got %q", q)
	}
}

func TestBuildBooleanQuery_PrefixOnLast(t *testing.T) {
	q, _ := repo.BuildBooleanQuery("parmigiano reggiano")
	if q != "+parmigiano +reggiano*" {
		t.Errorf("got %q", q)
	}
}
