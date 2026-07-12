package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	ErrPlaylistClipNotFound       = errors.New("playlist clip not found")
	ErrPlaylistClipLimit          = errors.New("playlist clip limit exceeded")
	ErrPlaylistMembershipMismatch = errors.New("clip IDs must exactly match playlist membership")
)

type PlaylistMembershipWriter struct{ repo *PlaylistRepository }

func NewPlaylistMembershipWriter(repo *PlaylistRepository) *PlaylistMembershipWriter {
	return &PlaylistMembershipWriter{repo: repo}
}

func (w *PlaylistMembershipWriter) Add(ctx context.Context, playlistID uuid.UUID, clipIDs []uuid.UUID) error {
	tx, err := w.repo.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin playlist membership transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	var lockedID uuid.UUID
	if err = tx.QueryRow(ctx, `SELECT id FROM playlists WHERE id = $1 AND deleted_at IS NULL FOR UPDATE`, playlistID).Scan(&lockedID); err != nil {
		return fmt.Errorf("lock playlist: %w", err)
	}
	var existingCount, requestedExisting, validClipCount int
	if err = tx.QueryRow(ctx, `SELECT COUNT(*) FROM playlist_items WHERE playlist_id = $1`, playlistID).Scan(&existingCount); err != nil {
		return fmt.Errorf("count playlist items: %w", err)
	}
	if err = tx.QueryRow(ctx, `SELECT COUNT(*) FROM clips WHERE id = ANY($1)`, clipIDs).Scan(&validClipCount); err != nil {
		return fmt.Errorf("validate clips: %w", err)
	}
	if validClipCount != len(clipIDs) {
		return ErrPlaylistClipNotFound
	}
	if err = tx.QueryRow(ctx, `SELECT COUNT(*) FROM playlist_items WHERE playlist_id = $1 AND clip_id = ANY($2)`, playlistID, clipIDs).Scan(&requestedExisting); err != nil {
		return fmt.Errorf("count existing requested clips: %w", err)
	}
	if existingCount+len(clipIDs)-requestedExisting > 1000 {
		return ErrPlaylistClipLimit
	}
	next := existingCount
	for _, clipID := range clipIDs {
		result, err := tx.Exec(ctx, `INSERT INTO playlist_items (playlist_id, clip_id, order_index) VALUES ($1, $2, $3) ON CONFLICT (playlist_id, clip_id) DO NOTHING`, playlistID, clipID, next)
		if err != nil {
			return fmt.Errorf("insert playlist item: %w", err)
		}
		if result.RowsAffected() == 1 {
			next++
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit playlist membership transaction: %w", err)
	}
	return nil
}

func (w *PlaylistMembershipWriter) Remove(ctx context.Context, playlistID, clipID uuid.UUID) error {
	tx, err := w.repo.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin playlist membership transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	var lockedID uuid.UUID
	if err = tx.QueryRow(ctx, `SELECT id FROM playlists WHERE id = $1 AND deleted_at IS NULL FOR UPDATE`, playlistID).Scan(&lockedID); err != nil {
		return fmt.Errorf("lock playlist: %w", err)
	}
	result, err := tx.Exec(ctx, `DELETE FROM playlist_items WHERE playlist_id = $1 AND clip_id = $2`, playlistID, clipID)
	if err != nil {
		return fmt.Errorf("delete playlist item: %w", err)
	}
	if result.RowsAffected() != 1 {
		return ErrPlaylistClipNotFound
	}
	if _, err = tx.Exec(ctx, `
		WITH ordered AS (SELECT clip_id, ROW_NUMBER() OVER (ORDER BY order_index, id) - 1 AS new_index FROM playlist_items WHERE playlist_id = $1)
		UPDATE playlist_items pi SET order_index = ordered.new_index FROM ordered WHERE pi.playlist_id = $1 AND pi.clip_id = ordered.clip_id
	`, playlistID); err != nil {
		return fmt.Errorf("reindex playlist items: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit playlist membership transaction: %w", err)
	}
	return nil
}

func (w *PlaylistMembershipWriter) Reorder(ctx context.Context, playlistID uuid.UUID, clipIDs []uuid.UUID) error {
	tx, err := w.repo.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin playlist reorder transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	var lockedID uuid.UUID
	if err = tx.QueryRow(ctx, `SELECT id FROM playlists WHERE id = $1 AND deleted_at IS NULL FOR UPDATE`, playlistID).Scan(&lockedID); err != nil {
		return fmt.Errorf("lock playlist: %w", err)
	}
	var matches bool
	if err = tx.QueryRow(ctx, `
		SELECT COUNT(*) = $2 AND COUNT(*) = (SELECT COUNT(*) FROM playlist_items WHERE playlist_id = $1)
		FROM (SELECT DISTINCT unnest($3::uuid[]) AS clip_id) requested
		WHERE clip_id IN (SELECT clip_id FROM playlist_items WHERE playlist_id = $1)
	`, playlistID, len(clipIDs), clipIDs).Scan(&matches); err != nil {
		return fmt.Errorf("validate playlist reorder membership: %w", err)
	}
	if !matches {
		return ErrPlaylistMembershipMismatch
	}
	if _, err = tx.Exec(ctx, `
		UPDATE playlist_items pi SET order_index = requested.position - 1
		FROM unnest($2::uuid[]) WITH ORDINALITY requested(clip_id, position)
		WHERE pi.playlist_id = $1 AND pi.clip_id = requested.clip_id
	`, playlistID, clipIDs); err != nil {
		return fmt.Errorf("reorder playlist items: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit playlist reorder transaction: %w", err)
	}
	return nil
}
