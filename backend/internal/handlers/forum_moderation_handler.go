package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ForumModerationHandler handles forum moderation operations
type ForumModerationHandler struct {
	db *pgxpool.Pool
}

// NewForumModerationHandler creates a new ForumModerationHandler
func NewForumModerationHandler(db *pgxpool.Pool) *ForumModerationHandler {
	return &ForumModerationHandler{
		db: db,
	}
}

// FlaggedContent represents flagged content in the queue
type FlaggedContent struct {
	ID         uuid.UUID  `json:"id"`
	TargetType string     `json:"target_type"`
	TargetID   uuid.UUID  `json:"target_id"`
	Reason     string     `json:"reason"`
	Details    *string    `json:"details,omitempty"`
	Status     string     `json:"status"`
	UserID     uuid.UUID  `json:"user_id"`
	Username   string     `json:"username"`
	Title      *string    `json:"title,omitempty"`
	Content    string     `json:"content"`
	FlagCount  int        `json:"flag_count"`
	CreatedAt  time.Time  `json:"created_at"`
	ReviewedBy *uuid.UUID `json:"reviewed_by,omitempty"`
	ReviewedAt *time.Time `json:"reviewed_at,omitempty"`
}

// ModerationAction represents a moderation action log entry
type ModerationAction struct {
	ID          uuid.UUID `json:"id"`
	ModeratorID uuid.UUID `json:"moderator_id"`
	Moderator   string    `json:"moderator"`
	ActionType  string    `json:"action_type"`
	TargetType  string    `json:"target_type"`
	TargetID    uuid.UUID `json:"target_id"`
	Reason      *string   `json:"reason,omitempty"`
	Metadata    *string   `json:"metadata,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// UserBan represents a user ban record
type UserBan struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	Username  string     `json:"username"`
	BannedBy  uuid.UUID  `json:"banned_by"`
	Moderator string     `json:"moderator"`
	Reason    string     `json:"reason"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Active    bool       `json:"active"`
	CreatedAt time.Time  `json:"created_at"`
}

// GetFlaggedContent retrieves flagged content for moderation
// GET /api/admin/forum/flagged
func (h *ForumModerationHandler) GetFlaggedContent(c *gin.Context) {
	status := c.DefaultQuery("status", "pending")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 || limit > 100 {
		limit = 50
	}

	query := `
		SELECT 
			cf.id, cf.target_type, cf.target_id, cf.reason, cf.details, 
			cf.status, cf.user_id, u.username, cf.created_at,
			cf.reviewed_by, cf.reviewed_at,
			CASE 
				WHEN cf.target_type = 'thread' THEN ft.title
				ELSE NULL
			END as title,
			CASE 
				WHEN cf.target_type = 'thread' THEN ft.content
				WHEN cf.target_type = 'reply' THEN fr.content
				ELSE ''
			END as content,
			CASE 
				WHEN cf.target_type = 'thread' THEN ft.flag_count
				ELSE 0
			END as flag_count
		FROM content_flags cf
		JOIN users u ON cf.user_id = u.id
		LEFT JOIN forum_threads ft ON cf.target_type = 'thread' AND cf.target_id = ft.id
		LEFT JOIN forum_replies fr ON cf.target_type = 'reply' AND cf.target_id = fr.id
		WHERE cf.status = $1
		ORDER BY cf.created_at DESC
		LIMIT $2
	`

	rows, err := h.db.Query(c.Request.Context(), query, status, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve flagged content",
		})
		return
	}
	defer rows.Close()

	var flaggedContent []FlaggedContent
	for rows.Next() {
		var item FlaggedContent
		err := rows.Scan(
			&item.ID, &item.TargetType, &item.TargetID, &item.Reason, &item.Details,
			&item.Status, &item.UserID, &item.Username, &item.CreatedAt,
			&item.ReviewedBy, &item.ReviewedAt, &item.Title, &item.Content, &item.FlagCount,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to scan flagged content",
			})
			return
		}
		flaggedContent = append(flaggedContent, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    flaggedContent,
		"meta": gin.H{
			"count":  len(flaggedContent),
			"limit":  limit,
			"status": status,
		},
	})
}

// LockThreadRequest represents the request to lock a thread
type LockThreadRequest struct {
	Reason string `json:"reason"`
	Locked bool   `json:"locked"`
}

// LockThread locks or unlocks a forum thread
// POST /api/admin/forum/threads/:id/lock
func (h *ForumModerationHandler) LockThread(c *gin.Context) {
	threadID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid thread ID",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var req LockThreadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	tx, err := h.db.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start transaction",
		})
		return
	}
	defer tx.Rollback(c.Request.Context())

	// Check if thread exists and update locked status
	result, err := tx.Exec(c.Request.Context(),
		`UPDATE forum_threads SET locked = $1, updated_at = NOW() WHERE id = $2 AND is_deleted = FALSE`,
		req.Locked, threadID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update thread",
		})
		return
	}

	// Check if thread was found and updated
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Thread not found or already deleted",
		})
		return
	}

	// Log moderation action
	actionType := "lock_thread"
	if !req.Locked {
		actionType = "unlock_thread"
	}

	_, err = tx.Exec(c.Request.Context(),
		`INSERT INTO moderation_actions (moderator_id, action_type, target_type, target_id, reason)
		VALUES ($1, $2, 'thread', $3, $4)`,
		userID, actionType, threadID, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to log moderation action",
		})
		return
	}

	if err := tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to commit transaction",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"status":  actionType,
	})
}

