// Package export produces personal-data exports (PRIV-01).
package export

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"strconv"
	"strings"
	"time"
)

// ErrTokenInvalid indicates a missing, malformed or tampered signed token.
var ErrTokenInvalid = errors.New("export token invalid")

// Sign returns a 43-char base64url (no padding) HMAC-SHA256 token binding
// jobID + expiry to the server secret. Format of signed payload is
// "jobID:unixExpiry"; the returned token carries only the HMAC — the caller
// stores jobID+expires_at alongside the token in the data_export_jobs row.
func Sign(jobID int64, expiresAt time.Time, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	var buf [16]byte
	binary.BigEndian.PutUint64(buf[0:8], uint64(jobID))
	binary.BigEndian.PutUint64(buf[8:16], uint64(expiresAt.Unix()))
	mac.Write(buf[:])
	sum := mac.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(sum)
}

// Verify recomputes the HMAC from (jobID, expiresAt) and compares in
// constant time. Returns ErrTokenInvalid on any mismatch or if the token is
// past expiry.
func Verify(token string, jobID int64, expiresAt time.Time, secret []byte) error {
	if len(token) != 43 {
		return ErrTokenInvalid
	}
	if time.Now().After(expiresAt) {
		return ErrTokenInvalid
	}
	expected := Sign(jobID, expiresAt, secret)
	if !hmac.Equal([]byte(expected), []byte(token)) {
		return ErrTokenInvalid
	}
	return nil
}

// TokenIsWellFormed returns true if token has the expected length/charset.
// Used as a cheap guard before hitting the DB.
func TokenIsWellFormed(token string) bool {
	if len(token) != 43 {
		return false
	}
	for _, c := range token {
		if !(c == '-' || c == '_' ||
			(c >= 'A' && c <= 'Z') ||
			(c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}

// FormatDuration renders a Go duration for debug/log use; stable formatting
// avoids non-deterministic unit pluralisation in fixture tests.
func FormatDuration(d time.Duration) string {
	if d >= time.Hour {
		return strconv.FormatFloat(d.Hours(), 'f', 1, 64) + "h"
	}
	return d.Truncate(time.Second).String()
}

// trimPseudonym returns the 16-char Phase 1 logger pseudonym given the full
// sha256 hex digest. Isolated to a helper so both service/erasure and
// service/export can share it.
func trimPseudonym(fullHex string) string {
	if len(fullHex) < 16 {
		return fullHex
	}
	return strings.ToLower(fullHex[:16])
}
