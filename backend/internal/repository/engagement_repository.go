package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrRankingExpired = errors.New("ranking expired: refresh the feed")

type EngagementMetadata struct {
	Generation          uuid.UUID  `json:"generation"`
	Period              string     `json:"period"`
	PublishedAt         time.Time  `json:"published_at"`
	TrackingStartedAt   time.Time  `json:"tracking_started_at"`
	CoverageStartedAt   time.Time  `json:"coverage_started_at"`
	WindowStart         time.Time  `json:"window_start"`
	OldestObservationAt *time.Time `json:"oldest_observation_at"`
	StaleClips          int64      `json:"stale_clips"`
	EligibleClips       int64      `json:"eligible_clips"`
	PartialCoverage     bool       `json:"partial_coverage"`
	Estimated           bool       `json:"estimated"`
}

func EngagementWindow(period string) (time.Duration, bool) {
	switch period {
	case "hour":
		return time.Hour, true
	case "day":
		return 24 * time.Hour, true
	case "week":
		return 7 * 24 * time.Hour, true
	case "month":
		return 30 * 24 * time.Hour, true
	case "year":
		return 365 * 24 * time.Hour, true
	case "all":
		return 0, true
	default:
		return 0, false
	}
}

type EngagementRepository struct{ pool *pgxpool.Pool }

func NewEngagementRepository(pool *pgxpool.Pool) *EngagementRepository {
	return &EngagementRepository{pool: pool}
}

// Resolve returns only a complete, retained generation. The timestamp captured
// by Publish is the common window end for every page of that generation.
func (r *EngagementRepository) Resolve(ctx context.Context, period string, id *uuid.UUID) (*EngagementMetadata, error) {
	duration, ok := EngagementWindow(period)
	if !ok {
		return nil, fmt.Errorf("invalid engagement period")
	}
	m := &EngagementMetadata{Period: period, Estimated: true}
	err := r.pool.QueryRow(ctx, `SELECT id,published_at,tracking_started_at,coverage_started_at,oldest_observation_at,stale_clips,eligible_clips
 FROM engagement_generations WHERE published_at > now()-interval '1 hour' AND ($1::uuid IS NULL OR id=$1)
 ORDER BY published_at DESC LIMIT 1`, id).Scan(&m.Generation, &m.PublishedAt, &m.TrackingStartedAt, &m.CoverageStartedAt, &m.OldestObservationAt, &m.StaleClips, &m.EligibleClips)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrRankingExpired
	}
	if err != nil {
		return nil, err
	}
	m.WindowStart = m.PublishedAt.Add(-duration)
	if period == "all" {
		m.WindowStart = m.TrackingStartedAt
	}
	m.PartialCoverage = m.WindowStart.Before(m.TrackingStartedAt) || m.WindowStart.Before(m.CoverageStartedAt) || m.StaleClips > 0
	return m, nil
}

