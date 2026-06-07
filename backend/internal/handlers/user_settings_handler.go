package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
)

// UserSettingsHandler handles user settings endpoints
type UserSettingsHandler struct {
	userSettingsService *services.UserSettingsService
	authService         *services.AuthService
}

// NewUserSettingsHandler creates a new user settings handler
func NewUserSettingsHandler(userSettingsService *services.UserSettingsService, authService *services.AuthService) *UserSettingsHandler {
	return &UserSettingsHandler{
		userSettingsService: userSettingsService,
		authService:         authService,
	}
}

// UpdateProfile handles PUT /api/v1/users/me/profile
func (h *UserSettingsHandler) UpdateProfile(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	// Parse request body
	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Update profile
	err := h.userSettingsService.UpdateProfile(c.Request.Context(), userID, req.DisplayName, req.Bio)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update profile",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
	})
}

// GetSettings handles GET /api/v1/users/me/settings
func (h *UserSettingsHandler) GetSettings(c *gin.Context) {
	// Get user ID from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	// Get settings
	settings, err := h.userSettingsService.GetSettings(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get settings",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"settings": settings,
	})
}

// UpdateSettings handles PUT /api/v1/users/me/settings
func (h *UserSettingsHandler) UpdateSettings(c *gin.Context) {
	// Get user ID from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	// Parse request body
	var req models.UpdateUserSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Update settings
	err := h.userSettingsService.UpdateSettings(c.Request.Context(), userID, req.ProfileVisibility, req.ShowKarmaPublicly)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Settings updated successfully",
	})
}

// ExportData handles GET /api/v1/users/me/export
func (h *UserSettingsHandler) ExportData(c *gin.Context) {
	// Get user ID from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	// Export data
	data, err := h.userSettingsService.ExportUserData(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to export data",
		})
		return
	}

	// Set headers for download
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", "attachment; filename=clpr_user_data_export.zip")
	c.Data(http.StatusOK, "application/zip", data)
}

// RequestAccountDeletion handles POST /api/v1/users/me/delete
func (h *UserSettingsHandler) RequestAccountDeletion(c *gin.Context) {
	// Get user ID from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	// Parse request body
	var req models.DeleteAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Request deletion
	deletion, err := h.userSettingsService.RequestAccountDeletion(c.Request.Context(), userID, req.Reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Account deletion scheduled",
		"scheduled_for": deletion.ScheduledFor,
	})
}

// CancelAccountDeletion handles POST /api/v1/users/me/delete/cancel
func (h *UserSettingsHandler) CancelAccountDeletion(c *gin.Context) {
	// Get user ID from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	// Cancel deletion
	err := h.userSettingsService.CancelAccountDeletion(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Account deletion cancelled",
	})
}

// GetDeletionStatus handles GET /api/v1/users/me/delete/status
func (h *UserSettingsHandler) GetDeletionStatus(c *gin.Context) {
	// Get user ID from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	// Get pending deletion
	deletion, err := h.userSettingsService.GetPendingDeletion(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get deletion status",
		})
		return
	}

	if deletion == nil {
		c.JSON(http.StatusOK, gin.H{
			"pending": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pending":       true,
		"scheduled_for": deletion.ScheduledFor,
		"requested_at":  deletion.RequestedAt,
	})
}

// UpdateSocialLinks handles PUT /api/v1/users/me/social-links
func (h *UserSettingsHandler) UpdateSocialLinks(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	// Parse request body
	var req models.UpdateSocialLinksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Update social links
	err := h.userSettingsService.UpdateSocialLinks(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update social links",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Social links updated successfully",
	})
}
