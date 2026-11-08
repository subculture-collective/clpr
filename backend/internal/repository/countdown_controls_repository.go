package repository

import (
	"context"
	"errors"
	"sort"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ErrCountdownRemovalNotFound means no uploaded video shows that clip under
// an exclusion.
var ErrCountdownRemovalNotFound = errors.New("countdown removal task not found")

// BroadcasterCountdownState returns a broadcaster's exclusion and opt-in.
func (r *CountdownRepository) BroadcasterCountdownState(ctx context.Context, broadcasterID string) (*models.BroadcasterCountdownState, error) {
	state := &models.BroadcasterCountdownState{BroadcasterID: broadcasterID, ConsentTextVersion: models.CountdownConsentTextVersion}
	var source string
	var at time.Time
	err := r.pool.QueryRow(ctx, `SELECT source, created_at FROM countdown_exclusions WHERE broadcaster_id = $1`, broadcasterID).Scan(&source, &at)
	switch {
	case err == nil:
		state.Excluded, state.ExclusionSource, state.ExcludedAt = true, &source, &at
	case !errors.Is(err, pgx.ErrNoRows):
		return nil, err
	}
	var granted time.Time
	err = r.pool.QueryRow(ctx, `SELECT granted_at FROM broadcaster_feature_optins WHERE broadcaster_id = $1`, broadcasterID).Scan(&granted)
	switch {
	case err == nil:
		state.Featured, state.FeaturedAt = true, &granted
	case !errors.Is(err, pgx.ErrNoRows):
		return nil, err
	}
	return state, nil
}

// SetSelfExclusion turns a broadcaster's own "leave my clips out" on or off.
// Turning it off removes only a self-service exclusion; one an admin made
// stays.
func (r *CountdownRepository) SetSelfExclusion(ctx context.Context, broadcasterID string, userID uuid.UUID, excluded bool) error {
	if !excluded {
		_, err := r.pool.Exec(ctx, `DELETE FROM countdown_exclusions WHERE broadcaster_id = $1 AND source = 'self_service'`, broadcasterID)
		return err
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO countdown_exclusions (broadcaster_id, reason, source, created_by, consent_text_version)
		VALUES ($1, 'request', 'self_service', $2, $3)
		ON CONFLICT (broadcaster_id) WHERE broadcaster_id IS NOT NULL DO NOTHING`,
		broadcasterID, userID, models.CountdownConsentTextVersion)
	return err
}

// RecordFeatureOptIn stores a broadcaster's "feature my clips" grant.
func (r *CountdownRepository) RecordFeatureOptIn(ctx context.Context, broadcasterID string, userID uuid.UUID, scope string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO broadcaster_feature_optins (broadcaster_id, user_id, scope, consent_text_version)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (broadcaster_id) DO UPDATE SET user_id = EXCLUDED.user_id, scope = EXCLUDED.scope,
			consent_text_version = EXCLUDED.consent_text_version, granted_at = NOW()`,
		broadcasterID, userID, scope, models.CountdownConsentTextVersion)
	return err
}

// DeleteFeatureOptIns removes the opt-ins a user recorded.
func (r *CountdownRepository) DeleteFeatureOptIns(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM broadcaster_feature_optins WHERE user_id = $1`, userID)
	return err
}

// ListRemovalTasks returns every clip in an uploaded video that an
// exclusion covers and no admin has resolved, the oldest due first.
func (r *CountdownRepository) ListRemovalTasks(ctx context.Context) ([]models.CountdownRemovalTask, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT ON (u.render_id, c.id)
		       u.render_id, u.kind, u.youtube_video_id, to_char(e.edition_date, 'YYYY-MM-DD'), s.slug,
		       c.id, c.title, c.broadcaster_name, c.broadcaster_id, x.reason, x.source, x.created_at
		FROM countdown_uploads u
		JOIN countdown_editions e ON e.id = u.edition_id
		JOIN countdown_series s ON s.id = e.series_id
		JOIN countdown_edition_clips ec ON ec.edition_id = u.edition_id AND (u.kind = 'countdown' OR ec.clip_id = u.clip_id)
		JOIN clips c ON c.id = ec.clip_id
		JOIN countdown_exclusions x ON x.clip_id = c.id OR (c.broadcaster_id IS NOT NULL AND x.broadcaster_id = c.broadcaster_id)
		WHERE NOT EXISTS (SELECT 1 FROM countdown_removals rm WHERE rm.render_id = u.render_id AND rm.clip_id = c.id)
		ORDER BY u.render_id, c.id, x.created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.CountdownRemovalTask{}
	for rows.Next() {
		var t models.CountdownRemovalTask
		if err := rows.Scan(&t.RenderID, &t.Kind, &t.YouTubeVideoID, &t.EditionDate, &t.SeriesSlug, &t.ClipID, &t.ClipTitle,
			&t.BroadcasterName, &t.BroadcasterID, &t.ExclusionReason, &t.ExclusionSource, &t.ExcludedAt); err != nil {
			return nil, err
		}
		t.DueAt = t.ExcludedAt.Add(7 * 24 * time.Hour)
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].DueAt.Before(out[j].DueAt) })
	return out, nil
}

// ResolveRemoval records how an admin handled a removal task.
func (r *CountdownRepository) ResolveRemoval(ctx context.Context, in *models.CountdownRemovalInput, adminID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `
		INSERT INTO countdown_removals (render_id, clip_id, resolution, note, resolved_by)
		SELECT $1, $2, $3, $4, $5
		WHERE EXISTS (SELECT 1 FROM countdown_uploads WHERE render_id = $1)
		ON CONFLICT (render_id, clip_id) DO UPDATE SET resolution = EXCLUDED.resolution, note = EXCLUDED.note,
			resolved_by = EXCLUDED.resolved_by, resolved_at = NOW()`,
		in.RenderID, in.ClipID, in.Resolution, in.Note, adminID)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return ErrCountdownRemovalNotFound
	}
	if err == nil && tag.RowsAffected() == 0 {
		return ErrCountdownRemovalNotFound
	}
	return err
}
