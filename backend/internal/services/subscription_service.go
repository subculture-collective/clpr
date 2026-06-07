package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v81"
	portalsession "github.com/stripe/stripe-go/v81/billingportal/session"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/customer"
	"github.com/stripe/stripe-go/v81/invoice"
	"github.com/stripe/stripe-go/v81/subscription"
	"github.com/stripe/stripe-go/v81/webhook"
	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

var (
	// ErrSubscriptionNotFound indicates the subscription was not found
	ErrSubscriptionNotFound = errors.New("subscription not found")
	// ErrInvalidPriceID indicates an invalid price ID was provided
	ErrInvalidPriceID = errors.New("invalid price ID")
	// ErrStripeCustomerNotFound indicates the Stripe customer was not found
	ErrStripeCustomerNotFound = errors.New("stripe customer not found")
)

const webhookLogComponent = "stripe_webhook"

func logWebhookInfo(message string, fields map[string]interface{}) {
	if fields == nil {
		fields = map[string]interface{}{}
	}
	fields["component"] = webhookLogComponent
	utils.Info(message, fields)
}

func logWebhookWarn(message string, fields map[string]interface{}) {
	if fields == nil {
		fields = map[string]interface{}{}
	}
	fields["component"] = webhookLogComponent
	utils.Warn(message, fields)
}

func logWebhookError(message string, err error, fields map[string]interface{}) {
	if fields == nil {
		fields = map[string]interface{}{}
	}
	fields["component"] = webhookLogComponent
	utils.Error(message, err, fields)
}

// SubscriptionService handles subscription business logic
type SubscriptionService struct {
	repo           repository.SubscriptionRepositoryInterface
	userRepo       repository.UserRepositoryInterface
	webhookRepo    repository.WebhookRepositoryInterface
	cfg            *config.Config
	auditLogSvc    *AuditLogService
	dunningService *DunningService
	emailService   *EmailService
}

// NewSubscriptionService creates a new subscription service
func NewSubscriptionService(
	repo repository.SubscriptionRepositoryInterface,
	userRepo repository.UserRepositoryInterface,
	webhookRepo repository.WebhookRepositoryInterface,
	cfg *config.Config,
	auditLogSvc *AuditLogService,
	dunningService *DunningService,
	emailService *EmailService,
) *SubscriptionService {
	// Initialize Stripe with secret key
	if cfg != nil && cfg.Stripe.SecretKey != "" {
		stripe.Key = cfg.Stripe.SecretKey
	}

	return &SubscriptionService{
		repo:           repo,
		userRepo:       userRepo,
		webhookRepo:    webhookRepo,
		cfg:            cfg,
		auditLogSvc:    auditLogSvc,
		dunningService: dunningService,
		emailService:   emailService,
	}
}

// GetRepository exposes the underlying subscription repository for tests and auxiliary services
// This maintains clear separation while allowing integration tests to inspect persisted state.
func (s *SubscriptionService) GetRepository() repository.SubscriptionRepositoryInterface {
	return s.repo
}

// GetOrCreateCustomer gets or creates a Stripe customer for the user
func (s *SubscriptionService) GetOrCreateCustomer(ctx context.Context, user *models.User) (string, error) {
	// Check if user already has a subscription with customer ID
	sub, err := s.repo.GetByUserID(ctx, user.ID)
	if err == nil && sub.StripeCustomerID != "" {
		return sub.StripeCustomerID, nil
	}

	// Create new Stripe customer
	params := &stripe.CustomerParams{
		Email: stripe.String(*user.Email),
		Metadata: map[string]string{
			"user_id":  user.ID.String(),
			"username": user.Username,
		},
	}

	if user.DisplayName != "" {
		params.Name = stripe.String(user.DisplayName)
	}

	cust, err := customer.New(params)
	if err != nil {
		return "", fmt.Errorf("failed to create Stripe customer: %w", err)
	}

	// Create or update subscription record with customer ID
	if sub == nil {
		sub = &models.Subscription{
			UserID:           user.ID,
			StripeCustomerID: cust.ID,
			Status:           "inactive",
			Tier:             "free",
		}
		if err := s.repo.Create(ctx, sub); err != nil {
			return "", fmt.Errorf("failed to create subscription record: %w", err)
		}
	}

	// Log audit event
	if s.auditLogSvc != nil {
		_ = s.auditLogSvc.LogSubscriptionEvent(ctx, user.ID, "customer_created", map[string]interface{}{
			"stripe_customer_id": cust.ID,
		})
	}

	return cust.ID, nil
}

// CreateCheckoutSession creates a Stripe Checkout session for subscription
func (s *SubscriptionService) CreateCheckoutSession(ctx context.Context, user *models.User, priceID string, couponCode *string) (*models.CreateCheckoutSessionResponse, error) {
	// If Stripe is not configured or premium feature flag is off, return a mock session
	if s.cfg.Stripe.SecretKey == "" || !s.cfg.FeatureFlags.PremiumSubscriptions {
		mockURL := s.cfg.Stripe.SuccessURL
		if mockURL == "" {
			mockURL = "http://localhost:5173/subscription/success"
		}
		return &models.CreateCheckoutSessionResponse{
			SessionID:  "cs_test_mock",
			SessionURL: fmt.Sprintf("%s?session_id=cs_test_mock", mockURL),
		}, nil
	}

	// Validate price ID
	if priceID != s.cfg.Stripe.ProMonthlyPriceID && priceID != s.cfg.Stripe.ProYearlyPriceID {
		return nil, ErrInvalidPriceID
	}

	// Get or create Stripe customer
	customerID, err := s.GetOrCreateCustomer(ctx, user)
	if err != nil {
		return nil, err
	}

	// Create checkout session with idempotency key
	idempotencyKey := fmt.Sprintf("checkout_%s_%s", user.ID.String(), priceID)

	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerID),
		Mode:     stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(s.cfg.Stripe.SuccessURL + "?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String(s.cfg.Stripe.CancelURL),
		Metadata: map[string]string{
			"user_id": user.ID.String(),
		},
		// Enable promotion codes by default
		AllowPromotionCodes: stripe.Bool(true),
	}

	// Enable Stripe Tax for automatic tax calculation if configured
	if s.cfg.Stripe.TaxEnabled {
		params.AutomaticTax = &stripe.CheckoutSessionAutomaticTaxParams{
			Enabled: stripe.Bool(true),
		}
		// Require billing address collection for tax calculation
		params.BillingAddressCollection = stripe.String("required")
	}

	// Apply coupon code if provided
	if couponCode != nil && *couponCode != "" {
		params.Discounts = []*stripe.CheckoutSessionDiscountParams{
			{
				Coupon: stripe.String(*couponCode),
			},
		}
	}

	params.SetIdempotencyKey(idempotencyKey)

	sess, err := session.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create checkout session: %w", err)
	}

	// Log audit event
	metadata := map[string]interface{}{
		"session_id": sess.ID,
		"price_id":   priceID,
	}
	if couponCode != nil && *couponCode != "" {
		metadata["coupon_code"] = *couponCode
	}
	if s.auditLogSvc != nil {
		_ = s.auditLogSvc.LogSubscriptionEvent(ctx, user.ID, "checkout_session_created", metadata)
	}

	return &models.CreateCheckoutSessionResponse{
		SessionID:  sess.ID,
		SessionURL: sess.URL,
	}, nil
}

