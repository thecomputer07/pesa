// Package srt implements the internal Subject Rights Tickets service
// (LEGAL-09/10): incoming privacy@ requests, 30-day SLA, staff-only.
package srt

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"time"
)

// Kinds enumerates the GDPR rights this ticketing system tracks.
var Kinds = map[string]bool{
	"access": true, "rectification": true, "erasure": true,
	"portability": true, "restriction": true, "objection": true,
	"complaint": true,
}

// ErrInvalidKind indicates an unknown ticket kind.
var ErrInvalidKind = errors.New("srt: invalid kind")

// ErrAlreadyClosed indicates a duplicate Close invocation.
var ErrAlreadyClosed = errors.New("srt: ticket already closed")

// ErrNotFound indicates the ticket id does not exist.
var ErrNotFound = errors.New("srt: ticket not found")

// Ticket mirrors a row of subject_rights_tickets with derived days_remaining.
type Ticket struct {
	ID            int64      `json:"id"`
	TicketRef     string     `json:"ticket_ref"`
	Kind          string     `json:"kind"`
	Status        string     `json:"status"`
	ReceivedAt    time.Time  `json:"received_at"`
	Deadline      time.Time  `json:"deadline"`
	ClosedAt      *time.Time `json:"closed_at,omitempty"`
	NotesMD       string     `json:"notes_md,omitempty"`
	EvidencePath  string     `json:"evidence_path,omitempty"`
	DaysRemaining int        `json:"days_remaining"`
}

// EraseInvoker is the subset of erasure.Service we need to perform an
// Art.17 close-with-erasure. Defined here to keep packages decoupled.
type EraseInvoker interface {
	Erase(ctx context.Context, userID int64, channel, notes string) (any, error)
}

// Service is the SRT business logic.
type Service struct {
	DB       *sql.DB
	HashSalt []byte
	Now      func() time.Time
}

// New constructs a Service with sane defaults.
func New(db *sql.DB, hashSalt []byte) *Service {
	return &Service{DB: db, HashSalt: hashSalt, Now: time.Now}
}

// CreateInput is the accepted payload of Service.Create.
type CreateInput struct {
	Kind           string
	RequesterEmail string
	NotesMD        string
}

// Create inserts a new ticket with deadline = received_at + 30d.
func (s *Service) Create(ctx context.Context, in CreateInput) (*Ticket, error) {
	if !Kinds[in.Kind] {
		return nil, ErrInvalidKind
	}
	now := s.now()
	deadline := now.Add(30 * 24 * time.Hour)
	ref, err := s.nextRef(ctx, now)
	if err != nil {
		return nil, err
	}
	emailHash := HashSalted(in.RequesterEmail, s.HashSalt)
	var id int64
	err = s.DB.QueryRowContext(ctx,
		`INSERT INTO subject_rights_tickets
		   (ticket_ref, requester_email_hash, kind, status, received_at, deadline, notes_md)
		 VALUES ($1, $2, $3, 'new', $4, $5, NULLIF($6, ''))
		 RETURNING id`,
		ref, emailHash, in.Kind, now, deadline, in.NotesMD).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &Ticket{
		ID: id, TicketRef: ref, Kind: in.Kind, Status: "new",
		ReceivedAt: now, Deadline: deadline, NotesMD: in.NotesMD,
		DaysRemaining: daysRemaining(deadline, now),
	}, nil
}

// ListFilter narrows Service.List.
type ListFilter struct {
	Status  string
	Overdue bool
}

// List returns matching tickets ordered by deadline ASC.
func (s *Service) List(ctx context.Context, f ListFilter) ([]Ticket, error) {
	q := `SELECT id, ticket_ref, kind, status, received_at, deadline, closed_at,
	             COALESCE(notes_md,''), COALESCE(evidence_path,'')
	      FROM subject_rights_tickets WHERE 1=1`
	args := []any{}
	n := 1
	if f.Status != "" {
		q += fmt.Sprintf(" AND status=$%d", n)
		args = append(args, f.Status)
		n++
	}
	if f.Overdue {
		q += " AND deadline < NOW() AND status NOT IN ('closed','rejected')"
	}
	q += " ORDER BY deadline ASC"
	rs, err := s.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rs.Close() }()
	now := s.now()
	out := []Ticket{}
	for rs.Next() {
		var t Ticket
		var closed sql.NullTime
		if err := rs.Scan(&t.ID, &t.TicketRef, &t.Kind, &t.Status,
			&t.ReceivedAt, &t.Deadline, &closed, &t.NotesMD, &t.EvidencePath); err != nil {
			return nil, err
		}
		if closed.Valid {
			ct := closed.Time
			t.ClosedAt = &ct
		}
		t.DaysRemaining = daysRemaining(t.Deadline, now)
		out = append(out, t)
	}
	return out, rs.Err()
}

