package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// TestGracePeriodDuration tests the grace period constant
func TestGracePeriodDuration(t *testing.T) {
	t.Run("grace period is 7 days", func(t *testing.T) {
		assert.Equal(t, 7*24*time.Hour, GracePeriodDuration)
	})
}

// TestIsInGracePeriod tests grace period checking logic
func TestIsInGracePeriod(t *testing.T) {
	service := &DunningService{}

	tests := []struct {
		name         string
		subscription *models.Subscription
		expected     bool
	}{
		{
			name: "No grace period set",
			subscription: &models.Subscription{
				GracePeriodEnd: nil,
			},
			expected: false,
		},
		{
			name: "Grace period in future",
			subscription: &models.Subscription{
				GracePeriodEnd: timePtr(time.Now().Add(24 * time.Hour)),
			},
			expected: true,
		},
		{
			name: "Grace period expired",
			subscription: &models.Subscription{
				GracePeriodEnd: timePtr(time.Now().Add(-24 * time.Hour)),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.IsInGracePeriod(tt.subscription)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestPrepareDunningEmailData tests email data preparation
func TestPrepareDunningEmailData(t *testing.T) {
	service := &DunningService{}

	subID := uuid.New()
	invoiceID := "inv_test123"
	gracePeriodEnd := time.Now().Add(5 * 24 * time.Hour)
	nextRetryAt := time.Now().Add(2 * 24 * time.Hour)

	sub := &models.Subscription{
		ID:             subID,
		GracePeriodEnd: &gracePeriodEnd,
	}

	failure := &models.PaymentFailure{
		ID:              uuid.New(),
		SubscriptionID:  subID,
		StripeInvoiceID: invoiceID,
		AmountDue:       999, // $9.99
		Currency:        "usd",
		AttemptCount:    2,
		NextRetryAt:     &nextRetryAt,
	}

	t.Run("payment_failed notification data", func(t *testing.T) {
		data := service.prepareDunningEmailData(sub, failure, models.NotificationTypePaymentFailed)

		assert.Equal(t, subID.String(), data["SubscriptionID"])
		assert.Equal(t, invoiceID, data["InvoiceID"])
		assert.Equal(t, "$9.99", data["AmountDue"])
		assert.Equal(t, "usd", data["Currency"])
		assert.Contains(t, data, "GracePeriodEnd")
		assert.Contains(t, data, "DaysRemaining")
		assert.Equal(t, 2, data["AttemptCount"])
	})

	t.Run("includes next retry time when available", func(t *testing.T) {
		data := service.prepareDunningEmailData(sub, failure, models.NotificationTypePaymentRetry)

		assert.Contains(t, data, "NextRetryAt")
		assert.NotEmpty(t, data["NextRetryAt"])
	})
}

// TestNotificationTypes tests notification type constants
func TestNotificationTypes(t *testing.T) {
	t.Run("notification types are defined", func(t *testing.T) {
		assert.Equal(t, "payment_failed", models.NotificationTypePaymentFailed)
		assert.Equal(t, "payment_retry", models.NotificationTypePaymentRetry)
		assert.Equal(t, "grace_period_warning", models.NotificationTypeGracePeriodWarning)
		assert.Equal(t, "subscription_downgraded", models.NotificationTypeSubscriptionDowngraded)
	})
}

// TestPaymentFailureModel tests payment failure model structure
func TestPaymentFailureModel(t *testing.T) {
	t.Run("creates payment failure record", func(t *testing.T) {
		subID := uuid.New()
		invoiceID := "inv_test123"
		paymentIntentID := "pi_test123"

		failure := &models.PaymentFailure{
			SubscriptionID:        subID,
			StripeInvoiceID:       invoiceID,
			StripePaymentIntentID: &paymentIntentID,
			AmountDue:             999,
			Currency:              "usd",
			AttemptCount:          1,
			Resolved:              false,
		}

		assert.Equal(t, subID, failure.SubscriptionID)
		assert.Equal(t, invoiceID, failure.StripeInvoiceID)
		assert.Equal(t, int64(999), failure.AmountDue)
		assert.Equal(t, "usd", failure.Currency)
		assert.Equal(t, 1, failure.AttemptCount)
		assert.False(t, failure.Resolved)
	})
}

// TestDunningAttemptModel tests dunning attempt model structure
func TestDunningAttemptModel(t *testing.T) {
	t.Run("creates dunning attempt record", func(t *testing.T) {
		failureID := uuid.New()
		userID := uuid.New()

		attempt := &models.DunningAttempt{
			PaymentFailureID: failureID,
			UserID:           userID,
			AttemptNumber:    1,
			NotificationType: models.NotificationTypePaymentFailed,
			EmailSent:        true,
		}

		assert.Equal(t, failureID, attempt.PaymentFailureID)
		assert.Equal(t, userID, attempt.UserID)
		assert.Equal(t, 1, attempt.AttemptNumber)
		assert.Equal(t, models.NotificationTypePaymentFailed, attempt.NotificationType)
		assert.True(t, attempt.EmailSent)
	})
}