// CreatePortalSession creates a Stripe Customer Portal session
func (s *SubscriptionService) CreatePortalSession(ctx context.Context, user *models.User) (*models.CreatePortalSessionResponse, error) {
	// If Stripe is not configured or premium feature flag is off, return a mock portal URL
	if s.cfg.Stripe.SecretKey == "" || !s.cfg.FeatureFlags.PremiumSubscriptions {
		mockURL := s.cfg.Stripe.SuccessURL
		if mockURL == "" {
			mockURL = "http://localhost:5173/subscription"
		}
		return &models.CreatePortalSessionResponse{PortalURL: mockURL}, nil
	}

	// Get subscription to find customer ID
	sub, err := s.repo.GetByUserID(ctx, user.ID)
	if err != nil {
		return nil, ErrSubscriptionNotFound
	}

	if sub.StripeCustomerID == "" {
		return nil, ErrStripeCustomerNotFound
	}

	// Create portal session
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(sub.StripeCustomerID),
		ReturnURL: stripe.String(s.cfg.Stripe.SuccessURL),
	}

	sess, err := portalsession.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create portal session: %w", err)
	}

	// Log audit event
	if s.auditLogSvc != nil {
		_ = s.auditLogSvc.LogSubscriptionEvent(ctx, user.ID, "portal_session_created", map[string]interface{}{
			"session_id": sess.ID,
		})
	}

	return &models.CreatePortalSessionResponse{
		PortalURL: sess.URL,
	}, nil
}

// GetSubscriptionByUserID retrieves a user's subscription
func (s *SubscriptionService) GetSubscriptionByUserID(ctx context.Context, userID uuid.UUID) (*models.Subscription, error) {
	return s.repo.GetByUserID(ctx, userID)
}

// HandleWebhook processes Stripe webhook events
func (s *SubscriptionService) HandleWebhook(ctx context.Context, payload []byte, signature string) error {
	// Verify webhook signature against all configured secrets
	event, err := s.verifyWebhookSignature(payload, signature)
	if err != nil {
		logWebhookError("Webhook signature verification failed", err, map[string]interface{}{
			"event_type": "unknown",
		})
		return fmt.Errorf("webhook signature verification failed: %w", err)
	}

	// Log webhook received
	logWebhookInfo("Received webhook event", map[string]interface{}{
		"event_id":   event.ID,
		"event_type": event.Type,
	})

	// Check for duplicate event (idempotency)
	existingEvent, err := s.repo.GetEventByStripeEventID(ctx, event.ID)
	if err == nil && existingEvent != nil {
		logWebhookInfo("Duplicate webhook event detected, skipping", map[string]interface{}{
			"event_id":   event.ID,
			"event_type": event.Type,
		})
		return nil
	}

	// Process the webhook with retry mechanism
	err = s.processWebhookWithRetry(ctx, event)
	if err != nil {
		logWebhookError("Failed to process webhook event", err, map[string]interface{}{
			"event_id":   event.ID,
			"event_type": event.Type,
		})
		// Add to retry queue if not already there
		if s.webhookRepo != nil {
			retryErr := s.webhookRepo.AddToRetryQueue(ctx, event.ID, string(event.Type), event, 3)
			if retryErr != nil {
				logWebhookError("Failed to add webhook event to retry queue", retryErr, map[string]interface{}{
					"event_id":   event.ID,
					"event_type": event.Type,
				})
			} else {
				logWebhookInfo("Added webhook event to retry queue", map[string]interface{}{
					"event_id":   event.ID,
					"event_type": event.Type,
				})
			}
		}
		return err
	}

	logWebhookInfo("Successfully processed webhook event", map[string]interface{}{
		"event_id":   event.ID,
		"event_type": event.Type,
	})
	return nil
}

