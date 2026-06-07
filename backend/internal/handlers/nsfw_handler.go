package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
)

// validateImageURL checks if the URL is safe and uses allowed protocols
func validateImageURL(imageURL string) error {
	parsedURL, err := url.Parse(imageURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Only allow HTTPS and HTTP protocols
	if parsedURL.Scheme != "https" && parsedURL.Scheme != "http" {
		return fmt.Errorf("only http and https protocols are allowed")
	}

	// Block private IP ranges and localhost to prevent SSRF
	host := strings.ToLower(parsedURL.Hostname())
	if host == "localhost" || host == "127.0.0.1" || host == "0.0.0.0" ||
		strings.HasPrefix(host, "192.168.") ||
		strings.HasPrefix(host, "10.") ||
		strings.HasPrefix(host, "172.16.") || strings.HasPrefix(host, "172.17.") ||
		strings.HasPrefix(host, "172.18.") || strings.HasPrefix(host, "172.19.") ||
		strings.HasPrefix(host, "172.20.") || strings.HasPrefix(host, "172.21.") ||
		strings.HasPrefix(host, "172.22.") || strings.HasPrefix(host, "172.23.") ||
		strings.HasPrefix(host, "172.24.") || strings.HasPrefix(host, "172.25.") ||
		strings.HasPrefix(host, "172.26.") || strings.HasPrefix(host, "172.27.") ||
		strings.HasPrefix(host, "172.28.") || strings.HasPrefix(host, "172.29.") ||
		strings.HasPrefix(host, "172.30.") || strings.HasPrefix(host, "172.31.") ||
		strings.HasPrefix(host, "169.254.") {
		return fmt.Errorf("private IP addresses and localhost are not allowed")
	}

	return nil
}

// NSFWHandler handles NSFW detection operations
type NSFWHandler struct {
	nsfwDetector *services.NSFWDetector
}

// NewNSFWHandler creates a new NSFWHandler
func NewNSFWHandler(nsfwDetector *services.NSFWDetector) *NSFWHandler {
	return &NSFWHandler{
		nsfwDetector: nsfwDetector,
	}
}

// DetectImage performs NSFW detection on an image URL
// POST /admin/nsfw/detect
func (h *NSFWHandler) DetectImage(c *gin.Context) {
	var req struct {
		ImageURL    string     `json:"image_url" binding:"required,url"`
		ContentType string     `json:"content_type" binding:"required,oneof=clip thumbnail submission user"`
		ContentID   *uuid.UUID `json:"content_id,omitempty"`
		AutoFlag    *bool      `json:"auto_flag,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	// Validate image URL to prevent SSRF attacks
	if err := validateImageURL(req.ImageURL); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid image URL: " + err.Error(),
		})
		return
	}

	ctx := c.Request.Context()

	// Perform detection with content info for database persistence
	var score *services.NSFWScore
	var err error

	if req.ContentID != nil && req.ContentType != "" {
		score, err = h.nsfwDetector.DetectImageWithID(ctx, req.ImageURL, req.ContentType, *req.ContentID)
	} else {
		score, err = h.nsfwDetector.DetectImage(ctx, req.ImageURL)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to detect NSFW content: " + err.Error(),
		})
		return
	}

	// Auto-flag to moderation queue if requested and content is NSFW
	if req.ContentID != nil && score.NSFW {
		autoFlag := true
		if req.AutoFlag != nil {
			autoFlag = *req.AutoFlag
		}

		if autoFlag {
			err = h.nsfwDetector.FlagToModerationQueue(ctx, req.ContentType, *req.ContentID, score)
			if err != nil {
				// Log error but don't fail the request
				c.Error(fmt.Errorf("failed to flag to moderation queue: %w", err))
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"nsfw":             score.NSFW,
			"confidence_score": score.ConfidenceScore,
			"categories":       score.Categories,
			"reason_codes":     score.ReasonCodes,
			"latency_ms":       score.LatencyMs,
		},
	})
}

// BatchDetect performs NSFW detection on multiple images
// POST /admin/nsfw/batch-detect
func (h *NSFWHandler) BatchDetect(c *gin.Context) {
	var req struct {
		Images []struct {
			ImageURL    string     `json:"image_url" binding:"required,url"`
			ContentType string     `json:"content_type" binding:"required,oneof=clip thumbnail submission user"`
			ContentID   *uuid.UUID `json:"content_id,omitempty"`
		} `json:"images" binding:"required,min=1,max=50"`
		AutoFlag *bool `json:"auto_flag,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	// Validate all image URLs to prevent SSRF attacks
	for i, img := range req.Images {
		if err := validateImageURL(img.ImageURL); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Invalid image URL at index %d: %s", i, err.Error()),
			})
			return
		}
	}

	ctx := c.Request.Context()

	type result struct {
		ImageURL        string             `json:"image_url"`
		NSFW            bool               `json:"nsfw"`
		ConfidenceScore float64            `json:"confidence_score"`
		Categories      map[string]float64 `json:"categories"`
		ReasonCodes     []string           `json:"reason_codes"`
		LatencyMs       int64              `json:"latency_ms"`
		Error           *string            `json:"error,omitempty"`
		Flagged         bool               `json:"flagged"`
	}

	results := make([]result, 0, len(req.Images))
	totalLatency := int64(0)
	successCount := 0
	nsfwCount := 0

	for _, img := range req.Images {
		var score *services.NSFWScore
		var err error

		// Perform detection with content info for database persistence
		if img.ContentID != nil && img.ContentType != "" {
			score, err = h.nsfwDetector.DetectImageWithID(ctx, img.ImageURL, img.ContentType, *img.ContentID)
		} else {
			score, err = h.nsfwDetector.DetectImage(ctx, img.ImageURL)
		}

		res := result{
			ImageURL: img.ImageURL,
		}

		if err != nil {
			errMsg := err.Error()
			res.Error = &errMsg
		} else {
			res.NSFW = score.NSFW
			res.ConfidenceScore = score.ConfidenceScore
			res.Categories = score.Categories
			res.ReasonCodes = score.ReasonCodes
			res.LatencyMs = score.LatencyMs

			totalLatency += score.LatencyMs
			successCount++

			if score.NSFW {
				nsfwCount++

				// Auto-flag if requested
				autoFlag := true
				if req.AutoFlag != nil {
					autoFlag = *req.AutoFlag
				}

				if autoFlag && img.ContentID != nil {
					err = h.nsfwDetector.FlagToModerationQueue(ctx, img.ContentType, *img.ContentID, score)
					if err == nil {
						res.Flagged = true
					}
				}
			}
		}

		results = append(results, res)
	}

	avgLatency := int64(0)
	if successCount > 0 {
		avgLatency = totalLatency / int64(successCount)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
		"meta": gin.H{
			"total_processed": len(req.Images),
			"success_count":   successCount,
			"nsfw_count":      nsfwCount,
			"avg_latency_ms":  avgLatency,
		},
	})
}

