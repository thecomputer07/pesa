package export

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// DefaultTTL is the signed-URL lifetime (PRIV-01 §decisions).
const DefaultTTL = 7 * 24 * time.Hour

// Worker processes pending data_export_jobs rows.
type Worker struct {
	DB       *sql.DB
	CacheDir string // .cache/exports by default
	Secret   []byte // PESA_EXPORT_TOKEN_SECRET
	LogSalt  []byte // PESA_LOG_SALT (for user pseudonym)
	Interval time.Duration
	TTL      time.Duration
	Now      func() time.Time
}

// NewWorker applies defaults.
func NewWorker(db *sql.DB, cacheDir string, secret, logSalt []byte) *Worker {
	return &Worker{
		DB: db, CacheDir: cacheDir, Secret: secret, LogSalt: logSalt,
		Interval: 5 * time.Second,
		TTL:      DefaultTTL,
		Now:      time.Now,
	}
}

// Run loops until ctx is cancelled. Each tick claims one pending job and/or
// performs a cleanup pass over expired rows.
func (w *Worker) Run(ctx context.Context) {
	t := time.NewTicker(w.Interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			w.Tick(ctx)
		}
	}
}

// Tick performs exactly one iteration (claim+build or cleanup). Returns true
// if a job was handled (tests drive it directly).
func (w *Worker) Tick(ctx context.Context) bool {
	_ = w.cleanupExpired(ctx)
	jobID, userID, ok, err := w.claimOne(ctx)
	if err != nil {
		slog.Error("export worker claim", "err", err)
		return false
	}
	if !ok {
		return false
	}
	if err := w.process(ctx, jobID, userID); err != nil {
		slog.Error("export worker process", "job_id", jobID, "err", err)
		_ = w.markFailed(ctx, jobID, err.Error())
	}
	return true
}

// claimOne atomically picks a pending job and flips it to running.
func (w *Worker) claimOne(ctx context.Context) (int64, int64, bool, error) {
	tx, err := w.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, 0, false, err
	}
	defer func() { _ = tx.Rollback() }()
	var jobID, userID int64
	err = tx.QueryRowContext(ctx,
		`SELECT id, user_id FROM data_export_jobs
		 WHERE status='pending' ORDER BY id LIMIT 1 FOR UPDATE`).
		Scan(&jobID, &userID)
	if err == sql.ErrNoRows {
		return 0, 0, false, nil
	}
	if err != nil {
		return 0, 0, false, err
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE data_export_jobs SET status='running' WHERE id=$1`, jobID); err != nil {
		return 0, 0, false, err
	}
	if err := tx.Commit(); err != nil {
		return 0, 0, false, err
	}
	return jobID, userID, true, nil
}

func (w *Worker) process(ctx context.Context, jobID, userID int64) error {
	// Lookup email for pseudonym.
	var email string
	if err := w.DB.QueryRowContext(ctx, `SELECT email FROM users WHERE id=$1`, userID).Scan(&email); err != nil {
		return fmt.Errorf("lookup user: %w", err)
	}
	pseudo := Pseudonym(email, w.LogSalt)
	dir := filepath.Join(w.CacheDir, pseudo)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	account, err := BuildAccountJSON(w.DB, userID)
	if err != nil {
		return err
	}
	tables, err := CollectTables(w.DB, userID)
	if err != nil {
		return err
	}
	parts, err := PartsFromTables(account, tables, ReadmeMarkdown())
	if err != nil {
		return err
	}
	fpath := filepath.Join(dir, fmt.Sprintf("%d.zip", jobID))
	f, err := os.OpenFile(fpath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	n, werr := WriteBundle(f, parts)
	if cerr := f.Close(); werr == nil {
		werr = cerr
	}
	if werr != nil {
		_ = os.Remove(fpath)
		return werr
	}
	expiresAt := w.Now().Add(w.TTL).UTC()
	token := Sign(jobID, expiresAt, w.Secret)
	_, err = w.DB.ExecContext(ctx,
		`UPDATE data_export_jobs
		 SET status='ready', file_path=$1, bytes=$2, download_token=$3, expires_at=$4, completed_at=NOW()
		 WHERE id=$5`,
		fpath, n, token, expiresAt, jobID)
	return err
}

func (w *Worker) markFailed(ctx context.Context, jobID int64, msg string) error {
	_, err := w.DB.ExecContext(ctx,
		`UPDATE data_export_jobs SET status='failed', error=$1, completed_at=NOW() WHERE id=$2`,
		msg, jobID)
	return err
}

func (w *Worker) cleanupExpired(ctx context.Context) error {
	rs, err := w.DB.QueryContext(ctx,
		`SELECT id, COALESCE(file_path,'') FROM data_export_jobs
		 WHERE expires_at IS NOT NULL AND expires_at < NOW()
		   AND status IN ('ready','failed')`)
	if err != nil {
		return err
	}
	type row struct {
		id   int64
		path string
	}
	var rows []row
	for rs.Next() {
		var r row
		if err := rs.Scan(&r.id, &r.path); err != nil {
			_ = rs.Close()
			return err
		}
		rows = append(rows, r)
	}
	_ = rs.Close()
	for _, r := range rows {
		if r.path != "" {
			_ = os.Remove(r.path)
		}
		if _, err := w.DB.ExecContext(ctx,
			`UPDATE data_export_jobs SET status='expired', download_token=NULL WHERE id=$1`, r.id); err != nil {
			return err
		}
	}
	return nil
}

// HashSecret returns a hex digest of the secret for non-reversible logging.
func HashSecret(secret []byte) string {
	h := sha256.Sum256(secret)
	return hex.EncodeToString(h[:8])
}