// verifyWebhookSignature attempts verification with each configured Stripe webhook secret
// so the service can honor Stripe's per-endpoint multiple secret requirement.
func (s *SubscriptionService) verifyWebhookSignature(payload []byte, signature string) (stripe.Event, error) {
	var lastErr error
	for _, secret := range s.cfg.Stripe.WebhookSecrets {
		if secret == "" {
			continue
		}
		event, err := webhook.ConstructEvent(payload, signature, secret)
		if err == nil {
			return event, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = errors.New("no webhook secrets configured")
	}
	return stripe.Event{}, lastErr
}

// processWebhookWithRetry processes a webhook event and handles the routing to specific handlers
func (s *SubscriptionService) processWebhookWithRetry(ctx context.Context, event stripe.Event) error {
	// Handle different event types
	switch event.Type {
	case "customer.subscription.created":
		return s.handleSubscriptionCreated(ctx, event)
	case "customer.subscription.updated":
		return s.handleSubscriptionUpdated(ctx, event)
	case "customer.subscription.deleted":
		return s.handleSubscriptionDeleted(ctx, event)
	case "invoice.paid", "invoice.payment_succeeded":
		return s.handleInvoicePaid(ctx, event)
	case "invoice.payment_failed":
		return s.handleInvoicePaymentFailed(ctx, event)
	case "invoice.finalized":
		return s.handleInvoiceFinalized(ctx, event)
	case "payment_intent.succeeded":
		return s.handlePaymentIntentSucceeded(ctx, event)
	case "payment_intent.payment_failed":
		return s.handlePaymentIntentFailed(ctx, event)
	case "charge.dispute.created":
		return s.handleDisputeCreated(ctx, event)
	default:
		logWebhookWarn("Unhandled webhook event type", map[string]interface{}{
			"event_type": event.Type,
		})
		return nil
	}
}

// handleSubscriptionCreated processes subscription.created events
func (s *SubscriptionService) handleSubscriptionCreated(ctx context.Context, event stripe.Event) error {
	var stripeSubscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &stripeSubscription); err != nil {
		logWebhookError("Failed to unmarshal subscription.created event", err, map[string]interface{}{
			"event_id":   event.ID,
			"event_type": event.Type,
		})
		return fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	logWebhookInfo("Processing subscription.created", map[string]interface{}{
		"event_id":        event.ID,
		"event_type":      event.Type,
		"customer_id":     stripeSubscription.Customer.ID,
		"subscription_id": stripeSubscription.ID,
	})

	// Get subscription by customer ID
	sub, err := s.repo.GetByStripeCustomerID(ctx, stripeSubscription.Customer.ID)
	if err != nil {
		logWebhookError("Failed to find subscription by customer ID", err, map[string]interface{}{
			"event_id":    event.ID,
			"event_type":  event.Type,
			"customer_id": stripeSubscription.Customer.ID,
		})
		return fmt.Errorf("failed to find subscription by customer ID: %w", err)
	}

	// Determine tier from price ID
	tier := s.getTierFromPriceID(stripeSubscription.Items.Data[0].Price.ID)

	// Update subscription with Stripe subscription details
	sub.StripeSubscriptionID = &stripeSubscription.ID
	sub.StripePriceID = &stripeSubscription.Items.Data[0].Price.ID
	sub.Status = string(stripeSubscription.Status)
	sub.Tier = tier
	sub.CurrentPeriodStart = timePtr(time.Unix(stripeSubscription.CurrentPeriodStart, 0))
	sub.CurrentPeriodEnd = timePtr(time.Unix(stripeSubscription.CurrentPeriodEnd, 0))
	sub.CancelAtPeriodEnd = stripeSubscription.CancelAtPeriodEnd

	if stripeSubscription.CanceledAt > 0 {
		sub.CanceledAt = timePtr(time.Unix(stripeSubscription.CanceledAt, 0))
	}

	if stripeSubscription.TrialStart > 0 {
		sub.TrialStart = timePtr(time.Unix(stripeSubscription.TrialStart, 0))
	}

	if stripeSubscription.TrialEnd > 0 {
		sub.TrialEnd = timePtr(time.Unix(stripeSubscription.TrialEnd, 0))
	}

	if err := s.repo.Update(ctx, sub); err != nil {
		logWebhookError("Failed to update subscription for customer", err, map[string]interface{}{
			"event_id":        event.ID,
			"event_type":      event.Type,
			"customer_id":     stripeSubscription.Customer.ID,
			"subscription_id": stripeSubscription.ID,
		})
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// Log event
	if err := s.repo.LogSubscriptionEvent(ctx, &sub.ID, "subscription_created", &event.ID, stripeSubscription); err != nil {
		logWebhookError("Failed to log subscription event", err, map[string]interface{}{
			"event_id":        event.ID,
			"event_type":      event.Type,
			"subscription_id": sub.ID,
		})
	}

	// Log audit event
	if s.auditLogSvc != nil {
		_ = s.auditLogSvc.LogSubscriptionEvent(ctx, sub.UserID, "subscription_created", map[string]interface{}{
			"subscription_id": stripeSubscription.ID,
			"tier":            tier,
			"status":          string(stripeSubscription.Status),
		})
	}

	logWebhookInfo("Successfully created subscription", map[string]interface{}{
		"event_id":        event.ID,
		"event_type":      event.Type,
		"user_id":         sub.UserID,
		"subscription_id": stripeSubscription.ID,
		"tier":            tier,
		"status":          string(stripeSubscription.Status),
	})
	return nil
}

// handleSubscriptionUpdated processes subscription.updated events
func (s *SubscriptionService) handleSubscriptionUpdated(ctx context.Context, event stripe.Event) error {
	var stripeSubscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &stripeSubscription); err != nil {
		logWebhookError("Failed to unmarshal subscription.updated event", err, map[string]interface{}{
			"event_id":   event.ID,
			"event_type": event.Type,
		})
		return fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	logWebhookInfo("Processing subscription.updated", map[string]interface{}{
		"event_id":        event.ID,
		"event_type":      event.Type,
		"subscription_id": stripeSubscription.ID,
	})

	// Get subscription by Stripe subscription ID
	sub, err := s.repo.GetByStripeSubscriptionID(ctx, stripeSubscription.ID)
	if err != nil {
		logWebhookError("Failed to find subscription", err, map[string]interface{}{
			"event_id":        event.ID,
			"event_type":      event.Type,
			"subscription_id": stripeSubscription.ID,
		})
		return fmt.Errorf("failed to find subscription: %w", err)
	}

	// Determine tier from price ID
	tier := s.getTierFromPriceID(stripeSubscription.Items.Data[0].Price.ID)

	// Update subscription details
	sub.StripePriceID = &stripeSubscription.Items.Data[0].Price.ID
	sub.Status = string(stripeSubscription.Status)
	sub.Tier = tier
	sub.CurrentPeriodStart = timePtr(time.Unix(stripeSubscription.CurrentPeriodStart, 0))
	sub.CurrentPeriodEnd = timePtr(time.Unix(stripeSubscription.CurrentPeriodEnd, 0))
	sub.CancelAtPeriodEnd = stripeSubscription.CancelAtPeriodEnd

	if stripeSubscription.CanceledAt > 0 {
		sub.CanceledAt = timePtr(time.Unix(stripeSubscription.CanceledAt, 0))
	}

	if err := s.repo.Update(ctx, sub); err != nil {
		logWebhookError("Failed to update subscription", err, map[string]interface{}{
			"event_id":        event.ID,
			"event_type":      event.Type,
			"subscription_id": stripeSubscription.ID,
		})
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// Log event
	if err := s.repo.LogSubscriptionEvent(ctx, &sub.ID, "subscription_updated", &event.ID, stripeSubscription); err != nil {
		logWebhookError("Failed to log subscription event", err, map[string]interface{}{
			"event_id":        event.ID,
			"event_type":      event.Type,
			"subscription_id": sub.ID,
		})
	}

	// Log audit event
	if s.auditLogSvc != nil {
		_ = s.auditLogSvc.LogSubscriptionEvent(ctx, sub.UserID, "subscription_updated", map[string]interface{}{
			"subscription_id": stripeSubscription.ID,
			"tier":            tier,
			"status":          string(stripeSubscription.Status),
		})
	}

	logWebhookInfo("Successfully updated subscription", map[string]interface{}{
		"event_id":        event.ID,
		"event_type":      event.Type,
		"subscription_id": stripeSubscription.ID,
		"tier":            tier,
		"status":          string(stripeSubscription.Status),
	})
	return nil
}

// handleSubscriptionDeleted processes subscription.deleted events
func (s *SubscriptionService) handleSubscriptionDeleted(ctx context.Context, event stripe.Event) error {
	var stripeSubscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &stripeSubscription); err != nil {
		logWebhookError("Failed to unmarshal subscription.deleted event", err, map[string]interface{}{
			"event_id":   event.ID,
			"event_type": event.Type,
		})
		return fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	logWebhookInfo("Processing subscription.deleted", map[string]interface{}{
		"event_id":        event.ID,
		"event_type":      event.Type,
		"subscription_id": stripeSubscription.ID,
	})

	// Get subscription by Stripe subscription ID
	sub, err := s.repo.GetByStripeSubscriptionID(ctx, stripeSubscription.ID)
	if err != nil {
		logWebhookError("Failed to find subscription", err, map[string]interface{}{
			"event_id":        event.ID,
			"event_type":      event.Type,
			"subscription_id": stripeSubscription.ID,
		})
		return fmt.Errorf("failed to find subscription: %w", err)
	}

	// Update subscription to canceled/inactive
	sub.Status = "canceled"
	sub.Tier = "free"
	sub.CanceledAt = timePtr(time.Now())

	if err := s.repo.Update(ctx, sub); err != nil {
		logWebhookError("Failed to update subscription to canceled", err, map[string]interface{}{
			"event_id":        event.ID,
			"event_type":      event.Type,
			"subscription_id": stripeSubscription.ID,
		})
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// Log event
	if err := s.repo.LogSubscriptionEvent(ctx, &sub.ID, "subscription_deleted", &event.ID, stripeSubscription); err != nil {
		logWebhookError("Failed to log subscription event", err, map[string]interface{}{
			"event_id":        event.ID,
			"event_type":      event.Type,
			"subscription_id": sub.ID,
		})
	}

	// Log audit event
	if s.auditLogSvc != nil {
		_ = s.auditLogSvc.LogSubscriptionEvent(ctx, sub.UserID, "subscription_deleted", map[string]interface{}{
			"subscription_id": stripeSubscription.ID,
		})
	}

	logWebhookInfo("Successfully deleted subscription", map[string]interface{}{
		"event_id":        event.ID,
		"event_type":      event.Type,
		"subscription_id": stripeSubscription.ID,
		"user_id":         sub.UserID,
	})
	return nil
}

// handleInvoicePaid processes invoice.paid events
func (s *SubscriptionService) handleInvoicePaid(ctx context.Context, event stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		logWebhookError("Failed to unmarshal invoice.paid event", err, map[string]interface{}{
			"event_id":   event.ID,
			"event_type": event.Type,
		})
		return fmt.Errorf("failed to unmarshal invoice: %w", err)
	}

	if invoice.Subscription == nil {
		logWebhookInfo("Invoice is not a subscription invoice, skipping", map[string]interface{}{
			"event_id":   event.ID,
			"event_type": event.Type,
			"invoice_id": invoice.ID,
		})
		return nil // Not a subscription invoice
	}

	logWebhookInfo("Processing invoice.paid", map[string]interface{}{
		"event_id":        event.ID,
		"event_type":      event.Type,
		"invoice_id":      invoice.ID,
		"subscription_id": invoice.Subscription.ID,
	})

	// Get subscription by Stripe subscription ID
	sub, err := s.repo.GetByStripeSubscriptionID(ctx, invoice.Subscription.ID)
	if err != nil {
		logWebhookError("Failed to find subscription for invoice", err, map[string]interface{}{
			"event_id":        event.ID,
			"event_type":      event.Type,
			"invoice_id":      invoice.ID,
			"subscription_id": invoice.Subscription.ID,
		})
		return nil // Not critical
	}

	// Process payment success for dunning (clears grace period and marks failures as resolved)
	if s.dunningService != nil {
		if err := s.dunningService.HandlePaymentSuccess(ctx, &invoice); err != nil {
			logWebhookError("Failed to process payment success in dunning service", err, map[string]interface{}{
				"event_id":   event.ID,
				"event_type": event.Type,
				"invoice_id": invoice.ID,
			})
		}
	}

	// Log event
	if err := s.repo.LogSubscriptionEvent(ctx, &sub.ID, "invoice_paid", &event.ID, invoice); err != nil {
		logWebhookError("Failed to log subscription event", err, map[string]interface{}{
			"event_id":        event.ID,
			"event_type":      event.Type,
			"subscription_id": sub.ID,
			"invoice_id":      invoice.ID,
		})
	}

	// Log audit event
	if s.auditLogSvc != nil {
		_ = s.auditLogSvc.LogSubscriptionEvent(ctx, sub.UserID, "invoice_paid", map[string]interface{}{
			"invoice_id":      invoice.ID,
			"amount_paid":     invoice.AmountPaid,
			"subscription_id": invoice.Subscription.ID,
		})
	}

	logWebhookInfo("Successfully processed invoice.paid", map[string]interface{}{
		"event_id":        event.ID,
		"event_type":      event.Type,
		"invoice_id":      invoice.ID,
		"subscription_id": invoice.Subscription.ID,
	})
	return nil
}

// handleInvoicePaymentFailed processes invoice.payment_failed events
func (s *SubscriptionService) handleInvoicePaymentFailed(ctx context.Context, event stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		logWebhookError("Failed to unmarshal invoice.payment_failed event", err, map[string]interface{}{
			"event_id":   event.ID,
			"event_type": event.Type,
		})
		return fmt.Errorf("failed to unmarshal invoice: %w", err)
	}

	if invoice.Subscription == nil {
		logWebhookInfo("Invoice is not a subscription invoice, skipping", map[string]interface{}{
			"event_id":   event.ID,
			"event_type": event.Type,
			"invoice_id": invoice.ID,
		})
		return nil // Not a subscription invoice
	}

	logWebhookInfo("Processing invoice.payment_failed", map[string]interface{}{
		"event_id":        event.ID,
		"event_type":      event.Type,
		"invoice_id":      invoice.ID,
		"subscription_id": invoice.Subscription.ID,
	})

	// Get subscription by Stripe subscription ID
	sub, err := s.repo.GetByStripeSubscriptionID(ctx, invoice.Subscription.ID)
	if err != nil {
		logWebhookError("Failed to find subscription for invoice", err, map[string]interface{}{
			"event_id":        event.ID,
			"event_type":      event.Type,
			"invoice_id":      invoice.ID,
			"subscription_id": invoice.Subscription.ID,
		})
		return nil // Not critical
	}

	// Update subscription status if needed
	if sub.Status != "past_due" && sub.Status != "unpaid" {
		sub.Status = "past_due"
		if err := s.repo.Update(ctx, sub); err != nil {
			logWebhookError("Failed to update subscription status to past_due", err, map[string]interface{}{
				"event_id":        event.ID,
				"event_type":      event.Type,
				"subscription_id": sub.ID,
			})
		} else {
			logWebhookInfo("Updated subscription status to past_due", map[string]interface{}{
				"event_id":        event.ID,
				"event_type":      event.Type,
				"subscription_id": sub.ID,
			})
		}
	}

	// Process payment failure through dunning service
	if s.dunningService != nil {
		if err := s.dunningService.HandlePaymentFailure(ctx, &invoice); err != nil {
			logWebhookError("Failed to process payment failure in dunning service", err, map[string]interface{}{
				"event_id":   event.ID,
				"event_type": event.Type,
				"invoice_id": invoice.ID,
			})
		}
	}

	// Log event
	if err := s.repo.LogSubscriptionEvent(ctx, &sub.ID, "invoice_payment_failed", &event.ID, invoice); err != nil {
		logWebhookError("Failed to log subscription event", err, map[string]interface{}{
			"event_id":        event.ID,
			"event_type":      event.Type,
			"subscription_id": sub.ID,
			"invoice_id":      invoice.ID,
		})
	}

	// Log audit event
	if s.auditLogSvc != nil {
		_ = s.auditLogSvc.LogSubscriptionEvent(ctx, sub.UserID, "invoice_payment_failed", map[string]interface{}{
			"invoice_id":      invoice.ID,
			"amount_due":      invoice.AmountDue,
			"subscription_id": invoice.Subscription.ID,
		})
	}

	logWebhookInfo("Successfully processed invoice.payment_failed", map[string]interface{}{
		"event_id":        event.ID,
		"event_type":      event.Type,
		"invoice_id":      invoice.ID,
		"subscription_id": invoice.Subscription.ID,
	})
	return nil
}

// handleInvoiceFinalized processes invoice.finalized events and sends invoice PDFs to customers
func (s *SubscriptionService) handleInvoiceFinalized(ctx context.Context, event stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		logWebhookError("Failed to unmarshal invoice.finalized event", err, map[string]interface{}{
			"event_id":   event.ID,
			"event_type": event.Type,
		})
		return fmt.Errorf("failed to unmarshal invoice: %w", err)
	}

	logWebhookInfo("Processing invoice.finalized", map[string]interface{}{
		"event_id":    event.ID,
		"event_type":  event.Type,
		"invoice_id":  invoice.ID,
		"customer_id": invoice.Customer.ID,
	})

	// Skip if invoice is not related to a subscription
	if invoice.Subscription == nil {
		logWebhookInfo("Invoice is not a subscription invoice, skipping", map[string]interface{}{
			"event_id":   event.ID,
			"event_type": event.Type,
			"invoice_id": invoice.ID,
		})
		return nil
	}

	// Skip if invoice PDF delivery is disabled
	if !s.cfg.Stripe.InvoicePDFEnabled {
		logWebhookInfo("Invoice PDF delivery disabled, skipping", map[string]interface{}{
			"event_id":   event.ID,
			"event_type": event.Type,
			"invoice_id": invoice.ID,
		})
		return nil
	}

	// Skip if no invoice PDF URL is available (shouldn't happen for finalized invoices)
	if invoice.InvoicePDF == "" {
		logWebhookWarn("No invoice PDF URL available", map[string]interface{}{
			"event_id":   event.ID,
			"event_type": event.Type,
			"invoice_id": invoice.ID,
		})
		return nil
	}

	// Get subscription by Stripe customer ID
	sub, err := s.repo.GetByStripeCustomerID(ctx, invoice.Customer.ID)
	if err != nil {
		logWebhookError("Failed to find subscription for customer", err, map[string]interface{}{
			"event_id":    event.ID,
			"event_type":  event.Type,
			"customer_id": invoice.Customer.ID,
		})
		return nil // Not critical, don't fail the webhook
	}

	// Get user for email
	user, err := s.userRepo.GetByID(ctx, sub.UserID)
	if err != nil {
		logWebhookError("Failed to get user for invoice email", err, map[string]interface{}{
			"event_id":   event.ID,
			"event_type": event.Type,
			"user_id":    sub.UserID,
			"invoice_id": invoice.ID,
		})
		return nil // Not critical
	}

	// Send invoice email with PDF link
	if s.emailService != nil {
		emailData := map[string]interface{}{
			"InvoiceID":        invoice.ID,
			"InvoicePDFURL":    invoice.InvoicePDF,
			"HostedInvoiceURL": invoice.HostedInvoiceURL,
			"AmountDue":        formatAmountForCurrency(invoice.AmountDue, string(invoice.Currency)),
			"Currency":         string(invoice.Currency),
			"InvoiceNumber":    invoice.Number,
		}

		// Add tax information if available
		if invoice.AutomaticTax != nil && invoice.AutomaticTax.Status != "" {
			emailData["TaxStatus"] = string(invoice.AutomaticTax.Status)
		}
		if invoice.Tax > 0 {
			emailData["TaxAmount"] = formatAmountForCurrency(invoice.Tax, string(invoice.Currency))
			// Always set Subtotal when TaxAmount is present
			emailData["Subtotal"] = formatAmountForCurrency(invoice.Subtotal, string(invoice.Currency))
		}
		// Always set Total; use AmountDue as fallback if Total is not positive
		if invoice.Total > 0 {
			emailData["Total"] = formatAmountForCurrency(invoice.Total, string(invoice.Currency))
		} else {
			emailData["Total"] = formatAmountForCurrency(invoice.AmountDue, string(invoice.Currency))
		}

		notificationID := uuid.New()
		if err := s.emailService.SendNotificationEmail(ctx, user, models.NotificationTypeInvoiceFinalized, notificationID, emailData); err != nil {
			logWebhookError("Failed to send invoice email", err, map[string]interface{}{
				"event_id":   event.ID,
				"event_type": event.Type,
				"user_id":    user.ID,
				"invoice_id": invoice.ID,
			})
			// Continue processing, email failure shouldn't fail the webhook
		} else {
			logWebhookInfo("Invoice email sent", map[string]interface{}{
				"event_id":   event.ID,
				"event_type": event.Type,
				"user_id":    user.ID,
				"invoice_id": invoice.ID,
			})
		}
	}

	// Log audit event
	if s.auditLogSvc != nil {
		_ = s.auditLogSvc.LogSubscriptionEvent(ctx, sub.UserID, "invoice_finalized", map[string]interface{}{
			"invoice_id":     invoice.ID,
			"invoice_number": invoice.Number,
			"amount_due":     invoice.AmountDue,
			"tax":            invoice.Tax,
			"pdf_url":        invoice.InvoicePDF,
		})
	}

	logWebhookInfo("Successfully processed invoice.finalized", map[string]interface{}{
		"event_id":   event.ID,
		"event_type": event.Type,
		"invoice_id": invoice.ID,
	})
	return nil
}

// formatAmountForCurrency formats an amount in smallest currency unit for display with currency
// Handles zero-decimal currencies (JPY, KRW, etc.) and three-decimal currencies (KWD, BHD, etc.)
func formatAmountForCurrency(amount int64, currency string) string {
	currency = strings.ToUpper(currency)

	// Zero-decimal currencies (no decimal places)
	zeroDecimalCurrencies := map[string]bool{
		"BIF": true, "CLP": true, "DJF": true, "GNF": true, "JPY": true,
		"KMF": true, "KRW": true, "MGA": true, "PYG": true, "RWF": true,
		"UGX": true, "VND": true, "VUV": true, "XAF": true, "XOF": true, "XPF": true,
	}

	// Three-decimal currencies
	threeDecimalCurrencies := map[string]bool{
		"BHD": true, "JOD": true, "KWD": true, "OMR": true, "TND": true,
	}

	if zeroDecimalCurrencies[currency] {
		return fmt.Sprintf("%d %s", amount, currency)
	}

	if threeDecimalCurrencies[currency] {
		return fmt.Sprintf("%.3f %s", float64(amount)/1000, currency)
	}

	// Default: two decimal places (most currencies)
	return fmt.Sprintf("%.2f %s", float64(amount)/100, currency)
}

// getTierFromPriceID determines the subscription tier from Stripe price ID
func (s *SubscriptionService) getTierFromPriceID(priceID string) string {
	if priceID == s.cfg.Stripe.ProMonthlyPriceID || priceID == s.cfg.Stripe.ProYearlyPriceID {
		return "pro"
	}
	return "free"
}

// ChangeSubscriptionPlan changes a user's subscription plan with proration
func (s *SubscriptionService) ChangeSubscriptionPlan(ctx context.Context, user *models.User, newPriceID string) error {
	// Validate new price ID
	if newPriceID != s.cfg.Stripe.ProMonthlyPriceID && newPriceID != s.cfg.Stripe.ProYearlyPriceID {
		return ErrInvalidPriceID
	}

	// Get existing subscription
	sub, err := s.repo.GetByUserID(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	if sub.StripeSubscriptionID == nil || *sub.StripeSubscriptionID == "" {
		return errors.New("no active stripe subscription found")
	}

	// Get the subscription from Stripe
	stripeSubscription, err := subscription.Get(*sub.StripeSubscriptionID, nil)
	if err != nil {
		return fmt.Errorf("failed to get stripe subscription: %w", err)
	}

	// Validate subscription has items
	if len(stripeSubscription.Items.Data) == 0 {
		return errors.New("subscription has no items")
	}

	// Validate first item has a price
	if stripeSubscription.Items.Data[0].Price == nil {
		return errors.New("subscription item has no price")
	}

	// Check if already on this plan
	if stripeSubscription.Items.Data[0].Price.ID == newPriceID {
		return errors.New("already subscribed to this plan")
	}

	// Update subscription with proration
	subscriptionItemID := stripeSubscription.Items.Data[0].ID
	params := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:    stripe.String(subscriptionItemID),
				Price: stripe.String(newPriceID),
			},
		},
		ProrationBehavior: stripe.String("always_invoice"),
	}

	_, err = subscription.Update(*sub.StripeSubscriptionID, params)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// Log audit event
	if s.auditLogSvc != nil {
		_ = s.auditLogSvc.LogSubscriptionEvent(ctx, user.ID, "subscription_plan_changed", map[string]interface{}{
			"old_price_id": stripeSubscription.Items.Data[0].Price.ID,
			"new_price_id": newPriceID,
			"proration":    "always_invoice",
		})
	}

	return nil
}

