package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v81"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
)

const (
	// Grace period duration - 7 days from first payment failure
	GracePeriodDuration = 7 * 24 * time.Hour
)

var (
	// ErrPaymentFailureNotFound indicates the payment failure was not found
	ErrPaymentFailureNotFound = errors.New("payment failure not found")
)

// DunningService handles dunning process for failed payments
type DunningService struct {
	dunningRepo      *repository.DunningRepository
	subscriptionRepo *repository.SubscriptionRepository
	userRepo         *repository.UserRepository
	emailService     *EmailService
	auditLogSvc      *AuditLogService
}

// NewDunningService creates a new dunning service
func NewDunningService(
	dunningRepo *repository.DunningRepository,
	subscriptionRepo *repository.SubscriptionRepository,
	userRepo *repository.UserRepository,
	emailService *EmailService,
	auditLogSvc *AuditLogService,
) *DunningService {
	return &DunningService{
		dunningRepo:      dunningRepo,
		subscriptionRepo: subscriptionRepo,
		userRepo:         userRepo,
		emailService:     emailService,
		auditLogSvc:      auditLogSvc,
	}
}

// HandlePaymentFailure processes a payment failure and initiates dunning
func (s *DunningService) HandlePaymentFailure(ctx context.Context, invoice *stripe.Invoice) error {
	if invoice.Subscription == nil {
		return nil // Not a subscription invoice
	}

	log.Printf("[DUNNING] Processing payment failure for invoice %s, subscription %s", invoice.ID, invoice.Subscription.ID)

	// Get subscription
	sub, err := s.subscriptionRepo.GetByStripeSubscriptionID(ctx, invoice.Subscription.ID)
	if err != nil {
		log.Printf("[DUNNING] Failed to find subscription %s: %v", invoice.Subscription.ID, err)
		return fmt.Errorf("failed to find subscription: %w", err)
	}

	// Check if payment failure already exists for this invoice
	existingFailure, err := s.dunningRepo.GetPaymentFailureByInvoiceID(ctx, invoice.ID)
	if err == nil && existingFailure != nil {
		log.Printf("[DUNNING] Payment failure already tracked for invoice %s, updating attempt count", invoice.ID)
		// Update existing failure
		existingFailure.AttemptCount++
		if invoice.NextPaymentAttempt > 0 {
			nextRetry := time.Unix(invoice.NextPaymentAttempt, 0)
			existingFailure.NextRetryAt = &nextRetry
		}
		if err := s.dunningRepo.UpdatePaymentFailure(ctx, existingFailure); err != nil {
			log.Printf("[DUNNING] Failed to update payment failure: %v", err)
		}

		// Send retry notification
		if err := s.sendDunningNotification(ctx, sub, existingFailure, models.NotificationTypePaymentRetry, existingFailure.AttemptCount); err != nil {
			log.Printf("[DUNNING] Failed to send retry notification: %v", err)
		}

		return nil
	}

	// Create new payment failure record
	paymentIntentID := ""
	if invoice.PaymentIntent != nil {
		paymentIntentID = invoice.PaymentIntent.ID
	}

	var failureReason *string
	if invoice.LastFinalizationError != nil {
		reason := string(invoice.LastFinalizationError.Code)
		failureReason = &reason
	}

	var nextRetryAt *time.Time
	if invoice.NextPaymentAttempt > 0 {
		nextRetry := time.Unix(invoice.NextPaymentAttempt, 0)
		nextRetryAt = &nextRetry
	}

	failure := &models.PaymentFailure{
		SubscriptionID:        sub.ID,
		StripeInvoiceID:       invoice.ID,
		StripePaymentIntentID: &paymentIntentID,
		AmountDue:             invoice.AmountDue,
		Currency:              string(invoice.Currency),
		AttemptCount:          1,
		FailureReason:         failureReason,
		NextRetryAt:           nextRetryAt,
		Resolved:              false,
	}

	if err := s.dunningRepo.CreatePaymentFailure(ctx, failure); err != nil {
		log.Printf("[DUNNING] Failed to create payment failure record: %v", err)
		return fmt.Errorf("failed to create payment failure: %w", err)
	}

	// Set grace period if not already set
	if sub.GracePeriodEnd == nil {
		gracePeriodEnd := time.Now().Add(GracePeriodDuration)
		if err := s.dunningRepo.SetGracePeriod(ctx, sub.ID, gracePeriodEnd); err != nil {
			log.Printf("[DUNNING] Failed to set grace period: %v", err)
		} else {
			log.Printf("[DUNNING] Grace period set until %s for subscription %s", gracePeriodEnd, sub.ID)
		}
	}

	// Send initial failure notification
	if err := s.sendDunningNotification(ctx, sub, failure, models.NotificationTypePaymentFailed, 1); err != nil {
		log.Printf("[DUNNING] Failed to send payment failed notification: %v", err)
	}

	// Log audit event
	if s.auditLogSvc != nil {
		_ = s.auditLogSvc.LogSubscriptionEvent(ctx, sub.UserID, "payment_failed", map[string]interface{}{
			"invoice_id":    invoice.ID,
			"amount_due":    invoice.AmountDue,
			"currency":      invoice.Currency,
			"attempt_count": 1,
		})
	}

	log.Printf("[DUNNING] Payment failure tracked for invoice %s, grace period active", invoice.ID)
	return nil
}

