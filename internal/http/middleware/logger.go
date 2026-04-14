package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type statusWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (s *statusWriter) WriteHeader(c int) { s.status = c; s.ResponseWriter.WriteHeader(c) }

func (s *statusWriter) Write(b []byte) (int, error) {
	n, err := s.ResponseWriter.Write(b)
	s.bytes += n
	return n, err
}

// logSalt is read once at process start. PRIV-06 / PRIV-07 — pseudonymize email.
var logSalt = os.Getenv("LOG_SALT") //nolint:gochecknoglobals

// Pseudonym returns sha256(email+salt)[:16] hex; "anon" for empty input.
func Pseudonym(email string) string {
	if email == "" {
		return "anon"
	}
	h := sha256.Sum256([]byte(email + logSalt))
	return hex.EncodeToString(h[:])[:16]
}

// Logger emits one structured slog line per request.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(sw, r)
		slog.Info("http",
			"req_id", ReqIDFrom(r.Context()),
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"bytes", sw.bytes,
			"duration_ms", time.Since(start).Milliseconds(),
			"user_pseudonym", Pseudonym(EmailFrom(r.Context())),
		)
	})
}
