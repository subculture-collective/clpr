package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
)

// mockTwitchModerationService implements a mock for testing
type mockTwitchModerationService struct {
	banUserOnTwitchFunc   func(moderatorUserID uuid.UUID, broadcasterID string, targetUserID string, reason *string, duration *int) error
	unbanUserOnTwitchFunc func(moderatorUserID uuid.UUID, broadcasterID string, targetUserID string) error
}

func (m *mockTwitchModerationService) BanUserOnTwitch(ctx context.Context, moderatorUserID uuid.UUID, broadcasterID string, targetUserID string, reason *string, duration *int) error {
	if m.banUserOnTwitchFunc != nil {
		return m.banUserOnTwitchFunc(moderatorUserID, broadcasterID, targetUserID, reason, duration)
	}
	return nil
}

func (m *mockTwitchModerationService) UnbanUserOnTwitch(ctx context.Context, moderatorUserID uuid.UUID, broadcasterID string, targetUserID string) error {
	if m.unbanUserOnTwitchFunc != nil {
		return m.unbanUserOnTwitchFunc(moderatorUserID, broadcasterID, targetUserID)
	}
	return nil
}

// TestTwitchBanUser_Success tests successful Twitch ban
func TestTwitchBanUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUserID := uuid.New()
	broadcasterID := "12345"
	targetUserID := "67890"

	mockService := &mockTwitchModerationService{
		banUserOnTwitchFunc: func(moderatorUserID uuid.UUID, bID string, tID string, reason *string, duration *int) error {
			if moderatorUserID != testUserID {
				t.Errorf("Expected moderator ID %s, got %s", testUserID, moderatorUserID)
			}
			if bID != broadcasterID {
				t.Errorf("Expected broadcaster ID %s, got %s", broadcasterID, bID)
			}
			if tID != targetUserID {
				t.Errorf("Expected target user ID %s, got %s", targetUserID, tID)
			}
			return nil
		},
	}

	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)
	handler.SetTwitchModerationService(mockService)

	requestBody := map[string]interface{}{
		"broadcasterID": broadcasterID,
		"userID":        targetUserID,
		"reason":        "Test reason",
	}
	jsonBody, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/moderation/twitch/ban", bytes.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", testUserID)

	handler.TwitchBanUser(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

// TestTwitchBanUser_SiteModeratorDenied tests that site moderators are denied
func TestTwitchBanUser_SiteModeratorDenied(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUserID := uuid.New()

	mockService := &mockTwitchModerationService{
		banUserOnTwitchFunc: func(moderatorUserID uuid.UUID, bID string, tID string, reason *string, duration *int) error {
			return services.ErrSiteModeratorsReadOnly
		},
	}

	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)
	handler.SetTwitchModerationService(mockService)

	requestBody := map[string]interface{}{
		"broadcasterID": "12345",
		"userID":        "67890",
	}
	jsonBody, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/moderation/twitch/ban", bytes.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", testUserID)

	handler.TwitchBanUser(c)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d for site moderator, got %d", http.StatusForbidden, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if code, ok := response["code"].(string); !ok || code != "SITE_MODERATORS_READ_ONLY" {
		t.Errorf("Expected error code SITE_MODERATORS_READ_ONLY, got %v", response["code"])
	}
}

// TestTwitchBanUser_NotAuthenticated tests that non-authenticated users are denied
func TestTwitchBanUser_NotAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUserID := uuid.New()

	mockService := &mockTwitchModerationService{
		banUserOnTwitchFunc: func(moderatorUserID uuid.UUID, bID string, tID string, reason *string, duration *int) error {
			return services.ErrTwitchNotAuthenticated
		},
	}

	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)
	handler.SetTwitchModerationService(mockService)

	requestBody := map[string]interface{}{
		"broadcasterID": "12345",
		"userID":        "67890",
	}
	jsonBody, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/moderation/twitch/ban", bytes.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", testUserID)

	handler.TwitchBanUser(c)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d for not authenticated, got %d", http.StatusForbidden, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if code, ok := response["code"].(string); !ok || code != "NOT_AUTHENTICATED" {
		t.Errorf("Expected error code NOT_AUTHENTICATED, got %v", response["code"])
	}
}

