package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCalculatePrecisionAtK tests precision calculation
func TestCalculatePrecisionAtK(t *testing.T) {
	tests := []struct {
		name               string
		relevances         []int
		k                  int
		relevanceThreshold int
		want               float64
	}{
		{
			name:               "Perfect precision at k=5",
			relevances:         []int{4, 3, 3, 2, 2, 1, 0},
			k:                  5,
			relevanceThreshold: 2,
			want:               1.0, // All 5 are relevant
		},
		{
			name:               "Half precision at k=4",
			relevances:         []int{4, 0, 3, 0, 2},
			k:                  4,
			relevanceThreshold: 2,
			want:               0.5, // 2 out of 4
		},
		{
			name:               "Zero precision",
			relevances:         []int{0, 1, 1, 0, 0},
			k:                  5,
			relevanceThreshold: 2,
			want:               0.0,
		},
		{
			name:               "k larger than results",
			relevances:         []int{4, 3, 2},
			k:                  10,
			relevanceThreshold: 2,
			want:               0.3, // 3 out of 10
		},
		{
			name:               "Empty relevances",
			relevances:         []int{},
			k:                  5,
			relevanceThreshold: 2,
			want:               0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculatePrecisionAtK(tt.relevances, tt.k, tt.relevanceThreshold)
			assert.InDelta(t, tt.want, got, 0.01, "Precision should match expected value")
		})
	}
}

// TestCalculateRecallAtK tests recall calculation
func TestCalculateRecallAtK(t *testing.T) {
	tests := []struct {
		name               string
		relevances         []int
		k                  int
		totalRelevant      int
		relevanceThreshold int
		want               float64
	}{
		{
			name:               "Perfect recall",
			relevances:         []int{4, 3, 2, 0, 0},
			k:                  5,
			totalRelevant:      3,
			relevanceThreshold: 2,
			want:               1.0, // Found all 3 relevant items
		},
		{
			name:               "Partial recall",
			relevances:         []int{4, 0, 3, 0, 0},
			k:                  5,
			totalRelevant:      4,
			relevanceThreshold: 2,
			want:               0.5, // Found 2 out of 4 relevant
		},
		{
			name:               "Zero recall",
			relevances:         []int{0, 1, 1, 0, 0},
			k:                  5,
			totalRelevant:      3,
			relevanceThreshold: 2,
			want:               0.0,
		},
		{
			name:               "k too small to find all",
			relevances:         []int{4, 3, 0, 0, 2, 2},
			k:                  3,
			totalRelevant:      4,
			relevanceThreshold: 2,
			want:               0.5, // Found 2 out of 4 in top 3
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateRecallAtK(tt.relevances, tt.k, tt.totalRelevant, tt.relevanceThreshold)
			assert.InDelta(t, tt.want, got, 0.01, "Recall should match expected value")
		})
	}
}

// TestCalculateDiversity tests diversity calculation
func TestCalculateDiversity(t *testing.T) {
	tests := []struct {
		name    string
		gameIDs []string
		k       int
		want    float64
	}{
		{
			name:    "All different games",
			gameIDs: []string{"game1", "game2", "game3", "game4", "game5"},
			k:       5,
			want:    5.0,
		},
		{
			name:    "All same game",
			gameIDs: []string{"game1", "game1", "game1", "game1", "game1"},
			k:       5,
			want:    1.0,
		},
		{
			name:    "Two unique games",
			gameIDs: []string{"game1", "game2", "game1", "game2", "game1"},
			k:       5,
			want:    2.0,
		},
		{
			name:    "k smaller than list",
			gameIDs: []string{"game1", "game2", "game3", "game4", "game5"},
			k:       3,
			want:    3.0,
		},
		{
			name:    "Empty game IDs",
			gameIDs: []string{"", "", "", "", ""},
			k:       5,
			want:    0.0,
		},
		{
			name:    "Mixed empty and games",
			gameIDs: []string{"game1", "", "game2", "", "game1"},
			k:       5,
			want:    2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateDiversity(tt.gameIDs, tt.k)
			assert.Equal(t, tt.want, got, "Diversity should match expected value")
		})
	}
}

