package main

import (
	"time"

	"github.com/gin-gonic/gin"
	"git.subcult.tv/subculture-collective/clpr/internal/middleware"
)

func registerClipRoutes(v1 *gin.RouterGroup, h *Handlers, svcs *Services, infra *Infrastructure) {
	// Clip routes
	clips := v1.Group("/clips")
	{
		// Public clip endpoints
		clips.GET("", h.Clip.ListClips)
		clips.GET("/:id", h.Clip.GetClip)
		clips.GET("/:id/related", h.Clip.GetRelatedClips)
		clips.GET("/:id/processing-status", middleware.RateLimitMiddleware(infra.Redis, 60, time.Minute), h.Clip.GetClipProcessingStatus)

		// Batch endpoint for media URLs (public, rate limited)
		clips.POST("/batch-media", middleware.RateLimitMiddleware(infra.Redis, 60, time.Minute), h.Clip.BatchGetClipMedia)

		// Clip tags (public)
		clips.GET("/:id/tags", h.Tag.GetClipTags)

		// Clip analytics (public)
		clips.GET("/:id/analytics", h.Analytics.GetClipAnalytics)
		clips.POST("/:id/track-view", h.Analytics.TrackClipView)

		// Clip engagement score (public)
		clips.GET("/:id/engagement", h.Engagement.GetContentEngagementScore)

		// Watch progress (optional authentication - works for both authenticated and anonymous users)
		clips.GET("/:id/progress", middleware.OptionalAuthMiddleware(svcs.Auth), h.WatchHistory.GetResumePosition)

		// List comments for a clip (public or authenticated)
		clips.GET("/:id/comments", h.Comment.ListComments)

		// Create comment (authenticated, rate limited)
		clips.POST("/:id/comments", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 10, time.Minute), h.Comment.CreateComment)

		// Protected clip endpoints (require authentication)
		clips.POST("/:id/vote", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 20, time.Minute), h.Clip.VoteOnClip)
		clips.POST("/:id/favorite", middleware.AuthMiddleware(svcs.Auth), h.Clip.AddFavorite)
		clips.DELETE("/:id/favorite", middleware.AuthMiddleware(svcs.Auth), h.Clip.RemoveFavorite)
		clips.POST("/:id/backfill", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 10, time.Minute), h.Clip.RequestClipBackfill)

		// Tag management for clips (authenticated, rate limited)
		clips.POST("/:id/tags", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 10, time.Minute), h.Tag.AddTagsToClip)
		clips.DELETE("/:id/tags/:slug", middleware.AuthMiddleware(svcs.Auth), h.Tag.RemoveTagFromClip)

		// Creator content management (authenticated)
		clips.PUT("/:id/metadata", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 10, time.Minute), h.Clip.UpdateClipMetadata)
		clips.PUT("/:id/visibility", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 10, time.Minute), h.Clip.UpdateClipVisibility)

		// User clip submission with rate limiting (10 per hour) - if Twitch client is available
		if h.ClipSync != nil {
			clips.POST("/request", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 10, time.Hour), h.ClipSync.RequestClip)
		}

		// Admin clip endpoints
		clips.PUT("/:id", middleware.AuthMiddleware(svcs.Auth), middleware.RequireRole("admin", "moderator"), h.Clip.UpdateClip)
		clips.DELETE("/:id", middleware.AuthMiddleware(svcs.Auth), middleware.RequireRole("admin"), h.Clip.DeleteClip)
	}

	// Scraped clips routes
	scrapedClips := v1.Group("/scraped-clips")
	{
		// Public endpoint for listing scraped clips (not claimed by users)
		scrapedClips.GET("", h.Clip.ListScrapedClips)
	}

	// Comment routes
	comments := v1.Group("/comments")
	{
		// Get replies to a comment (can be public or authenticated)
		comments.GET("/:id/replies", h.Comment.GetReplies)

		// Protected comment endpoints
		comments.PUT("/:id", middleware.AuthMiddleware(svcs.Auth), h.Comment.UpdateComment)
		comments.DELETE("/:id", middleware.AuthMiddleware(svcs.Auth), h.Comment.DeleteComment)
		comments.POST("/:id/vote", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 20, time.Minute), h.Comment.VoteOnComment)
	}

	// Favorite routes
	favorites := v1.Group("/favorites")
	{
		// Protected favorite endpoints (require authentication)
		favorites.GET("", middleware.AuthMiddleware(svcs.Auth), h.Favorite.ListUserFavorites)
	}
}
