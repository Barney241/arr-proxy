package proxy

import (
	"context"
	"log/slog"
	"net/http"

	"arr-proxy/internal/config"
	"arr-proxy/internal/handlers/rest"
	"arr-proxy/internal/usecases"
)

// Run starts the proxy server and returns a shutdown function.
func Run(cfg *config.Config) (func(ctx context.Context), error) {
	proxyUseCase := usecases.NewProxyUseCase()
	proxyHandler := rest.NewProxyHandler(cfg, proxyUseCase)
	infoHandler := rest.NewInfoHandler(cfg)

	srv, err := rest.New(cfg, proxyHandler, infoHandler)
	if err != nil {
		return nil, err
	}

	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "error", err)
		}
	}()

	stop := func(ctx context.Context) {
		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("HTTP server shutdown error", "error", err)
		}
	}

	return stop, nil
}
