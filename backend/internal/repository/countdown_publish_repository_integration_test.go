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

func TestCountdownPublishClaimRecordAndFinish(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "countdown_series", "clips", "users")
	ctx := context.Background()
	repo := NewCountdownRepository(pool)

	windowEnd := time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)
	ids := insertCountdownClips(t, pool, windowEnd, []countdownClip{
		{name: "a1", broadcaster: "alpha", velocity: 900, duration: 30, age: time.Hour},
	})
	series, err := repo.CreateSeries(ctx, &models.CountdownSeriesInput{Slug: "overall", Name: "Overall", Size: 1,
		WindowHours: 24, PerBroadcasterCap: 1, SelectionHourUTC: 12, Enabled: true, ShortsCount: 1})
	if err != nil {
		t.Fatal(err)
	}
	edition, err := repo.SelectEdition(ctx, series, "2026-11-07", windowEnd.Add(-24*time.Hour), windowEnd, false)
	if err != nil {
		t.Fatal(err)
	}
	if e, _, err := repo.ClaimEditionForPublish(ctx, time.Hour); err != nil || e != nil {
		t.Fatalf("a selected edition was claimed for publishing: %+v, %v", e, err)
	}
	if _, err := pool.Exec(ctx, `UPDATE countdown_editions SET status = 'rendered' WHERE id = $1`, edition.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO countdown_edition_media (edition_id, clip_id, source, file_name, sha256, bytes, duration_seconds, width, height, deleted_at)
		VALUES ($1, $2, 'player', 'x/a1.mp4', $3, 5, 29.5, 1920, 1080, NOW())`, edition.ID, ids["a1"], strings.Repeat("a", 64)); err != nil {
		t.Fatal(err)
	}
	var countdownID, shortID, goneID string
	for _, row := range []struct {
		kind, name string
		clip       any
		deleted    bool
		id         *string
	}{
		{"countdown", "countdown.mp4", nil, false, &countdownID},
		{"short", "short-01.mp4", ids["a1"], false, &shortID},
		{"thumbnail", "thumbnail.jpg", nil, true, &goneID},
	} {
		if err := pool.QueryRow(ctx, `
			INSERT INTO countdown_renders (edition_id, kind, clip_id, file_name, sha256, bytes, deleted_at)
			VALUES ($1, $2, $3, $4, $5, 9, CASE WHEN $6 THEN NOW() END) RETURNING id`,
			edition.ID, row.kind, row.clip, edition.ID.String()+"/"+row.name, strings.Repeat("c", 64), row.deleted).Scan(row.id); err != nil {
			t.Fatal(err)
		}
	}

	claimed, claimedSeries, err := repo.ClaimEditionForPublish(ctx, time.Hour)
	if err != nil || claimed == nil || claimed.ID != edition.ID || claimedSeries.ID != series.ID || len(claimed.Clips) != 1 {
		t.Fatalf("claim = %+v, %v", claimed, err)
	}
	renders, err := repo.ListEditionRenders(ctx, edition.ID)
	if err != nil || len(renders) != 2 {
		t.Fatalf("renders = %+v, %v (deleted files must be left out)", renders, err)
	}
	durations, err := repo.ClipDurations(ctx, edition.ID)
	if err != nil || durations[ids["a1"]] != 29.5 {
		t.Fatalf("durations = %v, %v (kept after the file is deleted)", durations, err)
	}

	at := time.Date(2026, 11, 7, 15, 0, 0, 0, time.UTC)
	for _, r := range renders {
		u := &models.CountdownUpload{RenderID: r.ID, EditionID: edition.ID, Kind: r.Kind, ClipID: r.ClipID,
			YouTubeVideoID: "vid" + r.Kind[:4] + "0001", PrivacyStatus: "private", PublishAt: &at}
		if err := repo.RecordUpload(ctx, u); err != nil {
			t.Fatal(err)
		}
		// Recording the same render again keeps the first video ID.
		u.YouTubeVideoID = "otherid0000"
		if err := repo.RecordUpload(ctx, u); err != nil {
			t.Fatal(err)
		}
	}
	uploads, err := repo.ListUploads(ctx, edition.ID)
	if err != nil || len(uploads) != 2 {
		t.Fatalf("uploads = %+v, %v", uploads, err)
	}
	for _, u := range uploads {
		if !strings.HasPrefix(u.YouTubeVideoID, "vid") || u.PublishAt == nil || !u.PublishAt.Equal(at) {
			t.Fatalf("upload = %+v", u)
		}
	}
	bad := &models.CountdownUpload{RenderID: renders[0].ID, EditionID: edition.ID, Kind: "countdown", YouTubeVideoID: "bad id!", PrivacyStatus: "private"}
	if _, err := pool.Exec(ctx, `DELETE FROM countdown_uploads WHERE render_id = $1`, bad.RenderID); err != nil {
		t.Fatal(err)
	}
	if err := repo.RecordUpload(ctx, bad); err == nil {
		t.Fatal("a malformed video ID was stored")
	}

	if id, err := repo.SeriesPlaylist(ctx, series.ID); err != nil || id != "" {
		t.Fatalf("playlist before = %q, %v", id, err)
	}
	if err := repo.SaveSeriesPlaylist(ctx, series.ID, "PLabcdefghijkl"); err != nil {
		t.Fatal(err)
	}
	if id, err := repo.SeriesPlaylist(ctx, series.ID); err != nil || id != "PLabcdefghijkl" {
		t.Fatalf("playlist = %q, %v", id, err)
	}

	if err := repo.FinishPublish(ctx, edition.ID, models.CountdownEditionPublished, nil); err != nil {
		t.Fatal(err)
	}
	progress, err := repo.ListSeriesProgress(ctx)
	if err != nil || len(progress) != 1 || progress[0].Slug != "overall" || progress[0].SelectionHourUTC != 12 ||
		progress[0].LatestRendered == nil || *progress[0].LatestRendered != "2026-11-07" ||
		progress[0].LatestPublished == nil || *progress[0].LatestPublished != "2026-11-07" {
		t.Fatalf("progress = %+v, %v", progress, err)
	}
	if err := repo.FinishPublish(ctx, edition.ID, models.CountdownEditionPublished, nil); !errors.Is(err, ErrCountdownEditionNotPublishing) {
		t.Fatalf("second finish = %v", err)
	}
	if err := repo.FinishPublish(ctx, edition.ID, "selected", nil); err == nil {
		t.Fatal("an invalid outcome was accepted")
	}
}
