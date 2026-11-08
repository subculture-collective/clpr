package services

import (
	"context"
	"errors"
	"fmt"
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

type fakeMediaStore struct {
	edition   *models.CountdownEdition
	series    *models.CountdownSeries
	claimed   bool
	backfill  []models.CountdownEditionClip
	recorded  []models.CountdownEditionMedia
	skipped   []uuid.UUID
	status    string
	reason    string
	toDelete  []models.CountdownEditionMedia
	deletedOK []uuid.UUID
}

func (s *fakeMediaStore) ClaimEditionForFetch(context.Context, time.Duration) (*models.CountdownEdition, *models.CountdownSeries, error) {
	if s.claimed || s.edition == nil {
		return nil, nil, nil
	}
	s.claimed = true
	return s.edition, s.series, nil
}

func (s *fakeMediaStore) RecordMedia(_ context.Context, m *models.CountdownEditionMedia) error {
	s.recorded = append(s.recorded, *m)
	return nil
}

func (s *fakeMediaStore) BackfillClip(_ context.Context, _ *models.CountdownEdition, _ *models.CountdownSeries, clipID uuid.UUID, _ string) (*models.CountdownEditionClip, error) {
	s.skipped = append(s.skipped, clipID)
	if len(s.backfill) == 0 {
		return nil, nil
	}
	next := s.backfill[0]
	s.backfill = s.backfill[1:]
	return &next, nil
}

func (s *fakeMediaStore) FinishFetch(_ context.Context, _ uuid.UUID, status string, reason *string) error {
	s.status = status
	if reason != nil {
		s.reason = *reason
	}
	return nil
}

func (s *fakeMediaStore) ListMediaToDelete(context.Context, time.Time) ([]models.CountdownEditionMedia, error) {
	return s.toDelete, nil
}

func (s *fakeMediaStore) MarkMediaDeleted(_ context.Context, _ uuid.UUID, clipID uuid.UUID) error {
	s.deletedOK = append(s.deletedOK, clipID)
	return nil
}

// fakeFetcher writes a small file, or fails for clips listed in failures.
type fakeFetcher struct {
	failures map[uuid.UUID]error
	fetched  []uuid.UUID
}

func (f *fakeFetcher) Fetch(_ context.Context, clip *models.CountdownEditionClip, dest string) (string, error) {
	f.fetched = append(f.fetched, clip.ClipID)
	if err := f.failures[clip.ClipID]; err != nil {
		return models.CountdownMediaSourcePlayer, err
	}
	return models.CountdownMediaSourcePlayer, os.WriteFile(dest, []byte("video:"+clip.ClipID.String()), 0o600)
}

type fakeProber struct{ duration float64 }

func (p fakeProber) Probe(context.Context, string) (*MediaInfo, error) {
	return &MediaInfo{DurationSeconds: p.duration, Width: 1920, Height: 1080}, nil
}

func mediaClip(position int) models.CountdownEditionClip {
	d := 30.0
	return models.CountdownEditionClip{Position: position, ClipID: uuid.New(), Duration: &d, TwitchClipURL: "https://clips.twitch.tv/x"}
}

func newMediaWorker(t *testing.T, store *fakeMediaStore, fetcher *fakeFetcher) *CountdownMediaWorker {
	t.Helper()
	return &CountdownMediaWorker{Store: store, Fetcher: fetcher, Prober: fakeProber{duration: 30.2}, Dir: t.TempDir(), MaxSkips: 2}
}

func mediaFixture(clips ...models.CountdownEditionClip) *fakeMediaStore {
	return &fakeMediaStore{
		edition: &models.CountdownEdition{ID: uuid.New(), SeriesSlug: "overall", EditionDate: "2026-11-07", Clips: clips},
		series:  &models.CountdownSeries{Slug: "overall"},
	}
}

func TestCountdownMediaWorkerFetchesEveryClip(t *testing.T) {
	a, b := mediaClip(1), mediaClip(2)
	store := mediaFixture(a, b)
	worker := newMediaWorker(t, store, &fakeFetcher{})

	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if store.status != models.CountdownEditionFetched || len(store.recorded) != 2 {
		t.Fatalf("status %q, recorded %d", store.status, len(store.recorded))
	}
	for _, m := range store.recorded {
		path := filepath.Join(worker.Dir, m.FileName)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("missing %s: %v", m.FileName, err)
		}
		if m.Bytes != int64(len(data)) || len(m.SHA256) != 64 || m.Width != 1920 || m.DurationSeconds != 30.2 {
			t.Fatalf("media record = %+v", m)
		}
		if _, err := os.Stat(path + ".partial"); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("partial file left behind")
		}
	}
}

