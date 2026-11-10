package services

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
)

type fakeTokenStore struct{ saved string }

func (f *fakeTokenStore) RefreshToken(_ context.Context, _ uuid.UUID, access, refresh, scopes string, _ time.Time) error {
	f.saved = access + "|" + refresh + "|" + scopes
	return nil
}

type fakeAuthLookup struct{ auth *models.TwitchAuth }

func (f fakeAuthLookup) GetTwitchAuthByTwitchUserID(context.Context, string) (*models.TwitchAuth, error) {
	copied := *f.auth
	return &copied, nil
}

type fakeDownloadURL struct{ url, token string }

func (f *fakeDownloadURL) GetDownloadURL(_ context.Context, _, _, token string) (string, error) {
	f.token = token
	return f.url, nil
}

func TestHelixOnlyFetchRefreshesAnExpiredToken(t *testing.T) {
	now := time.Date(2026, 11, 9, 7, 0, 0, 0, time.UTC)
	media := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "clip bytes")
	}))
	defer media.Close()
	tokens := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.Form.Get("refresh_token") != "refresh-1" || r.Form.Get("client_secret") != "secret" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_, _ = io.WriteString(w, `{"access_token":"fresh","refresh_token":"refresh-2","expires_in":14400,"scope":["channel:manage:clips"]}`)
	}))
	defer tokens.Close()
	store := &fakeTokenStore{}
	downloads := &fakeDownloadURL{url: media.URL + "/clip.mp4"}
	expired := &models.TwitchAuth{UserID: uuid.New(), AccessToken: "stale", RefreshToken: "refresh-1",
		Scopes: "channel:manage:clips", ExpiresAt: now.Add(-time.Hour)}
	f := &TwitchClipMediaFetcher{HelixOnly: true, Auth: fakeAuthLookup{expired}, Downloads: downloads, HTTPClient: media.Client(),
		Refresher: &TwitchTokenRefresher{ClientID: "id", ClientSecret: "secret", Store: store, TokenURL: tokens.URL,
			Now: func() time.Time { return now }},
		Now: func() time.Time { return now }}
	broadcaster := "1234"
	dest := filepath.Join(t.TempDir(), "clip.mp4")
	source, err := f.Fetch(context.Background(), &models.CountdownEditionClip{BroadcasterID: &broadcaster, TwitchClipID: "c"}, dest)
	if err != nil || source != models.CountdownMediaSourceHelix || downloads.token != "fresh" {
		t.Fatalf("fetch = %q, %v (token %q)", source, err, downloads.token)
	}
	if store.saved != "fresh|refresh-2|channel:manage:clips" {
		t.Fatalf("saved = %q", store.saved)
	}
	if b, _ := os.ReadFile(dest); string(b) != "clip bytes" {
		t.Fatalf("file = %q", b)
	}

	// A refresh Twitch refuses leaves the clip out instead of using the player.
	expired.RefreshToken = "revoked"
	if _, err := f.Fetch(context.Background(), &models.CountdownEditionClip{BroadcasterID: &broadcaster, TwitchClipID: "c"}, dest); err == nil ||
		!strings.Contains(err.Error(), "not authorized") {
		t.Fatalf("refused refresh = %v", err)
	}
}
