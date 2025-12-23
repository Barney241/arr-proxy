package middleware

import "net/http"

// SecurityHeaders middleware adds security headers to all responses.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")

		// Enable XSS filter in older browsers
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Control referrer information
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Prevent caching of sensitive responses
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")

		next.ServeHTTP(w, r)
	})
}
