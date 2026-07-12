package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupNSFWHandlerTest(t *testing.T) (*gin.Engine, *NSFWHandler) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"nudity":{"raw":0.01,"safe":0.99,"partial":0.01,"sexual":0.01},"offensive":{"prob":0.01}}`))
	}))
	t.Cleanup(provider.Close)

	detector := services.NewNSFWDetector(
		"test-key",
		provider.URL,
		true,
		0.80,
		true,
		true,
		200,
		5,
		nil,
	)

	handler := NewNSFWHandler(detector)
	return router, handler
}

func TestDetectImage_Success(t *testing.T) {
	router, handler := setupNSFWHandlerTest(t)
	router.POST("/detect", handler.DetectImage)

	requestBody := map[string]interface{}{
		"image_url":    "https://example.com/image.jpg",
		"content_type": "thumbnail",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/detect", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	assert.Contains(t, data, "nsfw")
	assert.Contains(t, data, "confidence_score")
	assert.Contains(t, data, "latency_ms")
}

func TestDetectImage_InvalidURL(t *testing.T) {
	router, handler := setupNSFWHandlerTest(t)
	router.POST("/detect", handler.DetectImage)

	requestBody := map[string]interface{}{
		"image_url":    "not-a-url",
		"content_type": "thumbnail",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/detect", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
}

func TestValidateImageURLRejectsUnsafeDestinations(t *testing.T) {
	for _, imageURL := range []string{
		"http://example.com/image.jpg",
		"https://localhost/image.jpg",
		"https://127.0.0.1/image.jpg",
		"https://[::1]/image.jpg",
		"https://user:password@example.com/image.jpg",
	} {
		t.Run(imageURL, func(t *testing.T) {
			assert.Error(t, validateImageURL(imageURL))
		})
	}
	assert.NoError(t, validateImageURL("https://cdn.example.com/image.jpg"))
}

func TestDetectImage_InvalidContentType(t *testing.T) {
	router, handler := setupNSFWHandlerTest(t)
	router.POST("/detect", handler.DetectImage)

	requestBody := map[string]interface{}{
		"image_url":    "https://example.com/image.jpg",
		"content_type": "invalid",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/detect", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDetectImage_WithContentID(t *testing.T) {
	router, handler := setupNSFWHandlerTest(t)
	router.POST("/detect", handler.DetectImage)

	contentID := uuid.New()
	requestBody := map[string]interface{}{
		"image_url":    "https://example.com/image.jpg",
		"content_type": "clip",
		"content_id":   contentID.String(),
		"auto_flag":    false,
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/detect", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestBatchDetect_Success(t *testing.T) {
	router, handler := setupNSFWHandlerTest(t)
	router.POST("/batch-detect", handler.BatchDetect)

	requestBody := map[string]interface{}{
		"images": []map[string]interface{}{
			{
				"image_url":    "https://example.com/image1.jpg",
				"content_type": "thumbnail",
			},
			{
				"image_url":    "https://example.com/image2.jpg",
				"content_type": "clip",
			},
		},
		"auto_flag": false,
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/batch-detect", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])
	assert.NotNil(t, response["meta"])

	data := response["data"].([]interface{})
	assert.Len(t, data, 2)

	meta := response["meta"].(map[string]interface{})
	assert.Equal(t, float64(2), meta["total_processed"])
	assert.Contains(t, meta, "avg_latency_ms")
}

func TestBatchDetect_EmptyImages(t *testing.T) {
	router, handler := setupNSFWHandlerTest(t)
	router.POST("/batch-detect", handler.BatchDetect)

	requestBody := map[string]interface{}{
		"images": []map[string]interface{}{},
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/batch-detect", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBatchDetect_TooManyImages(t *testing.T) {
	router, handler := setupNSFWHandlerTest(t)
	router.POST("/batch-detect", handler.BatchDetect)

	// Create more than 50 images
	images := make([]map[string]interface{}, 51)
	for i := 0; i < 51; i++ {
		images[i] = map[string]interface{}{
			"image_url":    "https://example.com/image.jpg",
			"content_type": "thumbnail",
		}
	}

	requestBody := map[string]interface{}{
		"images": images,
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/batch-detect", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetMetrics_Success(t *testing.T) {
	router, handler := setupNSFWHandlerTest(t)
	router.GET("/metrics", handler.GetMetrics)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Failed to retrieve NSFW metrics", response["error"])
}

func TestGetMetrics_WithDateRange(t *testing.T) {
	router, handler := setupNSFWHandlerTest(t)
	router.GET("/metrics", handler.GetMetrics)

	startDate := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	endDate := time.Now().Format("2006-01-02")

	req := httptest.NewRequest("GET", "/metrics?start_date="+startDate+"&end_date="+endDate, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestGetMetrics_InvalidDateFormat(t *testing.T) {
	router, handler := setupNSFWHandlerTest(t)
	router.GET("/metrics", handler.GetMetrics)

	req := httptest.NewRequest("GET", "/metrics?start_date=invalid-date", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetMetrics_RejectsReversedDateRange(t *testing.T) {
	router, handler := setupNSFWHandlerTest(t)
	router.GET("/metrics", handler.GetMetrics)

	req := httptest.NewRequest("GET", "/metrics?start_date=2026-07-12&end_date=2026-07-01", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetHealthCheck_Success(t *testing.T) {
	router, handler := setupNSFWHandlerTest(t)
	router.GET("/health", handler.GetHealthCheck)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
	assert.Equal(t, "healthy", response["status"])
	assert.Contains(t, response, "latency_ms")
}

func TestGetConfig_Success(t *testing.T) {
	router, handler := setupNSFWHandlerTest(t)
	router.GET("/config", handler.GetConfig)

	req := httptest.NewRequest("GET", "/config", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
	assert.NotNil(t, response["data"])

	data := response["data"].(map[string]interface{})
	assert.Contains(t, data, "enabled")
}
