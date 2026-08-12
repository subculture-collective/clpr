package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestAdminListUsersRejectsMalformedFiltersBeforeRepositoryAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []string{
		"?page=zero",
		"?page=0",
		"?per_page=101",
		"?role=owner",
		"?status=disabled",
		"?search=" + strings.Repeat("x", 201),
	}
	for _, query := range tests {
		t.Run(query, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/users"+query, http.NoBody)

			(&AdminUserHandler{}).ListUsers(ctx)

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
			}
		})
	}
}

func TestAdminUpdateKarmaAcceptsZeroAndRejectsMalformedIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/admin/users/"+uuid.NewString()+"/karma", strings.NewReader(`{"karma_points":0}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Params = gin.Params{{Key: "id", Value: uuid.NewString()}}
	ctx.Set("user_id", "not-a-uuid")

	(&AdminUserHandler{}).UpdateUserKarma(ctx)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected zero karma to bind before malformed identity returns %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

func TestPlatformModeratorEndpointsRejectMalformedInputsBeforeRepositoryAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		method string
		path   string
		body   string
		call   func(*gin.Context)
	}{
		{http.MethodGet, "/api/v1/admin/moderators?page=0", "", (&AdminUserHandler{}).ListPlatformModerators},
		{http.MethodGet, "/api/v1/admin/moderators?limit=101", "", (&AdminUserHandler{}).ListPlatformModerators},
		{http.MethodPost, "/api/v1/admin/moderators", `{"user_id":"invalid"}`, (&AdminUserHandler{}).AddPlatformModerator},
		{http.MethodPatch, "/api/v1/admin/moderators/invalid", `{}`, (&AdminUserHandler{}).UpdatePlatformModerator},
		{http.MethodDelete, "/api/v1/admin/moderators/invalid", `{}`, (&AdminUserHandler{}).RevokePlatformModerator},
	}
	for _, test := range tests {
		t.Run(test.method+test.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
			ctx.Request.Header.Set("Content-Type", "application/json")
			if strings.HasSuffix(test.path, "/invalid") {
				ctx.Params = gin.Params{{Key: "id", Value: "invalid"}}
			}
			test.call(ctx)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
			}
		})
	}
}

func TestAwardBadgeRejectsMalformedIdentityWithoutPanicking(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/users/"+uuid.NewString()+"/badges", strings.NewReader(`{"badge_id":"veteran"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Params = gin.Params{{Key: "id", Value: uuid.NewString()}}
	ctx.Set("user_id", "not-a-uuid")

	(&ReputationHandler{}).AwardBadge(ctx)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}
