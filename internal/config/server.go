package config

import (
	"log/slog"
	"time"

	"github.com/spf13/viper"
)

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	MaxBodySize       int64
	TLSMinVersion     string
	LogLevel          string
}

// LoadServerConfig loads server configuration from server.yaml with env overrides
func LoadServerConfig(configPaths ...string) ServerConfig {
	v := viper.New()
	v.SetConfigName("server")
	v.SetConfigType("yaml")

	// Add config paths
	for _, path := range configPaths {
		v.AddConfigPath(path)
	}
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	// Set defaults
	v.SetDefault("read_timeout", "30s")
	v.SetDefault("write_timeout", "30s")
	v.SetDefault("idle_timeout", "120s")
	v.SetDefault("read_header_timeout", "20s")
	v.SetDefault("max_body_size", 10*1024*1024) // 10MB
	v.SetDefault("tls_min_version", "1.2")
	v.SetDefault("log_level", "info")

	// Bind environment variables
	v.SetEnvPrefix("APP")
	_ = v.BindEnv("read_timeout", "APP_READ_TIMEOUT")
	_ = v.BindEnv("write_timeout", "APP_WRITE_TIMEOUT")
	_ = v.BindEnv("idle_timeout", "APP_IDLE_TIMEOUT")
	_ = v.BindEnv("read_header_timeout", "APP_READ_HEADER_TIMEOUT")
	_ = v.BindEnv("max_body_size", "APP_MAX_BODY_SIZE")
	_ = v.BindEnv("tls_min_version", "APP_TLS_MIN_VERSION")
	_ = v.BindEnv("log_level", "APP_LOG_LEVEL")

	// Read config file (optional - won't fail if missing)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			slog.Warn("Failed to read server config file", "error", err)
		}
	}

	readTimeout, err := time.ParseDuration(v.GetString("read_timeout"))
	if err != nil {
		slog.Warn("Invalid read_timeout, using default", "value", v.GetString("read_timeout"), "error", err)
		readTimeout = 30 * time.Second
	}

	writeTimeout, err := time.ParseDuration(v.GetString("write_timeout"))
	if err != nil {
		slog.Warn("Invalid write_timeout, using default", "value", v.GetString("write_timeout"), "error", err)
		writeTimeout = 30 * time.Second
	}

	idleTimeout, err := time.ParseDuration(v.GetString("idle_timeout"))
	if err != nil {
		slog.Warn("Invalid idle_timeout, using default", "value", v.GetString("idle_timeout"), "error", err)
		idleTimeout = 120 * time.Second
	}

	readHeaderTimeout, err := time.ParseDuration(v.GetString("read_header_timeout"))
	if err != nil {
		slog.Warn("Invalid read_header_timeout, using default", "value", v.GetString("read_header_timeout"), "error", err)
		readHeaderTimeout = 20 * time.Second
	}

	tlsMinVersion := v.GetString("tls_min_version")
	if tlsMinVersion != "1.2" && tlsMinVersion != "1.3" {
		slog.Warn("Invalid tls_min_version, using default 1.2", "value", tlsMinVersion)
		tlsMinVersion = "1.2"
	}

	maxBodySize := v.GetInt64("max_body_size")
	if maxBodySize <= 0 {
		slog.Warn("Invalid max_body_size, using default 10MB", "value", maxBodySize)
		maxBodySize = 10 * 1024 * 1024
	}

	logLevel := v.GetString("log_level")
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[logLevel] {
		slog.Warn("Invalid log_level, using default 'info'", "value", logLevel)
		logLevel = "info"
	}

	return ServerConfig{
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		MaxBodySize:       maxBodySize,
		TLSMinVersion:     tlsMinVersion,
		LogLevel:          logLevel,
	}
}
