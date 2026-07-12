package config

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
)

var validEnvironments = map[string]bool{
	"development": true,
	"local":       true,
	"test":        true,
	"staging":     true,
	"production":  true,
}

// Validate rejects incoherent or unsafe runtime profiles before infrastructure
// is initialized. Optional feature credentials are required only when their
// corresponding feature is enabled.
func (c *Config) Validate() error {
	environment := strings.ToLower(strings.TrimSpace(c.Server.Environment))
	if !validEnvironments[environment] {
		return fmt.Errorf("unsupported ENVIRONMENT %q", c.Server.Environment)
	}

	if err := validateURL("BASE_URL", c.Server.BaseURL, environment == "staging" || environment == "production"); err != nil {
		return err
	}
	if err := validateOrigins("CORS_ALLOWED_ORIGINS", c.CORS.AllowedOrigins, environment == "staging" || environment == "production"); err != nil {
		return err
	}
	for _, origin := range c.WebSocket.AllowedOrigins {
		if err := validateURL("WEBSOCKET_ALLOWED_ORIGINS", origin, environment == "staging" || environment == "production"); err != nil {
			return err
		}
	}

	if environment != "staging" && environment != "production" {
		return c.validateEnabledFeatures(false)
	}
	if c.Server.GinMode != "release" {
		return fmt.Errorf("GIN_MODE must be release in %s", environment)
	}
	if strings.TrimSpace(c.JWT.PrivateKey) == "" {
		return fmt.Errorf("JWT_PRIVATE_KEY is required in %s", environment)
	}
	if !validMFAKey(c.Security.MFAEncryptionKey) {
		return fmt.Errorf("MFA_ENCRYPTION_KEY must decode to at least 32 bytes in %s", environment)
	}
	if len(strings.TrimSpace(c.Security.OperationalToken)) < 32 {
		return fmt.Errorf("OPERATIONAL_AUTH_TOKEN must contain at least 32 characters in %s", environment)
	}
	if strings.TrimSpace(c.Database.Password) == "" || c.Database.Password == "CHANGEME_SECURE_PASSWORD_HERE" {
		return fmt.Errorf("DB_PASSWORD must be a non-default secret in %s", environment)
	}
	if strings.EqualFold(c.Database.SSLMode, "disable") {
		return fmt.Errorf("DB_SSLMODE must verify transport security in %s", environment)
	}
	if err := validateURL("OPENSEARCH_URL", c.OpenSearch.URL, true); err != nil {
		return err
	}
	if c.OpenSearch.InsecureSkipVerify {
		return fmt.Errorf("OPENSEARCH_INSECURE_SKIP_VERIFY must be false in %s", environment)
	}
	if strings.TrimSpace(c.Twitch.ClientID) == "" || strings.TrimSpace(c.Twitch.ClientSecret) == "" {
		return fmt.Errorf("TWITCH_CLIENT_ID and TWITCH_CLIENT_SECRET are required in %s", environment)
	}
	if err := validateURL("TWITCH_REDIRECT_URI", c.Twitch.RedirectURI, true); err != nil {
		return err
	}
	if c.Clip.StorageProvider == "" || strings.EqualFold(c.Clip.StorageProvider, "local") {
		return fmt.Errorf("CLIP_STORAGE_PROVIDER must be durable and non-local in %s", environment)
	}
	if strings.TrimSpace(c.Clip.StorageBucket) == "" || strings.TrimSpace(c.Clip.StoragePublicBaseURL) == "" {
		return fmt.Errorf("clip storage bucket and public base URL are required in %s", environment)
	}
	if err := validateURL("CLIP_STORAGE_PUBLIC_BASE_URL", c.Clip.StoragePublicBaseURL, true); err != nil {
		return err
	}
	if c.Telemetry.Enabled && c.Telemetry.Insecure {
		return fmt.Errorf("TELEMETRY_INSECURE must be false when telemetry is enabled in %s", environment)
	}

	return c.validateEnabledFeatures(true)
}

func (c *Config) validateEnabledFeatures(releaseProfile bool) error {
	if c.FeatureFlags.PremiumSubscriptions {
		if strings.TrimSpace(c.Stripe.SecretKey) == "" || len(c.Stripe.WebhookSecrets) == 0 ||
			strings.TrimSpace(c.Stripe.ProMonthlyPriceID) == "" || strings.TrimSpace(c.Stripe.ProYearlyPriceID) == "" {
			return fmt.Errorf("premium subscriptions require Stripe key, webhook secret, and both price IDs")
		}
		if err := validateURL("STRIPE_SUCCESS_URL", c.Stripe.SuccessURL, releaseProfile); err != nil {
			return err
		}
		if err := validateURL("STRIPE_CANCEL_URL", c.Stripe.CancelURL, releaseProfile); err != nil {
			return err
		}
	}
	if c.Email.Enabled && (strings.TrimSpace(c.Email.SendGridAPIKey) == "" || strings.TrimSpace(c.Email.FromEmail) == "") {
		return fmt.Errorf("email notifications require SENDGRID_API_KEY and EMAIL_FROM_ADDRESS")
	}
	if c.Embedding.Enabled && (strings.TrimSpace(c.Embedding.OpenAIAPIKey) == "" || strings.TrimSpace(c.Embedding.APIBaseURL) == "") {
		return fmt.Errorf("embeddings require an API key and EMBEDDING_API_BASE_URL")
	}
	if releaseProfile && (c.FeatureFlags.StreamClipCreation || c.FeatureFlags.LiveFeed || c.FeatureFlags.WatchParties) {
		return fmt.Errorf("incomplete stream clip, live feed, and watch party features must remain disabled in release profiles")
	}
	if releaseProfile && (c.CDN.Enabled || c.Mirror.Enabled) {
		return fmt.Errorf("CDN and mirroring must remain disabled until provider verification is production-ready")
	}
	return nil
}

func validateURL(name, raw string, requireHTTPS bool) error {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(raw))
	if err != nil || parsed.Host == "" {
		return fmt.Errorf("%s must be an absolute URL", name)
	}
	if requireHTTPS && parsed.Scheme != "https" {
		return fmt.Errorf("%s must use https in a release profile", name)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("%s must use http or https", name)
	}
	return nil
}

func validateOrigins(name, raw string, requireHTTPS bool) error {
	origins := parseCommaSeparatedList(raw)
	if len(origins) == 0 {
		return fmt.Errorf("%s must contain at least one origin", name)
	}
	for _, origin := range origins {
		if origin == "*" {
			return fmt.Errorf("%s must not contain a wildcard", name)
		}
		if err := validateURL(name, origin, requireHTTPS); err != nil {
			return err
		}
	}
	return nil
}

func validMFAKey(raw string) bool {
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(raw))
	return err == nil && len(decoded) >= 32
}