// PinThreadRequest represents the request to pin a thread
type PinThreadRequest struct {
	Reason string `json:"reason"`
	Pinned bool   `json:"pinned"`
}

// PinThread pins or unpins a forum thread
// POST /api/admin/forum/threads/:id/pin
func (h *ForumModerationHandler) PinThread(c *gin.Context) {
	threadID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid thread ID",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var req PinThreadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	tx, err := h.db.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start transaction",
		})
		return
	}
	defer tx.Rollback(c.Request.Context())

	// Check if thread exists and update pinned status
	result, err := tx.Exec(c.Request.Context(),
		`UPDATE forum_threads SET pinned = $1, updated_at = NOW() WHERE id = $2 AND is_deleted = FALSE`,
		req.Pinned, threadID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update thread",
		})
		return
	}

	// Check if thread was found and updated
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Thread not found or already deleted",
		})
		return
	}

	// Log moderation action
	actionType := "pin_thread"
	if !req.Pinned {
		actionType = "unpin_thread"
	}

	_, err = tx.Exec(c.Request.Context(),
		`INSERT INTO moderation_actions (moderator_id, action_type, target_type, target_id, reason)
		VALUES ($1, $2, 'thread', $3, $4)`,
		userID, actionType, threadID, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to log moderation action",
		})
		return
	}

	if err := tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to commit transaction",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"status":  actionType,
	})
}

// DeleteThreadRequest represents the request to delete a thread
type DeleteThreadRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// DeleteThread soft deletes a forum thread
// POST /api/admin/forum/threads/:id/delete
func (h *ForumModerationHandler) DeleteThread(c *gin.Context) {
	threadID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid thread ID",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var req DeleteThreadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Reason is required",
		})
		return
	}

	// Validate reason is not empty
	if strings.TrimSpace(req.Reason) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Reason must not be empty",
		})
		return
	}

	tx, err := h.db.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start transaction",
		})
		return
	}
	defer tx.Rollback(c.Request.Context())

	// Check if thread exists and soft delete it
	result, err := tx.Exec(c.Request.Context(),
		`UPDATE forum_threads SET is_deleted = TRUE, updated_at = NOW() WHERE id = $1 AND is_deleted = FALSE`,
		threadID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete thread",
		})
		return
	}

	// Check if thread was found and deleted
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Thread not found or already deleted",
		})
		return
	}

	// Log moderation action
	_, err = tx.Exec(c.Request.Context(),
		`INSERT INTO moderation_actions (moderator_id, action_type, target_type, target_id, reason)
		VALUES ($1, 'delete_thread', 'thread', $2, $3)`,
		userID, threadID, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to log moderation action",
		})
		return
	}

	if err := tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to commit transaction",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"status":  "deleted",
	})
}

// BanUserRequest represents the request to ban a user
type BanUserRequest struct {
	Reason       string `json:"reason" binding:"required"`
	DurationDays int    `json:"duration_days"` // 0 = permanent
}

// BanUser bans a user from the forum
// POST /api/admin/forum/users/:id/ban
func (h *ForumModerationHandler) BanUser(c *gin.Context) {
	targetUserID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	moderatorID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var req BanUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Validate reason is not empty
	if strings.TrimSpace(req.Reason) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Reason must not be empty",
		})
		return
	}

	// Validate duration is not negative
	if req.DurationDays < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Duration days must be non-negative",
		})
		return
	}

	// Prevent self-ban
	if moderatorID == targetUserID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot ban yourself",
		})
		return
	}

	var expiresAt *time.Time
	if req.DurationDays > 0 {
		expires := time.Now().AddDate(0, 0, req.DurationDays)
		expiresAt = &expires
	}

	tx, err := h.db.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start transaction",
		})
		return
	}
	defer tx.Rollback(c.Request.Context())

	// Check if user exists and is not already banned
	var isBanned bool
	err = tx.QueryRow(c.Request.Context(),
		`SELECT is_banned FROM users WHERE id = $1`,
		targetUserID).Scan(&isBanned)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check user status",
		})
		return
	}

	if isBanned {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User is already banned",
		})
		return
	}

	// Insert ban record
	_, err = tx.Exec(c.Request.Context(),
		`INSERT INTO user_bans (user_id, banned_by, reason, expires_at)
		VALUES ($1, $2, $3, $4)`,
		targetUserID, moderatorID, req.Reason, expiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create ban",
		})
		return
	}

	// Update user banned status
	_, err = tx.Exec(c.Request.Context(),
		`UPDATE users SET is_banned = TRUE WHERE id = $1`,
		targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update user status",
		})
		return
	}

	// Log moderation action
	_, err = tx.Exec(c.Request.Context(),
		`INSERT INTO moderation_actions (moderator_id, action_type, target_type, target_id, reason)
		VALUES ($1, 'ban_user', 'user', $2, $3)`,
		moderatorID, targetUserID, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to log moderation action",
		})
		return
	}

	if err := tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to commit transaction",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"status":  "banned",
	})
}

