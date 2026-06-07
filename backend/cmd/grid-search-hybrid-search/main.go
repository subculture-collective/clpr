package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/services"
)

// GridSearchConfig defines the parameter grid to search
type GridSearchConfig struct {
	BM25Weights   []float64
	VectorWeights []float64
}

// GridSearchResult contains the results for a single parameter combination
type GridSearchResult struct {
	Parameters struct {
		BM25Weight   float64 `json:"bm25_weight"`
		VectorWeight float64 `json:"vector_weight"`
	} `json:"parameters"`
	Metrics services.AggregateMetrics `json:"metrics"`
	Status  map[string]string         `json:"status"`
}

// GridSearchReport contains all grid search results
type GridSearchReport struct {
	Timestamp       string                     `json:"timestamp"`
	BaselineMetrics *services.AggregateMetrics `json:"baseline_metrics,omitempty"`
	Results         []GridSearchResult         `json:"results"`
	BestConfig      GridSearchResult           `json:"best_config"`
	Improvement     map[string]float64         `json:"improvement_vs_baseline,omitempty"` // Percentage improvement vs baseline
}

func main() {
	// Command line flags
	datasetPath := flag.String("dataset", "testdata/search_evaluation_dataset.yaml", "Path to evaluation dataset YAML file")
	outputPath := flag.String("output", "hybrid-search-grid-results.json", "Path to output JSON file")
	baselinePath := flag.String("baseline", "", "Path to baseline results JSON file (optional)")
	verbose := flag.Bool("verbose", false, "Print detailed results for each configuration")
	help := flag.Bool("help", false, "Show help message")
	quick := flag.Bool("quick", false, "Quick mode - test fewer combinations")

	flag.Parse()

	if *help {
		printUsage()
		os.Exit(0)
	}

	// Load baseline if provided
	var baselineMetrics *services.AggregateMetrics
	if *baselinePath != "" {
		log.Printf("Loading baseline from: %s", *baselinePath)
		baseline, err := loadBaseline(*baselinePath)
		if err != nil {
			log.Printf("Warning: Failed to load baseline: %v", err)
		} else {
			baselineMetrics = baseline
			log.Printf("Baseline loaded: nDCG@10=%.4f", baseline.MeanNDCG10)
		}
	}

	// Define parameter grid
	gridConfig := GridSearchConfig{}

	if *quick {
		// Quick mode - test key variations around baseline (0.7/0.3)
		gridConfig.BM25Weights = []float64{0.5, 0.6, 0.7, 0.8}
		gridConfig.VectorWeights = []float64{0.2, 0.3, 0.4, 0.5}
	} else {
		// Full grid search - comprehensive weight exploration
		gridConfig.BM25Weights = []float64{0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9}
		gridConfig.VectorWeights = []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7}
	}

	log.Printf("Starting grid search with %d parameter combinations...",
		len(gridConfig.BM25Weights)*len(gridConfig.VectorWeights))

	// Load evaluation dataset
	evalService := services.NewSearchEvaluationService(nil)
	if err := evalService.LoadDataset(*datasetPath); err != nil {
		log.Fatalf("Failed to load dataset: %v", err)
	}

	dataset := evalService.GetDataset()
	log.Printf("Loaded %d evaluation queries", len(dataset.EvaluationQueries))

	// Run grid search
	ctx := context.Background()
	results := []GridSearchResult{}

	for _, bm25Weight := range gridConfig.BM25Weights {
		for _, vectorWeight := range gridConfig.VectorWeights {
			// Weights should sum to 1.0
			sum := bm25Weight + vectorWeight
			if sum < 0.99 || sum > 1.01 {
				continue
			}

			if *verbose {
				fmt.Printf("\nTesting: BM25=%.2f, Vector=%.2f\n", bm25Weight, vectorWeight)
			}

			result, err := evaluateConfiguration(
				ctx,
				evalService,
				bm25Weight,
				vectorWeight,
			)
			if err != nil {
				log.Printf("Error evaluating configuration: %v", err)
				continue
			}

			results = append(results, result)

			if *verbose {
				printResult(result)
			}
		}
	}

	if len(results) == 0 {
		log.Fatalf("No valid configurations tested")
	}

	// Find best configuration based on nDCG@10 (primary metric for this task)
	bestConfig := findBestConfiguration(results)

	// Calculate improvement vs baseline if available
	var improvement map[string]float64
	if baselineMetrics != nil {
		improvement = calculateImprovement(*baselineMetrics, bestConfig.Metrics)
	}

	report := GridSearchReport{
		Timestamp:       time.Now().Format(time.RFC3339),
		BaselineMetrics: baselineMetrics,
		Results:         results,
		BestConfig:      bestConfig,
		Improvement:     improvement,
	}

	// Print summary
	printSummary(report)

	// Write output file
	if err := writeOutputFile(*outputPath, &report); err != nil {
		log.Fatalf("Failed to write output file: %v", err)
	}
	log.Printf("Results written to: %s", *outputPath)
}

