package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Countdown eligibility bounds from ADR-002: editions use short excerpts.
const (
	CountdownMinClipSeconds = 5
	CountdownMaxClipSeconds = 60
)

var (
	ErrCountdownSeriesNotFound     = errors.New("countdown series not found")
	ErrCountdownSeriesSlugTaken    = errors.New("countdown series slug already exists")
	ErrCountdownEditionExists      = errors.New("countdown edition already exists for this date")
	ErrCountdownEditionInProgress  = errors.New("countdown edition has moved past selection")
	ErrCountdownExclusionExists    = errors.New("countdown exclusion already exists")
	ErrCountdownExclusionNotFound  = errors.New("countdown exclusion not found")
	ErrCountdownExclusionClipUnset = errors.New("countdown exclusion clip does not exist")
	ErrCountdownClipRoomNotFound   = errors.New("clip room not found")
)

// CountdownRepository stores countdown series, editions and exclusions.
type CountdownRepository struct {
	pool *pgxpool.Pool
}

// NewCountdownRepository creates a CountdownRepository.
func NewCountdownRepository(pool *pgxpool.Pool) *CountdownRepository {
	return &CountdownRepository{pool: pool}
}

const countdownSeriesColumns = `id, slug, name, game_id, language, size, window_hours, per_broadcaster_cap,
	selection_hour_utc, priority, enabled, shorts_count, min_size, source, clip_room_id,
	(SELECT r.twitch_channel FROM streamer_clip_rooms r WHERE r.id = countdown_series.clip_room_id),
	created_at, updated_at`

func scanCountdownSeries(row pgx.Row) (*models.CountdownSeries, error) {
	var s models.CountdownSeries
	err := row.Scan(&s.ID, &s.Slug, &s.Name, &s.GameID, &s.Language, &s.Size, &s.WindowHours,
		&s.PerBroadcasterCap, &s.SelectionHourUTC, &s.Priority, &s.Enabled, &s.ShortsCount, &s.MinSize,
		&s.Source, &s.ClipRoomID, &s.ClipRoomChannel, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrCountdownSeriesNotFound
	}
	return &s, err
}

// ListSeries returns every series in selection order.
func (r *CountdownRepository) ListSeries(ctx context.Context) ([]models.CountdownSeries, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+countdownSeriesColumns+` FROM countdown_series ORDER BY priority, created_at, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.CountdownSeries{}
	for rows.Next() {
		s, err := scanCountdownSeries(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *s)
	}
	return out, rows.Err()
}

// GetSeries returns one series.
func (r *CountdownRepository) GetSeries(ctx context.Context, id uuid.UUID) (*models.CountdownSeries, error) {
	return scanCountdownSeries(r.pool.QueryRow(ctx, `SELECT `+countdownSeriesColumns+` FROM countdown_series WHERE id = $1`, id))
}

func countdownSeriesWriteErr(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrCountdownSeriesSlugTaken
	}
	if errors.As(err, &pgErr) && pgErr.Code == "23503" && pgErr.ConstraintName == "countdown_series_clip_room_id_fkey" {
		return ErrCountdownClipRoomNotFound
	}
	return err
}

// seriesSource is the input's source, velocity when unset.
func seriesSource(in *models.CountdownSeriesInput) string {
	if in.Source == "" {
		return models.CountdownSourceVelocity
	}
	return in.Source
}

// CreateSeries inserts a series. The input must already be validated.
func (r *CountdownRepository) CreateSeries(ctx context.Context, in *models.CountdownSeriesInput) (*models.CountdownSeries, error) {
	s, err := scanCountdownSeries(r.pool.QueryRow(ctx, `
		INSERT INTO countdown_series (slug, name, game_id, language, size, window_hours, per_broadcaster_cap,
			selection_hour_utc, priority, enabled, shorts_count, min_size, source, clip_room_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING `+countdownSeriesColumns,
		in.Slug, in.Name, in.GameID, in.Language, in.Size, in.WindowHours, in.PerBroadcasterCap,
		in.SelectionHourUTC, in.Priority, in.Enabled, in.ShortsCount, in.MinSize, seriesSource(in), in.ClipRoomID))
	return s, countdownSeriesWriteErr(err)
}

// UpdateSeries replaces a series configuration. Existing editions keep the
// window they were selected with.
func (r *CountdownRepository) UpdateSeries(ctx context.Context, id uuid.UUID, in *models.CountdownSeriesInput) (*models.CountdownSeries, error) {
	s, err := scanCountdownSeries(r.pool.QueryRow(ctx, `
		UPDATE countdown_series SET slug = $2, name = $3, game_id = $4, language = $5, size = $6,
			window_hours = $7, per_broadcaster_cap = $8, selection_hour_utc = $9, priority = $10,
			enabled = $11, shorts_count = $12, min_size = $13, source = $14, clip_room_id = $15, updated_at = NOW()
		WHERE id = $1
		RETURNING `+countdownSeriesColumns,
		id, in.Slug, in.Name, in.GameID, in.Language, in.Size, in.WindowHours, in.PerBroadcasterCap,
		in.SelectionHourUTC, in.Priority, in.Enabled, in.ShortsCount, in.MinSize, seriesSource(in), in.ClipRoomID))
	return s, countdownSeriesWriteErr(err)
}

// EditionExists reports whether the series already has an edition for date
// (YYYY-MM-DD).
func (r *CountdownRepository) EditionExists(ctx context.Context, seriesID uuid.UUID, date string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM countdown_editions WHERE series_id = $1 AND edition_date = $2::date)`,
		seriesID, date).Scan(&exists)
	return exists, err
}

