package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

const (
	secondsPerHour         = 3600
	banSyncTimeoutDuration = 5 * time.Minute
)

// TwitchModerationService defines the interface for Twitch-specific moderation operations
type TwitchModerationService interface {
	BanUserOnTwitch(ctx context.Context, moderatorUserID uuid.UUID, broadcasterID string, targetUserID string, reason *string, duration *int) error
	UnbanUserOnTwitch(ctx context.Context, moderatorUserID uuid.UUID, broadcasterID string, targetUserID string) error
}

// ModerationHandler handles moderation operations
type ModerationHandler struct {
	moderationEventService  *services.ModerationEventService
	moderationService       *services.ModerationService
	abuseDetector           *services.SubmissionAbuseDetector
	toxicityClassifier      *services.ToxicityClassifier
	twitchBanSyncService    *services.TwitchBanSyncService
	twitchModerationService TwitchModerationService
	communityRepo           *repository.CommunityRepository
	auditLogRepo            *repository.AuditLogRepository
	db                      *pgxpool.Pool
}

// NewModerationHandler creates a new ModerationHandler
func NewModerationHandler(moderationEventService *services.ModerationEventService, moderationService *services.ModerationService, abuseDetector *services.SubmissionAbuseDetector, toxicityClassifier *services.ToxicityClassifier, twitchBanSyncService *services.TwitchBanSyncService, communityRepo *repository.CommunityRepository, auditLogRepo *repository.AuditLogRepository, db *pgxpool.Pool) *ModerationHandler {
	return &ModerationHandler{
		moderationEventService:  moderationEventService,
		moderationService:       moderationService,
		abuseDetector:           abuseDetector,
		toxicityClassifier:      toxicityClassifier,
		twitchBanSyncService:    twitchBanSyncService,
		twitchModerationService: nil, // Will be set separately if configured
		communityRepo:           communityRepo,
		auditLogRepo:            auditLogRepo,
		db:                      db,
	}
}

// SetTwitchModerationService sets the Twitch moderation service (optional dependency)
func (h *ModerationHandler) SetTwitchModerationService(service TwitchModerationService) {
	h.twitchModerationService = service
}

// GetPendingEvents retrieves pending moderation events
// GET /admin/moderation/events
func (h *ModerationHandler) GetPendingEvents(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 || limit > 100 {
		limit = 50
	}

	events, err := h.moderationEventService.GetPendingEvents(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve pending events",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    events,
		"meta": gin.H{
			"count": len(events),
			"limit": limit,
		},
	})
}

// GetEventsByType retrieves events filtered by type
// GET /admin/moderation/events/:type
func (h *ModerationHandler) GetEventsByType(c *gin.Context) {
	eventType := services.ModerationEventType(c.Param("type"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 || limit > 100 {
		limit = 50
	}

	events, err := h.moderationEventService.GetEventsByType(c.Request.Context(), eventType, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve events",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    events,
		"meta": gin.H{
			"type":  eventType,
			"count": len(events),
			"limit": limit,
		},
	})
}

// MarkEventReviewed marks an event as reviewed
// POST /admin/moderation/events/:id/review
func (h *ModerationHandler) MarkEventReviewed(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid event ID",
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

	err = h.moderationEventService.MarkEventReviewed(c.Request.Context(), eventID, reviewerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to mark event as reviewed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Event marked as reviewed",
	})
}

