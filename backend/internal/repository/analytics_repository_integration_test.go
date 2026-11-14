//go:build integration

package repository

import (
	"context"
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/testutil"
	"github.com/google/uuid"
)

func TestPlatformOverviewCountsOnlySignedInProductUsers(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "users", "clips", "platform_analytics")

	ctx := context.Background()
	identities := []struct {
		username string
		role     string
		status   string
	}{
		{"signed-in-user", "user", "active"},
		{"imported-creator", "user", "unclaimed"},
		{"staff-admin", "admin", "active"},
		{"staff-moderator", "moderator", "active"},
		{"service-identity", "service", "active"},
		{"merged-user", "user", "merged"},
	}
	for _, identity := range identities {
		if _, err := pool.Exec(ctx, `
			INSERT INTO users (id, twitch_id, username, display_name, role, account_status)
			VALUES ($1, $2, $3, $3, $4, $5)
		`, uuid.New(), uuid.NewString(), identity.username, identity.role, identity.status); err != nil {
			t.Fatalf("insert identity %s: %v", identity.username, err)
		}
	}
	for _, removed := range []bool{false, true} {
		clipID := uuid.NewString()
		if _, err := pool.Exec(ctx, `
			INSERT INTO clips (
				id, twitch_clip_id, twitch_clip_url, embed_url, title,
				creator_name, broadcaster_name, created_at, is_removed
			) VALUES ($1, $2, $3, $4, $5, $6, $6, NOW(), $7)
		`, uuid.New(), clipID, "https://clips.twitch.tv/"+clipID, "https://clips.twitch.tv/embed?clip="+clipID, "Analytics clip", "creator", removed); err != nil {
			t.Fatalf("insert clip: %v", err)
		}
	}

	overview, err := NewAnalyticsRepository(pool).GetPlatformOverviewMetrics(ctx)
	if err != nil {
		t.Fatalf("get platform overview: %v", err)
	}
	if overview.TotalUsers != 1 {
		t.Fatalf("total_users = %d, want 1 active non-staff signed-in user", overview.TotalUsers)
	}
	if overview.TotalClips != 1 {
		t.Fatalf("total_clips = %d, want 1 visible canonical clip", overview.TotalClips)
	}
	if overview.ActiveUsersDaily != nil || overview.ActiveUsersMonthly != nil || overview.AvgSessionDuration != nil {
		t.Fatalf("unavailable engagement metrics must remain null: %#v", overview)
	}
}

func TestAdminIdentitySummarySeparatesUsersCreatorsAndStaff(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	defer testutil.CleanupTestDB(t, pool)
	testutil.TruncateTables(t, pool, "users")

	ctx := context.Background()
	identities := []struct {
		username string
		role     string
		status   string
	}{
		{"signed-in-user", "user", "active"},
		{"imported-creator", "user", "unclaimed"},
		{"staff-admin", "admin", "active"},
		{"staff-moderator", "moderator", "active"},
		{"service-identity", "service", "active"},
		{"merged-user", "user", "merged"},
	}
	for _, identity := range identities {
		if _, err := pool.Exec(ctx, `
			INSERT INTO users (id, twitch_id, username, display_name, role, account_status)
			VALUES ($1, $2, $3, $3, $4, $5)
		`, uuid.New(), uuid.NewString(), identity.username, identity.role, identity.status); err != nil {
			t.Fatalf("insert identity %s: %v", identity.username, err)
		}
	}

	summary, err := NewUserRepository(pool).GetAdminIdentitySummary(ctx)
	if err != nil {
		t.Fatalf("get identity summary: %v", err)
	}
	if summary.SignedInUsers != 1 || summary.UnclaimedCreators != 1 || summary.Staff != 2 || summary.Other != 2 {
		t.Fatalf("unexpected identity summary: %#v", summary)
	}
}

func TestCreatorAnalyticsClipQueriesSkipHiddenAndDMCARemovedClips(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	ctx := context.Background()
	repo := NewAnalyticsRepository(pool)
	creator := "analytics-" + uuid.NewString()[:8]

	visible := insertVisibilityClip(t, pool, creator, creator, false, false)
	hidden := insertVisibilityClip(t, pool, creator, creator, true, false)
	dmca := insertVisibilityClip(t, pool, creator, creator, false, true)
	for _, clipID := range []uuid.UUID{visible, hidden, dmca} {
		if _, err := pool.Exec(ctx, `
			INSERT INTO analytics_events (event_type, clip_id, user_agent)
			VALUES ('clip_view', $1, 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)')`, clipID); err != nil {
			t.Fatalf("insert view event: %v", err)
		}
		t.Cleanup(func() {
			_, _ = pool.Exec(context.Background(), `DELETE FROM analytics_events WHERE clip_id = $1`, clipID)
		})
	}

	for _, sortBy := range []string{"votes", "views", "comments"} {
		clips, err := repo.GetCreatorTopClips(ctx, creator, sortBy, 10)
		if err != nil {
			t.Fatalf("GetCreatorTopClips(%s): %v", sortBy, err)
		}
		if len(clips) != 1 || clips[0].ID != visible {
			t.Fatalf("top clips by %s = %d clips, want only the visible clip %s", sortBy, len(clips), visible)
		}
	}

	insights, err := repo.GetCreatorAudienceInsights(ctx, creator, 10)
	if err != nil {
		t.Fatalf("GetCreatorAudienceInsights: %v", err)
	}
	if insights.TotalViews != 1 {
		t.Fatalf("audience total_views = %d, want 1 view of the visible clip", insights.TotalViews)
	}
}
