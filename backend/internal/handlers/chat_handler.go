package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// ChatHandler handles chat-related requests
type ChatHandler struct {
	db *pgxpool.Pool
}

// NewChatHandler creates a new chat handler
func NewChatHandler(db *pgxpool.Pool) *ChatHandler {
	return &ChatHandler{
		db: db,
	}
}

// CreateChannel creates a new chat channel with the creator as owner
func (h *ChatHandler) CreateChannel(c *gin.Context) {
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

	var req models.CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default channel type if not provided
	channelType := "public"
	if req.ChannelType != "" {
		channelType = req.ChannelType
	}

	// Begin transaction to create channel and add creator as owner
	tx, err := h.db.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback(c.Request.Context())

	// Create channel
	channelQuery := `
		INSERT INTO chat_channels (name, description, creator_id, channel_type, max_participants, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
		RETURNING id, created_at, updated_at
	`

	var channel models.ChatChannel
	err = tx.QueryRow(c.Request.Context(), channelQuery,
		req.Name, req.Description, userID, channelType, req.MaxParticipants).Scan(
		&channel.ID, &channel.CreatedAt, &channel.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create channel"})
		return
	}

	// Add creator as owner in channel_members
	memberQuery := `
		INSERT INTO channel_members (channel_id, user_id, role)
		VALUES ($1, $2, 'owner')
	`
	_, err = tx.Exec(c.Request.Context(), memberQuery, channel.ID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add creator as owner"})
		return
	}

	// Commit transaction
	if err := tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
		return
	}

	// Populate the rest of the channel data
	channel.Name = req.Name
	channel.Description = req.Description
	channel.CreatorID = userID
	channel.ChannelType = channelType
	channel.MaxParticipants = req.MaxParticipants
	channel.IsActive = true

	c.JSON(http.StatusCreated, channel)
}

// ListChannels returns a list of chat channels
func (h *ChatHandler) ListChannels(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")
	channelType := c.Query("type")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Build query
	query := `
		SELECT id, name, description, creator_id, channel_type, is_active, 
		       max_participants, created_at, updated_at
		FROM chat_channels
		WHERE is_active = true
	`

	args := []interface{}{}
	argIndex := 1

	// Filter by channel type if provided
	if channelType != "" && (channelType == "public" || channelType == "private") {
		query += " AND channel_type = $" + strconv.Itoa(argIndex)
		args = append(args, channelType)
		argIndex++
	}

	query += " ORDER BY created_at DESC LIMIT $" + strconv.Itoa(argIndex) + " OFFSET $" + strconv.Itoa(argIndex+1)
	args = append(args, limit, offset)

	// Execute query
	rows, err := h.db.Query(c.Request.Context(), query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch channels"})
		return
	}
	defer rows.Close()

	channels := []models.ChatChannel{}
	for rows.Next() {
		var channel models.ChatChannel
		err := rows.Scan(
			&channel.ID,
			&channel.Name,
			&channel.Description,
			&channel.CreatorID,
			&channel.ChannelType,
			&channel.IsActive,
			&channel.MaxParticipants,
			&channel.CreatedAt,
			&channel.UpdatedAt,
		)
		if err != nil {
			continue
		}
		channels = append(channels, channel)
	}

	c.JSON(http.StatusOK, gin.H{
		"channels": channels,
		"limit":    limit,
		"offset":   offset,
	})
}

