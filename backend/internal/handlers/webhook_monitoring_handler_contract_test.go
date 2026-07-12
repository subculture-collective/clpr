package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type webhookRetryStatsServiceStub struct {
	stats map[string]interface{}
	err   error
}

func (s *webhookRetryStatsServiceStub) GetRetryQueueStats(context.Context) (map[string]interface{}, error) {
	return s.stats, s.err
}

type outboundWebhookStatsServiceStub struct {
	stats map[string]interface{}
	err   error
}

func (s *outboundWebhookStatsServiceStub) GetDeliveryStats(context.Context) (map[string]interface{}, error) {
	return s.stats, s.err
}

func TestWebhookMonitoringResponseContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	retry := &webhookRetryStatsServiceStub{stats: map[string]interface{}{
		"pending_retries": 2, "dlq_items": 1, "timestamp": time.Now(),
	}}
	delivery := &outboundWebhookStatsServiceStub{stats: map[string]interface{}{
		"active_subscriptions": 3, "pending_deliveries": 4,
		"recent_deliveries": map[string]int{"success": 5, "failed": 1},
	}}
	handler := NewWebhookMonitoringHandler(retry, delivery)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/internal/operations/webhooks", nil)
	handler.GetWebhookRetryStats(ctx)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"status":"healthy"`) || !strings.Contains(recorder.Body.String(), `"active_subscriptions":3`) {
		t.Fatalf("unexpected monitoring response: %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestWebhookMonitoringFailsClosedOnIncompleteStats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("retry queue unavailable", func(t *testing.T) {
		handler := NewWebhookMonitoringHandler(&webhookRetryStatsServiceStub{err: errors.New("database unavailable")}, &outboundWebhookStatsServiceStub{})
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodGet, "/internal/operations/webhooks", nil)
		handler.GetWebhookRetryStats(ctx)
		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", recorder.Code)
		}
	})

	t.Run("delivery stats unavailable", func(t *testing.T) {
		handler := NewWebhookMonitoringHandler(
			&webhookRetryStatsServiceStub{stats: map[string]interface{}{"pending_retries": 2, "dlq_items": 1, "timestamp": time.Now()}},
			&outboundWebhookStatsServiceStub{err: errors.New("database unavailable")},
		)
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodGet, "/internal/operations/webhooks", nil)
		handler.GetWebhookRetryStats(ctx)
		if recorder.Code != http.StatusServiceUnavailable || !strings.Contains(recorder.Body.String(), `"status":"degraded"`) {
			t.Fatalf("expected degraded 503, got %d %s", recorder.Code, recorder.Body.String())
		}
	})
}
