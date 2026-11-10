package services

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

type fakeYouTubeAPI struct {
	mu           sync.Mutex
	tokenCalls   int
	uploadMeta   map[string]any
	uploadBody   []byte
	thumbnailFor string
	playlistBody map[string]any
	itemBody     map[string]any
	failStart    string // error reason for the upload start, if any
}

func (f *fakeYouTubeAPI) server(t *testing.T) *httptest.Server {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		if r.URL.Path == "/token" {
			_ = r.ParseForm()
			if r.Form.Get("refresh_token") != "refresh" || r.Form.Get("grant_type") != "refresh_token" {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = io.WriteString(w, `{"error":"invalid_grant"}`)
				return
			}
			f.tokenCalls++
			_, _ = io.WriteString(w, `{"access_token":"access","expires_in":3600}`)
			return
		}
		if r.Header.Get("Authorization") != "Bearer access" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/upload/videos":
			if f.failStart != "" {
				w.WriteHeader(http.StatusForbidden)
				_, _ = io.WriteString(w, `{"error":{"message":"limit","errors":[{"reason":"`+f.failStart+`"}]}}`)
				return
			}
			if r.URL.Query().Get("uploadType") != "resumable" || r.Header.Get("X-Upload-Content-Length") == "" {
				t.Errorf("upload start %s with headers %v", r.URL, r.Header)
			}
			_ = json.NewDecoder(r.Body).Decode(&f.uploadMeta)
			w.Header().Set("Location", srv.URL+"/session/1")
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodPut && r.URL.Path == "/session/1":
			f.uploadBody, _ = io.ReadAll(r.Body)
			_, _ = io.WriteString(w, `{"id":"vid_ABC123"}`)
		case r.URL.Path == "/upload/thumbnails/set":
			f.thumbnailFor = r.URL.Query().Get("videoId")
			_, _ = io.WriteString(w, `{}`)
		case r.URL.Path == "/api/playlists":
			_ = json.NewDecoder(r.Body).Decode(&f.playlistBody)
			_, _ = io.WriteString(w, `{"id":"PLxxxxxxxxxxxx"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/api/videos":
			if r.URL.Query().Get("part") != "statistics,status" || r.URL.Query().Get("id") != "vid_A,vid_gone" {
				t.Errorf("videos.list %s", r.URL)
			}
			_, _ = io.WriteString(w, `{"items":[{"id":"vid_A","statistics":{"viewCount":"1234"},"status":{"privacyStatus":"public"}}]}`)
		case r.URL.Path == "/api/playlistItems":
			_ = json.NewDecoder(r.Body).Decode(&f.itemBody)
			_, _ = io.WriteString(w, `{"id":"item"}`)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	return srv
}

func testYouTubeClient(srv *httptest.Server, refresh string) *YouTubeClient {
	return &YouTubeClient{ClientID: "id", ClientSecret: "secret", RefreshToken: refresh,
		OAuthURL: srv.URL + "/token", APIBase: srv.URL + "/api", UploadBase: srv.URL + "/upload"}
}

func TestYouTubeClientUploadsAndOrganizes(t *testing.T) {
	api := &fakeYouTubeAPI{}
	srv := api.server(t)
	defer srv.Close()
	c := testYouTubeClient(srv, "refresh")
	video := filepath.Join(t.TempDir(), "countdown.mp4")
	if err := os.WriteFile(video, []byte("video bytes"), 0o600); err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 11, 8, 15, 0, 0, 0, time.UTC)
	id, err := c.UploadVideo(context.Background(), video, YouTubeVideo{Title: "Top 10", Description: "d", Tags: []string{"twitch"},
		CategoryID: "20", PrivacyStatus: "private", PublishAt: &at})
	if err != nil || id != "vid_ABC123" {
		t.Fatalf("upload = %q, %v", id, err)
	}
	if string(api.uploadBody) != "video bytes" {
		t.Fatalf("uploaded %q", api.uploadBody)
	}
	status, _ := api.uploadMeta["status"].(map[string]any)
	snippet, _ := api.uploadMeta["snippet"].(map[string]any)
	if status["privacyStatus"] != "private" || status["publishAt"] != "2026-11-08T15:00:00Z" || status["selfDeclaredMadeForKids"] != false ||
		snippet["title"] != "Top 10" || snippet["categoryId"] != "20" {
		t.Fatalf("upload metadata = %v", api.uploadMeta)
	}
	if err := c.SetThumbnail(context.Background(), id, video); err != nil || api.thumbnailFor != id {
		t.Fatalf("thumbnail for %q: %v", api.thumbnailFor, err)
	}
	playlist, err := c.CreatePlaylist(context.Background(), "Daily", "desc", "private")
	if err != nil || playlist != "PLxxxxxxxxxxxx" {
		t.Fatalf("playlist = %q, %v", playlist, err)
	}
	if err := c.AddToPlaylist(context.Background(), playlist, id); err != nil {
		t.Fatal(err)
	}
	if snippet, _ := api.itemBody["snippet"].(map[string]any); snippet["playlistId"] != playlist {
		t.Fatalf("playlist item = %v", api.itemBody)
	}
	if api.tokenCalls != 1 {
		t.Fatalf("token refreshed %d times, want once", api.tokenCalls)
	}
}

func TestYouTubeClientErrors(t *testing.T) {
	api := &fakeYouTubeAPI{failStart: "uploadLimitExceeded"}
	srv := api.server(t)
	defer srv.Close()
	video := filepath.Join(t.TempDir(), "v.mp4")
	_ = os.WriteFile(video, []byte("x"), 0o600)

	_, err := testYouTubeClient(srv, "refresh").UploadVideo(context.Background(), video, YouTubeVideo{PrivacyStatus: "private"})
	if !IsYouTubeQuotaError(err) {
		t.Fatalf("upload limit error = %v", err)
	}
	_, err = testYouTubeClient(srv, "revoked").UploadVideo(context.Background(), video, YouTubeVideo{PrivacyStatus: "private"})
	var yt *YouTubeError
	if !asYouTubeError(err, &yt) || !yt.AuthFailed() || strings.Contains(err.Error(), "revoked") {
		t.Fatalf("revoked token error = %v", err)
	}
	for _, tc := range []struct {
		err       YouTubeError
		transient bool
	}{
		{YouTubeError{Status: 503}, true},
		{YouTubeError{Status: 403, Reason: "rateLimitExceeded"}, true},
		{YouTubeError{Status: 400, Reason: "invalidTitle"}, false},
	} {
		if tc.err.Transient() != tc.transient || tc.err.QuotaExhausted() {
			t.Errorf("%+v: transient %v", tc.err, tc.err.Transient())
		}
	}
}

func asYouTubeError(err error, target **YouTubeError) bool {
	if e, ok := err.(*YouTubeError); ok {
		*target = e
		return true
	}
	return false
}

func TestYouTubeClientVideoStats(t *testing.T) {
	api := &fakeYouTubeAPI{}
	srv := api.server(t)
	defer srv.Close()
	c := testYouTubeClient(srv, "refresh")
	stats, err := c.VideoStats(context.Background(), []string{"vid_A", "vid_gone"})
	if err != nil || len(stats) != 1 || stats["vid_A"] != (YouTubeVideoStats{ViewCount: 1234, PrivacyStatus: "public"}) {
		t.Fatalf("stats = %+v, %v", stats, err)
	}
	if stats, err := c.VideoStats(context.Background(), nil); err != nil || len(stats) != 0 {
		t.Fatalf("no IDs = %+v, %v", stats, err)
	}
	if _, err := c.VideoStats(context.Background(), make([]string, 51)); err == nil {
		t.Fatal("51 IDs were accepted")
	}
}
