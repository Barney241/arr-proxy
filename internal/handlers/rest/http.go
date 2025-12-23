package rest

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"arr-proxy/internal/config"
	"arr-proxy/internal/middleware"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

// Server is the main server struct.
type Server struct {
	config *config.Config
	server *http.Server
}

// New creates a new server.
func New(cfg *config.Config, proxyHandler *ProxyHandler, infoHandler *InfoHandler) (*Server, error) {
	r := chi.NewRouter()

	// Core middleware (order matters)
	r.Use(middleware.RequestID)
	r.Use(middleware.SecurityHeaders)
	r.Use(chiMiddleware.Logger)

	// Apply authentication middleware based on mode
	switch cfg.Auth.Mode {
	case config.AuthModeBasic:
		r.Use(middleware.BasicAuth(cfg.Auth.BasicAuth.User, cfg.Auth.BasicAuth.Password))
	case config.AuthModeAPIKey:
		r.Use(middleware.APIKeyAuth(cfg.Auth.APIKey))
	case config.AuthModeMTLS:
		// mTLS auth is handled by TLS client cert verification
	default:
		return nil, fmt.Errorf("unknown auth mode: %s", cfg.Auth.Mode)
	}

	r.Get("/info", infoHandler.ServeHTTP)
	r.Handle("/*", proxyHandler)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
		ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
	}

	// Configure TLS if certificates are provided
	if cfg.TLSEnabled() {
		// Set TLS minimum version from config
		var minVersion uint16 = tls.VersionTLS12
		if cfg.Server.TLSMinVersion == "1.3" {
			minVersion = tls.VersionTLS13
		}

		tlsConfig := &tls.Config{
			MinVersion: minVersion,
		}

		// Load CA cert for client verification if provided
		if cfg.CACert != "" {
			caCert, err := os.ReadFile(cfg.CACert)
			if err != nil {
				return nil, fmt.Errorf("failed to read CA cert: %w", err)
			}
			caCertPool := x509.NewCertPool()
			if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
				return nil, fmt.Errorf("failed to parse CA certificate")
			}
			tlsConfig.ClientCAs = caCertPool
		}

		// Require client certs only for mTLS mode
		if cfg.Auth.Mode == config.AuthModeMTLS {
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		} else {
			tlsConfig.ClientAuth = tls.NoClientCert
		}

		server.TLSConfig = tlsConfig
	}

	return &Server{
		config: cfg,
		server: server,
	}, nil
}

// Start starts the server.
func (s *Server) Start() error {
	if s.config.TLSEnabled() {
		slog.Info("Starting HTTPS server", "port", s.config.Port)
		return s.server.ListenAndServeTLS(s.config.TLSCert, s.config.TLSKey)
	}
	slog.Info("Starting HTTP server", "port", s.config.Port)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// ErrServerClosed is an alias for net/http.ErrServerClosed
var ErrServerClosed = http.ErrServerClosed
