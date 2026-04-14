// Package config loads environment-driven configuration for the Pesa binary.
// Fails fast on missing/invalid values appropriate to the environment.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config is the top-level runtime configuration.
type Config struct {
	Env         string // "dev" | "prod"
	Addr        string
	DatabaseDSN string
	LogSalt     string

	// Cookies
	CookieDomain string

	// WebBaseURL is the public base URL of the web frontend used in emailed links
	// (e.g. "http://localhost:5173" in dev, "https://pesadia.com" in prod).
	WebBaseURL string

	// SMTP
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string

	// TLS / autocert
	EnableTLS         bool
	AutocertHosts     string
	AutocertEmail     string
	AutocertCacheDir  string

	// pprof gate
	EnablePprof bool
	PprofToken  string

	// Erasure log salt (read by the subject-rights tooling; bound here for surface visibility).
	ErasureHashSalt string

	// Image cache (Phase 5 — plan-02).
	ImageCacheDir      string
	ImageCacheMaxBytes int64
}

// Load reads env vars and validates.
func Load() (*Config, error) {
	c := &Config{
		Env:              getenv("PESA_ENV", "dev"),
		Addr:             getenv("PESA_ADDR", ":8080"),
		DatabaseDSN:      os.Getenv("DATABASE_DSN"),
		LogSalt:          os.Getenv("LOG_SALT"),
		CookieDomain:     os.Getenv("COOKIE_DOMAIN"),
		WebBaseURL:       os.Getenv("WEB_BASE_URL"),
		SMTPHost:         os.Getenv("SMTP_HOST"),
		SMTPUser:         os.Getenv("SMTP_USER"),
		SMTPPassword:     os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:         getenv("SMTP_FROM", "no-reply@example.invalid"),
		EnableTLS:        os.Getenv("PESA_ENABLE_TLS") == "1",
		AutocertHosts:    os.Getenv("AUTOCERT_HOSTS"),
		AutocertEmail:    os.Getenv("AUTOCERT_EMAIL"),
		AutocertCacheDir: getenv("AUTOCERT_CACHE_DIR", "/var/cache/pesa-autocert"),
		EnablePprof:      os.Getenv("PESA_ENABLE_PPROF") == "1",
		PprofToken:       os.Getenv("PPROF_TOKEN"),
		ErasureHashSalt:  os.Getenv("PESA_ERASURE_HASH_SALT"),
		ImageCacheDir:    getenv("PESA_IMAGE_CACHE_DIR", ".cache/images"),
	}
	if v := os.Getenv("PESA_IMAGE_CACHE_MAX_BYTES"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("PESA_IMAGE_CACHE_MAX_BYTES: %w", err)
		}
		c.ImageCacheMaxBytes = n
	} else {
		c.ImageCacheMaxBytes = 1 << 30 // 1 GiB default
	}
	if p := os.Getenv("SMTP_PORT"); p != "" {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("SMTP_PORT: %w", err)
		}
		c.SMTPPort = n
	} else {
		c.SMTPPort = 587
	}

	if c.DatabaseDSN == "" {
		return nil, errors.New("DATABASE_DSN is required")
	}

	if c.Env == "prod" {
		if c.LogSalt == "" {
			return nil, errors.New("LOG_SALT is required in prod")
		}
		if !strings.Contains(c.DatabaseDSN, "sslmode=") || strings.Contains(c.DatabaseDSN, "sslmode=disable") {
			return nil, errors.New("DATABASE_DSN must include sslmode=require (or stricter) in prod")
		}
		if c.EnableTLS {
			if c.AutocertHosts == "" || c.AutocertEmail == "" {
				return nil, errors.New("AUTOCERT_HOSTS and AUTOCERT_EMAIL required when PESA_ENABLE_TLS=1")
			}
		}
	}

	return c, nil
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
