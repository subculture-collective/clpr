package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TestGetBans_Unauthorized tests that getting bans requires authentication
func TestGetBans_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/moderation/bans?channelId="+uuid.New().String(), nil)
	// Not setting user_id to test authorization

	handler.GetBans(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d for unauthorized request, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestGetBans_MissingChannelID tests that getting bans without channelId returns service unavailable when service is nil
// Note: Missing channelId is now allowed (lists all bans for admins), but service must be available
func TestGetBans_MissingChannelID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUserID := uuid.New()
	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/moderation/bans", nil)
	c.Set("user_id", testUserID)

	handler.GetBans(c)

	// Service is nil, so we get 503 before reaching the GetAllBans call
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d for nil service, got %d", http.StatusServiceUnavailable, w.Code)
	}
}

// TestGetBans_InvalidChannelID tests that getting bans returns service unavailable when service is nil
// Note: Invalid channelId validation happens after service availability check
func TestGetBans_InvalidChannelID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUserID := uuid.New()
	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/moderation/bans?channelId=invalid-uuid", nil)
	c.Set("user_id", testUserID)

	handler.GetBans(c)

	// Service is nil, so we get 503 before reaching the channelId validation
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d for nil service, got %d", http.StatusServiceUnavailable, w.Code)
	}
}

// TestCreateBan_Unauthorized tests that creating a ban requires authentication
func TestCreateBan_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	requestBody := map[string]interface{}{
		"channelId": uuid.New().String(),
		"userId":    uuid.New().String(),
		"reason":    "Test reason",
	}
	jsonBody, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/moderation/ban", bytes.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	// Not setting user_id to test authorization

	handler.CreateBan(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d for unauthorized request, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestCreateBan_InvalidAuthenticationContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)
	body, _ := json.Marshal(map[string]interface{}{
		"channelId": uuid.New().String(),
		"userId":    uuid.New().String(),
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/moderation/ban", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", "not-a-uuid")

	handler.CreateBan(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestCreateBan_InvalidJSON tests that creating a ban validates JSON
func TestCreateBan_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUserID := uuid.New()
	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/moderation/ban", bytes.NewReader([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", testUserID)

	handler.CreateBan(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for invalid JSON, got %d", http.StatusBadRequest, w.Code)
	}
	if bytes.Contains(w.Body.Bytes(), []byte("invalid character")) {
		t.Errorf("response leaked parser details: %s", w.Body.String())
	}
}

func TestCreateBan_ReasonTooLong(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)
	body, _ := json.Marshal(map[string]interface{}{
		"channelId": uuid.New().String(),
		"userId":    uuid.New().String(),
		"reason":    string(make([]byte, 1001)),
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/moderation/ban", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", uuid.New())

	handler.CreateBan(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestCreateBan_MissingFields tests that creating a ban validates required fields
func TestCreateBan_MissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUserID := uuid.New()
	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	requestBody := map[string]interface{}{
		"channelId": uuid.New().String(),
		// Missing userId
	}
	jsonBody, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/moderation/ban", bytes.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", testUserID)

	handler.CreateBan(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for missing required fields, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestCreateBan_InvalidChannelID tests that creating a ban validates channelId format
func TestCreateBan_InvalidChannelID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUserID := uuid.New()
	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	requestBody := map[string]interface{}{
		"channelId": "invalid-uuid",
		"userId":    uuid.New().String(),
	}
	jsonBody, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/moderation/ban", bytes.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", testUserID)

	handler.CreateBan(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for invalid channelId, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestCreateBan_InvalidUserID tests that creating a ban validates userId format
func TestCreateBan_InvalidUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUserID := uuid.New()
	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	requestBody := map[string]interface{}{
		"channelId": uuid.New().String(),
		"userId":    "invalid-uuid",
	}
	jsonBody, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/moderation/ban", bytes.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", testUserID)

	handler.CreateBan(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for invalid userId, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestRevokeBan_Unauthorized tests that revoking a ban requires authentication
func TestRevokeBan_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/moderation/ban/"+uuid.New().String(), nil)
	// Not setting user_id to test authorization

	handler.RevokeBan(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d for unauthorized request, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestRevokeBan_InvalidBanID tests that revoking a ban validates ban ID format
func TestRevokeBan_InvalidBanID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUserID := uuid.New()
	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/moderation/ban/invalid-uuid", nil)
	c.Set("user_id", testUserID)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "invalid-uuid"}}

	handler.RevokeBan(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for invalid ban ID, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestGetBanDetails_Unauthorized tests that getting ban details requires authentication
func TestGetBanDetails_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/moderation/ban/"+uuid.New().String(), nil)
	// Not setting user_id to test authorization

	handler.GetBanDetails(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d for unauthorized request, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestGetBanDetails_InvalidBanID tests that getting ban details validates ban ID format
func TestGetBanDetails_InvalidBanID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUserID := uuid.New()
	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/moderation/ban/invalid-uuid", nil)
	c.Set("user_id", testUserID)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "invalid-uuid"}}

	handler.GetBanDetails(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for invalid ban ID, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestGetBans_ServiceUnavailable tests that GetBans handles service unavailability
func TestGetBans_ServiceUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUserID := uuid.New()
	// Pass nil for moderation service to simulate unavailability
	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/moderation/bans?channelId="+uuid.New().String(), nil)
	c.Set("user_id", testUserID)

	handler.GetBans(c)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d for service unavailable, got %d", http.StatusServiceUnavailable, w.Code)
	}
}

// TestCreateBan_ServiceUnavailable tests that CreateBan handles service unavailability
func TestCreateBan_ServiceUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUserID := uuid.New()
	// Pass nil for moderation service to simulate unavailability
	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	requestBody := map[string]interface{}{
		"channelId": uuid.New().String(),
		"userId":    uuid.New().String(),
		"reason":    "Test reason",
	}
	jsonBody, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/moderation/ban", bytes.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", testUserID)

	handler.CreateBan(c)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d for service unavailable, got %d", http.StatusServiceUnavailable, w.Code)
	}
}

// TestRevokeBan_ServiceUnavailable tests that RevokeBan handles service unavailability
func TestRevokeBan_ServiceUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUserID := uuid.New()
	// Pass nil for moderation service to simulate unavailability
	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/moderation/ban/"+uuid.New().String(), nil)
	c.Set("user_id", testUserID)
	c.Params = gin.Params{gin.Param{Key: "id", Value: uuid.New().String()}}

	handler.RevokeBan(c)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d for service unavailable, got %d", http.StatusServiceUnavailable, w.Code)
	}
}

// TestGetBanDetails_ServiceUnavailable tests that GetBanDetails handles service unavailability
func TestGetBanDetails_ServiceUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUserID := uuid.New()
	// Pass nil for moderation service to simulate unavailability
	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/moderation/ban/"+uuid.New().String(), nil)
	c.Set("user_id", testUserID)
	c.Params = gin.Params{gin.Param{Key: "id", Value: uuid.New().String()}}

	handler.GetBanDetails(c)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d for service unavailable, got %d", http.StatusServiceUnavailable, w.Code)
	}
}
