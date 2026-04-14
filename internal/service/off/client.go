package off

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

// UserAgent is the contact-bearing UA OFF asks clients to send.
const UserAgent = "Pesa/1.0 (+https://github.com/thecomputer07/pesa)"

// Client is a rate-limited OFF HTTP client (1 req/sec global).
type Client struct {
	HTTP    *http.Client
	Limiter *rate.Limiter
	BaseURL string
	UA      string
}

// NewClient builds a Client with sensible defaults.
func NewClient(baseURL string) *Client {
	return &Client{
		HTTP:    &http.Client{Timeout: 5 * time.Second},
		Limiter: rate.NewLimiter(rate.Every(time.Second), 1),
		BaseURL: strings.TrimRight(baseURL, "/"),
		UA:      UserAgent,
	}
}

// Fetch performs GET /api/v2/product/<ean>.json and returns the body + status.
// Caller treats status=404 as a legitimate miss (err is nil in that case).
func (c *Client) Fetch(ctx context.Context, ean string) ([]byte, int, error) {
	if err := c.Limiter.Wait(ctx); err != nil {
		return nil, 0, err
	}
	url := fmt.Sprintf("%s/api/v2/product/%s.json", c.BaseURL, ean)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("User-Agent", c.UA)
	req.Header.Set("Accept", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20)) // 2MB cap
	if err != nil {
		return nil, resp.StatusCode, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return body, http.StatusNotFound, nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return body, resp.StatusCode, fmt.Errorf("off: unexpected status %d", resp.StatusCode)
	}
	return body, resp.StatusCode, nil
}

// ErrNotFound is returned when an EAN has no OFF product.
var ErrNotFound = errors.New("off: product not found")
