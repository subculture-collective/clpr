package middleware

import (
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	sentrypkg "git.subcult.tv/subculture-collective/clpr/pkg/sentry"
)

// SentryMiddleware adds Sentry error tracking to Gin
func SentryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a new hub for this request
		hub := sentry.CurrentHub().Clone()
		ctx := sentry.SetHubOnContext(c.Request.Context(), hub)
		c.Request = c.Request.WithContext(ctx)

		// Set request ID as tag for tracing
		if requestID := requestid.Get(c); requestID != "" {
			hub.ConfigureScope(func(scope *sentry.Scope) {
				scope.SetTag("request_id", requestID)
			})
		}

		// Set route as tag
		hub.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetTag("route", c.FullPath())
			scope.SetTag("method", c.Request.Method)
		})

		// Start transaction for performance monitoring
		span := sentry.StartSpan(ctx, "http.server",
			sentry.WithTransactionName(c.Request.Method+" "+c.FullPath()),
		)
		defer span.Finish()

		// Add breadcrumb for this request
		hub.AddBreadcrumb(&sentry.Breadcrumb{
			Type:     "http",
			Category: "request",
			Data: map[string]interface{}{
				"method": c.Request.Method,
				"url":    c.Request.URL.String(),
			},
			Level:     sentry.LevelInfo,
			Timestamp: time.Now(),
		}, nil)

		// Process request
		c.Next()

		// Capture errors from Gin context
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				hub.CaptureException(err.Err)
			}
		}

		// Set HTTP status on span
		span.Status = sentry.HTTPtoSpanStatus(c.Writer.Status())
	}
}

// RecoverWithSentry replaces the default Gin recovery middleware
// and sends panics to Sentry
func RecoverWithSentry() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Capture panic in Sentry
				hub := sentry.GetHubFromContext(c.Request.Context())
				if hub != nil {
					hub.RecoverWithContext(c.Request.Context(), err)
				}

				// Set user context if available
				if userID, exists := c.Get("user_id"); exists {
					if username, ok := c.Get("username"); ok {
						sentrypkg.SetUser(c, userID.(string), username.(string))
					}
				}

				// Return 500 error
				c.AbortWithStatusJSON(500, gin.H{
					"error": "Internal server error",
				})
			}
		}()

		c.Next()
	}
}
