package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// MockApplicationLogRepository is a mock implementation of the repository for testing
type MockApplicationLogRepository struct {
	CreateFunc func(ctx context.Context, log *models.ApplicationLog) error
}

func (m *MockApplicationLogRepository) Create(ctx context.Context, log *models.ApplicationLog) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, log)
	}
	return nil
}

func (m *MockApplicationLogRepository) DeleteOldLogs(ctx context.Context, retentionDays int) (int64, error) {
	return 0, nil
}

func (m *MockApplicationLogRepository) GetLogStats(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"total_logs":     100,
		"unique_users":   10,
		"error_count":    20,
		"warn_count":     30,
		"info_count":     40,
		"debug_count":    10,
		"logs_last_hour": 50,
		"logs_last_24h":  100,
	}, nil
}

// TestCreateLog_Success tests successful log creation
func TestCreateLog_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockApplicationLogRepository{
		CreateFunc: func(ctx context.Context, log *models.ApplicationLog) error {
			// Verify log fields
			if log.Level != "error" {
				t.Errorf("expected level 'error', got '%s'", log.Level)
			}
			if log.Message != "Test error message" {
				t.Errorf("expected message 'Test error message', got '%s'", log.Message)
			}
			if log.Service != "clpr-frontend" {
				t.Errorf("expected service 'clpr-frontend', got '%s'", log.Service)
			}
			return nil
		},
	}

	handler := NewApplicationLogHandler(mockRepo)

	logPayload := map[string]interface{}{
		"level":     "error",
		"message":   "Test error message",
		"timestamp": time.Now().Format(time.RFC3339),
		"platform":  "web",
	}

	body, _ := json.Marshal(logPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/logs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/api/v1/logs", handler.CreateLog)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d, body: %s", http.StatusNoContent, w.Code, w.Body.String())
	}
}

// TestCreateLog_InvalidLevel tests that invalid log levels are rejected
func TestCreateLog_InvalidLevel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockApplicationLogRepository{}
	handler := NewApplicationLogHandler(mockRepo)

	logPayload := map[string]interface{}{
		"level":   "invalid",
		"message": "Test message",
	}

	body, _ := json.Marshal(logPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/logs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/api/v1/logs", handler.CreateLog)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestCreateLog_MissingMessage tests that missing message is rejected
func TestCreateLog_MissingMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockApplicationLogRepository{}
	handler := NewApplicationLogHandler(mockRepo)

	logPayload := map[string]interface{}{
		"level": "error",
		// Missing message
	}

	body, _ := json.Marshal(logPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/logs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/api/v1/logs", handler.CreateLog)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestCreateLog_SensitiveDataFiltering tests that sensitive data is filtered
func TestCreateLog_SensitiveDataFiltering(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var capturedLog *models.ApplicationLog

	mockRepo := &MockApplicationLogRepository{
		CreateFunc: func(ctx context.Context, log *models.ApplicationLog) error {
			capturedLog = log
			return nil
		},
	}

	handler := NewApplicationLogHandler(mockRepo)

	logPayload := map[string]interface{}{
		"level":   "error",
		"message": "Login failed for password: secret123",
		"context": map[string]interface{}{
			"password": "secret123",
			"token":    "abc123",
			"username": "testuser",
		},
	}

	body, _ := json.Marshal(logPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/logs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/api/v1/logs", handler.CreateLog)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}

	if capturedLog == nil {
		t.Fatal("log was not captured")
	}

	// Verify sensitive data was filtered
	if capturedLog.Message != "[REDACTED - contains sensitive data]" {
		t.Errorf("expected message to be redacted, got '%s'", capturedLog.Message)
	}

	// Verify context fields were filtered
	var contextMap map[string]interface{}
	if err := json.Unmarshal(capturedLog.Context, &contextMap); err != nil {
		t.Fatalf("failed to unmarshal context: %v", err)
	}

	if contextMap["password"] != "[REDACTED]" {
		t.Errorf("expected password to be redacted, got '%v'", contextMap["password"])
	}
	if contextMap["token"] != "[REDACTED]" {
		t.Errorf("expected token to be redacted, got '%v'", contextMap["token"])
	}
	if contextMap["username"] != "testuser" {
		t.Errorf("expected username to be preserved, got '%v'", contextMap["username"])
	}
}

// TestGetLogStats_Success tests successful retrieval of log stats
func TestGetLogStats_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &MockApplicationLogRepository{}
	handler := NewApplicationLogHandler(mockRepo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/logs/stats", http.NoBody)
	w := httptest.NewRecorder()

	router := gin.New()
	router.GET("/api/v1/logs/stats", handler.GetLogStats)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !response["success"].(bool) {
		t.Error("expected success to be true")
	}

	stats := response["stats"].(map[string]interface{})
	if stats["total_logs"].(float64) != 100 {
		t.Errorf("expected total_logs to be 100, got %v", stats["total_logs"])
	}
}
