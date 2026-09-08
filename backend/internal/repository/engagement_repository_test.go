//go:build integration

package repository

import (
	"context"
	"sync"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"git.subcult.tv/subculture-collective/clpr/internal/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// This contract deliberately exercises triggers through source mutations, not
// fabricated aggregate rows. PostgreSQL owns atomicity and retry behavior.
func TestEngagementContract(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer pool.Close()
	ctx := context.Background()
	exec := func(sql string, args ...any) {
		t.Helper()
		_, err := pool.Exec(ctx, sql, args...)
		require.NoError(t, err)
	}
	exec("TRUNCATE clips,users,engagement_generations CASCADE")
	repo := NewEngagementRepository(pool)
	newClip := func(lifetime int) uuid.UUID {
		t.Helper()
		id := uuid.New()
		exec(`INSERT INTO clips(id,twitch_clip_id,twitch_clip_url,embed_url,title,creator_name,broadcaster_name,created_at,view_count)
   VALUES($1,$2,'https://clips.twitch.tv/test','https://clips.twitch.tv/embed','test','test','test',now()-interval '2 years',$3)`, id, id.String(), lifetime)
		exec("DELETE FROM clip_view_observations WHERE clip_id=$1", id)
		exec("DELETE FROM clip_engagement_state WHERE clip_id=$1", id)
		return id
	}
	now := time.Now().UTC().Truncate(time.Second)
	active, inactive, weekly := newClip(100), newClip(1000000), newClip(500)
	observe := func(id uuid.UUID, count int, at time.Time) {
		t.Helper()
		require.NoError(t, repo.Observe(ctx, id, count, at))
	}
	observe(active, 100, now.Add(-2*time.Hour))
	observe(active, 220, now)
	observe(active, 220, now)                  // duplicate
	observe(active, 9999, now.Add(-time.Hour)) // out of order
	observe(inactive, 1000000, now.Add(-time.Hour))
	observe(inactive, 1000000, now)
	observe(weekly, 500, now.Add(-72*time.Hour))
	observe(weekly, 1000, now.Add(-48*time.Hour))
	require.NoError(t, repo.Publish(ctx))
	score := func(period string, id uuid.UUID) float64 {
		t.Helper()
		var n float64
		require.NoError(t, pool.QueryRow(ctx, `SELECT COALESCE(sum(score),0) FROM engagement_rankings WHERE period=$1 AND clip_id=$2`, period, id).Scan(&n))
		return n
	}
	require.InDelta(t, 60, score("hour", active), 0.2, "boundary is uniform observed growth, not lifetime views")
	require.InDelta(t, 120, score("day", active), 0.2)
	require.Zero(t, score("day", inactive))
	require.Zero(t, score("day", weekly))
	require.InDelta(t, 500, score("week", weekly), 0.2)
	m, err := repo.Resolve(ctx, "year", nil)
	require.NoError(t, err)
	require.True(t, m.PartialCoverage)
	original := m.Generation
	require.NoError(t, repo.Publish(ctx))
	m, err = repo.Resolve(ctx, "day", nil)
	require.NoError(t, err)
	require.Equal(t, original, m.Generation)
	observe(active, 10, now.Add(time.Second)) // counter reset
	observe(active, 15, now.Add(2*time.Second))
	var total float64
	require.NoError(t, pool.QueryRow(ctx, "SELECT view_gain FROM clip_engagement_state WHERE clip_id=$1", active).Scan(&total))
	require.Equal(t, 125.0, total)
	require.NoError(t, repo.PollFailed(ctx, active))
	exec("UPDATE engagement_generations SET published_at=now()-interval '2 hours'")
	_, err = repo.Resolve(ctx, "day", &original)
	require.ErrorIs(t, err, ErrRankingExpired)
	exec("UPDATE clips SET is_hidden=true WHERE id=$1", weekly)
	require.NoError(t, repo.Publish(ctx))
	require.Zero(t, score("week", weekly))
	m, err = repo.Resolve(ctx, "day", nil)
	require.NoError(t, err)
	require.Greater(t, m.StaleClips, int64(0))

	t.Run("local reversals rollback and concurrent retries", func(t *testing.T) {
		id := newClip(0)
		user := uuid.New()
		exec("INSERT INTO users(id,twitch_id,username) VALUES($1,$2,'engagement')", user, user.String())
		insert := `INSERT INTO votes(user_id,clip_id,vote_type) VALUES($1,$2,1) ON CONFLICT(user_id,clip_id) DO UPDATE SET vote_type=EXCLUDED.vote_type`
		var wg sync.WaitGroup
		errs := make(chan error, 8)
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func() { defer wg.Done(); _, e := pool.Exec(ctx, insert, user, id); errs <- e }()
		}
		wg.Wait()
		close(errs)
		for e := range errs {
			require.NoError(t, e)
		}
		exec("INSERT INTO favorites(user_id,clip_id) VALUES($1,$2) ON CONFLICT DO NOTHING", user, id)
		exec("INSERT INTO comments(user_id,clip_id,content) VALUES($1,$2,'test')", user, id)
		local := func() int64 {
			var n int64
			require.NoError(t, pool.QueryRow(ctx, "SELECT 2*vote_gain+3*comment_gain+2*favorite_gain FROM clip_engagement_state WHERE clip_id=$1", id).Scan(&n))
			return n
		}
		require.Equal(t, int64(7), local())
		tx, e := pool.Begin(ctx)
		require.NoError(t, e)
		_, e = tx.Exec(ctx, "DELETE FROM favorites WHERE user_id=$1 AND clip_id=$2", user, id)
		require.NoError(t, e)
		require.NoError(t, tx.Rollback(ctx))
		require.Equal(t, int64(7), local())
		exec("UPDATE votes SET vote_type=-1 WHERE user_id=$1 AND clip_id=$2", user, id)
		exec("UPDATE comments SET is_removed=true WHERE clip_id=$1", id)
		exec("DELETE FROM favorites WHERE clip_id=$1", id)
		require.Equal(t, int64(-2), local())
		exec("DELETE FROM votes WHERE clip_id=$1", id)
		require.Zero(t, local())
	})

	t.Run("observations and source mutations share a consistent lock order", func(t *testing.T) {
		id := newClip(100)
		user := uuid.New()
		exec("INSERT INTO users(id,twitch_id,username) VALUES($1,$2,'concurrent')", user, user.String())
		observe(id, 100, now.Add(-time.Hour))
		deadline, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		var wg sync.WaitGroup
		errs := make(chan error, 16)
		for i := 1; i <= 8; i++ {
			wg.Add(2)
			go func(i int) {
				defer wg.Done()
				errs <- repo.Observe(deadline, id, 100+i, now.Add(-time.Second+time.Duration(i)*time.Millisecond))
			}(i)
			go func() {
				defer wg.Done()
				_, err := pool.Exec(deadline, `INSERT INTO votes(user_id,clip_id,vote_type) VALUES($1,$2,1)
					ON CONFLICT(user_id,clip_id) DO UPDATE SET vote_type=EXCLUDED.vote_type`, user, id)
				errs <- err
			}()
		}
		wg.Wait()
		close(errs)
		for err := range errs {
			require.NoError(t, err)
		}
		var views float64
		var votes int64
		require.NoError(t, pool.QueryRow(ctx, "SELECT view_gain,vote_gain FROM clip_engagement_state WHERE clip_id=$1", id).Scan(&views, &votes))
		require.Equal(t, 8.0, views)
		require.Equal(t, int64(1), votes)
	})

	t.Run("resumable leases and immutable deterministic pagination", func(t *testing.T) {
		tie := newClip(9000)
		observe(tie, 9000, now.Add(-time.Hour))
		observe(tie, 9125, now)
		exec("UPDATE engagement_generations SET published_at=now()-interval '2 hours'")
		require.NoError(t, repo.Publish(ctx))
		generation, err := repo.Resolve(ctx, "all", nil)
		require.NoError(t, err)
		clips := NewClipRepository(pool)
		filters := ClipFilters{Sort: "trending", RankingGeneration: &generation.Generation, RankingPeriod: "all"}
		first, _, err := clips.ListWithFilters(ctx, filters, 1, 0)
		require.NoError(t, err)
		require.Len(t, first, 1)
		wantFirst, wantSecond := active, tie
		if tie.String() > active.String() {
			wantFirst, wantSecond = tie, active
		}
		require.Equal(t, wantFirst, first[0].ID)
		// Lifetime mutations after publication cannot reorder the retained ranking.
		exec("UPDATE clips SET vote_score=999999 WHERE id=$1", wantSecond)
		cursor := utils.EncodeCursor("trending", first[0].TrendingScore, first[0].ID, 0)
		filters.Cursor = &cursor
		second, _, err := clips.ListWithFilters(ctx, filters, 1, 0)
		require.NoError(t, err)
		require.Len(t, second, 1)
		require.Equal(t, wantSecond, second[0].ID)
		// Isolate the due catalog, then interrupt after claiming its only member.
		exec("UPDATE clip_engagement_state SET next_poll_at=now()+interval '1 day'")
		exec("UPDATE clip_engagement_state SET next_poll_at=now() WHERE clip_id=$1", tie)
		claimed, err := repo.ClaimDue(ctx, 1)
		require.NoError(t, err)
		require.Equal(t, []EngagementPollClip{{ID: tie, TwitchID: tie.String()}}, claimed)
		claimed, err = repo.ClaimDue(ctx, 1)
		require.NoError(t, err)
		require.Empty(t, claimed)
		exec("UPDATE clip_engagement_state SET next_poll_at=now()-interval '1 second' WHERE clip_id=$1", tie)
		claimed, err = repo.ClaimDue(ctx, 1)
		require.NoError(t, err)
		require.Len(t, claimed, 1)
		_, err = repo.ClaimDue(ctx, 101)
		require.Error(t, err)
	})
}
