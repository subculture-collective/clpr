package main

import (
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/middleware"
	"github.com/gin-gonic/gin"
)

func registerPlatformRoutes(v1 *gin.RouterGroup, h *Handlers, svcs *Services, infra *Infrastructure) {
	// Events tracking endpoint
	v1.POST("/events", middleware.RateLimitMiddleware(infra.Redis, 100, time.Minute), h.Event.TrackEvent)

	// Recommendation routes
	recommendations := v1.Group("/recommendations")
	{
		// All recommendation endpoints require authentication
		recommendations.Use(middleware.AuthMiddleware(svcs.Auth))

		// Get personalized clip recommendations
		recommendations.GET("/clips", middleware.RateLimitMiddleware(infra.Redis, 60, time.Minute), h.Recommendation.GetRecommendations)

		// Submit feedback on recommendations
		recommendations.POST("/feedback", middleware.RateLimitMiddleware(infra.Redis, 100, time.Minute), h.Recommendation.SubmitFeedback)

		// Get user preferences
		recommendations.GET("/preferences", h.Recommendation.GetPreferences)

		// Update user preferences
		recommendations.PUT("/preferences", middleware.RateLimitMiddleware(infra.Redis, 10, time.Minute), h.Recommendation.UpdatePreferences)

		// Complete onboarding flow
		recommendations.POST("/onboarding", middleware.RateLimitMiddleware(infra.Redis, 5, time.Minute), h.Recommendation.CompleteOnboarding)

		// Track view for recommendation engine
		recommendations.POST("/track-view/:id", middleware.RateLimitMiddleware(infra.Redis, 200, time.Minute), h.Recommendation.TrackView)
	}

	// Notification routes
	notifications := v1.Group("/notifications")
	{
		// Public unsubscribe endpoint (no auth required, uses token, but rate limited)
		notifications.GET("/unsubscribe", middleware.RateLimitMiddleware(infra.Redis, 10, time.Minute), h.Notification.Unsubscribe)

		// Protected notification endpoints (require authentication)
		notifications.Use(middleware.AuthMiddleware(svcs.Auth))

		// Get notifications list
		notifications.GET("", h.Notification.ListNotifications)

		// Get unread count
		notifications.GET("/count", h.Notification.GetUnreadCount)

		// Mark notification as read
		notifications.PUT("/:id/read", h.Notification.MarkAsRead)

		// Mark all notifications as read
		notifications.PUT("/read-all", h.Notification.MarkAllAsRead)

		// Delete notification
		notifications.DELETE("/:id", h.Notification.DeleteNotification)

		// Get/Update preferences
		notifications.GET("/preferences", h.Notification.GetPreferences)
		notifications.PUT("/preferences", h.Notification.UpdatePreferences)
		notifications.POST("/preferences/reset", h.Notification.ResetPreferences)

		// Device token registration for push notifications
		notifications.POST("/register", h.Notification.RegisterDeviceToken)
		notifications.DELETE("/unregister", h.Notification.UnregisterDeviceToken)
	}

	// Creator verification routes
	verification := v1.Group("/verification")
	{
		// Protected endpoints (require authentication)
		verification.Use(middleware.AuthMiddleware(svcs.Auth))
		verification.POST("/applications", middleware.RateLimitMiddleware(infra.Redis, 1, time.Hour), h.Verification.CreateApplication)
		verification.GET("/applications/me", h.Verification.GetApplication)
	}

	// SendGrid verifies its provider signature internally.
	v1.POST("/webhooks/sendgrid", h.SendGridWebhook.HandleWebhook)

	// Outbound webhook subscription routes
	webhooks := v1.Group("/webhooks")
	{
		// Get supported webhook events (public, rate-limited)
		webhooks.GET("/events", middleware.RateLimitMiddleware(infra.Redis, 60, time.Minute), h.WebhookSubscription.GetSupportedEvents)

		// Protected webhook subscription endpoints (require authentication)
		webhooks.Use(middleware.AuthMiddleware(svcs.Auth))

		// CRUD operations for webhook subscriptions
		webhooks.POST("", middleware.RateLimitMiddleware(infra.Redis, 10, time.Hour), h.WebhookSubscription.CreateSubscription)
		webhooks.GET("", h.WebhookSubscription.ListSubscriptions)
		webhooks.GET("/:id", h.WebhookSubscription.GetSubscription)
		webhooks.PATCH("/:id", h.WebhookSubscription.UpdateSubscription)
		webhooks.DELETE("/:id", h.WebhookSubscription.DeleteSubscription)

		// Secret regeneration
		webhooks.POST("/:id/regenerate-secret", middleware.RateLimitMiddleware(infra.Redis, 5, time.Hour), h.WebhookSubscription.RegenerateSecret)

		// Delivery history
		webhooks.GET("/:id/deliveries", h.WebhookSubscription.GetSubscriptionDeliveries)
	}

	// Contact routes
	contact := v1.Group("/contact")
	{
		// Public contact form submission with rate limiting
		contact.POST("", middleware.RateLimitMiddleware(infra.Redis, 3, time.Hour), h.Contact.SubmitContactMessage)
	}

	// Ad routes
	ads := v1.Group("/ads")
	{
		// Ad selection endpoint - rate limited to prevent abuse
		ads.GET("/select", middleware.OptionalAuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 60, time.Minute), h.Ad.SelectAd)
		// Ad tracking endpoint - higher rate limit for tracking callbacks
		ads.POST("/track/:id", middleware.RateLimitMiddleware(infra.Redis, 120, time.Minute), h.Ad.TrackImpression)
		// Get ad by ID (public)
		ads.GET("/:id", h.Ad.GetAd)
	}

	// Documentation routes (public access)
	docs := v1.Group("/docs")
	{
		docs.GET("", h.Docs.GetDocsList)
		docs.GET("/search", middleware.RateLimitMiddleware(infra.Redis, 60, time.Minute), h.Docs.SearchDocs)
		docs.GET("/content/*path", h.Docs.GetDoc)
		// Catch-all route must be last
		docs.GET("/:path", h.Docs.GetDoc) // Changed from /*path to /:path to avoid conflict
	}

	// Queue routes (clip playback queue)
	queue := v1.Group("/queue")
	queue.Use(middleware.AuthMiddleware(svcs.Auth))
	{
		// Queue management
		queue.GET("", h.Queue.GetQueue)
		queue.GET("/count", h.Queue.GetQueueCount)
		queue.POST("", middleware.RateLimitMiddleware(infra.Redis, 60, time.Minute), h.Queue.AddToQueue)
		queue.DELETE("", h.Queue.ClearQueue)
		queue.DELETE("/:id", h.Queue.RemoveFromQueue)
		queue.PATCH("/reorder", h.Queue.ReorderQueue)
		queue.POST("/:id/played", h.Queue.MarkAsPlayed)
		queue.POST("/convert-to-playlist", middleware.RateLimitMiddleware(infra.Redis, 10, time.Minute), h.Queue.ConvertToPlaylist)
	}

	// Watch history routes
	watchHistory := v1.Group("/watch-history")
	{
		// Record watch progress (authenticated, rate limited - 120 requests per minute)
		watchHistory.POST("", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 120, time.Minute), h.WatchHistory.RecordWatchProgress)

		// Get watch history (authenticated)
		watchHistory.GET("", middleware.AuthMiddleware(svcs.Auth), h.WatchHistory.GetWatchHistory)

		// Clear watch history (authenticated)
		watchHistory.DELETE("", middleware.AuthMiddleware(svcs.Auth), h.WatchHistory.ClearWatchHistory)
	}
}
