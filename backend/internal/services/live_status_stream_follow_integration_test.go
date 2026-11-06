//go:build integration

package services

import (
	"context"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"git.subcult.tv/subculture-collective/clpr/pkg/twitch"
	"github.com/google/uuid"
)

// The stream page and its follow/unfollow API are gone, so a broadcaster going
// live must not send stream_live notifications linking to /stream/<login>.
func TestLiveStatusDoesNotNotifyStreamFollowers(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	ctx := context.Background()

	suffix := uuid.NewString()[:8]
	broadcasterID := "lsf_" + suffix
	login := "lsf_login_" + suffix
	user := testutil.TestUser()
	user.Username, user.TwitchID, user.Email = "lsf_user_"+suffix, testutil.StringPtr("lsf_user_"+suffix), testutil.StringPtr("lsf_"+suffix+"@example.com")
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM notifications WHERE user_id = $1`, user.ID)
		_, _ = pool.Exec(ctx, `DELETE FROM stream_follows WHERE user_id = $1`, user.ID)
		_, _ = pool.Exec(ctx, `DELETE FROM broadcaster_sync_log WHERE broadcaster_id = $1`, broadcasterID)
		_, _ = pool.Exec(ctx, `DELETE FROM broadcaster_sync_status WHERE broadcaster_id = $1`, broadcasterID)
		_, _ = pool.Exec(ctx, `DELETE FROM broadcaster_live_status WHERE broadcaster_id = $1`, broadcasterID)
		_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, user.ID)
	})

	userRepo := repository.NewUserRepository(pool)
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	streamFollowRepo := repository.NewStreamFollowRepository(pool)
	if _, err := streamFollowRepo.FollowStreamer(ctx, user.ID, login, true); err != nil {
		t.Fatalf("follow streamer: %v", err)
	}

	client := &fakeStreamsClient{live: map[string]twitch.Stream{
		broadcasterID: {UserID: broadcasterID, UserLogin: login, UserName: login, Type: "live", Title: "stream", StartedAt: time.Now().Add(-time.Minute)},
	}}
	svc := &LiveStatusService{
		broadcasterRepo:  repository.NewBroadcasterRepository(pool),
		streamFollowRepo: streamFollowRepo,
		twitchClient:     client,
	}
	svc.SetNotificationService(NewNotificationService(repository.NewNotificationRepository(pool), userRepo, nil, nil, nil, nil))

	if err := svc.UpdateLiveStatusForBroadcasters(ctx, []string{broadcasterID}); err != nil {
		t.Fatalf("update live status: %v", err)
	}
	var wentLive int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM broadcaster_sync_log WHERE broadcaster_id = $1 AND status_change = 'went_live'`, broadcasterID).Scan(&wentLive); err != nil || wentLive != 1 {
		t.Fatalf("went_live log count = %d, %v; want 1", wentLive, err)
	}
	var streamLive int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND type = 'stream_live'`, user.ID).Scan(&streamLive); err != nil {
		t.Fatalf("count notifications: %v", err)
	}
	if streamLive != 0 {
		t.Fatalf("stream_live notifications = %d, want 0", streamLive)
	}
}
