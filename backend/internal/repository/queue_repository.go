package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// QueueRepository handles database operations for queue items
type QueueRepository struct {
	pool *pgxpool.Pool
}

// NewQueueRepository creates a new QueueRepository
func NewQueueRepository(pool *pgxpool.Pool) *QueueRepository {
	return &QueueRepository{
		pool: pool,
	}
}

// GetUserQueue retrieves a user's queue with optional limit
func (r *QueueRepository) GetUserQueue(ctx context.Context, userID uuid.UUID, limit int) ([]models.QueueItemWithClip, error) {
	query := `
		SELECT
			qi.id, qi.user_id, qi.clip_id, qi.position, qi.added_at, qi.played_at, qi.created_at, qi.updated_at,
			c.id, c.twitch_clip_id, c.twitch_clip_url, c.embed_url, c.title, c.creator_name,
			c.broadcaster_name, c.game_name, c.thumbnail_url, c.duration, c.view_count,
			c.vote_score, c.comment_count, c.favorite_count, c.is_nsfw, c.created_at,
			c.stream_source, c.status, c.video_url, c.processed_at, c.quality, c.start_time, c.end_time
		FROM queue_items qi
		LEFT JOIN clips c ON qi.clip_id = c.id
		WHERE qi.user_id = $1 AND qi.played_at IS NULL
		ORDER BY qi.position ASC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get user queue: %w", err)
	}
	defer rows.Close()

	var items []models.QueueItemWithClip
	for rows.Next() {
		var item models.QueueItemWithClip
		var clip models.Clip

		err := rows.Scan(
			&item.ID, &item.UserID, &item.ClipID, &item.Position, &item.AddedAt, &item.PlayedAt, &item.CreatedAt, &item.UpdatedAt,
			&clip.ID, &clip.TwitchClipID, &clip.TwitchClipURL, &clip.EmbedURL, &clip.Title, &clip.CreatorName,
			&clip.BroadcasterName, &clip.GameName, &clip.ThumbnailURL, &clip.Duration, &clip.ViewCount,
			&clip.VoteScore, &clip.CommentCount, &clip.FavoriteCount, &clip.IsNSFW, &clip.CreatedAt,
			&clip.StreamSource, &clip.Status, &clip.VideoURL, &clip.ProcessedAt, &clip.Quality, &clip.StartTime, &clip.EndTime,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan queue item: %w", err)
		}

		item.Clip = &clip
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating queue items: %w", err)
	}

	return items, nil
}

// GetQueueCount gets the total count of unplayed items in a user's queue
func (r *QueueRepository) GetQueueCount(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM queue_items WHERE user_id = $1 AND played_at IS NULL`

	var count int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get queue count: %w", err)
	}

	return count, nil
}

