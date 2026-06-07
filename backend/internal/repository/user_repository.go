package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

var (
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")
	// ErrUserAlreadyExists is returned when trying to create a duplicate user
	ErrUserAlreadyExists = errors.New("user already exists")
	// ErrBlockNotFound is returned when a block relationship is not found
	ErrBlockNotFound = errors.New("block not found")
)

// UserRepository handles user database operations
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	// Set default account_type if not specified
	if user.AccountType == "" {
		user.AccountType = models.AccountTypeMember
	}

	// Set default account_status if not specified
	if user.AccountStatus == "" {
		user.AccountStatus = "active"
	}

	query := `
		INSERT INTO users (
			id, twitch_id, username, display_name, email,
			avatar_url, bio, role, account_type, account_status, last_login_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRow(
		ctx, query,
		user.ID, user.TwitchID, user.Username, user.DisplayName, user.Email,
		user.AvatarURL, user.Bio, user.Role, user.AccountType, user.AccountStatus, user.LastLoginAt,
	).Scan(&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return err
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT
			id, twitch_id, username, display_name, email, avatar_url, bio,
			karma_points, role, account_type, is_banned, created_at, updated_at, last_login_at,
			COALESCE(moderator_scope, '') AS moderator_scope,
			COALESCE(moderation_channels, '{}'::uuid[]) AS moderation_channels,
			moderation_started_at
		FROM users
		WHERE id = $1
	`

	var user models.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.TwitchID, &user.Username, &user.DisplayName, &user.Email,
		&user.AvatarURL, &user.Bio, &user.KarmaPoints, &user.Role, &user.AccountType, &user.IsBanned,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
		&user.ModeratorScope, &user.ModerationChannels, &user.ModerationStartedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// GetByTwitchID retrieves a user by Twitch ID
func (r *UserRepository) GetByTwitchID(ctx context.Context, twitchID string) (*models.User, error) {
	query := `
		SELECT
			id, twitch_id, username, display_name, email, avatar_url, bio,
			karma_points, role, account_type, is_banned, created_at, updated_at, last_login_at,
			COALESCE(moderator_scope, '') AS moderator_scope,
			COALESCE(moderation_channels, '{}'::uuid[]) AS moderation_channels,
			moderation_started_at
		FROM users
		WHERE twitch_id = $1
	`

	var user models.User
	err := r.db.QueryRow(ctx, query, twitchID).Scan(
		&user.ID, &user.TwitchID, &user.Username, &user.DisplayName, &user.Email,
		&user.AvatarURL, &user.Bio, &user.KarmaPoints, &user.Role, &user.AccountType, &user.IsBanned,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
		&user.ModeratorScope, &user.ModerationChannels, &user.ModerationStartedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// GetByUsername retrieves a user by username (case-insensitive)
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT
			id, twitch_id, username, display_name, email, avatar_url, bio,
			karma_points, role, account_type, is_banned, created_at, updated_at, last_login_at,
			COALESCE(moderator_scope, '') AS moderator_scope,
			COALESCE(moderation_channels, '{}'::uuid[]) AS moderation_channels,
			moderation_started_at
		FROM users
		WHERE LOWER(username) = LOWER($1)
	`

	var user models.User
	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.TwitchID, &user.Username, &user.DisplayName, &user.Email,
		&user.AvatarURL, &user.Bio, &user.KarmaPoints, &user.Role, &user.AccountType, &user.IsBanned,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
		&user.ModeratorScope, &user.ModerationChannels, &user.ModerationStartedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// GetByIDs retrieves multiple users by their IDs in a single query
func (r *UserRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.User, error) {
	if len(ids) == 0 {
		return []*models.User{}, nil
	}

	query := `
		SELECT
			id, twitch_id, username, display_name, email, avatar_url, bio,
			karma_points, role, account_type, is_banned, created_at, updated_at, last_login_at,
			COALESCE(moderator_scope, '') AS moderator_scope,
			COALESCE(moderation_channels, '{}'::uuid[]) AS moderation_channels,
			moderation_started_at
		FROM users
		WHERE id = ANY($1)
	`

	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]*models.User, 0, len(ids))
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.TwitchID, &user.Username, &user.DisplayName, &user.Email,
			&user.AvatarURL, &user.Bio, &user.KarmaPoints, &user.Role, &user.AccountType, &user.IsBanned,
			&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
			&user.ModeratorScope, &user.ModerationChannels, &user.ModerationStartedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// Update updates an existing user
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET username = $2, display_name = $3, email = $4, avatar_url = $5,
		    bio = $6, last_login_at = $7, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(
		ctx, query,
		user.ID, user.Username, user.DisplayName, user.Email,
		user.AvatarURL, user.Bio, user.LastLoginAt,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdateProfile updates user's display name and bio
func (r *UserRepository) UpdateProfile(ctx context.Context, userID uuid.UUID, displayName string, bio *string) error {
	query := `
		UPDATE users
		SET display_name = $2, bio = $3, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, userID, displayName, bio)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdateLastLogin updates the user's last login timestamp
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users
		SET last_login_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, userID)
	return err
}

