// Package config provides functions to load application configuration.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/viper"
)

func compileWhitelist(patterns []string) []*regexp.Regexp {
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			slog.Warn("Failed to compile whitelist regex, skipping", "pattern", p, "error", err)
			continue
		}
		compiled = append(compiled, re)
	}
	return compiled
}

// Load loads the entire application configuration.
// Returns an error if configuration is invalid.
func Load() (Config, error) {
	configDir := os.Getenv("APP_CONFIG_DIR")
	if configDir == "" {
		configDir = "config"
	}

	appViper := viper.New()
	appViper.SetDefault("port", "8443")
	appViper.SetEnvPrefix("APP")
	appViper.AutomaticEnv()

	_ = appViper.BindEnv("tls_cert", "APP_TLS_CERT")
	_ = appViper.BindEnv("tls_key", "APP_TLS_KEY")
	_ = appViper.BindEnv("ca_cert", "APP_CA_CERT")
	_ = appViper.BindEnv("port", "APP_PORT")
	_ = appViper.BindEnv("auth_mode", "APP_AUTH_MODE")
	_ = appViper.BindEnv("basic_auth_user", "APP_BASIC_AUTH_USER")
	_ = appViper.BindEnv("basic_auth_pass", "APP_BASIC_AUTH_PASS")
	_ = appViper.BindEnv("api_key", "APP_API_KEY")

	authMode := appViper.GetString("auth_mode")
	if authMode == "" {
		authMode = AuthModeAPIKey
	}

	cfg := Config{
		Sonarr:  LoadSonarrConfig(configDir),
		Radarr:  LoadRadarrConfig(configDir),
		TLSCert: appViper.GetString("tls_cert"),
		TLSKey:  appViper.GetString("tls_key"),
		CACert:  appViper.GetString("ca_cert"),
		Port:    appViper.GetString("port"),
		Auth: AuthConfig{
			Mode: authMode,
			BasicAuth: BasicAuthConfig{
				User:     appViper.GetString("basic_auth_user"),
				Password: appViper.GetString("basic_auth_pass"),
			},
			APIKey: appViper.GetString("api_key"),
		},
		Server: LoadServerConfig(configDir),
	}

	// Validate configuration - collect all errors before failing
	var configErrors []string

	// At least one service must be configured
	if cfg.Sonarr == nil && cfg.Radarr == nil {
		configErrors = append(configErrors, "at least one service must be configured (set SONARR_URL or RADARR_URL)")
	}

	switch cfg.Auth.Mode {
	case AuthModeBasic:
		if cfg.Auth.BasicAuth.User == "" {
			configErrors = append(configErrors, "APP_BASIC_AUTH_USER required for basic auth mode")
		}
		if cfg.Auth.BasicAuth.Password == "" {
			configErrors = append(configErrors, "APP_BASIC_AUTH_PASS required for basic auth mode")
		}
	case AuthModeAPIKey:
		if cfg.Auth.APIKey == "" {
			configErrors = append(configErrors, "APP_API_KEY required for apikey auth mode")
		}
	case AuthModeMTLS:
		// mTLS requires all TLS certificates
		if cfg.TLSCert == "" {
			configErrors = append(configErrors, "APP_TLS_CERT required for mTLS mode")
		}
		if cfg.TLSKey == "" {
			configErrors = append(configErrors, "APP_TLS_KEY required for mTLS mode")
		}
		if cfg.CACert == "" {
			configErrors = append(configErrors, "APP_CA_CERT required for mTLS mode (CA to verify client certs)")
		}
	default:
		configErrors = append(configErrors, fmt.Sprintf("invalid APP_AUTH_MODE '%s' (valid modes: apikey, mtls, basic)", cfg.Auth.Mode))
	}

	// If TLS cert is provided, key must also be provided
	if (cfg.TLSCert != "" && cfg.TLSKey == "") || (cfg.TLSCert == "" && cfg.TLSKey != "") {
		configErrors = append(configErrors, "APP_TLS_CERT and APP_TLS_KEY must both be set for HTTPS")
	}

	if len(configErrors) > 0 {
		return Config{}, errors.New("configuration validation failed: " + strings.Join(configErrors, "; "))
	}

	// Log configured services
	if cfg.Sonarr != nil {
		slog.Info("Sonarr configured", "url", cfg.Sonarr.URL)
	}
	if cfg.Radarr != nil {
		slog.Info("Radarr configured", "url", cfg.Radarr.URL)
	}

	return cfg, nil
}
