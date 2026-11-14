package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	secondsPerHour         = 3600
	banSyncTimeoutDuration = 5 * time.Minute
)

func validModerationEventType(eventType services.ModerationEventType) bool {
	switch eventType {
	case services.ModerationEventSubmissionReceived,
		services.ModerationEventSubmissionSuspicious,
		services.ModerationEventSubmissionAutoRejected,
		services.ModerationEventSubmissionApproved,
		services.ModerationEventSubmissionRejected,
		services.ModerationEventSubmissionDuplicate,
		services.ModerationEventAbuseDetected,
		services.ModerationEventRateLimitExceeded,
		services.ModerationEventVelocityViolation,
		services.ModerationEventIPShareSuspicious,
		services.ModerationEventUserCooldownActivated:
		return true
	default:
		return false
	}
}

func parseAdminDateRange(c *gin.Context, defaultDays int) (time.Time, time.Time, bool) {
	now := time.Now().UTC()
	start := now.AddDate(0, 0, -defaultDays)
	end := now
	var err error
	if raw := c.Query("start_date"); raw != "" {
		start, err = time.Parse("2006-01-02", raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "start_date must use YYYY-MM-DD"})
			return time.Time{}, time.Time{}, false
		}
	}
	if raw := c.Query("end_date"); raw != "" {
		end, err = time.Parse("2006-01-02", raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "end_date must use YYYY-MM-DD"})
			return time.Time{}, time.Time{}, false
		}
	}
	if end.Before(start) || end.Sub(start) > 366*24*time.Hour {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date range must be ordered and no longer than 366 days"})
		return time.Time{}, time.Time{}, false
	}
	return start, end, true
}

// ModerationHandler handles moderation operations
type ModerationHandler struct {
	moderationEventService *services.ModerationEventService
	moderationService      *services.ModerationService
	abuseDetector          *services.SubmissionAbuseDetector
	toxicityClassifier     *services.ToxicityClassifier
	communityRepo          *repository.CommunityRepository
	auditLogRepo           *repository.AuditLogRepository
	db                     *pgxpool.Pool
}

// NewModerationHandler creates a new ModerationHandler
func NewModerationHandler(moderationEventService *services.ModerationEventService, moderationService *services.ModerationService, abuseDetector *services.SubmissionAbuseDetector, toxicityClassifier *services.ToxicityClassifier, communityRepo *repository.CommunityRepository, auditLogRepo *repository.AuditLogRepository, db *pgxpool.Pool) *ModerationHandler {
	return &ModerationHandler{
		moderationEventService: moderationEventService,
		moderationService:      moderationService,
		abuseDetector:          abuseDetector,
		toxicityClassifier:     toxicityClassifier,
		communityRepo:          communityRepo,
		auditLogRepo:           auditLogRepo,
		db:                     db,
	}
}

