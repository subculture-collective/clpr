package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func requiredFeedUserID(c *gin.Context) (uuid.UUID, bool) {
	value, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return uuid.Nil, false
	}
	userID, ok := value.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid authenticated user"})
		return uuid.Nil, false
	}
	routeUserID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return uuid.Nil, false
	}
	if routeUserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return uuid.Nil, false
	}
	return userID, true
}

func optionalFeedUserID(c *gin.Context) (*uuid.UUID, bool) {
	value, exists := c.Get("user_id")
	if !exists {
		return nil, true
	}
	userID, ok := value.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid authenticated user"})
		return nil, false
	}
	return &userID, true
}

func requiredFeedActorID(c *gin.Context) (uuid.UUID, bool) {
	value, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return uuid.Nil, false
	}
	userID, ok := value.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid authenticated user"})
		return uuid.Nil, false
	}
	return userID, true
}

func handleFeedError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, services.ErrFeedNotFound), errors.Is(err, services.ErrFeedClipNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
	case errors.Is(err, services.ErrFeedForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
	case errors.Is(err, services.ErrFeedMembershipMismatch):
		c.JSON(http.StatusConflict, gin.H{"error": "Feed membership conflict"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": fallback})
	}
}

// validateDateFilter validates and normalizes a date string expected to be in ISO 8601 format
func validateDateFilter(dateStr string) (string, error) {
	if dateStr == "" {
		return "", nil
	}
	// Try parsing as RFC3339 (ISO 8601)
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		// Try parsing as date only (YYYY-MM-DD)
		t, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return "", err
		}
	}
	// Return normalized RFC3339 format
	return t.Format(time.RFC3339), nil
}

type FeedHandler struct {
	feedService  *services.FeedService
	authService  *services.AuthService
	voteRepo     *repository.VoteRepository
	favoriteRepo *repository.FavoriteRepository
	userRepo     *repository.UserRepository
}

func NewFeedHandler(
	feedService *services.FeedService,
	authService *services.AuthService,
	voteRepo *repository.VoteRepository,
	favoriteRepo *repository.FavoriteRepository,
	userRepo *repository.UserRepository,
) *FeedHandler {
	return &FeedHandler{
		feedService:  feedService,
		authService:  authService,
		voteRepo:     voteRepo,
		favoriteRepo: favoriteRepo,
		userRepo:     userRepo,
	}
}

// CreateFeed creates a new feed
func (h *FeedHandler) CreateFeed(c *gin.Context) {
	userID, ok := requiredFeedUserID(c)
	if !ok {
		return
	}

	var req models.CreateFeedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	feed, err := h.feedService.CreateFeed(c.Request.Context(), userID, &req)
	if err != nil {
		handleFeedError(c, err, "Failed to create feed")
		return
	}

	c.JSON(http.StatusCreated, feed)
}

// ListUserFeeds lists all feeds for a user
func (h *FeedHandler) ListUserFeeds(c *gin.Context) {
	userIDParam := c.Param("id")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	requestingUserID, ok := optionalFeedUserID(c)
	if !ok {
		return
	}

	feeds, err := h.feedService.GetUserFeeds(c.Request.Context(), userID, requestingUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, feeds)
}

// GetFeed retrieves a feed by ID
func (h *FeedHandler) GetFeed(c *gin.Context) {
	feedIDParam := c.Param("feedId")
	feedID, err := uuid.Parse(feedIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feed ID"})
		return
	}

	requestingUserID, ok := optionalFeedUserID(c)
	if !ok {
		return
	}

	feed, err := h.feedService.GetFeed(c.Request.Context(), feedID, requestingUserID)
	if err != nil {
		handleFeedError(c, err, "Failed to get feed")
		return
	}
	ownerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	if feed.UserID != ownerID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}

	c.JSON(http.StatusOK, feed)
}

// UpdateFeed updates a feed
func (h *FeedHandler) UpdateFeed(c *gin.Context) {
	userID, ok := requiredFeedUserID(c)
	if !ok {
		return
	}

	feedIDParam := c.Param("feedId")
	feedID, err := uuid.Parse(feedIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feed ID"})
		return
	}

	var req models.UpdateFeedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	feed, err := h.feedService.UpdateFeed(c.Request.Context(), feedID, userID, &req)
	if err != nil {
		handleFeedError(c, err, "Failed to update feed")
		return
	}

	c.JSON(http.StatusOK, feed)
}

