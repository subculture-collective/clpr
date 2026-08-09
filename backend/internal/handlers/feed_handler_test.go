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

func TestUpdateFeedRejectsEmptyUpdate(t *testing.T) {
	userID, feedID := uuid.New(), uuid.New()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/users/"+userID.String()+"/feeds/"+feedID.String(), bytes.NewBufferString(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: userID.String()}, {Key: "feedId", Value: feedID.String()}}
	c.Set("user_id", userID)
	(&FeedHandler{}).UpdateFeed(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateFeedRejectsMismatchedRouteUser(t *testing.T) {
	userID := uuid.New()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/users/"+uuid.NewString()+"/feeds", bytes.NewBufferString(`{"name":"feed"}`))
	c.Params = gin.Params{{Key: "id", Value: uuid.NewString()}}
	c.Set("user_id", userID)
	(&FeedHandler{}).CreateFeed(c)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestFeedQueriesRejectInvalidBounds(t *testing.T) {
	h := &FeedHandler{}
	tests := []struct {
		target string
		call   func(*gin.Context)
	}{
		{"/api/v1/feeds/discover?limit=bad", h.DiscoverFeeds},
		{"/api/v1/feeds/search?q=&limit=20", h.SearchFeeds},
		{"/api/v1/feeds/clips?limit=9", h.GetFilteredClips},
		{"/api/v1/feeds/clips?filter%5Bgame%5D=a&filter%5Bgame%5D=b", h.GetFilteredClips},
	}
	for _, tt := range tests {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, tt.target, nil)
		tt.call(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("%s: expected 400, got %d", tt.target, w.Code)
		}
	}
}

func TestListUserFeeds_InvalidUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create handler with nil dependencies (not accessed in this test)
	handler := &FeedHandler{
		feedService: nil,
		authService: nil,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/invalid-uuid/feeds", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: "invalid-uuid"},
	}

	handler.ListUserFeeds(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateFeed_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &FeedHandler{
		feedService: nil,
		authService: nil,
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/550e8400-e29b-41d4-a716-446655440000/feeds", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// Don't set user_id to simulate unauthenticated request

	handler.CreateFeed(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}

	if _, exists := response["error"]; !exists {
		t.Error("expected error field in response")
	}
}

func TestUpdateFeed_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &FeedHandler{
		feedService: nil,
		authService: nil,
	}

	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/550e8400-e29b-41d4-a716-446655440000/feeds/650e8400-e29b-41d4-a716-446655440001", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// Don't set user_id to simulate unauthenticated request

	handler.UpdateFeed(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestDeleteFeed_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &FeedHandler{
		feedService: nil,
		authService: nil,
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/550e8400-e29b-41d4-a716-446655440000/feeds/650e8400-e29b-41d4-a716-446655440001", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// Don't set user_id to simulate unauthenticated request

	handler.DeleteFeed(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAddClipToFeed_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &FeedHandler{
		feedService: nil,
		authService: nil,
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/550e8400-e29b-41d4-a716-446655440000/feeds/650e8400-e29b-41d4-a716-446655440001/clips", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// Don't set user_id to simulate unauthenticated request

	handler.AddClipToFeed(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRemoveClipFromFeed_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &FeedHandler{
		feedService: nil,
		authService: nil,
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/550e8400-e29b-41d4-a716-446655440000/feeds/650e8400-e29b-41d4-a716-446655440001/clips/750e8400-e29b-41d4-a716-446655440002", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// Don't set user_id to simulate unauthenticated request

	handler.RemoveClipFromFeed(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestFollowFeed_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &FeedHandler{
		feedService: nil,
		authService: nil,
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/550e8400-e29b-41d4-a716-446655440000/feeds/650e8400-e29b-41d4-a716-446655440001/follow", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// Don't set user_id to simulate unauthenticated request

	handler.FollowFeed(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestUnfollowFeed_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &FeedHandler{
		feedService: nil,
		authService: nil,
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/550e8400-e29b-41d4-a716-446655440000/feeds/650e8400-e29b-41d4-a716-446655440001/follow", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// Don't set user_id to simulate unauthenticated request

	handler.UnfollowFeed(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}
