package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"git.subcult.tv/subculture-collective/clpr/internal/middleware"
)

func registerContentRoutes(v1 *gin.RouterGroup, h *Handlers, svcs *Services, infra *Infrastructure) {
	// Tag routes
	tags := v1.Group("/tags")
	{
		// Public tag endpoints
		tags.GET("", h.Tag.ListTags)
		tags.GET("/search", middleware.RateLimitMiddleware(infra.Redis, 60, time.Minute), h.Tag.SearchTags)
		tags.GET("/:slug", h.Tag.GetTag)
		tags.GET("/:slug/clips", h.Tag.GetClipsByTag)
	}

	// Search routes
	search := v1.Group("/search")
	{
		// Public search endpoints with rate limiting (60 requests/minute = 1 per second)
		search.GET("", middleware.RateLimitMiddleware(infra.Redis, 60, time.Minute), h.Search.Search)
		search.GET("/suggestions", middleware.RateLimitMiddleware(infra.Redis, 60, time.Minute), h.Search.GetSuggestions)
		search.GET("/scores", middleware.RateLimitMiddleware(infra.Redis, 60, time.Minute), h.Search.SearchWithScores) // Hybrid search with similarity scores

		// Search analytics endpoints
		search.GET("/trending", middleware.RateLimitMiddleware(infra.Redis, 30, time.Minute), h.Search.GetTrendingSearches) // Popular searches (public)
		search.GET("/history", middleware.AuthMiddleware(svcs.Auth), h.Search.GetSearchHistory)                              // User search history (authenticated)

		// Admin-only analytics endpoints
		searchAdmin := search.Group("")
		searchAdmin.Use(middleware.AuthMiddleware(svcs.Auth))
		searchAdmin.Use(middleware.RequireRole("admin"))
		{
			searchAdmin.GET("/failed", h.Search.GetFailedSearches)     // Failed searches (admin only)
			searchAdmin.GET("/analytics", h.Search.GetSearchAnalytics) // Search analytics summary (admin only)
		}
	}

	// Submission routes (if submission handler is available)
	if h.Submission != nil {
		submissions := v1.Group("/submissions")
		submissions.Use(middleware.AuthMiddleware(svcs.Auth))
		{
			// User submission endpoints (10 submissions per hour per user)
			submissions.POST("", middleware.RateLimitMiddleware(infra.Redis, 10, time.Hour), h.Submission.SubmitClip)
			submissions.GET("", h.Submission.GetUserSubmissions)
			submissions.GET("/stats", h.Submission.GetSubmissionStats)
			// Metadata endpoint with rate limiting (100 requests/hour per user)
			submissions.GET("/metadata", middleware.RateLimitMiddleware(infra.Redis, 100, time.Hour), h.Submission.GetClipMetadata)
			// Check clip status endpoint to see if it can be claimed
			submissions.GET("/check/:clip_id", middleware.RateLimitMiddleware(infra.Redis, 100, time.Hour), h.Submission.CheckClipStatus)
		}
	}

	// Report routes
	reports := v1.Group("/reports")
	{
		// Submit a report (authenticated, rate limited)
		reports.POST("", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 10, time.Hour), h.Report.SubmitReport)
	}

	// Moderation appeal routes (user-facing)
	if h.Moderation != nil {
		moderationAppeals := v1.Group("/moderation")
		{
			moderationAppeals.POST("/appeals", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 5, time.Hour), h.Moderation.CreateAppeal)
			moderationAppeals.GET("/appeals", middleware.AuthMiddleware(svcs.Auth), h.Moderation.GetUserAppeals)
			// Twitch ban sync endpoint with rate limiting
			moderationAppeals.POST("/sync-bans", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 5, time.Hour), h.Moderation.SyncBans)

			// Ban management endpoints
			moderationAppeals.GET("/bans", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 60, time.Minute), h.Moderation.GetBans)
			moderationAppeals.POST("/ban", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 10, time.Hour), h.Moderation.CreateBan)
			moderationAppeals.GET("/ban/:id", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 60, time.Minute), h.Moderation.GetBanDetails)
			moderationAppeals.DELETE("/ban/:id", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 10, time.Hour), h.Moderation.RevokeBan)

			// Twitch ban management endpoints (enforces Twitch-specific scope requirements)
			moderationAppeals.POST("/twitch/ban", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 10, time.Hour), h.Moderation.TwitchBanUser)
			moderationAppeals.DELETE("/twitch/ban", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 10, time.Hour), h.Moderation.TwitchUnbanUser)

			// Moderator management endpoints
			moderationAppeals.GET("/moderators", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 60, time.Minute), h.Moderation.ListModerators)
			moderationAppeals.POST("/moderators", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 10, time.Hour), h.Moderation.AddModerator)
			moderationAppeals.DELETE("/moderators/:id", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 10, time.Hour), h.Moderation.RemoveModerator)
			moderationAppeals.PATCH("/moderators/:id", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 10, time.Hour), h.Moderation.UpdateModeratorPermissions)

			// Audit log endpoints (requires moderator or admin role)
			moderationAppeals.GET("/audit-logs", middleware.AuthMiddleware(svcs.Auth), middleware.RequireRole("admin", "moderator"), middleware.RateLimitMiddleware(infra.Redis, 60, time.Minute), h.AuditLog.ListModerationAuditLogs)
			moderationAppeals.GET("/audit-logs/export", middleware.AuthMiddleware(svcs.Auth), middleware.RequireRole("admin", "moderator"), middleware.RateLimitMiddleware(infra.Redis, 10, time.Hour), h.AuditLog.ExportModerationAuditLogs)
			moderationAppeals.GET("/audit-logs/:id", middleware.AuthMiddleware(svcs.Auth), middleware.RequireRole("admin", "moderator"), middleware.RateLimitMiddleware(infra.Redis, 60, time.Minute), h.AuditLog.GetModerationAuditLog)

			// Ban reason template endpoints
			moderationAppeals.GET("/ban-templates", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 60, time.Minute), h.BanReasonTemplate.ListTemplates)
			moderationAppeals.GET("/ban-templates/stats", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 60, time.Minute), h.BanReasonTemplate.GetUsageStats)
			moderationAppeals.GET("/ban-templates/:id", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 60, time.Minute), h.BanReasonTemplate.GetTemplate)
			moderationAppeals.POST("/ban-templates", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 20, time.Hour), h.BanReasonTemplate.CreateTemplate)
			moderationAppeals.PATCH("/ban-templates/:id", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 20, time.Hour), h.BanReasonTemplate.UpdateTemplate)
			moderationAppeals.DELETE("/ban-templates/:id", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 20, time.Hour), h.BanReasonTemplate.DeleteTemplate)
		}
	}
}
