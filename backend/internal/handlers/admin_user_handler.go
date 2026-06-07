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

// AdminUserHandler handles admin user management endpoints
type AdminUserHandler struct {
	userRepo     *repository.UserRepository
	auditLogRepo *repository.AuditLogRepository
	authService  *services.AuthService
}

// NewAdminUserHandler creates a new admin user handler
func NewAdminUserHandler(
	userRepo *repository.UserRepository,
	auditLogRepo *repository.AuditLogRepository,
	authService *services.AuthService,
) *AdminUserHandler {
	return &AdminUserHandler{
		userRepo:     userRepo,
		auditLogRepo: auditLogRepo,
		authService:  authService,
	}
}

// ListUsers handles GET /api/v1/admin/users
func (h *AdminUserHandler) ListUsers(c *gin.Context) {
	// Get query parameters
	search := c.Query("search")
	role := c.Query("role")
	status := c.Query("status")

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	perPage, err := strconv.Atoi(c.DefaultQuery("per_page", "25"))
	if err != nil || perPage < 1 || perPage > 100 {
		perPage = 25
	}

	offset := (page - 1) * perPage

	// Search users with filters
	users, total, err := h.userRepo.AdminSearchUsers(
		c.Request.Context(),
		search,
		role,
		status,
		perPage,
		offset,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve users",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users":    users,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}

// BanUser handles POST /api/v1/admin/users/:id/ban
func (h *AdminUserHandler) BanUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	// Get reason from request body
	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Ban reason is required",
		})
		return
	}

	// Get admin user ID
	adminUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	// Ban the user
	err = h.userRepo.BanUser(c.Request.Context(), userID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to ban user",
		})
		return
	}

	// Log audit event
	auditLog := &models.ModerationAuditLog{
		ID:          uuid.New(),
		Action:      "ban_user",
		EntityType:  "user",
		EntityID:    userID,
		ModeratorID: adminUserID.(uuid.UUID),
		Reason:      &req.Reason,
	}
	if err := h.auditLogRepo.Create(c.Request.Context(), auditLog); err != nil {
		// Record audit log failure without affecting the main operation
		_ = c.Error(err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User banned successfully",
	})
}

// UnbanUser handles POST /api/v1/admin/users/:id/unban
func (h *AdminUserHandler) UnbanUser(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	// Get optional reason from request body
	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)

	// Get admin user ID
	adminUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	// Unban the user
	err = h.userRepo.UnbanUser(c.Request.Context(), userID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to unban user",
		})
		return
	}

	// Log audit event
	reason := "No reason provided"
	if req.Reason != "" {
		reason = req.Reason
	}
	auditLog := &models.ModerationAuditLog{
		ID:          uuid.New(),
		Action:      "unban_user",
		EntityType:  "user",
		EntityID:    userID,
		ModeratorID: adminUserID.(uuid.UUID),
		Reason:      &reason,
	}
	if err := h.auditLogRepo.Create(c.Request.Context(), auditLog); err != nil {
		// Record audit log failure without affecting the main operation
		_ = c.Error(err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User unbanned successfully",
	})
}

// UpdateUserRole handles PATCH /api/v1/admin/users/:id/role
func (h *AdminUserHandler) UpdateUserRole(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	// Get role and reason from request body
	var req struct {
		Role   string `json:"role" binding:"required"`
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Validate role
	if req.Role != "user" && req.Role != "moderator" && req.Role != "admin" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid role. Must be user, moderator, or admin",
		})
		return
	}

	// Get admin user ID
	adminUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	// Update user role
	err = h.userRepo.UpdateUserRole(c.Request.Context(), userID, req.Role)
	if err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update user role",
		})
		return
	}

	// Log audit event
	reason := req.Reason
	if reason == "" {
		reason = "Role changed to " + req.Role
	}
	auditLog := &models.ModerationAuditLog{
		ID:          uuid.New(),
		Action:      "update_user_role",
		EntityType:  "user",
		EntityID:    userID,
		ModeratorID: adminUserID.(uuid.UUID),
		Reason:      &reason,
	}
	if err := h.auditLogRepo.Create(c.Request.Context(), auditLog); err != nil {
		// Record audit log failure without affecting the main operation
		_ = c.Error(err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User role updated successfully",
	})
}

