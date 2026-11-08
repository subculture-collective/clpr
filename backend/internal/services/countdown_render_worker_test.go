package services

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/countdownrender"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
)

type fakeRenderStore struct {
	edition  *models.CountdownEdition
	series   *models.CountdownSeries
	media    []models.CountdownEditionMedia
	claimed  bool
	renders  []models.CountdownRender
	status   string
	reason   string
	deleted  []uuid.UUID
	expired  []models.CountdownRender
	unlisted []uuid.UUID
}

func (s *fakeRenderStore) ClaimEditionForRender(context.Context, time.Duration) (*models.CountdownEdition, *models.CountdownSeries, []models.CountdownEditionMedia, error) {
	if s.claimed || s.edition == nil {
		return nil, nil, nil, nil
	}
	s.claimed = true
	return s.edition, s.series, s.media, nil
}

func (s *fakeRenderStore) RecordRender(_ context.Context, x *models.CountdownRender) error {
	s.renders = append(s.renders, *x)
	return nil
}

func (s *fakeRenderStore) FinishRender(_ context.Context, _ uuid.UUID, status string, reason *string) error {
	s.status = status
	if reason != nil {
		s.reason = *reason
	}
	return nil
}

func (s *fakeRenderStore) MarkMediaDeleted(_ context.Context, _ uuid.UUID, clipID uuid.UUID) error {
	s.deleted = append(s.deleted, clipID)
	return nil
}

func (s *fakeRenderStore) ListRendersToDelete(context.Context, time.Time) ([]models.CountdownRender, error) {
	return s.expired, nil
}

func (s *fakeRenderStore) MarkRenderDeleted(_ context.Context, id uuid.UUID) error {
	s.unlisted = append(s.unlisted, id)
	return nil
}

// fakeRenderer writes a small file per output and remembers what it drew.
type fakeRenderer struct {
	countdown countdownrender.Edition
	shorts    []countdownrender.Clip
	fail      error
}

func (r *fakeRenderer) RenderCountdown(_ context.Context, ed countdownrender.Edition, out string) error {
	r.countdown = ed
	if r.fail != nil {
		return r.fail
	}
	return os.WriteFile(out, []byte("countdown"), 0o600)
}

func (r *fakeRenderer) RenderShort(_ context.Context, c countdownrender.Clip, out string) error {
	r.shorts = append(r.shorts, c)
	return os.WriteFile(out, []byte("short"), 0o600)
}

func (r *fakeRenderer) RenderThumbnail(_ context.Context, _ countdownrender.Edition, out string) error {
	return os.WriteFile(out, []byte("thumb"), 0o600)
}

func renderFixture(t *testing.T, shorts int) (*fakeRenderStore, *CountdownRenderWorker) {
	t.Helper()
	mediaDir, renderDir := t.TempDir(), t.TempDir()
	edition := &models.CountdownEdition{ID: uuid.New(), SeriesSlug: "overall", EditionDate: "2026-11-07"}
	store := &fakeRenderStore{edition: edition, series: &models.CountdownSeries{Name: "Overall", ShortsCount: shorts}}
	// Stored out of order to check that ranks follow positions.
	for _, pos := range []int{2, 1, 3} {
		clipID := uuid.New()
		name := filepath.Join(edition.ID.String(), clipID.String()+".mp4")
		if err := os.MkdirAll(filepath.Join(mediaDir, edition.ID.String()), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(mediaDir, name), []byte("clip"), 0o600); err != nil {
			t.Fatal(err)
		}
		edition.Clips = append(edition.Clips, models.CountdownEditionClip{Position: pos, ClipID: clipID, Title: "clip", BroadcasterName: "b"})
		store.media = append(store.media, models.CountdownEditionMedia{EditionID: edition.ID, ClipID: clipID, FileName: name})
	}
	worker := &CountdownRenderWorker{Store: store, Renderer: &fakeRenderer{}, Prober: fakeProber{duration: 40},
		MediaDir: mediaDir, RenderDir: renderDir}
	return store, worker
}

type fakeEditionNotifier struct{ editions []uuid.UUID }

func (f *fakeEditionNotifier) EditionPublished(_ context.Context, e *models.CountdownEdition) error {
	f.editions = append(f.editions, e.ID)
	return nil
}

