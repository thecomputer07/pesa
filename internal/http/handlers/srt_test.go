package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

// Only static behaviours covered at the unit level: bad JSON, bad id. The DB
// path is exercised by the integration suite.
func TestSRTBadID(t *testing.T) {
	h := NewSRT(nil, nil, nil)
	r := chi.NewRouter()
	r.Patch("/admin/subject-rights/tickets/{id}", h.Patch)
	req := httptest.NewRequest(http.MethodPatch, "/admin/subject-rights/tickets/abc",
		bytes.NewReader([]byte(`{}`)))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestSRTCreateBadJSON(t *testing.T) {
	h := NewSRT(nil, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/admin/subject-rights/tickets",
		bytes.NewReader([]byte(`not-json`)))
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", rec.Code)
	}
}

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	writeJSON(rec, http.StatusAccepted, map[string]string{"k": "v"})
	if rec.Code != http.StatusAccepted {
		t.Fatalf("code: %d", rec.Code)
	}
	var out map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out["k"] != "v" {
		t.Fatalf("roundtrip: %v", out)
	}
}
