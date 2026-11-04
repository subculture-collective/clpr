package handlers

import (
	"net/http"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		FavoriteGames:       firstNonEmptyStrings(req.TwitchCategories, req.FavoriteGames),
		FollowedStreamers:   firstNonEmptyStrings(req.FollowedCreators, req.FollowedStreamers),
		PreferredCategories: firstNonEmptyStrings(req.PreferredTopics, req.PreferredCategories),
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
	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	if req.TwitchCategories != nil {
		preferences.FavoriteGames = *req.TwitchCategories
	}
	if req.FollowedCreators != nil {
		preferences.FollowedStreamers = *req.FollowedCreators
	}
	if req.PreferredTopics != nil {
		preferences.PreferredCategories = *req.PreferredTopics
	}
	preferences.SyncCreatorFirstAliases()

	// Update preferences
	if err := h.service.UpdateUserPreferences(c.Request.Context(), preferences); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update preferences",
		})
		return
	}

	c.JSON(http.StatusOK, preferences)
}

func firstNonEmptyStrings(primary, legacy []string) []string {
	if len(primary) > 0 {
		return primary
	}
	return legacy
}
