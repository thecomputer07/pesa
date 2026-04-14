package service

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// KeyedLimiter applies a composite-key token-bucket rate limit.
// Keys combine a scope (e.g. "register:ip") with a value (IP or email).
// Stale buckets are GC'd every 10 minutes.
type KeyedLimiter struct {
	mu      sync.Mutex
	buckets map[string]*entry
	r       rate.Limit
	burst   int
	ttl     time.Duration
	stop    chan struct{}
}

type entry struct {
	lim      *rate.Limiter
	lastSeen time.Time
}

// NewKeyedLimiter returns a limiter with the given rate and burst.
func NewKeyedLimiter(r rate.Limit, burst int, ttl time.Duration) *KeyedLimiter {
	l := &KeyedLimiter{
		buckets: map[string]*entry{},
		r:       r,
		burst:   burst,
		ttl:     ttl,
		stop:    make(chan struct{}),
	}
	go l.gc()
	return l
}

// Allow reports whether an event for this key is permitted.
func (l *KeyedLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	e, ok := l.buckets[key]
	if !ok {
		e = &entry{lim: rate.NewLimiter(l.r, l.burst)}
		l.buckets[key] = e
	}
	e.lastSeen = time.Now()
	return e.lim.Allow()
}

// Close stops the internal GC goroutine.
func (l *KeyedLimiter) Close() { close(l.stop) }

func (l *KeyedLimiter) gc() {
	t := time.NewTicker(10 * time.Minute)
	defer t.Stop()
	for {
		select {
		case <-l.stop:
			return
		case <-t.C:
			l.mu.Lock()
			cutoff := time.Now().Add(-l.ttl)
			for k, e := range l.buckets {
				if e.lastSeen.Before(cutoff) {
					delete(l.buckets, k)
				}
			}
			l.mu.Unlock()
		}
	}
}
