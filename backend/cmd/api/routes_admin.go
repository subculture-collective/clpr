package main

import (
	"github.com/gin-gonic/gin"
	"github.com/subculture-collective/clipper/internal/middleware"
	"github.com/subculture-collective/clipper/internal/models"
)

func registerAdminRoutes(v1 *gin.RouterGroup, h *Handlers, svcs *Services, infra *Infrastructure) {
	// Admin routes
	admin := v1.Group("/admin")
	admin.Use(middleware.AuthMiddleware(svcs.Auth))
	admin.Use(middleware.RequireRole("admin", "moderator"))
	admin.Use(middleware.RequireMFAForAdminMiddleware(svcs.MFA)) // Enforce MFA for admin/moderator actions
	{
		// Clip sync (if available)
		if h.ClipSync != nil {
			sync := admin.Group("/sync")
			{
				sync.POST("/clips", h.ClipSync.TriggerSync)
				sync.GET("/status", h.ClipSync.GetSyncStatus)
			}
		}

		// Admin tag management
		adminTags := admin.Group("/tags")
		{
			adminTags.POST("", h.Tag.CreateTag)
			adminTags.PUT("/:id", h.Tag.UpdateTag)
			adminTags.DELETE("/:id", h.Tag.DeleteTag)

			// Tag blacklist management
			adminTags.GET("/blacklist", h.Tag.ListBlacklistedTags)
			adminTags.POST("/blacklist", h.Tag.AddBlacklistedTag)
			adminTags.DELETE("/blacklist/:id", h.Tag.RemoveBlacklistedTag)
		}

		// Submission moderation (if available)
		if h.Submission != nil {
			adminSubmissions := admin.Group("/submissions")
			{
				adminSubmissions.GET("", h.Submission.ListPendingSubmissions)
				adminSubmissions.GET("/rejection-reasons", h.Submission.GetRejectionReasonTemplates)
				adminSubmissions.POST("/:id/approve", h.Submission.ApproveSubmission)
				adminSubmissions.POST("/:id/reject", h.Submission.RejectSubmission)
				adminSubmissions.POST("/bulk-approve", h.Submission.BulkApproveSubmissions)
				adminSubmissions.POST("/bulk-reject", h.Submission.BulkRejectSubmissions)
			}
		}

		// Audit log routes
		auditLogs := admin.Group("/audit-logs")
		{
			auditLogs.GET("", h.AuditLog.ListAuditLogs)
			auditLogs.GET("/export", h.AuditLog.ExportAuditLogs)
		}

		// Report management
		adminReports := admin.Group("/reports")
		{
			adminReports.GET("", h.Report.ListReports)
			adminReports.GET("/:id", h.Report.GetReport)
			adminReports.PUT("/:id", h.Report.UpdateReport)
		}

		// User management (admin only - requires PermissionManageUsers)
		adminUsers := admin.Group("/users")
		{
			adminUsers.GET("", middleware.RequirePermission(models.PermissionManageUsers), h.AdminUser.ListUsers)
			adminUsers.POST("/:id/ban", middleware.RequirePermission(models.PermissionManageUsers), h.AdminUser.BanUser)
			adminUsers.POST("/:id/unban", middleware.RequirePermission(models.PermissionManageUsers), h.AdminUser.UnbanUser)
			adminUsers.PATCH("/:id/role", middleware.RequirePermission(models.PermissionManageUsers), h.AdminUser.UpdateUserRole)
			adminUsers.PATCH("/:id/karma", middleware.RequirePermission(models.PermissionManageUsers), h.AdminUser.UpdateUserKarma)
			adminUsers.POST("/:id/badges", middleware.RequirePermission(models.PermissionManageUsers), h.Reputation.AwardBadge)
			adminUsers.DELETE("/:id/badges/:badgeId", middleware.RequirePermission(models.PermissionManageUsers), h.Reputation.RemoveBadge)
			// Comment privilege suspension routes
			adminUsers.POST("/:id/suspend-comments", middleware.RequirePermission(models.PermissionManageUsers), h.AdminUser.SuspendCommentPrivileges)
			adminUsers.POST("/:id/lift-comment-suspension", middleware.RequirePermission(models.PermissionManageUsers), h.AdminUser.LiftCommentSuspension)
			adminUsers.GET("/:id/comment-suspension-history", middleware.RequirePermission(models.PermissionManageUsers), h.AdminUser.GetCommentSuspensionHistory)
			adminUsers.POST("/:id/toggle-comment-review", middleware.RequirePermission(models.PermissionManageUsers), h.AdminUser.ToggleCommentReview)
		}

		// Account type management (admin only)
		adminAccountTypes := admin.Group("/account-types")
		{
			adminAccountTypes.GET("/stats", middleware.RequirePermission(models.PermissionManageUsers), h.AccountType.GetAccountTypeStats)
			adminAccountTypes.GET("/conversions", middleware.RequirePermission(models.PermissionManageUsers), h.AccountType.GetRecentConversions)
			adminAccountTypes.POST("/users/:id/convert-to-moderator", middleware.RequirePermission(models.PermissionManageUsers), h.AccountType.ConvertToModerator)
		}

		// Analytics routes (admin only)
		analytics := admin.Group("/analytics")
		{
			analytics.GET("/overview", h.Analytics.GetPlatformOverview)
			analytics.GET("/content", h.Analytics.GetContentMetrics)
			analytics.GET("/trends", h.Analytics.GetPlatformTrends)

			// Engagement metrics routes
			analytics.GET("/health", h.Engagement.GetPlatformHealthMetrics)
			analytics.GET("/trending", h.Engagement.GetTrendingMetrics)
			analytics.GET("/alerts", h.Engagement.CheckAlerts)
			analytics.GET("/export", h.Engagement.ExportEngagementData)
		}

		// Revenue metrics (admin only)
		admin.GET("/revenue", h.Revenue.GetRevenueMetrics)

		// Contact message management (admin only)
		adminContact := admin.Group("/contact")
		{
			adminContact.GET("", h.Contact.GetContactMessages)
			adminContact.PUT("/:id/status", h.Contact.UpdateContactMessageStatus)
		}

		// Ad Campaign management (admin only)
		adminAds := admin.Group("/ads")
		{
			// Campaign CRUD
			adminAds.GET("/campaigns", h.Ad.ListCampaigns)
			adminAds.GET("/campaigns/:id", h.Ad.GetCampaign)
			adminAds.POST("/campaigns", h.Ad.CreateCampaign)
			adminAds.PUT("/campaigns/:id", h.Ad.UpdateCampaign)
			adminAds.DELETE("/campaigns/:id", h.Ad.DeleteCampaign)

			// Creative validation
			adminAds.POST("/validate-creative", h.Ad.ValidateCreative)

			// Campaign reports
			adminAds.GET("/reports/by-date", h.Ad.GetCampaignReportByDate)
			adminAds.GET("/reports/by-placement", h.Ad.GetCampaignReportByPlacement)
			adminAds.GET("/reports/by-campaign", h.Ad.GetCTRReportByCampaign)
			adminAds.GET("/reports/by-slot", h.Ad.GetCTRReportBySlot)

			// Experiments
			adminAds.GET("/experiments", h.Ad.ListExperiments)
			adminAds.GET("/experiments/:id/report", h.Ad.GetExperimentReport)
		}

		// Email monitoring and metrics (admin only)
		adminEmail := admin.Group("/email")
		{
			// Dashboard and metrics
			adminEmail.GET("/metrics/dashboard", h.EmailMetrics.GetDashboardMetrics)
			adminEmail.GET("/metrics", h.EmailMetrics.GetMetrics)
			adminEmail.GET("/metrics/templates", h.EmailMetrics.GetTemplateMetrics)

			// Email logs
			adminEmail.GET("/logs", h.EmailMetrics.SearchEmailLogs)

			// Alerts
			adminEmail.GET("/alerts", h.EmailMetrics.GetAlerts)
			adminEmail.POST("/alerts/:id/acknowledge", h.EmailMetrics.AcknowledgeAlert)
			adminEmail.POST("/alerts/:id/resolve", h.EmailMetrics.ResolveAlert)
		}

		// Moderation queue management (admin/moderator only)
		if h.Moderation != nil {
			moderation := admin.Group("/moderation")
			{
				// Event management (existing)
				moderation.GET("/events", h.Moderation.GetPendingEvents)
				moderation.GET("/events/:type", h.Moderation.GetEventsByType)
				moderation.POST("/events/:id/review", h.Moderation.MarkEventReviewed)
				moderation.POST("/events/:id/process", h.Moderation.ProcessEvent)
				moderation.GET("/stats", h.Moderation.GetEventStats)

				// Abuse detection (existing)
				moderation.GET("/abuse/:userId", h.Moderation.GetUserAbuseStats)

				// Moderation queue (new)
				moderation.GET("/queue", h.Moderation.GetModerationQueue)
				moderation.POST("/:id/approve", h.Moderation.ApproveContent)
				moderation.POST("/:id/reject", h.Moderation.RejectContent)
				moderation.POST("/bulk", h.Moderation.BulkModerate)
				moderation.GET("/queue/stats", h.Moderation.GetModerationStats)

				// Appeals management (admin)
				moderation.GET("/appeals", h.Moderation.GetAppeals)
				moderation.POST("/appeals/:id/resolve", h.Moderation.ResolveAppeal)

				// Audit logs and analytics
				moderation.GET("/audit", h.Moderation.GetModerationAuditLogs)
				moderation.GET("/analytics", h.Moderation.GetModerationAnalytics)

				// Toxicity classification metrics
				moderation.GET("/toxicity/metrics", h.Moderation.GetToxicityMetrics)
			}
		}

		// NSFW detection routes (admin only)
		nsfw := admin.Group("/nsfw")
		{
			nsfw.POST("/detect", h.NSFW.DetectImage)
			nsfw.POST("/batch-detect", h.NSFW.BatchDetect)
			nsfw.GET("/metrics", h.NSFW.GetMetrics)
			nsfw.GET("/health", h.NSFW.GetHealthCheck)
			nsfw.GET("/config", h.NSFW.GetConfig)
			nsfw.POST("/scan-clips", h.NSFW.ScanClipThumbnails)
		}

		// Creator verification management (admin only)
		adminVerification := admin.Group("/verification")
		{
			adminVerification.GET("/applications", h.Verification.ListApplications)
			adminVerification.GET("/applications/:id", h.Verification.GetApplicationByID)
			adminVerification.POST("/applications/:id/review", h.Verification.ReviewApplication)
			adminVerification.GET("/stats", h.Verification.GetApplicationStats)
			adminVerification.GET("/audit-logs", h.Verification.GetAuditLogs)
			adminVerification.GET("/users/:user_id/audit-logs", h.Verification.GetUserAuditHistory)
		}

		// Discovery list management (admin/moderator only)
		adminDiscoveryLists := admin.Group("/discovery-lists")
		{
			adminDiscoveryLists.GET("", h.DiscoveryList.AdminListDiscoveryLists)
			adminDiscoveryLists.POST("", h.DiscoveryList.AdminCreateDiscoveryList)
			adminDiscoveryLists.GET("/:id", h.DiscoveryList.GetDiscoveryList)
			adminDiscoveryLists.PUT("/:id", h.DiscoveryList.AdminUpdateDiscoveryList)
			adminDiscoveryLists.DELETE("/:id", h.DiscoveryList.AdminDeleteDiscoveryList)
			adminDiscoveryLists.GET("/:id/clips", h.DiscoveryList.GetDiscoveryListClips)
			adminDiscoveryLists.POST("/:id/clips", h.DiscoveryList.AdminAddClipToList)
			adminDiscoveryLists.DELETE("/:id/clips/:clipId", h.DiscoveryList.AdminRemoveClipFromList)
			adminDiscoveryLists.PUT("/:id/clips/reorder", h.DiscoveryList.AdminReorderListClips)
		}

		// Playlist script management (admin/moderator only)
		adminPlaylistScripts := admin.Group("/playlist-scripts")
		{
			adminPlaylistScripts.GET("", h.PlaylistScript.ListScripts)
			adminPlaylistScripts.POST("", h.PlaylistScript.CreateScript)
			adminPlaylistScripts.PUT("/:id", h.PlaylistScript.UpdateScript)
			adminPlaylistScripts.DELETE("/:id", h.PlaylistScript.DeleteScript)
			adminPlaylistScripts.POST("/:id/generate", h.PlaylistScript.GeneratePlaylist)
		}

		// Forum moderation management (admin/moderator only)
		adminForum := admin.Group("/forum")
		{
			adminForum.GET("/flagged", h.ForumModeration.GetFlaggedContent)
			adminForum.POST("/threads/:id/lock", h.ForumModeration.LockThread)
			adminForum.POST("/threads/:id/pin", h.ForumModeration.PinThread)
			adminForum.POST("/threads/:id/delete", h.ForumModeration.DeleteThread)
			adminForum.POST("/users/:id/ban", h.ForumModeration.BanUser)
			adminForum.GET("/moderation-log", h.ForumModeration.GetModerationLog)
			adminForum.GET("/bans", h.ForumModeration.GetUserBans)
		}

		// Broadcaster ranking management (admin only)
		adminBroadcasters := admin.Group("/broadcasters")
		{
			adminBroadcasters.POST("/refresh-rankings", h.Broadcaster.RefreshBroadcasterRankings)
		}

		// Webhook dead-letter queue management (admin only)
		webhookDLQ := admin.Group("/webhooks")
		{
			webhookDLQ.GET("/dlq", h.WebhookDLQ.GetDeadLetterQueue)
			webhookDLQ.POST("/dlq/:id/replay", h.WebhookDLQ.ReplayDeadLetterQueueItem)
			webhookDLQ.DELETE("/dlq/:id", h.WebhookDLQ.DeleteDeadLetterQueueItem)
		}
	}
}
