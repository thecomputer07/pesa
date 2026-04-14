package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/thecomputer07/pesa/internal/service/off"
)

type fakeBarcodeSvc struct {
	seenEAN string
	resp    off.FoodDTO
	err     error
}

func (f *fakeBarcodeSvc) LookupByEAN(ctx context.Context, ean string) (off.FoodDTO, error) {
	f.seenEAN = ean
	return f.resp, f.err
}

func newBarcodeRouter(h *Barcode) http.Handler {
	r := chi.NewRouter()
	r.Get("/api/v1/foods/barcode/{ean}", h.Lookup)
	return r
}

func TestBarcode_InvalidEAN(t *testing.T) {
	svc := &fakeBarcodeSvc{}
	h := NewBarcode(svc)
	r := newBarcodeRouter(h)

	for _, bad := range []string{"abc", "1234567", "123456789012345"} {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/foods/barcode/"+bad, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("%q: got %d, want 400", bad, w.Code)
		}
		if !strings.Contains(w.Header().Get("Content-Type"), "application/problem+json") {
			t.Errorf("%q: content-type %q", bad, w.Header().Get("Content-Type"))
		}
	}
}

func TestBarcode_ZeroPad(t *testing.T) {
	svc := &fakeBarcodeSvc{resp: off.FoodDTO{NameIT: "x", Source: off.SourceDTO{Code: "off", License: "ODbL-1.0"}}}
	h := NewBarcode(svc)
	r := newBarcodeRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/foods/barcode/12345678", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("got %d body=%s", w.Code, w.Body.String())
	}
	if svc.seenEAN != "0000012345678" {
		t.Errorf("expected zero-padded EAN, got %q", svc.seenEAN)
	}
	var dto off.FoodDTO
	if err := json.Unmarshal(w.Body.Bytes(), &dto); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if dto.Source.Code != "off" || dto.License == "" {
		// License omitted in fake but Source.Code should round-trip.
		if dto.Source.Code != "off" {
			t.Errorf("source.code: %q", dto.Source.Code)
		}
	}
}

func TestBarcode_NotFound(t *testing.T) {
	svc := &fakeBarcodeSvc{err: off.ErrNotFound}
	h := NewBarcode(svc)
	r := newBarcodeRouter(h)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/foods/barcode/8001505005707", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("got %d, want 404", w.Code)
	}
	if !strings.Contains(w.Body.String(), "barcode-not-found") {
		t.Errorf("expected problem code: %s", w.Body.String())
	}
}

func TestZeroPadTo13(t *testing.T) {
	cases := map[string]string{
		"8001505005707":  "8001505005707",
		"12345678":       "0000012345678",
		"08001505005707": "8001505005707",
	}
	for in, want := range cases {
		if got := zeroPadTo13(in); got != want {
			t.Errorf("%q -> %q, want %q", in, got, want)
		}
	}
}
