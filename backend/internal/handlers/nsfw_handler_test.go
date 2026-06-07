package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
)

func setupNSFWHandlerTest() (*gin.Engine, *NSFWHandler) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create mock NSFW detector with external API disabled to avoid network calls
	detector := services.NewNSFWDetector(
		"",
		"",
		false,
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
	router, handler := setupNSFWHandlerTest()
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
	router, handler := setupNSFWHandlerTest()
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

func TestDetectImage_InvalidContentType(t *testing.T) {
	router, handler := setupNSFWHandlerTest()
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
	router, handler := setupNSFWHandlerTest()
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
	router, handler := setupNSFWHandlerTest()
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
	router, handler := setupNSFWHandlerTest()
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
	router, handler := setupNSFWHandlerTest()
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
	router, handler := setupNSFWHandlerTest()
	router.GET("/metrics", handler.GetMetrics)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response["error"], "database not configured")
}

func TestGetMetrics_WithDateRange(t *testing.T) {
	router, handler := setupNSFWHandlerTest()
	router.GET("/metrics", handler.GetMetrics)

	startDate := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	endDate := time.Now().Format("2006-01-02")

	req := httptest.NewRequest("GET", "/metrics?start_date="+startDate+"&end_date="+endDate, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestGetMetrics_InvalidDateFormat(t *testing.T) {
	router, handler := setupNSFWHandlerTest()
	router.GET("/metrics", handler.GetMetrics)

	req := httptest.NewRequest("GET", "/metrics?start_date=invalid-date", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetHealthCheck_Success(t *testing.T) {
	router, handler := setupNSFWHandlerTest()
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
	router, handler := setupNSFWHandlerTest()
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

func TestScanClipThumbnails_Success(t *testing.T) {
	router, handler := setupNSFWHandlerTest()
	router.POST("/scan-clips", handler.ScanClipThumbnails)

	requestBody := map[string]interface{}{
		"limit":     100,
		"auto_flag": true,
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/scan-clips", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response["success"].(bool))
	assert.Contains(t, response, "job_id")
	assert.Equal(t, float64(100), response["limit"])
}

func TestScanClipThumbnails_InvalidLimit(t *testing.T) {
	router, handler := setupNSFWHandlerTest()
	router.POST("/scan-clips", handler.ScanClipThumbnails)

	requestBody := map[string]interface{}{
		"limit":     2000, // Exceeds max of 1000
		"auto_flag": true,
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/scan-clips", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
