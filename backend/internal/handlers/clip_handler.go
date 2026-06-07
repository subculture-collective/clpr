package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
)

// ClipSyncHandler handles clip sync operations
type ClipSyncHandler struct {
	syncService *services.ClipSyncService
	cfg         *config.Config
}

// NewClipSyncHandler creates a new ClipSyncHandler
func NewClipSyncHandler(syncService *services.ClipSyncService, cfg *config.Config) *ClipSyncHandler {
	return &ClipSyncHandler{
		syncService: syncService,
		cfg:         cfg,
	}
}

// TriggerSync handles manual sync trigger
// POST /admin/sync/clips
func (h *ClipSyncHandler) TriggerSync(c *gin.Context) {
	var req struct {
		GameID          string `json:"game_id"`
		BroadcasterID   string `json:"broadcaster_id"`
		Hours           int    `json:"hours"`
		Limit           int    `json:"limit"`
		Strategy        string `json:"strategy"` // "game", "broadcaster", "trending"
		RespectRotation bool   `json:"respect_rotation"`
		Language        string `json:"language"`
	}

	// Body is optional — all fields have defaults
	_ = c.ShouldBindJSON(&req)

	// Set defaults
	if req.Hours == 0 {
		req.Hours = 24
	}
	if req.Limit == 0 {
		req.Limit = 100
	}
	if req.Strategy == "" {
		if req.GameID != "" {
			req.Strategy = "game"
		} else if req.BroadcasterID != "" {
			req.Strategy = "broadcaster"
		} else {
			req.Strategy = "trending"
		}
	}

	// Execute sync based on strategy
	var stats *services.SyncStats
	var err error

	switch req.Strategy {
	case "game":
		if req.GameID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "game_id is required for game strategy",
			})
			return
		}
		stats, _, err = h.syncService.SyncClipsByGame(c.Request.Context(), req.GameID, req.Hours, req.Limit, &services.SyncClipsByGameOptions{LanguageFilter: req.Language})
	case "broadcaster":
		if req.BroadcasterID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "broadcaster_id is required for broadcaster strategy",
			})
			return
		}
		stats, err = h.syncService.SyncClipsByBroadcaster(c.Request.Context(), req.BroadcasterID, req.Hours, req.Limit, &services.SyncClipsByBroadcasterOptions{LanguageFilter: req.Language})
	case "trending":
		respectRotation := req.RespectRotation
		stats, err = h.syncService.SyncTrendingClips(c.Request.Context(), req.Hours, &services.TrendingSyncOptions{
			ForceResetPagination: !respectRotation,
			LanguageFilter:       req.Language,
		})
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid strategy. Must be 'game', 'broadcaster', or 'trending'",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Sync failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Sync completed",
		"strategy":      req.Strategy,
		"clips_fetched": stats.ClipsFetched,
		"clips_created": stats.ClipsCreated,
		"clips_updated": stats.ClipsUpdated,
		"clips_skipped": stats.ClipsSkipped,
		"errors":        stats.Errors,
		"duration_ms":   stats.EndTime.Sub(stats.StartTime).Milliseconds(),
		"started_at":    stats.StartTime,
		"completed_at":  stats.EndTime,
	})
}

// GetSyncStatus returns the current sync status
// GET /admin/sync/status
func (h *ClipSyncHandler) GetSyncStatus(c *gin.Context) {
	// Get statistics from the clip repository
	// This would be extended with a proper sync status tracking mechanism

	c.JSON(http.StatusOK, gin.H{
		"status":  "ready",
		"message": "Sync service is operational",
	})
}