// GetMaxPosition gets the maximum position in a user's queue (all items to avoid constraint violations)
func (r *QueueRepository) GetMaxPosition(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT COALESCE(MAX(position), 0) FROM queue_items WHERE user_id = $1`

	var maxPos int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&maxPos)
	if err != nil {
		return 0, fmt.Errorf("failed to get max position: %w", err)
	}

	return maxPos, nil
}

// AddItem adds a clip to the queue at the end (using transaction to avoid race conditions)
func (r *QueueRepository) AddItem(ctx context.Context, item *models.QueueItem) error {
	// Use a transaction to calculate position and insert atomically
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Use advisory lock based on user_id to prevent concurrent inserts
	// pg_advisory_xact_lock is automatically released at transaction end
	// We use a hash of the user_id bytes to get a consistent int64 lock key
	userIDBytes := item.UserID[:]
	lockKey := int64(userIDBytes[0])<<56 | int64(userIDBytes[1])<<48 | int64(userIDBytes[2])<<40 | int64(userIDBytes[3])<<32 |
		int64(userIDBytes[4])<<24 | int64(userIDBytes[5])<<16 | int64(userIDBytes[6])<<8 | int64(userIDBytes[7])
	_, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, lockKey)
	if err != nil {
		return fmt.Errorf("failed to acquire advisory lock: %w", err)
	}

	// Now calculate next position
	var maxPos int
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(MAX(position), 0)
		FROM queue_items
		WHERE user_id = $1
	`, item.UserID).Scan(&maxPos)
	if err != nil {
		return fmt.Errorf("failed to get max position: %w", err)
	}

	item.Position = maxPos + 1

	err = tx.QueryRow(ctx, `
		INSERT INTO queue_items (id, user_id, clip_id, position, added_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING created_at, updated_at, added_at
	`,
		item.ID,
		item.UserID,
		item.ClipID,
		item.Position,
	).Scan(&item.CreatedAt, &item.UpdatedAt, &item.AddedAt)

	if err != nil {
		return fmt.Errorf("failed to add queue item: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// AddItemAtTop adds a clip to the top of the queue with transaction support
func (r *QueueRepository) AddItemAtTop(ctx context.Context, item *models.QueueItem) error {
	// Use a transaction to shift positions and add item
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Shift ALL items down by 1 (including played items to avoid constraint violations)
	_, err = tx.Exec(ctx, `UPDATE queue_items SET position = position + 1 WHERE user_id = $1`, item.UserID)
	if err != nil {
		return fmt.Errorf("failed to shift positions: %w", err)
	}

	// Insert new item at position 1
	err = tx.QueryRow(ctx, `
		INSERT INTO queue_items (id, user_id, clip_id, position, added_at)
		VALUES ($1, $2, $3, 1, NOW())
		RETURNING created_at, updated_at, added_at
	`, item.ID, item.UserID, item.ClipID).Scan(&item.CreatedAt, &item.UpdatedAt, &item.AddedAt)

	if err != nil {
		return fmt.Errorf("failed to add queue item: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	item.Position = 1
	return nil
}

// RemoveItem removes an item from the queue and shifts positions
func (r *QueueRepository) RemoveItem(ctx context.Context, itemID uuid.UUID, userID uuid.UUID) error {
	// Use a transaction to remove item and shift positions
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get the position of the item being removed
	var position int
	err = tx.QueryRow(ctx, `SELECT position FROM queue_items WHERE id = $1 AND user_id = $2`, itemID, userID).Scan(&position)
	if err == pgx.ErrNoRows {
		return pgx.ErrNoRows
	}
	if err != nil {
		return fmt.Errorf("failed to get item position: %w", err)
	}

	// Delete the item
	result, err := tx.Exec(ctx, `DELETE FROM queue_items WHERE id = $1 AND user_id = $2`, itemID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove queue item: %w", err)
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	// Shift all items after the removed position down
	_, err = tx.Exec(ctx, `UPDATE queue_items SET position = position - 1 WHERE user_id = $1 AND position > $2`, userID, position)
	if err != nil {
		return fmt.Errorf("failed to shift positions: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ReorderItem moves a queue item to a new position
func (r *QueueRepository) ReorderItem(ctx context.Context, itemID uuid.UUID, userID uuid.UUID, newPosition int) error {
	// Use a transaction for reordering
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get current position
	var oldPosition int
	err = tx.QueryRow(ctx, `SELECT position FROM queue_items WHERE id = $1 AND user_id = $2`, itemID, userID).Scan(&oldPosition)
	if err == pgx.ErrNoRows {
		return fmt.Errorf("queue item not found")
	}
	if err != nil {
		return fmt.Errorf("failed to get old position: %w", err)
	}

	// Shift items between old and new position
	if oldPosition < newPosition {
		// Moving down: shift items up
		_, err = tx.Exec(ctx, `
			UPDATE queue_items
			SET position = position - 1
			WHERE user_id = $1 AND position > $2 AND position <= $3
		`, userID, oldPosition, newPosition)
	} else if oldPosition > newPosition {
		// Moving up: shift items down
		_, err = tx.Exec(ctx, `
			UPDATE queue_items
			SET position = position + 1
			WHERE user_id = $1 AND position >= $2 AND position < $3
		`, userID, newPosition, oldPosition)
	}

	if err != nil {
		return fmt.Errorf("failed to shift positions: %w", err)
	}

	// Update item to new position
	_, err = tx.Exec(ctx, `
		UPDATE queue_items
		SET position = $1, updated_at = NOW()
		WHERE id = $2 AND user_id = $3
	`, newPosition, itemID, userID)
	if err != nil {
		return fmt.Errorf("failed to update position: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// MarkAsPlayed marks a queue item as played
func (r *QueueRepository) MarkAsPlayed(ctx context.Context, itemID uuid.UUID, userID uuid.UUID) error {
	query := `
		UPDATE queue_items
		SET played_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND user_id = $2 AND played_at IS NULL
	`

	result, err := r.pool.Exec(ctx, query, itemID, userID)
	if err != nil {
		return fmt.Errorf("failed to mark as played: %w", err)
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// ClearQueue removes all unplayed items from the queue
func (r *QueueRepository) ClearQueue(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM queue_items WHERE user_id = $1 AND played_at IS NULL`

	_, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to clear queue: %w", err)
	}

	return nil
}

// GetItemByID retrieves a specific queue item
func (r *QueueRepository) GetItemByID(ctx context.Context, itemID uuid.UUID, userID uuid.UUID) (*models.QueueItem, error) {
	query := `
		SELECT id, user_id, clip_id, position, added_at, played_at, created_at, updated_at
		FROM queue_items
		WHERE id = $1 AND user_id = $2
	`

	var item models.QueueItem
	err := r.pool.QueryRow(ctx, query, itemID, userID).Scan(
		&item.ID,
		&item.UserID,
		&item.ClipID,
		&item.Position,
		&item.AddedAt,
		&item.PlayedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get queue item: %w", err)
	}

	return &item, nil
}

// CleanupStaleQueues removes old inactive queues
func (r *QueueRepository) CleanupStaleQueues(ctx context.Context) error {
	query := `SELECT cleanup_stale_queues()`

	_, err := r.pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to cleanup stale queues: %w", err)
	}

	return nil
}