// GetPendingEvents retrieves pending moderation events
// GET /admin/moderation/events
func (h *ModerationHandler) GetPendingEvents(c *gin.Context) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 100"})
		return
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
	if !validModerationEventType(eventType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid moderation event type"})
		return
	}
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 100"})
		return
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
	reviewerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	err = h.moderationEventService.MarkEventReviewed(c.Request.Context(), eventID, reviewerID)
	if err != nil {
		if errors.Is(err, services.ErrModerationEventNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
			return
		}
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
	reviewerID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	var req struct {
		Action string `json:"action" binding:"required,oneof=approve reject escalate"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Action is required",
		})
		return
	}

	err = h.moderationEventService.ProcessEvent(c.Request.Context(), eventID, reviewerID, req.Action)
	if err != nil {
		if errors.Is(err, services.ErrModerationEventNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
			return
		}
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
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 100"})
		return
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

	// Build query with filters. The joins add a short preview of the flagged
	// comment or clip; they yield NULL for other content types and for content
	// that no longer exists.
	query := `
		SELECT mq.id, mq.content_type, mq.content_id, mq.reason, mq.priority,
		       mq.status, mq.assigned_to, mq.reported_by, mq.report_count,
		       mq.auto_flagged, mq.confidence_score, mq.created_at,
		       mq.reviewed_at, mq.reviewed_by,
		       CASE mq.content_type
		           WHEN 'comment' THEN LEFT(cm.content, ` + strconv.Itoa(queueExcerptLength) + `)
		           WHEN 'clip' THEN cl.title
		       END AS content_excerpt,
		       CASE mq.content_type
		           WHEN 'comment' THEN cmu.username
		           WHEN 'clip' THEN COALESCE(clu.username, cl.creator_name)
		       END AS content_author,
		       CASE mq.content_type
		           WHEN 'comment' THEN '/clip/' || cm.clip_id::text || '#comment-' || cm.id::text
		           WHEN 'clip' THEN '/clip/' || cl.id::text
		       END AS content_url
		FROM moderation_queue mq
		LEFT JOIN comments cm ON mq.content_type = 'comment' AND cm.id = mq.content_id
		LEFT JOIN users cmu ON cmu.id = cm.user_id
		LEFT JOIN clips cl ON mq.content_type = 'clip' AND cl.id = mq.content_id
		LEFT JOIN users clu ON clu.id = cl.submitted_by_user_id
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

	var items []moderationQueueListItem
	for rows.Next() {
		var item moderationQueueListItem
		err := rows.Scan(
			&item.ID, &item.ContentType, &item.ContentID, &item.Reason,
			&item.Priority, &item.Status, &item.AssignedTo, &item.ReportedBy,
			&item.ReportCount, &item.AutoFlagged, &item.ConfidenceScore,
			&item.CreatedAt, &item.ReviewedAt, &item.ReviewedBy,
			&item.ContentExcerpt, &item.ContentAuthor, &item.ContentURL,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read moderation queue"})
			return
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read moderation queue"})
		return
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

// ApproveContent approves a moderation queue item. The flagged content is left as it is.
// POST /admin/moderation/:id/approve
func (h *ModerationHandler) ApproveContent(c *gin.Context) {
	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid item ID",
		})
		return
	}

	// Get moderator ID from context
	moderatorID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	if !h.decideQueueItem(c, itemID, moderatorID, "approve", nil) {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Content approved",
	})
}

// RejectContent rejects a moderation queue item and soft-removes the flagged
// comment or clip in the same transaction.
// POST /admin/moderation/:id/reject
func (h *ModerationHandler) RejectContent(c *gin.Context) {
	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid item ID",
		})
		return
	}

	// Get moderator ID from context
	moderatorID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	// Parse request body for optional reason
	var req struct {
		Reason *string `json:"reason" binding:"omitempty,max=1000"`
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

	if !h.decideQueueItem(c, itemID, moderatorID, "reject", req.Reason) {
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid moderation request"})
		return
	}

	// Get moderator ID from context
	moderatorID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	// Convert item IDs to UUIDs
	itemIDs := make([]uuid.UUID, 0, len(req.ItemIDs))
	seen := make(map[uuid.UUID]struct{}, len(req.ItemIDs))
	for _, idStr := range req.ItemIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid item ID: " + idStr,
			})
			return
		}
		if _, duplicate := seen[id]; duplicate {
			c.JSON(http.StatusBadRequest, gin.H{"error": "item_ids must be unique"})
			return
		}
		seen[id] = struct{}{}
		itemIDs = append(itemIDs, id)
	}

	// Lock rows in one order so two overlapping bulk requests cannot deadlock
	sort.Slice(itemIDs, func(i, j int) bool {
		return bytes.Compare(itemIDs[i][:], itemIDs[j][:]) < 0
	})

	// Begin transaction
	tx, err := h.db.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to begin transaction",
		})
		return
	}
	defer tx.Rollback(ctx)

	// Apply the action to every item that is still open. Items another
	// moderator already handled are skipped and reported back.
	processed := make([]uuid.UUID, 0, len(itemIDs))
	skipped := make([]uuid.UUID, 0)
	for _, itemID := range itemIDs {
		outcome, _, err := applyQueueDecision(ctx, tx, itemID, moderatorID, req.Action, "bulk_"+req.Action, req.Reason)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to moderate items"})
			return
		}
		if outcome == queueDecisionApplied {
			processed = append(processed, itemID)
		} else {
			skipped = append(skipped, itemID)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to commit transaction",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"processed":     len(processed),
		"processed_ids": processed,
		"skipped":       skipped,
		"total":         len(itemIDs),
	})
}

