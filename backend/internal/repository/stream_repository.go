package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// StreamRepository handles database operations for streams
type StreamRepository struct {
	pool *pgxpool.Pool
}

// NewStreamRepository creates a new stream repository
func NewStreamRepository(pool *pgxpool.Pool) *StreamRepository {
	return &StreamRepository{pool: pool}
}

// UpsertStream inserts or updates stream metadata in the database
func (r *StreamRepository) UpsertStream(ctx context.Context, stream *models.Stream) error {
	query := `
		INSERT INTO streams (
			streamer_username, streamer_user_id, display_name, is_live,
			last_went_live, last_went_offline, game_name, title, viewer_count
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (streamer_username) DO UPDATE SET
			streamer_user_id = EXCLUDED.streamer_user_id,
			display_name = EXCLUDED.display_name,
			is_live = EXCLUDED.is_live,
			last_went_live = EXCLUDED.last_went_live,
			last_went_offline = EXCLUDED.last_went_offline,
			game_name = EXCLUDED.game_name,
			title = EXCLUDED.title,
			viewer_count = EXCLUDED.viewer_count,
			updated_at = NOW()
		RETURNING id, created_at, updated_at
	`

	return r.pool.QueryRow(
		ctx, query,
		stream.StreamerUsername,
		stream.StreamerUserID,
		stream.DisplayName,
		stream.IsLive,
		stream.LastWentLive,
		stream.LastWentOffline,
		stream.GameName,
		stream.Title,
		stream.ViewerCount,
	).Scan(&stream.ID, &stream.CreatedAt, &stream.UpdatedAt)
}

// GetStreamByUsername retrieves stream metadata by username
func (r *StreamRepository) GetStreamByUsername(ctx context.Context, username string) (*models.Stream, error) {
	query := `
		SELECT id, streamer_username, streamer_user_id, display_name, is_live,
		       last_went_live, last_went_offline, game_name, title, viewer_count,
		       created_at, updated_at
		FROM streams
		WHERE streamer_username = $1
	`

	stream := &models.Stream{}
	err := r.pool.QueryRow(ctx, query, username).Scan(
		&stream.ID,
		&stream.StreamerUsername,
		&stream.StreamerUserID,
		&stream.DisplayName,
		&stream.IsLive,
		&stream.LastWentLive,
		&stream.LastWentOffline,
		&stream.GameName,
		&stream.Title,
		&stream.ViewerCount,
		&stream.CreatedAt,
		&stream.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return stream, nil
}

// UpdateStreamStatus updates the live status of a stream
func (r *StreamRepository) UpdateStreamStatus(ctx context.Context, username string, isLive bool, wentLiveOrOfflineAt time.Time) error {
	var query string
	if isLive {
		query = `
			UPDATE streams
			SET is_live = TRUE, last_went_live = $2, updated_at = NOW()
			WHERE streamer_username = $1
		`
	} else {
		query = `
			UPDATE streams
			SET is_live = FALSE, last_went_offline = $2, updated_at = NOW()
			WHERE streamer_username = $1
		`
	}

	_, err := r.pool.Exec(ctx, query, username, wentLiveOrOfflineAt)
	return err
}