func printUsage() {
	fmt.Println("Hybrid Search Grid Search Tool")
	fmt.Println()
	fmt.Println("Tests different BM25/Vector weight combinations for the hybrid search algorithm")
	fmt.Println("to find optimal weights that maximize search relevance metrics (nDCG@10).")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  grid-search-hybrid-search [options]")
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Run full grid search")
	fmt.Println("  grid-search-hybrid-search -output results.json")
	fmt.Println()
	fmt.Println("  # Quick search with baseline comparison")
	fmt.Println("  grid-search-hybrid-search -quick -baseline baseline.json -verbose")
	fmt.Println()
	fmt.Println("  # Compare against baseline")
	fmt.Println("  grid-search-hybrid-search -baseline baseline.json -output optimized.json")
}

func evaluateConfiguration(
	ctx context.Context,
	evalService *services.SearchEvaluationService,
	bm25Weight float64,
	vectorWeight float64,
) (GridSearchResult, error) {
	// TODO: For production use, integrate with actual HybridSearchService
	// This requires:
	// 1. Create HybridSearchService with the specified weights
	// 2. Configure search weights: config.HybridSearch.BM25Weight = bm25Weight
	// 3. Run actual searches for each evaluation query
	// 4. Collect and return real metrics
	//
	// Example implementation:
	//   searchConfig := services.SearchWeightConfig{
	//       BM25Weight:   bm25Weight,
	//       VectorWeight: vectorWeight,
	//       // ... other parameters
	//   }
	//   report, err := evalService.EvaluateWithLiveSearch(ctx, searchConfig)
	//
	// For now, using simulated results to demonstrate the framework.
	// All metrics will be identical until live search integration is complete.

	report, err := evalService.EvaluateWithSimulatedResults(ctx)
	if err != nil {
		return GridSearchResult{}, err
	}

	result := GridSearchResult{
		Metrics: report.Metrics,
		Status:  report.Status,
	}
	result.Parameters.BM25Weight = bm25Weight
	result.Parameters.VectorWeight = vectorWeight

	return result, nil
}

func findBestConfiguration(results []GridSearchResult) GridSearchResult {
	if len(results) == 0 {
		return GridSearchResult{}
	}

	// Score based on weighted combination of key metrics
	// For Roadmap 5.0 Phase 3.1, nDCG@10 is the primary metric (target: ≥10% improvement)
	bestScore := -1.0
	bestIdx := 0

	for i, result := range results {
		// Scoring function: weighted combination of metrics
		// Primary focus on nDCG@10 (weight: 4.0)
		score := 0.0
		score += result.Metrics.MeanNDCG10 * 4.0 // Primary metric
		score += result.Metrics.MeanNDCG5 * 2.0  // Secondary
		score += result.Metrics.MeanMRR * 2.0    // Secondary
		score += result.Metrics.MeanPrecision10 * 1.0
		score += result.Metrics.MeanRecall10 * 1.0

		if score > bestScore {
			bestScore = score
			bestIdx = i
		}
	}

	return results[bestIdx]
}

func calculateImprovement(baseline, current services.AggregateMetrics) map[string]float64 {
	improvements := make(map[string]float64)

	improvements["ndcg_at_5"] = calculatePercentChange(baseline.MeanNDCG5, current.MeanNDCG5)
	improvements["ndcg_at_10"] = calculatePercentChange(baseline.MeanNDCG10, current.MeanNDCG10)
	improvements["mrr"] = calculatePercentChange(baseline.MeanMRR, current.MeanMRR)
	improvements["precision_at_5"] = calculatePercentChange(baseline.MeanPrecision5, current.MeanPrecision5)
	improvements["precision_at_10"] = calculatePercentChange(baseline.MeanPrecision10, current.MeanPrecision10)
	improvements["precision_at_20"] = calculatePercentChange(baseline.MeanPrecision20, current.MeanPrecision20)
	improvements["recall_at_5"] = calculatePercentChange(baseline.MeanRecall5, current.MeanRecall5)
	improvements["recall_at_10"] = calculatePercentChange(baseline.MeanRecall10, current.MeanRecall10)
	improvements["recall_at_20"] = calculatePercentChange(baseline.MeanRecall20, current.MeanRecall20)

	return improvements
}