// UpdateUserKarma handles PATCH /api/v1/admin/users/:id/karma
func (h *AdminUserHandler) UpdateUserKarma(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	// Get karma points from request body
	var req struct {
		KarmaPoints int `json:"karma_points" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	// Get admin user ID
	adminUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	// Set user karma
	err = h.userRepo.SetUserKarma(c.Request.Context(), userID, req.KarmaPoints)
	if err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update user karma",
		})
		return
	}

	// Log audit event
	reason := "Karma manually adjusted by admin"
	auditLog := &models.ModerationAuditLog{
		ID:          uuid.New(),
		Action:      "update_user_karma",
		EntityType:  "user",
		EntityID:    userID,
		ModeratorID: adminUserID.(uuid.UUID),
		Reason:      &reason,
	}
	if err := h.auditLogRepo.Create(c.Request.Context(), auditLog); err != nil {
		// Record audit log failure without affecting the main operation
		_ = c.Error(err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "User karma updated successfully",
		"karma_points": req.KarmaPoints,
	})
}

// SuspendCommentPrivileges handles POST /api/v1/admin/users/:id/suspend-comments
func (h *AdminUserHandler) SuspendCommentPrivileges(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	var req models.CommentSuspensionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
		})
		return
	}

	// Validate duration is provided for temporary suspensions
	if req.SuspensionType == models.SuspensionTypeTemporary && req.DurationHours == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Duration is required for temporary suspensions",
		})
		return
	}

	// Get admin user ID
	adminUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	// Apply suspension via repository
	err = h.userRepo.SuspendCommentPrivileges(
		c.Request.Context(),
		userID,
		adminUserID.(uuid.UUID),
		req.SuspensionType,
		req.Reason,
		req.DurationHours,
	)

	if err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to suspend comment privileges",
		})
		return
	}

	// Log audit event
	auditLog := &models.ModerationAuditLog{
		ID:          uuid.New(),
		Action:      "suspend_comment_privileges",
		EntityType:  "user",
		EntityID:    userID,
		ModeratorID: adminUserID.(uuid.UUID),
		Reason:      &req.Reason,
	}
	if err := h.auditLogRepo.Create(c.Request.Context(), auditLog); err != nil {
		// Record audit log failure without affecting the main operation
		_ = c.Error(err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Comment privileges suspended successfully",
		"suspension_type": req.SuspensionType,
	})
}

// LiftCommentSuspension handles POST /api/v1/admin/users/:id/lift-comment-suspension
func (h *AdminUserHandler) LiftCommentSuspension(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	var req models.LiftSuspensionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Reason is required",
		})
		return
	}

	// Get admin user ID
	adminUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	// Lift suspension via repository
	err = h.userRepo.LiftCommentSuspension(
		c.Request.Context(),
		userID,
		adminUserID.(uuid.UUID),
		req.Reason,
	)

	if err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to lift comment suspension",
		})
		return
	}

	// Log audit event
	auditLog := &models.ModerationAuditLog{
		ID:          uuid.New(),
		Action:      "lift_comment_suspension",
		EntityType:  "user",
		EntityID:    userID,
		ModeratorID: adminUserID.(uuid.UUID),
		Reason:      &req.Reason,
	}
	if err := h.auditLogRepo.Create(c.Request.Context(), auditLog); err != nil {
		// Record audit log failure without affecting the main operation
		_ = c.Error(err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Comment suspension lifted successfully",
	})
}

// GetCommentSuspensionHistory handles GET /api/v1/admin/users/:id/comment-suspension-history
func (h *AdminUserHandler) GetCommentSuspensionHistory(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 50
	}

	history, err := h.userRepo.GetCommentSuspensionHistory(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve suspension history",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"history": history,
		"count":   len(history),
	})
}

// ToggleCommentReview handles POST /api/v1/admin/users/:id/toggle-comment-review
func (h *AdminUserHandler) ToggleCommentReview(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	var req struct {
		RequireReview bool   `json:"require_review"`
		Reason        string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
		})
		return
	}

	// Get admin user ID
	adminUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	// Toggle review requirement via repository
	err = h.userRepo.SetCommentReviewRequirement(
		c.Request.Context(),
		userID,
		req.RequireReview,
	)

	if err != nil {
		if err == repository.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update comment review requirement",
		})
		return
	}

	// Log audit event
	action := "enable_comment_review"
	if !req.RequireReview {
		action = "disable_comment_review"
	}
	auditLog := &models.ModerationAuditLog{
		ID:          uuid.New(),
		Action:      action,
		EntityType:  "user",
		EntityID:    userID,
		ModeratorID: adminUserID.(uuid.UUID),
		Reason:      &req.Reason,
	}
	if err := h.auditLogRepo.Create(c.Request.Context(), auditLog); err != nil {
		// Record audit log failure without affecting the main operation
		_ = c.Error(err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Comment review requirement updated successfully",
		"require_review": req.RequireReview,
	})
}