// GetModerationLog retrieves the moderation action log
// GET /api/admin/forum/moderation-log
func (h *ForumModerationHandler) GetModerationLog(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 || limit > 100 {
		limit = 50
	}

	actionType := c.Query("action_type")
	targetType := c.Query("target_type")

	// Use parameterized query with OR clauses to avoid dynamic SQL construction
	args := []interface{}{actionType, targetType, limit}

	query := `
		SELECT 
			ma.id, ma.moderator_id, u.username as moderator,
			ma.action_type, ma.target_type, ma.target_id,
			ma.reason, ma.metadata, ma.created_at
		FROM moderation_actions ma
		JOIN users u ON ma.moderator_id = u.id
		WHERE 
			($1 = '' OR ma.action_type = $1) AND
			($2 = '' OR ma.target_type = $2)
		ORDER BY ma.created_at DESC LIMIT $3
	`

	rows, err := h.db.Query(c.Request.Context(), query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve moderation log",
		})
		return
	}
	defer rows.Close()

	var actions []ModerationAction
	for rows.Next() {
		var action ModerationAction
		var metadata sql.NullString

		err := rows.Scan(
			&action.ID, &action.ModeratorID, &action.Moderator,
			&action.ActionType, &action.TargetType, &action.TargetID,
			&action.Reason, &metadata, &action.CreatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to scan moderation action",
			})
			return
		}

		if metadata.Valid {
			action.Metadata = &metadata.String
		}

		actions = append(actions, action)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    actions,
		"meta": gin.H{
			"count": len(actions),
			"limit": limit,
		},
	})
}

// FlagContentRequest represents the request to flag content
type FlagContentRequest struct {
	TargetType string  `json:"target_type" binding:"required"`
	TargetID   string  `json:"target_id" binding:"required"`
	Reason     string  `json:"reason" binding:"required"`
	Details    *string `json:"details"`
}

// FlagContent allows authenticated users to flag a thread or reply
// POST /api/v1/forum/flag
func (h *ForumModerationHandler) FlagContent(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var req FlagContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Validate target_type
	if req.TargetType != "thread" && req.TargetType != "reply" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "target_type must be 'thread' or 'reply'",
		})
		return
	}

	// Validate reason
	validReasons := map[string]bool{
		"spam":            true,
		"harassment":      true,
		"off-topic":       true,
		"misinformation":  true,
		"other":           true,
	}
	if !validReasons[req.Reason] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "reason must be one of: spam, harassment, off-topic, misinformation, other",
		})
		return
	}

	// Validate target_id as UUID
	targetID, err := uuid.Parse(req.TargetID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid target_id",
		})
		return
	}

	// Insert flag, handling duplicates gracefully
	_, err = h.db.Exec(c.Request.Context(),
		`INSERT INTO content_flags (user_id, target_type, target_id, reason, details, status)
		VALUES ($1, $2, $3, $4, $5, 'pending')
		ON CONFLICT (user_id, target_type, target_id) DO NOTHING`,
		userID, req.TargetType, targetID, req.Reason, req.Details)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to flag content",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Content has been flagged for review",
	})
}

// GetUserBans retrieves active user bans
// GET /api/admin/forum/bans
func (h *ForumModerationHandler) GetUserBans(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 || limit > 100 {
		limit = 50
	}

	activeOnly := c.DefaultQuery("active", "true") == "true"

	query := `
		SELECT 
			ub.id, ub.user_id, u.username,
			ub.banned_by, m.username as moderator,
			ub.reason, ub.expires_at, ub.active, ub.created_at
		FROM user_bans ub
		JOIN users u ON ub.user_id = u.id
		JOIN users m ON ub.banned_by = m.id
	`

	if activeOnly {
		query += ` WHERE ub.active = TRUE`
	}

	query += ` ORDER BY ub.created_at DESC LIMIT $1`

	rows, err := h.db.Query(c.Request.Context(), query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve bans",
		})
		return
	}
	defer rows.Close()

	var bans []UserBan
	for rows.Next() {
		var ban UserBan
		err := rows.Scan(
			&ban.ID, &ban.UserID, &ban.Username,
			&ban.BannedBy, &ban.Moderator,
			&ban.Reason, &ban.ExpiresAt, &ban.Active, &ban.CreatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to scan ban record",
			})
			return
		}
		bans = append(bans, ban)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    bans,
		"meta": gin.H{
			"count":       len(bans),
			"limit":       limit,
			"active_only": activeOnly,
		},
	})
}
