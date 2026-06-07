package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
)

// WebhookMonitoringHandler handles webhook monitoring endpoints
type WebhookMonitoringHandler struct {
	webhookRetryService    *services.WebhookRetryService
	outboundWebhookService *services.OutboundWebhookService
}

// NewWebhookMonitoringHandler creates a new webhook monitoring handler
func NewWebhookMonitoringHandler(
	webhookRetryService *services.WebhookRetryService,
	outboundWebhookService *services.OutboundWebhookService,
) *WebhookMonitoringHandler {
	return &WebhookMonitoringHandler{
		webhookRetryService:    webhookRetryService,
		outboundWebhookService: outboundWebhookService,
	}
}

// GetWebhookRetryStats returns webhook retry queue statistics
// @Summary Get webhook retry queue stats
// @Description Returns statistics about the webhook retry queue and dead-letter queue
// @Tags monitoring
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /health/webhooks [get]
func (h *WebhookMonitoringHandler) GetWebhookRetryStats(c *gin.Context) {
	// Get retry queue stats (includes queue sizes)
	stats, err := h.webhookRetryService.GetRetryQueueStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve webhook stats",
		})
		return
	}

	// Get additional metrics from outbound webhook service
	deliveryStats, err := h.outboundWebhookService.GetDeliveryStats(c.Request.Context())
	if err != nil {
		// Log error but don't fail the request
		c.JSON(http.StatusOK, gin.H{
			"status":   "healthy",
			"webhooks": stats,
		})
		return
	}

	// Combine stats
	combinedStats := make(map[string]interface{})
	for k, v := range stats {
		combinedStats[k] = v
	}
	for k, v := range deliveryStats {
		combinedStats[k] = v
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "healthy",
		"webhooks": combinedStats,
	})
}
