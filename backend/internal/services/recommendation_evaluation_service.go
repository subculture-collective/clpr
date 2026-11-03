package services

import (
	"context"
	"fmt"
	"math"
	"os"
	"sort"

	"gopkg.in/yaml.v3"
)

const (
	// RelevanceThreshold defines the minimum relevance score to consider an item as "relevant"
	// Using a 0-4 relevance scale, items with score >= 2 are considered relevant
	RelevanceThreshold = 2
)

// RecommendationEvaluationService evaluates recommendation quality using standard metrics
type RecommendationEvaluationService struct {
	recommendationService *RecommendationService
	dataset               *RecommendationEvaluationDataset
}

// RecommendationEvaluationDataset represents the loaded evaluation dataset
type RecommendationEvaluationDataset struct {
	Version              string                          `yaml:"version"`
	Description          string                          `yaml:"description"`
	EvaluationScenarios  []RecommendationScenario        `yaml:"evaluation_scenarios"`
	MetricTargets        map[string]RecommendationTarget `yaml:"metric_targets"`
	EvaluationGuidelines []string                        `yaml:"evaluation_guidelines"`
}

// RecommendationScenario represents a single evaluation scenario
type RecommendationScenario struct {
	ID            string                   `yaml:"id"`
	UserID        string                   `yaml:"user_id"`
	Description   string                   `yaml:"description"`
	UserProfile   UserProfileData          `yaml:"user_profile"`
	RelevantClips []RelevantRecommendation `yaml:"relevant_clips"`
	Algorithm     string                   `yaml:"algorithm,omitempty"`
	IsColdStart   bool                     `yaml:"is_cold_start"`
}

// UserProfileData represents user preferences and history
type UserProfileData struct {
	FavoriteGames       []string `yaml:"favorite_games,omitempty"`
	FollowedStreamers   []string `yaml:"followed_streamers,omitempty"`
	PreferredCategories []string `yaml:"preferred_categories,omitempty"`
	InteractionHistory  []string `yaml:"interaction_history,omitempty"`
}

// RelevantRecommendation represents a clip with its relevance score
type RelevantRecommendation struct {
	ClipID    string `yaml:"clip_id"`
	Relevance int    `yaml:"relevance"` // 0-4 scale
	Reason    string `yaml:"reason"`
	GameID    string `yaml:"game_id,omitempty"`
}

// RecommendationTarget defines the target and threshold for a metric
type RecommendationTarget struct {
	Target            float64 `yaml:"target"`
	Description       string  `yaml:"description"`
	WarningThreshold  float64 `yaml:"warning_threshold"`
	CriticalThreshold float64 `yaml:"critical_threshold"`
}

// RecommendationEvaluationResult contains results for a single scenario
type RecommendationEvaluationResult struct {
	ScenarioID         string                  `json:"scenario_id"`
	UserID             string                  `json:"user_id"`
	Description        string                  `json:"description"`
	Algorithm          string                  `json:"algorithm"`
	IsColdStart        bool                    `json:"is_cold_start"`
	Precision5         float64                 `json:"precision_at_5"`
	Precision10        float64                 `json:"precision_at_10"`
	Recall5            float64                 `json:"recall_at_5"`
	Recall10           float64                 `json:"recall_at_10"`
	NDCG5              float64                 `json:"ndcg_at_5"`
	NDCG10             float64                 `json:"ndcg_at_10"`
	Diversity5         float64                 `json:"diversity_at_5"`
	Diversity10        float64                 `json:"diversity_at_10"`
	SerendipityScore   float64                 `json:"serendipity_score"`
	Coverage           float64                 `json:"coverage"`
	RetrievedCount     int                     `json:"retrieved_count"`
	RelevantCount      int                     `json:"relevant_count"`
	RecommendedResults []RecommendedResultItem `json:"recommended_results,omitempty"`
}

// RecommendedResultItem represents a single recommendation with its relevance
type RecommendedResultItem struct {
	ClipID    string  `json:"clip_id"`
	Rank      int     `json:"rank"`
	Relevance int     `json:"relevance"`
	GameID    string  `json:"game_id,omitempty"`
	Score     float64 `json:"score,omitempty"`
}

// RecommendationAggregateMetrics contains aggregated metrics across all scenarios
type RecommendationAggregateMetrics struct {
	MeanPrecision5      float64 `json:"mean_precision_at_5"`
	MeanPrecision10     float64 `json:"mean_precision_at_10"`
	MeanRecall5         float64 `json:"mean_recall_at_5"`
	MeanRecall10        float64 `json:"mean_recall_at_10"`
	MeanNDCG5           float64 `json:"mean_ndcg_at_5"`
	MeanNDCG10          float64 `json:"mean_ndcg_at_10"`
	MeanDiversity5      float64 `json:"mean_diversity_at_5"`
	MeanDiversity10     float64 `json:"mean_diversity_at_10"`
	MeanSerendipity     float64 `json:"mean_serendipity"`
	MeanCoverage        float64 `json:"mean_coverage"`
	ScenarioCount       int     `json:"scenario_count"`
	ColdStartPrecision5 float64 `json:"cold_start_precision_at_5"`
	ColdStartRecall5    float64 `json:"cold_start_recall_at_5"`
	ColdStartCount      int     `json:"cold_start_count"`
}

