package srt

import (
	"testing"
	"time"
)

func TestKindValidation(t *testing.T) {
	for k := range Kinds {
		if !Kinds[k] {
			t.Fatalf("kind %s should be accepted", k)
		}
	}
	if Kinds["bogus"] {
		t.Fatal("bogus kind accepted")
	}
}

func TestDaysRemainingSLA(t *testing.T) {
	now := time.Date(2026, 4, 13, 12, 0, 0, 0, time.UTC)
	deadline := now.Add(30 * 24 * time.Hour)
	if got := daysRemaining(deadline, now); got != 30 {
		t.Fatalf("want 30 days, got %d", got)
	}
	overdue := now.Add(-5 * 24 * time.Hour)
	if got := daysRemaining(overdue, now); got >= 0 {
		t.Fatalf("want negative days_remaining, got %d", got)
	}
}

func TestRenderSubstitutes(t *testing.T) {
	got := Render("ref={{TICKET_REF}} due={{DEADLINE}} untouched={{X}}",
		map[string]string{"TICKET_REF": "SRT-2026-001", "DEADLINE": "2026-05-13T00:00:00Z"})
	want := "ref=SRT-2026-001 due=2026-05-13T00:00:00Z untouched={{X}}"
	if got != want {
		t.Fatalf("render mismatch:\n got: %q\nwant: %q", got, want)
	}
}

func TestHashSaltedStable(t *testing.T) {
	a := HashSalted("x@y.it", []byte("s"))
	b := HashSalted("x@y.it", []byte("s"))
	if a != b || len(a) != 64 {
		t.Fatalf("unstable or wrong-length hash: %s", a)
	}
}
