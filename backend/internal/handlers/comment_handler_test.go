package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"github.com/gin-gonic/gin"
)

// TestListComments_InvalidClipID tests that invalid clip IDs are rejected
func TestListComments_InvalidClipID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &CommentHandler{
		commentService: nil, // nil is ok since we never get to the service call
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clips/not-a-uuid/comments", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: "not-a-uuid"},
	}

	handler.ListComments(c)

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

// TestGetReplies_InvalidCommentID tests the GetReplies endpoint with invalid comment ID
func TestGetReplies_InvalidCommentID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &CommentHandler{
		commentService: nil,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/comments/not-a-uuid/replies", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: "not-a-uuid"},
	}

	handler.GetReplies(c)

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

func TestWriteCreatorModerationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	httpRecorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(httpRecorder)

	if handled := writeCreatorModerationError(c, &services.CreatorModerationError{Message: "restricted"}); !handled {
		t.Fatal("expected moderation error to be handled")
	}
	if httpRecorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", httpRecorder.Code, http.StatusForbidden)
	}

	var response map[string]any
	if err := json.Unmarshal(httpRecorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if response["error"] != "restricted" {
		t.Fatalf("error = %v, want restricted", response["error"])
	}
}
