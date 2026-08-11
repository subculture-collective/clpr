package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ReputationHandler handles reputation-related HTTP requests
type ReputationHandler struct {
	reputationService *services.ReputationService
	authService       *services.AuthService
}

// NewReputationHandler creates a new reputation handler
func NewReputationHandler(reputationService *services.ReputationService, authService *services.AuthService) *ReputationHandler {
	return &ReputationHandler{
		reputationService: reputationService,
		authService:       authService,
	}
}

// GetUserReputation retrieves complete reputation info for a user
// GET /users/:id/reputation
func (h *ReputationHandler) GetUserReputation(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"code":    "INVALID_USER_ID",
			"message": "The provided user ID is not valid",
		})
		return
	}

	reputation, err := h.reputationService.GetUserReputation(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user reputation",
			"code":    "REPUTATION_FETCH_ERROR",
			"message": "Unable to retrieve user reputation. Please try again later.",
		})
		return
	}

	c.JSON(http.StatusOK, reputation)
}

// GetUserKarma retrieves karma details for a user
// GET /users/:id/karma
func (h *ReputationHandler) GetUserKarma(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"code":    "INVALID_USER_ID",
			"message": "The provided user ID is not valid",
		})
		return
	}
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid limit",
			"code":    "INVALID_PAGINATION",
			"message": "limit must be an integer between 1 and 100",
		})
		return
	}

	// Get karma breakdown
	breakdown, err := h.reputationService.GetKarmaBreakdown(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get uppies breakdown",
			"code":    "KARMA_FETCH_ERROR",
			"message": "Unable to retrieve uppies breakdown. Please try again later.",
		})
		return
	}

	history, err := h.reputationService.GetUserKarmaHistory(c.Request.Context(), userID, limit)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get uppies history",
			"code":    "KARMA_HISTORY_FETCH_ERROR",
			"message": "Unable to retrieve uppies history. Please try again later.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"breakdown": breakdown,
		"history":   history,
	})
}

// GetUserBadges retrieves all badges for a user
// GET /users/:id/badges
func (h *ReputationHandler) GetUserBadges(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"code":    "INVALID_USER_ID",
			"message": "The provided user ID is not valid",
		})
		return
	}

	badges, err := h.reputationService.GetUserBadges(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user badges",
			"code":    "BADGES_FETCH_ERROR",
			"message": "Unable to retrieve user badges. Please try again later.",
		})
		return
	}

	// Enrich badges with definitions
	enrichedBadges := make([]gin.H, 0, len(badges))
	for _, badge := range badges {
		def, err := services.GetBadgeDefinition(badge.BadgeID)
		if err != nil {
			// Skip invalid badges
			continue
		}
		enrichedBadges = append(enrichedBadges, gin.H{
			"id":          badge.ID,
			"badge_id":    badge.BadgeID,
			"awarded_at":  badge.AwardedAt,
			"awarded_by":  badge.AwardedBy,
			"name":        def.Name,
			"description": def.Description,
			"icon":        def.Icon,
			"category":    def.Category,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"badges": enrichedBadges,
	})
}

// GetLeaderboard retrieves leaderboard by type
// GET /leaderboards/:type
func (h *ReputationHandler) GetLeaderboard(c *gin.Context) {
	leaderboardType := c.Param("type")

	// Get pagination params
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid limit",
			"code":    "INVALID_PAGINATION",
			"message": "limit must be an integer between 1 and 100",
		})
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 || page > 1_000_000 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid page",
			"code":    "INVALID_PAGINATION",
			"message": "page must be an integer between 1 and 1000000",
		})
		return
	}
	offset := (page - 1) * limit

	var entries interface{}
	switch leaderboardType {
	case "karma":
		entries, err = h.reputationService.GetKarmaLeaderboard(c.Request.Context(), limit, offset)
	case "engagement":
		entries, err = h.reputationService.GetEngagementLeaderboard(c.Request.Context(), limit, offset)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid leaderboard type",
			"code":    "INVALID_LEADERBOARD_TYPE",
			"message": "Leaderboard type must be 'karma' or 'engagement'",
		})
		return
	}

	if err != nil {
		// Log the error without exposing sensitive details
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve leaderboard",
			"code":    "LEADERBOARD_FETCH_ERROR",
			"message": "Unable to retrieve leaderboard data. Please try again later.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"type":    leaderboardType,
		"page":    page,
		"limit":   limit,
		"entries": entries,
	})
}