func TestCountdownMediaWorkerBackfillsAFailedClip(t *testing.T) {
	a, b, c := mediaClip(1), mediaClip(2), mediaClip(3)
	store := mediaFixture(a, b)
	store.backfill = []models.CountdownEditionClip{c}
	fetcher := &fakeFetcher{failures: map[uuid.UUID]error{a.ClipID: errors.New("clip was deleted")}}
	worker := newMediaWorker(t, store, fetcher)

	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if store.status != models.CountdownEditionFetched {
		t.Fatalf("status = %q (%s)", store.status, store.reason)
	}
	if len(store.skipped) != 1 || store.skipped[0] != a.ClipID {
		t.Fatalf("skipped = %v", store.skipped)
	}
	if len(store.recorded) != 2 || store.recorded[0].ClipID != b.ClipID || store.recorded[1].ClipID != c.ClipID {
		t.Fatalf("recorded = %+v", store.recorded)
	}
}

func TestCountdownMediaWorkerFailsAfterTooManySkips(t *testing.T) {
	a := mediaClip(1)
	store := mediaFixture(a)
	failing := map[uuid.UUID]error{a.ClipID: errors.New("HTTP 404")}
	for i := 0; i < 5; i++ {
		c := mediaClip(2 + i)
		failing[c.ClipID] = errors.New("HTTP 404")
		store.backfill = append(store.backfill, c)
	}
	worker := newMediaWorker(t, store, &fakeFetcher{failures: failing})

	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if store.status != models.CountdownEditionFailed || store.reason != "3 clips could not be fetched; last error: HTTP 404" {
		t.Fatalf("status %q reason %q", store.status, store.reason)
	}
	if len(store.skipped) != 2 {
		t.Fatalf("skipped %d clips before giving up, want 2", len(store.skipped))
	}
}

func TestCountdownMediaWorkerFailsWhenNoBackfillIsLeft(t *testing.T) {
	a := mediaClip(1)
	store := mediaFixture(a)
	worker := newMediaWorker(t, store, &fakeFetcher{failures: map[uuid.UUID]error{a.ClipID: errors.New("gone")}})

	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if store.status != models.CountdownEditionFailed || store.reason != "not enough eligible clips after 1 skips" {
		t.Fatalf("status %q reason %q", store.status, store.reason)
	}
}

func TestCountdownMediaWorkerFailsTheEditionWhenTheSourceIsDown(t *testing.T) {
	a, b := mediaClip(1), mediaClip(2)
	store := mediaFixture(a, b)
	down := fmt.Errorf("%w: yt-dlp not found", ErrCountdownMediaUnavailable)
	worker := newMediaWorker(t, store, &fakeFetcher{failures: map[uuid.UUID]error{a.ClipID: down}})

	err := worker.RunOnce(context.Background())
	if !errors.Is(err, ErrCountdownMediaUnavailable) {
		t.Fatalf("err = %v", err)
	}
	if store.status != models.CountdownEditionFailed || len(store.skipped) != 0 {
		t.Fatalf("status %q skipped %v", store.status, store.skipped)
	}
}

