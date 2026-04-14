package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/thecomputer07/pesa/internal/http/problem"
)

// Recover catches panics, logs with stack + req_id, and returns RFC 7807.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic",
					"err", rec,
					"stack", string(debug.Stack()),
					"req_id", ReqIDFrom(r.Context()),
				)
				problem.Write(w, r, http.StatusInternalServerError, "internal", "unexpected error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
