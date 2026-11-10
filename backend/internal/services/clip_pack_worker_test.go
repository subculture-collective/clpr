package services

import (
	"archive/zip"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/countdownrender"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
)

type fakePackStore struct {
	weeks     []time.Time
	pending   []*models.ClipPack
	completed map[uuid.UUID]string
	failed    map[uuid.UUID]string
	expires   map[uuid.UUID]time.Time
	toDelete  []models.ClipPack
	deleted   []uuid.UUID
}

func (f *fakePackStore) EnsureWeeklyPacks(_ context.Context, week time.Time, _ int) (int, error) {
	f.weeks = append(f.weeks, week)
	return 0, nil
}
func (f *fakePackStore) ClaimPackForBuild(context.Context, time.Duration) (*models.ClipPack, error) {
	if len(f.pending) == 0 {
		return nil, nil
	}
	p := f.pending[0]
	f.pending = f.pending[1:]
	return p, nil
}
func (f *fakePackStore) CompletePack(_ context.Context, id uuid.UUID, name, _ string, _ int64, _ int, expires time.Time) error {
	f.completed[id], f.expires[id] = name, expires
	return nil
}
func (f *fakePackStore) FailPack(_ context.Context, id uuid.UUID, reason string) error {
	f.failed[id] = reason
	return nil
}
func (f *fakePackStore) ListPacksToDelete(context.Context, time.Time) ([]models.ClipPack, error) {
	return f.toDelete, nil
}
func (f *fakePackStore) MarkPackDeleted(_ context.Context, id uuid.UUID) error {
	f.deleted = append(f.deleted, id)
	return nil
}

type fakePackFetcher struct{ fail map[string]bool }

func (f fakePackFetcher) Fetch(_ context.Context, c *models.CountdownEditionClip, dest string) (string, error) {
	if f.fail[c.Title] {
		return "", errors.New("no Helix authorization")
	}
	return "helix", os.WriteFile(dest, []byte("clip "+c.Title), 0o600)
}

type fakePackProber struct{}

func (fakePackProber) Probe(context.Context, string) (*MediaInfo, error) {
	return &MediaInfo{DurationSeconds: 20, Width: 1920, Height: 1080}, nil
}

type fakePackRenderer struct{ reel countdownrender.Edition }

func (f *fakePackRenderer) RenderCountdown(_ context.Context, ed countdownrender.Edition, out string) error {
	f.reel = ed
	return os.WriteFile(out, []byte("reel"), 0o600)
}
func (f *fakePackRenderer) RenderShort(_ context.Context, c countdownrender.Clip, out string) error {
	return os.WriteFile(out, []byte("short "+c.Title), 0o600)
}

func newPackWorker(t *testing.T, store *fakePackStore, fail map[string]bool, now time.Time) (*ClipPackWorker, *fakePackRenderer) {
	dir := t.TempDir()
	renderer := &fakePackRenderer{}
	return &ClipPackWorker{Store: store, Fetcher: fakePackFetcher{fail: fail}, Prober: fakePackProber{}, Renderer: renderer,
		WorkDir: filepath.Join(dir, "work"), PackDir: filepath.Join(dir, "packs"), Now: func() time.Time { return now }}, renderer
}

func newFakePackStore(packs ...*models.ClipPack) *fakePackStore {
	return &fakePackStore{pending: packs, completed: map[uuid.UUID]string{}, failed: map[uuid.UUID]string{}, expires: map[uuid.UUID]time.Time{}}
}