func TestCountdownMediaWorkerSkipsAFileOfTheWrongLength(t *testing.T) {
	a := mediaClip(1)
	store := mediaFixture(a)
	worker := newMediaWorker(t, store, &fakeFetcher{})
	worker.Prober = fakeProber{duration: 12}

	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if store.status != models.CountdownEditionFailed || len(store.recorded) != 0 {
		t.Fatalf("status %q recorded %d", store.status, len(store.recorded))
	}
	entries, _ := os.ReadDir(filepath.Join(worker.Dir, store.edition.ID.String()))
	if len(entries) != 0 {
		t.Fatalf("rejected file kept: %v", entries)
	}
}

func TestCountdownMediaSweepDeletesExpiredAndStrayFiles(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC)
	editionID, tracked := uuid.New(), uuid.New()
	if err := os.MkdirAll(filepath.Join(dir, editionID.String()), 0o700); err != nil {
		t.Fatal(err)
	}
	write := func(name string, age time.Duration) string {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		mtime := now.Add(-age)
		if err := os.Chtimes(path, mtime, mtime); err != nil {
			t.Fatal(err)
		}
		return path
	}
	trackedPath := write(filepath.Join(editionID.String(), tracked.String()+".mp4"), time.Hour)
	stray := write(filepath.Join(editionID.String(), "old.partial"), 25*time.Hour)
	fresh := write(filepath.Join(editionID.String(), "fresh.mp4"), time.Hour)

	store := &fakeMediaStore{toDelete: []models.CountdownEditionMedia{
		{EditionID: editionID, ClipID: tracked, FileName: filepath.Join(editionID.String(), tracked.String()+".mp4")},
		{EditionID: editionID, ClipID: uuid.New(), FileName: "../outside.mp4"},
	}}
	worker := &CountdownMediaWorker{Store: store, Dir: dir, Retention: 48 * time.Hour, Now: func() time.Time { return now }}

	err := worker.Sweep(context.Background())
	if err == nil {
		t.Fatal("a path outside the media directory was not refused")
	}
	for _, path := range []string{trackedPath, stray} {
		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("%s was not deleted", path)
		}
	}
	if _, err := os.Stat(fresh); err != nil {
		t.Errorf("fresh untracked file was deleted: %v", err)
	}
	if len(store.deletedOK) != 1 || store.deletedOK[0] != tracked {
		t.Fatalf("marked deleted = %v", store.deletedOK)
	}
}

func TestParseFFprobe(t *testing.T) {
	info, err := parseFFprobe([]byte(`{"streams":[{"codec_type":"audio"},{"codec_type":"video","width":1280,"height":720}],"format":{"duration":"29.983"}}`))
	if err != nil || info.Width != 1280 || info.Height != 720 || info.DurationSeconds != 29.983 {
		t.Fatalf("info = %+v, err = %v", info, err)
	}
	for _, bad := range []string{`{"streams":[{"codec_type":"audio"}],"format":{"duration":"3"}}`, `{"streams":[{"codec_type":"video","width":1,"height":1}],"format":{}}`, `nope`} {
		if _, err := parseFFprobe([]byte(bad)); err == nil {
			t.Errorf("parseFFprobe(%s) accepted", bad)
		}
	}
}

func TestCountdownClipURL(t *testing.T) {
	for _, ok := range []string{"https://clips.twitch.tv/BraveSlug-abc", "https://www.twitch.tv/someone/clip/BraveSlug-abc"} {
		if _, err := countdownClipURL(ok); err != nil {
			t.Errorf("%s rejected: %v", ok, err)
		}
	}
	for _, bad := range []string{"http://clips.twitch.tv/x", "https://evil.example/clip/x", "https://user:pw@clips.twitch.tv/x", "https://clips.twitch.tv/", "https://www.twitch.tv/someone"} {
		if _, err := countdownClipURL(bad); err == nil {
			t.Errorf("%s accepted", bad)
		}
	}
}

type mediaAuthStub struct{ auth *models.TwitchAuth }

func (s mediaAuthStub) GetTwitchAuthByTwitchUserID(context.Context, string) (*models.TwitchAuth, error) {
	return s.auth, nil
}

