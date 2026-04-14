package handlers

import (
	"context"
	"errors"
	"net/http"
	"os"
	"regexp"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/thecomputer07/pesa/internal/http/problem"
	"github.com/thecomputer07/pesa/internal/service/imagecache"
	"github.com/thecomputer07/pesa/internal/service/off"
)

// Images serves cached third-party images with license metadata headers.
type Images struct {
	Store *imagecache.Store
}

// NewImages builds the handler bound to an imagecache.Store.
func NewImages(s *imagecache.Store) *Images { return &Images{Store: s} }

var imageFileRegex = regexp.MustCompile(`^[a-f0-9]{64}\.(jpg|png|webp|avif|gif)$`)

// Serve handles GET /static/images/{source}/{file}.
func (h *Images) Serve(w http.ResponseWriter, r *http.Request) {
	source := chi.URLParam(r, "source")
	file := chi.URLParam(r, "file")
	if _, ok := imagecache.AllowedSources[source]; !ok {
		problem.Write(w, r, http.StatusNotFound, "image-not-found", "unknown source")
		return
	}
	if !imageFileRegex.MatchString(file) {
		problem.Write(w, r, http.StatusNotFound, "image-not-found", "invalid file name")
		return
	}
	meta, err := h.Store.ReadMeta(source, file)
	if err != nil {
		problem.Write(w, r, http.StatusNotFound, "image-not-found", "no cached image")
		return
	}
	path, err := h.Store.Path(source, file)
	if err != nil {
		problem.Write(w, r, http.StatusNotFound, "image-not-found", "path rejected")
		return
	}
	f, err := os.Open(path)
	if err != nil {
		problem.Write(w, r, http.StatusNotFound, "image-not-found", "file missing")
		return
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil {
		problem.Write(w, r, http.StatusInternalServerError, "internal-error", "stat failed")
		return
	}

	w.Header().Set("Content-Type", meta.ContentType)
	w.Header().Set("Cache-Control", "public, max-age=2592000, immutable")
	w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
	if meta.LicenseURL != "" {
		w.Header().Set("Link", "<"+meta.LicenseURL+`>; rel="license"`)
	}
	if meta.Attribution != "" {
		// Keep attribution ASCII-only for safety in headers.
		w.Header().Set("X-Attribution", meta.Attribution)
	}
	etag := `"` + file + `"`
	w.Header().Set("ETag", etag)
	http.ServeContent(w, r, file, info.ModTime(), f)
}

// offImagePersister adapts imagecache.Store into the narrow interface
// required by off.Service so the OFF normalizer can rewrite image URLs.
type offImagePersister struct {
	Store   *imagecache.Store
	Fetcher *imagecache.Fetcher
}

// NewOFFImagePersister wires a Store to an off.ImagePersister.
func NewOFFImagePersister(s *imagecache.Store) off.ImagePersister {
	return &offImagePersister{Store: s, Fetcher: imagecache.NewFetcher()}
}

// PersistOFFImage downloads originalURL and stores it under source=off.
// ok=false means the caller should leave the DTO image untouched.
func (p *offImagePersister) PersistOFFImage(ctx context.Context, ean, originalURL string) (string, string, string, bool) {
	body, ct, err := p.Fetcher.Fetch(ctx, originalURL)
	if err != nil {
		return "", "", "", false
	}
	meta := imagecache.SidecarMeta{
		Source:      "off",
		License:     "ODbL-1.0",
		LicenseURL:  "https://opendatacommons.org/licenses/odbl/1-0/",
		Attribution: "Open Food Facts contributors — ODbL",
		SourceURL:   "https://world.openfoodfacts.org/product/" + ean,
		OriginalURL: originalURL,
	}
	sha, ext, err := p.Store.Put("off", body, ct, meta)
	if err != nil {
		if errors.Is(err, imagecache.ErrUnsupportedContentType) {
			return "", "", "", false
		}
		return "", "", "", false
	}
	return "/static/images/off/" + sha + ext, meta.License, meta.Attribution, true
}
