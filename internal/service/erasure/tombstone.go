// Package erasure implements Art. 17 cascade deletion plus tombstone
// propagation to backups (PRIV-02/03/04).
package erasure

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// Tombstone is a row in the tombstones table.
type Tombstone struct {
	ID               int64
	UserIDPseudonym  string
	EmailHash        string
	ErasedAt         time.Time
	AppliedToBackups string
	LastAttemptedAt  sql.NullTime
}

// HashSalted returns hex(sha256(input || salt)) — matches Phase 1 logger
// pseudonym convention and erasure_log columns.
func HashSalted(input string, salt []byte) string {
	h := sha256.New()
	h.Write([]byte(input))
	h.Write(salt)
	return hex.EncodeToString(h.Sum(nil))
}

// InsertTombstone records a pending strip-from-backup entry.
func InsertTombstone(ctx context.Context, tx *sql.Tx, pseudonym, emailHash string) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO tombstones (user_id_pseudonym, email_hash, erased_at)
		 VALUES ($1, $2, NOW())`,
		pseudonym, emailHash)
	return err
}

// ListPending returns tombstones whose applied_to_backups does not yet contain
// backupID. ORDER BY erased_at ASC so older entries strip first.
func ListPending(ctx context.Context, db *sql.DB, backupID string) ([]Tombstone, error) {
	rs, err := db.QueryContext(ctx,
		`SELECT id, user_id_pseudonym, email_hash, erased_at,
		        COALESCE(applied_to_backups,''), last_attempted_at
		 FROM tombstones
		 WHERE applied_to_backups IS NULL OR applied_to_backups NOT LIKE $1
		 ORDER BY erased_at ASC`,
		"%"+backupID+"%")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rs.Close() }()
	out := []Tombstone{}
	for rs.Next() {
		var t Tombstone
		if err := rs.Scan(&t.ID, &t.UserIDPseudonym, &t.EmailHash, &t.ErasedAt,
			&t.AppliedToBackups, &t.LastAttemptedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rs.Err()
}

// MarkApplied appends backupID to applied_to_backups for the given tombstone
// (idempotent).
func MarkApplied(ctx context.Context, db *sql.DB, id int64, backupID string) error {
	var existing sql.NullString
	if err := db.QueryRowContext(ctx,
		`SELECT applied_to_backups FROM tombstones WHERE id=$1`, id).Scan(&existing); err != nil {
		return err
	}
	merged := appendCSV(existing.String, backupID)
	_, err := db.ExecContext(ctx,
		`UPDATE tombstones SET applied_to_backups=$1, last_attempted_at=NOW() WHERE id=$2`,
		merged, id)
	if err != nil {
		return err
	}
	// Mirror into erasure_log.backup_tombstone_at for audit.
	_, _ = db.ExecContext(ctx,
		`UPDATE erasure_log SET backup_tombstone_at=NOW()
		 WHERE user_id_hash=$1 AND backup_tombstone_at IS NULL`,
		existing.String) // best-effort; legacy rows may be left as-is
	return nil
}

func appendCSV(current, id string) string {
	if current == "" {
		return id
	}
	for _, p := range strings.Split(current, ",") {
		if strings.TrimSpace(p) == id {
			return current
		}
	}
	return current + "," + id
}

// FormatPseudonym returns the 16-char prefix used for filesystem directories
// (.cache/exports/<prefix>) given the full pseudonym hex.
func FormatPseudonym(fullHex string) string {
	if len(fullHex) < 16 {
		return fullHex
	}
	return fullHex[:16]
}

// DeleteStatement returns the staff-CLI strip SQL as a human-readable string.
// Exposed for logging and dry-run mode.
func DeleteStatement(pseudonym string) string {
	return fmt.Sprintf("DELETE FROM users WHERE encode(digest(email || @salt, 'sha256'), 'hex') = %q", pseudonym)
}
