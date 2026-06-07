package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type fakeSubmissionService struct {
	submitClipFn         func(context.Context, uuid.UUID, *services.SubmitClipRequest, string, string) (*models.ClipSubmission, error)
	submitUploadFn       func(context.Context, uuid.UUID, *services.SubmitUploadRequest, string, string) (*models.ClipSubmission, error)
	getUserSubmissionsFn func(context.Context, uuid.UUID, int, int) ([]*models.ClipSubmission, int, error)
}

func (f *fakeSubmissionService) SubmitClip(ctx context.Context, userID uuid.UUID, req *services.SubmitClipRequest, ip string, deviceFingerprint string) (*models.ClipSubmission, error) {
	if f.submitClipFn != nil {
		return f.submitClipFn(ctx, userID, req, ip, deviceFingerprint)
	}
	return nil, nil
}

func (f *fakeSubmissionService) SubmitUpload(ctx context.Context, userID uuid.UUID, req *services.SubmitUploadRequest, ip string, deviceFingerprint string) (*models.ClipSubmission, error) {
	if f.submitUploadFn != nil {
		return f.submitUploadFn(ctx, userID, req, ip, deviceFingerprint)
	}
	return nil, nil
}

func (f *fakeSubmissionService) GetUserSubmissions(ctx context.Context, userID uuid.UUID, page, limit int) ([]*models.ClipSubmission, int, error) {
	if f.getUserSubmissionsFn != nil {
		return f.getUserSubmissionsFn(ctx, userID, page, limit)
	}
	return nil, 0, nil
}

func (f *fakeSubmissionService) GetSubmissionStats(ctx context.Context, userID uuid.UUID) (*models.SubmissionStats, error) {
	return nil, nil
}

func (f *fakeSubmissionService) GetClipMetadata(ctx context.Context, clipURLOrID string) (*services.ClipMetadata, error) {
	return nil, nil
}

func (f *fakeSubmissionService) CheckClipExistence(ctx context.Context, twitchClipID string) (*services.ClipExistenceResult, error) {
	return nil, nil
}

func (f *fakeSubmissionService) GetPendingSubmissionsWithFilters(ctx context.Context, filters repository.SubmissionFilters, page, limit int) ([]*models.ClipSubmissionWithUser, int, error) {
	return nil, 0, nil
}

func (f *fakeSubmissionService) ApproveSubmission(ctx context.Context, submissionID, reviewerID uuid.UUID) error {
	return nil
}
func (f *fakeSubmissionService) RejectSubmission(ctx context.Context, submissionID, reviewerID uuid.UUID, reason string) error {
	return nil
}
func (f *fakeSubmissionService) BulkApproveSubmissions(ctx context.Context, submissionIDs []uuid.UUID, reviewerID uuid.UUID) error {
	return nil
}
func (f *fakeSubmissionService) BulkRejectSubmissions(ctx context.Context, submissionIDs []uuid.UUID, reviewerID uuid.UUID, reason string) error {
	return nil
}

type fakeClipStorage struct {
	putCalls int
	putInput storage.PutObjectInput
	putErr   error
	result   storage.PutObjectResult
}

func (f *fakeClipStorage) PutObject(ctx context.Context, input storage.PutObjectInput) (storage.PutObjectResult, error) {
	f.putCalls++
	f.putInput = input
	if f.putErr != nil {
		return storage.PutObjectResult{}, f.putErr
	}
	if f.result.Provider == "" {
		f.result = storage.PutObjectResult{Provider: "s3", Bucket: "clips", Key: input.Key}
	}
	return f.result, nil
}

func (f *fakeClipStorage) CopyObject(ctx context.Context, srcKey, dstKey, contentType string) (storage.PutObjectResult, error) {
	return storage.PutObjectResult{Provider: "s3", Bucket: "clips", Key: dstKey}, nil
}

func (f *fakeClipStorage) DeleteObject(ctx context.Context, key string) error { return nil }
func (f *fakeClipStorage) PublicURL(key string) string                        { return "" }