func TestCountdownRenderWorkerRendersAndDeletesSources(t *testing.T) {
	store, worker := renderFixture(t, 2)
	notifier := &fakeEditionNotifier{}
	worker.Notifier = notifier
	renderer := worker.Renderer.(*fakeRenderer)
	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(notifier.editions) != 1 || notifier.editions[0] != store.edition.ID {
		t.Fatalf("announced editions = %v", notifier.editions)
	}
	if store.status != models.CountdownEditionRendered {
		t.Fatalf("status %q (%s)", store.status, store.reason)
	}
	ranks := []int{}
	for _, c := range renderer.countdown.Clips {
		ranks = append(ranks, c.Rank)
	}
	if len(ranks) != 3 || ranks[0] != 1 || ranks[1] != 2 || ranks[2] != 3 {
		t.Fatalf("ranks = %v", ranks)
	}
	if len(renderer.shorts) != 2 || renderer.shorts[0].Rank != 1 || renderer.shorts[1].Rank != 2 {
		t.Fatalf("shorts = %+v", renderer.shorts)
	}
	kinds := map[string]int{}
	for _, x := range store.renders {
		kinds[x.Kind]++
		path := filepath.Join(worker.RenderDir, x.FileName)
		if _, err := os.Stat(path); err != nil || len(x.SHA256) != 64 || x.Bytes == 0 {
			t.Fatalf("render %+v: %v", x, err)
		}
		if (x.Kind == models.CountdownRenderShort) != (x.ClipID != nil) || (x.Kind == models.CountdownRenderThumbnail) != (x.DurationSeconds == nil) {
			t.Fatalf("render fields for %s: %+v", x.Kind, x)
		}
	}
	if kinds["countdown"] != 1 || kinds["thumbnail"] != 1 || kinds["short"] != 2 {
		t.Fatalf("render kinds = %v", kinds)
	}
	if len(store.deleted) != 3 {
		t.Fatalf("deleted sources = %d, want 3", len(store.deleted))
	}
	for _, m := range store.media {
		if _, err := os.Stat(filepath.Join(worker.MediaDir, m.FileName)); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("source %s still on disk", m.FileName)
		}
	}
}

func TestCountdownRenderWorkerFailureStillDeletesSources(t *testing.T) {
	store, worker := renderFixture(t, 0)
	notifier := &fakeEditionNotifier{}
	worker.Notifier = notifier
	worker.Renderer.(*fakeRenderer).fail = errors.New("ffmpeg failed: boom")
	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(notifier.editions) != 0 {
		t.Fatalf("a failed edition was announced: %v", notifier.editions)
	}
	if store.status != models.CountdownEditionFailed || store.reason != "render countdown: ffmpeg failed: boom" {
		t.Fatalf("status %q reason %q", store.status, store.reason)
	}
	if len(store.deleted) != 3 || len(store.renders) != 0 {
		t.Fatalf("deleted %d sources, recorded %d renders", len(store.deleted), len(store.renders))
	}
	if _, err := os.Stat(filepath.Join(worker.RenderDir, store.edition.ID.String())); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("render directory of a failed edition kept: %v", err)
	}
}

func TestCountdownRenderWorkerFailsWhenMediaIsMissing(t *testing.T) {
	store, worker := renderFixture(t, 0)
	store.media = store.media[1:] // the first stored clip holds position 2
	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if store.status != models.CountdownEditionFailed || store.reason != "media for clip #2 is missing" {
		t.Fatalf("status %q reason %q", store.status, store.reason)
	}
}

func TestCountdownRenderWorkerKeepsSourcesOnShutdown(t *testing.T) {
	store, worker := renderFixture(t, 0)
	ctx, cancel := context.WithCancel(context.Background())
	worker.Renderer.(*fakeRenderer).fail = context.Canceled
	cancel()
	if err := worker.RunOnce(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v", err)
	}
	if store.status != "" || len(store.deleted) != 0 {
		t.Fatalf("shutdown finished the edition (%q) or deleted %d sources", store.status, len(store.deleted))
	}
}

func TestCountdownRenderSweepDeletesExpiredRenders(t *testing.T) {
	store, worker := renderFixture(t, 0)
	now := time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC)
	worker.Now = func() time.Time { return now }
	name := filepath.Join(store.edition.ID.String(), "countdown.mp4")
	if err := os.MkdirAll(filepath.Join(worker.RenderDir, store.edition.ID.String()), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(worker.RenderDir, name), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	expired := models.CountdownRender{ID: uuid.New(), FileName: name}
	store.expired = []models.CountdownRender{expired, {ID: uuid.New(), FileName: "../escape.mp4"}}
	if err := worker.Sweep(context.Background()); err == nil {
		t.Fatal("a path outside the render directory was not refused")
	}
	if _, err := os.Stat(filepath.Join(worker.RenderDir, name)); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("expired render kept")
	}
	if len(store.unlisted) != 1 || store.unlisted[0] != expired.ID {
		t.Fatalf("marked deleted = %v", store.unlisted)
	}
}
