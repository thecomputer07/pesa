// Package middleware bundles the standard HTTP middleware chain.
package middleware

import (
	"context"

	"golang.org/x/time/rate"

	"github.com/thecomputer07/pesa/internal/service"
)

// Context-key types — exported names kept unexported to prevent collisions across packages.
type ctxKeyReqID struct{}
type ctxKeyEmail struct{}
type ctxKeyUserID struct{}
type ctxKeyLocale struct{}

// ReqIDFrom returns the request id from context (empty if absent).
func ReqIDFrom(ctx context.Context) string {
	v, _ := ctx.Value(ctxKeyReqID{}).(string)
	return v
}

// UserIDFrom returns the authenticated user id from context (0 if absent).
func UserIDFrom(ctx context.Context) int64 {
	v, _ := ctx.Value(ctxKeyUserID{}).(int64)
	return v
}

// EmailFrom returns the authenticated email (empty if absent).
func EmailFrom(ctx context.Context) string {
	v, _ := ctx.Value(ctxKeyEmail{}).(string)
	return v
}

// LocaleFrom returns the negotiated locale ("it" or "en"; defaults to "it").
func LocaleFrom(ctx context.Context) string {
	v, _ := ctx.Value(ctxKeyLocale{}).(string)
	if v == "" {
		return "it"
	}
	return v
}

// globalLimiter is the singleton SEC-17 limiter: per IP, burst 100, refill one
// token every 6 ms (~10k req/min). Generous on purpose: assets and photos go
// through the same router.
var globalLimiter = service.NewKeyedLimiter(rate.Every(600*1e6/100), 100, 30*60*1e9) //nolint:gochecknoglobals
