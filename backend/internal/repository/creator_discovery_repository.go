package repository

import (
	"context"
	"fmt"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// ListCreatorDiscovery builds distinct creator rails from current clip
// momentum, growth relative to catalog size, and first-seen freshness.
func (r *BroadcasterRepository) ListCreatorDiscovery(
	ctx context.Context,
	limit int,
) (*models.CreatorDiscoveryRails, error) {
	if limit < 1 || limit > 24 {
		limit = 12
	}

	const query = `
		WITH creator_stats AS (
			SELECT
				c.broadcaster_id,
				MAX(c.broadcaster_name) AS broadcaster_name,
				COUNT(*)::INTEGER AS total_clips,
				COUNT(*) FILTER (WHERE c.created_at >= NOW() - INTERVAL '7 days')::INTEGER AS recent_clips,
				COALESCE(SUM(c.view_count), 0)::BIGINT AS total_views,
				COALESCE(SUM(c.view_count) FILTER (WHERE c.created_at >= NOW() - INTERVAL '7 days'), 0)::BIGINT AS recent_views,
				COALESCE(SUM(c.view_velocity) FILTER (WHERE c.created_at >= NOW() - INTERVAL '7 days'), 0)::DOUBLE PRECISION AS view_velocity,
				COALESCE(SUM(GREATEST(c.vote_score, 0) + c.comment_count * 2)
					FILTER (WHERE c.created_at >= NOW() - INTERVAL '7 days'), 0)::DOUBLE PRECISION AS recent_engagement,
				MIN(c.imported_at) AS first_discovered_at,
				MAX(c.created_at) AS latest_clip_at
			FROM clips c
			WHERE c.is_removed = FALSE
			  AND c.is_hidden = FALSE
			  AND COALESCE(c.dmca_removed, FALSE) = FALSE
			  AND c.broadcaster_id IS NOT NULL
			  AND c.broadcaster_id != ''
			  AND c.broadcaster_name != ''
			GROUP BY c.broadcaster_id
		), enriched AS (
			SELECT
				cs.*,
				COALESCE(f.followers, 0)::INTEGER AS follower_count,
				latest.thumbnail_url AS latest_clip_thumbnail,
				latest.title AS latest_clip_title,
				latest.game_name AS twitch_category_name,
				(
					cs.view_velocity
					+ SQRT(GREATEST(cs.recent_views, 0)) * 12
					+ cs.recent_engagement * 20
				) AS trending_score,
				(
					cs.view_velocity
					+ SQRT(GREATEST(cs.recent_views, 0)) * 4
					+ cs.recent_engagement * 30
				) / SQRT(GREATEST(cs.total_clips, 1)) AS rising_score
			FROM creator_stats cs
			LEFT JOIN LATERAL (
				SELECT COUNT(*) AS followers
				FROM broadcaster_follows bf
				WHERE bf.broadcaster_id = cs.broadcaster_id
			) f ON TRUE
			LEFT JOIN LATERAL (
				SELECT c.thumbnail_url, c.title, c.game_name
				FROM clips c
				WHERE c.broadcaster_id = cs.broadcaster_id
				  AND c.is_removed = FALSE
				  AND c.is_hidden = FALSE
				  AND COALESCE(c.dmca_removed, FALSE) = FALSE
				ORDER BY c.created_at DESC, c.imported_at DESC
				LIMIT 1
			) latest ON TRUE
		), ranked AS (
			SELECT
				e.*,
				ROW_NUMBER() OVER (ORDER BY e.trending_score DESC, e.latest_clip_at DESC) AS trending_rank,
				ROW_NUMBER() OVER (ORDER BY e.rising_score DESC, e.latest_clip_at DESC) AS rising_rank,
				ROW_NUMBER() OVER (ORDER BY e.first_discovered_at DESC, e.latest_clip_at DESC) AS new_rank
			FROM enriched e
			WHERE e.recent_clips > 0
		)
		SELECT section, broadcaster_id, broadcaster_name, total_clips, recent_clips,
		       total_views, recent_views, view_velocity, follower_count,
		       first_discovered_at, latest_clip_at, latest_clip_thumbnail,
		       latest_clip_title, twitch_category_name, score
		FROM (
			SELECT 'trending' AS section, r.*, r.trending_score AS score
			FROM ranked r WHERE r.trending_rank <= $1
			UNION ALL
			SELECT 'rising' AS section, r.*, r.rising_score AS score
			FROM ranked r WHERE r.rising_rank <= $1
			UNION ALL
			SELECT 'new' AS section, r.*, EXTRACT(EPOCH FROM r.first_discovered_at) AS score
			FROM ranked r
			WHERE r.new_rank <= $1
			  AND r.first_discovered_at >= NOW() - INTERVAL '7 days'
		) rails
		ORDER BY section, score DESC
	`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list creator discovery rails: %w", err)
	}
	defer rows.Close()

	rails := &models.CreatorDiscoveryRails{
		Trending: make([]models.CreatorDiscoveryProfile, 0, limit),
		Rising:   make([]models.CreatorDiscoveryProfile, 0, limit),
		New:      make([]models.CreatorDiscoveryProfile, 0, limit),
	}
	for rows.Next() {
		var section string
		var creator models.CreatorDiscoveryProfile
		if err := rows.Scan(
			&section,
			&creator.BroadcasterID,
			&creator.BroadcasterName,
			&creator.TotalClips,
			&creator.RecentClips,
			&creator.TotalViews,
			&creator.RecentViews,
			&creator.ViewVelocity,
			&creator.FollowerCount,
			&creator.FirstDiscoveredAt,
			&creator.LatestClipAt,
			&creator.LatestClipThumbnail,
			&creator.LatestClipTitle,
			&creator.TwitchCategoryName,
			&creator.Score,
		); err != nil {
			return nil, fmt.Errorf("failed to scan creator discovery profile: %w", err)
		}
		switch section {
		case "trending":
			rails.Trending = append(rails.Trending, creator)
		case "rising":
			rails.Rising = append(rails.Rising, creator)
		case "new":
			rails.New = append(rails.New, creator)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating creator discovery profiles: %w", err)
	}

	return rails, nil
}
