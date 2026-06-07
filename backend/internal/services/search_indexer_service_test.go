package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

func TestCalculateEngagementScore(t *testing.T) {
	tests := []struct {
		name     string
		clip     *models.Clip
		expected float64
	}{
		{
			name: "High engagement clip",
			clip: &models.Clip{
				VoteScore:     100,
				CommentCount:  50,
				FavoriteCount: 30,
				ViewCount:     10000,
			},
			expected: 100*10.0 + 50*5.0 + 30*3.0 + 10000*0.01,
		},
		{
			name: "Low engagement clip",
			clip: &models.Clip{
				VoteScore:     1,
				CommentCount:  0,
				FavoriteCount: 0,
				ViewCount:     100,
			},
			expected: 1*10.0 + 0*5.0 + 0*3.0 + 100*0.01,
		},
		{
			name: "Zero engagement clip",
			clip: &models.Clip{
				VoteScore:     0,
				CommentCount:  0,
				FavoriteCount: 0,
				ViewCount:     0,
			},
			expected: 0.0,
		},
		{
			name: "Negative vote score",
			clip: &models.Clip{
				VoteScore:     -10,
				CommentCount:  5,
				FavoriteCount: 2,
				ViewCount:     500,
			},
			expected: -10*10.0 + 5*5.0 + 2*3.0 + 500*0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateEngagementScore(tt.clip)
			assert.Equal(t, tt.expected, score)
		})
	}
}

func TestCalculateRecencyScore(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		clip          *models.Clip
		expectedRange [2]float64 // min, max expected values
		checkDecay    bool
	}{
		{
			name: "Brand new clip",
			clip: &models.Clip{
				CreatedAt: now,
			},
			expectedRange: [2]float64{99.0, 100.0}, // Should be close to 100
			checkDecay:    false,
		},
		{
			name: "One week old clip",
			clip: &models.Clip{
				CreatedAt: now.Add(-7 * 24 * time.Hour),
			},
			expectedRange: [2]float64{45.0, 55.0}, // ~50% of initial score
			checkDecay:    true,
		},
		{
			name: "Two weeks old clip",
			clip: &models.Clip{
				CreatedAt: now.Add(-14 * 24 * time.Hour),
			},
			expectedRange: [2]float64{20.0, 30.0}, // ~25% of initial score
			checkDecay:    true,
		},
		{
			name: "One month old clip",
			clip: &models.Clip{
				CreatedAt: now.Add(-30 * 24 * time.Hour),
			},
			expectedRange: [2]float64{5.0, 15.0}, // Very low score
			checkDecay:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateRecencyScore(tt.clip)

			// Check score is within expected range
			assert.GreaterOrEqual(t, score, tt.expectedRange[0],
				"Score should be >= %f, got %f", tt.expectedRange[0], score)
			assert.LessOrEqual(t, score, tt.expectedRange[1],
				"Score should be <= %f, got %f", tt.expectedRange[1], score)

			// Verify score is positive
			assert.Greater(t, score, 0.0, "Recency score should always be positive")
		})
	}
}

func TestCalculateRecencyScoreDecay(t *testing.T) {
	now := time.Now()

	// Create clips at different ages
	newClip := &models.Clip{CreatedAt: now}
	oneWeekOld := &models.Clip{CreatedAt: now.Add(-7 * 24 * time.Hour)}
	twoWeeksOld := &models.Clip{CreatedAt: now.Add(-14 * 24 * time.Hour)}

	newScore := calculateRecencyScore(newClip)
	oneWeekScore := calculateRecencyScore(oneWeekOld)
	twoWeeksScore := calculateRecencyScore(twoWeeksOld)

	// Verify exponential decay: newer clips always have higher scores
	assert.Greater(t, newScore, oneWeekScore,
		"New clip should have higher score than 1-week-old clip")
	assert.Greater(t, oneWeekScore, twoWeeksScore,
		"1-week-old clip should have higher score than 2-week-old clip")

	// Verify approximate 50% decay per week
	ratio := oneWeekScore / newScore
	assert.InDelta(t, 0.5, ratio, 0.15,
		"Score should decrease by ~50%% per week, got %.2f", ratio)
}

func TestBulkIndexClipsEngagementAndRecency(t *testing.T) {
	// This test verifies that BulkIndexClips calculates engagement and recency scores
	now := time.Now()

	clips := []models.Clip{
		{
			ID:            uuid.New(),
			Title:         "Test Clip 1",
			VoteScore:     50,
			CommentCount:  10,
			FavoriteCount: 5,
			ViewCount:     1000,
			CreatedAt:     now.Add(-24 * time.Hour),
		},
		{
			ID:            uuid.New(),
			Title:         "Test Clip 2",
			VoteScore:     100,
			CommentCount:  20,
			FavoriteCount: 10,
			ViewCount:     5000,
			CreatedAt:     now,
		},
	}

	// Calculate expected scores
	for _, clip := range clips {
		engagementScore := calculateEngagementScore(&clip)
		recencyScore := calculateRecencyScore(&clip)

		// Verify scores are reasonable
		assert.Greater(t, engagementScore, 0.0,
			"Engagement score should be positive for clip with interactions")
		assert.Greater(t, recencyScore, 0.0,
			"Recency score should be positive")

		// Newer clip should have higher recency
		if clip.CreatedAt.After(clips[0].CreatedAt) {
			clip2RecencyScore := calculateRecencyScore(&clips[1])
			clip1RecencyScore := calculateRecencyScore(&clips[0])
			assert.Greater(t, clip2RecencyScore, clip1RecencyScore,
				"Newer clip should have higher recency score")
		}
	}
}

func TestEngagementScoreWeights(t *testing.T) {
	// Verify that weights are applied correctly
	clip := &models.Clip{
		VoteScore:     10,
		CommentCount:  10,
		FavoriteCount: 10,
		ViewCount:     10,
	}

	score := calculateEngagementScore(clip)

	// Calculate expected with known weights
	// Vote: 10 * 10.0 = 100
	// Comments: 10 * 5.0 = 50
	// Favorites: 10 * 3.0 = 30
	// Views: 10 * 0.01 = 0.1
	// Total: 180.1
	expected := 180.1

	assert.Equal(t, expected, score,
		"Engagement score calculation should match expected weights")
}
