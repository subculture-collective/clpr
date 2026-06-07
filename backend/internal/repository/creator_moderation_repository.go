package repository

import (
	"context"
	"fmt"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type creatorModerationDB interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

// CreatorModerationRepository handles creator moderation persistence.
type CreatorModerationRepository struct {
	db creatorModerationDB
}

// NewCreatorModerationRepository creates a new CreatorModerationRepository.
func NewCreatorModerationRepository(pool *pgxpool.Pool) *CreatorModerationRepository {
	return &CreatorModerationRepository{db: pool}
}

const creatorModeratorSelectColumns = `id, creator_id, user_id, platform, platform_user_id, permissions, source, created_at`

const creatorBanSelectColumns = `id, creator_id, target_user_id, target_platform, target_platform_user_id, source, reason, scopes, expires_at, created_by_user_id, sync_status, created_at, updated_at`

func scanCreatorModerator(scanner interface{ Scan(...any) error }, moderator *models.CreatorModerator) error {
	return scanner.Scan(
		&moderator.ID,
		&moderator.CreatorID,
		&moderator.UserID,
		&moderator.Platform,
		&moderator.PlatformUserID,
		&moderator.Permissions,
		&moderator.Source,
		&moderator.CreatedAt,
	)
}

func scanCreatorBan(scanner interface{ Scan(...any) error }, ban *models.CreatorBan) error {
	return scanner.Scan(
		&ban.ID,
		&ban.CreatorID,
		&ban.TargetUserID,
		&ban.TargetPlatform,
		&ban.TargetPlatformUserID,
		&ban.Source,
		&ban.Reason,
		&ban.Scopes,
		&ban.ExpiresAt,
		&ban.CreatedByUserID,
		&ban.SyncStatus,
		&ban.CreatedAt,
		&ban.UpdatedAt,
	)
}

// CreateModerator inserts a creator moderator assignment.
func (r *CreatorModerationRepository) CreateModerator(ctx context.Context, moderator *models.CreatorModerator) error {
	query := `
		INSERT INTO creator_moderators (
			id, creator_id, user_id, platform, platform_user_id, permissions, source
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at
	`

	if err := r.db.QueryRow(ctx, query,
		moderator.ID,
		moderator.CreatorID,
		moderator.UserID,
		moderator.Platform,
		moderator.PlatformUserID,
		moderator.Permissions,
		moderator.Source,
	).Scan(&moderator.CreatedAt); err != nil {
		return fmt.Errorf("failed to create creator moderator: %w", err)
	}

	return nil
}

// ListModeratorsByCreator lists creator moderators ordered by creation time.
func (r *CreatorModerationRepository) ListModeratorsByCreator(ctx context.Context, creatorID uuid.UUID) ([]*models.CreatorModerator, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM creator_moderators
		WHERE creator_id = $1
		ORDER BY created_at DESC, id DESC
	`, creatorModeratorSelectColumns)

	rows, err := r.db.Query(ctx, query, creatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to list creator moderators: %w", err)
	}
	defer rows.Close()

	moderators := make([]*models.CreatorModerator, 0)
	for rows.Next() {
		var moderator models.CreatorModerator
		if err := scanCreatorModerator(rows, &moderator); err != nil {
			return nil, fmt.Errorf("failed to scan creator moderator: %w", err)
		}
		moderators = append(moderators, &moderator)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate creator moderators: %w", err)
	}

	return moderators, nil
}

// CreateBan inserts a creator ban.
func (r *CreatorModerationRepository) CreateBan(ctx context.Context, ban *models.CreatorBan) error {
	query := `
		INSERT INTO creator_bans (
			id, creator_id, target_user_id, target_platform, target_platform_user_id,
			source, reason, scopes, expires_at, created_by_user_id, sync_status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at, updated_at
	`

	if err := r.db.QueryRow(ctx, query,
		ban.ID,
		ban.CreatorID,
		ban.TargetUserID,
		ban.TargetPlatform,
		ban.TargetPlatformUserID,
		ban.Source,
		ban.Reason,
		ban.Scopes,
		ban.ExpiresAt,
		ban.CreatedByUserID,
		ban.SyncStatus,
	).Scan(&ban.CreatedAt, &ban.UpdatedAt); err != nil {
		return fmt.Errorf("failed to create creator ban: %w", err)
	}

	return nil
}

// GetActiveBanForUser retrieves an active ban for a local user and scope.
func (r *CreatorModerationRepository) GetActiveBanForUser(ctx context.Context, creatorID, targetUserID uuid.UUID, scope string) (*models.CreatorBan, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM creator_bans
		WHERE creator_id = $1
		  AND target_user_id = $2
		  AND scopes @> ARRAY[$3]::TEXT[]
		  AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY created_at DESC, id DESC
		LIMIT 1
	`, creatorBanSelectColumns)

	var ban models.CreatorBan
	if err := scanCreatorBan(r.db.QueryRow(ctx, query, creatorID, targetUserID, scope), &ban); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get active creator ban for user: %w", err)
	}

	return &ban, nil
}

// GetActiveBanForPlatformIdentity retrieves an active ban for a platform identity and scope.
func (r *CreatorModerationRepository) GetActiveBanForPlatformIdentity(ctx context.Context, creatorID uuid.UUID, platform, platformUserID, scope string) (*models.CreatorBan, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM creator_bans
		WHERE creator_id = $1
		  AND target_platform = $2
		  AND target_platform_user_id = $3
		  AND scopes @> ARRAY[$4]::TEXT[]
		  AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY created_at DESC, id DESC
		LIMIT 1
	`, creatorBanSelectColumns)

	var ban models.CreatorBan
	if err := scanCreatorBan(r.db.QueryRow(ctx, query, creatorID, platform, platformUserID, scope), &ban); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get active creator ban for platform identity: %w", err)
	}

	return &ban, nil
}
