package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

// WebhookDLQHandler handles webhook dead-letter queue management endpoints
type WebhookDLQHandler struct {
	webhookService *services.OutboundWebhookService
}

// NewWebhookDLQHandler creates a new webhook DLQ handler
func NewWebhookDLQHandler(webhookService *services.OutboundWebhookService) *WebhookDLQHandler {
	return &WebhookDLQHandler{
		webhookService: webhookService,
	}
}

// GetDeadLetterQueue returns items from the webhook dead-letter queue
// @Summary Get webhook DLQ items
// @Description Returns paginated list of failed webhook deliveries in the dead-letter queue
// @Tags admin,webhooks
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/webhooks/dlq [get]
func (h *WebhookDLQHandler) GetDeadLetterQueue(c *gin.Context) {
	// Parse pagination parameters
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l <= 0 || l > 100 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid 'limit' query parameter; must be an integer between 1 and 100",
			})
			return
		}
		limit = l
	}

	// Get DLQ items
	items, total, err := h.webhookService.GetDeadLetterQueueItems(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve DLQ items",
		})
		return
	}

	totalPages := (total + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"pagination": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// ReplayDeadLetterQueueItem attempts to replay a failed webhook delivery
// @Summary Replay a DLQ item
// @Description Attempts to redeliver a failed webhook from the dead-letter queue
// @Tags admin,webhooks
// @Produce json
// @Param id path string true "DLQ item ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/webhooks/dlq/{id}/replay [post]
func (h *WebhookDLQHandler) ReplayDeadLetterQueueItem(c *gin.Context) {
	// Parse DLQ item ID
	idStr := c.Param("id")
	dlqID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid DLQ item ID",
		})
		return
	}

	// Replay the item
	if err := h.webhookService.ReplayDeadLetterQueueItem(c.Request.Context(), dlqID); err != nil {
		// Log detailed error but return generic message to avoid leaking internal details
		utils.Error("Failed to replay webhook DLQ item", err, map[string]interface{}{
			"component": "webhook_dlq",
			"dlq_id":    dlqID,
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to replay webhook delivery",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Webhook replayed successfully",
	})
}

// DeleteDeadLetterQueueItem deletes a DLQ item
// @Summary Delete a DLQ item
// @Description Permanently deletes a failed webhook from the dead-letter queue
// @Tags admin,webhooks
// @Produce json
// @Param id path string true "DLQ item ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/webhooks/dlq/{id} [delete]
func (h *WebhookDLQHandler) DeleteDeadLetterQueueItem(c *gin.Context) {
	// Parse DLQ item ID
	idStr := c.Param("id")
	dlqID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid DLQ item ID",
		})
		return
	}

	// Delete the item
	if err := h.webhookService.DeleteDeadLetterQueueItem(c.Request.Context(), dlqID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete DLQ item",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "DLQ item deleted successfully",
	})
}
