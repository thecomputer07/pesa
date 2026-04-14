package erasure

import (
	"strings"
	"testing"
	"time"
)

func TestHashSaltedStable(t *testing.T) {
	a := HashSalted("u@example.it", []byte("salt"))
	b := HashSalted("u@example.it", []byte("salt"))
	if a != b {
		t.Fatalf("unstable hash: %s vs %s", a, b)
	}
	if len(a) != 64 {
		t.Fatalf("want 64-char hex, got %d", len(a))
	}
	c := HashSalted("u@example.it", []byte("other"))
	if a == c {
		t.Fatalf("hash should differ on different salt")
	}
}

func TestTicketRefFormat(t *testing.T) {
	ref := generateTicketRef(time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC))
	if !strings.HasPrefix(ref, "SS-20260413-") {
		t.Fatalf("ticket ref malformed: %s", ref)
	}
	if len(ref) != len("SS-20260413-")+6 {
		t.Fatalf("unexpected ticket_ref length: %s", ref)
	}
}

func TestAppendCSVIdempotent(t *testing.T) {
	if appendCSV("", "b1") != "b1" {
		t.Fatal("empty start")
	}
	if appendCSV("b1", "b2") != "b1,b2" {
		t.Fatal("append new")
	}
	if appendCSV("b1,b2", "b1") != "b1,b2" {
		t.Fatal("idempotent on dup")
	}
}

func TestFormatPseudonym(t *testing.T) {
	full := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	if got := FormatPseudonym(full); got != "0123456789abcdef" {
		t.Fatalf("bad prefix: %s", got)
	}
}

func TestDeleteStatementQuoted(t *testing.T) {
	s := DeleteStatement("abc")
	if !strings.Contains(s, `"abc"`) {
		t.Fatalf("want quoted pseudonym in stmt: %s", s)
	}
}