// ProcessEvent processes an event with an action
// POST /admin/moderation/events/:id/process
func (h *ModerationHandler) ProcessEvent(c *gin.Context) {
	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid event ID",
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

	var req struct {
		Action string `json:"action" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Action is required",
		})
		return
	}

	err = h.moderationEventService.ProcessEvent(c.Request.Context(), eventID, reviewerID, req.Action)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to process event",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Event processed",
		"action":  req.Action,
	})
}

// GetEventStats returns statistics about moderation events
// GET /admin/moderation/stats
func (h *ModerationHandler) GetEventStats(c *gin.Context) {
	stats, err := h.moderationEventService.GetEventStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve stats",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetUserAbuseStats returns abuse statistics for a specific user
// GET /admin/moderation/abuse/:userId
func (h *ModerationHandler) GetUserAbuseStats(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID",
		})
		return
	}

	stats, err := h.abuseDetector.GetAbuseStats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve abuse stats",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
		"user_id": userID,
	})
}

// GetModerationQueue retrieves moderation queue items with optional filters
// GET /admin/moderation/queue
func (h *ModerationHandler) GetModerationQueue(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse and validate query parameters
	contentType := c.Query("type")
	status := c.DefaultQuery("status", "pending")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 || limit > 100 {
		limit = 50
	}

	// Validate status parameter
	validStatuses := map[string]bool{
		"pending":   true,
		"approved":  true,
		"rejected":  true,
		"escalated": true,
	}
	if !validStatuses[status] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid status. Must be one of: pending, approved, rejected, escalated",
		})
		return
	}

	// Validate content type if provided
	if contentType != "" {
		validContentTypes := map[string]bool{
			"comment":    true,
			"clip":       true,
			"user":       true,
			"submission": true,
		}
		if !validContentTypes[contentType] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid content type. Must be one of: comment, clip, user, submission",
			})
			return
		}
	}

	// Build query with filters
	query := `
		SELECT mq.id, mq.content_type, mq.content_id, mq.reason, mq.priority,
		       mq.status, mq.assigned_to, mq.reported_by, mq.report_count,
		       mq.auto_flagged, mq.confidence_score, mq.created_at,
		       mq.reviewed_at, mq.reviewed_by
		FROM moderation_queue mq
		WHERE mq.status = $1
	`
	args := []interface{}{status}
	argIdx := 2

	if contentType != "" {
		query += fmt.Sprintf(" AND mq.content_type = $%d", argIdx)
		args = append(args, contentType)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY mq.priority DESC, mq.created_at ASC LIMIT $%d", argIdx)
	args = append(args, limit)

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve moderation queue",
		})
		return
	}
	defer rows.Close()

	var items []models.ModerationQueueItem
	for rows.Next() {
		var item models.ModerationQueueItem
		err := rows.Scan(
			&item.ID, &item.ContentType, &item.ContentID, &item.Reason,
			&item.Priority, &item.Status, &item.AssignedTo, &item.ReportedBy,
			&item.ReportCount, &item.AutoFlagged, &item.ConfidenceScore,
			&item.CreatedAt, &item.ReviewedAt, &item.ReviewedBy,
		)
		if err != nil {
			// Log scan error for debugging but continue processing other rows
			c.Error(fmt.Errorf("failed to scan moderation queue item: %w", err))
			continue
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    items,
		"meta": gin.H{
			"count":  len(items),
			"limit":  limit,
			"status": status,
		},
	})
}

// ApproveContent approves a moderation queue item
// POST /admin/moderation/:id/approve
func (h *ModerationHandler) ApproveContent(c *gin.Context) {
	ctx := c.Request.Context()
	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid item ID",
		})
		return
	}

	// Get moderator ID from context
	moderatorIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	moderatorID := moderatorIDVal.(uuid.UUID)

	// Begin transaction
	tx, err := h.db.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to begin transaction",
		})
		return
	}
	defer tx.Rollback(ctx)

	// Update queue item
	cmdTag, err := tx.Exec(ctx, `
		UPDATE moderation_queue
		SET status = 'approved', reviewed_by = $1
		WHERE id = $2 AND status = 'pending'
	`, moderatorID, itemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to approve item",
		})
		return
	}

	// Check if any rows were updated
	if cmdTag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Item not found or not in pending status",
		})
		return
	}

	// Record decision
	_, err = tx.Exec(ctx, `
		INSERT INTO moderation_decisions (queue_item_id, moderator_id, action)
		VALUES ($1, $2, 'approve')
	`, itemID, moderatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to record decision",
		})
		return
	}

	if err = tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to commit transaction",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Content approved",
	})
}

// RejectContent rejects a moderation queue item
// POST /admin/moderation/:id/reject
func (h *ModerationHandler) RejectContent(c *gin.Context) {
	ctx := c.Request.Context()
	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid item ID",
		})
		return
	}

	// Get moderator ID from context
	moderatorIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	moderatorID := moderatorIDVal.(uuid.UUID)

	// Parse request body for optional reason
	var req struct {
		Reason *string `json:"reason"`
	}
	// Enforce JSON Content-Type for consistency
	contentType := c.GetHeader("Content-Type")
	if c.Request.ContentLength > 0 {
		if contentType != "application/json" {
			c.JSON(http.StatusUnsupportedMediaType, gin.H{
				"error": "Content-Type must be application/json",
			})
			return
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid JSON in request body",
			})
			return
		}
	}

	// Begin transaction
	tx, err := h.db.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to begin transaction",
		})
		return
	}
	defer tx.Rollback(ctx)

	// Update queue item
	cmdTag, err := tx.Exec(ctx, `
		UPDATE moderation_queue
		SET status = 'rejected', reviewed_by = $1
		WHERE id = $2 AND status = 'pending'
	`, moderatorID, itemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to reject item",
		})
		return
	}

	// Check if any rows were updated
	if cmdTag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Item not found or not in pending status",
		})
		return
	}

	// Record decision
	_, err = tx.Exec(ctx, `
		INSERT INTO moderation_decisions (queue_item_id, moderator_id, action, reason)
		VALUES ($1, $2, 'reject', $3)
	`, itemID, moderatorID, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to record decision",
		})
		return
	}

	if err = tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to commit transaction",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Content rejected",
	})
}

// BulkModerate performs bulk moderation actions
// POST /admin/moderation/bulk
func (h *ModerationHandler) BulkModerate(c *gin.Context) {
	ctx := c.Request.Context()

	var req models.BulkModerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	// Get moderator ID from context
	moderatorIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	moderatorID := moderatorIDVal.(uuid.UUID)

	// Convert item IDs to UUIDs
	itemIDs := make([]uuid.UUID, 0, len(req.ItemIDs))
	for _, idStr := range req.ItemIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid item ID: " + idStr,
			})
			return
		}
		itemIDs = append(itemIDs, id)
	}

	// Begin transaction
	tx, err := h.db.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to begin transaction",
		})
		return
	}
	defer tx.Rollback(ctx)

	// Determine status based on action
	status := "approved"
	if req.Action == "reject" {
		status = "rejected"
	} else if req.Action == "escalate" {
		status = "escalated"
	}

	// Update all items
	processedCount := 0
	failedItems := make([]string, 0)
	for _, itemID := range itemIDs {
		// Update queue item
		cmdTag, err := tx.Exec(ctx, `
			UPDATE moderation_queue
			SET status = $1, reviewed_by = $2
			WHERE id = $3 AND status = 'pending'
		`, status, moderatorID, itemID)
		if err != nil {
			failedItems = append(failedItems, itemID.String())
			continue
		}
		if cmdTag.RowsAffected() == 0 {
			failedItems = append(failedItems, itemID.String())
			continue
		}

		// Record decision
		_, err = tx.Exec(ctx, `
			INSERT INTO moderation_decisions (queue_item_id, moderator_id, action, reason)
			VALUES ($1, $2, $3, $4)
		`, itemID, moderatorID, req.Action, req.Reason)
		if err != nil {
			failedItems = append(failedItems, itemID.String())
			continue
		}

		processedCount++
	}

	if err = tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to commit transaction",
		})
		return
	}

	response := gin.H{
		"success":   true,
		"processed": processedCount,
		"total":     len(itemIDs),
	}
	if len(failedItems) > 0 {
		response["failed"] = failedItems
		response["message"] = fmt.Sprintf("Processed %d items, %d failed", processedCount, len(failedItems))
	}

	c.JSON(http.StatusOK, response)
}

// GetModerationStats returns statistics about the moderation queue
// GET /admin/moderation/queue/stats
func (h *ModerationHandler) GetModerationStats(c *gin.Context) {
	ctx := c.Request.Context()

	stats := models.ModerationQueueStats{
		ByContentType: make(map[string]int),
		ByReason:      make(map[string]int),
	}

	// Get all stats in a single optimized query using CTEs
	rows, err := h.db.Query(ctx, `
		WITH status_counts AS (
			SELECT
				COUNT(*) FILTER (WHERE status = 'pending') as total_pending,
				COUNT(*) FILTER (WHERE status = 'approved') as total_approved,
				COUNT(*) FILTER (WHERE status = 'rejected') as total_rejected,
				COUNT(*) FILTER (WHERE status = 'escalated') as total_escalated,
				COUNT(*) FILTER (WHERE status = 'pending' AND auto_flagged = true) as auto_flagged_count,
				COUNT(*) FILTER (WHERE status = 'pending' AND report_count > 0) as user_reported_count,
				COUNT(*) FILTER (WHERE status = 'pending' AND priority >= 75) as high_priority_count,
				EXTRACT(EPOCH FROM (NOW() - MIN(created_at) FILTER (WHERE status = 'pending')))/`+fmt.Sprintf("%d", secondsPerHour)+` as oldest_age
			FROM moderation_queue
		),
		type_counts AS (
			SELECT 'type' as category, content_type as name, COUNT(*) as count
			FROM moderation_queue
			WHERE status = 'pending'
			GROUP BY content_type
		),
		reason_counts AS (
			SELECT 'reason' as category, reason as name, COUNT(*) as count
			FROM moderation_queue
			WHERE status = 'pending'
			GROUP BY reason
		)
		SELECT 'status' as type, NULL as name,
			   total_pending, total_approved, total_rejected, total_escalated,
			   auto_flagged_count, user_reported_count, high_priority_count, oldest_age
		FROM status_counts
		UNION ALL
		SELECT category, name, count, 0, 0, 0, 0, 0, 0, 0
		FROM type_counts
		UNION ALL
		SELECT category, name, count, 0, 0, 0, 0, 0, 0, 0
		FROM reason_counts
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve stats",
		})
		return
	}
	defer rows.Close()

	// Process results
	for rows.Next() {
		var rowType string
		var name *string
		var count, totalPending, totalApproved, totalRejected, totalEscalated int
		var autoFlagged, userReported, highPriority int
		var oldestAge *int

		err := rows.Scan(&rowType, &name, &count, &totalPending, &totalApproved,
			&totalRejected, &totalEscalated, &autoFlagged, &userReported,
			&highPriority, &oldestAge)
		if err != nil {
			continue
		}

		if rowType == "status" {
			// Status row contains aggregate stats
			stats.TotalPending = totalPending
			stats.TotalApproved = totalApproved
			stats.TotalRejected = totalRejected
			stats.TotalEscalated = totalEscalated
			stats.AutoFlaggedCount = autoFlagged
			stats.UserReportedCount = userReported
			stats.HighPriorityCount = highPriority
			stats.OldestPendingAge = oldestAge
		} else if rowType == "type" && name != nil {
			stats.ByContentType[*name] = count
		} else if rowType == "reason" && name != nil {
			stats.ByReason[*name] = count
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// CreateAppeal creates a new appeal for a moderation decision
// POST /api/moderation/appeals
func (h *ModerationHandler) CreateAppeal(c *gin.Context) {
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

	var req models.CreateAppealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	moderationActionID, err := uuid.Parse(req.ModerationActionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid moderation action ID",
		})
		return
	}

	// Verify the moderation action exists and get content details for ownership check
	var contentType string
	var contentID uuid.UUID
	err = h.db.QueryRow(ctx, `
		SELECT mq.content_type, mq.content_id
		FROM moderation_decisions md
		JOIN moderation_queue mq ON md.queue_item_id = mq.id
		WHERE md.id = $1
	`, moderationActionID).Scan(&contentType, &contentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Moderation action not found",
		})
		return
	}

	// Verify content ownership based on content type
	var ownsContent bool
	switch contentType {
	case "comment":
		err = h.db.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM comments WHERE id = $1 AND user_id = $2)
		`, contentID, userID).Scan(&ownsContent)
	case "clip":
		err = h.db.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM clips WHERE id = $1 AND submitted_by_user_id = $2)
		`, contentID, userID).Scan(&ownsContent)
	case "user":
		// For user moderation actions, check if the moderated user is the same as the requester
		err = h.db.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND id = $2)
		`, contentID, userID).Scan(&ownsContent)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Unsupported content type for appeals: %s", contentType),
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to verify content ownership",
		})
		return
	}

	if !ownsContent {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "You do not have permission to appeal this moderation action",
		})
		return
	}

	// Insert appeal
	var appealID uuid.UUID
	err = h.db.QueryRow(ctx, `
		INSERT INTO moderation_appeals (user_id, moderation_action_id, reason)
		VALUES ($1, $2, $3)
		RETURNING id
	`, userID, moderationActionID, req.Reason).Scan(&appealID)
	if err != nil {
		// Check for unique constraint violation (duplicate appeal)
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "uq_appeals_action_pending") {
			c.JSON(http.StatusConflict, gin.H{
				"error": "An appeal for this moderation action is already pending",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create appeal",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":   true,
		"appeal_id": appealID,
		"message":   "Appeal submitted successfully",
	})
}

// GetAppeals retrieves appeals for admin review
// GET /admin/moderation/appeals
func (h *ModerationHandler) GetAppeals(c *gin.Context) {
	ctx := c.Request.Context()

	status := c.DefaultQuery("status", "pending")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 || limit > 100 {
		limit = 50
	}

	// Validate status parameter
	validStatuses := map[string]bool{
		"pending":  true,
		"approved": true,
		"rejected": true,
	}
	if !validStatuses[status] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid status. Must be one of: pending, approved, rejected",
		})
		return
	}

	query := `
		SELECT ma.id, ma.user_id, ma.moderation_action_id, ma.reason,
		       ma.status, ma.resolved_by, ma.resolution,
		       ma.created_at, ma.resolved_at,
		       u.username, u.display_name,
		       md.action, md.reason as decision_reason,
		       mq.content_type, mq.content_id
		FROM moderation_appeals ma
		JOIN users u ON ma.user_id = u.id
		JOIN moderation_decisions md ON ma.moderation_action_id = md.id
		JOIN moderation_queue mq ON md.queue_item_id = mq.id
		WHERE ma.status = $1
		ORDER BY ma.created_at ASC
		LIMIT $2
	`

	rows, err := h.db.Query(ctx, query, status, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve appeals",
		})
		return
	}
	defer rows.Close()

	type AppealWithDetails struct {
		models.ModerationAppeal
		Username       string    `json:"username"`
		DisplayName    string    `json:"display_name"`
		DecisionAction string    `json:"decision_action"`
		DecisionReason *string   `json:"decision_reason,omitempty"`
		ContentType    string    `json:"content_type"`
		ContentID      uuid.UUID `json:"content_id"`
	}

	var appeals []AppealWithDetails
	for rows.Next() {
		var appeal AppealWithDetails
		err := rows.Scan(
			&appeal.ID, &appeal.UserID, &appeal.ModerationActionID,
			&appeal.Reason, &appeal.Status, &appeal.ResolvedBy,
			&appeal.Resolution, &appeal.CreatedAt, &appeal.ResolvedAt,
			&appeal.Username, &appeal.DisplayName,
			&appeal.DecisionAction, &appeal.DecisionReason,
			&appeal.ContentType, &appeal.ContentID,
		)
		if err != nil {
			c.Error(fmt.Errorf("failed to scan appeal: %w", err))
			continue
		}
		appeals = append(appeals, appeal)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    appeals,
		"meta": gin.H{
			"count":  len(appeals),
			"limit":  limit,
			"status": status,
		},
	})
}

// ResolveAppeal resolves an appeal
// POST /admin/moderation/appeals/:id/resolve
func (h *ModerationHandler) ResolveAppeal(c *gin.Context) {
	ctx := c.Request.Context()

	appealID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid appeal ID",
		})
		return
	}

	// Get admin ID from context
	adminIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	adminID := adminIDVal.(uuid.UUID)

	var req models.ResolveAppealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	// Map decision to status
	status := "rejected"
	if req.Decision == "approve" {
		status = "approved"
	}

	// Update appeal
	cmdTag, err := h.db.Exec(ctx, `
		UPDATE moderation_appeals
		SET status = $1, resolved_by = $2, resolution = $3
		WHERE id = $4 AND status = 'pending'
	`, status, adminID, req.Resolution, appealID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to resolve appeal",
		})
		return
	}

	if cmdTag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Appeal not found or not in pending status",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Appeal resolved successfully",
		"status":  status,
	})
}

// GetUserAppeals retrieves appeals for the authenticated user
// GET /api/moderation/appeals
func (h *ModerationHandler) GetUserAppeals(c *gin.Context) {
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

	query := `
		SELECT ma.id, ma.user_id, ma.moderation_action_id, ma.reason,
		       ma.status, ma.resolved_by, ma.resolution,
		       ma.created_at, ma.resolved_at,
		       md.action, md.reason as decision_reason,
		       mq.content_type, mq.content_id
		FROM moderation_appeals ma
		JOIN moderation_decisions md ON ma.moderation_action_id = md.id
		JOIN moderation_queue mq ON md.queue_item_id = mq.id
		WHERE ma.user_id = $1
		ORDER BY ma.created_at DESC
		LIMIT 50
	`

	rows, err := h.db.Query(ctx, query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve appeals",
		})
		return
	}
	defer rows.Close()

	type UserAppealWithDetails struct {
		models.ModerationAppeal
		DecisionAction string    `json:"decision_action"`
		DecisionReason *string   `json:"decision_reason,omitempty"`
		ContentType    string    `json:"content_type"`
		ContentID      uuid.UUID `json:"content_id"`
	}

	var appeals []UserAppealWithDetails
	for rows.Next() {
		var appeal UserAppealWithDetails
		err := rows.Scan(
			&appeal.ID, &appeal.UserID, &appeal.ModerationActionID,
			&appeal.Reason, &appeal.Status, &appeal.ResolvedBy,
			&appeal.Resolution, &appeal.CreatedAt, &appeal.ResolvedAt,
			&appeal.DecisionAction, &appeal.DecisionReason,
			&appeal.ContentType, &appeal.ContentID,
		)
		if err != nil {
			c.Error(fmt.Errorf("failed to scan appeal: %w", err))
			continue
		}
		appeals = append(appeals, appeal)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    appeals,
	})
}

// GetModerationAuditLogs retrieves audit logs with optional filters
// GET /admin/moderation/audit
func (h *ModerationHandler) GetModerationAuditLogs(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse query parameters
	moderatorID := c.Query("moderator_id")
	actionType := c.Query("action")
	startDate := c.DefaultQuery("start_date", "")
	endDate := c.DefaultQuery("end_date", "")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit < 1 || limit > 1000 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	// Build query with filters - use CTE for counting
	baseWhere := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if moderatorID != "" {
		moderatorUUID, err := uuid.Parse(moderatorID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid moderator ID",
			})
			return
		}
		baseWhere += fmt.Sprintf(" AND md.moderator_id = $%d", argIdx)
		args = append(args, moderatorUUID)
		argIdx++
	}

	if actionType != "" {
		// Validate action type against allowed values
		validActions := map[string]bool{
			"approve":  true,
			"reject":   true,
			"escalate": true,
			"ban_user": true,
		}
		if !validActions[actionType] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid action type. Must be one of: approve, reject, escalate, ban_user",
			})
			return
		}
		baseWhere += fmt.Sprintf(" AND md.action = $%d", argIdx)
		args = append(args, actionType)
		argIdx++
	}

	if startDate != "" {
		baseWhere += fmt.Sprintf(" AND md.created_at >= $%d", argIdx)
		args = append(args, startDate)
		argIdx++
	}

	if endDate != "" {
		baseWhere += fmt.Sprintf(" AND md.created_at < ($%d::date + interval '1 day')", argIdx)
		args = append(args, endDate)
		argIdx++
	}

	// Count total records using CTE
	countQuery := fmt.Sprintf(`
		WITH filtered_decisions AS (
			SELECT md.id
			FROM moderation_decisions md
			JOIN moderation_queue mq ON md.queue_item_id = mq.id
			LEFT JOIN users u ON md.moderator_id = u.id
			%s
		)
		SELECT COUNT(*) FROM filtered_decisions
	`, baseWhere)

	// Check if database is available before querying
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Database service unavailable",
		})
		return
	}

	var totalCount int
	err := h.db.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to count audit logs",
		})
		return
	}

	// Check if database is available
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Database service unavailable",
		})
		return
	}

	// Get paginated data
	query := fmt.Sprintf(`
		SELECT
			md.id, md.queue_item_id, md.moderator_id,
			u.username as moderator_name,
			md.action,
			mq.content_type, mq.content_id,
			md.reason, md.metadata, md.created_at
		FROM moderation_decisions md
		JOIN moderation_queue mq ON md.queue_item_id = mq.id
		LEFT JOIN users u ON md.moderator_id = u.id
		%s
		ORDER BY md.created_at DESC LIMIT $%d OFFSET $%d
	`, baseWhere, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve audit logs",
		})
		return
	}
	defer rows.Close()

	logs := []models.ModerationDecisionWithDetails{}
	for rows.Next() {
		var log models.ModerationDecisionWithDetails
		err := rows.Scan(
			&log.ID, &log.QueueItemID, &log.ModeratorID,
			&log.ModeratorName, &log.Action,
			&log.ContentType, &log.ContentID,
			&log.Reason, &log.Metadata, &log.CreatedAt,
		)
		if err != nil {
			c.Error(fmt.Errorf("failed to scan audit log: %w", err))
			continue
		}
		logs = append(logs, log)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    logs,
		"meta": gin.H{
			"total":  totalCount,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// GetModerationAnalytics retrieves analytics data for moderation actions
// GET /admin/moderation/analytics
func (h *ModerationHandler) GetModerationAnalytics(c *gin.Context) {
	ctx := c.Request.Context()

	// Check if database is available
	if h.db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Database service unavailable",
		})
		return
	}

	// Parse query parameters for date range
	startDate := c.DefaultQuery("start_date", "")
	endDate := c.DefaultQuery("end_date", "")

	// Default to last 30 days if not specified
	if startDate == "" {
		startDate = time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	analytics := models.ModerationAnalytics{
		ActionsByType:        make(map[string]int),
		ActionsByModerator:   make(map[string]int),
		ContentTypeBreakdown: make(map[string]int),
		ActionsOverTime:      []models.TimeSeriesPoint{},
		BanReasons:           make(map[string]int),
		MostBannedUsers:      []models.BannedUserStat{},
	}

	// Total actions in date range
	err := h.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM moderation_decisions
		WHERE created_at >= $1 AND created_at < $2::date + interval '1 day'
	`, startDate, endDate).Scan(&analytics.TotalActions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve analytics",
		})
		return
	}

	// Actions by type
	rows, err := h.db.Query(ctx, `
		SELECT action, COUNT(*)
		FROM moderation_decisions
		WHERE created_at >= $1 AND created_at < $2::date + interval '1 day'
		GROUP BY action
		ORDER BY COUNT(*) DESC
	`, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve action breakdown",
		})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var action string
		var count int
		if err := rows.Scan(&action, &count); err != nil {
			continue
		}
		analytics.ActionsByType[action] = count
	}
	rows.Close()

	// Actions by moderator (top 10)
	rows, err = h.db.Query(ctx, `
		SELECT u.username, COUNT(*)
		FROM moderation_decisions md
		JOIN users u ON md.moderator_id = u.id
		WHERE md.created_at >= $1 AND md.created_at < $2::date + interval '1 day'
		GROUP BY u.username
		ORDER BY COUNT(*) DESC
		LIMIT 10
	`, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve moderator stats",
		})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var username string
		var count int
		if err := rows.Scan(&username, &count); err != nil {
			continue
		}
		analytics.ActionsByModerator[username] = count
	}
	rows.Close()

	// Content type breakdown
	rows, err = h.db.Query(ctx, `
		SELECT mq.content_type, COUNT(*)
		FROM moderation_decisions md
		JOIN moderation_queue mq ON md.queue_item_id = mq.id
		WHERE md.created_at >= $1 AND md.created_at < $2::date + interval '1 day'
		GROUP BY mq.content_type
		ORDER BY COUNT(*) DESC
	`, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve content type breakdown",
		})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var contentType string
		var count int
		if err := rows.Scan(&contentType, &count); err != nil {
			continue
		}
		analytics.ContentTypeBreakdown[contentType] = count
	}
	rows.Close()

	// Actions over time (daily aggregation)
	rows, err = h.db.Query(ctx, `
		SELECT
			DATE(created_at) as date,
			COUNT(*) as count
		FROM moderation_decisions
		WHERE created_at >= $1 AND created_at < $2::date + interval '1 day'
		GROUP BY DATE(created_at)
		ORDER BY date ASC
	`, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve time series data",
		})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var point models.TimeSeriesPoint
		if err := rows.Scan(&point.Date, &point.Count); err != nil {
			continue
		}
		analytics.ActionsOverTime = append(analytics.ActionsOverTime, point)
	}
	rows.Close()

	// Calculate average response time (time from queue creation to decision)
	var avgResponseMinutes *float64
	err = h.db.QueryRow(ctx, `
		SELECT AVG(EXTRACT(EPOCH FROM (md.created_at - mq.created_at)) / 60)
		FROM moderation_decisions md
		JOIN moderation_queue mq ON md.queue_item_id = mq.id
		WHERE md.created_at >= $1 AND md.created_at < $2::date + interval '1 day'
	`, startDate, endDate).Scan(&avgResponseMinutes)
	if err == nil && avgResponseMinutes != nil {
		analytics.AverageResponseTime = avgResponseMinutes
	}

	// Ban reasons distribution (from user_bans table)
	analytics.BanReasons = make(map[string]int)
	rows, err = h.db.Query(ctx, `
		SELECT reason, COUNT(*)
		FROM user_bans
		WHERE created_at >= $1 AND created_at < $2::date + interval '1 day'
		GROUP BY reason
		ORDER BY COUNT(*) DESC
	`, startDate, endDate)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var reason string
			var count int
			if err := rows.Scan(&reason, &count); err == nil {
				analytics.BanReasons[reason] = count
			}
		}
	}

	// Most banned users (users with most bans in the date range)
	rows, err = h.db.Query(ctx, `
		SELECT
			ub.user_id,
			COALESCE(u.username, 'Unknown') as username,
			COUNT(*) as ban_count,
			MAX(ub.created_at) as last_ban_at
		FROM user_bans ub
		LEFT JOIN users u ON ub.user_id = u.id
		WHERE ub.created_at >= $1 AND ub.created_at < $2::date + interval '1 day'
		GROUP BY ub.user_id, u.username
		ORDER BY ban_count DESC, last_ban_at DESC
		LIMIT 10
	`, startDate, endDate)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var stat models.BannedUserStat
			var lastBanAt time.Time
			if err := rows.Scan(&stat.UserID, &stat.Username, &stat.BanCount, &lastBanAt); err == nil {
				stat.LastBanAt = lastBanAt.Format(time.RFC3339)
				analytics.MostBannedUsers = append(analytics.MostBannedUsers, stat)
			}
		}
	}

	// Appeals statistics
	appealStats := &models.AppealStats{}
	err = h.db.QueryRow(ctx, `
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'pending') as pending,
			COUNT(*) FILTER (WHERE status = 'approved') as approved,
			COUNT(*) FILTER (WHERE status = 'rejected') as rejected
		FROM moderation_appeals
		WHERE created_at >= $1 AND created_at < $2::date + interval '1 day'
	`, startDate, endDate).Scan(
		&appealStats.TotalAppeals,
		&appealStats.PendingAppeals,
		&appealStats.ApprovedAppeals,
		&appealStats.RejectedAppeals,
	)
	if err == nil && appealStats.TotalAppeals > 0 {
		// Calculate false positive rate (approved appeals / total appeals)
		rate := float64(appealStats.ApprovedAppeals) / float64(appealStats.TotalAppeals) * 100
		appealStats.FalsePositiveRate = &rate
		analytics.Appeals = appealStats
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    analytics,
	})
}

