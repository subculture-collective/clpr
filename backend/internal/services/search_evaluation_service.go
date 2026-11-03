package services

import (
	"context"
	"fmt"
	"math"
	"os"
	"sort"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// SearchEvaluationService evaluates search quality using standard IR metrics
type SearchEvaluationService struct {
	hybridSearchService *HybridSearchService
	dataset             *EvaluationDataset
}

// EvaluationDataset represents the loaded evaluation dataset
type EvaluationDataset struct {
	Version              string                  `yaml:"version"`
	Description          string                  `yaml:"description"`
	EvaluationQueries    []EvaluationQuery       `yaml:"evaluation_queries"`
	MetricTargets        map[string]MetricTarget `yaml:"metric_targets"`
	EvaluationGuidelines []string                `yaml:"evaluation_guidelines"`
}

// EvaluationQuery represents a single evaluation query with relevance judgments
type EvaluationQuery struct {
	ID                string             `yaml:"id"`
	Query             string             `yaml:"query"`
	Description       string             `yaml:"description"`
	RelevantDocuments []RelevantDocument `yaml:"relevant_documents"`
}

// RelevantDocument represents a document with its relevance score
type RelevantDocument struct {
	ClipID    string `yaml:"clip_id"`
	Relevance int    `yaml:"relevance"` // 0-4 scale
	Reason    string `yaml:"reason"`
}

// MetricTarget defines the target and threshold for a metric
type MetricTarget struct {
	Target            float64 `yaml:"target"`
	Description       string  `yaml:"description"`
	WarningThreshold  float64 `yaml:"warning_threshold"`
	CriticalThreshold float64 `yaml:"critical_threshold"`
}

// EvaluationResult contains the results of a search evaluation
type EvaluationResult struct {
	QueryID          string            `json:"query_id"`
	Query            string            `json:"query"`
	NDCG5            float64           `json:"ndcg_at_5"`
	NDCG10           float64           `json:"ndcg_at_10"`
	MRR              float64           `json:"mrr"`
	Precision5       float64           `json:"precision_at_5"`
	Precision10      float64           `json:"precision_at_10"`
	Precision20      float64           `json:"precision_at_20"`
	Recall5          float64           `json:"recall_at_5"`
	Recall10         float64           `json:"recall_at_10"`
	Recall20         float64           `json:"recall_at_20"`
	RetrievedResults int               `json:"retrieved_results"`
	RelevantResults  int               `json:"relevant_results"`
	QueryResults     []QueryResultItem `json:"query_results,omitempty"`
}

// QueryResultItem represents a single result with its relevance
type QueryResultItem struct {
	ClipID    string  `json:"clip_id"`
	Rank      int     `json:"rank"`
	Relevance int     `json:"relevance"`
	Score     float64 `json:"score,omitempty"`
}

// AggregateMetrics contains aggregated metrics across all queries
type AggregateMetrics struct {
	MeanNDCG5       float64 `json:"mean_ndcg_at_5"`
	MeanNDCG10      float64 `json:"mean_ndcg_at_10"`
	MeanMRR         float64 `json:"mean_mrr"`
	MeanPrecision5  float64 `json:"mean_precision_at_5"`
	MeanPrecision10 float64 `json:"mean_precision_at_10"`
	MeanPrecision20 float64 `json:"mean_precision_at_20"`
	MeanRecall5     float64 `json:"mean_recall_at_5"`
	MeanRecall10    float64 `json:"mean_recall_at_10"`
	MeanRecall20    float64 `json:"mean_recall_at_20"`
	QueryCount      int     `json:"query_count"`
}

// EvaluationReport contains the full evaluation report
type EvaluationReport struct {
	Metrics      AggregateMetrics        `json:"aggregate_metrics"`
	QueryResults []EvaluationResult      `json:"query_results"`
	Targets      map[string]MetricTarget `json:"targets,omitempty"`
	Status       map[string]string       `json:"status,omitempty"` // "pass", "warning", "critical"
}

// NewSearchEvaluationService creates a new search evaluation service
func NewSearchEvaluationService(hybridSearchService *HybridSearchService) *SearchEvaluationService {
	return &SearchEvaluationService{
		hybridSearchService: hybridSearchService,
	}
}

// LoadDataset loads the evaluation dataset from a YAML file
func (s *SearchEvaluationService) LoadDataset(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read dataset file: %w", err)
	}

	var dataset EvaluationDataset
	if err := yaml.Unmarshal(data, &dataset); err != nil {
		return fmt.Errorf("failed to parse dataset YAML: %w", err)
	}

	s.dataset = &dataset
	return nil
}

