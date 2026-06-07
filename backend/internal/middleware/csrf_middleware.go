package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	redispkg "git.subcult.tv/subculture-collective/clpr/pkg/redis"
)

const (
	// CSRFTokenLength is the length of CSRF tokens in bytes
	CSRFTokenLength = 32
	// CSRFTokenHeader is the header name for CSRF token
	CSRFTokenHeader = "X-CSRF-Token" // #nosec G101 -- not a credential, just header name
	// CSRFCookieName is the cookie name for CSRF token
	CSRFCookieName = "csrf_token"
	// CSRFTokenTTL is the time-to-live for CSRF tokens
	CSRFTokenTTL = 24 * time.Hour
)

// CSRFMiddleware provides CSRF protection for state-changing requests
// It uses the double-submit cookie pattern with server-side validation
func CSRFMiddleware(redis *redispkg.Client, secure bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip if Redis client is nil (for testing)
		if redis == nil {
			c.Next()
			return
		}

		// Skip CSRF check for safe methods (GET, HEAD, OPTIONS)
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			// Generate and set CSRF token for safe methods to be used in subsequent requests
			ensureCSRFToken(c, redis, secure)
			c.Next()
			return
		}

		// For state-changing methods (POST, PUT, DELETE, PATCH), validate CSRF token
		// Only enforce CSRF if using cookie-based authentication
		// (JWT in Authorization header is immune to CSRF)
		_, cookieErr := c.Cookie("access_token")
		if cookieErr != nil {
			// No cookie authentication, skip CSRF check
			c.Next()
			return
		}

		// Get CSRF token from header
		headerToken := c.GetHeader(CSRFTokenHeader)
		if headerToken == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "CSRF token missing",
			})
			c.Abort()
			return
		}

		// Get CSRF token from cookie
		cookieToken, err := c.Cookie(CSRFCookieName)
		if err != nil || cookieToken == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "CSRF token invalid or expired",
			})
			c.Abort()
			return
		}

		// Verify token matches and is valid
		if !verifyCSRFToken(c, redis, cookieToken, headerToken) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "CSRF token validation failed",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ensureCSRFToken generates and sets a CSRF token if one doesn't exist
func ensureCSRFToken(c *gin.Context, redis *redispkg.Client, secure bool) {
	// Check if token already exists in cookie
	existingToken, err := c.Cookie(CSRFCookieName)
	if err == nil && existingToken != "" {
		// Verify token is still valid in Redis
		ctx := c.Request.Context()
		key := fmt.Sprintf("csrf:%s", existingToken)
		exists, _ := redis.Exists(ctx, key)
		if exists {
			// Token still valid, set it in response header for frontend
			c.Header(CSRFTokenHeader, existingToken)
			return
		}
	}

	// Generate new token
	token, err := generateCSRFToken()
	if err != nil {
		// Log error but don't fail the request
		_ = c.Error(fmt.Errorf("failed to generate CSRF token: %w", err))
		return
	}

	// Store token in Redis with expiration
	ctx := c.Request.Context()
	key := fmt.Sprintf("csrf:%s", token)
	if err := redis.Set(ctx, key, "1", CSRFTokenTTL); err != nil {
		// Log error but don't fail the request
		_ = c.Error(fmt.Errorf("failed to store CSRF token: %w", err))
		return
	}

	// Set token in cookie
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		CSRFCookieName,
		token,
		int(CSRFTokenTTL.Seconds()),
		"/",
		"",
		secure,
		true, // HttpOnly to prevent XSS
	)

	// Also set in response header for immediate use
	c.Header(CSRFTokenHeader, token)
}

// verifyCSRFToken verifies that the CSRF token is valid
func verifyCSRFToken(ctx *gin.Context, redis *redispkg.Client, cookieToken, headerToken string) bool {
	// Constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(cookieToken), []byte(headerToken)) != 1 {
		return false
	}

	// Verify token exists in Redis
	key := fmt.Sprintf("csrf:%s", cookieToken)
	exists, err := redis.Exists(ctx, key)
	if err != nil || !exists {
		return false
	}

	return true
}

// generateCSRFToken generates a cryptographically secure random token
func generateCSRFToken() (string, error) {
	bytes := make([]byte, CSRFTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
