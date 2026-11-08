//go:build integration

package repository

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
)

func TestCountdownMediaClaimBackfillAndFinish(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "countdown_series", "clips", "users")
	ctx := context.Background()
	repo := NewCountdownRepository(pool)

	day := "2026-11-07"
	windowEnd := time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)
	windowStart := windowEnd.Add(-24 * time.Hour)
	h := time.Hour
	ids := insertCountdownClips(t, pool, windowEnd, []countdownClip{
		{name: "a1", broadcaster: "alpha", velocity: 900, duration: 30, age: h},
		{name: "b1", broadcaster: "bravo", velocity: 800, duration: 30, age: h},
		{name: "c1", broadcaster: "charlie", velocity: 700, duration: 30, age: h},
		// Next in line: bravo's second clip would break the cap of 1, so
		// the backfill must take delta instead.
		{name: "b2", broadcaster: "bravo", velocity: 650, duration: 30, age: h},
		{name: "d1", broadcaster: "delta", velocity: 600, duration: 30, age: h},
		{name: "e1", broadcaster: "echo", velocity: 500, duration: 30, age: h},
	})
	series, err := repo.CreateSeries(ctx, &models.CountdownSeriesInput{Slug: "overall", Name: "Overall", Size: 3,
		WindowHours: 24, PerBroadcasterCap: 1, SelectionHourUTC: 12, Priority: 1, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.SelectEdition(ctx, series, day, windowStart, windowEnd, false); err != nil {
		t.Fatal(err)
	}

	edition, claimedSeries, err := repo.ClaimEditionForFetch(ctx, time.Hour)
	if err != nil || edition == nil || claimedSeries.ID != series.ID {
		t.Fatalf("claim = %+v, %v", edition, err)
	}
	if edition.Clips[0].TwitchClipID == "" || !strings.HasPrefix(edition.Clips[0].TwitchClipURL, "https://clips.twitch.tv/") {
		t.Fatalf("claimed clip lacks Twitch fields: %+v", edition.Clips[0])
	}
	if again, _, err := repo.ClaimEditionForFetch(ctx, time.Hour); err != nil || again != nil {
		t.Fatalf("second claim = %+v, %v", again, err)
	}

	// c1 (position 3) fails. Bravo still has b1 in the edition, so b2 would
	// break the cap of 1 and delta is appended instead.
	next, err := repo.BackfillClip(ctx, edition, series, ids["c1"], "HTTP 404")
	if err != nil || next == nil || next.Title != "d1" || next.Position != 3 || next.TwitchClipID == "" {
		t.Fatalf("backfill = %+v, %v", next, err)
	}
	listed, err := repo.ListEditions(ctx, &day, nil, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := editionClipNames(&listed[0]), []string{"a1", "b1", "d1"}; !equalStrings(got, want) {
		t.Fatalf("after backfill = %v, want %v", got, want)
	}

	// a1 (position 1) fails: the others move up one place and echo is appended.
	next, err = repo.BackfillClip(ctx, edition, series, ids["a1"], "HTTP 404")
	if err != nil || next == nil || next.Title != "e1" || next.Position != 3 {
		t.Fatalf("second backfill = %+v, %v", next, err)
	}
	listed, err = repo.ListEditions(ctx, &day, nil, 1)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := editionClipNames(&listed[0]), []string{"b1", "d1", "e1"}; !equalStrings(got, want) {
		t.Fatalf("after second backfill = %v, want %v", got, want)
	}
	for i, c := range listed[0].Clips {
		if c.Position != i+1 {
			t.Fatalf("positions = %+v", listed[0].Clips)
		}
	}

	// A skipped clip stays out of other editions that day.
	other, err := repo.CreateSeries(ctx, &models.CountdownSeriesInput{Slug: "second", Name: "Second", Size: 1,
		WindowHours: 24, PerBroadcasterCap: 2, SelectionHourUTC: 12, Priority: 2, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	second, err := repo.SelectEdition(ctx, other, day, windowStart, windowEnd, false)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := editionClipNames(second), []string{"b2"}; !equalStrings(got, want) {
		t.Fatalf("second series = %v, want %v (a1 and c1 were skipped today)", got, want)
	}

	media := &models.CountdownEditionMedia{EditionID: edition.ID, ClipID: ids["b1"], Source: "player",
		FileName: edition.ID.String() + "/b1.mp4", SHA256: strings.Repeat("a", 64), Bytes: 10, DurationSeconds: 30, Width: 1920, Height: 1080}
	if err := repo.RecordMedia(ctx, media); err != nil {
		t.Fatal(err)
	}
	if err := repo.RecordMedia(ctx, media); err != nil {
		t.Fatalf("recording the same clip again: %v", err)
	}

	// Every other clip is taken or skipped today: no backfill.
	none, err := repo.BackfillClip(ctx, edition, series, ids["d1"], "HTTP 404")
	if err != nil || none != nil {
		t.Fatalf("exhausted backfill = %+v, %v", none, err)
	}

	if err := repo.FinishFetch(ctx, edition.ID, models.CountdownEditionFailed, strPtr("not enough eligible clips after 3 skips")); err != nil {
		t.Fatal(err)
	}
	if err := repo.FinishFetch(ctx, edition.ID, models.CountdownEditionFetched, nil); !errors.Is(err, ErrCountdownEditionNotFetching) {
		t.Fatalf("finish twice err = %v", err)
	}
	if err := repo.RecordMedia(ctx, media); !errors.Is(err, ErrCountdownEditionNotFetching) {
		t.Fatalf("record after finish err = %v", err)
	}

	// A failed edition's files are listed for deletion right away.
	toDelete, err := repo.ListMediaToDelete(ctx, time.Now().Add(-24*time.Hour))
	if err != nil || len(toDelete) != 1 || toDelete[0].ClipID != ids["b1"] {
		t.Fatalf("to delete = %+v, %v", toDelete, err)
	}
	if err := repo.MarkMediaDeleted(ctx, edition.ID, ids["b1"]); err != nil {
		t.Fatal(err)
	}
	if toDelete, _ = repo.ListMediaToDelete(ctx, time.Now()); len(toDelete) != 0 {
		t.Fatalf("deleted media listed again: %+v", toDelete)
	}

	// A claim left in fetching past staleAfter is taken again.
	if _, err := pool.Exec(ctx, `UPDATE countdown_editions SET status = 'fetching', updated_at = NOW() - INTERVAL '2 hours' WHERE id = $1`, second.ID); err != nil {
		t.Fatal(err)
	}
	stale, _, err := repo.ClaimEditionForFetch(ctx, time.Hour)
	if err != nil || stale == nil || stale.ID != second.ID {
		t.Fatalf("stale claim = %+v, %v", stale, err)
	}
}
