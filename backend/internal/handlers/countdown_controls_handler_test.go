package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type fakeControlsStore struct {
	excluded map[string]bool
	featured bool
	resolved *models.CountdownRemovalInput
}

func (f *fakeControlsStore) BroadcasterCountdownState(_ context.Context, id string) (*models.BroadcasterCountdownState, error) {
	return &models.BroadcasterCountdownState{BroadcasterID: id, Excluded: f.excluded[id], Featured: f.featured,
		ConsentTextVersion: models.CountdownConsentTextVersion}, nil
}
func (f *fakeControlsStore) SetSelfExclusion(_ context.Context, id string, _ uuid.UUID, excluded bool) error {
	f.excluded[id] = excluded
	return nil
}
func (f *fakeControlsStore) DeleteFeatureOptIns(context.Context, uuid.UUID) error {
	f.featured = false
	return nil
}
func (f *fakeControlsStore) ListRemovalTasks(context.Context) ([]models.CountdownRemovalTask, error) {
	return []models.CountdownRemovalTask{}, nil
}
func (f *fakeControlsStore) ResolveRemoval(_ context.Context, in *models.CountdownRemovalInput, _ uuid.UUID) error {
	if in.ClipID == uuid.Nil || in.RenderID.String() == "00000000-0000-0000-0000-000000000009" {
		return repository.ErrCountdownRemovalNotFound
	}
	f.resolved = in
	return nil
}

type fakeControlsUsers map[uuid.UUID]*models.User

func (f fakeControlsUsers) GetByID(_ context.Context, id uuid.UUID) (*models.User, error) {
	return f[id], nil
}

type fakeControlsTwitch struct{ auth *models.TwitchAuth }

func (f fakeControlsTwitch) GetTwitchAuth(context.Context, uuid.UUID) (*models.TwitchAuth, error) {
	return f.auth, nil
}

func TestCountdownControlsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	streamer, viewer := uuid.New(), uuid.New()
	twitchID := "1234"
	store := &fakeControlsStore{excluded: map[string]bool{}, featured: true}
	h := NewCountdownControlsHandler(store,
		fakeControlsUsers{streamer: {ID: streamer, TwitchID: &twitchID}, viewer: {ID: viewer}},
		fakeControlsTwitch{&models.TwitchAuth{TwitchUserID: "1234", Scopes: "channel:bot channel:manage:clips"}})
	router := gin.New()
	router.Use(func(c *gin.Context) {
		if id := c.GetHeader("X-User"); id != "" {
			c.Set("user_id", uuid.MustParse(id))
		}
	})
	router.GET("/me/countdowns", h.GetMyCountdownControls)
	router.PUT("/me/countdowns/exclusion", h.SetMyCountdownExclusion)
	router.DELETE("/me/countdowns/feature", h.DeleteMyFeatureOptIn)
	router.POST("/admin/removals", h.ResolveRemoval)

	do := func(method, path, user, body string) (int, models.BroadcasterCountdownState) {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if user != "" {
			req.Header.Set("X-User", user)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		var resp struct {
			Data models.BroadcasterCountdownState `json:"data"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		return w.Code, resp.Data
	}

	if code, _ := do("GET", "/me/countdowns", "", ""); code != http.StatusUnauthorized {
		t.Fatalf("anonymous = %d", code)
	}
	if code, _ := do("GET", "/me/countdowns", viewer.String(), ""); code != http.StatusConflict {
		t.Fatalf("user without Twitch = %d", code)
	}
	code, state := do("GET", "/me/countdowns", streamer.String(), "")
	if code != 200 || state.BroadcasterID != "1234" || state.Excluded || !state.ClipDownloadAuthorized || !state.Featured {
		t.Fatalf("state = %d %+v", code, state)
	}
	if code, _ := do("PUT", "/me/countdowns/exclusion", streamer.String(), `{}`); code != http.StatusBadRequest {
		t.Fatalf("missing excluded = %d", code)
	}
	if code, state := do("PUT", "/me/countdowns/exclusion", streamer.String(), `{"excluded":true}`); code != 200 || !state.Excluded {
		t.Fatalf("exclude = %d %+v", code, state)
	}
	if code, state := do("PUT", "/me/countdowns/exclusion", streamer.String(), `{"excluded":false}`); code != 200 || state.Excluded {
		t.Fatalf("include = %d %+v", code, state)
	}
	if code, state := do("DELETE", "/me/countdowns/feature", streamer.String(), ""); code != 200 || state.Featured {
		t.Fatalf("stop featuring = %d %+v", code, state)
	}

	render, clip := uuid.New(), uuid.New()
	for _, tc := range []struct {
		body string
		want int
	}{
		{`{"render_id":"` + render.String() + `","clip_id":"` + clip.String() + `","resolution":"removed"}`, 200},
		{`{"render_id":"` + render.String() + `","clip_id":"` + clip.String() + `","resolution":"deleted"}`, 400},
		{`{"render_id":"` + render.String() + `","resolution":"removed"}`, 400},
		{`{"render_id":"00000000-0000-0000-0000-000000000009","clip_id":"` + clip.String() + `","resolution":"kept"}`, 404},
		{`{"render_id":"` + render.String() + `","clip_id":"` + clip.String() + `","resolution":"kept","note":"` + strings.Repeat("n", 501) + `"}`, 400},
	} {
		if code, _ := do("POST", "/admin/removals", streamer.String(), tc.body); code != tc.want {
			t.Errorf("resolve %s = %d, want %d", tc.body[:40], code, tc.want)
		}
	}
	if store.resolved == nil || store.resolved.Resolution != "removed" {
		t.Fatalf("resolved = %+v", store.resolved)
	}
}