// RecommendationEvaluationReport contains the full evaluation report
type RecommendationEvaluationReport struct {
	Metrics         RecommendationAggregateMetrics   `json:"aggregate_metrics"`
	ScenarioResults []RecommendationEvaluationResult `json:"scenario_results"`
	Targets         map[string]RecommendationTarget  `json:"targets,omitempty"`
	Status          map[string]string                `json:"status,omitempty"` // "pass", "warning", "critical"
}

// NewRecommendationEvaluationService creates a new recommendation evaluation service
func NewRecommendationEvaluationService(recommendationService *RecommendationService) *RecommendationEvaluationService {
	return &RecommendationEvaluationService{
		recommendationService: recommendationService,
	}
}

// LoadDataset loads the evaluation dataset from a YAML file
func (s *RecommendationEvaluationService) LoadDataset(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read dataset file: %w", err)
	}

	var dataset RecommendationEvaluationDataset
	if err := yaml.Unmarshal(data, &dataset); err != nil {
		return fmt.Errorf("failed to parse dataset YAML: %w", err)
	}

	s.dataset = &dataset
	return nil
}

// GetDataset returns the loaded dataset
func (s *RecommendationEvaluationService) GetDataset() *RecommendationEvaluationDataset {
	return s.dataset
}

// CalculatePrecisionAtK calculates precision@k for recommendations
func CalculatePrecisionAtK(relevances []int, k int, relevanceThreshold int) float64 {
	if k <= 0 {
		return 0.0
	}

	limit := k
	if limit > len(relevances) {
		limit = len(relevances)
	}

	relevant := 0
	for i := 0; i < limit; i++ {
		if relevances[i] >= relevanceThreshold {
			relevant++
		}
	}

	return float64(relevant) / float64(k)
}

// CalculateRecallAtK calculates recall@k for recommendations
func CalculateRecallAtK(relevances []int, k int, totalRelevant int, relevanceThreshold int) float64 {
	if totalRelevant <= 0 || k <= 0 {
		return 0.0
	}

	limit := k
	if limit > len(relevances) {
		limit = len(relevances)
	}

	foundRelevant := 0
	for i := 0; i < limit; i++ {
		if relevances[i] >= relevanceThreshold {
			foundRelevant++
		}
	}

	return float64(foundRelevant) / float64(totalRelevant)
}

// CalculateDiversity calculates diversity (unique games) in top-k recommendations
func CalculateDiversity(gameIDs []string, k int) float64 {
	if k <= 0 {
		return 0.0
	}

	limit := k
	if limit > len(gameIDs) {
		limit = len(gameIDs)
	}

	uniqueGames := make(map[string]bool)
	for i := 0; i < limit; i++ {
		if gameIDs[i] != "" {
			uniqueGames[gameIDs[i]] = true
		}
	}

	return float64(len(uniqueGames))
}

// CalculateSerendipity calculates serendipity score (unexpected relevant items)
// Serendipity = relevant items that are not from user's favorite games / total relevant items
func CalculateSerendipity(relevances []int, gameIDs []string, favoriteGames []string, relevanceThreshold int) float64 {
	favoriteSet := make(map[string]bool)
	for _, game := range favoriteGames {
		favoriteSet[game] = true
	}

	relevantCount := 0
	serendipitousCount := 0

	for i := 0; i < len(relevances); i++ {
		if relevances[i] >= relevanceThreshold {
			relevantCount++
			// Serendipitous if relevant but not from favorite games
			if i < len(gameIDs) && gameIDs[i] != "" && !favoriteSet[gameIDs[i]] {
				serendipitousCount++
			}
		}
	}

	if relevantCount == 0 {
		return 0.0
	}

	return float64(serendipitousCount) / float64(relevantCount)
}

// CalculateCoverage calculates catalog coverage (fraction of unique items recommended)
func CalculateCoverage(recommendedIDs []string, catalogSize int) float64 {
	if catalogSize <= 0 {
		return 0.0
	}

	unique := make(map[string]bool)
	for _, id := range recommendedIDs {
		unique[id] = true
	}

	return float64(len(unique)) / float64(catalogSize)
}

