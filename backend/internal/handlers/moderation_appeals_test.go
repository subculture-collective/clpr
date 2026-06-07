package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// TestCreateAppeal_Unauthorized tests that creating an appeal requires authentication
func TestCreateAppeal_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	requestBody := models.CreateAppealRequest{
		ModerationActionID: uuid.New().String(),
		Reason:             "Test appeal reason",
	}
	jsonBody, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/moderation/appeals", bytes.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	// Not setting user_id to test authorization

	handler.CreateAppeal(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d for unauthorized request, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestCreateAppeal_InvalidJSON tests appeal creation with invalid JSON
func TestCreateAppeal_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUserID := uuid.New()
	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/moderation/appeals", bytes.NewReader([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", testUserID)

	handler.CreateAppeal(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for invalid JSON, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestCreateAppeal_InvalidModerationActionID tests appeal creation with invalid moderation action ID
func TestCreateAppeal_InvalidModerationActionID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUserID := uuid.New()
	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	requestBody := models.CreateAppealRequest{
		ModerationActionID: "invalid-uuid",
		Reason:             "Test appeal reason with at least 10 characters",
	}
	jsonBody, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/moderation/appeals", bytes.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", testUserID)

	handler.CreateAppeal(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for invalid moderation action ID, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestGetAppeals_InvalidStatus tests getting appeals with invalid status parameter
func TestGetAppeals_InvalidStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUserID := uuid.New()
	// Pass nil for db pool since we're testing validation before DB access
	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/admin/moderation/appeals?status=invalid", nil)
	c.Set("user_id", testUserID)

	handler.GetAppeals(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for invalid status parameter, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestResolveAppeal_Unauthorized tests that resolving an appeal requires authentication
func TestResolveAppeal_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	requestBody := models.ResolveAppealRequest{
		Decision: "approve",
	}
	jsonBody, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/moderation/appeals/"+uuid.New().String()+"/resolve", bytes.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}
	// Not setting user_id to test authorization

	handler.ResolveAppeal(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d for unauthorized request, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestResolveAppeal_InvalidAppealID tests resolving an appeal with invalid appeal ID
func TestResolveAppeal_InvalidAppealID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUserID := uuid.New()
	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	requestBody := models.ResolveAppealRequest{
		Decision: "approve",
	}
	jsonBody, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/moderation/appeals/invalid-id/resolve", bytes.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "invalid-id"}}
	c.Set("user_id", testUserID)

	handler.ResolveAppeal(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for invalid appeal ID, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestResolveAppeal_InvalidJSON tests resolving an appeal with invalid JSON
func TestResolveAppeal_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUserID := uuid.New()
	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/admin/moderation/appeals/"+uuid.New().String()+"/resolve", bytes.NewReader([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}
	c.Set("user_id", testUserID)

	handler.ResolveAppeal(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for invalid JSON, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestGetUserAppeals_Unauthorized tests that getting user appeals requires authentication
func TestGetUserAppeals_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/moderation/appeals", nil)
	// Not setting user_id to test authorization

	handler.GetUserAppeals(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d for unauthorized request, got %d", http.StatusUnauthorized, w.Code)
	}
}