// HandlePaymentSuccess processes successful payment after previous failures
func (s *DunningService) HandlePaymentSuccess(ctx context.Context, invoice *stripe.Invoice) error {
	if invoice.Subscription == nil {
		return nil
	}

	log.Printf("[DUNNING] Processing payment success for invoice %s, subscription %s", invoice.ID, invoice.Subscription.ID)

	// Get subscription
	sub, err := s.subscriptionRepo.GetByStripeSubscriptionID(ctx, invoice.Subscription.ID)
	if err != nil {
		return fmt.Errorf("failed to find subscription: %w", err)
	}

	// Check for existing payment failures
	failures, err := s.dunningRepo.GetPaymentFailuresBySubscriptionID(ctx, sub.ID)
	if err != nil {
		return fmt.Errorf("failed to get payment failures: %w", err)
	}

	// Mark all unresolved failures as resolved
	for _, failure := range failures {
		if !failure.Resolved {
			if err := s.dunningRepo.MarkPaymentFailureResolved(ctx, failure.ID); err != nil {
				log.Printf("[DUNNING] Failed to mark failure %s as resolved: %v", failure.ID, err)
			}
		}
	}

	// Clear grace period
	if sub.GracePeriodEnd != nil {
		if err := s.dunningRepo.ClearGracePeriod(ctx, sub.ID); err != nil {
			log.Printf("[DUNNING] Failed to clear grace period: %v", err)
		} else {
			log.Printf("[DUNNING] Grace period cleared for subscription %s", sub.ID)
		}
	}

	// Log audit event
	if s.auditLogSvc != nil {
		_ = s.auditLogSvc.LogSubscriptionEvent(ctx, sub.UserID, "payment_recovered", map[string]interface{}{
			"invoice_id":  invoice.ID,
			"amount_paid": invoice.AmountPaid,
		})
	}

	log.Printf("[DUNNING] Payment recovered for subscription %s", sub.ID)
	return nil
}

// ProcessExpiredGracePeriods processes subscriptions whose grace periods have expired
func (s *DunningService) ProcessExpiredGracePeriods(ctx context.Context) error {
	log.Printf("[DUNNING] Processing expired grace periods")

	subscriptions, err := s.dunningRepo.GetExpiredGracePeriodSubscriptions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get expired grace period subscriptions: %w", err)
	}

	log.Printf("[DUNNING] Found %d subscriptions with expired grace periods", len(subscriptions))

	for _, sub := range subscriptions {
		if err := s.downgradeSubscription(ctx, sub); err != nil {
			log.Printf("[DUNNING] Failed to downgrade subscription %s: %v", sub.ID, err)
			continue
		}
	}

	return nil
}

