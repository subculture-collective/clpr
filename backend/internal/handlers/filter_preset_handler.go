package handlers

import (
	"errors"
	"net/http"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FilterPresetHandler struct {
	presetService *services.FilterPresetService
}

func filterPresetUserID(c *gin.Context) (uuid.UUID, bool) {
	value, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return uuid.Nil, false
	}
	userID, ok := value.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid authenticated user"})
		return uuid.Nil, false
	}
	routeUserID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return uuid.Nil, false
	}
	if routeUserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return uuid.Nil, false
	}
	return userID, true
}

func handleFilterPresetError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, repository.ErrPresetNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Preset not found"})
	case errors.Is(err, repository.ErrUnauthorizedPresetAccess):
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
	case errors.Is(err, repository.ErrMaxPresetsReached):
		c.JSON(http.StatusConflict, gin.H{"error": "Maximum preset count reached"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": fallback})
	}
}

func publicFilterPreset(preset *models.UserFilterPreset) (*models.PublicFilterPreset, error) {
	filters, err := repository.ParseFiltersJSON(preset.FiltersJSON)
	if err != nil {
		return nil, err
	}
	return &models.PublicFilterPreset{ID: preset.ID, UserID: preset.UserID, Name: preset.Name, Filters: *filters, CreatedAt: preset.CreatedAt, UpdatedAt: preset.UpdatedAt}, nil
}

func NewFilterPresetHandler(presetService *services.FilterPresetService) *FilterPresetHandler {
	return &FilterPresetHandler{
		presetService: presetService,
	}
}

// CreatePreset creates a new filter preset
// POST /api/v1/users/:id/filter-presets
func (h *FilterPresetHandler) CreatePreset(c *gin.Context) {
	userID, ok := filterPresetUserID(c)
	if !ok {
		return
	}

	var req models.CreateFilterPresetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := req.Filters.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	preset, err := h.presetService.CreatePreset(c.Request.Context(), userID, &req)
	if err != nil {
		handleFilterPresetError(c, err, "Failed to create preset")
		return
	}

	response, err := publicFilterPreset(preset)
	if err != nil {
		handleFilterPresetError(c, err, "Failed to serialize preset")
		return
	}
	c.JSON(http.StatusCreated, response)
}

// GetUserPresets retrieves all filter presets for a user
// GET /api/v1/users/:id/filter-presets
func (h *FilterPresetHandler) GetUserPresets(c *gin.Context) {
	userID, ok := filterPresetUserID(c)
	if !ok {
		return
	}

	presets, err := h.presetService.GetUserPresets(c.Request.Context(), userID)
	if err != nil {
		handleFilterPresetError(c, err, "Failed to list presets")
		return
	}

	responses := make([]*models.PublicFilterPreset, 0, len(presets))
	for _, preset := range presets {
		response, err := publicFilterPreset(preset)
		if err != nil {
			handleFilterPresetError(c, err, "Failed to serialize presets")
			return
		}
		responses = append(responses, response)
	}
	c.JSON(http.StatusOK, responses)
}

// GetPreset retrieves a specific filter preset
// GET /api/v1/users/:id/filter-presets/:presetId
func (h *FilterPresetHandler) GetPreset(c *gin.Context) {
	userID, ok := filterPresetUserID(c)
	if !ok {
		return
	}

	presetIDParam := c.Param("presetId")
	presetID, err := uuid.Parse(presetIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid preset ID"})
		return
	}

	preset, err := h.presetService.GetPreset(c.Request.Context(), presetID, userID)
	if err != nil {
		handleFilterPresetError(c, err, "Failed to get preset")
		return
	}

	response, err := publicFilterPreset(preset)
	if err != nil {
		handleFilterPresetError(c, err, "Failed to serialize preset")
		return
	}
	c.JSON(http.StatusOK, response)
}

// UpdatePreset updates a filter preset
// PUT /api/v1/users/:id/filter-presets/:presetId
func (h *FilterPresetHandler) UpdatePreset(c *gin.Context) {
	userID, ok := filterPresetUserID(c)
	if !ok {
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
	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	preset, err := h.presetService.UpdatePreset(c.Request.Context(), presetID, userID, &req)
	if err != nil {
		handleFilterPresetError(c, err, "Failed to update preset")
		return
	}

	response, err := publicFilterPreset(preset)
	if err != nil {
		handleFilterPresetError(c, err, "Failed to serialize preset")
		return
	}
	c.JSON(http.StatusOK, response)
}

// DeletePreset deletes a filter preset
// DELETE /api/v1/users/:id/filter-presets/:presetId
func (h *FilterPresetHandler) DeletePreset(c *gin.Context) {
	userID, ok := filterPresetUserID(c)
	if !ok {
		return
	}

	presetIDParam := c.Param("presetId")
	presetID, err := uuid.Parse(presetIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid preset ID"})
		return
	}

	err = h.presetService.DeletePreset(c.Request.Context(), presetID, userID)
	if err != nil {
		handleFilterPresetError(c, err, "Failed to delete preset")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Preset deleted successfully"})
}