// countdownCandidatesQuery ranks the eligible clips of one series window by
// view_velocity, keeping at most per_broadcaster_cap clips per broadcaster
// and leaving out clips already in an edition that day or skipped that day
// because their media could not be fetched. When $10 names an edition (a
// backfill), its remaining clips count toward the per-broadcaster cap.
const countdownCandidatesQuery = `
	WITH eligible AS (
		SELECT c.id, c.view_velocity, c.view_count,
		       COALESCE(c.broadcaster_id, lower(c.broadcaster_name)) AS broadcaster_key,
		       ROW_NUMBER() OVER (
		           PARTITION BY COALESCE(c.broadcaster_id, lower(c.broadcaster_name))
		           ORDER BY c.view_velocity DESC, c.view_count DESC, c.id
		       ) AS broadcaster_rank
		FROM clips c
		WHERE c.source_type = 'twitch'
		  AND c.is_removed = false
		  AND c.is_hidden = false
		  AND c.is_nsfw = false
		  AND COALESCE(c.dmca_removed, false) = false
		  AND c.duration BETWEEN $1 AND $2
		  AND c.created_at >= $3 AND c.created_at < $4
		  AND c.view_velocity > 0
		  AND ($5::text IS NULL OR c.game_id = $5)
		  AND ($6::text IS NULL OR c.language = $6)
		  AND NOT EXISTS (
		      SELECT 1 FROM countdown_exclusions x
		      WHERE x.clip_id = c.id OR (c.broadcaster_id IS NOT NULL AND x.broadcaster_id = c.broadcaster_id)
		  )
		  AND NOT EXISTS (
		      SELECT 1 FROM countdown_edition_clips ec
		      WHERE ec.edition_date = $7::date AND ec.clip_id = c.id
		  )
		  AND NOT EXISTS (
		      SELECT 1 FROM countdown_edition_skips sk
		      WHERE sk.edition_date = $7::date AND sk.clip_id = c.id
		  )
	), taken AS (
		SELECT COALESCE(c.broadcaster_id, lower(c.broadcaster_name)) AS broadcaster_key, count(*) AS clips
		FROM countdown_edition_clips ec
		JOIN clips c ON c.id = ec.clip_id
		WHERE ec.edition_id = $10::uuid
		GROUP BY 1
	)
	SELECT e.id, e.view_velocity
	FROM eligible e
	LEFT JOIN taken t ON t.broadcaster_key = e.broadcaster_key
	WHERE e.broadcaster_rank + COALESCE(t.clips, 0) <= $8
	ORDER BY e.view_velocity DESC, e.view_count DESC, e.id
	LIMIT $9`

