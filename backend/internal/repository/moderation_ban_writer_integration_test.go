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

func setupBanWriterFixture(t *testing.T) (*pgxpool.Pool, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	pool := testutil.SetupTestDB(t)
	ctx := context.Background()
	moderator, target, community := uuid.New(), uuid.New(), uuid.New()
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM moderation_audit_logs WHERE channel_id = $1`, community)
		// Deleting the users cascades to the community and its bans.
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id = ANY($1)`, []uuid.UUID{moderator, target})
	})
	for _, id := range []uuid.UUID{moderator, target} {
		name := "banwriter-" + id.String()[:18]
		if _, err := pool.Exec(ctx, `INSERT INTO users(id,twitch_id,username) VALUES($1,$2,$2)`, id, name); err != nil {
			t.Fatal(err)
		}
	}
	slug := "banwriter-" + community.String()
	if _, err := pool.Exec(ctx, `INSERT INTO communities(id,name,slug,owner_id) VALUES($1,$2,$2,$3)`, community, slug, moderator); err != nil {
		t.Fatal(err)
	}
	return pool, moderator, target, community
}

func banAudit(moderator, community uuid.UUID, action string) *models.ModerationAuditLog {
	return &models.ModerationAuditLog{
		ID: uuid.New(), Action: action, EntityType: "community_ban", EntityID: community,
		ModeratorID: moderator, ChannelID: &community, CreatedAt: time.Now(),
	}
}

func TestModerationBanWriter_RepeatedBanUpdatesExistingBan(t *testing.T) {
	pool, moderator, target, community := setupBanWriterFixture(t)
	ctx := context.Background()
	writer := NewModerationBanWriter(pool)

	var firstID uuid.UUID
	for i, reason := range []string{"first", "second"} {
		reason := reason
		ban := &models.CommunityBan{ID: uuid.New(), CommunityID: community, BannedUserID: target,
			BannedByUserID: &moderator, Reason: &reason, BannedAt: time.Now()}
		audit := banAudit(moderator, community, "ban_user")
		audit.EntityID = ban.ID
		if err := writer.CreateBanWithAudit(ctx, ban, audit); err != nil {
			t.Fatalf("ban %d: %v", i, err)
		}
		if i == 0 {
			firstID = ban.ID
		} else if ban.ID != firstID || audit.EntityID != firstID {
			t.Fatalf("repeated ban id=%s audit entity=%s, want existing %s", ban.ID, audit.EntityID, firstID)
		}
	}
	var bans, audits int
	var reason string
	if err := pool.QueryRow(ctx, `SELECT COUNT(*), MAX(reason) FROM community_bans WHERE community_id=$1 AND banned_user_id=$2`, community, target).Scan(&bans, &reason); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM moderation_audit_logs WHERE channel_id=$1 AND action='ban_user' AND entity_id=$2`, community, firstID).Scan(&audits); err != nil {
		t.Fatal(err)
	}
	if bans != 1 || reason != "second" || audits != 2 {
		t.Fatalf("bans=%d reason=%q audits=%d", bans, reason, audits)
	}
}

func TestModerationBanWriter_UnbanWithAuditIsAtomic(t *testing.T) {
	pool, moderator, target, community := setupBanWriterFixture(t)
	ctx := context.Background()
	writer := NewModerationBanWriter(pool)

	ban := &models.CommunityBan{ID: uuid.New(), CommunityID: community, BannedUserID: target, BannedByUserID: &moderator, BannedAt: time.Now()}
	if err := writer.CreateBanWithAudit(ctx, ban, banAudit(moderator, community, "ban_user")); err != nil {
		t.Fatal(err)
	}

	// A failing audit insert (unknown moderator FK) must leave the ban in place.
	bad := banAudit(uuid.New(), community, "unban_user")
	if err := writer.UnbanWithAudit(ctx, community, target, bad); err == nil {
		t.Fatal("expected audit failure")
	}
	var bans int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM community_bans WHERE community_id=$1 AND banned_user_id=$2`, community, target).Scan(&bans); err != nil || bans != 1 {
		t.Fatalf("ban removed despite audit failure: bans=%d err=%v", bans, err)
	}

	if err := writer.UnbanWithAudit(ctx, community, target, banAudit(moderator, community, "unban_user")); err != nil {
		t.Fatal(err)
	}
	var audits int
	if err := pool.QueryRow(ctx, `SELECT
		(SELECT COUNT(*) FROM community_bans WHERE community_id=$1 AND banned_user_id=$2),
		(SELECT COUNT(*) FROM moderation_audit_logs WHERE channel_id=$1 AND action='unban_user')`, community, target).Scan(&bans, &audits); err != nil {
		t.Fatal(err)
	}
	if bans != 0 || audits != 1 {
		t.Fatalf("bans=%d unban audits=%d", bans, audits)
	}
	if err := writer.UnbanWithAudit(ctx, community, target, banAudit(moderator, community, "unban_user")); !errors.Is(err, ErrCommunityBanNotFound) {
		t.Fatalf("second unban err=%v, want ErrCommunityBanNotFound", err)
	}
}
