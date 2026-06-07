package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"git.subcult.tv/subculture-collective/clpr/internal/handlers"
)

// TestSyncBans_Unauthorized tests that syncing bans requires authentication
func TestSyncBans_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	requestBody := map[string]string{
		"channel_id": "123456",
	}
	jsonBody, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/moderation/sync-bans", bytes.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	// Not setting user_id to test authorization

	handler.SyncBans(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestSyncBans_InvalidJSON tests sync-bans with invalid JSON
func TestSyncBans_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/moderation/sync-bans", bytes.NewReader([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", testUserID)

	handler.SyncBans(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestSyncBans_MissingChannelID tests sync-bans with missing channelId
func TestSyncBans_MissingChannelID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()
	requestBody := map[string]string{}
	jsonBody, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/moderation/sync-bans", bytes.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", testUserID)

	handler.SyncBans(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestSyncBans_EmptyChannelID tests sync-bans with empty channelId
func TestSyncBans_EmptyChannelID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()
	requestBody := map[string]string{
		"channel_id": "",
	}
	jsonBody, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/moderation/sync-bans", bytes.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", testUserID)

	handler.SyncBans(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestSyncBans_ServiceUnavailable tests sync-bans when service is not available
func TestSyncBans_ServiceUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testUserID := uuid.New()
	// Create handler without TwitchBanSyncService (nil)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	requestBody := map[string]string{
		"channel_id": "123456",
	}
	jsonBody, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/moderation/sync-bans", bytes.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", testUserID)

	handler.SyncBans(c)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

