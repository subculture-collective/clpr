package services

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateNDCG(t *testing.T) {
	tests := []struct {
		name       string
		relevances []int
		judged     []int
		k          int
		expected   float64
		tolerance  float64
	}{
		{
			name:       "perfect ranking",
			relevances: []int{4, 3, 2, 1, 0},
			k:          5,
			expected:   1.0,
			tolerance:  0.001,
		},
		{
			name:       "inverse ranking",
			relevances: []int{0, 1, 2, 3, 4},
			k:          5,
			expected:   0.513, // DCG of inverse / IDCG
			tolerance:  0.01,
		},
		{
			name:       "all zeros",
			relevances: []int{0, 0, 0, 0, 0},
			k:          5,
			expected:   0.0,
			tolerance:  0.001,
		},
		{
			name:       "single relevant at top",
			relevances: []int{4, 0, 0, 0, 0},
			k:          5,
			expected:   1.0,
			tolerance:  0.001,
		},
		{
			name:       "k smaller than results",
			relevances: []int{4, 3, 2, 1, 0, 0, 0},
			k:          3,
			expected:   1.0,
			tolerance:  0.001,
		},
		{
			name:       "k larger than results",
			relevances: []int{4, 3},
			k:          10,
			expected:   1.0,
			tolerance:  0.001,
		},
		{
			name:       "empty relevances",
			relevances: []int{},
			k:          5,
			expected:   0.0,
			tolerance:  0.001,
		},
		{
			name:       "k is zero",
			relevances: []int{4, 3, 2, 1, 0},
			k:          0,
			expected:   0.0,
			tolerance:  0.001,
		},
		{
			name:       "binary relevance",
			relevances: []int{1, 0, 1, 0, 1},
			k:          5,
			expected:   0.885, // Calculated: DCG / IDCG for binary
			tolerance:  0.01,
		},
		{
			// Retrieved docs are in perfect order among themselves, but the
			// two best judged docs (4, 3) were never retrieved.
			// DCG = 3/1 + 1/log2(3) = 3.631; IDCG = 15 + 7/log2(3) + 3/2 + 1/log2(5) = 21.347
			name:       "missing best judged documents",
			relevances: []int{2, 1},
			judged:     []int{4, 3, 2, 1},
			k:          5,
			expected:   0.170,
			tolerance:  0.001,
		},
		{
			// Unjudged retrieved docs count as relevance 0; the ideal ranking
			// still comes from all four judgments.
			// DCG = 15/log2(3) = 9.464; IDCG = 15 + 15/log2(3) + 7/2 + 7/log2(5) = 30.979
			name:       "best documents ranked below an unjudged one",
			relevances: []int{0, 4},
			judged:     []int{4, 4, 3, 3},
			k:          5,
			expected:   0.305,
			tolerance:  0.001,
		},
		{
			name:       "nothing judged relevant",
			relevances: []int{0, 0},
			judged:     []int{},
			k:          5,
			expected:   0.0,
			tolerance:  0.001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Unless a case sets judged, the ground truth is exactly the
			// retrieved set, so the ideal ranking is those relevances sorted.
			judged := tt.judged
			if judged == nil {
				judged = tt.relevances
			}
			result := CalculateNDCG(tt.relevances, judged, tt.k)
			assert.InDelta(t, tt.expected, result, tt.tolerance,
				"NDCG should be approximately %v, got %v", tt.expected, result)
		})
	}
}

func TestCalculateRR(t *testing.T) {
	tests := []struct {
		name       string
		relevances []int
		threshold  int
		expected   float64
	}{
		{
			name:       "first result relevant",
			relevances: []int{3, 1, 0},
			threshold:  2,
			expected:   1.0, // 1/1
		},
		{
			name:       "second result relevant",
			relevances: []int{1, 3, 0},
			threshold:  2,
			expected:   0.5, // 1/2
		},
		{
			name:       "third result relevant",
			relevances: []int{1, 1, 3, 0},
			threshold:  2,
			expected:   1.0 / 3.0, // 1/3
		},
		{
			name:       "no relevant results",
			relevances: []int{1, 0, 1, 0},
			threshold:  2,
			expected:   0.0,
		},
		{
			name:       "empty relevances",
			relevances: []int{},
			threshold:  2,
			expected:   0.0,
		},
		{
			name:       "all relevant",
			relevances: []int{4, 4, 4},
			threshold:  2,
			expected:   1.0,
		},
		{
			name:       "threshold at exact value",
			relevances: []int{1, 2, 3},
			threshold:  2,
			expected:   0.5, // 1/2 (second result is exactly at threshold)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateRR(tt.relevances, tt.threshold)
			assert.Equal(t, tt.expected, result,
				"MRR should be %v, got %v", tt.expected, result)
		})
	}
}

