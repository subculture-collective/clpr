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
)

type eventTrackingServiceStub struct {
	trackErrors []error
	trackCalls  int
	feedMetrics map[string]interface{}
	hourly      []models.HourlyMetric
	err         error
}

func (s *eventTrackingServiceStub) TrackEvent(models.Event) error {
	index := s.trackCalls
	s.trackCalls++
	if index < len(s.trackErrors) {
		return s.trackErrors[index]
	}
	return s.err
}
func (s *eventTrackingServiceStub) GetFeedMetrics(context.Context, int) (map[string]interface{}, error) {
	return s.feedMetrics, s.err
}
func (s *eventTrackingServiceStub) GetHourlyMetrics(context.Context, string, int) ([]models.HourlyMetric, error) {
	return s.hourly, s.err
}

func TestEventHandlerLiveResponseContracts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("single event accepted", func(t *testing.T) {
		service := &eventTrackingServiceStub{}
		handler := NewEventHandler(service)
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/events", strings.NewReader(`{"event_type":"feed_viewed","properties":{"source":"home"}}`))
		ctx.Request.Header.Set("Content-Type", "application/json")
		ctx.Request.Header.Set("X-Session-ID", "session-1")
		handler.TrackEvent(ctx)
		if recorder.Code != http.StatusAccepted || service.trackCalls != 1 {
			t.Fatalf("expected accepted event, got %d %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("queue saturation is unavailable", func(t *testing.T) {
		service := &eventTrackingServiceStub{err: services.ErrEventQueueFull}
		handler := NewEventHandler(service)
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/events", strings.NewReader(`{"event_type":"feed_viewed"}`))
		handler.TrackEvent(ctx)
		if recorder.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected 503, got %d: %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("partial batch is explicit", func(t *testing.T) {
		service := &eventTrackingServiceStub{trackErrors: []error{nil, services.ErrEventQueueFull}}
		handler := NewEventHandler(service)
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/events", strings.NewReader(`{"events":[{"event_type":"feed_viewed"},{"event_type":"feed_engaged"}]}`))
		handler.TrackEvent(ctx)
		if recorder.Code != http.StatusMultiStatus {
			t.Fatalf("expected 207, got %d: %s", recorder.Code, recorder.Body.String())
		}
		if !strings.Contains(recorder.Body.String(), `"success_count":1`) {
			t.Fatalf("unexpected partial response: %s", recorder.Body.String())
		}
	})

	t.Run("feed metrics", func(t *testing.T) {
		service := &eventTrackingServiceStub{feedMetrics: map[string]interface{}{
			"events":       []map[string]interface{}{{"event_type": "feed_viewed", "total_count": int64(2), "unique_users": int64(1), "unique_sessions": int64(1)}},
			"period_hours": 24,
		}}
		handler := NewEventHandler(service)
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/feeds/analytics?hours=24", nil)
		handler.GetFeedMetrics(ctx)
		if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"period_hours":24`) {
			t.Fatalf("unexpected metrics response: %d %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("hourly metrics", func(t *testing.T) {
		service := &eventTrackingServiceStub{hourly: []models.HourlyMetric{{Hour: time.Now(), EventType: "feed_viewed", Count: 2, UniqueUsers: 1, UniqueSessions: 1}}}
		handler := NewEventHandler(service)
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/feeds/analytics/hourly?event_type=feed_viewed&hours=24", nil)
		handler.GetHourlyMetrics(ctx)
		if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"event_type":"feed_viewed"`) {
			t.Fatalf("unexpected hourly response: %d %s", recorder.Code, recorder.Body.String())
		}
	})
}

func TestEventHandlerRejectsUnboundedOrInvalidInput(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewEventHandler(&eventTrackingServiceStub{})

	tests := []struct {
		name, method, target, body string
		header                     string
		handle                     func(*gin.Context)
		want                       int
	}{
		{"empty batch event", http.MethodPost, "/api/v1/events", `{"events":[{"event_type":""}]}`, "", handler.TrackEvent, http.StatusBadRequest},
		{"long session", http.MethodPost, "/api/v1/events", `{"event_type":"feed_viewed"}`, strings.Repeat("x", 101), handler.TrackEvent, http.StatusBadRequest},
		{"invalid hours", http.MethodGet, "/api/v1/feeds/analytics?hours=invalid", "", "", handler.GetFeedMetrics, http.StatusBadRequest},
		{"excessive hours", http.MethodGet, "/api/v1/feeds/analytics/hourly?event_type=feed_viewed&hours=721", "", "", handler.GetHourlyMetrics, http.StatusBadRequest},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(test.method, test.target, strings.NewReader(test.body))
			if test.header != "" {
				ctx.Request.Header.Set("X-Session-ID", test.header)
			}
			test.handle(ctx)
			if recorder.Code != test.want {
				t.Fatalf("expected %d, got %d: %s", test.want, recorder.Code, recorder.Body.String())
			}
		})
	}

	t.Run("oversized body", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/events", strings.NewReader(`{"event_type":"feed_viewed","properties":{"value":"`+strings.Repeat("x", maxEventRequestBytes)+`"}}`))
		handler.TrackEvent(ctx)
		if recorder.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("expected 413, got %d: %s", recorder.Code, recorder.Body.String())
		}
	})
}