// UpdateAccountStatus updates a user's account status
func (r *UserRepository) UpdateAccountStatus(ctx context.Context, userID uuid.UUID, status string) error {
	query := `
		UPDATE users
		SET account_status = $2, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, userID, status)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdateDisplayName updates a user's display name
func (r *UserRepository) UpdateDisplayName(ctx context.Context, userID uuid.UUID, displayName string) error {
	query := `
		UPDATE users
		SET display_name = $2, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, userID, displayName)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdateKarma updates a user's karma points
func (r *UserRepository) UpdateKarma(ctx context.Context, userID uuid.UUID, delta int) error {
	query := `
		UPDATE users
		SET karma_points = karma_points + $2
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, userID, delta)
	return err
}

// BanUser bans a user
func (r *UserRepository) BanUser(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users
		SET is_banned = true, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, userID)
	return err
}

// UnbanUser unbans a user
func (r *UserRepository) UnbanUser(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users
		SET is_banned = false, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, userID)
	return err
}

// RefreshTokenRepository handles refresh token database operations
type RefreshTokenRepository struct {
	db *pgxpool.Pool
}

// NewRefreshTokenRepository creates a new refresh token repository
func NewRefreshTokenRepository(db *pgxpool.Pool) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Create creates a new refresh token
func (r *RefreshTokenRepository) Create(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.Exec(ctx, query, userID, tokenHash, expiresAt)
	return err
}

// GetByHash retrieves a refresh token by its hash
func (r *RefreshTokenRepository) GetByHash(ctx context.Context, tokenHash string) (userID uuid.UUID, expiresAt time.Time, isRevoked bool, err error) {
	query := `
		SELECT user_id, expires_at, is_revoked
		FROM refresh_tokens
		WHERE token_hash = $1
	`

	err = r.db.QueryRow(ctx, query, tokenHash).Scan(&userID, &expiresAt, &isRevoked)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, time.Time{}, false, errors.New("refresh token not found")
		}
		return uuid.Nil, time.Time{}, false, err
	}

	return userID, expiresAt, isRevoked, nil
}

// Revoke marks a refresh token as revoked
func (r *RefreshTokenRepository) Revoke(ctx context.Context, tokenHash string) error {
	query := `
		UPDATE refresh_tokens
		SET is_revoked = true, revoked_at = NOW()
		WHERE token_hash = $1
	`

	_, err := r.db.Exec(ctx, query, tokenHash)
	return err
}

// RevokeAllForUser revokes all refresh tokens for a user
func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE refresh_tokens
		SET is_revoked = true, revoked_at = NOW()
		WHERE user_id = $1 AND is_revoked = false
	`

	_, err := r.db.Exec(ctx, query, userID)
	return err
}

// GetAllActiveUserIDs retrieves all user IDs that are not banned
func (r *UserRepository) GetAllActiveUserIDs(ctx context.Context) ([]uuid.UUID, error) {
	query := `
		SELECT id FROM users
		WHERE is_banned = false
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, id)
	}

	return userIDs, rows.Err()
}

// DeleteExpired deletes expired refresh tokens
func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE expires_at < NOW() - INTERVAL '7 days'
	`

	_, err := r.db.Exec(ctx, query)
	return err
}

// UserSettingsRepository handles user settings database operations
type UserSettingsRepository struct {
	db *pgxpool.Pool
}

// NewUserSettingsRepository creates a new user settings repository
func NewUserSettingsRepository(db *pgxpool.Pool) *UserSettingsRepository {
	return &UserSettingsRepository{db: db}
}

// GetByUserID retrieves user settings by user ID
func (r *UserSettingsRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.UserSettings, error) {
	query := `
		SELECT user_id, profile_visibility, show_karma_publicly, created_at, updated_at
		FROM user_settings
		WHERE user_id = $1
	`

	var settings models.UserSettings
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&settings.UserID, &settings.ProfileVisibility, &settings.ShowKarmaPublicly,
		&settings.CreatedAt, &settings.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &settings, nil
}

