//go:build stripe_mock

package services

import (
	"context"
	"errors"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	_ "github.com/stripe/stripe-go/v81/testing"
)

// TestCreateCheckoutSessionAgainstStripeMock crosses the real stripe-go SDK
// boundary while keeping the test incapable of creating a real charge. The
// stripe-go testing package points all SDK backends at the official
// stripe-mock service exposed through STRIPE_MOCK_PORT.
func TestCreateCheckoutSessionAgainstStripeMock(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	email := "checkout-contract@example.com"
	user := &models.User{
		ID:          userID,
		Username:    "checkout-contract",
		DisplayName: "Checkout Contract",
		Email:       &email,
	}
	repo := new(MockSubscriptionRepository)
	repo.On("GetByUserID", mock.Anything, userID).Return(nil, errors.New("not found")).Once()
	repo.On("Create", mock.Anything, mock.MatchedBy(func(sub *models.Subscription) bool {
		return sub != nil && sub.UserID == userID && sub.StripeCustomerID != "" && sub.Status == "inactive" && sub.Tier == "free"
	})).Return(nil).Once()

	cfg := &config.Config{
		Stripe: config.StripeConfig{
			SecretKey:         "sk_test_123",
			ProMonthlyPriceID: "price_monthly_contract",
			ProYearlyPriceID:  "price_yearly_contract",
			SuccessURL:        "https://app.example.test/subscription/success",
			CancelURL:         "https://app.example.test/subscription",
		},
		FeatureFlags: config.FeatureFlagsConfig{PremiumSubscriptions: true},
	}

	service := NewSubscriptionService(repo, nil, nil, cfg, nil, nil, nil)
	result, err := service.CreateCheckoutSession(ctx, user, cfg.Stripe.ProMonthlyPriceID, nil)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Regexp(t, `^cs_`, result.SessionID)
	assert.NotEmpty(t, result.SessionURL)
	repo.AssertExpectations(t)
}

func TestCreateCheckoutSessionAgainstStripeMockWithoutEmail(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	user := &models.User{ID: userID, Username: "no-email-contract"}
	repo := new(MockSubscriptionRepository)
	repo.On("GetByUserID", mock.Anything, userID).Return(nil, errors.New("not found")).Once()
	repo.On("Create", mock.Anything, mock.MatchedBy(func(sub *models.Subscription) bool {
		return sub != nil && sub.StripeCustomerID != ""
	})).Return(nil).Once()
	cfg := &config.Config{
		Stripe: config.StripeConfig{
			SecretKey:         "sk_test_123",
			ProMonthlyPriceID: "price_monthly_contract",
			ProYearlyPriceID:  "price_yearly_contract",
			SuccessURL:        "https://app.example.test/subscription/success",
			CancelURL:         "https://app.example.test/subscription",
		},
		FeatureFlags: config.FeatureFlagsConfig{PremiumSubscriptions: true},
	}

	result, err := NewSubscriptionService(repo, nil, nil, cfg, nil, nil, nil).
		CreateCheckoutSession(ctx, user, cfg.Stripe.ProMonthlyPriceID, nil)

	require.NoError(t, err)
	assert.Regexp(t, `^cs_`, result.SessionID)
	repo.AssertExpectations(t)
}
