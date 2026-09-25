package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/internal/tagtaxonomy"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var adminTagSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var tagBlacklistGlobPattern = regexp.MustCompile(`^[a-z0-9/_.%?*\\-]+$`)

// Children returned per parent by GET /tags/tree.
const (
	defaultTagTreeChildLimit = 50
	maxTagTreeChildLimit     = 500
)

// TagTreeNode represents a tag in a hierarchical tree response.
type TagTreeNode struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	ParentSlug  *string   `json:"parent_slug,omitempty"`
	Description *string   `json:"description,omitempty"`
	Color       *string   `json:"color,omitempty"`
	UsageCount  int       `json:"usage_count"`
	CreatedAt   time.Time `json:"created_at"`
	Lane        string    `json:"lane"`
	DisplayName string    `json:"display_name"`
	Evidence    string    `json:"evidence,omitempty"`
	// ChildCount is the number of visible children, including any beyond the
	// per-parent limit that were omitted from Children.
	ChildCount int            `json:"child_count,omitempty"`
	Children   []*TagTreeNode `json:"children,omitempty"`
}

func tagToTreeNode(tag *models.Tag) *TagTreeNode {
	labels := tagtaxonomy.Classify(tag.Slug, tag.Name)
	return &TagTreeNode{
		Lane:        string(labels.Lane),
		DisplayName: labels.DisplayName,
		Evidence:    string(labels.Evidence),
		ID:          tag.ID,
		Name:        tag.Name,
		Slug:        tag.Slug,
		ParentSlug:  tag.ParentSlug,
		Description: tag.Description,
		Color:       tag.Color,
		UsageCount:  tag.UsageCount,
		CreatedAt:   tag.CreatedAt,
	}
}

// buildTagTree nests rows from a tag tree query under their depth-0 roots.
// Rows arrive ordered by depth and usage, so children keep that order.
func buildTagTree(rows []repository.TagTreeRow) []*TagTreeNode {
	childrenOf := make(map[string][]*repository.TagTreeRow)
	var roots []*repository.TagTreeRow
	for i := range rows {
		row := &rows[i]
		if row.Depth == 0 {
			roots = append(roots, row)
		} else if row.ParentSlug != nil {
			childrenOf[*row.ParentSlug] = append(childrenOf[*row.ParentSlug], row)
		}
	}

	var build func(row *repository.TagTreeRow, depth int) *TagTreeNode
	build = func(row *repository.TagTreeRow, depth int) *TagTreeNode {
		node := tagToTreeNode(&row.Tag)
		node.ChildCount = row.ChildCount
		if depth >= repository.MaxTagTreeDepth {
			return node
		}
		for _, child := range childrenOf[row.Slug] {
			node.Children = append(node.Children, build(child, depth+1))
		}
		return node
	}

	nodes := make([]*TagTreeNode, 0, len(roots))
	for _, root := range roots {
		nodes = append(nodes, build(root, 0))
	}
	return nodes
}

func validateAdminTagFields(name, slug string, description *string) (string, string, bool) {
	name = strings.TrimSpace(name)
	slug = strings.ToLower(strings.TrimSpace(slug))
	if len(name) < 2 || len(name) > 50 || len(slug) < 2 || len(slug) > 50 || !adminTagSlugPattern.MatchString(slug) {
		return "", "", false
	}
	if description != nil && len(*description) > 1000 {
		return "", "", false
	}
	return name, slug, true
}

