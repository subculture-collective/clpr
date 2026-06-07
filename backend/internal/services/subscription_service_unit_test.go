package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
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

// TestGetOrCreateCustomer tests customer creation with mocked dependencies
func TestGetOrCreateCustomer(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	email := "test@example.com"
	user := &models.User{
		ID:          userID,
		Username:    "testuser",
		Email:       &email,
		DisplayName: "Test User",
	}

	t.Run("returns existing customer ID when subscription exists", func(t *testing.T) {
		mockSubRepo := new(MockSubscriptionRepository)
		mockUserRepo := new(MockUserRepository)
		mockWebhookRepo := new(MockWebhookRepository)
		cfg := &config.Config{
			Stripe: config.StripeConfig{
				SecretKey: "",
			},
		}

		existingCustomerID := "cus_existing123"
		existingSub := &models.Subscription{
			ID:               uuid.New(),
			UserID:           userID,
			StripeCustomerID: existingCustomerID,
			Status:           "active",
			Tier:             "pro",
		}

		mockSubRepo.On("GetByUserID", ctx, userID).Return(existingSub, nil)

		service := newTestSubscriptionService(mockSubRepo, mockUserRepo, mockWebhookRepo, cfg)

		customerID, err := service.GetOrCreateCustomer(ctx, user)

		assert.NoError(t, err)
		assert.Equal(t, existingCustomerID, customerID)
		mockSubRepo.AssertExpectations(t)
	})

	t.Run("handles subscription not found error", func(t *testing.T) {
		mockSubRepo := new(MockSubscriptionRepository)
		mockUserRepo := new(MockUserRepository)
		mockWebhookRepo := new(MockWebhookRepository)
		cfg := &config.Config{
			Stripe: config.StripeConfig{
				SecretKey: "",
			},
		}

		mockSubRepo.On("GetByUserID", ctx, userID).Return(nil, errors.New("not found"))

		service := newTestSubscriptionService(mockSubRepo, mockUserRepo, mockWebhookRepo, cfg)

		// This will attempt to create a Stripe customer, which will fail without real Stripe API
		// In a real test, we'd mock the Stripe client as well
		_, err := service.GetOrCreateCustomer(ctx, user)

		// We expect an error since we're not mocking Stripe API
		assert.Error(t, err)
		mockSubRepo.AssertExpectations(t)
	})
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

// TestHasActiveSubscription tests subscription status checking with mocked repository
func TestHasActiveSubscription(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("returns true for active subscription", func(t *testing.T) {
		mockSubRepo := new(MockSubscriptionRepository)
		mockUserRepo := new(MockUserRepository)
		mockWebhookRepo := new(MockWebhookRepository)
		cfg := &config.Config{}

		sub := &models.Subscription{
			ID:     uuid.New(),
			UserID: userID,
			Status: "active",
			Tier:   "pro",
		}

		mockSubRepo.On("GetByUserID", ctx, userID).Return(sub, nil)

		service := newTestSubscriptionService(mockSubRepo, mockUserRepo, mockWebhookRepo, cfg)

		result := service.HasActiveSubscription(ctx, userID)

		assert.True(t, result)
		mockSubRepo.AssertExpectations(t)
	})

	t.Run("returns true for trialing subscription", func(t *testing.T) {
		mockSubRepo := new(MockSubscriptionRepository)
		mockUserRepo := new(MockUserRepository)
		mockWebhookRepo := new(MockWebhookRepository)
		cfg := &config.Config{}

		sub := &models.Subscription{
			ID:     uuid.New(),
			UserID: userID,
			Status: "trialing",
			Tier:   "pro",
		}

		mockSubRepo.On("GetByUserID", ctx, userID).Return(sub, nil)

		service := NewSubscriptionService(
			mockSubRepo,
			mockUserRepo,
			mockWebhookRepo,
			cfg,
			nil,
			nil,
			nil,
		)

		result := service.HasActiveSubscription(ctx, userID)

		assert.True(t, result)
		mockSubRepo.AssertExpectations(t)
	})

	t.Run("returns true for past_due subscription in grace period", func(t *testing.T) {
		mockSubRepo := new(MockSubscriptionRepository)
		mockUserRepo := new(MockUserRepository)
		mockWebhookRepo := new(MockWebhookRepository)
		cfg := &config.Config{}

		gracePeriodEnd := time.Now().Add(24 * time.Hour)
		sub := &models.Subscription{
			ID:             uuid.New(),
			UserID:         userID,
			Status:         "past_due",
			Tier:           "pro",
			GracePeriodEnd: &gracePeriodEnd,
		}

		mockSubRepo.On("GetByUserID", ctx, userID).Return(sub, nil)

		service := NewSubscriptionService(
			mockSubRepo,
			mockUserRepo,
			mockWebhookRepo,
			cfg,
			nil,
			nil,
			nil,
		)

		result := service.HasActiveSubscription(ctx, userID)

		assert.True(t, result)
		mockSubRepo.AssertExpectations(t)
	})

	t.Run("returns false for past_due subscription outside grace period", func(t *testing.T) {
		mockSubRepo := new(MockSubscriptionRepository)
		mockUserRepo := new(MockUserRepository)
		mockWebhookRepo := new(MockWebhookRepository)
		cfg := &config.Config{}

		gracePeriodEnd := time.Now().Add(-24 * time.Hour)
		sub := &models.Subscription{
			ID:             uuid.New(),
			UserID:         userID,
			Status:         "past_due",
			Tier:           "pro",
			GracePeriodEnd: &gracePeriodEnd,
		}

		mockSubRepo.On("GetByUserID", ctx, userID).Return(sub, nil)

		service := NewSubscriptionService(
			mockSubRepo,
			mockUserRepo,
			mockWebhookRepo,
			cfg,
			nil,
			nil,
			nil,
		)

		result := service.HasActiveSubscription(ctx, userID)

		assert.False(t, result)
		mockSubRepo.AssertExpectations(t)
	})

	t.Run("returns false when subscription not found", func(t *testing.T) {
		mockSubRepo := new(MockSubscriptionRepository)
		mockUserRepo := new(MockUserRepository)
		mockWebhookRepo := new(MockWebhookRepository)
		cfg := &config.Config{}

		mockSubRepo.On("GetByUserID", ctx, userID).Return(nil, errors.New("not found"))

		service := NewSubscriptionService(
			mockSubRepo,
			mockUserRepo,
			mockWebhookRepo,
			cfg,
			nil,
			nil,
			nil,
		)

		result := service.HasActiveSubscription(ctx, userID)

		assert.False(t, result)
		mockSubRepo.AssertExpectations(t)
	})

	t.Run("returns false for canceled subscription", func(t *testing.T) {
		mockSubRepo := new(MockSubscriptionRepository)
		mockUserRepo := new(MockUserRepository)
		mockWebhookRepo := new(MockWebhookRepository)
		cfg := &config.Config{}

		sub := &models.Subscription{
			ID:     uuid.New(),
			UserID: userID,
			Status: "canceled",
			Tier:   "free",
		}

		mockSubRepo.On("GetByUserID", ctx, userID).Return(sub, nil)

		service := NewSubscriptionService(
			mockSubRepo,
			mockUserRepo,
			mockWebhookRepo,
			cfg,
			nil,
			nil,
			nil,
		)

		result := service.HasActiveSubscription(ctx, userID)

		assert.False(t, result)
		mockSubRepo.AssertExpectations(t)
	})
}

// TestIsProUser tests Pro user checking with mocked repository
func TestIsProUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("returns true for active Pro subscription", func(t *testing.T) {
		mockSubRepo := new(MockSubscriptionRepository)
		mockUserRepo := new(MockUserRepository)
		mockWebhookRepo := new(MockWebhookRepository)
		cfg := &config.Config{}

		sub := &models.Subscription{
			ID:     uuid.New(),
			UserID: userID,
			Status: "active",
			Tier:   "pro",
		}

		mockSubRepo.On("GetByUserID", ctx, userID).Return(sub, nil)

		service := newTestSubscriptionService(mockSubRepo, mockUserRepo, mockWebhookRepo, cfg)

		result := service.IsProUser(ctx, userID)

		assert.True(t, result)
		mockSubRepo.AssertExpectations(t)
	})

	t.Run("returns false for active free subscription", func(t *testing.T) {
		mockSubRepo := new(MockSubscriptionRepository)
		mockUserRepo := new(MockUserRepository)
		mockWebhookRepo := new(MockWebhookRepository)
		cfg := &config.Config{}

		sub := &models.Subscription{
			ID:     uuid.New(),
			UserID: userID,
			Status: "active",
			Tier:   "free",
		}

		mockSubRepo.On("GetByUserID", ctx, userID).Return(sub, nil)

		service := NewSubscriptionService(
			mockSubRepo,
			mockUserRepo,
			mockWebhookRepo,
			cfg,
			nil,
			nil,
			nil,
		)

		result := service.IsProUser(ctx, userID)

		assert.False(t, result)
		mockSubRepo.AssertExpectations(t)
	})

	t.Run("returns false for canceled Pro subscription", func(t *testing.T) {
		mockSubRepo := new(MockSubscriptionRepository)
		mockUserRepo := new(MockUserRepository)
		mockWebhookRepo := new(MockWebhookRepository)
		cfg := &config.Config{}

		sub := &models.Subscription{
			ID:     uuid.New(),
			UserID: userID,
			Status: "canceled",
			Tier:   "pro",
		}

		mockSubRepo.On("GetByUserID", ctx, userID).Return(sub, nil)

		service := NewSubscriptionService(
			mockSubRepo,
			mockUserRepo,
			mockWebhookRepo,
			cfg,
			nil,
			nil,
			nil,
		)

		result := service.IsProUser(ctx, userID)

		assert.False(t, result)
		mockSubRepo.AssertExpectations(t)
	})
}

