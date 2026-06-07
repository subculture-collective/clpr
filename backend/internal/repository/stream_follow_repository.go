package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// StreamFollowRepository handles database operations for stream follows
type StreamFollowRepository struct {
	pool *pgxpool.Pool
}

// NewStreamFollowRepository creates a new stream follow repository
func NewStreamFollowRepository(pool *pgxpool.Pool) *StreamFollowRepository {
	return &StreamFollowRepository{pool: pool}
}

// FollowStreamer creates a follow relationship for a user and streamer
func (r *StreamFollowRepository) FollowStreamer(ctx context.Context, userID uuid.UUID, streamerUsername string, notificationsEnabled bool) (*models.StreamFollow, error) {
	query := `
		INSERT INTO stream_follows (user_id, streamer_username, notifications_enabled)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, streamer_username) 
		DO UPDATE SET notifications_enabled = EXCLUDED.notifications_enabled, updated_at = NOW()
		RETURNING id, user_id, streamer_username, notifications_enabled, created_at, updated_at
	`

	var follow models.StreamFollow
	err := r.pool.QueryRow(ctx, query, userID, streamerUsername, notificationsEnabled).Scan(
		&follow.ID,
		&follow.UserID,
		&follow.StreamerUsername,
		&follow.NotificationsEnabled,
		&follow.CreatedAt,
		&follow.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to follow streamer: %w", err)
	}

	return &follow, nil
}

// UnfollowStreamer removes a follow relationship
func (r *StreamFollowRepository) UnfollowStreamer(ctx context.Context, userID uuid.UUID, streamerUsername string) error {
	query := `
		DELETE FROM stream_follows
		WHERE user_id = $1 AND streamer_username = $2
	`

	result, err := r.pool.Exec(ctx, query, userID, streamerUsername)
	if err != nil {
		return fmt.Errorf("failed to unfollow streamer: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("follow relationship not found")
	}

	return nil
}

// IsFollowing checks if a user is following a streamer
func (r *StreamFollowRepository) IsFollowing(ctx context.Context, userID uuid.UUID, streamerUsername string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM stream_follows
			WHERE user_id = $1 AND streamer_username = $2
		)
	`

	var exists bool
	err := r.pool.QueryRow(ctx, query, userID, streamerUsername).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check follow status: %w", err)
	}

	return exists, nil
}

// GetFollowedStreamers returns all streamers a user is following
func (r *StreamFollowRepository) GetFollowedStreamers(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.StreamFollow, error) {
	query := `
		SELECT id, user_id, streamer_username, notifications_enabled, created_at, updated_at
		FROM stream_follows
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get followed streamers: %w", err)
	}
	defer rows.Close()

	var follows []models.StreamFollow
	for rows.Next() {
		var follow models.StreamFollow
		err := rows.Scan(
			&follow.ID,
			&follow.UserID,
			&follow.StreamerUsername,
			&follow.NotificationsEnabled,
			&follow.CreatedAt,
			&follow.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan follow: %w", err)
		}
		follows = append(follows, follow)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating follows: %w", err)
	}

	return follows, nil
}

// GetFollowersForStreamer returns all users following a specific streamer with notifications enabled
func (r *StreamFollowRepository) GetFollowersForStreamer(ctx context.Context, streamerUsername string) ([]uuid.UUID, error) {
	query := `
		SELECT user_id
		FROM stream_follows
		WHERE streamer_username = $1 AND notifications_enabled = TRUE
	`

	rows, err := r.pool.Query(ctx, query, streamerUsername)
	if err != nil {
		return nil, fmt.Errorf("failed to get followers: %w", err)
	}
	defer rows.Close()

	var userIDs []uuid.UUID
	for rows.Next() {
		var userID uuid.UUID
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("failed to scan user ID: %w", err)
		}
		userIDs = append(userIDs, userID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating followers: %w", err)
	}

	return userIDs, nil
}

// GetFollow retrieves a specific follow relationship
func (r *StreamFollowRepository) GetFollow(ctx context.Context, userID uuid.UUID, streamerUsername string) (*models.StreamFollow, error) {
	query := `
		SELECT id, user_id, streamer_username, notifications_enabled, created_at, updated_at
		FROM stream_follows
		WHERE user_id = $1 AND streamer_username = $2
	`

	var follow models.StreamFollow
	err := r.pool.QueryRow(ctx, query, userID, streamerUsername).Scan(
		&follow.ID,
		&follow.UserID,
		&follow.StreamerUsername,
		&follow.NotificationsEnabled,
		&follow.CreatedAt,
		&follow.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("follow relationship not found")
		}
		return nil, fmt.Errorf("failed to get follow: %w", err)
	}

	return &follow, nil
}

// UpdateNotificationPreference updates the notification preference for a follow
func (r *StreamFollowRepository) UpdateNotificationPreference(ctx context.Context, userID uuid.UUID, streamerUsername string, enabled bool) error {
	query := `
		UPDATE stream_follows
		SET notifications_enabled = $3, updated_at = NOW()
		WHERE user_id = $1 AND streamer_username = $2
	`

	result, err := r.pool.Exec(ctx, query, userID, streamerUsername, enabled)
	if err != nil {
		return fmt.Errorf("failed to update notification preference: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("follow relationship not found")
	}

	return nil
}

// GetFollowCount returns the number of streamers a user is following
func (r *StreamFollowRepository) GetFollowCount(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM stream_follows
		WHERE user_id = $1
	`

	var count int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get follow count: %w", err)
	}

	return count, nil
}
