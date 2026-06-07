package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

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
	days := 7
	if daysParam := c.Query("days"); daysParam != "" {
		if d, err := strconv.Atoi(daysParam); err == nil && d > 0 && d <= 30 {
			days = d
		}
	}

	// Get daily metrics
	dailyMetrics, err := h.emailMetricsService.GetDailyMetrics(c.Request.Context(), days)
	if err != nil {
		h.logger.Error("Failed to get dashboard metrics", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve metrics"})
		return
	}

	// Get current day metrics
	now := time.Now()
	todayStart := now.Truncate(24 * time.Hour)
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
		templateMetrics = make([]models.EmailMetricsSummary, 0) // Return empty array on error
	}

	// Get recent bounces
	recentBounces, err := h.emailLogRepo.GetRecentBounces(c.Request.Context(), 10)
	if err != nil {
		h.logger.Error("Failed to get recent bounces", err)
		recentBounces = make([]models.EmailLog, 0) // Return empty array on error
	}

	// Get unresolved alerts
	alerts, err := h.emailMetricsService.GetUnresolvedAlerts(c.Request.Context(), 10)
	if err != nil {
		h.logger.Error("Failed to get alerts", err)
		alerts = make([]models.EmailAlert, 0) // Return empty array on error
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
	// Parse time parameters
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	templateParam := c.Query("template")

	var startTime, endTime time.Time
	var err error

	if startDateStr != "" {
		startTime, err = time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format"})
			return
		}
	} else {
		startTime = time.Now().AddDate(0, 0, -7).Truncate(24 * time.Hour)
	}

	if endDateStr != "" {
		endTime, err = time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format"})
			return
		}
	} else {
		endTime = time.Now()
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
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	limitStr := c.Query("limit")

	var startTime, endTime time.Time
	var err error

	if startDateStr != "" {
		startTime, err = time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format"})
			return
		}
	} else {
		startTime = time.Now().AddDate(0, 0, -7).Truncate(24 * time.Hour)
	}

	if endDateStr != "" {
		endTime, err = time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format"})
			return
		}
	} else {
		endTime = time.Now()
	}

	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
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
		filters["status"] = status
	}
	if template := c.Query("template"); template != "" {
		filters["template"] = template
	}
	if recipient := c.Query("recipient"); recipient != "" {
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
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 200 {
			limit = l
		}
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
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	alertIDStr := c.Param("id")
	alertID, err := uuid.Parse(alertIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	if err := h.emailLogRepo.AcknowledgeAlert(c.Request.Context(), alertID, userID.(uuid.UUID)); err != nil {
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
		h.logger.Error("Failed to resolve alert", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve alert"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alert resolved"})
}
