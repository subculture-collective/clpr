package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

type fakeStatsStore struct {
	ids     []string
	listed  [2]time.Time
	saved   []models.CountdownVideoStats
	missing []string
}

func (f *fakeStatsStore) ListUploadsForStats(_ context.Context, after, stale time.Time, limit int) ([]string, error) {
	f.listed = [2]time.Time{after, stale}
	return f.ids[:min(limit, len(f.ids))], nil
}

func (f *fakeStatsStore) SaveUploadStats(_ context.Context, stats []models.CountdownVideoStats, missing []string, _ time.Time) error {
	f.saved = append(f.saved, stats...)
	f.missing = append(f.missing, missing...)
	return nil
}

type fakeStatsReader struct {
	calls [][]string
	err   error
}

func (f *fakeStatsReader) VideoStats(_ context.Context, ids []string) (map[string]YouTubeVideoStats, error) {
	f.calls = append(f.calls, ids)
	if f.err != nil {
		return nil, f.err
	}
	out := map[string]YouTubeVideoStats{}
	for i, id := range ids {
		if id != "gone" {
			out[id] = YouTubeVideoStats{ViewCount: int64(i), PrivacyStatus: "public"}
		}
	}
	return out, nil
}

func TestCountdownStatsRefresherBatchesAndMarksMissing(t *testing.T) {
	now := time.Date(2026, 11, 10, 3, 0, 0, 0, time.UTC)
	store := &fakeStatsStore{ids: []string{"gone"}}
	for i := 0; i < 120; i++ {
		store.ids = append(store.ids, fmt.Sprintf("vid%03d", i))
	}
	reader := &fakeStatsReader{}
	r := &CountdownStatsRefresher{Store: store, YouTube: reader, PerRun: 101, Now: func() time.Time { return now }}
	if err := r.RunOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(reader.calls) != 3 || len(reader.calls[0]) != 50 || len(reader.calls[2]) != 1 {
		t.Fatalf("calls = %d, sizes %d/%d", len(reader.calls), len(reader.calls[0]), len(reader.calls[len(reader.calls)-1]))
	}
	if len(store.saved) != 100 || len(store.missing) != 1 || store.missing[0] != "gone" {
		t.Fatalf("saved %d, missing %v", len(store.saved), store.missing)
	}
	if store.listed[0] != now.Add(-90*24*time.Hour) || store.listed[1] != now.Add(-6*time.Hour) {
		t.Fatalf("listed with %v", store.listed)
	}
	// The next cycle a minute later waits for the hourly interval.
	now = now.Add(time.Minute)
	if err := r.RunOnce(context.Background()); err != nil || len(reader.calls) != 3 {
		t.Fatalf("ran again within the interval: %d calls, %v", len(reader.calls), err)
	}
	now = now.Add(time.Hour)
	if err := r.RunOnce(context.Background()); err != nil || len(reader.calls) != 6 {
		t.Fatalf("did not run after the interval: %d calls, %v", len(reader.calls), err)
	}
}

func TestCountdownStatsRefresherPausesOnYouTubeErrors(t *testing.T) {
	now := time.Date(2026, 11, 10, 3, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		err  error
		want time.Time
	}{
		{&YouTubeError{Status: 403, Reason: "quotaExceeded"}, time.Date(2026, 11, 10, 8, 0, 0, 0, time.UTC)},
		{&YouTubeError{Status: 401}, now.Add(time.Hour)},
		{&YouTubeError{Status: 503}, now.Add(15 * time.Minute)},
		{&YouTubeError{Status: 400, Reason: "badRequest"}, now.Add(6 * time.Hour)},
	} {
		store := &fakeStatsStore{ids: []string{"a"}}
		reader := &fakeStatsReader{err: tc.err}
		r := &CountdownStatsRefresher{Store: store, YouTube: reader, Now: func() time.Time { return now }}
		if err := r.RunOnce(context.Background()); err != nil {
			t.Fatalf("%v: RunOnce returned %v", tc.err, err)
		}
		if !r.pausedUntil.Equal(tc.want) {
			t.Errorf("%v: paused until %v, want %v", tc.err, r.pausedUntil, tc.want)
		}
		r.lastRun = time.Time{} // only the pause holds it back
		if err := r.RunOnce(context.Background()); err != nil || len(reader.calls) != 1 {
			t.Errorf("%v: a paused refresher called YouTube again (%d calls, %v)", tc.err, len(reader.calls), err)
		}
	}
}
