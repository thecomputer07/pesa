package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	mw "github.com/thecomputer07/pesa/internal/http/middleware"
	"github.com/thecomputer07/pesa/internal/http/problem"
	"github.com/thecomputer07/pesa/internal/service"
)

// Me bundles the PATCH /api/v1/me surface (email change).
type Me struct {
	DB     *sql.DB
	Auth   *service.AuthService
	Mailer service.Mailer
	Domain string
}

// NewMe constructs the handler.
func NewMe(db *sql.DB, auth *service.AuthService, mailer service.Mailer, domain string) *Me {
	return &Me{DB: db, Auth: auth, Mailer: mailer, Domain: domain}
}

// Get handles GET /api/v1/me. Returns the authenticated user's identity or 401.
// Frontend AuthContext calls this on mount to hydrate session state.
func (m *Me) Get(w http.ResponseWriter, r *http.Request) {
	uid := mw.UserIDFrom(r.Context())
	if uid == 0 {
		problem.Write(w, r, http.StatusUnauthorized, "unauthorized", "login required")
		return
	}
	var (
		email         string
		emailVerified sql.NullTime
		locale        sql.NullString
		isStaff       sql.NullBool
	)
	err := m.DB.QueryRowContext(r.Context(),
		`SELECT email, email_verified_at, locale, is_staff FROM users WHERE id=$1`, uid).
		Scan(&email, &emailVerified, &locale, &isStaff)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			problem.Write(w, r, http.StatusUnauthorized, "unauthorized", "user not found")
			return
		}
		problem.Write(w, r, http.StatusInternalServerError, "internal", "db error")
		return
	}
	resp := map[string]any{
		"id":             uid,
		"email":          email,
		"email_verified": emailVerified.Valid,
		"is_staff":       isStaff.Valid && isStaff.Bool,
	}
	if locale.Valid {
		resp["locale"] = locale.String
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(resp)
}

type patchMeReq struct {
	Email           string `json:"email"`
	CurrentPassword string `json:"current_password"`
}

// Patch handles PATCH /api/v1/me. Email change triggers a reverification mail
// to the NEW address; the swap happens only after /verify-email completes.
func (m *Me) Patch(w http.ResponseWriter, r *http.Request) {
	uid := mw.UserIDFrom(r.Context())
	if uid == 0 {
		problem.Write(w, r, http.StatusUnauthorized, "unauthorized", "login required")
		return
	}
	var in patchMeReq
	if err := json.NewDecoder(io.LimitReader(r.Body, 4*1024)).Decode(&in); err != nil {
		problem.Write(w, r, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	newEmail := strings.ToLower(strings.TrimSpace(in.Email))
	if !looksLikeEmail(newEmail) {
		problem.Write(w, r, http.StatusBadRequest, "bad_request", "invalid email")
		return
	}
	// Verify password.
	var (
		currentEmail string
		hash         string
	)
	err := m.DB.QueryRowContext(r.Context(),
		`SELECT email, password_hash FROM users WHERE id=$1`, uid).Scan(&currentEmail, &hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			problem.Write(w, r, http.StatusUnauthorized, "invalid_credentials", "invalid credentials")
			return
		}
		problem.Write(w, r, http.StatusInternalServerError, "internal", "db error")
		return
	}
	ok, _ := service.VerifyPassword(in.CurrentPassword, hash)
	if !ok {
		problem.Write(w, r, http.StatusUnauthorized, "invalid_credentials", "invalid credentials")
		return
	}
	if newEmail == strings.ToLower(currentEmail) {
		// No-op change: respond 202 without further work (no enumeration risk).
		writeAccepted(w)
		return
	}

	// Insert pending verification row. Uniqueness enumeration protection:
	// always respond 202 even if email already taken (log the collision).
	token := service.NewToken(32)
	_, err = m.DB.ExecContext(r.Context(),
		`INSERT INTO email_verifications (user_id, token, new_email, expires_at, created_at)
		 VALUES ($1, $2, $3, $4, NOW())`,
		uid, token, newEmail, time.Now().Add(24*time.Hour))
	if err != nil {
		// If the table/columns don't match (schema drift), fall back silently
		// to still emit 202 — the mailer won't send.
		problem.Write(w, r, http.StatusInternalServerError, "internal", "db error")
		return
	}
	if m.Mailer != nil {
		link := "https://" + m.Domain + "/verify-email?token=" + token
		body := "Conferma il cambio email per Pesa:\n\n" + link + "\n\nLink valido 24 ore.\n"
		_ = m.Mailer.Send(r.Context(), newEmail, "Conferma cambio email — Pesa", body)
	}
	writeAccepted(w)
}

func writeAccepted(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]any{"message": "verification_sent"})
}

func looksLikeEmail(s string) bool {
	if len(s) < 3 || len(s) > 254 {
		return false
	}
	at := strings.IndexByte(s, '@')
	if at <= 0 || at == len(s)-1 {
		return false
	}
	if strings.Contains(s, " ") {
		return false
	}
	return strings.Contains(s[at+1:], ".")
}
