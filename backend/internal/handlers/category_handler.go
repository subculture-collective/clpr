package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
)

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

	categories, err := h.categoryRepo.List(c.Request.Context(), categoryType, featured)
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

	category, err := h.categoryRepo.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Category not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"category": category,
	})
}

// ListCategoryGames handles GET /api/v1/categories/:slug/games
func (h *CategoryHandler) ListCategoryGames(c *gin.Context) {
	slug := c.Param("slug")

	// Parse pagination parameters
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

	// Get user ID if authenticated
	var userID *uuid.UUID
	if user, exists := c.Get("user"); exists {
		if u, ok := user.(*models.User); ok {
			userID = &u.ID
		}
	}

	// Get category
	category, err := h.categoryRepo.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Category not found",
		})
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

	// Parse pagination and filter parameters
	limitStr := c.DefaultQuery("limit", "20")
	pageStr := c.DefaultQuery("page", "1")
	sort := c.DefaultQuery("sort", "hot")
	timeframe := c.Query("timeframe")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	offset := (page - 1) * limit

	// Get category
	category, err := h.categoryRepo.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Category not found",
		})
		return
	}

	// Get user ID if authenticated (for is_following field)
	var userID *uuid.UUID
	if user, exists := c.Get("user"); exists {
		if u, ok := user.(*models.User); ok {
			userID = &u.ID
		}
	}

	// Get games in category (to filter clips by game IDs)
	games, err := h.categoryRepo.GetGamesInCategory(c.Request.Context(), category.ID, userID, 100, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch category games",
		})
		return
	}

	// Collect all twitch game IDs from the category
	gameIDs := make([]string, 0, len(games))
	for _, game := range games {
		if game.TwitchGameID != "" {
			gameIDs = append(gameIDs, game.TwitchGameID)
		}
	}

	var allClips []models.Clip
	var total int

	if len(gameIDs) > 0 {
		// Fetch clips matching any of the category's games in a single query
		for _, gameID := range gameIDs {
			gid := gameID
			filters := repository.ClipFilters{
				GameID:    &gid,
				Sort:      sort,
				Timeframe: &timeframe,
			}

			clips, _, err := h.clipRepo.ListWithFilters(c.Request.Context(), filters, 200, 0)
			if err != nil {
				continue
			}
			allClips = append(allClips, clips...)
		}
	}

	// Fallback: if no clips found via games, show recent clips
	if len(allClips) == 0 {
		filters := repository.ClipFilters{
			Sort:      sort,
			Timeframe: &timeframe,
		}
		clips, count, err := h.clipRepo.ListWithFilters(c.Request.Context(), filters, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch clips",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"clips":    clips,
			"total":    count,
			"page":     page,
			"limit":    limit,
			"has_more": len(clips) == limit,
		})
		return
	}

	// Calculate pagination over combined results
	total = len(allClips)
	start := offset
	end := offset + limit

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	paginatedClips := make([]models.Clip, 0)
	if start < total {
		paginatedClips = allClips[start:end]
	}

	c.JSON(http.StatusOK, gin.H{
		"clips":    paginatedClips,
		"total":    total,
		"page":     page,
		"limit":    limit,
		"has_more": end < total,
	})
}