func TestSubmissionHandlerSubmitUploadRejectsOversizeBeforeStorage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	probeCalled := false
	validator := services.NewUploadValidator(config.ClipSourceConfig{
		MaxDurationSeconds:     600,
		MaxUploadBytes:         1024,
		AllowedUploadMimeTypes: []string{"video/mp4"},
	}, func(ctx context.Context, filePath string) (int64, error) {
		probeCalled = true
		return 45, nil
	})

	storageBackend := &fakeClipStorage{}
	handler := NewSubmissionHandler(&services.SubmissionService{}, validator, storageBackend)

	userID := uuid.New()
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Set("user_id", userID)
	req := newUploadTestRequest(t, "/api/v1/submissions/upload", map[string]string{}, "file", "clip.mp4", bytes.Repeat([]byte("a"), 1025))
	ctx.Request = req

	handler.SubmitUpload(ctx)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusRequestEntityTooLarge)
	}
	if storageBackend.putCalls != 0 {
		t.Fatalf("storage PutObject called %d times, want 0", storageBackend.putCalls)
	}
	if probeCalled {
		t.Fatal("duration probe called for oversize upload")
	}
	assertUploadErrorContains(t, rec.Body.String(), "uploaded file exceeds maximum allowed size")
}

func TestSubmissionHandlerSubmitUploadMapsStorageFailureToUploadError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	probeCalled := false
	validator := services.NewUploadValidator(config.ClipSourceConfig{
		MaxDurationSeconds:     600,
		MaxUploadBytes:         2048,
		AllowedUploadMimeTypes: []string{"video/mp4"},
	}, func(ctx context.Context, filePath string) (int64, error) {
		probeCalled = true
		return 45, nil
	})

	storageBackend := &fakeClipStorage{putErr: errors.New("s3 unavailable")}
	handler := NewSubmissionHandler(&services.SubmissionService{}, validator, storageBackend)

	userID := uuid.New()
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Set("user_id", userID)
	req := newUploadTestRequest(t, "/api/v1/submissions/upload", map[string]string{
		"custom_title":      "  Great play  ",
		"submission_reason": "  because  ",
		"is_nsfw":           "false",
	}, "file", "clip.mp4", bytes.Repeat([]byte{0}, 1024))
	ctx.Request = req

	handler.SubmitUpload(ctx)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	if storageBackend.putCalls != 1 {
		t.Fatalf("storage PutObject called %d times, want 1", storageBackend.putCalls)
	}
	if !probeCalled {
		t.Fatal("duration probe not called for valid upload")
	}
	assertUploadErrorContains(t, rec.Body.String(), "failed to store uploaded clip")
}