// TestCalculateSerendipity tests serendipity calculation
func TestCalculateSerendipity(t *testing.T) {
	tests := []struct {
		name               string
		relevances         []int
		gameIDs            []string
		favoriteGames      []string
		relevanceThreshold int
		want               float64
	}{
		{
			name:               "All serendipitous",
			relevances:         []int{4, 3, 2, 0, 0},
			gameIDs:            []string{"gameA", "gameB", "gameC", "gameD", "gameE"},
			favoriteGames:      []string{"gameX", "gameY"},
			relevanceThreshold: 2,
			want:               1.0, // 3 relevant, all from non-favorite games
		},
		{
			name:               "None serendipitous",
			relevances:         []int{4, 3, 2, 0, 0},
			gameIDs:            []string{"gameX", "gameY", "gameX", "gameZ", "gameW"},
			favoriteGames:      []string{"gameX", "gameY"},
			relevanceThreshold: 2,
			want:               0.0, // All relevant are from favorites
		},
		{
			name:               "Half serendipitous",
			relevances:         []int{4, 3, 2, 2, 0},
			gameIDs:            []string{"gameX", "gameA", "gameY", "gameB", "gameC"},
			favoriteGames:      []string{"gameX", "gameY"},
			relevanceThreshold: 2,
			want:               0.5, // 2 out of 4 relevant are from non-favorites
		},
		{
			name:               "No relevant items",
			relevances:         []int{0, 1, 1, 0, 0},
			gameIDs:            []string{"gameA", "gameB", "gameC", "gameD", "gameE"},
			favoriteGames:      []string{"gameX"},
			relevanceThreshold: 2,
			want:               0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateSerendipity(tt.relevances, tt.gameIDs, tt.favoriteGames, tt.relevanceThreshold)
			assert.InDelta(t, tt.want, got, 0.01, "Serendipity should match expected value")
		})
	}
}

// TestCalculateCoverage tests coverage calculation
func TestCalculateCoverage(t *testing.T) {
	tests := []struct {
		name           string
		recommendedIDs []string
		catalogSize    int
		want           float64
	}{
		{
			name:           "Full coverage",
			recommendedIDs: []string{"clip1", "clip2", "clip3", "clip4", "clip5"},
			catalogSize:    5,
			want:           1.0,
		},
		{
			name:           "Half coverage",
			recommendedIDs: []string{"clip1", "clip2", "clip3"},
			catalogSize:    6,
			want:           0.5,
		},
		{
			name:           "Duplicates in recommendations",
			recommendedIDs: []string{"clip1", "clip2", "clip1", "clip2", "clip3"},
			catalogSize:    10,
			want:           0.3, // 3 unique out of 10
		},
		{
			name:           "Empty recommendations",
			recommendedIDs: []string{},
			catalogSize:    10,
			want:           0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateCoverage(tt.recommendedIDs, tt.catalogSize)
			assert.InDelta(t, tt.want, got, 0.01, "Coverage should match expected value")
		})
	}
}

// TestCalculateAggregateMetrics tests aggregate metric calculation
func TestCalculateAggregateMetrics(t *testing.T) {
	service := &RecommendationEvaluationService{}

	results := []RecommendationEvaluationResult{
		{
			Precision5:       0.8,
			Precision10:      0.7,
			Recall5:          0.6,
			Recall10:         0.8,
			NDCG5:            0.9,
			NDCG10:           0.85,
			Diversity5:       4.0,
			Diversity10:      7.0,
			SerendipityScore: 0.5,
			Coverage:         0.3,
			IsColdStart:      false,
		},
		{
			Precision5:       0.6,
			Precision10:      0.5,
			Recall5:          0.4,
			Recall10:         0.6,
			NDCG5:            0.7,
			NDCG10:           0.65,
			Diversity5:       3.0,
			Diversity10:      5.0,
			SerendipityScore: 0.3,
			Coverage:         0.2,
			IsColdStart:      true,
		},
		{
			Precision5:       1.0,
			Precision10:      0.9,
			Recall5:          0.8,
			Recall10:         0.9,
			NDCG5:            0.95,
			NDCG10:           0.9,
			Diversity5:       5.0,
			Diversity10:      8.0,
			SerendipityScore: 0.7,
			Coverage:         0.4,
			IsColdStart:      true,
		},
	}

	aggregate := service.calculateAggregateMetrics(results)

	// Test mean values
	assert.InDelta(t, 0.8, aggregate.MeanPrecision5, 0.01, "Mean precision@5 should be average")
	assert.InDelta(t, 0.7, aggregate.MeanPrecision10, 0.01, "Mean precision@10 should be average")
	assert.InDelta(t, 0.6, aggregate.MeanRecall5, 0.01, "Mean recall@5 should be average")
	assert.InDelta(t, 0.77, aggregate.MeanRecall10, 0.01, "Mean recall@10 should be average")
	assert.InDelta(t, 0.85, aggregate.MeanNDCG5, 0.01, "Mean nDCG@5 should be average")
	assert.InDelta(t, 0.8, aggregate.MeanNDCG10, 0.01, "Mean nDCG@10 should be average")
	assert.InDelta(t, 4.0, aggregate.MeanDiversity5, 0.01, "Mean diversity@5 should be average")
	assert.InDelta(t, 6.67, aggregate.MeanDiversity10, 0.01, "Mean diversity@10 should be average")
	assert.InDelta(t, 0.5, aggregate.MeanSerendipity, 0.01, "Mean serendipity should be average")
	assert.InDelta(t, 0.3, aggregate.MeanCoverage, 0.01, "Mean coverage should be average")

	// Test cold start metrics
	assert.Equal(t, 3, aggregate.ScenarioCount, "Should count all scenarios")
	assert.Equal(t, 2, aggregate.ColdStartCount, "Should count cold start scenarios")
	assert.InDelta(t, 0.8, aggregate.ColdStartPrecision5, 0.01, "Cold start precision should be average of cold start scenarios")
	assert.InDelta(t, 0.6, aggregate.ColdStartRecall5, 0.01, "Cold start recall should be average of cold start scenarios")
}

