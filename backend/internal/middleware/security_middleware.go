package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"git.subcult.tv/subculture-collective/clpr/config"
)

// SecurityHeadersMiddleware adds security headers to all responses
func SecurityHeadersMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Strict-Transport-Security (HSTS)
		// Forces HTTPS for 1 year, includes subdomains, allows preloading
		if cfg.Server.GinMode == "release" {
			c.Writer.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		// X-Content-Type-Options
		// Prevents MIME type sniffing
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")

		// X-Frame-Options
		// Prevents clickjacking attacks
		c.Writer.Header().Set("X-Frame-Options", "DENY")

		// X-XSS-Protection
		// Enables XSS filter in older browsers
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")

		// Referrer-Policy
		// Controls how much referrer information is sent
		c.Writer.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content-Security-Policy
		// Helps prevent XSS, clickjacking, and other code injection attacks
		// Note: Twitch embed requires specific frame-src and media-src allowances
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.jsdelivr.net https://embed.twitch.tv; " +
			"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; " +
			"font-src 'self' https://fonts.gstatic.com; " +
			"img-src 'self' data: https: blob:; " +
			"media-src 'self' https://clips-media-assets2.twitch.tv https://clips.twitch.tv https://static.twitchcdn.net; " +
			"frame-src 'self' https://clips.twitch.tv https://player.twitch.tv https://embed.twitch.tv; " +
			"connect-src 'self' https://api.twitch.tv https://gql.twitch.tv; " +
			"object-src 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self'; " +
			"frame-ancestors 'none'; " +
			"upgrade-insecure-requests; " +
			"block-all-mixed-content"
		c.Writer.Header().Set("Content-Security-Policy", csp)

		// Permissions-Policy (formerly Feature-Policy)
		// Controls which browser features can be used
		c.Writer.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=()")

		c.Next()
	}
}

// SecureCookieOptions returns secure cookie configuration
type SecureCookieOptions struct {
	HTTPOnly bool
	Secure   bool
	SameSite string
	MaxAge   int
	Domain   string
	Path     string
}

// GetSecureCookieOptions returns cookie options based on environment
func GetSecureCookieOptions(cfg *config.Config) SecureCookieOptions {
	isProduction := cfg.Server.GinMode == "release"

	return SecureCookieOptions{
		HTTPOnly: true,         // Prevents JavaScript access to cookies
		Secure:   isProduction, // Only send over HTTPS in production
		SameSite: "lax",        // CSRF protection
		MaxAge:   86400,        // 24 hours
		Domain:   "",           // Will be set per-cookie based on config
		Path:     "/",          // Available to entire domain
	}
}

// SetSecureCookie sets a secure cookie with proper flags
func SetSecureCookie(c *gin.Context, name, value string, options SecureCookieOptions) {
	sameSite := http.SameSiteLaxMode // Default to Lax for better compatibility
	switch options.SameSite {
	case "strict":
		sameSite = http.SameSiteStrictMode
	case "lax":
		sameSite = http.SameSiteLaxMode
	case "none":
		sameSite = http.SameSiteNoneMode
	}

	c.SetSameSite(sameSite)
	c.SetCookie(
		name,
		value,
		options.MaxAge,
		options.Path,
		options.Domain,
		options.Secure,
		options.HTTPOnly,
	)
}
