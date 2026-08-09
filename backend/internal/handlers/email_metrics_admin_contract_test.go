package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func emailAdminTestContext(target string) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, target, nil)
	return ctx, recorder
}

func TestEmailMetricsRangeRejectsInvalidAndExcessiveRanges(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, target := range []string{
		"/?start_date=invalid",
		"/?start_date=2026-07-12T00:00:00Z&end_date=2026-07-11T00:00:00Z",
		"/?start_date=2025-01-01T00:00:00Z&end_date=2026-07-12T00:00:00Z",
	} {
		t.Run(target, func(t *testing.T) {
			ctx, recorder := emailAdminTestContext(target)
			_, _, ok := parseEmailMetricsRange(ctx, 7)
			if ok || recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected rejected range with 400, got ok=%v status=%d", ok, recorder.Code)
			}
		})
	}
}

func TestEmailMetricsBoundedIntegerRejectsMalformedValues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, target := range []string{"/?limit=invalid", "/?limit=0", "/?limit=201"} {
		t.Run(target, func(t *testing.T) {
			ctx, recorder := emailAdminTestContext(target)
			_, ok := parseBoundedEmailInt(ctx, "limit", 50, 1, 200)
			if ok || recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected rejected limit with 400, got ok=%v status=%d", ok, recorder.Code)
			}
		})
	}
}

func TestAcknowledgeEmailAlertRejectsInvalidIdentityType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &EmailMetricsHandler{}
	ctx, recorder := emailAdminTestContext("/api/v1/admin/email/alerts/invalid/acknowledge")
	ctx.Set("user_id", "not-a-uuid")

	handler.AcknowledgeAlert(ctx)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}