// DeleteFeed deletes a feed
func (h *FeedHandler) DeleteFeed(c *gin.Context) {
	userID, ok := requiredFeedUserID(c)
	if !ok {
		return
	}

	feedIDParam := c.Param("feedId")
	feedID, err := uuid.Parse(feedIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feed ID"})
		return
	}

	err = h.feedService.DeleteFeed(c.Request.Context(), feedID, userID)
	if err != nil {
		handleFeedError(c, err, "Failed to delete feed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Feed deleted successfully"})
}

// AddClipToFeed adds a clip to a feed
func (h *FeedHandler) AddClipToFeed(c *gin.Context) {
	userID, ok := requiredFeedUserID(c)
	if !ok {
		return
	}

	feedIDParam := c.Param("feedId")
	feedID, err := uuid.Parse(feedIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feed ID"})
		return
	}

	var req models.AddClipToFeedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	feedItem, err := h.feedService.AddClipToFeed(c.Request.Context(), feedID, userID, req.ClipID)
	if err != nil {
		handleFeedError(c, err, "Failed to add clip to feed")
		return
	}

	c.JSON(http.StatusCreated, feedItem)
}

// RemoveClipFromFeed removes a clip from a feed
func (h *FeedHandler) RemoveClipFromFeed(c *gin.Context) {
	userID, ok := requiredFeedUserID(c)
	if !ok {
		return
	}

	feedIDParam := c.Param("feedId")
	feedID, err := uuid.Parse(feedIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feed ID"})
		return
	}

	clipIDParam := c.Param("clipId")
	clipID, err := uuid.Parse(clipIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid clip ID"})
		return
	}

	err = h.feedService.RemoveClipFromFeed(c.Request.Context(), feedID, userID, clipID)
	if err != nil {
		handleFeedError(c, err, "Failed to remove clip from feed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Clip removed from feed successfully"})
}

// GetFeedClips retrieves all clips in a feed
func (h *FeedHandler) GetFeedClips(c *gin.Context) {
	feedIDParam := c.Param("feedId")
	feedID, err := uuid.Parse(feedIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feed ID"})
		return
	}

	requestingUserID, ok := optionalFeedUserID(c)
	if !ok {
		return
	}
	ownerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	clips, err := h.feedService.GetFeedClips(c.Request.Context(), feedID, ownerID, requestingUserID)
	if err != nil {
		handleFeedError(c, err, "Failed to get feed clips")
		return
	}

	c.JSON(http.StatusOK, clips)
}

// ReorderFeedClips reorders clips in a feed
func (h *FeedHandler) ReorderFeedClips(c *gin.Context) {
	userID, ok := requiredFeedUserID(c)
	if !ok {
		return
	}

	feedIDParam := c.Param("feedId")
	feedID, err := uuid.Parse(feedIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feed ID"})
		return
	}

	var req models.ReorderFeedClipsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.feedService.ReorderFeedClips(c.Request.Context(), feedID, userID, req.ClipIDs)
	if err != nil {
		handleFeedError(c, err, "Failed to reorder feed clips")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Clips reordered successfully"})
}

// FollowFeed follows a feed
func (h *FeedHandler) FollowFeed(c *gin.Context) {
	userID, ok := requiredFeedActorID(c)
	if !ok {
		return
	}
	ownerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	feedIDParam := c.Param("feedId")
	feedID, err := uuid.Parse(feedIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feed ID"})
		return
	}

	err = h.feedService.FollowFeed(c.Request.Context(), userID, ownerID, feedID)
	if err != nil {
		handleFeedError(c, err, "Failed to follow feed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Feed followed successfully"})
}

// UnfollowFeed unfollows a feed
func (h *FeedHandler) UnfollowFeed(c *gin.Context) {
	userID, ok := requiredFeedActorID(c)
	if !ok {
		return
	}
	ownerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	feedIDParam := c.Param("feedId")
	feedID, err := uuid.Parse(feedIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feed ID"})
		return
	}

	err = h.feedService.UnfollowFeed(c.Request.Context(), userID, ownerID, feedID)
	if err != nil {
		handleFeedError(c, err, "Failed to unfollow feed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Feed unfollowed successfully"})
}

// DiscoverFeeds retrieves public feeds for discovery
func (h *FeedHandler) DiscoverFeeds(c *gin.Context) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 100"})
		return
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 || offset > 100000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "offset must be between 0 and 100000"})
		return
	}

	feeds, err := h.feedService.DiscoverPublicFeeds(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, feeds)
}

// SearchFeeds searches for public feeds
func (h *FeedHandler) SearchFeeds(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" || len(query) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 100"})
		return
	}

	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 || offset > 100000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "offset must be between 0 and 100000"})
		return
	}

	feeds, err := h.feedService.SearchFeeds(c.Request.Context(), query, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, feeds)
}