// GetDataset returns the loaded dataset
func (s *SearchEvaluationService) GetDataset() *EvaluationDataset {
	return s.dataset
}

// CalculateNDCG calculates Normalized Discounted Cumulative Gain at k
// DCG = sum((2^relevance_i - 1) / log2(i + 1)) for i in 1..k
// IDCG = DCG of the ideal ranking of all judged documents
// NDCG = DCG / IDCG
//
// relevances are the graded relevances of the retrieved documents in rank
// order. judgedRelevances are the relevances of every ground-truth judgment
// for the query, in any order; the ideal ranking is built from them, so a
// result list that misses the most relevant documents scores below 1.0.
func CalculateNDCG(relevances []int, judgedRelevances []int, k int) float64 {
	if len(relevances) == 0 || k <= 0 {
		return 0.0
	}

	// Calculate DCG for actual ranking
	dcg := calculateDCG(relevances, k)

	// Calculate IDCG (ideal DCG) by sorting the ground-truth relevances descending
	ideal := make([]int, len(judgedRelevances))
	copy(ideal, judgedRelevances)
	sort.Sort(sort.Reverse(sort.IntSlice(ideal)))
	idcg := calculateDCG(ideal, k)

	if idcg == 0 {
		return 0.0
	}

	return dcg / idcg
}

// calculateDCG computes Discounted Cumulative Gain at k
func calculateDCG(relevances []int, k int) float64 {
	dcg := 0.0
	limit := k
	if limit > len(relevances) {
		limit = len(relevances)
	}

	for i := 0; i < limit; i++ {
		// Using the formula: (2^rel - 1) / log2(i + 2)
		// This gives more weight to highly relevant documents
		// Relevances are expected to be 0-4, but we clamp to prevent overflow
		rel := relevances[i]
		if rel > 4 {
			rel = 4 // Clamp to max expected relevance
		}
		if rel < 0 {
			rel = 0 // Clamp to min relevance
		}
		gain := float64(int(1<<uint(rel)) - 1)
		discount := math.Log2(float64(i + 2))
		dcg += gain / discount
	}

	return dcg
}

// CalculateRR calculates Reciprocal Rank for a single query
// RR = 1 / rank of first relevant document
// Returns 0 if no relevant documents found
// Note: MRR (Mean Reciprocal Rank) is the mean of RR across multiple queries
func CalculateRR(relevances []int, relevanceThreshold int) float64 {
	for i, rel := range relevances {
		if rel >= relevanceThreshold {
			return 1.0 / float64(i+1)
		}
	}
	return 0.0
}

