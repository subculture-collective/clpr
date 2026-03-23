package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/subculture-collective/clipper/internal/middleware"
)

func registerUserRoutes(v1 *gin.RouterGroup, h *Handlers, svcs *Services, infra *Infrastructure) {
	// Reputation routes
	users := v1.Group("/users")
	{
		// Public user profile
		users.GET("/by-username/:username", h.User.GetUserByUsername)

		// User autocomplete for mentions/suggestions - must be before /:id to avoid route conflicts
		users.GET("/autocomplete", middleware.RateLimitMiddleware(infra.Redis, 100, time.Hour), h.User.SearchUsersAutocomplete)

		users.GET("/:id", middleware.OptionalAuthMiddleware(svcs.Auth), h.User.GetUserProfile)

		// Account claiming for unclaimed profiles
		users.POST("/claim-account", middleware.AuthMiddleware(svcs.Auth), h.User.ClaimAccount)

		// Public reputation endpoints
		users.GET("/:id/reputation", h.Reputation.GetUserReputation)
		users.GET("/:id/karma", h.Reputation.GetUserKarma)
		users.GET("/:id/badges", h.Reputation.GetUserBadges)

		// User activity endpoints
		users.GET("/:id/comments", h.User.GetUserComments)
		users.GET("/:id/clips", middleware.OptionalAuthMiddleware(svcs.Auth), h.User.GetUserClips)
		users.GET("/:id/activity", middleware.OptionalAuthMiddleware(svcs.Auth), h.User.GetUserActivity)
		users.GET("/:id/upvoted", h.User.GetUserUpvotedClips)
		users.GET("/:id/downvoted", h.User.GetUserDownvotedClips)

		// User social connections
		users.GET("/:id/followers", middleware.OptionalAuthMiddleware(svcs.Auth), h.User.GetUserFollowers)
		users.GET("/:id/following", middleware.OptionalAuthMiddleware(svcs.Auth), h.User.GetUserFollowing)
		users.GET("/:id/following/broadcasters", middleware.OptionalAuthMiddleware(svcs.Auth), h.User.GetFollowedBroadcasters)
		users.POST("/:id/follow", middleware.AuthMiddleware(svcs.Auth), h.User.FollowUser)
		users.DELETE("/:id/follow", middleware.AuthMiddleware(svcs.Auth), h.User.UnfollowUser)

		// User blocking
		users.POST("/:id/block", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 20, time.Minute), h.User.BlockUser)
		users.DELETE("/:id/block", middleware.AuthMiddleware(svcs.Auth), h.User.UnblockUser)
		users.GET("/me/blocked", middleware.AuthMiddleware(svcs.Auth), h.User.GetBlockedUsers)

		// Personal statistics (authenticated)
		users.GET("/me/stats", middleware.AuthMiddleware(svcs.Auth), h.Analytics.GetUserStats)

		// User engagement score (authenticated)
		users.GET("/:id/engagement", middleware.AuthMiddleware(svcs.Auth), h.Engagement.GetUserEngagementScore)

		// Profile management (authenticated)
		users.PUT("/me/profile", middleware.AuthMiddleware(svcs.Auth), h.UserSettings.UpdateProfile)
		users.PUT("/me/social-links", middleware.AuthMiddleware(svcs.Auth), h.UserSettings.UpdateSocialLinks)
		users.GET("/me/settings", middleware.AuthMiddleware(svcs.Auth), h.UserSettings.GetSettings)
		users.PUT("/me/settings", middleware.AuthMiddleware(svcs.Auth), h.UserSettings.UpdateSettings)

		// Data export (authenticated, rate limited)
		users.GET("/me/export", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 1, time.Hour), h.UserSettings.ExportData)

		// Cookie consent management (authenticated, rate limited)
		users.GET("/me/consent", middleware.AuthMiddleware(svcs.Auth), h.Consent.GetConsent)
		users.POST("/me/consent", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 30, time.Minute), h.Consent.SaveConsent)

		// Account deletion (authenticated, rate limited)
		users.POST("/me/delete", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 1, time.Hour), h.UserSettings.RequestAccountDeletion)
		users.POST("/me/delete/cancel", middleware.AuthMiddleware(svcs.Auth), h.UserSettings.CancelAccountDeletion)
		users.GET("/me/delete/status", middleware.AuthMiddleware(svcs.Auth), h.UserSettings.GetDeletionStatus)

		// Email logs for current user (authenticated)
		users.GET("/me/email-logs", middleware.AuthMiddleware(svcs.Auth), h.EmailMetrics.GetUserEmailLogs)

		// Account type endpoints
		users.GET("/:id/account-type", middleware.OptionalAuthMiddleware(svcs.Auth), h.AccountType.GetAccountType)
		users.GET("/:id/account-type/history", middleware.OptionalAuthMiddleware(svcs.Auth), h.AccountType.GetConversionHistory)
		users.POST("/me/convert-to-broadcaster", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 3, 24*time.Hour), h.AccountType.ConvertToBroadcaster)

		// Discovery list follows for current user (authenticated)
		users.GET("/me/discovery-list-follows", middleware.AuthMiddleware(svcs.Auth), h.DiscoveryList.GetUserFollowedLists)

		// Game follows for a user
		users.GET("/:id/games/following", h.Game.GetFollowedGames)
		// User feeds routes
		users.GET("/:id/feeds", middleware.OptionalAuthMiddleware(svcs.Auth), h.Feed.ListUserFeeds)
		users.POST("/:id/feeds", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 10, time.Hour), h.Feed.CreateFeed)
		users.GET("/:id/feeds/:feedId", middleware.OptionalAuthMiddleware(svcs.Auth), h.Feed.GetFeed)
		users.PUT("/:id/feeds/:feedId", middleware.AuthMiddleware(svcs.Auth), h.Feed.UpdateFeed)
		users.DELETE("/:id/feeds/:feedId", middleware.AuthMiddleware(svcs.Auth), h.Feed.DeleteFeed)
		users.GET("/:id/feeds/:feedId/clips", middleware.OptionalAuthMiddleware(svcs.Auth), h.Feed.GetFeedClips)
		users.POST("/:id/feeds/:feedId/clips", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 20, time.Minute), h.Feed.AddClipToFeed)
		users.DELETE("/:id/feeds/:feedId/clips/:clipId", middleware.AuthMiddleware(svcs.Auth), h.Feed.RemoveClipFromFeed)
		users.PUT("/:id/feeds/:feedId/clips/reorder", middleware.AuthMiddleware(svcs.Auth), h.Feed.ReorderFeedClips)
		users.POST("/:id/feeds/:feedId/follow", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 20, time.Minute), h.Feed.FollowFeed)
		users.DELETE("/:id/feeds/:feedId/follow", middleware.AuthMiddleware(svcs.Auth), h.Feed.UnfollowFeed)

		// Filter preset routes
		users.GET("/:id/filter-presets", middleware.AuthMiddleware(svcs.Auth), h.FilterPreset.GetUserPresets)
		users.POST("/:id/filter-presets", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 10, time.Hour), h.FilterPreset.CreatePreset)
		users.GET("/:id/filter-presets/:presetId", middleware.AuthMiddleware(svcs.Auth), h.FilterPreset.GetPreset)
		users.PUT("/:id/filter-presets/:presetId", middleware.AuthMiddleware(svcs.Auth), h.FilterPreset.UpdatePreset)
		users.DELETE("/:id/filter-presets/:presetId", middleware.AuthMiddleware(svcs.Auth), h.FilterPreset.DeletePreset)
	}

	// Creator analytics routes
	creators := v1.Group("/creators")
	{
		// Public creator analytics endpoints
		creators.GET("/:creatorName/analytics/overview", h.Analytics.GetCreatorAnalyticsOverview)
		creators.GET("/:creatorName/analytics/clips", h.Analytics.GetCreatorTopClips)
		creators.GET("/:creatorName/analytics/trends", h.Analytics.GetCreatorTrends)
		creators.GET("/:creatorName/analytics/audience", h.Analytics.GetCreatorAudienceInsights)

		// Creator clips listing (shows hidden clips if authenticated as creator)
		creators.GET("/:creatorName/clips", middleware.OptionalAuthMiddleware(svcs.Auth), h.Clip.ListCreatorClips)

		// Creator data export routes (authenticated, rate limited)
		creators.POST("/me/export/request", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 3, 24*time.Hour), h.Export.RequestExport)
		creators.GET("/me/exports", middleware.AuthMiddleware(svcs.Auth), h.Export.ListExportRequests)
		creators.GET("/me/export/status/:id", middleware.AuthMiddleware(svcs.Auth), h.Export.GetExportStatus)
		creators.GET("/me/export/download/:id", middleware.AuthMiddleware(svcs.Auth), h.Export.DownloadExport)
	}

	// Broadcaster routes
	broadcasters := v1.Group("/broadcasters")
	{
		// Popular broadcasters (must come before /:id route)
		broadcasters.GET("/popular", h.Broadcaster.ListPopularBroadcasters)

		// Broadcaster rankings by engagement score (must come before /:id route)
		broadcasters.GET("/rankings", h.Broadcaster.GetBroadcasterRankings)

		// Live status endpoints (must come before /:id route)
		if h.LiveStatus != nil {
			// Public list of all live broadcasters
			broadcasters.GET("/live", h.LiveStatus.ListLiveBroadcasters)
		}

		// Public broadcaster profile endpoint (with optional auth for follow status)
		broadcasters.GET("/:id", middleware.OptionalAuthMiddleware(svcs.Auth), h.Broadcaster.GetBroadcasterProfile)

		// Public broadcaster clips endpoint
		broadcasters.GET("/:id/clips", h.Broadcaster.ListBroadcasterClips)

		// Live status for specific broadcaster
		if h.LiveStatus != nil {
			broadcasters.GET("/:id/live-status", h.LiveStatus.GetBroadcasterLiveStatus)
		}

		// Protected broadcaster endpoints (require authentication)
		broadcasters.POST("/:id/follow", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 20, time.Minute), h.Broadcaster.FollowBroadcaster)
		broadcasters.DELETE("/:id/follow", middleware.AuthMiddleware(svcs.Auth), h.Broadcaster.UnfollowBroadcaster)
	}

	// Category routes
	categories := v1.Group("/categories")
	{
		// Public category endpoints
		categories.GET("", h.Category.ListCategories)
		categories.GET("/:slug", h.Category.GetCategory)
		categories.GET("/:slug/games", h.Category.ListCategoryGames)
		categories.GET("/:slug/clips", h.Category.ListCategoryClips)
	}
}