// CancelSubscription cancels a user's subscription
// If immediate is true, cancels immediately. Otherwise, cancels at period end.
func (s *SubscriptionService) CancelSubscription(ctx context.Context, user *models.User, immediate bool) error {
	// Get existing subscription
	sub, err := s.repo.GetByUserID(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	if sub.StripeSubscriptionID == nil || *sub.StripeSubscriptionID == "" {
		return errors.New("no active stripe subscription found")
	}

	// Cancel the subscription in Stripe
	var canceledSub *stripe.Subscription

	if immediate {
		// Cancel immediately
		cancelParams := &stripe.SubscriptionCancelParams{}
		canceledSub, err = subscription.Cancel(*sub.StripeSubscriptionID, cancelParams)
	} else {
		// Cancel at period end
		updateParams := &stripe.SubscriptionParams{CancelAtPeriodEnd: stripe.Bool(true)}
		canceledSub, err = subscription.Update(*sub.StripeSubscriptionID, updateParams)
	}

	if err != nil {
		return fmt.Errorf("failed to cancel subscription: %w", err)
	}

	// Update local subscription record
	sub.CancelAtPeriodEnd = canceledSub.CancelAtPeriodEnd
	if canceledSub.Status == "canceled" {
		sub.Status = "canceled"
		sub.CanceledAt = timePtr(time.Now())
	}

	if err := s.repo.Update(ctx, sub); err != nil {
		utils.Error("Failed to update subscription after cancellation", err, map[string]interface{}{
			"subscription_id": sub.ID,
			"user_id":         user.ID,
		})
		return fmt.Errorf("subscription cancelled in Stripe but failed to update local record: %w", err)
	}

	// Log audit event
	if s.auditLogSvc != nil {
		_ = s.auditLogSvc.LogSubscriptionEvent(ctx, user.ID, "subscription_canceled", map[string]interface{}{
			"subscription_id":      *sub.StripeSubscriptionID,
			"immediate":            immediate,
			"cancel_at_period_end": canceledSub.CancelAtPeriodEnd,
		})
	}

	return nil
}

// GetInvoices retrieves a user's invoices from Stripe
func (s *SubscriptionService) GetInvoices(ctx context.Context, user *models.User, limit int64) ([]*stripe.Invoice, error) {
	// Get subscription to find customer ID
	sub, err := s.repo.GetByUserID(ctx, user.ID)
	if err != nil {
		return nil, ErrSubscriptionNotFound
	}

	if sub.StripeCustomerID == "" {
		return nil, ErrStripeCustomerNotFound
	}

	// Set default limit if not provided
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100 // Cap at 100 invoices
	}

	// Fetch invoices from Stripe
	params := &stripe.InvoiceListParams{
		Customer: stripe.String(sub.StripeCustomerID),
	}
	params.Limit = stripe.Int64(limit)

	invoices := []*stripe.Invoice{}
	iter := invoice.List(params)
	for iter.Next() {
		invoices = append(invoices, iter.Invoice())
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to fetch invoices: %w", err)
	}

	return invoices, nil
}