// GetToxicityMetrics retrieves toxicity classification metrics
// GET /admin/moderation/toxicity/metrics
func (h *ModerationHandler) GetToxicityMetrics(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse query parameters for date range
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// Default to last 30 days if not specified
	var startDate, endDate time.Time
	if startDateStr == "" {
		startDate = time.Now().AddDate(0, 0, -30)
	} else {
		parsed, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid start_date format. Use YYYY-MM-DD",
			})
			return
		}
		startDate = parsed
	}

	if endDateStr == "" {
		endDate = time.Now()
	} else {
		parsed, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid end_date format. Use YYYY-MM-DD",
			})
			return
		}
		endDate = parsed.Add(24 * time.Hour) // Include the end date
	}

	// Get metrics from toxicity classifier
	if h.toxicityClassifier == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Toxicity classification service not available",
		})
		return
	}

	metrics, err := h.toxicityClassifier.GetMetrics(ctx, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve toxicity metrics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"data":       metrics,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	})
}

// SyncBans initiates Twitch ban synchronization for a channel
// POST /api/v1/moderation/sync-bans
func (h *ModerationHandler) SyncBans(c *gin.Context) {
	logger := utils.GetLogger()

	// Get user ID from context (authentication required)
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	userID := userIDVal.(uuid.UUID)

	// Parse request body
	var req struct {
		ChannelID string `json:"channel_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: channel_id is required",
		})
		return
	}

	// Validate channel_id is not empty
	if strings.TrimSpace(req.ChannelID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "channel_id cannot be empty",
		})
		return
	}

	// Check if Twitch ban sync service is available
	if h.twitchBanSyncService == nil {
		logger.Warn("Twitch ban sync service not available", map[string]interface{}{
			"user_id": userID.String(),
		})
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Twitch ban sync service not available. Please ensure Twitch integration is configured.",
		})
		return
	}

	// Create job ID for tracking
	jobID := uuid.New()

	// Log sync request
	logger.Info("Ban sync requested", map[string]interface{}{
		"user_id":    userID.String(),
		"channel_id": req.ChannelID,
		"job_id":     jobID.String(),
	})

	// Start async ban sync in a goroutine (fire and forget)
	go func() {
		// Create a new context for the background job with timeout
		bgCtx, cancel := context.WithTimeout(context.Background(), banSyncTimeoutDuration)
		defer cancel()

		err := h.twitchBanSyncService.SyncChannelBans(bgCtx, userID.String(), req.ChannelID)
		if err != nil {
			// Log error but don't fail the request
			var authErr *services.AuthenticationError
			var authzErr *services.AuthorizationError
			var apiErr *services.BanSyncTwitchAPIError
			var dbErr *services.DatabaseError

			if errors.As(err, &authErr) {
				logger.Error("Ban sync authentication failed", err, map[string]interface{}{
					"user_id":    userID.String(),
					"channel_id": req.ChannelID,
					"job_id":     jobID.String(),
				})
			} else if errors.As(err, &authzErr) {
				logger.Error("Ban sync authorization failed", err, map[string]interface{}{
					"user_id":    userID.String(),
					"channel_id": req.ChannelID,
					"job_id":     jobID.String(),
				})
			} else if errors.As(err, &apiErr) {
				logger.Error("Ban sync Twitch API error", err, map[string]interface{}{
					"user_id":    userID.String(),
					"channel_id": req.ChannelID,
					"job_id":     jobID.String(),
				})
			} else if errors.As(err, &dbErr) {
				logger.Error("Ban sync database error", err, map[string]interface{}{
					"user_id":    userID.String(),
					"channel_id": req.ChannelID,
					"job_id":     jobID.String(),
				})
			} else {
				logger.Error("Ban sync failed", err, map[string]interface{}{
					"user_id":    userID.String(),
					"channel_id": req.ChannelID,
					"job_id":     jobID.String(),
				})
			}
		} else {
			logger.Info("Ban sync completed successfully", map[string]interface{}{
				"user_id":    userID.String(),
				"channel_id": req.ChannelID,
				"job_id":     jobID.String(),
			})
		}
	}()

	// Return immediate response
	c.JSON(http.StatusOK, gin.H{
		"status":  "syncing",
		"job_id":  jobID.String(),
		"message": "Ban sync started",
	})
}

// GetBans retrieves bans for a channel with filtering and pagination
// GET /api/v1/moderation/bans
func (h *ModerationHandler) GetBans(c *gin.Context) {
	ctx := c.Request.Context()

	// Get moderator ID from context
	moderatorIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	moderatorID := moderatorIDVal.(uuid.UUID)

	// Check if moderation service is available before parsing pagination
	if h.moderationService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Moderation service not available",
		})
		return
	}

	// Parse pagination parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Validate pagination parameters
	if limit < 1 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	// Ensure offset is a multiple of limit for proper page calculation
	if offset%limit != 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "offset must be a multiple of limit",
		})
		return
	}

	// Convert offset to page number (GetBans expects page, not offset)
	page := (offset / limit) + 1

	// Parse optional channelId - if absent, list all bans (admin only)
	channelIDStr := c.Query("channelId")
	if channelIDStr == "" {
		// List all bans across all communities (admin/site mod only)
		bans, total, err := h.moderationService.GetAllBans(ctx, moderatorID, page, limit)
		if err != nil {
			if errors.Is(err, services.ErrModerationPermissionDenied) || errors.Is(err, services.ErrModerationNotAuthorized) {
				c.JSON(http.StatusForbidden, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to retrieve bans",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"bans":   bans,
			"total":  total,
			"limit":  limit,
			"offset": offset,
		})
		return
	}

	channelID, err := uuid.Parse(channelIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid channelId format",
		})
		return
	}

	// Get bans from moderation service
	bans, total, err := h.moderationService.GetBans(ctx, channelID, moderatorID, page, limit)
	if err != nil {
		if errors.Is(err, services.ErrModerationPermissionDenied) || errors.Is(err, services.ErrModerationNotAuthorized) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": err.Error(),
			})
			return
		}
		if errors.Is(err, services.ErrModerationCommunityNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve bans",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bans":   bans,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// CreateBan creates a new ban
// POST /api/v1/moderation/ban
func (h *ModerationHandler) CreateBan(c *gin.Context) {
	ctx := c.Request.Context()

	// Get moderator ID from context
	moderatorIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	moderatorID := moderatorIDVal.(uuid.UUID)

	// Parse request body
	var req struct {
		ChannelID string  `json:"channelId" binding:"required"`
		UserID    string  `json:"userId" binding:"required"`
		Reason    *string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	// Parse UUIDs
	channelID, err := uuid.Parse(req.ChannelID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid channelId format",
		})
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid userId format",
		})
		return
	}

	// Check if moderation service is available
	if h.moderationService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Moderation service not available",
		})
		return
	}

	// Create ban using moderation service
	err = h.moderationService.BanUser(ctx, channelID, moderatorID, userID, req.Reason)
	if err != nil {
		if errors.Is(err, services.ErrModerationPermissionDenied) || errors.Is(err, services.ErrModerationNotAuthorized) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": err.Error(),
			})
			return
		}
		if errors.Is(err, services.ErrModerationCommunityNotFound) || errors.Is(err, services.ErrModerationUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}
		if errors.Is(err, services.ErrModerationCannotBanOwner) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create ban",
		})
		return
	}

	// Retrieve the created ban to return full details using direct query
	// Note: We use a direct query here because we need the most recently created ban
	// and CommunityRepository doesn't have a method to get ban by channel+user
	var ban models.CommunityBan
	err = h.db.QueryRow(ctx, `
		SELECT id, community_id, banned_user_id, banned_by_user_id, reason, banned_at
		FROM community_bans
		WHERE community_id = $1 AND banned_user_id = $2
		ORDER BY banned_at DESC
		LIMIT 1
	`, channelID, userID).Scan(
		&ban.ID, &ban.CommunityID, &ban.BannedUserID, &ban.BannedByUserID, &ban.Reason, &ban.BannedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Ban created but failed to retrieve details",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        ban.ID,
		"channelId": ban.CommunityID,
		"userId":    ban.BannedUserID,
		"bannedBy":  ban.BannedByUserID,
		"reason":    ban.Reason,
		"bannedAt":  ban.BannedAt,
	})
}

// RevokeBan revokes/deletes a ban
// DELETE /api/v1/moderation/ban/:id
func (h *ModerationHandler) RevokeBan(c *gin.Context) {
	ctx := c.Request.Context()

	// Get moderator ID from context
	moderatorIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	moderatorID := moderatorIDVal.(uuid.UUID)

	// Parse ban ID from URL
	banIDStr := c.Param("id")
	banID, err := uuid.Parse(banIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ban ID format",
		})
		return
	}

	// Check if moderation service is available
	if h.moderationService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Moderation service not available",
		})
		return
	}

	// Check if community repository is available
	if h.communityRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Community repository not available",
		})
		return
	}

	// Retrieve the ban to get channel and user IDs using repository
	ban, err := h.communityRepo.GetBanByID(ctx, banID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Ban not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve ban details",
		})
		return
	}

	// Unban user using moderation service
	err = h.moderationService.UnbanUser(ctx, ban.CommunityID, moderatorID, ban.BannedUserID)
	if err != nil {
		if errors.Is(err, services.ErrModerationPermissionDenied) || errors.Is(err, services.ErrModerationNotAuthorized) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": err.Error(),
			})
			return
		}
		if errors.Is(err, services.ErrModerationNotBanned) || errors.Is(err, services.ErrModerationCommunityNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to revoke ban",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Ban revoked successfully",
	})
}

// GetBanDetails retrieves details of a specific ban
// GET /api/v1/moderation/ban/:id
func (h *ModerationHandler) GetBanDetails(c *gin.Context) {
	ctx := c.Request.Context()

	// Get moderator ID from context
	moderatorIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	moderatorID := moderatorIDVal.(uuid.UUID)

	// Parse ban ID from URL
	banIDStr := c.Param("id")
	banID, err := uuid.Parse(banIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ban ID format",
		})
		return
	}

	// Check if moderation service is available
	if h.moderationService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Moderation service not available",
		})
		return
	}

	// Check if community repository is available
	if h.communityRepo == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Community repository not available",
		})
		return
	}

	// Retrieve ban from repository
	ban, err := h.communityRepo.GetBanByID(ctx, banID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Ban not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve ban details",
		})
		return
	}

	// Verify moderator has permission to view bans for this channel
	err = h.moderationService.HasModerationPermission(ctx, ban.CommunityID, moderatorID)
	if err != nil {
		if errors.Is(err, services.ErrModerationPermissionDenied) || errors.Is(err, services.ErrModerationNotAuthorized) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "You do not have permission to view this ban",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to verify permissions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        ban.ID,
		"channelId": ban.CommunityID,
		"userId":    ban.BannedUserID,
		"bannedBy":  ban.BannedByUserID,
		"reason":    ban.Reason,
		"bannedAt":  ban.BannedAt,
	})
}

// ListModerators retrieves moderators for a channel
// GET /api/v1/moderation/moderators
func (h *ModerationHandler) ListModerators(c *gin.Context) {
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

	// Parse query parameters
	channelIDStr := c.Query("channelId")
	if channelIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "channelId query parameter is required",
		})
		return
	}

	channelID, err := uuid.Parse(channelIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid channelId format",
		})
		return
	}

	// Parse pagination parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Validate pagination parameters
	if limit < 1 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	// Check permission to view moderators - must be admin, site moderator, or channel owner/admin
	if err := h.validateModeratorListPermission(ctx, channelID, userID); err != nil {
		if errors.Is(err, services.ErrModerationPermissionDenied) || errors.Is(err, services.ErrModerationNotAuthorized) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "You do not have permission to view moderators for this channel",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to verify permissions",
		})
		return
	}

	// Get all moderators (both mods and admins) with proper pagination
	// Query both roles in a single database query
	query := `
		SELECT id, community_id, user_id, role, joined_at
		FROM community_members
		WHERE community_id = $1 AND (role = $2 OR role = $3)
		ORDER BY joined_at DESC
		LIMIT $4 OFFSET $5
	`

	// Count total
	var total int
	err = h.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM community_members
		WHERE community_id = $1 AND (role = $2 OR role = $3)
	`, channelID, models.CommunityRoleMod, models.CommunityRoleAdmin).Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to count moderators",
		})
		return
	}

	// Get moderators
	rows, err := h.db.Query(ctx, query, channelID, models.CommunityRoleMod, models.CommunityRoleAdmin, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve moderators",
		})
		return
	}
	defer rows.Close()

	var moderators []*models.CommunityMember
	for rows.Next() {
		member := &models.CommunityMember{}
		err := rows.Scan(&member.ID, &member.CommunityID, &member.UserID, &member.Role, &member.JoinedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to scan moderator",
			})
			return
		}
		moderators = append(moderators, member)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to iterate moderators",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"moderators": moderators,
		"total":      total,
		"limit":      limit,
		"offset":     offset,
	})
}

