package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func serveFilteredClips(t *testing.T, handler *FeedHandler, cursor string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/feeds/clips", handler.GetFilteredClips)
	query := url.Values{"sort": {"new"}, "limit": {"20"}, "cursor": {cursor}}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/feeds/clips?"+query.Encode(), nil))
	return rec
}

// A database failure on a cursor page is a server error, not "invalid cursor"
// with the driver's error text.
func TestFilteredClipsCursorPageReportsQueryFailureAsServerError(t *testing.T) {
	pool := unreachablePool(t)
	clipRepo := repository.NewClipRepository(pool)
	handler := NewFeedHandler(services.NewFeedService(nil, clipRepo, nil, nil, nil, nil), nil, nil, nil, nil)

	cursor := utils.EncodeCursor("new", 0, uuid.New(), 1_700_000_000)
	rec := serveFilteredClips(t, handler, cursor)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500: %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "invalid cursor") || strings.Contains(rec.Body.String(), "127.0.0.1") {
		t.Fatalf("response blames the cursor or leaks the error: %s", rec.Body.String())
	}
}

func TestFilteredClipsRejectsUndecodableCursor(t *testing.T) {
	rec := serveFilteredClips(t, &FeedHandler{}, "not-a-cursor")
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "invalid cursor") {
		t.Fatalf("got %d %s, want 400 invalid cursor", rec.Code, rec.Body.String())
	}
}