type mediaDownloadStub struct {
	url   string
	calls int
}

func (s *mediaDownloadStub) GetDownloadURL(_ context.Context, broadcasterID, clipID, token string) (string, error) {
	s.calls++
	if broadcasterID != "42" || clipID != "BraveSlug" || token != "user-token" {
		return "", fmt.Errorf("unexpected request %s %s", broadcasterID, clipID)
	}
	return s.url, nil
}

// fakeYTDLP writes a script that records its arguments and writes the -o file.
func fakeYTDLP(t *testing.T) (path, argsFile string) {
	t.Helper()
	dir := t.TempDir()
	argsFile = filepath.Join(dir, "args")
	path = filepath.Join(dir, "yt-dlp")
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > " + argsFile + "\nwhile [ $# -gt 0 ]; do if [ \"$1\" = -o ]; then printf player > \"$2\"; fi; shift; done\n"
	if err := os.WriteFile(path, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	return path, argsFile
}

func TestTwitchClipMediaFetcherSources(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("helix"))
	}))
	defer server.Close()
	ytdlp, argsFile := fakeYTDLP(t)
	broadcaster := "42"
	clip := &models.CountdownEditionClip{ClipID: uuid.New(), BroadcasterID: &broadcaster, TwitchClipID: "BraveSlug",
		TwitchClipURL: "https://clips.twitch.tv/BraveSlug"}
	now := time.Date(2026, 11, 8, 0, 0, 0, 0, time.UTC)
	granted := &models.TwitchAuth{AccessToken: "user-token", Scopes: "user:read:email channel:manage:clips", ExpiresAt: now.Add(time.Hour)}

	cases := []struct {
		name   string
		auth   *models.TwitchAuth
		want   string
		helix  int
		player bool
	}{
		{"granted scope uses Helix", granted, "helix", 1, false},
		{"no authorization uses the player files", nil, "player", 0, true},
		{"expired token uses the player files", &models.TwitchAuth{AccessToken: "user-token", Scopes: granted.Scopes, ExpiresAt: now.Add(-time.Minute)}, "player", 0, true},
		{"missing scope uses the player files", &models.TwitchAuth{AccessToken: "user-token", Scopes: "user:read:email", ExpiresAt: now.Add(time.Hour)}, "player", 0, true},
	}
	for _, tc := range cases {
		_ = os.Remove(argsFile)
		downloads := &mediaDownloadStub{url: server.URL + "/clip.mp4"}
		fetcher := &TwitchClipMediaFetcher{Auth: mediaAuthStub{tc.auth}, Downloads: downloads, HTTPClient: server.Client(),
			YTDLPPath: ytdlp, Now: func() time.Time { return now }}
		dest := filepath.Join(t.TempDir(), "clip.mp4.partial")
		source, err := fetcher.Fetch(context.Background(), clip, dest)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		data, _ := os.ReadFile(dest)
		if source != tc.want || string(data) != tc.want || downloads.calls != tc.helix {
			t.Fatalf("%s: source %q, file %q, helix calls %d", tc.name, source, data, downloads.calls)
		}
		args, err := os.ReadFile(argsFile)
		if tc.player != (err == nil) {
			t.Fatalf("%s: yt-dlp ran = %v", tc.name, err == nil)
		}
		if tc.player && !strings.Contains(string(args), "--no-config\n") || tc.player && !strings.HasSuffix(string(args), "--\nhttps://clips.twitch.tv/BraveSlug\n") {
			t.Fatalf("%s: yt-dlp args = %q", tc.name, args)
		}
	}

	missing := &TwitchClipMediaFetcher{YTDLPPath: filepath.Join(t.TempDir(), "absent")}
	if _, err := missing.Fetch(context.Background(), clip, filepath.Join(t.TempDir(), "x")); !errors.Is(err, ErrCountdownMediaUnavailable) {
		t.Fatalf("missing yt-dlp err = %v", err)
	}
}
