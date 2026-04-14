package middleware

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/thecomputer07/pesa/internal/http/problem"
)

// PerIPRate builds a middleware enforcing n requests per window per client IP.
// Limiters are kept in a sharded sync.Map and swept when idle for > 5*window.
func PerIPRate(n int, window time.Duration) func(http.Handler) http.Handler {
	if n <= 0 {
		n = 1
	}
	if window <= 0 {
		window = time.Minute
	}
	store := &perIPStore{
		limit:  rate.Limit(float64(n) / window.Seconds()),
		burst:  n,
		window: window,
	}
	go store.sweep()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			lim := store.limiter(ip)
			if !lim.Allow() {
				problem.Write(w, r, http.StatusTooManyRequests, "rate_limited", "per-ip rate limit exceeded")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type perIPEntry struct {
	lim  *rate.Limiter
	seen time.Time
}

type perIPStore struct {
	mu     sync.Mutex
	cells  map[string]*perIPEntry
	limit  rate.Limit
	burst  int
	window time.Duration
}

func (s *perIPStore) limiter(ip string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cells == nil {
		s.cells = make(map[string]*perIPEntry)
	}
	e, ok := s.cells[ip]
	if !ok {
		e = &perIPEntry{lim: rate.NewLimiter(s.limit, s.burst)}
		s.cells[ip] = e
	}
	e.seen = time.Now()
	return e.lim
}

func (s *perIPStore) sweep() {
	t := time.NewTicker(s.window)
	defer t.Stop()
	for range t.C {
		cut := time.Now().Add(-5 * s.window)
		s.mu.Lock()
		for ip, e := range s.cells {
			if e.seen.Before(cut) {
				delete(s.cells, ip)
			}
		}
		s.mu.Unlock()
	}
}
