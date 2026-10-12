// Package cache is a small shared response cache (Redis) for public data:
// anonymous search results and public food details. It is strictly optional:
// without REDIS_ADDR, or while Redis is down, every call is a miss and the
// backend reads PostgreSQL as before.
package cache

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache stores opaque byte values under string keys with a TTL.
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, bool)
	Set(ctx context.Context, key string, val []byte, ttl time.Duration)
}

// Noop is the cache used without Redis: always a miss.
type Noop struct{}

func (Noop) Get(context.Context, string) ([]byte, bool)         { return nil, false }
func (Noop) Set(context.Context, string, []byte, time.Duration) {}

// Redis is a Cache on a Redis server. Short timeouts: a slow or missing Redis
// must never slow a request down more than a few milliseconds.
type Redis struct {
	c      *redis.Client
	prefix string
	// lastErrLog throttles error logs to one per minute (unix seconds).
	lastErrLog atomic.Int64
}

// New returns a Redis cache for addr ("host:port"), or Noop when addr is empty.
// Keys are namespaced with prefix (e.g. "pesa:").
func New(addr, prefix string) Cache {
	if addr == "" {
		return Noop{}
	}
	return &Redis{
		prefix: prefix,
		c: redis.NewClient(&redis.Options{
			Addr:         addr,
			DialTimeout:  200 * time.Millisecond,
			ReadTimeout:  50 * time.Millisecond,
			WriteTimeout: 50 * time.Millisecond,
			PoolSize:     32,
			MaxRetries:   -1, // a miss is cheaper than a retry
		}),
	}
}

func (r *Redis) Get(ctx context.Context, key string) ([]byte, bool) {
	b, err := r.c.Get(ctx, r.prefix+key).Bytes()
	if err != nil {
		if err != redis.Nil {
			r.logErr("get", err)
		}
		return nil, false
	}
	return b, true
}

func (r *Redis) Set(ctx context.Context, key string, val []byte, ttl time.Duration) {
	if err := r.c.Set(ctx, r.prefix+key, val, ttl).Err(); err != nil {
		r.logErr("set", err)
	}
}

func (r *Redis) logErr(op string, err error) {
	now := time.Now().Unix()
	if last := r.lastErrLog.Load(); now-last >= 60 && r.lastErrLog.CompareAndSwap(last, now) {
		slog.Warn("redis cache unavailable, using the database", "op", op, "err", err)
	}
}