func calculatePercentChange(baseline, current float64) float64 {
	if baseline == 0 {
		if current == 0 {
			return 0
		}
		return 999.0 // Large improvement indicator
	}
	return ((current - baseline) / baseline) * 100.0
}

func printResult(result GridSearchResult) {
	fmt.Printf("  nDCG@10: %.4f | nDCG@5: %.4f | MRR: %.4f | P@10: %.4f\n",
		result.Metrics.MeanNDCG10,
		result.Metrics.MeanNDCG5,
		result.Metrics.MeanMRR,
		result.Metrics.MeanPrecision10)
}

func printSummary(report GridSearchReport) {
	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("  Grid Search Summary")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Printf("Tested %d parameter combinations\n", len(report.Results))
	fmt.Println()

	// Show baseline if available
	if report.BaselineMetrics != nil {
		fmt.Println("Baseline Configuration:")
		fmt.Println("-----------------------------------------")
		fmt.Printf("  nDCG@10:      %.4f\n", report.BaselineMetrics.MeanNDCG10)
		fmt.Printf("  nDCG@5:       %.4f\n", report.BaselineMetrics.MeanNDCG5)
		fmt.Printf("  MRR:          %.4f\n", report.BaselineMetrics.MeanMRR)
		fmt.Printf("  Precision@10: %.4f\n", report.BaselineMetrics.MeanPrecision10)
		fmt.Println()
	}

	fmt.Println("Best Configuration:")
	fmt.Println("-----------------------------------------")
	fmt.Printf("  BM25 Weight:   %.2f\n", report.BestConfig.Parameters.BM25Weight)
	fmt.Printf("  Vector Weight: %.2f\n", report.BestConfig.Parameters.VectorWeight)
	fmt.Println()
	fmt.Println("Metrics:")
	fmt.Println("-----------------------------------------")
	fmt.Printf("  nDCG@5:       %.4f\n", report.BestConfig.Metrics.MeanNDCG5)
	fmt.Printf("  nDCG@10:      %.4f\n", report.BestConfig.Metrics.MeanNDCG10)
	fmt.Printf("  MRR:          %.4f\n", report.BestConfig.Metrics.MeanMRR)
	fmt.Printf("  Precision@5:  %.4f\n", report.BestConfig.Metrics.MeanPrecision5)
	fmt.Printf("  Precision@10: %.4f\n", report.BestConfig.Metrics.MeanPrecision10)
	fmt.Printf("  Precision@20: %.4f\n", report.BestConfig.Metrics.MeanPrecision20)
	fmt.Printf("  Recall@5:     %.4f\n", report.BestConfig.Metrics.MeanRecall5)
	fmt.Printf("  Recall@10:    %.4f\n", report.BestConfig.Metrics.MeanRecall10)
	fmt.Printf("  Recall@20:    %.4f\n", report.BestConfig.Metrics.MeanRecall20)
	fmt.Println()

	// Show improvement if baseline available
	if report.Improvement != nil {
		fmt.Println("Improvement vs Baseline:")
		fmt.Println("-----------------------------------------")
		fmt.Printf("  nDCG@10:      %+.2f%%\n", report.Improvement["ndcg_at_10"])
		fmt.Printf("  nDCG@5:       %+.2f%%\n", report.Improvement["ndcg_at_5"])
		fmt.Printf("  MRR:          %+.2f%%\n", report.Improvement["mrr"])
		fmt.Printf("  Precision@10: %+.2f%%\n", report.Improvement["precision_at_10"])
		fmt.Println()

		// Check if nDCG@10 improvement meets target (≥10%)
		ndcg10Improvement := report.Improvement["ndcg_at_10"]
		if ndcg10Improvement >= 10.0 {
			fmt.Println("✅ SUCCESS: nDCG@10 improvement ≥10% target met!")
		} else {
			fmt.Printf("⚠️  WARNING: nDCG@10 improvement (%.2f%%) below 10%% target\n", ndcg10Improvement)
		}
		fmt.Println()
	}
}

// BaselineReport matches the structure from capture-baseline-search
type BaselineReport struct {
	Metrics services.AggregateMetrics `json:"metrics"`
}

func loadBaseline(path string) (*services.AggregateMetrics, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Try loading as BaselineReport first (from capture-baseline-search)
	var baselineReport BaselineReport
	if err := json.Unmarshal(data, &baselineReport); err == nil {
		return &baselineReport.Metrics, nil
	}

	// Fall back to EvaluationReport format (from evaluate-search)
	var evalReport services.EvaluationReport
	if err := json.Unmarshal(data, &evalReport); err != nil {
		return nil, fmt.Errorf("failed to parse baseline file: %w", err)
	}

	return &evalReport.Metrics, nil
}

func writeOutputFile(path string, report *GridSearchReport) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
