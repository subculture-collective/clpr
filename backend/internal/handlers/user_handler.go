package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	clipRepo            *repository.ClipRepository
	voteRepo            *repository.VoteRepository
	commentRepo         *repository.CommentRepository
	userRepo            *repository.UserRepository
	broadcasterRepo     *repository.BroadcasterRepository
	accountMergeService *services.AccountMergeService
}

// NewUserHandler creates a new user handler
func NewUserHandler(
	clipRepo *repository.ClipRepository,
	voteRepo *repository.VoteRepository,
	commentRepo *repository.CommentRepository,
	userRepo *repository.UserRepository,
	broadcasterRepo *repository.BroadcasterRepository,
	accountMergeService *services.AccountMergeService,
) *UserHandler {
	return &UserHandler{
		clipRepo:            clipRepo,
		voteRepo:            voteRepo,
		commentRepo:         commentRepo,
		userRepo:            userRepo,
		broadcasterRepo:     broadcasterRepo,
		accountMergeService: accountMergeService,
	}
}

// GetUserByUsername retrieves a user's public profile by username
// GET /api/v1/users/by-username/:username
func (h *UserHandler) GetUserByUsername(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	user, err := h.userRepo.GetByUsername(c.Request.Context(), username)
	if err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve user"})
		return
	}

	// Check if full profile with stats is requested
	fullProfile := c.Query("full") == "true"
	if fullProfile {
		// Get current user ID if authenticated
		var currentUserID *uuid.UUID
		if userIDInterface, exists := c.Get("user_id"); exists {
			if uid, ok := userIDInterface.(uuid.UUID); ok {
				currentUserID = &uid
			}
		}

		profile, err := h.userRepo.GetUserProfile(c.Request.Context(), user.ID, currentUserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve user profile"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    profile,
		})
		return
	}

	// Return only basic public user information
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":           user.ID,
			"username":     user.Username,
			"display_name": user.DisplayName,
			"avatar_url":   user.AvatarURL,
			"bio":          user.Bio,
			"karma_points": user.KarmaPoints,
			"role":         user.Role,
			"created_at":   user.CreatedAt,
		},
	})
}

// SearchUsersAutocomplete searches users by username for autocomplete/mention suggestions
// GET /api/v1/users/autocomplete?q=username_prefix
func (h *UserHandler) SearchUsersAutocomplete(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []gin.H{},
		})
		return
	}

	// Validate query length to prevent performance issues
	if len(query) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "query parameter must be 50 characters or less",
		})
		return
	}

	// Parse limit with default of 10 and max of 20
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 {
		limit = 10
	} else if limit > 20 {
		limit = 20
	}

	users, err := h.userRepo.SearchUsersForAutocomplete(c.Request.Context(), query, limit)
	if err != nil {
		log.Printf("Failed to search users for autocomplete: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search users"})
		return
	}

	// Format response with only necessary fields for autocomplete
	suggestions := make([]gin.H, 0, len(users))
	for _, user := range users {
		suggestions = append(suggestions, gin.H{
			"id":           user.ID,
			"username":     user.Username,
			"display_name": user.DisplayName,
			"avatar_url":   user.AvatarURL,
			"is_verified":  user.IsVerified,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    suggestions,
	})
}