// TestGetBans_Unauthorized tests that getting bans requires authentication
func TestGetBans_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/moderation/bans?channelId="+uuid.New().String(), nil)
	// Not setting user_id to test authorization

	handler.GetBans(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestGetBans_MissingChannelID tests that getting bans without channelId returns service unavailable when service is nil
// Note: Missing channelId is now allowed (lists all bans for admins), but service must be available
func TestGetBans_MissingChannelID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/moderation/bans", nil)
	c.Set("user_id", testUserID)

	handler.GetBans(c)

	// Service is nil, so we get 503 before reaching the GetAllBans call
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

// TestGetBans_InvalidChannelID tests that getting bans returns service unavailable when service is nil
// Note: Invalid channelId validation happens after service availability check
func TestGetBans_InvalidChannelID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/moderation/bans?channelId=invalid-uuid", nil)
	c.Set("user_id", testUserID)

	handler.GetBans(c)

	// Service is nil, so we get 503 before reaching the channelId validation
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

// TestGetBans_ServiceUnavailable tests getting bans when service is unavailable
func TestGetBans_ServiceUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()
	channelID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/moderation/bans?channelId="+channelID.String(), nil)
	c.Set("user_id", testUserID)

	handler.GetBans(c)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

// TestGetBans_PaginationValidation tests pagination parameter validation
func TestGetBans_PaginationValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()
	channelID := uuid.New()

	tests := []struct {
		name         string
		queryParams  string
		expectedCode int
	}{
		{
			name:         "negative limit defaults to valid value",
			queryParams:  "?channelId=" + channelID.String() + "&limit=-1",
			expectedCode: http.StatusServiceUnavailable, // Will fail on service check, not validation
		},
		{
			name:         "limit over 100 defaults to valid value",
			queryParams:  "?channelId=" + channelID.String() + "&limit=1000",
			expectedCode: http.StatusServiceUnavailable, // Will fail on service check, not validation
		},
		{
			name:         "negative offset",
			queryParams:  "?channelId=" + channelID.String() + "&offset=-5",
			expectedCode: http.StatusServiceUnavailable, // Will fail on service check, not validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/moderation/bans"+tt.queryParams, nil)
			c.Set("user_id", testUserID)

			handler.GetBans(c)

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}

// TestCreateBan_Unauthorized tests that creating a ban requires authentication
func TestCreateBan_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

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

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestCreateBan_InvalidJSON tests that creating a ban validates JSON
func TestCreateBan_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/moderation/ban", bytes.NewReader([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", testUserID)

	handler.CreateBan(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestCreateBan_MissingRequiredFields tests creating a ban with missing required fields
func TestCreateBan_MissingRequiredFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()
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

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestCreateBan_InvalidChannelID tests creating a ban with invalid channelId
func TestCreateBan_InvalidChannelID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()
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

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestCreateBan_ServiceUnavailable tests creating a ban when service is unavailable
func TestCreateBan_ServiceUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()
	requestBody := map[string]interface{}{
		"channelId": uuid.New().String(),
		"userId":    uuid.New().String(),
	}
	jsonBody, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/moderation/ban", bytes.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", testUserID)

	handler.CreateBan(c)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

// TestListModerators_Unauthorized tests that listing moderators requires authentication
func TestListModerators_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/moderation/moderators?channelId="+uuid.New().String(), nil)
	// Not setting user_id to test authorization

	handler.ListModerators(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestListModerators_MissingChannelID tests that listing moderators requires channelId
func TestListModerators_MissingChannelID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/moderation/moderators", nil)
	c.Set("user_id", testUserID)

	handler.ListModerators(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestListModerators_InvalidChannelID tests that listing moderators validates channelId format
func TestListModerators_InvalidChannelID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/moderation/moderators?channelId=invalid-uuid", nil)
	c.Set("user_id", testUserID)

	handler.ListModerators(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestAddModerator_Unauthorized tests that adding moderators requires authentication
func TestAddModerator_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/moderation/moderators", nil)
	// Not setting user_id to test authorization

	handler.AddModerator(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestRemoveModerator_Unauthorized tests that removing moderators requires authentication
func TestRemoveModerator_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/moderation/moderators/"+uuid.New().String(), nil)
	// Not setting user_id to test authorization

	handler.RemoveModerator(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestUpdateModeratorPermissions_Unauthorized tests that updating moderator permissions requires authentication
func TestUpdateModeratorPermissions_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/moderation/moderators/"+uuid.New().String(), nil)
	// Not setting user_id to test authorization

	handler.UpdateModeratorPermissions(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestGetModerationAuditLogs_InvalidModeratorID tests audit logs with invalid moderator ID
func TestGetModerationAuditLogs_InvalidModeratorID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/moderation/audit?moderator_id=invalid-uuid", nil)
	c.Set("user_id", testUserID)

	handler.GetModerationAuditLogs(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestGetModerationAuditLogs_InvalidAction tests audit logs with invalid action type
func TestGetModerationAuditLogs_InvalidAction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/moderation/audit?action=invalid_action", nil)
	c.Set("user_id", testUserID)

	handler.GetModerationAuditLogs(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestGetModerationAnalytics_ServiceUnavailable tests analytics when database is unavailable
func TestGetModerationAnalytics_ServiceUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/moderation/analytics", nil)
	c.Set("user_id", testUserID)

	handler.GetModerationAnalytics(c)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

// TestRevokeBan_Unauthorized tests that revoking bans requires authentication
func TestRevokeBan_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/moderation/ban/"+uuid.New().String(), nil)
	// Not setting user_id to test authorization

	handler.RevokeBan(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestRevokeBan_InvalidBanID tests revoking a ban with invalid ID
func TestRevokeBan_InvalidBanID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/moderation/ban/invalid-uuid", nil)
	c.Set("user_id", testUserID)

	handler.RevokeBan(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestRevokeBan_ServiceUnavailable tests revoking a ban when service is unavailable
func TestRevokeBan_ServiceUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()
	banID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: banID.String()}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/moderation/ban/"+banID.String(), nil)
	c.Set("user_id", testUserID)

	handler.RevokeBan(c)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

// TestGetBanDetails_Unauthorized tests that getting ban details requires authentication
func TestGetBanDetails_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/moderation/ban/"+uuid.New().String(), nil)
	// Not setting user_id to test authorization

	handler.GetBanDetails(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestGetBanDetails_InvalidBanID tests getting ban details with invalid ID
func TestGetBanDetails_InvalidBanID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/moderation/ban/invalid-uuid", nil)
	c.Set("user_id", testUserID)

	handler.GetBanDetails(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestGetBanDetails_ServiceUnavailable tests getting ban details when service is unavailable
func TestGetBanDetails_ServiceUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()
	banID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: banID.String()}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/moderation/ban/"+banID.String(), nil)
	c.Set("user_id", testUserID)

	handler.GetBanDetails(c)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

// ==============================================================================
// Moderation Queue Tests
// ==============================================================================

// TestGetModerationQueue_InvalidStatus tests queue endpoint with invalid status
func TestGetModerationQueue_InvalidStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/moderation/queue?status=invalid", nil)
	c.Set("user_id", testUserID)

	handler.GetModerationQueue(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestGetModerationQueue_InvalidContentType tests queue endpoint with invalid content type
func TestGetModerationQueue_InvalidContentType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/moderation/queue?type=invalid_type", nil)
	c.Set("user_id", testUserID)

	handler.GetModerationQueue(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestApproveContent_Unauthorized tests that approving content requires authentication
func TestApproveContent_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}
	c.Request = httptest.NewRequest(http.MethodPost, "/admin/moderation/"+uuid.New().String()+"/approve", nil)
	// Not setting user_id to test authorization

	handler.ApproveContent(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestApproveContent_InvalidItemID tests approving content with invalid ID
func TestApproveContent_InvalidItemID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}
	c.Request = httptest.NewRequest(http.MethodPost, "/admin/moderation/invalid-uuid/approve", nil)
	c.Set("user_id", testUserID)

	handler.ApproveContent(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestRejectContent_Unauthorized tests that rejecting content requires authentication
func TestRejectContent_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}
	c.Request = httptest.NewRequest(http.MethodPost, "/admin/moderation/"+uuid.New().String()+"/reject", nil)
	// Not setting user_id to test authorization

	handler.RejectContent(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestRejectContent_InvalidItemID tests rejecting content with invalid ID
func TestRejectContent_InvalidItemID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}
	c.Request = httptest.NewRequest(http.MethodPost, "/admin/moderation/invalid-uuid/reject", nil)
	c.Set("user_id", testUserID)

	handler.RejectContent(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestBulkModerate_Unauthorized tests that bulk moderation requires authentication
func TestBulkModerate_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	// Send valid JSON with item IDs since body is checked first
	requestBody := map[string]interface{}{
		"item_ids": []string{uuid.New().String()},
		"action":   "approve",
	}
	jsonBody, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/admin/moderation/bulk", bytes.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	// Not setting user_id to test authorization

	handler.BulkModerate(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestBulkModerate_InvalidJSON tests bulk moderation with invalid JSON
func TestBulkModerate_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/admin/moderation/bulk", bytes.NewReader([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", testUserID)

	handler.BulkModerate(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ==============================================================================
// Moderation Event Tests
// ==============================================================================

// TestMarkEventReviewed_Unauthorized tests that marking event as reviewed requires authentication
func TestMarkEventReviewed_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}
	c.Request = httptest.NewRequest(http.MethodPost, "/admin/moderation/events/"+uuid.New().String()+"/review", nil)
	// Not setting user_id to test authorization

	handler.MarkEventReviewed(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestMarkEventReviewed_InvalidEventID tests marking event as reviewed with invalid ID
func TestMarkEventReviewed_InvalidEventID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}
	c.Request = httptest.NewRequest(http.MethodPost, "/admin/moderation/events/invalid-uuid/review", nil)
	c.Set("user_id", testUserID)

	handler.MarkEventReviewed(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestProcessEvent_Unauthorized tests that processing an event requires authentication
func TestProcessEvent_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}
	c.Request = httptest.NewRequest(http.MethodPost, "/admin/moderation/events/"+uuid.New().String()+"/process", nil)
	// Not setting user_id to test authorization

	handler.ProcessEvent(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestProcessEvent_InvalidEventID tests processing an event with invalid ID
func TestProcessEvent_InvalidEventID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}
	c.Request = httptest.NewRequest(http.MethodPost, "/admin/moderation/events/invalid-uuid/process", nil)
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", testUserID)

	handler.ProcessEvent(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestProcessEvent_MissingAction tests processing an event without action
func TestProcessEvent_MissingAction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()
	eventID := uuid.New()
	requestBody := map[string]string{}
	jsonBody, _ := json.Marshal(requestBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: eventID.String()}}
	c.Request = httptest.NewRequest(http.MethodPost, "/admin/moderation/events/"+eventID.String()+"/process", bytes.NewReader(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", testUserID)

	handler.ProcessEvent(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestGetUserAbuseStats_InvalidUserID tests abuse stats with invalid user ID
func TestGetUserAbuseStats_InvalidUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "userId", Value: "invalid-uuid"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/moderation/abuse/invalid-uuid", nil)
	c.Set("user_id", testUserID)

	handler.GetUserAbuseStats(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ==============================================================================
// Appeal Tests
// ==============================================================================

// TestCreateAppeal_Unauthorized tests that creating an appeal requires authentication
func TestCreateAppeal_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/moderation/appeals", nil)
	// Not setting user_id to test authorization

	handler.CreateAppeal(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestCreateAppeal_InvalidJSON tests creating an appeal with invalid JSON
func TestCreateAppeal_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/moderation/appeals", bytes.NewReader([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", testUserID)

	handler.CreateAppeal(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestGetAppeals_InvalidStatus tests getting appeals with invalid status
func TestGetAppeals_InvalidStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/moderation/appeals?status=invalid", nil)
	c.Set("user_id", testUserID)

	handler.GetAppeals(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestResolveAppeal_Unauthorized tests that resolving an appeal requires authentication
func TestResolveAppeal_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}
	c.Request = httptest.NewRequest(http.MethodPost, "/admin/moderation/appeals/"+uuid.New().String()+"/resolve", nil)
	// Not setting user_id to test authorization

	handler.ResolveAppeal(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestResolveAppeal_InvalidAppealID tests resolving an appeal with invalid ID
func TestResolveAppeal_InvalidAppealID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}
	c.Request = httptest.NewRequest(http.MethodPost, "/admin/moderation/appeals/invalid-uuid/resolve", nil)
	c.Set("user_id", testUserID)

	handler.ResolveAppeal(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestGetUserAppeals_Unauthorized tests that getting user's appeals requires authentication
func TestGetUserAppeals_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/moderation/appeals", nil)
	// Not setting user_id to test authorization

	handler.GetUserAppeals(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ==============================================================================
// Toxicity Metrics Tests
// ==============================================================================

// TestGetToxicityMetrics_InvalidDateFormat tests toxicity metrics with invalid date format
func TestGetToxicityMetrics_InvalidDateFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/moderation/toxicity/metrics?start_date=invalid-date", nil)
	c.Set("user_id", testUserID)

	handler.GetToxicityMetrics(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestGetToxicityMetrics_ServiceUnavailable tests toxicity metrics when classifier is unavailable
func TestGetToxicityMetrics_ServiceUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := handlers.NewModerationHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	testUserID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/moderation/toxicity/metrics", nil)
	c.Set("user_id", testUserID)

	handler.GetToxicityMetrics(c)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

// ==============================================================================
// Additional Moderation Queue and Event Tests
// ==============================================================================
