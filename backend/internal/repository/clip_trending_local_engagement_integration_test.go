//go:build integration

package repository

import (
	"context"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/google/uuid"
)

// An old clip that receives new local votes, comments or favorites must be
// rewritten by the windowed refresh even though it was created, imported and
// last view-observed outside the window.
func TestClipRepository_TrendingRefreshIncludesOldClipsWithRecentLocalEngagement(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	ctx := context.Background()
	id := uuid.New()
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM clips WHERE id = $1`, id) })
	if _, err := pool.Exec(ctx, `
		INSERT INTO clips (id,twitch_clip_id,twitch_clip_url,embed_url,title,creator_name,broadcaster_name,view_count,
			created_at,imported_at,view_count_observed_at)
		VALUES ($1,$2,'https://clips.twitch.tv/x','https://clips.twitch.tv/embed?clip=x','old clip','creator','broadcaster',40,
			NOW()-interval '60 days',NOW()-interval '60 days',NOW()-interval '60 days')`, id, "local-engagement-"+id.String()); err != nil {
		t.Fatal(err)
	}
	repo := NewClipRepository(pool)
	if _, err := repo.updateTrendingScoresBatched(ctx, 200, 0); err != nil {
		t.Fatalf("baseline refresh: %v", err)
	}

	// A new favorite (what the favorites trigger records) on the old clip.
	if _, err := pool.Exec(ctx, `UPDATE clips SET favorite_count = favorite_count + 1 WHERE id = $1`, id); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO clip_local_engagement(clip_id,occurred_at,score_delta) VALUES($1,NOW(),2)`, id); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.UpdateTrendingScoresBatched(ctx, 200); err != nil {
		t.Fatalf("windowed refresh: %v", err)
	}
	var popularity, engagement, want int64
	if err := pool.QueryRow(ctx, `SELECT popularity_index, engagement_count,
		view_count + vote_score*2 + comment_count*3 + favorite_count*2 FROM clips WHERE id = $1`, id).Scan(&popularity, &engagement, &want); err != nil {
		t.Fatal(err)
	}
	if popularity != want || engagement != want {
		t.Fatalf("popularity_index=%d engagement_count=%d, want %d", popularity, engagement, want)
	}
}
