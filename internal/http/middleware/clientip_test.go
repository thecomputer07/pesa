package middleware

import (
	"net/http/httptest"
	"testing"
)

func TestClientIPUsesRealIPNotForwardedFor(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "127.0.0.1:5555"
	r.Header.Set("X-Forwarded-For", "6.6.6.6, 1.2.3.4")
	if got := clientIP(r); got != "127.0.0.1" {
		t.Fatalf("X-Forwarded-For must be ignored, got %q", got)
	}
	r.Header.Set("X-Real-IP", "1.2.3.4")
	if got := clientIP(r); got != "1.2.3.4" {
		t.Fatalf("want X-Real-IP, got %q", got)
	}
}