// AddModerator adds a moderator to a channel
// POST /api/v1/moderation/moderators
func (h *ModerationHandler) AddModerator(c *gin.Context) {
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

	// Parse request body
	var req struct {
		UserID    string  `json:"userId" binding:"required"`
		ChannelID string  `json:"channelId" binding:"required"`
		Reason    *string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	// Parse UUIDs
	channelID, err := uuid.Parse(req.ChannelID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid channelId format",
		})
		return
	}

	targetUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid userId format",
		})
		return
	}

	// Check permission to add moderators - must be admin or channel owner/admin
	if err := h.validateModeratorManagementPermission(ctx, channelID, userID); err != nil {
		if errors.Is(err, services.ErrModerationPermissionDenied) || errors.Is(err, services.ErrModerationNotAuthorized) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "You do not have permission to add moderators to this channel",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to verify permissions",
		})
		return
	}

	// Validate scope - community moderators can't assign mods to other channels
	if err := h.validateModeratorScope(ctx, channelID, userID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Community moderators can only manage moderators for their assigned channels",
		})
		return
	}

	// Check if target user exists
	var userExists bool
	err = h.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", targetUserID).Scan(&userExists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to verify user",
		})
		return
	}
	if !userExists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	// Check if user is already a member
	existingMember, err := h.communityRepo.GetMember(ctx, channelID, targetUserID)
	if err == nil && existingMember != nil {
		// Check if user is already an admin - don't demote them
		if existingMember.Role == models.CommunityRoleAdmin {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "User is already a channel admin. Cannot demote to moderator.",
			})
			return
		}
		// User is already a member, update their role to mod
		if err := h.communityRepo.UpdateMemberRole(ctx, channelID, targetUserID, models.CommunityRoleMod); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update member role",
			})
			return
		}
	} else {
		// User is not a member, add them as a moderator
		member := &models.CommunityMember{
			ID:          uuid.New(),
			CommunityID: channelID,
			UserID:      targetUserID,
			Role:        models.CommunityRoleMod,
		}
		if err := h.communityRepo.AddMember(ctx, member); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to add moderator",
			})
			return
		}
	}

	// Create audit log
	metadata := map[string]interface{}{
		"channel_id":     channelID.String(),
		"target_user_id": targetUserID.String(),
		"assigned_by":    userID.String(),
		"new_role":       models.CommunityRoleMod,
	}
	if req.Reason != nil {
		metadata["reason"] = *req.Reason
	}

	auditLog := &models.ModerationAuditLog{
		Action:      "add_moderator",
		EntityType:  "community_member",
		EntityID:    targetUserID,
		ModeratorID: userID,
		Reason:      req.Reason,
		Metadata:    metadata,
	}

	// Create audit log
	h.createModeratorAuditLog(ctx, auditLog)

	// Retrieve the updated member to return
	member, err := h.communityRepo.GetMember(ctx, channelID, targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Moderator added but failed to retrieve details",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":   true,
		"moderator": member,
		"message":   "Moderator added successfully",
	})
}