// GetUserComments retrieves comments by a user
// GET /api/v1/users/:id/comments
func (h *UserHandler) GetUserComments(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Get user comments
	comments, total, err := h.commentRepo.ListByUserID(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve comments"})
		return
	}

	// Transform to response format
	type CommentResponse struct {
		ID            string  `json:"id"`
		ClipID        string  `json:"clip_id"`
		UserID        string  `json:"user_id"`
		Username      string  `json:"username"`
		UserAvatar    *string `json:"user_avatar"`
		UserKarma     int     `json:"user_karma"`
		UserRole      string  `json:"user_role"`
		ParentID      *string `json:"parent_id"`
		Content       string  `json:"content"`
		VoteScore     int     `json:"vote_score"`
		CreatedAt     string  `json:"created_at"`
		UpdatedAt     string  `json:"updated_at"`
		IsDeleted     bool    `json:"is_deleted"`
		IsRemoved     bool    `json:"is_removed"`
		RemovedReason *string `json:"removed_reason"`
		Depth         int     `json:"depth"`
		ChildCount    int     `json:"child_count"`
		UserVote      *int16  `json:"user_vote"`
	}

	responses := make([]CommentResponse, len(comments))
	for i, comment := range comments {
		var parentID *string
		if comment.ParentCommentID != nil {
			pid := comment.ParentCommentID.String()
			parentID = &pid
		}

		responses[i] = CommentResponse{
			ID:            comment.ID.String(),
			ClipID:        comment.ClipID.String(),
			UserID:        comment.UserID.String(),
			Username:      comment.AuthorUsername,
			UserAvatar:    comment.AuthorAvatarURL,
			UserKarma:     comment.AuthorKarma,
			UserRole:      comment.AuthorRole,
			ParentID:      parentID,
			Content:       comment.Content,
			VoteScore:     comment.VoteScore,
			CreatedAt:     comment.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:     comment.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			IsDeleted:     false,
			IsRemoved:     comment.IsRemoved,
			RemovedReason: comment.RemovedReason,
			Depth:         0,
			ChildCount:    comment.ReplyCount,
			UserVote:      comment.UserVote,
		}
	}

	totalPages := (total + limit - 1) / limit
	c.JSON(http.StatusOK, gin.H{
		"comments": responses,
		"total":    total,
		"page":     page,
		"limit":    limit,
		"has_more": page < totalPages,
	})
}

// GetUserUpvotedClips retrieves clips that a user has upvoted
// GET /api/v1/users/:id/upvoted
func (h *UserHandler) GetUserUpvotedClips(c *gin.Context) {
	h.getUserVotedClips(c, 1)
}

// GetUserDownvotedClips retrieves clips that a user has downvoted
// GET /api/v1/users/:id/downvoted
func (h *UserHandler) GetUserDownvotedClips(c *gin.Context) {
	h.getUserVotedClips(c, -1)
}

// getUserVotedClips is a helper function to retrieve clips that a user has voted on
func (h *UserHandler) getUserVotedClips(c *gin.Context, voteType int16) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Get voted clip IDs
	clipIDs, total, err := h.voteRepo.GetUserVotedClips(c.Request.Context(), userID, voteType, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve voted clips"})
		return
	}

	// Get clip details
	clips := []models.Clip{}
	if len(clipIDs) > 0 {
		clips, err = h.clipRepo.GetByIDs(c.Request.Context(), clipIDs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve clip details"})
			return
		}
	}

	// Transform to response format
	type ClipResponse struct {
		ID              string   `json:"id"`
		TwitchClipID    string   `json:"twitch_clip_id"`
		TwitchClipURL   string   `json:"twitch_clip_url"`
		EmbedURL        string   `json:"embed_url"`
		Title           string   `json:"title"`
		CreatorName     string   `json:"creator_name"`
		CreatorID       *string  `json:"creator_id"`
		BroadcasterName string   `json:"broadcaster_name"`
		BroadcasterID   *string  `json:"broadcaster_id"`
		GameID          *string  `json:"game_id"`
		GameName        *string  `json:"game_name"`
		Language        *string  `json:"language"`
		ThumbnailURL    *string  `json:"thumbnail_url"`
		Duration        *float64 `json:"duration"`
		ViewCount       int      `json:"view_count"`
		CreatedAt       string   `json:"created_at"`
		ImportedAt      string   `json:"imported_at"`
		VoteScore       int      `json:"vote_score"`
		CommentCount    int      `json:"comment_count"`
		FavoriteCount   int      `json:"favorite_count"`
		IsFeatured      bool     `json:"is_featured"`
		IsNSFW          bool     `json:"is_nsfw"`
		IsRemoved       bool     `json:"is_removed"`
		RemovedReason   *string  `json:"removed_reason"`
		UserVote        *int16   `json:"user_vote"`
	}

	responses := make([]ClipResponse, len(clips))
	for i, clip := range clips {
		userVote := voteType
		responses[i] = ClipResponse{
			ID:              clip.ID.String(),
			TwitchClipID:    clip.TwitchClipID,
			TwitchClipURL:   clip.TwitchClipURL,
			EmbedURL:        clip.EmbedURL,
			Title:           clip.Title,
			CreatorName:     clip.CreatorName,
			CreatorID:       clip.CreatorID,
			BroadcasterName: clip.BroadcasterName,
			BroadcasterID:   clip.BroadcasterID,
			GameID:          clip.GameID,
			GameName:        clip.GameName,
			Language:        clip.Language,
			ThumbnailURL:    clip.ThumbnailURL,
			Duration:        clip.Duration,
			ViewCount:       clip.ViewCount,
			CreatedAt:       clip.CreatedAt.Format("2006-01-02T15:04:05Z"),
			ImportedAt:      clip.ImportedAt.Format("2006-01-02T15:04:05Z"),
			VoteScore:       clip.VoteScore,
			CommentCount:    clip.CommentCount,
			FavoriteCount:   clip.FavoriteCount,
			IsFeatured:      clip.IsFeatured,
			IsNSFW:          clip.IsNSFW,
			IsRemoved:       clip.IsRemoved,
			RemovedReason:   clip.RemovedReason,
			UserVote:        &userVote,
		}
	}

	totalPages := (total + limit - 1) / limit
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    responses,
		"meta": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
			"has_next":    page < totalPages,
			"has_prev":    page > 1,
		},
	})
}

