//go:build integration

package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type countdownClip struct {
	name        string
	broadcaster string
	game        string
	language    string
	velocity    float64
	duration    float64
	age         time.Duration // before the window end
	source      string
	nsfw        bool
	removed     bool
	hidden      bool
	dmca        bool
}

func insertCountdownClips(t *testing.T, pool *pgxpool.Pool, windowEnd time.Time, clips []countdownClip) map[string]uuid.UUID {
	t.Helper()
	ids := map[string]uuid.UUID{}
	for _, c := range clips {
		id := uuid.New()
		source := c.source
		if source == "" {
			source = "twitch"
		}
		_, err := pool.Exec(context.Background(), `
			INSERT INTO clips (id, twitch_clip_id, twitch_clip_url, embed_url, title, creator_name, broadcaster_name,
				broadcaster_id, game_id, language, duration, view_count, view_velocity, created_at, imported_at,
				source_type, is_nsfw, is_removed, is_hidden, dmca_removed)
			VALUES ($1, $2, $3, $3, $4, 'clipper', $5, $5, $6, $7, $8, 100, $9, $10, NOW(), $11, $12, $13, $14, $15)`,
			id, "tw-"+id.String(), "https://clips.twitch.tv/"+c.name, c.name, c.broadcaster, c.game, c.language,
			c.duration, c.velocity, windowEnd.Add(-c.age).UTC(), source, c.nsfw, c.removed, c.hidden, c.dmca)
		if err != nil {
			t.Fatalf("insert clip %s: %v", c.name, err)
		}
		ids[c.name] = id
	}
	return ids
}

func editionClipNames(e *models.CountdownEdition) []string {
	names := make([]string, len(e.Clips))
	for i, c := range e.Clips {
		names[i] = c.Title
	}
	return names
}

