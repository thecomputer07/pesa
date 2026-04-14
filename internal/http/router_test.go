package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/thecomputer07/pesa/internal/http/health"
)

func TestHealthz(t *testing.T) {
	rr := httptest.NewRecorder()
	health.Liveness(rr, httptest.NewRequest("GET", "/healthz", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rr.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("want status=ok, got %v", body)
	}
}

func TestReadyzNilDBSkips(t *testing.T) {
	rr := httptest.NewRecorder()
	health.Readiness(nil)(rr, httptest.NewRequest("GET", "/readyz", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("want 200 (skipped), got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "ok") {
		t.Fatalf("want ok body, got %s", rr.Body.String())
	}
}
