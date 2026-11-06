package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// {"vote":0} removes a vote and must reach the service; only a missing or
// out-of-range vote is a bad request.
func TestVoteOnClipAcceptsZeroVote(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pool := unreachablePool(t)
	clipService := services.NewClipService(repository.NewClipRepository(pool), nil, nil, nil, nil, nil, nil, nil, nil)
	handler := NewClipHandler(clipService, nil)

	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set("user_id", uuid.New()) })
	router.POST("/clips/:id/vote", handler.VoteOnClip)

	for _, tt := range []struct {
		body     string
		wantCode int
		wantText string
	}{
		// The service is reached and fails on the unreachable database.
		{`{"vote":0}`, http.StatusInternalServerError, "VOTE_FAILED"},
		{`{"vote":1}`, http.StatusInternalServerError, "VOTE_FAILED"},
		{`{}`, http.StatusBadRequest, "INVALID_REQUEST"},
		{`{"vote":2}`, http.StatusBadRequest, "INVALID_VOTE"},
	} {
		t.Run(tt.body, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/clips/"+uuid.NewString()+"/vote", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(rec, req)
			if rec.Code != tt.wantCode || !strings.Contains(rec.Body.String(), tt.wantText) {
				t.Fatalf("got %d %s, want %d %s", rec.Code, rec.Body.String(), tt.wantCode, tt.wantText)
			}
		})
	}
}
