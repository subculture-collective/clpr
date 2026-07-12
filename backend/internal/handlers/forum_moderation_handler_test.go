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
