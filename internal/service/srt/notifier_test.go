package srt

import (
	"context"
	"strings"
	"testing"
	"time"
)

type capMailer struct {
	to, subj, body string
}

func (c *capMailer) Send(_ context.Context, to, subject, body string) error {
	c.to, c.subj, c.body = to, subject, body
	return nil
}

func TestNotifierUsesFallbackAndSubstitutes(t *testing.T) {
	m := &capMailer{}
	n := NewNotifier("/nonexistent/path", m)
	tk := &Ticket{
		TicketRef: "SRT-2026-001",
		Deadline:  time.Date(2026, 5, 13, 0, 0, 0, 0, time.UTC),
	}
	if err := n.SendAck(context.Background(), tk, "user@example.it"); err != nil {
		t.Fatal(err)
	}
	if m.to != "user@example.it" {
		t.Fatalf("recipient: %s", m.to)
	}
	if !strings.Contains(m.body, "SRT-2026-001") {
		t.Fatalf("body missing ticket ref: %s", m.body)
	}
	if !strings.Contains(m.body, "2026-05-13T00:00:00Z") {
		t.Fatalf("body missing deadline: %s", m.body)
	}
	if !strings.Contains(m.subj, "SRT-2026-001") {
		t.Fatalf("subject missing ref: %s", m.subj)
	}
}
