package middleware

import (
	"net/http"
	"strings"

	"github.com/thecomputer07/pesa/internal/http/problem"
)

// csrfExempt are pre-session endpoints protected by rate limits / GET semantics.
var csrfExempt = map[string]struct{}{
	"/api/v1/auth/register":               {},
	"/api/v1/auth/login":                  {},
	"/api/v1/auth/password-reset/request": {},
	"/verify-email":                       {},
}

// CSRF enforces the double-submit token pattern on mutating verbs.
// pesa_csrf cookie value MUST equal the X-CSRF-Token request header.
func CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
			return
		}
		if _, ok := csrfExempt[r.URL.Path]; ok {
			next.ServeHTTP(w, r)
			return
		}
		c, err := r.Cookie("pesa_csrf")
		if err != nil || c.Value == "" {
			problem.Write(w, r, http.StatusForbidden, "csrf", "CSRF token missing")
			return
		}
		hdr := strings.TrimSpace(r.Header.Get("X-CSRF-Token"))
		if hdr == "" || hdr != c.Value {
			problem.Write(w, r, http.StatusForbidden, "csrf", "CSRF token mismatched")
			return
		}
		next.ServeHTTP(w, r)
	})
}
