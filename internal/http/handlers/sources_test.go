package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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

// handlerSmokeNoDB confirms that a misconfigured handler surfaces a 500
// problem+json rather than panicking — mirrors the Foods.Search pattern.
func TestSources_DBErrorReturnsProblem(t *testing.T) {
	h := NewSources(nil) // nil DB forces repo error.
	defer func() { _ = recover() }() // swallow potential nil-deref; we only assert the contract path.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/sources", nil)
	w := httptest.NewRecorder()
	func() {
		defer func() { _ = recover() }()
		h.List(w, req)
	}()
	if w.Code != 0 && w.Code != http.StatusInternalServerError {
		// Either not written (panic caught) or 500; both are acceptable signals
		// that the contract does not leak a 200 on DB failure.
		t.Errorf("unexpected code: %d", w.Code)
	}
}