func writeAdminTagError(c *gin.Context, err error, message string) {
	switch {
	case errors.Is(err, repository.ErrTagNotFound), errors.Is(err, repository.ErrBlacklistedTagNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
	case errors.Is(err, repository.ErrTagConflict), errors.Is(err, repository.ErrBlacklistedTagConflict):
		c.JSON(http.StatusConflict, gin.H{"error": "Tag already exists"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": message})
	}
}

// Public tag listings are cached as serialized responses. Admin tag writes
// clear them; usage counts may otherwise lag by up to one TTL.
const (
	tagResponseCachePrefix = "tags:api:"
	tagListCacheTTL        = time.Minute
	tagTreeCacheTTL        = 5 * time.Minute
	maxCachedTagRootLength = 100
)

// TagResponseCache stores serialized public tag responses. *redis.Client from
// pkg/redis satisfies it.
type TagResponseCache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	DeletePattern(ctx context.Context, pattern string) error
}

// TagHandler handles tag-related HTTP requests
type TagHandler struct {
	tagRepo        *repository.TagRepository
	clipRepo       *repository.ClipRepository
	autoTagService *services.AutoTagService
	cache          TagResponseCache
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

// SetResponseCache enables caching of GET /tags and GET /tags/tree.
func (h *TagHandler) SetResponseCache(cache TagResponseCache) {
	h.cache = cache
}

// writeCachedTagResponse serves key from the cache and reports whether it did.
func (h *TagHandler) writeCachedTagResponse(c *gin.Context, key string) bool {
	if h.cache == nil {
		return false
	}
	body, err := h.cache.Get(c.Request.Context(), key)
	if err != nil || body == "" {
		return false
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(body))
	return true
}

// writeTagResponse sends payload and stores it under key for ttl.
func (h *TagHandler) writeTagResponse(c *gin.Context, key string, ttl time.Duration, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		utils.GetLogger().Error("Failed to encode tag response", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode tags"})
		return
	}
	if h.cache != nil {
		if err := h.cache.Set(c.Request.Context(), key, body, ttl); err != nil {
			utils.GetLogger().Warn("Failed to cache tag response", map[string]interface{}{"key": key, "error": err.Error()})
		}
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", body)
}

// invalidateTagResponses clears cached public tag responses after a
// successful admin write. Deferred at the top of each write handler.
func (h *TagHandler) invalidateTagResponses(c *gin.Context) {
	if h.cache == nil || c.Writer.Status() >= http.StatusMultipleChoices {
		return
	}
	if err := h.cache.DeletePattern(c.Request.Context(), tagResponseCachePrefix+"*"); err != nil {
		utils.GetLogger().Warn("Failed to invalidate cached tag responses", map[string]interface{}{"error": err.Error()})
	}
}

// ListTags handles GET /tags
func (h *TagHandler) ListTags(c *gin.Context) {
	// Parse query parameters
	sort := c.DefaultQuery("sort", "popularity")
	switch sort {
	case "popularity", "trending", "alphabetical", "recent", "curated":
	default:
		sort = "popularity" // the repository treats unknown sorts as popularity
	}
	lane := c.Query("lane")
	switch tagtaxonomy.Lane(lane) {
	case tagtaxonomy.LaneCategory, tagtaxonomy.LaneDetected, tagtaxonomy.LaneStreamer,
		tagtaxonomy.LaneCommunity, tagtaxonomy.LaneDuration, tagtaxonomy.LaneLanguage:
	default:
		lane = "" // the repository treats unknown lanes as all lanes
	}
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

	cacheKey := fmt.Sprintf("%slist:%s:%s:%d:%d", tagResponseCachePrefix, sort, lane, limit, page)
	if h.writeCachedTagResponse(c, cacheKey) {
		return
	}

	// Get tags from repository
	tags, err := h.tagRepo.List(c.Request.Context(), sort, lane, limit, offset)
	if err != nil {
		utils.GetLogger().Error("Failed to list tags", err, map[string]interface{}{"sort": sort, "lane": lane})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch tags",
		})
		return
	}

	// Get total count
	total, err := h.tagRepo.Count(c.Request.Context(), lane)
	if err != nil {
		utils.GetLogger().Error("Failed to count tags", err, map[string]interface{}{"lane": lane})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to count tags",
		})
		return
	}

	h.writeTagResponse(c, cacheKey, tagListCacheTTL, gin.H{
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
	clipCount, err := h.tagRepo.CountClipsByTag(c.Request.Context(), tag.Slug)
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
	clipIDs, err := h.tagRepo.GetClipsByTag(c.Request.Context(), tag.Slug, limit, offset)
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
	total, err := h.tagRepo.CountClipsByTag(c.Request.Context(), tag.Slug)
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

// GetTagTree returns the tag hierarchy.
// Without ?root it returns the top-level tags that have children, each with
// its highest-usage children; childless top-level tags are listed by GET /tags.
// With ?root=<slug> it returns that tag's subtree. ?limit (default 50, max
// 500) caps the children returned under each parent; child_count reports
// the full number.
// GET /api/v1/tags/tree
func (h *TagHandler) GetTagTree(c *gin.Context) {
	rootSlug := c.DefaultQuery("root", "")
	limit, err := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(defaultTagTreeChildLimit)))
	if err != nil || limit < 1 || limit > maxTagTreeChildLimit {
		limit = defaultTagTreeChildLimit
	}

	cacheKey := fmt.Sprintf("%stree:%d:%s", tagResponseCachePrefix, limit, rootSlug)
	if h.writeCachedTagResponse(c, cacheKey) {
		return
	}

	var rows []repository.TagTreeRow
	if rootSlug != "" {
		rows, err = h.tagRepo.GetTagTree(c.Request.Context(), rootSlug, limit)
	} else {
		rows, err = h.tagRepo.GetTagForest(c.Request.Context(), limit)
	}
	if err != nil {
		utils.GetLogger().Error("Failed to fetch tag tree", err, map[string]interface{}{"root": rootSlug})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch tag tree"})
		return
	}

	payload := gin.H{"tags": buildTagTree(rows)}
	if len(rows) == 0 || len(rootSlug) > maxCachedTagRootLength {
		// Unknown roots are not cached, so arbitrary slugs cannot fill Redis.
		c.JSON(http.StatusOK, payload)
		return
	}
	h.writeTagResponse(c, cacheKey, tagTreeCacheTTL, payload)
}

// CreateTag handles POST /admin/tags
func (h *TagHandler) CreateTag(c *gin.Context) {
	defer h.invalidateTagResponses(c)
	var req struct {
		Name        string  `json:"name" binding:"required,min=2,max=50"`
		Slug        string  `json:"slug" binding:"required,min=2,max=50"`
		ParentSlug  *string `json:"parent_slug,omitempty" binding:"omitempty,max=100"`
		Description *string `json:"description"`
		Color       *string `json:"color"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}
	name, slug, valid := validateAdminTagFields(req.Name, req.Slug, req.Description)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and slug must be valid and description cannot exceed 1000 characters"})
		return
	}

	// Validate color format if provided
	if req.Color != nil && !isValidHexColor(*req.Color) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid color format. Use hex format like #FF0000",
		})
		return
	}

	// Validate parent_slug references an existing tag
	if req.ParentSlug != nil {
		parent, err := h.tagRepo.GetBySlug(c.Request.Context(), *req.ParentSlug)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "parent_slug does not reference an existing tag"})
			return
		}
		_ = parent // parent exists, validation passed
	}

	// Create tag
	tag := &models.Tag{
		ID:          uuid.New(),
		Name:        name,
		Slug:        slug,
		ParentSlug:  req.ParentSlug,
		Description: req.Description,
		Color:       req.Color,
		UsageCount:  0,
	}

	blacklisted, err := h.tagRepo.IsBlacklisted(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate tag"})
		return
	}
	if blacklisted {
		c.JSON(http.StatusConflict, gin.H{"error": "Tag slug is blacklisted"})
		return
	}
	err = h.tagRepo.Create(c.Request.Context(), tag)
	if err != nil {
		writeAdminTagError(c, err, "Failed to create tag")
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
	defer h.invalidateTagResponses(c)
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
		ParentSlug  *string `json:"parent_slug,omitempty" binding:"omitempty,max=100"`
		Description *string `json:"description"`
		Color       *string `json:"color"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}
	name, slug, valid := validateAdminTagFields(req.Name, req.Slug, req.Description)
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and slug must be valid and description cannot exceed 1000 characters"})
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

	// Validate parent_slug references an existing tag (skip self-references)
	if req.ParentSlug != nil && *req.ParentSlug != tag.Slug {
		parent, err := h.tagRepo.GetBySlug(c.Request.Context(), *req.ParentSlug)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "parent_slug does not reference an existing tag"})
			return
		}
		_ = parent
	}

	// Update tag fields
	blacklisted, err := h.tagRepo.IsBlacklisted(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate tag"})
		return
	}
	if blacklisted {
		c.JSON(http.StatusConflict, gin.H{"error": "Tag slug is blacklisted"})
		return
	}
	tag.Name = name
	tag.Slug = slug
	tag.ParentSlug = req.ParentSlug
	tag.Description = req.Description
	tag.Color = req.Color

	err = h.tagRepo.Update(c.Request.Context(), tag)
	if err != nil {
		writeAdminTagError(c, err, "Failed to update tag")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tag updated successfully",
		"tag":     tag,
	})
}

