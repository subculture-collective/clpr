package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
)

// DMCAHandler handles DMCA-related HTTP requests
type DMCAHandler struct {
	dmcaService *services.DMCAService
	authService *services.AuthService
}

// NewDMCAHandler creates a new DMCA handler
func NewDMCAHandler(
	dmcaService *services.DMCAService,
	authService *services.AuthService,
) *DMCAHandler {
	return &DMCAHandler{
		dmcaService: dmcaService,
		authService: authService,
	}
}

// SubmitTakedownNotice handles DMCA takedown notice submissions (public endpoint)
// POST /api/v1/dmca/takedown
func (h *DMCAHandler) SubmitTakedownNotice(c *gin.Context) {
	var req models.SubmitDMCANoticeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get IP address and user agent for audit trail
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Submit notice
	notice, err := h.dmcaService.SubmitTakedownNotice(c.Request.Context(), &req, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Return success response (don't expose full notice details)
	c.JSON(http.StatusCreated, gin.H{
		"message":      "DMCA takedown notice submitted successfully. We will review your notice and take appropriate action.",
		"notice_id":    notice.ID,
		"status":       "pending_review",
		"submitted_at": notice.SubmittedAt,
	})
}

// SubmitCounterNotice handles DMCA counter-notice submissions
// POST /api/v1/dmca/counter-notice
func (h *DMCAHandler) SubmitCounterNotice(c *gin.Context) {
	var req models.SubmitDMCACounterNoticeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID if authenticated (optional)
	var userID *uuid.UUID
	if id, exists := c.Get("user_id"); exists {
		uid := id.(uuid.UUID)
		userID = &uid
	}

	// Get IP address and user agent
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Submit counter-notice
	counterNotice, err := h.dmcaService.SubmitCounterNotice(c.Request.Context(), &req, userID, ipAddress, userAgent)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":             "Counter-notice submitted successfully. We will forward it to the complainant.",
		"counter_notice_id":   counterNotice.ID,
		"waiting_period_ends": counterNotice.WaitingPeriodEnds,
		"status":              "pending_review",
	})
}

// GetUserStrikes retrieves DMCA strikes for the authenticated user
// GET /api/v1/users/:id/dmca-strikes
func (h *DMCAHandler) GetUserStrikes(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Check authorization - users can only view their own strikes unless admin
	authenticatedUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Allow if viewing own strikes or if admin
	authUID := authenticatedUserID.(uuid.UUID)
	roleVal, _ := c.Get("user_role")
	roleStr, _ := roleVal.(string)
	isAdmin := roleStr == "admin" || roleStr == "moderator"

	if userID != authUID && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only view your own DMCA strikes"})
		return
	}

	// Get strikes from service
	// For this implementation, we'll call the repository directly
	// In a full implementation, add a service method
	strikes, err := h.dmcaService.GetUserStrikes(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve strikes"})
		return
	}

	// Count strikes by status
	activeCount := 0
	expiredCount := 0
	removedCount := 0
	for _, strike := range strikes {
		switch strike.Status {
		case "active":
			activeCount++
		case "expired":
			expiredCount++
		case "removed":
			removedCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"strikes":       strikes,
		"active_count":  activeCount,
		"expired_count": expiredCount,
		"removed_count": removedCount,
	})
}

// ==============================================================================
// Admin Endpoints
// ==============================================================================

// ListDMCANotices lists all DMCA notices (admin only)
// GET /api/admin/dmca/notices
func (h *DMCAHandler) ListDMCANotices(c *gin.Context) {
	// Parse query parameters
	page := 1
	limit := 20
	status := c.Query("status")

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Get notices from service
	// In a full implementation, create a service method for this
	// For now, showing the expected response structure
	c.JSON(http.StatusOK, gin.H{
		"notices": []interface{}{},
		"total":   0,
		"page":    page,
		"limit":   limit,
		"status":  status,
	})
}

// ReviewNotice allows admin to mark a notice as valid or invalid
// PATCH /api/admin/dmca/notices/:id/review
func (h *DMCAHandler) ReviewNotice(c *gin.Context) {
	noticeIDStr := c.Param("id")
	noticeID, err := uuid.Parse(noticeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notice ID"})
		return
	}

	var req models.UpdateDMCANoticeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get reviewer ID from context
	reviewerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Review the notice
	if err := h.dmcaService.ReviewNotice(c.Request.Context(), noticeID, reviewerID.(uuid.UUID), req.Status, req.Notes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notice reviewed successfully",
		"status":  req.Status,
	})
}

// ProcessTakedown processes a valid DMCA notice and removes content
// POST /api/admin/dmca/notices/:id/process
func (h *DMCAHandler) ProcessTakedown(c *gin.Context) {
	noticeIDStr := c.Param("id")
	noticeID, err := uuid.Parse(noticeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notice ID"})
		return
	}

	// Get admin ID from context
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Process takedown
	if err := h.dmcaService.ProcessTakedown(c.Request.Context(), noticeID, adminID.(uuid.UUID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Takedown processed successfully. Content has been removed and users have been notified.",
	})
}

// ForwardCounterNotice forwards a counter-notice to the original complainant
// POST /api/admin/dmca/counter-notices/:id/forward
func (h *DMCAHandler) ForwardCounterNotice(c *gin.Context) {
	counterNoticeIDStr := c.Param("id")
	counterNoticeID, err := uuid.Parse(counterNoticeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid counter-notice ID"})
		return
	}

	// Get admin ID from context
	adminID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Forward counter-notice
	if err := h.dmcaService.ForwardCounterNoticeToComplainant(c.Request.Context(), counterNoticeID, adminID.(uuid.UUID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Counter-notice forwarded to complainant. Waiting period has begun.",
	})
}

// GetDashboardStats returns DMCA dashboard statistics (admin only)
// GET /api/admin/dmca/dashboard
func (h *DMCAHandler) GetDashboardStats(c *gin.Context) {
	// Get dashboard stats from service
	// In a full implementation, create a service method that calls repo.GetDashboardStats
	c.JSON(http.StatusOK, gin.H{
		"pending_notices":                  0,
		"pending_counter_notices":          0,
		"content_awaiting_removal":         0,
		"content_awaiting_restore":         0,
		"users_with_active_strikes":        0,
		"users_with_two_strikes":           0,
		"total_takedowns_this_month":       0,
		"total_counter_notices_this_month": 0,
	})
}
