package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func writeAdminUserMutationError(c *gin.Context, err error, message string) {
	switch {
	case errors.Is(err, repository.ErrUserNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
	case errors.Is(err, repository.ErrAdminRequired), errors.Is(err, repository.ErrAdminSelfMutation), errors.Is(err, repository.ErrProtectedAdminTarget):
		c.JSON(http.StatusForbidden, gin.H{"error": "User mutation not permitted"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": message})
	}
}

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
	search := strings.TrimSpace(c.Query("search"))
	if len(search) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search cannot exceed 200 characters"})
		return
	}
	role := c.Query("role")
	status := c.Query("status")
	if role != "" && !models.IsValidRole(role) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role must be user, moderator, or admin"})
		return
	}
	if status != "" && status != "active" && status != "banned" && status != "unclaimed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status must be active, banned, or unclaimed"})
		return
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page must be a positive integer"})
		return
	}

	perPage, err := strconv.Atoi(c.DefaultQuery("per_page", "25"))
	if err != nil || perPage < 1 || perPage > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "per_page must be between 1 and 100"})
		return
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

// ListPlatformModerators handles GET /api/v1/admin/moderators.
func (h *AdminUserHandler) ListPlatformModerators(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page must be a positive integer"})
		return
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "25"))
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 100"})
		return
	}
	users, total, err := h.userRepo.AdminSearchUsers(c.Request.Context(), "", models.RoleModerator, "", limit, (page-1)*limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve platform moderators"})
		return
	}
	totalPages := 0
	if total > 0 {
		totalPages = (total + limit - 1) / limit
	}
	c.JSON(http.StatusOK, gin.H{"items": users, "total": total, "page": page, "limit": limit, "total_pages": totalPages})
}

func (h *AdminUserHandler) applyPlatformModeratorRole(c *gin.Context, userID uuid.UUID, role, reason string) {
	actorID, ok := authenticatedUserID(c)
	if !ok {
		return
	}
	if reason == "" {
		reason = "Platform moderator role changed to " + role
	}
	if err := h.userRepo.ApplyAdminUserMutation(c.Request.Context(), userID, actorID, repository.AdminUserActionRole, role, reason); err != nil {
		writeAdminUserMutationError(c, err, "Failed to update platform moderator")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Platform moderator updated successfully"})
}

// AddPlatformModerator handles POST /api/v1/admin/moderators.
func (h *AdminUserHandler) AddPlatformModerator(c *gin.Context) {
	var req struct {
		UserID uuid.UUID `json:"user_id" binding:"required"`
		Reason string    `json:"reason" binding:"omitempty,max=1000"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.UserID == uuid.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A valid user_id is required"})
		return
	}
	h.applyPlatformModeratorRole(c, req.UserID, models.RoleModerator, req.Reason)
}

// UpdatePlatformModerator handles PATCH /api/v1/admin/moderators/:id.
func (h *AdminUserHandler) UpdatePlatformModerator(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	var req struct {
		Reason string `json:"reason" binding:"omitempty,max=1000"`
	}
	if c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
	}
	h.applyPlatformModeratorRole(c, userID, models.RoleModerator, req.Reason)
}

// RevokePlatformModerator handles DELETE /api/v1/admin/moderators/:id.
func (h *AdminUserHandler) RevokePlatformModerator(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	var req struct {
		Reason string `json:"reason" binding:"omitempty,max=1000"`
	}
	if c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
	}
	h.applyPlatformModeratorRole(c, userID, models.RoleUser, req.Reason)
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
		Reason string `json:"reason" binding:"required,min=3,max=1000"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Ban reason is required",
		})
		return
	}

	// Get admin user ID
	adminUserID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	// Ban the user
	err = h.userRepo.ApplyAdminUserMutation(c.Request.Context(), userID, adminUserID, repository.AdminUserActionBan, nil, req.Reason)
	if err != nil {
		writeAdminUserMutationError(c, err, "Failed to ban user")
		return
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
		Reason string `json:"reason" binding:"omitempty,max=1000"`
	}
	if c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unban request"})
			return
		}
	}

	// Get admin user ID
	adminUserID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	// Unban the user
	reason := "No reason provided"
	if req.Reason != "" {
		reason = req.Reason
	}
	err = h.userRepo.ApplyAdminUserMutation(c.Request.Context(), userID, adminUserID, repository.AdminUserActionUnban, nil, reason)
	if err != nil {
		writeAdminUserMutationError(c, err, "Failed to unban user")
		return
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
		Role   string `json:"role" binding:"required,oneof=user moderator admin"`
		Reason string `json:"reason" binding:"omitempty,max=1000"`
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
	adminUserID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	// Update user role
	reason := req.Reason
	if reason == "" {
		reason = "Role changed to " + req.Role
	}
	err = h.userRepo.ApplyAdminUserMutation(c.Request.Context(), userID, adminUserID, repository.AdminUserActionRole, req.Role, reason)
	if err != nil {
		writeAdminUserMutationError(c, err, "Failed to update user role")
		return
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
		KarmaPoints *int `json:"karma_points" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}
	if *req.KarmaPoints < -1_000_000 || *req.KarmaPoints > 1_000_000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "karma_points must be between -1000000 and 1000000"})
		return
	}

	// Get admin user ID
	adminUserID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	// Set user karma
	reason := "Karma manually adjusted by admin"
	err = h.userRepo.ApplyAdminUserMutation(c.Request.Context(), userID, adminUserID, repository.AdminUserActionKarma, *req.KarmaPoints, reason)
	if err != nil {
		writeAdminUserMutationError(c, err, "Failed to update user uppies")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "User uppies updated successfully",
		"karma_points": *req.KarmaPoints,
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
	adminUserID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	// Apply suspension via repository
	err = h.userRepo.SuspendCommentPrivileges(
		c.Request.Context(),
		userID,
		adminUserID,
		req.SuspensionType,
		req.Reason,
		req.DurationHours,
	)

	if err != nil {
		writeAdminUserMutationError(c, err, "Failed to suspend comment privileges")
		return
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
	adminUserID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	// Lift suspension via repository
	err = h.userRepo.LiftCommentSuspension(
		c.Request.Context(),
		userID,
		adminUserID,
		req.Reason,
	)

	if err != nil {
		writeAdminUserMutationError(c, err, "Failed to lift comment suspension")
		return
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 100"})
		return
	}
	if _, err := h.userRepo.GetByID(c.Request.Context(), userID); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
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
		Reason        string `json:"reason" binding:"required,min=3,max=1000"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
		})
		return
	}

	// Get admin user ID
	adminUserID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	// Toggle review requirement via repository
	action := repository.AdminUserActionReview
	err = h.userRepo.ApplyAdminUserMutation(c.Request.Context(), userID, adminUserID, action, req.RequireReview, req.Reason)

	if err != nil {
		writeAdminUserMutationError(c, err, "Failed to update comment review requirement")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Comment review requirement updated successfully",
		"require_review": req.RequireReview,
	})
}
