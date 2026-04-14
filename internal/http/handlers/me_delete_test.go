package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMeDeleteRejectsUnauth(t *testing.T) {
	h := NewMeDelete(nil, nil, nil, nil, false)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/me", nil)
	rec := httptest.NewRecorder()
	h.Handle(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
}

func TestMeDeleteRejectsBadJSON(t *testing.T) {
	// Even unauth returns 401 before JSON parse — so add a dummy body is
	// still 401. Regression guard against silent body-size explosion.
	h := NewMeDelete(nil, nil, nil, nil, false)
	body, _ := json.Marshal(map[string]any{"current_password": "x", "confirmation_text": "nope"})
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/me", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Handle(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
}