// ReactivateSubscription reactivates a subscription that was set to cancel at period end
func (s *SubscriptionService) ReactivateSubscription(ctx context.Context, user *models.User) error {
	// Get existing subscription
	sub, err := s.repo.GetByUserID(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	if sub.StripeSubscriptionID == nil || *sub.StripeSubscriptionID == "" {
		return errors.New("no active stripe subscription found")
	}

	if !sub.CancelAtPeriodEnd {
		return errors.New("subscription is not scheduled for cancellation")
	}

	// Reactivate the subscription in Stripe
	params := &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(false),
	}

	reactivatedSub, err := subscription.Update(*sub.StripeSubscriptionID, params)
	if err != nil {
		return fmt.Errorf("failed to reactivate subscription: %w", err)
	}

	// Update local subscription record
	sub.CancelAtPeriodEnd = false
	sub.Status = string(reactivatedSub.Status)

	if err := s.repo.Update(ctx, sub); err != nil {
		utils.Error("Failed to update subscription after reactivation", err, map[string]interface{}{
			"subscription_id": sub.ID,
			"user_id":         user.ID,
		})
		return fmt.Errorf("subscription reactivated in Stripe but failed to update local record: %w", err)
	}

	// Log audit event
	if s.auditLogSvc != nil {
		_ = s.auditLogSvc.LogSubscriptionEvent(ctx, user.ID, "subscription_reactivated", map[string]interface{}{
			"subscription_id": *sub.StripeSubscriptionID,
		})
	}

	return nil
}

