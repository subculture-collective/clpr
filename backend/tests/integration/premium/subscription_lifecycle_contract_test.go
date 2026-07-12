//go:build integration

package premium

import (
	"context"
	"fmt"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v81/webhook"
)

func TestSignedStripeWebhookLifecyclePersistsEntitlementsIdempotently(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	t.Cleanup(pool.Close)
	ctx := context.Background()
	userID := uuid.New()
	twitchID := userID.String()
	_, err := pool.Exec(ctx, `INSERT INTO users (id, twitch_id, username, display_name, email) VALUES ($1, $2, $3, $3, $4)`,
		userID, twitchID, "premium-contract", "premium-contract@example.com")
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", userID) })

	repo := repository.NewSubscriptionRepository(pool)
	customerID := "cus_" + userID.String()
	subscriptionID := "sub_" + userID.String()
	require.NoError(t, repo.Create(ctx, &models.Subscription{
		UserID: userID, StripeCustomerID: customerID, Status: "inactive", Tier: "free",
	}))

	const secret = "whsec_subscription_contract"
	cfg := &config.Config{Stripe: config.StripeConfig{
		SecretKey: "sk_test_123", WebhookSecrets: []string{secret},
		ProMonthlyPriceID: "price_monthly_contract", ProYearlyPriceID: "price_yearly_contract",
	}}
	service := services.NewSubscriptionService(repo, repository.NewUserRepository(pool), repository.NewWebhookRepository(pool), cfg, nil, nil, nil)
	baseTime := time.Now().Unix()

	created := subscriptionEvent("evt_created_"+userID.String(), "customer.subscription.created", baseTime,
		subscriptionID, customerID, "active", "price_monthly_contract")
	require.NoError(t, deliverSigned(service, secret, created))
	assertSubscription(t, ctx, repo, service, userID, "active", "pro", true)

	// A completed receipt makes replay a no-op and the unique event log remains singular.
	require.NoError(t, deliverSigned(service, secret, created))
	var createdEvents int
	require.NoError(t, pool.QueryRow(ctx, "SELECT COUNT(*) FROM subscription_events WHERE stripe_event_id = $1", "evt_created_"+userID.String()).Scan(&createdEvents))
	require.Equal(t, 1, createdEvents)

	// A unique but older update must not revoke newer entitlement state.
	stale := subscriptionEvent("evt_stale_"+userID.String(), "customer.subscription.updated", baseTime-60,
		subscriptionID, customerID, "canceled", "price_monthly_contract")
	require.NoError(t, deliverSigned(service, secret, stale))
	assertSubscription(t, ctx, repo, service, userID, "active", "pro", true)

	failedInvoice := fmt.Sprintf(`{"id":"evt_failed_%s","object":"event","api_version":"2025-02-24.acacia","type":"invoice.payment_failed","created":%d,"data":{"object":{"id":"in_failed_%s","object":"invoice","subscription":"%s","customer":"%s","amount_due":999,"currency":"usd"}}}`,
		userID, baseTime+60, userID, subscriptionID, customerID)
	require.NoError(t, deliverSigned(service, secret, []byte(failedInvoice)))
	assertSubscription(t, ctx, repo, service, userID, "past_due", "pro", false)

	deleted := subscriptionEvent("evt_deleted_"+userID.String(), "customer.subscription.deleted", baseTime+120,
		subscriptionID, customerID, "canceled", "price_monthly_contract")
	require.NoError(t, deliverSigned(service, secret, deleted))
	assertSubscription(t, ctx, repo, service, userID, "canceled", "free", false)

	require.Error(t, service.HandleWebhook(ctx, created, "t=1,v1=invalid"))
	assertSubscription(t, ctx, repo, service, userID, "canceled", "free", false)
}

func subscriptionEvent(eventID, eventType string, created int64, subscriptionID, customerID, status, priceID string) []byte {
	return []byte(fmt.Sprintf(`{"id":"%s","object":"event","api_version":"2025-02-24.acacia","type":"%s","created":%d,"data":{"object":{"id":"%s","object":"subscription","customer":"%s","status":"%s","items":{"data":[{"price":{"id":"%s","object":"price"}}]},"current_period_start":%d,"current_period_end":%d}}}`,
		eventID, eventType, created, subscriptionID, customerID, status, priceID, created, created+2592000))
}

func deliverSigned(service *services.SubscriptionService, secret string, payload []byte) error {
	signed := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{Payload: payload, Secret: secret})
	return service.HandleWebhook(context.Background(), payload, signed.Header)
}

func assertSubscription(t *testing.T, ctx context.Context, repo *repository.SubscriptionRepository, service *services.SubscriptionService, userID uuid.UUID, status, tier string, entitled bool) {
	t.Helper()
	subscription, err := repo.GetByUserID(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, status, subscription.Status)
	require.Equal(t, tier, subscription.Tier)
	require.Equal(t, entitled, service.IsProUser(ctx, userID))
}