// countdownRoomCandidatesQuery lists the approved clips of a clip room
// series (#88) in the room's order, under the same clip rules as the
// velocity ranking apart from view velocity and the per-broadcaster cap: the
// streamer chose them. A clip approved twice counts once, at its first place.
const countdownRoomCandidatesQuery = `
	WITH items AS (
		SELECT i.clip_id, i.position, i.approved_at, i.id,
		       ROW_NUMBER() OVER (PARTITION BY i.clip_id ORDER BY i.position NULLS LAST, i.approved_at, i.id) AS copy
		FROM streamer_clip_room_items i
		WHERE i.room_id = $5 AND i.status = 'approved' AND i.clip_id IS NOT NULL
		  AND i.approved_at >= $3 AND i.approved_at < $4
	)
	SELECT c.id, COALESCE(c.view_velocity, 0)
	FROM items i
	JOIN clips c ON c.id = i.clip_id
	WHERE i.copy = 1
	  AND c.source_type = 'twitch'
	  AND c.is_removed = false
	  AND c.is_hidden = false
	  AND c.is_nsfw = false
	  AND COALESCE(c.dmca_removed, false) = false
	  AND c.duration BETWEEN $1 AND $2
	  AND NOT EXISTS (
	      SELECT 1 FROM countdown_exclusions x
	      WHERE x.clip_id = c.id OR (c.broadcaster_id IS NOT NULL AND x.broadcaster_id = c.broadcaster_id)
	  )
	  AND NOT EXISTS (
	      SELECT 1 FROM countdown_edition_clips ec
	      WHERE ec.edition_date = $6::date AND ec.clip_id = c.id
	  )
	  AND NOT EXISTS (
	      SELECT 1 FROM countdown_edition_skips sk
	      WHERE sk.edition_date = $6::date AND sk.clip_id = c.id
	  )
	ORDER BY i.position NULLS LAST, i.approved_at, i.id
	LIMIT $7`

// errCountdownRoomGone means a clip room series lost its room.
var errCountdownRoomGone = errors.New("the clip room of this series no longer exists")

// countdownCandidates returns the query and arguments that rank the
// candidates of series for date, at most limit. backfill names the edition
// whose clips count toward the per-broadcaster cap of a velocity series.
func countdownCandidates(series *models.CountdownSeries, windowStart, windowEnd time.Time, date string, limit int, backfill *uuid.UUID) (string, []any, error) {
	if series.Source == models.CountdownSourceClipRoom {
		if series.ClipRoomID == nil {
			return "", nil, errCountdownRoomGone
		}
		return countdownRoomCandidatesQuery, []any{CountdownMinClipSeconds, CountdownMaxClipSeconds,
			windowStart.UTC(), windowEnd.UTC(), *series.ClipRoomID, date, limit}, nil
	}
	return countdownCandidatesQuery, []any{CountdownMinClipSeconds, CountdownMaxClipSeconds, windowStart.UTC(), windowEnd.UTC(),
		series.GameID, series.Language, date, series.PerBroadcasterCap, limit, backfill}, nil
}

