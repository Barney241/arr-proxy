package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"regexp"
)

type contextKey string

const (
	requestIDKey   contextKey = "request_id"
	requestIDBytes            = 16
)

// validRequestID matches alphanumeric request IDs (hex, uuid-like, etc.)
var validRequestID = regexp.MustCompile(`^[a-zA-Z0-9\-_]{1,64}$`)

// RequestID middleware adds a unique request ID to each request.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for existing request ID from client
		requestID := r.Header.Get("X-Request-ID")

		// Validate or generate new ID
		if requestID == "" || !validRequestID.MatchString(requestID) {
			requestID = generateRequestID()
		}

		// Add to response header
		w.Header().Set("X-Request-ID", requestID)

		// Add to request context
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID retrieves the request ID from the context.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

func generateRequestID() string {
	b := make([]byte, requestIDBytes)
	if _, err := rand.Read(b); err != nil {
		slog.Warn("Failed to generate random request ID, using fallback", "error", err)
		return "fallback-request-id"
	}
	return hex.EncodeToString(b)
}
