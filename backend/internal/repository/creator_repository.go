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

type creatorDB interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

// CreatorRepository handles database operations for creator accounts.
type CreatorRepository struct {
	db creatorDB
}

// NewCreatorRepository creates a new CreatorRepository.
func NewCreatorRepository(pool *pgxpool.Pool) *CreatorRepository {
	return &CreatorRepository{db: pool}
}

const creatorAccountSelectColumns = `id, owner_user_id, display_name, slug, created_at, updated_at`

const creatorPlatformAccountSelectColumns = `id, creator_id, platform, platform_user_id, platform_display_name, profile_url,
	can_import_bans, can_sync_bans_outbound, can_import_moderators, can_verify_ownership,
	can_fetch_metadata, access_token_encrypted, refresh_token_encrypted, token_expires_at,
	created_at, updated_at`

func scanCreatorAccount(scanner interface{ Scan(...any) error }, account *models.CreatorAccount) error {
	return scanner.Scan(
		&account.ID,
		&account.OwnerUserID,
		&account.DisplayName,
		&account.Slug,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
}

func scanCreatorPlatformAccount(scanner interface{ Scan(...any) error }, account *models.CreatorPlatformAccount) error {
	return scanner.Scan(
		&account.ID,
		&account.CreatorID,
		&account.Platform,
		&account.PlatformUserID,
		&account.PlatformDisplayName,
		&account.ProfileURL,
		&account.CanImportBans,
		&account.CanSyncBansOutbound,
		&account.CanImportModerators,
		&account.CanVerifyOwnership,
		&account.CanFetchMetadata,
		&account.AccessTokenEncrypted,
		&account.RefreshTokenEncrypted,
		&account.TokenExpiresAt,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
}

// CreateCreatorAccount inserts a new creator account.
func (r *CreatorRepository) CreateCreatorAccount(ctx context.Context, account *models.CreatorAccount) error {
	query := fmt.Sprintf(`
		INSERT INTO creator_accounts (id, owner_user_id, display_name, slug)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at, updated_at
	`)

	if err := r.db.QueryRow(ctx, query, account.ID, account.OwnerUserID, account.DisplayName, account.Slug).Scan(&account.CreatedAt, &account.UpdatedAt); err != nil {
		return fmt.Errorf("failed to create creator account: %w", err)
	}

	return nil
}

// GetCreatorAccountByID retrieves a creator account by ID.
func (r *CreatorRepository) GetCreatorAccountByID(ctx context.Context, id uuid.UUID) (*models.CreatorAccount, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM creator_accounts
		WHERE id = $1
	`, creatorAccountSelectColumns)

	var account models.CreatorAccount
	if err := scanCreatorAccount(r.db.QueryRow(ctx, query, id), &account); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get creator account by id: %w", err)
	}

	return &account, nil
}

// GetCreatorAccountBySlug retrieves a creator account by slug.
func (r *CreatorRepository) GetCreatorAccountBySlug(ctx context.Context, slug string) (*models.CreatorAccount, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM creator_accounts
		WHERE slug = $1
	`, creatorAccountSelectColumns)

	var account models.CreatorAccount
	if err := scanCreatorAccount(r.db.QueryRow(ctx, query, slug), &account); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get creator account by slug: %w", err)
	}

	return &account, nil
}

// ListCreatorAccountsByOwner lists creator accounts for a given owner.
func (r *CreatorRepository) ListCreatorAccountsByOwner(ctx context.Context, ownerUserID uuid.UUID) ([]*models.CreatorAccount, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM creator_accounts
		WHERE owner_user_id = $1
		ORDER BY created_at DESC
	`, creatorAccountSelectColumns)

	rows, err := r.db.Query(ctx, query, ownerUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to list creator accounts by owner: %w", err)
	}
	defer rows.Close()

	accounts := make([]*models.CreatorAccount, 0)
	for rows.Next() {
		var account models.CreatorAccount
		if err := scanCreatorAccount(rows, &account); err != nil {
			return nil, fmt.Errorf("failed to scan creator account: %w", err)
		}
		accounts = append(accounts, &account)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate creator accounts: %w", err)
	}

	return accounts, nil
}

// CreateCreatorPlatformAccount inserts a platform account link for a creator.
func (r *CreatorRepository) CreateCreatorPlatformAccount(ctx context.Context, account *models.CreatorPlatformAccount) error {
	query := fmt.Sprintf(`
		INSERT INTO creator_platform_accounts (
			id, creator_id, platform, platform_user_id, platform_display_name,
			profile_url, can_import_bans, can_sync_bans_outbound, can_import_moderators,
			can_verify_ownership, can_fetch_metadata, access_token_encrypted,
			refresh_token_encrypted, token_expires_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12,
			$13, $14
		)
		RETURNING created_at, updated_at
	`)

	if err := r.db.QueryRow(ctx, query,
		account.ID,
		account.CreatorID,
		account.Platform,
		account.PlatformUserID,
		account.PlatformDisplayName,
		account.ProfileURL,
		account.CanImportBans,
		account.CanSyncBansOutbound,
		account.CanImportModerators,
		account.CanVerifyOwnership,
		account.CanFetchMetadata,
		account.AccessTokenEncrypted,
		account.RefreshTokenEncrypted,
		account.TokenExpiresAt,
	).Scan(&account.CreatedAt, &account.UpdatedAt); err != nil {
		return fmt.Errorf("failed to create creator platform account: %w", err)
	}

	return nil
}

// GetCreatorPlatformAccountByPlatformAndUserID retrieves a platform account by unique platform identity.
func (r *CreatorRepository) GetCreatorPlatformAccountByPlatformAndUserID(ctx context.Context, platform, platformUserID string) (*models.CreatorPlatformAccount, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM creator_platform_accounts
		WHERE platform = $1 AND platform_user_id = $2
	`, creatorPlatformAccountSelectColumns)

	var account models.CreatorPlatformAccount
	if err := scanCreatorPlatformAccount(r.db.QueryRow(ctx, query, platform, platformUserID), &account); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get creator platform account: %w", err)
	}

	return &account, nil
}

// ListCreatorPlatformAccounts lists platform links for a creator.
func (r *CreatorRepository) ListCreatorPlatformAccounts(ctx context.Context, creatorID uuid.UUID) ([]*models.CreatorPlatformAccount, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM creator_platform_accounts
		WHERE creator_id = $1
		ORDER BY created_at DESC
	`, creatorPlatformAccountSelectColumns)

	rows, err := r.db.Query(ctx, query, creatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to list creator platform accounts: %w", err)
	}
	defer rows.Close()

	accounts := make([]*models.CreatorPlatformAccount, 0)
	for rows.Next() {
		var account models.CreatorPlatformAccount
		if err := scanCreatorPlatformAccount(rows, &account); err != nil {
			return nil, fmt.Errorf("failed to scan creator platform account: %w", err)
		}
		accounts = append(accounts, &account)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate creator platform accounts: %w", err)
	}

	return accounts, nil
}
