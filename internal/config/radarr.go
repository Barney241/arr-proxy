package config

import (
	"log/slog"
	"net/url"

	"github.com/spf13/viper"
)

// LoadRadarrConfig loads Radarr configuration from file and environment.
// Returns nil if no configuration is found.
func LoadRadarrConfig(configPaths ...string) *ServiceConfig {
	v := viper.New()
	v.SetConfigName("radarr")
	v.SetConfigType("yaml")
	for _, p := range configPaths {
		v.AddConfigPath(p)
	}

	// Bind environment variables
	_ = v.BindEnv("url", "RADARR_URL")
	_ = v.BindEnv("api_key", "RADARR_API_KEY")

	// Try to read config file (optional)
	_ = v.ReadInConfig()

	urlStr := v.GetString("url")
	if urlStr == "" {
		return nil // Service not configured
	}

	// Validate URL is parseable
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		slog.Error("Invalid Radarr URL", "url", urlStr, "error", err)
		return nil
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		slog.Error("Invalid Radarr URL scheme (must be http or https)", "url", urlStr, "scheme", parsedURL.Scheme)
		return nil
	}
	if parsedURL.Host == "" {
		slog.Error("Invalid Radarr URL: missing host", "url", urlStr)
		return nil
	}

	whitelist := v.GetStringSlice("whitelist")
	compiledWhitelist := compileWhitelist(whitelist)
	if len(whitelist) > 0 && len(compiledWhitelist) == 0 {
		slog.Error("All Radarr whitelist patterns failed to compile")
		return nil
	}

	cfg := &ServiceConfig{
		URL:               urlStr,
		APIKey:            v.GetString("api_key"),
		Whitelist:         whitelist,
		CompiledWhitelist: compiledWhitelist,
		ParsedURL:         parsedURL,
	}

	return cfg
}
