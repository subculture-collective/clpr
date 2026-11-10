package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type fakePackList struct{ asked string }

func (f *fakePackList) ListBroadcasterPacks(_ context.Context, id string, _ int) ([]models.ClipPack, error) {
	f.asked = id
	later := time.Date(2026, 11, 16, 7, 0, 0, 0, time.UTC)
	earlier := time.Date(2026, 11, 9, 7, 0, 0, 0, time.UTC)
	return []models.ClipPack{
		{ID: uuid.New(), WeekStart: "2026-11-02", Status: models.ClipPackReady, ExpiresAt: &later},
		{ID: uuid.New(), WeekStart: "2026-10-26", Status: models.ClipPackExpired, ExpiresAt: &earlier},
		{ID: uuid.New(), WeekStart: "2026-10-19", Status: models.ClipPackFailed},
	}, nil
}

func TestClipPackHandlerListsSignedLinks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	streamer, viewer := uuid.New(), uuid.New()
	twitchID := "1234"
	users := fakeControlsUsers{streamer: {ID: streamer, TwitchID: &twitchID}, viewer: {ID: viewer}}
	store := &fakePackList{}
	links := services.ClipPackLinks{Secret: []byte("s"), BaseURL: "https://packs.example"}
	h := NewClipPackHandler(store, users, links)
	h.now = func() time.Time { return time.Date(2026, 11, 10, 7, 0, 0, 0, time.UTC) }
	router := gin.New()
	router.Use(func(c *gin.Context) {
		if id := c.GetHeader("X-User"); id != "" {
			c.Set("user_id", uuid.MustParse(id))
		}
	})
	router.GET("/me/clip-packs", h.ListMyPacks)
	get := func(user string) (int, []map[string]any) {
		req := httptest.NewRequest(http.MethodGet, "/me/clip-packs", nil)
		if user != "" {
			req.Header.Set("X-User", user)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		var resp struct {
			Data []map[string]any `json:"data"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		return w.Code, resp.Data
	}
	if code, _ := get(""); code != http.StatusUnauthorized {
		t.Fatalf("anonymous = %d", code)
	}
	if code, _ := get(viewer.String()); code != http.StatusConflict {
		t.Fatalf("no Twitch = %d", code)
	}
	code, packs := get(streamer.String())
	if code != http.StatusOK || store.asked != "1234" || len(packs) != 3 {
		t.Fatalf("packs = %d %v (asked %q)", code, packs, store.asked)
	}
	link, _ := packs[0]["download_url"].(string)
	if !strings.HasPrefix(link, "https://packs.example/packs/") || packs[1]["download_url"] != nil || packs[2]["download_url"] != nil {
		t.Fatalf("links = %v", packs)
	}
	if _, leaked := packs[0]["file_name"]; leaked {
		t.Fatal("the file name is in the response")
	}

	// Without a secret, packs are listed without links.
	h.links = services.ClipPackLinks{}
	if _, packs := get(streamer.String()); packs[0]["download_url"] != nil {
		t.Fatalf("unsigned link = %v", packs[0])
	}
}
