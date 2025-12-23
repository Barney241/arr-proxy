package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"arr-proxy/cmd/proxy"
	"arr-proxy/internal/config"
)

func setupLogger(level string) {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	})
	slog.SetDefault(slog.New(handler))
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Configuration error", "error", err)
		os.Exit(1)
	}
	setupLogger(cfg.Server.LogLevel)

	stop, err := proxy.Run(&cfg)
	if err != nil {
		slog.Error("Application startup error", "error", err)
		os.Exit(1)
	}

	// Wait for interrupt signal to gracefully shut down the server
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	<-ctx.Done()
	slog.Info("Shutting down server...")

	// Give the server 30 seconds to finish current requests
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	stop(shutdownCtx)
}
