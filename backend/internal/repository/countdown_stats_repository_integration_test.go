//go:build integration

package repository

import (
	"context"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/google/uuid"
)

func TestCountdownPlacementsAndViewCounts(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "countdown_series", "countdown_exclusions", "clips", "users")
	ctx := context.Background()
	repo := NewCountdownRepository(pool)

	windowEnd := time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC)
	ids := insertCountdownClips(t, pool, windowEnd, []countdownClip{
		{name: "a1", broadcaster: "alpha", velocity: 900, duration: 30, age: time.Hour},
		{name: "b1", broadcaster: "bravo", velocity: 800, duration: 30, age: time.Hour},
	})
	series, err := repo.CreateSeries(ctx, &models.CountdownSeriesInput{Slug: "overall", Name: "Overall", Size: 2,
		WindowHours: 24, PerBroadcasterCap: 1, SelectionHourUTC: 12, Enabled: true, ShortsCount: 1})
	if err != nil {
		t.Fatal(err)
	}
	edition, err := repo.SelectEdition(ctx, series, "2026-11-08", windowEnd.Add(-24*time.Hour), windowEnd, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `UPDATE countdown_editions SET status = 'published' WHERE id = $1`, edition.ID); err != nil {
		t.Fatal(err)
	}
	upload := func(kind string, clip any, video string, publishAt *time.Time) {
		t.Helper()
		var renderID uuid.UUID
		if err := pool.QueryRow(ctx, `
			INSERT INTO countdown_renders (edition_id, kind, clip_id, file_name, sha256, bytes)
			VALUES ($1, $2, $3, $4, $5, 9) RETURNING id`, edition.ID, kind, clip, video+".mp4", strings.Repeat("b", 64)).Scan(&renderID); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, `
			INSERT INTO countdown_uploads (render_id, edition_id, kind, clip_id, youtube_video_id, privacy_status, publish_at)
			VALUES ($1, $2, $3, $4, $5, 'private', $6)`, renderID, edition.ID, kind, clip, video, publishAt); err != nil {
			t.Fatal(err)
		}
	}
	past := time.Now().Add(-time.Hour)
	upload("countdown", nil, "vidCountdown1", &past) // scheduled, now public
	upload("short", ids["a1"], "vidShortA1", nil)    // private

	placements, err := repo.ListBroadcasterPlacements(ctx, "alpha", 50)
	if err != nil || len(placements) != 1 {
		t.Fatalf("placements = %+v, %v", placements, err)
	}
	p := placements[0]
	if p.ClipID != ids["a1"] || p.Position != 1 || p.SeriesSlug != "overall" || p.EditionDate != "2026-11-08" ||
		p.Countdown == nil || p.Countdown.YouTubeVideoID != "vidCountdown1" || p.Countdown.ViewCount != nil || p.Short != nil {
		t.Fatalf("placement before refresh = %+v (countdown %+v, short %+v)", p, p.Countdown, p.Short)
	}

	due, err := repo.ListUploadsForStats(ctx, windowEnd.Add(-48*time.Hour), time.Now(), 10)
	if err != nil || len(due) != 2 {
		t.Fatalf("due = %v, %v", due, err)
	}
	if due, err := repo.ListUploadsForStats(ctx, time.Now(), time.Now(), 10); err != nil || len(due) != 0 {
		t.Fatalf("uploads older than the window are due: %v, %v", due, err)
	}
	now := time.Now()
	if err := repo.SaveUploadStats(ctx, []models.CountdownVideoStats{
		{YouTubeVideoID: "vidCountdown1", ViewCount: 500, PrivacyStatus: "public"},
		{YouTubeVideoID: "vidShortA1", ViewCount: 20, PrivacyStatus: "public"},
	}, nil, now); err != nil {
		t.Fatal(err)
	}
	placements, err = repo.ListBroadcasterPlacements(ctx, "alpha", 50)
	if err != nil || len(placements) != 1 || placements[0].Countdown == nil || *placements[0].Countdown.ViewCount != 500 ||
		placements[0].Short == nil || *placements[0].Short.ViewCount != 20 || placements[0].Short.ViewsUpdatedAt == nil {
		t.Fatalf("placements after refresh = %+v, %v", placements, err)
	}
	if due, err := repo.ListUploadsForStats(ctx, windowEnd.Add(-48*time.Hour), now.Add(-time.Minute), 10); err != nil || len(due) != 0 {
		t.Fatalf("fresh uploads are due: %v, %v", due, err)
	}

	// A Short taken down on YouTube is no longer linked or refreshed.
	if err := repo.SaveUploadStats(ctx, nil, []string{"vidShortA1"}, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	placements, err = repo.ListBroadcasterPlacements(ctx, "alpha", 50)
	if err != nil || len(placements) != 1 || placements[0].Short != nil || placements[0].Countdown == nil {
		t.Fatalf("placements after the Short went missing = %+v, %v", placements, err)
	}
	if due, err := repo.ListUploadsForStats(ctx, windowEnd.Add(-48*time.Hour), now.Add(2*time.Hour), 10); err != nil || len(due) != 1 || due[0] != "vidCountdown1" {
		t.Fatalf("due after missing = %v, %v", due, err)
	}

	// Excluding the broadcaster hides their placements, like the public pages.
	alpha := "alpha"
	if _, err := repo.CreateExclusion(ctx, &models.CountdownExclusionInput{BroadcasterID: &alpha, Reason: "request"}, "admin", nil); err != nil {
		t.Fatal(err)
	}
	if placements, err := repo.ListBroadcasterPlacements(ctx, "alpha", 50); err != nil || len(placements) != 0 {
		t.Fatalf("excluded placements = %+v, %v", placements, err)
	}
	if placements, err := repo.ListBroadcasterPlacements(ctx, "bravo", 50); err != nil || len(placements) != 1 || placements[0].Countdown == nil {
		t.Fatalf("bravo placements = %+v, %v", placements, err)
	}
}
