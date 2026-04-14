package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/thecomputer07/pesa/internal/http/problem"
)

// GlobalRateLimit applies SEC-17 — 100 req/min/IP via the package-level limiter.
func GlobalRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if !globalLimiter.Allow("global:" + ip) {
			problem.Write(w, r, http.StatusTooManyRequests, "rate_limited", "global rate limit exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