// GetUserProfile retrieves a user's complete profile with stats
// GET /api/v1/users/:id
func (h *UserHandler) GetUserProfile(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Get current user ID if authenticated
	var currentUserID *uuid.UUID
	if userIDInterface, exists := c.Get("user_id"); exists {
		if uid, ok := userIDInterface.(uuid.UUID); ok {
			currentUserID = &uid
		}
	}

	profile, err := h.userRepo.GetUserProfile(c.Request.Context(), userID, currentUserID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve user profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    profile,
	})
}

// GetUserClips retrieves clips submitted by a user
// GET /api/v1/users/:id/clips
func (h *UserHandler) GetUserClips(c *gin.Context) {
	userIDStr := c.Param("id")
	_, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Get current user to check permissions
	var currentUserID *uuid.UUID
	if userIDInterface, exists := c.Get("user_id"); exists {
		if uid, ok := userIDInterface.(uuid.UUID); ok {
			currentUserID = &uid
		}
	}

	// Build filters with submitted_by_user_id
	filters := repository.ClipFilters{
		SubmittedByUserID: &userIDStr,
		ShowHidden:        currentUserID != nil, // Show hidden clips if authenticated
	}

	clips, total, err := h.clipRepo.ListWithFilters(c.Request.Context(), filters, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve clips"})
		return
	}

	totalPages := (total + limit - 1) / limit
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    clips,
		"meta": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
			"has_next":    page < totalPages,
			"has_prev":    page > 1,
		},
	})
}

// GetUserActivity retrieves a user's activity feed
// GET /api/v1/users/:id/activity
func (h *UserHandler) GetUserActivity(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	activities, total, err := h.userRepo.GetUserActivity(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve user activity"})
		return
	}

	totalPages := (total + limit - 1) / limit
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    activities,
		"meta": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
			"has_next":    page < totalPages,
			"has_prev":    page > 1,
		},
	})
}

// GetUserFollowers retrieves users who follow the specified user
// GET /api/v1/users/:id/followers
func (h *UserHandler) GetUserFollowers(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Get current user ID if authenticated
	var currentUserID *uuid.UUID
	if userIDInterface, exists := c.Get("user_id"); exists {
		if uid, ok := userIDInterface.(uuid.UUID); ok {
			currentUserID = &uid
		}
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	followers, total, err := h.userRepo.GetFollowers(c.Request.Context(), userID, currentUserID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve followers"})
		return
	}

	totalPages := (total + limit - 1) / limit
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    followers,
		"meta": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
			"has_next":    page < totalPages,
			"has_prev":    page > 1,
		},
	})
}

// GetUserFollowing retrieves users that the specified user follows
// GET /api/v1/users/:id/following
func (h *UserHandler) GetUserFollowing(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Get current user ID if authenticated
	var currentUserID *uuid.UUID
	if userIDInterface, exists := c.Get("user_id"); exists {
		if uid, ok := userIDInterface.(uuid.UUID); ok {
			currentUserID = &uid
		}
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	following, total, err := h.userRepo.GetFollowing(c.Request.Context(), userID, currentUserID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve following"})
		return
	}

	totalPages := (total + limit - 1) / limit
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    following,
		"meta": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
			"has_next":    page < totalPages,
			"has_prev":    page > 1,
		},
	})
}

