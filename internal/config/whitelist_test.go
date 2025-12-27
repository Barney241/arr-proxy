package config

import (
	"testing"
)

func TestCompileWhitelist(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		method   string
		path     string
		want     bool
	}{
		// Basic patterns (all methods allowed)
		{
			name:     "simple pattern matches any method",
			patterns: []string{"^/api/v3/status$"},
			method:   "GET",
			path:     "/api/v3/status",
			want:     true,
		},
		{
			name:     "simple pattern matches POST",
			patterns: []string{"^/api/v3/status$"},
			method:   "POST",
			path:     "/api/v3/status",
			want:     true,
		},
		{
			name:     "simple pattern no match",
			patterns: []string{"^/api/v3/status$"},
			method:   "GET",
			path:     "/api/v3/other",
			want:     false,
		},

		// Method-specific patterns
		{
			name:     "GET only pattern matches GET",
			patterns: []string{"GET:^/api/v3/movie$"},
			method:   "GET",
			path:     "/api/v3/movie",
			want:     true,
		},
		{
			name:     "GET only pattern blocks POST",
			patterns: []string{"GET:^/api/v3/movie$"},
			method:   "POST",
			path:     "/api/v3/movie",
			want:     false,
		},
		{
			name:     "GET only pattern blocks DELETE",
			patterns: []string{"GET:^/api/v3/movie$"},
			method:   "DELETE",
			path:     "/api/v3/movie",
			want:     false,
		},

		// Multiple methods
		{
			name:     "GET,POST pattern matches GET",
			patterns: []string{"GET,POST:^/api/v3/movie$"},
			method:   "GET",
			path:     "/api/v3/movie",
			want:     true,
		},
		{
			name:     "GET,POST pattern matches POST",
			patterns: []string{"GET,POST:^/api/v3/movie$"},
			method:   "POST",
			path:     "/api/v3/movie",
			want:     true,
		},
		{
			name:     "GET,POST pattern blocks DELETE",
			patterns: []string{"GET,POST:^/api/v3/movie$"},
			method:   "DELETE",
			path:     "/api/v3/movie",
			want:     false,
		},

		// Multiple rules
		{
			name:     "multiple rules - first matches",
			patterns: []string{"GET:^/api/v3/movie$", "POST:^/api/v3/movie$"},
			method:   "GET",
			path:     "/api/v3/movie",
			want:     true,
		},
		{
			name:     "multiple rules - second matches",
			patterns: []string{"GET:^/api/v3/movie$", "POST:^/api/v3/movie$"},
			method:   "POST",
			path:     "/api/v3/movie",
			want:     true,
		},
		{
			name:     "multiple rules - none match method",
			patterns: []string{"GET:^/api/v3/movie$", "POST:^/api/v3/movie$"},
			method:   "DELETE",
			path:     "/api/v3/movie",
			want:     false,
		},

		// Regex with colon (should not be confused with method prefix)
		{
			name:     "regex with colon in pattern",
			patterns: []string{"^/api/v3/movie:\\d+$"},
			method:   "GET",
			path:     "/api/v3/movie:123",
			want:     true,
		},

		// Mixed rules
		{
			name:     "mixed - method-specific and open",
			patterns: []string{"GET:^/api/v3/readonly$", "^/api/v3/any$"},
			method:   "DELETE",
			path:     "/api/v3/any",
			want:     true,
		},
		{
			name:     "mixed - method-specific blocks on readonly",
			patterns: []string{"GET:^/api/v3/readonly$", "^/api/v3/any$"},
			method:   "DELETE",
			path:     "/api/v3/readonly",
			want:     false,
		},

		// All common methods
		{
			name:     "DELETE method",
			patterns: []string{"DELETE:^/api/v3/movie/\\d+$"},
			method:   "DELETE",
			path:     "/api/v3/movie/123",
			want:     true,
		},
		{
			name:     "PUT method",
			patterns: []string{"PUT:^/api/v3/movie/\\d+$"},
			method:   "PUT",
			path:     "/api/v3/movie/123",
			want:     true,
		},
		{
			name:     "PATCH method",
			patterns: []string{"PATCH:^/api/v3/movie/\\d+$"},
			method:   "PATCH",
			path:     "/api/v3/movie/123",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := compileWhitelist(tt.patterns)
			cfg := &ServiceConfig{CompiledWhitelist: rules}
			got := cfg.IsWhitelisted(tt.method, tt.path)
			if got != tt.want {
				t.Errorf("IsWhitelisted(%q, %q) = %v, want %v", tt.method, tt.path, got, tt.want)
			}
		})
	}
}

func TestIsValidMethodSpec(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"GET", true},
		{"POST", true},
		{"PUT", true},
		{"PATCH", true},
		{"DELETE", true},
		{"HEAD", true},
		{"OPTIONS", true},
		{"GET,POST", true},
		{"GET,POST,DELETE", true},
		{"GET, POST", true}, // with space
		{"get", false},      // lowercase
		{"INVALID", false},
		{"GET,INVALID", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isValidMethodSpec(tt.input)
			if got != tt.want {
				t.Errorf("isValidMethodSpec(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