// CalculatePrecision calculates Precision at k
// Precision@k = (relevant documents in top k) / k
func CalculatePrecision(relevances []int, k int, relevanceThreshold int) float64 {
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

// CalculateRecall calculates Recall at k
// Recall@k = (relevant documents in top k) / (total relevant documents)
func CalculateRecall(relevances []int, k int, totalRelevant int, relevanceThreshold int) float64 {
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

// EvaluateQuery runs a single query and evaluates results against ground truth
func (s *SearchEvaluationService) EvaluateQuery(ctx context.Context, evalQuery EvaluationQuery, retrievedIDs []string) EvaluationResult {
	// Build relevance map from ground truth
	relevanceMap := make(map[string]int)
	totalRelevant := 0
	for _, doc := range evalQuery.RelevantDocuments {
		relevanceMap[doc.ClipID] = doc.Relevance
		if doc.Relevance >= 2 { // Consider relevance >= 2 as "relevant"
			totalRelevant++
		}
	}

	// Ground-truth relevances (one per judged document) for the ideal ranking
	judgedRelevances := make([]int, 0, len(relevanceMap))
	for _, rel := range relevanceMap {
		judgedRelevances = append(judgedRelevances, rel)
	}

	// Get relevances for retrieved documents
	relevances := make([]int, len(retrievedIDs))
	queryResults := make([]QueryResultItem, len(retrievedIDs))
	relevantRetrieved := 0

	for i, id := range retrievedIDs {
		rel := 0
		if r, ok := relevanceMap[id]; ok {
			rel = r
		}
		relevances[i] = rel
		queryResults[i] = QueryResultItem{
			ClipID:    id,
			Rank:      i + 1,
			Relevance: rel,
		}
		if rel >= 2 {
			relevantRetrieved++
		}
	}

	// Calculate metrics
	result := EvaluationResult{
		QueryID:          evalQuery.ID,
		Query:            evalQuery.Query,
		NDCG5:            CalculateNDCG(relevances, judgedRelevances, 5),
		NDCG10:           CalculateNDCG(relevances, judgedRelevances, 10),
		MRR:              CalculateRR(relevances, 2), // Threshold of 2 for "relevant"
		Precision5:       CalculatePrecision(relevances, 5, 2),
		Precision10:      CalculatePrecision(relevances, 10, 2),
		Precision20:      CalculatePrecision(relevances, 20, 2),
		Recall5:          CalculateRecall(relevances, 5, totalRelevant, 2),
		Recall10:         CalculateRecall(relevances, 10, totalRelevant, 2),
		Recall20:         CalculateRecall(relevances, 20, totalRelevant, 2),
		RetrievedResults: len(retrievedIDs),
		RelevantResults:  relevantRetrieved,
		QueryResults:     queryResults,
	}

	return result
}

// EvaluateDataset runs evaluation on the entire dataset
func (s *SearchEvaluationService) EvaluateDataset(ctx context.Context, resultsProvider func(query string) ([]string, error)) (*EvaluationReport, error) {
	if s.dataset == nil {
		return nil, fmt.Errorf("no dataset loaded")
	}

	results := make([]EvaluationResult, 0, len(s.dataset.EvaluationQueries))

	for _, evalQuery := range s.dataset.EvaluationQueries {
		// Get search results
		retrievedIDs, err := resultsProvider(evalQuery.Query)
		if err != nil {
			// Log error but continue with empty results
			retrievedIDs = []string{}
		}

		result := s.EvaluateQuery(ctx, evalQuery, retrievedIDs)
		results = append(results, result)
	}

	// Calculate aggregate metrics
	aggregate := s.calculateAggregateMetrics(results)

	// Check against targets
	status := s.checkTargets(aggregate)

	return &EvaluationReport{
		Metrics:      aggregate,
		QueryResults: results,
		Targets:      s.dataset.MetricTargets,
		Status:       status,
	}, nil
}

// calculateAggregateMetrics computes mean metrics across all queries
func (s *SearchEvaluationService) calculateAggregateMetrics(results []EvaluationResult) AggregateMetrics {
	if len(results) == 0 {
		return AggregateMetrics{}
	}

	var sumNDCG5, sumNDCG10, sumMRR, sumPrec5, sumPrec10, sumPrec20, sumRecall5, sumRecall10, sumRecall20 float64

	for _, r := range results {
		sumNDCG5 += r.NDCG5
		sumNDCG10 += r.NDCG10
		sumMRR += r.MRR
		sumPrec5 += r.Precision5
		sumPrec10 += r.Precision10
		sumPrec20 += r.Precision20
		sumRecall5 += r.Recall5
		sumRecall10 += r.Recall10
		sumRecall20 += r.Recall20
	}

	n := float64(len(results))
	return AggregateMetrics{
		MeanNDCG5:       sumNDCG5 / n,
		MeanNDCG10:      sumNDCG10 / n,
		MeanMRR:         sumMRR / n,
		MeanPrecision5:  sumPrec5 / n,
		MeanPrecision10: sumPrec10 / n,
		MeanPrecision20: sumPrec20 / n,
		MeanRecall5:     sumRecall5 / n,
		MeanRecall10:    sumRecall10 / n,
		MeanRecall20:    sumRecall20 / n,
		QueryCount:      len(results),
	}
}

// checkTargets compares metrics against targets and returns status
func (s *SearchEvaluationService) checkTargets(metrics AggregateMetrics) map[string]string {
	if s.dataset == nil || s.dataset.MetricTargets == nil {
		return nil
	}

	status := make(map[string]string)

	// Check nDCG@5
	if target, ok := s.dataset.MetricTargets["ndcg_at_5"]; ok {
		status["ndcg_at_5"] = evaluateMetric(metrics.MeanNDCG5, target)
	}

	// Check nDCG@10
	if target, ok := s.dataset.MetricTargets["ndcg_at_10"]; ok {
		status["ndcg_at_10"] = evaluateMetric(metrics.MeanNDCG10, target)
	}

	// Check MRR
	if target, ok := s.dataset.MetricTargets["mrr"]; ok {
		status["mrr"] = evaluateMetric(metrics.MeanMRR, target)
	}

	// Check Precision@5
	if target, ok := s.dataset.MetricTargets["precision_at_5"]; ok {
		status["precision_at_5"] = evaluateMetric(metrics.MeanPrecision5, target)
	}

	// Check Precision@10
	if target, ok := s.dataset.MetricTargets["precision_at_10"]; ok {
		status["precision_at_10"] = evaluateMetric(metrics.MeanPrecision10, target)
	}

	// Check Precision@20
	if target, ok := s.dataset.MetricTargets["precision_at_20"]; ok {
		status["precision_at_20"] = evaluateMetric(metrics.MeanPrecision20, target)
	}

	// Check Recall@5
	if target, ok := s.dataset.MetricTargets["recall_at_5"]; ok {
		status["recall_at_5"] = evaluateMetric(metrics.MeanRecall5, target)
	}

	// Check Recall@10
	if target, ok := s.dataset.MetricTargets["recall_at_10"]; ok {
		status["recall_at_10"] = evaluateMetric(metrics.MeanRecall10, target)
	}

	// Check Recall@20
	if target, ok := s.dataset.MetricTargets["recall_at_20"]; ok {
		status["recall_at_20"] = evaluateMetric(metrics.MeanRecall20, target)
	}

	return status
}

// evaluateMetric returns "pass", "warning", or "critical" based on value vs target
func evaluateMetric(value float64, target MetricTarget) string {
	if value >= target.Target {
		return "pass"
	}
	if value >= target.WarningThreshold {
		return "warning"
	}
	return "critical"
}

// SimulateResults creates simulated results for testing without live search
// This uses the ground truth data to simulate what a perfect search would return
func (s *SearchEvaluationService) SimulateResults(evalQuery EvaluationQuery) []string {
	// Sort by relevance descending to simulate ideal ranking
	docs := make([]RelevantDocument, len(evalQuery.RelevantDocuments))
	copy(docs, evalQuery.RelevantDocuments)
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].Relevance > docs[j].Relevance
	})

	ids := make([]string, len(docs))
	for i, doc := range docs {
		ids[i] = doc.ClipID
	}
	return ids
}

// EvaluateWithSimulatedResults runs evaluation using simulated (ideal) results
// This is useful for establishing baseline metrics and testing the evaluation system
func (s *SearchEvaluationService) EvaluateWithSimulatedResults(ctx context.Context) (*EvaluationReport, error) {
	return s.EvaluateDataset(ctx, func(query string) ([]string, error) {
		// Find the query in dataset and return simulated results
		for _, eq := range s.dataset.EvaluationQueries {
			if eq.Query == query {
				return s.SimulateResults(eq), nil
			}
		}
		return []string{}, nil
	})
}

// ConvertClipIDsToUUIDs attempts to convert clip IDs to UUIDs for actual search
// If the ID is already a valid UUID, it's used as-is
// Otherwise, a deterministic UUID is generated from the string
func ConvertClipIDsToUUIDs(ids []string) []uuid.UUID {
	uuids := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if u, err := uuid.Parse(id); err == nil {
			uuids = append(uuids, u)
		}
		// Skip non-UUID strings for now
	}
	return uuids
}
