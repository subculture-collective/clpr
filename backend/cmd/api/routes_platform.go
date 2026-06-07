package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"git.subcult.tv/subculture-collective/clpr/internal/middleware"
)

func registerPlatformRoutes(v1 *gin.RouterGroup, h *Handlers, svcs *Services, infra *Infrastructure) {
	// Stream routes
	if h.Stream != nil {
		streams := v1.Group("/streams")
		{
			// Get followed streamers (authenticated) - must be before /:streamer
			streams.GET("/following", middleware.AuthMiddleware(svcs.Auth), h.Stream.GetFollowedStreamers)

			// Public stream status endpoint
			streams.GET("/:streamer", h.Stream.GetStreamStatus)

			// Protected stream follow endpoints (authenticated, rate limited)
			streams.POST("/:streamer/follow", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 20, time.Minute), h.Stream.FollowStreamer)
			streams.DELETE("/:streamer/follow", middleware.AuthMiddleware(svcs.Auth), h.Stream.UnfollowStreamer)
			streams.GET("/:streamer/follow-status", middleware.AuthMiddleware(svcs.Auth), h.Stream.GetStreamFollowStatus)

			// Protected stream clip creation endpoint (authenticated, rate limited)
			streams.POST("/:streamer/clips", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 10, time.Hour), h.Stream.CreateClipFromStream)
		}
	}

	// Twitch OAuth routes for chat integration
	if h.TwitchOAuth != nil {
		twitch := v1.Group("/twitch")
		{
			// OAuth endpoints (authenticated, rate limited)
			twitch.GET("/oauth/authorize", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 20, time.Minute), h.TwitchOAuth.InitiateTwitchOAuth)
			twitch.GET("/oauth/callback", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 20, time.Minute), h.TwitchOAuth.TwitchOAuthCallback)

			// Auth status endpoint (can be called without auth to check status)
			twitch.GET("/auth/status", middleware.OptionalAuthMiddleware(svcs.Auth), h.TwitchOAuth.GetTwitchAuthStatus)

			// Revoke endpoint (authenticated)
			twitch.DELETE("/auth", middleware.AuthMiddleware(svcs.Auth), h.TwitchOAuth.RevokeTwitchAuth)
		}
	}

	if h.StreamerClipRoom != nil {
		streamerClipRooms := v1.Group("/streamer-clip-rooms")
		streamerClipRooms.Use(middleware.AuthMiddleware(svcs.Auth))
		{
			streamerClipRooms.GET("/:channel", h.StreamerClipRoom.GetRoom)
			streamerClipRooms.POST("/:channel/start", middleware.RateLimitMiddleware(infra.Redis, 10, time.Minute), h.StreamerClipRoom.StartRoom)
			streamerClipRooms.POST("/:channel/stop", h.StreamerClipRoom.StopRoom)
			// Gin requires sibling wildcard routes to use the same parameter name.
			// These routes still expose the API contract shape `/:roomId/...`; the
			// shared internal name avoids registration conflicts with `/:channel`.
			streamerClipRooms.GET("/:channel/items", h.StreamerClipRoom.ListItems)
			streamerClipRooms.POST("/:channel/items/:itemId/approve", middleware.RateLimitMiddleware(infra.Redis, 120, time.Minute), h.StreamerClipRoom.ApproveItem)
			streamerClipRooms.POST("/:channel/items/:itemId/reject", middleware.RateLimitMiddleware(infra.Redis, 120, time.Minute), h.StreamerClipRoom.RejectItem)
			streamerClipRooms.PUT("/:channel/items/order", h.StreamerClipRoom.ReorderItems)
			streamerClipRooms.GET("/:channel/ws", h.StreamerClipRoom.WebSocket)
		}
	}

	// Game routes
	games := v1.Group("/games")
	{
		// Public game endpoints
		games.GET("/trending", h.Game.GetTrendingGames)
		games.GET("/:gameId", middleware.OptionalAuthMiddleware(svcs.Auth), h.Game.GetGame)
		games.GET("/:gameId/clips", h.Game.ListGameClips)

		// Protected game endpoints (require authentication)
		games.POST("/:gameId/follow", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 20, time.Minute), h.Game.FollowGame)
		games.DELETE("/:gameId/follow", middleware.AuthMiddleware(svcs.Auth), h.Game.UnfollowGame)
	}

	// Discovery list routes
	discoveryLists := v1.Group("/discovery-lists")
	{
		// Public discovery list endpoints
		discoveryLists.GET("", middleware.OptionalAuthMiddleware(svcs.Auth), h.DiscoveryList.ListDiscoveryLists)
		discoveryLists.GET("/:id", middleware.OptionalAuthMiddleware(svcs.Auth), h.DiscoveryList.GetDiscoveryList)
		discoveryLists.GET("/:id/clips", middleware.OptionalAuthMiddleware(svcs.Auth), h.DiscoveryList.GetDiscoveryListClips)

		// Protected discovery list endpoints (require authentication)
		discoveryLists.POST("/:id/follow", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 20, time.Minute), h.DiscoveryList.FollowDiscoveryList)
		discoveryLists.DELETE("/:id/follow", middleware.AuthMiddleware(svcs.Auth), h.DiscoveryList.UnfollowDiscoveryList)
		discoveryLists.POST("/:id/bookmark", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 20, time.Minute), h.DiscoveryList.BookmarkDiscoveryList)
		discoveryLists.DELETE("/:id/bookmark", middleware.AuthMiddleware(svcs.Auth), h.DiscoveryList.UnbookmarkDiscoveryList)
	}

	// Leaderboard routes
	leaderboards := v1.Group("/leaderboards")
	{
		// Public leaderboard endpoints
		leaderboards.GET("/:type", h.Reputation.GetLeaderboard)
	}

	// Badge definitions (public)
	v1.GET("/badges", h.Reputation.GetBadgeDefinitions)

	// Feed discovery and search routes
	feeds := v1.Group("/feeds")
	{
		// Public feed discovery endpoints
		feeds.GET("/discover", h.Feed.DiscoverFeeds)
		feeds.GET("/search", middleware.RateLimitMiddleware(infra.Redis, 60, time.Minute), h.Feed.SearchFeeds)

		// Comprehensive feed filtering endpoint
		feeds.GET("/clips", middleware.OptionalAuthMiddleware(svcs.Auth), h.Feed.GetFilteredClips)

		// Following feed (authenticated)
		feeds.GET("/following", middleware.AuthMiddleware(svcs.Auth), h.Feed.GetFollowingFeed)

		// Feed analytics routes (admin only)
		feeds.GET("/analytics", middleware.AuthMiddleware(svcs.Auth), middleware.RequireRole("admin"), h.Event.GetFeedMetrics)
		feeds.GET("/analytics/hourly", middleware.AuthMiddleware(svcs.Auth), middleware.RequireRole("admin"), h.Event.GetHourlyMetrics)
	}

	// Events tracking endpoint
	v1.POST("/events", middleware.RateLimitMiddleware(infra.Redis, 100, time.Minute), h.Event.TrackEvent)

	// Live feed (authenticated)
	if h.LiveStatus != nil {
		v1.GET("/feed/live", middleware.AuthMiddleware(svcs.Auth), h.LiveStatus.GetFollowedLiveBroadcasters)
	}

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

	// Subscription routes
	subscriptions := v1.Group("/subscriptions")
	{
		// Webhook endpoint (public, no auth required)
		v1.POST("/webhooks/stripe", h.Subscription.HandleWebhook)
		// SendGrid webhook endpoint (public, no auth required, signature verified internally)
		v1.POST("/webhooks/sendgrid", h.SendGridWebhook.HandleWebhook)

		// Protected subscription endpoints (require authentication)
		subscriptions.Use(middleware.AuthMiddleware(svcs.Auth))
		subscriptions.GET("/me", h.Subscription.GetSubscription)
		subscriptions.POST("/checkout", middleware.RateLimitMiddleware(infra.Redis, 5, time.Minute), h.Subscription.CreateCheckoutSession)
		subscriptions.POST("/portal", middleware.RateLimitMiddleware(infra.Redis, 10, time.Minute), h.Subscription.CreatePortalSession)
		subscriptions.POST("/change-plan", middleware.RateLimitMiddleware(infra.Redis, 5, time.Minute), h.Subscription.ChangeSubscriptionPlan)
		subscriptions.POST("/cancel", middleware.RateLimitMiddleware(infra.Redis, 5, time.Minute), h.Subscription.CancelSubscription)
		subscriptions.POST("/reactivate", middleware.RateLimitMiddleware(infra.Redis, 5, time.Minute), h.Subscription.ReactivateSubscription)
		subscriptions.GET("/invoices", middleware.RateLimitMiddleware(infra.Redis, 10, time.Minute), h.Subscription.GetInvoices)
	}

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
		ads.GET("/select", middleware.RateLimitMiddleware(infra.Redis, 60, time.Minute), h.Ad.SelectAd)
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
