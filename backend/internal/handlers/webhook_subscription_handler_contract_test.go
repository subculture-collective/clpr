package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type webhookSubscriptionServiceStub struct {
	subscription  *models.WebhookSubscription
	subscriptions []*models.WebhookSubscription
	deliveries    []*models.WebhookDelivery
	total         int
	secret        string
	err           error
}

func (s *webhookSubscriptionServiceStub) CreateSubscription(context.Context, uuid.UUID, *models.CreateWebhookSubscriptionRequest) (*models.WebhookSubscription, error) {
	return s.subscription, s.err
}
func (s *webhookSubscriptionServiceStub) GetSubscriptionsByUserID(context.Context, uuid.UUID) ([]*models.WebhookSubscription, error) {
	return s.subscriptions, s.err
}
func (s *webhookSubscriptionServiceStub) GetSubscriptionByID(context.Context, uuid.UUID, uuid.UUID) (*models.WebhookSubscription, error) {
	return s.subscription, s.err
}
func (s *webhookSubscriptionServiceStub) UpdateSubscription(context.Context, uuid.UUID, uuid.UUID, *models.UpdateWebhookSubscriptionRequest) error {
	return s.err
}
func (s *webhookSubscriptionServiceStub) DeleteSubscription(context.Context, uuid.UUID, uuid.UUID) error {
	return s.err
}
func (s *webhookSubscriptionServiceStub) GetDeliveriesBySubscriptionID(context.Context, uuid.UUID, uuid.UUID, int, int) ([]*models.WebhookDelivery, int, error) {
	return s.deliveries, s.total, s.err
}
func (s *webhookSubscriptionServiceStub) RegenerateSecret(context.Context, uuid.UUID, uuid.UUID) (string, error) {
	return s.secret, s.err
}

func webhookTestContext(method, target string, body *strings.Reader, userID uuid.UUID) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	if body == nil {
		ctx.Request = httptest.NewRequest(method, target, nil)
	} else {
		ctx.Request = httptest.NewRequest(method, target, body)
		ctx.Request.Header.Set("Content-Type", "application/json")
	}
	ctx.Set("user_id", userID)
	return ctx, recorder
}

func TestWebhookSubscriptionResponseContracts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	subscriptionID := uuid.New()
	now := time.Now().UTC()
	subscription := &models.WebhookSubscription{
		ID: subscriptionID, UserID: userID, URL: "https://example.com/hooks/clpr",
		Secret: "create-only-secret", Events: []string{models.WebhookEventClipApproved},
		IsActive: true, CreatedAt: now, UpdatedAt: now,
	}
	delivery := &models.WebhookDelivery{
		ID: uuid.New(), SubscriptionID: subscriptionID, EventType: models.WebhookEventClipApproved,
		EventID: uuid.New(), Payload: `{}`, Status: "delivered", AttemptCount: 1,
		MaxAttempts: 5, CreatedAt: now, UpdatedAt: now,
	}
	service := &webhookSubscriptionServiceStub{
		subscription: subscription, subscriptions: []*models.WebhookSubscription{subscription},
		deliveries: []*models.WebhookDelivery{delivery}, total: 1, secret: "regenerated-secret",
	}
	handler := NewWebhookSubscriptionHandler(service)

	t.Run("create", func(t *testing.T) {
		ctx, recorder := webhookTestContext(http.MethodPost, "/api/v1/webhooks", strings.NewReader(`{"url":"https://example.com/hooks/clpr","events":["clip.approved"]}`), userID)
		handler.CreateSubscription(ctx)
		require.Equal(t, http.StatusCreated, recorder.Code)
		var body struct {
			Success bool                        `json:"success"`
			Data    *models.WebhookSubscription `json:"data"`
			Secret  string                      `json:"secret"`
			Message string                      `json:"message"`
		}
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
		require.True(t, body.Success)
		require.Equal(t, subscriptionID, body.Data.ID)
		require.Equal(t, "create-only-secret", body.Secret)
		var rawBody struct {
			Data map[string]any `json:"data"`
		}
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &rawBody))
		require.NotContains(t, rawBody.Data, "secret", "secret must not leak inside the subscription object")
	})

	t.Run("list and get", func(t *testing.T) {
		ctx, recorder := webhookTestContext(http.MethodGet, "/api/v1/webhooks", nil, userID)
		handler.ListSubscriptions(ctx)
		require.Equal(t, http.StatusOK, recorder.Code)
		require.Contains(t, recorder.Body.String(), subscriptionID.String())

		ctx, recorder = webhookTestContext(http.MethodGet, "/api/v1/webhooks/"+subscriptionID.String(), nil, userID)
		ctx.Params = gin.Params{{Key: "id", Value: subscriptionID.String()}}
		handler.GetSubscription(ctx)
		require.Equal(t, http.StatusOK, recorder.Code)
		require.NotContains(t, recorder.Body.String(), "create-only-secret")
	})

	t.Run("update and delete", func(t *testing.T) {
		ctx, recorder := webhookTestContext(http.MethodPatch, "/api/v1/webhooks/"+subscriptionID.String(), strings.NewReader(`{"is_active":false}`), userID)
		ctx.Params = gin.Params{{Key: "id", Value: subscriptionID.String()}}
		handler.UpdateSubscription(ctx)
		require.Equal(t, http.StatusOK, recorder.Code)

		ctx, recorder = webhookTestContext(http.MethodDelete, "/api/v1/webhooks/"+subscriptionID.String(), nil, userID)
		ctx.Params = gin.Params{{Key: "id", Value: subscriptionID.String()}}
		handler.DeleteSubscription(ctx)
		require.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run("deliveries", func(t *testing.T) {
		ctx, recorder := webhookTestContext(http.MethodGet, "/api/v1/webhooks/"+subscriptionID.String()+"/deliveries?page=2&limit=10", nil, userID)
		ctx.Params = gin.Params{{Key: "id", Value: subscriptionID.String()}}
		handler.GetSubscriptionDeliveries(ctx)
		require.Equal(t, http.StatusOK, recorder.Code)
		var body struct {
			Success bool                      `json:"success"`
			Data    []*models.WebhookDelivery `json:"data"`
			Meta    struct {
				Page, Limit, Total, TotalPages int
			} `json:"meta"`
		}
		require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
		require.True(t, body.Success)
		require.Len(t, body.Data, 1)
		require.Equal(t, 2, body.Meta.Page)
		require.Equal(t, 10, body.Meta.Limit)
	})

	t.Run("events and regenerated secret", func(t *testing.T) {
		ctx, recorder := webhookTestContext(http.MethodGet, "/api/v1/webhooks/events", nil, userID)
		handler.GetSupportedEvents(ctx)
		require.Equal(t, http.StatusOK, recorder.Code)
		require.Contains(t, recorder.Body.String(), models.WebhookEventClipApproved)

		ctx, recorder = webhookTestContext(http.MethodPost, "/api/v1/webhooks/"+subscriptionID.String()+"/regenerate-secret", nil, userID)
		ctx.Params = gin.Params{{Key: "id", Value: subscriptionID.String()}}
		handler.RegenerateSecret(ctx)
		require.Equal(t, http.StatusOK, recorder.Code)
		require.Contains(t, recorder.Body.String(), "regenerated-secret")
	})
}

func TestWebhookSubscriptionRejectsMalformedIdentityWithoutPanicking(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewWebhookSubscriptionHandler(&webhookSubscriptionServiceStub{})
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/webhooks", nil)
	ctx.Set("user_id", "not-a-uuid")
	require.NotPanics(t, func() { handler.ListSubscriptions(ctx) })
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
}
