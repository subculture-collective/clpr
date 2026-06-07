package handlers

import (
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
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

// CreateCheckoutSession creates a Stripe Checkout session
// @Summary Create checkout session
// @Description Creates a Stripe Checkout session for subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param request body models.CreateCheckoutSessionRequest true "Checkout session request"
// @Success 200 {object} models.CreateCheckoutSessionResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/subscriptions/checkout [post]
func (h *SubscriptionHandler) CreateCheckoutSession(c *gin.Context) {
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

	// Check if user has email (required for Stripe)
	if currentUser.Email == nil || *currentUser.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required for subscriptions"})
		return
	}

	// Parse request
	var req models.CreateCheckoutSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Create checkout session with optional coupon code
	response, err := h.subscriptionService.CreateCheckoutSession(c.Request.Context(), currentUser, req.PriceID, req.CouponCode)
	if err != nil {
		log.Printf("Failed to create checkout session: %v", err)
		if err == services.ErrInvalidPriceID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid price ID"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create checkout session"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CreatePortalSession creates a Stripe Customer Portal session
// @Summary Create portal session
// @Description Creates a Stripe Customer Portal session for managing subscription
// @Tags subscriptions
// @Produce json
// @Success 200 {object} models.CreatePortalSessionResponse
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/subscriptions/portal [post]
func (h *SubscriptionHandler) CreatePortalSession(c *gin.Context) {
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

	// Create portal session
	response, err := h.subscriptionService.CreatePortalSession(c.Request.Context(), currentUser)
	if err != nil {
		log.Printf("Failed to create portal session: %v", err)
		if err == services.ErrSubscriptionNotFound || err == services.ErrStripeCustomerNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "No subscription found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create portal session"})
		return
	}

	c.JSON(http.StatusOK, response)
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
		c.JSON(http.StatusNotFound, gin.H{"error": "No subscription found"})
		return
	}

	c.JSON(http.StatusOK, subscription)
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
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("Failed to read webhook body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Get the Stripe signature header
	signature := c.GetHeader("Stripe-Signature")
	if signature == "" {
		log.Printf("Missing Stripe-Signature header")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing signature"})
		return
	}

	// Process webhook
	if err := h.subscriptionService.HandleWebhook(c.Request.Context(), payload, signature); err != nil {
		log.Printf("Failed to process webhook: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}

// ChangeSubscriptionPlan changes the user's subscription plan
// @Summary Change subscription plan
// @Description Changes the user's subscription plan (e.g., monthly to yearly) with proration
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param request body models.ChangeSubscriptionPlanRequest true "New plan request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/subscriptions/change-plan [post]
func (h *SubscriptionHandler) ChangeSubscriptionPlan(c *gin.Context) {
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
	var req models.ChangeSubscriptionPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Change subscription plan
	if err := h.subscriptionService.ChangeSubscriptionPlan(c.Request.Context(), currentUser, req.PriceID); err != nil {
		log.Printf("Failed to change subscription plan: %v", err)
		if err == services.ErrInvalidPriceID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid price ID"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscription plan changed successfully"})
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
	if err := h.subscriptionService.CancelSubscription(c.Request.Context(), currentUser, req.Immediate); err != nil {
		log.Printf("Failed to cancel subscription: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message := "Subscription will be canceled at the end of the billing period"
	if req.Immediate {
		message = "Subscription canceled immediately"
	}

	c.JSON(http.StatusOK, gin.H{"message": message})
}

// ReactivateSubscription reactivates a subscription scheduled for cancellation
// @Summary Reactivate subscription
// @Description Reactivates a subscription that was set to cancel at period end
// @Tags subscriptions
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/subscriptions/reactivate [post]
func (h *SubscriptionHandler) ReactivateSubscription(c *gin.Context) {
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

	// Reactivate subscription
	if err := h.subscriptionService.ReactivateSubscription(c.Request.Context(), currentUser); err != nil {
		log.Printf("Failed to reactivate subscription: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscription reactivated successfully"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve invoices"})
		return
	}

	c.JSON(http.StatusOK, invoices)
}
