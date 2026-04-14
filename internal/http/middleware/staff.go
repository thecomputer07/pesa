package middleware

import (
	"database/sql"
	"net/http"

	"github.com/thecomputer07/pesa/internal/http/problem"
)

// RequireStaff guards a route group: only users with users.is_staff=TRUE
// proceed. Unauthenticated or non-staff callers receive 403 problem+json.
func RequireStaff(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			uid := UserIDFrom(r.Context())
			if uid == 0 {
				problem.Write(w, r, http.StatusForbidden, "forbidden", "staff only")
				return
			}
			var isStaff bool
			err := db.QueryRowContext(r.Context(),
				`SELECT is_staff FROM users WHERE id=$1`, uid).Scan(&isStaff)
			if err != nil || !isStaff {
				problem.Write(w, r, http.StatusForbidden, "forbidden", "staff only")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
