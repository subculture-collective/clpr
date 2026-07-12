package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	nsfwMetricsOnce      sync.Once
	nsfwDetectionCounter *prometheus.CounterVec
	nsfwLatencyHistogram prometheus.Histogram
	nsfwFlaggedCounter   *prometheus.CounterVec
	nsfwErrorCounter     *prometheus.CounterVec

	ErrNSFWMetricsDBUnavailable = errors.New("nsfw metrics database not configured")
	ErrNSFWDetectorUnavailable  = errors.New("nsfw detector is disabled or provider is not configured")
)

// NSFWDetector handles NSFW detection for images/thumbnails
type NSFWDetector struct {
	httpClient     *http.Client
	apiKey         string
	apiURL         string
	enabled        bool
	threshold      float64
	scanThumbnails bool
	autoFlag       bool
	maxLatencyMs   int
	db             *pgxpool.Pool
}

// Operational reports whether real provider-backed detection is available.
func (nd *NSFWDetector) Operational() bool {
	return nd != nil && nd.enabled && nd.apiKey != "" && nd.apiURL != ""
}

// Enabled reports the configured feature flag without exposing provider secrets.
func (nd *NSFWDetector) Enabled() bool {
	return nd != nil && nd.enabled
}

// NSFWScore represents the result of NSFW detection
type NSFWScore struct {
	NSFW            bool               `json:"nsfw"`
	ConfidenceScore float64            `json:"confidence_score"`
	Categories      map[string]float64 `json:"categories"`
	ReasonCodes     []string           `json:"reason_codes"`
	LatencyMs       int64              `json:"latency_ms"`
}

// NewNSFWDetector creates a new NSFWDetector
func NewNSFWDetector(
	apiKey, apiURL string,
	enabled bool,
	threshold float64,
	scanThumbnails, autoFlag bool,
	maxLatencyMs, timeoutSeconds int,
	db *pgxpool.Pool,
) *NSFWDetector {
	detector := &NSFWDetector{
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		},
		apiKey:         apiKey,
		apiURL:         apiURL,
		enabled:        enabled,
		threshold:      threshold,
		scanThumbnails: scanThumbnails,
		autoFlag:       autoFlag,
		maxLatencyMs:   maxLatencyMs,
		db:             db,
	}

	// Initialize Prometheus metrics once
	detector.initMetrics()

	return detector
}

// initMetrics initializes Prometheus metrics using sync.Once
func (nd *NSFWDetector) initMetrics() {
	nsfwMetricsOnce.Do(func() {
		nsfwDetectionCounter = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "nsfw_detection_total",
				Help: "Total number of NSFW detections performed",
			},
			[]string{"result"},
		)

		nsfwLatencyHistogram = promauto.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "nsfw_detection_latency_ms",
				Help:    "Latency of NSFW detection in milliseconds",
				Buckets: []float64{10, 25, 50, 100, 150, 200, 300, 500, 1000},
			},
		)

		nsfwFlaggedCounter = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "nsfw_content_flagged_total",
				Help: "Total number of NSFW content flagged",
			},
			[]string{"content_type"},
		)

		nsfwErrorCounter = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "nsfw_detection_errors_total",
				Help: "Total number of NSFW detection errors",
			},
			[]string{"error_type"},
		)
	})
}

// DetectImage analyzes an image for NSFW content
func (nd *NSFWDetector) DetectImage(ctx context.Context, imageURL string) (*NSFWScore, error) {
	return nd.DetectImageWithID(ctx, imageURL, "", uuid.Nil)
}

