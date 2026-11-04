package main

import (
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/middleware"
	"github.com/gin-gonic/gin"
)

func registerSocialRoutes(v1 *gin.RouterGroup, h *Handlers, svcs *Services, infra *Infrastructure) {
	// Chat routes
	chat := v1.Group("/chat")
	{
		// Chat channel routes
		channels := chat.Group("/channels")
		channels.Use(middleware.AuthMiddleware(svcs.Auth))
		{
			// Channel CRUD operations
			channels.POST("", middleware.RateLimitMiddleware(infra.Redis, 10, time.Minute), h.Chat.CreateChannel)
			channels.GET("", h.Chat.ListChannels)
			channels.GET("/:id", h.Chat.GetChannel)
			channels.PATCH("/:id", middleware.RateLimitMiddleware(infra.Redis, 10, time.Minute), h.Chat.UpdateChannel)
			channels.DELETE("/:id", middleware.RateLimitMiddleware(infra.Redis, 10, time.Minute), h.Chat.DeleteChannel)

			// Channel member management
			channels.GET("/:id/members", h.Chat.ListChannelMembers)
			channels.POST("/:id/members", middleware.RateLimitMiddleware(infra.Redis, 20, time.Minute), h.Chat.AddChannelMember)
			channels.DELETE("/:id/members/:user_id", middleware.RateLimitMiddleware(infra.Redis, 20, time.Minute), h.Chat.RemoveChannelMember)
			channels.PATCH("/:id/members/:user_id", middleware.RateLimitMiddleware(infra.Redis, 20, time.Minute), h.Chat.UpdateChannelMemberRole)
			channels.GET("/:id/role", h.Chat.GetCurrentUserRole)

			// WebSocket connection endpoint
			channels.GET("/:id/ws", h.WebSocket.HandleConnection)

			// Message history endpoint
			channels.GET("/:id/messages", h.WebSocket.GetMessageHistory)

			// Moderation endpoints enforce owner/admin/moderator membership in the handler.
			channels.POST("/:id/ban", middleware.RateLimitMiddleware(infra.Redis, 30, time.Minute), h.Chat.BanUser)
			channels.DELETE("/:id/ban/:user_id", middleware.RateLimitMiddleware(infra.Redis, 30, time.Minute), h.Chat.UnbanUser)
			channels.POST("/:id/mute", middleware.RateLimitMiddleware(infra.Redis, 30, time.Minute), h.Chat.MuteUser)
			channels.POST("/:id/timeout", middleware.RateLimitMiddleware(infra.Redis, 30, time.Minute), h.Chat.TimeoutUser)
			channels.GET("/:id/moderation-log", h.Chat.GetModerationLog)
			channels.GET("/:id/check-ban", h.Chat.CheckUserBan)
		}

		// Chat message routes
		messages := chat.Group("/messages")
		messages.Use(middleware.AuthMiddleware(svcs.Auth))
		{
			messages.DELETE("/:id", middleware.RateLimitMiddleware(infra.Redis, 30, time.Minute), h.Chat.DeleteMessage)
		}

		// Health check endpoint for WebSocket server
		chat.GET("/health", h.WebSocket.GetHealthCheck)
		chat.GET("/stats", middleware.AuthMiddleware(svcs.Auth), middleware.RequireRole("admin"), h.WebSocket.GetChannelStats)
	}

	// Community routes
	communities := v1.Group("/communities")
	{
		// Public community endpoints
		communities.GET("", h.Community.ListCommunities)
		communities.GET("/search", middleware.RateLimitMiddleware(infra.Redis, 60, time.Minute), h.Community.SearchCommunities)
		communities.GET("/:id", middleware.OptionalAuthMiddleware(svcs.Auth), h.Community.GetCommunity)
		communities.GET("/:id/members", middleware.OptionalAuthMiddleware(svcs.Auth), h.Community.GetMembers)
		communities.GET("/:id/feed", middleware.OptionalAuthMiddleware(svcs.Auth), h.Community.GetCommunityFeed)
		communities.GET("/:id/discussions", middleware.OptionalAuthMiddleware(svcs.Auth), h.Community.ListDiscussions)
		communities.GET("/:id/discussions/:discussionId", middleware.OptionalAuthMiddleware(svcs.Auth), h.Community.GetDiscussion)

		// Protected community endpoints (require authentication)
		communities.POST("", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 5, time.Hour), h.Community.CreateCommunity)
		communities.PUT("/:id", middleware.AuthMiddleware(svcs.Auth), h.Community.UpdateCommunity)
		communities.DELETE("/:id", middleware.AuthMiddleware(svcs.Auth), h.Community.DeleteCommunity)

		// Member management
		communities.POST("/:id/join", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 10, time.Minute), h.Community.JoinCommunity)
		communities.POST("/:id/leave", middleware.AuthMiddleware(svcs.Auth), h.Community.LeaveCommunity)
		communities.PUT("/:id/members/:userId/role", middleware.AuthMiddleware(svcs.Auth), h.Community.UpdateMemberRole)

		// Moderation
		communities.POST("/:id/ban", middleware.AuthMiddleware(svcs.Auth), h.Community.BanMember)
		communities.DELETE("/:id/ban/:userId", middleware.AuthMiddleware(svcs.Auth), h.Community.UnbanMember)
		communities.GET("/:id/bans", middleware.AuthMiddleware(svcs.Auth), h.Community.GetBannedMembers)

		// Community feed management
		communities.POST("/:id/clips", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 20, time.Minute), h.Community.AddClipToCommunity)
		communities.DELETE("/:id/clips/:clipId", middleware.AuthMiddleware(svcs.Auth), h.Community.RemoveClipFromCommunity)

		// Discussions
		communities.POST("/:id/discussions", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 10, time.Minute), h.Community.CreateDiscussion)
		communities.PUT("/:id/discussions/:discussionId", middleware.AuthMiddleware(svcs.Auth), h.Community.UpdateDiscussion)
		communities.DELETE("/:id/discussions/:discussionId", middleware.AuthMiddleware(svcs.Auth), h.Community.DeleteDiscussion)
	}

	// Playlist routes
	playlists := v1.Group("/playlists")
	{
		// Public playlist endpoints
		playlists.GET("/public", middleware.OptionalAuthMiddleware(svcs.Auth), h.Playlist.ListPublicPlaylists)
		playlists.GET("/featured", middleware.OptionalAuthMiddleware(svcs.Auth), h.Playlist.ListFeaturedPlaylists)
		playlists.GET("/today", middleware.OptionalAuthMiddleware(svcs.Auth), h.Playlist.GetPlaylistOfTheDay)
		playlists.GET("/share/:token", middleware.OptionalAuthMiddleware(svcs.Auth), h.Playlist.GetPlaylistByShareToken)
		playlists.GET("/bookmarks", middleware.AuthMiddleware(svcs.Auth), h.Playlist.ListBookmarkedPlaylists)
		playlists.GET("/:id", middleware.OptionalAuthMiddleware(svcs.Auth), h.Playlist.GetPlaylist)

		// Protected playlist endpoints (require authentication)
		playlists.POST("", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 20, time.Hour), h.Playlist.CreatePlaylist)
		playlists.GET("", middleware.AuthMiddleware(svcs.Auth), h.Playlist.ListUserPlaylists)
		playlists.PATCH("/:id", middleware.AuthMiddleware(svcs.Auth), h.Playlist.UpdatePlaylist)
		playlists.DELETE("/:id", middleware.AuthMiddleware(svcs.Auth), h.Playlist.DeletePlaylist)

		// Playlist item management
		playlists.POST("/:id/clips", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 60, time.Minute), h.Playlist.AddClipsToPlaylist)
		playlists.DELETE("/:id/clips/:clip_id", middleware.AuthMiddleware(svcs.Auth), h.Playlist.RemoveClipFromPlaylist)
		playlists.PUT("/:id/clips/order", middleware.AuthMiddleware(svcs.Auth), h.Playlist.ReorderPlaylistClips)
		playlists.POST("/:id/copy", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 10, time.Hour), h.Playlist.CopyPlaylist)

		// Playlist likes (social engagement)
		playlists.POST("/:id/like", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 30, time.Minute), h.Playlist.LikePlaylist)
		playlists.DELETE("/:id/like", middleware.AuthMiddleware(svcs.Auth), h.Playlist.UnlikePlaylist)
		playlists.POST("/:id/bookmark", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 30, time.Minute), h.Playlist.BookmarkPlaylist)
		playlists.DELETE("/:id/bookmark", middleware.AuthMiddleware(svcs.Auth), h.Playlist.UnbookmarkPlaylist)

		// Playlist sharing and collaboration
		playlists.GET("/:id/share-link", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 10, time.Hour), h.Playlist.GetShareLink)
		playlists.POST("/:id/track-share", middleware.RateLimitMiddleware(infra.Redis, 60, time.Minute), h.Playlist.TrackShare) // Public endpoint for analytics with rate limiting
		playlists.GET("/:id/collaborators", middleware.OptionalAuthMiddleware(svcs.Auth), h.Playlist.GetCollaborators)
		playlists.POST("/:id/collaborators", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 20, time.Hour), h.Playlist.AddCollaborator)
		playlists.DELETE("/:id/collaborators/:user_id", middleware.AuthMiddleware(svcs.Auth), h.Playlist.RemoveCollaborator)
		playlists.PATCH("/:id/collaborators/:user_id", middleware.AuthMiddleware(svcs.Auth), h.Playlist.UpdateCollaboratorPermission)
	}

	// User-scoped playlist script routes (smart playlists)
	playlistScripts := v1.Group("/playlist-scripts")
	playlistScripts.Use(middleware.AuthMiddleware(svcs.Auth))
	{
		playlistScripts.GET("", h.PlaylistScript.ListMyScripts)
		playlistScripts.POST("", middleware.RateLimitMiddleware(infra.Redis, 20, time.Hour), h.PlaylistScript.CreateMyScript)
		playlistScripts.PUT("/:id", h.PlaylistScript.UpdateMyScript)
		playlistScripts.DELETE("/:id", h.PlaylistScript.DeleteMyScript)
		playlistScripts.POST("/:id/generate", middleware.RateLimitMiddleware(infra.Redis, 10, time.Hour), h.PlaylistScript.GenerateMyPlaylist)
	}

	// Forum routes
	forum := v1.Group("/forum")
	{
		// Public forum endpoints
		forum.GET("/threads", h.Forum.ListThreads)
		forum.GET("/threads/:id", h.Forum.GetThread)
		forum.GET("/search", middleware.RateLimitMiddleware(infra.Redis, 30, time.Minute), h.Forum.SearchThreads)
		forum.GET("/replies/:id/votes", middleware.OptionalAuthMiddleware(svcs.Auth), h.Forum.GetReplyVotes)
		forum.GET("/users/:id/reputation", h.Forum.GetUserReputation)
		forum.GET("/analytics", middleware.RateLimitMiddleware(infra.Redis, 30, time.Minute), h.Forum.GetForumAnalytics)
		forum.GET("/popular", middleware.RateLimitMiddleware(infra.Redis, 30, time.Minute), h.Forum.GetPopularDiscussions)
		forum.GET("/helpful-replies", middleware.RateLimitMiddleware(infra.Redis, 30, time.Minute), h.Forum.GetMostHelpfulReplies)

		// Protected forum endpoints (require authentication)
		forum.POST("/threads", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 10, time.Hour), h.Forum.CreateThread)
		forum.POST("/threads/:id/replies", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 30, time.Minute), h.Forum.CreateReply)
		forum.PATCH("/replies/:id", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 20, time.Minute), h.Forum.UpdateReply)
		forum.DELETE("/replies/:id", middleware.AuthMiddleware(svcs.Auth), h.Forum.DeleteReply)
		forum.POST("/replies/:id/vote", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 50, time.Minute), h.Forum.VoteOnReply)
		forum.POST("/flag", middleware.AuthMiddleware(svcs.Auth), middleware.RateLimitMiddleware(infra.Redis, 10, time.Minute), h.ForumModeration.FlagContent)
	}

}