func TestCalculatePrecision(t *testing.T) {
	tests := []struct {
		name       string
		relevances []int
		k          int
		threshold  int
		expected   float64
	}{
		{
			name:       "all relevant",
			relevances: []int{4, 3, 2, 3, 4},
			k:          5,
			threshold:  2,
			expected:   1.0,
		},
		{
			name:       "none relevant",
			relevances: []int{0, 1, 0, 1, 0},
			k:          5,
			threshold:  2,
			expected:   0.0,
		},
		{
			name:       "half relevant",
			relevances: []int{4, 0, 3, 0, 2},
			k:          5,
			threshold:  2,
			expected:   0.6, // 3/5
		},
		{
			name:       "k smaller than results",
			relevances: []int{4, 3, 0, 0, 0, 0},
			k:          3,
			threshold:  2,
			expected:   2.0 / 3.0, // 2/3
		},
		{
			name:       "k larger than results",
			relevances: []int{4, 3},
			k:          5,
			threshold:  2,
			expected:   0.4, // 2/5 (zeros padded)
		},
		{
			name:       "empty relevances",
			relevances: []int{},
			k:          5,
			threshold:  2,
			expected:   0.0,
		},
		{
			name:       "k is zero",
			relevances: []int{4, 3, 2},
			k:          0,
			threshold:  2,
			expected:   0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculatePrecision(tt.relevances, tt.k, tt.threshold)
			assert.InDelta(t, tt.expected, result, 0.001,
				"Precision@%d should be %v, got %v", tt.k, tt.expected, result)
		})
	}
}

func TestCalculateRecall(t *testing.T) {
	tests := []struct {
		name          string
		relevances    []int
		k             int
		totalRelevant int
		threshold     int
		expected      float64
	}{
		{
			name:          "all found",
			relevances:    []int{4, 3, 2},
			k:             5,
			totalRelevant: 3,
			threshold:     2,
			expected:      1.0,
		},
		{
			name:          "half found",
			relevances:    []int{4, 0, 3, 0, 0},
			k:             5,
			totalRelevant: 4,
			threshold:     2,
			expected:      0.5, // 2/4
		},
		{
			name:          "none found",
			relevances:    []int{1, 0, 1, 0, 0},
			k:             5,
			totalRelevant: 3,
			threshold:     2,
			expected:      0.0,
		},
		{
			name:          "zero total relevant",
			relevances:    []int{4, 3, 2},
			k:             5,
			totalRelevant: 0,
			threshold:     2,
			expected:      0.0,
		},
		{
			name:          "k smaller than results",
			relevances:    []int{4, 3, 2, 0, 0},
			k:             2,
			totalRelevant: 3,
			threshold:     2,
			expected:      2.0 / 3.0, // 2/3
		},
		{
			name:          "empty relevances",
			relevances:    []int{},
			k:             5,
			totalRelevant: 3,
			threshold:     2,
			expected:      0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateRecall(tt.relevances, tt.k, tt.totalRelevant, tt.threshold)
			assert.InDelta(t, tt.expected, result, 0.001,
				"Recall@%d should be %v, got %v", tt.k, tt.expected, result)
		})
	}
}

