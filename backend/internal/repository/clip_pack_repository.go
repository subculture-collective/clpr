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

// ErrClipPackNotBuilding means the pack is no longer claimed for building.
var ErrClipPackNotBuilding = errors.New("clip pack is not being built")

// EnsureWeeklyPacks creates a pending pack for the week starting weekStart
// (a Monday, UTC) for every broadcaster who features their clips and had a
// clip that week, with their perPack most-viewed clips. Broadcasters left out
// of countdowns get none. A pack that already exists is left alone. It
// returns how many packs it created.
func (r *CountdownRepository) EnsureWeeklyPacks(ctx context.Context, weekStart time.Time, perPack int) (int, error) {
	start := time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, time.UTC)
	if start.Weekday() != time.Monday {
		return 0, fmt.Errorf("week start %s is not a Monday", start.Format("2006-01-02"))
	}
	var created int
	err := r.pool.QueryRow(ctx, `
		WITH eligible AS (
			SELECT o.broadcaster_id, c.id AS clip_id, left(c.broadcaster_name, 100) AS broadcaster_name,
			       ROW_NUMBER() OVER (
			           PARTITION BY o.broadcaster_id
			           ORDER BY COALESCE(c.view_count, 0) DESC, c.created_at, c.id
			       ) AS pos
			FROM broadcaster_feature_optins o
			JOIN clips c ON c.broadcaster_id = o.broadcaster_id
			WHERE c.source_type = 'twitch' AND c.created_at >= $1 AND c.created_at < $2
			  AND c.is_removed = false AND c.is_hidden = false AND COALESCE(c.dmca_removed, false) = false
			  AND c.broadcaster_name <> ''
			  AND NOT EXISTS (
			      SELECT 1 FROM countdown_exclusions x WHERE x.clip_id = c.id OR x.broadcaster_id = o.broadcaster_id
			  )
		), packs AS (
			INSERT INTO clip_packs (broadcaster_id, broadcaster_name, week_start)
			SELECT broadcaster_id, broadcaster_name, $1::date FROM eligible WHERE pos = 1
			ON CONFLICT (broadcaster_id, week_start) DO NOTHING
			RETURNING id, broadcaster_id
		), chosen AS (
			INSERT INTO clip_pack_clips (pack_id, position, clip_id)
			SELECT p.id, e.pos, e.clip_id FROM packs p JOIN eligible e ON e.broadcaster_id = p.broadcaster_id
			WHERE e.pos <= $3
		)
		SELECT count(*) FROM packs`, start, start.AddDate(0, 0, 7), perPack).Scan(&created)
	return created, err
}

// clipPackStillAllowed: the broadcaster still features their clips and is
// not left out of countdowns. p is the clip_packs row.
const clipPackStillAllowed = `EXISTS (SELECT 1 FROM broadcaster_feature_optins o WHERE o.broadcaster_id = p.broadcaster_id)
	AND NOT EXISTS (SELECT 1 FROM countdown_exclusions x WHERE x.broadcaster_id = p.broadcaster_id)`

// clipPackClipAllowed: a chosen clip may still go in the pack. c is the clip.
const clipPackClipAllowed = `c.is_removed = false AND c.is_hidden = false AND COALESCE(c.dmca_removed, false) = false
	AND NOT EXISTS (SELECT 1 FROM countdown_exclusions x WHERE x.clip_id = c.id)`

// ClaimPackForBuild moves the oldest pending pack to "building" and returns
// it with its clips that may still go in: clips removed, hidden, taken down
// or excluded since the pack was queued are left out, and a broadcaster who
// stopped featuring their clips or was excluded gets none. A claim older
// than staleAfter is taken again.
func (r *CountdownRepository) ClaimPackForBuild(ctx context.Context, staleAfter time.Duration) (*models.ClipPack, error) {
	var p models.ClipPack
	err := r.pool.QueryRow(ctx, `
		UPDATE clip_packs SET status = 'building', updated_at = NOW()
		WHERE id = (
			SELECT id FROM clip_packs
			WHERE status = 'pending' OR (status = 'building' AND updated_at < NOW() - make_interval(secs => $1))
			ORDER BY created_at, id
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, broadcaster_id, broadcaster_name, to_char(week_start, 'YYYY-MM-DD'), status`,
		staleAfter.Seconds()).Scan(&p.ID, &p.BroadcasterID, &p.BroadcasterName, &p.WeekStart, &p.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT pc.position, c.id, c.title, c.broadcaster_name, c.broadcaster_id, c.twitch_clip_id, c.twitch_clip_url,
		       c.duration, COALESCE(c.view_count, 0)
		FROM clip_pack_clips pc
		JOIN clip_packs p ON p.id = pc.pack_id
		JOIN clips c ON c.id = pc.clip_id
		WHERE pc.pack_id = $1 AND %s AND %s
		ORDER BY pc.position`, clipPackStillAllowed, clipPackClipAllowed), p.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var c models.ClipPackClip
		if err := rows.Scan(&c.Position, &c.ClipID, &c.Title, &c.BroadcasterName, &c.BroadcasterID, &c.TwitchClipID,
			&c.TwitchClipURL, &c.Duration, &c.ViewCount); err != nil {
			return nil, err
		}
		p.Clips = append(p.Clips, c)
	}
	return &p, rows.Err()
}

// CompletePack records a built pack's zip and when it expires.
func (r *CountdownRepository) CompletePack(ctx context.Context, id uuid.UUID, fileName, sha256 string, bytes int64, clipCount int, expiresAt time.Time) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE clip_packs
		SET status = 'ready', file_name = $2, sha256 = $3, bytes = $4, clip_count = $5,
		    ready_at = NOW(), expires_at = $6, failure_reason = NULL, updated_at = NOW()
		WHERE id = $1 AND status = 'building'`, id, fileName, sha256, bytes, clipCount, expiresAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrClipPackNotBuilding
	}
	return nil
}

