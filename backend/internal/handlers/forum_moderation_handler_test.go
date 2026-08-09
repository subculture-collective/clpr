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

func TestFlagContentRejectsMalformedIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &ForumModerationHandler{db: nil}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/forum/flag", strings.NewReader(`{"target_type":"thread","target_id":"`+uuid.NewString()+`","reason":"spam"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", "not-a-uuid")

	handler.FlagContent(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestFlagContentRejectsInvalidBoundedRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &ForumModerationHandler{db: nil}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/forum/flag", strings.NewReader(`{"target_type":"clip","target_id":"`+uuid.NewString()+`","reason":"spam"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", uuid.New())

	handler.FlagContent(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestLockThread_InvalidThreadID tests that invalid thread IDs are rejected
func TestLockThread_InvalidThreadID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &ForumModerationHandler{
		db: nil, // nil is ok since we never get to the DB call
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/forum/threads/not-a-uuid/lock", strings.NewReader(`{"reason":"test","locked":true}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: "not-a-uuid"},
	}

	handler.LockThread(c)

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

// TestLockThread_Unauthorized tests that unauthorized requests are rejected
func TestLockThread_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &ForumModerationHandler{
		db: nil, // nil is ok since we never get to the DB call
	}

	threadID := uuid.New().String()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/forum/threads/"+threadID+"/lock", strings.NewReader(`{"reason":"test","locked":true}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: threadID},
	}
	// Don't set user_id in context - should fail with unauthorized

	handler.LockThread(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}

	if _, ok := response["error"]; !ok {
		t.Error("expected error field in response")
	}
}

func TestForumAdminMutationsRejectMalformedIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		path string
		body string
		call func(*ForumModerationHandler, *gin.Context)
	}{
		{"/api/v1/admin/forum/threads/" + uuid.NewString() + "/lock", `{"locked":true}`, (*ForumModerationHandler).LockThread},
		{"/api/v1/admin/forum/threads/" + uuid.NewString() + "/pin", `{"pinned":true}`, (*ForumModerationHandler).PinThread},
		{"/api/v1/admin/forum/threads/" + uuid.NewString() + "/delete", `{"reason":"remove thread"}`, (*ForumModerationHandler).DeleteThread},
	} {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodPost, tc.path, strings.NewReader(tc.body))
		ctx.Request.Header.Set("Content-Type", "application/json")
		ctx.Params = gin.Params{{Key: "id", Value: uuid.NewString()}}
		ctx.Set("user_id", "not-a-uuid")
		tc.call(&ForumModerationHandler{}, ctx)
		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("%s: expected 401, got %d", tc.path, recorder.Code)
		}
	}
}

func TestForumAdminListsRejectMalformedFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		path string
		call func(*ForumModerationHandler, *gin.Context)
	}{
		{"/api/v1/admin/forum/flagged?status=unknown", (*ForumModerationHandler).GetFlaggedContent},
		{"/api/v1/admin/forum/flagged?limit=zero", (*ForumModerationHandler).GetFlaggedContent},
		{"/api/v1/admin/forum/moderation-log?action_type=delete_all", (*ForumModerationHandler).GetModerationLog},
		{"/api/v1/admin/forum/moderation-log?target_type=clip", (*ForumModerationHandler).GetModerationLog},
		{"/api/v1/admin/forum/bans?active=yes", (*ForumModerationHandler).GetUserBans},
		{"/api/v1/admin/forum/bans?limit=101", (*ForumModerationHandler).GetUserBans},
	} {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodGet, tc.path, http.NoBody)
		tc.call(&ForumModerationHandler{}, ctx)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("%s: expected 400, got %d", tc.path, recorder.Code)
		}
	}
}

// TestBanUser_InvalidUserID tests that invalid user IDs are rejected
func TestBanUser_InvalidUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &ForumModerationHandler{
		db: nil, // nil is ok since we never get to the DB call
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/forum/users/not-a-uuid/ban", strings.NewReader(`{"reason":"test","duration_days":7}`))
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

// TestDeleteThread_InvalidThreadID tests that invalid thread IDs are rejected
func TestDeleteThread_InvalidThreadID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &ForumModerationHandler{
		db: nil, // nil is ok since we never get to the DB call
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/forum/threads/not-a-uuid/delete", strings.NewReader(`{"reason":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: "not-a-uuid"},
	}

	handler.DeleteThread(c)

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
