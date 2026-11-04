package main

import (
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/middleware"
	"github.com/gin-gonic/gin"
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
		if cfg.AllowsTestLogin() {
			auth.POST("/test-login", middleware.RateLimitMiddleware(infra.Redis, 30, time.Minute), h.Auth.TestLogin)
		}
		auth.POST("/refresh", middleware.RateLimitMiddleware(infra.Redis, 50, time.Minute), h.Auth.RefreshToken)
		auth.POST("/logout", h.Auth.Logout)

		// Protected auth endpoints
		auth.GET("/me", middleware.AuthMiddleware(svcs.Auth), h.Auth.GetCurrentUser)
		auth.POST("/twitch/reauthorize", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 3, time.Hour), h.Auth.ReauthorizeTwitch)

	}
}
