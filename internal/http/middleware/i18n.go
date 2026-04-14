package middleware

import (
	"context"
	"net/http"
	"strings"
)

// I18n parses Accept-Language and pins the locale to "it" (default) or "en".
func I18n(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loc := pickLocale(r.Header.Get("Accept-Language"))
		ctx := context.WithValue(r.Context(), ctxKeyLocale{}, loc)
		w.Header().Set("Content-Language", loc)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func pickLocale(acceptLang string) string {
	al := strings.ToLower(acceptLang)
	// Look for the first IT or EN tag preference.
	for _, part := range strings.Split(al, ",") {
		tag := strings.TrimSpace(strings.SplitN(part, ";", 2)[0])
		switch {
		case strings.HasPrefix(tag, "it"):
			return "it"
		case strings.HasPrefix(tag, "en"):
			return "en"
		}
	}
	return "it"
}
