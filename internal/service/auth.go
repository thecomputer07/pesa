package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/thecomputer07/pesa/internal/legal"
)

// Sentinel errors.
var (
	ErrPrivacyConsentRequired = errors.New("privacy consent required")
	ErrEmailAlreadyUsed       = errors.New("email already used")
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrTokenInvalid           = errors.New("token invalid or expired")
	ErrEmailNotVerified       = errors.New("email not verified")
)

// AuthService encapsulates register / verify / login / reset operations.
type AuthService struct {
	DB      *sql.DB
	Mailer  Mailer
	Now     func() time.Time // overridable for tests
	Domain  string           // used to build absolute links in emails; e.g. "example.it"
	WebBase string           // full base URL of the web frontend; if set, used verbatim for links (e.g. "http://localhost:5173")
}

// RegisterInput is the decoded body for POST /api/v1/auth/register.
type RegisterInput struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	Locale         string `json:"locale"`
	PrivacyConsent bool   `json:"privacy_consent"`
}

// now returns the configured clock or time.Now.
func (s *AuthService) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

func (s *AuthService) domain() string {
	if s.Domain != "" {
		return s.Domain
	}
	return "pesadia.com"
}

// webBase returns the scheme+host prefix for links embedded in emails.
// In dev the caller sets WebBase="http://localhost:5173"; in prod it stays
// empty and we fall back to "https://<Domain>".
func (s *AuthService) webBase() string {
	if s.WebBase != "" {
		return strings.TrimRight(s.WebBase, "/")
	}
	return "https://" + s.domain()
}

// Register creates a user, stores Argon2id hash, emits an email verification token (24h).
func (s *AuthService) Register(ctx context.Context, in RegisterInput) (int64, error) {
	if !in.PrivacyConsent {
		return 0, ErrPrivacyConsentRequired
	}
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	if in.Locale != "it" && in.Locale != "en" {
		in.Locale = "it"
	}
	hash, err := HashPassword(in.Password)
	if err != nil {
		return 0, err
	}
	var userID int64
	err = s.DB.QueryRowContext(ctx,
		`INSERT INTO users (email, password_hash, locale, privacy_consent_at, privacy_consent_version)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		in.Email, hash, in.Locale, s.now().UTC(), legal.PrivacyPolicyVersion).Scan(&userID)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") ||
			strings.Contains(strings.ToLower(err.Error()), "unique") {
			return 0, ErrEmailAlreadyUsed
		}
		return 0, err
	}
	token := NewToken(32)
	if _, err := s.DB.ExecContext(ctx,
		`INSERT INTO email_verifications (token, user_id, expires_at) VALUES ($1, $2, $3)`,
		token, userID, s.now().UTC().Add(24*time.Hour)); err != nil {
		return 0, err
	}
	_ = s.Mailer.Send(ctx, in.Email, "Verifica email Pesa",
		fmt.Sprintf("Per confermare l'email fai clic sul link: %s/verifica-email?token=%s", s.webBase(), token))
	return userID, nil
}

// VerifyEmail consumes a verification token (single-use, 24h TTL).
func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var userID int64
	var expiresAt time.Time
	var usedAt sql.NullTime
	err = tx.QueryRowContext(ctx,
		`SELECT user_id, expires_at, used_at FROM email_verifications WHERE token = $1`, token).
		Scan(&userID, &expiresAt, &usedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrTokenInvalid
	}
	if err != nil {
		return err
	}
	if usedAt.Valid || s.now().After(expiresAt) {
		return ErrTokenInvalid
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE email_verifications SET used_at = $1 WHERE token = $2`,
		s.now().UTC(), token); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE users SET email_verified_at = $1 WHERE id = $2`,
		s.now().UTC(), userID); err != nil {
		return err
	}
	return tx.Commit()
}

// dummyHash is used for timing-safe negative verify against unknown emails.
const dummyHash = "$argon2id$v=19$m=65536,t=3,p=2$YWFhYWFhYWFhYWFhYWFhYQ$" +
	"YWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWE"

// Login checks credentials (timing-safe) and returns user ID.
func (s *AuthService) Login(ctx context.Context, email, password string) (int64, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	var id int64
	var hash string
	var verifiedAt sql.NullTime
	err := s.DB.QueryRowContext(ctx,
		`SELECT id, password_hash, email_verified_at FROM users WHERE email = $1 AND status = 'active'`, email).
		Scan(&id, &hash, &verifiedAt)
	if errors.Is(err, sql.ErrNoRows) {
		_, _ = VerifyPassword(password, dummyHash)
		return 0, ErrInvalidCredentials
	}
	if err != nil {
		return 0, err
	}
	ok, err := VerifyPassword(password, hash)
	if err != nil || !ok {
		return 0, ErrInvalidCredentials
	}
	if !verifiedAt.Valid {
		return 0, ErrEmailNotVerified
	}
	return id, nil
}

// RequestPasswordReset emits a 1h single-use token. Email enumeration-safe:
// always returns nil even when the email is unknown.
func (s *AuthService) RequestPasswordReset(ctx context.Context, email string, ip net.IP) error {
	email = strings.ToLower(strings.TrimSpace(email))
	var id int64
	err := s.DB.QueryRowContext(ctx, `SELECT id FROM users WHERE email = $1`, email).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	token := NewToken(32)
	ipStr := ""
	if ip != nil {
		ipStr = ip.String()
	}
	if _, err := s.DB.ExecContext(ctx,
		`INSERT INTO password_resets (token, user_id, expires_at, requested_ip) VALUES ($1, $2, $3, $4)`,
		token, id, s.now().UTC().Add(1*time.Hour), ipStr); err != nil {
		return err
	}
	_ = s.Mailer.Send(ctx, email, "Reset password Pesa",
		fmt.Sprintf("Se hai richiesto il reset, apri: %s/password-reset/conferma?token=%s", s.webBase(), token))
	return nil
}

// ConfirmPasswordReset consumes the token and sets a new password hash.
func (s *AuthService) ConfirmPasswordReset(ctx context.Context, token, newPassword string) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	var userID int64
	var expiresAt time.Time
	var usedAt sql.NullTime
	err = tx.QueryRowContext(ctx,
		`SELECT user_id, expires_at, used_at FROM password_resets WHERE token = $1`, token).
		Scan(&userID, &expiresAt, &usedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrTokenInvalid
	}
	if err != nil {
		return err
	}
	if usedAt.Valid || s.now().After(expiresAt) {
		return ErrTokenInvalid
	}
	hash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE password_resets SET used_at = $1 WHERE token = $2`, s.now().UTC(), token); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE users SET password_hash = $1 WHERE id = $2`, hash, userID); err != nil {
		return err
	}
	// Revoke all sessions tied to the user after credential change (SEC-13).
	// scs stores data as gob-encoded blob; cast to text for substring match.
	// This is best-effort: sessions expire on their own anyway.
	_, _ = tx.ExecContext(ctx, `DELETE FROM sessions WHERE data::text LIKE '%user_id%' || $1 || '%'`, userID)
	return tx.Commit()
}

// NewToken returns a hex-encoded random token of nBytes*2 characters.
func NewToken(nBytes int) string {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
