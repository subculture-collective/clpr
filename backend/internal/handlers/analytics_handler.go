package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
)

// AnalyticsHandler handles analytics-related HTTP requests
type AnalyticsHandler struct {
	analyticsService *services.AnalyticsService
	authService      *services.AuthService
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(analyticsService *services.AnalyticsService, authService *services.AuthService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
		authService:      authService,
	}
}

// GetCreatorAnalyticsOverview returns summary metrics for a creator
// GET /api/v1/creators/:creatorName/analytics/overview
func (h *AnalyticsHandler) GetCreatorAnalyticsOverview(c *gin.Context) {
	creatorName := c.Param("creatorName")
	if creatorName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "creator name is required"})
		return
	}

	overview, err := h.analyticsService.GetCreatorAnalyticsOverview(c.Request.Context(), creatorName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve creator analytics"})
		return
	}

	c.JSON(http.StatusOK, overview)
}

// GetCreatorTopClips returns top-performing clips for a creator
// GET /api/v1/creators/:creatorName/analytics/clips?sort=views&limit=10
func (h *AnalyticsHandler) GetCreatorTopClips(c *gin.Context) {
	creatorName := c.Param("creatorName")
	if creatorName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "creator name is required"})
		return
	}

	sortBy := c.DefaultQuery("sort", "votes")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	clips, err := h.analyticsService.GetCreatorTopClips(c.Request.Context(), creatorName, sortBy, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve top clips"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"clips": clips,
		"count": len(clips),
	})
}

// GetCreatorTrends returns time-series data for creator metrics
// GET /api/v1/creators/:creatorName/analytics/trends?metric=views&days=30
func (h *AnalyticsHandler) GetCreatorTrends(c *gin.Context) {
	creatorName := c.Param("creatorName")
	if creatorName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "creator name is required"})
		return
	}

	metricType := c.DefaultQuery("metric", "clip_views")
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))

	trends, err := h.analyticsService.GetCreatorTrends(c.Request.Context(), creatorName, metricType, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve trends"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metric": metricType,
		"days":   days,
		"data":   trends,
	})
}

// GetClipAnalytics returns analytics for a specific clip
// GET /api/v1/clips/:id/analytics
func (h *AnalyticsHandler) GetClipAnalytics(c *gin.Context) {
	clipIDStr := c.Param("id")
	clipID, err := uuid.Parse(clipIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid clip ID"})
		return
	}

	analytics, err := h.analyticsService.GetClipAnalytics(c.Request.Context(), clipID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "analytics not found for this clip"})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// GetUserStats returns personal statistics for the authenticated user
// GET /api/v1/users/me/stats
func (h *AnalyticsHandler) GetUserStats(c *gin.Context) {
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

	analytics, err := h.analyticsService.GetUserAnalytics(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user statistics not found"})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// GetPlatformOverview returns platform KPIs for admin dashboard
// GET /api/v1/admin/analytics/overview
func (h *AnalyticsHandler) GetPlatformOverview(c *gin.Context) {
	overview, err := h.analyticsService.GetPlatformOverview(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve platform overview"})
		return
	}

	c.JSON(http.StatusOK, overview)
}

// GetContentMetrics returns content-related metrics for admin dashboard
// GET /api/v1/admin/analytics/content
func (h *AnalyticsHandler) GetContentMetrics(c *gin.Context) {
	metrics, err := h.analyticsService.GetContentMetrics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve content metrics"})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetPlatformTrends returns time-series data for platform metrics
// GET /api/v1/admin/analytics/trends?metric=users&days=30
func (h *AnalyticsHandler) GetPlatformTrends(c *gin.Context) {
	metricType := c.DefaultQuery("metric", "users")
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))

	trends, err := h.analyticsService.GetPlatformTrends(c.Request.Context(), metricType, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve platform trends"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metric": metricType,
		"days":   days,
		"data":   trends,
	})
}

// GetCreatorAudienceInsights returns audience insights (geography and devices) for a creator
// GET /api/v1/creators/:creatorName/analytics/audience?limit=10
func (h *AnalyticsHandler) GetCreatorAudienceInsights(c *gin.Context) {
	creatorName := c.Param("creatorName")
	if creatorName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "creator name is required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	insights, err := h.analyticsService.GetCreatorAudienceInsights(c.Request.Context(), creatorName, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve audience insights"})
		return
	}

	c.JSON(http.StatusOK, insights)
}

// TrackClipView tracks a clip view event
// POST /api/v1/clips/:id/track-view
func (h *AnalyticsHandler) TrackClipView(c *gin.Context) {
	clipIDStr := c.Param("id")
	clipID, err := uuid.Parse(clipIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid clip ID"})
		return
	}

	// Get user ID if authenticated
	var userID *uuid.UUID
	if userInterface, exists := c.Get("user"); exists {
		if user, ok := userInterface.(*models.User); ok {
			userID = &user.ID
		}
	}

	// Get request metadata
	metadata := map[string]interface{}{
		"user_agent": c.Request.UserAgent(),
		"referrer":   c.Request.Referer(),
	}

	// Track the view event
	err = h.analyticsService.TrackEvent(
		c.Request.Context(),
		"clip_view",
		userID,
		&clipID,
		metadata,
		c.ClientIP(),
		c.Request.UserAgent(),
		c.Request.Referer(),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to track view"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "view tracked"})
}
