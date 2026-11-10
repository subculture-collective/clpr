package repository

import (
	"context"
	"fmt"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/jackc/pgx/v5"
)

// ListUploadsForStats returns the YouTube video IDs of uploads made after
// uploadedAfter whose view counts were last read before staleBefore, or never,
// oldest reading first.
func (r *CountdownRepository) ListUploadsForStats(ctx context.Context, uploadedAfter, staleBefore time.Time, limit int) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT youtube_video_id
		FROM countdown_uploads
		WHERE youtube_missing = FALSE AND uploaded_at >= $1
		  AND (stats_refreshed_at IS NULL OR stats_refreshed_at < $2)
		ORDER BY stats_refreshed_at NULLS FIRST, uploaded_at DESC
		LIMIT $3`, uploadedAfter, staleBefore, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// SaveUploadStats stores what YouTube reported for each video, and marks the
// videos it no longer returns as missing.
func (r *CountdownRepository) SaveUploadStats(ctx context.Context, stats []models.CountdownVideoStats, missing []string, at time.Time) error {
	batch := &pgx.Batch{}
	for _, s := range stats {
		var privacy *string
		switch s.PrivacyStatus {
		case "private", "unlisted", "public":
			privacy = &s.PrivacyStatus
		}
		batch.Queue(`
			UPDATE countdown_uploads
			SET view_count = $2, youtube_privacy_status = COALESCE($3, youtube_privacy_status), stats_refreshed_at = $4
			WHERE youtube_video_id = $1`, s.YouTubeVideoID, max(s.ViewCount, 0), privacy, at)
	}
	for _, id := range missing {
		batch.Queue(`UPDATE countdown_uploads SET youtube_missing = TRUE, stats_refreshed_at = $2 WHERE youtube_video_id = $1`, id, at)
	}
	if batch.Len() == 0 {
		return nil
	}
	return r.pool.SendBatch(ctx, batch).Close()
}

// countdownVideoVisible: viewers can open the video on YouTube. Before the
// worker first reads a video, a scheduled upload counts as public from its
// publish time.
const countdownVideoVisible = `u.youtube_missing = FALSE AND (
	u.youtube_privacy_status IN ('public', 'unlisted')
	OR (u.youtube_privacy_status IS NULL AND (u.privacy_status IN ('public', 'unlisted') OR u.publish_at <= NOW()))
)`

// ListBroadcasterPlacements returns the broadcaster's clips in public
// editions, newest first, with the countdown video and the Short each one
// appears in when viewers can open them.
func (r *CountdownRepository) ListBroadcasterPlacements(ctx context.Context, broadcasterID string, limit int) ([]models.CountdownPlacement, error) {
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT to_char(e.edition_date, 'YYYY-MM-DD'), s.slug, s.name, ec.position, c.id, c.title, c.thumbnail_url,
		       cv.youtube_video_id, cv.view_count, cv.stats_refreshed_at,
		       sv.youtube_video_id, sv.view_count, sv.stats_refreshed_at
		FROM countdown_edition_clips ec
		JOIN countdown_editions e ON e.id = ec.edition_id
		JOIN countdown_series s ON s.id = e.series_id
		JOIN clips c ON c.id = ec.clip_id
		LEFT JOIN LATERAL (
			SELECT u.youtube_video_id, u.view_count, u.stats_refreshed_at FROM countdown_uploads u
			WHERE u.edition_id = e.id AND u.kind = 'countdown' AND %[3]s
			ORDER BY u.uploaded_at DESC LIMIT 1
		) cv ON TRUE
		LEFT JOIN LATERAL (
			SELECT u.youtube_video_id, u.view_count, u.stats_refreshed_at FROM countdown_uploads u
			WHERE u.edition_id = e.id AND u.kind = 'short' AND u.clip_id = c.id AND %[3]s
			ORDER BY u.uploaded_at DESC LIMIT 1
		) sv ON TRUE
		WHERE %[1]s AND %[2]s AND c.broadcaster_id = $1
		ORDER BY e.edition_date DESC, ec.position, s.priority
		LIMIT $2`, countdownPublicStatus, countdownPublicClip, countdownVideoVisible), broadcasterID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	placements := []models.CountdownPlacement{}
	for rows.Next() {
		var p models.CountdownPlacement
		var countdownID, shortID *string
		var countdown, short models.CountdownVideo
		if err := rows.Scan(&p.EditionDate, &p.SeriesSlug, &p.SeriesName, &p.Position, &p.ClipID, &p.Title, &p.ThumbnailURL,
			&countdownID, &countdown.ViewCount, &countdown.ViewsUpdatedAt,
			&shortID, &short.ViewCount, &short.ViewsUpdatedAt); err != nil {
			return nil, err
		}
		if countdownID != nil {
			countdown.YouTubeVideoID = *countdownID
			p.Countdown = &countdown
		}
		if shortID != nil {
			short.YouTubeVideoID = *shortID
			p.Short = &short
		}
		placements = append(placements, p)
	}
	return placements, rows.Err()
}
