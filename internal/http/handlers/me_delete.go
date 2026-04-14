package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/alexedwards/scs/v2"

	mw "github.com/thecomputer07/pesa/internal/http/middleware"
	"github.com/thecomputer07/pesa/internal/http/problem"
	"github.com/thecomputer07/pesa/internal/service"
	"github.com/thecomputer07/pesa/internal/service/erasure"
)

// MeDelete handles DELETE /api/v1/me (self-service erasure).
type MeDelete struct {
	DB      *sql.DB
	Session *scs.SessionManager
	Erasure *erasure.Service
	Secure  bool
	Auth    *service.AuthService
}

// NewMeDelete constructs the handler.
func NewMeDelete(db *sql.DB, sm *scs.SessionManager, er *erasure.Service, auth *service.AuthService, secure bool) *MeDelete {
	return &MeDelete{DB: db, Session: sm, Erasure: er, Auth: auth, Secure: secure}
}

type deleteReq struct {
	CurrentPassword  string `json:"current_password"`
	ConfirmationText string `json:"confirmation_text"`
	Notes            string `json:"notes"`
}

// Handle serves DELETE /api/v1/me.
func (h *MeDelete) Handle(w http.ResponseWriter, r *http.Request) {
	uid := mw.UserIDFrom(r.Context())
	if uid == 0 {
		problem.Write(w, r, http.StatusUnauthorized, "unauthorized", "login required")
		return
	}
	var in deleteReq
	if err := json.NewDecoder(io.LimitReader(r.Body, 4*1024)).Decode(&in); err != nil {
		problem.Write(w, r, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	if strings.TrimSpace(in.ConfirmationText) != "ELIMINA" {
		problem.Write(w, r, http.StatusBadRequest, "confirmation_mismatch", "confirmation_text must be ELIMINA")
		return
	}
	// Verify password.
	var hash string
	err := h.DB.QueryRowContext(r.Context(), `SELECT password_hash FROM users WHERE id=$1`, uid).Scan(&hash)
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
	// Cascade.
	if _, err := h.Erasure.Erase(r.Context(), uid, "self-service", in.Notes); err != nil {
		if errors.Is(err, erasure.ErrUserNotFound) {
			problem.Write(w, r, http.StatusNotFound, "not_found", "user already erased")
			return
		}
		problem.Write(w, r, http.StatusInternalServerError, "internal", "erasure failed")
		return
	}
	// Clear session.
	_ = h.Session.Destroy(r.Context())
	http.SetCookie(w, &http.Cookie{
		Name: "pesa_csrf", Value: "", Path: "/", MaxAge: -1,
		Secure: h.Secure, SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusNoContent)
}
