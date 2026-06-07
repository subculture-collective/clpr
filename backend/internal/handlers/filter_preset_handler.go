package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
)

type FilterPresetHandler struct {
	presetService *services.FilterPresetService
}

func NewFilterPresetHandler(presetService *services.FilterPresetService) *FilterPresetHandler {
	return &FilterPresetHandler{
		presetService: presetService,
	}
}

// CreatePreset creates a new filter preset
// POST /api/v1/users/:id/filter-presets
func (h *FilterPresetHandler) CreatePreset(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Verify the requesting user matches the URL parameter
	userIDParam := c.Param("id")
	urlUserID, err := uuid.Parse(userIDParam)
	if err != nil || urlUserID != userID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot create presets for other users"})
		return
	}

	var req models.CreateFilterPresetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	preset, err := h.presetService.CreatePreset(c.Request.Context(), userID.(uuid.UUID), &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, repository.ErrMaxPresetsReached) {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, preset)
}

// GetUserPresets retrieves all filter presets for a user
// GET /api/v1/users/:id/filter-presets
func (h *FilterPresetHandler) GetUserPresets(c *gin.Context) {
	userIDParam := c.Param("id")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Verify the requesting user is the same as the target user
	requestingUserID, exists := c.Get("user_id")
	if !exists || requestingUserID.(uuid.UUID) != userID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	presets, err := h.presetService.GetUserPresets(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, presets)
}

// GetPreset retrieves a specific filter preset
// GET /api/v1/users/:id/filter-presets/:presetId
func (h *FilterPresetHandler) GetPreset(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	presetIDParam := c.Param("presetId")
	presetID, err := uuid.Parse(presetIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid preset ID"})
		return
	}

	preset, err := h.presetService.GetPreset(c.Request.Context(), presetID, userID.(uuid.UUID))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, repository.ErrPresetNotFound) {
			statusCode = http.StatusNotFound
		} else if errors.Is(err, repository.ErrUnauthorizedPresetAccess) {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, preset)
}

// UpdatePreset updates a filter preset
// PUT /api/v1/users/:id/filter-presets/:presetId
func (h *FilterPresetHandler) UpdatePreset(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	presetIDParam := c.Param("presetId")
	presetID, err := uuid.Parse(presetIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid preset ID"})
		return
	}

	var req models.UpdateFilterPresetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	preset, err := h.presetService.UpdatePreset(c.Request.Context(), presetID, userID.(uuid.UUID), &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, repository.ErrPresetNotFound) {
			statusCode = http.StatusNotFound
		} else if errors.Is(err, repository.ErrUnauthorizedPresetAccess) {
			statusCode = http.StatusForbidden
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, preset)
}

// DeletePreset deletes a filter preset
// DELETE /api/v1/users/:id/filter-presets/:presetId
func (h *FilterPresetHandler) DeletePreset(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	presetIDParam := c.Param("presetId")
	presetID, err := uuid.Parse(presetIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid preset ID"})
		return
	}

	err = h.presetService.DeletePreset(c.Request.Context(), presetID, userID.(uuid.UUID))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, repository.ErrPresetNotFound) {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Preset deleted successfully"})
}
