package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// PlaylistCurationRepository provides strategy-based clip queries for automated playlist curation.
type PlaylistCurationRepository struct {
	pool *pgxpool.Pool
}

// NewPlaylistCurationRepository creates a new PlaylistCurationRepository
func NewPlaylistCurationRepository(pool *pgxpool.Pool) *PlaylistCurationRepository {
	return &PlaylistCurationRepository{pool: pool}
}

// baseClipFilter returns the common WHERE clause fragment that all strategies share.
// It handles removed/hidden/nsfw filtering and optional game/broadcaster/tag/language filters.
func baseClipFilter(script *models.PlaylistScript) (string, []interface{}) {
	clause := "c.is_removed = false AND c.is_hidden = false"
	args := []interface{}{}
	idx := 1

	if script.ExcludeNSFW {
		clause += " AND c.is_nsfw = false"
	}
	if script.GameID != nil {
		clause += fmt.Sprintf(" AND c.game_id = $%d", idx)
		args = append(args, *script.GameID)
		idx++
	}
	if script.BroadcasterID != nil {
		clause += fmt.Sprintf(" AND c.broadcaster_id = $%d", idx)
		args = append(args, *script.BroadcasterID)
		idx++
	}
	if script.Tag != nil {
		clause += fmt.Sprintf(` AND EXISTS (
			SELECT 1 FROM clip_tags ct JOIN tags t ON t.id = ct.tag_id
			WHERE ct.clip_id = c.id AND t.slug = $%d
		)`, idx)
		args = append(args, *script.Tag)
		idx++
	}
	if len(script.ExcludeTags) > 0 {
		clause += fmt.Sprintf(` AND NOT EXISTS (
			SELECT 1 FROM clip_tags ct JOIN tags t ON t.id = ct.tag_id
			WHERE ct.clip_id = c.id AND t.slug = ANY($%d)
		)`, idx)
		args = append(args, script.ExcludeTags)
		idx++
	}
	if script.Language != nil {
		clause += fmt.Sprintf(" AND c.language = $%d", idx)
		args = append(args, *script.Language)
		idx++
	}
	if script.MinVoteScore != nil {
		clause += fmt.Sprintf(" AND c.vote_score >= $%d", idx)
		args = append(args, *script.MinVoteScore)
		idx++
	}
	if script.MinViewCount != nil {
		clause += fmt.Sprintf(" AND c.view_count >= $%d", idx)
		args = append(args, *script.MinViewCount)
		idx++
	}
	if script.Top10kStreamers {
		clause += " AND c.broadcaster_id IN (SELECT broadcaster_id FROM top_streamers)"
	}

	return clause, args
}

// timeframeClause returns a SQL time constraint clause based on the script's timeframe setting.
func timeframeClause(script *models.PlaylistScript) string {
	if script.Timeframe == nil {
		return ""
	}
	switch *script.Timeframe {
	case "hour":
		return " AND c.created_at > NOW() - INTERVAL '1 hour'"
	case "day":
		return " AND c.created_at > NOW() - INTERVAL '1 day'"
	case "week":
		return " AND c.created_at > NOW() - INTERVAL '7 days'"
	case "month":
		return " AND c.created_at > NOW() - INTERVAL '30 days'"
	case "year":
		return " AND c.created_at > NOW() - INTERVAL '365 days'"
	default:
		return ""
	}
}

// scanClipIDs scans rows returning only clip IDs
func (r *PlaylistCurationRepository) scanClipIDs(ctx context.Context, query string, args []interface{}) ([]models.Clip, error) {
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clips []models.Clip
	for rows.Next() {
		var clip models.Clip
		if err := rows.Scan(&clip.ID); err != nil {
			return nil, err
		}
		clips = append(clips, clip)
	}
	return clips, nil
}

// SleeperHits finds high retention/completion clips with low view counts — hidden gems.
func (r *PlaylistCurationRepository) SleeperHits(ctx context.Context, script *models.PlaylistScript) ([]models.Clip, error) {
	where, args := baseClipFilter(script)
	where += timeframeClause(script)
	nextArg := len(args) + 1

	query := fmt.Sprintf(`
		SELECT c.id
		FROM clips c
		JOIN clip_analytics ca ON ca.clip_id = c.id
		WHERE %s
		  AND ca.retention_rate IS NOT NULL
		  AND ca.retention_rate > 0.6
		  AND c.view_count < 500
		ORDER BY ca.retention_rate DESC, c.vote_score DESC
		LIMIT $%d
	`, where, nextArg)
	args = append(args, script.ClipLimit)

	return r.scanClipIDs(ctx, query, args)
}

