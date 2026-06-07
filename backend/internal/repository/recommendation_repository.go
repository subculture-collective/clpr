package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// RecommendationRepository handles database operations for recommendations
type RecommendationRepository struct {
	pool *pgxpool.Pool
}

// NewRecommendationRepository creates a new recommendation repository
func NewRecommendationRepository(pool *pgxpool.Pool) *RecommendationRepository {
	return &RecommendationRepository{
		pool: pool,
	}
}

// GetUserPreferences retrieves user preferences
func (r *RecommendationRepository) GetUserPreferences(ctx context.Context, userID uuid.UUID) (*models.UserPreference, error) {
	query := `
		SELECT user_id, favorite_games, followed_streamers, preferred_categories,
		       preferred_tags, onboarding_completed, onboarding_completed_at,
		       cold_start_source, updated_at, created_at
		FROM user_preferences
		WHERE user_id = $1
	`

	var pref models.UserPreference
	var favoriteGames, followedStreamers, preferredCategories []string
	var preferredTags []string

	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&pref.UserID,
		&favoriteGames,
		&followedStreamers,
		&preferredCategories,
		&preferredTags,
		&pref.OnboardingCompleted,
		&pref.OnboardingCompletedAt,
		&pref.ColdStartSource,
		&pref.UpdatedAt,
		&pref.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		// Return empty preferences for users without any
		return &models.UserPreference{
			UserID:              userID,
			FavoriteGames:       []string{},
			FollowedStreamers:   []string{},
			PreferredCategories: []string{},
			PreferredTags:       []uuid.UUID{},
			OnboardingCompleted: false,
			UpdatedAt:           time.Now(),
			CreatedAt:           time.Now(),
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}

	pref.FavoriteGames = favoriteGames
	pref.FollowedStreamers = followedStreamers
	pref.PreferredCategories = preferredCategories

	// Convert string UUIDs to uuid.UUID
	pref.PreferredTags = make([]uuid.UUID, 0, len(preferredTags))
	for _, tagStr := range preferredTags {
		tagID, err := uuid.Parse(tagStr)
		if err == nil {
			pref.PreferredTags = append(pref.PreferredTags, tagID)
		}
	}

	return &pref, nil
}

// UpdateUserPreferences updates or creates user preferences
func (r *RecommendationRepository) UpdateUserPreferences(ctx context.Context, pref *models.UserPreference) error {
	query := `
		INSERT INTO user_preferences (
			user_id, favorite_games, followed_streamers,
			preferred_categories, preferred_tags, updated_at, created_at
		)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (user_id)
		DO UPDATE SET
			favorite_games = COALESCE($2, user_preferences.favorite_games),
			followed_streamers = COALESCE($3, user_preferences.followed_streamers),
			preferred_categories = COALESCE($4, user_preferences.preferred_categories),
			preferred_tags = COALESCE($5, user_preferences.preferred_tags),
			updated_at = NOW()
	`

	// Convert uuid.UUID to strings for tags
	tagStrings := make([]string, len(pref.PreferredTags))
	for i, tag := range pref.PreferredTags {
		tagStrings[i] = tag.String()
	}

	_, err := r.pool.Exec(ctx, query,
		pref.UserID,
		pref.FavoriteGames,
		pref.FollowedStreamers,
		pref.PreferredCategories,
		tagStrings,
	)

	if err != nil {
		return fmt.Errorf("failed to update user preferences: %w", err)
	}

	return nil
}