// Update updates user settings
func (r *UserSettingsRepository) Update(ctx context.Context, userID uuid.UUID, profileVisibility *string, showKarmaPublicly *bool) error {
	// Build query based on which fields are provided
	var query string
	var args []interface{}

	if profileVisibility != nil && showKarmaPublicly != nil {
		query = `UPDATE user_settings SET profile_visibility = $2, show_karma_publicly = $3, updated_at = NOW() WHERE user_id = $1`
		args = []interface{}{userID, *profileVisibility, *showKarmaPublicly}
	} else if profileVisibility != nil {
		query = `UPDATE user_settings SET profile_visibility = $2, updated_at = NOW() WHERE user_id = $1`
		args = []interface{}{userID, *profileVisibility}
	} else if showKarmaPublicly != nil {
		query = `UPDATE user_settings SET show_karma_publicly = $2, updated_at = NOW() WHERE user_id = $1`
		args = []interface{}{userID, *showKarmaPublicly}
	} else {
		// Nothing to update
		return nil
	}

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

// AccountDeletionRepository handles account deletion database operations
type AccountDeletionRepository struct {
	db *pgxpool.Pool
}

// NewAccountDeletionRepository creates a new account deletion repository
func NewAccountDeletionRepository(db *pgxpool.Pool) *AccountDeletionRepository {
	return &AccountDeletionRepository{db: db}
}

// Create creates a new account deletion request
func (r *AccountDeletionRepository) Create(ctx context.Context, deletion *models.AccountDeletion) error {
	query := `
		INSERT INTO account_deletions (id, user_id, scheduled_for, reason)
		VALUES ($1, $2, $3, $4)
		RETURNING requested_at, is_cancelled
	`

	err := r.db.QueryRow(
		ctx, query,
		deletion.ID, deletion.UserID, deletion.ScheduledFor, deletion.Reason,
	).Scan(&deletion.RequestedAt, &deletion.IsCancelled)

	return err
}

// GetPendingByUserID retrieves a pending deletion request for a user
func (r *AccountDeletionRepository) GetPendingByUserID(ctx context.Context, userID uuid.UUID) (*models.AccountDeletion, error) {
	query := `
		SELECT id, user_id, requested_at, scheduled_for, reason, is_cancelled, cancelled_at, completed_at
		FROM account_deletions
		WHERE user_id = $1 AND is_cancelled = false AND completed_at IS NULL
		ORDER BY requested_at DESC
		LIMIT 1
	`

	var deletion models.AccountDeletion
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&deletion.ID, &deletion.UserID, &deletion.RequestedAt, &deletion.ScheduledFor,
		&deletion.Reason, &deletion.IsCancelled, &deletion.CancelledAt, &deletion.CompletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // No pending deletion
		}
		return nil, err
	}

	return &deletion, nil
}

