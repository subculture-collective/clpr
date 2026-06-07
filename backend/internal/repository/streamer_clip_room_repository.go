package repository

import (
	"context"
	"fmt"
	"strings"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// StreamerClipRoomRepository handles database operations for streamer clip rooms.
type StreamerClipRoomRepository struct {
	pool *pgxpool.Pool
}

// NewStreamerClipRoomRepository creates a new StreamerClipRoomRepository.
func NewStreamerClipRoomRepository(pool *pgxpool.Pool) *StreamerClipRoomRepository {
	return &StreamerClipRoomRepository{pool: pool}
}

func scanStreamerClipRoom(row pgx.Row) (*models.StreamerClipRoom, error) {
	var room models.StreamerClipRoom
	if err := row.Scan(
		&room.ID,
		&room.OwnerUserID,
		&room.TwitchChannel,
		&room.ApprovalMode,
		&room.IsActive,
		&room.LastListenerError,
		&room.ListenerStartedAt,
		&room.CreatedAt,
		&room.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &room, nil
}

func scanStreamerClipRoomItem(row pgx.Row) (*models.StreamerClipRoomItem, error) {
	var item models.StreamerClipRoomItem
	if err := row.Scan(
		&item.ID,
		&item.RoomID,
		&item.ClipID,
		&item.SourceURL,
		&item.SourceType,
		&item.Status,
		&item.Position,
		&item.TwitchMessageID,
		&item.TwitchUserID,
		&item.TwitchUsername,
		&item.MessageText,
		&item.SkipReason,
		&item.DetectedAt,
		&item.ApprovedAt,
		&item.ApprovedByUserID,
		&item.RejectedAt,
		&item.RejectedByUserID,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &item, nil
}

// GetOrCreateRoom retrieves a room by owner/channel or creates it if missing.
func (r *StreamerClipRoomRepository) GetOrCreateRoom(ctx context.Context, ownerUserID uuid.UUID, channel string) (*models.StreamerClipRoom, error) {
	normalized := strings.ToLower(strings.TrimSpace(channel))
	if normalized == "" {
		return nil, fmt.Errorf("channel is required")
	}

	const insertQuery = `
		INSERT INTO streamer_clip_rooms (owner_user_id, twitch_channel)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
		RETURNING id, owner_user_id, twitch_channel, approval_mode, is_active,
			last_listener_error, listener_started_at, created_at, updated_at
	`

	room, err := scanStreamerClipRoom(r.pool.QueryRow(ctx, insertQuery, ownerUserID, normalized))
	if err == nil {
		return room, nil
	}
	if err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to get or create streamer clip room: %w", err)
	}

	const selectQuery = `
		SELECT id, owner_user_id, twitch_channel, approval_mode, is_active,
			last_listener_error, listener_started_at, created_at, updated_at
		FROM streamer_clip_rooms
		WHERE owner_user_id = $1
		  AND lower(twitch_channel) = $2
	`

	room, err = scanStreamerClipRoom(r.pool.QueryRow(ctx, selectQuery, ownerUserID, normalized))
	if err != nil {
		return nil, fmt.Errorf("failed to get or create streamer clip room: %w", err)
	}

	return room, nil
}

// GetRoomByID retrieves a streamer clip room by its ID.
func (r *StreamerClipRoomRepository) GetRoomByID(ctx context.Context, roomID uuid.UUID) (*models.StreamerClipRoom, error) {
	const query = `
		SELECT id, owner_user_id, twitch_channel, approval_mode, is_active,
			last_listener_error, listener_started_at, created_at, updated_at
		FROM streamer_clip_rooms
		WHERE id = $1
	`

	room, err := scanStreamerClipRoom(r.pool.QueryRow(ctx, query, roomID))
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get streamer clip room: %w", err)
	}

	return room, nil
}

// SetRoomActive updates the active state and listener error for a room.
func (r *StreamerClipRoomRepository) SetRoomActive(ctx context.Context, roomID uuid.UUID, active bool, listenerError *string) error {
	const query = `
		UPDATE streamer_clip_rooms
		SET is_active = $2,
			last_listener_error = $3,
			listener_started_at = CASE WHEN $2 THEN NOW() ELSE NULL END,
			updated_at = NOW()
		WHERE id = $1
	`

	if _, err := r.pool.Exec(ctx, query, roomID, active, listenerError); err != nil {
		return fmt.Errorf("failed to update streamer clip room active state: %w", err)
	}

	return nil
}

// CreateItem inserts a new streamer clip room item.
func (r *StreamerClipRoomRepository) CreateItem(ctx context.Context, item *models.StreamerClipRoomItem) error {
	const query = `
		INSERT INTO streamer_clip_room_items (
			id, room_id, clip_id, source_url, source_type, status, position,
			twitch_message_id, twitch_user_id, twitch_username, message_text, skip_reason,
			detected_at, approved_at, approved_by_user_id, rejected_at, rejected_by_user_id
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12,
			$13, $14, $15, $16, $17
		)
		RETURNING created_at, updated_at
	`

	if err := r.pool.QueryRow(ctx, query,
		item.ID,
		item.RoomID,
		item.ClipID,
		item.SourceURL,
		item.SourceType,
		item.Status,
		item.Position,
		item.TwitchMessageID,
		item.TwitchUserID,
		item.TwitchUsername,
		item.MessageText,
		item.SkipReason,
		item.DetectedAt,
		item.ApprovedAt,
		item.ApprovedByUserID,
		item.RejectedAt,
		item.RejectedByUserID,
	).Scan(&item.CreatedAt, &item.UpdatedAt); err != nil {
		return fmt.Errorf("failed to create streamer clip room item: %w", err)
	}

	return nil
}

// ListItems lists room items filtered by status.
func (r *StreamerClipRoomRepository) ListItems(ctx context.Context, roomID uuid.UUID, status string, limit int) ([]models.StreamerClipRoomItem, error) {
	if limit <= 0 {
		limit = 50
	}

	status = strings.ToLower(strings.TrimSpace(status))

	query := `
		SELECT id, room_id, clip_id, source_url, source_type, status, position,
			twitch_message_id, twitch_user_id, twitch_username, message_text, skip_reason,
			detected_at, approved_at, approved_by_user_id, rejected_at, rejected_by_user_id,
			created_at, updated_at
		FROM streamer_clip_room_items
		WHERE room_id = $1
	`
	args := []any{roomID}
	if status != "" && status != "all" {
		query += " AND status = $2"
		args = append(args, status)
	}
	query += `
		ORDER BY CASE WHEN status = 'approved' THEN position END ASC NULLS LAST, detected_at DESC
		LIMIT $` + fmt.Sprintf("%d", len(args)+1)
	args = append(args, limit)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list streamer clip room items: %w", err)
	}
	defer rows.Close()

	items := make([]models.StreamerClipRoomItem, 0)
	for rows.Next() {
		var item models.StreamerClipRoomItem
		if err := rows.Scan(
			&item.ID,
			&item.RoomID,
			&item.ClipID,
			&item.SourceURL,
			&item.SourceType,
			&item.Status,
			&item.Position,
			&item.TwitchMessageID,
			&item.TwitchUserID,
			&item.TwitchUsername,
			&item.MessageText,
			&item.SkipReason,
			&item.DetectedAt,
			&item.ApprovedAt,
			&item.ApprovedByUserID,
			&item.RejectedAt,
			&item.RejectedByUserID,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan streamer clip room item: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate streamer clip room items: %w", err)
	}

	return items, nil
}

// GetItem retrieves a room item by room and item ID.
func (r *StreamerClipRoomRepository) GetItem(ctx context.Context, roomID, itemID uuid.UUID) (*models.StreamerClipRoomItem, error) {
	const query = `
		SELECT id, room_id, clip_id, source_url, source_type, status, position,
			twitch_message_id, twitch_user_id, twitch_username, message_text, skip_reason,
			detected_at, approved_at, approved_by_user_id, rejected_at, rejected_by_user_id,
			created_at, updated_at
		FROM streamer_clip_room_items
		WHERE room_id = $1 AND id = $2
	`

	item, err := scanStreamerClipRoomItem(r.pool.QueryRow(ctx, query, roomID, itemID))
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get streamer clip room item: %w", err)
	}

	return item, nil
}