// TestSimulateRecommendations tests simulated recommendation generation
func TestSimulateRecommendations(t *testing.T) {
	service := &RecommendationEvaluationService{}

	scenario := RecommendationScenario{
		ID:     "test-scenario",
		UserID: "user123",
		RelevantClips: []RelevantRecommendation{
			{ClipID: "clip1", Relevance: 2, GameID: "gameA"},
			{ClipID: "clip2", Relevance: 4, GameID: "gameB"},
			{ClipID: "clip3", Relevance: 1, GameID: "gameC"},
			{ClipID: "clip4", Relevance: 3, GameID: "gameD"},
		},
	}

	ids, gameIDs := service.SimulateRecommendations(scenario)

	// Should be sorted by relevance descending
	assert.Equal(t, []string{"clip2", "clip4", "clip1", "clip3"}, ids, "Should be sorted by relevance")
	assert.Equal(t, []string{"gameB", "gameD", "gameA", "gameC"}, gameIDs, "Game IDs should match sorted order")
}

// TestEvaluateRecommendationMetric tests metric evaluation against targets
func TestEvaluateRecommendationMetric(t *testing.T) {
	target := RecommendationTarget{
		Target:            0.8,
		WarningThreshold:  0.7,
		CriticalThreshold: 0.6,
	}

	tests := []struct {
		name  string
		value float64
		want  string
	}{
		{"Pass - above target", 0.85, "pass"},
		{"Pass - at target", 0.8, "pass"},
		{"Warning - above warning threshold", 0.75, "warning"},
		{"Warning - at warning threshold", 0.7, "warning"},
		{"Critical - below warning threshold", 0.65, "critical"},
		{"Critical - very low", 0.3, "critical"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := evaluateRecommendationMetric(tt.value, target)
			assert.Equal(t, tt.want, got, "Status should match expected")
		})
	}
}