// RequestClip handles user clip submission
// POST /clips/request
func (h *ClipSyncHandler) RequestClip(c *gin.Context) {
	var req struct {
		ClipURL string `json:"clip_url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "clip_url is required",
		})
		return
	}

	// Fetch and save the clip
	clip, err := h.syncService.FetchClipByURL(c.Request.Context(), req.ClipURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch clip: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Clip added successfully",
		"clip":    clip,
	})
}

// ClipHandler handles clip retrieval operations
type ClipHandler struct {
	clipService *services.ClipService
	authService *services.AuthService
	cdnProvider services.CDNProvider
	jobService  *services.ClipExtractionJobService
}

// NewClipHandler creates a new ClipHandler
func NewClipHandler(clipService *services.ClipService, authService *services.AuthService, opts ...ClipHandlerOption) *ClipHandler {
	handler := &ClipHandler{
		clipService: clipService,
		authService: authService,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(handler)
		}
	}

	return handler
}

// ClipHandlerOption configures optional ClipHandler behavior.
type ClipHandlerOption func(*ClipHandler)

// WithCDNProvider enables CDN-aware responses with retry/backoff and failover headers.
func WithCDNProvider(provider services.CDNProvider) ClipHandlerOption {
	return func(h *ClipHandler) {
		h.cdnProvider = provider
	}
}

// WithClipExtractionJobService enables clip processing backfill requests.
func WithClipExtractionJobService(service *services.ClipExtractionJobService) ClipHandlerOption {
	return func(h *ClipHandler) {
		h.jobService = service
	}
}

// StandardResponse represents a standard API response
type StandardResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo represents error information
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// ListClips handles GET /clips
func (h *ClipHandler) ListClips(c *gin.Context) {
	// Parse query parameters
	sort := c.DefaultQuery("sort", "hot")
	timeframe := c.Query("timeframe")
	gameID := c.Query("game_id")
	broadcasterID := c.Query("broadcaster_id")
	tag := c.Query("tag")
	excludeTagsParam := c.Query("exclude_tags")
	search := c.Query("search")
	language := c.Query("language")
	submittedByUserID := c.Query("submitted_by_user_id")
	top10kStreamers := c.Query("top10k_streamers") == "true"
	// By default, only show user-submitted clips. Set show_all_clips=true to include scraped clips (for discovery)
	showAllClips := c.Query("show_all_clips") == "true"
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "25"))

	// Validate and constrain parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 25
	}

	// Validate submitted_by_user_id as UUID if provided
	if submittedByUserID != "" {
		if _, err := uuid.Parse(submittedByUserID); err != nil {
			c.JSON(http.StatusBadRequest, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "INVALID_UUID",
					Message: "Invalid submitted_by_user_id: must be a valid UUID",
				},
			})
			return
		}
	}

	// Build filters
	filters := repository.ClipFilters{
		Sort:              sort,
		Top10kStreamers:   top10kStreamers,
		UserSubmittedOnly: !showAllClips, // Only show user-submitted unless explicitly requesting all
	}

	if gameID != "" {
		filters.GameID = &gameID
	}
	if broadcasterID != "" {
		filters.BroadcasterID = &broadcasterID
	}
	if tag != "" {
		filters.Tag = &tag
	}
	// Parse exclude_tags as comma-separated list with max limit of 10
	if excludeTagsParam != "" {
		excludeTags := []string{}
		for _, t := range strings.Split(excludeTagsParam, ",") {
			trimmed := strings.TrimSpace(t)
			if trimmed != "" {
				excludeTags = append(excludeTags, trimmed)
			}
			// Limit to prevent abuse
			if len(excludeTags) >= 10 {
				break
			}
		}
		if len(excludeTags) > 0 {
			filters.ExcludeTags = excludeTags
		}
	}
	if search != "" {
		filters.Search = &search
	}
	if language != "" {
		filters.Language = &language
	}
	if timeframe != "" {
		filters.Timeframe = &timeframe
	}
	if submittedByUserID != "" {
		filters.SubmittedByUserID = &submittedByUserID
	}

	// Get user ID if authenticated
	var userID *uuid.UUID
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(uuid.UUID); ok {
			userID = &uid
		}
	}

	// Fetch clips
	clips, total, err := h.clipService.ListClips(c.Request.Context(), filters, page, limit, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to fetch clips",
			},
		})
		return
	}

	// Build pagination metadata
	totalPages := (total + limit - 1) / limit
	meta := PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data:    clips,
		Meta:    meta,
	})
}

// ListScrapedClips handles GET /scraped-clips
// Returns clips that have not been claimed/submitted by any user (submitted_by_user_id IS NULL)
func (h *ClipHandler) ListScrapedClips(c *gin.Context) {
	// Parse query parameters
	sort := c.DefaultQuery("sort", "new")
	timeframe := c.Query("timeframe")
	gameID := c.Query("game_id")
	broadcasterID := c.Query("broadcaster_id")
	tag := c.Query("tag")
	excludeTagsParam := c.Query("exclude_tags")
	search := c.Query("search")
	language := c.Query("language")
	top10kStreamers := c.Query("top10k_streamers") == "true"
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "25"))

	// Validate and constrain parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 25
	}

	// Build filters
	filters := repository.ClipFilters{
		Sort:            sort,
		Top10kStreamers: top10kStreamers,
	}

	if gameID != "" {
		filters.GameID = &gameID
	}
	if broadcasterID != "" {
		filters.BroadcasterID = &broadcasterID
	}
	if tag != "" {
		filters.Tag = &tag
	}
	// Parse exclude_tags as comma-separated list with max limit of 10
	if excludeTagsParam != "" {
		excludeTags := []string{}
		for _, t := range strings.Split(excludeTagsParam, ",") {
			trimmed := strings.TrimSpace(t)
			if trimmed != "" {
				excludeTags = append(excludeTags, trimmed)
			}
			// Limit to prevent abuse
			if len(excludeTags) >= 10 {
				break
			}
		}
		if len(excludeTags) > 0 {
			filters.ExcludeTags = excludeTags
		}
	}
	if search != "" {
		filters.Search = &search
	}
	if language != "" {
		filters.Language = &language
	}
	if timeframe != "" {
		filters.Timeframe = &timeframe
	}

	// Get user ID if authenticated
	var userID *uuid.UUID
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(uuid.UUID); ok {
			userID = &uid
		}
	}

	// Fetch scraped clips only
	clips, total, err := h.clipService.ListScrapedClips(c.Request.Context(), filters, page, limit, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to fetch scraped clips",
			},
		})
		return
	}

	// Build pagination metadata
	totalPages := (total + limit - 1) / limit
	meta := PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data:    clips,
		Meta:    meta,
	})
}

// GetClip handles GET /clips/:id
// Accepts both UUID and Twitch clip ID formats
func (h *ClipHandler) GetClip(c *gin.Context) {
	clipIDParam := c.Param("id")

	// Get user ID if authenticated
	var userID *uuid.UUID
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(uuid.UUID); ok {
			userID = &uid
		}
	}

	var clip *services.ClipWithUserData
	var err error

	// Try to parse as UUID first
	if clipID, parseErr := uuid.Parse(clipIDParam); parseErr == nil {
		// It's a valid UUID, fetch by database ID
		clip, err = h.clipService.GetClip(c.Request.Context(), clipID, userID)
	} else {
		// Not a UUID, treat as Twitch clip ID
		clip, err = h.clipService.GetClipByTwitchID(c.Request.Context(), clipIDParam, userID)
	}

	if err != nil {
		c.JSON(http.StatusNotFound, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "CLIP_NOT_FOUND",
				Message: "Clip not found or has been removed",
			},
		})
		return
	}

	// Attempt to generate CDN URL with retry/backoff and fall back to origin on failure
	if h.cdnProvider != nil {
		cdnURL, cdnErr := h.getCDNURLWithRetry(&clip.Clip)
		if cdnErr != nil {
			h.applyFailoverHeaders(c, cdnErr)
			h.applyOriginFallbackCacheHeaders(c)
			if clip.VideoURL != nil {
				clip.PrimaryCDNURL = clip.VideoURL
			}
		} else if cdnURL != "" {
			clip.PrimaryCDNURL = &cdnURL
			h.applyCDNCacheHeaders(c)
		}
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data:    clip,
	})
}

// GetHLSMasterPlaylist handles GET /video/:clipId/master.m3u8
func (h *ClipHandler) GetHLSMasterPlaylist(c *gin.Context) {
	clipIDParam := c.Param("clipId")

	var clip *services.ClipWithUserData
	var err error

	if clipID, parseErr := uuid.Parse(clipIDParam); parseErr == nil {
		clip, err = h.clipService.GetClip(c.Request.Context(), clipID, nil)
	} else {
		clip, err = h.clipService.GetClipByTwitchID(c.Request.Context(), clipIDParam, nil)
	}

	if err != nil {
		c.JSON(http.StatusNotFound, StandardResponse{
			Success: false,
			Error:   &ErrorInfo{Code: "CLIP_NOT_FOUND", Message: "Clip not found or has been removed"},
		})
		return
	}

	var playlistURL string
	if h.cdnProvider != nil {
		cdnURL, cdnErr := h.getCDNURLWithRetry(&clip.Clip)
		if cdnErr != nil {
			h.applyFailoverHeaders(c, cdnErr)
			h.applyOriginFallbackCacheHeaders(c)
		} else if cdnURL != "" {
			playlistURL = cdnURL
			h.applyCDNCacheHeaders(c)
		}
	}

	if playlistURL == "" && clip.VideoURL != nil {
		playlistURL = *clip.VideoURL
	}

	if playlistURL == "" {
		if clip.TwitchClipURL != "" {
			playlistURL = clip.TwitchClipURL
		} else if clip.EmbedURL != "" {
			playlistURL = clip.EmbedURL
		}
	}

	if playlistURL == "" {
		c.JSON(http.StatusNotFound, StandardResponse{
			Success: false,
			Error:   &ErrorInfo{Code: "CLIP_NOT_FOUND", Message: "Clip not found or has been removed"},
		})
		return
	}

	c.Header("Content-Type", "application/vnd.apple.mpegurl; charset=utf-8")
	c.String(http.StatusOK, "#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-STREAM-INF:BANDWIDTH=1500000\n%s\n", playlistURL)
}

// GetClipProcessingStatus handles GET /clips/:id/processing-status
// Returns live processing status from Redis when available.
func (h *ClipHandler) GetClipProcessingStatus(c *gin.Context) {
	clipIDParam := c.Param("id")

	var clip *services.ClipWithUserData
	var err error

	if clipID, parseErr := uuid.Parse(clipIDParam); parseErr == nil {
		clip, err = h.clipService.GetClip(c.Request.Context(), clipID, nil)
	} else {
		clip, err = h.clipService.GetClipByTwitchID(c.Request.Context(), clipIDParam, nil)
	}

	if err != nil {
		c.JSON(http.StatusNotFound, StandardResponse{
			Success: false,
			Error:   &ErrorInfo{Code: "CLIP_NOT_FOUND", Message: "Clip not found or has been removed"},
		})
		return
	}

	if clip.VideoURL != nil && *clip.VideoURL != "" {
		c.JSON(http.StatusOK, StandardResponse{
			Success: true,
			Data: map[string]interface{}{
				"status": "completed",
			},
		})
		return
	}

	if clip.Status != nil && *clip.Status != "" {
		c.JSON(http.StatusOK, StandardResponse{
			Success: true,
			Data: map[string]interface{}{
				"status": *clip.Status,
			},
		})
		return
	}

	if h.jobService == nil {
		c.JSON(http.StatusOK, StandardResponse{
			Success: true,
			Data: map[string]interface{}{
				"status": "unavailable",
			},
		})
		return
	}

	jobStatus, err := h.jobService.GetJobStatus(c.Request.Context(), clip.ID.String())
	if err != nil {
		c.JSON(http.StatusOK, StandardResponse{
			Success: true,
			Data: map[string]interface{}{
				"status": "not_queued",
			},
		})
		return
	}

	statusValue := "unknown"
	if rawStatus, ok := jobStatus["status"]; ok {
		switch v := rawStatus.(type) {
		case string:
			statusValue = v
		default:
			statusValue = fmt.Sprint(v)
		}
	}

	jobStatus["status"] = statusValue

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: map[string]interface{}{
			"status": statusValue,
			"job":    jobStatus,
		},
	})
}

// RequestClipBackfill handles POST /clips/:id/backfill
// Enqueues a clip processing job when HLS is not yet available.
func (h *ClipHandler) RequestClipBackfill(c *gin.Context) {
	if h.jobService == nil {
		c.JSON(http.StatusServiceUnavailable, StandardResponse{
			Success: false,
			Error: &ErrorInfo{Code: "PROCESSING_UNAVAILABLE", Message: "Clip processing is not configured"},
		})
		return
	}

	clipIDParam := c.Param("id")

	var clip *services.ClipWithUserData
	var err error

	if clipID, parseErr := uuid.Parse(clipIDParam); parseErr == nil {
		clip, err = h.clipService.GetClip(c.Request.Context(), clipID, nil)
	} else {
		clip, err = h.clipService.GetClipByTwitchID(c.Request.Context(), clipIDParam, nil)
	}

	if err != nil {
		c.JSON(http.StatusNotFound, StandardResponse{
			Success: false,
			Error:   &ErrorInfo{Code: "CLIP_NOT_FOUND", Message: "Clip not found or has been removed"},
		})
		return
	}

	if clip.VideoURL != nil && *clip.VideoURL != "" {
		c.JSON(http.StatusOK, StandardResponse{
			Success: true,
			Data: map[string]interface{}{
				"status": "ready",
			},
		})
		return
	}

	if clip.Status != nil && *clip.Status == "processing" {
		c.JSON(http.StatusAccepted, StandardResponse{
			Success: true,
			Data: map[string]interface{}{
				"status": "processing",
			},
		})
		return
	}

	sourceURL := ""
	if clip.VideoURL != nil && *clip.VideoURL != "" {
		sourceURL = *clip.VideoURL
	} else if clip.TwitchClipURL != "" {
		sourceURL = clip.TwitchClipURL
	} else if clip.EmbedURL != "" {
		sourceURL = clip.EmbedURL
	}

	if sourceURL == "" {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error:   &ErrorInfo{Code: "NO_SOURCE_URL", Message: "Clip source URL unavailable"},
		})
		return
	}

	startTime := 0.0
	endTime := 0.0
	if clip.StartTime != nil {
		startTime = *clip.StartTime
	}
	if clip.EndTime != nil {
		endTime = *clip.EndTime
	}
	if endTime <= startTime {
		if clip.Duration != nil && *clip.Duration > 0 {
			endTime = *clip.Duration
		}
	}
	if endTime <= startTime {
		endTime = startTime + 30
	}

	quality := "source"
	if clip.Quality != nil && *clip.Quality != "" {
		quality = *clip.Quality
	}

	job := &models.ClipExtractionJob{
		ClipID:    clip.ID.String(),
		VODURL:    sourceURL,
		StartTime: startTime,
		EndTime:   endTime,
		Quality:   quality,
	}

	if err := h.jobService.EnqueueJob(c.Request.Context(), job); err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error:   &ErrorInfo{Code: "ENQUEUE_FAILED", Message: "Failed to queue clip processing"},
		})
		return
	}

	c.JSON(http.StatusAccepted, StandardResponse{
		Success: true,
		Data: map[string]interface{}{
			"status": "queued",
		},
	})
}

// getCDNURLWithRetry attempts to generate a CDN URL with simple exponential backoff.
func (h *ClipHandler) getCDNURLWithRetry(clip *models.Clip) (string, error) {
	if h.cdnProvider == nil {
		return "", nil
	}

	backoff := 50 * time.Millisecond
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		cdnURL, err := h.cdnProvider.GenerateURL(clip)
		if err == nil && cdnURL != "" {
			return cdnURL, nil
		}

		lastErr = err
		if attempt < 3 {
			time.Sleep(backoff)
			backoff *= 2
		}
	}

	if lastErr == nil {
		lastErr = errors.New("failed to generate CDN URL")
	}

	return "", lastErr
}

// applyFailoverHeaders annotates the response with CDN failover metadata.
func (h *ClipHandler) applyFailoverHeaders(c *gin.Context, err error) {
	reason := "error"
	if errors.Is(err, context.DeadlineExceeded) {
		reason = "timeout"
	}

	c.Header("X-CDN-Failover", "true")
	c.Header("X-CDN-Failover-Reason", reason)
	c.Header("X-CDN-Failover-Service", "origin")
}

// applyOriginFallbackCacheHeaders sets conservative cache headers for origin fallback responses.
func (h *ClipHandler) applyOriginFallbackCacheHeaders(c *gin.Context) {
	c.Header("Cache-Control", "public, max-age=120, stale-while-revalidate=60")
	c.Header("X-CDN-Failover-Service", "origin")
}

// applyCDNCacheHeaders uses provider-specific cache headers when available.
func (h *ClipHandler) applyCDNCacheHeaders(c *gin.Context) {
	if h.cdnProvider == nil {
		return
	}

	for key, val := range h.cdnProvider.GetCacheHeaders() {
		c.Header(key, val)
	}
}

// BatchGetClipMedia handles POST /clips/batch-media
func (h *ClipHandler) BatchGetClipMedia(c *gin.Context) {
	var req struct {
		ClipIDs []string `json:"clip_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "clip_ids array is required",
			},
		})
		return
	}

	// Validate and parse UUIDs
	if len(req.ClipIDs) == 0 {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "EMPTY_REQUEST",
				Message: "clip_ids array cannot be empty",
			},
		})
		return
	}

	if len(req.ClipIDs) > 100 {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "TOO_MANY_CLIPS",
				Message: "Maximum 100 clips per batch request",
			},
		})
		return
	}

	clipIDs := make([]uuid.UUID, 0, len(req.ClipIDs))
	for _, idStr := range req.ClipIDs {
		clipID, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "INVALID_CLIP_ID",
					Message: fmt.Sprintf("Invalid clip ID: %s", idStr),
				},
			})
			return
		}
		clipIDs = append(clipIDs, clipID)
	}

	// Fetch batch media
	mediaInfo, err := h.clipService.BatchGetClipMedia(c.Request.Context(), clipIDs)
	if err != nil {
		// Log the actual error for debugging
		c.Error(err)
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to fetch clip media",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data:    mediaInfo,
	})
}