// HasActiveSubscription checks if user has an active subscription (including grace period)
func (s *SubscriptionService) HasActiveSubscription(ctx context.Context, userID uuid.UUID) bool {
	sub, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return false
	}

	// Active or trialing status
	if sub.Status == "active" || sub.Status == "trialing" {
		return true
	}

	// In grace period for past_due or unpaid subscriptions
	if (sub.Status == "past_due" || sub.Status == "unpaid") && s.isInGracePeriod(sub) {
		return true
	}

	return false
}

// IsProUser checks if user has an active Pro subscription (including grace period)
func (s *SubscriptionService) IsProUser(ctx context.Context, userID uuid.UUID) bool {
	sub, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return false
	}

	// Must be Pro tier
	if sub.Tier != "pro" {
		return false
	}

	// Active or trialing status
	if sub.Status == "active" || sub.Status == "trialing" {
		return true
	}

	// In grace period for past_due or unpaid subscriptions
	if (sub.Status == "past_due" || sub.Status == "unpaid") && s.isInGracePeriod(sub) {
		return true
	}

	return false
}

// isInGracePeriod checks if a subscription is currently in grace period
func (s *SubscriptionService) isInGracePeriod(sub *models.Subscription) bool {
	if sub.GracePeriodEnd == nil {
		return false
	}
	return time.Now().Before(*sub.GracePeriodEnd)
}

