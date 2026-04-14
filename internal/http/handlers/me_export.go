package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/thecomputer07/pesa/internal/http/problem"
	mw "github.com/thecomputer07/pesa/internal/http/middleware"
	"github.com/thecomputer07/pesa/internal/service/export"
)

// MeExport bundles the PRIV-01 endpoints (POST, GET status, GET download).
type MeExport struct {
	DB     *sql.DB
	Secret []byte
}

// NewMeExport constructs the handler.
func NewMeExport(db *sql.DB, secret []byte) *MeExport {
	return &MeExport{DB: db, Secret: secret}
}

// Request handles POST /api/v1/me/export. Returns an existing pending/ready
// job if one is present, else inserts a new pending row.
func (m *MeExport) Request(w http.ResponseWriter, r *http.Request) {
	uid := mw.UserIDFrom(r.Context())
	if uid == 0 {
		problem.Write(w, r, http.StatusUnauthorized, "unauthorized", "login required")
		return
	}
	// Reuse an active job to enforce "one active per user" (T-06-03).
	var jobID int64
	err := m.DB.QueryRowContext(r.Context(),
		`SELECT id FROM data_export_jobs
		 WHERE user_id=$1 AND status IN ('pending','running','ready')
		 ORDER BY id DESC LIMIT 1`, uid).Scan(&jobID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		problem.Write(w, r, http.StatusInternalServerError, "internal", "db error")
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		err = m.DB.QueryRowContext(r.Context(),
			`INSERT INTO data_export_jobs (user_id, status, requested_at)
			 VALUES ($1, 'pending', NOW()) RETURNING id`, uid).Scan(&jobID)
		if err != nil {
			problem.Write(w, r, http.StatusInternalServerError, "internal", "insert failed")
			return
		}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]any{"job_id": jobID})
}

// Status handles GET /api/v1/me/export/{job_id}.
func (m *MeExport) Status(w http.ResponseWriter, r *http.Request) {
	uid := mw.UserIDFrom(r.Context())
	if uid == 0 {
		problem.Write(w, r, http.StatusUnauthorized, "unauthorized", "login required")
		return
	}
	jobID, err := strconv.ParseInt(chi.URLParam(r, "job_id"), 10, 64)
	if err != nil || jobID <= 0 {
		problem.Write(w, r, http.StatusNotFound, "not_found", "job not found")
		return
	}
	var (
		status    string
		bytesOut  sql.NullInt64
		tok       sql.NullString
		expiresAt sql.NullTime
	)
	err = m.DB.QueryRowContext(r.Context(),
		`SELECT status, bytes, download_token, expires_at
		 FROM data_export_jobs WHERE id=$1 AND user_id=$2`, jobID, uid).
		Scan(&status, &bytesOut, &tok, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		problem.Write(w, r, http.StatusNotFound, "not_found", "job not found")
		return
	}
	if err != nil {
		problem.Write(w, r, http.StatusInternalServerError, "internal", "db error")
		return
	}
	out := map[string]any{"status": status}
	if bytesOut.Valid {
		out["bytes"] = bytesOut.Int64
	}
	if status == "ready" && tok.Valid && expiresAt.Valid {
		out["download_url"] = "/api/v1/me/export/download?t=" + tok.String
		out["expires_at"] = expiresAt.Time.UTC().Format(time.RFC3339)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(out)
}

// Download handles GET /api/v1/me/export/download?t=<token>.
// Requires both a valid signed token AND authenticated session whose user_id
// matches the job owner (T-06-01, defense in depth).
func (m *MeExport) Download(w http.ResponseWriter, r *http.Request) {
	uid := mw.UserIDFrom(r.Context())
	if uid == 0 {
		problem.Write(w, r, http.StatusUnauthorized, "unauthorized", "login required")
		return
	}
	tok := r.URL.Query().Get("t")
	if !export.TokenIsWellFormed(tok) {
		problem.Write(w, r, http.StatusUnauthorized, "token_invalid", "bad token")
		return
	}
	var (
		jobID     int64
		ownerID   int64
		filePath  sql.NullString
		expiresAt sql.NullTime
		status    string
	)
	err := m.DB.QueryRowContext(r.Context(),
		`SELECT id, user_id, file_path, expires_at, status
		 FROM data_export_jobs WHERE download_token=$1`, tok).
		Scan(&jobID, &ownerID, &filePath, &expiresAt, &status)
	if errors.Is(err, sql.ErrNoRows) {
		problem.Write(w, r, http.StatusUnauthorized, "token_invalid", "bad token")
		return
	}
	if err != nil {
		problem.Write(w, r, http.StatusInternalServerError, "internal", "db error")
		return
	}
	if ownerID != uid {
		problem.Write(w, r, http.StatusUnauthorized, "token_invalid", "bad token")
		return
	}
	if status == "expired" || !expiresAt.Valid {
		problem.Write(w, r, http.StatusGone, "expired", "export expired")
		return
	}
	if err := export.Verify(tok, jobID, expiresAt.Time, m.Secret); err != nil {
		problem.Write(w, r, http.StatusGone, "expired", "export expired")
		return
	}
	if !filePath.Valid {
		problem.Write(w, r, http.StatusGone, "expired", "export expired")
		return
	}
	f, err := os.Open(filePath.String)
	if err != nil {
		problem.Write(w, r, http.StatusGone, "expired", "export expired")
		return
	}
	defer func() { _ = f.Close() }()
	st, _ := f.Stat()
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="pesa-export.zip"`)
	if st != nil {
		w.Header().Set("Content-Length", strconv.FormatInt(st.Size(), 10))
	}
	_, _ = io.Copy(w, f)
}
