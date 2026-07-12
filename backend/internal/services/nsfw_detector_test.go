package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNSFWDetector(t *testing.T) {
	detector := NewNSFWDetector(
		"test-api-key",
		"https://api.example.com",
		true,
		0.80,
		true,
		true,
		200,
		5,
		nil,
	)

	assert.NotNil(t, detector)
	assert.True(t, detector.enabled)
	assert.Equal(t, 0.80, detector.threshold)
	assert.True(t, detector.scanThumbnails)
	assert.True(t, detector.autoFlag)
	assert.Equal(t, 200, detector.maxLatencyMs)
}

func TestDetectImage_Disabled(t *testing.T) {
	detector := NewNSFWDetector(
		"",
		"",
		false, // disabled
		0.80,
		true,
		true,
		200,
		5,
		nil,
	)

	ctx := context.Background()
	score, err := detector.DetectImage(ctx, "https://example.com/image.jpg")

	require.ErrorIs(t, err, ErrNSFWDetectorUnavailable)
	assert.Nil(t, score)
}

func TestDetectImage_WithAPI_Safe(t *testing.T) {
	// Mock API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.Header.Get("Authorization"), "Bearer test-api-key")

		response := map[string]interface{}{
			"nudity": map[string]float64{
				"raw":     0.05,
				"safe":    0.95,
				"partial": 0.02,
				"sexual":  0.01,
			},
			"offensive": map[string]float64{
				"prob": 0.03,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	detector := NewNSFWDetector(
		"test-api-key",
		server.URL,
		true,
		0.80,
		true,
		true,
		200,
		5,
		nil,
	)

	ctx := context.Background()
	score, err := detector.DetectImage(ctx, "https://example.com/safe-image.jpg")

	require.NoError(t, err)
	assert.NotNil(t, score)
	assert.False(t, score.NSFW)
	assert.Less(t, score.ConfidenceScore, 0.80)
	assert.GreaterOrEqual(t, score.LatencyMs, int64(0)) // Allow 0 latency for very fast responses
	assert.Contains(t, score.Categories, "nudity_raw")
	assert.Contains(t, score.Categories, "offensive")
}

func TestDetectImage_WithAPI_NSFW(t *testing.T) {
	// Mock API server returning NSFW content
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"nudity": map[string]float64{
				"raw":     0.92,
				"safe":    0.01,
				"partial": 0.15,
				"sexual":  0.88,
			},
			"offensive": map[string]float64{
				"prob": 0.15,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	detector := NewNSFWDetector(
		"test-api-key",
		server.URL,
		true,
		0.80,
		true,
		true,
		200,
		5,
		nil,
	)

	ctx := context.Background()
	score, err := detector.DetectImage(ctx, "https://example.com/nsfw-image.jpg")

	require.NoError(t, err)
	assert.NotNil(t, score)
	assert.True(t, score.NSFW)
	assert.GreaterOrEqual(t, score.ConfidenceScore, 0.80)
	assert.Greater(t, len(score.ReasonCodes), 0)
	assert.Contains(t, score.ReasonCodes, "nudity_explicit")
}

func TestDetectImage_WithAPI_Error(t *testing.T) {
	// Mock API server returning error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
	}))
	defer server.Close()

	detector := NewNSFWDetector(
		"test-api-key",
		server.URL,
		true,
		0.80,
		true,
		true,
		200,
		5,
		nil,
	)

	ctx := context.Background()
	score, err := detector.DetectImage(ctx, "https://example.com/image.jpg")

	assert.Error(t, err)
	assert.Nil(t, score)
	assert.Contains(t, err.Error(), "500")
}

func TestDetectImage_Timeout(t *testing.T) {
	// Mock API server with delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	detector := NewNSFWDetector(
		"test-api-key",
		server.URL,
		true,
		0.80,
		true,
		true,
		200,
		1, // 1 second timeout
		nil,
	)

	ctx := context.Background()
	score, err := detector.DetectImage(ctx, "https://example.com/image.jpg")

	assert.Error(t, err)
	assert.Nil(t, score)
	// Check for common timeout-related errors
	assert.True(t,
		strings.Contains(err.Error(), "timeout") ||
			strings.Contains(err.Error(), "deadline exceeded") ||
			strings.Contains(err.Error(), "Client.Timeout"),
		"Expected timeout-related error, got: %v", err)
}