// downgradeSubscription downgrades a subscription to free tier after grace period expiry
func (s *DunningService) downgradeSubscription(ctx context.Context, sub *models.Subscription) error {
	log.Printf("[DUNNING] Downgrading subscription %s (user %s) to free tier", sub.ID, sub.UserID)

	// Store the previous tier for audit logging
	previousTier := sub.Tier

	// Update subscription to free tier
	sub.Tier = "free"
	sub.Status = "canceled"
	if err := s.subscriptionRepo.Update(ctx, sub); err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// Clear grace period
	if err := s.dunningRepo.ClearGracePeriod(ctx, sub.ID); err != nil {
		log.Printf("[DUNNING] Failed to clear grace period: %v", err)
	}

	// Get payment failures for notification
	failures, err := s.dunningRepo.GetPaymentFailuresBySubscriptionID(ctx, sub.ID)
	if err != nil || len(failures) == 0 {
		log.Printf("[DUNNING] No payment failures found for subscription %s", sub.ID)
	} else {
		// Send downgrade notification
		if err := s.sendDunningNotification(ctx, sub, failures[0], models.NotificationTypeSubscriptionDowngraded, 0); err != nil {
			log.Printf("[DUNNING] Failed to send downgrade notification: %v", err)
		}
	}

	// Log audit event with actual previous tier
	if s.auditLogSvc != nil {
		_ = s.auditLogSvc.LogSubscriptionEvent(ctx, sub.UserID, "subscription_downgraded", map[string]interface{}{
			"reason":        "grace_period_expired",
			"previous_tier": previousTier,
			"new_tier":      "free",
		})
	}

	log.Printf("[DUNNING] Successfully downgraded subscription %s to free tier", sub.ID)
	return nil
}

// SendGracePeriodWarnings sends warnings to users approaching grace period expiry
func (s *DunningService) SendGracePeriodWarnings(ctx context.Context) error {
	log.Printf("[DUNNING] Sending grace period warnings")

	subscriptions, err := s.dunningRepo.GetSubscriptionsInGracePeriod(ctx)
	if err != nil {
		return fmt.Errorf("failed to get subscriptions in grace period: %w", err)
	}

	log.Printf("[DUNNING] Found %d subscriptions in grace period", len(subscriptions))

	for _, sub := range subscriptions {
		// Send warning if grace period expires in 2 days or less
		if sub.GracePeriodEnd != nil {
			timeUntilExpiry := time.Until(*sub.GracePeriodEnd)
			if timeUntilExpiry > 0 && timeUntilExpiry <= 2*24*time.Hour {
				failures, err := s.dunningRepo.GetPaymentFailuresBySubscriptionID(ctx, sub.ID)
				if err != nil || len(failures) == 0 {
					continue
				}

				// Check if warning was already sent by looking at dunning attempts
				attempts, err := s.dunningRepo.GetDunningAttemptsByFailureID(ctx, failures[0].ID)
				if err != nil {
					log.Printf("[DUNNING] Failed to get dunning attempts for failure %s: %v", failures[0].ID, err)
				}

				// Skip if warning was already sent
				alreadySent := false
				for _, attempt := range attempts {
					if attempt.NotificationType == models.NotificationTypeGracePeriodWarning && attempt.EmailSent {
						alreadySent = true
						break
					}
				}

				if alreadySent {
					log.Printf("[DUNNING] Grace period warning already sent for subscription %s, skipping", sub.ID)
					continue
				}

				if err := s.sendDunningNotification(ctx, sub, failures[0], models.NotificationTypeGracePeriodWarning, 0); err != nil {
					log.Printf("[DUNNING] Failed to send grace period warning for subscription %s: %v", sub.ID, err)
				}
			}
		}
	}

	return nil
}

