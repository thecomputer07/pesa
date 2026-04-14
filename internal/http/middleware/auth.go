package middleware

import (
	"context"
	"net/http"

	"github.com/alexedwards/scs/v2"
)

// AuthContext copies the authenticated user_id (and optional email) from the
// scs session into the request context so downstream handlers + the logger can
// see it. Does NOT reject unauthenticated requests; route guards do that.
func AuthContext(sm *scs.SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			if uid := sm.GetInt64(ctx, "user_id"); uid != 0 {
				ctx = context.WithValue(ctx, ctxKeyUserID{}, uid)
			}
			if email := sm.GetString(ctx, "email"); email != "" {
				ctx = context.WithValue(ctx, ctxKeyEmail{}, email)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
