package middleware

import (
	"context"
	"net/http"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/gin-gonic/gin"
)

// TokenUserResolver is the narrow authentication dependency required by the
// global mutation boundary.
type TokenUserResolver interface {
	GetUserFromToken(context.Context, string) (*models.User, error)
}

// PublicMutationPaths is the reviewed set of unauthenticated state-changing
// endpoints. Each endpoint has an independent integrity control (OAuth/MFA
// challenge, signed webhook, rate limit, or non-authoritative telemetry).
var PublicMutationPaths = map[string]string{
	"POST /api/v1/auth/twitch/callback":      "OAuth PKCE exchange",
	"POST /api/v1/auth/test-login":           "non-release test profile only",
	"POST /api/v1/auth/refresh":              "signed refresh token",
	"POST /api/v1/auth/logout":               "idempotent credential clearing",
	"POST /api/v1/auth/mfa/verify-login":     "pending MFA challenge",
	"POST /api/v1/clips/batch-media":         "read-only batch lookup",
	"POST /api/v1/clips/:id/track-view":      "rate-limited aggregate telemetry",
	"POST /api/v1/events":                    "rate-limited aggregate telemetry",
	"POST /api/v1/webhooks/stripe":           "Stripe signature verification",
	"POST /api/v1/webhooks/sendgrid":         "SendGrid signature verification",
	"POST /api/v1/contact":                   "rate-limited public contact form",
	"POST /api/v1/ads/track/:id":             "rate-limited aggregate telemetry",
	"POST /api/v1/logs":                      "rate-limited sanitized client logs",
	"POST /api/v1/playlists/:id/track-share": "rate-limited aggregate telemetry",
}

func isMutation(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func attachAuthenticatedUser(c *gin.Context, user *models.User) {
	c.Set("user", user)
	c.Set("user_id", user.ID)
	c.Set("user_role", user.Role)
}

// MutationAuthorizationBoundary makes authentication the default for every
// registered state-changing route. Public exceptions must be explicit above.
func MutationAuthorizationBoundary(auth TokenUserResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isMutation(c.Request.Method) || c.FullPath() == "" {
			c.Next()
			return
		}
		key := c.Request.Method + " " + c.FullPath()
		if _, allowed := PublicMutationPaths[key]; allowed {
			c.Next()
			return
		}
		if _, exists := c.Get("user_id"); exists {
			c.Next()
			return
		}
		token := extractToken(c)
		if token == "" || auth == nil {
			abortMutationUnauthorized(c)
			return
		}
		user, err := auth.GetUserFromToken(c.Request.Context(), token)
		if err != nil || user == nil {
			abortMutationUnauthorized(c)
			return
		}
		attachAuthenticatedUser(c, user)
		c.Next()
	}
}

func abortMutationUnauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"success": false,
		"error":   gin.H{"code": "UNAUTHORIZED", "message": "Authentication required for state-changing requests"},
	})
}
