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

// ErrCountdownEditionNotFetching means the edition is no longer claimed for
// fetching, for example because an admin selected it again.
var ErrCountdownEditionNotFetching = errors.New("countdown edition is not being fetched")

// ClaimEditionForFetch moves the oldest selected edition to "fetching" and
// returns it with its series. An edition left in "fetching" for longer than
// staleAfter (a worker that died) is claimed again. It returns nil, nil when
// nothing is waiting.
func (r *CountdownRepository) ClaimEditionForFetch(ctx context.Context, staleAfter time.Duration) (*models.CountdownEdition, *models.CountdownSeries, error) {
	return r.claimEdition(ctx, models.CountdownEditionSelected, models.CountdownEditionFetching, staleAfter)
}

// claimEdition moves the oldest edition in status from (or left in status
// to for longer than staleAfter) to status to.
func (r *CountdownRepository) claimEdition(ctx context.Context, from, to string, staleAfter time.Duration) (*models.CountdownEdition, *models.CountdownSeries, error) {
	var editionID, seriesID uuid.UUID
	err := r.pool.QueryRow(ctx, `
		UPDATE countdown_editions e SET status = $3, failure_reason = NULL, updated_at = NOW()
		WHERE e.id = (
			SELECT id FROM countdown_editions
			WHERE status = $2
			   OR (status = $3 AND updated_at < NOW() - make_interval(secs => $1))
			ORDER BY edition_date, created_at, id
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		RETURNING e.id, e.series_id`, staleAfter.Seconds(), from, to).Scan(&editionID, &seriesID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	editions, err := r.listEditions(ctx, `e.id = $1`, []any{editionID}, 1)
	if err != nil {
		return nil, nil, err
	}
	if len(editions) == 0 {
		return nil, nil, fmt.Errorf("claimed countdown edition %s vanished", editionID)
	}
	series, err := r.GetSeries(ctx, seriesID)
	if err != nil {
		return nil, nil, err
	}
	return &editions[0], series, nil
}

// RecordMedia stores a fetched file and keeps the edition's claim fresh.
func (r *CountdownRepository) RecordMedia(ctx context.Context, m *models.CountdownEditionMedia) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	tag, err := tx.Exec(ctx, `UPDATE countdown_editions SET updated_at = NOW() WHERE id = $1 AND status = 'fetching'`, m.EditionID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrCountdownEditionNotFetching
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO countdown_edition_media (edition_id, clip_id, source, file_name, sha256, bytes, duration_seconds, width, height)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (edition_id, clip_id) DO UPDATE SET source = EXCLUDED.source, file_name = EXCLUDED.file_name,
			sha256 = EXCLUDED.sha256, bytes = EXCLUDED.bytes, duration_seconds = EXCLUDED.duration_seconds,
			width = EXCLUDED.width, height = EXCLUDED.height, fetched_at = NOW(), deleted_at = NULL`,
		m.EditionID, m.ClipID, m.Source, m.FileName, m.SHA256, m.Bytes, m.DurationSeconds, m.Width, m.Height); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// BackfillClip removes a clip whose media could not be fetched from the
// edition, records the skip, moves the clips below it up one place, and
// appends the next eligible clip at the end. It returns the new clip, or
// nil when no clip is left to take its place.
func (r *CountdownRepository) BackfillClip(ctx context.Context, edition *models.CountdownEdition, series *models.CountdownSeries, clipID uuid.UUID, reason string) (*models.CountdownEditionClip, error) {
	if len(reason) > 500 {
		reason = reason[:500]
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext('clpr:countdown:' || $1))`, edition.EditionDate); err != nil {
		return nil, err
	}
	var status string
	if err := tx.QueryRow(ctx, `SELECT status FROM countdown_editions WHERE id = $1 FOR UPDATE`, edition.ID).Scan(&status); err != nil {
		return nil, err
	}
	if status != models.CountdownEditionFetching {
		return nil, ErrCountdownEditionNotFetching
	}

	var position int
	err = tx.QueryRow(ctx, `DELETE FROM countdown_edition_clips WHERE edition_id = $1 AND clip_id = $2 RETURNING position`,
		edition.ID, clipID).Scan(&position)
	if err != nil {
		return nil, fmt.Errorf("remove skipped clip: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO countdown_edition_skips (edition_id, edition_date, clip_id, reason)
		VALUES ($1, $2::date, $3, $4) ON CONFLICT DO NOTHING`, edition.ID, edition.EditionDate, clipID, reason); err != nil {
		return nil, err
	}
	// Two steps through positions above any edition size (at most 50),
	// because positions are a primary key checked row by row.
	if _, err := tx.Exec(ctx, `UPDATE countdown_edition_clips SET position = position + 1000 WHERE edition_id = $1 AND position > $2`,
		edition.ID, position); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `UPDATE countdown_edition_clips SET position = position - 1001 WHERE edition_id = $1 AND position > 1000`,
		edition.ID); err != nil {
		return nil, err
	}

	var remaining int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM countdown_edition_clips WHERE edition_id = $1`, edition.ID).Scan(&remaining); err != nil {
		return nil, err
	}
	var next uuid.UUID
	var velocity float64
	query, args, err := countdownCandidates(series, edition.WindowStart, edition.WindowEnd, edition.EditionDate, 1, &edition.ID)
	if err == nil {
		err = tx.QueryRow(ctx, query, args...).Scan(&next, &velocity)
	}
	if errors.Is(err, pgx.ErrNoRows) || errors.Is(err, errCountdownRoomGone) {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("select backfill candidate: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO countdown_edition_clips (edition_id, edition_date, position, clip_id, view_velocity)
		VALUES ($1, $2::date, $3, $4, $5)`, edition.ID, edition.EditionDate, remaining+1, next, velocity); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `UPDATE countdown_editions SET updated_at = NOW() WHERE id = $1`, edition.ID); err != nil {
		return nil, err
	}
	clip := models.CountdownEditionClip{Position: remaining + 1, ClipID: next, ViewVelocity: velocity}
	if err := tx.QueryRow(ctx, `
		SELECT title, broadcaster_name, broadcaster_id, twitch_clip_id, twitch_clip_url, duration FROM clips WHERE id = $1`, next).
		Scan(&clip.Title, &clip.BroadcasterName, &clip.BroadcasterID, &clip.TwitchClipID, &clip.TwitchClipURL, &clip.Duration); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &clip, nil
}

// FinishFetch moves an edition out of "fetching" to "fetched", or to
// "failed" with a reason.
func (r *CountdownRepository) FinishFetch(ctx context.Context, editionID uuid.UUID, status string, reason *string) error {
	if status != models.CountdownEditionFetched && status != models.CountdownEditionFailed {
		return fmt.Errorf("invalid fetch outcome %q", status)
	}
	if reason != nil && len(*reason) > 500 {
		short := (*reason)[:500]
		reason = &short
	}
	tag, err := r.pool.Exec(ctx, `
		UPDATE countdown_editions SET status = $2, failure_reason = $3, updated_at = NOW()
		WHERE id = $1 AND status = 'fetching'`, editionID, status, reason)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrCountdownEditionNotFetching
	}
	return nil
}

// ListMediaToDelete returns stored files fetched before cutoff, and every
// stored file of a failed edition, that are not marked deleted yet.
func (r *CountdownRepository) ListMediaToDelete(ctx context.Context, cutoff time.Time) ([]models.CountdownEditionMedia, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT m.edition_id, m.clip_id, m.file_name
		FROM countdown_edition_media m
		JOIN countdown_editions e ON e.id = m.edition_id
		WHERE m.deleted_at IS NULL AND (m.fetched_at < $1 OR e.status = 'failed')
		ORDER BY m.fetched_at
		LIMIT 1000`, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.CountdownEditionMedia{}
	for rows.Next() {
		var m models.CountdownEditionMedia
		if err := rows.Scan(&m.EditionID, &m.ClipID, &m.FileName); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// MarkMediaDeleted records that a file was removed from the render host.
func (r *CountdownRepository) MarkMediaDeleted(ctx context.Context, editionID, clipID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE countdown_edition_media SET deleted_at = NOW()
		WHERE edition_id = $1 AND clip_id = $2 AND deleted_at IS NULL`, editionID, clipID)
	return err
}