// FollowUser creates a follow relationship
// POST /api/v1/users/:id/follow
func (h *UserHandler) FollowUser(c *gin.Context) {
	userIDStr := c.Param("id")
	followingID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Get current user ID
	followerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	followerUUID := followerID.(uuid.UUID)

	// Can't follow yourself
	if followerUUID == followingID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot follow yourself"})
		return
	}

	// Create follow relationship
	err = h.userRepo.FollowUser(c.Request.Context(), followerUUID, followingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to follow user"})
		return
	}

	// Record activity
	activity := &models.UserActivity{
		ID:           uuid.New(),
		UserID:       followerUUID,
		ActivityType: models.ActivityTypeUserFollowed,
		TargetID:     &followingID,
		TargetType:   strPtr("user"),
	}
	if err := h.userRepo.CreateUserActivity(c.Request.Context(), activity); err != nil {
		log.Printf("Warning: Failed to record follow activity for user %s: %v", followerUUID, err)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "user followed successfully",
	})
}

// UnfollowUser removes a follow relationship
// DELETE /api/v1/users/:id/follow
func (h *UserHandler) UnfollowUser(c *gin.Context) {
	userIDStr := c.Param("id")
	followingID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Get current user ID
	followerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	followerUUID := followerID.(uuid.UUID)

	// Remove follow relationship
	err = h.userRepo.UnfollowUser(c.Request.Context(), followerUUID, followingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unfollow user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "user unfollowed successfully",
	})
}

// Helper function
func strPtr(s string) *string {
	return &s
}

// BlockUser creates a block relationship
// POST /api/v1/users/:id/block
func (h *UserHandler) BlockUser(c *gin.Context) {
	// Get target user ID from URL parameter
	targetUserIDStr := c.Param("id")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format"})
		return
	}

	// Get current user ID from auth middleware
	blockerIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	blockerUUID := blockerIDInterface.(uuid.UUID)

	// Prevent self-blocking
	if blockerUUID == targetUserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot block yourself"})
		return
	}

	// Check if target user exists
	_, err = h.userRepo.GetByID(c.Request.Context(), targetUserID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find user"})
		return
	}

	// Create block relationship

	// Check if target user has already blocked the current user
	isBlocked, err := h.userRepo.IsBlocked(c.Request.Context(), targetUserID, blockerUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check block status"})
		return
	}

	err = h.userRepo.BlockUser(c.Request.Context(), blockerUUID, targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to block user"})
		return
	}

	// If they were following each other, remove the follow relationships
	_ = h.userRepo.UnfollowUser(c.Request.Context(), blockerUUID, targetUserID)
	_ = h.userRepo.UnfollowUser(c.Request.Context(), targetUserID, blockerUUID)

	response := gin.H{
		"success": true,
		"message": "user blocked successfully",
	}

	// If there's a mutual block, add that to the response
	if isBlocked {
		response["note"] = "both users have blocked each other"
	}

	c.JSON(http.StatusOK, response)
}

// UnblockUser removes a block relationship
// DELETE /api/v1/users/:id/block
func (h *UserHandler) UnblockUser(c *gin.Context) {
	// Get target user ID from URL parameter
	targetUserIDStr := c.Param("id")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format"})
		return
	}

	// Get current user ID from auth middleware
	blockerIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	blockerUUID := blockerIDInterface.(uuid.UUID)

	// Remove block relationship
	err = h.userRepo.UnblockUser(c.Request.Context(), blockerUUID, targetUserID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "block relationship not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unblock user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "user unblocked successfully",
	})
}

// GetBlockedUsers retrieves users blocked by the current user
// GET /api/v1/users/me/blocked
func (h *UserHandler) GetBlockedUsers(c *gin.Context) {
	// Get current user ID from auth middleware
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userUUID := userIDInterface.(uuid.UUID)

	// Parse pagination parameters
	limit := 20
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get blocked users
	blockedUsers, total, err := h.userRepo.GetBlockedUsers(c.Request.Context(), userUUID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve blocked users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    blockedUsers,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"total":  total,
		},
	})
}

