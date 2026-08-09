package config

import (
	"encoding/base64"
	"strings"
	"testing"
)

func validReleaseConfig() *Config {
	return &Config{
		Server:     ServerConfig{Environment: "production", GinMode: "release", BaseURL: "https://clpr.example"},
		Database:   DatabaseConfig{Password: "not-a-default", SSLMode: "verify-full"},
		JWT:        JWTConfig{PrivateKey: "configured"},
		Twitch:     TwitchConfig{ClientID: "client", ClientSecret: "secret", RedirectURI: "https://clpr.example/api/v1/auth/twitch/callback"},
		Clip:       ClipConfig{StorageProvider: "s3", StorageBucket: "clips", StoragePublicBaseURL: "https://media.clpr.example"},
		CORS:       CORSConfig{AllowedOrigins: "https://clpr.example"},
		WebSocket:  WebSocketConfig{AllowedOrigins: []string{"https://clpr.example"}},
		OpenSearch: OpenSearchConfig{URL: "https://search.internal.example", InsecureSkipVerify: false},
		Security: SecurityConfig{
			MFAEncryptionKey: base64.StdEncoding.EncodeToString(make([]byte, 32)),
			OperationalToken: strings.Repeat("o", 32),
		},
	}
}

func TestValidateProfiles(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr string
	}{
		{name: "valid production"},
		{name: "valid staging", mutate: func(c *Config) { c.Server.Environment = "staging" }},
		{name: "unknown environment", mutate: func(c *Config) { c.Server.Environment = "prod" }, wantErr: "unsupported ENVIRONMENT"},
		{name: "debug production", mutate: func(c *Config) { c.Server.GinMode = "debug" }, wantErr: "GIN_MODE"},
		{name: "http base URL", mutate: func(c *Config) { c.Server.BaseURL = "http://clpr.example" }, wantErr: "BASE_URL must use https"},
		{name: "wildcard CORS", mutate: func(c *Config) { c.CORS.AllowedOrigins = "*" }, wantErr: "wildcard"},
		{name: "missing JWT key", mutate: func(c *Config) { c.JWT.PrivateKey = "" }, wantErr: "JWT_PRIVATE_KEY"},
		{name: "invalid MFA key", mutate: func(c *Config) { c.Security.MFAEncryptionKey = "short" }, wantErr: "MFA_ENCRYPTION_KEY"},
		{name: "missing operational token", mutate: func(c *Config) { c.Security.OperationalToken = "" }, wantErr: "OPERATIONAL_AUTH_TOKEN"},
		{name: "default DB password", mutate: func(c *Config) { c.Database.Password = "CHANGEME_SECURE_PASSWORD_HERE" }, wantErr: "DB_PASSWORD"},
		{name: "database TLS disabled", mutate: func(c *Config) { c.Database.SSLMode = "disable" }, wantErr: "DB_SSLMODE"},
		{name: "OpenSearch TLS disabled", mutate: func(c *Config) { c.OpenSearch.URL = "http://search.internal.example" }, wantErr: "OPENSEARCH_URL must use https"},
		{name: "OpenSearch verification disabled", mutate: func(c *Config) { c.OpenSearch.InsecureSkipVerify = true }, wantErr: "OPENSEARCH_INSECURE_SKIP_VERIFY"},
		{name: "local clip storage", mutate: func(c *Config) { c.Clip.StorageProvider = "local" }, wantErr: "durable and non-local"},
		{name: "incomplete feature enabled", mutate: func(c *Config) { c.FeatureFlags.WatchParties = true }, wantErr: "must remain disabled"},
		{name: "CDN enabled", mutate: func(c *Config) { c.CDN.Enabled = true }, wantErr: "CDN and mirroring"},
		{name: "premium missing Stripe config", mutate: func(c *Config) { c.FeatureFlags.PremiumSubscriptions = true }, wantErr: "premium subscriptions require"},
		{name: "insecure telemetry", mutate: func(c *Config) { c.Telemetry.Enabled = true; c.Telemetry.Insecure = true }, wantErr: "TELEMETRY_INSECURE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validReleaseConfig()
			if tt.mutate != nil {
				tt.mutate(cfg)
			}
			err := cfg.Validate()
			if tt.wantErr == "" && err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
			if tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)) {
				t.Fatalf("Validate() error = %v, want substring %q", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDevelopmentAllowsLocalDefaultsButValidatesEnabledFeatures(t *testing.T) {
	cfg := &Config{
		Server:       ServerConfig{Environment: "development", GinMode: "debug", BaseURL: "http://localhost:5173"},
		CORS:         CORSConfig{AllowedOrigins: "http://localhost:5173"},
		WebSocket:    WebSocketConfig{AllowedOrigins: []string{"http://localhost:5173"}},
		FeatureFlags: FeatureFlagsConfig{PremiumSubscriptions: true},
	}
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "Stripe") {
		t.Fatalf("Validate() error = %v, want missing Stripe configuration", err)
	}
}
