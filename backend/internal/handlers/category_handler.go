package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
	"github.com/gin-gonic/gin"
)

// respondCategoryLookupError answers 404 only for a missing category. Other
// failures (such as statement timeouts) are logged and reported as 500 rather
// than disguised as "Category not found".
func respondCategoryLookupError(c *gin.Context, slug string, err error) {
	if errors.Is(err, repository.ErrCategoryNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}
	utils.GetLogger().Error("Failed to load category", err, map[string]interface{}{"category": slug})
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load category"})
}

// CategoryHandler handles category-related HTTP requests
type CategoryHandler struct {
	categoryRepo *repository.CategoryRepository
	clipRepo     *repository.ClipRepository
}

// NewCategoryHandler creates a new CategoryHandler
func NewCategoryHandler(
	categoryRepo *repository.CategoryRepository,
	clipRepo *repository.ClipRepository,
) *CategoryHandler {
	return &CategoryHandler{
		categoryRepo: categoryRepo,
		clipRepo:     clipRepo,
	}
}

// ListCategories handles GET /api/v1/categories
func (h *CategoryHandler) ListCategories(c *gin.Context) {
	var categoryType *string
	if typeParam := c.Query("type"); typeParam != "" {
		if typeParam != "game" && typeParam != "topic" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "type must be game or topic"})
			return
		}
		categoryType = &typeParam
	}

	var featured *bool
	if featuredParam := c.Query("featured"); featuredParam != "" {
		parsed, err := strconv.ParseBool(featuredParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid featured parameter",
			})
			return
		}
		featured = &parsed
	}
	var public *bool
	if publicParam := c.Query("public"); publicParam != "" {
		parsed, err := strconv.ParseBool(publicParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid public parameter"})
			return
		}
		public = &parsed
	}

	categories, err := h.categoryRepo.List(c.Request.Context(), categoryType, featured, public)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch categories",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"categories": categories,
	})
}

// GetCategory handles GET /api/v1/categories/:slug
func (h *CategoryHandler) GetCategory(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" || len(slug) > 255 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category slug"})
		return
	}

	category, err := h.categoryRepo.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		respondCategoryLookupError(c, slug, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"category": category,
	})
}

// ListCategoryGames handles GET /api/v1/categories/:slug/games
func (h *CategoryHandler) ListCategoryGames(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" || len(slug) > 255 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category slug"})
		return
	}

	// Parse pagination parameters
	page, limit, ok := parseCommunityPagination(c, 50)
	if !ok {
		return
	}

	offset := (page - 1) * limit

	// Get user ID if authenticated
	userID, ok := optionalCommunityUserID(c)
	if !ok {
		return
	}

	// Get category
	category, err := h.categoryRepo.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		respondCategoryLookupError(c, slug, err)
		return
	}

	// Get games in category
	games, err := h.categoryRepo.GetGamesInCategory(c.Request.Context(), category.ID, userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch games",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"games":    games,
		"page":     page,
		"limit":    limit,
		"has_more": len(games) == limit,
	})
}

// ListCategoryClips handles GET /api/v1/categories/:slug/clips
func (h *CategoryHandler) ListCategoryClips(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" || len(slug) > 255 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category slug"})
		return
	}

	// Parse pagination and filter parameters
	sort := c.DefaultQuery("sort", "hot")
	timeframe := c.Query("timeframe")
	validSorts := map[string]bool{"hot": true, "new": true, "top": true, "rising": true, "discussed": true, "trending": true, "popular": true}
	if !validSorts[sort] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sort"})
		return
	}
	validTimeframes := map[string]bool{"": true, "hour": true, "day": true, "week": true, "month": true, "year": true, "all": true}
	if !validTimeframes[timeframe] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid timeframe"})
		return
	}
	page, limit, ok := parseCommunityPagination(c, 20)
	if !ok {
		return
	}

	offset := (page - 1) * limit

	// Get category
	category, err := h.categoryRepo.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		respondCategoryLookupError(c, slug, err)
		return
	}

	filters := repository.ClipFilters{CategoryID: &category.ID, Sort: sort, Timeframe: &timeframe}
	clips, total, err := h.clipRepo.ListWithFilters(c.Request.Context(), filters, limit, offset)
	if err != nil {
		utils.GetLogger().Error("Failed to list category clips", err, map[string]interface{}{"category": slug, "sort": sort})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch clips"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"clips":    clips,
		"total":    total,
		"page":     page,
		"limit":    limit,
		"has_more": offset+len(clips) < total,
	})
}
