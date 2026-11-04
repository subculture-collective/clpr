package repository

import (
	"context"
	"fmt"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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
		pref := &models.UserPreference{
			UserID:              userID,
			FavoriteGames:       []string{},
			FollowedStreamers:   []string{},
			PreferredCategories: []string{},
			PreferredTags:       []uuid.UUID{},
			OnboardingCompleted: false,
			UpdatedAt:           time.Now(),
			CreatedAt:           time.Now(),
		}
		pref.SyncCreatorFirstAliases()
		return pref, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}

	pref.FavoriteGames = favoriteGames
	pref.FollowedStreamers = followedStreamers
	pref.PreferredCategories = preferredCategories
	pref.SyncCreatorFirstAliases()

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