// TestCreateCheckoutSessionWithMocks tests checkout session creation with mock config
func TestCreateCheckoutSessionWithMocks(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	email := "test@example.com"
	user := &models.User{
		ID:       userID,
		Username: "testuser",
		Email:    &email,
	}

	t.Run("returns mock session when Stripe not configured", func(t *testing.T) {
		mockSubRepo := new(MockSubscriptionRepository)
		mockUserRepo := new(MockUserRepository)
		mockWebhookRepo := new(MockWebhookRepository)
		cfg := &config.Config{
			Stripe: config.StripeConfig{
				SecretKey:  "",
				SuccessURL: "http://localhost:5173/success",
			},
			FeatureFlags: config.FeatureFlagsConfig{
				PremiumSubscriptions: true,
			},
		}

		service := NewSubscriptionService(
			mockSubRepo,
			mockUserRepo,
			mockWebhookRepo,
			cfg,
			nil,
			nil,
			nil,
		)

		priceID := "price_test_monthly"
		result, err := service.CreateCheckoutSession(ctx, user, priceID, nil)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "cs_test_mock", result.SessionID)
		assert.Contains(t, result.SessionURL, "success")
	})

	t.Run("returns mock session when premium feature flag is off", func(t *testing.T) {
		mockSubRepo := new(MockSubscriptionRepository)
		mockUserRepo := new(MockUserRepository)
		mockWebhookRepo := new(MockWebhookRepository)
		cfg := &config.Config{
			Stripe: config.StripeConfig{
				SecretKey:  "sk_test_123",
				SuccessURL: "http://localhost:5173/success",
			},
			FeatureFlags: config.FeatureFlagsConfig{
				PremiumSubscriptions: false,
			},
		}

		service := NewSubscriptionService(
			mockSubRepo,
			mockUserRepo,
			mockWebhookRepo,
			cfg,
			nil,
			nil,
			nil,
		)

		priceID := "price_test_monthly"
		result, err := service.CreateCheckoutSession(ctx, user, priceID, nil)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "cs_test_mock", result.SessionID)
	})
}