// SelectEdition picks and stores the edition of series for date, up to
// series.Size clips. A series with MinSize takes a shorter list on a thin day
// and stores a skipped edition below MinSize. A series without it stores a
// failed edition when fewer than Size clips are eligible, so a short list
// never goes out by accident. With replace, an existing edition that is still
// "selected", "failed" or "skipped" is selected again.
func (r *CountdownRepository) SelectEdition(ctx context.Context, series *models.CountdownSeries, date string, windowStart, windowEnd time.Time, replace bool) (*models.CountdownEdition, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// One selection per day at a time keeps the cross-series dedupe exact.
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext('clpr:countdown:' || $1))`, date); err != nil {
		return nil, err
	}

	var existingID uuid.UUID
	var existingStatus string
	err = tx.QueryRow(ctx, `SELECT id, status FROM countdown_editions WHERE series_id = $1 AND edition_date = $2::date`,
		series.ID, date).Scan(&existingID, &existingStatus)
	switch {
	case err == nil:
		if !replace {
			return nil, ErrCountdownEditionExists
		}
		if existingStatus != models.CountdownEditionSelected && existingStatus != models.CountdownEditionFailed &&
			existingStatus != models.CountdownEditionSkipped {
			return nil, ErrCountdownEditionInProgress
		}
		if _, err := tx.Exec(ctx, `DELETE FROM countdown_editions WHERE id = $1`, existingID); err != nil {
			return nil, err
		}
	case !errors.Is(err, pgx.ErrNoRows):
		return nil, err
	}

	type candidate struct {
		id       uuid.UUID
		velocity float64
	}
	var picked []candidate
	var failure *string
	query, args, err := countdownCandidates(series, windowStart, windowEnd, date, series.Size, nil)
	if errors.Is(err, errCountdownRoomGone) {
		reason := err.Error()
		failure = &reason
	} else {
		rows, err := tx.Query(ctx, query, args...)
		if err != nil {
			return nil, fmt.Errorf("select countdown candidates: %w", err)
		}
		for rows.Next() {
			var c candidate
			if err := rows.Scan(&c.id, &c.velocity); err != nil {
				rows.Close()
				return nil, err
			}
			picked = append(picked, c)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	status := models.CountdownEditionSelected
	required := series.Size
	if series.MinSize != nil {
		required = *series.MinSize
	}
	switch {
	case failure != nil:
		status = models.CountdownEditionFailed
	case len(picked) < required:
		status = models.CountdownEditionFailed
		if series.MinSize != nil {
			status = models.CountdownEditionSkipped
		}
		reason := fmt.Sprintf("only %d of %d clips were eligible", len(picked), required)
		failure = &reason
		picked = nil
	}

	var editionID uuid.UUID
	if err := tx.QueryRow(ctx, `
		INSERT INTO countdown_editions (series_id, edition_date, status, window_start, window_end, failure_reason)
		VALUES ($1, $2::date, $3, $4, $5, $6)
		RETURNING id`,
		series.ID, date, status, windowStart.UTC(), windowEnd.UTC(), failure).Scan(&editionID); err != nil {
		return nil, err
	}
	for i, c := range picked {
		if _, err := tx.Exec(ctx, `
			INSERT INTO countdown_edition_clips (edition_id, edition_date, position, clip_id, view_velocity)
			VALUES ($1, $2::date, $3, $4, $5)`,
			editionID, date, i+1, c.id, c.velocity); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	editions, err := r.listEditions(ctx, `e.id = $1`, []any{editionID}, 1)
	if err != nil {
		return nil, err
	}
	if len(editions) == 0 {
		return nil, fmt.Errorf("countdown edition %s vanished after insert", editionID)
	}
	return &editions[0], nil
}

// ListEditions returns editions, newest first, optionally for one date
// (YYYY-MM-DD) or one series, with their ranked clips.
func (r *CountdownRepository) ListEditions(ctx context.Context, date *string, seriesID *uuid.UUID, limit int) ([]models.CountdownEdition, error) {
	return r.listEditions(ctx, `($1::date IS NULL OR e.edition_date = $1::date) AND ($2::uuid IS NULL OR e.series_id = $2)`,
		[]any{date, seriesID}, limit)
}

func (r *CountdownRepository) listEditions(ctx context.Context, where string, args []any, limit int) ([]models.CountdownEdition, error) {
	args = append(args, limit)
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT e.id, e.series_id, s.slug, to_char(e.edition_date, 'YYYY-MM-DD'), e.status, e.window_start,
		       e.window_end, e.failure_reason, e.created_at, e.updated_at
		FROM countdown_editions e
		JOIN countdown_series s ON s.id = e.series_id
		WHERE %s
		ORDER BY e.edition_date DESC, s.priority, s.created_at, s.id
		LIMIT $%d`, where, len(args)), args...)
	if err != nil {
		return nil, err
	}
	editions := []models.CountdownEdition{}
	index := map[uuid.UUID]int{}
	var ids []uuid.UUID
	for rows.Next() {
		var e models.CountdownEdition
		if err := rows.Scan(&e.ID, &e.SeriesID, &e.SeriesSlug, &e.EditionDate, &e.Status, &e.WindowStart,
			&e.WindowEnd, &e.FailureReason, &e.CreatedAt, &e.UpdatedAt); err != nil {
			rows.Close()
			return nil, err
		}
		e.Clips = []models.CountdownEditionClip{}
		index[e.ID] = len(editions)
		ids = append(ids, e.ID)
		editions = append(editions, e)
	}
	rows.Close()
	if err := rows.Err(); err != nil || len(ids) == 0 {
		return editions, err
	}

	clipRows, err := r.pool.Query(ctx, `
		SELECT ec.edition_id, ec.position, ec.clip_id, ec.view_velocity, c.title, c.broadcaster_name,
		       c.broadcaster_id, c.twitch_clip_id, c.twitch_clip_url, c.duration
		FROM countdown_edition_clips ec
		JOIN clips c ON c.id = ec.clip_id
		WHERE ec.edition_id = ANY($1)
		ORDER BY ec.edition_id, ec.position`, ids)
	if err != nil {
		return nil, err
	}
	defer clipRows.Close()
	for clipRows.Next() {
		var editionID uuid.UUID
		var c models.CountdownEditionClip
		if err := clipRows.Scan(&editionID, &c.Position, &c.ClipID, &c.ViewVelocity, &c.Title,
			&c.BroadcasterName, &c.BroadcasterID, &c.TwitchClipID, &c.TwitchClipURL, &c.Duration); err != nil {
			return nil, err
		}
		e := &editions[index[editionID]]
		e.Clips = append(e.Clips, c)
	}
	return editions, clipRows.Err()
}

