package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type webhookDLQServiceStub struct {
	items []*models.OutboundWebhookDeadLetterQueue
	total int
	err   error
}

func (s *webhookDLQServiceStub) GetDeadLetterQueueItems(context.Context, int, int) ([]*models.OutboundWebhookDeadLetterQueue, int, error) {
	return s.items, s.total, s.err
}
func (s *webhookDLQServiceStub) ReplayDeadLetterQueueItem(context.Context, uuid.UUID) error {
	return s.err
}
func (s *webhookDLQServiceStub) DeleteDeadLetterQueueItem(context.Context, uuid.UUID) error {
	return s.err
}

func TestWebhookDLQHandlerResponseContracts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	id := uuid.New()
	item := &models.OutboundWebhookDeadLetterQueue{
		ID: id, SubscriptionID: uuid.New(), DeliveryID: uuid.New(), EventType: "clip.approved",
		EventID: uuid.New(), Payload: `{}`, ErrorMessage: "timeout", AttemptCount: 5,
		OriginalCreatedAt: time.Now(), MovedToDLQAt: time.Now(), CreatedAt: time.Now(),
	}

	t.Run("list", func(t *testing.T) {
		handler := NewWebhookDLQHandler(&webhookDLQServiceStub{items: []*models.OutboundWebhookDeadLetterQueue{item}, total: 1})
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/webhooks/dlq?page=1&limit=20", nil)
		handler.GetDeadLetterQueue(ctx)
		if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), id.String()) {
			t.Fatalf("unexpected list response: %d %s", recorder.Code, recorder.Body.String())
		}
	})

	for _, operation := range []struct {
		name, method, suffix string
		handle               func(*WebhookDLQHandler, *gin.Context)
	}{
		{"replay", http.MethodPost, "/replay", func(h *WebhookDLQHandler, c *gin.Context) { h.ReplayDeadLetterQueueItem(c) }},
		{"delete", http.MethodDelete, "", func(h *WebhookDLQHandler, c *gin.Context) { h.DeleteDeadLetterQueueItem(c) }},
	} {
		t.Run(operation.name, func(t *testing.T) {
			handler := NewWebhookDLQHandler(&webhookDLQServiceStub{})
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Params = gin.Params{{Key: "id", Value: id.String()}}
			ctx.Request = httptest.NewRequest(operation.method, "/api/v1/admin/webhooks/dlq/"+id.String()+operation.suffix, nil)
			operation.handle(handler, ctx)
			if recorder.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
			}
		})
	}
}

func TestWebhookDLQHandlerFailClosedStatuses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	id := uuid.NewString()
	tests := []struct {
		name   string
		err    error
		method string
		handle func(*WebhookDLQHandler, *gin.Context)
		want   int
	}{
		{"missing replay", services.ErrWebhookDLQItemNotFound, http.MethodPost, func(h *WebhookDLQHandler, c *gin.Context) { h.ReplayDeadLetterQueueItem(c) }, http.StatusNotFound},
		{"concurrent replay", services.ErrWebhookDLQReplayUnavailable, http.MethodPost, func(h *WebhookDLQHandler, c *gin.Context) { h.ReplayDeadLetterQueueItem(c) }, http.StatusConflict},
		{"missing delete", services.ErrWebhookDLQItemNotFound, http.MethodDelete, func(h *WebhookDLQHandler, c *gin.Context) { h.DeleteDeadLetterQueueItem(c) }, http.StatusNotFound},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := NewWebhookDLQHandler(&webhookDLQServiceStub{err: test.err})
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Params = gin.Params{{Key: "id", Value: id}}
			ctx.Request = httptest.NewRequest(test.method, "/api/v1/admin/webhooks/dlq/"+id, nil)
			test.handle(handler, ctx)
			if recorder.Code != test.want {
				t.Fatalf("expected %d, got %d: %s", test.want, recorder.Code, recorder.Body.String())
			}
		})
	}
}
