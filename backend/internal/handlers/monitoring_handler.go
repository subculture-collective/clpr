package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	redispkg "git.subcult.tv/subculture-collective/clpr/pkg/redis"
)

// MonitoringHandler handles monitoring and health check endpoints
type MonitoringHandler struct {
	redis *redispkg.Client
}

// NewMonitoringHandler creates a new monitoring handler
func NewMonitoringHandler(redis *redispkg.Client) *MonitoringHandler {
	return &MonitoringHandler{
		redis: redis,
	}
}

// GetCacheStats returns Redis cache statistics
// GET /health/cache
func (h *MonitoringHandler) GetCacheStats(c *gin.Context) {
	ctx := c.Request.Context()

	stats, err := h.redis.GetStats(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve cache stats",
		})
		return
	}

	// Parse key stats
	hitRate := 0.0
	if hits, exists := stats["keyspace_hits"]; exists {
		if misses, exists := stats["keyspace_misses"]; exists {
			hitsInt, _ := strconv.ParseInt(hits, 10, 64)
			missesInt, _ := strconv.ParseInt(misses, 10, 64)
			total := hitsInt + missesInt
			if total > 0 {
				hitRate = float64(hitsInt) / float64(total) * 100
			}
		}
	}

	// Build response
	response := gin.H{
		"status": "healthy",
		"cache": gin.H{
			"hit_rate":               hitRate,
			"keyspace_hits":          stats["keyspace_hits"],
			"keyspace_misses":        stats["keyspace_misses"],
			"used_memory":            stats["used_memory"],
			"used_memory_human":      stats["used_memory_human"],
			"used_memory_peak":       stats["used_memory_peak"],
			"used_memory_peak_human": stats["used_memory_peak_human"],
			"total_commands":         stats["total_commands_processed"],
			"instantaneous_ops":      stats["instantaneous_ops_per_sec"],
			"connected_clients":      stats["connected_clients"],
			"evicted_keys":           stats["evicted_keys"],
			"expired_keys":           stats["expired_keys"],
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetCacheHealth returns a simple cache health check
// GET /health/cache/check
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
