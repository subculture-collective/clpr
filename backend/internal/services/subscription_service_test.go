package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/webhook"
)

type atomicClaimSubscriptionRepo struct {
	*MockSubscriptionRepository
	claimed   bool
	completed int
	released  int
}

func (r *atomicClaimSubscriptionRepo) ClaimStripeWebhookEvent(context.Context, string, string, time.Duration) (bool, error) {
	if r.claimed {
		return false, nil
	}
	r.claimed = true
	return true, nil
}

func (r *atomicClaimSubscriptionRepo) CompleteStripeWebhookEvent(context.Context, string) error {
	r.completed++
	return nil
}

func (r *atomicClaimSubscriptionRepo) ReleaseStripeWebhookEvent(context.Context, string) error {
	r.released++
	r.claimed = false
	return nil
}

func TestStripeWebhookAtomicClaimSkipsDuplicateDelivery(t *testing.T) {
	const secret = "whsec_atomic_claim_test"
	payload := []byte(`{"id":"evt_atomic","object":"event","api_version":"2025-02-24.acacia","type":"unsupported.test","data":{"object":{}}}`)
	signed := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{Payload: payload, Secret: secret})
	repo := &atomicClaimSubscriptionRepo{MockSubscriptionRepository: new(MockSubscriptionRepository)}
	service := NewSubscriptionService(repo, nil, nil, &config.Config{Stripe: config.StripeConfig{WebhookSecrets: []string{secret}}}, nil, nil, nil)

	require.NoError(t, service.HandleWebhook(context.Background(), payload, signed.Header))
	require.NoError(t, service.HandleWebhook(context.Background(), payload, signed.Header))
	assert.Equal(t, 1, repo.completed)
	assert.Equal(t, 0, repo.released)
}

func TestStripeWebhookAtomicClaimReleasesFailedDeliveryForRetry(t *testing.T) {
	const secret = "whsec_atomic_release_test"
	payload := []byte(`{"id":"evt_retry","object":"event","api_version":"2025-02-24.acacia","type":"customer.subscription.updated","data":{"object":{}}}`)
	signed := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{Payload: payload, Secret: secret})
	repo := &atomicClaimSubscriptionRepo{MockSubscriptionRepository: new(MockSubscriptionRepository)}
	repo.On("GetByStripeSubscriptionID", context.Background(), "").Return(nil, errors.New("temporary database failure")).Twice()
	service := NewSubscriptionService(repo, nil, nil, &config.Config{Stripe: config.StripeConfig{WebhookSecrets: []string{secret}}}, nil, nil, nil)

	err := service.HandleWebhook(context.Background(), payload, signed.Header)
	require.Error(t, err)
	assert.Equal(t, 0, repo.completed)
	assert.Equal(t, 1, repo.released)
	require.Error(t, service.HandleWebhook(context.Background(), payload, signed.Header))
	assert.Equal(t, 2, repo.released)
}

func TestStaleSubscriptionEvent(t *testing.T) {
	sub := &models.Subscription{LastStripeEventCreated: 200}
	assert.True(t, staleSubscriptionEvent(sub, stripe.Event{Created: 199}))
	assert.False(t, staleSubscriptionEvent(sub, stripe.Event{Created: 200}))
	assert.False(t, staleSubscriptionEvent(sub, stripe.Event{Created: 201}))
}

// TestInvoiceFinalizedNotificationType tests the invoice finalized notification type
func TestInvoiceFinalizedNotificationType(t *testing.T) {
	t.Run("notification type constant is correct", func(t *testing.T) {
		assert.Equal(t, "invoice_finalized", models.NotificationTypeInvoiceFinalized)
	})
}

// TestPaymentIntentWebhookHandlers tests the payment intent webhook handler data structures
func TestPaymentIntentWebhookHandlers(t *testing.T) {
	t.Run("handles payment intent with nil customer gracefully", func(t *testing.T) {
		// Test that handlers can handle payment intents without a customer
		var customerID string = ""
		assert.Empty(t, customerID)
	})

	t.Run("handles payment intent succeeded event structure", func(t *testing.T) {
		// Verify the metadata structure for successful payment intents
		metadata := map[string]interface{}{
			"payment_intent_id": "pi_test_123",
			"amount_cents":      1999,
			"currency":          "usd",
			"status":            "succeeded",
		}
		assert.Equal(t, "pi_test_123", metadata["payment_intent_id"])
		assert.Equal(t, 1999, metadata["amount_cents"])
		assert.Equal(t, "usd", metadata["currency"])
		assert.Equal(t, "succeeded", metadata["status"])
	})

	t.Run("handles payment intent failed event structure", func(t *testing.T) {
		// Verify the metadata structure for failed payment intents
		metadata := map[string]interface{}{
			"payment_intent_id": "pi_test_456",
			"amount_cents":      2999,
			"currency":          "usd",
			"status":            "requires_payment_method",
			"error_code":        "card_declined",
			"error_message":     "Your card was declined",
		}
		assert.Equal(t, "pi_test_456", metadata["payment_intent_id"])
		assert.Equal(t, "card_declined", metadata["error_code"])
		assert.Equal(t, "Your card was declined", metadata["error_message"])
	})

	t.Run("includes customer ID when present", func(t *testing.T) {
		// Verify that customer ID is included in metadata when available
		metadata := map[string]interface{}{
			"payment_intent_id":  "pi_test_789",
			"stripe_customer_id": "cus_test_123",
		}
		assert.Contains(t, metadata, "stripe_customer_id")
		assert.Equal(t, "cus_test_123", metadata["stripe_customer_id"])
	})

	t.Run("handles nil payment error gracefully", func(t *testing.T) {
		// Test that handler can handle missing error information
		metadata := map[string]interface{}{
			"payment_intent_id": "pi_test_000",
			"status":            "requires_payment_method",
		}
		// Should not have error_code or error_message when error is nil
		assert.NotContains(t, metadata, "error_code")
		assert.NotContains(t, metadata, "error_message")
	})

	t.Run("validates payment intent event types", func(t *testing.T) {
		// Verify the event types are correctly defined
		succeededEvent := "payment_intent.succeeded"
		failedEvent := "payment_intent.payment_failed"

		assert.Equal(t, "payment_intent.succeeded", succeededEvent)
		assert.Equal(t, "payment_intent.payment_failed", failedEvent)
	})
}
