package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type outboundWebhookStatsService interface {
	GetDeliveryStats(context.Context) (map[string]interface{}, error)
}

// WebhookMonitoringHandler handles webhook monitoring endpoints
type WebhookMonitoringHandler struct {
	outboundWebhookService outboundWebhookStatsService
}

// NewWebhookMonitoringHandler creates a new webhook monitoring handler
func NewWebhookMonitoringHandler(
	outboundWebhookService outboundWebhookStatsService,
) *WebhookMonitoringHandler {
	return &WebhookMonitoringHandler{
		outboundWebhookService: outboundWebhookService,
	}
}

// GetWebhookStats returns outbound webhook delivery statistics
// @Summary Get outbound webhook delivery stats
// @Description Returns statistics about outbound webhook subscriptions and deliveries
// @Tags monitoring
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 503 {object} map[string]string
// @Router /internal/operations/webhooks [get]
func (h *WebhookMonitoringHandler) GetWebhookStats(c *gin.Context) {
	// Get additional metrics from outbound webhook service
	deliveryStats, err := h.outboundWebhookService.GetDeliveryStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":                 "degraded",
			"unavailable_components": []string{"delivery_stats"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "healthy",
		"webhooks": deliveryStats,
	})
}
