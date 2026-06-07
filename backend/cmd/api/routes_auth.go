package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"git.subcult.tv/subculture-collective/clpr/internal/middleware"
)

func registerAuthRoutes(v1 *gin.RouterGroup, h *Handlers, svcs *Services, infra *Infrastructure) {
	cfg := infra.Config

	// Auth routes
	auth := v1.Group("/auth")
	{
		// Public auth endpoints with rate limiting (increased for legitimate OAuth flows)
		auth.GET("/twitch", middleware.RateLimitMiddleware(infra.Redis, 30, time.Minute), h.Auth.InitiateOAuth)
		auth.GET("/twitch/callback", middleware.RateLimitMiddleware(infra.Redis, 50, time.Minute), h.Auth.HandleCallback)
		auth.POST("/twitch/callback", middleware.RateLimitMiddleware(infra.Redis, 50, time.Minute), h.Auth.HandlePKCECallback)
		if cfg.Server.GinMode != "release" {
			auth.POST("/test-login", middleware.RateLimitMiddleware(infra.Redis, 30, time.Minute), h.Auth.TestLogin)
		}
		auth.POST("/refresh", middleware.RateLimitMiddleware(infra.Redis, 50, time.Minute), h.Auth.RefreshToken)
		auth.POST("/logout", h.Auth.Logout)

		// Protected auth endpoints
		auth.GET("/me", middleware.AuthMiddleware(svcs.Auth), h.Auth.GetCurrentUser)
		auth.POST("/twitch/reauthorize", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 3, time.Hour), h.Auth.ReauthorizeTwitch)

		// MFA routes (protected)
		mfa := auth.Group("/mfa")
		mfa.Use(middleware.AuthMiddleware(svcs.Auth))
		{
			// MFA enrollment
			mfa.POST("/enroll", middleware.RateLimitMiddleware(infra.Redis, 3, time.Hour), h.MFA.StartEnrollment)
			mfa.POST("/verify-enrollment", middleware.RateLimitMiddleware(infra.Redis, 10, time.Minute), h.MFA.VerifyEnrollment)

			// MFA status
			mfa.GET("/status", h.MFA.GetStatus)

			// MFA management
			mfa.POST("/regenerate-backup-codes", middleware.RateLimitMiddleware(infra.Redis, 5, time.Hour), h.MFA.RegenerateBackupCodes)
			mfa.POST("/disable", middleware.RateLimitMiddleware(infra.Redis, 3, time.Hour), h.MFA.DisableMFA)

			// Trusted devices
			mfa.GET("/trusted-devices", h.MFA.GetTrustedDevices)
			mfa.DELETE("/trusted-devices/:id", h.MFA.RevokeTrustedDevice)
		}

		// MFA login verification (special case - uses different middleware)
		auth.POST("/mfa/verify-login", middleware.RateLimitMiddleware(infra.Redis, 10, time.Minute), h.MFA.VerifyLogin)
	}
}
