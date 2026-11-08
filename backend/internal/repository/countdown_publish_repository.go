package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ErrCountdownEditionNotPublishing means the edition is no longer claimed
// for publishing.
var ErrCountdownEditionNotPublishing = errors.New("countdown edition is not being published")

// ClaimEditionForPublish moves the oldest rendered edition to "publishing"
// and returns it with its series. A claim older than staleAfter is taken
// again.
func (r *CountdownRepository) ClaimEditionForPublish(ctx context.Context, staleAfter time.Duration) (*models.CountdownEdition, *models.CountdownSeries, error) {
	return r.claimEdition(ctx, models.CountdownEditionRendered, models.CountdownEditionPublishing, staleAfter)
}

// ListEditionRenders returns the edition's rendered files that are still on
// the render host.
func (r *CountdownRepository) ListEditionRenders(ctx context.Context, editionID uuid.UUID) ([]models.CountdownRender, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, edition_id, kind, clip_id, file_name, sha256, bytes, duration_seconds, rendered_at
		FROM countdown_renders
		WHERE edition_id = $1 AND deleted_at IS NULL
		ORDER BY kind, file_name`, editionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.CountdownRender{}
	for rows.Next() {
		var x models.CountdownRender
		if err := rows.Scan(&x.ID, &x.EditionID, &x.Kind, &x.ClipID, &x.FileName, &x.SHA256, &x.Bytes,
			&x.DurationSeconds, &x.RenderedAt); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

// ClipDurations returns the duration of each fetched clip of an edition,
// including clips whose files are already deleted.
func (r *CountdownRepository) ClipDurations(ctx context.Context, editionID uuid.UUID) (map[uuid.UUID]float64, error) {
	rows, err := r.pool.Query(ctx, `SELECT clip_id, duration_seconds FROM countdown_edition_media WHERE edition_id = $1`, editionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[uuid.UUID]float64{}
	for rows.Next() {
		var id uuid.UUID
		var d float64
		if err := rows.Scan(&id, &d); err != nil {
			return nil, err
		}
		out[id] = d
	}
	return out, rows.Err()
}

// ListUploads returns the edition's uploads by render ID.
func (r *CountdownRepository) ListUploads(ctx context.Context, editionID uuid.UUID) (map[uuid.UUID]models.CountdownUpload, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT render_id, edition_id, kind, clip_id, youtube_video_id, privacy_status, publish_at, uploaded_at
		FROM countdown_uploads WHERE edition_id = $1`, editionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[uuid.UUID]models.CountdownUpload{}
	for rows.Next() {
		var u models.CountdownUpload
		if err := rows.Scan(&u.RenderID, &u.EditionID, &u.Kind, &u.ClipID, &u.YouTubeVideoID, &u.PrivacyStatus,
			&u.PublishAt, &u.UploadedAt); err != nil {
			return nil, err
		}
		out[u.RenderID] = u
	}
	return out, rows.Err()
}

// RecordUpload stores an uploaded video and keeps the edition's claim fresh.
// It records the upload even when the claim was lost, so the video is never
// uploaded twice.
func (r *CountdownRepository) RecordUpload(ctx context.Context, u *models.CountdownUpload) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `
		INSERT INTO countdown_uploads (render_id, edition_id, kind, clip_id, youtube_video_id, privacy_status, publish_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (render_id) DO NOTHING`,
		u.RenderID, u.EditionID, u.Kind, u.ClipID, u.YouTubeVideoID, u.PrivacyStatus, u.PublishAt); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE countdown_editions SET updated_at = NOW() WHERE id = $1 AND status = 'publishing'`, u.EditionID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// FinishPublish moves an edition out of "publishing": to "published", back
// to "rendered" to retry later, or to "failed" with a reason.
func (r *CountdownRepository) FinishPublish(ctx context.Context, editionID uuid.UUID, status string, reason *string) error {
	switch status {
	case models.CountdownEditionPublished, models.CountdownEditionRendered, models.CountdownEditionFailed:
	default:
		return fmt.Errorf("invalid publish outcome %q", status)
	}
	if reason != nil && len(*reason) > 500 {
		short := (*reason)[:500]
		reason = &short
	}
	tag, err := r.pool.Exec(ctx, `
		UPDATE countdown_editions SET status = $2, failure_reason = $3, updated_at = NOW()
		WHERE id = $1 AND status = 'publishing'`, editionID, status, reason)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrCountdownEditionNotPublishing
	}
	return nil
}

// SeriesPlaylist returns the series' YouTube playlist ID, or "" when it has
// none yet.
func (r *CountdownRepository) SeriesPlaylist(ctx context.Context, seriesID uuid.UUID) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx, `SELECT playlist_id FROM countdown_youtube_playlists WHERE series_id = $1`, seriesID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return id, err
}

// SaveSeriesPlaylist stores the series' YouTube playlist ID.
func (r *CountdownRepository) SaveSeriesPlaylist(ctx context.Context, seriesID uuid.UUID, playlistID string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO countdown_youtube_playlists (series_id, playlist_id) VALUES ($1, $2)
		ON CONFLICT (series_id) DO UPDATE SET playlist_id = EXCLUDED.playlist_id`, seriesID, playlistID)
	return err
}

// ListSeriesProgress returns, for each enabled series, its newest rendered
// (or later) and newest published edition dates. A skipped day counts as
// reached at both stages: nothing is owed for it.
func (r *CountdownRepository) ListSeriesProgress(ctx context.Context) ([]models.CountdownSeriesProgress, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT s.slug, s.selection_hour_utc, s.created_at,
		       to_char(max(e.edition_date) FILTER (WHERE e.status IN ('rendered', 'publishing', 'published', 'skipped')), 'YYYY-MM-DD'),
		       to_char(max(e.edition_date) FILTER (WHERE e.status IN ('published', 'skipped')), 'YYYY-MM-DD')
		FROM countdown_series s
		LEFT JOIN countdown_editions e ON e.series_id = s.id
		WHERE s.enabled
		GROUP BY s.id
		ORDER BY s.priority, s.slug`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.CountdownSeriesProgress{}
	for rows.Next() {
		var p models.CountdownSeriesProgress
		if err := rows.Scan(&p.Slug, &p.SelectionHourUTC, &p.CreatedAt, &p.LatestRendered, &p.LatestPublished); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