// ListExclusions returns every exclusion, newest first.
func (r *CountdownRepository) ListExclusions(ctx context.Context) ([]models.CountdownExclusion, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, broadcaster_id, clip_id, reason, source, note, created_by, created_at
		FROM countdown_exclusions ORDER BY created_at DESC, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.CountdownExclusion{}
	for rows.Next() {
		var x models.CountdownExclusion
		if err := rows.Scan(&x.ID, &x.BroadcasterID, &x.ClipID, &x.Reason, &x.Source, &x.Note, &x.CreatedBy, &x.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

// CreateExclusion stores an exclusion. The input must already be validated.
func (r *CountdownRepository) CreateExclusion(ctx context.Context, in *models.CountdownExclusionInput, source string, createdBy *uuid.UUID) (*models.CountdownExclusion, error) {
	var x models.CountdownExclusion
	err := r.pool.QueryRow(ctx, `
		INSERT INTO countdown_exclusions (broadcaster_id, clip_id, reason, source, note, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, broadcaster_id, clip_id, reason, source, note, created_by, created_at`,
		in.BroadcasterID, in.ClipID, in.Reason, source, in.Note, createdBy).
		Scan(&x.ID, &x.BroadcasterID, &x.ClipID, &x.Reason, &x.Source, &x.Note, &x.CreatedBy, &x.CreatedAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return nil, ErrCountdownExclusionExists
		case "23503":
			return nil, ErrCountdownExclusionClipUnset
		}
	}
	if err != nil {
		return nil, err
	}
	return &x, nil
}

// DeleteExclusion removes an exclusion.
func (r *CountdownRepository) DeleteExclusion(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM countdown_exclusions WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrCountdownExclusionNotFound
	}
	return nil
}
