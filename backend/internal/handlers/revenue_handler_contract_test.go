package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/gin-gonic/gin"
)

type revenueMetricsServiceStub struct {
	metrics *models.RevenueMetrics
	err     error
}

func (s *revenueMetricsServiceStub) GetRevenueMetrics(context.Context) (*models.RevenueMetrics, error) {
	return s.metrics, s.err
}

func TestRevenueHandlerResponseContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewRevenueHandler(&revenueMetricsServiceStub{metrics: &models.RevenueMetrics{
		MRR: 999, ARPU: 999, ActiveSubscribers: 1, TotalRevenue: 11988,
		PlanDistribution: []models.PlanDistributionMetric{}, CohortRetention: []models.CohortRetentionMetric{},
		RevenueByMonth:   []models.RevenueByMonthMetric{{Month: "2026-07", Revenue: 11988, MRR: 999}},
		SubscriberGrowth: []models.SubscriberGrowthMetric{{Month: "2026-07", Total: 1, New: 1}}, UpdatedAt: time.Now(),
	}})
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/revenue", nil)
	handler.GetRevenueMetrics(ctx)
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"total_revenue":11988`) || !strings.Contains(recorder.Body.String(), `"mrr":999`) {
		t.Fatalf("unexpected revenue response: %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestRevenueHandlerFailsClosedOnIncompleteMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewRevenueHandler(&revenueMetricsServiceStub{err: errors.New("cohort query failed")})
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/revenue", nil)
	handler.GetRevenueMetrics(ctx)
	if recorder.Code != http.StatusInternalServerError || !strings.Contains(recorder.Body.String(), `"code":"METRICS_ERROR"`) {
		t.Fatalf("expected fail-closed metrics response, got %d %s", recorder.Code, recorder.Body.String())
	}
}
