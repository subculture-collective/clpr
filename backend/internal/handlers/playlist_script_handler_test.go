package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestPlaylistScriptMalformedIdentityReturnsUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/playlist-scripts", nil)
	c.Set("user_id", "not-a-uuid")

	NewPlaylistScriptHandler(nil).ListMyScripts(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreatePlaylistScriptRejectsInvalidBodyWithoutLeakingBinderDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/playlist-scripts", bytes.NewBufferString(`{"name":`))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", uuid.New())

	NewPlaylistScriptHandler(nil).CreateMyScript(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if strings.Contains(w.Body.String(), "unexpected EOF") {
		t.Fatalf("response leaked binder details: %s", w.Body.String())
	}
}

func TestCreatePlaylistScriptBoundsListFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gameIDs := `"` + strings.Repeat("x", 10) + `"`
	body := `{"name":"bounded","sort":"hot","clip_limit":10,"game_ids":[` + strings.Repeat(gameIDs+",", 50) + gameIDs + `]}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/playlist-scripts", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", uuid.New())

	NewPlaylistScriptHandler(nil).CreateMyScript(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAdminPlaylistScriptListRejectsMalformedLimitBeforeServiceWork(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, target := range []string{"/api/v1/admin/playlist-scripts?limit=invalid", "/api/v1/admin/playlist-scripts?limit=501"} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, target, nil)

		NewPlaylistScriptHandler(nil).ListScripts(c)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for %s, got %d", target, w.Code)
		}
	}
}
