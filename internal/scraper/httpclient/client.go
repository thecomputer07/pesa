// Package httpclient provides a polite HTTP client for the Pesa scrapers.
//
// Features:
//   - User-Agent identification
//   - Configurable min-interval rate limiter with jitter
//   - Exponential-backoff retries on 5xx / timeout (not on 4xx)
//   - Filesystem response cache keyed by sha256(url)
package httpclient

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// PoliteClient wraps *http.Client with rate limiting and disk cache.
type PoliteClient struct {
	HTTP        *http.Client
	UserAgent   string
	MinInterval time.Duration
	Jitter      time.Duration
	CacheDir    string

	mu      sync.Mutex
	nextOK  time.Time
	backoff []time.Duration
}

// New constructs a PoliteClient with the plan-02 defaults.
func New(userAgent string, minInterval time.Duration, cacheDir string) *PoliteClient {
	return &PoliteClient{
		HTTP:        &http.Client{Timeout: 30 * time.Second},
		UserAgent:   userAgent,
		MinInterval: minInterval,
		Jitter:      500 * time.Millisecond,
		CacheDir:    cacheDir,
		backoff:     []time.Duration{2 * time.Second, 5 * time.Second, 15 * time.Second},
	}
}

// Get performs a GET against url, returning body bytes.
// fromCache is true when the response came from CacheDir and no HTTP call was made.
func (c *PoliteClient) Get(ctx context.Context, url string) (body []byte, fromCache bool, err error) {
	if c.CacheDir != "" {
		if b, ok := c.readCache(url); ok {
			return b, true, nil
		}
	}

	c.waitSlot()

	var lastErr error
	for attempt := 0; attempt <= len(c.backoff); attempt++ {
		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if reqErr != nil {
			return nil, false, reqErr
		}
		req.Header.Set("User-Agent", c.UserAgent)
		req.Header.Set("Accept", "text/html,application/xhtml+xml,*/*;q=0.8")

		resp, httpErr := c.HTTP.Do(req)
		if httpErr != nil {
			lastErr = httpErr
			if attempt < len(c.backoff) {
				c.sleep(ctx, c.backoff[attempt])
				continue
			}
			return nil, false, lastErr
		}

		// 4xx: do not retry.
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			b, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			return nil, false, fmt.Errorf("http %d: %s", resp.StatusCode, string(b))
		}

		// 5xx: retry with backoff.
		if resp.StatusCode >= 500 {
			_ = resp.Body.Close()
			lastErr = fmt.Errorf("http %d", resp.StatusCode)
			if attempt < len(c.backoff) {
				c.sleep(ctx, c.backoff[attempt])
				continue
			}
			return nil, false, lastErr
		}

		b, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			return nil, false, readErr
		}

		if c.CacheDir != "" {
			_ = c.writeCache(url, b)
		}
		return b, false, nil
	}
	return nil, false, lastErr
}

func (c *PoliteClient) waitSlot() {
	c.mu.Lock()
	now := time.Now()
	if now.Before(c.nextOK) {
		wait := c.nextOK.Sub(now)
		c.mu.Unlock()
		time.Sleep(wait)
		c.mu.Lock()
	}
	jitter := time.Duration(0)
	if c.Jitter > 0 {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(2*c.Jitter)))
		jitter = time.Duration(n.Int64()) - c.Jitter
	}
	c.nextOK = time.Now().Add(c.MinInterval + jitter)
	c.mu.Unlock()
}

func (c *PoliteClient) sleep(ctx context.Context, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}

// HashURL returns the sha256 hex digest used as cache key.
func HashURL(url string) string {
	sum := sha256.Sum256([]byte(url))
	return hex.EncodeToString(sum[:])
}

func (c *PoliteClient) cachePath(url string) string {
	return filepath.Join(c.CacheDir, HashURL(url)+".html")
}

func (c *PoliteClient) readCache(url string) ([]byte, bool) {
	b, err := os.ReadFile(c.cachePath(url))
	if err != nil {
		return nil, false
	}
	return b, true
}

func (c *PoliteClient) writeCache(url string, body []byte) error {
	if err := os.MkdirAll(c.CacheDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(c.cachePath(url), body, 0o644)
}