// queueExcerptLength is the most characters of a flagged comment shown in the queue.
const queueExcerptLength = 280

// moderationQueueListItem is a queue item with a preview of the flagged content.
type moderationQueueListItem struct {
	models.ModerationQueueItem
	ContentExcerpt *string `json:"content_excerpt,omitempty"`
	ContentAuthor  *string `json:"content_author,omitempty"`
	ContentURL     *string `json:"content_url,omitempty"`
}

type queueDecisionOutcome int

const (
	queueDecisionApplied  queueDecisionOutcome = iota
	queueDecisionNotFound                      // no queue item has this ID
	queueDecisionStale                         // the item was already handled
)

// queueDecisionStatus maps a queue action to the status it sets.
var queueDecisionStatus = map[string]string{
	"approve":  "approved",
	"reject":   "rejected",
	"escalate": "escalated",
}

// applyQueueDecision applies approve, reject or escalate to one queue item
// inside tx. Approve and reject work on pending and escalated items; escalate
// works on pending items only. A reject also soft-removes the flagged comment
// or clip, with the same update as CommentRepository.RemoveComment and
// ClipRepository.RemoveClip. Every applied decision writes a
// moderation_decisions row and a moderation_audit_logs row under auditAction.
// For a stale item it returns the item's current status and changes nothing.
func applyQueueDecision(ctx context.Context, tx pgx.Tx, itemID, moderatorID uuid.UUID, action, auditAction string, reason *string) (queueDecisionOutcome, string, error) {
	newStatus, ok := queueDecisionStatus[action]
	if !ok {
		return 0, "", fmt.Errorf("unknown moderation action %q", action)
	}

	var contentType, queueReason, currentStatus string
	var contentID uuid.UUID
	err := tx.QueryRow(ctx, `
		SELECT content_type, content_id, reason, status
		FROM moderation_queue
		WHERE id = $1
		FOR UPDATE
	`, itemID).Scan(&contentType, &contentID, &queueReason, &currentStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return queueDecisionNotFound, "", nil
	}
	if err != nil {
		return 0, "", fmt.Errorf("load moderation queue item: %w", err)
	}
	if currentStatus != "pending" && (currentStatus != "escalated" || action == "escalate") {
		return queueDecisionStale, currentStatus, nil
	}

	if _, err = tx.Exec(ctx, `
		UPDATE moderation_queue
		SET status = $1, reviewed_by = $2, reviewed_at = NOW()
		WHERE id = $3
	`, newStatus, moderatorID, itemID); err != nil {
		return 0, "", fmt.Errorf("update moderation queue item: %w", err)
	}

	if _, err = tx.Exec(ctx, `
		INSERT INTO moderation_decisions (queue_item_id, moderator_id, action, reason)
		VALUES ($1, $2, $3, $4)
	`, itemID, moderatorID, action, reason); err != nil {
		return 0, "", fmt.Errorf("record moderation decision: %w", err)
	}

	contentRemoved := false
	if action == "reject" {
		removedReason := "Content removed: " + queueReason
		if reason != nil && strings.TrimSpace(*reason) != "" {
			removedReason = *reason
		}
		var tag pgconn.CommandTag
		switch contentType {
		case "comment":
			tag, err = tx.Exec(ctx, `
				UPDATE comments
				SET is_removed = true, removed_reason = $2
				WHERE id = $1
			`, contentID, removedReason)
		case "clip":
			tag, err = tx.Exec(ctx, `
				UPDATE clips
				SET is_removed = true, removed_reason = $2
				WHERE id = $1
			`, contentID, removedReason)
		}
		if err != nil {
			return 0, "", fmt.Errorf("remove rejected %s: %w", contentType, err)
		}
		// False when the content was already deleted
		contentRemoved = tag.RowsAffected() > 0
	}

	metadata, err := json.Marshal(map[string]interface{}{
		"content_type":    contentType,
		"content_id":      contentID.String(),
		"previous_status": currentStatus,
		"new_status":      newStatus,
		"content_removed": contentRemoved,
	})
	if err != nil {
		return 0, "", fmt.Errorf("marshal moderation audit metadata: %w", err)
	}
	if _, err = tx.Exec(ctx, `
		INSERT INTO moderation_audit_logs (
			id, action, entity_type, entity_id, moderator_id, actor_id, reason, metadata, created_at
		) VALUES ($1, $2, 'moderation_queue', $3, $4, $4, $5, $6, $7)
	`, uuid.New(), auditAction, itemID, moderatorID, reason, metadata, time.Now().UTC()); err != nil {
		return 0, "", fmt.Errorf("insert moderation audit log: %w", err)
	}

	return queueDecisionApplied, newStatus, nil
}

