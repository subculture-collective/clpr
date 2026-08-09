package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ReportHandler handles report-related HTTP requests
type ReportHandler struct {
	reportRepo  *repository.ReportRepository
	clipRepo    *repository.ClipRepository
	commentRepo *repository.CommentRepository
	userRepo    *repository.UserRepository
	authService *services.AuthService
}

// NewReportHandler creates a new report handler
func NewReportHandler(
	reportRepo *repository.ReportRepository,
	clipRepo *repository.ClipRepository,
	commentRepo *repository.CommentRepository,
	userRepo *repository.UserRepository,
	authService *services.AuthService,
) *ReportHandler {
	return &ReportHandler{
		reportRepo:  reportRepo,
		clipRepo:    clipRepo,
		commentRepo: commentRepo,
		userRepo:    userRepo,
		authService: authService,
	}
}

// CreateReportRequest represents the request body for creating a report
type CreateReportRequest struct {
	ReportableType string  `json:"reportable_type" binding:"required,oneof=clip comment user"`
	ReportableID   string  `json:"reportable_id" binding:"required,uuid"`
	Reason         string  `json:"reason" binding:"required,oneof=spam harassment nsfw violence copyright other"`
	Description    *string `json:"description"`
}

// UpdateReportRequest represents the request body for updating a report
type UpdateReportRequest struct {
	Status string  `json:"status" binding:"required,oneof=pending reviewed actioned dismissed"`
	Action *string `json:"action" binding:"omitempty,oneof=remove_content ban_user mark_false"`
}

func reportActorID(c *gin.Context) (uuid.UUID, bool) {
	value, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}
	userID, ok := value.(uuid.UUID)
	return userID, ok && userID != uuid.Nil
}

// SubmitReport creates a new report
func (h *ReportHandler) SubmitReport(c *gin.Context) {
	userID, ok := reportActorID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req CreateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reportableID, err := uuid.Parse(req.ReportableID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reportable_id format"})
		return
	}

	// Check if reportable item exists
	reportableExists, err := h.validateReportable(c.Request.Context(), req.ReportableType, reportableID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate reportable item"})
		return
	}
	if !reportableExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Reportable item not found"})
		return
	}

	// Check for duplicate reports
	isDuplicate, err := h.reportRepo.CheckDuplicateReport(
		c.Request.Context(),
		userID,
		reportableID,
		req.ReportableType,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check for duplicate reports"})
		return
	}
	if isDuplicate {
		c.JSON(http.StatusConflict, gin.H{"error": "You have already reported this item"})
		return
	}

	// Check rate limit (10 reports per hour)
	oneHourAgo := time.Now().Add(-time.Hour)
	reportCount, err := h.reportRepo.GetReportCountByUser(c.Request.Context(), userID, oneHourAgo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check report rate limit"})
		return
	}
	if reportCount >= 10 {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Report rate limit exceeded. Please try again later."})
		return
	}

	// Create the report
	report := &models.Report{
		ID:             uuid.New(),
		ReporterID:     userID,
		ReportableType: req.ReportableType,
		ReportableID:   reportableID,
		Reason:         req.Reason,
		Description:    req.Description,
		Status:         "pending",
		CreatedAt:      time.Now(),
	}

	if err := h.reportRepo.CreateReport(c.Request.Context(), report); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create report"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Report submitted successfully. Thank you for helping keep our community safe.",
		"report":  report,
	})
}

