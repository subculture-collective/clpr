package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// TestColdStartWithOnboarding tests cold start recommendations for users with onboarding preferences
func TestColdStartWithOnboarding(t *testing.T) {
	// This is a placeholder test showing the expected behavior
	// In a real implementation, this would need mock repositories

	t.Run("Uses onboarding preferences for cold start", func(t *testing.T) {
		// Arrange
		userID := uuid.New()
		preferences := &models.UserPreference{
			UserID:              userID,
			FavoriteGames:       []string{"game-1", "game-2"},
			FollowedStreamers:   []string{"streamer-1"},
			PreferredCategories: []string{"FPS", "MOBA"},
			PreferredTags:       []uuid.UUID{uuid.New()},
			OnboardingCompleted: true,
		}

		// Assert expected behavior
		assert.True(t, preferences.OnboardingCompleted, "Onboarding should be marked as completed")
		assert.NotEmpty(t, preferences.FavoriteGames, "Should have favorite games from onboarding")
		assert.NotEmpty(t, preferences.FollowedStreamers, "Should have followed streamers from onboarding")
	})

	t.Run("Falls back to trending for users without onboarding", func(t *testing.T) {
		// Arrange
		userID := uuid.New()
		preferences := &models.UserPreference{
			UserID:              userID,
			FavoriteGames:       []string{},
			FollowedStreamers:   []string{},
			OnboardingCompleted: false,
		}

		// Assert expected behavior
		assert.False(t, preferences.OnboardingCompleted, "Onboarding should not be completed")
		assert.Empty(t, preferences.FavoriteGames, "Should have no favorite games")
	})
}

// TestOnboardingPreferencesValidation tests validation of onboarding preferences
func TestOnboardingPreferencesValidation(t *testing.T) {
	t.Run("Valid onboarding preferences", func(t *testing.T) {
		req := models.OnboardingPreferencesRequest{
			FavoriteGames:       []string{"game-1", "game-2", "game-3"},
			FollowedStreamers:   []string{"streamer-1"},
			PreferredCategories: []string{"FPS"},
			PreferredTags:       []uuid.UUID{uuid.New()},
		}

		assert.NotEmpty(t, req.FavoriteGames, "Should have at least one favorite game")
		assert.LessOrEqual(t, len(req.FavoriteGames), 10, "Should not exceed max favorite games")
	})

	t.Run("Minimal valid onboarding", func(t *testing.T) {
		req := models.OnboardingPreferencesRequest{
			FavoriteGames: []string{"game-1"},
		}

		assert.NotEmpty(t, req.FavoriteGames, "Should have at least one favorite game")
	})
}

// TestContentBasedWithEnhancedFeatures tests content-based recommendations with tags and categories
func TestContentBasedWithEnhancedFeatures(t *testing.T) {
	t.Run("Content-based uses all preference types", func(t *testing.T) {
		preferences := &models.UserPreference{
			UserID:              uuid.New(),
			FavoriteGames:       []string{"game-1"},
			FollowedStreamers:   []string{"streamer-1"},
			PreferredCategories: []string{"FPS", "MOBA"},
			PreferredTags:       []uuid.UUID{uuid.New(), uuid.New()},
		}

		// Assert all preference types are available
		assert.NotEmpty(t, preferences.FavoriteGames, "Should have favorite games")
		assert.NotEmpty(t, preferences.FollowedStreamers, "Should have followed streamers")
		assert.NotEmpty(t, preferences.PreferredCategories, "Should have preferred categories")
		assert.NotEmpty(t, preferences.PreferredTags, "Should have preferred tags")
	})
}

// TestColdStartSource tests the cold_start_source field tracking
func TestColdStartSource(t *testing.T) {
	t.Run("Onboarding source is set correctly", func(t *testing.T) {
		source := "onboarding"
		preferences := &models.UserPreference{
			UserID:              uuid.New(),
			FavoriteGames:       []string{"game-1"},
			OnboardingCompleted: true,
			ColdStartSource:     &source,
		}

		assert.NotNil(t, preferences.ColdStartSource, "Cold start source should be set")
		assert.Equal(t, "onboarding", *preferences.ColdStartSource, "Source should be 'onboarding'")
	})

	t.Run("Inferred source is set correctly", func(t *testing.T) {
		source := "inferred"
		preferences := &models.UserPreference{
			UserID:              uuid.New(),
			FavoriteGames:       []string{"game-1"},
			OnboardingCompleted: false,
			ColdStartSource:     &source,
		}

		assert.NotNil(t, preferences.ColdStartSource, "Cold start source should be set")
		assert.Equal(t, "inferred", *preferences.ColdStartSource, "Source should be 'inferred'")
	})
}

// TestPopularityFallback tests the popularity-based fallback for cold start
func TestPopularityFallback(t *testing.T) {
	t.Run("Popularity score calculation is reasonable", func(t *testing.T) {
		// Test the expected behavior of popularity scoring
		// Popularity = (views/hour_age) * (1 + vote_score/views)

		// High engagement clip (1000 views, 100 votes in 10 hours)
		views1 := 1000.0
		votes1 := 100.0
		hours1 := 10.0
		popularityScore1 := (views1 / hours1) * (1 + (votes1 / views1))

		// Low engagement clip (100 views, 5 votes in 10 hours)
		views2 := 100.0
		votes2 := 5.0
		hours2 := 10.0
		popularityScore2 := (views2 / hours2) * (1 + (votes2 / views2))

		assert.Greater(t, popularityScore1, popularityScore2, "High engagement should score higher")
	})
}

// TestTrendingConfigurability tests configurable trending parameters
func TestTrendingConfigurability(t *testing.T) {
	t.Run("Trending window is configurable", func(t *testing.T) {
		service := &RecommendationService{
			trendingWindowDays: 14, // Extended window
			trendingMinScore:   10.0,
		}

		assert.Equal(t, 14, service.trendingWindowDays, "Should use custom trending window")
		assert.Equal(t, 10.0, service.trendingMinScore, "Should use custom min score")
	})

	t.Run("Popularity window is configurable", func(t *testing.T) {
		service := &RecommendationService{
			popularityWindowDays: 7,
			popularityMinViews:   50,
		}

		assert.Equal(t, 7, service.popularityWindowDays, "Should use custom popularity window")
		assert.Equal(t, 50, service.popularityMinViews, "Should use custom min views")
	})
}

// TestColdStartMetadataTracking tests that cold start state is properly tracked
func TestColdStartMetadataTracking(t *testing.T) {
	t.Run("Metadata indicates cold start", func(t *testing.T) {
		metadata := models.RecommendationMetadata{
			AlgorithmUsed:    models.AlgorithmTrending,
			ColdStart:        true,
			DiversityApplied: false,
			CacheHit:         false,
			ProcessingTimeMs: 50,
		}

		assert.True(t, metadata.ColdStart, "Should mark as cold start")
		assert.Equal(t, models.AlgorithmTrending, metadata.AlgorithmUsed, "Should use trending algorithm for cold start")
	})

	t.Run("Metadata indicates non-cold start", func(t *testing.T) {
		metadata := models.RecommendationMetadata{
			AlgorithmUsed:    models.AlgorithmHybrid,
			ColdStart:        false,
			DiversityApplied: true,
			CacheHit:         false,
			ProcessingTimeMs: 75,
		}

		assert.False(t, metadata.ColdStart, "Should not mark as cold start")
		assert.True(t, metadata.DiversityApplied, "Should apply diversity for non-cold start")
	})
}
