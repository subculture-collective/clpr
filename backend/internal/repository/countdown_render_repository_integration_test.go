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

func TestCountdownRenderClaimRecordAndFinish(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "countdown_series", "clips", "users")
	ctx := context.Background()
	repo := NewCountdownRepository(pool)

	day := "2026-11-07"
	windowEnd := time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)
	ids := insertCountdownClips(t, pool, windowEnd, []countdownClip{
		{name: "a1", broadcaster: "alpha", velocity: 900, duration: 30, age: time.Hour},
		{name: "b1", broadcaster: "bravo", velocity: 800, duration: 30, age: time.Hour},
	})
	series, err := repo.CreateSeries(ctx, &models.CountdownSeriesInput{Slug: "overall", Name: "Overall", Size: 2,
		WindowHours: 24, PerBroadcasterCap: 1, SelectionHourUTC: 12, Enabled: true, ShortsCount: 1})
	if err != nil || series.ShortsCount != 1 {
		t.Fatalf("series = %+v, %v", series, err)
	}
	if _, err := repo.SelectEdition(ctx, series, day, windowEnd.Add(-24*time.Hour), windowEnd, false); err != nil {
		t.Fatal(err)
	}
	if edition, _, _, err := repo.ClaimEditionForRender(ctx, time.Hour); err != nil || edition != nil {
		t.Fatalf("a selected edition was claimed for rendering: %+v, %v", edition, err)
	}

	edition, _, err := repo.ClaimEditionForFetch(ctx, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"a1", "b1"} {
		if err := repo.RecordMedia(ctx, &models.CountdownEditionMedia{EditionID: edition.ID, ClipID: ids[name], Source: "player",
			FileName: edition.ID.String() + "/" + name + ".mp4", SHA256: strings.Repeat("b", 64), Bytes: 5, DurationSeconds: 30,
			Width: 1920, Height: 1080}); err != nil {
			t.Fatal(err)
		}
	}
	if err := repo.MarkMediaDeleted(ctx, edition.ID, ids["b1"]); err != nil {
		t.Fatal(err)
	}
	if err := repo.FinishFetch(ctx, edition.ID, models.CountdownEditionFetched, nil); err != nil {
		t.Fatal(err)
	}

	claimed, claimedSeries, media, err := repo.ClaimEditionForRender(ctx, time.Hour)
	if err != nil || claimed == nil || claimed.ID != edition.ID || claimedSeries.ShortsCount != 1 {
		t.Fatalf("render claim = %+v, %v", claimed, err)
	}
	if len(media) != 1 || media[0].ClipID != ids["a1"] {
		t.Fatalf("media = %+v (deleted files must be left out)", media)
	}

	duration := 70.0
	short := ids["a1"]
	outputs := []*models.CountdownRender{
		{EditionID: edition.ID, Kind: "countdown", FileName: edition.ID.String() + "/countdown.mp4", SHA256: strings.Repeat("c", 64), Bytes: 9, DurationSeconds: &duration},
		{EditionID: edition.ID, Kind: "thumbnail", FileName: edition.ID.String() + "/thumbnail.jpg", SHA256: strings.Repeat("d", 64), Bytes: 3},
		{EditionID: edition.ID, Kind: "short", ClipID: &short, FileName: edition.ID.String() + "/short-01.mp4", SHA256: strings.Repeat("e", 64), Bytes: 4, DurationSeconds: &duration},
	}
	for _, x := range outputs {
		if err := repo.RecordRender(ctx, x); err != nil {
			t.Fatalf("record %s: %v", x.Kind, err)
		}
	}
	// Rendering again replaces the row instead of adding one, also for the
	// kinds without a clip.
	again := *outputs[0]
	again.SHA256 = strings.Repeat("f", 64)
	if err := repo.RecordRender(ctx, &again); err != nil {
		t.Fatal(err)
	}
	var rows int
	var digest string
	if err := pool.QueryRow(ctx, `SELECT count(*), max(sha256) FILTER (WHERE kind = 'countdown') FROM countdown_renders WHERE edition_id = $1`, edition.ID).Scan(&rows, &digest); err != nil {
		t.Fatal(err)
	}
	if rows != 3 || digest != strings.Repeat("f", 64) {
		t.Fatalf("renders = %d, countdown sha %s", rows, digest)
	}
	bad := &models.CountdownRender{EditionID: edition.ID, Kind: "short", FileName: "x.mp4", SHA256: strings.Repeat("a", 64), Bytes: 1}
	if err := repo.RecordRender(ctx, bad); err == nil {
		t.Fatal("a Short without a clip was stored")
	}

	if err := repo.FinishRender(ctx, edition.ID, models.CountdownEditionRendered, nil); err != nil {
		t.Fatal(err)
	}
	if err := repo.RecordRender(ctx, outputs[1]); !errors.Is(err, ErrCountdownEditionNotRendering) {
		t.Fatalf("record after finish err = %v", err)
	}
	if err := repo.FinishRender(ctx, edition.ID, models.CountdownEditionFailed, nil); !errors.Is(err, ErrCountdownEditionNotRendering) {
		t.Fatalf("finish twice err = %v", err)
	}

	if expired, err := repo.ListRendersToDelete(ctx, time.Now().Add(-time.Hour)); err != nil || len(expired) != 0 {
		t.Fatalf("fresh renders listed for deletion: %+v, %v", expired, err)
	}
	expired, err := repo.ListRendersToDelete(ctx, time.Now().Add(time.Hour))
	if err != nil || len(expired) != 3 {
		t.Fatalf("expired renders = %+v, %v", expired, err)
	}
	for _, x := range expired {
		if err := repo.MarkRenderDeleted(ctx, x.ID); err != nil {
			t.Fatal(err)
		}
	}
	if left, _ := repo.ListRendersToDelete(ctx, time.Now().Add(time.Hour)); len(left) != 0 {
		t.Fatalf("deleted renders listed again: %+v", left)
	}
}
