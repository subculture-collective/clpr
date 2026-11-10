//go:build integration

package repository

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/google/uuid"
)

func TestSearchTranscriptionSelection(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "transcription_usage_daily", "countdown_exclusions", "clips", "users")
	ctx := context.Background()
	repo := NewSearchTranscriptionRepository(pool)
	now := time.Now()

	ids := insertCountdownClips(t, pool, now, []countdownClip{
		{name: "popular", broadcaster: "alpha", language: "en", duration: 30, age: time.Hour},
		{name: "most-popular", broadcaster: "alpha", language: "en", duration: 30, age: time.Hour},
		{name: "few-views", broadcaster: "bravo", language: "en", duration: 30, age: time.Hour},
		{name: "old", broadcaster: "bravo", language: "en", duration: 30, age: 40 * 24 * time.Hour},
		{name: "german", broadcaster: "bravo", language: "de", duration: 30, age: time.Hour},
		{name: "nsfw", broadcaster: "bravo", language: "en", duration: 30, age: time.Hour, nsfw: true},
		{name: "dmca", broadcaster: "bravo", language: "en", duration: 30, age: time.Hour, dmca: true},
		{name: "has-transcript", broadcaster: "bravo", language: "en", duration: 30, age: time.Hour},
		{name: "excluded-broadcaster", broadcaster: "charlie", language: "en", duration: 30, age: time.Hour},
		{name: "excluded-clip", broadcaster: "bravo", language: "en", duration: 30, age: time.Hour},
		{name: "silent", broadcaster: "bravo", language: "en", duration: 30, age: time.Hour},
		{name: "failed-recently", broadcaster: "bravo", language: "en", duration: 30, age: time.Hour},
		{name: "failed-earlier", broadcaster: "bravo", language: "en", duration: 30, age: time.Hour},
		{name: "failed-thrice", broadcaster: "bravo", language: "en", duration: 30, age: time.Hour},
	})
	if _, err := pool.Exec(ctx, `UPDATE clips SET view_count = 5000 WHERE id <> $1`, ids["few-views"]); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `UPDATE clips SET view_count = 9000 WHERE id = $1`, ids["most-popular"]); err != nil {
		t.Fatal(err)
	}
	words := json.RawMessage(`[]`)
	if stored, err := repo.SaveSearchTranscript(ctx, ids["has-transcript"], "en", "already here", words); err != nil || !stored {
		t.Fatalf("store = %v, %v", stored, err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO countdown_exclusions (broadcaster_id, reason, source) VALUES ('charlie', 'request', 'admin');
		INSERT INTO countdown_exclusions (clip_id, reason, source) VALUES ('`+ids["excluded-clip"].String()+`', 'request', 'admin')`); err != nil {
		t.Fatal(err)
	}
	record := func(name, outcome string, times int, at time.Time) {
		for i := 0; i < times; i++ {
			if err := repo.RecordSearchTranscriptionAttempt(ctx, ids[name], outcome); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := pool.Exec(ctx, `UPDATE search_transcription_attempts SET attempted_at = $2 WHERE clip_id = $1`, ids[name], at); err != nil {
			t.Fatal(err)
		}
	}
	record("silent", SearchTranscriptionNoSpeech, 1, now.Add(-48*time.Hour))
	record("failed-recently", SearchTranscriptionFailed, 1, now.Add(-time.Hour))
	record("failed-earlier", SearchTranscriptionFailed, 2, now.Add(-12*time.Hour))
	record("failed-thrice", SearchTranscriptionFailed, 3, now.Add(-12*time.Hour))

	got, err := repo.NextClipsToTranscribe(ctx, SearchTranscriptionQuery{MinViews: 1000, CreatedAfter: now.Add(-30 * 24 * time.Hour),
		RetryBefore: now.Add(-6 * time.Hour), MaxAttempts: 3, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	want := []uuid.UUID{ids["most-popular"]}
	// popular and failed-earlier tie on views; ids break the tie.
	rest := []uuid.UUID{ids["popular"], ids["failed-earlier"]}
	if rest[0].String() > rest[1].String() {
		rest[0], rest[1] = rest[1], rest[0]
	}
	want = append(want, rest...)
	if len(got) != len(want) {
		t.Fatalf("got %d clips, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i].ClipID != want[i] {
			t.Fatalf("clip %d = %s, want %s", i, got[i].ClipID, want[i])
		}
	}
	if got[0].TwitchClipURL != "https://clips.twitch.tv/most-popular" || got[0].BroadcasterID == nil || *got[0].BroadcasterID != "alpha" {
		t.Fatalf("clip fields = %+v", got[0])
	}

	// A transcript from another source is never replaced.
	if stored, err := NewCountdownRepository(pool).SaveCountdownTranscript(ctx, ids["popular"], "en", "countdown words", words); err != nil || !stored {
		t.Fatalf("countdown store = %v, %v", stored, err)
	}
	if stored, err := repo.SaveSearchTranscript(ctx, ids["popular"], "en", "search words", words); err != nil || stored {
		t.Fatalf("search transcript replaced another: %v, %v", stored, err)
	}

	// The daily limit counts every clip attempted.
	for i := 0; i < 3; i++ {
		if err := repo.RecordTranscriptionUsage(ctx, TranscriptionUsage{Day: now, Source: TranscriptSourceSearch, Transcribed: i == 0}); err != nil {
			t.Fatal(err)
		}
	}
	if n, err := repo.SearchClipsAttemptedOn(ctx, now); err != nil || n != 3 {
		t.Fatalf("attempted today = %d, %v", n, err)
	}
	if n, err := repo.SearchClipsAttemptedOn(ctx, now.Add(-48*time.Hour)); err != nil || n != 0 {
		t.Fatalf("attempted two days ago = %d, %v", n, err)
	}

	// Excluding a broadcaster empties the render host's transcripts of their
	// clips and queues the change for the search index.
	if stored, err := repo.SaveSearchTranscript(ctx, ids["most-popular"], "en", "alpha said this", words); err != nil || !stored {
		t.Fatalf("store = %v, %v", stored, err)
	}
	if _, err := pool.Exec(ctx, `UPDATE clip_transcripts SET search_indexed_at = updated_at`); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO countdown_exclusions (broadcaster_id, reason, source) VALUES ('alpha', 'request', 'self_service')`); err != nil {
		t.Fatal(err)
	}
	if n, err := repo.ClearExcludedTranscripts(ctx); err != nil || n != 2 {
		t.Fatalf("cleared = %d, %v", n, err)
	}
	if n, err := repo.ClearExcludedTranscripts(ctx); err != nil || n != 0 {
		t.Fatalf("cleared again = %d, %v", n, err)
	}
	pending, err := NewClipRepository(pool).ListTranscriptsToIndex(ctx, 10)
	if err != nil || len(pending) != 2 {
		t.Fatalf("pending = %+v, %v", pending, err)
	}
	for _, p := range pending {
		if p.FullText != "" {
			t.Fatalf("excluded transcript kept its text: %+v", p)
		}
	}
}
