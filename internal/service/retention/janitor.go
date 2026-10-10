// Package retention implements PRIV-06/PRIV-03: periodic cleanup of expired
// auth tokens, completed export jobs, and stale erasure_log rows.
package retention

import (
	"context"
	"database/sql"
	"log/slog"
	"time"
)

const (
	// ErasureLogRetention is the upper bound for anonymized erasure records
	// per RoPA Activity B (target 12 months).
	ErasureLogRetention = 365 * 24 * time.Hour

	// ExpiredTokenGrace delays deletion of expired auth tokens to aid debugging
	// of race conditions between client retries and server cleanup.
	ExpiredTokenGrace = 24 * time.Hour
)

// Janitor runs periodic DELETEs to honor retention commitments.
type Janitor struct {
	DB       *sql.DB
	Interval time.Duration
	Now      func() time.Time
}

func New(db *sql.DB) *Janitor {
	return &Janitor{DB: db, Interval: 1 * time.Hour, Now: time.Now}
}

func (j *Janitor) Run(ctx context.Context) {
	j.Tick(ctx)
	t := time.NewTicker(j.Interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			j.Tick(ctx)
		}
	}
}

// Tick runs one cleanup pass. Errors are logged and swallowed — the next tick
// will retry. Each DELETE is independent so one failing statement does not
// block the others.
func (j *Janitor) Tick(ctx context.Context) {
	now := j.Now().UTC()
	tokenCutoff := now.Add(-ExpiredTokenGrace)
	erasureCutoff := now.Add(-ErasureLogRetention)

	exec := func(label, query string, args ...interface{}) {
		res, err := j.DB.ExecContext(ctx, query, args...)
		if err != nil {
			slog.Error("retention janitor", "op", label, "err", err)
			return
		}
		n, _ := res.RowsAffected()
		if n > 0 {
			slog.Info("retention janitor", "op", label, "rows", n)
		}
	}

	exec("email_verifications",
		`DELETE FROM email_verifications WHERE expires_at < $1`, tokenCutoff)
	exec("password_resets",
		`DELETE FROM password_resets WHERE expires_at < $1`, tokenCutoff)
	exec("data_export_jobs",
		`DELETE FROM data_export_jobs
		 WHERE status IN ('expired','failed')
		   AND COALESCE(expires_at, requested_at) < $1`, tokenCutoff)
	exec("erasure_log",
		`DELETE FROM erasure_log WHERE executed_at < $1`, erasureCutoff)
}