func TestClipPackWorkerBuildsZip(t *testing.T) {
	now := time.Date(2026, 11, 9, 7, 0, 0, 0, time.UTC) // Monday
	pack := &models.ClipPack{ID: uuid.New(), BroadcasterID: "1234", BroadcasterName: "Moon Moon", WeekStart: "2026-11-02", Clips: []models.ClipPackClip{
		{Position: 1, ClipID: uuid.New(), Title: "first", BroadcasterName: "Moon Moon", TwitchClipURL: "https://clips.twitch.tv/a", ViewCount: 900},
		{Position: 2, ClipID: uuid.New(), Title: "gone", BroadcasterName: "Moon Moon"},
		{Position: 3, ClipID: uuid.New(), Title: "third", BroadcasterName: "Moon Moon", TwitchClipURL: "https://clips.twitch.tv/c", ViewCount: 50},
	}}
	store := newFakePackStore(pack)
	w, renderer := newPackWorker(t, store, map[string]bool{"gone": true}, now)
	if err := w.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(store.weeks) != 1 || store.weeks[0] != time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC) {
		t.Fatalf("ensured weeks = %v", store.weeks)
	}
	name := store.completed[pack.ID]
	if name != pack.ID.String()+"/clpr-pack-moon-moon-2026-11-02.zip" || !store.expires[pack.ID].Equal(now.Add(7*24*time.Hour)) {
		t.Fatalf("completed %q expiring %v (failed: %v)", name, store.expires[pack.ID], store.failed)
	}
	if len(renderer.reel.Clips) != 2 || renderer.reel.Clips[1].Rank != 2 || renderer.reel.Clips[1].Title != "third" {
		t.Fatalf("reel clips = %+v", renderer.reel.Clips)
	}
	zr, err := zip.OpenReader(filepath.Join(w.PackDir, name))
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	var names []string
	contents := map[string]string{}
	for _, f := range zr.File {
		names = append(names, f.Name)
		rc, _ := f.Open()
		b, _ := io.ReadAll(rc)
		rc.Close()
		contents[f.Name] = string(b)
	}
	sort.Strings(names)
	if strings.Join(names, ",") != "01-9x16.mp4,02-9x16.mp4,credits.txt,highlights-16x9.mp4" {
		t.Fatalf("zip holds %v", names)
	}
	if contents["02-9x16.mp4"] != "short third" || !strings.Contains(contents["credits.txt"], "02. third (50 views)\n    https://clips.twitch.tv/c") {
		t.Fatalf("zip contents = %q", contents)
	}
	if entries, _ := os.ReadDir(w.WorkDir); len(entries) != 0 {
		t.Fatalf("work directory kept %d entries", len(entries))
	}

	// The next cycle within the hour does not query for new packs again.
	if err := w.RunOnce(context.Background()); err != nil || len(store.weeks) != 1 {
		t.Fatalf("ensured again: %v, %v", store.weeks, err)
	}
}

func TestClipPackWorkerFailsWithoutClips(t *testing.T) {
	pack := &models.ClipPack{ID: uuid.New(), BroadcasterID: "1234", BroadcasterName: "x", WeekStart: "2026-11-02",
		Clips: []models.ClipPackClip{{Position: 1, ClipID: uuid.New(), Title: "gone"}}}
	store := newFakePackStore(pack)
	w, _ := newPackWorker(t, store, map[string]bool{"gone": true}, time.Date(2026, 11, 9, 7, 0, 0, 0, time.UTC))
	if err := w.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if store.failed[pack.ID] != "no clip of the week could be downloaded" || store.completed[pack.ID] != "" {
		t.Fatalf("failed = %v, completed = %v", store.failed, store.completed)
	}
}

func TestClipPackLastWeek(t *testing.T) {
	w := &ClipPackWorker{ReadyHourUTC: 6}
	for _, tc := range []struct{ now, want string }{
		{"2026-11-09T05:59:00Z", "2026-10-26"}, // Monday before the hour: the week before last
		{"2026-11-09T06:00:00Z", "2026-11-02"},
		{"2026-11-15T23:00:00Z", "2026-11-02"}, // Sunday
		{"2026-11-16T07:00:00Z", "2026-11-09"},
	} {
		now, _ := time.Parse(time.RFC3339, tc.now)
		if got := w.lastWeek(now).Format("2006-01-02"); got != tc.want {
			t.Errorf("lastWeek(%s) = %s, want %s", tc.now, got, tc.want)
		}
	}
}

