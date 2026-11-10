//go:build integration

package repository

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/google/uuid"
)

func TestTranscriptSearchStorage(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "transcription_usage_daily", "clips", "users")
	ctx := context.Background()
	countdowns := NewCountdownRepository(pool)
	clips := NewClipRepository(pool)

	ids := insertCountdownClips(t, pool, time.Now(), []countdownClip{
		{name: "said-it", broadcaster: "alpha", velocity: 1, duration: 30, age: time.Hour},
		{name: "authorized", broadcaster: "bravo", velocity: 1, duration: 30, age: time.Hour},
		{name: "quiet", broadcaster: "charlie", velocity: 1, duration: 30, age: time.Hour},
	})
	words := json.RawMessage(`[{"start":0,"end":1,"text":"no"}]`)
	if stored, err := countdowns.SaveCountdownTranscript(ctx, ids["said-it"], "en", "oh my god that was the cleanest ace ever", words); err != nil || !stored {
		t.Fatalf("store = %v, %v", stored, err)
	}
	// A transcript from the broadcaster's authorization is never replaced.
	if err := clips.RecordClipTranscript(ctx, &models.ClipTranscript{ClipID: ids["authorized"], Language: "de",
		FullText: "authorized words", Segments: json.RawMessage(`[]`), Source: TranscriptSourceAuthorized}); err != nil {
		t.Fatal(err)
	}
	if stored, err := countdowns.SaveCountdownTranscript(ctx, ids["authorized"], "en", "render host words", words); err != nil || stored {
		t.Fatalf("render host replaced an authorized transcript: %v, %v", stored, err)
	}
	if got, err := clips.GetClipTranscript(ctx, ids["authorized"]); err != nil || got.FullText != "authorized words" {
		t.Fatalf("authorized transcript = %+v, %v", got, err)
	}

	day := time.Date(2026, 11, 9, 23, 0, 0, 0, time.UTC)
	for _, u := range []TranscriptionUsage{
		{Day: day, Source: TranscriptSourceCountdown, Transcribed: true, AudioSeconds: 30, ComputeSeconds: 4},
		{Day: day.Add(30 * time.Minute), Source: TranscriptSourceCountdown, AudioSeconds: 20, ComputeSeconds: 2.5},
		{Day: day, Source: TranscriptSourceAuthorized, Transcribed: true, AudioSeconds: 10, ComputeSeconds: 1},
		{Day: day.Add(2 * time.Hour), Source: TranscriptSourceCountdown, AudioSeconds: 5, ComputeSeconds: 1},
	} {
		if err := countdowns.RecordTranscriptionUsage(ctx, u); err != nil {
			t.Fatal(err)
		}
	}
	var clipCount, transcripts int
	var audio, compute float64
	if err := pool.QueryRow(ctx, `SELECT clips, transcripts, audio_seconds, compute_seconds FROM transcription_usage_daily
		WHERE day = '2026-11-09' AND source = 'countdown_render_whisper'`).Scan(&clipCount, &transcripts, &audio, &compute); err != nil {
		t.Fatal(err)
	}
	if clipCount != 2 || transcripts != 1 || audio != 50 || compute != 6.5 {
		t.Fatalf("usage = %d clips, %d transcripts, %v s audio, %v s compute", clipCount, transcripts, audio, compute)
	}
	var days int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM transcription_usage_daily WHERE day = '2026-11-10'`).Scan(&days); err != nil || days != 1 {
		t.Fatalf("usage across midnight UTC: %d rows for 2026-11-10, %v", days, err)
	}

	pending, err := clips.ListTranscriptsToIndex(ctx, 10)
	if err != nil || len(pending) != 2 {
		t.Fatalf("pending = %+v, %v", pending, err)
	}
	if err := clips.MarkTranscriptsIndexed(ctx, pending[:1]); err != nil {
		t.Fatal(err)
	}
	if again, err := clips.ListTranscriptsToIndex(ctx, 10); err != nil || len(again) != 1 || again[0].ClipID != pending[1].ClipID {
		t.Fatalf("pending after marking one = %+v, %v", again, err)
	}
	// A changed transcript is pending again.
	if pending[0].ClipID != ids["said-it"] {
		t.Fatalf("oldest pending = %v", pending[0].ClipID)
	}
	if _, err := countdowns.SaveCountdownTranscript(ctx, ids["said-it"], "en", "oh my god that was the cleanest ace ever!", words); err != nil {
		t.Fatal(err)
	}
	if again, err := clips.ListTranscriptsToIndex(ctx, 10); err != nil || len(again) != 2 {
		t.Fatalf("pending after a change = %+v, %v", again, err)
	}

	known, err := countdowns.ClipsWithTranscripts(ctx, []uuid.UUID{ids["said-it"], ids["authorized"], ids["quiet"]})
	if err != nil || len(known) != 2 || !known[ids["said-it"]] || !known[ids["authorized"]] || known[ids["quiet"]] {
		t.Fatalf("known transcripts = %v, %v", known, err)
	}

	// A full index run requeues the transcripts written since it started.
	all, _ := clips.ListTranscriptsToIndex(ctx, 10)
	if err := clips.MarkTranscriptsIndexed(ctx, all); err != nil {
		t.Fatal(err)
	}
	if err := RequeueTranscriptsSince(ctx, pool, time.Now().Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
	if again, err := clips.ListTranscriptsToIndex(ctx, 10); err != nil || len(again) != 2 {
		t.Fatalf("requeued = %+v, %v", again, err)
	}

	// The database search finds the clip by what was said.
	search := NewSearchRepository(pool)
	said := "cleanest ace"
	res, err := search.Search(ctx, &models.SearchRequest{Type: "clips", Said: &said, Page: 1, Limit: 10, Sort: "relevance"})
	if err != nil || len(res.Results.Clips) != 1 || res.Results.Clips[0].ID != ids["said-it"] {
		t.Fatalf("said search = %+v, %v", res, err)
	}
	snippet := res.Results.Clips[0].TranscriptSnippet
	if snippet == nil || !strings.Contains(*snippet, "cleanest ace") {
		t.Fatalf("snippet = %v", snippet)
	}
	other := "ace cleanest"
	if res, err := search.Search(ctx, &models.SearchRequest{Type: "clips", Said: &other, Page: 1, Limit: 10}); err != nil || len(res.Results.Clips) != 0 {
		t.Fatalf("out-of-order phrase matched: %+v, %v", res, err)
	}
	if res, err := search.Search(ctx, &models.SearchRequest{Type: "clips", Query: "said", Page: 1, Limit: 10}); err != nil || res.Results.Clips == nil {
		t.Fatalf("search without said = %+v, %v", res, err)
	}
}
