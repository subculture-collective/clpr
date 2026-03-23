package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/subculture-collective/clipper/internal/models"
	"github.com/subculture-collective/clipper/internal/repository"
	"github.com/subculture-collective/clipper/internal/services"
)

// TagHandler handles tag-related HTTP requests
type TagHandler struct {
	tagRepo        *repository.TagRepository
	clipRepo       *repository.ClipRepository
	autoTagService *services.AutoTagService
}

// NewTagHandler creates a new TagHandler
func NewTagHandler(
	tagRepo *repository.TagRepository,
	clipRepo *repository.ClipRepository,
	autoTagService *services.AutoTagService,
) *TagHandler {
	return &TagHandler{
		tagRepo:        tagRepo,
		clipRepo:       clipRepo,
		autoTagService: autoTagService,
	}
}

// ListTags handles GET /tags
func (h *TagHandler) ListTags(c *gin.Context) {
	// Parse query parameters
	sort := c.DefaultQuery("sort", "popularity")
	limitStr := c.DefaultQuery("limit", "50")
	pageStr := c.DefaultQuery("page", "1")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 50
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	offset := (page - 1) * limit

	// Get tags from repository
	tags, err := h.tagRepo.List(c.Request.Context(), sort, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch tags",
		})
		return
	}

	// Get total count
	total, err := h.tagRepo.Count(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to count tags",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tags":     tags,
		"total":    total,
		"page":     page,
		"limit":    limit,
		"has_more": offset+len(tags) < total,
	})
}

// GetTag handles GET /tags/:slug
func (h *TagHandler) GetTag(c *gin.Context) {
	slug := c.Param("slug")

	tag, err := h.tagRepo.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Tag not found",
		})
		return
	}

	// Get clip count for this tag
	clipCount, err := h.tagRepo.CountClipsByTag(c.Request.Context(), slug)
	if err != nil {
		clipCount = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"tag":        tag,
		"clip_count": clipCount,
	})
}

// GetClipsByTag handles GET /tags/:slug/clips
func (h *TagHandler) GetClipsByTag(c *gin.Context) {
	slug := c.Param("slug")

	// Verify tag exists
	tag, err := h.tagRepo.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Tag not found",
		})
		return
	}

	// Parse pagination
	limitStr := c.DefaultQuery("limit", "20")
	pageStr := c.DefaultQuery("page", "1")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	offset := (page - 1) * limit

	// Get clip IDs with this tag
	clipIDs, err := h.tagRepo.GetClipsByTag(c.Request.Context(), slug, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch clips",
		})
		return
	}

	// Fetch full clip details
	clips := make([]*models.Clip, 0, len(clipIDs))
	for _, clipID := range clipIDs {
		clip, err := h.clipRepo.GetByID(c.Request.Context(), clipID)
		if err != nil {
			continue // Skip clips that fail to load
		}
		clips = append(clips, clip)
	}

	// Get total count
	total, err := h.tagRepo.CountClipsByTag(c.Request.Context(), slug)
	if err != nil {
		total = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"tag":      tag,
		"clips":    clips,
		"total":    total,
		"page":     page,
		"limit":    limit,
		"has_more": offset+len(clips) < total,
	})
}

