package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestModerationAdminListEndpointsRejectMalformedLimits(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		path string
		call func(*ModerationHandler, *gin.Context)
	}{
		{"/api/v1/admin/moderation/events?limit=zero", (*ModerationHandler).GetPendingEvents},
		{"/api/v1/admin/moderation/events/submission_received?limit=101", (*ModerationHandler).GetEventsByType},
		{"/api/v1/admin/moderation/queue?limit=0", (*ModerationHandler).GetModerationQueue},
		{"/api/v1/admin/moderation/appeals?limit=101", (*ModerationHandler).GetAppeals},
		{"/api/v1/admin/moderation/audit?offset=bad", (*ModerationHandler).GetModerationAuditLogs},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodGet, tt.path, http.NoBody)
			ctx.Params = gin.Params{{Key: "type", Value: "submission_received"}}
			tt.call(&ModerationHandler{}, ctx)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
			}
		})
	}
}

func TestModerationAdminRejectsUnknownEventTypeAndAction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		path string
		body string
		call func(*ModerationHandler, *gin.Context)
	}{
		{"/api/v1/admin/moderation/events/unknown", "", (*ModerationHandler).GetEventsByType},
		{"/api/v1/admin/moderation/events/" + uuid.NewString() + "/process", `{"action":"delete_everything"}`, (*ModerationHandler).ProcessEvent},
	}
	for _, tt := range tests {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodPost, tt.path, strings.NewReader(tt.body))
		ctx.Request.Header.Set("Content-Type", "application/json")
		ctx.Params = gin.Params{{Key: "type", Value: "unknown"}, {Key: "id", Value: uuid.NewString()}}
		ctx.Set("user_id", uuid.New())
		tt.call(&ModerationHandler{}, ctx)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
		}
	}
}

func TestModerationAdminRejectsMalformedIdentityWithoutPanicking(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/moderation/"+uuid.NewString()+"/approve", http.NoBody)
	ctx.Params = gin.Params{{Key: "id", Value: uuid.NewString()}}
	ctx.Set("user_id", "not-a-uuid")

	(&ModerationHandler{}).ApproveContent(ctx)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestBulkModerationRejectsDuplicateItems(t *testing.T) {
	gin.SetMode(gin.TestMode)
	id := uuid.NewString()
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/moderation/bulk", strings.NewReader(`{"item_ids":["`+id+`","`+id+`"],"action":"approve"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Set("user_id", uuid.New())

	(&ModerationHandler{}).BulkModerate(ctx)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
}

func TestModerationDateRangesFailClosed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, path := range []string{
		"/api/v1/admin/moderation/audit?start_date=2026-02-30",
		"/api/v1/admin/moderation/analytics?start_date=2026-07-10&end_date=2026-07-01",
		"/api/v1/admin/moderation/toxicity/metrics?start_date=2020-01-01&end_date=2026-01-01",
	} {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodGet, path, http.NoBody)
		handler := &ModerationHandler{}
		switch {
		case strings.Contains(path, "/audit"):
			handler.GetModerationAuditLogs(ctx)
		case strings.Contains(path, "/analytics"):
			handler.GetModerationAnalytics(ctx)
		default:
			handler.GetToxicityMetrics(ctx)
		}
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("%s: expected status %d, got %d", path, http.StatusBadRequest, recorder.Code)
		}
	}
}