// ApproveItem marks an item approved and assigns the next available position.
func (r *StreamerClipRoomRepository) ApproveItem(ctx context.Context, roomID, itemID, approverID uuid.UUID) (*models.StreamerClipRoomItem, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin approve transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `SELECT id FROM streamer_clip_rooms WHERE id = $1 FOR UPDATE`, roomID); err != nil {
		return nil, fmt.Errorf("failed to lock streamer clip room: %w", err)
	}

	const selectQuery = `
		SELECT id, room_id, clip_id, source_url, source_type, status, position,
			twitch_message_id, twitch_user_id, twitch_username, message_text, skip_reason,
			detected_at, approved_at, approved_by_user_id, rejected_at, rejected_by_user_id,
			created_at, updated_at
		FROM streamer_clip_room_items
		WHERE room_id = $1 AND id = $2
		FOR UPDATE
	`

	item, err := scanStreamerClipRoomItem(tx.QueryRow(ctx, selectQuery, roomID, itemID))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("failed to lock streamer clip room item: %w", err)
	}

	if item.Status == models.StreamerClipRoomItemStatusApproved {
		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("failed to commit approve transaction: %w", err)
		}
		return item, nil
	}

	var nextPosition int
	if err := tx.QueryRow(ctx, `
		SELECT COALESCE(MAX(position), 0) + 1
		FROM streamer_clip_room_items
		WHERE room_id = $1 AND status = 'approved'
	`, roomID).Scan(&nextPosition); err != nil {
		return nil, fmt.Errorf("failed to get next approved position: %w", err)
	}

	const updateQuery = `
		UPDATE streamer_clip_room_items
		SET status = 'approved',
			position = $3,
			approved_at = NOW(),
			approved_by_user_id = $4,
			rejected_at = NULL,
			rejected_by_user_id = NULL,
			updated_at = NOW()
		WHERE room_id = $1 AND id = $2
		RETURNING id, room_id, clip_id, source_url, source_type, status, position,
			twitch_message_id, twitch_user_id, twitch_username, message_text, skip_reason,
			detected_at, approved_at, approved_by_user_id, rejected_at, rejected_by_user_id,
			created_at, updated_at
	`

	item, err = scanStreamerClipRoomItem(tx.QueryRow(ctx, updateQuery, roomID, itemID, nextPosition, approverID))
	if err != nil {
		return nil, fmt.Errorf("failed to approve streamer clip room item: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit approve transaction: %w", err)
	}

	return item, nil
}