// CompleteOnboarding marks onboarding as completed and saves initial preferences
func (r *RecommendationRepository) CompleteOnboarding(ctx context.Context, pref *models.UserPreference) error {
	// Convert uuid.UUID to strings for tags
	tagStrings := make([]string, len(pref.PreferredTags))
	for i, tag := range pref.PreferredTags {
		tagStrings[i] = tag.String()
	}

	query := `
		SELECT complete_user_onboarding($1, $2, $3, $4, $5)
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		pref.UserID,
		pref.FavoriteGames,
		pref.FollowedStreamers,
		pref.PreferredCategories,
		tagStrings,
	)

	if err != nil {
		return fmt.Errorf("failed to complete onboarding: %w", err)
	}

	return nil
}

// RecordInteraction records a user interaction with a clip
func (r *RecommendationRepository) RecordInteraction(ctx context.Context, interaction *models.UserClipInteraction) error {
	query := `
		INSERT INTO user_clip_interactions (
			id, user_id, clip_id, interaction_type, dwell_time, timestamp
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, clip_id, interaction_type)
		DO UPDATE SET
			dwell_time = COALESCE($5, user_clip_interactions.dwell_time),
			timestamp = $6
	`

	if interaction.ID == uuid.Nil {
		interaction.ID = uuid.New()
	}
	if interaction.Timestamp.IsZero() {
		interaction.Timestamp = time.Now()
	}

	_, err := r.pool.Exec(ctx, query,
		interaction.ID,
		interaction.UserID,
		interaction.ClipID,
		interaction.InteractionType,
		interaction.DwellTime,
		interaction.Timestamp,
	)

	if err != nil {
		return fmt.Errorf("failed to record interaction: %w", err)
	}

	return nil
}

// GetContentBasedRecommendations gets recommendations based on user's content preferences
func (r *RecommendationRepository) GetContentBasedRecommendations(
	ctx context.Context,
	userID uuid.UUID,
	preferences *models.UserPreference,
	excludeClipIDs []uuid.UUID,
	limit int,
) ([]models.ClipScore, error) {
	query := `
		WITH user_excluded AS (
			SELECT clip_id FROM user_clip_interactions
			WHERE user_id = $1 AND interaction_type IN ('like', 'view')
		),
		max_vote AS (
			SELECT MAX(vote_score) AS max_vote_score
			FROM clips
			WHERE created_at > NOW() - INTERVAL '7 days'
		),
		clip_with_tags AS (
			SELECT c.id,
			       ARRAY_AGG(DISTINCT ct.tag_id) FILTER (WHERE ct.tag_id IS NOT NULL) AS clip_tags
			FROM clips c
			LEFT JOIN clip_tags ct ON c.id = ct.clip_id
			WHERE c.created_at > NOW() - INTERVAL '30 days'
			  AND c.is_removed = false
			  AND c.dmca_removed = false
			  AND c.id NOT IN (SELECT clip_id FROM user_excluded)
			  AND ($6::uuid[] IS NULL OR c.id != ALL($6::uuid[]))
			GROUP BY c.id
		),
		scored_clips AS (
			SELECT
				c.id as clip_id,
				(
					CASE WHEN c.game_id = ANY($2::text[]) THEN 0.35 ELSE 0 END +
					CASE WHEN c.broadcaster_id = ANY($3::text[]) THEN 0.25 ELSE 0 END +
					CASE WHEN c.game_name = ANY($4::text[]) THEN 0.15 ELSE 0 END +
					CASE
						WHEN $5::uuid[] IS NOT NULL AND cwt.clip_tags && $5::uuid[] THEN 0.15
						ELSE 0
					END +
					(c.vote_score::float / NULLIF((SELECT max_vote_score FROM max_vote), 0)) * 0.1
				) as similarity_score
			FROM clips c
			JOIN clip_with_tags cwt ON c.id = cwt.id
		)
		SELECT
			clip_id,
			similarity_score,
			ROW_NUMBER() OVER (ORDER BY similarity_score DESC) as similarity_rank
		FROM scored_clips
		ORDER BY similarity_score DESC
		LIMIT $7
	`

	rows, err := r.pool.Query(ctx, query,
		userID,
		preferences.FavoriteGames,
		preferences.FollowedStreamers,
		preferences.PreferredCategories,
		preferences.PreferredTags,
		excludeClipIDs,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get content-based recommendations: %w", err)
	}
	defer rows.Close()

	var scores []models.ClipScore
	for rows.Next() {
		var score models.ClipScore
		if err := rows.Scan(&score.ClipID, &score.SimilarityScore, &score.SimilarityRank); err != nil {
			return nil, fmt.Errorf("failed to scan clip score: %w", err)
		}
		scores = append(scores, score)
	}

	return scores, nil
}

// GetCollaborativeRecommendations gets recommendations based on similar users
func (r *RecommendationRepository) GetCollaborativeRecommendations(
	ctx context.Context,
	userID uuid.UUID,
	excludeClipIDs []uuid.UUID,
	limit int,
) ([]models.ClipScore, error) {
	query := `
		WITH user_likes AS (
			SELECT clip_id FROM user_clip_interactions
			WHERE user_id = $1 AND interaction_type = 'like'
		),
		similar_users AS (
			SELECT
				uci.user_id,
				COUNT(*) as common_likes
			FROM user_clip_interactions uci
			WHERE uci.interaction_type = 'like'
			  AND uci.clip_id IN (SELECT clip_id FROM user_likes)
			  AND uci.user_id != $1
			GROUP BY uci.user_id
			ORDER BY common_likes DESC
			LIMIT 50
		),
		user_excluded AS (
			SELECT clip_id FROM user_clip_interactions
			WHERE user_id = $1
		)
		SELECT
			c.id as clip_id,
			COALESCE(COUNT(*) * 1.0 / NULLIF((SELECT COUNT(*) FROM similar_users), 0), 0) as similarity_score,
			ROW_NUMBER() OVER (ORDER BY COUNT(*) DESC) as similarity_rank
		FROM clips c
		JOIN user_clip_interactions uci ON c.id = uci.clip_id
		JOIN similar_users su ON uci.user_id = su.user_id
		WHERE uci.interaction_type = 'like'
		  AND c.created_at > NOW() - INTERVAL '30 days'
		  AND c.is_removed = false
		  AND c.dmca_removed = false
		  AND c.id NOT IN (SELECT clip_id FROM user_excluded)
		  AND ($2::uuid[] IS NULL OR c.id != ALL($2::uuid[]))
		GROUP BY c.id
		ORDER BY similarity_score DESC, c.vote_score DESC
		LIMIT $3
	`

	rows, err := r.pool.Query(ctx, query, userID, excludeClipIDs, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get collaborative recommendations: %w", err)
	}
	defer rows.Close()

	var scores []models.ClipScore
	for rows.Next() {
		var score models.ClipScore
		if err := rows.Scan(&score.ClipID, &score.SimilarityScore, &score.SimilarityRank); err != nil {
			return nil, fmt.Errorf("failed to scan clip score: %w", err)
		}
		scores = append(scores, score)
	}

	return scores, nil
}

// GetTrendingClips gets trending clips for cold start
func (r *RecommendationRepository) GetTrendingClips(
	ctx context.Context,
	excludeClipIDs []uuid.UUID,
	windowDays int,
	minScore float64,
	limit int,
) ([]models.ClipScore, error) {
	query := `
		SELECT
			id as clip_id,
			trending_score as similarity_score,
			ROW_NUMBER() OVER (ORDER BY trending_score DESC) as similarity_rank
		FROM clips
		WHERE created_at > NOW() - INTERVAL '1 day' * $1
		  AND is_removed = false
		  AND dmca_removed = false
		  AND trending_score > $2
		  AND ($3::uuid[] IS NULL OR id != ALL($3::uuid[]))
		ORDER BY trending_score DESC
		LIMIT $4
	`

	rows, err := r.pool.Query(ctx, query, windowDays, minScore, excludeClipIDs, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get trending clips: %w", err)
	}
	defer rows.Close()

	var scores []models.ClipScore
	for rows.Next() {
		var score models.ClipScore
		if err := rows.Scan(&score.ClipID, &score.SimilarityScore, &score.SimilarityRank); err != nil {
			return nil, fmt.Errorf("failed to scan clip score: %w", err)
		}
		scores = append(scores, score)
	}

	return scores, nil
}

// GetPopularClips gets popular clips for cold start fallback (new clips with good engagement)
func (r *RecommendationRepository) GetPopularClips(
	ctx context.Context,
	excludeClipIDs []uuid.UUID,
	windowDays int,
	minViews int,
	limit int,
) ([]models.ClipScore, error) {
	query := `
		WITH ranked_clips AS (
			SELECT
				id AS clip_id,
				(view_count::float / GREATEST(1, EXTRACT(EPOCH FROM (NOW() - created_at)) / 3600)) *
				(1 + (vote_score::float / GREATEST(1, view_count::float))) AS popularity_score
			FROM clips
			WHERE created_at > NOW() - INTERVAL '1 day' * $1
			  AND is_removed = false
			  AND dmca_removed = false
			  AND view_count >= $2
			  AND ($3::uuid[] IS NULL OR id != ALL($3::uuid[]))
		)
		SELECT
			clip_id,
			popularity_score AS similarity_score,
			ROW_NUMBER() OVER (ORDER BY popularity_score DESC) AS similarity_rank
		FROM ranked_clips
		ORDER BY popularity_score DESC
		LIMIT $4
	`

	rows, err := r.pool.Query(ctx, query, windowDays, minViews, excludeClipIDs, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular clips: %w", err)
	}
	defer rows.Close()

	var scores []models.ClipScore
	for rows.Next() {
		var score models.ClipScore
		if err := rows.Scan(&score.ClipID, &score.SimilarityScore, &score.SimilarityRank); err != nil {
			return nil, fmt.Errorf("failed to scan clip score: %w", err)
		}
		scores = append(scores, score)
	}

	return scores, nil
}

// GetClipsByIDs retrieves clips by their IDs while preserving order
func (r *RecommendationRepository) GetClipsByIDs(ctx context.Context, clipIDs []uuid.UUID) ([]models.Clip, error) {
	if len(clipIDs) == 0 {
		return []models.Clip{}, nil
	}

	query := `
		SELECT id, twitch_clip_id, twitch_clip_url, embed_url, title,
		       creator_name, creator_id, broadcaster_name, broadcaster_id,
		       game_id, game_name, language, thumbnail_url, duration,
		       view_count, created_at, imported_at, vote_score, comment_count,
		       favorite_count, is_featured, is_nsfw, is_removed, removed_reason,
		       trending_score, hot_score, popularity_index, engagement_count
		FROM clips
		WHERE id = ANY($1)
		ORDER BY array_position($1, id)
	`

	rows, err := r.pool.Query(ctx, query, clipIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get clips by IDs: %w", err)
	}
	defer rows.Close()

	var clips []models.Clip
	for rows.Next() {
		var clip models.Clip
		err := rows.Scan(
			&clip.ID, &clip.TwitchClipID, &clip.TwitchClipURL, &clip.EmbedURL, &clip.Title,
			&clip.CreatorName, &clip.CreatorID, &clip.BroadcasterName, &clip.BroadcasterID,
			&clip.GameID, &clip.GameName, &clip.Language, &clip.ThumbnailURL, &clip.Duration,
			&clip.ViewCount, &clip.CreatedAt, &clip.ImportedAt, &clip.VoteScore, &clip.CommentCount,
			&clip.FavoriteCount, &clip.IsFeatured, &clip.IsNSFW, &clip.IsRemoved, &clip.RemovedReason,
			&clip.TrendingScore, &clip.HotScore, &clip.PopularityIndex, &clip.EngagementCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan clip: %w", err)
		}
		clips = append(clips, clip)
	}

	return clips, nil
}

// HasUserInteractions checks if a user has any interaction history
func (r *RecommendationRepository) HasUserInteractions(ctx context.Context, userID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM user_clip_interactions
			WHERE user_id = $1
			LIMIT 1
		)
	`

	var exists bool
	err := r.pool.QueryRow(ctx, query, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user interactions: %w", err)
	}

	return exists, nil
}

// UpdateUserPreferencesFromInteractions triggers the stored procedure to update preferences
func (r *RecommendationRepository) UpdateUserPreferencesFromInteractions(ctx context.Context, userID uuid.UUID) error {
	query := `SELECT update_user_preferences_from_interactions($1)`

	_, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to update user preferences from interactions: %w", err)
	}

	return nil
}