func TestDetectImage_Latency(t *testing.T) {
	// Mock API server with controlled delay
	delay := 50 * time.Millisecond
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)

		response := map[string]interface{}{
			"nudity": map[string]float64{
				"raw":     0.05,
				"safe":    0.95,
				"partial": 0.02,
				"sexual":  0.01,
			},
			"offensive": map[string]float64{
				"prob": 0.03,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	detector := NewNSFWDetector(
		"test-api-key",
		server.URL,
		true,
		0.80,
		true,
		true,
		200,
		5,
		nil,
	)

	ctx := context.Background()
	score, err := detector.DetectImage(ctx, "https://example.com/image.jpg")

	require.NoError(t, err)
	assert.NotNil(t, score)
	assert.GreaterOrEqual(t, score.LatencyMs, int64(delay.Milliseconds()))
	assert.Less(t, score.LatencyMs, int64(200)) // Should be well under max latency
}

func TestDetectWithoutProviderFailsClosed(t *testing.T) {
	detector := NewNSFWDetector(
		"", // No API key
		"", // No API URL
		true,
		0.80,
		true,
		true,
		200,
		5,
		nil,
	)

	ctx := context.Background()
	score, err := detector.DetectImage(ctx, "https://example.com/image.jpg")

	require.ErrorIs(t, err, ErrNSFWDetectorUnavailable)
	assert.Nil(t, score)
}

func TestReasonCodes(t *testing.T) {
	tests := []struct {
		name          string
		nudityRaw     float64
		nuditySexual  float64
		offensiveProb float64
		threshold     float64
		expectedCodes []string
	}{
		{
			name:          "explicit nudity",
			nudityRaw:     0.95,
			nuditySexual:  0.50,
			offensiveProb: 0.10,
			threshold:     0.80,
			expectedCodes: []string{"nudity_explicit"},
		},
		{
			name:          "sexual content",
			nudityRaw:     0.50,
			nuditySexual:  0.90,
			offensiveProb: 0.10,
			threshold:     0.80,
			expectedCodes: []string{"sexual_content"},
		},
		{
			name:          "offensive content",
			nudityRaw:     0.20,
			nuditySexual:  0.30,
			offensiveProb: 0.85,
			threshold:     0.80,
			expectedCodes: []string{"offensive_content"},
		},
		{
			name:          "multiple violations",
			nudityRaw:     0.92,
			nuditySexual:  0.88,
			offensiveProb: 0.85,
			threshold:     0.80,
			expectedCodes: []string{"nudity_explicit", "sexual_content", "offensive_content"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				response := map[string]interface{}{
					"nudity": map[string]float64{
						"raw":     tt.nudityRaw,
						"safe":    1.0 - tt.nudityRaw,
						"partial": 0.10,
						"sexual":  tt.nuditySexual,
					},
					"offensive": map[string]float64{
						"prob": tt.offensiveProb,
					},
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			detector := NewNSFWDetector(
				"test-api-key",
				server.URL,
				true,
				tt.threshold,
				true,
				true,
				200,
				5,
				nil,
			)

			ctx := context.Background()
			score, err := detector.DetectImage(ctx, "https://example.com/image.jpg")

			require.NoError(t, err)
			assert.NotNil(t, score)

			for _, expectedCode := range tt.expectedCodes {
				assert.Contains(t, score.ReasonCodes, expectedCode)
			}
		})
	}
}

// Note: Tests for FlagToModerationQueue and GetMetrics require database setup
// These would be integration tests rather than unit tests
func TestFlagToModerationQueue_NoAutoFlag(t *testing.T) {
	detector := NewNSFWDetector(
		"test-api-key",
		"https://api.example.com",
		true,
		0.80,
		true,
		false, // autoFlag disabled
		200,
		5,
		nil,
	)

	ctx := context.Background()
	score := &NSFWScore{
		NSFW:            true,
		ConfidenceScore: 0.95,
		ReasonCodes:     []string{"nudity_explicit"},
	}

	err := detector.FlagToModerationQueue(ctx, "clip", uuid.New(), score)
	assert.NoError(t, err) // Should not error, just skip
}

// Benchmark tests
func BenchmarkDetectImage(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"nudity": map[string]float64{
				"raw":     0.05,
				"safe":    0.95,
				"partial": 0.02,
				"sexual":  0.01,
			},
			"offensive": map[string]float64{
				"prob": 0.03,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	detector := NewNSFWDetector(
		"test-api-key",
		server.URL,
		true,
		0.80,
		true,
		true,
		200,
		5,
		nil,
	)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = detector.DetectImage(ctx, "https://example.com/image.jpg")
	}
}
