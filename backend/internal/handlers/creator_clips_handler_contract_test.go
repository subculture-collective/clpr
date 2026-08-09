package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type creatorClipServiceStub struct {
	creatorID string
	userID    *uuid.UUID
	page      int
	limit     int
	clips     []services.ClipWithUserData
	total     int
	err       error
}

func (s *creatorClipServiceStub) ListCreatorClips(_ context.Context, creatorID string, userID *uuid.UUID, page, limit int) ([]services.ClipWithUserData, int, error) {
	s.creatorID, s.userID, s.page, s.limit = creatorID, userID, page, limit
	return s.clips, s.total, s.err
}

func creatorClipsContext(target, creatorID string) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Params = gin.Params{{Key: "creator", Value: creatorID}}
	ctx.Request = httptest.NewRequest(http.MethodGet, target, nil)
	return ctx, recorder
}

func TestListCreatorClipsRejectsInvalidRouteAndPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name      string
		creatorID string
		target    string
	}{
		{"empty creator", "", "/api/v1/creators//clips"},
		{"invalid page", "1234", "/api/v1/creators/1234/clips?page=nope"},
		{"zero page", "1234", "/api/v1/creators/1234/clips?page=0"},
		{"oversized limit", "1234", "/api/v1/creators/1234/clips?limit=101"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &creatorClipServiceStub{}
			handler := &ClipHandler{creatorClipService: service}
			ctx, recorder := creatorClipsContext(tt.target, tt.creatorID)
			handler.ListCreatorClips(ctx)
			if recorder.Code != http.StatusBadRequest || service.creatorID != "" {
				t.Fatalf("expected rejected request, got %d %s service creator=%q", recorder.Code, recorder.Body.String(), service.creatorID)
			}
		})
	}
}

func TestListCreatorClipsPassesTwitchIDAndSafeOptionalIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &creatorClipServiceStub{clips: []services.ClipWithUserData{}, total: 26}
	handler := &ClipHandler{creatorClipService: service}
	ctx, recorder := creatorClipsContext("/api/v1/creators/1234/clips?page=2&limit=25", " 1234 ")
	ctx.Set("user_id", "malformed")
	handler.ListCreatorClips(ctx)

	if recorder.Code != http.StatusOK || service.creatorID != "1234" || service.userID != nil || service.page != 2 || service.limit != 25 {
		t.Fatalf("unexpected request mapping: code=%d creator=%q user=%v page=%d limit=%d body=%s", recorder.Code, service.creatorID, service.userID, service.page, service.limit, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"total":26`) || !strings.Contains(recorder.Body.String(), `"total_pages":2`) {
		t.Fatalf("pagination metadata missing: %s", recorder.Body.String())
	}
}

func TestListCreatorClipsPassesAuthenticatedIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := &creatorClipServiceStub{}
	handler := &ClipHandler{creatorClipService: service}
	ctx, recorder := creatorClipsContext("/api/v1/creators/1234/clips", "1234")
	userID := uuid.New()
	ctx.Set("user_id", userID)
	handler.ListCreatorClips(ctx)

	if recorder.Code != http.StatusOK || service.userID == nil || *service.userID != userID {
		t.Fatalf("authenticated identity not forwarded safely: %d %v", recorder.Code, service.userID)
	}
}