// sendDunningNotification sends a dunning notification email to the user
func (s *DunningService) sendDunningNotification(ctx context.Context, sub *models.Subscription, failure *models.PaymentFailure, notificationType string, attemptNumber int) error {
	// Get user
	user, err := s.userRepo.GetByID(ctx, sub.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user.Email == nil || *user.Email == "" {
		log.Printf("[DUNNING] User %s has no email, skipping notification", user.ID)
		return nil
	}

	// Create dunning attempt record
	attempt := &models.DunningAttempt{
		PaymentFailureID: failure.ID,
		UserID:           user.ID,
		AttemptNumber:    attemptNumber,
		NotificationType: notificationType,
		EmailSent:        false,
	}

	// Prepare email data based on notification type
	emailData := s.prepareDunningEmailData(sub, failure, notificationType)

	// Send email using the injected emailService
	// Use a dummy notification ID for dunning emails (they're not tied to in-app notifications)
	notificationID := uuid.New()
	err = s.emailService.SendNotificationEmail(ctx, user, notificationType, notificationID, emailData)
	if err != nil {
		log.Printf("[DUNNING] Failed to send %s notification to user %s (email: %s): %v", notificationType, user.ID, *user.Email, err)
		// Do not mark attempt as sent on failure
		attempt.EmailSent = false
	} else {
		// Mark attempt as sent on success
		log.Printf("[DUNNING] Successfully sent %s notification to user %s (email: %s)", notificationType, user.ID, *user.Email)
		attempt.EmailSent = true
		now := time.Now()
		attempt.EmailSentAt = &now
	}

	// Create dunning attempt record regardless of email success
	if err := s.dunningRepo.CreateDunningAttempt(ctx, attempt); err != nil {
		log.Printf("[DUNNING] Failed to create dunning attempt record: %v", err)
	}

	log.Printf("[DUNNING] Dunning notification logged: type=%s, user=%s, attempt=%d, sent=%v", notificationType, user.ID, attemptNumber, attempt.EmailSent)
	return nil
}

// prepareDunningEmailData prepares email data for dunning notifications
func (s *DunningService) prepareDunningEmailData(sub *models.Subscription, failure *models.PaymentFailure, notificationType string) map[string]interface{} {
	data := map[string]interface{}{
		"SubscriptionID": sub.ID.String(),
		"InvoiceID":      failure.StripeInvoiceID,
		"AmountDue":      fmt.Sprintf("$%.2f", float64(failure.AmountDue)/100),
		"Currency":       failure.Currency,
	}

	if sub.GracePeriodEnd != nil {
		data["GracePeriodEnd"] = sub.GracePeriodEnd.Format("January 2, 2006")
		data["DaysRemaining"] = int(time.Until(*sub.GracePeriodEnd).Hours() / 24)
	}

	if failure.NextRetryAt != nil {
		data["NextRetryAt"] = failure.NextRetryAt.Format("January 2, 2006 at 3:04 PM")
	}

	data["AttemptCount"] = failure.AttemptCount

	return data
}

// GetPaymentFailuresBySubscriptionID retrieves payment failures for a subscription
func (s *DunningService) GetPaymentFailuresBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) ([]*models.PaymentFailure, error) {
	return s.dunningRepo.GetPaymentFailuresBySubscriptionID(ctx, subscriptionID)
}

// GetDunningAttemptsByUserID retrieves dunning attempts for a user
func (s *DunningService) GetDunningAttemptsByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]*models.DunningAttempt, error) {
	return s.dunningRepo.GetDunningAttemptsByUserID(ctx, userID, limit)
}

// IsInGracePeriod checks if a subscription is currently in grace period
func (s *DunningService) IsInGracePeriod(sub *models.Subscription) bool {
	if sub.GracePeriodEnd == nil {
		return false
	}
	return time.Now().Before(*sub.GracePeriodEnd)
}
