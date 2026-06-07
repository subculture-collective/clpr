package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
)

// AbuseAnalyticsHandler handles abuse detection analytics and metrics
type AbuseAnalyticsHandler struct {
	anomalyScorer *services.AnomalyScorer
	autoFlagger   *services.AbuseAutoFlagger
}

// NewAbuseAnalyticsHandler creates a new abuse analytics handler
func NewAbuseAnalyticsHandler(
	anomalyScorer *services.AnomalyScorer,
	autoFlagger *services.AbuseAutoFlagger,
) *AbuseAnalyticsHandler {
	return &AbuseAnalyticsHandler{
		anomalyScorer: anomalyScorer,
		autoFlagger:   autoFlagger,
	}
}

// GetAbuseMetrics returns real-time abuse detection metrics
// GET /api/v1/admin/abuse/metrics
func (h *AbuseAnalyticsHandler) GetAbuseMetrics(c *gin.Context) {
	// This endpoint is intended to be admin-only.
	// Authentication/authorization MUST be enforced by middleware on the /api/v1/admin route group
	// in the router configuration where this handler is registered.

	ctx := c.Request.Context()

	// Get anomaly scorer metrics
	scorerMetrics, err := h.anomalyScorer.GetMetrics(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve anomaly metrics",
		})
		return
	}

	// Get auto-flagger stats (last 24 hours)
	since := time.Now().Add(-24 * time.Hour)
	flaggerStats, err := h.autoFlagger.GetAutoFlagStats(ctx, since)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve auto-flag statistics",
		})
		return
	}

	// Combine metrics
	response := gin.H{
		"anomaly_detection": scorerMetrics,
		"auto_flagging":     flaggerStats,
		"timestamp":         time.Now().UTC(),
		"time_range":        "24h",
	}

	c.JSON(http.StatusOK, response)
}

// GetAbuseMetricsHistory returns historical abuse detection metrics
// GET /api/v1/admin/abuse/metrics/history
func (h *AbuseAnalyticsHandler) GetAbuseMetricsHistory(c *gin.Context) {
	// Parse time range from query parameters
	hoursStr := c.DefaultQuery("hours", "24")
	var hours int
	if _, err := fmt.Sscanf(hoursStr, "%d", &hours); err != nil || hours <= 0 {
		hours = 24
	}

	// Cap at 7 days
	if hours > 168 {
		hours = 168
	}

	ctx := c.Request.Context()
	since := time.Now().Add(-time.Duration(hours) * time.Hour)

	// Get auto-flagger stats for the time range
	flaggerStats, err := h.autoFlagger.GetAutoFlagStats(ctx, since)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve historical statistics",
		})
		return
	}

	// Calculate false positive rate if we have review data
	fpr := h.calculateFalsePositiveRate(flaggerStats)

	response := gin.H{
		"stats":                 flaggerStats,
		"false_positive_rate":   fpr,
		"time_range_hours":      hours,
		"timestamp":             time.Now().UTC(),
		"fpr_target":            0.02, // 2% target
		"fpr_meets_requirement": fpr <= 0.02,
	}

	c.JSON(http.StatusOK, response)
}

// calculateFalsePositiveRate calculates the false positive rate from stats.
// In this context, all items in the stats are auto-flagged as potential abuse.
// We define:
//   - "approved": item was reviewed and approved as legitimate (not abuse).
//   - "rejected": item was reviewed and confirmed as abuse.
//
// The false positive rate is therefore: approved / (approved + rejected),
// i.e., the fraction of reviewed auto-flagged items that turned out to be legitimate.
func (h *AbuseAnalyticsHandler) calculateFalsePositiveRate(stats map[string]interface{}) float64 {
	// Extract status counts if available
	if byStatus, ok := stats["by_status"].(map[string]int); ok {
		approved := byStatus["approved"]

		// Only include reviewed statuses in the denominator.
		// Pending or other non-final statuses should not affect the false positive rate.
		reviewedTotal := 0
		for status, count := range byStatus {
			if status == "approved" || status == "rejected" {
				reviewedTotal += count
			}
		}

		if reviewedTotal > 0 {
			return float64(approved) / float64(reviewedTotal)
		}
	}

	// Return 0 if we can't calculate (not enough data)
	return 0.0
}

// GetAbuseDetectionHealth returns health status of abuse detection system
// GET /api/v1/admin/abuse/health
func (h *AbuseAnalyticsHandler) GetAbuseDetectionHealth(c *gin.Context) {
	ctx := c.Request.Context()

	// Get recent metrics
	scorerMetrics, err := h.anomalyScorer.GetMetrics(ctx)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  "Failed to retrieve metrics",
		})
		return
	}

	// Check if system is processing
	isProcessing := false
	if scorerMetrics != nil {
		// If we have any anomaly counts, system is working
		if anomaliesVote, ok := scorerMetrics["anomalies_vote"].(string); ok && anomaliesVote != "0" {
			isProcessing = true
		}
		if anomaliesFollow, ok := scorerMetrics["anomalies_follow"].(string); ok && anomaliesFollow != "0" {
			isProcessing = true
		}
		if anomaliesSubmission, ok := scorerMetrics["anomalies_submission"].(string); ok && anomaliesSubmission != "0" {
			isProcessing = true
		}
	}

	// Get FPR
	since := time.Now().Add(-24 * time.Hour)
	flaggerStats, _ := h.autoFlagger.GetAutoFlagStats(ctx, since)
	fpr := h.calculateFalsePositiveRate(flaggerStats)

	status := "healthy"
	if fpr > 0.02 {
		status = "degraded" // FPR exceeds target
	}

	response := gin.H{
		"status":                status,
		"processing":            isProcessing,
		"false_positive_rate":   fpr,
		"fpr_target":            0.02,
		"fpr_meets_requirement": fpr <= 0.02,
		"timestamp":             time.Now().UTC(),
	}

	c.JSON(http.StatusOK, response)
}