// TestTwitchBanUser_InsufficientScopes tests that users without scopes are denied
func TestTwitchBanUser_InsufficientScopes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUserID := uuid.New()

	mockService := &mockTwitchModerationService{
		banUserOnTwitchFunc: func(moderatorUserID uuid.UUID, bID string, tID string, reason *string, duration *int) error {
			return services.ErrTwitchScopeInsufficient
		},
	}

	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)
	handler.SetTwitchModerationService(mockService)

	requestBody := map[string]interface{}{
		"broadcasterID": "12345",
		"userID":        "67890",
	}
	jsonBody, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/moderation/twitch/ban", bytes.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", testUserID)

	handler.TwitchBanUser(c)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d for insufficient scopes, got %d", http.StatusForbidden, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if code, ok := response["code"].(string); !ok || code != "INSUFFICIENT_SCOPES" {
		t.Errorf("Expected error code INSUFFICIENT_SCOPES, got %v", response["code"])
	}
}

// TestTwitchBanUser_NotBroadcaster tests that non-broadcasters are denied
func TestTwitchBanUser_NotBroadcaster(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUserID := uuid.New()

	mockService := &mockTwitchModerationService{
		banUserOnTwitchFunc: func(moderatorUserID uuid.UUID, bID string, tID string, reason *string, duration *int) error {
			return services.ErrTwitchNotBroadcaster
		},
	}

	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)
	handler.SetTwitchModerationService(mockService)

	requestBody := map[string]interface{}{
		"broadcasterID": "12345",
		"userID":        "67890",
	}
	jsonBody, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/moderation/twitch/ban", bytes.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", testUserID)

	handler.TwitchBanUser(c)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d for not broadcaster, got %d", http.StatusForbidden, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if code, ok := response["code"].(string); !ok || code != "NOT_BROADCASTER" {
		t.Errorf("Expected error code NOT_BROADCASTER, got %v", response["code"])
	}
}

// TestTwitchUnbanUser_Success tests successful Twitch unban
func TestTwitchUnbanUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUserID := uuid.New()
	broadcasterID := "12345"
	targetUserID := "67890"

	mockService := &mockTwitchModerationService{
		unbanUserOnTwitchFunc: func(moderatorUserID uuid.UUID, bID string, tID string) error {
			if moderatorUserID != testUserID {
				t.Errorf("Expected moderator ID %s, got %s", testUserID, moderatorUserID)
			}
			if bID != broadcasterID {
				t.Errorf("Expected broadcaster ID %s, got %s", broadcasterID, bID)
			}
			if tID != targetUserID {
				t.Errorf("Expected target user ID %s, got %s", targetUserID, tID)
			}
			return nil
		},
	}

	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)
	handler.SetTwitchModerationService(mockService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/moderation/twitch/ban?broadcasterID="+broadcasterID+"&userID="+targetUserID, nil)
	c.Set("user_id", testUserID)

	handler.TwitchUnbanUser(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

// TestTwitchUnbanUser_SiteModeratorDenied tests that site moderators are denied
func TestTwitchUnbanUser_SiteModeratorDenied(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUserID := uuid.New()

	mockService := &mockTwitchModerationService{
		unbanUserOnTwitchFunc: func(moderatorUserID uuid.UUID, bID string, tID string) error {
			return services.ErrSiteModeratorsReadOnly
		},
	}

	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)
	handler.SetTwitchModerationService(mockService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/moderation/twitch/ban?broadcasterID=12345&userID=67890", nil)
	c.Set("user_id", testUserID)

	handler.TwitchUnbanUser(c)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d for site moderator, got %d", http.StatusForbidden, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if code, ok := response["code"].(string); !ok || code != "SITE_MODERATORS_READ_ONLY" {
		t.Errorf("Expected error code SITE_MODERATORS_READ_ONLY, got %v", response["code"])
	}
}

// TestTwitchBanUser_ServiceNotConfigured tests that service unavailable is returned when not configured
func TestTwitchBanUser_ServiceNotConfigured(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUserID := uuid.New()

	handler := NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)
	// Don't set Twitch moderation service

	requestBody := map[string]interface{}{
		"broadcasterID": "12345",
		"userID":        "67890",
	}
	jsonBody, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/moderation/twitch/ban", bytes.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", testUserID)

	handler.TwitchBanUser(c)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d when service not configured, got %d", http.StatusServiceUnavailable, w.Code)
	}
}

// Suppress unused errors warning
var _ = errors.New("")