// AddTagsToClip handles POST /clips/:id/tags
func (h *TagHandler) AddTagsToClip(c *gin.Context) {
	clipIDStr := c.Param("id")
	clipID, err := uuid.Parse(clipIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid clip ID",
		})
		return
	}

	// Parse request body
	var req struct {
		TagSlugs []string `json:"tag_slugs" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Validate number of tags
	if len(req.TagSlugs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "At least one tag slug is required",
		})
		return
	}

	if len(req.TagSlugs) > 10 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Maximum 10 tags can be added at once",
		})
		return
	}

	// Check if clip exists
	clip, err := h.clipRepo.GetByID(c.Request.Context(), clipID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Clip not found",
		})
		return
	}

	// Check total tag count for clip (max 15 tags per clip)
	currentCount, err := h.tagRepo.GetClipTagCount(c.Request.Context(), clipID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check tag count",
		})
		return
	}

	if currentCount+len(req.TagSlugs) > 15 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Clip can have maximum 15 tags",
		})
		return
	}

	// Add tags to clip
	addedTags := make([]*models.Tag, 0)
	for _, slug := range req.TagSlugs {
		// Validate slug format
		slug = strings.TrimSpace(strings.ToLower(slug))
		if len(slug) < 2 || len(slug) > 50 {
			continue
		}

		// Get or create tag
		tag, err := h.tagRepo.GetBySlug(c.Request.Context(), slug)
		if err != nil {
			// Tag doesn't exist, create it (for now, allow users to create tags)
			// In production, you might want to require admin approval
			// Simple title case: capitalize first letter of each word
			name := strings.ToUpper(slug[:1]) + slug[1:]
			tag, err = h.tagRepo.GetOrCreateTag(c.Request.Context(), name, slug, nil)
			if err != nil {
				continue
			}
		} // Add tag to clip
		err = h.tagRepo.AddTagToClip(c.Request.Context(), clipID, tag.ID)
		if err != nil {
			continue // Skip if already exists or other error
		}

		addedTags = append(addedTags, tag)
	}

	// Get updated tag list for clip
	tags, err := h.tagRepo.GetClipTags(c.Request.Context(), clipID)
	if err != nil {
		tags = addedTags // Fallback to just added tags
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tags added successfully",
		"clip":    clip,
		"tags":    tags,
	})
}

// RemoveTagFromClip handles DELETE /clips/:id/tags/:slug
func (h *TagHandler) RemoveTagFromClip(c *gin.Context) {
	clipIDStr := c.Param("id")
	clipID, err := uuid.Parse(clipIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid clip ID",
		})
		return
	}

	tagSlug := c.Param("slug")

	// Get tag by slug
	tag, err := h.tagRepo.GetBySlug(c.Request.Context(), tagSlug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Tag not found",
		})
		return
	}

	// Remove tag from clip
	err = h.tagRepo.RemoveTagFromClip(c.Request.Context(), clipID, tag.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Tag association not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tag removed successfully",
	})
}

// SearchTags handles GET /tags/search
func (h *TagHandler) SearchTags(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Search query is required",
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 50 {
		limit = 10
	}

	tags, err := h.tagRepo.Search(c.Request.Context(), query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to search tags",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tags": tags,
	})
}

// CreateTag handles POST /admin/tags
func (h *TagHandler) CreateTag(c *gin.Context) {
	var req struct {
		Name        string  `json:"name" binding:"required,min=2,max=50"`
		Slug        string  `json:"slug" binding:"required,min=2,max=50"`
		Description *string `json:"description"`
		Color       *string `json:"color"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Validate color format if provided
	if req.Color != nil && !isValidHexColor(*req.Color) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid color format. Use hex format like #FF0000",
		})
		return
	}

	// Create tag
	tag := &models.Tag{
		ID:          uuid.New(),
		Name:        req.Name,
		Slug:        strings.ToLower(req.Slug),
		Description: req.Description,
		Color:       req.Color,
		UsageCount:  0,
	}

	err := h.tagRepo.Create(c.Request.Context(), tag)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Tag with this name or slug already exists",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create tag",
		})
		return
	}

	// Fetch created tag
	createdTag, _ := h.tagRepo.GetBySlug(c.Request.Context(), tag.Slug)
	if createdTag != nil {
		tag = createdTag
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Tag created successfully",
		"tag":     tag,
	})
}

// UpdateTag handles PUT /admin/tags/:id
func (h *TagHandler) UpdateTag(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tag ID",
		})
		return
	}

	var req struct {
		Name        string  `json:"name" binding:"required,min=2,max=50"`
		Slug        string  `json:"slug" binding:"required,min=2,max=50"`
		Description *string `json:"description"`
		Color       *string `json:"color"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Validate color format if provided
	if req.Color != nil && !isValidHexColor(*req.Color) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid color format. Use hex format like #FF0000",
		})
		return
	}

	// Get existing tag
	tag, err := h.tagRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Tag not found",
		})
		return
	}

	// Update tag fields
	tag.Name = req.Name
	tag.Slug = strings.ToLower(req.Slug)
	tag.Description = req.Description
	tag.Color = req.Color

	err = h.tagRepo.Update(c.Request.Context(), tag)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Tag with this name or slug already exists",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update tag",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tag updated successfully",
		"tag":     tag,
	})
}

// DeleteTag handles DELETE /admin/tags/:id
func (h *TagHandler) DeleteTag(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tag ID",
		})
		return
	}

	// Delete tag (will also delete clip associations)
	err = h.tagRepo.Delete(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Tag not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete tag",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tag deleted successfully",
	})
}

// GetClipTags handles GET /clips/:id/tags
func (h *TagHandler) GetClipTags(c *gin.Context) {
	clipIDStr := c.Param("id")
	clipID, err := uuid.Parse(clipIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid clip ID",
		})
		return
	}

	tags, err := h.tagRepo.GetClipTags(c.Request.Context(), clipID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch tags",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tags": tags,
	})
}

// ListBlacklistedTags returns all blacklisted tag patterns (admin only)
func (h *TagHandler) ListBlacklistedTags(c *gin.Context) {
	tags, err := h.tagRepo.GetBlacklistedTags(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get blacklisted tags"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": tags})
}

// AddBlacklistedTag adds a pattern to the tag blacklist (admin only)
func (h *TagHandler) AddBlacklistedTag(c *gin.Context) {
	var req struct {
		Pattern string  `json:"pattern" binding:"required"`
		Reason  *string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pattern is required"})
		return
	}

	// Get admin user ID from context
	var createdBy *uuid.UUID
	if userIDValue, exists := c.Get("user_id"); exists {
		if uid, ok := userIDValue.(uuid.UUID); ok {
			createdBy = &uid
		}
	}

	if err := h.tagRepo.AddBlacklistedTag(c.Request.Context(), req.Pattern, req.Reason, createdBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add blacklisted tag"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// RemoveBlacklistedTag removes a pattern from the tag blacklist (admin only)
func (h *TagHandler) RemoveBlacklistedTag(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	if err := h.tagRepo.RemoveBlacklistedTag(c.Request.Context(), id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Blacklisted tag not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove blacklisted tag"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// isValidHexColor validates hex color format
func isValidHexColor(color string) bool {
	if len(color) != 7 {
		return false
	}
	if color[0] != '#' {
		return false
	}
	for _, c := range color[1:] {
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}
