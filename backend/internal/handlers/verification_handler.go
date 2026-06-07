package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
)

// VerificationHandler handles creator verification operations
type VerificationHandler struct {
	verificationRepo    *repository.VerificationRepository
	notificationService *services.NotificationService
	db                  *pgxpool.Pool
}

// NewVerificationHandler creates a new VerificationHandler
func NewVerificationHandler(
	verificationRepo *repository.VerificationRepository,
	notificationService *services.NotificationService,
	db *pgxpool.Pool,
) *VerificationHandler {
	return &VerificationHandler{
		verificationRepo:    verificationRepo,
		notificationService: notificationService,
		db:                  db,
	}
}

// CreateApplication creates a new verification application
// POST /api/v1/verification/applications
func (h *VerificationHandler) CreateApplication(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user ID from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	userID := userIDVal.(uuid.UUID)

	var req models.CreateVerificationApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// === ABUSE PREVENTION CHECKS ===

	// 0. Check if user is already verified
	isVerified, err := h.verificationRepo.IsUserVerified(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check verification status",
		})
		return
	}

	if isVerified {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "You are already verified",
		})
		return
	}

	// 1. Check if user already has a pending application
	existing, err := h.verificationRepo.GetApplicationByUserID(ctx, userID, models.VerificationStatusPending)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check for existing application",
		})
		return
	}

	if existing != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "You already have a pending verification application",
		})
		return
	}

	// 2. Check for recently rejected applications (30-day cooldown)
	recentRejection, err := h.verificationRepo.GetRecentRejectedApplicationByUserID(ctx, userID, 30)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check application history",
		})
		return
	}

	if recentRejection != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":       "You must wait 30 days after a rejection before reapplying",
			"rejected_at": recentRejection.ReviewedAt,
		})
		return
	}

	// 3. Check total application count to detect potential abuse
	totalApps, err := h.verificationRepo.GetApplicationCountByUserID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check application history",
		})
		return
	}

	if totalApps >= 5 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "You have reached the maximum number of verification applications",
		})
		return
	}

	// 4. Check for duplicate Twitch URLs from other users
	duplicateApps, err := h.verificationRepo.GetApplicationsByTwitchURL(ctx, req.TwitchChannelURL, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to validate Twitch channel",
		})
		return
	}

	if len(duplicateApps) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "This Twitch channel is already associated with another verification application",
		})
		return
	}

	// Create application
	app := &models.CreatorVerificationApplication{
		UserID:             userID,
		TwitchChannelURL:   req.TwitchChannelURL,
		FollowerCount:      req.FollowerCount,
		SubscriberCount:    req.SubscriberCount,
		AvgViewers:         req.AvgViewers,
		ContentDescription: req.ContentDescription,
		SocialMediaLinks:   make(map[string]interface{}),
		Status:             models.VerificationStatusPending,
		Priority:           50, // Default priority
	}

	// Convert string map to interface map
	for k, v := range req.SocialMediaLinks {
		app.SocialMediaLinks[k] = v
	}

	err = h.verificationRepo.CreateApplication(ctx, app)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create verification application",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    app,
		"message": "Verification application submitted successfully",
	})
}

// GetApplication retrieves the current user's verification application
// GET /api/v1/verification/applications/me
func (h *VerificationHandler) GetApplication(c *gin.Context) {
	ctx := c.Request.Context()

	// Get user ID from context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	userID := userIDVal.(uuid.UUID)

	// Get latest application regardless of status (empty string means no status filter)
	app, err := h.verificationRepo.GetApplicationByUserID(ctx, userID, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve verification application",
		})
		return
	}

	if app == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No verification application found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    app,
	})
}

// ListApplications lists all verification applications (admin only)
// GET /admin/verification/applications
func (h *VerificationHandler) ListApplications(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse query parameters
	status := c.DefaultQuery("status", models.VerificationStatusPending)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	// Validate status parameter
	validStatuses := map[string]bool{
		models.VerificationStatusPending:  true,
		models.VerificationStatusApproved: true,
		models.VerificationStatusRejected: true,
		"":                                true, // Allow empty for all statuses
	}
	if !validStatuses[status] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid status. Must be one of: pending, approved, rejected",
		})
		return
	}

	if limit < 1 || limit > 100 {
		limit = 50
	}
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * limit

	apps, err := h.verificationRepo.ListApplications(ctx, status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve verification applications",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    apps,
		"meta": gin.H{
			"count":  len(apps),
			"limit":  limit,
			"page":   page,
			"status": status,
		},
	})
}

