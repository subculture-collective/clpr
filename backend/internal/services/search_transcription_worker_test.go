package services

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/countdownrender"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

type fakeSearchTranscriptionStore struct {
	clips     []repository.SearchTranscriptionClip
	query     repository.SearchTranscriptionQuery
	attempted int
	saved     map[uuid.UUID]string
	outcomes  map[uuid.UUID]string
	usage     []repository.TranscriptionUsage
	cleared   int
}

func (s *fakeSearchTranscriptionStore) NextClipsToTranscribe(_ context.Context, q repository.SearchTranscriptionQuery) ([]repository.SearchTranscriptionClip, error) {
	s.query = q
	return s.clips[:min(q.Limit, len(s.clips))], nil
}

func (s *fakeSearchTranscriptionStore) SaveSearchTranscript(_ context.Context, clipID uuid.UUID, _, text string, _ json.RawMessage) (bool, error) {
	if s.saved == nil {
		s.saved = map[uuid.UUID]string{}
	}
	s.saved[clipID] = text
	return true, nil
}

func (s *fakeSearchTranscriptionStore) RecordSearchTranscriptionAttempt(_ context.Context, clipID uuid.UUID, outcome string) error {
	if s.outcomes == nil {
		s.outcomes = map[uuid.UUID]string{}
	}
	s.outcomes[clipID] = outcome
	return nil
}

func (s *fakeSearchTranscriptionStore) SearchClipsAttemptedOn(context.Context, time.Time) (int, error) {
	return s.attempted, nil
}

func (s *fakeSearchTranscriptionStore) ClearExcludedTranscripts(context.Context) (int64, error) {
	s.cleared++
	return 0, nil
}

func (s *fakeSearchTranscriptionStore) RecordTranscriptionUsage(_ context.Context, u repository.TranscriptionUsage) error {
	s.usage = append(s.usage, u)
	return nil
}

func searchTranscriptionFixture(t *testing.T, n int) (*fakeSearchTranscriptionStore, *SearchTranscriptionWorker, *fakeCaptioner) {
	t.Helper()
	store := &fakeSearchTranscriptionStore{}
	for i := 0; i < n; i++ {
		store.clips = append(store.clips, repository.SearchTranscriptionClip{ClipID: uuid.New(), TwitchClipURL: "https://clips.twitch.tv/x"})
	}
	captioner := &fakeCaptioner{words: map[string][]countdownrender.Word{}}
	now := time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)
	worker := &SearchTranscriptionWorker{
		Store: store, Fetcher: &fakeFetcher{}, Prober: fakeProber{duration: 20}, Captioner: captioner,
		Dir: t.TempDir(), MinViews: 1000, MaxAge: 30 * 24 * time.Hour, DailyLimit: 500, BatchSize: 5,
		Now: func() time.Time { return now },
	}
	return store, worker, captioner
}

func TestSearchTranscriptionStoresTranscriptsAndDeletesFiles(t *testing.T) {
	store, worker, captioner := searchTranscriptionFixture(t, 2)
	spoken, silent := store.clips[0].ClipID, store.clips[1].ClipID
	spokenPath := filepath.Join(worker.Dir, ".search", spoken.String()+".mp4")
	captioner.words[spokenPath] = []countdownrender.Word{{Start: 0, End: 0.3, Text: " nice"}, {Start: 0.3, End: 0.6, Text: " shot"}}

	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if store.saved[spoken] != "nice shot" || len(store.saved) != 1 {
		t.Fatalf("saved = %v", store.saved)
	}
	if store.outcomes[spoken] != repository.SearchTranscriptionTranscribed || store.outcomes[silent] != repository.SearchTranscriptionNoSpeech {
		t.Fatalf("outcomes = %v", store.outcomes)
	}
	if len(store.usage) != 2 || !store.usage[0].Transcribed || store.usage[1].Transcribed ||
		store.usage[0].Source != repository.TranscriptSourceSearch || store.usage[0].AudioSeconds != 20 {
		t.Fatalf("usage = %+v", store.usage)
	}
	entries, err := os.ReadDir(filepath.Join(worker.Dir, ".search"))
	if err != nil || len(entries) != 0 {
		t.Fatalf("files left on the render host: %v %v", entries, err)
	}
	if store.cleared != 1 {
		t.Fatalf("excluded transcripts not cleared")
	}
	q := store.query
	if q.MinViews != 1000 || q.MaxAttempts != 3 || !q.CreatedAfter.Equal(worker.now().Add(-30*24*time.Hour)) ||
		!q.RetryBefore.Equal(worker.now().Add(-6*time.Hour)) {
		t.Fatalf("query = %+v", q)
	}
	if got := testutil.ToFloat64(worker.Metrics.WithLabelValues(repository.SearchTranscriptionNoSpeech)); got != 1 {
		t.Fatalf("no_speech metric = %v", got)
	}
}

func TestSearchTranscriptionKeepsWithinDailyLimitAndBatch(t *testing.T) {
	store, worker, _ := searchTranscriptionFixture(t, 10)
	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if store.query.Limit != 5 || len(store.outcomes) != 5 {
		t.Fatalf("batch: limit %d, attempted %d", store.query.Limit, len(store.outcomes))
	}

	store, worker, _ = searchTranscriptionFixture(t, 10)
	store.attempted = 498
	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(store.outcomes) != 2 {
		t.Fatalf("attempted %d clips with 2 left today", len(store.outcomes))
	}

	store, worker, _ = searchTranscriptionFixture(t, 10)
	store.attempted = 500
	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(store.outcomes) != 0 {
		t.Fatalf("attempted %d clips over the daily limit", len(store.outcomes))
	}
}

func TestSearchTranscriptionFailures(t *testing.T) {
	store, worker, _ := searchTranscriptionFixture(t, 2)
	gone, next := store.clips[0].ClipID, store.clips[1].ClipID
	worker.Fetcher = &fakeFetcher{failures: map[uuid.UUID]error{gone: errors.New("yt-dlp failed: clip deleted")}}
	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if store.outcomes[gone] != repository.SearchTranscriptionFailed || store.outcomes[next] != repository.SearchTranscriptionNoSpeech {
		t.Fatalf("a failed clip stopped the pass: %v", store.outcomes)
	}

	// Missing tools stop the pass without using up the clip's attempts.
	store, worker, _ = searchTranscriptionFixture(t, 2)
	worker.Fetcher = &fakeFetcher{failures: map[uuid.UUID]error{store.clips[0].ClipID: ErrCountdownMediaUnavailable}}
	if err := worker.RunOnce(context.Background()); !errors.Is(err, ErrCountdownMediaUnavailable) {
		t.Fatalf("err = %v", err)
	}
	if len(store.outcomes) != 0 {
		t.Fatalf("outcomes recorded: %v", store.outcomes)
	}

	store, worker, captioner := searchTranscriptionFixture(t, 1)
	captioner.err = errors.New("whisper crashed")
	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if store.outcomes[store.clips[0].ClipID] != repository.SearchTranscriptionFailed || len(store.saved) != 0 {
		t.Fatalf("outcomes = %v, saved = %v", store.outcomes, store.saved)
	}
}
