package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type cacheMonitor interface {
	GetStats(context.Context) (map[string]string, error)
	HealthCheck(context.Context) error
}

// MonitoringHandler handles monitoring and health check endpoints
type MonitoringHandler struct {
	redis cacheMonitor
}

// NewMonitoringHandler creates a new monitoring handler
func NewMonitoringHandler(redis cacheMonitor) *MonitoringHandler {
	return &MonitoringHandler{
		redis: redis,
	}
}

// GetCacheStats returns Redis cache statistics
// GET /internal/operations/cache
func (h *MonitoringHandler) GetCacheStats(c *gin.Context) {
	ctx := c.Request.Context()

	stats, err := h.redis.GetStats(ctx)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "degraded",
			"error":  "Failed to retrieve cache stats",
		})
		return
	}

	numericFields := []string{
		"keyspace_hits", "keyspace_misses", "used_memory", "used_memory_peak",
		"total_commands_processed", "instantaneous_ops_per_sec", "connected_clients",
		"evicted_keys", "expired_keys",
	}
	parsed := make(map[string]int64, len(numericFields))
	for _, field := range numericFields {
		value, exists := stats[field]
		if !exists {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "degraded", "error": fmt.Sprintf("Cache stats missing %s", field)})
			return
		}
		parsedValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil || parsedValue < 0 {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "degraded", "error": fmt.Sprintf("Cache stats contain invalid %s", field)})
			return
		}
		parsed[field] = parsedValue
	}

	hitRate := 0.0
	total := parsed["keyspace_hits"] + parsed["keyspace_misses"]
	if total > 0 {
		hitRate = float64(parsed["keyspace_hits"]) / float64(total) * 100
	}

	// Build response
	response := gin.H{
		"status": "healthy",
		"cache": gin.H{
			"hit_rate":               hitRate,
			"keyspace_hits":          parsed["keyspace_hits"],
			"keyspace_misses":        parsed["keyspace_misses"],
			"used_memory":            parsed["used_memory"],
			"used_memory_human":      stats["used_memory_human"],
			"used_memory_peak":       parsed["used_memory_peak"],
			"used_memory_peak_human": stats["used_memory_peak_human"],
			"total_commands":         parsed["total_commands_processed"],
			"instantaneous_ops":      parsed["instantaneous_ops_per_sec"],
			"connected_clients":      parsed["connected_clients"],
			"evicted_keys":           parsed["evicted_keys"],
			"expired_keys":           parsed["expired_keys"],
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetCacheHealth returns a simple cache health check
// GET /internal/operations/cache/check
func (h *MonitoringHandler) GetCacheHealth(c *gin.Context) {
	ctx := c.Request.Context()

	if err := h.redis.HealthCheck(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  "Redis is not accessible",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"cache":  "ok",
	})
}