// handlePaymentIntentSucceeded processes payment_intent.succeeded events
func (s *SubscriptionService) handlePaymentIntentSucceeded(ctx context.Context, event stripe.Event) error {
	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		logWebhookError("Failed to unmarshal payment_intent.succeeded event", err, map[string]interface{}{
			"event_id":   event.ID,
			"event_type": event.Type,
		})
		return fmt.Errorf("failed to unmarshal payment intent: %w", err)
	}

	customerID := ""
	if paymentIntent.Customer != nil {
		customerID = paymentIntent.Customer.ID
	}

	logWebhookInfo("Processing payment_intent.succeeded", map[string]interface{}{
		"event_id":          event.ID,
		"event_type":        event.Type,
		"payment_intent_id": paymentIntent.ID,
		"customer_id":       customerID,
		"amount":            paymentIntent.Amount,
		"currency":          paymentIntent.Currency,
	})

	// Log successful payment
	if s.auditLogSvc != nil {
		metadata := map[string]interface{}{
			"payment_intent_id": paymentIntent.ID,
			"amount_cents":      paymentIntent.Amount,
			"currency":          paymentIntent.Currency,
			"status":            string(paymentIntent.Status),
		}
		if paymentIntent.Customer != nil {
			metadata["stripe_customer_id"] = paymentIntent.Customer.ID
		}

		// Try to get user ID from subscription if available
		if paymentIntent.Invoice != nil && paymentIntent.Invoice.Subscription != nil {
			sub, err := s.repo.GetByStripeSubscriptionID(ctx, paymentIntent.Invoice.Subscription.ID)
			if err == nil {
				_ = s.auditLogSvc.LogSubscriptionEvent(ctx, sub.UserID, "payment_intent_succeeded", metadata)
			}
		}
	}

	logWebhookInfo("Successfully processed payment_intent.succeeded", map[string]interface{}{
		"event_id":          event.ID,
		"event_type":        event.Type,
		"payment_intent_id": paymentIntent.ID,
	})
	return nil
}

