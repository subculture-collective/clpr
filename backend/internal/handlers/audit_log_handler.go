package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
)

// AuditLogResponse represents the API response format for audit log entries
type AuditLogResponse struct {
	ID         string                 `json:"id"`
	Action     string                 `json:"action"`
	EntityType string                 `json:"entityType"`
	Actor      map[string]interface{} `json:"actor"`
	Target     map[string]interface{} `json:"target"`
	Reason     string                 `json:"reason"`
	CreatedAt  string                 `json:"createdAt"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// transformAuditLog converts a ModerationAuditLogWithUser to the API response format
// Note: Target username is not populated as it would require additional database joins.
// The entity_id is provided for clients to fetch user details if needed.
func transformAuditLog(log *models.ModerationAuditLogWithUser) AuditLogResponse {
	var reason string
	if log.Reason != nil {
		reason = *log.Reason
	}

	actor := map[string]interface{}{
		"id":       log.ModeratorID.String(),
		"username": "",
	}
	if log.Moderator != nil {
		actor["username"] = log.Moderator.Username
	}

	// Note: Target username is not included as it would require additional joins.
	// Clients can fetch user details using the entity_id if entity_type is "user"
	target := map[string]interface{}{
		"id":       log.EntityID.String(),
		"username": "",
	}

	metadata := log.Metadata
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	return AuditLogResponse{
		ID:         log.ID.String(),
		Action:     log.Action,
		EntityType: log.EntityType,
		Actor:      actor,
		Target:     target,
		Reason:     reason,
		CreatedAt:  log.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Metadata:   metadata,
	}
}

// AuditLogHandler handles audit log operations
type AuditLogHandler struct {
	auditLogService *services.AuditLogService
}

// NewAuditLogHandler creates a new AuditLogHandler
func NewAuditLogHandler(auditLogService *services.AuditLogService) *AuditLogHandler {
	return &AuditLogHandler{
		auditLogService: auditLogService,
	}
}

// ListAuditLogs retrieves audit logs with filters
// GET /admin/audit-logs
// Supports filters: moderator_id, action, entity_type, entity_id, channel_id, start_date (RFC3339), end_date (RFC3339), search
func (h *AuditLogHandler) ListAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}

	// Parse filters
	filters, err := services.ParseAuditLogFilters(
		c.Query("moderator_id"),
		c.Query("action"),
		c.Query("entity_type"),
		c.Query("entity_id"),
		c.Query("channel_id"),
		c.Query("start_date"),
		c.Query("end_date"),
		c.Query("search"),
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	logs, total, err := h.auditLogService.GetAuditLogs(c.Request.Context(), filters, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve audit logs",
		})
		return
	}

	totalPages := (total + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    logs,
		"meta": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// ExportAuditLogs exports audit logs to CSV
// GET /admin/audit-logs/export
// Supports same filters as ListAuditLogs
func (h *AuditLogHandler) ExportAuditLogs(c *gin.Context) {
	// Parse filters
	filters, err := services.ParseAuditLogFilters(
		c.Query("moderator_id"),
		c.Query("action"),
		c.Query("entity_type"),
		c.Query("entity_id"),
		c.Query("channel_id"),
		c.Query("start_date"),
		c.Query("end_date"),
		c.Query("search"),
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Set response headers for CSV download
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=audit_logs.csv")

	// Export to CSV
	if err := h.auditLogService.ExportAuditLogsCSV(c.Request.Context(), filters, c.Writer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to export audit logs",
		})
		return
	}
}

// ListModerationAuditLogs retrieves moderation audit logs with filters and offset-based pagination
// GET /api/v1/moderation/audit-logs
// Supports filters: action, actor (moderator_id), target (entity_id), channel, startDate, endDate, limit, offset, search
// Note: For optimal results, offset should be a multiple of limit due to underlying page-based repository implementation
func (h *AuditLogHandler) ListModerationAuditLogs(c *gin.Context) {
	// Get pagination params using offset instead of page
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit < 1 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	// Parse filters - use "actor" and "target" as per requirement
	filters, err := services.ParseAuditLogFilters(
		c.Query("actor"),     // moderator_id in filter
		c.Query("action"),    // action type
		"",                   // entity_type - not specified in requirements
		c.Query("target"),    // entity_id in filter
		c.Query("channel"),   // channel_id
		c.Query("startDate"), // start_date
		c.Query("endDate"),   // end_date
		c.Query("search"),    // search term for reason field
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// The repository uses page-based pagination internally, so we calculate the page
	// and adjust offset to align with page boundaries. This means the actual offset
	// returned may differ from the requested offset if it's not a multiple of limit.
	page := (offset / limit) + 1
	actualOffset := (page - 1) * limit

	logs, total, err := h.auditLogService.GetAuditLogs(c.Request.Context(), filters, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve audit logs",
		})
		return
	}

	// Transform logs to match the required response format
	response := make([]AuditLogResponse, 0, len(logs))
	for _, log := range logs {
		response = append(response, transformAuditLog(log))
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":   response,
		"total":  total,
		"limit":  limit,
		"offset": actualOffset,
	})
}

// GetModerationAuditLog retrieves a single audit log entry by ID
// GET /api/v1/moderation/audit-logs/:id
func (h *AuditLogHandler) GetModerationAuditLog(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid audit log ID",
		})
		return
	}

	log, err := h.auditLogService.GetAuditLogByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Audit log not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve audit log",
		})
		return
	}

	// Transform to match response format
	response := transformAuditLog(log)
	c.JSON(http.StatusOK, response)
}

// ExportModerationAuditLogs exports moderation audit logs to CSV
// GET /api/v1/moderation/audit-logs/export
func (h *AuditLogHandler) ExportModerationAuditLogs(c *gin.Context) {
	// Parse filters using moderation-specific param names
	filters, err := services.ParseAuditLogFilters(
		c.Query("actor"),
		c.Query("action"),
		"",
		c.Query("target"),
		c.Query("channel"),
		c.Query("startDate"),
		c.Query("endDate"),
		c.Query("search"),
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Set response headers for CSV download before writing
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=moderation_audit_logs.csv")

	// Export to CSV
	// Note: If export fails after headers are sent, we can only set the status code.
	// The client may receive partial CSV data followed by an incomplete response.
	if err := h.auditLogService.ExportAuditLogsCSV(c.Request.Context(), filters, c.Writer); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
}
