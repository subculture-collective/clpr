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

func TestClipRoomFeedsCountdownAndPack(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "countdown_series", "countdown_exclusions", "clip_packs", "broadcaster_feature_optins",
		"streamer_clip_rooms", "clips", "users")
	ctx := context.Background()
	repo := NewCountdownRepository(pool)

	windowEnd := time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)
	ids := insertCountdownClips(t, pool, windowEnd, []countdownClip{
		// Old clips are fine: a room series looks at when the streamer approved them.
		{name: "own-1", broadcaster: "alpha", velocity: 1, duration: 30, age: 10 * 24 * time.Hour},
		{name: "own-2", broadcaster: "alpha", velocity: 900, duration: 30, age: time.Hour},
		{name: "viewer", broadcaster: "bravo", velocity: 0, duration: 30, age: 2 * time.Hour},
		{name: "too-long", broadcaster: "bravo", velocity: 5, duration: 120, age: time.Hour},
		{name: "rejected", broadcaster: "bravo", velocity: 5, duration: 30, age: time.Hour},
		{name: "old-approval", broadcaster: "bravo", velocity: 6, duration: 30, age: time.Hour},
	})
	owner := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO users (id, twitch_id, username, display_name, role, account_type)
		VALUES ($1, 'alpha', 'alpha', 'Alpha', 'user', 'member')`, owner); err != nil {
		t.Fatal(err)
	}
	var roomID uuid.UUID
	if err := pool.QueryRow(ctx, `INSERT INTO streamer_clip_rooms (owner_user_id, twitch_channel) VALUES ($1, 'alpha') RETURNING id`,
		owner).Scan(&roomID); err != nil {
		t.Fatal(err)
	}
	approved := windowEnd.Add(-3 * time.Hour)
	for _, item := range []struct {
		clip     string
		status   string
		position any
		at       time.Time
	}{
		{"viewer", "approved", 1, approved},
		{"own-1", "approved", 2, approved},
		{"too-long", "approved", 3, approved},
		{"own-2", "approved", 4, approved},
		{"rejected", "rejected", nil, approved},
		{"old-approval", "approved", 5, windowEnd.Add(-48 * time.Hour)},
	} {
		if _, err := pool.Exec(ctx, `
			INSERT INTO streamer_clip_room_items (room_id, clip_id, source_url, source_type, status, position, approved_at)
			VALUES ($1, $2, 'https://clips.twitch.tv/x', 'twitch', $3, $4, $5)`, roomID, ids[item.clip], item.status, item.position, item.at); err != nil {
			t.Fatal(err)
		}
	}

	series, err := repo.CreateSeries(ctx, &models.CountdownSeriesInput{Slug: "alpha-room", Name: "Alpha's room", Size: 5, MinSize: intPtrRoom(1),
		WindowHours: 24, PerBroadcasterCap: 1, SelectionHourUTC: 12, Enabled: true, Source: models.CountdownSourceClipRoom, ClipRoomID: &roomID})
	if err != nil {
		t.Fatal(err)
	}
	if series.Source != models.CountdownSourceClipRoom || series.ClipRoomChannel == nil || *series.ClipRoomChannel != "alpha" {
		t.Fatalf("series = %+v", series)
	}
	if listed, err := repo.ListSeries(ctx); err != nil || len(listed) != 1 || listed[0].ClipRoomID == nil || *listed[0].ClipRoomID != roomID {
		t.Fatalf("listed = %+v, %v", listed, err)
	}
	missing := uuid.New()
	if _, err := repo.CreateSeries(ctx, &models.CountdownSeriesInput{Slug: "ghost", Name: "Ghost", Size: 1, WindowHours: 24,
		PerBroadcasterCap: 1, Source: models.CountdownSourceClipRoom, ClipRoomID: &missing}); err != ErrCountdownClipRoomNotFound {
		t.Fatalf("unknown room = %v", err)
	}

	edition, err := repo.SelectEdition(ctx, series, "2026-11-07", windowEnd.Add(-24*time.Hour), windowEnd, false)
	if err != nil {
		t.Fatal(err)
	}
	// The room's order, without the clip that is too long, the rejected one
	// and the one approved before the window; two clips of alpha despite the cap.
	if got := strings.Join(editionClipNames(edition), ","); edition.Status != models.CountdownEditionSelected || got != "viewer,own-1,own-2" {
		t.Fatalf("edition %s clips = %s", edition.Status, got)
	}

	// A velocity series the same day does not get the room's clips again.
	overall, err := repo.CreateSeries(ctx, &models.CountdownSeriesInput{Slug: "overall", Name: "Overall", Size: 1,
		WindowHours: 24, PerBroadcasterCap: 1, SelectionHourUTC: 12, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	if e, err := repo.SelectEdition(ctx, overall, "2026-11-07", windowEnd.Add(-24*time.Hour), windowEnd, false); err != nil ||
		len(e.Clips) != 1 || e.Clips[0].Title != "old-approval" {
		t.Fatalf("overall = %+v, %v", e, err)
	}

	// A series whose room is gone fails with the reason.
	if _, err := pool.Exec(ctx, `DELETE FROM streamer_clip_rooms WHERE id = $1`, roomID); err != nil {
		t.Fatal(err)
	}
	series, err = repo.GetSeries(ctx, series.ID)
	if err != nil || series.ClipRoomID != nil || series.Source != models.CountdownSourceClipRoom {
		t.Fatalf("series after the room was deleted = %+v, %v", series, err)
	}
	gone, err := repo.SelectEdition(ctx, series, "2026-11-08", windowEnd, windowEnd.Add(24*time.Hour), false)
	if err != nil || gone.Status != models.CountdownEditionFailed || gone.FailureReason == nil ||
		!strings.Contains(*gone.FailureReason, "no longer exists") {
		t.Fatalf("edition without a room = %+v, %v", gone, err)
	}
}

func TestClipRoomOrderLeadsThePack(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "clip_packs", "countdown_exclusions", "broadcaster_feature_optins", "streamer_clip_rooms", "clips", "users")
	ctx := context.Background()
	repo := NewCountdownRepository(pool)

	weekStart := time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)
	weekEnd := weekStart.AddDate(0, 0, 7)
	ids := insertCountdownClips(t, pool, weekEnd, []countdownClip{
		{name: "popular", broadcaster: "alpha", velocity: 1, duration: 30, age: 24 * time.Hour},
		{name: "picked", broadcaster: "alpha", velocity: 1, duration: 30, age: 48 * time.Hour},
		{name: "older-picked", broadcaster: "alpha", velocity: 1, duration: 30, age: 30 * 24 * time.Hour},
		{name: "someone-else", broadcaster: "bravo", velocity: 1, duration: 30, age: time.Hour},
	})
	if _, err := pool.Exec(ctx, `UPDATE clips SET view_count = 9000 WHERE id = $1`, ids["popular"]); err != nil {
		t.Fatal(err)
	}
	owner := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO users (id, twitch_id, username, display_name, role, account_type)
		VALUES ($1, 'alpha', 'alpha', 'Alpha', 'user', 'member')`, owner); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO broadcaster_feature_optins (broadcaster_id, user_id, scope, consent_text_version)
		VALUES ('alpha', $1, 'channel:manage:clips', '2026-11-07')`, owner); err != nil {
		t.Fatal(err)
	}
	var roomID uuid.UUID
	if err := pool.QueryRow(ctx, `INSERT INTO streamer_clip_rooms (owner_user_id, twitch_channel) VALUES ($1, 'alpha') RETURNING id`,
		owner).Scan(&roomID); err != nil {
		t.Fatal(err)
	}
	for _, item := range []struct {
		clip     string
		position int
	}{{"older-picked", 1}, {"picked", 2}, {"someone-else", 3}} {
		if _, err := pool.Exec(ctx, `
			INSERT INTO streamer_clip_room_items (room_id, clip_id, source_url, source_type, status, position, approved_at)
			VALUES ($1, $2, 'https://clips.twitch.tv/x', 'twitch', 'approved', $3, $4)`,
			roomID, ids[item.clip], item.position, weekStart.Add(72*time.Hour)); err != nil {
			t.Fatal(err)
		}
	}
	if created, err := repo.EnsureWeeklyPacks(ctx, weekStart, 10); err != nil || created != 1 {
		t.Fatalf("created = %d, %v", created, err)
	}
	pack, err := repo.ClaimPackForBuild(ctx, time.Hour)
	if err != nil || pack == nil {
		t.Fatalf("pack = %+v, %v", pack, err)
	}
	var got []string
	for _, c := range pack.Clips {
		got = append(got, c.Title)
	}
	// Room order first (an older clip the streamer picked this week counts),
	// then views; another broadcaster's clip from the room is not theirs to pack.
	if strings.Join(got, ",") != "older-picked,picked,popular" {
		t.Fatalf("pack clips = %v", got)
	}
}

func intPtrRoom(n int) *int { return &n }
