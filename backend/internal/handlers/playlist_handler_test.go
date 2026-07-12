package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func playlistTestContext(method, target string, body []byte, identity interface{}) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, target, bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	if identity != nil {
		c.Set("user_id", identity)
	}
	return c, w
}

func TestPlaylistCoreRoutesRejectMalformedIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name   string
		invoke func(*PlaylistHandler, *gin.Context)
	}{
		{"create", func(h *PlaylistHandler, c *gin.Context) { h.CreatePlaylist(c) }},
		{"list", func(h *PlaylistHandler, c *gin.Context) { h.ListUserPlaylists(c) }},
		{"bookmarks", func(h *PlaylistHandler, c *gin.Context) { h.ListBookmarkedPlaylists(c) }},
		{"optional public", func(h *PlaylistHandler, c *gin.Context) { h.ListPublicPlaylists(c) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, w := playlistTestContext(http.MethodGet, "/api/v1/playlists", nil, "bad")
			tc.invoke(NewPlaylistHandler(nil), c)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestPlaylistPaginationFailsClosed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	id := uuid.New()
	for _, target := range []string{
		"/api/v1/playlists?page=0", "/api/v1/playlists?limit=101",
		"/api/v1/playlists/" + id.String() + "?page=bad",
	} {
		c, w := playlistTestContext(http.MethodGet, target, nil, uuid.New())
		c.Params = gin.Params{{Key: "id", Value: id.String()}}
		if strings.Contains(target, id.String()) {
			NewPlaylistHandler(nil).GetPlaylist(c)
		} else {
			NewPlaylistHandler(nil).ListUserPlaylists(c)
		}
		if w.Code != http.StatusBadRequest {
			t.Fatalf("%s: expected 400, got %d", target, w.Code)
		}
	}
}

func TestPlaylistCreateRedactsBinderDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, w := playlistTestContext(http.MethodPost, "/api/v1/playlists", []byte(`{"title":`), uuid.New())
	NewPlaylistHandler(nil).CreatePlaylist(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	if strings.Contains(w.Body.String(), "unexpected EOF") {
		t.Fatalf("leaked binder error: %s", w.Body.String())
	}
}

func TestPlaylistUpdateRejectsEmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	id := uuid.New()
	c, w := playlistTestContext(http.MethodPatch, "/api/v1/playlists/"+id.String(), []byte(`{}`), uuid.New())
	c.Params = gin.Params{{Key: "id", Value: id.String()}}
	h := NewPlaylistHandler(services.NewPlaylistService(nil, nil, ""))
	h.UpdatePlaylist(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestPlaylistMembershipInputsFailClosed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	playlistID, clipID := uuid.New(), uuid.New()
	for _, tc := range []struct {
		name, method, body string
		identity           interface{}
		invoke             func(*PlaylistHandler, *gin.Context)
		status             int
	}{
		{"identity", http.MethodPost, `{"clip_ids":["` + clipID.String() + `"]}`, "bad", func(h *PlaylistHandler, c *gin.Context) { h.AddClipsToPlaylist(c) }, http.StatusUnauthorized},
		{"duplicate add", http.MethodPost, `{"clip_ids":["` + clipID.String() + `","` + clipID.String() + `"]}`, uuid.New(), func(h *PlaylistHandler, c *gin.Context) { h.AddClipsToPlaylist(c) }, http.StatusBadRequest},
		{"duplicate reorder", http.MethodPut, `{"clip_ids":["` + clipID.String() + `","` + clipID.String() + `"]}`, uuid.New(), func(h *PlaylistHandler, c *gin.Context) { h.ReorderPlaylistClips(c) }, http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, w := playlistTestContext(tc.method, "/api/v1/playlists/"+playlistID.String()+"/clips", []byte(tc.body), tc.identity)
			c.Params = gin.Params{{Key: "id", Value: playlistID.String()}}
			tc.invoke(NewPlaylistHandler(nil), c)
			if w.Code != tc.status {
				t.Fatalf("expected %d, got %d: %s", tc.status, w.Code, w.Body.String())
			}
		})
	}
}