// VoteOnClip handles POST /clips/:id/vote
func (h *ClipHandler) VoteOnClip(c *gin.Context) {
	clipID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_CLIP_ID",
				Message: "Invalid clip ID format",
			},
		})
		return
	}

	// Get user ID (required)
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "UNAUTHORIZED",
				Message: "Authentication required",
			},
		})
		return
	}
	userID := userIDVal.(uuid.UUID)

	// Parse request body
	var req struct {
		Vote int16 `json:"vote" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
			},
		})
		return
	}

	// Validate vote value
	if req.Vote != -1 && req.Vote != 0 && req.Vote != 1 {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_VOTE",
				Message: "Vote must be -1, 0, or 1",
			},
		})
		return
	}

	// Process vote
	err = h.clipService.VoteOnClip(c.Request.Context(), userID, clipID, req.Vote)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "VOTE_FAILED",
				Message: "Failed to process vote",
			},
		})
		return
	}

	// Fetch updated clip data
	clip, err := h.clipService.GetClip(c.Request.Context(), clipID, &userID)
	if err != nil {
		c.JSON(http.StatusOK, StandardResponse{
			Success: true,
			Data: gin.H{
				"message": "Vote processed successfully",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: gin.H{
			"message":        "Vote processed successfully",
			"vote_score":     clip.VoteScore,
			"upvote_count":   clip.UpvoteCount,
			"downvote_count": clip.DownvoteCount,
			"user_vote":      clip.UserVote,
		},
	})
}

// AddFavorite handles POST /clips/:id/favorite
func (h *ClipHandler) AddFavorite(c *gin.Context) {
	clipID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_CLIP_ID",
				Message: "Invalid clip ID format",
			},
		})
		return
	}

	// Get user ID (required)
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "UNAUTHORIZED",
				Message: "Authentication required",
			},
		})
		return
	}
	userID := userIDVal.(uuid.UUID)

	// Add favorite
	err = h.clipService.AddFavorite(c.Request.Context(), userID, clipID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "FAVORITE_FAILED",
				Message: "Failed to add favorite",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: gin.H{
			"message":      "Clip added to favorites",
			"is_favorited": true,
		},
	})
}

// RemoveFavorite handles DELETE /clips/:id/favorite
func (h *ClipHandler) RemoveFavorite(c *gin.Context) {
	clipID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_CLIP_ID",
				Message: "Invalid clip ID format",
			},
		})
		return
	}

	// Get user ID (required)
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "UNAUTHORIZED",
				Message: "Authentication required",
			},
		})
		return
	}
	userID := userIDVal.(uuid.UUID)

	// Remove favorite
	err = h.clipService.RemoveFavorite(c.Request.Context(), userID, clipID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "UNFAVORITE_FAILED",
				Message: "Failed to remove favorite",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: gin.H{
			"message":      "Clip removed from favorites",
			"is_favorited": false,
		},
	})
}

// GetRelatedClips handles GET /clips/:id/related
func (h *ClipHandler) GetRelatedClips(c *gin.Context) {
	clipID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_CLIP_ID",
				Message: "Invalid clip ID format",
			},
		})
		return
	}

	// Get related clips
	clips, err := h.clipService.GetRelatedClips(c.Request.Context(), clipID, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to fetch related clips",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data:    clips,
	})
}

// UpdateClip handles PUT /clips/:id (admin only)
func (h *ClipHandler) UpdateClip(c *gin.Context) {
	if h.clipService == nil {
		c.JSON(http.StatusNotFound, StandardResponse{
			Success: false,
			Error:   &ErrorInfo{Code: "NOT_FOUND", Message: "Clip not found"},
		})
		return
	}

	clipID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_CLIP_ID",
				Message: "Invalid clip ID format",
			},
		})
		return
	}

	// Parse request body
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
			},
		})
		return
	}

	// Ensure the clip exists before attempting an update
	if _, err := h.clipService.GetClip(c.Request.Context(), clipID, nil); err != nil {
		c.JSON(http.StatusNotFound, StandardResponse{
			Success: false,
			Error:   &ErrorInfo{Code: "NOT_FOUND", Message: "Clip not found"},
		})
		return
	}

	// Update clip
	if err := h.clipService.UpdateClip(c.Request.Context(), clipID, req); err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "UPDATE_FAILED",
				Message: err.Error(),
			},
		})
		return
	}

	// Fetch updated clip
	clip, err := h.clipService.GetClip(c.Request.Context(), clipID, nil)
	if err != nil {
		c.JSON(http.StatusOK, StandardResponse{
			Success: true,
			Data: gin.H{
				"message": "Clip updated successfully",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data:    clip,
	})
}

// DeleteClip handles DELETE /clips/:id (admin only)
func (h *ClipHandler) DeleteClip(c *gin.Context) {
	if h.clipService == nil {
		c.JSON(http.StatusNotFound, StandardResponse{
			Success: false,
			Error:   &ErrorInfo{Code: "NOT_FOUND", Message: "Clip not found"},
		})
		return
	}

	clipID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_CLIP_ID",
				Message: "Invalid clip ID format",
			},
		})
		return
	}

	// Parse request body for reason
	var req struct {
		Reason string `json:"reason"`
	}

	// Body is optional for administrative deletes; use a default reason if not provided
	_ = c.ShouldBindJSON(&req)
	if req.Reason == "" {
		req.Reason = "Removed by admin"
	}

	// Ensure clip exists before attempting delete
	if _, err := h.clipService.GetClip(c.Request.Context(), clipID, nil); err != nil {
		c.JSON(http.StatusNotFound, StandardResponse{
			Success: false,
			Error:   &ErrorInfo{Code: "NOT_FOUND", Message: "Clip not found"},
		})
		return
	}

	// Delete clip
	if err := h.clipService.DeleteClip(c.Request.Context(), clipID, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "DELETE_FAILED",
				Message: "Failed to delete clip",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: gin.H{
			"message": "Clip deleted successfully",
		},
	})
}

// UpdateClipMetadata handles PUT /clips/:id/metadata
// Updates clip metadata (title) - only accessible by creator or admin
func (h *ClipHandler) UpdateClipMetadata(c *gin.Context) {
	// Get clip ID from URL
	clipIDStr := c.Param("id")
	clipID, err := uuid.Parse(clipIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_CLIP_ID",
				Message: "Invalid clip ID format",
			},
		})
		return
	}

	// Get user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "UNAUTHORIZED",
				Message: "Authentication required",
			},
		})
		return
	}

	// Parse request body
	var req models.UpdateClipMetadataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
			},
		})
		return
	}

	// Validate that at least one field is provided
	if req.Title == nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "At least one field must be provided for update",
			},
		})
		return
	}

	// Update metadata
	err = h.clipService.UpdateClipMetadata(c.Request.Context(), userID.(uuid.UUID), clipID, req.Title)
	if err != nil {
		if errors.Is(err, services.ErrUnauthorized) {
			c.JSON(http.StatusForbidden, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "FORBIDDEN",
					Message: err.Error(),
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "UPDATE_FAILED",
				Message: "Failed to update clip metadata",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: gin.H{
			"message": "Clip metadata updated successfully",
		},
	})
}

// UpdateClipVisibility handles PUT /clips/:id/visibility
// Updates clip visibility (hidden status) - only accessible by creator or admin
func (h *ClipHandler) UpdateClipVisibility(c *gin.Context) {
	// Get clip ID from URL
	clipIDStr := c.Param("id")
	clipID, err := uuid.Parse(clipIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_CLIP_ID",
				Message: "Invalid clip ID format",
			},
		})
		return
	}

	// Get user from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "UNAUTHORIZED",
				Message: "Authentication required",
			},
		})
		return
	}

	// Parse request body
	var req models.UpdateClipVisibilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
			},
		})
		return
	}

	// Update visibility
	err = h.clipService.UpdateClipVisibility(c.Request.Context(), userID.(uuid.UUID), clipID, req.IsHidden)
	if err != nil {
		if errors.Is(err, services.ErrUnauthorized) {
			c.JSON(http.StatusForbidden, StandardResponse{
				Success: false,
				Error: &ErrorInfo{
					Code:    "FORBIDDEN",
					Message: err.Error(),
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "UPDATE_FAILED",
				Message: "Failed to update clip visibility",
			},
		})
		return
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data: gin.H{
			"message":   "Clip visibility updated successfully",
			"is_hidden": req.IsHidden,
		},
	})
}

// ListCreatorClips handles GET /creators/:creatorId/clips
// Lists clips for a specific creator
func (h *ClipHandler) ListCreatorClips(c *gin.Context) {
	creatorID := c.Param("creatorId")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "25"))

	// Validate and constrain parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 25
	}

	// Get user ID from context (optional)
	var userID *uuid.UUID
	if uid, exists := c.Get("user_id"); exists {
		id := uid.(uuid.UUID)
		userID = &id
	}

	// List clips
	clips, total, err := h.clipService.ListCreatorClips(c.Request.Context(), creatorID, userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "LIST_FAILED",
				Message: "Failed to list creator clips",
			},
		})
		return
	}

	// Calculate pagination metadata
	totalPages := (total + limit - 1) / limit
	meta := PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data:    clips,
		Meta:    meta,
	})
}
