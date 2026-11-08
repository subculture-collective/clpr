package repository

import (
	"context"
	"fmt"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
)

// Editions become public once rendered. Clips removed, hidden, taken down
// or excluded after the edition was made drop out of the public pages.
const (
	countdownPublicStatus = `e.status IN ('rendered', 'published')`
	countdownPublicClip   = `c.is_removed = false AND c.is_hidden = false AND COALESCE(c.dmca_removed, false) = false
		AND NOT EXISTS (
			SELECT 1 FROM countdown_exclusions x
			WHERE x.clip_id = c.id OR (c.broadcaster_id IS NOT NULL AND x.broadcaster_id = c.broadcaster_id)
		)`
)

// ListPublicEditions returns the public editions of the newest days dates
// before before (YYYY-MM-DD, optional), newest first, optionally for one
// series. next is the date to pass as before for the following page, or nil
// on the last page.
func (r *CountdownRepository) ListPublicEditions(ctx context.Context, seriesSlug, before *string, days int) (editions []models.PublicCountdownEdition, next *string, err error) {
	editions, err = r.listPublicEditions(ctx, fmt.Sprintf(`
		e.edition_date IN (
			SELECT DISTINCT e.edition_date
			FROM countdown_editions e
			JOIN countdown_series s ON s.id = e.series_id
			WHERE %s AND ($1::text IS NULL OR s.slug = $1) AND ($2::date IS NULL OR e.edition_date < $2::date)
			ORDER BY e.edition_date DESC
			LIMIT $3
		) AND ($1::text IS NULL OR s.slug = $1)`, countdownPublicStatus), seriesSlug, before, days)
	if err != nil || len(editions) == 0 {
		return editions, nil, err
	}
	oldest := editions[len(editions)-1].EditionDate
	var more bool
	if err := r.pool.QueryRow(ctx, fmt.Sprintf(`
		SELECT EXISTS (
			SELECT 1 FROM countdown_editions e
			JOIN countdown_series s ON s.id = e.series_id
			WHERE %s AND ($1::text IS NULL OR s.slug = $1) AND e.edition_date < $2::date
		)`, countdownPublicStatus), seriesSlug, oldest).Scan(&more); err != nil {
		return nil, nil, err
	}
	if more {
		next = &oldest
	}
	return editions, next, nil
}

// GetPublicEdition returns one public edition, or nil when there is none.
func (r *CountdownRepository) GetPublicEdition(ctx context.Context, date, seriesSlug string) (*models.PublicCountdownEdition, error) {
	editions, err := r.listPublicEditions(ctx, `e.edition_date = $1::date AND s.slug = $2`, date, seriesSlug)
	if err != nil || len(editions) == 0 {
		return nil, err
	}
	return &editions[0], nil
}

func (r *CountdownRepository) listPublicEditions(ctx context.Context, where string, args ...any) ([]models.PublicCountdownEdition, error) {
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT e.id, to_char(e.edition_date, 'YYYY-MM-DD'), s.slug, s.name,
		       (SELECT count(*) FROM countdown_edition_clips ec WHERE ec.edition_id = e.id)
		FROM countdown_editions e
		JOIN countdown_series s ON s.id = e.series_id
		WHERE %s AND %s
		ORDER BY e.edition_date DESC, s.priority, s.created_at, s.id`, countdownPublicStatus, where), args...)
	if err != nil {
		return nil, err
	}
	editions := []models.PublicCountdownEdition{}
	index := map[uuid.UUID]int{}
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		var e models.PublicCountdownEdition
		if err := rows.Scan(&id, &e.EditionDate, &e.SeriesSlug, &e.SeriesName, &e.Size); err != nil {
			rows.Close()
			return nil, err
		}
		e.Clips = []models.PublicCountdownClip{}
		index[id] = len(editions)
		ids = append(ids, id)
		editions = append(editions, e)
	}
	rows.Close()
	if err := rows.Err(); err != nil || len(ids) == 0 {
		return editions, err
	}

	clipRows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT ec.edition_id, ec.position, c.id, c.title, c.broadcaster_name, c.broadcaster_id,
		       c.twitch_clip_id, c.twitch_clip_url, c.thumbnail_url, c.duration
		FROM countdown_edition_clips ec
		JOIN clips c ON c.id = ec.clip_id
		WHERE ec.edition_id = ANY($1) AND %s
		ORDER BY ec.edition_id, ec.position`, countdownPublicClip), ids)
	if err != nil {
		return nil, err
	}
	defer clipRows.Close()
	for clipRows.Next() {
		var editionID uuid.UUID
		var c models.PublicCountdownClip
		if err := clipRows.Scan(&editionID, &c.Position, &c.ClipID, &c.Title, &c.BroadcasterName, &c.BroadcasterID,
			&c.TwitchClipID, &c.TwitchClipURL, &c.ThumbnailURL, &c.Duration); err != nil {
			return nil, err
		}
		e := &editions[index[editionID]]
		e.Clips = append(e.Clips, c)
	}
	return editions, clipRows.Err()
}

// ListCountdownFeatures returns the public editions a clip, or any clip of a
// broadcaster, placed in, newest first.
func (r *CountdownRepository) ListCountdownFeatures(ctx context.Context, clipID *uuid.UUID, broadcasterID *string, limit int) ([]models.CountdownFeature, error) {
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT to_char(e.edition_date, 'YYYY-MM-DD'), s.slug, s.name, ec.position, c.id, c.title
		FROM countdown_edition_clips ec
		JOIN countdown_editions e ON e.id = ec.edition_id
		JOIN countdown_series s ON s.id = e.series_id
		JOIN clips c ON c.id = ec.clip_id
		WHERE %s AND %s
		  AND ($1::uuid IS NULL OR ec.clip_id = $1)
		  AND ($2::text IS NULL OR c.broadcaster_id = $2)
		ORDER BY e.edition_date DESC, ec.position, s.priority
		LIMIT $3`, countdownPublicStatus, countdownPublicClip), clipID, broadcasterID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	features := []models.CountdownFeature{}
	for rows.Next() {
		var f models.CountdownFeature
		if err := rows.Scan(&f.EditionDate, &f.SeriesSlug, &f.SeriesName, &f.Position, &f.ClipID, &f.Title); err != nil {
			return nil, err
		}
		features = append(features, f)
	}
	return features, rows.Err()
}

// ListPublicEditionKeys returns the newest public edition pages for the
// sitemap.
func (r *CountdownRepository) ListPublicEditionKeys(ctx context.Context, limit int) ([]models.CountdownEditionKey, error) {
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT to_char(e.edition_date, 'YYYY-MM-DD'), s.slug
		FROM countdown_editions e
		JOIN countdown_series s ON s.id = e.series_id
		WHERE %s
		ORDER BY e.edition_date DESC, s.priority, s.slug
		LIMIT $1`, countdownPublicStatus), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	keys := []models.CountdownEditionKey{}
	for rows.Next() {
		var k models.CountdownEditionKey
		if err := rows.Scan(&k.EditionDate, &k.SeriesSlug); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}
