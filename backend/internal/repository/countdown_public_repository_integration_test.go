//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
)

func TestCountdownPublicPages(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "countdown_series", "countdown_exclusions", "clips", "users")
	ctx := context.Background()
	repo := NewCountdownRepository(pool)

	day1, day2 := "2026-11-06", "2026-11-07"
	end1 := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	end2 := end1.Add(24 * time.Hour)
	ids := insertCountdownClips(t, pool, end1, []countdownClip{
		{name: "a1", broadcaster: "alpha", velocity: 900, duration: 30, age: time.Hour},
		{name: "b1", broadcaster: "bravo", velocity: 800, duration: 30, age: time.Hour},
		{name: "c1", broadcaster: "charlie", velocity: 700, duration: 30, age: time.Hour},
	})
	for k, v := range insertCountdownClips(t, pool, end2, []countdownClip{
		{name: "a2", broadcaster: "alpha", velocity: 900, duration: 30, age: time.Hour},
		{name: "d2", broadcaster: "delta", velocity: 600, duration: 30, age: time.Hour},
	}) {
		ids[k] = v
	}
	overall, err := repo.CreateSeries(ctx, &models.CountdownSeriesInput{Slug: "overall", Name: "Overall", Size: 2,
		WindowHours: 24, PerBroadcasterCap: 1, SelectionHourUTC: 12, Priority: 0, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	late, err := repo.CreateSeries(ctx, &models.CountdownSeriesInput{Slug: "late", Name: "Late", Size: 1,
		WindowHours: 24, PerBroadcasterCap: 1, SelectionHourUTC: 12, Priority: 1, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	setStatus := func(e *models.CountdownEdition, status string) {
		if _, err := pool.Exec(ctx, `UPDATE countdown_editions SET status = $2 WHERE id = $1`, e.ID, status); err != nil {
			t.Fatal(err)
		}
	}
	e1, err := repo.SelectEdition(ctx, overall, day1, end1.Add(-24*time.Hour), end1, false)
	if err != nil {
		t.Fatal(err)
	}
	setStatus(e1, models.CountdownEditionRendered)
	if _, err := repo.SelectEdition(ctx, late, day1, end1.Add(-24*time.Hour), end1, false); err != nil {
		t.Fatal(err) // stays "selected", so it is not public
	}
	e2, err := repo.SelectEdition(ctx, overall, day2, end2.Add(-24*time.Hour), end2, false)
	if err != nil {
		t.Fatal(err)
	}
	setStatus(e2, models.CountdownEditionPublished)
	bravo := "bravo"
	if _, err := repo.CreateExclusion(ctx, &models.CountdownExclusionInput{BroadcasterID: &bravo, Reason: "request"}, "admin", nil); err != nil {
		t.Fatal(err)
	}

	page, next, err := repo.ListPublicEditions(ctx, nil, nil, 1)
	if err != nil || len(page) != 1 || page[0].EditionDate != day2 || next == nil || *next != day2 {
		t.Fatalf("first page = %+v, next %v, %v", page, next, err)
	}
	page, next, err = repo.ListPublicEditions(ctx, nil, next, 14)
	if err != nil || len(page) != 1 || next != nil {
		t.Fatalf("second page = %+v, next %v, %v", page, next, err)
	}
	got := page[0]
	if got.EditionDate != day1 || got.SeriesSlug != "overall" || got.SeriesName != "Overall" || got.Size != 2 {
		t.Fatalf("edition = %+v", got)
	}
	// bravo was excluded after the edition was made, so b1 drops out.
	if len(got.Clips) != 1 || got.Clips[0].ClipID != ids["a1"] || got.Clips[0].Position != 1 {
		t.Fatalf("clips = %+v", got.Clips)
	}
	if page, _, err := repo.ListPublicEditions(ctx, &late.Slug, nil, 14); err != nil || len(page) != 0 {
		t.Fatalf("late series pages = %+v, %v", page, err)
	}

	if e, err := repo.GetPublicEdition(ctx, day2, "overall"); err != nil || e == nil || len(e.Clips) != 2 {
		t.Fatalf("day 2 = %+v, %v", e, err)
	}
	if e, err := repo.GetPublicEdition(ctx, day1, "late"); err != nil || e != nil {
		t.Fatalf("a selected edition is public: %+v, %v", e, err)
	}

	features, err := repo.ListCountdownFeatures(ctx, nil, strPtrCountdown("alpha"), 20)
	if err != nil || len(features) != 2 || features[0].EditionDate != day2 || features[0].ClipID != ids["a2"] || features[1].ClipID != ids["a1"] {
		t.Fatalf("alpha features = %+v, %v", features, err)
	}
	clip := ids["b1"]
	if features, err := repo.ListCountdownFeatures(ctx, &clip, nil, 20); err != nil || len(features) != 0 {
		t.Fatalf("excluded clip features = %+v, %v", features, err)
	}

	keys, err := repo.ListPublicEditionKeys(ctx, 100)
	if err != nil || len(keys) != 2 || keys[0] != (models.CountdownEditionKey{EditionDate: day2, SeriesSlug: "overall"}) {
		t.Fatalf("keys = %+v, %v", keys, err)
	}
}

func strPtrCountdown(s string) *string { return &s }
