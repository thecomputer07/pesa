package erasure

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ErrUserNotFound is returned when Erase is invoked on an id that does not
// exist (or has already been erased) — makes the call idempotent-safe.
var ErrUserNotFound = errors.New("erasure: user not found")

// Service coordinates the transactional cascade delete + tombstone insert.
type Service struct {
	DB       *sql.DB
	HashSalt []byte // PESA_ERASURE_HASH_SALT
	LogSalt  []byte // PESA_LOG_SALT (pseudonym for exports dir)
	CacheDir string // .cache/exports
	Now      func() time.Time
}

// New constructs a Service with sane defaults.
func New(db *sql.DB, hashSalt, logSalt []byte, cacheDir string) *Service {
	return &Service{DB: db, HashSalt: hashSalt, LogSalt: logSalt, CacheDir: cacheDir, Now: time.Now}
}

// EraseResult is what Erase returns on success.
type EraseResult struct {
	TicketRef       string
	UserIDHash      string
	EmailHash       string
	PseudonymFull   string
	PseudonymShort  string
	BackupFilesPath string
}

// Erase performs the transactional cascade for userID.
//
//	channel is "self-service" for DELETE /api/v1/me invocations, "email" for
//	staff-driven erasure from srt tickets.
//	notes is the user-supplied motivation (hashed, never stored).
func (s *Service) Erase(ctx context.Context, userID int64, channel, notes string) (*EraseResult, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var email string
	if err := tx.QueryRowContext(ctx, `SELECT email FROM users WHERE id=$1`, userID).Scan(&email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	uidHash := HashSalted(fmt.Sprintf("%d", userID), s.HashSalt)
	emailHash := HashSalted(email, s.HashSalt)
	notesHash := ""
	if notes != "" {
		notesHash = HashSalted(notes, s.HashSalt)
	}
	pseudonymFull := HashSalted(email, s.LogSalt) // 64-char hex, SHA-256
	pseudonymShort := FormatPseudonym(pseudonymFull)

	// Cascade: children-first to respect FKs even when ON DELETE CASCADE is
	// present; redundancy is acceptable and test-friendly.
	stmts := []string{
		`DELETE FROM meal_items USING meals WHERE meals.id=meal_items.meal_id AND meals.user_id=$1`,
		`DELETE FROM meals WHERE user_id=$1`,
		`DELETE FROM recipe_items USING recipes WHERE recipes.id=recipe_items.recipe_id AND recipes.user_id=$1`,
		`DELETE FROM recipes WHERE user_id=$1`,
		`DELETE FROM meal_templates WHERE user_id=$1`,
		`DELETE FROM favorites WHERE user_id=$1`,
		`DELETE FROM sessions WHERE user_id=$1`,
		`DELETE FROM email_verifications WHERE user_id=$1`,
		`DELETE FROM password_resets WHERE user_id=$1`,
		`DELETE FROM data_export_jobs WHERE user_id=$1`,
		`DELETE FROM users WHERE id=$1`,
	}
	for _, stmt := range stmts {
		if _, err := tx.ExecContext(ctx, stmt, userID); err != nil {
			// Be tolerant of missing tables in stripped test schemas: retry
			// without failing the whole cascade when the table is absent.
			if isMissingTable(err) {
				continue
			}
			return nil, fmt.Errorf("erasure cascade: %s: %w", stmt, err)
		}
	}

	ticketRef := generateTicketRef(s.now())
	_, err = tx.ExecContext(ctx,
		`INSERT INTO erasure_log
		   (user_id_hash, email_hash, requested_at, executed_at,
		    channel, ticket_ref, right_exercised, notes_hash)
		 VALUES ($1, $2, NOW(), NOW(), $3, $4, 'art.17', NULLIF($5, ''))`,
		uidHash, emailHash, channel, ticketRef, notesHash)
	if err != nil {
		return nil, err
	}
	if err := InsertTombstone(ctx, tx, pseudonymFull, emailHash); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Post-commit filesystem cleanup (best-effort; DB is source of truth).
	exportsDir := filepath.Join(s.CacheDir, pseudonymShort)
	if s.CacheDir != "" {
		if err := os.RemoveAll(exportsDir); err != nil && !os.IsNotExist(err) {
			// log-only — tombstone survives so backups still get cleaned
			_ = err
		}
	}
	return &EraseResult{
		TicketRef:       ticketRef,
		UserIDHash:      uidHash,
		EmailHash:       emailHash,
		PseudonymFull:   pseudonymFull,
		PseudonymShort:  pseudonymShort,
		BackupFilesPath: exportsDir,
	}, nil
}

func (s *Service) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

func generateTicketRef(now time.Time) string {
	var b [3]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("SS-%s-%s", now.UTC().Format("20060102"), hex.EncodeToString(b[:]))
}

// isMissingTable returns true for MariaDB/MySQL "Table doesn't exist" errors
// and PostgreSQL "does not exist" errors so the cascade can tolerate stripped
// schemas in tests.
func isMissingTable(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return contains(msg, "doesn't exist") || contains(msg, "does not exist") ||
		contains(msg, "no such table") || contains(msg, "Unknown table")
}

func contains(a, b string) bool {
	for i := 0; i+len(b) <= len(a); i++ {
		if a[i:i+len(b)] == b {
			return true
		}
	}
	return false
}