// ViralVelocity finds clips with the fastest engagement growth rate in the time window.
func (r *PlaylistCurationRepository) ViralVelocity(ctx context.Context, script *models.PlaylistScript) ([]models.Clip, error) {
	where, args := baseClipFilter(script)
	where += timeframeClause(script)
	nextArg := len(args) + 1

	query := fmt.Sprintf(`
		SELECT c.id
		FROM clips c
		WHERE %s
		  AND c.created_at > NOW() - INTERVAL '48 hours'
		  AND EXTRACT(EPOCH FROM (NOW() - c.created_at)) / 3600 > 0.5
		ORDER BY (c.view_count + c.vote_score * 10 + c.comment_count * 5 + c.favorite_count * 3)::FLOAT
		         / (EXTRACT(EPOCH FROM (NOW() - c.created_at)) / 3600) DESC
		LIMIT $%d
	`, where, nextArg)
	args = append(args, script.ClipLimit)

	return r.scanClipIDs(ctx, query, args)
}

// CommunityFavorites finds clips with the highest save-to-view ratio.
func (r *PlaylistCurationRepository) CommunityFavorites(ctx context.Context, script *models.PlaylistScript) ([]models.Clip, error) {
	where, args := baseClipFilter(script)
	where += timeframeClause(script)
	nextArg := len(args) + 1

	query := fmt.Sprintf(`
		SELECT c.id
		FROM clips c
		WHERE %s
		  AND c.view_count > 10
		  AND c.favorite_count > 0
		ORDER BY c.favorite_count::FLOAT / c.view_count DESC, c.favorite_count DESC
		LIMIT $%d
	`, where, nextArg)
	args = append(args, script.ClipLimit)

	return r.scanClipIDs(ctx, query, args)
}

// DeepCuts finds clips with high dwell time and votes but not trending.
func (r *PlaylistCurationRepository) DeepCuts(ctx context.Context, script *models.PlaylistScript) ([]models.Clip, error) {
	where, args := baseClipFilter(script)
	where += timeframeClause(script)
	nextArg := len(args) + 1

	query := fmt.Sprintf(`
		SELECT c.id
		FROM clips c
		JOIN (
			SELECT clip_id, AVG(progress_seconds) AS avg_progress
			FROM watch_history
			GROUP BY clip_id
			HAVING AVG(progress_seconds) > 15
		) wh ON wh.clip_id = c.id
		WHERE %s
		  AND c.vote_score > 3
		  AND c.trending_score < 5
		ORDER BY wh.avg_progress DESC, c.vote_score DESC
		LIMIT $%d
	`, where, nextArg)
	args = append(args, script.ClipLimit)

	return r.scanClipIDs(ctx, query, args)
}

// FreshFaces finds top clips from new creators with 5 or fewer total clips.
func (r *PlaylistCurationRepository) FreshFaces(ctx context.Context, script *models.PlaylistScript) ([]models.Clip, error) {
	where, args := baseClipFilter(script)
	where += timeframeClause(script)
	nextArg := len(args) + 1

	query := fmt.Sprintf(`
		SELECT c.id
		FROM clips c
		JOIN (
			SELECT creator_id
			FROM clips
			WHERE is_removed = false AND creator_id IS NOT NULL
			GROUP BY creator_id
			HAVING COUNT(*) <= 5
		) new_creators ON new_creators.creator_id = c.creator_id
		WHERE %s
		ORDER BY c.vote_score DESC, c.view_count DESC
		LIMIT $%d
	`, where, nextArg)
	args = append(args, script.ClipLimit)

	return r.scanClipIDs(ctx, query, args)
}

// OnePerCreator finds a single standout clip per creator to maximize variety.
func (r *PlaylistCurationRepository) OnePerCreator(ctx context.Context, script *models.PlaylistScript) ([]models.Clip, error) {
	where, args := baseClipFilter(script)
	where += timeframeClause(script)
	nextArg := len(args) + 1

	query := fmt.Sprintf(`
		WITH ranked AS (
			SELECT c.id,
			       ROW_NUMBER() OVER (
				   PARTITION BY COALESCE(c.creator_id, c.broadcaster_id, c.id::text)
				   ORDER BY c.vote_score DESC, c.view_count DESC, c.created_at DESC
			   ) AS creator_rank
			FROM clips c
			WHERE %s
		)
		SELECT id
		FROM ranked
		WHERE creator_rank = 1
		LIMIT $%d
	`, where, nextArg)
	args = append(args, script.ClipLimit)

	return r.scanClipIDs(ctx, query, args)
}

