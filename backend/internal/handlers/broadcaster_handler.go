package handlers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/pkg/twitch"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
	"github.com/gin-gonic/gin"
)

// BroadcasterHandler handles broadcaster-related HTTP requests
type BroadcasterHandler struct {
	broadcasterRepo      *repository.BroadcasterRepository
	rankingRefresher     broadcasterRankingRefresher
	creatorDiscoveryRepo creatorDiscoveryLister
	clipRepo             *repository.ClipRepository
	twitchClient         *twitch.Client
	authService          *services.AuthService
}

type broadcasterRankingRefresher interface {
	RefreshRankings(context.Context) error
}

type creatorDiscoveryLister interface {
	ListCreatorDiscovery(context.Context, int) (*models.CreatorDiscoveryRails, error)
}

// NewBroadcasterHandler creates a new broadcaster handler
func NewBroadcasterHandler(
	broadcasterRepo *repository.BroadcasterRepository,
	clipRepo *repository.ClipRepository,
	twitchClient *twitch.Client,
	authService *services.AuthService,
) *BroadcasterHandler {
	return &BroadcasterHandler{
		broadcasterRepo:      broadcasterRepo,
		rankingRefresher:     broadcasterRepo,
		creatorDiscoveryRepo: broadcasterRepo,
		clipRepo:             clipRepo,
		twitchClient:         twitchClient,
		authService:          authService,
	}
}

// GetBroadcasterProfile returns a broadcaster's profile information
// GET /api/v1/broadcasters/:id
func (h *BroadcasterHandler) GetBroadcasterProfile(c *gin.Context) {
	broadcasterID := c.Param("id")
	if broadcasterID == "" || len(broadcasterID) > 128 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "broadcaster_id is required"})
		return
	}

	ctx := c.Request.Context()

	// Get broadcaster info from database
	broadcasterName, err := h.broadcasterRepo.GetBroadcasterByID(ctx, broadcasterID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "broadcaster not found"})
			return
		}
		utils.GetLogger().Error("Failed to get broadcaster by ID", err, map[string]interface{}{"broadcaster_id": broadcasterID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get broadcaster"})
		return
	}

	// Display name defaults to broadcaster name, will be overridden by Twitch API if available
	displayName := broadcasterName

	// Get broadcaster stats
	totalClips, totalViews, avgVoteScore, err := h.broadcasterRepo.GetBroadcasterStats(ctx, broadcasterID)
	if err != nil {
		utils.GetLogger().Error("Failed to get broadcaster stats", err, map[string]interface{}{"broadcaster_id": broadcasterID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get broadcaster stats"})
		return
	}

	// Get follower count
	followerCount, err := h.broadcasterRepo.GetFollowerCount(ctx, broadcasterID)
	if err != nil {
		utils.GetLogger().Error("Failed to get follower count", err, map[string]interface{}{"broadcaster_id": broadcasterID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get follower count"})
		return
	}

	// Check if current user is following (if authenticated)
	isFollowing := false
	userID, ok := optionalCommunityUserID(c)
	if !ok {
		return
	}
	if userID != nil {
		isFollowing, err = h.broadcasterRepo.IsFollowing(ctx, *userID, broadcasterID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get follow status"})
			return
		}
	}

	// Fetch fresh Twitch metadata if available
	var avatarURL *string
	var bio *string
	if h.twitchClient != nil {
		users, err := h.twitchClient.GetUsers(ctx, []string{broadcasterID}, nil)
		if err == nil && len(users.Data) > 0 {
			user := users.Data[0]
			avatarURL = &user.ProfileImageURL
			bio = &user.Description
			displayName = user.DisplayName
		}
	}

	profile := models.BroadcasterProfile{
		BroadcasterID:   broadcasterID,
		BroadcasterName: broadcasterName,
		DisplayName:     displayName,
		AvatarURL:       avatarURL,
		Bio:             bio,
		TwitchURL:       "https://twitch.tv/" + broadcasterName,
		TotalClips:      totalClips,
		FollowerCount:   followerCount,
		TotalViews:      totalViews,
		AvgVoteScore:    avgVoteScore,
		IsFollowing:     isFollowing,
		UpdatedAt:       time.Now(),
	}

	c.JSON(http.StatusOK, profile)
}

// FollowBroadcaster allows a user to follow a broadcaster
// POST /api/v1/broadcasters/:id/follow
func (h *BroadcasterHandler) FollowBroadcaster(c *gin.Context) {
	broadcasterID := c.Param("id")
	if broadcasterID == "" || len(broadcasterID) > 128 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "broadcaster_id is required"})
		return
	}

	// Get authenticated user
	userUUID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	ctx := c.Request.Context()

	// Get broadcaster name
	broadcasterName, err := h.broadcasterRepo.GetBroadcasterByID(ctx, broadcasterID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "broadcaster not found"})
			return
		}
		utils.GetLogger().Error("Failed to get broadcaster by ID", err, map[string]interface{}{"broadcaster_id": broadcasterID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get broadcaster"})
		return
	}

	// Follow broadcaster
	if err := h.broadcasterRepo.FollowBroadcaster(ctx, userUUID, broadcasterID, broadcasterName); err != nil {
		utils.GetLogger().Error("Failed to follow broadcaster", err, map[string]interface{}{"user_id": userUUID.String(), "broadcaster_id": broadcasterID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to follow broadcaster"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "successfully followed broadcaster"})
}

// UnfollowBroadcaster allows a user to unfollow a broadcaster
// DELETE /api/v1/broadcasters/:id/follow
func (h *BroadcasterHandler) UnfollowBroadcaster(c *gin.Context) {
	broadcasterID := c.Param("id")
	if broadcasterID == "" || len(broadcasterID) > 128 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "broadcaster_id is required"})
		return
	}

	// Get authenticated user
	userUUID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	ctx := c.Request.Context()

	// Unfollow broadcaster
	if err := h.broadcasterRepo.UnfollowBroadcaster(ctx, userUUID, broadcasterID); err != nil {
		if err.Error() == "follow relationship not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "not following this broadcaster"})
			return
		}
		utils.GetLogger().Error("Failed to unfollow broadcaster", err, map[string]interface{}{"user_id": userUUID.String(), "broadcaster_id": broadcasterID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unfollow broadcaster"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "successfully unfollowed broadcaster"})
}

// ListPopularBroadcasters returns popular broadcasters by clip count
// GET /api/v1/broadcasters/popular
func (h *BroadcasterHandler) ListPopularBroadcasters(c *gin.Context) {
	limit := 15
	if l := c.Query("limit"); l != "" {
		parsed, err := strconv.Atoi(l)
		if err != nil || parsed < 1 || parsed > 50 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 50"})
			return
		}
		limit = parsed
	}

	broadcasters, err := h.broadcasterRepo.ListPopularBroadcasters(c.Request.Context(), limit)
	if err != nil {
		utils.GetLogger().Error("Failed to list popular broadcasters", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list popular broadcasters"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"broadcasters": broadcasters})
}

// GetBroadcasterRankings returns the ranked broadcaster list
// GET /api/v1/broadcasters/rankings
func (h *BroadcasterHandler) GetBroadcasterRankings(c *gin.Context) {
	limit := 20
	offset := 0
	if l := c.Query("limit"); l != "" {
		parsed, err := strconv.Atoi(l)
		if err != nil || parsed < 1 || parsed > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 100"})
			return
		}
		limit = parsed
	}
	if o := c.Query("offset"); o != "" {
		parsed, err := strconv.Atoi(o)
		if err != nil || parsed < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "offset must be a non-negative integer"})
			return
		}
		offset = parsed
	}

	rankings, total, err := h.broadcasterRepo.GetRankedBroadcasters(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get broadcaster rankings"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    rankings,
		"meta": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// RefreshBroadcasterRankings triggers a refresh of the rankings materialized view (admin only)
// POST /api/v1/admin/broadcasters/refresh-rankings
func (h *BroadcasterHandler) RefreshBroadcasterRankings(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	if err := h.rankingRefresher.RefreshRankings(ctx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "Ranking refresh timed out"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh rankings"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Rankings refreshed"})
}

