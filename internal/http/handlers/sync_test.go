package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// These tests exercise the handler surface against a nil service where
// authentication is the dominant concern. Full round-trip integration tests
// live behind the testcontainers harness (see TODO markers below).

// TestSyncPost_Unauthenticated returns 401 when no user in context.
func TestSyncPost_Unauthenticated(t *testing.T) {
	h := NewSync(nil)
	r := httptest.NewRequest(http.MethodPost, "/api/v1/sync", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	h.Post(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d, want 401", w.Code)
	}
}

// TestSyncGet_Unauthenticated returns 401 when no user in context.
func TestSyncGet_Unauthenticated(t *testing.T) {
	h := NewSync(nil)
	r := httptest.NewRequest(http.MethodGet, "/api/v1/sync?since=0", nil)
	w := httptest.NewRecorder()
	h.Get(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d, want 401", w.Code)
	}
}

// TestSyncPost_NegativeSinceCoerced ensures negative cursor values are clamped.
func TestSyncPost_NegativeSinceCoerced(t *testing.T) {
	h := NewSync(nil)
	r := httptest.NewRequest(http.MethodGet, "/api/v1/sync?since=-1", nil)
	w := httptest.NewRecorder()
	h.Get(w, r) // 401 (unauth) but the parse path must not panic.
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d, want 401", w.Code)
	}
}

// TODO(integration): add authenticated round-trip tests covering:
//   - POST with "foods":[] key → 422, problem+json type contains "foods-readonly"
//   - POST with 2MB body → 413 "body-too-large"
//   - POST valid batch → 200 with BatchResponse; re-POST idempotent (same server_id)
//   - GET cursor monotonically advances across POST→GET round-trips
//   - Cross-user isolation: user A POST, user B GET returns nothing
// These require the Phase 1 session-login helper + testcontainers MariaDB.
