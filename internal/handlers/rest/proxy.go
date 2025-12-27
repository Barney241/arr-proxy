// Package rest contains the HTTP handlers for the application.
package rest

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"arr-proxy/internal/config"
	"arr-proxy/internal/usecases"
)

// ProxyHandler is the handler for proxying requests.
type ProxyHandler struct {
	config       *config.Config
	proxyUseCase *usecases.ProxyUseCase
}

// NewProxyHandler creates a new ProxyHandler.
func NewProxyHandler(cfg *config.Config, proxyUseCase *usecases.ProxyUseCase) *ProxyHandler {
	return &ProxyHandler{
		config:       cfg,
		proxyUseCase: proxyUseCase,
	}
}

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// ServeHTTP is the entry point for the handler.
func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	clientCN := "unknown"
	if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
		clientCN = r.TLS.PeerCertificates[0].Subject.CommonName
	}

	var serviceConfig *config.ServiceConfig

	if strings.HasPrefix(r.URL.Path, "/sonarr") {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/sonarr")
		if r.URL.Path == "" {
			r.URL.Path = "/"
		}
		serviceConfig = h.config.Sonarr
	} else if strings.HasPrefix(r.URL.Path, "/radarr") {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/radarr")
		if r.URL.Path == "" {
			r.URL.Path = "/"
		}
		serviceConfig = h.config.Radarr
	} else {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	}

	// Check if the service is configured
	if serviceConfig == nil {
		http.Error(w, "503 Service Not Configured", http.StatusServiceUnavailable)
		return
	}

	if !serviceConfig.IsWhitelisted(r.Method, r.URL.Path) {
		slog.Warn("Request blocked", "method", r.Method, "path", r.URL.Path, "client_cn", clientCN, "reason", "method/endpoint not whitelisted", "status", 403)
		http.Error(w, "403 Forbidden", http.StatusForbidden)
		return
	}

	// Enforce body size limit on all requests (protection against DoS)
	maxBodySize := h.config.Server.MaxBodySize
	if r.ContentLength > maxBodySize {
		slog.Warn("Request blocked", "method", r.Method, "path", r.URL.Path, "client_cn", clientCN, "reason", "payload too large", "status", 413)
		http.Error(w, "413 Payload Too Large", http.StatusRequestEntityTooLarge)
		return
	}

	// Validate JSON payload for methods with request bodies
	if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
		bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, maxBodySize+1))
		if err != nil {
			slog.Warn("Request blocked", "method", r.Method, "path", r.URL.Path, "client_cn", clientCN, "reason", "failed to read payload", "status", 400)
			http.Error(w, "400 Bad Request", http.StatusBadRequest)
			return
		}
		_ = r.Body.Close()

		// Check if body exceeded max size
		if int64(len(bodyBytes)) > maxBodySize {
			slog.Warn("Request blocked", "method", r.Method, "path", r.URL.Path, "client_cn", clientCN, "reason", "payload too large", "status", 413)
			http.Error(w, "413 Payload Too Large", http.StatusRequestEntityTooLarge)
			return
		}

		// Validate JSON if content-type indicates JSON
		contentType := r.Header.Get("Content-Type")
		if strings.Contains(contentType, "application/json") && len(bodyBytes) > 0 {
			var payload map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &payload); err != nil {
				slog.Warn("Request blocked", "method", r.Method, "path", r.URL.Path, "client_cn", clientCN, "reason", "invalid JSON payload", "status", 400)
				http.Error(w, "400 Bad Request", http.StatusBadRequest)
				return
			}
		}
		// Re-encode body so it can be read again by the proxy
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		r.ContentLength = int64(len(bodyBytes))
	}

	sw := &statusResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
	h.proxyUseCase.ServeHTTP(sw, r, serviceConfig.ParsedURL, serviceConfig.APIKey)
	latency := time.Since(start)
	slog.Info("Request completed", "method", r.Method, "path", r.URL.Path, "client_cn", clientCN, "status", sw.statusCode, "latency", latency)
}
