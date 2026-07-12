package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const maxEmailMetricsRange = 366 * 24 * time.Hour

func parseEmailMetricsRange(c *gin.Context, defaultDays int) (time.Time, time.Time, bool) {
	now := time.Now().UTC()
	startTime := now.AddDate(0, 0, -defaultDays)
	endTime := now
	var err error
	if value := c.Query("start_date"); value != "" {
		startTime, err = time.Parse(time.RFC3339, value)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format"})
			return time.Time{}, time.Time{}, false
		}
	}
	if value := c.Query("end_date"); value != "" {
		endTime, err = time.Parse(time.RFC3339, value)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format"})
			return time.Time{}, time.Time{}, false
		}
	}
	if endTime.Before(startTime) || endTime.Sub(startTime) > maxEmailMetricsRange {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Date range must be ordered and no longer than 366 days"})
		return time.Time{}, time.Time{}, false
	}
	return startTime, endTime, true
}

func parseBoundedEmailInt(c *gin.Context, name string, defaultValue, minimum, maximum int) (int, bool) {
	value := c.Query(name)
	if value == "" {
		return defaultValue, true
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < minimum || parsed > maximum {
		c.JSON(http.StatusBadRequest, gin.H{"error": name + " is out of range"})
		return 0, false
	}
	return parsed, true
}

// EmailMetricsHandler handles email metrics and monitoring endpoints
type EmailMetricsHandler struct {
	emailMetricsService *services.EmailMetricsService
	emailLogRepo        *repository.EmailLogRepository
	logger              *utils.StructuredLogger
}

// NewEmailMetricsHandler creates a new email metrics handler
func NewEmailMetricsHandler(emailMetricsService *services.EmailMetricsService, emailLogRepo *repository.EmailLogRepository) *EmailMetricsHandler {
	return &EmailMetricsHandler{
		emailMetricsService: emailMetricsService,
		emailLogRepo:        emailLogRepo,
		logger:              utils.GetLogger(),
	}
}

// GetDashboardMetrics returns email metrics for the dashboard
// @Summary Get email dashboard metrics
// @Description Returns email delivery metrics for the dashboard including 7-day trends
// @Tags email-metrics
// @Produce json
// @Param days query int false "Number of days to include (default: 7)"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/v1/email/metrics/dashboard [get]
func (h *EmailMetricsHandler) GetDashboardMetrics(c *gin.Context) {
	days, ok := parseBoundedEmailInt(c, "days", 7, 1, 30)
	if !ok {
		return
	}

	// Get daily metrics
	dailyMetrics, err := h.emailMetricsService.GetDailyMetrics(c.Request.Context(), days)
	if err != nil {
		h.logger.Error("Failed to get dashboard metrics", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve metrics"})
		return
	}

	// Get current day metrics
	now := time.Now().UTC()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	currentMetrics, err := h.emailMetricsService.GetMetricsForPeriod(c.Request.Context(), todayStart, now, nil)
	if err != nil {
		h.logger.Error("Failed to get current metrics", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve metrics"})
		return
	}

	// Get metrics by template (top 10)
	weekAgo := now.AddDate(0, 0, -7)
	templateMetrics, err := h.emailMetricsService.GetMetricsByTemplate(c.Request.Context(), weekAgo, now, 10)
	if err != nil {
		h.logger.Error("Failed to get template metrics", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve metrics"})
		return
	}

	// Get recent bounces
	recentBounces, err := h.emailLogRepo.GetRecentBounces(c.Request.Context(), 10)
	if err != nil {
		h.logger.Error("Failed to get recent bounces", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve metrics"})
		return
	}

	// Get unresolved alerts
	alerts, err := h.emailMetricsService.GetUnresolvedAlerts(c.Request.Context(), 10)
	if err != nil {
		h.logger.Error("Failed to get alerts", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve metrics"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"daily_metrics":    dailyMetrics,
		"current_metrics":  currentMetrics,
		"template_metrics": templateMetrics,
		"recent_bounces":   recentBounces,
		"alerts":           alerts,
	})
}

// GetMetrics returns email metrics for a specific period
// @Summary Get email metrics
// @Description Returns email metrics for a specific time period
// @Tags email-metrics
// @Produce json
// @Param start_date query string false "Start date (RFC3339 format)"
// @Param end_date query string false "End date (RFC3339 format)"
// @Param template query string false "Filter by template"
// @Success 200 {object} models.EmailMetricsSummary
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/email/metrics [get]
func (h *EmailMetricsHandler) GetMetrics(c *gin.Context) {
	templateParam := c.Query("template")
	if len(templateParam) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "template is too long"})
		return
	}
	startTime, endTime, ok := parseEmailMetricsRange(c, 7)
	if !ok {
		return
	}

	var template *string
	if templateParam != "" {
		template = &templateParam
	}

	metrics, err := h.emailMetricsService.GetMetricsForPeriod(c.Request.Context(), startTime, endTime, template)
	if err != nil {
		h.logger.Error("Failed to get metrics", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve metrics"})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetTemplateMetrics returns metrics grouped by template
// @Summary Get metrics by template
// @Description Returns email metrics grouped by template
// @Tags email-metrics
// @Produce json
// @Param start_date query string false "Start date (RFC3339 format)"
// @Param end_date query string false "End date (RFC3339 format)"
// @Param limit query int false "Maximum number of templates to return (default: 10)"
// @Success 200 {array} models.EmailMetricsSummary
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/email/metrics/templates [get]
func (h *EmailMetricsHandler) GetTemplateMetrics(c *gin.Context) {
	startTime, endTime, ok := parseEmailMetricsRange(c, 7)
	if !ok {
		return
	}
	limit, ok := parseBoundedEmailInt(c, "limit", 10, 1, 100)
	if !ok {
		return
	}

	metrics, err := h.emailMetricsService.GetMetricsByTemplate(c.Request.Context(), startTime, endTime, limit)
	if err != nil {
		h.logger.Error("Failed to get template metrics", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve metrics"})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// SearchEmailLogs searches email logs with filters
// @Summary Search email logs
// @Description Searches email logs with various filters
// @Tags email-metrics
// @Produce json
// @Param status query string false "Filter by status"
// @Param template query string false "Filter by template"
// @Param recipient query string false "Filter by recipient email"
// @Param start_date query string false "Start date (RFC3339 format)"
// @Param end_date query string false "End date (RFC3339 format)"
// @Param limit query int false "Limit results (default: 50)"
// @Param offset query int false "Offset for pagination (default: 0)"
// @Success 200 {array} models.EmailLog
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/email/logs [get]
func (h *EmailMetricsHandler) SearchEmailLogs(c *gin.Context) {
	filters := make(map[string]interface{})

	if status := c.Query("status"); status != "" {
		validStatuses := map[string]bool{
			models.EmailLogStatusSent: true, models.EmailLogStatusDelivered: true,
			models.EmailLogStatusBounce: true, models.EmailLogStatusDropped: true,
			models.EmailLogStatusOpen: true, models.EmailLogStatusClick: true,
			models.EmailLogStatusSpamReport: true, models.EmailLogStatusUnsubscribe: true,
			models.EmailLogStatusDeferred: true, models.EmailLogStatusProcessed: true,
		}
		if !validStatuses[status] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
			return
		}
		filters["status"] = status
	}
	if template := c.Query("template"); template != "" {
		if len(template) > 200 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "template is too long"})
			return
		}
		filters["template"] = template
	}
	if recipient := c.Query("recipient"); recipient != "" {
		if len(recipient) > 254 || strings.ContainsAny(recipient, "\r\n") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recipient filter"})
			return
		}
		filters["recipient"] = recipient
	}

	if startDate := c.Query("start_date"); startDate != "" {
		t, err := time.Parse(time.RFC3339, startDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format"})
			return
		}
		filters["start_date"] = t
	}

	if endDate := c.Query("end_date"); endDate != "" {
		t, err := time.Parse(time.RFC3339, endDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format"})
			return
		}
		filters["end_date"] = t
	}

	if start, startOK := filters["start_date"].(time.Time); startOK {
		if end, endOK := filters["end_date"].(time.Time); endOK && (end.Before(start) || end.Sub(start) > maxEmailMetricsRange) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Date range must be ordered and no longer than 366 days"})
			return
		}
	}
	limit, ok := parseBoundedEmailInt(c, "limit", 50, 1, 200)
	if !ok {
		return
	}
	offset, ok := parseBoundedEmailInt(c, "offset", 0, 0, 100000)
	if !ok {
		return
	}

	logs, err := h.emailLogRepo.SearchEmailLogs(c.Request.Context(), filters, limit, offset)
	if err != nil {
		h.logger.Error("Failed to search email logs", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search logs"})
		return
	}

	c.JSON(http.StatusOK, logs)
}

// GetUserEmailLogs returns email logs for a specific user
// @Summary Get user email logs
// @Description Returns email logs for the authenticated user
// @Tags email-metrics
// @Produce json
// @Param limit query int false "Limit results (default: 50)"
// @Param offset query int false "Offset for pagination (default: 0)"
// @Success 200 {array} models.EmailLog
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/email/logs/me [get]
func (h *EmailMetricsHandler) GetUserEmailLogs(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 200 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	logs, err := h.emailLogRepo.GetEmailLogsByUserID(c.Request.Context(), userID.(uuid.UUID), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get user email logs", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve logs"})
		return
	}

	c.JSON(http.StatusOK, logs)
}

// GetAlerts returns email alerts
// @Summary Get email alerts
// @Description Returns unresolved email alerts
// @Tags email-metrics
// @Produce json
// @Param limit query int false "Limit results (default: 50)"
// @Success 200 {array} models.EmailAlert
// @Failure 500 {object} map[string]string
// @Router /api/v1/email/alerts [get]
func (h *EmailMetricsHandler) GetAlerts(c *gin.Context) {
	limit, ok := parseBoundedEmailInt(c, "limit", 50, 1, 200)
	if !ok {
		return
	}

	alerts, err := h.emailMetricsService.GetUnresolvedAlerts(c.Request.Context(), limit)
	if err != nil {
		h.logger.Error("Failed to get alerts", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve alerts"})
		return
	}

	c.JSON(http.StatusOK, alerts)
}

// AcknowledgeAlert acknowledges an email alert
// @Summary Acknowledge an email alert
// @Description Acknowledges an unresolved email alert
// @Tags email-metrics
// @Produce json
// @Param id path string true "Alert ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/email/alerts/:id/acknowledge [post]
func (h *EmailMetricsHandler) AcknowledgeAlert(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, ok := userIDValue.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	alertIDStr := c.Param("id")
	alertID, err := uuid.Parse(alertIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	if err := h.emailLogRepo.AcknowledgeAlert(c.Request.Context(), alertID, userID); err != nil {
		if errors.Is(err, repository.ErrEmailAlertNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Alert not found or already acknowledged"})
			return
		}
		h.logger.Error("Failed to acknowledge alert", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to acknowledge alert"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alert acknowledged"})
}

// ResolveAlert resolves an email alert
// @Summary Resolve an email alert
// @Description Resolves an unresolved email alert
// @Tags email-metrics
// @Produce json
// @Param id path string true "Alert ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/email/alerts/:id/resolve [post]
func (h *EmailMetricsHandler) ResolveAlert(c *gin.Context) {
	alertIDStr := c.Param("id")
	alertID, err := uuid.Parse(alertIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	if err := h.emailLogRepo.ResolveAlert(c.Request.Context(), alertID); err != nil {
		if errors.Is(err, repository.ErrEmailAlertNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Alert not found or already resolved"})
			return
		}
		h.logger.Error("Failed to resolve alert", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve alert"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alert resolved"})
}