// RemoveModerator removes a moderator from a channel
// DELETE /api/v1/moderation/moderators/:id
func (h *ModerationHandler) RemoveModerator(c *gin.Context) {
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

	// Parse member ID from URL
	memberIDStr := c.Param("id")
	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid moderator ID format",
		})
		return
	}

	// Get the member to find channel and user IDs
	var member models.CommunityMember
	err = h.db.QueryRow(ctx, `
		SELECT id, community_id, user_id, role
		FROM community_members
		WHERE id = $1
	`, memberID).Scan(&member.ID, &member.CommunityID, &member.UserID, &member.Role)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Moderator not found",
		})
		return
	}

	// Check permission to remove moderators
	if err := h.validateModeratorManagementPermission(ctx, member.CommunityID, userID); err != nil {
		if errors.Is(err, services.ErrModerationPermissionDenied) || errors.Is(err, services.ErrModerationNotAuthorized) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "You do not have permission to remove moderators from this channel",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to verify permissions",
		})
		return
	}

	// Validate scope - community moderators can't manage mods in other channels
	if err := h.validateModeratorScope(ctx, member.CommunityID, userID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Community moderators can only manage moderators for their assigned channels",
		})
		return
	}

	// Check if trying to remove a channel owner (admin role typically)
	community, err := h.communityRepo.GetCommunityByID(ctx, member.CommunityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve community",
		})
		return
	}
	if community.OwnerID == member.UserID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot remove the channel owner",
		})
		return
	}

	// Update member role to 'member' instead of removing them entirely
	if err := h.communityRepo.UpdateMemberRole(ctx, member.CommunityID, member.UserID, models.CommunityRoleMember); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to remove moderator",
		})
		return
	}

	// Create audit log
	metadata := map[string]interface{}{
		"channel_id":     member.CommunityID.String(),
		"target_user_id": member.UserID.String(),
		"removed_by":     userID.String(),
		"previous_role":  member.Role,
		"new_role":       models.CommunityRoleMember,
	}

	auditLog := &models.ModerationAuditLog{
		Action:      "remove_moderator",
		EntityType:  "community_member",
		EntityID:    member.UserID,
		ModeratorID: userID,
		Metadata:    metadata,
	}

	// Create audit log
	h.createModeratorAuditLog(ctx, auditLog)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Moderator removed successfully",
	})
}