// EvaluateScenario runs evaluation for a single scenario
func (s *RecommendationEvaluationService) EvaluateScenario(
	ctx context.Context,
	scenario RecommendationScenario,
	recommendedIDs []string,
	recommendedGameIDs []string,
) RecommendationEvaluationResult {
	// Build relevance map from ground truth
	relevanceMap := make(map[string]int)
	gameIDMap := make(map[string]string)
	totalRelevant := 0

	for _, clip := range scenario.RelevantClips {
		relevanceMap[clip.ClipID] = clip.Relevance
		gameIDMap[clip.ClipID] = clip.GameID
		if clip.Relevance >= RelevanceThreshold { // Consider relevance >= 2 as "relevant"
			totalRelevant++
		}
	}

	// Ground-truth relevances (one per judged clip) for the nDCG ideal ranking
	judgedRelevances := make([]int, 0, len(relevanceMap))
	for _, rel := range relevanceMap {
		judgedRelevances = append(judgedRelevances, rel)
	}

	// Get relevances for recommended items
	relevances := make([]int, len(recommendedIDs))
	gameIDs := make([]string, len(recommendedIDs))
	recommendedResults := make([]RecommendedResultItem, len(recommendedIDs))
	relevantRetrieved := 0

	for i, id := range recommendedIDs {
		rel := 0
		if r, ok := relevanceMap[id]; ok {
			rel = r
		}
		relevances[i] = rel

		gameID := ""
		if i < len(recommendedGameIDs) {
			gameID = recommendedGameIDs[i]
		} else if gid, ok := gameIDMap[id]; ok {
			gameID = gid
		}
		gameIDs[i] = gameID

		recommendedResults[i] = RecommendedResultItem{
			ClipID:    id,
			Rank:      i + 1,
			Relevance: rel,
			GameID:    gameID,
		}

		if rel >= RelevanceThreshold {
			relevantRetrieved++
		}
	}

	// Calculate all metrics
	result := RecommendationEvaluationResult{
		ScenarioID:  scenario.ID,
		UserID:      scenario.UserID,
		Description: scenario.Description,
		Algorithm:   scenario.Algorithm,
		IsColdStart: scenario.IsColdStart,
		Precision5:  CalculatePrecisionAtK(relevances, 5, RelevanceThreshold),
		Precision10: CalculatePrecisionAtK(relevances, 10, RelevanceThreshold),
		Recall5:     CalculateRecallAtK(relevances, 5, totalRelevant, RelevanceThreshold),
		Recall10:    CalculateRecallAtK(relevances, 10, totalRelevant, RelevanceThreshold),
		// Note: CalculateNDCG is implemented in search_evaluation_service.go (same package).
		NDCG5:              CalculateNDCG(relevances, judgedRelevances, 5),
		NDCG10:             CalculateNDCG(relevances, judgedRelevances, 10),
		Diversity5:         CalculateDiversity(gameIDs, 5),
		Diversity10:        CalculateDiversity(gameIDs, 10),
		SerendipityScore:   CalculateSerendipity(relevances, gameIDs, scenario.UserProfile.FavoriteGames, RelevanceThreshold),
		Coverage:           CalculateCoverage(recommendedIDs, len(scenario.RelevantClips)),
		RetrievedCount:     len(recommendedIDs),
		RelevantCount:      relevantRetrieved,
		RecommendedResults: recommendedResults,
	}

	return result
}

// EvaluateDataset runs evaluation on the entire dataset
func (s *RecommendationEvaluationService) EvaluateDataset(
	ctx context.Context,
	resultsProvider func(scenario RecommendationScenario) ([]string, []string, error),
) (*RecommendationEvaluationReport, error) {
	if s.dataset == nil {
		return nil, fmt.Errorf("no dataset loaded")
	}

	results := make([]RecommendationEvaluationResult, 0, len(s.dataset.EvaluationScenarios))

	for _, scenario := range s.dataset.EvaluationScenarios {
		// Get recommendations
		recommendedIDs, gameIDs, err := resultsProvider(scenario)
		if err != nil {
			// Log error but continue with empty results
			recommendedIDs = []string{}
			gameIDs = []string{}
		}

		result := s.EvaluateScenario(ctx, scenario, recommendedIDs, gameIDs)
		results = append(results, result)
	}

	// Calculate aggregate metrics
	aggregate := s.calculateAggregateMetrics(results)

	// Check against targets
	status := s.checkTargets(aggregate)

	return &RecommendationEvaluationReport{
		Metrics:         aggregate,
		ScenarioResults: results,
		Targets:         s.dataset.MetricTargets,
		Status:          status,
	}, nil
}

