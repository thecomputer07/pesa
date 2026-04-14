// Package robots wraps temoto/robotstxt with a tiny in-memory cache keyed by host.
package robots

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/temoto/robotstxt"
)

// Checker loads and caches robots.txt per host.
type Checker struct {
	UserAgent string
	HTTP      *http.Client

	mu    sync.Mutex
	cache map[string]*robotstxt.RobotsData
}

// NewChecker builds a Checker with a 10s timeout HTTP client.
func NewChecker(userAgent string) *Checker {
	return &Checker{
		UserAgent: userAgent,
		HTTP:      &http.Client{Timeout: 10 * time.Second},
		cache:     map[string]*robotstxt.RobotsData{},
	}
}

// Allowed reports whether targetURL is allowed for this checker's UA.
// A missing robots.txt (404) is treated as fully-allowed per convention.
func (c *Checker) Allowed(ctx context.Context, targetURL string) (bool, error) {
	u, err := url.Parse(targetURL)
	if err != nil {
		return false, err
	}
	host := u.Scheme + "://" + u.Host

	c.mu.Lock()
	data, ok := c.cache[host]
	c.mu.Unlock()

	if !ok {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, host+"/robots.txt", nil)
		req.Header.Set("User-Agent", c.UserAgent)
		resp, rerr := c.HTTP.Do(req)
		if rerr != nil {
			return false, fmt.Errorf("robots fetch: %w", rerr)
		}
		defer func() { _ = resp.Body.Close() }()
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusNotFound {
			body = []byte{} // fully allowed
		}
		parsed, perr := robotstxt.FromBytes(body)
		if perr != nil {
			return false, fmt.Errorf("robots parse: %w", perr)
		}
		c.mu.Lock()
		c.cache[host] = parsed
		c.mu.Unlock()
		data = parsed
	}

	group := data.FindGroup(c.UserAgent)
	return group.Test(u.Path), nil
}
