package export

import (
	"testing"
	"time"
)

func TestSignVerifyRoundTrip(t *testing.T) {
	secret := []byte("unit-test-secret")
	exp := time.Now().Add(1 * time.Hour)
	tok := Sign(42, exp, secret)
	if len(tok) != 43 {
		t.Fatalf("token length want 43 got %d", len(tok))
	}
	if !TokenIsWellFormed(tok) {
		t.Fatalf("well-formed check rejected valid token")
	}
	if err := Verify(tok, 42, exp, secret); err != nil {
		t.Fatalf("verify legit token: %v", err)
	}
}

func TestVerifyRejectsTamper(t *testing.T) {
	secret := []byte("unit-test-secret")
	exp := time.Now().Add(1 * time.Hour)
	tok := Sign(42, exp, secret)
	// flip one char
	tampered := "A" + tok[1:]
	if tampered == tok {
		tampered = "B" + tok[1:]
	}
	if err := Verify(tampered, 42, exp, secret); err == nil {
		t.Fatal("verify accepted tampered token")
	}
	// wrong jobID
	if err := Verify(tok, 99, exp, secret); err == nil {
		t.Fatal("verify accepted wrong jobID")
	}
	// expired
	if err := Verify(tok, 42, time.Now().Add(-1*time.Hour), secret); err == nil {
		t.Fatal("verify accepted expired token")
	}
}

func TestTokenWellFormedRejectsBadInput(t *testing.T) {
	cases := []string{"", "short", "has space in ittttttttttttttttttttttttttttt", "!!!"}
	for _, c := range cases {
		if TokenIsWellFormed(c) {
			t.Fatalf("accepted malformed token: %q", c)
		}
	}
}
