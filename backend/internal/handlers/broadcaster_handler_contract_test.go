package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type broadcasterRankingRefresherStub struct {
	err      error
	deadline bool
}

func (s *broadcasterRankingRefresherStub) RefreshRankings(ctx context.Context) error {
	_, s.deadline = ctx.Deadline()
	return s.err
}

func TestRefreshBroadcasterRankingsContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{"success", nil, http.StatusOK},
		{"dependency failure", errors.New("database unavailable"), http.StatusInternalServerError},
		{"timeout", context.DeadlineExceeded, http.StatusGatewayTimeout},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &broadcasterRankingRefresherStub{err: tt.err}
			handler := &BroadcasterHandler{rankingRefresher: service}
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/broadcasters/refresh-rankings", http.NoBody)
			handler.RefreshBroadcasterRankings(ctx)
			if recorder.Code != tt.status {
				t.Fatalf("expected %d, got %d: %s", tt.status, recorder.Code, recorder.Body.String())
			}
			if !service.deadline {
				t.Fatal("refresh was called without a deadline")
			}
		})
	}
}