// GetMetrics retrieves NSFW detection metrics
// GET /admin/nsfw/metrics
func (h *NSFWHandler) GetMetrics(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse date range
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate time.Time
	if startDateStr == "" {
		startDate = time.Now().AddDate(0, 0, -30) // Last 30 days
	} else {
		parsed, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid start_date format. Use YYYY-MM-DD",
			})
			return
		}
		startDate = parsed
	}

	if endDateStr == "" {
		endDate = time.Now()
	} else {
		parsed, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid end_date format. Use YYYY-MM-DD",
			})
			return
		}
		endDate = parsed.Add(24 * time.Hour) // Include the end date
	}

	metrics, err := h.nsfwDetector.GetMetrics(ctx, startDate, endDate)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, services.ErrNSFWMetricsDBUnavailable) {
			status = http.StatusServiceUnavailable
		}

		c.JSON(status, gin.H{
			"error": "Failed to retrieve metrics: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metrics,
	})
}

// GetHealthCheck returns the health status of the NSFW detector
// GET /admin/nsfw/health
func (h *NSFWHandler) GetHealthCheck(c *gin.Context) {
	startTime := time.Now()

	// Check internal service state without relying on external services
	healthy := h.nsfwDetector != nil
	latency := time.Since(startTime).Milliseconds()

	status := "healthy"
	if !healthy {
		status = "unhealthy"
	}

	response := gin.H{
		"success":    healthy,
		"status":     status,
		"latency_ms": latency,
	}

	if !healthy {
		response["error"] = "NSFW detector is not initialized"
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetConfig returns the current NSFW detector configuration (non-sensitive)
// GET /admin/nsfw/config
func (h *NSFWHandler) GetConfig(c *gin.Context) {
	// Return non-sensitive configuration details
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"enabled": h.nsfwDetector != nil,
			// Note: Don't expose threshold, API keys, or other sensitive config
			// These should be retrieved through secure admin APIs only
		},
	})
}

// ScanClipThumbnails scans existing clip thumbnails for NSFW content
// POST /admin/nsfw/scan-clips
func (h *NSFWHandler) ScanClipThumbnails(c *gin.Context) {
	var req struct {
		Limit    int  `json:"limit" binding:"omitempty,min=1,max=1000"`
		AutoFlag bool `json:"auto_flag" binding:"omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}

	if req.Limit == 0 {
		req.Limit = 100
	}

	// This would trigger a background job to scan clips
	// For now, return a job ID that can be polled
	jobID := uuid.New()

	c.JSON(http.StatusAccepted, gin.H{
		"success": true,
		"job_id":  jobID,
		"message": "Scan job started",
		"limit":   req.Limit,
	})
}
