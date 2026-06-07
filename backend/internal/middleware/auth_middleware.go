package middleware

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	sentrypkg "git.subcult.tv/subculture-collective/clpr/pkg/sentry"
)

// AuthMiddleware creates middleware that requires authentication
func AuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Missing authentication token",
				},
			})
			c.Abort()
			return
		}

		// Get user from token
		user, err := authService.GetUserFromToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Invalid or expired token",
				},
			})
			c.Abort()
			return
		}

		// Attach user to context
		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("user_role", user.Role)

		// Set user context in Sentry for error tracking
		sentrypkg.SetUser(c, user.ID.String(), user.Username)

		c.Next()
	}
}

// OptionalAuthMiddleware creates middleware that attaches user if authenticated
func OptionalAuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token != "" {
			user, err := authService.GetUserFromToken(c.Request.Context(), token)
			if err == nil {
				c.Set("user", user)
				c.Set("user_id", user.ID)
				c.Set("user_role", user.Role)

				// Set user context in Sentry for error tracking
				sentrypkg.SetUser(c, user.ID.String(), user.Username)
			}
		}
		c.Next()
	}
}

// RequireRole creates middleware that requires a specific role
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Authentication required",
				},
			})
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Invalid role format",
				},
			})
			c.Abort()
			return
		}

		// Check if user has required role
		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "FORBIDDEN",
				"message": "Insufficient permissions",
			},
		})
		c.Abort()
	}
}

// extractToken extracts JWT token from Authorization header, WebSocket subprotocol, or cookie
// WebSocket subprotocol support prevents tokens from appearing in URLs/query parameters
// which could be logged by proxies, load balancers, or access logs
func extractToken(c *gin.Context) string {
	// Try Authorization header first
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	// Try WebSocket subprotocol (Sec-WebSocket-Protocol header)
	// Format: "auth.bearer.<base64_token>"
	// This prevents tokens from appearing in URLs which are logged
	wsProtocol := c.GetHeader("Sec-WebSocket-Protocol")
	if wsProtocol != "" {
		// Parse subprotocol format: auth.bearer.<base64_token>
		parts := strings.Split(wsProtocol, ".")
		if len(parts) == 3 && parts[0] == "auth" && parts[1] == "bearer" {
			// Decode base64 token
			decoded, err := base64.StdEncoding.DecodeString(parts[2])
			if err == nil {
				return string(decoded)
			}
		}
	}

	// Fall back to cookie
	token, err := c.Cookie("access_token")
	if err == nil && token != "" {
		return token
	}

	return ""
}
