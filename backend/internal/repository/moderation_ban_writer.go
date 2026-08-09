package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ModerationBanWriter persists a community ban and its audit record atomically.
type ModerationBanWriter struct {
	db *pgxpool.Pool
}

func NewModerationBanWriter(db *pgxpool.Pool) *ModerationBanWriter {
	return &ModerationBanWriter{db: db}
}

func (w *ModerationBanWriter) CreateBanWithAudit(ctx context.Context, ban *models.CommunityBan, audit *models.ModerationAuditLog) error {
	tx, err := w.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin moderation ban transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err = tx.Exec(ctx, "DELETE FROM community_members WHERE community_id = $1 AND user_id = $2", ban.CommunityID, ban.BannedUserID); err != nil {
		return fmt.Errorf("remove banned community member: %w", err)
	}
	if err = tx.QueryRow(ctx, `
		INSERT INTO community_bans (id, community_id, banned_user_id, banned_by_user_id, reason, banned_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, banned_at
	`, ban.ID, ban.CommunityID, ban.BannedUserID, ban.BannedByUserID, ban.Reason, ban.BannedAt).Scan(&ban.ID, &ban.BannedAt); err != nil {
		return fmt.Errorf("insert community ban: %w", err)
	}

	metadata, err := json.Marshal(audit.Metadata)
	if err != nil {
		return fmt.Errorf("marshal moderation audit metadata: %w", err)
	}
	if _, err = tx.Exec(ctx, `
		INSERT INTO moderation_audit_logs (
			id, action, entity_type, entity_id, moderator_id, actor_id, reason, metadata,
			ip_address, user_agent, channel_id, created_at
		) VALUES ($1, $2, $3, $4, $5, $5, $6, $7, $8, $9, $10, $11)
	`, audit.ID, audit.Action, audit.EntityType, audit.EntityID, audit.ModeratorID, audit.Reason,
		metadata, audit.IPAddress, audit.UserAgent, audit.ChannelID, audit.CreatedAt.UTC()); err != nil {
		return fmt.Errorf("insert moderation audit log: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit moderation ban transaction: %w", err)
	}
	return nil
}
