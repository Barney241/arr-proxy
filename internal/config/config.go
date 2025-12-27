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

// WhitelistRule represents a single whitelist entry with optional method restrictions.
// Format: "METHOD1,METHOD2:pattern" or just "pattern" (allows all methods)
type WhitelistRule struct {
	Methods map[string]bool // nil means all methods allowed
	Pattern *regexp.Regexp
}

// Matches checks if the rule matches the given method and path.
func (r *WhitelistRule) Matches(method, path string) bool {
	if !r.Pattern.MatchString(path) {
		return false
	}
	// If no methods specified, allow all
	if r.Methods == nil {
		return true
	}
	return r.Methods[method]
}

// ServiceConfig holds the configuration for a single service (like Sonarr or Radarr).
type ServiceConfig struct {
	URL               string   `yaml:"url" mapstructure:"url"`
	APIKey            string   `yaml:"api_key" mapstructure:"api_key"`
	Whitelist         []string `yaml:"whitelist" mapstructure:"whitelist"`
	CompiledWhitelist []WhitelistRule
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

// IsWhitelisted checks if a given method and path combination is whitelisted for the service.
func (sc *ServiceConfig) IsWhitelisted(method, path string) bool {
	for _, rule := range sc.CompiledWhitelist {
		if rule.Matches(method, path) {
			return true
		}
	}
	return false
}

// TLSEnabled returns true if TLS certificates are configured.
func (c *Config) TLSEnabled() bool {
	return c.TLSCert != "" && c.TLSKey != ""
}
