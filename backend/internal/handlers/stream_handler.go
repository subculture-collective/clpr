package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/pkg/twitch"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

// StreamHandler handles stream-related HTTP requests
type StreamHandler struct {
	twitchClient     *twitch.Client
	streamRepo       *repository.StreamRepository
	clipRepo         *repository.ClipRepository
	streamFollowRepo *repository.StreamFollowRepository
	jobService       *services.ClipExtractionJobService
}

// NewStreamHandler creates a new stream handler
func NewStreamHandler(twitchClient *twitch.Client, streamRepo *repository.StreamRepository, clipRepo *repository.ClipRepository, streamFollowRepo *repository.StreamFollowRepository, jobService *services.ClipExtractionJobService) *StreamHandler {
	return &StreamHandler{
		twitchClient:     twitchClient,
		streamRepo:       streamRepo,
		clipRepo:         clipRepo,
		streamFollowRepo: streamFollowRepo,
		jobService:       jobService,
	}
}

// validateStreamerUsername validates a Twitch username
// Twitch usernames: 4-25 characters, alphanumeric + underscore only
func validateStreamerUsername(username string) error {
	if len(username) < 4 || len(username) > 25 {
		return fmt.Errorf("username must be between 4 and 25 characters")
	}
	for _, ch := range username {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_') {
			return fmt.Errorf("username can only contain letters, numbers, and underscores")
		}
	}
	return nil
}

// GetStreamStatus returns stream status for a specific streamer
// GET /api/v1/streams/:streamer
func (h *StreamHandler) GetStreamStatus(c *gin.Context) {
	streamer := c.Param("streamer")
	if streamer == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "streamer username is required"})
		return
	}

	ctx := c.Request.Context()

	// Get stream status from Twitch API (with caching)
	stream, user, err := h.twitchClient.GetStreamStatusByUsername(ctx, streamer)
	if err != nil {
		// Check if user not found
		if apiErr, ok := err.(*twitch.APIError); ok {
			if apiErr.StatusCode == 404 {
				c.JSON(http.StatusNotFound, gin.H{"error": "streamer not found"})
				return
			}
		}
		utils.GetLogger().Error("Failed to get stream status", err, map[string]interface{}{
			"streamer": streamer,
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":    "failed to get stream status",
			"streamer": streamer,
		})
		return
	}

	// Build stream model for database
	dbStream := &models.Stream{
		StreamerUsername: user.Login,
		StreamerUserID:   &user.ID,
		DisplayName:      &user.DisplayName,
		IsLive:           stream != nil,
		ViewerCount:      0,
	}

	if stream != nil {
		// Stream is live
		dbStream.Title = &stream.Title
		dbStream.GameName = &stream.GameName
		dbStream.ViewerCount = stream.ViewerCount
		dbStream.LastWentLive = &stream.StartedAt
	} else {
		// Stream is offline - set last went offline to now if this is a transition
		now := time.Now()
		dbStream.LastWentOffline = &now
	}

	// Persist to database
	if err := h.streamRepo.UpsertStream(ctx, dbStream); err != nil {
		utils.GetLogger().Error("Failed to persist stream to database", err, map[string]interface{}{
			"streamer": streamer,
		})
		// Continue anyway - don't fail the request if DB persistence fails
	}

	// Try to get historical data from database for offline streams
	var lastWentOffline *time.Time
	if !dbStream.IsLive {
		if existingStream, err := h.streamRepo.GetStreamByUsername(ctx, streamer); err == nil {
			lastWentOffline = existingStream.LastWentOffline
		} else if err != sql.ErrNoRows {
			utils.GetLogger().Warn("Failed to get stream from database", map[string]interface{}{
				"streamer": streamer,
				"error":    err.Error(),
			})
		}
	}

	// Build response
	streamInfo := models.StreamInfo{
		StreamerUsername: user.Login,
		IsLive:           stream != nil,
		ViewerCount:      0,
		LastWentOffline:  lastWentOffline,
	}

	if stream != nil {
		// Stream is live
		streamInfo.Title = &stream.Title
		streamInfo.GameName = &stream.GameName
		streamInfo.ViewerCount = stream.ViewerCount
		streamInfo.StartedAt = &stream.StartedAt
		streamInfo.ThumbnailURL = &stream.ThumbnailURL
	}

	c.JSON(http.StatusOK, streamInfo)
}

