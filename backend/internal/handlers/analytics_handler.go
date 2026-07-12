package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type analyticsService interface {
	GetCreatorAnalyticsOverview(context.Context, string) (*models.CreatorAnalyticsOverview, error)
	GetCreatorTopClips(context.Context, string, string, int) ([]models.CreatorTopClip, error)
	GetCreatorTrends(context.Context, string, string, int) ([]models.TrendDataPoint, error)
	GetCreatorAudienceInsights(context.Context, string, int) (*models.CreatorAudienceInsights, error)
	GetClipAnalytics(context.Context, uuid.UUID) (*models.ClipAnalytics, error)
	GetUserAnalytics(context.Context, uuid.UUID) (*models.UserAnalytics, error)
	GetPlatformOverview(context.Context) (*models.PlatformOverviewMetrics, error)
	GetContentMetrics(context.Context) (*models.ContentMetrics, error)
	GetPlatformTrends(context.Context, string, int) ([]models.TrendDataPoint, error)
	TrackEvent(context.Context, string, *uuid.UUID, *uuid.UUID, map[string]interface{}, string, string, string) error
}

// AnalyticsHandler handles analytics-related HTTP requests
type AnalyticsHandler struct {
	analyticsService analyticsService
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(analyticsService analyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsService: analyticsService}
}

func creatorAnalyticsName(c *gin.Context) (string, bool) {
	name := strings.TrimSpace(c.Param("creator"))
	if name == "" || len(name) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid creator name"})
		return "", false
	}
	return name, true
}

func boundedAnalyticsInt(c *gin.Context, key string, defaultValue, maximum int) (int, bool) {
	raw, exists := c.GetQuery(key)
	if !exists {
		return defaultValue, true
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 || value > maximum {
		c.JSON(http.StatusBadRequest, gin.H{"error": key + " must be an integer between 1 and " + strconv.Itoa(maximum)})
		return 0, false
	}
	return value, true
}

// GetCreatorAnalyticsOverview returns summary metrics for a creator
// GET /api/v1/creators/:creatorName/analytics/overview
func (h *AnalyticsHandler) GetCreatorAnalyticsOverview(c *gin.Context) {
	creatorName, ok := creatorAnalyticsName(c)
	if !ok {
		return
	}

	overview, err := h.analyticsService.GetCreatorAnalyticsOverview(c.Request.Context(), creatorName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "creator analytics not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve creator analytics"})
		}
		return
	}

	c.JSON(http.StatusOK, overview)
}

// GetCreatorTopClips returns top-performing clips for a creator
// GET /api/v1/creators/:creatorName/analytics/clips?sort=views&limit=10
func (h *AnalyticsHandler) GetCreatorTopClips(c *gin.Context) {
	creatorName, ok := creatorAnalyticsName(c)
	if !ok {
		return
	}

	sortBy := c.DefaultQuery("sort", "votes")
	if sortBy != "views" && sortBy != "votes" && sortBy != "comments" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sort must be one of views, votes, or comments"})
		return
	}
	limit, ok := boundedAnalyticsInt(c, "limit", 10, 100)
	if !ok {
		return
	}

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
	creatorName, ok := creatorAnalyticsName(c)
	if !ok {
		return
	}

	metricType := c.DefaultQuery("metric", "clip_views")
	if metricType != "clip_views" && metricType != "votes" && metricType != "comments" && metricType != "favorites" && metricType != "shares" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported creator metric"})
		return
	}
	days, ok := boundedAnalyticsInt(c, "days", 30, 365)
	if !ok {
		return
	}

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
	validMetrics := map[string]bool{"users": true, "clips": true, "views": true, "votes": true, "comments": true}
	if !validMetrics[metricType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "metric must be users, clips, views, votes, or comments"})
		return
	}
	days, err := strconv.Atoi(c.DefaultQuery("days", "30"))
	if err != nil || days < 1 || days > 365 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "days must be between 1 and 365"})
		return
	}

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
	creatorName, ok := creatorAnalyticsName(c)
	if !ok {
		return
	}

	limit, ok := boundedAnalyticsInt(c, "limit", 10, 50)
	if !ok {
		return
	}

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