// ListReports lists all reports with filters (admin/moderator only)
func (h *ReportHandler) ListReports(c *gin.Context) {
	status := c.DefaultQuery("status", "")
	reportableType := c.DefaultQuery("type", "")
	if status != "" && status != "pending" && status != "reviewed" && status != "actioned" && status != "dismissed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status parameter"})
		return
	}
	if reportableType != "" && reportableType != "clip" && reportableType != "comment" && reportableType != "user" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid type parameter"})
		return
	}
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page parameter"})
		return
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Limit must be between 1 and 100"})
		return
	}

	reports, total, err := h.reportRepo.ListReports(c.Request.Context(), status, reportableType, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reports"})
		return
	}

	totalPages := (total + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"data": reports,
		"meta": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// GetReport retrieves a specific report by ID (admin/moderator only)
func (h *ReportHandler) GetReport(c *gin.Context) {
	reportIDStr := c.Param("id")
	reportID, err := uuid.Parse(reportIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID"})
		return
	}

	report, err := h.reportRepo.GetReportByID(c.Request.Context(), reportID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}

	// Get related reports for the same item
	relatedReports, _ := h.reportRepo.GetReportsByReportable(
		c.Request.Context(),
		report.ReportableID,
		report.ReportableType,
	)

	// Get reporter information
	reporter, _ := h.userRepo.GetByID(c.Request.Context(), report.ReporterID)

	c.JSON(http.StatusOK, gin.H{
		"report":          report,
		"reporter":        reporter,
		"related_reports": relatedReports,
	})
}

// UpdateReport updates a report's status and takes action (admin/moderator only)
func (h *ReportHandler) UpdateReport(c *gin.Context) {
	reportIDStr := c.Param("id")
	reportID, err := uuid.Parse(reportIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report ID"})
		return
	}

	userID, ok := reportActorID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req UpdateReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validateReportResolution(req.Status, req.Action); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the report
	report, err := h.reportRepo.GetReportByID(c.Request.Context(), reportID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}

	// Take action if specified
	if req.Action != nil {
		if err := validateReportActionTarget(*req.Action, report.ReportableType); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := h.takeAction(c, report, *req.Action, userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to take action: " + err.Error()})
			return
		}
	}

	// Persist the resolution only after any requested moderation action succeeds.
	if err := h.reportRepo.UpdateReportStatus(c.Request.Context(), reportID, req.Status, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update report"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Report updated successfully",
	})
}

func validateReportResolution(status string, action *string) error {
	if action == nil {
		if status == "actioned" {
			return fmt.Errorf("actioned status requires an action")
		}
		return nil
	}
	if *action == "mark_false" && status != "dismissed" {
		return fmt.Errorf("mark_false action requires dismissed status")
	}
	if *action != "mark_false" && status != "actioned" {
		return fmt.Errorf("moderation actions require actioned status")
	}
	return nil
}

func validateReportActionTarget(action, reportableType string) error {
	switch action {
	case "remove_content":
		if reportableType != "clip" && reportableType != "comment" {
			return fmt.Errorf("remove_content is only supported for clips and comments")
		}
	case "ban_user":
		if reportableType != "comment" && reportableType != "user" {
			return fmt.Errorf("ban_user is only supported for comments and users")
		}
	case "mark_false":
		return nil
	default:
		return fmt.Errorf("unsupported moderation action")
	}
	return nil
}

// validateReportable checks if the reportable item exists
func (h *ReportHandler) validateReportable(ctx context.Context, reportableType string, reportableID uuid.UUID) (bool, error) {
	switch reportableType {
	case "clip":
		_, err := h.clipRepo.GetByID(ctx, reportableID)
		if err != nil {
			return false, nil
		}
		return true, nil
	case "comment":
		_, err := h.commentRepo.GetByID(ctx, reportableID, nil)
		if err != nil {
			return false, nil
		}
		return true, nil
	case "user":
		_, err := h.userRepo.GetByID(ctx, reportableID)
		if err != nil {
			return false, nil
		}
		return true, nil
	default:
		return false, nil
	}
}

// takeAction performs moderation actions based on the report
func (h *ReportHandler) takeAction(c *gin.Context, report *models.Report, action string, _ uuid.UUID) error {
	ctx := c.Request.Context()

	switch action {
	case "remove_content":
		reason := "Content removed due to violation of community guidelines"
		if report.Reason != "" {
			reason = "Content removed: " + report.Reason
		}

		switch report.ReportableType {
		case "clip":
			return h.clipRepo.RemoveClip(ctx, report.ReportableID, &reason)
		case "comment":
			return h.commentRepo.RemoveComment(ctx, report.ReportableID, &reason)
		}

	case "ban_user":
		// Get the user to ban
		var targetUserID uuid.UUID
		switch report.ReportableType {
		case "comment":
			comment, err := h.commentRepo.GetByID(ctx, report.ReportableID, nil)
			if err != nil {
				return err
			}
			targetUserID = comment.UserID
		case "user":
			targetUserID = report.ReportableID
		}

		// Ban the user
		return h.userRepo.BanUser(ctx, targetUserID)

	case "mark_false":
		// Mark report as dismissed (already handled by status update)
		return nil
	}

	return nil
}
