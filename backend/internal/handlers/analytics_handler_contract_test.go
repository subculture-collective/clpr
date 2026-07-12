package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type analyticsServiceStub struct {
	overview       *models.CreatorAnalyticsOverview
	clips          []models.CreatorTopClip
	trends         []models.TrendDataPoint
	audience       *models.CreatorAudienceInsights
	err            error
	creatorName    string
	sortBy         string
	metric         string
	limit          int
	days           int
	unexpectedCall bool
}

func (s *analyticsServiceStub) GetCreatorAnalyticsOverview(_ context.Context, creatorName string) (*models.CreatorAnalyticsOverview, error) {
	s.creatorName = creatorName
	return s.overview, s.err
}
func (s *analyticsServiceStub) GetCreatorTopClips(_ context.Context, creatorName, sortBy string, limit int) ([]models.CreatorTopClip, error) {
	s.creatorName, s.sortBy, s.limit = creatorName, sortBy, limit
	return s.clips, s.err
}
func (s *analyticsServiceStub) GetCreatorTrends(_ context.Context, creatorName, metric string, days int) ([]models.TrendDataPoint, error) {
	s.creatorName, s.metric, s.days = creatorName, metric, days
	return s.trends, s.err
}
func (s *analyticsServiceStub) GetCreatorAudienceInsights(_ context.Context, creatorName string, limit int) (*models.CreatorAudienceInsights, error) {
	s.creatorName, s.limit = creatorName, limit
	return s.audience, s.err
}
func (s *analyticsServiceStub) GetClipAnalytics(context.Context, uuid.UUID) (*models.ClipAnalytics, error) {
	s.unexpectedCall = true
	return nil, s.err
}
func (s *analyticsServiceStub) GetUserAnalytics(context.Context, uuid.UUID) (*models.UserAnalytics, error) {
	s.unexpectedCall = true
	return nil, s.err
}
func (s *analyticsServiceStub) GetPlatformOverview(context.Context) (*models.PlatformOverviewMetrics, error) {
	s.unexpectedCall = true
	return nil, s.err
}
func (s *analyticsServiceStub) GetContentMetrics(context.Context) (*models.ContentMetrics, error) {
	s.unexpectedCall = true
	return nil, s.err
}
func (s *analyticsServiceStub) GetPlatformTrends(context.Context, string, int) ([]models.TrendDataPoint, error) {
	s.unexpectedCall = true
	return nil, s.err
}
func (s *analyticsServiceStub) TrackEvent(context.Context, string, *uuid.UUID, *uuid.UUID, map[string]interface{}, string, string, string) error {
	s.unexpectedCall = true
	return s.err
}

func creatorAnalyticsContext(method, target, creatorName string) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Params = gin.Params{{Key: "creator", Value: creatorName}}
	ctx.Request = httptest.NewRequest(method, target, nil)
	return ctx, recorder
}

func TestCreatorAnalyticsHandlersValidatePublicQueries(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name   string
		target string
		call   func(*AnalyticsHandler, *gin.Context)
	}{
		{"invalid clip limit", "/api/v1/creators/alice/analytics/clips?limit=lots", (*AnalyticsHandler).GetCreatorTopClips},
		{"out of range clip limit", "/api/v1/creators/alice/analytics/clips?limit=101", (*AnalyticsHandler).GetCreatorTopClips},
		{"invalid clip sort", "/api/v1/creators/alice/analytics/clips?sort=secret", (*AnalyticsHandler).GetCreatorTopClips},
		{"invalid trend days", "/api/v1/creators/alice/analytics/trends?days=0", (*AnalyticsHandler).GetCreatorTrends},
		{"unsupported trend metric", "/api/v1/creators/alice/analytics/trends?metric=users_active", (*AnalyticsHandler).GetCreatorTrends},
		{"out of range audience limit", "/api/v1/creators/alice/analytics/audience?limit=51", (*AnalyticsHandler).GetCreatorAudienceInsights},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &analyticsServiceStub{}
			handler := NewAnalyticsHandler(service)
			ctx, recorder := creatorAnalyticsContext(http.MethodGet, tt.target, "alice")
			tt.call(handler, ctx)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", recorder.Code, recorder.Body.String())
			}
			if service.creatorName != "" || service.unexpectedCall {
				t.Fatal("invalid input reached the analytics service")
			}
		})
	}
}

func TestCreatorAnalyticsHandlerContracts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("overview distinguishes absence from dependency failure", func(t *testing.T) {
		handler := NewAnalyticsHandler(&analyticsServiceStub{err: pgx.ErrNoRows})
		ctx, recorder := creatorAnalyticsContext(http.MethodGet, "/api/v1/creators/missing/analytics/overview", "missing")
		handler.GetCreatorAnalyticsOverview(ctx)
		if recorder.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("trend preserves validated request", func(t *testing.T) {
		service := &analyticsServiceStub{trends: []models.TrendDataPoint{{Date: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC), Value: 12}}}
		handler := NewAnalyticsHandler(service)
		ctx, recorder := creatorAnalyticsContext(http.MethodGet, "/api/v1/creators/alice/analytics/trends?metric=votes&days=90", "alice")
		handler.GetCreatorTrends(ctx)
		if recorder.Code != http.StatusOK || service.metric != "votes" || service.days != 90 || !strings.Contains(recorder.Body.String(), `"value":12`) {
			t.Fatalf("unexpected response/service call: %d %s metric=%q days=%d", recorder.Code, recorder.Body.String(), service.metric, service.days)
		}
	})

	t.Run("audience contract makes no fabricated country claim", func(t *testing.T) {
		service := &analyticsServiceStub{audience: &models.CreatorAudienceInsights{
			TopCountries: []models.GeographyMetric{},
			DeviceTypes:  []models.DeviceMetric{{DeviceType: "mobile", ViewCount: 2, Percentage: 100}},
			TotalViews:   2,
		}}
		handler := NewAnalyticsHandler(service)
		ctx, recorder := creatorAnalyticsContext(http.MethodGet, "/api/v1/creators/alice/analytics/audience?limit=5", "alice")
		handler.GetCreatorAudienceInsights(ctx)
		if recorder.Code != http.StatusOK || strings.Contains(recorder.Body.String(), `"country":"XX"`) || !strings.Contains(recorder.Body.String(), `"top_countries":[]`) {
			t.Fatalf("unexpected audience response: %d %s", recorder.Code, recorder.Body.String())
		}
	})
}
