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

type clipSyncServiceStub struct {
	stats    *services.SyncStats
	lastSync *time.Time
	err      error
}

func (s *clipSyncServiceStub) SyncClipsByGame(context.Context, string, int, int, *services.SyncClipsByGameOptions) (*services.SyncStats, string, error) {
	return s.stats, "", s.err
}
func (s *clipSyncServiceStub) SyncClipsByBroadcaster(context.Context, string, int, int, *services.SyncClipsByBroadcasterOptions) (*services.SyncStats, error) {
	return s.stats, s.err
}
func (s *clipSyncServiceStub) SyncTrendingClips(context.Context, int, *services.TrendingSyncOptions) (*services.SyncStats, error) {
	return s.stats, s.err
}
func (s *clipSyncServiceStub) GetLastSyncTime(context.Context) (*time.Time, error) {
	return s.lastSync, s.err
}
func (s *clipSyncServiceStub) FetchClipByURL(context.Context, string) (*models.Clip, error) {
	return nil, s.err
}

func TestClipSyncHandlerResponseContracts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	start := time.Now().Add(-time.Second)
	end := time.Now()
	lastSync := end.Add(-time.Minute)

	t.Run("complete sync", func(t *testing.T) {
		handler := NewClipSyncHandler(&clipSyncServiceStub{stats: &services.SyncStats{ClipsFetched: 3, ClipsCreated: 2, ClipsUpdated: 1, StartTime: start, EndTime: end}})
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/sync/clips", strings.NewReader(`{"strategy":"trending","hours":24,"limit":100}`))
		ctx.Request.Header.Set("Content-Type", "application/json")
		handler.TriggerSync(ctx)
		if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"clips_created":2`) {
			t.Fatalf("unexpected sync response: %d %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("partial sync", func(t *testing.T) {
		handler := NewClipSyncHandler(&clipSyncServiceStub{stats: &services.SyncStats{ClipsFetched: 3, ClipsCreated: 2, Errors: []string{"one clip failed"}, StartTime: start, EndTime: end}})
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/sync/clips", strings.NewReader(`{"strategy":"trending"}`))
		ctx.Request.Header.Set("Content-Type", "application/json")
		handler.TriggerSync(ctx)
		if recorder.Code != http.StatusMultiStatus || !strings.Contains(recorder.Body.String(), "completed with errors") {
			t.Fatalf("expected partial response, got %d %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("persisted status", func(t *testing.T) {
		handler := NewClipSyncHandler(&clipSyncServiceStub{lastSync: &lastSync})
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/sync/status", nil)
		handler.GetSyncStatus(ctx)
		if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"status":"ready"`) || !strings.Contains(recorder.Body.String(), "last_sync_at") {
			t.Fatalf("unexpected status response: %d %s", recorder.Code, recorder.Body.String())
		}
	})
}

func TestClipSyncHandlerRejectsInvalidBounds(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewClipSyncHandler(&clipSyncServiceStub{})
	for _, body := range []string{`{`, `{"hours":169}`, `{"limit":101}`, `{"strategy":"game"}`} {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/sync/clips", strings.NewReader(body))
		ctx.Request.Header.Set("Content-Type", "application/json")
		handler.TriggerSync(ctx)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for %s, got %d: %s", body, recorder.Code, recorder.Body.String())
		}
	}
}