// Publish atomically exposes six immutable rankings. An advisory transaction
// lock prevents two API instances from publishing the same scheduled refresh.
func (r *EngagementRepository) Publish(ctx context.Context) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	// Publishing is a background batch over every tracked clip. It gets its own
	// statement budget instead of the 15 s pool default meant for requests.
	if _, err = tx.Exec(ctx, `SET LOCAL statement_timeout = '120s'`); err != nil {
		return err
	}
	var locked bool
	if err = tx.QueryRow(ctx, `SELECT pg_try_advisory_xact_lock(141001)`).Scan(&locked); err != nil {
		return err
	}
	if !locked {
		return nil
	}
	var recent bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM engagement_generations WHERE published_at>now()-interval '15 minutes')`).Scan(&recent); err != nil {
		return err
	}
	if recent {
		return nil
	}
	var id uuid.UUID
	err = tx.QueryRow(ctx, `INSERT INTO engagement_generations(published_at,tracking_started_at,coverage_started_at,oldest_observation_at,stale_clips,eligible_clips)
 SELECT now(),(SELECT started_at FROM engagement_tracking),
 COALESCE(max(s.first_observed_at) FILTER(WHERE c.twitch_clip_id IS NOT NULL),(SELECT started_at FROM engagement_tracking)),min(s.observed_at),
 count(*) FILTER(WHERE c.twitch_clip_id IS NOT NULL AND (s.observed_at IS NULL OR s.observed_at<now()-interval '25 hours' OR (s.active_until>now() AND s.observed_at<now()-interval '30 minutes') OR s.next_poll_at<now()-interval '15 minutes' OR s.last_poll_error_at IS NOT NULL)),count(*)
 FROM clips c LEFT JOIN clip_engagement_state s ON s.clip_id=c.id
 WHERE NOT c.is_hidden AND NOT c.is_removed
 RETURNING id`).Scan(&id)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO engagement_rankings(generation_id,period,clip_id,score)
 WITH periods(period,duration) AS (VALUES ('hour',interval '1 hour'),('day',interval '1 day'),('week',interval '7 days'),('month',interval '30 days'),('year',interval '365 days')),
 windows AS (SELECT period,now()-duration AS start_at,
 date_trunc('hour',now()-duration,'UTC')+interval '1 hour' AS interior_start,
 date_trunc('hour',now(),'UTC') AS interior_end FROM periods),
 interior AS (SELECT (SELECT interior_start FROM windows WHERE period='hour') AS hour_start,
 (SELECT interior_start FROM windows WHERE period='day') AS day_start,
 (SELECT interior_start FROM windows WHERE period='week') AS week_start,
 (SELECT interior_start FROM windows WHERE period='month') AS month_start,
 (SELECT interior_start FROM windows WHERE period='year') AS year_start,
 date_trunc('hour',now(),'UTC') AS end_at),
 hourly AS (
 -- One pass over complete hours: every window's interior total per clip.
 SELECT h.clip_id,
 sum(h.view_gain+2*h.vote_gain+3*h.comment_gain+2*h.favorite_gain) FILTER (WHERE h.bucket_start>=i.hour_start) AS hour_score,
 sum(h.view_gain+2*h.vote_gain+3*h.comment_gain+2*h.favorite_gain) FILTER (WHERE h.bucket_start>=i.day_start) AS day_score,
 sum(h.view_gain+2*h.vote_gain+3*h.comment_gain+2*h.favorite_gain) FILTER (WHERE h.bucket_start>=i.week_start) AS week_score,
 sum(h.view_gain+2*h.vote_gain+3*h.comment_gain+2*h.favorite_gain) FILTER (WHERE h.bucket_start>=i.month_start) AS month_score,
 sum(h.view_gain+2*h.vote_gain+3*h.comment_gain+2*h.favorite_gain) AS year_score
 FROM interior i JOIN clip_engagement_hourly h ON h.bucket_start>=i.year_start AND h.bucket_start<i.end_at
 GROUP BY h.clip_id),
 contributions AS (
 SELECT p.period,h.clip_id,p.score FROM hourly h
 CROSS JOIN LATERAL (VALUES ('hour',h.hour_score),('day',h.day_score),('week',h.week_score),('month',h.month_score),('year',h.year_score)) AS p(period,score)
 WHERE p.score IS NOT NULL
 UNION ALL
 -- Observations only matter where they overlap a window's partial first or last hour.
 SELECT w.period,o.clip_id,o.view_gain * (
 GREATEST(0,EXTRACT(EPOCH FROM (LEAST(o.observed_at,w.interior_start)-GREATEST(o.interval_start,w.start_at))))+
 GREATEST(0,EXTRACT(EPOCH FROM (LEAST(o.observed_at,now())-GREATEST(o.interval_start,w.interior_end)))))
 /NULLIF(EXTRACT(EPOCH FROM (o.observed_at-o.interval_start)),0)
 FROM windows w JOIN clip_view_observations o ON o.observed_at>w.start_at AND o.interval_start<now() AND o.view_gain>0
 AND (o.interval_start<w.interior_start OR o.observed_at>w.interior_end)
 UNION ALL
 SELECT w.period,e.clip_id,e.score_delta FROM windows w JOIN clip_local_engagement e
 ON e.occurred_at>=w.start_at AND e.occurred_at<now()
 AND (e.occurred_at<w.interior_start OR e.occurred_at>=w.interior_end)),
 weighted AS (
 SELECT period,clip_id,sum(score) AS score FROM contributions GROUP BY period,clip_id
 UNION ALL SELECT 'all',clip_id,view_gain+2*vote_gain+3*comment_gain+2*favorite_gain FROM clip_engagement_state)
 SELECT $1,w.period,w.clip_id,w.score FROM weighted w JOIN clips c ON c.id=w.clip_id
 WHERE w.score>0 AND NOT c.is_removed AND NOT c.is_hidden`, id)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `DELETE FROM engagement_generations WHERE published_at<=now()-interval '1 hour';
 DELETE FROM clip_engagement_hourly WHERE bucket_start<now()-interval '400 days';
 DELETE FROM clip_view_observations WHERE observed_at<now()-interval '400 days';
 DELETE FROM clip_local_engagement WHERE occurred_at<now()-interval '400 days'`)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

type EngagementPollClip struct {
	ID       uuid.UUID
	TwitchID string
}

// ClaimDue persists a short lease before network I/O, so interruption is
// resumable and concurrent workers do not repeatedly fetch the same batch.
func (r *EngagementRepository) ClaimDue(ctx context.Context, limit int) ([]EngagementPollClip, error) {
	if limit < 1 || limit > 100 {
		return nil, fmt.Errorf("invalid poll batch size")
	}
	_, err := r.pool.Exec(ctx, `INSERT INTO clip_engagement_state(clip_id)
 SELECT id FROM clips WHERE twitch_clip_id IS NOT NULL AND NOT is_removed AND NOT is_hidden
 AND NOT EXISTS(SELECT 1 FROM clip_engagement_state s WHERE s.clip_id=clips.id) ON CONFLICT DO NOTHING`)
	if err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, `WITH due AS (
 SELECT s.clip_id FROM clip_engagement_state s JOIN clips c ON c.id=s.clip_id
 WHERE s.next_poll_at<=now() AND NOT c.is_removed AND NOT c.is_hidden AND c.twitch_clip_id IS NOT NULL
 ORDER BY s.next_poll_at,s.clip_id FOR UPDATE OF s SKIP LOCKED LIMIT $1),
 claimed AS(UPDATE clip_engagement_state SET next_poll_at=now()+interval '5 minutes' WHERE clip_id IN(SELECT clip_id FROM due) RETURNING clip_id)
 SELECT c.id,c.twitch_clip_id FROM claimed JOIN clips c ON c.id=claimed.clip_id`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []EngagementPollClip{}
	for rows.Next() {
		var c EngagementPollClip
		if err = rows.Scan(&c.ID, &c.TwitchID); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

func (r *EngagementRepository) Observe(ctx context.Context, id uuid.UUID, count int, at time.Time) error {
	_, err := r.pool.Exec(ctx, `SELECT record_twitch_observation($1,$2,$3)`, id, count, at)
	return err
}
func (r *EngagementRepository) PollFailed(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE clip_engagement_state SET last_poll_error_at=now(),next_poll_at=now()+interval '15 minutes' WHERE clip_id=$1`, id)
	return err
}
