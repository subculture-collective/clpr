package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

type BanReasonTemplateHandler struct {
	service *services.BanReasonTemplateService
	logger  *utils.StructuredLogger
}

func NewBanReasonTemplateHandler(service *services.BanReasonTemplateService, logger *utils.StructuredLogger) *BanReasonTemplateHandler {
	return &BanReasonTemplateHandler{
		service: service,
		logger:  logger,
	}
}

// GetTemplate retrieves a single template by ID
// GET /api/v1/moderation/ban-templates/:id
func (h *BanReasonTemplateHandler) GetTemplate(c *gin.Context) {
	ctx := c.Request.Context()

	templateID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	template, err := h.service.GetTemplate(ctx, templateID)
	if err != nil {
		if err == services.ErrTemplateNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
			return
		}
		h.logger.Error("Failed to get template", err, map[string]interface{}{
			"template_id": templateID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get template"})
		return
	}

	c.JSON(http.StatusOK, template)
}

// ListTemplates retrieves templates with optional filtering
// GET /api/v1/moderation/ban-templates?broadcasterID=...&includeDefaults=true
func (h *BanReasonTemplateHandler) ListTemplates(c *gin.Context) {
	ctx := c.Request.Context()

	broadcasterID := c.Query("broadcasterID")
	includeDefaults := c.DefaultQuery("includeDefaults", "true") == "true"

	var broadcasterIDPtr *string
	if broadcasterID != "" {
		broadcasterIDPtr = &broadcasterID
	}

	templates, err := h.service.ListTemplates(ctx, broadcasterIDPtr, includeDefaults)
	if err != nil {
		h.logger.Error("Failed to list templates", err, map[string]interface{}{
			"broadcaster_id":   broadcasterID,
			"include_defaults": includeDefaults,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list templates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

// CreateTemplate creates a new template
// POST /api/v1/moderation/ban-templates
func (h *BanReasonTemplateHandler) CreateTemplate(c *gin.Context) {
	ctx := c.Request.Context()

	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	var req models.CreateBanReasonTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	template, err := h.service.CreateTemplate(ctx, userID, &req)
	if err != nil {
		if err == services.ErrTemplateNameExists {
			c.JSON(http.StatusConflict, gin.H{"error": "Template name already exists for this broadcaster"})
			return
		}
		h.logger.Error("Failed to create template", err, map[string]interface{}{
			"user_id": userID,
			"name":    req.Name,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create template"})
		return
	}

	c.JSON(http.StatusCreated, template)
}

// UpdateTemplate updates an existing template
// PATCH /api/v1/moderation/ban-templates/:id
func (h *BanReasonTemplateHandler) UpdateTemplate(c *gin.Context) {
	ctx := c.Request.Context()

	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	templateID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	var req models.UpdateBanReasonTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	template, err := h.service.UpdateTemplate(ctx, userID, templateID, &req)
	if err != nil {
		if err == services.ErrTemplateNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
			return
		}
		if err == services.ErrCannotDeleteDefault {
			c.JSON(http.StatusForbidden, gin.H{"error": "Cannot modify default templates"})
			return
		}
		if err == services.ErrUnauthorizedTemplate {
			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to modify this template"})
			return
		}
		h.logger.Error("Failed to update template", err, map[string]interface{}{
			"user_id":     userID,
			"template_id": templateID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update template"})
		return
	}

	c.JSON(http.StatusOK, template)
}

// DeleteTemplate deletes a template
// DELETE /api/v1/moderation/ban-templates/:id
func (h *BanReasonTemplateHandler) DeleteTemplate(c *gin.Context) {
	ctx := c.Request.Context()

	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	templateID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template ID"})
		return
	}

	err = h.service.DeleteTemplate(ctx, userID, templateID)
	if err != nil {
		if err == services.ErrTemplateNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
			return
		}
		if err == services.ErrCannotDeleteDefault {
			c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete default templates"})
			return
		}
		if err == services.ErrUnauthorizedTemplate {
			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to delete this template"})
			return
		}
		h.logger.Error("Failed to delete template", err, map[string]interface{}{
			"user_id":     userID,
			"template_id": templateID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete template"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetUsageStats retrieves usage statistics for templates
// GET /api/v1/moderation/ban-templates/stats?broadcasterID=...
func (h *BanReasonTemplateHandler) GetUsageStats(c *gin.Context) {
	ctx := c.Request.Context()

	broadcasterID := c.Query("broadcasterID")
	var broadcasterIDPtr *string
	if broadcasterID != "" {
		broadcasterIDPtr = &broadcasterID
	}

	stats, err := h.service.GetUsageStats(ctx, broadcasterIDPtr)
	if err != nil {
		h.logger.Error("Failed to get usage stats", err, map[string]interface{}{
			"broadcaster_id": broadcasterID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get usage stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"templates": stats})
}
