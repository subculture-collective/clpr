package sentry

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"git.subcult.tv/subculture-collective/clpr/config"
)

// Init initializes Sentry SDK with the given configuration
func Init(cfg *config.SentryConfig) error {
	if !cfg.Enabled || cfg.DSN == "" {
		return nil
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.DSN,
		Environment:      cfg.Environment,
		Release:          cfg.Release,
		TracesSampleRate: cfg.TracesSampleRate,
		// Attach stack traces to messages
		AttachStacktrace: true,
		// Before sending events, scrub sensitive data
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			return scrubSensitiveData(event)
		},
		// Sample errors more aggressively in production
		SampleRate: 1.0,
	})

	if err != nil {
		return fmt.Errorf("failed to initialize Sentry: %w", err)
	}

	return nil
}

// Close flushes any buffered events and shuts down Sentry
func Close() {
	sentry.Flush(2 * time.Second)
}

// scrubSensitiveData removes or masks PII from Sentry events
func scrubSensitiveData(event *sentry.Event) *sentry.Event {
	if event == nil {
		return nil
	}

	// Scrub sensitive request data
	if event.Request != nil {
		// Remove sensitive headers
		if event.Request.Headers != nil {
			delete(event.Request.Headers, "Authorization")
			delete(event.Request.Headers, "Cookie")
			delete(event.Request.Headers, "X-CSRF-Token")
		}

		// Remove sensitive query parameters
		if event.Request.QueryString != "" {
			// Don't include query strings that might contain tokens
			event.Request.QueryString = "[REDACTED]"
		}
	}

	// Scrub user data - keep only hashed ID
	if event.User.ID != "" {
		event.User.ID = hashUserID(event.User.ID)
		event.User.Email = ""
		event.User.Username = ""
		event.User.IPAddress = ""
	}

	// Remove breadcrumbs that might contain sensitive data
	filteredBreadcrumbs := make([]*sentry.Breadcrumb, 0, len(event.Breadcrumbs))
	for _, bc := range event.Breadcrumbs {
		if bc.Data != nil {
			delete(bc.Data, "password")
			delete(bc.Data, "token")
			delete(bc.Data, "secret")
			delete(bc.Data, "api_key")
		}
		filteredBreadcrumbs = append(filteredBreadcrumbs, bc)
	}
	event.Breadcrumbs = filteredBreadcrumbs

	return event
}

// hashUserID creates a SHA-256 hash of user ID for privacy
func hashUserID(userID string) string {
	hash := sha256.Sum256([]byte(userID))
	return hex.EncodeToString(hash[:8]) // Use first 8 bytes for shorter hash
}

// SetUser sets the user context for Sentry
func SetUser(c *gin.Context, userID, username string) {
	if hub := sentry.GetHubFromContext(c.Request.Context()); hub != nil {
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetUser(sentry.User{
				ID:       hashUserID(userID), // Hash for privacy
				Username: username,
			})
		})
	}
}

// SetTag sets a tag for Sentry context
func SetTag(c *gin.Context, key, value string) {
	if hub := sentry.GetHubFromContext(c.Request.Context()); hub != nil {
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetTag(key, value)
		})
	}
}

// SetContext sets additional context for Sentry
func SetContext(c *gin.Context, key string, data map[string]interface{}) {
	if hub := sentry.GetHubFromContext(c.Request.Context()); hub != nil {
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetContext(key, data)
		})
	}
}

// CaptureException captures an exception and sends it to Sentry
func CaptureException(c *gin.Context, err error) {
	if hub := sentry.GetHubFromContext(c.Request.Context()); hub != nil {
		hub.CaptureException(err)
	}
}

// CaptureMessage captures a message and sends it to Sentry
func CaptureMessage(c *gin.Context, message string) {
	if hub := sentry.GetHubFromContext(c.Request.Context()); hub != nil {
		hub.CaptureMessage(message)
	}
}

// AddBreadcrumb adds a breadcrumb to the current scope
func AddBreadcrumb(c *gin.Context, breadcrumb *sentry.Breadcrumb) {
	if hub := sentry.GetHubFromContext(c.Request.Context()); hub != nil {
		hub.AddBreadcrumb(breadcrumb, nil)
	}
}
