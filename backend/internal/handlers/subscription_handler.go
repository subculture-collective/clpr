package handlers

import (
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"github.com/gin-gonic/gin"
)

// SubscriptionHandler handles subscription-related HTTP requests
type SubscriptionHandler struct {
	subscriptionService *services.SubscriptionService
}

// NewSubscriptionHandler creates a new subscription handler
func NewSubscriptionHandler(subscriptionService *services.SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionService: subscriptionService,
	}
}

// GetSubscription retrieves the current user's subscription
// @Summary Get subscription
// @Description Retrieves the authenticated user's subscription information
// @Tags subscriptions
// @Produce json
// @Success 200 {object} models.Subscription
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/subscriptions/me [get]
func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	// Get authenticated user from context
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	currentUser, ok := user.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user information"})
		return
	}

	// Get subscription
	subscription, err := h.subscriptionService.GetSubscriptionByUserID(c.Request.Context(), currentUser.ID)
	if err != nil {
		if errors.Is(err, services.ErrSubscriptionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "No subscription found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve subscription"})
		return
	}

	c.JSON(http.StatusOK, models.PublicSubscription{
		Status: subscription.Status, Tier: subscription.Tier,
		CurrentPeriodStart: subscription.CurrentPeriodStart, CurrentPeriodEnd: subscription.CurrentPeriodEnd,
		CancelAtPeriodEnd: subscription.CancelAtPeriodEnd, CanceledAt: subscription.CanceledAt,
		TrialStart: subscription.TrialStart, TrialEnd: subscription.TrialEnd, GracePeriodEnd: subscription.GracePeriodEnd,
	})
}

// HandleWebhook handles Stripe webhook events
// @Summary Handle Stripe webhook
// @Description Processes Stripe webhook events for subscription lifecycle
// @Tags webhooks
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/webhooks/stripe [post]
func (h *SubscriptionHandler) HandleWebhook(c *gin.Context) {
	// Read the request body
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("Failed to read webhook body: %v", err)
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Webhook payload exceeds 1 MiB"})
		return
	}

	// Get the Stripe signature header
	signature := c.GetHeader("Stripe-Signature")
	if signature == "" || len(signature) > 8192 {
		log.Printf("Missing Stripe-Signature header")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing signature"})
		return
	}

	// Process webhook
	if err := h.subscriptionService.HandleWebhook(c.Request.Context(), payload, signature); err != nil {
		log.Printf("Failed to process webhook: %v", err)
		if errors.Is(err, services.ErrInvalidWebhookSignature) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signature"})
			return
		}
		if errors.Is(err, services.ErrSubscriptionsUnavailable) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Webhook processing is unavailable"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Webhook processing failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}

// CancelSubscription cancels the user's subscription
// @Summary Cancel subscription
// @Description Cancels the authenticated user's subscription (immediate or at period end)
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param request body models.CancelSubscriptionRequest true "Cancel subscription request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/subscriptions/cancel [post]
func (h *SubscriptionHandler) CancelSubscription(c *gin.Context) {
	// Get authenticated user from context
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	currentUser, ok := user.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user information"})
		return
	}

	// Parse request
	var req models.CancelSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Cancel subscription
	if err := h.subscriptionService.CancelSubscription(c.Request.Context(), currentUser, *req.Immediate); err != nil {
		log.Printf("Failed to cancel subscription: %v", err)
		if errors.Is(err, services.ErrSubscriptionsUnavailable) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Subscriptions are currently unavailable"})
			return
		}
		if errors.Is(err, services.ErrSubscriptionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "No subscription found"})
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": "Subscription cannot be canceled"})
		return
	}

	message := "Subscription will be canceled at the end of the billing period"
	if *req.Immediate {
		message = "Subscription canceled immediately"
	}

	c.JSON(http.StatusOK, gin.H{"message": message})
}

// GetInvoices retrieves the user's invoices
// @Summary Get invoices
// @Description Retrieves the authenticated user's subscription invoices
// @Tags subscriptions
// @Produce json
// @Param limit query int false "Number of invoices to retrieve (default: 10, max: 100)"
// @Success 200 {array} stripe.Invoice
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/subscriptions/invoices [get]
func (h *SubscriptionHandler) GetInvoices(c *gin.Context) {
	// Get authenticated user from context
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	currentUser, ok := user.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user information"})
		return
	}

	// Parse query parameters
	var limit int64 = 10
	if limitStr := c.Query("limit"); limitStr != "" {
		parsedLimit, err := strconv.ParseInt(limitStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
			return
		}
		if parsedLimit < 1 || parsedLimit > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Limit must be between 1 and 100"})
			return
		}
		limit = parsedLimit
	}

	// Get invoices
	invoices, err := h.subscriptionService.GetInvoices(c.Request.Context(), currentUser, limit)
	if err != nil {
		log.Printf("Failed to get invoices: %v", err)
		if err == services.ErrSubscriptionNotFound || err == services.ErrStripeCustomerNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "No subscription found"})
			return
		}
		if errors.Is(err, services.ErrSubscriptionsUnavailable) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Subscriptions are currently unavailable"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve invoices"})
		return
	}

	publicInvoices := make([]models.PublicInvoice, 0, len(invoices))
	for _, invoice := range invoices {
		if invoice == nil {
			continue
		}
		publicInvoices = append(publicInvoices, models.PublicInvoice{
			ID: invoice.ID, Status: string(invoice.Status), Currency: string(invoice.Currency),
			AmountDue: invoice.AmountDue, AmountPaid: invoice.AmountPaid,
			HostedInvoiceURL: invoice.HostedInvoiceURL, InvoicePDF: invoice.InvoicePDF,
			PeriodStart: invoice.PeriodStart, PeriodEnd: invoice.PeriodEnd, Created: invoice.Created,
		})
	}
	c.JSON(http.StatusOK, publicInvoices)
}
