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
)

type liveStatusServiceStub struct {
	status       *models.BroadcasterLiveStatus
	broadcasters []models.BroadcasterLiveStatus
	total        int
	err          error
}

func (s *liveStatusServiceStub) GetLiveStatus(context.Context, string) (*models.BroadcasterLiveStatus, error) {
	return s.status, s.err
}
func (s *liveStatusServiceStub) ListLiveBroadcasters(context.Context, int, int) ([]models.BroadcasterLiveStatus, int, error) {
	return s.broadcasters, s.total, s.err
}
func (s *liveStatusServiceStub) GetFollowedLiveBroadcasters(context.Context, uuid.UUID) ([]models.BroadcasterLiveStatus, error) {
	return s.broadcasters, s.err
}

func TestLiveStatusHandlerResponseContracts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	status := models.BroadcasterLiveStatus{BroadcasterID: "123", IsLive: true, ViewerCount: 42, LastChecked: time.Now(), CreatedAt: time.Now(), UpdatedAt: time.Now()}
	handler := NewLiveStatusHandler(&liveStatusServiceStub{status: &status, broadcasters: []models.BroadcasterLiveStatus{status}, total: 1})

	t.Run("individual", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Params = gin.Params{{Key: "id", Value: "123"}}
		ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/broadcasters/123/live-status", nil)
		handler.GetBroadcasterLiveStatus(ctx)
		if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"viewer_count":42`) || !strings.Contains(recorder.Body.String(), `"is_stale":false`) {
			t.Fatalf("unexpected individual status: %d %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("list", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/broadcasters/live?page=1&limit=50", nil)
		handler.ListLiveBroadcasters(ctx)
		if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"total_items":1`) {
			t.Fatalf("unexpected live list: %d %s", recorder.Code, recorder.Body.String())
		}
	})
}

func TestLiveStatusHandlerRejectsInvalidPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewLiveStatusHandler(&liveStatusServiceStub{})
	for _, target := range []string{"/api/v1/broadcasters/live?page=0", "/api/v1/broadcasters/live?limit=101", "/api/v1/broadcasters/live?page=invalid"} {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodGet, target, nil)
		handler.ListLiveBroadcasters(ctx)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for %s, got %d", target, recorder.Code)
		}
	}
}
