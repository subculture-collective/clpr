package handlers

import (
	"context"
	"net/http"
	"strconv"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// WebhookSubscriptionHandler handles webhook subscription endpoints
type webhookSubscriptionService interface {
	CreateSubscription(context.Context, uuid.UUID, *models.CreateWebhookSubscriptionRequest) (*models.WebhookSubscription, error)
	GetSubscriptionsByUserID(context.Context, uuid.UUID) ([]*models.WebhookSubscription, error)
	GetSubscriptionByID(context.Context, uuid.UUID, uuid.UUID) (*models.WebhookSubscription, error)
	UpdateSubscription(context.Context, uuid.UUID, uuid.UUID, *models.UpdateWebhookSubscriptionRequest) error
	DeleteSubscription(context.Context, uuid.UUID, uuid.UUID) error
	GetDeliveriesBySubscriptionID(context.Context, uuid.UUID, uuid.UUID, int, int) ([]*models.WebhookDelivery, int, error)
	RegenerateSecret(context.Context, uuid.UUID, uuid.UUID) (string, error)
}

type WebhookSubscriptionHandler struct {
	webhookService webhookSubscriptionService
}

// NewWebhookSubscriptionHandler creates a new webhook subscription handler
func NewWebhookSubscriptionHandler(webhookService webhookSubscriptionService) *WebhookSubscriptionHandler {
	return &WebhookSubscriptionHandler{
		webhookService: webhookService,
	}
}

func webhookUserID(c *gin.Context) (uuid.UUID, bool) {
	value, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}
	userID, ok := value.(uuid.UUID)
	return userID, ok && userID != uuid.Nil
}

// CreateSubscription creates a new webhook subscription
// POST /webhooks
func (h *WebhookSubscriptionHandler) CreateSubscription(c *gin.Context) {
	// Get user ID from context
	userID, ok := webhookUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	var req models.CreateWebhookSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body: " + err.Error(),
		})
		return
	}

	subscription, err := h.webhookService.CreateSubscription(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create webhook subscription: " + err.Error(),
		})
		return
	}

	// Return the subscription with the secret (only time it's shown)
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    subscription,
		"secret":  subscription.Secret,
		"message": "Webhook subscription created. Save the secret safely - it won't be shown again.",
	})
}

// ListSubscriptions lists all webhook subscriptions for the authenticated user
// GET /webhooks
func (h *WebhookSubscriptionHandler) ListSubscriptions(c *gin.Context) {
	// Get user ID from context
	userID, ok := webhookUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	subscriptions, err := h.webhookService.GetSubscriptionsByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve webhook subscriptions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    subscriptions,
	})
}

// GetSubscription retrieves a specific webhook subscription
// GET /webhooks/:id
func (h *WebhookSubscriptionHandler) GetSubscription(c *gin.Context) {
	// Get user ID from context
	userID, ok := webhookUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	// Get subscription ID from URL
	subscriptionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid subscription ID",
		})
		return
	}

	subscription, err := h.webhookService.GetSubscriptionByID(c.Request.Context(), subscriptionID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Webhook subscription not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    subscription,
	})
}

// UpdateSubscription updates a webhook subscription
// PATCH /webhooks/:id
func (h *WebhookSubscriptionHandler) UpdateSubscription(c *gin.Context) {
	// Get user ID from context
	userID, ok := webhookUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	// Get subscription ID from URL
	subscriptionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid subscription ID",
		})
		return
	}

	var req models.UpdateWebhookSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body: " + err.Error(),
		})
		return
	}

	if err := h.webhookService.UpdateSubscription(c.Request.Context(), subscriptionID, userID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update webhook subscription: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Webhook subscription updated",
	})
}

// DeleteSubscription deletes a webhook subscription
// DELETE /webhooks/:id
func (h *WebhookSubscriptionHandler) DeleteSubscription(c *gin.Context) {
	// Get user ID from context
	userID, ok := webhookUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	// Get subscription ID from URL
	subscriptionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid subscription ID",
		})
		return
	}

	if err := h.webhookService.DeleteSubscription(c.Request.Context(), subscriptionID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete webhook subscription: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Webhook subscription deleted",
	})
}

// GetSubscriptionDeliveries retrieves delivery history for a webhook subscription
// GET /webhooks/:id/deliveries
func (h *WebhookSubscriptionHandler) GetSubscriptionDeliveries(c *gin.Context) {
	// Get user ID from context
	userID, ok := webhookUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	// Get subscription ID from URL
	subscriptionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid subscription ID",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	deliveries, total, err := h.webhookService.GetDeliveriesBySubscriptionID(c.Request.Context(), subscriptionID, userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve deliveries: " + err.Error(),
		})
		return
	}

	totalPages := (total + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    deliveries,
		"meta": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// GetSupportedEvents returns the list of supported webhook events
// GET /webhooks/events
func (h *WebhookSubscriptionHandler) GetSupportedEvents(c *gin.Context) {
	events := models.GetSupportedWebhookEvents()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    events,
	})
}

// RegenerateSecret regenerates the secret for a webhook subscription
// POST /webhooks/:id/regenerate-secret
func (h *WebhookSubscriptionHandler) RegenerateSecret(c *gin.Context) {
	userID, ok := webhookUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	subscriptionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	newSecret, err := h.webhookService.RegenerateSecret(c.Request.Context(), subscriptionID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to regenerate secret: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"secret":  newSecret,
		"message": "Secret regenerated. Save it safely - it won't be shown again.",
	})
}