// PatchInput optionally updates status, appends to notes, sets evidence.
type PatchInput struct {
	Status       *string
	AppendNotes  string
	EvidencePath *string
}

// Update applies PatchInput.
func (s *Service) Update(ctx context.Context, id int64, in PatchInput) (*Ticket, error) {
	if in.Status != nil {
		if _, err := s.DB.ExecContext(ctx,
			`UPDATE subject_rights_tickets SET status=$1 WHERE id=$2`, *in.Status, id); err != nil {
			return nil, err
		}
	}
	if in.EvidencePath != nil {
		if _, err := s.DB.ExecContext(ctx,
			`UPDATE subject_rights_tickets SET evidence_path=$1 WHERE id=$2`, *in.EvidencePath, id); err != nil {
			return nil, err
		}
	}
	if in.AppendNotes != "" {
		stamp := s.now().UTC().Format(time.RFC3339)
		line := fmt.Sprintf("\n\n[%s] %s", stamp, in.AppendNotes)
		if _, err := s.DB.ExecContext(ctx,
			`UPDATE subject_rights_tickets
			 SET notes_md = COALESCE(notes_md,'') || $1 WHERE id=$2`, line, id); err != nil {
			return nil, err
		}
	}
	return s.byID(ctx, id)
}

// Close sets the terminal status and closed_at; idempotent returns ErrAlreadyClosed.
func (s *Service) Close(ctx context.Context, id int64, terminal, finalNotes string) (*Ticket, error) {
	if terminal != "closed" && terminal != "rejected" {
		return nil, errors.New("srt: terminal must be closed|rejected")
	}
	t, err := s.byID(ctx, id)
	if err != nil {
		return nil, err
	}
	if t.Status == "closed" || t.Status == "rejected" {
		return nil, ErrAlreadyClosed
	}
	if finalNotes != "" {
		if _, err := s.Update(ctx, id, PatchInput{AppendNotes: finalNotes}); err != nil {
			return nil, err
		}
	}
	if _, err := s.DB.ExecContext(ctx,
		`UPDATE subject_rights_tickets SET status=$1, closed_at=NOW() WHERE id=$2`,
		terminal, id); err != nil {
		return nil, err
	}
	return s.byID(ctx, id)
}

// CloseErasure resolves the erasure_user_id via inv.Erase(...) and marks the
// ticket closed, appending the returned erasure ticket reference.
func (s *Service) CloseErasure(ctx context.Context, id int64, userID int64, inv EraseInvoker) (*Ticket, error) {
	t, err := s.byID(ctx, id)
	if err != nil {
		return nil, err
	}
	if t.Kind != "erasure" {
		return nil, errors.New("srt: ticket is not kind=erasure")
	}
	if t.Status == "closed" || t.Status == "rejected" {
		return nil, ErrAlreadyClosed
	}
	result, err := inv.Erase(ctx, userID, "email", "via SRT "+t.TicketRef)
	if err != nil {
		return nil, err
	}
	note := fmt.Sprintf("erasure applied: %v", result)
	return s.Close(ctx, id, "closed", note)
}

func (s *Service) byID(ctx context.Context, id int64) (*Ticket, error) {
	var t Ticket
	var closed sql.NullTime
	err := s.DB.QueryRowContext(ctx,
		`SELECT id, ticket_ref, kind, status, received_at, deadline, closed_at,
		        COALESCE(notes_md,''), COALESCE(evidence_path,'')
		 FROM subject_rights_tickets WHERE id=$1`, id).
		Scan(&t.ID, &t.TicketRef, &t.Kind, &t.Status,
			&t.ReceivedAt, &t.Deadline, &closed, &t.NotesMD, &t.EvidencePath)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if closed.Valid {
		ct := closed.Time
		t.ClosedAt = &ct
	}
	t.DaysRemaining = daysRemaining(t.Deadline, s.now())
	return &t, nil
}

func (s *Service) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

func (s *Service) nextRef(ctx context.Context, now time.Time) (string, error) {
	year := now.Year()
	prefix := fmt.Sprintf("SRT-%d-", year)
	var count int
	err := s.DB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM subject_rights_tickets WHERE ticket_ref LIKE $1`,
		prefix+"%").Scan(&count)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%03d", prefix, count+1), nil
}

func daysRemaining(deadline, now time.Time) int {
	return int(math.Floor(deadline.Sub(now).Hours() / 24))
}

// HashSalted mirrors the erasure package hasher so SRT rows can cross-reference
// erasure_log entries by requester email.
func HashSalted(input string, salt []byte) string {
	h := sha256.New()
	h.Write([]byte(input))
	h.Write(salt)
	return hex.EncodeToString(h.Sum(nil))
}