// DeleteTag handles DELETE /admin/tags/:id
func (h *TagHandler) DeleteTag(c *gin.Context) {
	defer h.invalidateTagResponses(c)
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
		writeAdminTagError(c, err, "Failed to delete tag")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tag deleted successfully",
	})
}

func (h *TagHandler) ListAdminTags(c *gin.Context) {
	tags, err := h.tagRepo.ListAdmin(c.Request.Context(), 500, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tags"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tags": tags})
}

func (h *TagHandler) SuppressTag(c *gin.Context) {
	defer h.invalidateTagResponses(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tag ID"})
		return
	}
	actor, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(strings.TrimSpace(req.Reason)) < 1 || len(req.Reason) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reason is required and must be at most 500 characters"})
		return
	}
	if err := h.tagRepo.Suppress(c.Request.Context(), id, actor, strings.TrimSpace(req.Reason)); err != nil {
		writeAdminTagError(c, err, "Failed to suppress tag")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tag suppressed"})
}

func (h *TagHandler) RestoreTag(c *gin.Context) {
	defer h.invalidateTagResponses(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tag ID"})
		return
	}
	if err := h.tagRepo.Restore(c.Request.Context(), id); err != nil {
		writeAdminTagError(c, err, "Failed to restore tag")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tag restored"})
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
	defer h.invalidateTagResponses(c)
	var req struct {
		Pattern string  `json:"pattern" binding:"required,min=2,max=50"`
		Reason  *string `json:"reason" binding:"omitempty,max=500"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pattern is required"})
		return
	}

	createdByID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	pattern := strings.ToLower(strings.TrimSpace(req.Pattern))
	if !tagBlacklistGlobPattern.MatchString(pattern) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pattern must be a safe tag glob using * and ? wildcards"})
		return
	}
	createdBy := &createdByID

	if err := h.tagRepo.AddBlacklistedTag(c.Request.Context(), pattern, req.Reason, createdBy); err != nil {
		writeAdminTagError(c, err, "Failed to add blacklisted tag")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// RemoveBlacklistedTag removes a pattern from the tag blacklist (admin only)
func (h *TagHandler) RemoveBlacklistedTag(c *gin.Context) {
	defer h.invalidateTagResponses(c)
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	if err := h.tagRepo.RemoveBlacklistedTag(c.Request.Context(), id); err != nil {
		writeAdminTagError(c, err, "Failed to remove blacklisted tag")
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
