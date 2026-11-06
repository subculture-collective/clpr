package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrCommunityBanNotFound is returned when an unban finds no ban to remove.
var ErrCommunityBanNotFound = errors.New("community ban not found")

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
	// A repeated ban updates the existing restriction (UNIQUE community_id,
	// banned_user_id), matching CommunityRepository.BanMemberAndRemoveMembership.
	if err = tx.QueryRow(ctx, `
		INSERT INTO community_bans (id, community_id, banned_user_id, banned_by_user_id, reason, banned_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (community_id, banned_user_id) DO UPDATE SET
			banned_by_user_id = EXCLUDED.banned_by_user_id,
			reason = EXCLUDED.reason,
			banned_at = EXCLUDED.banned_at
		RETURNING id, banned_at
	`, ban.ID, ban.CommunityID, ban.BannedUserID, ban.BannedByUserID, ban.Reason, ban.BannedAt).Scan(&ban.ID, &ban.BannedAt); err != nil {
		return fmt.Errorf("insert community ban: %w", err)
	}
	// On a repeated ban the stored row keeps its original id; audit that row.
	audit.EntityID = ban.ID

	if err = insertModerationAudit(ctx, tx, audit); err != nil {
		return err
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit moderation ban transaction: %w", err)
	}
	return nil
}

// UnbanWithAudit removes a community ban and records its audit entry in one
// transaction. It returns ErrCommunityBanNotFound when no ban exists.
func (w *ModerationBanWriter) UnbanWithAudit(ctx context.Context, communityID, userID uuid.UUID, audit *models.ModerationAuditLog) error {
	tx, err := w.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin moderation unban transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	result, err := tx.Exec(ctx, `DELETE FROM community_bans WHERE community_id = $1 AND banned_user_id = $2`, communityID, userID)
	if err != nil {
		return fmt.Errorf("delete community ban: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrCommunityBanNotFound
	}
	if err = insertModerationAudit(ctx, tx, audit); err != nil {
		return err
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit moderation unban transaction: %w", err)
	}
	return nil
}

func insertModerationAudit(ctx context.Context, tx pgx.Tx, audit *models.ModerationAuditLog) error {
	if audit.ID == uuid.Nil {
		audit.ID = uuid.New()
	}
	if audit.CreatedAt.IsZero() {
		audit.CreatedAt = time.Now()
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
	return nil
}