// CreateClipFromStream creates a clip from a live stream VOD
// POST /api/v1/streams/:streamer/clips
func (h *StreamHandler) CreateClipFromStream(c *gin.Context) {
	streamer := c.Param("streamer")
	if streamer == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "streamer username is required"})
		return
	}

	// Get user ID from context (middleware sets this)
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	// Get user object from context for creator information
	userVal, userExists := c.Get("user")
	if !userExists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	authenticatedUser := userVal.(*models.User)

	// Parse request body
	var req models.ClipFromStreamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// Validate title length explicitly for clearer error messages
	if len(req.Title) < 3 || len(req.Title) > 255 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title must be between 3 and 255 characters"})
		return
	}

	// Validate that end_time is greater than start_time
	if req.EndTime <= req.StartTime {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end time must be greater than start time"})
		return
	}

	// Validate duration
	duration := req.EndTime - req.StartTime
	if duration < 5 || duration > 60 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "clip duration must be between 5 and 60 seconds"})
		return
	}

	// Get stream info to verify it exists
	stream, user, err := h.twitchClient.GetStreamStatusByUsername(ctx, streamer)
	if err != nil {
		utils.GetLogger().Error("Failed to get stream status for clip creation", err, map[string]interface{}{
			"streamer": streamer,
		})
		c.JSON(http.StatusNotFound, gin.H{"error": "stream not found or VOD not available"})
		return
	}

	// Ensure the stream is currently live before allowing clip creation
	if stream == nil {
		utils.GetLogger().Error("Attempt to create clip for non-live stream", nil, map[string]interface{}{
			"streamer": streamer,
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "stream must be live to create a clip"})
		return
	}

	// For this initial implementation, we'll create a clip record in "processing" state
	// In a production implementation, this would trigger an async job to extract the video

	// Generate unique clip ID
	clipID := uuid.New()
	twitchClipID := fmt.Sprintf("stream_%s_%s", streamer, clipID.String())

	// Create clip record
	streamSource := "stream"
	status := "processing"
	now := time.Now()

	// Use placeholder URLs for user-created clips until video processing is complete
	placeholderURL := fmt.Sprintf("/clips/%s", clipID.String())

	clip := &models.Clip{
		ID:                clipID,
		TwitchClipID:      twitchClipID,
		TwitchClipURL:     placeholderURL,
		EmbedURL:          fmt.Sprintf("/clips/%s/embed", clipID.String()),
		Title:             req.Title,
		CreatorName:       authenticatedUser.Username,
		CreatorID:         authenticatedUser.TwitchID,
		BroadcasterName:   user.Login,
		BroadcasterID:     &user.ID,
		Duration:          &duration,
		ViewCount:         0,
		CreatedAt:         now,
		ImportedAt:        now,
		VoteScore:         0,
		CommentCount:      0,
		FavoriteCount:     0,
		IsFeatured:        false,
		IsNSFW:            false,
		IsRemoved:         false,
		IsHidden:          false,
		SubmittedByUserID: &userID,
		SubmittedAt:       &now,
		StreamSource:      &streamSource,
		Status:            &status,
		Quality:           &req.Quality,
		StartTime:         &req.StartTime,
		EndTime:           &req.EndTime,
	}

	// Add stream metadata (stream is guaranteed non-nil from earlier checks)
	clip.GameName = &stream.GameName
	clip.GameID = &stream.GameID
	clip.Language = &stream.Language

	// Insert clip into database using repository method
	err = h.clipRepo.CreateStreamClip(ctx, clip)

	if err != nil {
		utils.GetLogger().Error("Failed to create clip from stream", err, map[string]interface{}{
			"streamer": streamer,
			"user_id":  userID.String(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create clip"})
		return
	}

	// Enqueue FFmpeg extraction job for background processing
	if h.jobService != nil {
		// Construct placeholder VOD URL (in production, this would be the actual stream VOD URL)
		// TODO: Replace with actual Twitch VOD API call to get the real VOD URL
		vodURL := fmt.Sprintf("placeholder://vod/%s/%s", user.ID, stream.ID)

		job := &models.ClipExtractionJob{
			ClipID:    clipID.String(),
			VODURL:    vodURL,
			StartTime: req.StartTime,
			EndTime:   req.EndTime,
			Quality:   req.Quality,
		}

		if err := h.jobService.EnqueueJob(ctx, job); err != nil {
			// Log error but don't fail the request - clip is already created
			utils.GetLogger().Error("Failed to enqueue clip extraction job", err, map[string]interface{}{
				"clip_id":  clipID.String(),
				"streamer": streamer,
				"user_id":  userID.String(),
			})
		}
	} else {
		// Log warning if job service is not configured
		utils.GetLogger().Warn("Clip extraction job service not configured - clip will remain in processing state", map[string]interface{}{
			"clip_id":  clipID.String(),
			"streamer": streamer,
		})
	}

	utils.GetLogger().Info("Clip from stream created", map[string]interface{}{
		"clip_id":  clipID.String(),
		"streamer": streamer,
		"user_id":  userID.String(),
	})

	// Return response
	response := models.ClipFromStreamResponse{
		ClipID: clipID.String(),
		Status: "processing",
	}

	c.JSON(http.StatusCreated, response)
}

// FollowStreamer allows a user to follow a streamer for live notifications
// POST /api/v1/streams/:streamer/follow
func (h *StreamHandler) FollowStreamer(c *gin.Context) {
	streamer := c.Param("streamer")
	if streamer == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "streamer username is required"})
		return
	}

	// Validate streamer username
	if err := validateStreamerUsername(streamer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (middleware sets this)
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	ctx := c.Request.Context()

	// Parse optional request body for notification preferences
	var req models.StreamFollowRequest
	notificationsEnabled := true // Default to enabled
	if err := c.ShouldBindJSON(&req); err == nil && req.NotificationsEnabled != nil {
		notificationsEnabled = *req.NotificationsEnabled
	}
	// Note: ShouldBindJSON returns an error for empty bodies, which is fine - we'll use the default

	// Follow the streamer
	follow, err := h.streamFollowRepo.FollowStreamer(ctx, userID, streamer, notificationsEnabled)
	if err != nil {
		utils.GetLogger().Error("Failed to follow streamer", err, map[string]interface{}{
			"user_id":  userID.String(),
			"streamer": streamer,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to follow streamer"})
		return
	}

	utils.GetLogger().Info("User followed streamer", map[string]interface{}{
		"user_id":  userID.String(),
		"streamer": streamer,
	})

	c.JSON(http.StatusOK, gin.H{
		"following":             true,
		"notifications_enabled": follow.NotificationsEnabled,
		"message":               fmt.Sprintf("Successfully following %s", streamer),
	})
}

// UnfollowStreamer allows a user to unfollow a streamer
// DELETE /api/v1/streams/:streamer/follow
func (h *StreamHandler) UnfollowStreamer(c *gin.Context) {
	streamer := c.Param("streamer")
	if streamer == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "streamer username is required"})
		return
	}

	// Validate streamer username
	if err := validateStreamerUsername(streamer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (middleware sets this)
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	ctx := c.Request.Context()

	// Unfollow the streamer
	err := h.streamFollowRepo.UnfollowStreamer(ctx, userID, streamer)
	if err != nil {
		utils.GetLogger().Error("Failed to unfollow streamer", err, map[string]interface{}{
			"user_id":  userID.String(),
			"streamer": streamer,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unfollow streamer"})
		return
	}

	utils.GetLogger().Info("User unfollowed streamer", map[string]interface{}{
		"user_id":  userID.String(),
		"streamer": streamer,
	})

	c.JSON(http.StatusOK, gin.H{
		"following": false,
		"message":   fmt.Sprintf("Successfully unfollowed %s", streamer),
	})
}

// GetFollowedStreamers returns the list of streamers a user is following
// GET /api/v1/streams/following
func (h *StreamHandler) GetFollowedStreamers(c *gin.Context) {
	// Get user ID from context (middleware sets this)
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	ctx := c.Request.Context()

	// Get pagination parameters
	limit := 50
	offset := 0

	// Get followed streamers
	follows, err := h.streamFollowRepo.GetFollowedStreamers(ctx, userID, limit, offset)
	if err != nil {
		utils.GetLogger().Error("Failed to get followed streamers", err, map[string]interface{}{
			"user_id": userID.String(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get followed streamers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"follows": follows,
		"count":   len(follows),
	})
}

// GetStreamFollowStatus returns whether a user is following a specific streamer
// GET /api/v1/streams/:streamer/follow-status
func (h *StreamHandler) GetStreamFollowStatus(c *gin.Context) {
	streamer := c.Param("streamer")
	if streamer == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "streamer username is required"})
		return
	}

	// Validate streamer username
	if err := validateStreamerUsername(streamer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (middleware sets this)
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	ctx := c.Request.Context()

	// Check if following
	isFollowing, err := h.streamFollowRepo.IsFollowing(ctx, userID, streamer)
	if err != nil {
		utils.GetLogger().Error("Failed to check follow status", err, map[string]interface{}{
			"user_id":  userID.String(),
			"streamer": streamer,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check follow status"})
		return
	}

	response := gin.H{
		"following":             isFollowing,
		"notifications_enabled": false,
	}

	// If following, get the notification preference
	if isFollowing {
		follow, err := h.streamFollowRepo.GetFollow(ctx, userID, streamer)
		if err == nil {
			response["notifications_enabled"] = follow.NotificationsEnabled
		}
	}

	c.JSON(http.StatusOK, response)
}
