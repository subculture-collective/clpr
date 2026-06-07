package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
)

// FavoriteHandler handles favorite-related requests
type FavoriteHandler struct {
	favoriteRepo *repository.FavoriteRepository
	voteRepo     *repository.VoteRepository
	clipService  *services.ClipService
}

// NewFavoriteHandler creates a new FavoriteHandler
func NewFavoriteHandler(favoriteRepo *repository.FavoriteRepository, voteRepo *repository.VoteRepository, clipService *services.ClipService) *FavoriteHandler {
	return &FavoriteHandler{
		favoriteRepo: favoriteRepo,
		voteRepo:     voteRepo,
		clipService:  clipService,
	}
}

// ListUserFavorites handles GET /favorites
func (h *FavoriteHandler) ListUserFavorites(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "UNAUTHORIZED",
				Message: "Authentication required",
			},
		})
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Invalid user ID format",
			},
		})
		return
	}

	// Parse query parameters
	sort := c.DefaultQuery("sort", "newest")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "25"))

	// Validate and constrain parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 25
	}

	// Validate sort parameter
	validSorts := map[string]bool{
		"newest":    true,
		"top":       true,
		"discussed": true,
	}
	if !validSorts[sort] {
		sort = "newest"
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Fetch favorite clips
	clips, total, err := h.favoriteRepo.GetClipsByUserID(c.Request.Context(), userID, sort, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Success: false,
			Error: &ErrorInfo{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to fetch favorites",
			},
		})
		return
	}

	// Enrich clips with user interaction data
	clipsWithData := make([]services.ClipWithUserData, len(clips))
	for i, clip := range clips {
		clipsWithData[i] = services.ClipWithUserData{
			Clip:        clip,
			IsFavorited: true, // All clips in favorites are favorited
		}

		// Get vote counts
		upvotes, downvotes, err := h.voteRepo.GetVoteCounts(c.Request.Context(), clip.ID)
		if err == nil {
			clipsWithData[i].UpvoteCount = upvotes
			clipsWithData[i].DownvoteCount = downvotes
		}

		// Get user vote
		vote, err := h.voteRepo.GetVote(c.Request.Context(), userID, clip.ID)
		if err == nil && vote != nil {
			clipsWithData[i].UserVote = &vote.VoteType
		}
	}

	// Build pagination metadata
	totalPages := (total + limit - 1) / limit
	meta := PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}

	c.JSON(http.StatusOK, StandardResponse{
		Success: true,
		Data:    clipsWithData,
		Meta:    meta,
	})
}