// handlePaymentIntentFailed processes payment_intent.payment_failed events
func (s *SubscriptionService) handlePaymentIntentFailed(ctx context.Context, event stripe.Event) error {
	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		logWebhookError("Failed to unmarshal payment_intent.payment_failed event", err, map[string]interface{}{
			"event_id":   event.ID,
			"event_type": event.Type,
		})
		return fmt.Errorf("failed to unmarshal payment intent: %w", err)
	}

	customerID := ""
	if paymentIntent.Customer != nil {
		customerID = paymentIntent.Customer.ID
	}

	logWebhookInfo("Processing payment_intent.payment_failed", map[string]interface{}{
		"event_id":          event.ID,
		"event_type":        event.Type,
		"payment_intent_id": paymentIntent.ID,
		"customer_id":       customerID,
		"amount":            paymentIntent.Amount,
		"currency":          paymentIntent.Currency,
	})

	// Log failed payment
	if s.auditLogSvc != nil {
		metadata := map[string]interface{}{
			"payment_intent_id": paymentIntent.ID,
			"amount_cents":      paymentIntent.Amount,
			"currency":          paymentIntent.Currency,
			"status":            string(paymentIntent.Status),
		}
		if paymentIntent.LastPaymentError != nil {
			metadata["error_code"] = paymentIntent.LastPaymentError.Code
			metadata["error_message"] = paymentIntent.LastPaymentError.Msg
		}
		if paymentIntent.Customer != nil {
			metadata["stripe_customer_id"] = paymentIntent.Customer.ID
		}

		// Try to get user ID from subscription if available
		if paymentIntent.Invoice != nil && paymentIntent.Invoice.Subscription != nil {
			sub, err := s.repo.GetByStripeSubscriptionID(ctx, paymentIntent.Invoice.Subscription.ID)
			if err == nil {
				_ = s.auditLogSvc.LogSubscriptionEvent(ctx, sub.UserID, "payment_intent_failed", metadata)
			}
		}
	}

	logWebhookInfo("Successfully processed payment_intent.payment_failed", map[string]interface{}{
		"event_id":          event.ID,
		"event_type":        event.Type,
		"payment_intent_id": paymentIntent.ID,
	})
	return nil
}

// handleDisputeCreated processes charge.dispute.created events
func (s *SubscriptionService) handleDisputeCreated(ctx context.Context, event stripe.Event) error {
	var dispute stripe.Dispute
	if err := json.Unmarshal(event.Data.Raw, &dispute); err != nil {
		logWebhookError("Failed to unmarshal charge.dispute.created event", err, map[string]interface{}{
			"event_id":   event.ID,
			"event_type": event.Type,
		})
		return fmt.Errorf("failed to unmarshal dispute: %w", err)
	}

	// Get the charge to find the customer
	customerID := ""
	chargeID := "unknown"
	if dispute.Charge != nil {
		chargeID = dispute.Charge.ID
		if dispute.Charge.Customer != nil {
			customerID = dispute.Charge.Customer.ID
		}
	}

	logWebhookInfo("Processing charge.dispute.created", map[string]interface{}{
		"event_id":   event.ID,
		"event_type": event.Type,
		"dispute_id": dispute.ID,
		"charge_id":  chargeID,
		"amount":     dispute.Amount,
		"currency":   dispute.Currency,
		"reason":     dispute.Reason,
	})

	// Try to find subscription by customer ID (single lookup)
	var userID uuid.UUID
	var subscription *models.Subscription
	if customerID != "" {
		sub, err := s.repo.GetByStripeCustomerID(ctx, customerID)
		if err == nil {
			userID = sub.UserID
			subscription = sub
		} else {
			logWebhookError("Could not find subscription for customer", err, map[string]interface{}{
				"event_id":    event.ID,
				"event_type":  event.Type,
				"customer_id": customerID,
			})
		}
	}

	// Log dispute event
	if s.auditLogSvc != nil && userID != uuid.Nil {
		metadata := map[string]interface{}{
			"dispute_id":         dispute.ID,
			"amount_cents":       dispute.Amount,
			"currency":           dispute.Currency,
			"reason":             dispute.Reason,
			"status":             string(dispute.Status),
			"stripe_customer_id": customerID,
		}
		// Only add charge_id if charge exists
		if dispute.Charge != nil {
			metadata["charge_id"] = dispute.Charge.ID
		}
		if dispute.Evidence != nil {
			metadata["has_evidence"] = true
		}
		_ = s.auditLogSvc.LogSubscriptionEvent(ctx, userID, "dispute_created", metadata)
	}

	// Send email notification about dispute if email service is available
	if s.emailService != nil && userID != uuid.Nil {
		// Get user details
		user, err := s.userRepo.GetByID(ctx, userID)
		if err == nil && user.Email != nil && *user.Email != "" {
			// Send dispute notification email
			emailErr := s.emailService.SendDisputeNotification(ctx, user, &dispute)
			if emailErr != nil {
				logWebhookError("Failed to send dispute notification email", emailErr, map[string]interface{}{
					"event_id":   event.ID,
					"event_type": event.Type,
					"user_id":    userID,
				})
			} else {
				logWebhookInfo("Sent dispute notification email", map[string]interface{}{
					"event_id":   event.ID,
					"event_type": event.Type,
					"user_id":    userID,
				})
			}
		}
	}

	// Log subscription event for record keeping (reuse subscription from earlier lookup)
	if subscription != nil {
		if logErr := s.repo.LogSubscriptionEvent(ctx, &subscription.ID, "dispute_created", &event.ID, dispute); logErr != nil {
			logWebhookError("Failed to log dispute event", logErr, map[string]interface{}{
				"event_id":        event.ID,
				"event_type":      event.Type,
				"subscription_id": subscription.ID,
			})
		}
	}

	logWebhookInfo("Successfully processed charge.dispute.created", map[string]interface{}{
		"event_id":   event.ID,
		"event_type": event.Type,
		"dispute_id": dispute.ID,
	})
	return nil
}

// timePtr returns a pointer to a time.Time
func timePtr(t time.Time) *time.Time {
	return &t
}
