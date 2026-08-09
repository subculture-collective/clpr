package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GameHandler handles game-related HTTP requests
type GameHandler struct {
	gameRepo    *repository.GameRepository
	clipRepo    *repository.ClipRepository
	authService *services.AuthService
}

// NewGameHandler creates a new GameHandler
func NewGameHandler(
	gameRepo *repository.GameRepository,
	clipRepo *repository.ClipRepository,
	authService *services.AuthService,
) *GameHandler {
	return &GameHandler{
		gameRepo:    gameRepo,
		clipRepo:    clipRepo,
		authService: authService,
	}
}

// GetGame handles GET /api/v1/games/:gameId
func (h *GameHandler) GetGame(c *gin.Context) {
	gameIDStr := c.Param("gameId")
	if gameIDStr == "" || len(gameIDStr) > 128 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID"})
		return
	}

	// Try to parse as UUID first (internal ID)
	gameID, err := uuid.Parse(gameIDStr)
	var game *models.GameWithStats

	// Get user ID if authenticated
	var userID *uuid.UUID
	if user, exists := c.Get("user"); exists {
		if u, ok := user.(*models.User); ok {
			userID = &u.ID
		}
	}

	if err == nil {
		// It's a UUID, get by internal ID
		game, err = h.gameRepo.GetWithStats(c.Request.Context(), gameID, userID)
	} else {
		// Assume it's a Twitch game ID
		basicGame, err := h.gameRepo.GetByTwitchGameID(c.Request.Context(), gameIDStr)
		if err == nil && basicGame != nil {
			game, err = h.gameRepo.GetWithStats(c.Request.Context(), basicGame.ID, userID)
		}
	}

	if err != nil || game == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Game not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"game": game,
	})
}

// ListGameClips handles GET /api/v1/games/:gameId/clips
func (h *GameHandler) ListGameClips(c *gin.Context) {
	gameIDStr := c.Param("gameId")
	if gameIDStr == "" || len(gameIDStr) > 128 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID"})
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

	// Get game to validate it exists and get Twitch game ID
	var twitchGameID string
	gameID, err := uuid.Parse(gameIDStr)
	if err == nil {
		// It's a UUID, get by internal ID
		game, err := h.gameRepo.GetByID(c.Request.Context(), gameID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Game not found",
			})
			return
		}
		twitchGameID = game.TwitchGameID
	} else {
		game, lookupErr := h.gameRepo.GetByTwitchGameID(c.Request.Context(), gameIDStr)
		if lookupErr != nil || game == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
			return
		}
		twitchGameID = game.TwitchGameID
	}

	// Build filters for clips
	filters := repository.ClipFilters{
		GameID:    &twitchGameID,
		Sort:      sort,
		Timeframe: &timeframe,
	}

	clips, total, err := h.clipRepo.ListWithFilters(c.Request.Context(), filters, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch clips",
		})
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

// GetTrendingGames handles GET /api/v1/games/trending
func (h *GameHandler) GetTrendingGames(c *gin.Context) {
	// Parse pagination parameters
	page, limit, ok := parseCommunityPagination(c, 20)
	if !ok {
		return
	}

	offset := (page - 1) * limit

	games, err := h.gameRepo.GetTrending(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch trending games",
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

// FollowGame handles POST /api/v1/games/:gameId/follow
func (h *GameHandler) FollowGame(c *gin.Context) {
	// Get authenticated user
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	gameIDStr := c.Param("gameId")
	if gameIDStr == "" || len(gameIDStr) > 128 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID"})
		return
	}

	// Parse game ID
	gameID, err := uuid.Parse(gameIDStr)
	if err != nil {
		// Try to find by Twitch game ID
		game, err := h.gameRepo.GetByTwitchGameID(c.Request.Context(), gameIDStr)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Game not found",
			})
			return
		}
		gameID = game.ID
	} else if _, err := h.gameRepo.GetByID(c.Request.Context(), gameID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		return
	}

	// Follow the game
	err = h.gameRepo.FollowGame(c.Request.Context(), userID, gameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to follow game",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Game followed successfully",
	})
}

// UnfollowGame handles DELETE /api/v1/games/:gameId/follow
func (h *GameHandler) UnfollowGame(c *gin.Context) {
	// Get authenticated user
	userID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	gameIDStr := c.Param("gameId")
	if gameIDStr == "" || len(gameIDStr) > 128 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID"})
		return
	}

	// Parse game ID
	gameID, err := uuid.Parse(gameIDStr)
	if err != nil {
		// Try to find by Twitch game ID
		game, err := h.gameRepo.GetByTwitchGameID(c.Request.Context(), gameIDStr)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Game not found",
			})
			return
		}
		gameID = game.ID
	} else if _, err := h.gameRepo.GetByID(c.Request.Context(), gameID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		return
	}

	// Unfollow the game
	err = h.gameRepo.UnfollowGame(c.Request.Context(), userID, gameID)
	if err != nil {
		if errors.Is(err, repository.ErrGameFollowNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Game follow not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to unfollow game",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Game unfollowed successfully",
	})
}

// GetFollowedGames handles GET /api/v1/users/:userId/games/following
func (h *GameHandler) GetFollowedGames(c *gin.Context) {
	userIDStr := c.Param("id")

	// Parse user ID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "20")
	pageStr := c.DefaultQuery("page", "1")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 100"})
		return
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 || page > 100000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page must be between 1 and 100000"})
		return
	}

	offset := (page - 1) * limit

	games, err := h.gameRepo.GetFollowedGames(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch followed games",
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
