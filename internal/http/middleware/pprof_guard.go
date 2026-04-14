package middleware

import (
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/thecomputer07/pesa/internal/http/problem"
)

// MountPprof attaches /debug/pprof/* to the chi router only when
// PESA_ENABLE_PPROF=1 AND the request carries a matching PPROF_TOKEN header.
// Interim measure (PERF-14): an admin-flag-gated session check arrives later.
func MountPprof(r chi.Router) {
	if os.Getenv("PESA_ENABLE_PPROF") != "1" {
		return
	}
	tok := os.Getenv("PPROF_TOKEN")
	gate := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			if tok == "" || req.Header.Get("X-Pprof-Token") != tok {
				problem.Write(w, req, http.StatusForbidden, "forbidden", "pprof requires admin token")
				return
			}
			h(w, req)
		}
	}
	r.Get("/debug/pprof/", gate(pprof.Index))
	r.Get("/debug/pprof/cmdline", gate(pprof.Cmdline))
	r.Get("/debug/pprof/profile", gate(pprof.Profile))
	r.Get("/debug/pprof/symbol", gate(pprof.Symbol))
	r.Get("/debug/pprof/trace", gate(pprof.Trace))
	for _, p := range []string{"goroutine", "heap", "threadcreate", "block", "mutex", "allocs"} {
		r.Get("/debug/pprof/"+p, gate(pprof.Handler(p).ServeHTTP))
	}
}
