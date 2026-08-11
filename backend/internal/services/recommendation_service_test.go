package services

import (
	"testing"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnforceCreatorDiversity tests that creator repetition is capped first.
func TestEnforceCreatorDiversity(t *testing.T) {
	service := &RecommendationService{
		contentWeight:       0.5,
		collaborativeWeight: 0.3,
		trendingWeight:      0.2,
	}

	gameIDs := []string{"category-1", "category-2", "category-3"}
	creatorIDs := []string{"creator-1", "creator-2", "creator-3", "creator-4", "creator-5"}
	var recommendations []models.ClipRecommendation
	for index := 0; index < 3; index++ {
		for creatorIndex := range creatorIDs {
			creatorID := creatorIDs[creatorIndex]
			gameID := gameIDs[(creatorIndex+index)%len(gameIDs)]
			recommendations = append(recommendations, models.ClipRecommendation{
				Clip:  models.Clip{ID: uuid.New(), BroadcasterID: &creatorID, GameID: &gameID},
				Score: 1 - float64(len(recommendations))/100,
			})
		}
	}

	// Enforce diversity with limit of 10
	diversified := service.enforceCreatorDiversity(recommendations, 10)

	// Should have exactly 10 recommendations
	assert.Len(t, diversified, 10, "Should have exactly 10 recommendations")

	creatorCounts := make(map[string]int)
	for _, rec := range diversified {
		creatorCounts[*rec.Clip.BroadcasterID]++
	}
	for creatorID, count := range creatorCounts {
		assert.LessOrEqual(t, count, 2, "should cap clips from %s", creatorID)
	}
}

// TestMergeAndRank tests the merging and ranking of scores
func TestMergeAndRank(t *testing.T) {
	service := &RecommendationService{
		contentWeight:       0.5,
		collaborativeWeight: 0.3,
		trendingWeight:      0.2,
	}

	clipID1 := uuid.New()
	clipID2 := uuid.New()
	clipID3 := uuid.New()

	// Create test scores
	contentScores := []models.ClipScore{
		{ClipID: clipID1, SimilarityScore: 0.8},
		{ClipID: clipID2, SimilarityScore: 0.6},
	}

	collaborativeScores := []models.ClipScore{
		{ClipID: clipID2, SimilarityScore: 0.9},
		{ClipID: clipID3, SimilarityScore: 0.7},
	}

	trendingScores := []models.ClipScore{
		{ClipID: clipID1, SimilarityScore: 0.5},
		{ClipID: clipID3, SimilarityScore: 0.8},
	}

	// Merge and rank
	merged := service.mergeAndRank(contentScores, collaborativeScores, trendingScores)

	// Should have 3 unique clips
	require.Len(t, merged, 3, "Should have 3 unique clips")

	// Check scores are weighted correctly
	// clipID1: 0.8*0.5 + 0*0.3 + 0.5*0.2 = 0.4 + 0 + 0.1 = 0.5
	// clipID2: 0.6*0.5 + 0.9*0.3 + 0*0.2 = 0.3 + 0.27 + 0 = 0.57
	// clipID3: 0*0.5 + 0.7*0.3 + 0.8*0.2 = 0 + 0.21 + 0.16 = 0.37

	// Find each clip in merged results
	var clip1Score, clip2Score, clip3Score float64
	for _, score := range merged {
		switch score.ClipID {
		case clipID1:
			clip1Score = score.SimilarityScore
		case clipID2:
			clip2Score = score.SimilarityScore
		case clipID3:
			clip3Score = score.SimilarityScore
		}
	}

	// Check weighted scores
	assert.InDelta(t, 0.5, clip1Score, 0.01, "clipID1 score should be ~0.5")
	assert.InDelta(t, 0.57, clip2Score, 0.01, "clipID2 score should be ~0.57")
	assert.InDelta(t, 0.37, clip3Score, 0.01, "clipID3 score should be ~0.37")

	// Check ranking - clipID2 should be first (highest score)
	assert.Equal(t, clipID2, merged[0].ClipID, "clipID2 should be ranked first")
	assert.Equal(t, 1, merged[0].SimilarityRank, "First clip should have rank 1")

	// Verify descending order
	for i := 1; i < len(merged); i++ {
		assert.GreaterOrEqual(t, merged[i-1].SimilarityScore, merged[i].SimilarityScore,
			"Scores should be in descending order")
		assert.Equal(t, i+1, merged[i].SimilarityRank, "Ranks should be sequential")
	}
}

// TestGenerateReason tests reason generation
func TestGenerateReason(t *testing.T) {
	service := &RecommendationService{}

	gameName := "Valorant"
	broadcasterName := "shroud"

	tests := []struct {
		name      string
		clip      *models.Clip
		algorithm string
		wantMatch string
	}{
		{
			name: "Clip with game name",
			clip: &models.Clip{
				ID:              uuid.New(),
				GameName:        &gameName,
				BroadcasterName: "",
			},
			algorithm: models.AlgorithmContent,
			wantMatch: "More from the Valorant Twitch category",
		},
		{
			name: "Clip with broadcaster name",
			clip: &models.Clip{
				ID:              uuid.New(),
				GameName:        nil,
				BroadcasterName: broadcasterName,
			},
			algorithm: models.AlgorithmContent,
			wantMatch: "Because you watched shroud",
		},
		{
			name: "Trending algorithm",
			clip: &models.Clip{
				ID:              uuid.New(),
				GameName:        nil,
				BroadcasterName: "",
			},
			algorithm: models.AlgorithmTrending,
			wantMatch: "Trending now",
		},
		{
			name: "Collaborative algorithm",
			clip: &models.Clip{
				ID:              uuid.New(),
				GameName:        nil,
				BroadcasterName: "",
			},
			algorithm: models.AlgorithmCollaborative,
			wantMatch: "Popular with users like you",
		},
		{
			name: "Hybrid algorithm",
			clip: &models.Clip{
				ID:              uuid.New(),
				GameName:        nil,
				BroadcasterName: "",
			},
			algorithm: models.AlgorithmHybrid,
			wantMatch: "Recommended for you",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reason := service.generateReason(tt.clip, tt.algorithm, 0.8)
			assert.NotEmpty(t, reason, "Reason should not be empty")
			if tt.wantMatch != "" {
				assert.Equal(t, tt.wantMatch, reason, "Reason should match expected")
			}
		})
	}
}