// GetApplicationByID retrieves a verification application by ID (admin only)
// GET /admin/verification/applications/:id
func (h *VerificationHandler) GetApplicationByID(c *gin.Context) {
	ctx := c.Request.Context()

	appID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid application ID",
		})
		return
	}

	app, err := h.verificationRepo.GetApplicationWithUser(ctx, appID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve verification application",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    app,
	})
}

// ReviewApplication reviews a verification application (admin only)
// POST /admin/verification/applications/:id/review
func (h *VerificationHandler) ReviewApplication(c *gin.Context) {
	ctx := c.Request.Context()

	appID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid application ID",
		})
		return
	}

	// Get reviewer ID from context
	reviewerIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	reviewerID := reviewerIDVal.(uuid.UUID)

	var req models.ReviewVerificationApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Get the application first to get user ID
	app, err := h.verificationRepo.GetApplicationByID(ctx, appID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve verification application",
		})
		return
	}

	if app.Status != models.VerificationStatusPending {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Application has already been reviewed",
		})
		return
	}

	// Update application status
	var notes string
	if req.Notes != nil {
		notes = *req.Notes
	}
	err = h.verificationRepo.UpdateApplicationStatus(ctx, appID, reviewerID, req.Decision, notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update application status",
		})
		return
	}

	// Create decision audit entry
	decision := &models.CreatorVerificationDecision{
		ApplicationID: appID,
		ReviewerID:    reviewerID,
		Decision:      req.Decision,
		Notes:         req.Notes,
		Metadata:      make(map[string]interface{}),
	}

	err = h.verificationRepo.CreateDecision(ctx, decision)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create decision audit entry",
		})
		return
	}

	// Send notification to user (async - don't fail request if notification fails)
	// Use background context to avoid cancellation when HTTP request completes
	go func() {
		bgCtx := context.Background()
		var notificationType string
		var title string
		var message string

		if req.Decision == models.VerificationDecisionApproved {
			notificationType = models.NotificationTypeSystemAlert
			title = "Verification Approved"
			message = "Your creator verification application has been approved! You now have a verified badge."
		} else {
			notificationType = models.NotificationTypeSystemAlert
			title = "Verification Application Reviewed"
			message = "Your creator verification application has been reviewed."
			if notes != "" {
				message += " Note: " + notes
			}
		}

		_, _ = h.notificationService.CreateNotification(
			bgCtx,
			app.UserID,
			notificationType,
			title,
			message,
			nil, // link
			nil, // sourceUserID
			nil, // sourceContentID
			nil, // sourceContentType
		)
	}()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Application reviewed successfully",
		"data": gin.H{
			"id":       appID,
			"decision": req.Decision,
		},
	})
}

// GetApplicationStats retrieves verification application statistics (admin only)
// GET /admin/verification/stats
func (h *VerificationHandler) GetApplicationStats(c *gin.Context) {
	ctx := c.Request.Context()

	stats, err := h.verificationRepo.GetApplicationStats(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve verification statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetAuditLogs retrieves audit logs for verification (admin only)
// GET /admin/verification/audit-logs
func (h *VerificationHandler) GetAuditLogs(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse query parameters
	userIDParam := c.Query("user_id")
	onlyFlagged := c.DefaultQuery("only_flagged", "false") == "true"
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	if limit < 1 || limit > 100 {
		limit = 50
	}
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * limit

	var logs []*models.VerificationAuditLog
	var err error

	if onlyFlagged {
		logs, err = h.verificationRepo.GetFlaggedAudits(ctx, limit, offset)
	} else if userIDParam != "" {
		userID, parseErr := uuid.Parse(userIDParam)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid user_id parameter",
			})
			return
		}
		logs, err = h.verificationRepo.GetAuditLogsByUserID(ctx, userID, limit, offset)
	} else {
		// If neither flagged nor user_id specified, return flagged by default for admin view
		logs, err = h.verificationRepo.GetFlaggedAudits(ctx, limit, offset)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve audit logs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    logs,
		"meta": gin.H{
			"count": len(logs),
			"limit": limit,
			"page":  page,
		},
	})
}

// GetUserAuditHistory retrieves audit history for a specific user (admin only)
// GET /admin/verification/users/:user_id/audit-logs
func (h *VerificationHandler) GetUserAuditHistory(c *gin.Context) {
	ctx := c.Request.Context()

	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	if limit < 1 || limit > 100 {
		limit = 50
	}
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * limit

	logs, err := h.verificationRepo.GetAuditLogsByUserID(ctx, userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve audit logs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    logs,
		"meta": gin.H{
			"count": len(logs),
			"limit": limit,
			"page":  page,
		},
	})
}
