package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/thecomputer07/pesa/internal/service/imagecache"
)

func newImagesRouter(h *Images) http.Handler {
	r := chi.NewRouter()
	r.Get("/static/images/{source}/{file}", h.Serve)
	return r
}

func TestImages_HappyPath(t *testing.T) {
	dir := t.TempDir()
	store, err := imagecache.NewStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	sha, ext, err := store.Put("off", []byte("fake-jpg"), "image/jpeg", imagecache.SidecarMeta{
		Source: "off", License: "ODbL-1.0",
		LicenseURL:  "https://opendatacommons.org/licenses/odbl/1-0/",
		Attribution: "OFF contributors",
	})
	if err != nil {
		t.Fatal(err)
	}
	r := newImagesRouter(NewImages(store))

	req := httptest.NewRequest(http.MethodGet, "/static/images/off/"+sha+ext, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("got %d, body=%s", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "image/jpeg") {
		t.Errorf("content-type: %q", ct)
	}
	if cc := w.Header().Get("Cache-Control"); !strings.Contains(cc, "immutable") {
		t.Errorf("cache-control: %q", cc)
	}
	if link := w.Header().Get("Link"); !strings.Contains(link, `rel="license"`) {
		t.Errorf("link: %q", link)
	}
}

func TestImages_BadSource(t *testing.T) {
	dir := t.TempDir()
	store, _ := imagecache.NewStore(dir)
	r := newImagesRouter(NewImages(store))
	req := httptest.NewRequest(http.MethodGet, "/static/images/evil/abc.jpg", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("got %d", w.Code)
	}
}

func TestImages_BadFileRegex(t *testing.T) {
	dir := t.TempDir()
	store, _ := imagecache.NewStore(dir)
	r := newImagesRouter(NewImages(store))
	req := httptest.NewRequest(http.MethodGet, "/static/images/off/../secret", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("got %d", w.Code)
	}
}

func TestImages_UnknownSHA(t *testing.T) {
	dir := t.TempDir()
	store, _ := imagecache.NewStore(dir)
	r := newImagesRouter(NewImages(store))
	sha := strings.Repeat("a", 64)
	req := httptest.NewRequest(http.MethodGet, "/static/images/off/"+sha+".jpg", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("got %d", w.Code)
	}
	if !strings.Contains(w.Header().Get("Content-Type"), "problem+json") {
		t.Errorf("expected problem+json")
	}
}