// TestCreatePortalSessionWithMocks tests portal session creation with mock config
func TestCreatePortalSessionWithMocks(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	email := "test@example.com"
	user := &models.User{
		ID:       userID,
		Username: "testuser",
		Email:    &email,
	}

	t.Run("returns mock portal URL when Stripe not configured", func(t *testing.T) {
		mockSubRepo := new(MockSubscriptionRepository)
		mockUserRepo := new(MockUserRepository)
		mockWebhookRepo := new(MockWebhookRepository)
		cfg := &config.Config{
			Stripe: config.StripeConfig{
				SecretKey:  "",
				SuccessURL: "http://localhost:5173/subscription",
			},
			FeatureFlags: config.FeatureFlagsConfig{
				PremiumSubscriptions: true,
			},
		}

		service := NewSubscriptionService(
			mockSubRepo,
			mockUserRepo,
			mockWebhookRepo,
			cfg,
			nil,
			nil,
			nil,
		)

		result, err := service.CreatePortalSession(ctx, user)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "http://localhost:5173/subscription", result.PortalURL)
	})

	t.Run("returns error when subscription not found", func(t *testing.T) {
		mockSubRepo := new(MockSubscriptionRepository)
		mockUserRepo := new(MockUserRepository)
		mockWebhookRepo := new(MockWebhookRepository)
		cfg := &config.Config{
			Stripe: config.StripeConfig{
				SecretKey: "sk_test_123",
			},
			FeatureFlags: config.FeatureFlagsConfig{
				PremiumSubscriptions: true,
			},
		}

		mockSubRepo.On("GetByUserID", ctx, userID).Return(nil, errors.New("not found"))

		service := NewSubscriptionService(
			mockSubRepo,
			mockUserRepo,
			mockWebhookRepo,
			cfg,
			nil,
			nil,
			nil,
		)

		result, err := service.CreatePortalSession(ctx, user)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrSubscriptionNotFound, err)
		mockSubRepo.AssertExpectations(t)
	})

	t.Run("returns error when customer ID is missing", func(t *testing.T) {
		mockSubRepo := new(MockSubscriptionRepository)
		mockUserRepo := new(MockUserRepository)
		mockWebhookRepo := new(MockWebhookRepository)
		cfg := &config.Config{
			Stripe: config.StripeConfig{
				SecretKey: "sk_test_123",
			},
			FeatureFlags: config.FeatureFlagsConfig{
				PremiumSubscriptions: true,
			},
		}

		sub := &models.Subscription{
			ID:               uuid.New(),
			UserID:           userID,
			StripeCustomerID: "",
			Status:           "active",
		}

		mockSubRepo.On("GetByUserID", ctx, userID).Return(sub, nil)

		service := NewSubscriptionService(
			mockSubRepo,
			mockUserRepo,
			mockWebhookRepo,
			cfg,
			nil,
			nil,
			nil,
		)

		result, err := service.CreatePortalSession(ctx, user)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrStripeCustomerNotFound, err)
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