// GetChannel returns details about a specific chat channel
func (h *ChatHandler) GetChannel(c *gin.Context) {
	channelID := c.Param("id")
	channelUUID, err := uuid.Parse(channelID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	query := `
		SELECT id, name, description, creator_id, channel_type, is_active, 
		       max_participants, created_at, updated_at
		FROM chat_channels
		WHERE id = $1
	`

	var channel models.ChatChannel
	err = h.db.QueryRow(c.Request.Context(), query, channelUUID).Scan(
		&channel.ID,
		&channel.Name,
		&channel.Description,
		&channel.CreatorID,
		&channel.ChannelType,
		&channel.IsActive,
		&channel.MaxParticipants,
		&channel.CreatedAt,
		&channel.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch channel"})
		return
	}

	c.JSON(http.StatusOK, channel)
}

// UpdateChannel updates a chat channel
func (h *ChatHandler) UpdateChannel(c *gin.Context) {
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

	var req models.UpdateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user is the creator
	var creatorID uuid.UUID
	err = h.db.QueryRow(c.Request.Context(), "SELECT creator_id FROM chat_channels WHERE id = $1", channelUUID).Scan(&creatorID)
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify channel"})
		return
	}

	if creatorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the channel creator can update the channel"})
		return
	}

	// Build update query dynamically
	updateParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Name != nil {
		updateParts = append(updateParts, "name = $"+strconv.Itoa(argIndex))
		args = append(args, *req.Name)
		argIndex++
	}

	if req.Description != nil {
		updateParts = append(updateParts, "description = $"+strconv.Itoa(argIndex))
		args = append(args, req.Description)
		argIndex++
	}

	if req.IsActive != nil {
		updateParts = append(updateParts, "is_active = $"+strconv.Itoa(argIndex))
		args = append(args, *req.IsActive)
		argIndex++
	}

	if req.MaxParticipants != nil {
		updateParts = append(updateParts, "max_participants = $"+strconv.Itoa(argIndex))
		args = append(args, req.MaxParticipants)
		argIndex++
	}

	if len(updateParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	// Add channel ID to args
	args = append(args, channelUUID)

	query := "UPDATE chat_channels SET " +
		joinStrings(updateParts, ", ") +
		" WHERE id = $" + strconv.Itoa(argIndex) +
		" RETURNING id, name, description, creator_id, channel_type, is_active, max_participants, created_at, updated_at"

	var channel models.ChatChannel
	err = h.db.QueryRow(c.Request.Context(), query, args...).Scan(
		&channel.ID,
		&channel.Name,
		&channel.Description,
		&channel.CreatorID,
		&channel.ChannelType,
		&channel.IsActive,
		&channel.MaxParticipants,
		&channel.CreatedAt,
		&channel.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update channel"})
		return
	}

	c.JSON(http.StatusOK, channel)
}

// joinStrings is a helper function to join strings
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// BanUser bans a user from a channel
func (h *ChatHandler) BanUser(c *gin.Context) {
	channelID := c.Param("id")
	if _, err := uuid.Parse(channelID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	moderatorID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.BanUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Calculate expiration time if duration is provided
	var expiresAt *time.Time
	if req.DurationMinutes != nil {
		expires := time.Now().Add(time.Duration(*req.DurationMinutes) * time.Minute)
		expiresAt = &expires
	}

	// Insert or update ban record
	query := `
		INSERT INTO chat_bans (channel_id, user_id, banned_by, reason, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (channel_id, user_id) 
		DO UPDATE SET 
			reason = EXCLUDED.reason,
			expires_at = EXCLUDED.expires_at,
			banned_by = EXCLUDED.banned_by,
			created_at = NOW()
		RETURNING id`

	var banID uuid.UUID
	err := h.db.QueryRow(c.Request.Context(), query,
		channelID, req.UserID, moderatorID, req.Reason, expiresAt).Scan(&banID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to ban user"})
		return
	}

	// Log the moderation action
	logQuery := `
		INSERT INTO chat_moderation_log (channel_id, moderator_id, target_user_id, action, reason, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)`

	metadata, _ := json.Marshal(map[string]interface{}{
		"duration_minutes": req.DurationMinutes,
		"expires_at":       expiresAt,
	})

	_, err = h.db.Exec(c.Request.Context(), logQuery,
		channelID, moderatorID, req.UserID, models.ChatActionBan, req.Reason, metadata)
	if err != nil {
		// Log error but don't fail the request
		_ = err
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "banned",
		"ban_id":     banID,
		"expires_at": expiresAt,
	})
}