// UpdateModeratorPermissions updates a moderator's permissions (role)
// PATCH /api/v1/moderation/moderators/:id
func (h *ModerationHandler) UpdateModeratorPermissions(c *gin.Context) {
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

	// Parse member ID from URL
	memberIDStr := c.Param("id")
	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid moderator ID format",
		})
		return
	}

	// Parse request body
	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: role is required",
		})
		return
	}

	// Validate role using constants
	if req.Role != models.CommunityRoleMod && req.Role != models.CommunityRoleAdmin {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid role: must be 'mod' or 'admin'",
		})
		return
	}

	// Get the member to find channel and user IDs
	var member models.CommunityMember
	err = h.db.QueryRow(ctx, `
		SELECT id, community_id, user_id, role
		FROM community_members
		WHERE id = $1
	`, memberID).Scan(&member.ID, &member.CommunityID, &member.UserID, &member.Role)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Moderator not found",
		})
		return
	}

	// Check permission to update moderators
	if err := h.validateModeratorManagementPermission(ctx, member.CommunityID, userID); err != nil {
		if errors.Is(err, services.ErrModerationPermissionDenied) || errors.Is(err, services.ErrModerationNotAuthorized) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "You do not have permission to update moderators in this channel",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to verify permissions",
		})
		return
	}

	// Validate scope - community moderators can't manage mods in other channels
	if err := h.validateModeratorScope(ctx, member.CommunityID, userID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Community moderators can only manage moderators for their assigned channels",
		})
		return
	}

	// Check if trying to modify a channel owner
	community, err := h.communityRepo.GetCommunityByID(ctx, member.CommunityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve community",
		})
		return
	}
	if community.OwnerID == member.UserID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot modify the channel owner's role",
		})
		return
	}

	// Update member role
	if err := h.communityRepo.UpdateMemberRole(ctx, member.CommunityID, member.UserID, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update moderator permissions",
		})
		return
	}

	// Create audit log
	metadata := map[string]interface{}{
		"channel_id":     member.CommunityID.String(),
		"target_user_id": member.UserID.String(),
		"updated_by":     userID.String(),
		"previous_role":  member.Role,
		"new_role":       req.Role,
	}

	auditLog := &models.ModerationAuditLog{
		Action:      "update_moderator_permissions",
		EntityType:  "community_member",
		EntityID:    member.UserID,
		ModeratorID: userID,
		Metadata:    metadata,
	}

	// Create audit log
	h.createModeratorAuditLog(ctx, auditLog)

	// Retrieve updated member
	updatedMember, err := h.communityRepo.GetMember(ctx, member.CommunityID, member.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Permissions updated but failed to retrieve details",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"moderator": updatedMember,
		"message":   "Moderator permissions updated successfully",
	})
}

