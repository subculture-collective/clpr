package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
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
		// Assume it's a Twitch game ID
		twitchGameID = gameIDStr
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
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	currentUser := user.(*models.User)
	gameIDStr := c.Param("gameId")

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
	}

	// Follow the game
	err = h.gameRepo.FollowGame(c.Request.Context(), currentUser.ID, gameID)
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
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	currentUser := user.(*models.User)
	gameIDStr := c.Param("gameId")

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
	}

	// Unfollow the game
	err = h.gameRepo.UnfollowGame(c.Request.Context(), currentUser.ID, gameID)
	if err != nil {
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
	userIDStr := c.Param("userId")

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
		limit = 20
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
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