// decideQueueItem applies one approve or reject and writes the error response
// when it does not go through. It reports whether the decision was committed.
func (h *ModerationHandler) decideQueueItem(c *gin.Context, itemID, moderatorID uuid.UUID, action string, reason *string) bool {
	ctx := c.Request.Context()

	tx, err := h.db.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to begin transaction",
		})
		return false
	}
	defer tx.Rollback(ctx)

	outcome, currentStatus, err := applyQueueDecision(ctx, tx, itemID, moderatorID, action, action, reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to " + action + " item",
		})
		return false
	}
	switch outcome {
	case queueDecisionNotFound:
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Item not found",
		})
		return false
	case queueDecisionStale:
		c.JSON(http.StatusConflict, gin.H{
			"error":  "This item was already handled. It is now " + currentStatus + ".",
			"code":   "ALREADY_HANDLED",
			"status": currentStatus,
		})
		return false
	}

	if err = tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to commit transaction",
		})
		return false
	}
	return true
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
				FLOOR(EXTRACT(EPOCH FROM (NOW() - MIN(created_at) FILTER (WHERE status = 'pending')))/`+fmt.Sprintf("%d", secondsPerHour)+`)::int as oldest_age
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
		SELECT 'status' as type, NULL::text as name, 0::bigint as count,
			   total_pending, total_approved, total_rejected, total_escalated,
			   auto_flagged_count, user_reported_count, high_priority_count, oldest_age
		FROM status_counts
		UNION ALL
		SELECT category, name::text, count, 0, 0, 0, 0, 0, 0, 0, NULL::int
		FROM type_counts
		UNION ALL
		SELECT category, name::text, count, 0, 0, 0, 0, 0, 0, 0, NULL::int
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
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to read stats",
			})
			return
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
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read stats",
		})
		return
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
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 100"})
		return
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read appeals"})
			return
		}
		appeals = append(appeals, appeal)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read appeals"})
		return
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
	adminID, ok := authenticatedUserID(c)
	if !ok {
		return
	}

	var req models.ResolveAppealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appeal resolution"})
		return
	}

	// Map decision to status
	status := "rejected"
	if req.Decision == "approve" {
		status = "approved"
	}

	tx, err := h.db.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve appeal"})
		return
	}
	defer tx.Rollback(ctx)
	var targetUserID uuid.UUID
	if err := tx.QueryRow(ctx, `SELECT user_id FROM moderation_appeals WHERE id = $1 FOR UPDATE`, appealID).Scan(&targetUserID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Appeal not found"})
		return
	}
	cmdTag, err := tx.Exec(ctx, `
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
		c.JSON(http.StatusConflict, gin.H{
			"error": "Appeal not found or not in pending status",
		})
		return
	}
	reason := "appeal " + status
	if req.Resolution != nil {
		reason = *req.Resolution
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO moderation_audit_logs
			(id, action, entity_type, entity_id, moderator_id, actor_id, target_user_id, reason, created_at)
		VALUES ($1, 'resolve_appeal', 'moderation_appeal', $2, $3, $3, $4, $5, NOW())`,
		uuid.New(), appealID, adminID, targetUserID, reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve appeal"})
		return
	}
	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve appeal"})
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
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "100"))
	if err != nil || limit < 1 || limit > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 1000"})
		return
	}
	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 || offset > 1_000_000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "offset must be between 0 and 1000000"})
		return
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

	var parsedStart, parsedEnd time.Time
	if startDate != "" {
		parsedStart, err = time.Parse("2006-01-02", startDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "start_date must use YYYY-MM-DD"})
			return
		}
		baseWhere += fmt.Sprintf(" AND md.created_at >= $%d", argIdx)
		args = append(args, parsedStart)
		argIdx++
	}

	if endDate != "" {
		parsedEnd, err = time.Parse("2006-01-02", endDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "end_date must use YYYY-MM-DD"})
			return
		}
		baseWhere += fmt.Sprintf(" AND md.created_at < ($%d::date + interval '1 day')", argIdx)
		args = append(args, parsedEnd)
		argIdx++
	}
	if !parsedStart.IsZero() && !parsedEnd.IsZero() && (parsedEnd.Before(parsedStart) || parsedEnd.Sub(parsedStart) > 366*24*time.Hour) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date range must be ordered and no longer than 366 days"})
		return
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
	err = h.db.QueryRow(ctx, countQuery, args...).Scan(&totalCount)
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
		// The moderator is NULL once their account is deleted.
		var moderatorID *uuid.UUID
		var moderatorName *string
		err := rows.Scan(
			&log.ID, &log.QueueItemID, &moderatorID,
			&moderatorName, &log.Action,
			&log.ContentType, &log.ContentID,
			&log.Reason, &log.Metadata, &log.CreatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read audit logs"})
			return
		}
		if moderatorID != nil {
			log.ModeratorID = *moderatorID
		}
		if moderatorName != nil {
			log.ModeratorName = *moderatorName
		}
		logs = append(logs, log)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read audit logs"})
		return
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

