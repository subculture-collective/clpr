package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestReportResolutionValidation(t *testing.T) {
	removeContent := "remove_content"
	markFalse := "mark_false"
	tests := []struct {
		name, status string
		action       *string
		wantError    bool
	}{
		{"review without action", "reviewed", nil, false},
		{"actioned requires action", "actioned", nil, true},
		{"removal requires actioned", "reviewed", &removeContent, true},
		{"removal actioned", "actioned", &removeContent, false},
		{"false report dismissed", "dismissed", &markFalse, false},
		{"false report cannot action", "actioned", &markFalse, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateReportResolution(test.status, test.action)
			if (err != nil) != test.wantError {
				t.Fatalf("validateReportResolution() error = %v, wantError %v", err, test.wantError)
			}
		})
	}
}

func TestReportActionTargetValidation(t *testing.T) {
	tests := []struct {
		action, target string
		wantError      bool
	}{
		{"remove_content", "clip", false},
		{"remove_content", "comment", false},
		{"remove_content", "user", true},
		{"ban_user", "comment", false},
		{"ban_user", "user", false},
		{"ban_user", "clip", true},
		{"warn_user", "user", true},
	}
	for _, test := range tests {
		t.Run(test.action+"/"+test.target, func(t *testing.T) {
			err := validateReportActionTarget(test.action, test.target)
			if (err != nil) != test.wantError {
				t.Fatalf("validateReportActionTarget() error = %v, wantError %v", err, test.wantError)
			}
		})
	}
}

func TestReportHandlersRejectInvalidInputBeforeRepositoryAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &ReportHandler{}

	for _, target := range []string{
		"/api/v1/admin/reports?status=unknown",
		"/api/v1/admin/reports?type=unknown",
		"/api/v1/admin/reports?page=0",
		"/api/v1/admin/reports?limit=101",
	} {
		t.Run(target, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodGet, target, nil)
			handler.ListReports(ctx)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", recorder.Code)
			}
		})
	}

	t.Run("malformed actor context", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Params = gin.Params{{Key: "id", Value: uuid.NewString()}}
		ctx.Set("user_id", "not-a-uuid")
		ctx.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/reports/"+uuid.NewString(), strings.NewReader(`{"status":"reviewed"}`))
		ctx.Request.Header.Set("Content-Type", "application/json")
		handler.UpdateReport(ctx)
		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", recorder.Code)
		}
	})

	t.Run("unsupported warning action", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Params = gin.Params{{Key: "id", Value: uuid.NewString()}}
		ctx.Set("user_id", uuid.New())
		ctx.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/reports/"+uuid.NewString(), strings.NewReader(`{"status":"actioned","action":"warn_user"}`))
		ctx.Request.Header.Set("Content-Type", "application/json")
		handler.UpdateReport(ctx)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", recorder.Code, recorder.Body.String())
		}
	})
}