// FailPack records why a pack could not be built.
func (r *CountdownRepository) FailPack(ctx context.Context, id uuid.UUID, reason string) error {
	if len(reason) > 500 {
		reason = reason[:500]
	}
	tag, err := r.pool.Exec(ctx, `
		UPDATE clip_packs SET status = 'failed', failure_reason = $2, updated_at = NOW()
		WHERE id = $1 AND status = 'building'`, id, reason)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrClipPackNotBuilding
	}
	return nil
}

// ListPacksToDelete returns the packs whose zip has expired by now and is
// still on the render host.
func (r *CountdownRepository) ListPacksToDelete(ctx context.Context, now time.Time) ([]models.ClipPack, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, file_name FROM clip_packs
		WHERE file_name IS NOT NULL AND deleted_at IS NULL AND expires_at <= $1
		ORDER BY expires_at`, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.ClipPack{}
	for rows.Next() {
		var p models.ClipPack
		if err := rows.Scan(&p.ID, &p.FileName); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// MarkPackDeleted records that the pack's zip was deleted; a ready pack
// becomes "expired".
func (r *CountdownRepository) MarkPackDeleted(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE clip_packs
		SET deleted_at = NOW(), updated_at = NOW(), status = CASE WHEN status = 'ready' THEN 'expired' ELSE status END
		WHERE id = $1`, id)
	return err
}

// clipPackWithdrawn: since the pack was made, the broadcaster stopped
// featuring their clips or was excluded, or a clip in it was removed,
// hidden, taken down or excluded. Such a pack is no longer served.
var clipPackWithdrawn = fmt.Sprintf(`NOT (%s) OR EXISTS (
	SELECT 1 FROM clip_pack_clips pc JOIN clips c ON c.id = pc.clip_id
	WHERE pc.pack_id = p.id AND NOT (%s))`, clipPackStillAllowed, clipPackClipAllowed)

// PackFile returns what the zip server needs about a pack, or nil when there
// is no such pack.
func (r *CountdownRepository) PackFile(ctx context.Context, id uuid.UUID) (*models.ClipPackFile, error) {
	var f models.ClipPackFile
	var name *string
	err := r.pool.QueryRow(ctx, `
		SELECT p.file_name, p.status, p.expires_at, p.deleted_at IS NOT NULL, `+clipPackWithdrawn+`
		FROM clip_packs p WHERE p.id = $1`, id).
		Scan(&name, &f.Status, &f.ExpiresAt, &f.Deleted, &f.Withdrawn)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if name != nil {
		f.FileName = *name
	}
	return &f, nil
}

// ListBroadcasterPacks returns the broadcaster's packs, newest week first.
func (r *CountdownRepository) ListBroadcasterPacks(ctx context.Context, broadcasterID string, limit int) ([]models.ClipPack, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT p.id, p.broadcaster_id, p.broadcaster_name, to_char(p.week_start, 'YYYY-MM-DD'), p.status, p.sha256, p.bytes,
		       p.clip_count, p.ready_at, p.expires_at, `+clipPackWithdrawn+`
		FROM clip_packs p
		WHERE p.broadcaster_id = $1
		ORDER BY p.week_start DESC
		LIMIT $2`, broadcasterID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.ClipPack{}
	for rows.Next() {
		var p models.ClipPack
		if err := rows.Scan(&p.ID, &p.BroadcasterID, &p.BroadcasterName, &p.WeekStart, &p.Status, &p.SHA256, &p.Bytes,
			&p.ClipCount, &p.ReadyAt, &p.ExpiresAt, &p.Withdrawn); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
