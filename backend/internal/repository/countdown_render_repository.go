package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
)

// ErrCountdownEditionNotRendering means the edition is no longer claimed for
// rendering.
var ErrCountdownEditionNotRendering = errors.New("countdown edition is not being rendered")

// ClaimEditionForRender moves the oldest fetched edition to "rendering" and
// returns it with its series and its stored media. A claim left in
// "rendering" for longer than staleAfter is taken again. It returns nil
// values when nothing is waiting.
func (r *CountdownRepository) ClaimEditionForRender(ctx context.Context, staleAfter time.Duration) (*models.CountdownEdition, *models.CountdownSeries, []models.CountdownEditionMedia, error) {
	edition, series, err := r.claimEdition(ctx, models.CountdownEditionFetched, models.CountdownEditionRendering, staleAfter)
	if err != nil || edition == nil {
		return nil, nil, nil, err
	}
	media, err := r.ListEditionMedia(ctx, edition.ID)
	if err != nil {
		return nil, nil, nil, err
	}
	return edition, series, media, nil
}

// ListEditionMedia returns the edition's fetched files that are not deleted.
func (r *CountdownRepository) ListEditionMedia(ctx context.Context, editionID uuid.UUID) ([]models.CountdownEditionMedia, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT edition_id, clip_id, source, file_name, sha256, bytes, duration_seconds, width, height, fetched_at
		FROM countdown_edition_media
		WHERE edition_id = $1 AND deleted_at IS NULL`, editionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.CountdownEditionMedia{}
	for rows.Next() {
		var m models.CountdownEditionMedia
		if err := rows.Scan(&m.EditionID, &m.ClipID, &m.Source, &m.FileName, &m.SHA256, &m.Bytes,
			&m.DurationSeconds, &m.Width, &m.Height, &m.FetchedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// RecordRender stores a rendered file and keeps the edition's claim fresh.
func (r *CountdownRepository) RecordRender(ctx context.Context, x *models.CountdownRender) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	tag, err := tx.Exec(ctx, `UPDATE countdown_editions SET updated_at = NOW() WHERE id = $1 AND status = 'rendering'`, x.EditionID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrCountdownEditionNotRendering
	}
	if err := tx.QueryRow(ctx, `
		INSERT INTO countdown_renders (edition_id, kind, clip_id, file_name, sha256, bytes, duration_seconds)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (edition_id, kind, clip_id) DO UPDATE SET file_name = EXCLUDED.file_name, sha256 = EXCLUDED.sha256,
			bytes = EXCLUDED.bytes, duration_seconds = EXCLUDED.duration_seconds, rendered_at = NOW(), deleted_at = NULL
		RETURNING id, rendered_at`,
		x.EditionID, x.Kind, x.ClipID, x.FileName, x.SHA256, x.Bytes, x.DurationSeconds).Scan(&x.ID, &x.RenderedAt); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// FinishRender moves an edition out of "rendering" to "rendered", or to
// "failed" with a reason.
func (r *CountdownRepository) FinishRender(ctx context.Context, editionID uuid.UUID, status string, reason *string) error {
	if status != models.CountdownEditionRendered && status != models.CountdownEditionFailed {
		return fmt.Errorf("invalid render outcome %q", status)
	}
	if reason != nil && len(*reason) > 500 {
		short := (*reason)[:500]
		reason = &short
	}
	tag, err := r.pool.Exec(ctx, `
		UPDATE countdown_editions SET status = $2, failure_reason = $3, updated_at = NOW()
		WHERE id = $1 AND status = 'rendering'`, editionID, status, reason)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrCountdownEditionNotRendering
	}
	return nil
}

// ListRendersToDelete returns rendered files older than cutoff, and every
// rendered file of a failed edition, that are not marked deleted yet.
func (r *CountdownRepository) ListRendersToDelete(ctx context.Context, cutoff time.Time) ([]models.CountdownRender, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT x.id, x.edition_id, x.kind, x.file_name
		FROM countdown_renders x
		JOIN countdown_editions e ON e.id = x.edition_id
		WHERE x.deleted_at IS NULL AND (x.rendered_at < $1 OR e.status = 'failed')
		ORDER BY x.rendered_at
		LIMIT 1000`, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.CountdownRender{}
	for rows.Next() {
		var x models.CountdownRender
		if err := rows.Scan(&x.ID, &x.EditionID, &x.Kind, &x.FileName); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

// MarkRenderDeleted records that a rendered file was removed.
func (r *CountdownRepository) MarkRenderDeleted(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE countdown_renders SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	return err
}
