// Package config provides the configuration structures for the application.
package config

import (
	"net/url"
	"regexp"
)

const (
	AuthModeMTLS   = "mtls"
	AuthModeBasic  = "basic"
	AuthModeAPIKey = "apikey"
)

// ServiceConfig holds the configuration for a single service (like Sonarr or Radarr).
type ServiceConfig struct {
	URL               string   `yaml:"url" mapstructure:"url"`
	APIKey            string   `yaml:"api_key" mapstructure:"api_key"`
	Whitelist         []string `yaml:"whitelist" mapstructure:"whitelist"`
	CompiledWhitelist []*regexp.Regexp
	ParsedURL         *url.URL
}

type BasicAuthConfig struct {
	User     string
	Password string
}

type AuthConfig struct {
	Mode      string
	BasicAuth BasicAuthConfig
	APIKey    string
}

// Config holds the application configuration.
type Config struct {
	Sonarr  *ServiceConfig // nil if not configured
	Radarr  *ServiceConfig // nil if not configured
	TLSCert string
	TLSKey  string
	CACert  string
	Port    string
	Auth    AuthConfig
	Server  ServerConfig
}

// IsWhitelisted checks if a given path is whitelisted for the service.
func (sc *ServiceConfig) IsWhitelisted(path string) bool {
	for _, re := range sc.CompiledWhitelist {
		if re.MatchString(path) {
			return true
		}
	}
	return false
}

// TLSEnabled returns true if TLS certificates are configured.
func (c *Config) TLSEnabled() bool {
	return c.TLSCert != "" && c.TLSKey != ""
}
