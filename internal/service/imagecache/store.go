// Package imagecache provides content-addressed on-disk storage for
// third-party images (OFF, Wikimedia) with a paired license sidecar JSON.
//
// Files land under <Dir>/<source>/<sha>.<ext> with <sha>.license.json next to
// them. Writes are atomic (tmp + rename). Reads are paired: a bare binary with
// a missing sidecar is treated as not-found.
package imagecache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SidecarMeta is persisted alongside every cached image.
type SidecarMeta struct {
	Source      string `json:"source"`
	License     string `json:"license"`
	LicenseURL  string `json:"license_url"`
	Attribution string `json:"attribution"`
	SourceURL   string `json:"source_url,omitempty"`
	OriginalURL string `json:"original_url"`
	FetchedAt   string `json:"fetched_at"`
	ContentType string `json:"content_type"`
	Bytes       int64  `json:"bytes"`
}

// Store is a rooted directory holding content-addressed images.
type Store struct {
	Dir string
}

// AllowedSources restricts {source} path segments to a small allowlist.
var AllowedSources = map[string]struct{}{
	"off":       {},
	"wikimedia": {},
}

// ErrNotFound is returned when a requested sha has no cached file.
var ErrNotFound = errors.New("imagecache: not found")

// ErrUnsupportedContentType is returned when Put receives an unknown MIME.
var ErrUnsupportedContentType = errors.New("imagecache: unsupported content-type")

// ErrUnknownSource is returned when source is outside the allowlist.
var ErrUnknownSource = errors.New("imagecache: unknown source")

// NewStore ensures dir exists and returns a Store rooted there.
func NewStore(dir string) (*Store, error) {
	if dir == "" {
		return nil, errors.New("imagecache: empty dir")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &Store{Dir: dir}, nil
}

// extFromContentType maps a response content-type to a safe file extension.
func extFromContentType(ct string) (string, error) {
	ct = strings.ToLower(strings.TrimSpace(strings.SplitN(ct, ";", 2)[0]))
	switch ct {
	case "image/jpeg":
		return ".jpg", nil
	case "image/png":
		return ".png", nil
	case "image/webp":
		return ".webp", nil
	case "image/avif":
		return ".avif", nil
	case "image/gif":
		return ".gif", nil
	default:
		return "", fmt.Errorf("%w: %q", ErrUnsupportedContentType, ct)
	}
}

// Put stores body under <source>/<sha>.<ext> with a paired sidecar.
// Idempotent: if the binary already exists, the sidecar is (re)written.
// Returns (sha, ext) or an error.
func (s *Store) Put(source string, body []byte, contentType string, meta SidecarMeta) (string, string, error) {
	if _, ok := AllowedSources[source]; !ok {
		return "", "", ErrUnknownSource
	}
	ext, err := extFromContentType(contentType)
	if err != nil {
		return "", "", err
	}
	sum := sha256.Sum256(body)
	sha := hex.EncodeToString(sum[:])

	dir := filepath.Join(s.Dir, source)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", err
	}
	bin := filepath.Join(dir, sha+ext)
	side := filepath.Join(dir, sha+".license.json")

	if _, err := os.Stat(bin); errors.Is(err, os.ErrNotExist) {
		if err := atomicWrite(bin, body); err != nil {
			return "", "", err
		}
	} else if err != nil {
		return "", "", err
	}

	meta.ContentType = strings.TrimSpace(strings.SplitN(contentType, ";", 2)[0])
	meta.Bytes = int64(len(body))
	if meta.FetchedAt == "" {
		meta.FetchedAt = time.Now().UTC().Format(time.RFC3339)
	}
	sidebytes, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return "", "", err
	}
	if err := atomicWrite(side, sidebytes); err != nil {
		return "", "", err
	}
	return sha, ext, nil
}

// Exists reports whether a cached file + sidecar are present for (source, sha).
// The caller is expected to have validated source and sha already.
func (s *Store) Exists(source, shaWithExt string) bool {
	bin := filepath.Join(s.Dir, source, shaWithExt)
	if _, err := os.Stat(bin); err != nil {
		return false
	}
	base := strings.TrimSuffix(shaWithExt, filepath.Ext(shaWithExt))
	if _, err := os.Stat(filepath.Join(s.Dir, source, base+".license.json")); err != nil {
		return false
	}
	return true
}

// ReadMeta returns the sidecar meta for (source, shaWithExt).
func (s *Store) ReadMeta(source, shaWithExt string) (SidecarMeta, error) {
	if _, ok := AllowedSources[source]; !ok {
		return SidecarMeta{}, ErrUnknownSource
	}
	base := strings.TrimSuffix(shaWithExt, filepath.Ext(shaWithExt))
	side := filepath.Join(s.Dir, source, base+".license.json")
	data, err := os.ReadFile(side)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return SidecarMeta{}, ErrNotFound
		}
		return SidecarMeta{}, err
	}
	var m SidecarMeta
	if err := json.Unmarshal(data, &m); err != nil {
		return SidecarMeta{}, err
	}
	return m, nil
}

// Path returns the absolute file path for (source, shaWithExt). The caller
// must validate the path through filepath.Clean before serving.
func (s *Store) Path(source, shaWithExt string) (string, error) {
	if _, ok := AllowedSources[source]; !ok {
		return "", ErrUnknownSource
	}
	base := filepath.Join(s.Dir, source, shaWithExt)
	clean := filepath.Clean(base)
	root := filepath.Clean(filepath.Join(s.Dir, source))
	if !strings.HasPrefix(clean, root+string(filepath.Separator)) && clean != root {
		return "", errors.New("imagecache: path traversal rejected")
	}
	return clean, nil
}

func atomicWrite(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	if _, err := io.Copy(tmp, strings.NewReader(string(data))); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name())
		return err
	}
	return os.Rename(tmp.Name(), path)
}