// ListBroadcasterClips returns all clips for a broadcaster
// GET /api/v1/broadcasters/:id/clips
func (h *BroadcasterHandler) ListBroadcasterClips(c *gin.Context) {
	broadcasterID := c.Param("id")
	if broadcasterID == "" || len(broadcasterID) > 128 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "broadcaster_id is required"})
		return
	}

	// Parse pagination parameters
	page := 1
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err != nil || parsed < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page parameter"})
			return
		} else {
			page = parsed
		}
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err != nil || parsed < 1 || parsed > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter (1-100)"})
			return
		} else {
			limit = parsed
		}
	}

	offset := (page - 1) * limit

	// Parse sort parameter
	sort := c.DefaultQuery("sort", "recent")
	validSorts := map[string]bool{"recent": true, "popular": true, "trending": true}
	if !validSorts[sort] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sort parameter (recent, popular, trending)"})
		return
	}

	ctx := c.Request.Context()

	// List clips for broadcaster
	clips, total, err := h.clipRepo.ListClipsByBroadcaster(ctx, broadcasterID, sort, limit, offset)
	if err != nil {
		utils.GetLogger().Error("Failed to list broadcaster clips", err, map[string]interface{}{"broadcaster_id": broadcasterID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list clips"})
		return
	}

	// Calculate pagination metadata
	totalPages := (total + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    clips,
		"meta": gin.H{
			"page":        page,
			"limit":       limit,
			"total_items": total,
			"total_pages": totalPages,
		},
	})
}