// getUserDetails fetches user details from the database
// This helper reduces code duplication across permission validation functions
func (h *ModerationHandler) getUserDetails(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var user models.User
	err := h.db.QueryRow(ctx, `
		SELECT id, role, account_type, moderator_scope
		FROM users
		WHERE id = $1
	`, userID).Scan(&user.ID, &user.Role, &user.AccountType, &user.ModeratorScope)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// isAdminOrSiteModerator checks if a user has admin or site-wide moderator privileges
func isAdminOrSiteModerator(user *models.User) bool {
	if user.Role == models.RoleAdmin || user.AccountType == models.AccountTypeAdmin {
		return true
	}
	if user.AccountType == models.AccountTypeModerator && user.ModeratorScope == models.ModeratorScopeSite {
		return true
	}
	return false
}

// validateModeratorListPermission checks if a user can view moderators for a channel
func (h *ModerationHandler) validateModeratorListPermission(ctx context.Context, channelID, userID uuid.UUID) error {
	// Check if user exists and get their details
	user, err := h.getUserDetails(ctx, userID)
	if err != nil {
		return err
	}

	// Admins and site moderators can view any channel's moderators
	if isAdminOrSiteModerator(user) {
		return nil
	}

	// Channel owners and admins can view their channel's moderators
	community, err := h.communityRepo.GetCommunityByID(ctx, channelID)
	if err != nil {
		return services.ErrModerationCommunityNotFound
	}
	if community.OwnerID == userID {
		return nil
	}

	// Channel admins can view
	member, err := h.communityRepo.GetMember(ctx, channelID, userID)
	if err == nil && member != nil && member.Role == models.CommunityRoleAdmin {
		return nil
	}

	return services.ErrModerationPermissionDenied
}

// validateModeratorManagementPermission checks if a user can add/remove/update moderators
func (h *ModerationHandler) validateModeratorManagementPermission(ctx context.Context, channelID, userID uuid.UUID) error {
	// Check if user exists and get their details
	user, err := h.getUserDetails(ctx, userID)
	if err != nil {
		return err
	}

	// Admins can manage any channel's moderators
	if isAdminOrSiteModerator(user) {
		return nil
	}

	// Channel owners can manage their channel's moderators
	community, err := h.communityRepo.GetCommunityByID(ctx, channelID)
	if err != nil {
		return services.ErrModerationCommunityNotFound
	}
	if community.OwnerID == userID {
		return nil
	}

	// Channel admins can manage moderators
	member, err := h.communityRepo.GetMember(ctx, channelID, userID)
	if err == nil && member != nil && member.Role == models.CommunityRoleAdmin {
		return nil
	}

	return services.ErrModerationPermissionDenied
}

// validateModeratorScope checks if a community moderator is managing their assigned channels
func (h *ModerationHandler) validateModeratorScope(ctx context.Context, channelID, userID uuid.UUID) error {
	// Get user details with moderation channels
	var user models.User
	var moderationChannels []uuid.UUID
	err := h.db.QueryRow(ctx, `
		SELECT id, role, account_type, moderator_scope, moderation_channels
		FROM users
		WHERE id = $1
	`, userID).Scan(&user.ID, &user.Role, &user.AccountType, &user.ModeratorScope, &moderationChannels)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Admins and site moderators have no scope restrictions
	if isAdminOrSiteModerator(&user) {
		return nil
	}

	// Channel owners have no restrictions for their channels
	community, err := h.communityRepo.GetCommunityByID(ctx, channelID)
	if err != nil {
		return services.ErrModerationCommunityNotFound
	}
	if community.OwnerID == userID {
		return nil
	}

	// Community moderators must have the channel in their moderation scope
	if user.AccountType == models.AccountTypeCommunityModerator {
		if user.ModeratorScope != models.ModeratorScopeCommunity {
			return services.ErrModerationNotAuthorized
		}

		// Check if this channel is in their authorized scope
		for _, authorizedChannelID := range moderationChannels {
			if authorizedChannelID == channelID {
				return nil
			}
		}
		return services.ErrModerationNotAuthorized
	}

	// Default: deny access for any other case
	return services.ErrModerationPermissionDenied
}

// createModeratorAuditLog creates an audit log entry for moderator management actions
// Uses AuditLogRepository for consistency with other audit logging in the codebase
func (h *ModerationHandler) createModeratorAuditLog(ctx context.Context, auditLog *models.ModerationAuditLog) {
	logger := utils.GetLogger()

	if err := h.auditLogRepo.Create(ctx, auditLog); err != nil {
		logger.Error("Failed to create audit log", err, map[string]interface{}{
			"action":      auditLog.Action,
			"entity_type": auditLog.EntityType,
			"entity_id":   auditLog.EntityID,
			"moderator":   auditLog.ModeratorID,
			"reason":      auditLog.Reason,
			"metadata":    auditLog.Metadata,
			"error":       err.Error(),
		})
	}
}

// TwitchBanUser bans a user on Twitch via the Twitch API
// POST /api/v1/moderation/twitch/ban
// This enforces Twitch-specific scope requirements and blocks site moderators
func (h *ModerationHandler) TwitchBanUser(c *gin.Context) {
	ctx := c.Request.Context()
	logger := utils.GetLogger()

	// Get moderator ID from context
	moderatorIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	moderatorID := moderatorIDVal.(uuid.UUID)

	// Parse request body
	var req struct {
		BroadcasterID string  `json:"broadcasterID" binding:"required"` // Twitch broadcaster ID (string, not UUID)
		UserID        string  `json:"userID" binding:"required"`        // Twitch user ID to ban (string)
		Reason        *string `json:"reason"`
		Duration      *int    `json:"duration"` // Duration in seconds for timeout, omit for permanent ban
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	// Validate Twitch moderation service is available
	if h.twitchModerationService == nil {
		logger.Error("Twitch moderation service not configured", nil, map[string]interface{}{
			"moderator_id": moderatorID.String(),
		})
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Twitch moderation service not available. Please ensure Twitch integration is configured.",
		})
		return
	}

	// Ban user on Twitch with scope validation
	err := h.twitchModerationService.BanUserOnTwitch(ctx, moderatorID, req.BroadcasterID, req.UserID, req.Reason, req.Duration)
	if err != nil {
		// Handle specific error types with appropriate HTTP status codes
		if errors.Is(err, services.ErrSiteModeratorsReadOnly) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":  err.Error(),
				"code":   "SITE_MODERATORS_READ_ONLY",
				"detail": "Site moderators can view Twitch bans but cannot create them. You must be the Twitch channel broadcaster or a Twitch-recognized moderator for this channel.",
			})
			return
		}
		if errors.Is(err, services.ErrTwitchNotAuthenticated) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":  err.Error(),
				"code":   "NOT_AUTHENTICATED",
				"detail": "You must connect your Twitch account to perform this action.",
			})
			return
		}
		if errors.Is(err, services.ErrTwitchScopeInsufficient) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":  err.Error(),
				"code":   "INSUFFICIENT_SCOPES",
				"detail": "Your Twitch account does not have the required permissions (moderator:manage:banned_users or channel:manage:banned_users). Please re-authenticate with Twitch.",
			})
			return
		}
		if errors.Is(err, services.ErrTwitchNotBroadcaster) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":  err.Error(),
				"code":   "NOT_BROADCASTER",
				"detail": "Only the channel broadcaster can perform Twitch ban actions for this channel.",
			})
			return
		}

		// Log error for debugging
		logger.Error("Twitch ban failed", err, map[string]interface{}{
			"moderator_id":   moderatorID.String(),
			"broadcaster_id": req.BroadcasterID,
			"target_user_id": req.UserID,
		})

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to ban user on Twitch",
		})
		return
	}

	logger.Info("User banned on Twitch", map[string]interface{}{
		"moderator_id":   moderatorID.String(),
		"broadcaster_id": req.BroadcasterID,
		"target_user_id": req.UserID,
	})

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"message":       "User banned on Twitch successfully",
		"broadcasterID": req.BroadcasterID,
		"userID":        req.UserID,
	})
}

