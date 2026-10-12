// Package handlers contains HTTP handlers for the Pesa API surface.
package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/scs/v2"
	"golang.org/x/time/rate"

	"github.com/thecomputer07/pesa/internal/http/problem"
	"github.com/thecomputer07/pesa/internal/service"
)

// Handlers groups the auth surface with its dependencies.
type Handlers struct {
	Auth    *service.AuthService
	Session *scs.SessionManager
	Secure  bool // sets Secure flag on cookies (true in prod)

	RegisterLimiter     *service.KeyedLimiter
	LoginLimiter        *service.KeyedLimiter
	ResetRequestLimiter *service.KeyedLimiter
}

// New builds a Handlers with the rate-limiter profiles from plan-03 §Task 4.
// In dev (secure=false) i limiti sono molto più permissivi per non bloccare il test locale.
func New(a *service.AuthService, sm *scs.SessionManager, secure bool) *Handlers {
	if !secure {
		// dev: 100 burst ogni 1s → praticamente illimitato
		return &Handlers{
			Auth:                a,
			Session:             sm,
			Secure:              secure,
			RegisterLimiter:     service.NewKeyedLimiter(rate.Every(time.Second), 100, 10*time.Minute),
			LoginLimiter:        service.NewKeyedLimiter(rate.Every(time.Second), 100, 10*time.Minute),
			ResetRequestLimiter: service.NewKeyedLimiter(rate.Every(time.Second), 100, 10*time.Minute),
		}
	}
	return &Handlers{
		Auth:                a,
		Session:             sm,
		Secure:              secure,
		RegisterLimiter:     service.NewKeyedLimiter(rate.Every(12*time.Minute), 5, 2*time.Hour),
		LoginLimiter:        service.NewKeyedLimiter(rate.Every(90*time.Second), 10, 1*time.Hour),
		ResetRequestLimiter: service.NewKeyedLimiter(rate.Every(20*time.Minute), 3, 2*time.Hour),
	}
}

// ClientIP extracts the client IP: X-Real-IP from the reverse proxy, else RemoteAddr.
func ClientIP(r *http.Request) string {
	// X-Real-IP is set by Caddy from {client_ip}: the visitor's address even
	// behind Cloudflare (CF-Connecting-IP from trusted proxies), and it always
	// overwrites what the client sent. X-Forwarded-For is not used: behind
	// Cloudflare its first entry can be forged by the client.
	if ip := strings.TrimSpace(r.Header.Get("X-Real-IP")); ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// Register handles POST /api/v1/auth/register.
func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	ip := ClientIP(r)
	if !h.RegisterLimiter.Allow("register:ip:" + ip) {
		problem.Write(w, r, http.StatusTooManyRequests, "rate_limited", "Too many registrations from this IP")
		return
	}
	var in service.RegisterInput
	if err := json.NewDecoder(io.LimitReader(r.Body, 4*1024)).Decode(&in); err != nil {
		problem.Write(w, r, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	_, err := h.Auth.Register(r.Context(), in)
	switch {
	case errors.Is(err, service.ErrPrivacyConsentRequired):
		problem.Write(w, r, http.StatusUnprocessableEntity, "privacy_consent_required", "Consent to privacy policy required")
	case errors.Is(err, service.ErrEmailAlreadyUsed):
		problem.Write(w, r, http.StatusConflict, "email_already_used", "Email is already registered")
	case errors.Is(err, service.ErrPasswordTooShort):
		problem.Write(w, r, http.StatusBadRequest, "password_too_short", "Password too short (min 12 chars)")
	case err != nil:
		problem.Write(w, r, http.StatusInternalServerError, "internal", "Unexpected error")
	default:
		w.WriteHeader(http.StatusCreated)
	}
}

// VerifyEmail handles GET /verify-email?token=...
func (h *Handlers) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	tok := r.URL.Query().Get("token")
	if tok == "" {
		problem.Write(w, r, http.StatusBadRequest, "bad_request", "token required")
		return
	}
	if err := h.Auth.VerifyEmail(r.Context(), tok); err != nil {
		problem.Write(w, r, http.StatusBadRequest, "token_invalid", "token invalid or expired")
		return
	}
	http.Redirect(w, r, "/verified", http.StatusFound)
}

// Login handles POST /api/v1/auth/login.
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	ip := ClientIP(r)
	if !h.LoginLimiter.Allow("login:ip:" + ip) {
		problem.Write(w, r, http.StatusTooManyRequests, "rate_limited", "Too many login attempts")
		return
	}
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 4*1024)).Decode(&in); err != nil {
		problem.Write(w, r, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	uid, err := h.Auth.Login(r.Context(), in.Email, in.Password)
	switch {
	case errors.Is(err, service.ErrEmailNotVerified):
		problem.Write(w, r, http.StatusForbidden, "email_not_verified", "Email not verified")
		return
	case errors.Is(err, service.ErrInvalidCredentials):
		problem.Write(w, r, http.StatusUnauthorized, "invalid_credentials", "Invalid credentials")
		return
	case err != nil:
		problem.Write(w, r, http.StatusInternalServerError, "internal", "Unexpected error")
		return
	}
	// Bind the session.
	if err := h.Session.RenewToken(r.Context()); err != nil {
		problem.Write(w, r, http.StatusInternalServerError, "internal", "session renew failed")
		return
	}
	h.Session.Put(r.Context(), "user_id", uid)

	// Issue CSRF cookie (non-HttpOnly, double-submit pattern — cookie-policy.it.md §2).
	http.SetCookie(w, &http.Cookie{
		Name:     "pesa_csrf",
		Value:    service.NewToken(16),
		Path:     "/",
		Secure:   h.Secure,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: false,
	})
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

// Logout handles POST /api/v1/auth/logout.
func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	if err := h.Session.Destroy(r.Context()); err != nil {
		problem.Write(w, r, http.StatusInternalServerError, "internal", "logout failed")
		return
	}
	// Clear CSRF cookie.
	http.SetCookie(w, &http.Cookie{
		Name:     "pesa_csrf",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   h.Secure,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusNoContent)
}

// RequestPasswordReset handles POST /api/v1/auth/password-reset/request.
func (h *Handlers) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 1024)).Decode(&in); err != nil {
		problem.Write(w, r, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	if !h.ResetRequestLimiter.Allow("reset:email:" + strings.ToLower(strings.TrimSpace(in.Email))) {
		problem.Write(w, r, http.StatusTooManyRequests, "rate_limited", "Too many reset requests")
		return
	}
	ip := net.ParseIP(ClientIP(r))
	if err := h.Auth.RequestPasswordReset(r.Context(), in.Email, ip); err != nil {
		problem.Write(w, r, http.StatusInternalServerError, "internal", "Unexpected error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ConfirmPasswordReset handles POST /api/v1/auth/password-reset/confirm.
func (h *Handlers) ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 4*1024)).Decode(&in); err != nil {
		problem.Write(w, r, http.StatusBadRequest, "bad_request", "invalid JSON")
		return
	}
	err := h.Auth.ConfirmPasswordReset(r.Context(), in.Token, in.NewPassword)
	switch {
	case errors.Is(err, service.ErrTokenInvalid):
		problem.Write(w, r, http.StatusBadRequest, "token_invalid", "token invalid or expired")
	case errors.Is(err, service.ErrPasswordTooShort):
		problem.Write(w, r, http.StatusBadRequest, "password_too_short", "Password too short (min 12 chars)")
	case err != nil:
		problem.Write(w, r, http.StatusInternalServerError, "internal", "Unexpected error")
	default:
		w.WriteHeader(http.StatusNoContent)
	}
}
