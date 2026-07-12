package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var errChatAccessDenied = errors.New("channel access denied")

// authenticatedUserID returns the identity established by AuthMiddleware.
// Treating a missing or malformed value as unauthenticated keeps handlers from
// accidentally passing interface{} values to SQL authorization checks.
func authenticatedUserID(c *gin.Context) (uuid.UUID, bool) {
	value, exists := c.Get("user_id")
	userID, ok := value.(uuid.UUID)
	if !exists || !ok || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return uuid.Nil, false
	}
	return userID, true
}

// requireChatChannelAccess allows every authenticated user to enter an active
// public channel, while private channels are visible only to their members.
func requireChatChannelAccess(c *gin.Context, db *pgxpool.Pool, channelID, userID uuid.UUID) bool {
	var allowed bool
	err := db.QueryRow(c.Request.Context(), `
		SELECT cc.is_active AND (
			cc.channel_type = 'public' OR EXISTS (
				SELECT 1 FROM channel_members cm
				WHERE cm.channel_id = cc.id AND cm.user_id = $2
			)
		)
		FROM chat_channels cc
		WHERE cc.id = $1`, channelID, userID).Scan(&allowed)
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return false
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify channel access"})
		return false
	}
	if !allowed {
		c.JSON(http.StatusForbidden, gin.H{"error": errChatAccessDenied.Error()})
		return false
	}
	return true
}

func isChatStaffRole(role string) bool {
	return role == "owner" || role == "admin" || role == "moderator"
}

// requireChatStaff applies channel-scoped authorization. Global application
// roles alone must never grant moderation rights over an unrelated channel.
func requireChatStaff(c *gin.Context, db *pgxpool.Pool, channelID, userID uuid.UUID) bool {
	var role string
	err := db.QueryRow(c.Request.Context(), `
		SELECT cm.role
		FROM channel_members cm
		JOIN chat_channels cc ON cc.id = cm.channel_id
		WHERE cm.channel_id = $1 AND cm.user_id = $2 AND cc.is_active = true`, channelID, userID).Scan(&role)
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Channel staff role required"})
		return false
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify channel role"})
		return false
	}
	if !isChatStaffRole(role) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Channel staff role required"})
		return false
	}
	return true
}
