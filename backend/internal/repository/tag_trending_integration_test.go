//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// The trending list counts seven-day assignments per tag, ignores hidden and
// removed clips, and orders by that count. This pins the aggregate-first
// query shape, which exists so the blacklist filter runs per tag, not per
// assignment.
func TestTagListTrendingCountsRecentVisibleAssignments(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "clip_tags", "tags", "clips")
	ctx := context.Background()
	exec := func(sql string, args ...any) {
		t.Helper()
		_, err := pool.Exec(ctx, sql, args...)
		require.NoError(t, err)
	}
	newClip := func(hidden bool) uuid.UUID {
		id := uuid.New()
		exec(`INSERT INTO clips (id,twitch_clip_id,twitch_clip_url,embed_url,title,creator_name,broadcaster_name,created_at,is_hidden)
			VALUES ($1,$2,'https://clips.example/t','https://clips.example/e','trending','creator','broadcaster',NOW(),$3)`,
			id, "trending-"+id.String(), hidden)
		return id
	}
	quiet, busy := uuid.New(), uuid.New()
	exec(`INSERT INTO tags(id,name,slug,parent_slug) VALUES ($1,'Quiet','community/quiet','community'),($2,'Busy','community/busy','community')`, quiet, busy)
	assign := func(tag uuid.UUID, clip uuid.UUID, at time.Time) {
		exec(`INSERT INTO clip_tags(clip_id,tag_id,created_at) VALUES ($1,$2,$3)`, clip, tag, at)
	}
	now := time.Now()
	assign(quiet, newClip(false), now)
	// A hidden clip and an assignment outside the window do not count.
	assign(quiet, newClip(true), now)
	assign(quiet, newClip(false), now.Add(-8*24*time.Hour))
	for i := 0; i < 3; i++ {
		assign(busy, newClip(false), now.Add(-time.Duration(i)*time.Hour))
	}

	tags, err := NewTagRepository(pool).List(ctx, "trending", "", 10, 0)
	require.NoError(t, err)
	require.Len(t, tags, 2)
	require.Equal(t, "community/busy", tags[0].Slug)
	require.Equal(t, 3, tags[0].UsageCount)
	require.Equal(t, "community/quiet", tags[1].Slug)
	require.Equal(t, 1, tags[1].UsageCount)
}
