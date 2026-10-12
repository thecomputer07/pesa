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