func TestSearchEvaluationService_LoadDataset(t *testing.T) {
	// Create a temporary test dataset
	yamlContent := `
version: "1.0"
description: "Test dataset"
evaluation_queries:
  - id: "test-001"
    query: "test query"
    description: "Test description"
    relevant_documents:
      - clip_id: "clip-1"
        relevance: 4
        reason: "Perfect match"
      - clip_id: "clip-2"
        relevance: 2
        reason: "Partial match"
metric_targets:
  ndcg_at_5:
    target: 0.75
    description: "nDCG@5"
    warning_threshold: 0.70
    critical_threshold: 0.60
`
	// Write temp file
	tmpFile, err := os.CreateTemp("", "test_dataset_*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(yamlContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Test loading
	service := NewSearchEvaluationService(nil)
	err = service.LoadDataset(tmpFile.Name())
	require.NoError(t, err)

	dataset := service.GetDataset()
	assert.NotNil(t, dataset)
	assert.Equal(t, "1.0", dataset.Version)
	assert.Len(t, dataset.EvaluationQueries, 1)
	assert.Len(t, dataset.EvaluationQueries[0].RelevantDocuments, 2)
}

func TestSearchEvaluationService_LoadDataset_FileNotFound(t *testing.T) {
	service := NewSearchEvaluationService(nil)
	err := service.LoadDataset("/nonexistent/file.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read dataset file")
}

func TestSearchEvaluationService_LoadDataset_InvalidYAML(t *testing.T) {
	// Create invalid YAML file
	tmpFile, err := os.CreateTemp("", "invalid_yaml_*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("invalid: yaml: content: ][")
	require.NoError(t, err)
	tmpFile.Close()

	service := NewSearchEvaluationService(nil)
	err = service.LoadDataset(tmpFile.Name())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse dataset YAML")
}

func TestSearchEvaluationService_EvaluateQuery(t *testing.T) {
	service := NewSearchEvaluationService(nil)

	evalQuery := EvaluationQuery{
		ID:          "test-001",
		Query:       "test query",
		Description: "Test",
		RelevantDocuments: []RelevantDocument{
			{ClipID: "clip-1", Relevance: 4, Reason: "Perfect"},
			{ClipID: "clip-2", Relevance: 3, Reason: "Good"},
			{ClipID: "clip-3", Relevance: 2, Reason: "Fair"},
			{ClipID: "clip-4", Relevance: 1, Reason: "Poor"},
			{ClipID: "clip-5", Relevance: 0, Reason: "Not relevant"},
		},
	}

	t.Run("perfect ranking", func(t *testing.T) {
		// Perfect ranking: all relevant docs at top in order
		retrieved := []string{"clip-1", "clip-2", "clip-3", "clip-4", "clip-5"}
		result := service.EvaluateQuery(context.Background(), evalQuery, retrieved)

		assert.Equal(t, "test-001", result.QueryID)
		assert.Equal(t, "test query", result.Query)
		assert.Equal(t, 1.0, result.NDCG5)
		assert.Equal(t, 1.0, result.MRR)        // First result is relevant (rel >= 2)
		assert.Equal(t, 0.6, result.Precision5) // 3 relevant in top 5 (rel >= 2)
		assert.Equal(t, 5, result.RetrievedResults)
	})

	t.Run("worst ranking", func(t *testing.T) {
		// Inverse ranking
		retrieved := []string{"clip-5", "clip-4", "clip-3", "clip-2", "clip-1"}
		result := service.EvaluateQuery(context.Background(), evalQuery, retrieved)

		assert.Less(t, result.NDCG5, 1.0)
		assert.InDelta(t, 1.0/3.0, result.MRR, 0.001) // First relevant at position 3
	})

	t.Run("no results", func(t *testing.T) {
		retrieved := []string{}
		result := service.EvaluateQuery(context.Background(), evalQuery, retrieved)

		assert.Equal(t, 0.0, result.NDCG5)
		assert.Equal(t, 0.0, result.MRR)
		assert.Equal(t, 0, result.RetrievedResults)
	})

	t.Run("unknown documents", func(t *testing.T) {
		// Documents not in ground truth get relevance 0
		retrieved := []string{"unknown-1", "unknown-2", "clip-1"}
		result := service.EvaluateQuery(context.Background(), evalQuery, retrieved)

		assert.InDelta(t, 1.0/3.0, result.MRR, 0.001) // First relevant at position 3
	})
}

func TestSearchEvaluationService_EvaluateWithSimulatedResults(t *testing.T) {
	// Create a test dataset
	yamlContent := `
version: "1.0"
description: "Test dataset"
evaluation_queries:
  - id: "eval-001"
    query: "test query 1"
    description: "Test 1"
    relevant_documents:
      - clip_id: "clip-1"
        relevance: 4
        reason: "Perfect"
      - clip_id: "clip-2"
        relevance: 2
        reason: "Fair"
  - id: "eval-002"
    query: "test query 2"
    description: "Test 2"
    relevant_documents:
      - clip_id: "clip-3"
        relevance: 3
        reason: "Good"
metric_targets:
  ndcg_at_5:
    target: 0.75
    warning_threshold: 0.70
    critical_threshold: 0.60
`
	tmpFile, err := os.CreateTemp("", "test_eval_*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(yamlContent)
	require.NoError(t, err)
	tmpFile.Close()

	service := NewSearchEvaluationService(nil)
	err = service.LoadDataset(tmpFile.Name())
	require.NoError(t, err)

	// Run evaluation with simulated results (ideal ranking)
	report, err := service.EvaluateWithSimulatedResults(context.Background())
	require.NoError(t, err)

	assert.Equal(t, 2, report.Metrics.QueryCount)
	assert.Equal(t, 1.0, report.Metrics.MeanNDCG5) // Perfect ranking should give 1.0
	assert.NotNil(t, report.Status)
}

func TestSearchEvaluationService_SimulateResults(t *testing.T) {
	service := NewSearchEvaluationService(nil)

	evalQuery := EvaluationQuery{
		ID:          "test-001",
		Query:       "test",
		Description: "Test",
		RelevantDocuments: []RelevantDocument{
			{ClipID: "clip-3", Relevance: 2, Reason: "Fair"},
			{ClipID: "clip-1", Relevance: 4, Reason: "Perfect"},
			{ClipID: "clip-2", Relevance: 3, Reason: "Good"},
		},
	}

	results := service.SimulateResults(evalQuery)

	// Should be sorted by relevance descending
	assert.Equal(t, []string{"clip-1", "clip-2", "clip-3"}, results)
}

func TestEvaluateMetric(t *testing.T) {
	target := MetricTarget{
		Target:            0.75,
		WarningThreshold:  0.70,
		CriticalThreshold: 0.60,
	}

	tests := []struct {
		value    float64
		expected string
	}{
		{0.80, "pass"},
		{0.75, "pass"},
		{0.74, "warning"},
		{0.70, "warning"},
		{0.69, "critical"},
		{0.50, "critical"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := evaluateMetric(tt.value, target)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAggregateMetrics(t *testing.T) {
	service := NewSearchEvaluationService(nil)

	results := []EvaluationResult{
		{
			NDCG5:      0.8,
			NDCG10:     0.85,
			MRR:        1.0,
			Precision5: 0.6,
			Recall10:   0.7,
		},
		{
			NDCG5:      0.6,
			NDCG10:     0.65,
			MRR:        0.5,
			Precision5: 0.4,
			Recall10:   0.5,
		},
	}

	aggregate := service.calculateAggregateMetrics(results)

	assert.Equal(t, 2, aggregate.QueryCount)
	assert.InDelta(t, 0.7, aggregate.MeanNDCG5, 0.001)
	assert.InDelta(t, 0.75, aggregate.MeanNDCG10, 0.001)
	assert.InDelta(t, 0.75, aggregate.MeanMRR, 0.001)
	assert.InDelta(t, 0.5, aggregate.MeanPrecision5, 0.001)
	assert.InDelta(t, 0.6, aggregate.MeanRecall10, 0.001)
}

func TestConvertClipIDsToUUIDs(t *testing.T) {
	ids := []string{
		"550e8400-e29b-41d4-a716-446655440000", // Valid UUID
		"not-a-uuid",                           // Invalid
		"660e8400-e29b-41d4-a716-446655440001", // Valid UUID
	}

	uuids := ConvertClipIDsToUUIDs(ids)

	assert.Len(t, uuids, 2)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", uuids[0].String())
	assert.Equal(t, "660e8400-e29b-41d4-a716-446655440001", uuids[1].String())
}