// TwitchUnbanUser unbans a user on Twitch via the Twitch API
// DELETE /api/v1/moderation/twitch/ban
// This enforces Twitch-specific scope requirements and blocks site moderators
func (h *ModerationHandler) TwitchUnbanUser(c *gin.Context) {
	ctx := c.Request.Context()
	logger := utils.GetLogger()

	// Get moderator ID from context
	moderatorIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	moderatorID := moderatorIDVal.(uuid.UUID)

	// Parse request body or query params
	broadcasterID := c.Query("broadcasterID")
	userID := c.Query("userID")

	if broadcasterID == "" || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "broadcasterID and userID query parameters are required",
		})
		return
	}

	// Validate Twitch moderation service is available
	if h.twitchModerationService == nil {
		logger.Error("Twitch moderation service not configured", nil, map[string]interface{}{
			"moderator_id": moderatorID.String(),
		})
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Twitch moderation service not available. Please ensure Twitch integration is configured.",
		})
		return
	}

	// Unban user on Twitch with scope validation
	err := h.twitchModerationService.UnbanUserOnTwitch(ctx, moderatorID, broadcasterID, userID)
	if err != nil {
		// Handle specific error types with appropriate HTTP status codes
		if errors.Is(err, services.ErrSiteModeratorsReadOnly) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":  err.Error(),
				"code":   "SITE_MODERATORS_READ_ONLY",
				"detail": "Site moderators can view Twitch bans but cannot remove them. You must be the Twitch channel broadcaster or a Twitch-recognized moderator for this channel.",
			})
			return
		}
		if errors.Is(err, services.ErrTwitchNotAuthenticated) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":  err.Error(),
				"code":   "NOT_AUTHENTICATED",
				"detail": "You must connect your Twitch account to perform this action.",
			})
			return
		}
		if errors.Is(err, services.ErrTwitchScopeInsufficient) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":  err.Error(),
				"code":   "INSUFFICIENT_SCOPES",
				"detail": "Your Twitch account does not have the required permissions (moderator:manage:banned_users or channel:manage:banned_users). Please re-authenticate with Twitch.",
			})
			return
		}
		if errors.Is(err, services.ErrTwitchNotBroadcaster) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":  err.Error(),
				"code":   "NOT_BROADCASTER",
				"detail": "Only the channel broadcaster can perform Twitch unban actions for this channel.",
			})
			return
		}

		// Log error for debugging
		logger.Error("Twitch unban failed", err, map[string]interface{}{
			"moderator_id":   moderatorID.String(),
			"broadcaster_id": broadcasterID,
			"target_user_id": userID,
		})

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to unban user on Twitch",
		})
		return
	}

	logger.Info("User unbanned on Twitch", map[string]interface{}{
		"moderator_id":   moderatorID.String(),
		"broadcaster_id": broadcasterID,
		"target_user_id": userID,
	})

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"message":       "User unbanned on Twitch successfully",
		"broadcasterID": broadcasterID,
		"userID":        userID,
	})
}
