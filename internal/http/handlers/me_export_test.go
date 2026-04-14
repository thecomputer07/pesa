package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// me_export handlers are fully exercised by integration tests in a DB-backed
// suite; the unit-level test here verifies that unauthenticated callers are
// rejected with 401 (defense in depth for T-06-01).
func TestMeExportRequestsRejectUnauth(t *testing.T) {
	m := NewMeExport(nil, []byte("k"))
	for _, tc := range []struct {
		name string
		h    http.HandlerFunc
	}{
		{"request", m.Request},
		{"status", m.Status},
		{"download", m.Download},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			tc.h(rec, req)
			if rec.Code != http.StatusUnauthorized && rec.Code != http.StatusNotFound {
				t.Fatalf("want 401/404, got %d", rec.Code)
			}
		})
	}
}
