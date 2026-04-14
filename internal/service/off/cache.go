// Package off implements the Open Food Facts proxy: cache-through lookup,
// rate-limited HTTP client, and a normalizer that maps OFF JSON onto Pesa's
// Food DTO. OFF data stays out of the foods table (ODbL isolation).
package off

import (
	"context"
	"database/sql"
	"time"
)

// TTLs per cache row status.
const (
	TTLHit   = 7 * 24 * time.Hour
	TTLMiss  = 24 * time.Hour
	TTLError = 1 * time.Hour
)

// Row-level statuses stored in off_cache.status.
const (
	StatusHit   = "hit"
	StatusMiss  = "miss"
	StatusError = "error"
)

// Repo wraps off_cache with TTL-aware Get/Put semantics.
type Repo struct {
	DB  *sql.DB
	Now func() time.Time // injectable for tests
}

// NewRepo builds a Repo bound to the given sql.DB.
func NewRepo(db *sql.DB) *Repo {
	return &Repo{DB: db, Now: time.Now}
}

// GetResult is the tuple returned by Get.
type GetResult struct {
	Payload   []byte
	Status    string
	FetchedAt time.Time
	Stale     bool
	Found     bool
}

// Get reads a cache row for ean and classifies it by TTL.
// Rows past their status-specific TTL are still returned but with Stale=true.
func (r *Repo) Get(ctx context.Context, ean string) (GetResult, error) {
	var (
		payload   sql.NullString
		status    string
		fetchedAt time.Time
	)
	err := r.DB.QueryRowContext(ctx,
		`SELECT payload_json, status, fetched_at FROM off_cache WHERE ean = $1`, ean,
	).Scan(&payload, &status, &fetchedAt)
	if err == sql.ErrNoRows {
		return GetResult{}, nil
	}
	if err != nil {
		return GetResult{}, err
	}
	ttl := ttlForStatus(status)
	stale := r.Now().Sub(fetchedAt) > ttl
	out := GetResult{
		Status:    status,
		FetchedAt: fetchedAt,
		Stale:     stale,
		Found:     true,
	}
	if payload.Valid {
		out.Payload = []byte(payload.String)
	}
	return out, nil
}

// Put upserts a cache row with status, payload and optional error message.
func (r *Repo) Put(ctx context.Context, ean, status string, payload []byte, errMsg string) error {
	var payloadArg interface{}
	if payload != nil {
		payloadArg = string(payload)
	} else {
		payloadArg = nil
	}
	var errArg interface{}
	if errMsg != "" {
		if len(errMsg) > 255 {
			errMsg = errMsg[:255]
		}
		errArg = errMsg
	} else {
		errArg = nil
	}
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO off_cache (ean, payload_json, status, fetched_at, error_msg)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (ean) DO UPDATE SET
		  payload_json = EXCLUDED.payload_json,
		  status       = EXCLUDED.status,
		  fetched_at   = EXCLUDED.fetched_at,
		  error_msg    = EXCLUDED.error_msg`,
		ean, payloadArg, status, r.Now().UTC(), errArg,
	)
	return err
}

func ttlForStatus(status string) time.Duration {
	switch status {
	case StatusHit:
		return TTLHit
	case StatusMiss:
		return TTLMiss
	case StatusError:
		return TTLError
	default:
		return TTLError
	}
}