// DetectImageWithID analyzes an image for NSFW content and optionally persists to database
func (nd *NSFWDetector) DetectImageWithID(ctx context.Context, imageURL string, contentType string, contentID uuid.UUID) (*NSFWScore, error) {
	startTime := time.Now()

	if !nd.Operational() {
		return nil, ErrNSFWDetectorUnavailable
	}

	var score *NSFWScore
	var err error

	score, err = nd.detectWithAPI(ctx, imageURL)

	// Record latency
	latencyMs := time.Since(startTime).Milliseconds()
	if nsfwLatencyHistogram != nil {
		nsfwLatencyHistogram.Observe(float64(latencyMs))
	}

	if score != nil {
		score.LatencyMs = latencyMs
	}

	// Record detection result
	if err != nil {
		if nsfwErrorCounter != nil {
			nsfwErrorCounter.WithLabelValues("detection_error").Inc()
		}
		if nsfwDetectionCounter != nil {
			nsfwDetectionCounter.WithLabelValues("error").Inc()
		}
		return nil, err
	}

	if nsfwDetectionCounter != nil {
		if score.NSFW {
			nsfwDetectionCounter.WithLabelValues("nsfw").Inc()
		} else {
			nsfwDetectionCounter.WithLabelValues("safe").Inc()
		}
	}

	// Persist detection to database if content info provided and database available
	if nd.db != nil && contentType != "" && contentID != uuid.Nil {
		if err := nd.persistDetection(ctx, imageURL, contentType, contentID, score); err != nil {
			// Log error but don't fail the detection
			if nsfwErrorCounter != nil {
				nsfwErrorCounter.WithLabelValues("persist_error").Inc()
			}
		}
	}

	return score, nil
}

// persistDetection saves detection results to the database
func (nd *NSFWDetector) persistDetection(ctx context.Context, imageURL, contentType string, contentID uuid.UUID, score *NSFWScore) error {
	reasonCodes := score.ReasonCodes
	if reasonCodes == nil {
		reasonCodes = []string{}
	}

	categoriesJSON, err := json.Marshal(score.Categories)
	if err != nil {
		return fmt.Errorf("failed to marshal categories: %w", err)
	}

	query := `
		INSERT INTO nsfw_detection_metrics (
			content_type, content_id, image_url, nsfw, confidence_score,
			categories, reason_codes, latency_ms, detected_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
	`

	_, err = nd.db.Exec(ctx, query,
		contentType,
		contentID,
		imageURL,
		score.NSFW,
		score.ConfidenceScore,
		categoriesJSON,
		reasonCodes,
		score.LatencyMs,
	)

	return err
}