// Cancel cancels a deletion request
func (r *AccountDeletionRepository) Cancel(ctx context.Context, deletionID uuid.UUID) error {
	query := `
		UPDATE account_deletions
		SET is_cancelled = true, cancelled_at = NOW()
		WHERE id = $1 AND is_cancelled = false AND completed_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, deletionID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("deletion request not found or already completed")
	}

	return nil
}

// GetScheduledDeletions retrieves all scheduled deletions that are ready to be executed
func (r *AccountDeletionRepository) GetScheduledDeletions(ctx context.Context) ([]*models.AccountDeletion, error) {
	query := `
		SELECT id, user_id, requested_at, scheduled_for, reason, is_cancelled, cancelled_at, completed_at
		FROM account_deletions
		WHERE scheduled_for <= NOW() AND is_cancelled = false AND completed_at IS NULL
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deletions []*models.AccountDeletion
	for rows.Next() {
		var deletion models.AccountDeletion
		err := rows.Scan(
			&deletion.ID, &deletion.UserID, &deletion.RequestedAt, &deletion.ScheduledFor,
			&deletion.Reason, &deletion.IsCancelled, &deletion.CancelledAt, &deletion.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		deletions = append(deletions, &deletion)
	}

	return deletions, rows.Err()
}

// MarkCompleted marks a deletion as completed
func (r *AccountDeletionRepository) MarkCompleted(ctx context.Context, deletionID uuid.UUID) error {
	query := `
		UPDATE account_deletions
		SET completed_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, deletionID)
	return err
}

// UpdateDeviceToken updates a user's device token and platform
func (r *UserRepository) UpdateDeviceToken(ctx context.Context, userID uuid.UUID, deviceToken string, devicePlatform string) error {
	query := `
		UPDATE users
		SET device_token = $1, device_platform = $2, updated_at = NOW()
		WHERE id = $3
	`

	_, err := r.db.Exec(ctx, query, deviceToken, devicePlatform, userID)
	return err
}

// ClearDeviceToken clears a user's device token
func (r *UserRepository) ClearDeviceToken(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users
		SET device_token = NULL, device_platform = NULL, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, userID)
	return err
}

// GetUserProfile retrieves a complete user profile with stats
func (r *UserRepository) GetUserProfile(ctx context.Context, userID uuid.UUID, currentUserID *uuid.UUID) (*models.UserProfile, error) {
	// First get basic user info
	query := `
		SELECT
			u.id, u.twitch_id, u.username, u.display_name, u.email, u.avatar_url, u.bio,
			u.social_links, u.karma_points, u.trust_score, u.trust_score_updated_at,
			u.role, u.account_type, u.is_banned, u.follower_count, u.following_count,
			u.created_at, u.updated_at, u.last_login_at
		FROM users u
		WHERE u.id = $1
	`

	var profile models.UserProfile
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&profile.ID, &profile.TwitchID, &profile.Username, &profile.DisplayName, &profile.Email,
		&profile.AvatarURL, &profile.Bio, &profile.SocialLinks, &profile.KarmaPoints,
		&profile.TrustScore, &profile.TrustScoreUpdatedAt, &profile.Role, &profile.AccountType, &profile.IsBanned,
		&profile.FollowerCount, &profile.FollowingCount,
		&profile.CreatedAt, &profile.UpdatedAt, &profile.LastLoginAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Check if current user is following this user
	if currentUserID != nil {
		followQuery := `SELECT EXISTS(SELECT 1 FROM user_follows WHERE follower_id = $1 AND following_id = $2)`
		err = r.db.QueryRow(ctx, followQuery, *currentUserID, userID).Scan(&profile.IsFollowing)
		if err != nil {
			return nil, err
		}

		followBackQuery := `SELECT EXISTS(SELECT 1 FROM user_follows WHERE follower_id = $1 AND following_id = $2)`
		err = r.db.QueryRow(ctx, followBackQuery, userID, *currentUserID).Scan(&profile.IsFollowedBy)
		if err != nil {
			return nil, err
		}
	}

	// Get additional stats - using optimized queries with conditional aggregation
	statsQuery := `
		SELECT
			COALESCE(clips.total_count, 0) AS clips_submitted,
			COALESCE(clips.total_upvotes, 0) AS total_upvotes,
			COALESCE(clips.featured_count, 0) AS clips_featured,
			COALESCE((SELECT COUNT(*) FROM comments WHERE user_id = $1), 0) AS total_comments,
			COALESCE((SELECT COUNT(*) FROM broadcaster_follows WHERE user_id = $1), 0) AS broadcasters_followed
		FROM (
			SELECT
				COUNT(*) AS total_count,
				COALESCE(SUM(vote_score), 0) AS total_upvotes,
				COUNT(*) FILTER (WHERE is_featured = true) AS featured_count
			FROM clips
			WHERE submitted_by_user_id = $1
		) clips
	`

	err = r.db.QueryRow(ctx, statsQuery, userID).Scan(
		&profile.Stats.ClipsSubmitted,
		&profile.Stats.TotalUpvotes,
		&profile.Stats.TotalComments,
		&profile.Stats.ClipsFeatured,
		&profile.Stats.BroadcastersFollowed,
	)

	if err != nil {
		return nil, err
	}

	return &profile, nil
}

// UpdateSocialLinks updates a user's social media links
func (r *UserRepository) UpdateSocialLinks(ctx context.Context, userID uuid.UUID, socialLinks string) error {
	query := `
		UPDATE users
		SET social_links = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.db.Exec(ctx, query, socialLinks, userID)
	return err
}

// FollowUser creates a follow relationship between two users
func (r *UserRepository) FollowUser(ctx context.Context, followerID, followingID uuid.UUID) error {
	query := `
		INSERT INTO user_follows (follower_id, following_id)
		VALUES ($1, $2)
		ON CONFLICT (follower_id, following_id) DO NOTHING
	`

	_, err := r.db.Exec(ctx, query, followerID, followingID)
	return err
}

// ErrNotFollowing is returned when trying to unfollow a user that is not being followed
var ErrNotFollowing = errors.New("not following this user")

// UnfollowUser removes a follow relationship between two users
func (r *UserRepository) UnfollowUser(ctx context.Context, followerID, followingID uuid.UUID) error {
	query := `
		DELETE FROM user_follows
		WHERE follower_id = $1 AND following_id = $2
	`

	result, err := r.db.Exec(ctx, query, followerID, followingID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFollowing
	}

	return nil
}

// GetFollowers retrieves a list of users who follow the specified user
func (r *UserRepository) GetFollowers(ctx context.Context, userID uuid.UUID, currentUserID *uuid.UUID, limit, offset int) ([]models.FollowerUser, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM user_follows WHERE following_id = $1`
	var total int
	err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get followers with follow status in a single query
	var query string
	var rows pgx.Rows

	if currentUserID != nil {
		query = `
			SELECT
				u.id, u.username, u.display_name, u.avatar_url, u.bio, u.karma_points, uf.created_at,
				EXISTS(SELECT 1 FROM user_follows WHERE follower_id = $2 AND following_id = u.id) AS is_following
			FROM user_follows uf
			JOIN users u ON u.id = uf.follower_id
			WHERE uf.following_id = $1
			ORDER BY uf.created_at DESC
			LIMIT $3 OFFSET $4
		`
		rows, err = r.db.Query(ctx, query, userID, *currentUserID, limit, offset)
	} else {
		query = `
			SELECT
				u.id, u.username, u.display_name, u.avatar_url, u.bio, u.karma_points, uf.created_at,
				false AS is_following
			FROM user_follows uf
			JOIN users u ON u.id = uf.follower_id
			WHERE uf.following_id = $1
			ORDER BY uf.created_at DESC
			LIMIT $2 OFFSET $3
		`
		rows, err = r.db.Query(ctx, query, userID, limit, offset)
	}

	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	followers := []models.FollowerUser{}
	for rows.Next() {
		var follower models.FollowerUser
		err := rows.Scan(
			&follower.ID, &follower.Username, &follower.DisplayName, &follower.AvatarURL,
			&follower.Bio, &follower.KarmaPoints, &follower.FollowedAt, &follower.IsFollowing,
		)
		if err != nil {
			return nil, 0, err
		}

		followers = append(followers, follower)
	}

	return followers, total, nil
}

// GetFollowing retrieves a list of users that the specified user follows
func (r *UserRepository) GetFollowing(ctx context.Context, userID uuid.UUID, currentUserID *uuid.UUID, limit, offset int) ([]models.FollowerUser, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM user_follows WHERE follower_id = $1`
	var total int
	err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get following with follow status in a single query
	var query string
	var rows pgx.Rows

	if currentUserID != nil {
		query = `
			SELECT
				u.id, u.username, u.display_name, u.avatar_url, u.bio, u.karma_points, uf.created_at,
				EXISTS(SELECT 1 FROM user_follows WHERE follower_id = $2 AND following_id = u.id) AS is_following
			FROM user_follows uf
			JOIN users u ON u.id = uf.following_id
			WHERE uf.follower_id = $1
			ORDER BY uf.created_at DESC
			LIMIT $3 OFFSET $4
		`
		rows, err = r.db.Query(ctx, query, userID, *currentUserID, limit, offset)
	} else {
		query = `
			SELECT
				u.id, u.username, u.display_name, u.avatar_url, u.bio, u.karma_points, uf.created_at,
				false AS is_following
			FROM user_follows uf
			JOIN users u ON u.id = uf.following_id
			WHERE uf.follower_id = $1
			ORDER BY uf.created_at DESC
			LIMIT $2 OFFSET $3
		`
		rows, err = r.db.Query(ctx, query, userID, limit, offset)
	}

	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	following := []models.FollowerUser{}
	for rows.Next() {
		var user models.FollowerUser
		err := rows.Scan(
			&user.ID, &user.Username, &user.DisplayName, &user.AvatarURL,
			&user.Bio, &user.KarmaPoints, &user.FollowedAt, &user.IsFollowing,
		)
		if err != nil {
			return nil, 0, err
		}

		following = append(following, user)
	}

	return following, total, nil
}

// GetUserActivity retrieves a user's activity feed
func (r *UserRepository) GetUserActivity(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.UserActivityItem, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM user_activity WHERE user_id = $1`
	var total int
	err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get activity items
	query := `
		SELECT
			ua.id, ua.user_id, ua.activity_type, ua.target_id, ua.target_type,
			ua.metadata, ua.created_at,
			u.username, u.avatar_url,
			c.title as clip_title, c.id as clip_id,
			co.content as comment_text,
			u2.username as target_user
		FROM user_activity ua
		JOIN users u ON u.id = ua.user_id
		LEFT JOIN clips c ON c.id = ua.target_id AND ua.target_type = 'clip'
		LEFT JOIN comments co ON co.id = ua.target_id AND ua.target_type = 'comment'
		LEFT JOIN users u2 ON u2.id = ua.target_id AND ua.target_type = 'user'
		WHERE ua.user_id = $1
		ORDER BY ua.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	activities := []models.UserActivityItem{}
	for rows.Next() {
		var activity models.UserActivityItem
		var clipID *uuid.UUID
		err := rows.Scan(
			&activity.ID, &activity.UserID, &activity.ActivityType, &activity.TargetID,
			&activity.TargetType, &activity.Metadata, &activity.CreatedAt,
			&activity.Username, &activity.UserAvatar,
			&activity.ClipTitle, &clipID,
			&activity.CommentText,
			&activity.TargetUser,
		)
		if err != nil {
			return nil, 0, err
		}

		if clipID != nil {
			clipIDStr := clipID.String()
			activity.ClipID = &clipIDStr
		}

		activities = append(activities, activity)
	}

	return activities, total, nil
}

// CreateUserActivity records a user activity
func (r *UserRepository) CreateUserActivity(ctx context.Context, activity *models.UserActivity) error {
	query := `
		INSERT INTO user_activity (id, user_id, activity_type, target_id, target_type, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at
	`

	err := r.db.QueryRow(
		ctx, query,
		activity.ID, activity.UserID, activity.ActivityType, activity.TargetID,
		activity.TargetType, activity.Metadata,
	).Scan(&activity.CreatedAt)

	return err
}

// UpdateAccountType updates a user's account type
func (r *UserRepository) UpdateAccountType(ctx context.Context, userID uuid.UUID, accountType string) error {
	query := `
		UPDATE users
		SET account_type = $2,
		    account_type_updated_at = NOW(),
		    updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, userID, accountType)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

// GetAccountType retrieves a user's current account type
func (r *UserRepository) GetAccountType(ctx context.Context, userID uuid.UUID) (string, error) {
	query := `
		SELECT account_type
		FROM users
		WHERE id = $1
	`

	var accountType string
	err := r.db.QueryRow(ctx, query, userID).Scan(&accountType)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrUserNotFound
		}
		return "", err
	}

	return accountType, nil
}

// BlockUser creates a block relationship between two users
func (r *UserRepository) BlockUser(ctx context.Context, userID, blockedUserID uuid.UUID) error {
	query := `
		INSERT INTO user_blocks (user_id, blocked_user_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, blocked_user_id) DO NOTHING
	`

	_, err := r.db.Exec(ctx, query, userID, blockedUserID)
	return err
}

// UnblockUser removes a block relationship between two users
func (r *UserRepository) UnblockUser(ctx context.Context, userID, blockedUserID uuid.UUID) error {
	query := `
		DELETE FROM user_blocks
		WHERE user_id = $1 AND blocked_user_id = $2
	`

	commandTag, err := r.db.Exec(ctx, query, userID, blockedUserID)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return ErrUserNotFound // Reusing this error for simplicity
	}

	return nil
}

// IsBlocked checks if userID has blocked blockedUserID
func (r *UserRepository) IsBlocked(ctx context.Context, userID, blockedUserID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM user_blocks
			WHERE user_id = $1 AND blocked_user_id = $2
		)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, userID, blockedUserID).Scan(&exists)
	return exists, err
}

// GetBlockedUsers retrieves users blocked by the specified user
func (r *UserRepository) GetBlockedUsers(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.BlockedUser, int, error) {
	// Get blocked users with their info
	query := `
		SELECT
			u.id, u.username, u.display_name, u.avatar_url, u.bio, u.karma_points,
			ub.blocked_at
		FROM user_blocks ub
		JOIN users u ON u.id = ub.blocked_user_id
		WHERE ub.user_id = $1
		ORDER BY ub.blocked_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var blockedUsers []models.BlockedUser
	for rows.Next() {
		var user models.BlockedUser
		err := rows.Scan(
			&user.ID, &user.Username, &user.DisplayName, &user.AvatarURL,
			&user.Bio, &user.KarmaPoints, &user.BlockedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		blockedUsers = append(blockedUsers, user)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM user_blocks WHERE user_id = $1`
	err = r.db.QueryRow(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return blockedUsers, total, nil
}

// AdminSearchUsers searches users with filtering for admin dashboard
func (r *UserRepository) AdminSearchUsers(ctx context.Context, searchQuery string, role string, status string, limit, offset int) ([]*models.User, int, error) {
	// Ensure limit is within a safe range to avoid excessive result sets
	if limit < 1 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}

	// Build dynamic query based on filters
	baseQuery := `
		SELECT
			id, twitch_id, username, display_name, email, avatar_url, bio,
			karma_points, role, account_type, is_banned, account_status, created_at, updated_at, last_login_at
		FROM users
		WHERE 1=1
	`
	countQuery := `SELECT COUNT(*) FROM users WHERE 1=1`

	args := []interface{}{}
	argNum := 1
	whereClause := ""

	// Add search filter
	if searchQuery != "" {
		whereClause += ` AND (LOWER(username) LIKE LOWER($` + strconv.Itoa(argNum) + `)
			OR LOWER(display_name) LIKE LOWER($` + strconv.Itoa(argNum) + `)
			OR LOWER(email) LIKE LOWER($` + strconv.Itoa(argNum) + `))`
		args = append(args, "%"+searchQuery+"%")
		argNum++
	}

	// Add role filter
	if role != "" && role != "all" {
		whereClause += ` AND role = $` + strconv.Itoa(argNum)
		args = append(args, role)
		argNum++
	}

	// Add status filter
	if status != "" && status != "all" {
		if status == "banned" {
			whereClause += ` AND is_banned = true`
		} else if status == "active" {
			whereClause += ` AND is_banned = false`
		} else if status == "unclaimed" {
			whereClause += ` AND account_status = 'unclaimed'`
		}
	}

	// Get total count
	var total int
	err := r.db.QueryRow(ctx, countQuery+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Add pagination
	paginationArgs := append(args, limit, offset)
	finalQuery := baseQuery + whereClause + ` ORDER BY created_at DESC LIMIT $` + strconv.Itoa(argNum) + ` OFFSET $` + strconv.Itoa(argNum+1)

	rows, err := r.db.Query(ctx, finalQuery, paginationArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := []*models.User{}
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.TwitchID, &user.Username, &user.DisplayName, &user.Email,
			&user.AvatarURL, &user.Bio, &user.KarmaPoints, &user.Role, &user.AccountType,
			&user.IsBanned, &user.AccountStatus, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
		)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// SearchUsersForAutocomplete searches users by username prefix for chat mentions and autocomplete
// Returns up to 'limit' users whose username starts with the query string
func (r *UserRepository) SearchUsersForAutocomplete(ctx context.Context, query string, limit int) ([]*models.User, error) {
	// Ensure limit is within a safe range
	if limit < 1 {
		limit = 10
	} else if limit > 20 {
		limit = 20
	}

	// Only search if query is non-empty
	if query == "" {
		return []*models.User{}, nil
	}

	// Note: This query uses LOWER() which prevents index usage on standard btree indexes.
	// For better performance with large user tables, consider:
	// 1. Creating a functional index: CREATE INDEX idx_users_username_lower ON users(LOWER(username))
	// 2. Using PostgreSQL's citext extension for case-insensitive username column
	// 3. Using pg_trgm extension with GIN index for fuzzy matching
	searchQuery := `
		SELECT
			id, username, display_name, avatar_url, is_verified
		FROM users
		WHERE LOWER(username) LIKE LOWER($1)
			AND is_banned = false
			AND account_status = 'active'
		ORDER BY
			CASE WHEN is_verified THEN 0 ELSE 1 END,
			LENGTH(username),
			username
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, searchQuery, query+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.DisplayName,
			&user.AvatarURL,
			&user.IsVerified,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// UpdateUserRole updates a user's role (user, moderator, admin)
func (r *UserRepository) UpdateUserRole(ctx context.Context, userID uuid.UUID, role string) error {
	// Validate role before updating
	if role != "user" && role != "moderator" && role != "admin" {
		return errors.New("invalid role: must be user, moderator, or admin")
	}

	query := `
		UPDATE users
		SET role = $2, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, userID, role)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

// SetUserKarma sets a user's karma points to a specific value (admin override)
func (r *UserRepository) SetUserKarma(ctx context.Context, userID uuid.UUID, karma int) error {
	query := `
		UPDATE users
		SET karma_points = $2, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, userID, karma)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

// SuspendCommentPrivileges suspends a user's comment privileges
func (r *UserRepository) SuspendCommentPrivileges(
	ctx context.Context,
	userID uuid.UUID,
	suspendedBy uuid.UUID,
	suspensionType string,
	reason string,
	durationHours *int,
) error {
	// Validate suspension type before performing any database operations
	if suspensionType != models.SuspensionTypeWarning &&
		suspensionType != models.SuspensionTypeTemporary &&
		suspensionType != models.SuspensionTypePermanent {
		return fmt.Errorf("invalid suspension type: must be warning, temporary, or permanent")
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Calculate expiration time for temporary suspensions
	var expiresAt *time.Time
	var suspendedUntil *time.Time
	if suspensionType == models.SuspensionTypeTemporary && durationHours != nil {
		expiry := time.Now().Add(time.Duration(*durationHours) * time.Hour)
		expiresAt = &expiry
		suspendedUntil = &expiry
	}

	// Update user record
	var updateQuery string
	if suspensionType == models.SuspensionTypePermanent {
		// For permanent suspensions, set to far future date (year 9999)
		permanentDate := time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
		updateQuery = `
			UPDATE users
			SET comment_suspended_until = $2,
				updated_at = NOW()
			WHERE id = $1
		`
		_, err = tx.Exec(ctx, updateQuery, userID, permanentDate)
	} else if suspensionType == models.SuspensionTypeTemporary {
		updateQuery = `
			UPDATE users
			SET comment_suspended_until = $2,
				updated_at = NOW()
			WHERE id = $1
		`
		_, err = tx.Exec(ctx, updateQuery, userID, suspendedUntil)
	} else {
		// Warning only - increment warning count
		updateQuery = `
			UPDATE users
			SET comment_warning_count = comment_warning_count + 1,
				updated_at = NOW()
			WHERE id = $1
		`
		_, err = tx.Exec(ctx, updateQuery, userID)
	}

	if err != nil {
		return err
	}

	// Insert into suspension history
	insertQuery := `
		INSERT INTO comment_suspension_history (
			id, user_id, suspended_by, suspension_type, reason,
			duration_hours, suspended_at, expires_at, is_active
		) VALUES ($1, $2, $3, $4, $5, $6, NOW(), $7, true)
	`

	_, err = tx.Exec(ctx, insertQuery,
		uuid.New(), userID, suspendedBy, suspensionType, reason,
		durationHours, expiresAt,
	)

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// LiftCommentSuspension lifts a user's comment suspension
func (r *UserRepository) LiftCommentSuspension(
	ctx context.Context,
	userID uuid.UUID,
	liftedBy uuid.UUID,
	reason string,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Clear suspension from user record
	updateQuery := `
		UPDATE users
		SET comment_suspended_until = NULL,
			updated_at = NOW()
		WHERE id = $1
	`
	result, err := tx.Exec(ctx, updateQuery, userID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	// Mark active suspensions as lifted in history
	historyQuery := `
		UPDATE comment_suspension_history
		SET is_active = false,
			lifted_at = NOW(),
			lifted_by = $2,
			lift_reason = $3
		WHERE user_id = $1 AND is_active = true
	`
	_, err = tx.Exec(ctx, historyQuery, userID, liftedBy, reason)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// GetCommentSuspensionHistory retrieves a user's comment suspension history
func (r *UserRepository) GetCommentSuspensionHistory(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
) ([]*models.CommentSuspensionHistory, error) {
	query := `
		SELECT
			id, user_id, suspended_by, suspension_type, reason,
			duration_hours, suspended_at, expires_at, is_active,
			lifted_at, lifted_by, lift_reason, metadata
		FROM comment_suspension_history
		WHERE user_id = $1
		ORDER BY suspended_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*models.CommentSuspensionHistory
	for rows.Next() {
		h := &models.CommentSuspensionHistory{}
		err := rows.Scan(
			&h.ID, &h.UserID, &h.SuspendedBy, &h.SuspensionType, &h.Reason,
			&h.DurationHours, &h.SuspendedAt, &h.ExpiresAt, &h.IsActive,
			&h.LiftedAt, &h.LiftedBy, &h.LiftReason, &h.Metadata,
		)
		if err != nil {
			return nil, err
		}
		history = append(history, h)
	}

	return history, rows.Err()
}

// SetCommentReviewRequirement sets whether a user's comments require review
func (r *UserRepository) SetCommentReviewRequirement(
	ctx context.Context,
	userID uuid.UUID,
	requireReview bool,
) error {
	query := `
		UPDATE users
		SET comments_require_review = $2, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, userID, requireReview)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

// CanUserComment checks if a user can post comments (not suspended)
func (r *UserRepository) CanUserComment(ctx context.Context, userID uuid.UUID) (bool, error) {
	query := `
		SELECT
			CASE
				WHEN is_banned THEN false
				WHEN comment_suspended_until IS NOT NULL AND comment_suspended_until > NOW() THEN false
				ELSE true
			END as can_comment
		FROM users
		WHERE id = $1
	`

	var canComment bool
	err := r.db.QueryRow(ctx, query, userID).Scan(&canComment)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, ErrUserNotFound
		}
		return false, err
	}

	return canComment, nil
}

// GetCommentSuspensionInfo returns suspension details for better error messaging
func (r *UserRepository) GetCommentSuspensionInfo(ctx context.Context, userID uuid.UUID) (*time.Time, error) {
	query := `
		SELECT comment_suspended_until
		FROM users
		WHERE id = $1
	`

	var suspendedUntil *time.Time
	err := r.db.QueryRow(ctx, query, userID).Scan(&suspendedUntil)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return suspendedUntil, nil
}

// DoesUserRequireCommentReview checks if user's comments need review
func (r *UserRepository) DoesUserRequireCommentReview(ctx context.Context, userID uuid.UUID) (bool, error) {
	query := `
		SELECT comments_require_review
		FROM users
		WHERE id = $1
	`

	var requireReview bool
	err := r.db.QueryRow(ctx, query, userID).Scan(&requireReview)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, ErrUserNotFound
		}
		return false, err
	}

	return requireReview, nil
}