// UnbanUser removes a ban from a user
func (h *ChatHandler) UnbanUser(c *gin.Context) {
	channelID := c.Param("id")
	userID := c.Param("user_id")

	if _, err := uuid.Parse(channelID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	if _, err := uuid.Parse(userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	moderatorID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Delete the ban
	query := `DELETE FROM chat_bans WHERE channel_id = $1 AND user_id = $2`
	result, err := h.db.Exec(c.Request.Context(), query, channelID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unban user"})
		return
	}

	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ban not found"})
		return
	}

	// Log the moderation action
	logQuery := `
		INSERT INTO chat_moderation_log (channel_id, moderator_id, target_user_id, action, reason)
		VALUES ($1, $2, $3, $4, $5)`

	_, err = h.db.Exec(c.Request.Context(), logQuery,
		channelID, moderatorID, userID, models.ChatActionUnban, "Manual unban")
	if err != nil {
		// Log error but don't fail the request
		_ = err
	}

	c.JSON(http.StatusOK, gin.H{"status": "unbanned"})
}

// MuteUser mutes a user in a channel (implemented similar to ban)
func (h *ChatHandler) MuteUser(c *gin.Context) {
	channelID := c.Param("id")
	if _, err := uuid.Parse(channelID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	moderatorID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.MuteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Calculate expiration time if duration is provided
	var expiresAt *time.Time
	if req.DurationMinutes != nil {
		expires := time.Now().Add(time.Duration(*req.DurationMinutes) * time.Minute)
		expiresAt = &expires
	}

	// Use same table as bans but with different action type in log
	query := `
		INSERT INTO chat_bans (channel_id, user_id, banned_by, reason, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (channel_id, user_id) 
		DO UPDATE SET 
			reason = EXCLUDED.reason,
			expires_at = EXCLUDED.expires_at,
			banned_by = EXCLUDED.banned_by,
			created_at = NOW()
		RETURNING id`

	var muteID uuid.UUID
	err := h.db.QueryRow(c.Request.Context(), query,
		channelID, req.UserID, moderatorID, req.Reason, expiresAt).Scan(&muteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mute user"})
		return
	}

	// Log the moderation action
	logQuery := `
		INSERT INTO chat_moderation_log (channel_id, moderator_id, target_user_id, action, reason, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)`

	metadata, _ := json.Marshal(map[string]interface{}{
		"duration_minutes": req.DurationMinutes,
		"expires_at":       expiresAt,
	})

	_, err = h.db.Exec(c.Request.Context(), logQuery,
		channelID, moderatorID, req.UserID, models.ChatActionMute, req.Reason, metadata)
	if err != nil {
		// Log error but don't fail the request
		_ = err
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "muted",
		"mute_id":    muteID,
		"expires_at": expiresAt,
	})
}

// TimeoutUser temporarily bans a user from a channel
func (h *ChatHandler) TimeoutUser(c *gin.Context) {
	channelID := c.Param("id")
	if _, err := uuid.Parse(channelID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	moderatorID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.TimeoutUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Calculate expiration time
	expiresAt := time.Now().Add(time.Duration(req.DurationMinutes) * time.Minute)

	// Insert or update timeout
	query := `
		INSERT INTO chat_bans (channel_id, user_id, banned_by, reason, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (channel_id, user_id) 
		DO UPDATE SET 
			reason = EXCLUDED.reason,
			expires_at = EXCLUDED.expires_at,
			banned_by = EXCLUDED.banned_by,
			created_at = NOW()
		RETURNING id`

	var timeoutID uuid.UUID
	err := h.db.QueryRow(c.Request.Context(), query,
		channelID, req.UserID, moderatorID, req.Reason, expiresAt).Scan(&timeoutID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to timeout user"})
		return
	}

	// Log the moderation action
	logQuery := `
		INSERT INTO chat_moderation_log (channel_id, moderator_id, target_user_id, action, reason, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)`

	metadata, _ := json.Marshal(map[string]interface{}{
		"duration_minutes": req.DurationMinutes,
		"expires_at":       expiresAt,
	})

	_, err = h.db.Exec(c.Request.Context(), logQuery,
		channelID, moderatorID, req.UserID, models.ChatActionTimeout, req.Reason, metadata)
	if err != nil {
		// Log error but don't fail the request
		_ = err
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "timed_out",
		"timeout_id": timeoutID,
		"expires_at": expiresAt,
	})
}

// DeleteMessage deletes a message from a channel
func (h *ChatHandler) DeleteMessage(c *gin.Context) {
	messageID := c.Param("id")
	if _, err := uuid.Parse(messageID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	moderatorID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.DeleteMessageRequest
	// Bind JSON but only enforce required fields
	_ = c.ShouldBindJSON(&req)

	// Get message details before deleting
	var msg models.ChatMessage
	msgQuery := `SELECT id, channel_id, user_id, content FROM chat_messages WHERE id = $1 AND is_deleted = false`
	err := h.db.QueryRow(c.Request.Context(), msgQuery, messageID).Scan(
		&msg.ID, &msg.ChannelID, &msg.UserID, &msg.Content)
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found or already deleted"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch message"})
		return
	}

	// Mark message as deleted
	now := time.Now()
	modID, ok := moderatorID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid moderator ID"})
		return
	}
	deleteQuery := `
		UPDATE chat_messages 
		SET is_deleted = true, deleted_at = $1, deleted_by = $2, updated_at = $1
		WHERE id = $3`

	_, err = h.db.Exec(c.Request.Context(), deleteQuery, now, modID, messageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete message"})
		return
	}

	// Log the moderation action
	logQuery := `
		INSERT INTO chat_moderation_log (channel_id, moderator_id, target_user_id, action, reason, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)`

	metadata, _ := json.Marshal(map[string]interface{}{
		"message_id":      messageID,
		"message_content": msg.Content,
	})

	_, err = h.db.Exec(c.Request.Context(), logQuery,
		msg.ChannelID, moderatorID, msg.UserID, models.ChatActionDelete, req.Reason, metadata)
	if err != nil {
		// Log error but don't fail the request
		_ = err
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// GetModerationLog retrieves the moderation log for a channel
func (h *ChatHandler) GetModerationLog(c *gin.Context) {
	channelID := c.Param("id")
	if _, err := uuid.Parse(channelID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	// Parse pagination parameters
	page := 1
	limit := 50
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	offset := (page - 1) * limit

	query := `
		SELECT 
			l.id, l.channel_id, l.moderator_id, l.target_user_id, 
			l.action, l.reason, l.metadata, l.created_at,
			m.username as moderator_username,
			t.username as target_username
		FROM chat_moderation_log l
		JOIN users m ON l.moderator_id = m.id
		LEFT JOIN users t ON l.target_user_id = t.id
		WHERE l.channel_id = $1
		ORDER BY l.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := h.db.Query(c.Request.Context(), query, channelID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch moderation log"})
		return
	}
	defer rows.Close()

	var logs []models.ChatModerationLog
	for rows.Next() {
		var log models.ChatModerationLog
		err := rows.Scan(
			&log.ID, &log.ChannelID, &log.ModeratorID, &log.TargetUserID,
			&log.Action, &log.Reason, &log.Metadata, &log.CreatedAt,
			&log.ModeratorUsername, &log.TargetUsername,
		)
		if err != nil {
			continue
		}
		logs = append(logs, log)
	}

	// Get total count for pagination
	var total int
	countQuery := `SELECT COUNT(*) FROM chat_moderation_log WHERE channel_id = $1`
	_ = h.db.QueryRow(c.Request.Context(), countQuery, channelID).Scan(&total)

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// CheckUserBan checks if a user is banned in a channel
func (h *ChatHandler) CheckUserBan(c *gin.Context) {
	channelID := c.Param("id")
	userID := c.Query("user_id")

	if _, err := uuid.Parse(channelID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	if _, err := uuid.Parse(userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	query := `
		SELECT id, expires_at, reason
		FROM chat_bans 
		WHERE channel_id = $1 AND user_id = $2 
		AND (expires_at IS NULL OR expires_at > NOW())`

	var ban models.ChatBan
	err := h.db.QueryRow(c.Request.Context(), query, channelID, userID).Scan(
		&ban.ID, &ban.ExpiresAt, &ban.Reason)

	if err == pgx.ErrNoRows {
		c.JSON(http.StatusOK, gin.H{"is_banned": false})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check ban status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"is_banned":  true,
		"ban_id":     ban.ID,
		"expires_at": ban.ExpiresAt,
		"reason":     ban.Reason,
	})
}

// DeleteChannel deletes a chat channel (only by owner/creator)
func (h *ChatHandler) DeleteChannel(c *gin.Context) {
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

	// Check if user is the creator
	var creatorID uuid.UUID
	err = h.db.QueryRow(c.Request.Context(), "SELECT creator_id FROM chat_channels WHERE id = $1", channelUUID).Scan(&creatorID)
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify channel"})
		return
	}

	if creatorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the channel creator can delete the channel"})
		return
	}

	// Delete channel (cascade will delete related records)
	deleteQuery := `DELETE FROM chat_channels WHERE id = $1`
	result, err := h.db.Exec(c.Request.Context(), deleteQuery, channelUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete channel"})
		return
	}

	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// AddChannelMember adds a member to a channel
func (h *ChatHandler) AddChannelMember(c *gin.Context) {
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

	inviterID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.AddChannelMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	targetUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID in request"})
		return
	}

	// Check if inviter has permission (must be owner, admin, or moderator)
	var inviterRole string
	err = h.db.QueryRow(c.Request.Context(),
		`SELECT role FROM channel_members WHERE channel_id = $1 AND user_id = $2`,
		channelUUID, inviterID).Scan(&inviterRole)
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not a member of this channel"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify membership"})
		return
	}

	if inviterRole != "owner" && inviterRole != "admin" && inviterRole != "moderator" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only channel staff can add members"})
		return
	}

	// Set default role
	role := "member"
	if req.Role != "" {
		role = req.Role
	}

	// Only owner can add admins or moderators
	if (role == "admin" || role == "moderator") && inviterRole != "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the channel owner can add admins or moderators"})
		return
	}

	// Check if target user exists
	var userExists bool
	err = h.db.QueryRow(c.Request.Context(),
		`SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, targetUserID).Scan(&userExists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify user"})
		return
	}
	if !userExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Add member
	query := `
		INSERT INTO channel_members (channel_id, user_id, role, invited_by)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (channel_id, user_id) DO NOTHING
		RETURNING id, joined_at
	`

	var member models.ChannelMember
	err = h.db.QueryRow(c.Request.Context(), query,
		channelUUID, targetUserID, role, inviterID).Scan(&member.ID, &member.JoinedAt)

	if err == pgx.ErrNoRows {
		c.JSON(http.StatusConflict, gin.H{"error": "User is already a member"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member"})
		return
	}

	member.ChannelID = channelUUID
	member.UserID = targetUserID
	member.Role = role
	member.InvitedBy = &inviterID

	c.JSON(http.StatusCreated, member)
}

// RemoveChannelMember removes a member from a channel
func (h *ChatHandler) RemoveChannelMember(c *gin.Context) {
	channelID := c.Param("id")
	memberUserID := c.Param("user_id")

	channelUUID, err := uuid.Parse(channelID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	memberUUID, err := uuid.Parse(memberUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get authenticated user
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	removerID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	// Check remover's role
	var removerRole string
	err = h.db.QueryRow(c.Request.Context(),
		`SELECT role FROM channel_members WHERE channel_id = $1 AND user_id = $2`,
		channelUUID, removerID).Scan(&removerRole)
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not a member of this channel"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify membership"})
		return
	}

	// Check target member's role
	var targetRole string
	err = h.db.QueryRow(c.Request.Context(),
		`SELECT role FROM channel_members WHERE channel_id = $1 AND user_id = $2`,
		channelUUID, memberUUID).Scan(&targetRole)
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "User is not a member of this channel"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify target membership"})
		return
	}

	// Users can remove themselves
	if removerID == memberUUID {
		// Owners cannot leave their own channel
		if targetRole == "owner" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Channel owner cannot leave the channel"})
			return
		}
	} else {
		// Only owner and admin can remove other members
		if removerRole != "owner" && removerRole != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only channel owner or admins can remove members"})
			return
		}

		// Cannot remove the owner
		if targetRole == "owner" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Cannot remove the channel owner"})
			return
		}

		// Only owner can remove admins
		if targetRole == "admin" && removerRole != "owner" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only the channel owner can remove admins"})
			return
		}
	}

	// Remove member
	deleteQuery := `DELETE FROM channel_members WHERE channel_id = $1 AND user_id = $2`
	result, err := h.db.Exec(c.Request.Context(), deleteQuery, channelUUID, memberUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member"})
		return
	}

	if result.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Member not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}