// GetToxicityMetrics retrieves toxicity classification metrics
// GET /admin/moderation/toxicity/metrics
func (h *ModerationHandler) GetToxicityMetrics(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse query parameters for date range
	startDate, endDate, ok := parseAdminDateRange(c, 30)
	if !ok {
		return
	}
	endDate = endDate.Add(24 * time.Hour)

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
	moderatorID, authenticated := moderatorIDVal.(uuid.UUID)
	if !exists || !authenticated || moderatorID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	// Parse request body
	var req struct {
		ChannelID string  `json:"channelId" binding:"required"`
		UserID    string  `json:"userId" binding:"required"`
		Reason    *string `json:"reason" binding:"omitempty,max=1000"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request",
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
	ban, err := h.moderationService.BanUserWithResult(ctx, channelID, moderatorID, userID, req.Reason)
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
		if errors.Is(err, services.ErrModerationCannotBanOwner) || errors.Is(err, services.ErrModerationCannotBanSelf) {
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

	c.JSON(http.StatusCreated, gin.H{
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

	// A banned user cannot be made a moderator; the ban has to be lifted first
	banned, err := h.communityRepo.IsBanned(ctx, channelID, targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check ban status",
		})
		return
	}
	if banned {
		c.JSON(http.StatusConflict, gin.H{
			"error": "This user is banned from the channel. Lift the ban before making them a moderator.",
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
			JoinedAt:    time.Now(),
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
		SELECT id, role, COALESCE(account_type, ''), COALESCE(moderator_scope, '')
		FROM users
		WHERE id = $1
	`, userID).Scan(&user.ID, &user.Role, &user.AccountType, &user.ModeratorScope)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// isAdminOrSiteModerator checks if a user has admin or site-wide moderator privileges.
// Platform moderators are users with role moderator.
func isAdminOrSiteModerator(user *models.User) bool {
	if user.Role == models.RoleAdmin || user.Role == models.RoleModerator || user.AccountType == models.AccountTypeAdmin {
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
		SELECT id, role, COALESCE(account_type, ''), COALESCE(moderator_scope, ''),
		       COALESCE(moderation_channels, '{}'::uuid[])
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

	// Channel admins manage their own channel
	member, err := h.communityRepo.GetMember(ctx, channelID, userID)
	if err == nil && member != nil && member.Role == models.CommunityRoleAdmin {
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