func TestSubmissionHandlerGetUserSubmissionsOmitsStorageInternals(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &fakeSubmissionService{
		getUserSubmissionsFn: func(ctx context.Context, userID uuid.UUID, page, limit int) ([]*models.ClipSubmission, int, error) {
			clipID := uuid.New()
			return []*models.ClipSubmission{{
				ID:               uuid.New(),
				UserID:           userID,
				ClipID:           &clipID,
				TwitchClipID:     "upload:123",
				TwitchClipURL:    "https://clips.twitch.tv/upload:123",
				Title:            ptrString("Title"),
				CustomTitle:      ptrString("Custom title"),
				IsNSFW:           true,
				Status:           "pending",
				CreatedAt:        time.Now().UTC(),
				UpdatedAt:        time.Now().UTC(),
				SourceType:       "upload",
				SourcePlatform:   "upload",
				SourceURL:        ptrString("https://example.com/source"),
				SourceID:         ptrString("source-123"),
				SourceMetadata:   json.RawMessage(`{"storage_key":"uploads/pending/u/s/original.mp4","storage_bucket":"clips","storage_provider":"s3"}`),
				StorageProvider:  ptrString("s3"),
				StorageBucket:    ptrString("clips"),
				StorageKey:       ptrString("uploads/pending/u/s/original.mp4"),
				OriginalFilename: ptrString("clip.mp4"),
				DurationSeconds:  ptrInt(45),
				DurationVerified: true,
				CreatorName:      ptrString("Creator"),
				BroadcasterName:  ptrString("Broadcaster"),
				GameName:         ptrString("Game"),
				ThumbnailURL:     ptrString("https://example.com/thumb.jpg"),
				Duration:         ptrFloat64(45.5),
				ViewCount:        123,
			}}, 1, nil
		},
	}
	handler := NewSubmissionHandler(service, &services.UploadValidator{}, &fakeClipStorage{})

	userID := uuid.New()
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Set("user_id", userID)
	req, err := http.NewRequest(http.MethodGet, "/api/v1/submissions?page=1&limit=20", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	ctx.Request = req

	handler.GetUserSubmissions(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	for _, forbidden := range []string{"storage_key", "storage_bucket", "storage_provider", "source_metadata"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("response JSON leaked %q: %s", forbidden, body)
		}
	}
	for _, want := range []string{`"source_type":"upload"`, `"source_platform":"upload"`, `"twitch_clip_id":"upload:123"`, `"status":"pending"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("response JSON missing %q: %s", want, body)
		}
	}
}

func TestSubmissionHandlerSubmitUploadOmitsStorageInternals(t *testing.T) {
	gin.SetMode(gin.TestMode)

	reviewedBy := uuid.New()
	clipID := uuid.New()
	userID := uuid.New()
	service := &fakeSubmissionService{
		submitUploadFn: func(ctx context.Context, gotUserID uuid.UUID, req *services.SubmitUploadRequest, ip string, deviceFingerprint string) (*models.ClipSubmission, error) {
			if gotUserID != userID {
				t.Fatalf("userID = %s, want %s", gotUserID, userID)
			}
			return &models.ClipSubmission{
				ID:               uuid.New(),
				UserID:           gotUserID,
				ClipID:           &clipID,
				TwitchClipID:     "upload:123",
				TwitchClipURL:    "",
				Title:            ptrString("Title"),
				CustomTitle:      ptrString("Custom title"),
				IsNSFW:           true,
				Status:           "pending",
				ReviewedBy:       &reviewedBy,
				CreatedAt:        time.Now().UTC(),
				UpdatedAt:        time.Now().UTC(),
				SourceType:       "upload",
				SourcePlatform:   "upload",
				SourceURL:        ptrString("https://example.com/source"),
				SourceID:         ptrString("source-123"),
				SourceMetadata:   json.RawMessage(`{"storage_key":"uploads/pending/u/s/original.mp4","storage_bucket":"clips","storage_provider":"s3"}`),
				DurationSeconds:  ptrInt(45),
				DurationVerified: true,
				CreatorName:      ptrString("Creator"),
				BroadcasterName:  ptrString("Broadcaster"),
				GameName:         ptrString("Game"),
				ThumbnailURL:     ptrString("https://example.com/thumb.jpg"),
				Duration:         ptrFloat64(45.5),
				ViewCount:        123,
				StorageProvider:  ptrString("s3"),
				StorageBucket:    ptrString("clips"),
				StorageKey:       ptrString("uploads/pending/u/s/original.mp4"),
				OriginalFilename: ptrString("clip.mp4"),
			}, nil
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

	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Set("user_id", userID)
	req := newUploadTestRequest(t, "/api/v1/submissions/upload", map[string]string{
		"custom_title":      " Great title ",
		"submission_reason": " because ",
		"is_nsfw":           "false",
	}, "file", "clip.mp4", bytes.Repeat([]byte{0}, 1024))
	ctx.Request = req

	handler.SubmitUpload(ctx)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	body := rec.Body.String()
	for _, forbidden := range []string{"storage_key", "storage_bucket", "storage_provider", "source_metadata"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("response JSON leaked %q: %s", forbidden, body)
		}
	}
	for _, want := range []string{`"title":"Title"`, `"status":"pending"`, `"source_type":"upload"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("response JSON missing %q: %s", want, body)
		}
	}
}

func newUploadTestRequest(t *testing.T, path string, fields map[string]string, fileField, filename string, fileContent []byte) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("WriteField(%s) error = %v", key, err)
		}
	}
	part, err := writer.CreateFormFile(fileField, filename)
	if err != nil {
		t.Fatalf("CreateFormFile() error = %v", err)
	}
	if _, err := part.Write(fileContent); err != nil {
		t.Fatalf("writing file content failed: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close() error = %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, path, &body)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func assertUploadErrorContains(t *testing.T, rawBody, want string) {
	t.Helper()

	var payload map[string]any
	if err := json.Unmarshal([]byte(rawBody), &payload); err != nil {
		t.Fatalf("response body is not JSON: %v (%s)", err, rawBody)
	}
	msg, _ := payload["error"].(string)
	if !strings.Contains(strings.ToLower(msg), strings.ToLower(want)) {
		t.Fatalf("error = %q, want contains %q", msg, want)
	}
}

func ptrString(s string) *string { return &s }

func ptrInt(v int) *int { return &v }

func ptrFloat64(v float64) *float64 { return &v }
