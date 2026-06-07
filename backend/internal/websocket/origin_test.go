package websocket

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"git.subcult.tv/subculture-collective/clpr/config"
)

func TestIsOriginAllowed(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		allowedOrigins []string
		expected       bool
	}{
		{
			name:           "exact match - localhost",
			origin:         "http://localhost:5173",
			allowedOrigins: []string{"http://localhost:5173", "http://localhost:3000"},
			expected:       true,
		},
		{
			name:           "exact match - production",
			origin:         "https://clpr.tv",
			allowedOrigins: []string{"https://clpr.tv", "https://www.clpr.tv"},
			expected:       true,
		},
		{
			name:           "no match",
			origin:         "https://evil.com",
			allowedOrigins: []string{"https://clpr.tv"},
			expected:       false,
		},
		{
			name:           "empty origin",
			origin:         "",
			allowedOrigins: []string{"https://clpr.tv"},
			expected:       false,
		},
		{
			name:           "wildcard subdomain match",
			origin:         "https://staging.clpr.tv",
			allowedOrigins: []string{"*.clpr.tv"},
			expected:       true,
		},
		{
			name:           "wildcard subdomain - multiple levels",
			origin:         "https://api.staging.clpr.tv",
			allowedOrigins: []string{"*.clpr.tv"},
			expected:       true,
		},
		{
			name:           "wildcard - base domain match",
			origin:         "https://clpr.tv",
			allowedOrigins: []string{"*.clpr.tv"},
			expected:       true,
		},
		{
			name:           "wildcard - no match different domain",
			origin:         "https://clpr.evil.com",
			allowedOrigins: []string{"*.clpr.tv"},
			expected:       false,
		},
		{
			name:           "wildcard - partial domain match should fail",
			origin:         "https://fakeclpr.tv",
			allowedOrigins: []string{"*.clpr.tv"},
			expected:       false,
		},
		{
			name:           "http with port",
			origin:         "http://localhost:3000",
			allowedOrigins: []string{"http://localhost:3000"},
			expected:       true,
		},
		{
			name:           "multiple patterns - match first",
			origin:         "https://clpr.tv",
			allowedOrigins: []string{"https://clpr.tv", "https://www.clpr.tv", "*.staging.clpr.tv"},
			expected:       true,
		},
		{
			name:           "multiple patterns - match wildcard",
			origin:         "https://beta.staging.clpr.tv",
			allowedOrigins: []string{"https://clpr.tv", "*.staging.clpr.tv"},
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isOriginAllowed(tt.origin, tt.allowedOrigins)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		name     string
		origin   string
		pattern  string
		expected bool
	}{
		{
			name:     "exact match",
			origin:   "https://clpr.tv",
			pattern:  "https://clpr.tv",
			expected: true,
		},
		{
			name:     "wildcard subdomain",
			origin:   "https://staging.clpr.tv",
			pattern:  "*.clpr.tv",
			expected: true,
		},
		{
			name:     "wildcard base domain",
			origin:   "https://clpr.tv",
			pattern:  "*.clpr.tv",
			expected: true,
		},
		{
			name:     "wildcard with http",
			origin:   "http://staging.clpr.tv",
			pattern:  "*.clpr.tv",
			expected: true,
		},
		{
			name:     "wildcard with port",
			origin:   "http://staging.clpr.tv:8080",
			pattern:  "*.clpr.tv",
			expected: true,
		},
		{
			name:     "no match - different domain",
			origin:   "https://example.com",
			pattern:  "*.clpr.tv",
			expected: false,
		},
		{
			name:     "no match - partial domain",
			origin:   "https://fakeclpr.tv",
			pattern:  "*.clpr.tv",
			expected: false,
		},
		{
			name:     "no match - suffix but no dot",
			origin:   "https://evilclpr.tv",
			pattern:  "*.clpr.tv",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesPattern(tt.origin, tt.pattern)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestServerCheckOrigin(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		allowedOrigins []string
		expected       bool
	}{
		{
			name:           "allowed origin",
			origin:         "http://localhost:5173",
			allowedOrigins: []string{"http://localhost:5173"},
			expected:       true,
		},
		{
			name:           "disallowed origin",
			origin:         "https://evil.com",
			allowedOrigins: []string{"http://localhost:5173"},
			expected:       false,
		},
		{
			name:           "wildcard match",
			origin:         "https://staging.clpr.tv",
			allowedOrigins: []string{"*.clpr.tv"},
			expected:       true,
		},
		{
			name:           "missing origin header",
			origin:         "",
			allowedOrigins: []string{"http://localhost:5173"},
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.WebSocketConfig{
				AllowedOrigins: tt.allowedOrigins,
			}
			server := NewServer(nil, nil, cfg)

			req := &http.Request{
				Header: http.Header{},
			}
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			result := server.Upgrader.CheckOrigin(req)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateAllowedOrigins(t *testing.T) {
	// These tests just ensure validateAllowedOrigins doesn't panic
	tests := []struct {
		name    string
		origins []string
	}{
		{
			name:    "empty list",
			origins: []string{},
		},
		{
			name:    "valid origins",
			origins: []string{"https://clpr.tv", "http://localhost:5173"},
		},
		{
			name:    "wildcard origin",
			origins: []string{"*"},
		},
		{
			name:    "wildcard subdomain",
			origins: []string{"*.clpr.tv"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			assert.NotPanics(t, func() {
				validateAllowedOrigins(tt.origins)
			})
		})
	}
}
