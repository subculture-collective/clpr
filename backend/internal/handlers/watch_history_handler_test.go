package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type watchHistoryRepositoryStub struct {
	enabled bool
	history []models.WatchHistoryEntry
	err     error
}

func (s *watchHistoryRepositoryStub) IsWatchHistoryEnabled(context.Context, uuid.UUID) (bool, error) {
	return s.enabled, s.err
}
func (s *watchHistoryRepositoryStub) RecordWatchProgress(context.Context, uuid.UUID, uuid.UUID, int, int, string) error {
	return s.err
}
func (s *watchHistoryRepositoryStub) GetWatchHistory(context.Context, uuid.UUID, string, int) ([]models.WatchHistoryEntry, error) {
	return s.history, s.err
}
func (s *watchHistoryRepositoryStub) GetResumePosition(context.Context, uuid.UUID, uuid.UUID) (int, bool, error) {
	return 0, false, s.err
}
func (s *watchHistoryRepositoryStub) ClearWatchHistory(context.Context, uuid.UUID) error {
	return s.err
}

func TestRecordWatchProgress_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &WatchHistoryHandler{
		repo: nil,
	}

	clipID := uuid.New().String()
	reqBody := map[string]interface{}{
		"clip_id":          clipID,
		"progress_seconds": 30,
		"duration_seconds": 120,
		"session_id":       "test_session_123",
	}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/watch-history", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.RecordWatchProgress(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}

	if errMsg, exists := response["error"]; !exists || errMsg != "User not authenticated" {
		t.Errorf("expected error message 'User not authenticated', got %v", errMsg)
	}
}

func TestGetWatchHistory_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &WatchHistoryHandler{
		repo: nil,
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/watch-history", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handler.GetWatchHistory(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestGetResumePosition_Unauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &WatchHistoryHandler{
		repo: nil,
	}

	clipID := uuid.New().String()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/clips/"+clipID+"/progress", http.NoBody)
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Params = gin.Params{
		{Key: "id", Value: clipID},
	}

	handler.GetResumePosition(c)

	// For unauthenticated users, should return no progress
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("response is not valid JSON: %v", err)
	}

	if hasProgress, exists := response["has_progress"]; !exists || hasProgress != false {
		t.Errorf("expected has_progress to be false, got %v", hasProgress)
	}
}

func TestProgressPercentCalculation(t *testing.T) {
	tests := []struct {
		name             string
		progressSeconds  int
		durationSeconds  int
		expectedPercent  float64
		expectedComplete bool
	}{
		{
			name:             "0% progress",
			progressSeconds:  0,
			durationSeconds:  100,
			expectedPercent:  0.0,
			expectedComplete: false,
		},
		{
			name:             "50% progress",
			progressSeconds:  50,
			durationSeconds:  100,
			expectedPercent:  50.0,
			expectedComplete: false,
		},
		{
			name:             "90% progress - completed",
			progressSeconds:  90,
			durationSeconds:  100,
			expectedPercent:  90.0,
			expectedComplete: true,
		},
		{
			name:             "95% progress - completed",
			progressSeconds:  95,
			durationSeconds:  100,
			expectedPercent:  95.0,
			expectedComplete: true,
		},
		{
			name:             "100% progress - completed",
			progressSeconds:  100,
			durationSeconds:  100,
			expectedPercent:  100.0,
			expectedComplete: true,
		},
		{
			name:             "89% progress - not completed",
			progressSeconds:  89,
			durationSeconds:  100,
			expectedPercent:  89.0,
			expectedComplete: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			progressPercent := float64(tt.progressSeconds) / float64(tt.durationSeconds)
			completed := progressPercent >= 0.9

			actualPercent := progressPercent * 100
			if actualPercent != tt.expectedPercent {
				t.Errorf("expected progress percent %.1f, got %.1f", tt.expectedPercent, actualPercent)
			}

			if completed != tt.expectedComplete {
				t.Errorf("expected completed %v, got %v", tt.expectedComplete, completed)
			}
		})
	}
}

func TestWatchHistoryLiveResponseContracts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	clipID := uuid.New()
	repo := &watchHistoryRepositoryStub{
		enabled: true,
		history: []models.WatchHistoryEntry{{
			ID: uuid.New(), UserID: userID, ClipID: clipID, ProgressSeconds: 30,
			DurationSeconds: 120, ProgressPercent: 25, SessionID: "session-1",
			WatchedAt: time.Now(), CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}},
	}
	handler := NewWatchHistoryHandler(repo)

	t.Run("records zero-second progress", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/watch-history", bytes.NewBufferString(`{"clip_id":"`+clipID.String()+`","progress_seconds":0,"duration_seconds":120,"session_id":"session-1"}`))
		request.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = request
		ctx.Set("user_id", userID)
		handler.RecordWatchProgress(ctx)
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
		}
		var response struct {
			Status          string  `json:"status"`
			Completed       bool    `json:"completed"`
			ProgressPercent float64 `json:"progress_percent"`
		}
		if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		if response.Status != "recorded" || response.Completed || response.ProgressPercent != 0 {
			t.Fatalf("unexpected response: %+v", response)
		}
	})

	t.Run("lists history", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/v1/watch-history?filter=in-progress&limit=25", nil)
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = request
		ctx.Set("user_id", userID)
		handler.GetWatchHistory(ctx)
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", recorder.Code)
		}
		var response models.WatchHistoryResponse
		if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		if response.Total != 1 || len(response.History) != 1 || response.History[0].ClipID != clipID {
			t.Fatalf("unexpected history response: %+v", response)
		}
	})

	t.Run("clears history", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodDelete, "/api/v1/watch-history", nil)
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = request
		ctx.Set("user_id", userID)
		handler.ClearWatchHistory(ctx)
		if recorder.Code != http.StatusOK || recorder.Body.String() != `{"status":"cleared"}` {
			t.Fatalf("unexpected clear response: %d %s", recorder.Code, recorder.Body.String())
		}
	})
}

func TestWatchHistoryRejectsInvalidBounds(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	clipID := uuid.New()
	handler := NewWatchHistoryHandler(&watchHistoryRepositoryStub{enabled: true})

	tests := []struct {
		name, method, target, body string
		handle                     func(*gin.Context)
	}{
		{"unknown filter", http.MethodGet, "/api/v1/watch-history?filter=unknown", "", handler.GetWatchHistory},
		{"excessive limit", http.MethodGet, "/api/v1/watch-history?limit=1000", "", handler.GetWatchHistory},
		{"progress beyond duration", http.MethodPost, "/api/v1/watch-history", `{"clip_id":"` + clipID.String() + `","progress_seconds":121,"duration_seconds":120,"session_id":"session-1"}`, handler.RecordWatchProgress},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.target, bytes.NewBufferString(test.body))
			request.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = request
			ctx.Set("user_id", userID)
			test.handle(ctx)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("expected status 400, got %d: %s", recorder.Code, recorder.Body.String())
			}
		})
	}
}
