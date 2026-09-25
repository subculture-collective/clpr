//go:build integration

package repository

import (
	"context"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/google/uuid"
)

// users.display_name is nullable. A single user without one made every
// leaderboard fail with "cannot scan NULL into *string" (HTTP 500).
func TestLeaderboardsFallBackToUsernameWithoutDisplayName(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "user_stats", "users")

	ctx := context.Background()
	userID := uuid.New()
	if _, err := pool.Exec(ctx, `INSERT INTO users (id, twitch_id, username) VALUES ($1, $2, $3)`, userID, uuid.NewString(), "no-display-name"); err != nil {
		t.Fatalf("create user: %v", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO user_stats (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING`, userID); err != nil {
		t.Fatalf("create user stats: %v", err)
	}

	repo := NewReputationRepository(pool)
	boards := map[string]func() (string, error){
		"karma": func() (string, error) {
			entries, err := repo.GetKarmaLeaderboard(ctx, 10, 0)
			if err != nil || len(entries) != 1 {
				return "", err
			}
			return entries[0].DisplayName, nil
		},
		"engagement": func() (string, error) {
			entries, err := repo.GetEngagementLeaderboard(ctx, 10, 0)
			if err != nil || len(entries) != 1 {
				return "", err
			}
			return entries[0].DisplayName, nil
		},
		"trust score": func() (string, error) {
			entries, err := repo.GetTrustScoreLeaderboard(ctx, 10, 0)
			if err != nil || len(entries) != 1 {
				return "", err
			}
			return entries[0].DisplayName, nil
		},
	}
	for name, load := range boards {
		displayName, err := load()
		if err != nil {
			t.Fatalf("%s leaderboard: %v", name, err)
		}
		if displayName != "no-display-name" {
			t.Errorf("%s leaderboard display name = %q, want the username", name, displayName)
		}
	}
}
