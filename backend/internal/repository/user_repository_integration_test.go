//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/google/uuid"
)

func TestUserRepository_GetUserIDsActiveSinceSkipsIdleAccounts(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "user_activity", "users")
	ctx := context.Background()

	recentLogin, recentActivity, idle, banned := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	if _, err := pool.Exec(ctx, `
		INSERT INTO users (id, twitch_id, username, display_name, last_login_at, is_banned, created_at) VALUES
		($1::uuid, 'tw-'||$1, 'recent-login', 'Recent Login', NOW() - interval '3 days', false, NOW() - interval '400 days'),
		($2::uuid, 'tw-'||$2, 'recent-activity', 'Recent Activity', NOW() - interval '200 days', false, NOW() - interval '300 days'),
		($3::uuid, 'tw-'||$3, 'idle', 'Idle', NOW() - interval '200 days', false, NOW() - interval '200 days'),
		($4::uuid, 'tw-'||$4, 'banned', 'Banned', NOW() - interval '1 day', true, NOW() - interval '100 days')`,
		recentLogin.String(), recentActivity.String(), idle.String(), banned.String()); err != nil {
		t.Fatalf("seed users: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO user_activity (user_id, activity_type, created_at) VALUES ($1, 'vote', NOW() - interval '10 days')`, recentActivity); err != nil {
		t.Fatalf("seed activity: %v", err)
	}

	ids, err := NewUserRepository(pool).GetUserIDsActiveSince(ctx, time.Now().Add(-90*24*time.Hour))
	if err != nil {
		t.Fatalf("GetUserIDsActiveSince: %v", err)
	}
	if len(ids) != 2 || ids[0] != recentLogin || ids[1] != recentActivity {
		t.Fatalf("active users = %v, want [%s %s]", ids, recentLogin, recentActivity)
	}
}
