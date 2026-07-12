package handlers

import (
	"net/http"
	"strconv"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// EngagementHandler handles engagement metrics HTTP requests
type EngagementHandler struct {
	engagementService *services.EngagementService
	authService       *services.AuthService
}

// NewEngagementHandler creates a new engagement handler
func NewEngagementHandler(engagementService *services.EngagementService, authService *services.AuthService) *EngagementHandler {
	return &EngagementHandler{
		engagementService: engagementService,
		authService:       authService,
	}
}

// GetUserEngagementScore returns engagement score for a user
// GET /api/v1/users/:userId/engagement
func (h *EngagementHandler) GetUserEngagementScore(c *gin.Context) {
	userIDStr := c.Param("userId")

	// Allow users to use "me" to get their own score
	var userID uuid.UUID
	var err error

	if userIDStr == "me" {
		// Get user from context (set by auth middleware)
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		user, ok := userInterface.(*models.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user context"})
			return
		}
		userID = user.ID
	} else {
		userID, err = uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
			return
		}
	}

	// Check if requester is authorized to view this user's engagement
	// Only the user themselves or admins can view detailed engagement
	requestUserInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	requestUser, ok := requestUserInterface.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user context"})
		return
	}
	if requestUser.Role != "admin" && requestUser.ID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	score, err := h.engagementService.GetUserEngagementScore(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to calculate engagement score"})
		return
	}

	c.JSON(http.StatusOK, score)
}

// GetPlatformHealthMetrics returns platform-wide health metrics
// GET /api/v1/admin/analytics/health
func (h *EngagementHandler) GetPlatformHealthMetrics(c *gin.Context) {
	// Verify admin role
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, ok := userInterface.(*models.User)
	if !ok || user.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	metrics, err := h.engagementService.GetPlatformHealthMetrics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve platform health metrics"})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetTrendingMetrics returns trending data with week-over-week changes
// GET /api/v1/admin/analytics/trending?metric=dau&days=7
func (h *EngagementHandler) GetTrendingMetrics(c *gin.Context) {
	// Verify admin role
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, ok := userInterface.(*models.User)
	if !ok || user.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	metric := c.DefaultQuery("metric", "dau")
	validMetrics := map[string]bool{"dau": true, "clips": true, "votes": true, "comments": true}
	if !validMetrics[metric] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "metric must be dau, clips, votes, or comments"})
		return
	}
	days, err := strconv.Atoi(c.DefaultQuery("days", "7"))

	if err != nil || days <= 0 || days > 365 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "days must be between 1 and 365"})
		return
	}

	trending, err := h.engagementService.GetTrendingMetrics(c.Request.Context(), metric, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve trending metrics"})
		return
	}

	c.JSON(http.StatusOK, trending)
}

// GetContentEngagementScore returns engagement score for a clip
// GET /api/v1/clips/:id/engagement
func (h *EngagementHandler) GetContentEngagementScore(c *gin.Context) {
	clipIDStr := c.Param("id")
	clipID, err := uuid.Parse(clipIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid clip ID"})
		return
	}

	score, err := h.engagementService.GetContentEngagementScore(c.Request.Context(), clipID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to calculate content engagement score"})
		return
	}

	c.JSON(http.StatusOK, score)
}

// CheckAlerts checks and returns current engagement alerts
// GET /api/v1/admin/analytics/alerts
func (h *EngagementHandler) CheckAlerts(c *gin.Context) {
	// Verify admin role
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, ok := userInterface.(*models.User)
	if !ok || user.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	alerts, err := h.engagementService.CheckAlertThresholds(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check alert thresholds"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"count":  len(alerts),
	})
}
