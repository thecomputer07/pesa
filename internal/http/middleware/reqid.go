package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// ReqID injects an X-Request-ID into context and echoes it on the response.
func ReqID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			u, err := uuid.NewV7()
			if err == nil {
				id = u.String()
			} else {
				id = uuid.NewString()
			}
		}
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), ctxKeyReqID{}, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
