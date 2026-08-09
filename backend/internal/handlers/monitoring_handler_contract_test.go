package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type cacheMonitorStub struct {
	stats     map[string]string
	statsErr  error
	healthErr error
}

func (s *cacheMonitorStub) GetStats(context.Context) (map[string]string, error) {
	return s.stats, s.statsErr
}
func (s *cacheMonitorStub) HealthCheck(context.Context) error { return s.healthErr }

func validCacheStats() map[string]string {
	return map[string]string{
		"keyspace_hits": "75", "keyspace_misses": "25", "used_memory": "1024",
		"used_memory_human": "1K", "used_memory_peak": "2048", "used_memory_peak_human": "2K",
		"total_commands_processed": "1000", "instantaneous_ops_per_sec": "5",
		"connected_clients": "3", "evicted_keys": "0", "expired_keys": "10",
	}
}

func TestCacheMonitoringResponseContracts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewMonitoringHandler(&cacheMonitorStub{stats: validCacheStats()})

	t.Run("typed stats", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodGet, "/internal/operations/cache", nil)
		handler.GetCacheStats(ctx)
		if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"hit_rate":75`) || !strings.Contains(recorder.Body.String(), `"keyspace_hits":75`) {
			t.Fatalf("unexpected cache stats: %d %s", recorder.Code, recorder.Body.String())
		}
	})

	t.Run("health", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodGet, "/internal/operations/cache/check", nil)
		handler.GetCacheHealth(ctx)
		if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"status":"healthy"`) {
			t.Fatalf("unexpected cache health: %d %s", recorder.Code, recorder.Body.String())
		}
	})
}

func TestCacheMonitoringFailsClosed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name    string
		monitor *cacheMonitorStub
		handle  func(*MonitoringHandler, *gin.Context)
	}{
		{"stats dependency", &cacheMonitorStub{statsErr: errors.New("redis unavailable")}, func(h *MonitoringHandler, c *gin.Context) { h.GetCacheStats(c) }},
		{"malformed stats", &cacheMonitorStub{stats: map[string]string{"keyspace_hits": "not-a-number"}}, func(h *MonitoringHandler, c *gin.Context) { h.GetCacheStats(c) }},
		{"health dependency", &cacheMonitorStub{healthErr: errors.New("redis unavailable")}, func(h *MonitoringHandler, c *gin.Context) { h.GetCacheHealth(c) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := NewMonitoringHandler(test.monitor)
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/internal/operations/cache", nil)
			test.handle(handler, ctx)
			if recorder.Code != http.StatusServiceUnavailable {
				t.Fatalf("expected 503, got %d: %s", recorder.Code, recorder.Body.String())
			}
		})
	}
}