// detectWithAPI uses an external API for NSFW detection
func (nd *NSFWDetector) detectWithAPI(ctx context.Context, imageURL string) (*NSFWScore, error) {
	// Generic API request structure - adapt based on actual provider
	// This example uses a generic format that can be adapted for:
	// - Sightengine API
	// - AWS Rekognition
	// - Google Cloud Vision AI
	// - Azure Content Moderator

	requestBody := map[string]interface{}{
		"url":    imageURL,
		"models": []string{"nudity", "offensive"},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", nd.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+nd.apiKey)

	resp, err := nd.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read body once
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response - adapt based on actual provider
	var apiResponse struct {
		Nudity struct {
			Raw     float64 `json:"raw"`
			Safe    float64 `json:"safe"`
			Partial float64 `json:"partial"`
			Sexual  float64 `json:"sexual"`
		} `json:"nudity"`
		Offensive struct {
			Prob float64 `json:"prob"`
		} `json:"offensive"`
	}

	if err := json.Unmarshal(bodyBytes, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Calculate overall score using built-in max (available since Go 1.21)
	maxScore := max(
		apiResponse.Nudity.Raw,
		apiResponse.Nudity.Sexual,
		apiResponse.Offensive.Prob,
	)

	score := &NSFWScore{
		NSFW:            maxScore >= nd.threshold,
		ConfidenceScore: maxScore,
		Categories: map[string]float64{
			"nudity_raw":     apiResponse.Nudity.Raw,
			"nudity_safe":    apiResponse.Nudity.Safe,
			"nudity_partial": apiResponse.Nudity.Partial,
			"nudity_sexual":  apiResponse.Nudity.Sexual,
			"offensive":      apiResponse.Offensive.Prob,
		},
		ReasonCodes: []string{},
	}

	// Build reason codes
	if apiResponse.Nudity.Raw >= nd.threshold {
		score.ReasonCodes = append(score.ReasonCodes, "nudity_explicit")
	}
	if apiResponse.Nudity.Sexual >= nd.threshold {
		score.ReasonCodes = append(score.ReasonCodes, "sexual_content")
	}
	if apiResponse.Offensive.Prob >= nd.threshold {
		score.ReasonCodes = append(score.ReasonCodes, "offensive_content")
	}

	return score, nil
}

// FlagToModerationQueue flags NSFW content to the moderation queue
func (nd *NSFWDetector) FlagToModerationQueue(ctx context.Context, contentType string, contentID uuid.UUID, score *NSFWScore) error {
	if !nd.autoFlag {
		return nil
	}

	// Build reason string
	reason := "nsfw_detected"
	if len(score.ReasonCodes) > 0 {
		reason = "nsfw_" + score.ReasonCodes[0]
	}

	// Calculate priority based on confidence score (higher confidence = higher priority)
	priority := int(score.ConfidenceScore * 100)

	// Prepare NSFW metadata for moderation queue
	nsfwCategoriesJSON, err := json.Marshal(score.Categories)
	if err != nil {
		if nsfwErrorCounter != nil {
			nsfwErrorCounter.WithLabelValues("categories_marshal_error").Inc()
		}
		return fmt.Errorf("failed to marshal NSFW category scores: %w", err)
	}

	detectedAt := time.Now()

	// Insert into moderation queue
	query := `
		INSERT INTO moderation_queue (
			content_type, content_id, reason, priority,
			auto_flagged, confidence_score, status,
			nsfw_categories, nsfw_detected_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, 'pending', $7, $8)
		ON CONFLICT (content_type, content_id) WHERE status = 'pending'
		DO UPDATE SET
			confidence_score = GREATEST(moderation_queue.confidence_score, EXCLUDED.confidence_score),
			priority = GREATEST(moderation_queue.priority, EXCLUDED.priority),
			nsfw_categories = EXCLUDED.nsfw_categories,
			nsfw_detected_at = EXCLUDED.nsfw_detected_at
	`

	_, err = nd.db.Exec(ctx, query,
		contentType,
		contentID,
		reason,
		priority,
		true,
		score.ConfidenceScore,
		nsfwCategoriesJSON,
		detectedAt,
	)
	if err != nil {
		if nsfwErrorCounter != nil {
			nsfwErrorCounter.WithLabelValues("queue_insert_error").Inc()
		}
		return fmt.Errorf("failed to flag to moderation queue: %w", err)
	}

	if nsfwFlaggedCounter != nil {
		nsfwFlaggedCounter.WithLabelValues(contentType).Inc()
	}

	return nil
}

// GetMetrics retrieves NSFW detection metrics
func (nd *NSFWDetector) GetMetrics(ctx context.Context, startDate, endDate time.Time) (map[string]interface{}, error) {
	if nd.db == nil {
		return nil, ErrNSFWMetricsDBUnavailable
	}

	metrics := make(map[string]interface{})

	// Get detection counts by result
	rows, err := nd.db.Query(ctx, `
		SELECT
			reason,
			COUNT(*) as count,
			AVG(confidence_score) as avg_confidence
		FROM moderation_queue
		WHERE reason IN ('nsfw_detected', 'nsfw_nudity_explicit', 'nsfw_sexual_content', 'nsfw_offensive_content')
			AND created_at >= $1
			AND created_at < $2
		GROUP BY reason
		ORDER BY count DESC
	`, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}
	defer rows.Close()

	detectionsByReason := make(map[string]map[string]interface{})
	totalDetections := 0

	for rows.Next() {
		var reason string
		var count int
		var avgConfidence float64

		if err := rows.Scan(&reason, &count, &avgConfidence); err != nil {
			// Log error and continue to avoid partial data
			if nsfwErrorCounter != nil {
				nsfwErrorCounter.WithLabelValues("metrics_scan_error").Inc()
			}
			continue
		}

		detectionsByReason[reason] = map[string]interface{}{
			"count":          count,
			"avg_confidence": avgConfidence,
		}
		totalDetections += count
	}

	metrics["total_detections"] = totalDetections
	metrics["detections_by_reason"] = detectionsByReason
	metrics["start_date"] = startDate.Format("2006-01-02")
	metrics["end_date"] = endDate.Format("2006-01-02")

	// Get average review time
	var avgReviewMinutes *float64
	err = nd.db.QueryRow(ctx, `
		SELECT AVG(EXTRACT(EPOCH FROM (reviewed_at - created_at)) / 60)
		FROM moderation_queue
		WHERE reason IN ('nsfw_detected', 'nsfw_nudity_explicit', 'nsfw_sexual_content', 'nsfw_offensive_content')
			AND reviewed_at IS NOT NULL
			AND created_at >= $1
			AND created_at < $2
	`, startDate, endDate).Scan(&avgReviewMinutes)
	if err == nil && avgReviewMinutes != nil {
		metrics["avg_review_time_minutes"] = *avgReviewMinutes
	}

	return metrics, nil
}
