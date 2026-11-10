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

// ListDigestRecipients returns the broadcasters who feature their clips,
// have an email address, have email notifications and the digest turned
// on, are not left out of countdowns, and have no digest for the week yet.
func (r *CountdownRepository) ListDigestRecipients(ctx context.Context, weekStart time.Time) ([]models.StreamerDigestRecipient, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT ON (u.id) u.id, o.broadcaster_id, COALESCE(NULLIF(u.display_name, ''), u.username), u.email
		FROM broadcaster_feature_optins o
		JOIN users u ON u.id = o.user_id
		JOIN notification_preferences p ON p.user_id = u.id
		WHERE u.email IS NOT NULL AND u.email <> '' AND u.is_banned = false AND u.account_status = 'active'
		  AND p.email_enabled AND p.notify_streamer_digest AND p.email_digest <> 'never'
		  AND NOT EXISTS (SELECT 1 FROM countdown_exclusions x WHERE x.broadcaster_id = o.broadcaster_id)
		  AND NOT EXISTS (SELECT 1 FROM streamer_digest_sends d WHERE d.user_id = u.id AND d.week_start = $1::date)
		ORDER BY u.id, o.granted_at`, weekStart.UTC().Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.StreamerDigestRecipient{}
	for rows.Next() {
		var rcpt models.StreamerDigestRecipient
		if err := rows.Scan(&rcpt.UserID, &rcpt.BroadcasterID, &rcpt.Name, &rcpt.Email); err != nil {
			return nil, err
		}
		out = append(out, rcpt)
	}
	return out, rows.Err()
}

// ClaimDigest records that the user's digest for the week is being sent. It
// reports false when another run already claimed it.
func (r *CountdownRepository) ClaimDigest(ctx context.Context, userID uuid.UUID, weekStart time.Time) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
		INSERT INTO streamer_digest_sends (user_id, week_start) VALUES ($1, $2::date)
		ON CONFLICT (user_id, week_start) DO NOTHING`, userID, weekStart.UTC().Format("2006-01-02"))
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

// FinishDigest records how the user's digest for the week ended.
func (r *CountdownRepository) FinishDigest(ctx context.Context, userID uuid.UUID, weekStart time.Time, status string, reason *string) error {
	switch status {
	case models.StreamerDigestSent, models.StreamerDigestSkipped, models.StreamerDigestFailed:
	default:
		return fmt.Errorf("invalid digest status %q", status)
	}
	if reason != nil && len(*reason) > 500 {
		short := (*reason)[:500]
		reason = &short
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE streamer_digest_sends SET status = $3, reason = $4, updated_at = NOW()
		WHERE user_id = $1 AND week_start = $2::date`, userID, weekStart.UTC().Format("2006-01-02"), status, reason)
	return err
}

// StreamerDigestContent gathers the week's top clips, countdown placements
// and clip pack of a broadcaster, under the public-page rules.
func (r *CountdownRepository) StreamerDigestContent(ctx context.Context, broadcasterID string, weekStart time.Time) (*models.StreamerDigest, error) {
	start := time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 7)
	d := &models.StreamerDigest{BroadcasterID: broadcasterID, WeekStart: start.Format("2006-01-02"),
		TopClips: []models.StreamerDigestClip{}, Placements: []models.CountdownFeature{}}

	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT c.id, c.title, COALESCE(c.view_count, 0)
		FROM clips c
		WHERE c.broadcaster_id = $1 AND c.source_type = 'twitch' AND c.created_at >= $2 AND c.created_at < $3 AND %s
		ORDER BY COALESCE(c.view_count, 0) DESC, c.created_at, c.id
		LIMIT 5`, countdownPublicClip), broadcasterID, start, end)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var c models.StreamerDigestClip
		if err := rows.Scan(&c.ClipID, &c.Title, &c.ViewCount); err != nil {
			rows.Close()
			return nil, err
		}
		d.TopClips = append(d.TopClips, c)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	rows, err = r.pool.Query(ctx, fmt.Sprintf(`
		SELECT to_char(e.edition_date, 'YYYY-MM-DD'), s.slug, s.name, ec.position, c.id, c.title
		FROM countdown_edition_clips ec
		JOIN countdown_editions e ON e.id = ec.edition_id
		JOIN countdown_series s ON s.id = e.series_id
		JOIN clips c ON c.id = ec.clip_id
		WHERE %s AND %s AND c.broadcaster_id = $1 AND e.edition_date >= $2::date AND e.edition_date < $3::date
		ORDER BY e.edition_date, ec.position, s.priority
		LIMIT 20`, countdownPublicStatus, countdownPublicClip), broadcasterID, start.Format("2006-01-02"), end.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var f models.CountdownFeature
		if err := rows.Scan(&f.EditionDate, &f.SeriesSlug, &f.SeriesName, &f.Position, &f.ClipID, &f.Title); err != nil {
			rows.Close()
			return nil, err
		}
		d.Placements = append(d.Placements, f)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var clipCount *int
	err = r.pool.QueryRow(ctx, `SELECT status, clip_count FROM clip_packs WHERE broadcaster_id = $1 AND week_start = $2::date`,
		broadcasterID, start.Format("2006-01-02")).Scan(&d.PackStatus, &clipCount)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if clipCount != nil {
		d.PackClipCount = *clipCount
	}
	return d, nil
}
