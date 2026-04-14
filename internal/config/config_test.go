package config

import "testing"

func TestLogSaltRequiredInProd(t *testing.T) {
	t.Setenv("PESA_ENV", "prod")
	t.Setenv("DATABASE_DSN", "user:p@tcp(h:3306)/db?tls=true")
	t.Setenv("LOG_SALT", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTLSRequiredInProdDSN(t *testing.T) {
	t.Setenv("PESA_ENV", "prod")
	t.Setenv("DATABASE_DSN", "user:p@tcp(h:3306)/db?parseTime=true")
	t.Setenv("LOG_SALT", "deadbeef")
	if _, err := Load(); err == nil {
		t.Fatal("expected tls=true error, got nil")
	}
}

func TestDevDefaults(t *testing.T) {
	t.Setenv("PESA_ENV", "dev")
	t.Setenv("DATABASE_DSN", "x:y@tcp(h:3306)/d?parseTime=true")
	t.Setenv("LOG_SALT", "")
	c, err := Load()
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if c.Addr != ":8080" {
		t.Fatalf("want :8080, got %q", c.Addr)
	}
}