func TestClipPackSweepDeletesExpiredZips(t *testing.T) {
	id := uuid.New()
	name := id.String() + "/pack.zip"
	store := newFakePackStore()
	store.toDelete = []models.ClipPack{{ID: id, FileName: &name}, {ID: uuid.New(), FileName: strPtrPack("../outside.zip")}}
	w, _ := newPackWorker(t, store, nil, time.Now())
	if err := os.MkdirAll(filepath.Join(w.PackDir, id.String()), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(w.PackDir, name), []byte("zip"), 0o600); err != nil {
		t.Fatal(err)
	}
	err := w.Sweep(context.Background())
	if err == nil || !strings.Contains(err.Error(), "outside the pack directory") {
		t.Fatalf("sweep error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(w.PackDir, id.String())); !os.IsNotExist(err) {
		t.Fatalf("pack directory still there: %v", err)
	}
	if len(store.deleted) != 1 || store.deleted[0] != id {
		t.Fatalf("deleted = %v", store.deleted)
	}
}

func strPtrPack(s string) *string { return &s }

type fakePackFiles map[uuid.UUID]*models.ClipPackFile

func (f fakePackFiles) PackFile(_ context.Context, id uuid.UUID) (*models.ClipPackFile, error) {
	return f[id], nil
}

func TestClipPackLinksAndServer(t *testing.T) {
	now := time.Date(2026, 11, 9, 12, 0, 0, 0, time.UTC)
	links := ClipPackLinks{Secret: []byte("secret"), BaseURL: "https://packs.example/"}
	ready, expired, gone := uuid.New(), uuid.New(), uuid.New()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ready.String()), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ready.String(), "clpr-pack-x-2026-11-02.zip"), []byte("PK zip"), 0o600); err != nil {
		t.Fatal(err)
	}
	later, earlier := now.Add(6*24*time.Hour), now.Add(-time.Minute)
	server := &ClipPackServer{Dir: dir, Links: links, Now: func() time.Time { return now }, Store: fakePackFiles{
		ready:   {FileName: ready.String() + "/clpr-pack-x-2026-11-02.zip", Status: models.ClipPackReady, ExpiresAt: &later},
		expired: {FileName: expired.String() + "/a.zip", Status: models.ClipPackExpired, ExpiresAt: &earlier, Deleted: true},
		gone:    {FileName: gone.String() + "/missing.zip", Status: models.ClipPackReady, ExpiresAt: &later},
	}}

	get := func(target string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		server.ServeHTTP(w, httptest.NewRequest(http.MethodGet, strings.TrimPrefix(target, "https://packs.example"), nil))
		return w
	}
	url := links.URL(ready, later, now)
	if !strings.HasPrefix(url, "https://packs.example/packs/"+ready.String()+".zip?expires=") {
		t.Fatalf("url = %s", url)
	}
	w := get(url)
	if w.Code != http.StatusOK || w.Body.String() != "PK zip" || w.Header().Get("Content-Type") != "application/zip" ||
		w.Header().Get("Content-Disposition") != `attachment; filename="clpr-pack-x-2026-11-02.zip"` {
		t.Fatalf("download = %d %q %v", w.Code, w.Body.String(), w.Header())
	}
	if w := get(strings.Replace(url, "sig=", "sig=0", 1)); w.Code != http.StatusForbidden {
		t.Fatalf("tampered link = %d", w.Code)
	}
	if w := get(links.URL(expired, earlier, now)); w.Code != http.StatusForbidden {
		t.Fatalf("link past the pack's expiry = %d", w.Code)
	}
	if w := get(links.URL(expired, later, now)); w.Code != http.StatusGone {
		t.Fatalf("deleted pack = %d", w.Code)
	}
	if w := get(links.URL(gone, later, now)); w.Code != http.StatusGone {
		t.Fatalf("missing file = %d", w.Code)
	}
	if w := get("/packs/../etc/passwd.zip"); w.Code != http.StatusNotFound {
		t.Fatalf("bad path = %d", w.Code)
	}
	// A link expires after its TTL, even while the pack is still kept.
	server.Now = func() time.Time { return now.Add(25 * time.Hour) }
	if w := get(url); w.Code != http.StatusForbidden {
		t.Fatalf("day-old link = %d", w.Code)
	}
	if (ClipPackLinks{}).Valid(ready, "9999999999", "x", now) {
		t.Fatal("a link without a secret was valid")
	}
}

func TestHelixOnlyFetcherRefusesThePlayer(t *testing.T) {
	f := &TwitchClipMediaFetcher{HelixOnly: true, YTDLPPath: "/must/not/run"}
	_, err := f.Fetch(context.Background(), &models.CountdownEditionClip{TwitchClipURL: "https://clips.twitch.tv/x"}, filepath.Join(t.TempDir(), "x.mp4"))
	if !errors.Is(err, ErrCountdownMediaUnavailable) || !strings.Contains(err.Error(), "not authorized") {
		t.Fatalf("Helix-only fetch without authorization = %v", err)
	}
}