// RejectItem marks an item rejected and clears its approved position.
func (r *StreamerClipRoomRepository) RejectItem(ctx context.Context, roomID, itemID, rejecterID uuid.UUID) (*models.StreamerClipRoomItem, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin reject transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `SELECT id FROM streamer_clip_rooms WHERE id = $1 FOR UPDATE`, roomID); err != nil {
		return nil, fmt.Errorf("failed to lock streamer clip room: %w", err)
	}

	const selectQuery = `
		SELECT id, room_id, clip_id, source_url, source_type, status, position,
			twitch_message_id, twitch_user_id, twitch_username, message_text, skip_reason,
			detected_at, approved_at, approved_by_user_id, rejected_at, rejected_by_user_id,
			created_at, updated_at
		FROM streamer_clip_room_items
		WHERE room_id = $1 AND id = $2
		FOR UPDATE
	`

	if _, err := scanStreamerClipRoomItem(tx.QueryRow(ctx, selectQuery, roomID, itemID)); err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("failed to lock streamer clip room item: %w", err)
	}

	const updateQuery = `
		UPDATE streamer_clip_room_items
		SET status = 'rejected',
			position = NULL,
			approved_at = NULL,
			approved_by_user_id = NULL,
			rejected_at = NOW(),
			rejected_by_user_id = $3,
			updated_at = NOW()
		WHERE room_id = $1 AND id = $2
		RETURNING id, room_id, clip_id, source_url, source_type, status, position,
			twitch_message_id, twitch_user_id, twitch_username, message_text, skip_reason,
			detected_at, approved_at, approved_by_user_id, rejected_at, rejected_by_user_id,
			created_at, updated_at
	`

	item, err := scanStreamerClipRoomItem(tx.QueryRow(ctx, updateQuery, roomID, itemID, rejecterID))
	if err != nil {
		return nil, fmt.Errorf("failed to reject streamer clip room item: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit reject transaction: %w", err)
	}

	return item, nil
}

// ReorderApprovedItems reassigns approved item positions in the provided order.
func (r *StreamerClipRoomRepository) ReorderApprovedItems(ctx context.Context, roomID uuid.UUID, itemIDs []uuid.UUID) error {
	if len(itemIDs) == 0 {
		return nil
	}

	seen := make(map[uuid.UUID]struct{}, len(itemIDs))
	for _, id := range itemIDs {
		if _, ok := seen[id]; ok {
			return fmt.Errorf("duplicate streamer clip room item ID provided")
		}
		seen[id] = struct{}{}
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin reorder transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `SELECT id FROM streamer_clip_rooms WHERE id = $1 FOR UPDATE`, roomID); err != nil {
		return fmt.Errorf("failed to lock streamer clip room: %w", err)
	}

	for _, itemID := range itemIDs {
		var status string
		if err := tx.QueryRow(ctx, `
			SELECT status
			FROM streamer_clip_room_items
			WHERE room_id = $1 AND id = $2
			FOR UPDATE
		`, roomID, itemID).Scan(&status); err != nil {
			if err == pgx.ErrNoRows {
				return pgx.ErrNoRows
			}
			return fmt.Errorf("failed to lock streamer clip room item: %w", err)
		}
		if status != models.StreamerClipRoomItemStatusApproved {
			return fmt.Errorf("streamer clip room item %s is not approved", itemID)
		}
	}

	if _, err := tx.Exec(ctx, `
		UPDATE streamer_clip_room_items
		SET position = position + 100000,
			updated_at = NOW()
		WHERE room_id = $1
		  AND id = ANY($2)
		  AND status = 'approved'
	`, roomID, itemIDs); err != nil {
		return fmt.Errorf("failed to stage streamer clip room item reorder: %w", err)
	}

	for idx, itemID := range itemIDs {
		if _, err := tx.Exec(ctx, `
			UPDATE streamer_clip_room_items
			SET position = $3,
				updated_at = NOW()
			WHERE room_id = $1
			  AND id = $2
			  AND status = 'approved'
		`, roomID, itemID, idx+1); err != nil {
			return fmt.Errorf("failed to finalize streamer clip room item reorder: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit reorder transaction: %w", err)
	}

	return nil
}
