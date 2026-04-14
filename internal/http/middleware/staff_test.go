package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// RequireStaff without a DB + unauthenticated caller must 403.
func TestRequireStaffRejectsUnauth(t *testing.T) {
	h := RequireStaff(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/admin/anything", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", rec.Code)
	}
}