func TestEnforceCreatorDiversityEmptyList(t *testing.T) {
	service := &RecommendationService{}
	diversified := service.enforceCreatorDiversity([]models.ClipRecommendation{}, 10)
	assert.Empty(t, diversified, "Should return empty list for empty input")
}

func TestEnforceCreatorDiversityLessThanLimit(t *testing.T) {
	service := &RecommendationService{}

	gameID := "game-1"
	recommendations := []models.ClipRecommendation{
		{Clip: models.Clip{ID: uuid.New(), GameID: &gameID}, Score: 0.9},
		{Clip: models.Clip{ID: uuid.New(), GameID: &gameID}, Score: 0.8},
	}

	diversified := service.enforceCreatorDiversity(recommendations, 10)
	assert.Len(t, diversified, 2, "Should return all recommendations when less than limit")
}

// TestMergeAndRankNoOverlap tests merging with no overlapping clips
func TestMergeAndRankNoOverlap(t *testing.T) {
	service := &RecommendationService{
		contentWeight:       0.5,
		collaborativeWeight: 0.3,
		trendingWeight:      0.2,
	}

	clipID1 := uuid.New()
	clipID2 := uuid.New()
	clipID3 := uuid.New()

	contentScores := []models.ClipScore{
		{ClipID: clipID1, SimilarityScore: 0.8},
	}

	collaborativeScores := []models.ClipScore{
		{ClipID: clipID2, SimilarityScore: 0.9},
	}

	trendingScores := []models.ClipScore{
		{ClipID: clipID3, SimilarityScore: 0.7},
	}

	merged := service.mergeAndRank(contentScores, collaborativeScores, trendingScores)

	// Should have 3 clips with no overlap
	assert.Len(t, merged, 3, "Should have 3 unique clips")

	// Each should only have score from one algorithm
	for _, score := range merged {
		if score.ClipID == clipID1 {
			assert.InDelta(t, 0.8*0.5, score.SimilarityScore, 0.01, "Should only have content score")
		} else if score.ClipID == clipID2 {
			assert.InDelta(t, 0.9*0.3, score.SimilarityScore, 0.01, "Should only have collaborative score")
		} else if score.ClipID == clipID3 {
			assert.InDelta(t, 0.7*0.2, score.SimilarityScore, 0.01, "Should only have trending score")
		}
	}
}

// TestMergeAndRankEmptyScores tests merging with empty score lists
func TestMergeAndRankEmptyScores(t *testing.T) {
	service := &RecommendationService{
		contentWeight:       0.5,
		collaborativeWeight: 0.3,
		trendingWeight:      0.2,
	}

	merged := service.mergeAndRank(
		[]models.ClipScore{},
		[]models.ClipScore{},
		[]models.ClipScore{},
	)

	assert.Empty(t, merged, "Should return empty list for empty scores")
}
