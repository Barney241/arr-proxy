package middleware

import (
	"crypto/subtle"
	"log/slog"
	"net/http"
)

// BasicAuth middleware enforces HTTP Basic Authentication
func BasicAuth(user, pass string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, p, ok := r.BasicAuth()
			if !ok || subtle.ConstantTimeCompare([]byte(u), []byte(user)) != 1 || subtle.ConstantTimeCompare([]byte(p), []byte(pass)) != 1 {
				slog.Warn("Authentication failed", "method", "basic", "path", r.URL.Path, "remote_addr", r.RemoteAddr)
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// APIKeyAuth middleware enforces API Key Authentication via header or query param
func APIKeyAuth(keyValue string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check Header first
			key := r.Header.Get("X-Api-Key")
			fromQueryParam := false

			if key == "" {
				// Check Query Param (less secure - warn about usage)
				key = r.URL.Query().Get("apikey")
				if key != "" {
					fromQueryParam = true
					slog.Warn("API key provided via query parameter (less secure)", "path", r.URL.Path, "remote_addr", r.RemoteAddr)
				}
			}

			// Reject empty keys to prevent bypass when keyValue is also empty
			if key == "" {
				slog.Warn("Authentication failed", "method", "apikey", "reason", "no key provided", "path", r.URL.Path, "remote_addr", r.RemoteAddr)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if subtle.ConstantTimeCompare([]byte(key), []byte(keyValue)) != 1 {
				slog.Warn("Authentication failed", "method", "apikey", "reason", "invalid key", "path", r.URL.Path, "remote_addr", r.RemoteAddr, "via_query_param", fromQueryParam)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
