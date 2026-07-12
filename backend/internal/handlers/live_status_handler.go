package handlers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type liveStatusService interface {
	GetLiveStatus(context.Context, string) (*models.BroadcasterLiveStatus, error)
	ListLiveBroadcasters(context.Context, int, int) ([]models.BroadcasterLiveStatus, int, error)
	GetFollowedLiveBroadcasters(context.Context, uuid.UUID) ([]models.BroadcasterLiveStatus, error)
}

// LiveStatusHandler handles live status HTTP requests
type LiveStatusHandler struct {
	liveStatusService liveStatusService
}

// NewLiveStatusHandler creates a new live status handler
func NewLiveStatusHandler(liveStatusService liveStatusService) *LiveStatusHandler {
	return &LiveStatusHandler{liveStatusService: liveStatusService}
}

// GetBroadcasterLiveStatus returns live status for a specific broadcaster
// GET /api/v1/broadcasters/:id/live-status
func (h *LiveStatusHandler) GetBroadcasterLiveStatus(c *gin.Context) {
	broadcasterID := c.Param("id")
	if broadcasterID == "" || len(broadcasterID) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "broadcaster_id is required"})
		return
	}

	ctx := c.Request.Context()

	status, err := h.liveStatusService.GetLiveStatus(ctx, broadcasterID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// If no status record exists, return default offline status
			c.JSON(http.StatusOK, gin.H{
				"broadcaster_id": broadcasterID,
				"is_live":        false,
				"is_stale":       false,
				"viewer_count":   0,
			})
			return
		}
		utils.GetLogger().Error("Failed to get live status", err, map[string]interface{}{"broadcaster_id": broadcasterID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get live status"})
		return
	}

	c.JSON(http.StatusOK, status)
}

// ListLiveBroadcasters returns all currently live broadcasters
// GET /api/v1/broadcasters/live
func (h *LiveStatusHandler) ListLiveBroadcasters(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse pagination parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page must be a positive integer"})
		return
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 100"})
		return
	}

	offset := (page - 1) * limit

	broadcasters, total, err := h.liveStatusService.ListLiveBroadcasters(ctx, limit, offset)
	if err != nil {
		utils.GetLogger().Error("Failed to list live broadcasters", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list live broadcasters"})
		return
	}

	totalPages := (total + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    broadcasters,
		"meta": gin.H{
			"page":        page,
			"limit":       limit,
			"total_items": total,
			"total_pages": totalPages,
		},
	})
}

// GetFollowedLiveBroadcasters returns live broadcasters that the authenticated user follows
// GET /api/v1/feed/live
func (h *LiveStatusHandler) GetFollowedLiveBroadcasters(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID"})
		return
	}

	ctx := c.Request.Context()

	broadcasters, err := h.liveStatusService.GetFollowedLiveBroadcasters(ctx, userID)
	if err != nil {
		utils.GetLogger().Error("Failed to get followed live broadcasters", err, map[string]interface{}{"user_id": userID})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get followed live broadcasters"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    broadcasters,
	})
}
