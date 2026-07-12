package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/gin-gonic/gin"
)

func productRouteContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, target, strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c, w
}

func TestProductListRoutesRejectMalformedPaginationAndFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name string
		path string
		call func(*gin.Context)
	}{
		{"popular broadcasters", "/?limit=zero", (&BroadcasterHandler{}).ListPopularBroadcasters},
		{"broadcaster rankings", "/?offset=-1", (&BroadcasterHandler{}).GetBroadcasterRankings},
		{"category games", "/?page=zero", (&CategoryHandler{}).ListCategoryGames},
		{"category clips", "/?sort=unknown", (&CategoryHandler{}).ListCategoryClips},
		{"game clips", "/?timeframe=decade", (&GameHandler{}).ListGameClips},
		{"trending games", "/?limit=101", (&GameHandler{}).GetTrendingGames},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c, w := productRouteContext(http.MethodGet, test.path, "")
			c.Params = gin.Params{{Key: "slug", Value: "games"}, {Key: "gameId", Value: "1234"}}
			test.call(c)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected %d, got %d", http.StatusBadRequest, w.Code)
			}
		})
	}
}

func TestAdTrackingDoesNotRequireDuplicatedBodyID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := productRouteContext(http.MethodPost, "/ads/track/id", `{"viewability_time_ms":1000,"is_viewable":true}`)
	var request models.AdTrackingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		t.Fatalf("tracking body should not duplicate the path impression ID: %v", err)
	}
}

func TestClipStatusRoutesBoundIdentifiers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &ClipHandler{}
	c, w := productRouteContext(http.MethodGet, "/clips/id/processing-status", "")
	c.Params = gin.Params{{Key: "id", Value: strings.Repeat("x", 129)}}
	h.GetClipProcessingStatus(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, w.Code)
	}
}
