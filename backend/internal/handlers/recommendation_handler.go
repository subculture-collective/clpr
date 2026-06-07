package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
)

// RecommendationHandler handles recommendation endpoints
type RecommendationHandler struct {
	service     *services.RecommendationService
	authService *services.AuthService
}

// NewRecommendationHandler creates a new recommendation handler
func NewRecommendationHandler(
	service *services.RecommendationService,
	authService *services.AuthService,
) *RecommendationHandler {
	return &RecommendationHandler{
		service:     service,
		authService: authService,
	}
}

// GetRecommendations handles GET /api/v1/recommendations/clips
func (h *RecommendationHandler) GetRecommendations(c *gin.Context) {
	// Get user ID from context (from auth middleware)
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user not authenticated",
		})
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "invalid user ID",
		})
		return
	}

	// Parse query parameters
	var req models.RecommendationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Override user_id with authenticated user
	req.UserID = userID

	// Set defaults
	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.Algorithm == "" {
		req.Algorithm = models.AlgorithmHybrid
	}

	// Get recommendations
	response, err := h.service.GetRecommendations(
		c.Request.Context(),
		req.UserID,
		req.Algorithm,
		req.Limit,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get recommendations",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// SubmitFeedback handles POST /api/v1/recommendations/feedback
func (h *RecommendationHandler) SubmitFeedback(c *gin.Context) {
	// Get user ID from context
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user not authenticated",
		})
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "invalid user ID",
		})
		return
	}

	// Parse request body
	var req models.SubmitFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Record interaction based on feedback
	interactionType := models.InteractionTypeLike
	if req.FeedbackType == "negative" {
		interactionType = models.InteractionTypeDislike
	}

	interaction := &models.UserClipInteraction{
		UserID:          userID,
		ClipID:          req.ClipID,
		InteractionType: interactionType,
	}

	if err := h.service.RecordInteraction(c.Request.Context(), interaction); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to record feedback",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "feedback recorded successfully",
	})
}

// GetPreferences handles GET /api/v1/recommendations/preferences
func (h *RecommendationHandler) GetPreferences(c *gin.Context) {
	// Get user ID from context
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user not authenticated",
		})
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "invalid user ID",
		})
		return
	}

	// Get preferences
	preferences, err := h.service.GetUserPreferences(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get preferences",
		})
		return
	}

	c.JSON(http.StatusOK, preferences)
}

// CompleteOnboarding handles POST /api/v1/recommendations/onboarding
func (h *RecommendationHandler) CompleteOnboarding(c *gin.Context) {
	// Get user ID from context
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user not authenticated",
		})
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "invalid user ID",
		})
		return
	}

	// Parse request body
	var req models.OnboardingPreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Validate that at least one preference type is provided
	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Create preferences object
	preferences := &models.UserPreference{
		UserID:              userID,
		FavoriteGames:       req.FavoriteGames,
		FollowedStreamers:   req.FollowedStreamers,
		PreferredCategories: req.PreferredCategories,
		PreferredTags:       req.PreferredTags,
	}

	// Complete onboarding
	if err := h.service.CompleteOnboarding(c.Request.Context(), preferences); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to complete onboarding",
		})
		return
	}

	// Get updated preferences to return
	updatedPreferences, err := h.service.GetUserPreferences(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get updated preferences",
		})
		return
	}

	c.JSON(http.StatusOK, updatedPreferences)
}

// UpdatePreferences handles PUT /api/v1/recommendations/preferences
func (h *RecommendationHandler) UpdatePreferences(c *gin.Context) {
	// Get user ID from context
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user not authenticated",
		})
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "invalid user ID",
		})
		return
	}

	// Parse request body
	var req models.UpdatePreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Get current preferences
	preferences, err := h.service.GetUserPreferences(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get current preferences",
		})
		return
	}

	// Update fields if provided
	if req.FavoriteGames != nil {
		preferences.FavoriteGames = *req.FavoriteGames
	}
	if req.FollowedStreamers != nil {
		preferences.FollowedStreamers = *req.FollowedStreamers
	}
	if req.PreferredCategories != nil {
		preferences.PreferredCategories = *req.PreferredCategories
	}
	if req.PreferredTags != nil {
		preferences.PreferredTags = *req.PreferredTags
	}

	// Update preferences
	if err := h.service.UpdateUserPreferences(c.Request.Context(), preferences); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update preferences",
		})
		return
	}

	c.JSON(http.StatusOK, preferences)
}

// TrackView handles POST /api/v1/recommendations/track-view
func (h *RecommendationHandler) TrackView(c *gin.Context) {
	// Get user ID from context (optional for tracking)
	userIDValue, exists := c.Get("user_id")
	if !exists {
		// Anonymous view tracking could be supported, but we skip for now
		c.JSON(http.StatusOK, gin.H{
			"message": "view not tracked (user not authenticated)",
		})
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "invalid user ID",
		})
		return
	}

	// Parse clip ID from route parameter
	clipIDStr := c.Param("id")
	clipID, err := uuid.Parse(clipIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid clip ID",
		})
		return
	}

	// Parse request body for dwell time (optional)
	var body struct {
		DwellTime *int `json:"dwell_time,omitempty"`
	}
	_ = c.ShouldBindJSON(&body)

	// Record view interaction
	interaction := &models.UserClipInteraction{
		UserID:          userID,
		ClipID:          clipID,
		InteractionType: models.InteractionTypeView,
		DwellTime:       body.DwellTime,
	}

	if err := h.service.RecordInteraction(c.Request.Context(), interaction); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to track view",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "view tracked successfully",
	})
}
