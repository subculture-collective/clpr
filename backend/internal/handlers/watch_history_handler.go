package handlers

import (
	"context"
	"fmt"
	"net/http"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// WatchHistoryHandler handles watch history related endpoints
type watchHistoryRepository interface {
	IsWatchHistoryEnabled(context.Context, uuid.UUID) (bool, error)
	RecordWatchProgress(context.Context, uuid.UUID, uuid.UUID, int, int, string) error
	GetWatchHistory(context.Context, uuid.UUID, string, int) ([]models.WatchHistoryEntry, error)
	GetResumePosition(context.Context, uuid.UUID, uuid.UUID) (int, bool, error)
	ClearWatchHistory(context.Context, uuid.UUID) error
}

type WatchHistoryHandler struct {
	repo watchHistoryRepository
}

// NewWatchHistoryHandler creates a new watch history handler
func NewWatchHistoryHandler(repo watchHistoryRepository) *WatchHistoryHandler {
	return &WatchHistoryHandler{repo: repo}
}

// RecordWatchProgress records watch progress for a clip
// POST /api/v1/watch-history
func (h *WatchHistoryHandler) RecordWatchProgress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Check if history tracking is enabled
	historyEnabled, err := h.repo.IsWatchHistoryEnabled(c.Request.Context(), userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check history settings"})
		return
	}

	if !historyEnabled {
		c.JSON(http.StatusOK, gin.H{"status": "tracking_disabled"})
		return
	}

	var req models.RecordWatchProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.ProgressSeconds > req.DurationSeconds {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Progress cannot exceed duration"})
		return
	}

	clipID, err := uuid.Parse(req.ClipID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid clip ID"})
		return
	}

	// Record progress
	err = h.repo.RecordWatchProgress(c.Request.Context(), userUUID, clipID, req.ProgressSeconds, req.DurationSeconds, req.SessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record watch progress"})
		return
	}

	// Calculate progress percent and completion
	// Guard against division by zero (validation at line 51 checks min=1, but be defensive)
	var progressPercent float64
	if req.DurationSeconds > 0 {
		progressPercent = float64(req.ProgressSeconds) / float64(req.DurationSeconds) * 100
	}
	completed := progressPercent >= 90.0

	c.JSON(http.StatusOK, gin.H{
		"status":           "recorded",
		"completed":        completed,
		"progress_percent": progressPercent,
	})
}

// GetWatchHistory retrieves watch history for the authenticated user
// GET /api/v1/watch-history?filter=all|completed|in-progress&limit=50
func (h *WatchHistoryHandler) GetWatchHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	filterType := c.DefaultQuery("filter", "all")
	if filterType != "all" && filterType != "completed" && filterType != "in-progress" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter parameter"})
		return
	}
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if _, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
			return
		}
		if limit < 1 || limit > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Limit must be between 1 and 100"})
			return
		}
	}

	history, err := h.repo.GetWatchHistory(c.Request.Context(), userUUID, filterType, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch watch history"})
		return
	}

	if history == nil {
		history = []models.WatchHistoryEntry{}
	}

	c.JSON(http.StatusOK, models.WatchHistoryResponse{
		History: history,
		Total:   len(history),
	})
}

// GetResumePosition gets the resume position for a specific clip
// GET /api/v1/clips/:id/progress
func (h *WatchHistoryHandler) GetResumePosition(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		// If user is not authenticated, return no progress
		c.JSON(http.StatusOK, models.ResumePositionResponse{
			HasProgress:     false,
			ProgressSeconds: 0,
			Completed:       false,
		})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	clipIDStr := c.Param("id")
	clipID, err := uuid.Parse(clipIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid clip ID"})
		return
	}

	progressSeconds, completed, err := h.repo.GetResumePosition(c.Request.Context(), userUUID, clipID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get resume position"})
		return
	}

	// Users should be able to resume from progress even if completed
	hasProgress := progressSeconds > 0

	c.JSON(http.StatusOK, models.ResumePositionResponse{
		HasProgress:     hasProgress,
		ProgressSeconds: progressSeconds,
		Completed:       completed,
	})
}

// ClearWatchHistory clears all watch history for the authenticated user
// DELETE /api/v1/watch-history
func (h *WatchHistoryHandler) ClearWatchHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	err := h.repo.ClearWatchHistory(c.Request.Context(), userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "cleared"})
}
