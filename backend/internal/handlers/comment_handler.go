package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
)

// CommentHandler handles comment-related HTTP requests
type CommentHandler struct {
	commentService *services.CommentService
}

// NewCommentHandler creates a new CommentHandler
func NewCommentHandler(commentService *services.CommentService) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
	}
}

// ListComments handles GET /clips/:id/comments
func (h *CommentHandler) ListComments(c *gin.Context) {
	// Parse clip ID
	clipIDStr := c.Param("id")
	clipID, err := uuid.Parse(clipIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid clip ID",
		})
		return
	}

	// Parse query parameters
	sortBy := c.DefaultQuery("sort", "best")
	limitStr := c.DefaultQuery("limit", "50")
	cursorStr := c.DefaultQuery("cursor", "0")
	includeRepliesStr := c.DefaultQuery("include_replies", "false")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 50
	}

	offset, err := strconv.Atoi(cursorStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	includeReplies := includeRepliesStr == "true"

	// Get user ID if authenticated
	var userID *uuid.UUID
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(uuid.UUID); ok {
			userID = &uid
		}
	}

	// List comments with optional nested replies
	comments, err := h.commentService.ListCommentsWithReplies(c.Request.Context(), clipID, sortBy, limit, offset, userID, includeReplies)
	if err != nil {
		// Log the actual error for debugging
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve comments",
		})
		return
	}

	// Calculate next cursor
	nextCursor := offset + len(comments)
	hasMore := len(comments) == limit

	c.JSON(http.StatusOK, gin.H{
		"comments":    comments,
		"next_cursor": nextCursor,
		"has_more":    hasMore,
	})
}

// CreateComment handles POST /clips/:id/comments
func (h *CommentHandler) CreateComment(c *gin.Context) {
	// Parse clip ID
	clipIDStr := c.Param("id")
	clipID, err := uuid.Parse(clipIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid clip ID",
		})
		return
	}

	// Get user ID from context (set by auth middleware)
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	// Parse request body
	var req services.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Create comment
	comment, err := h.commentService.CreateComment(c.Request.Context(), &req, clipID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, comment)
}

// UpdateComment handles PUT /comments/:id
func (h *CommentHandler) UpdateComment(c *gin.Context) {
	// Parse comment ID
	commentIDStr := c.Param("id")
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid comment ID",
		})
		return
	}

	// Get user ID and role from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	role, _ := c.Get("user_role")
	roleStr, _ := role.(string)
	isAdmin := roleStr == "admin"

	// Parse request body
	var req struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Update comment
	if err := h.commentService.UpdateComment(c.Request.Context(), commentID, userID, req.Content, isAdmin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Comment updated successfully",
	})
}

// DeleteComment handles DELETE /comments/:id
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	// Parse comment ID
	commentIDStr := c.Param("id")
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid comment ID",
		})
		return
	}

	// Get user ID and role from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	role, _ := c.Get("user_role")
	roleStr, _ := role.(string)

	// Parse optional reason (for mod/admin deletions)
	var req struct {
		Reason *string `json:"reason,omitempty"`
	}
	_ = c.ShouldBindJSON(&req) // Optional

	// Delete comment
	if err := h.commentService.DeleteComment(c.Request.Context(), commentID, userID, roleStr, req.Reason); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Comment deleted successfully",
	})
}

// VoteOnComment handles POST /comments/:id/vote
func (h *CommentHandler) VoteOnComment(c *gin.Context) {
	// Parse comment ID
	commentIDStr := c.Param("id")
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid comment ID",
		})
		return
	}

	// Get user ID from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required",
		})
		return
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	// Parse request body
	var req struct {
		Vote int16 `json:"vote"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Vote on comment
	if err := h.commentService.VoteOnComment(c.Request.Context(), commentID, userID, req.Vote); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Vote recorded successfully",
	})
}

// GetReplies handles GET /comments/:id/replies
func (h *CommentHandler) GetReplies(c *gin.Context) {
	// Parse comment ID
	commentIDStr := c.Param("id")
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid comment ID",
		})
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "50")
	cursorStr := c.DefaultQuery("cursor", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 50
	}

	offset, err := strconv.Atoi(cursorStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get user ID if authenticated
	var userID *uuid.UUID
	if userIDVal, exists := c.Get("user_id"); exists {
		if uid, ok := userIDVal.(uuid.UUID); ok {
			userID = &uid
		}
	}

	// Get replies
	replies, err := h.commentService.GetReplies(c.Request.Context(), commentID, limit, offset, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve replies",
		})
		return
	}

	// Calculate next cursor
	nextCursor := offset + len(replies)
	hasMore := len(replies) == limit

	c.JSON(http.StatusOK, gin.H{
		"replies":     replies,
		"next_cursor": nextCursor,
		"has_more":    hasMore,
	})
}
