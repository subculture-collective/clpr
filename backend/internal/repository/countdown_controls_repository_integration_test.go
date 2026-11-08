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
	"github.com/google/uuid"
)

func TestCountdownBroadcasterControls(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "countdown_series", "countdown_exclusions", "broadcaster_feature_optins", "clips", "users")
	ctx := context.Background()
	repo := NewCountdownRepository(pool)

	userID := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO users (id, twitch_id, username, display_name, role, account_type)
		VALUES ($1, 'alpha', 'alpha', 'Alpha', 'user', 'member')`, userID); err != nil {
		t.Fatal(err)
	}
	windowEnd := time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)
	ids := insertCountdownClips(t, pool, windowEnd, []countdownClip{
		{name: "a1", broadcaster: "alpha", velocity: 900, duration: 30, age: time.Hour},
		{name: "b1", broadcaster: "bravo", velocity: 800, duration: 30, age: time.Hour},
	})
	series, err := repo.CreateSeries(ctx, &models.CountdownSeriesInput{Slug: "overall", Name: "Overall", Size: 2,
		WindowHours: 24, PerBroadcasterCap: 1, SelectionHourUTC: 12, Enabled: true, ShortsCount: 1})
	if err != nil {
		t.Fatal(err)
	}
	edition, err := repo.SelectEdition(ctx, series, "2026-11-07", windowEnd.Add(-24*time.Hour), windowEnd, false)
	if err != nil || len(edition.Clips) != 2 {
		t.Fatalf("edition = %+v, %v", edition, err)
	}
	if _, err := pool.Exec(ctx, `UPDATE countdown_editions SET status = 'published' WHERE id = $1`, edition.ID); err != nil {
		t.Fatal(err)
	}
	// A published countdown with both clips, and a Short of bravo's clip.
	upload := func(kind, name string, clip *uuid.UUID, video string) uuid.UUID {
		var renderID uuid.UUID
		if err := pool.QueryRow(ctx, `INSERT INTO countdown_renders (edition_id, kind, clip_id, file_name, sha256, bytes)
			VALUES ($1, $2, $3, $4, $5, 9) RETURNING id`, edition.ID, kind, clip, edition.ID.String()+"/"+name, strings.Repeat("c", 64)).Scan(&renderID); err != nil {
			t.Fatal(err)
		}
		if err := repo.RecordUpload(ctx, &models.CountdownUpload{RenderID: renderID, EditionID: edition.ID, Kind: kind, ClipID: clip,
			YouTubeVideoID: video, PrivacyStatus: "private"}); err != nil {
			t.Fatal(err)
		}
		return renderID
	}
	bravo := ids["b1"]
	countdownRender := upload("countdown", "countdown.mp4", nil, "countdown01")
	upload("short", "short-02.mp4", &bravo, "shortbravo1")

	state, err := repo.BroadcasterCountdownState(ctx, "alpha")
	if err != nil || state.Excluded || state.Featured || state.ConsentTextVersion != models.CountdownConsentTextVersion {
		t.Fatalf("initial state = %+v, %v", state, err)
	}
	if tasks, err := repo.ListRemovalTasks(ctx); err != nil || len(tasks) != 0 {
		t.Fatalf("tasks before any exclusion = %+v, %v", tasks, err)
	}

	// Leave my clips out: alpha's clip in the published countdown becomes a task.
	if err := repo.SetSelfExclusion(ctx, "alpha", userID, true); err != nil {
		t.Fatal(err)
	}
	if err := repo.SetSelfExclusion(ctx, "alpha", userID, true); err != nil {
		t.Fatalf("turning it on twice: %v", err)
	}
	state, _ = repo.BroadcasterCountdownState(ctx, "alpha")
	if !state.Excluded || state.ExclusionSource == nil || *state.ExclusionSource != "self_service" {
		t.Fatalf("excluded state = %+v", state)
	}
	var version string
	if err := pool.QueryRow(ctx, `SELECT consent_text_version FROM countdown_exclusions WHERE broadcaster_id = 'alpha'`).Scan(&version); err != nil || version != models.CountdownConsentTextVersion {
		t.Fatalf("consent version %q, %v", version, err)
	}
	tasks, err := repo.ListRemovalTasks(ctx)
	if err != nil || len(tasks) != 1 || tasks[0].ClipID != ids["a1"] || tasks[0].YouTubeVideoID != "countdown01" ||
		tasks[0].ExclusionSource != "self_service" || !tasks[0].DueAt.Equal(tasks[0].ExcludedAt.Add(7*24*time.Hour)) {
		t.Fatalf("tasks = %+v, %v", tasks, err)
	}

	// An admin excludes bravo, who has no account: the countdown and the Short.
	bravoID := "bravo"
	if _, err := repo.CreateExclusion(ctx, &models.CountdownExclusionInput{BroadcasterID: &bravoID, Reason: "request"}, "admin", nil); err != nil {
		t.Fatal(err)
	}
	tasks, _ = repo.ListRemovalTasks(ctx)
	if len(tasks) != 3 {
		t.Fatalf("tasks after bravo's exclusion = %+v", tasks)
	}
	// A broadcaster turning their toggle off does not lift an admin exclusion.
	if err := repo.SetSelfExclusion(ctx, "bravo", userID, false); err != nil {
		t.Fatal(err)
	}
	if s, _ := repo.BroadcasterCountdownState(ctx, "bravo"); !s.Excluded || *s.ExclusionSource != "admin" {
		t.Fatalf("bravo state = %+v", s)
	}

	note := "cut from the video in Studio"
	if err := repo.ResolveRemoval(ctx, &models.CountdownRemovalInput{RenderID: countdownRender, ClipID: ids["a1"], Resolution: "edited", Note: &note}, userID); err != nil {
		t.Fatal(err)
	}
	if err := repo.ResolveRemoval(ctx, &models.CountdownRemovalInput{RenderID: uuid.New(), ClipID: ids["a1"], Resolution: "removed"}, userID); !errors.Is(err, ErrCountdownRemovalNotFound) {
		t.Fatalf("resolving an unknown render = %v", err)
	}
	if tasks, _ = repo.ListRemovalTasks(ctx); len(tasks) != 2 {
		t.Fatalf("tasks after resolving one = %+v", tasks)
	}

	// Turning the toggle off removes alpha's own exclusion.
	if err := repo.SetSelfExclusion(ctx, "alpha", userID, false); err != nil {
		t.Fatal(err)
	}
	if s, _ := repo.BroadcasterCountdownState(ctx, "alpha"); s.Excluded {
		t.Fatalf("alpha still excluded: %+v", s)
	}

	// Feature my clips.
	if err := repo.RecordFeatureOptIn(ctx, "alpha", userID, "channel:manage:clips"); err != nil {
		t.Fatal(err)
	}
	if s, _ := repo.BroadcasterCountdownState(ctx, "alpha"); !s.Featured || s.FeaturedAt == nil {
		t.Fatalf("featured state = %+v", s)
	}
	if err := repo.DeleteFeatureOptIns(ctx, userID); err != nil {
		t.Fatal(err)
	}
	if s, _ := repo.BroadcasterCountdownState(ctx, "alpha"); s.Featured {
		t.Fatalf("opt-in survived revocation: %+v", s)
	}
}
