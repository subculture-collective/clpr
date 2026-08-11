package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/gin-gonic/gin"
)

type creatorDiscoveryListerStub struct {
	limit int
	rails *models.CreatorDiscoveryRails
	err   error
}

func (s *creatorDiscoveryListerStub) ListCreatorDiscovery(
	_ context.Context,
	limit int,
) (*models.CreatorDiscoveryRails, error) {
	s.limit = limit
	return s.rails, s.err
}

func TestListCreatorDiscoveryContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	stub := &creatorDiscoveryListerStub{rails: &models.CreatorDiscoveryRails{
		Trending: []models.CreatorDiscoveryProfile{{BroadcasterID: "creator-1"}},
		Rising:   []models.CreatorDiscoveryProfile{},
		New:      []models.CreatorDiscoveryProfile{},
	}}
	handler := &BroadcasterHandler{creatorDiscoveryRepo: stub}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/broadcasters/discover?limit=8", http.NoBody)

	handler.ListCreatorDiscovery(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	if stub.limit != 8 {
		t.Fatalf("repository limit = %d, want 8", stub.limit)
	}
	var response struct {
		Success bool                         `json:"success"`
		Data    models.CreatorDiscoveryRails `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !response.Success || len(response.Data.Trending) != 1 {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestListCreatorDiscoveryRejectsInvalidLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &BroadcasterHandler{}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/broadcasters/discover?limit=25", http.NoBody)

	handler.ListCreatorDiscovery(ctx)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", recorder.Code)
	}
}
