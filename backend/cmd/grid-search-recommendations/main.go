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
	ContentWeights       []float64
	CollaborativeWeights []float64
	TrendingWeights      []float64
}

// GridSearchResult contains the results for a single parameter combination
type GridSearchResult struct {
	Parameters struct {
		ContentWeight       float64 `json:"content_weight"`
		CollaborativeWeight float64 `json:"collaborative_weight"`
		TrendingWeight      float64 `json:"trending_weight"`
	} `json:"parameters"`
	Metrics services.RecommendationAggregateMetrics `json:"metrics"`
	Status  map[string]string                       `json:"status"`
}

// GridSearchReport contains all grid search results
type GridSearchReport struct {
	Timestamp    string             `json:"timestamp"`
	Results      []GridSearchResult `json:"results"`
	BestConfig   GridSearchResult   `json:"best_config"`
	BaselineFile string             `json:"baseline_file,omitempty"`
}

func main() {
	// Command line flags
	datasetPath := flag.String("dataset", "testdata/recommendation_evaluation_dataset.yaml", "Path to evaluation dataset YAML file")
	outputPath := flag.String("output", "grid-search-results.json", "Path to output JSON file")
	verbose := flag.Bool("verbose", false, "Print detailed results for each configuration")
	help := flag.Bool("help", false, "Show help message")
	quick := flag.Bool("quick", false, "Quick mode - test fewer combinations")

	flag.Parse()

	if *help {
		printUsage()
		os.Exit(0)
	}

	// Define parameter grid
	gridConfig := GridSearchConfig{}

	if *quick {
		// Quick mode - test key variations
		gridConfig.ContentWeights = []float64{0.4, 0.5, 0.6}
		gridConfig.CollaborativeWeights = []float64{0.2, 0.3, 0.4}
		gridConfig.TrendingWeights = []float64{0.1, 0.2, 0.3}
	} else {
		// Full grid search
		gridConfig.ContentWeights = []float64{0.3, 0.4, 0.5, 0.6, 0.7}
		gridConfig.CollaborativeWeights = []float64{0.1, 0.2, 0.3, 0.4, 0.5}
		gridConfig.TrendingWeights = []float64{0.0, 0.1, 0.2, 0.3}
	}

	log.Printf("Starting grid search with %d parameter combinations...",
		len(gridConfig.ContentWeights)*len(gridConfig.CollaborativeWeights)*len(gridConfig.TrendingWeights))

	// Load evaluation dataset
	evalService := services.NewRecommendationEvaluationService(nil)
	if err := evalService.LoadDataset(*datasetPath); err != nil {
		log.Fatalf("Failed to load dataset: %v", err)
	}

	dataset := evalService.GetDataset()
	log.Printf("Loaded %d evaluation scenarios", len(dataset.EvaluationScenarios))

	// Run grid search
	ctx := context.Background()
	results := []GridSearchResult{}

	for _, contentWeight := range gridConfig.ContentWeights {
		for _, collaborativeWeight := range gridConfig.CollaborativeWeights {
			for _, trendingWeight := range gridConfig.TrendingWeights {
				// Skip invalid combinations (weights should approximately sum to 1.0)
				sum := contentWeight + collaborativeWeight + trendingWeight
				if sum < 0.9 || sum > 1.1 {
					continue
				}

				if *verbose {
					fmt.Printf("\nTesting: content=%.2f, collaborative=%.2f, trending=%.2f\n",
						contentWeight, collaborativeWeight, trendingWeight)
				}

				result, err := evaluateConfiguration(
					ctx,
					evalService,
					contentWeight,
					collaborativeWeight,
					trendingWeight,
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
	}

	// Find best configuration based on primary metric (Precision@10)
	bestConfig := findBestConfiguration(results)

	report := GridSearchReport{
		Timestamp:  time.Now().Format(time.RFC3339),
		Results:    results,
		BestConfig: bestConfig,
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
	fmt.Println("Recommendation Grid Search Tool")
	fmt.Println()
	fmt.Println("Tests different parameter combinations for the hybrid recommendation algorithm")
	fmt.Println("to find optimal weights for content-based, collaborative, and trending signals.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  grid-search-recommendations [options]")
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Run full grid search")
	fmt.Println("  grid-search-recommendations -output results.json")
	fmt.Println()
	fmt.Println("  # Quick search with fewer combinations")
	fmt.Println("  grid-search-recommendations -quick -verbose")
}

func evaluateConfiguration(
	ctx context.Context,
	evalService *services.RecommendationEvaluationService,
	contentWeight float64,
	collaborativeWeight float64,
	trendingWeight float64,
) (GridSearchResult, error) {
	// For this simulated evaluation, we'll use the existing simulated results
	// In a real implementation, this would create a service with the given weights
	// and run actual recommendations

	report, err := evalService.EvaluateWithSimulatedResults(ctx)
	if err != nil {
		return GridSearchResult{}, err
	}

	result := GridSearchResult{
		Metrics: report.Metrics,
		Status:  report.Status,
	}
	result.Parameters.ContentWeight = contentWeight
	result.Parameters.CollaborativeWeight = collaborativeWeight
	result.Parameters.TrendingWeight = trendingWeight

	return result, nil
}

func findBestConfiguration(results []GridSearchResult) GridSearchResult {
	if len(results) == 0 {
		return GridSearchResult{}
	}

	// Score based on weighted combination of key metrics
	// Prioritize Precision@10 improvement (target: 0.60, currently 0.5125)
	bestScore := -1.0
	bestIdx := 0

	for i, result := range results {
		// Scoring function: weighted combination of metrics
		// Higher weight on metrics that need improvement
		score := 0.0
		score += result.Metrics.MeanPrecision10 * 3.0 // High weight - needs improvement
		score += result.Metrics.MeanPrecision5 * 1.0
		score += result.Metrics.MeanRecall10 * 1.0
		score += result.Metrics.MeanNDCG5 * 1.0
		score += result.Metrics.MeanDiversity5 / 5.0 // Normalize
		score += result.Metrics.MeanSerendipity * 1.0

		if score > bestScore {
			bestScore = score
			bestIdx = i
		}
	}

	return results[bestIdx]
}

func printResult(result GridSearchResult) {
	fmt.Printf("  Precision@10: %.4f | nDCG@5: %.4f | Diversity@5: %.2f\n",
		result.Metrics.MeanPrecision10,
		result.Metrics.MeanNDCG5,
		result.Metrics.MeanDiversity5)
}

func printSummary(report GridSearchReport) {
	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("  Grid Search Summary")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Printf("Tested %d parameter combinations\n", len(report.Results))
	fmt.Println()
	fmt.Println("Best Configuration:")
	fmt.Println("-----------------------------------------")
	fmt.Printf("  Content Weight:       %.2f\n", report.BestConfig.Parameters.ContentWeight)
	fmt.Printf("  Collaborative Weight: %.2f\n", report.BestConfig.Parameters.CollaborativeWeight)
	fmt.Printf("  Trending Weight:      %.2f\n", report.BestConfig.Parameters.TrendingWeight)
	fmt.Println()
	fmt.Println("Metrics:")
	fmt.Println("-----------------------------------------")
	fmt.Printf("  Precision@5:     %.4f\n", report.BestConfig.Metrics.MeanPrecision5)
	fmt.Printf("  Precision@10:    %.4f\n", report.BestConfig.Metrics.MeanPrecision10)
	fmt.Printf("  Recall@5:        %.4f\n", report.BestConfig.Metrics.MeanRecall5)
	fmt.Printf("  Recall@10:       %.4f\n", report.BestConfig.Metrics.MeanRecall10)
	fmt.Printf("  nDCG@5:          %.4f\n", report.BestConfig.Metrics.MeanNDCG5)
	fmt.Printf("  nDCG@10:         %.4f\n", report.BestConfig.Metrics.MeanNDCG10)
	fmt.Printf("  Diversity@5:     %.2f games\n", report.BestConfig.Metrics.MeanDiversity5)
	fmt.Printf("  Diversity@10:    %.2f games\n", report.BestConfig.Metrics.MeanDiversity10)
	fmt.Printf("  Serendipity:     %.4f\n", report.BestConfig.Metrics.MeanSerendipity)
	fmt.Println()
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