// TestEvaluateScenario tests full scenario evaluation
func TestEvaluateScenario(t *testing.T) {
	service := &RecommendationEvaluationService{}

	scenario := RecommendationScenario{
		ID:          "test-001",
		UserID:      "user123",
		Description: "Test scenario",
		Algorithm:   "hybrid",
		IsColdStart: false,
		UserProfile: UserProfileData{
			FavoriteGames: []string{"gameA", "gameB"},
		},
		RelevantClips: []RelevantRecommendation{
			{ClipID: "clip1", Relevance: 4, GameID: "gameA"},
			{ClipID: "clip2", Relevance: 3, GameID: "gameB"},
			{ClipID: "clip3", Relevance: 2, GameID: "gameC"},
			{ClipID: "clip4", Relevance: 1, GameID: "gameD"},
			{ClipID: "clip5", Relevance: 0, GameID: "gameE"},
		},
	}

	// Simulated recommendations (sorted by relevance)
	recommendedIDs := []string{"clip1", "clip2", "clip3", "clip4", "clip5"}
	recommendedGameIDs := []string{"gameA", "gameB", "gameC", "gameD", "gameE"}

	result := service.EvaluateScenario(context.Background(), scenario, recommendedIDs, recommendedGameIDs)

	// Verify basic fields
	assert.Equal(t, "test-001", result.ScenarioID)
	assert.Equal(t, "user123", result.UserID)
	assert.Equal(t, "hybrid", result.Algorithm)
	assert.False(t, result.IsColdStart)

	// Verify metrics
	// 3 out of 5 clips are relevant (relevance >= 2)
	assert.Equal(t, 0.6, result.Precision5, "3 out of 5 should be relevant (>=2)")
	// Precision@10 = relevant in top 10 / 10 = 3/10 = 0.3 (even though only 5 results exist)
	assert.InDelta(t, 0.3, result.Precision10, 0.01, "3 relevant out of 10 requested")
	assert.Equal(t, 1.0, result.Recall5, "All 3 relevant items found in top 5")
	assert.Equal(t, 5.0, result.Diversity5, "All 5 clips have different games")

	// Serendipity: clip3 is relevant and not from favorite games
	// 1 serendipitous out of 3 relevant = 0.333...
	assert.InDelta(t, 0.33, result.SerendipityScore, 0.01, "One relevant clip from non-favorite game")

	// Coverage: 5 unique clips out of 5 in catalog
	assert.Equal(t, 1.0, result.Coverage, "All clips in catalog recommended")

	// Verify counts
	assert.Equal(t, 5, result.RetrievedCount)
	assert.Equal(t, 3, result.RelevantCount)
}

// TestCalculateNDCGInRecommendationContext tests nDCG calculation for recommendation evaluation
// This verifies the metric works correctly in recommendation scenarios with relevance scores
func TestCalculateNDCGInRecommendationContext(t *testing.T) {
	tests := []struct {
		name       string
		relevances []int
		judged     []int
		k          int
		want       float64
	}{
		{
			name:       "Perfect recommendation ranking",
			relevances: []int{4, 3, 2, 1, 0},
			k:          5,
			want:       1.0, // Already in ideal order
		},
		{
			name:       "Poor recommendation ranking",
			relevances: []int{0, 1, 2, 3, 4},
			k:          5,
			want:       0.51, // Reversed order
		},
		{
			name:       "Top 3 recommendations ideal",
			relevances: []int{4, 3, 2, 1, 0},
			k:          3,
			want:       1.0, // Top 3 are in ideal order
		},
		{
			name:       "Mixed recommendation quality",
			relevances: []int{3, 4, 2, 1, 0},
			k:          5,
			want:       0.86, // Close to ideal but not perfect (4 and 3 are swapped)
		},
		{
			name:       "All same relevance recommendations",
			relevances: []int{2, 2, 2, 2, 2},
			k:          5,
			want:       1.0, // Any order is ideal when all have same relevance
		},
		{
			name:       "All irrelevant recommendations",
			relevances: []int{0, 0, 0, 0, 0},
			k:          5,
			want:       0.0, // No relevant items
		},
		{
			name:       "Fewer recommendations than k",
			relevances: []int{4, 3, 2},
			k:          10,
			want:       1.0, // All results in ideal order
		},
		{
			name:       "Empty recommendations",
			relevances: []int{},
			k:          5,
			want:       0.0,
		},
		{
			name:       "Highly relevant items first",
			relevances: []int{4, 4, 3, 2, 2, 1, 0},
			k:          5,
			want:       1.0, // Top items are in descending relevance order
		},
		{
			name:       "Relevant but suboptimal order",
			relevances: []int{2, 4, 3, 2, 1},
			k:          5,
			want:       0.78, // Relevant items present but not in ideal order
		},
		{
			name:       "Ideal order but best judged clips missing",
			relevances: []int{2, 1, 0},
			judged:     []int{4, 4, 2, 1},
			k:          5,
			want:       0.14, // 3.63 / 26.39: missing the 4s must not score 1.0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Unless a case sets judged, the ground truth is exactly the
			// recommended set.
			judged := tt.judged
			if judged == nil {
				judged = tt.relevances
			}
			got := CalculateNDCG(tt.relevances, judged, tt.k)
			assert.InDelta(t, tt.want, got, 0.05, "nDCG should match expected value within tolerance")
		})
	}
}