// GetFollowingFeed retrieves clips from followed users and broadcasters
// GET /api/v1/feed/following
func (h *FeedHandler) GetFollowingFeed(c *gin.Context) {
	// Get current user ID from auth middleware
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	// Parse pagination and filter parameters
	page := 1
	limit := 20

	if pageStr := c.Query("page"); pageStr != "" {
		parsedPage, err := strconv.Atoi(pageStr)
		if err != nil || parsedPage < 1 || parsedPage > 100000 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "page must be between 1 and 100000"})
			return
		}
		page = parsedPage
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit < 1 || parsedLimit > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 100"})
			return
		}
		limit = parsedLimit
	}

	offset := (page - 1) * limit

	// Get clips from the following feed
	clips, total, err := h.feedService.GetFollowingFeed(c.Request.Context(), userID, limit, offset)
	if err != nil {
		log.Printf("Error retrieving following feed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve following feed"})
		return
	}

	totalPages := (total + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    clips,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// GetFilteredClips handles comprehensive feed filtering with multiple criteria
// GET /api/v1/feeds/clips
// Supports both offset-based (legacy) and cursor-based pagination
func (h *FeedHandler) GetFilteredClips(c *gin.Context) {
	// Parse query parameters
	// Use flat parameter names because the global request validator deliberately
	// rejects bracketed query keys. Keep the bracketed lookups as a compatibility
	// fallback for callers mounted without that middleware.
	games := c.QueryArray("game_id")
	if len(games) == 0 {
		games = c.QueryArray("filter[game]")
	}
	streamers := c.QueryArray("broadcaster_id")
	if len(streamers) == 0 {
		streamers = c.QueryArray("filter[streamer]")
	}
	tags := c.QueryArray("tag")
	if len(tags) == 0 {
		tags = c.QueryArray("filter[tags]")
	}
	dateFrom := c.Query("date_from")
	if dateFrom == "" {
		dateFrom = c.Query("filter[date_from]")
	}
	dateTo := c.Query("date_to")
	if dateTo == "" {
		dateTo = c.Query("filter[date_to]")
	}
	sort := c.DefaultQuery("sort", "trending")
	limit, limitErr := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, offsetErr := strconv.Atoi(c.DefaultQuery("offset", "0"))
	cursor := c.Query("cursor") // Cursor for cursor-based pagination
	if limitErr != nil || limit < 10 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 10 and 100"})
		return
	}
	if offsetErr != nil || offset < 0 || offset > 100000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "offset must be between 0 and 100000"})
		return
	}
	if len(cursor) > 2048 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cursor is too long"})
		return
	}
	if len(games) > 1 || len(streamers) > 1 || len(tags) > 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game, streamer, and tag filters currently accept one value each"})
		return
	}
	for _, values := range [][]string{games, streamers, tags} {
		for _, value := range values {
			if value == "" || len(value) > 100 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "filter values must be between 1 and 100 characters"})
				return
			}
		}
	}

	// Validate and normalize sort parameter
	validSorts := map[string]bool{
		"trending": true, "popular": true, "new": true,
		"top": true, "discussed": true, "hot": true, "rising": true,
	}
	if !validSorts[sort] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sort"})
		return
	}

	// Resolve the viewer before creating the stateless feed-session seed. A new
	// first-page request gets a fresh personalized arrangement; subsequent pages
	// recover the same seed from the cursor.
	userID, ok := optionalFeedUserID(c)
	if !ok {
		return
	}
	shuffleSeed := ""
	if sort == "trending" {
		if cursor != "" {
			decodedCursor, decodeErr := utils.DecodeCursor(cursor)
			if decodeErr != nil || decodedCursor == nil || decodedCursor.ShuffleSeed == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cursor: refresh the feed to start a new session"})
				return
			}
			shuffleSeed = decodedCursor.ShuffleSeed
		} else if userID != nil {
			shuffleSeed = uuid.NewSHA1(*userID, []byte(uuid.NewString())).String()
		} else {
			shuffleSeed = uuid.NewString()
		}
	}

	// Validate date filters to prevent SQL injection
	var validatedDateFrom, validatedDateTo string
	var err error
	if dateFrom != "" {
		validatedDateFrom, err = validateDateFilter(dateFrom)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date_from format. Use ISO 8601 (YYYY-MM-DD or RFC3339)"})
			return
		}
	}
	if dateTo != "" {
		validatedDateTo, err = validateDateFilter(dateTo)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date_to format. Use ISO 8601 (YYYY-MM-DD or RFC3339)"})
			return
		}
	}
	if validatedDateFrom != "" && validatedDateTo != "" {
		fromTime, _ := time.Parse(time.RFC3339, validatedDateFrom)
		toTime, _ := time.Parse(time.RFC3339, validatedDateTo)
		if fromTime.After(toTime) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "date_from must not be after date_to"})
			return
		}
	}

	// Build filters for clip repository
	filters := repository.ClipFilters{
		Sort:                sort,
		UserSubmittedOnly:   false, // Automated and user-submitted clips share the main feed.
		TrendingShuffleSeed: &shuffleSeed,
	}

	// Apply cursor if provided (takes precedence over offset)
	if cursor != "" {
		filters.Cursor = &cursor
		offset = 0 // Ignore offset when using cursor
	}

	// Apply game filters (currently single-select; multi-select requires backend changes)
	if len(games) > 0 {
		filters.GameID = &games[0]
	}

	// Apply streamer filters (currently single-select; multi-select requires backend changes)
	if len(streamers) > 0 {
		filters.BroadcasterID = &streamers[0]
	}

	// Apply tag filters (currently single-select; multi-select requires backend changes)
	if len(tags) > 0 {
		filters.Tag = &tags[0]
	}

	// Apply validated date range filters
	if validatedDateFrom != "" {
		filters.DateFrom = &validatedDateFrom
	}
	if validatedDateTo != "" {
		filters.DateTo = &validatedDateTo
	}

	// Fetch clips using feed service with user data enrichment (fetch limit+1 to check if there are more)
	fetchLimit := limit + 1
	clips, total, err := h.feedService.GetFilteredClipsWithUserData(c.Request.Context(), filters, fetchLimit, offset, userID)
	if err != nil {
		log.Printf("Error fetching filtered clips: %v", err)
		// Check if error is cursor-related (client error)
		if filters.Cursor != nil && *filters.Cursor != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cursor: " + err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve filtered clips"})
		}
		return
	}
	// Determine if there are more results
	hasMore := len(clips) > limit
	if hasMore {
		clips = clips[:limit] // Trim to requested limit
	}

	// Generate next cursor from the last clip
	var nextCursor *string
	if hasMore && len(clips) > 0 {
		lastClip := clips[len(clips)-1]
		var sortValue float64
		switch sort {
		case "trending":
			// Use coalesced value to match query behavior
			if lastClip.TrendingScore != 0 {
				sortValue = lastClip.TrendingScore
			} else {
				// Fallback: estimate trending score if not calculated
				// This matches the COALESCE behavior in the query
				sortValue = float64(lastClip.CreatedAt.Unix())
			}
		case "popular":
			// Use coalesced value to match query behavior
			if lastClip.PopularityIndex != 0 {
				sortValue = float64(lastClip.PopularityIndex)
			} else if lastClip.EngagementCount != 0 {
				sortValue = float64(lastClip.EngagementCount)
			} else {
				// Fallback calculation matching COALESCE
				sortValue = float64(lastClip.ViewCount + lastClip.VoteScore*2 + lastClip.CommentCount*3 + lastClip.FavoriteCount*2)
			}
		case "new":
			sortValue = float64(lastClip.CreatedAt.Unix())
		case "top":
			sortValue = float64(lastClip.VoteScore)
		case "discussed":
			sortValue = float64(lastClip.CommentCount)
		case "hot", "rising":
			sortValue = float64(lastClip.CreatedAt.Unix())
		default:
			sortValue = float64(lastClip.CreatedAt.Unix())
		}
		encodedCursor := utils.EncodeCursorWithShuffleSeed(sort, sortValue, lastClip.ID, lastClip.CreatedAt.Unix(), shuffleSeed)
		nextCursor = &encodedCursor
	}

	// Build pagination response
	paginationResponse := gin.H{
		"limit":    limit,
		"offset":   offset,
		"has_more": hasMore,
	}

	// Only include total and total_pages for offset-based pagination to avoid COUNT(*) overhead with cursors
	if cursor == "" {
		totalPages := (total + limit - 1) / limit
		paginationResponse["total"] = total
		paginationResponse["total_pages"] = totalPages
	}

	// Add cursor to response if generated
	if nextCursor != nil {
		paginationResponse["cursor"] = *nextCursor
	}

	// Return response with filter metadata
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"clips":      clips,
		"pagination": paginationResponse,
		"filters_applied": gin.H{
			"games":     games,
			"streamers": streamers,
			"tags":      tags,
			"date_from": dateFrom,
			"date_to":   dateTo,
			"sort":      sort,
		},
	})
}