// SimilarVibes finds clips semantically similar to a seed clip using pgvector cosine distance.
func (r *PlaylistCurationRepository) SimilarVibes(ctx context.Context, script *models.PlaylistScript, seedClipID uuid.UUID) ([]models.Clip, error) {
	where, args := baseClipFilter(script)
	nextArg := len(args) + 1

	query := fmt.Sprintf(`
		SELECT c.id
		FROM clips c,
		     (SELECT embedding FROM clips WHERE id = $%d AND embedding IS NOT NULL) seed
		WHERE %s
		  AND c.id != $%d
		  AND c.embedding IS NOT NULL
		ORDER BY c.embedding <=> seed.embedding ASC
		LIMIT $%d
	`, nextArg, where, nextArg, nextArg+1)
	args = append(args, seedClipID, script.ClipLimit)

	return r.scanClipIDs(ctx, query, args)
}

// CrossGameHits finds the best clips across multiple specified games.
func (r *PlaylistCurationRepository) CrossGameHits(ctx context.Context, script *models.PlaylistScript, gameIDs []string) ([]models.Clip, error) {
	where, args := baseClipFilter(script)
	where += timeframeClause(script)
	nextArg := len(args) + 1

	query := fmt.Sprintf(`
		SELECT c.id
		FROM clips c
		WHERE %s
		  AND c.game_id = ANY($%d)
		ORDER BY c.vote_score DESC, c.view_count DESC
		LIMIT $%d
	`, where, nextArg, nextArg+1)
	args = append(args, gameIDs, script.ClipLimit)

	return r.scanClipIDs(ctx, query, args)
}

// Controversial finds clips with high comment activity and polarizing engagement.
func (r *PlaylistCurationRepository) Controversial(ctx context.Context, script *models.PlaylistScript) ([]models.Clip, error) {
	where, args := baseClipFilter(script)
	where += timeframeClause(script)
	nextArg := len(args) + 1

	query := fmt.Sprintf(`
		SELECT c.id
		FROM clips c
		WHERE %s
		  AND c.comment_count > 3
		  AND c.view_count > 10
		ORDER BY c.comment_count::FLOAT * (c.comment_count::FLOAT / c.view_count) DESC
		LIMIT $%d
	`, where, nextArg)
	args = append(args, script.ClipLimit)

	return r.scanClipIDs(ctx, query, args)
}

// BingeWorthy finds clips from sessions where users watched 3+ clips.
func (r *PlaylistCurationRepository) BingeWorthy(ctx context.Context, script *models.PlaylistScript) ([]models.Clip, error) {
	where, args := baseClipFilter(script)
	where += timeframeClause(script)
	nextArg := len(args) + 1

	query := fmt.Sprintf(`
		SELECT c.id
		FROM clips c
		WHERE %s
		  AND c.id IN (
			SELECT wh.clip_id
			FROM watch_history wh
			WHERE wh.session_id IS NOT NULL
			  AND wh.session_id IN (
				SELECT session_id
				FROM watch_history
				WHERE session_id IS NOT NULL
				GROUP BY session_id
				HAVING COUNT(DISTINCT clip_id) >= 3
			  )
		  )
		ORDER BY c.view_count DESC, c.vote_score DESC
		LIMIT $%d
	`, where, nextArg)
	args = append(args, script.ClipLimit)

	return r.scanClipIDs(ctx, query, args)
}

// RisingStars finds creators whose recent engagement outpaces their overall average.
func (r *PlaylistCurationRepository) RisingStars(ctx context.Context, script *models.PlaylistScript) ([]models.Clip, error) {
	where, args := baseClipFilter(script)
	where += timeframeClause(script)
	nextArg := len(args) + 1

	query := fmt.Sprintf(`
		WITH creator_stats AS (
			SELECT creator_id,
			       AVG(vote_score) AS overall_avg
			FROM clips
			WHERE is_removed = false AND creator_id IS NOT NULL
			GROUP BY creator_id
			HAVING COUNT(*) >= 3
		),
		recent_stats AS (
			SELECT creator_id,
			       AVG(vote_score) AS recent_avg
			FROM clips
			WHERE is_removed = false
			  AND creator_id IS NOT NULL
			  AND created_at > NOW() - INTERVAL '30 days'
			GROUP BY creator_id
			HAVING COUNT(*) >= 1
		)
		SELECT c.id
		FROM clips c
		JOIN creator_stats cs ON cs.creator_id = c.creator_id
		JOIN recent_stats rs ON rs.creator_id = c.creator_id
		WHERE %s
		  AND c.created_at > NOW() - INTERVAL '30 days'
		  AND rs.recent_avg > cs.overall_avg * 1.5
		ORDER BY rs.recent_avg / GREATEST(cs.overall_avg, 1) DESC, c.vote_score DESC
		LIMIT $%d
	`, where, nextArg)
	args = append(args, script.ClipLimit)

	return r.scanClipIDs(ctx, query, args)
}
