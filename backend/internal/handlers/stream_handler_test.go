package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// TestGetStreamStatus_MissingStreamer tests request with missing streamer parameter
func TestGetStreamStatus_MissingStreamer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewStreamHandler(nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/streams/", nil)

	handler.GetStreamStatus(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for missing streamer, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if _, ok := response["error"]; !ok {
		t.Error("Expected error field in response")
	}
}

// TestStreamInfoStructure tests that StreamInfo model has correct structure
func TestStreamInfoStructure(t *testing.T) {
	streamInfo := models.StreamInfo{
		StreamerUsername: "teststreamer",
		IsLive:           true,
		ViewerCount:      100,
	}

	if streamInfo.StreamerUsername != "teststreamer" {
		t.Errorf("Expected streamer_username to be 'teststreamer', got %s", streamInfo.StreamerUsername)
	}

	if !streamInfo.IsLive {
		t.Error("Expected is_live to be true")
	}

	if streamInfo.ViewerCount != 100 {
		t.Errorf("Expected viewer_count to be 100, got %d", streamInfo.ViewerCount)
	}
}

// TestStreamHandler_Initialization tests that handler initializes correctly
func TestStreamHandler_Initialization(t *testing.T) {
	handler := NewStreamHandler(nil, nil, nil, nil, nil)

	if handler == nil {
		t.Error("Expected handler to be created")
	}

	if handler.twitchClient != nil {
		t.Error("Expected twitchClient to be nil in test setup")
	}

	if handler.streamRepo != nil {
		t.Error("Expected streamRepo to be nil in test setup")
	}

	if handler.clipRepo != nil {
		t.Error("Expected clipRepo to be nil in test setup")
	}

	if handler.streamFollowRepo != nil {
		t.Error("Expected streamFollowRepo to be nil in test setup")
	}

	if handler.jobService != nil {
		t.Error("Expected jobService to be nil in test setup")
	}
}

// TestFollowStreamer_MissingStreamer tests follow request with missing streamer parameter
func TestFollowStreamer_MissingStreamer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewStreamHandler(nil, nil, nil, nil, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/streams//follow", nil)
	c.Params = gin.Params{gin.Param{Key: "streamer", Value: ""}}

	handler.FollowStreamer(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for missing streamer, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestFollowStreamer_InvalidUsername tests validation of streamer username
func TestFollowStreamer_InvalidUsername(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		streamer       string
		expectedStatus int
	}{
		{
			name:           "Username too short",
			streamer:       "abc",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Username too long",
			streamer:       "thisusernameiswaytooolongfortwitch",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid characters - @",
			streamer:       "user@name",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewStreamHandler(nil, nil, nil, nil, nil)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/streams/"+tt.streamer+"/follow", nil)
			c.Params = gin.Params{gin.Param{Key: "streamer", Value: tt.streamer}}

			handler.FollowStreamer(c)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// TestValidateStreamerUsername tests the validation helper directly
func TestValidateStreamerUsername(t *testing.T) {
	tests := []struct {
		name      string
		username  string
		wantError bool
	}{
		{"Valid 4 chars", "test", false},
		{"Valid 25 chars", "TwentyFiveCharUsername1", false},
		{"Valid with underscore", "test_user", false},
		{"Too short", "abc", true},
		{"Too long", "thisusernameiswaytooolongfortwitch", true},
		{"Invalid char @", "user@name", true},
		{"Empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStreamerUsername(tt.username)
			if (err != nil) != tt.wantError {
				t.Errorf("validateStreamerUsername(%s) error = %v, wantError %v", tt.username, err, tt.wantError)
			}
		})
	}
}