// ListChannelMembers lists all members of a channel
func (h *ChatHandler) ListChannelMembers(c *gin.Context) {
	channelID := c.Param("id")
	channelUUID, err := uuid.Parse(channelID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	// Parse pagination
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	query := `
		SELECT cm.id, cm.channel_id, cm.user_id, cm.role, cm.joined_at, cm.invited_by,
		       u.username, u.display_name, u.avatar_url
		FROM channel_members cm
		JOIN users u ON cm.user_id = u.id
		WHERE cm.channel_id = $1
		ORDER BY cm.joined_at ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := h.db.Query(c.Request.Context(), query, channelUUID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch members"})
		return
	}
	defer rows.Close()

	members := []models.ChannelMember{}
	for rows.Next() {
		var member models.ChannelMember
		err := rows.Scan(
			&member.ID,
			&member.ChannelID,
			&member.UserID,
			&member.Role,
			&member.JoinedAt,
			&member.InvitedBy,
			&member.Username,
			&member.DisplayName,
			&member.AvatarURL,
		)
		if err != nil {
			continue
		}
		members = append(members, member)
	}

	c.JSON(http.StatusOK, gin.H{
		"members": members,
		"limit":   limit,
		"offset":  offset,
	})
}

// UpdateChannelMemberRole updates a member's role in a channel
func (h *ChatHandler) UpdateChannelMemberRole(c *gin.Context) {
	channelID := c.Param("id")
	memberUserID := c.Param("user_id")

	channelUUID, err := uuid.Parse(channelID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	memberUUID, err := uuid.Parse(memberUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get authenticated user
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	updaterID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.UpdateChannelMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check updater's role (must be owner)
	var updaterRole string
	err = h.db.QueryRow(c.Request.Context(),
		`SELECT role FROM channel_members WHERE channel_id = $1 AND user_id = $2`,
		channelUUID, updaterID).Scan(&updaterRole)
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not a member of this channel"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify membership"})
		return
	}

	if updaterRole != "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the channel owner can update member roles"})
		return
	}

	// Cannot change owner role
	var targetRole string
	err = h.db.QueryRow(c.Request.Context(),
		`SELECT role FROM channel_members WHERE channel_id = $1 AND user_id = $2`,
		channelUUID, memberUUID).Scan(&targetRole)
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "User is not a member of this channel"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify target membership"})
		return
	}

	if targetRole == "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot change the owner's role"})
		return
	}

	// Update role
	updateQuery := `
		UPDATE channel_members 
		SET role = $1 
		WHERE channel_id = $2 AND user_id = $3
		RETURNING id, channel_id, user_id, role, joined_at, invited_by
	`

	var member models.ChannelMember
	err = h.db.QueryRow(c.Request.Context(), updateQuery,
		req.Role, channelUUID, memberUUID).Scan(
		&member.ID, &member.ChannelID, &member.UserID, &member.Role, &member.JoinedAt, &member.InvitedBy)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update member role"})
		return
	}

	c.JSON(http.StatusOK, member)
}

// GetCurrentUserRole gets the current authenticated user's role in a channel
func (h *ChatHandler) GetCurrentUserRole(c *gin.Context) {
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

	// Get user's role in the channel
	var role string
	err = h.db.QueryRow(c.Request.Context(),
		`SELECT role FROM channel_members WHERE channel_id = $1 AND user_id = $2`,
		channelUUID, userID).Scan(&role)
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "You are not a member of this channel"})
		return
	} else if err != nil {
		// Log error for debugging but don't expose internal details to client
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve channel membership"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"role": role})
}
