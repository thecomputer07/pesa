package imagecache

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// UserAgent is the UA used for image downloads.
const UserAgent = "Pesa/1.0 (+https://github.com/thecomputer07/pesa)"

// MaxBodyBytes caps individual image downloads at 5 MiB.
const MaxBodyBytes = 5 << 20

// Fetcher is a small HTTP client bound to defaults suitable for image egress.
type Fetcher struct {
	HTTP *http.Client
}

// NewFetcher builds a Fetcher with conservative timeouts.
func NewFetcher() *Fetcher {
	return &Fetcher{
		HTTP: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				ResponseHeaderTimeout: 3 * time.Second,
			},
		},
	}
}

// Fetch downloads url and returns (body, contentType). Rejects non-image
// content-types and payloads larger than MaxBodyBytes.
func (f *Fetcher) Fetch(ctx context.Context, url string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", UserAgent)
	resp, err := f.HTTP.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", fmt.Errorf("imagecache: unexpected status %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(strings.ToLower(ct), "image/") {
		return nil, "", errors.New("imagecache: non-image content-type: " + ct)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxBodyBytes+1))
	if err != nil {
		return nil, "", err
	}
	if int64(len(body)) > MaxBodyBytes {
		return nil, "", errors.New("imagecache: body exceeds 5MB")
	}
	return body, ct, nil
}
