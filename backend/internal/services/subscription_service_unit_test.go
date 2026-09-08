package services

import (
	"context"
	"errors"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// newTestSubscriptionService creates a subscription service with mock dependencies for testing
func newTestSubscriptionService(
	subRepo *MockSubscriptionRepository,
	userRepo *MockUserRepository,
	webhookRepo *MockWebhookRepository,
	cfg *config.Config,
) *SubscriptionService {
	return NewSubscriptionService(
		subRepo,
		userRepo,
		webhookRepo,
		cfg,
		nil, // auditLogSvc - can be mocked if needed
		nil, // dunningService - can be mocked if needed
		nil, // emailService - can be mocked if needed
	)
}

// TestNewSubscriptionService tests service creation with dependency injection
func TestNewSubscriptionService(t *testing.T) {
	t.Run("creates service with all dependencies", func(t *testing.T) {
		mockSubRepo := new(MockSubscriptionRepository)
		mockUserRepo := new(MockUserRepository)
		mockWebhookRepo := new(MockWebhookRepository)
		cfg := &config.Config{
			Stripe: config.StripeConfig{
				SecretKey: "",
			},
		}

		service := newTestSubscriptionService(mockSubRepo, mockUserRepo, mockWebhookRepo, cfg)

		assert.NotNil(t, service)
	})

	t.Run("creates service with empty config", func(t *testing.T) {
		mockSubRepo := new(MockSubscriptionRepository)
		mockUserRepo := new(MockUserRepository)
		mockWebhookRepo := new(MockWebhookRepository)
		cfg := &config.Config{}

		service := newTestSubscriptionService(mockSubRepo, mockUserRepo, mockWebhookRepo, cfg)

		assert.NotNil(t, service)
	})
}

func TestSubscriptionOutboundOperationsFailClosedWhenDisabled(t *testing.T) {
	service := newTestSubscriptionService(new(MockSubscriptionRepository), new(MockUserRepository), new(MockWebhookRepository), &config.Config{})
	ctx := context.Background()
	user := &models.User{ID: uuid.New()}
	tests := []struct {
		name string
		call func() error
	}{
		{"cancel", func() error { return service.CancelSubscription(ctx, user, false) }},
		{"invoices", func() error { _, err := service.GetInvoices(ctx, user, 10); return err }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { assert.ErrorIs(t, tt.call(), ErrSubscriptionsUnavailable) })
	}
}

// TestGetSubscriptionByUserID tests subscription retrieval with mocked repository
func TestGetSubscriptionByUserID(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("returns subscription when found", func(t *testing.T) {
		mockSubRepo := new(MockSubscriptionRepository)
		mockUserRepo := new(MockUserRepository)
		mockWebhookRepo := new(MockWebhookRepository)
		cfg := &config.Config{}

		expectedSub := &models.Subscription{
			ID:               uuid.New(),
			UserID:           userID,
			StripeCustomerID: "cus_test123",
			Status:           "active",
			Tier:             "pro",
		}

		mockSubRepo.On("GetByUserID", ctx, userID).Return(expectedSub, nil)

		service := newTestSubscriptionService(mockSubRepo, mockUserRepo, mockWebhookRepo, cfg)

		result, err := service.GetSubscriptionByUserID(ctx, userID)

		assert.NoError(t, err)
		assert.Equal(t, expectedSub, result)
		mockSubRepo.AssertExpectations(t)
	})

	t.Run("returns error when subscription not found", func(t *testing.T) {
		mockSubRepo := new(MockSubscriptionRepository)
		mockUserRepo := new(MockUserRepository)
		mockWebhookRepo := new(MockWebhookRepository)
		cfg := &config.Config{}

		expectedError := errors.New("subscription not found")
		mockSubRepo.On("GetByUserID", ctx, userID).Return(nil, expectedError)

		service := newTestSubscriptionService(mockSubRepo, mockUserRepo, mockWebhookRepo, cfg)

		result, err := service.GetSubscriptionByUserID(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
		mockSubRepo.AssertExpectations(t)
	})
}

// TestFormatAmountForCurrency tests currency amount formatting (kept from original tests)
func TestFormatAmountForCurrency(t *testing.T) {
	t.Run("formats USD amount correctly", func(t *testing.T) {
		result := formatAmountForCurrency(1999, "usd")
		assert.Equal(t, "19.99 USD", result)
	})

	t.Run("formats EUR amount correctly", func(t *testing.T) {
		result := formatAmountForCurrency(2500, "eur")
		assert.Equal(t, "25.00 EUR", result)
	})

	t.Run("handles zero amount", func(t *testing.T) {
		result := formatAmountForCurrency(0, "usd")
		assert.Equal(t, "0.00 USD", result)
	})

	t.Run("formats JPY (zero-decimal currency) correctly", func(t *testing.T) {
		result := formatAmountForCurrency(1000, "jpy")
		assert.Equal(t, "1000 JPY", result)
	})

	t.Run("formats KRW (zero-decimal currency) correctly", func(t *testing.T) {
		result := formatAmountForCurrency(50000, "krw")
		assert.Equal(t, "50000 KRW", result)
	})

	t.Run("formats KWD (three-decimal currency) correctly", func(t *testing.T) {
		result := formatAmountForCurrency(1500, "kwd")
		assert.Equal(t, "1.500 KWD", result)
	})

	t.Run("handles case insensitivity", func(t *testing.T) {
		result1 := formatAmountForCurrency(1000, "USD")
		result2 := formatAmountForCurrency(1000, "usd")
		assert.Equal(t, result1, result2)
	})
}
