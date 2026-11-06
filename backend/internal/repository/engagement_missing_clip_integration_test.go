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

// A clip Twitch stops returning backs off toward daily polls and, once it has
// been missing for over an hour, stops marking every ranking as partial. A
// failed request keeps its 15-minute retry and stays stale.
func TestEngagementMissingClipBacksOffAndLeavesCoverage(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	ctx := context.Background()
	repo := NewEngagementRepository(pool)
	gone, justMissing, failed := uuid.New(), uuid.New(), uuid.New()
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM clips WHERE id = ANY($1)`, []uuid.UUID{gone, justMissing, failed})
	})
	lastObserved := map[uuid.UUID]time.Duration{gone: 3 * 24 * time.Hour, justMissing: 5 * time.Minute, failed: 3 * 24 * time.Hour}
	for id, age := range lastObserved {
		_, err := pool.Exec(ctx, `INSERT INTO clips(id,twitch_clip_id,twitch_clip_url,embed_url,title,creator_name,broadcaster_name,created_at,view_count)
			VALUES($1,$2,'https://clips.twitch.tv/x','https://clips.twitch.tv/embed','t','c','b',now()-interval '30 days',10)`, id, "missing-"+id.String())
		require.NoError(t, err)
		_, err = pool.Exec(ctx, `UPDATE clip_engagement_state SET observed_at=now()-$2::interval, next_poll_at=now() WHERE clip_id=$1`,
			id, age.String())
		require.NoError(t, err)
	}
	require.NoError(t, repo.PollMissing(ctx, gone))
	require.NoError(t, repo.PollMissing(ctx, justMissing))
	require.NoError(t, repo.PollFailed(ctx, failed))

	retryIn := func(id uuid.UUID) time.Duration {
		var seconds float64
		require.NoError(t, pool.QueryRow(ctx, `SELECT EXTRACT(EPOCH FROM next_poll_at-now()) FROM clip_engagement_state WHERE clip_id=$1`, id).Scan(&seconds))
		return time.Duration(seconds) * time.Second
	}
	stale := func(id uuid.UUID) bool {
		var v bool
		require.NoError(t, pool.QueryRow(ctx, `SELECT `+engagementStaleClip+`
			FROM clips c LEFT JOIN clip_engagement_state s ON s.clip_id=c.id WHERE c.id=$1`, id).Scan(&v))
		return v
	}
	require.InDelta(t, (24 * time.Hour).Seconds(), retryIn(gone).Seconds(), 60, "long-missing clip retries daily")
	require.InDelta(t, (15 * time.Minute).Seconds(), retryIn(justMissing).Seconds(), 60, "recently missing clip retries soon")
	require.InDelta(t, (15 * time.Minute).Seconds(), retryIn(failed).Seconds(), 60)
	require.False(t, stale(gone), "a long-missing clip is not a coverage gap")
	require.True(t, stale(justMissing), "a recently missing clip is still stale")
	require.True(t, stale(failed), "a failed request is still stale")

	// Claiming the gone clip for its daily retry keeps it out of the stale count.
	_, err := pool.Exec(ctx, `UPDATE clip_engagement_state SET next_poll_at=now()+interval '5 minutes', last_poll_error_at=now()-interval '1 day' WHERE clip_id=$1`, gone)
	require.NoError(t, err)
	require.False(t, stale(gone))
}