// calculateAggregateMetrics computes mean metrics across all scenarios
func (s *RecommendationEvaluationService) calculateAggregateMetrics(
	results []RecommendationEvaluationResult,
) RecommendationAggregateMetrics {
	if len(results) == 0 {
		return RecommendationAggregateMetrics{}
	}

	var sumPrec5, sumPrec10, sumRecall5, sumRecall10 float64
	var sumNDCG5, sumNDCG10, sumDiv5, sumDiv10 float64
	var sumSerendipity, sumCoverage float64
	var coldStartPrec5, coldStartRecall5 float64
	coldStartCount := 0

	for _, r := range results {
		sumPrec5 += r.Precision5
		sumPrec10 += r.Precision10
		sumRecall5 += r.Recall5
		sumRecall10 += r.Recall10
		sumNDCG5 += r.NDCG5
		sumNDCG10 += r.NDCG10
		sumDiv5 += r.Diversity5
		sumDiv10 += r.Diversity10
		sumSerendipity += r.SerendipityScore
		sumCoverage += r.Coverage

		if r.IsColdStart {
			coldStartPrec5 += r.Precision5
			coldStartRecall5 += r.Recall5
			coldStartCount++
		}
	}

	n := float64(len(results))
	metrics := RecommendationAggregateMetrics{
		MeanPrecision5:  sumPrec5 / n,
		MeanPrecision10: sumPrec10 / n,
		MeanRecall5:     sumRecall5 / n,
		MeanRecall10:    sumRecall10 / n,
		MeanNDCG5:       sumNDCG5 / n,
		MeanNDCG10:      sumNDCG10 / n,
		MeanDiversity5:  sumDiv5 / n,
		MeanDiversity10: sumDiv10 / n,
		MeanSerendipity: sumSerendipity / n,
		MeanCoverage:    sumCoverage / n,
		ScenarioCount:   len(results),
		ColdStartCount:  coldStartCount,
	}

	if coldStartCount > 0 {
		metrics.ColdStartPrecision5 = coldStartPrec5 / float64(coldStartCount)
		metrics.ColdStartRecall5 = coldStartRecall5 / float64(coldStartCount)
	}

	return metrics
}

// checkTargets compares metrics against targets and returns status
func (s *RecommendationEvaluationService) checkTargets(
	metrics RecommendationAggregateMetrics,
) map[string]string {
	if s.dataset == nil || s.dataset.MetricTargets == nil {
		return nil
	}

	status := make(map[string]string)

	metricValues := map[string]float64{
		"precision_at_5":         metrics.MeanPrecision5,
		"precision_at_10":        metrics.MeanPrecision10,
		"recall_at_5":            metrics.MeanRecall5,
		"recall_at_10":           metrics.MeanRecall10,
		"ndcg_at_5":              metrics.MeanNDCG5,
		"ndcg_at_10":             metrics.MeanNDCG10,
		"diversity_at_5":         metrics.MeanDiversity5,
		"diversity_at_10":        metrics.MeanDiversity10,
		"serendipity":            metrics.MeanSerendipity,
		"cold_start_precision_5": metrics.ColdStartPrecision5,
	}

	for metricName, value := range metricValues {
		if target, ok := s.dataset.MetricTargets[metricName]; ok {
			status[metricName] = evaluateRecommendationMetric(value, target)
		}
	}

	return status
}

// evaluateRecommendationMetric returns "pass", "warning", or "critical" based on value vs target
func evaluateRecommendationMetric(value float64, target RecommendationTarget) string {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return "critical"
	}
	if value >= target.Target {
		return "pass"
	}
	if value >= target.WarningThreshold {
		return "warning"
	}
	return "critical"
}

// SimulateRecommendations creates simulated recommendations for testing
// Returns recommendations sorted by relevance (simulating ideal ranking)
func (s *RecommendationEvaluationService) SimulateRecommendations(
	scenario RecommendationScenario,
) ([]string, []string) {
	// Sort by relevance descending to simulate ideal ranking
	clips := make([]RelevantRecommendation, len(scenario.RelevantClips))
	copy(clips, scenario.RelevantClips)
	sort.Slice(clips, func(i, j int) bool {
		return clips[i].Relevance > clips[j].Relevance
	})

	ids := make([]string, len(clips))
	gameIDs := make([]string, len(clips))
	for i, clip := range clips {
		ids[i] = clip.ClipID
		gameIDs[i] = clip.GameID
	}

	return ids, gameIDs
}

// EvaluateWithSimulatedResults runs evaluation using simulated (ideal) results
func (s *RecommendationEvaluationService) EvaluateWithSimulatedResults(
	ctx context.Context,
) (*RecommendationEvaluationReport, error) {
	return s.EvaluateDataset(ctx, func(scenario RecommendationScenario) ([]string, []string, error) {
		ids, gameIDs := s.SimulateRecommendations(scenario)
		return ids, gameIDs, nil
	})
}
