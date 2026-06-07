package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	ws "git.subcult.tv/subculture-collective/clpr/internal/websocket"
)

// WebSocketHandler handles WebSocket chat connections
type WebSocketHandler struct {
	db     *pgxpool.Pool
	server *ws.Server
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(db *pgxpool.Pool, server *ws.Server) *WebSocketHandler {
	return &WebSocketHandler{
		db:     db,
		server: server,
	}
}

// HandleConnection handles WebSocket connection upgrades
func (h *WebSocketHandler) HandleConnection(c *gin.Context) {
	// Extract channel ID from URL parameter and validate
	channelID := c.Param("id")
	channelUUID, err := uuid.Parse(channelID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	// Extract user from context (set by AuthMiddleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	// Verify channel exists and is active
	var channelExists bool
	err = h.db.QueryRow(c.Request.Context(),
		"SELECT EXISTS(SELECT 1 FROM chat_channels WHERE id = $1 AND is_active = true)",
		channelUUID).Scan(&channelExists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify channel"})
		return
	}
	if !channelExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found or inactive"})
		return
	}

	// Get user details
	var username string
	err = h.db.QueryRow(c.Request.Context(), "SELECT username FROM users WHERE id = $1", userID).Scan(&username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user details"})
		return
	}

	// Check if user is banned from the channel
	var isBanned bool
	banQuery := `
		SELECT EXISTS(
			SELECT 1 FROM chat_bans 
			WHERE channel_id = $1 AND user_id = $2 
			AND (expires_at IS NULL OR expires_at > NOW())
		)
	`
	err = h.db.QueryRow(c.Request.Context(), banQuery, channelID, userID).Scan(&isBanned)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check ban status"})
		return
	}

	if isBanned {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are banned from this channel"})
		return
	}

	// Handle the WebSocket upgrade
	// Note: After upgrade, we cannot send HTTP responses anymore
	_ = h.server.HandleWebSocket(c.Writer, c.Request, userID, username, channelID)
}

// GetMessageHistory returns message history for a channel
func (h *WebSocketHandler) GetMessageHistory(c *gin.Context) {
	channelID := c.Param("id")
	channelUUID, err := uuid.Parse(channelID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	// Get authenticated user
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	// Verify channel exists
	var channelExists bool
	err = h.db.QueryRow(c.Request.Context(),
		"SELECT EXISTS(SELECT 1 FROM chat_channels WHERE id = $1 AND is_active = true)",
		channelUUID).Scan(&channelExists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify channel"})
		return
	}
	if !channelExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found or inactive"})
		return
	}

	// Check if user is banned from the channel
	var isBanned bool
	banQuery := `
		SELECT EXISTS(
			SELECT 1 FROM chat_bans 
			WHERE channel_id = $1 AND user_id = $2 
			AND (expires_at IS NULL OR expires_at > NOW())
		)
	`
	err = h.db.QueryRow(c.Request.Context(), banQuery, channelUUID, userID).Scan(&isBanned)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check ban status"})
		return
	}

	if isBanned {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are banned from this channel"})
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "50")
	cursor := c.Query("cursor")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 50
	}

	// Build query
	query := `
		SELECT cm.id, cm.channel_id, cm.user_id, cm.content, cm.created_at,
		       u.username, u.display_name, u.avatar_url
		FROM chat_messages cm
		JOIN users u ON cm.user_id = u.id
		WHERE cm.channel_id = $1 AND cm.is_deleted = false
	`

	args := []interface{}{channelID}
	argIndex := 2

	// Add cursor condition if provided
	if cursor != "" {
		cursorTime, err := time.Parse(time.RFC3339, cursor)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cursor format; expected RFC3339 timestamp"})
			return
		}
		query += fmt.Sprintf(" AND cm.created_at < $%d", argIndex)
		args = append(args, cursorTime)
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY cm.created_at DESC LIMIT $%d", argIndex)
	args = append(args, limit)

	// Execute query
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}
	defer rows.Close()

	messages := []models.ChatMessage{}
	var nextCursor *string

	for rows.Next() {
		var msg models.ChatMessage
		err := rows.Scan(
			&msg.ID,
			&msg.ChannelID,
			&msg.UserID,
			&msg.Content,
			&msg.CreatedAt,
			&msg.Username,
			&msg.DisplayName,
			&msg.AvatarURL,
		)
		if err != nil {
			log.Printf("Failed to scan message row: %v", err)
			continue
		}
		messages = append(messages, msg)
	}

	// Set next cursor if we have messages
	if len(messages) > 0 {
		lastMessage := messages[len(messages)-1]
		cursorStr := lastMessage.CreatedAt.Format(time.RFC3339)
		nextCursor = &cursorStr
	}

	// Reverse messages to return oldest first
	for i := 0; i < len(messages)/2; i++ {
		j := len(messages) - 1 - i
		messages[i], messages[j] = messages[j], messages[i]
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"cursor":   nextCursor,
		"limit":    limit,
	})
}

// GetHealthCheck returns health information about WebSocket connections
func (h *WebSocketHandler) GetHealthCheck(c *gin.Context) {
	stats := h.server.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"status":            "healthy",
		"total_connections": stats["total_connections"],
		"active_channels":   stats["active_channels"],
		"timestamp":         time.Now(),
	})
}

// GetChannelStats returns detailed statistics for all channels
func (h *WebSocketHandler) GetChannelStats(c *gin.Context) {
	stats := h.server.GetStats()

	c.JSON(http.StatusOK, stats)
}