// AwardBadge awards a badge to a user (admin only)
// POST /admin/users/:id/badges
func (h *ReputationHandler) AwardBadge(c *gin.Context) {
	// Get user ID from path
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"code":    "INVALID_USER_ID",
			"message": "The provided user ID is not valid",
		})
		return
	}

	adminID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	// Parse request body
	var req struct {
		BadgeID string `json:"badge_id" binding:"required,max=100"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"code":    "INVALID_REQUEST",
			"message": "Request body is missing or malformed",
		})
		return
	}

	// Award badge
	err = h.reputationService.ApplyAdminBadgeMutation(c.Request.Context(), userID, adminID, req.BadgeID, true)
	if err != nil {
		if strings.HasPrefix(err.Error(), "invalid badge ID") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid badge ID", "code": "INVALID_BADGE_ID"})
			return
		}
		if errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found", "code": "USER_NOT_FOUND"})
			return
		}
		if errors.Is(err, repository.ErrBadgeAlreadyAwarded) {
			c.JSON(http.StatusConflict, gin.H{"error": "Badge already awarded", "code": "BADGE_ALREADY_AWARDED"})
			return
		}
		if errors.Is(err, repository.ErrAdminRequired) || errors.Is(err, repository.ErrAdminSelfMutation) || errors.Is(err, repository.ErrProtectedAdminTarget) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Badge mutation not permitted", "code": "BADGE_MUTATION_FORBIDDEN"})
			return
		}
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to award badge",
			"code":    "BADGE_AWARD_ERROR",
			"message": "Unable to award badge. Please try again later.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Badge awarded successfully",
	})
}

// RemoveBadge removes a badge from a user (admin only)
// DELETE /admin/users/:id/badges/:badgeId
func (h *ReputationHandler) RemoveBadge(c *gin.Context) {
	// Get user ID from path
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"code":    "INVALID_USER_ID",
			"message": "The provided user ID is not valid",
		})
		return
	}

	// Get badge ID from path
	badgeID := c.Param("badgeId")
	if len(badgeID) == 0 || len(badgeID) > 100 || !services.IsValidBadge(badgeID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid badge ID", "code": "INVALID_BADGE_ID"})
		return
	}
	adminID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	// Remove badge
	err = h.reputationService.ApplyAdminBadgeMutation(c.Request.Context(), userID, adminID, badgeID, false)
	if err != nil {
		if errors.Is(err, repository.ErrBadgeNotFound) || errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User or badge award not found", "code": "BADGE_NOT_FOUND"})
			return
		}
		if errors.Is(err, repository.ErrAdminRequired) || errors.Is(err, repository.ErrAdminSelfMutation) || errors.Is(err, repository.ErrProtectedAdminTarget) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Badge mutation not permitted", "code": "BADGE_MUTATION_FORBIDDEN"})
			return
		}
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to remove badge",
			"code":    "BADGE_REMOVE_ERROR",
			"message": "Unable to remove badge. Please try again later.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Badge removed successfully",
	})
}

// GetBadgeDefinitions retrieves all badge definitions
// GET /badges
func (h *ReputationHandler) GetBadgeDefinitions(c *gin.Context) {
	accept := c.GetHeader("Accept")
	if accept != "" && !strings.Contains(accept, "application/json") && !strings.Contains(accept, "*/*") {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": "Only application/json is available"})
		return
	}
	badges := services.GetAllBadgeDefinitions()
	c.JSON(http.StatusOK, gin.H{
		"badges": badges,
	})
}

// Note: Trust score admin endpoints (breakdown, history, manual adjustment, leaderboard)
// are defined but not yet wired to the TrustScoreService.
// These will be implemented in a follow-up PR once the service is integrated into the main application.
// See backend/docs/trust-score-implementation.md for integration details.