func TestCountdownEditionSelection(t *testing.T) {
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
		{name: "a1", broadcaster: "alpha", game: "g1", language: "en", velocity: 900, duration: 30, age: h},
		{name: "a2", broadcaster: "alpha", game: "g1", language: "en", velocity: 800, duration: 30, age: h},
		{name: "b1", broadcaster: "bravo", game: "g1", language: "en", velocity: 700, duration: 30, age: h},
		{name: "c1", broadcaster: "charlie", game: "g2", language: "de", velocity: 600, duration: 30, age: h},
		{name: "d1", broadcaster: "delta", game: "g1", language: "en", velocity: 500, duration: 5, age: 2 * h},
		{name: "e1", broadcaster: "echo", game: "g2", language: "en", velocity: 400, duration: 60, age: 23 * h},
		// Ineligible, each for one reason, all ranked above the eligible clips.
		{name: "too-short", broadcaster: "x1", velocity: 5000, duration: 4.9, age: h},
		{name: "too-long", broadcaster: "x2", velocity: 5000, duration: 60.5, age: h},
		{name: "too-old", broadcaster: "x3", velocity: 5000, duration: 30, age: 25 * h},
		{name: "after-window", broadcaster: "x4", velocity: 5000, duration: 30, age: -h},
		{name: "nsfw", broadcaster: "x5", velocity: 5000, duration: 30, age: h, nsfw: true},
		{name: "removed", broadcaster: "x6", velocity: 5000, duration: 30, age: h, removed: true},
		{name: "hidden", broadcaster: "x7", velocity: 5000, duration: 30, age: h, hidden: true},
		{name: "dmca", broadcaster: "x8", velocity: 5000, duration: 30, age: h, dmca: true},
		{name: "upload", broadcaster: "x9", velocity: 5000, duration: 30, age: h, source: "upload"},
		{name: "no-velocity", broadcaster: "x10", velocity: 0, duration: 30, age: h},
		{name: "excluded-broadcaster", broadcaster: "optout", velocity: 5000, duration: 30, age: h},
		{name: "excluded-clip", broadcaster: "x11", velocity: 5000, duration: 30, age: h},
	})

	optout := "optout"
	if _, err := repo.CreateExclusion(ctx, &models.CountdownExclusionInput{BroadcasterID: &optout, Reason: "request"}, "admin", nil); err != nil {
		t.Fatalf("exclude broadcaster: %v", err)
	}
	clipID := ids["excluded-clip"]
	if _, err := repo.CreateExclusion(ctx, &models.CountdownExclusionInput{ClipID: &clipID, Reason: "takedown"}, "admin", nil); err != nil {
		t.Fatalf("exclude clip: %v", err)
	}
	if _, err := repo.CreateExclusion(ctx, &models.CountdownExclusionInput{BroadcasterID: &optout, Reason: "dmca"}, "admin", nil); !errors.Is(err, ErrCountdownExclusionExists) {
		t.Fatalf("duplicate exclusion err = %v", err)
	}
	missing := uuid.New()
	if _, err := repo.CreateExclusion(ctx, &models.CountdownExclusionInput{ClipID: &missing, Reason: "dmca"}, "admin", nil); !errors.Is(err, ErrCountdownExclusionClipUnset) {
		t.Fatalf("unknown clip exclusion err = %v", err)
	}

	newSeries := func(slug string, size, cap int, game, language *string, priority int) *models.CountdownSeries {
		s, err := repo.CreateSeries(ctx, &models.CountdownSeriesInput{Slug: slug, Name: slug, GameID: game, Language: language,
			Size: size, WindowHours: 24, PerBroadcasterCap: cap, SelectionHourUTC: 12, Priority: priority, Enabled: true})
		if err != nil {
			t.Fatalf("create series %s: %v", slug, err)
		}
		return s
	}
	g1 := "g1"
	overall := newSeries("overall", 4, 1, nil, nil, 10)
	category := newSeries("category-g1", 2, 2, &g1, nil, 20)

	// Overall: one clip per broadcaster, ranked by view velocity.
	edition, err := repo.SelectEdition(ctx, overall, day, windowStart, windowEnd, false)
	if err != nil {
		t.Fatalf("select overall: %v", err)
	}
	if got, want := editionClipNames(edition), []string{"a1", "b1", "c1", "d1"}; !equalStrings(got, want) {
		t.Fatalf("overall clips = %v, want %v", got, want)
	}
	if edition.Status != models.CountdownEditionSelected || edition.EditionDate != day || edition.Clips[0].Position != 1 {
		t.Fatalf("overall edition = %+v", edition)
	}

	// Category g1 with a cap of 2: a1, b1 and d1 already went to the overall
	// edition today, so only a2 is left and the edition cannot be filled.
	edition, err = repo.SelectEdition(ctx, category, day, windowStart, windowEnd, false)
	if err != nil {
		t.Fatalf("select category: %v", err)
	}
	if edition.Status != models.CountdownEditionFailed || len(edition.Clips) != 0 || edition.FailureReason == nil ||
		*edition.FailureReason != "only 1 of 2 clips were eligible" {
		t.Fatalf("category edition = %+v", edition)
	}

	if _, err := repo.SelectEdition(ctx, overall, day, windowStart, windowEnd, false); !errors.Is(err, ErrCountdownEditionExists) {
		t.Fatalf("second selection err = %v, want ErrCountdownEditionExists", err)
	}

	// A different day does not see today's editions, so the category gets
	// both alpha clips.
	nextDay := "2026-11-08"
	edition, err = repo.SelectEdition(ctx, category, nextDay, windowStart, windowEnd, false)
	if err != nil {
		t.Fatalf("select category next day: %v", err)
	}
	if got, want := editionClipNames(edition), []string{"a1", "a2"}; !equalStrings(got, want) {
		t.Fatalf("category next day = %v, want %v", got, want)
	}

	// The database refuses a clip in two editions on one day.
	_, err = pool.Exec(ctx, `
		INSERT INTO countdown_edition_clips (edition_id, edition_date, position, clip_id, view_velocity)
		SELECT e.id, e.edition_date, 99, $1, 1 FROM countdown_editions e WHERE e.series_id = $2 AND e.edition_date = $3::date`,
		ids["a1"], category.ID, day)
	if err == nil {
		t.Fatal("a clip was stored in two editions on one day")
	}

	// Reselecting the overall edition after excluding alpha moves everyone up.
	alpha := "alpha"
	if _, err := repo.CreateExclusion(ctx, &models.CountdownExclusionInput{BroadcasterID: &alpha, Reason: "request"}, "self_service", nil); err != nil {
		t.Fatalf("exclude alpha: %v", err)
	}
	edition, err = repo.SelectEdition(ctx, overall, day, windowStart, windowEnd, true)
	if err != nil {
		t.Fatalf("reselect overall: %v", err)
	}
	if got, want := editionClipNames(edition), []string{"b1", "c1", "d1", "e1"}; !equalStrings(got, want) {
		t.Fatalf("reselected overall = %v, want %v", got, want)
	}

	// An edition that has moved on to fetching is not replaced.
	if _, err := pool.Exec(ctx, `UPDATE countdown_editions SET status = 'fetching' WHERE id = $1`, edition.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.SelectEdition(ctx, overall, day, windowStart, windowEnd, true); !errors.Is(err, ErrCountdownEditionInProgress) {
		t.Fatalf("reselect in progress err = %v", err)
	}

	listed, err := repo.ListEditions(ctx, &day, nil, 10)
	if err != nil {
		t.Fatalf("list editions: %v", err)
	}
	if len(listed) != 2 || listed[0].SeriesSlug != "overall" || listed[1].SeriesSlug != "category-g1" || len(listed[0].Clips) != 4 {
		t.Fatalf("listed = %+v", listed)
	}
	exists, err := repo.EditionExists(ctx, overall.ID, day)
	if err != nil || !exists {
		t.Fatalf("EditionExists = %v, %v", exists, err)
	}

	if _, err := repo.CreateSeries(ctx, &models.CountdownSeriesInput{Slug: "overall", Name: "dup", Size: 1, WindowHours: 1, PerBroadcasterCap: 1}); !errors.Is(err, ErrCountdownSeriesSlugTaken) {
		t.Fatalf("duplicate slug err = %v", err)
	}
	exclusions, err := repo.ListExclusions(ctx)
	if err != nil || len(exclusions) != 3 {
		t.Fatalf("exclusions = %+v, %v", exclusions, err)
	}
	if err := repo.DeleteExclusion(ctx, exclusions[0].ID); err != nil {
		t.Fatalf("delete exclusion: %v", err)
	}
	if err := repo.DeleteExclusion(ctx, exclusions[0].ID); !errors.Is(err, ErrCountdownExclusionNotFound) {
		t.Fatalf("second delete err = %v", err)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestCountdownMinSize(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "countdown_series", "clips", "users")
	ctx := context.Background()
	repo := NewCountdownRepository(pool)

	day := "2026-11-07"
	windowEnd := time.Date(2026, 11, 7, 12, 0, 0, 0, time.UTC)
	windowStart := windowEnd.Add(-24 * time.Hour)
	insertCountdownClips(t, pool, windowEnd, []countdownClip{
		{name: "a", broadcaster: "alpha", game: "g1", language: "en", velocity: 900, duration: 30, age: time.Hour},
		{name: "b", broadcaster: "bravo", game: "g1", language: "en", velocity: 800, duration: 30, age: time.Hour},
		{name: "c", broadcaster: "charlie", game: "g1", language: "en", velocity: 700, duration: 30, age: time.Hour},
		{name: "d", broadcaster: "delta", game: "g2", language: "en", velocity: 600, duration: 30, age: time.Hour},
	})
	three := 3
	create := func(slug, game string, priority int) *models.CountdownSeries {
		t.Helper()
		series, err := repo.CreateSeries(ctx, &models.CountdownSeriesInput{
			Slug: slug, Name: slug, GameID: &game, Size: 5, MinSize: &three, WindowHours: 24,
			PerBroadcasterCap: 1, SelectionHourUTC: 12, Priority: priority, Enabled: true,
		})
		if err != nil {
			t.Fatalf("create %s: %v", slug, err)
		}
		if series.MinSize == nil || *series.MinSize != 3 {
			t.Fatalf("%s min_size = %v", slug, series.MinSize)
		}
		return series
	}
	short := create("short-list", "g1", 0)
	thin := create("thin-day", "g2", 1)

	edition, err := repo.SelectEdition(ctx, short, day, windowStart, windowEnd, false)
	if err != nil {
		t.Fatalf("select short list: %v", err)
	}
	if got, want := editionClipNames(edition), []string{"a", "b", "c"}; edition.Status != models.CountdownEditionSelected || !equalStrings(got, want) {
		t.Fatalf("short list = %s %v, want selected %v", edition.Status, got, want)
	}

	edition, err = repo.SelectEdition(ctx, thin, day, windowStart, windowEnd, false)
	if err != nil {
		t.Fatalf("select thin day: %v", err)
	}
	if edition.Status != models.CountdownEditionSkipped || len(edition.Clips) != 0 || edition.FailureReason == nil ||
		*edition.FailureReason != "only 1 of 3 clips were eligible" {
		t.Fatalf("thin day = %+v", edition)
	}

	progress, err := repo.ListSeriesProgress(ctx)
	if err != nil {
		t.Fatalf("progress: %v", err)
	}
	if len(progress) != 2 || progress[1].LatestRendered == nil || *progress[1].LatestRendered != day ||
		progress[1].LatestPublished == nil || *progress[1].LatestPublished != day {
		t.Fatalf("a skipped day should count as reached: %+v", progress)
	}

	// A skipped edition can be selected again once more clips exist.
	insertCountdownClips(t, pool, windowEnd, []countdownClip{
		{name: "e", broadcaster: "echo", game: "g2", language: "en", velocity: 500, duration: 30, age: time.Hour},
		{name: "f", broadcaster: "foxtrot", game: "g2", language: "en", velocity: 400, duration: 30, age: time.Hour},
	})
	edition, err = repo.SelectEdition(ctx, thin, day, windowStart, windowEnd, true)
	if err != nil {
		t.Fatalf("reselect thin day: %v", err)
	}
	if got, want := editionClipNames(edition), []string{"d", "e", "f"}; edition.Status != models.CountdownEditionSelected || !equalStrings(got, want) {
		t.Fatalf("reselected = %s %v, want selected %v", edition.Status, got, want)
	}
}
