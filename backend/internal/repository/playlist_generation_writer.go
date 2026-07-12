package repository

import (
	"context"
	"fmt"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
)

// PlaylistGenerationWriter commits every state change for one generated playlist atomically.
type PlaylistGenerationWriter struct{ repo *PlaylistScriptRepository }

func NewPlaylistGenerationWriter(repo *PlaylistScriptRepository) *PlaylistGenerationWriter {
	return &PlaylistGenerationWriter{repo: repo}
}

func (w *PlaylistGenerationWriter) Persist(ctx context.Context, script *models.PlaylistScript, playlist *models.Playlist, clips []models.Clip) error {
	tx, err := w.repo.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin playlist generation transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	var previousPlaylistID *uuid.UUID
	if err = tx.QueryRow(ctx, `SELECT last_generated_playlist_id FROM playlist_scripts WHERE id = $1 AND is_active = true FOR UPDATE`, script.ID).Scan(&previousPlaylistID); err != nil {
		return fmt.Errorf("lock active playlist script: %w", err)
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO playlists (id, user_id, title, description, cover_url, visibility, is_curated, is_featured, display_order, script_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at, updated_at
	`, playlist.ID, playlist.UserID, playlist.Title, playlist.Description, playlist.CoverURL, playlist.Visibility,
		playlist.IsCurated, playlist.IsFeatured, playlist.DisplayOrder, playlist.ScriptID).Scan(&playlist.CreatedAt, &playlist.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert generated playlist: %w", err)
	}

	for index, clip := range clips {
		if _, err = tx.Exec(ctx, `INSERT INTO playlist_items (playlist_id, clip_id, order_index) VALUES ($1, $2, $3)`, playlist.ID, clip.ID, index); err != nil {
			return fmt.Errorf("insert generated playlist item: %w", err)
		}
	}
	now := time.Now()
	if _, err = tx.Exec(ctx, `INSERT INTO generated_playlists (id, script_id, playlist_id, generated_at) VALUES ($1, $2, $3, $4)`, uuid.New(), script.ID, playlist.ID, now); err != nil {
		return fmt.Errorf("record generated playlist: %w", err)
	}
	if previousPlaylistID != nil {
		if _, err = tx.Exec(ctx, `UPDATE playlists SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`, now, *previousPlaylistID); err != nil {
			return fmt.Errorf("retire previous generated playlist: %w", err)
		}
	}
	result, err := tx.Exec(ctx, `UPDATE playlist_scripts SET last_run_at = $1, last_generated_playlist_id = $2 WHERE id = $3 AND is_active = true`, now, playlist.ID, script.ID)
	if err != nil {
		return fmt.Errorf("update playlist script run metadata: %w", err)
	}
	if result.RowsAffected() != 1 {
		return fmt.Errorf("playlist script is inactive or missing")
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit playlist generation transaction: %w", err)
	}
	return nil
}
