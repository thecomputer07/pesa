package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Structural test: encoding/decoding of the DTO survives a round-trip.
// Full DB integration is exercised by higher-level tests (phase 2 harness).
func TestSourcesDTO_RoundTrip(t *testing.T) {
	lic := "https://opendatacommons.org/licenses/odbl/1-0/"
	sc := "data+images"
	in := sourcesResponseDTO{Sources: []sourceDTO{
		{Code: "off", Name: "Open Food Facts", License: "ODbL-1.0",
			LicenseURL: &lic, Attribution: "OFF contributors", Scope: &sc},
	}}
	body, _ := json.Marshal(in)
	var out sourcesResponseDTO
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Sources) != 1 || out.Sources[0].Code != "off" {
		t.Fatalf("round-trip failed: %+v", out)
	}
	if out.Sources[0].LicenseURL == nil || *out.Sources[0].LicenseURL != lic {
		t.Errorf("license url lost")
	}
}

// A DB failure must surface as a 500 problem+json, never a 200. A closed
// *sql.DB fails every query immediately, without touching the network.
func TestSources_DBErrorReturnsProblem(t *testing.T) {
	db, err := sql.Open("pgx", "postgres://pesa@127.0.0.1:1/pesa")
	if err != nil {
		t.Fatal(err)
	}
	_ = db.Close()
	h := NewSources(db)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/sources", nil)
	w := httptest.NewRecorder()
	h.List(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("code = %d, want 500", w.Code)
	}
}