// GetFollowedBroadcasters retrieves broadcasters followed by the specified user
// GET /api/v1/users/:id/following/broadcasters
func (h *UserHandler) GetFollowedBroadcasters(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID format"})
		return
	}

	// Check if user exists
	_, err = h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find user"})
		return
	}

	// Parse pagination parameters
	limit := 20
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get followed broadcasters
	follows, total, err := h.broadcasterRepo.ListUserFollows(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve followed broadcasters"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    follows,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"total":  total,
		},
	})
}

// ClaimAccountRequest represents the request to claim an unclaimed account
type ClaimAccountRequest struct {
	TwitchID string `json:"twitch_id" binding:"required"`
}

// ClaimAccount allows a user to claim an unclaimed profile
// POST /api/v1/users/claim-account
func (h *UserHandler) ClaimAccount(c *gin.Context) {
	// Get authenticated user
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	authenticatedUserID := userIDValue.(uuid.UUID)

	// Parse request
	var req ClaimAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Get authenticated user
	authenticatedUser, err := h.userRepo.GetByID(c.Request.Context(), authenticatedUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve user"})
		return
	}

	// Verify the authenticated user's Twitch ID matches the claim request
	if authenticatedUser.TwitchID == nil || *authenticatedUser.TwitchID != req.TwitchID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only claim accounts matching your Twitch ID"})
		return
	}

	// Find the unclaimed account
	unclaimedUser, err := h.userRepo.GetByTwitchID(c.Request.Context(), req.TwitchID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "no unclaimed account found for this Twitch ID"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find unclaimed account"})
		return
	}

	// Verify the account is unclaimed
	if unclaimedUser.AccountStatus != "unclaimed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "this account is not unclaimed"})
		return
	}

	// Prevent claiming if the authenticated user already has an active account
	if authenticatedUser.AccountStatus == "active" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "you already have an active account"})
		return
	}

	// Transfer data from unclaimed account to authenticated user
	// Update the authenticated user with data from the unclaimed account
	if err := h.userRepo.UpdateDisplayName(c.Request.Context(), authenticatedUserID, unclaimedUser.DisplayName); err != nil {
		log.Printf("Warning: failed to update display name during claim: %v", err)
	}

	// Update account status to active
	if err := h.userRepo.UpdateAccountStatus(c.Request.Context(), authenticatedUserID, "active"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to activate account"})
		return
	}

	// Perform account merge: transfer all data from unclaimed to authenticated account
	if h.accountMergeService != nil {
		mergeResult, err := h.accountMergeService.MergeAccounts(c.Request.Context(), unclaimedUser.ID, authenticatedUserID)
		if err != nil {
			log.Printf("Error: account merge failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to merge account data"})
			return
		}

		log.Printf("Account merge completed: clips=%d, votes=%d, favorites=%d, comments=%d, follows=%d, watch_history=%d, duplicates_skipped=%d",
			mergeResult.ClipsMerged,
			mergeResult.VotesMerged,
			mergeResult.FavoritesMerged,
			mergeResult.CommentsMerged,
			mergeResult.FollowsMerged,
			mergeResult.WatchHistoryMerged,
			mergeResult.DuplicatesSkipped,
		)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "account claimed successfully",
			"merge_stats": gin.H{
				"clips_merged":         mergeResult.ClipsMerged,
				"votes_merged":         mergeResult.VotesMerged,
				"favorites_merged":     mergeResult.FavoritesMerged,
				"comments_merged":      mergeResult.CommentsMerged,
				"follows_merged":       mergeResult.FollowsMerged,
				"watch_history_merged": mergeResult.WatchHistoryMerged,
				"duplicates_skipped":   mergeResult.DuplicatesSkipped,
			},
		})
		return
	}

	// Fallback if merge service is not available
	// Just mark the unclaimed account status as pending to prevent reuse
	if err := h.userRepo.UpdateAccountStatus(c.Request.Context(), unclaimedUser.ID, "pending"); err != nil {
		log.Printf("Warning: failed to mark unclaimed account as pending: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "account claimed successfully",
	})
}
