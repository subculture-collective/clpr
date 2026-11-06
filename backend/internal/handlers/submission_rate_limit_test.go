package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// The submit page recognizes a 429 by error == "rate_limit_exceeded" and
// numeric limit, window and retry_after (SubmitClipPage.tsx).
func assertRateLimitBody(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Error      string `json:"error"`
		Limit      *int   `json:"limit"`
		Window     *int   `json:"window"`
		RetryAfter *int64 `json:"retry_after"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode 429 body %q: %v", rec.Body.String(), err)
	}
	if body.Error != "rate_limit_exceeded" || body.Limit == nil || *body.Limit != 10 || body.Window == nil || *body.Window != 3600 || body.RetryAfter == nil || *body.RetryAfter != 1_700_003_600 {
		t.Fatalf("429 body = %s, want rate_limit_exceeded with limit, window and retry_after", rec.Body.String())
	}
}

func TestSubmissionRateLimitResponseMatchesFrontendContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rateLimited := &services.RateLimitError{Message: "rate_limit_exceeded", Limit: 10, Window: 3600, RetryAfter: 1_700_003_600}
	service := &fakeSubmissionService{
		submitClipFn: func(context.Context, uuid.UUID, *services.SubmitClipRequest, string, string) (*models.ClipSubmission, error) {
			return nil, rateLimited
		},
		submitUploadFn: func(context.Context, uuid.UUID, *services.SubmitUploadRequest, string, string) (*models.ClipSubmission, error) {
			return nil, rateLimited
		},
	}
	validator := services.NewUploadValidator(config.ClipSourceConfig{
		MaxDurationSeconds:     600,
		MaxUploadBytes:         2048,
		AllowedUploadMimeTypes: []string{"video/mp4"},
	}, func(ctx context.Context, filePath string) (int64, error) {
		return 45, nil
	})
	handler := NewSubmissionHandler(service, validator, &fakeClipStorage{})

	t.Run("link submission", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Set("user_id", uuid.New())
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/submissions", bytes.NewBufferString(`{"clip_url":"https://clips.twitch.tv/SomeClip"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		handler.SubmitClip(c)
		assertRateLimitBody(t, rec)
	})

	t.Run("upload submission", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Set("user_id", uuid.New())
		c.Request = newUploadTestRequest(t, "/api/v1/submissions/upload", map[string]string{}, "file", "clip.mp4", bytes.Repeat([]byte{0}, 1024))
		handler.SubmitUpload(c)
		assertRateLimitBody(t, rec)
	})
}

func TestSubmissionModerationErrorMapsClaimedClipToConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	writeSubmissionModerationError(c, fmt.Errorf("%w: submission x", repository.ErrSubmissionClipClaimed), "Failed to approve submission")
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusConflict)
	}
}
