package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TestCreateChannel_Unauthorized tests that unauthorized requests are rejected
func TestCreateChannel_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &ChatHandler{
		db: nil,
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat/channels", strings.NewReader(`{"name":"Test Channel"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// Note: not setting user_id in context to simulate unauthorized

	handler.CreateChannel(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestCreateChannel_InvalidRequest tests that invalid requests are rejected
func TestCreateChannel_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &ChatHandler{
		db: nil,
	}

	userID := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat/channels", strings.NewReader(`{"name":""}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)

	handler.CreateChannel(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}

	if _, ok := response["error"]; !ok {
		t.Error("expected error field in response")
	}
}

func TestListChannels_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &ChatHandler{db: nil}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/chat/channels", http.NoBody)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.ListChannels(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestListChannels_RejectsMalformedParameters(t *testing.T) {
	tests := []string{
		"?limit=zero",
		"?limit=101",
		"?offset=-1",
		"?type=direct",
	}
	for _, query := range tests {
		t.Run(query, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			handler := &ChatHandler{db: nil}
			req := httptest.NewRequest(http.MethodGet, "/api/v1/chat/channels"+query, http.NoBody)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Set("user_id", uuid.New())

			handler.ListChannels(c)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
			}
		})
	}
}

func TestAuthenticatedUserID_RejectsMalformedContextIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "not-a-uuid")

	if _, ok := authenticatedUserID(c); ok {
		t.Fatal("expected malformed context identity to be rejected")
	}
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestIsChatStaffRole(t *testing.T) {
	for _, role := range []string{"owner", "admin", "moderator"} {
		if !isChatStaffRole(role) {
			t.Fatalf("expected %q to be a staff role", role)
		}
	}
	for _, role := range []string{"", "member", "administrator"} {
		if isChatStaffRole(role) {
			t.Fatalf("expected %q not to be a staff role", role)
		}
	}
}

// TestGetChannel_InvalidChannelID tests that invalid channel IDs are rejected
func TestGetChannel_InvalidChannelID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &ChatHandler{
		db: nil,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/chat/channels/not-a-uuid", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: "not-a-uuid"},
	}

	handler.GetChannel(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}

	if _, ok := response["error"]; !ok {
		t.Error("expected error field in response")
	}
}

// TestUpdateChannel_InvalidChannelID tests that invalid channel IDs are rejected
func TestUpdateChannel_InvalidChannelID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &ChatHandler{
		db: nil,
	}

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/chat/channels/not-a-uuid", strings.NewReader(`{"name":"Updated"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: "not-a-uuid"},
	}

	handler.UpdateChannel(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}

	if _, ok := response["error"]; !ok {
		t.Error("expected error field in response")
	}
}

// TestUpdateChannel_Unauthorized tests that unauthorized requests are rejected
func TestUpdateChannel_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &ChatHandler{
		db: nil,
	}

	channelID := uuid.New().String()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/chat/channels/"+channelID, strings.NewReader(`{"name":"Updated"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: channelID},
	}
	// Note: not setting user_id in context to simulate unauthorized

	handler.UpdateChannel(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestBanUser_InvalidChannelID tests that invalid channel IDs are rejected
func TestBanUser_InvalidChannelID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &ChatHandler{
		db: nil, // nil is ok since we never get to the DB call
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat/channels/not-a-uuid/ban", strings.NewReader(`{"user_id":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: "not-a-uuid"},
	}

	handler.BanUser(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}

	if _, ok := response["error"]; !ok {
		t.Error("expected error field in response")
	}
}

// TestBanUser_Unauthorized tests that unauthorized requests are rejected
func TestBanUser_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &ChatHandler{
		db: nil,
	}

	channelID := uuid.New().String()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat/channels/"+channelID+"/ban", strings.NewReader(`{"user_id":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: channelID},
	}
	// Note: not setting user_id in context to simulate unauthorized

	handler.BanUser(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestDeleteMessage_InvalidMessageID tests that invalid message IDs are rejected
func TestDeleteMessage_InvalidMessageID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &ChatHandler{
		db: nil,
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/chat/messages/not-a-uuid", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: "not-a-uuid"},
	}

	handler.DeleteMessage(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}

	if _, ok := response["error"]; !ok {
		t.Error("expected error field in response")
	}
}

// TestMuteUser_InvalidChannelID tests that invalid channel IDs are rejected
func TestMuteUser_InvalidChannelID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &ChatHandler{
		db: nil,
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat/channels/invalid/mute", strings.NewReader(`{"user_id":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: "invalid"},
	}

	handler.MuteUser(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestTimeoutUser_InvalidChannelID tests that invalid channel IDs are rejected
func TestTimeoutUser_InvalidChannelID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &ChatHandler{
		db: nil,
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat/channels/invalid/timeout", strings.NewReader(`{"user_id":"test","duration_minutes":10}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: "invalid"},
	}

	handler.TimeoutUser(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestGetModerationLog_InvalidChannelID tests that invalid channel IDs are rejected
func TestGetModerationLog_InvalidChannelID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &ChatHandler{
		db: nil,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/chat/channels/invalid/moderation-log", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: "invalid"},
	}

	handler.GetModerationLog(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestCheckUserBan_InvalidChannelID tests that invalid channel IDs are rejected
func TestCheckUserBan_InvalidChannelID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &ChatHandler{
		db: nil,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/chat/channels/invalid/check-ban?user_id=test", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: "invalid"},
	}

	handler.CheckUserBan(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestAddChannelMember_Unauthorized tests that unauthorized requests are rejected
func TestAddChannelMember_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &ChatHandler{
		db: nil,
	}

	channelID := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat/channels/"+channelID.String()+"/members",
		strings.NewReader(`{"user_id":"`+uuid.New().String()+`"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: channelID.String()},
	}
	// Note: not setting user_id in context to simulate unauthorized

	handler.AddChannelMember(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestRemoveChannelMember_InvalidChannelID tests that invalid channel IDs are rejected
func TestRemoveChannelMember_InvalidChannelID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &ChatHandler{
		db: nil,
	}

	userID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/chat/channels/not-a-uuid/members/"+uuid.New().String(), http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)
	c.Params = gin.Params{
		{Key: "id", Value: "not-a-uuid"},
		{Key: "user_id", Value: uuid.New().String()},
	}

	handler.RemoveChannelMember(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}

	if _, ok := response["error"]; !ok {
		t.Error("expected error field in response")
	}
}

// TestDeleteChannel_Unauthorized tests that unauthorized requests are rejected
func TestDeleteChannel_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &ChatHandler{
		db: nil,
	}

	channelID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/chat/channels/"+channelID.String(), http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: channelID.String()},
	}
	// Note: not setting user_id in context to simulate unauthorized

	handler.DeleteChannel(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestGetCurrentUserRole_Unauthorized tests that unauthorized requests are rejected
func TestGetCurrentUserRole_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &ChatHandler{
		db: nil,
	}

	channelID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/chat/channels/"+channelID.String()+"/role", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: channelID.String()},
	}
	// Note: not setting user_id in context to simulate unauthorized

	handler.GetCurrentUserRole(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestGetCurrentUserRole_InvalidChannelID tests that invalid channel IDs are rejected
func TestGetCurrentUserRole_InvalidChannelID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &ChatHandler{
		db: nil,
	}

	userID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/chat/channels/not-a-uuid/role", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", userID)
	c.Params = gin.Params{
		{Key: "id", Value: "not-a-uuid"},
	}

	handler.GetCurrentUserRole(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}

	if _, ok := response["error"]; !ok {
		t.Error("expected error field in response")
	}
}
