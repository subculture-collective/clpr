package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TestListClips_InvalidSubmittedByUserID tests that invalid UUIDs in submitted_by_user_id parameter are rejected
func TestListClips_InvalidSubmittedByUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &ClipHandler{
		clipService: nil, // nil is ok since we never get to the service call with invalid UUID
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clips?submitted_by_user_id=not-a-uuid", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.ListClips(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response StandardResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}

	if response.Success {
		t.Error("expected Success to be false")
	}

	if response.Error == nil {
		t.Error("expected Error to be set")
	} else if response.Error.Code != "INVALID_UUID" {
		t.Errorf("expected error code INVALID_UUID, got %s", response.Error.Code)
	}
}

func TestClipHandler_AppMediaURLDefaultsToRelativeAPIPath(t *testing.T) {
	clipID := uuid.MustParse("00000000-0000-0000-0000-000000000123")
	handler := &ClipHandler{}

	got := handler.appMediaURL(clipID)
	want := "/api/v1/clips/00000000-0000-0000-0000-000000000123/media"
	if got != want {
		t.Fatalf("appMediaURL() = %q, want %q", got, want)
	}
}

func TestClipHandler_AppMediaURLUsesConfiguredPublicBase(t *testing.T) {
	clipID := uuid.MustParse("00000000-0000-0000-0000-000000000123")
	handler := &ClipHandler{clipConfig: &config.ClipConfig{MediaPublicBaseURL: "https://clpr.example.invalid/api/v1/clips/"}}

	got := handler.appMediaURL(clipID)
	want := "https://clpr.example.invalid/api/v1/clips/00000000-0000-0000-0000-000000000123/media"
	if got != want {
		t.Fatalf("appMediaURL() = %q, want %q", got, want)
	}
}

func TestClipHandler_ApplyAppMediaURLRewritesDirectMediaOnly(t *testing.T) {
	clipID := uuid.MustParse("00000000-0000-0000-0000-000000000123")
	videoURL := "https://clips-s3.example.invalid/clpr-clips/path/to/video.mp4"
	clip := &services.ClipWithUserData{Clip: models.Clip{ID: clipID, VideoURL: &videoURL}}
	handler := &ClipHandler{clipConfig: &config.ClipConfig{MediaPublicBaseURL: "https://clpr.example.invalid/api/v1/clips"}}

	handler.applyAppMediaURL(clip)

	want := "https://clpr.example.invalid/api/v1/clips/00000000-0000-0000-0000-000000000123/media"
	if clip.VideoURL == nil || *clip.VideoURL != want {
		t.Fatalf("VideoURL = %v, want %q", clip.VideoURL, want)
	}
}

func TestClipHandler_OriginMediaURLCanonicalizesStorageBase(t *testing.T) {
	videoURL := "https://clips-s3.example.invalid/clpr-clips/path/to/video.mp4"
	clip := &services.ClipWithUserData{Clip: models.Clip{VideoURL: &videoURL}}
	handler := &ClipHandler{clipConfig: &config.ClipConfig{StoragePublicBaseURL: "https://clips-s3.example.invalid/clpr-clips/"}}

	got := handler.originMediaURL(clip)
	want := "https://clips-s3.example.invalid/clpr-clips/path/to/video.mp4"
	if got != want {
		t.Fatalf("originMediaURL() = %q, want %q", got, want)
	}
}

func TestStorageObjectKey(t *testing.T) {
	tests := []struct {
		name string
		url  string
		base string
		want string
	}{
		{
			name: "matching storage URL",
			url:  "https://clips-s3.example.invalid/clpr-clips/path/to/video.mp4",
			base: "https://clips-s3.example.invalid/clpr-clips",
			want: "path/to/video.mp4",
		},
		{
			name: "encoded key",
			url:  "https://clips-s3.example.invalid/clpr-clips/path%20with%20spaces/video.mp4",
			base: "https://clips-s3.example.invalid/clpr-clips",
			want: "path with spaces/video.mp4",
		},
		{
			name: "different host",
			url:  "https://other.example.invalid/clpr-clips/video.mp4",
			base: "https://clips-s3.example.invalid/clpr-clips",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := storageObjectKey(tt.url, tt.base); got != tt.want {
				t.Fatalf("storageObjectKey() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestListClips_MultipleInvalidSubmittedByUserIDs tests various invalid UUID formats
func TestListClips_MultipleInvalidSubmittedByUserIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &ClipHandler{
		clipService: nil,
	}

	testCases := []string{
		"not-a-uuid",
		"12345",
		"invalid-uuid-format",
		"zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz",
	}

	for _, invalidUUID := range testCases {
		t.Run(invalidUUID, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/clips?submitted_by_user_id="+invalidUUID, http.NoBody)
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.ListClips(c)

			if w.Code != http.StatusBadRequest {
				t.Errorf("for UUID '%s': expected status %d, got %d", invalidUUID, http.StatusBadRequest, w.Code)
			}

			var response StandardResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Errorf("for UUID '%s': response is not valid JSON: %v", invalidUUID, err)
			}

			if response.Error == nil || response.Error.Code != "INVALID_UUID" {
				t.Errorf("for UUID '%s': expected INVALID_UUID error", invalidUUID)
			}
		})
	}
}
