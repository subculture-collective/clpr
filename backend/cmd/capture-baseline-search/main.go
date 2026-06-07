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

// BaselineReport contains baseline metrics and configuration
type BaselineReport struct {
	Timestamp     string                    `json:"timestamp"`
	Configuration BaselineConfiguration     `json:"configuration"`
	Metrics       services.AggregateMetrics `json:"metrics"`
	Status        map[string]string         `json:"status"`
	QueryCount    int                       `json:"query_count"`
	Notes         string                    `json:"notes"`
}

// BaselineConfiguration captures the current search configuration
type BaselineConfiguration struct {
	BM25Weight      float64 `json:"bm25_weight"`
	VectorWeight    float64 `json:"vector_weight"`
	TitleBoost      float64 `json:"title_boost"`
	CreatorBoost    float64 `json:"creator_boost"`
	GameBoost       float64 `json:"game_boost"`
	EngagementBoost float64 `json:"engagement_boost"`
	RecencyBoost    float64 `json:"recency_boost"`
	Description     string  `json:"description"`
}

func main() {
	// Command line flags
	datasetPath := flag.String("dataset", "testdata/search_evaluation_dataset.yaml", "Path to evaluation dataset YAML file")
	outputPath := flag.String("output", "baseline-search-metrics.json", "Path to output JSON file")
	notes := flag.String("notes", "Baseline capture for hybrid search weight optimization", "Notes about this baseline capture")
	help := flag.Bool("help", false, "Show help message")

	flag.Parse()

	if *help {
		printUsage()
		os.Exit(0)
	}

	log.Println("Capturing baseline search metrics...")
	log.Printf("Dataset: %s", *datasetPath)

	// Create evaluation service
	evalService := services.NewSearchEvaluationService(nil)

	// Load dataset
	if err := evalService.LoadDataset(*datasetPath); err != nil {
		log.Fatalf("Failed to load dataset: %v", err)
	}

	dataset := evalService.GetDataset()
	log.Printf("Loaded %d evaluation queries", len(dataset.EvaluationQueries))

	// Run evaluation with current/baseline configuration
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	report, err := evalService.EvaluateWithSimulatedResults(ctx)
	if err != nil {
		log.Fatalf("Evaluation failed: %v", err)
	}

	// Get current configuration from search_ab_testing.go DefaultConfigs
	// Using "baseline" configuration as reference
	currentConfig := BaselineConfiguration{
		BM25Weight:      0.7,
		VectorWeight:    0.3,
		TitleBoost:      3.0,
		CreatorBoost:    2.0,
		GameBoost:       1.0,
		EngagementBoost: 0.1,
		RecencyBoost:    0.5,
		Description:     "Current production configuration (baseline)",
	}

	// Create baseline report
	baselineReport := BaselineReport{
		Timestamp:     time.Now().Format(time.RFC3339),
		Configuration: currentConfig,
		Metrics:       report.Metrics,
		Status:        report.Status,
		QueryCount:    len(dataset.EvaluationQueries),
		Notes:         *notes,
	}

	// Print summary
	printSummary(&baselineReport)

	// Write output file
	if err := writeOutputFile(*outputPath, &baselineReport); err != nil {
		log.Fatalf("Failed to write output file: %v", err)
	}
	log.Printf("Baseline metrics saved to: %s", *outputPath)
	log.Println("Use this file with grid-search-hybrid-search -baseline flag for comparison")
}

func printUsage() {
	fmt.Println("Baseline Search Metrics Capture Tool")
	fmt.Println()
	fmt.Println("Captures current search performance metrics as a baseline for comparison.")
	fmt.Println("This baseline can be used with grid-search-hybrid-search to measure improvements.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  capture-baseline-search [options]")
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Capture baseline with default settings")
	fmt.Println("  capture-baseline-search")
	fmt.Println()
	fmt.Println("  # Capture baseline with custom notes")
	fmt.Println("  capture-baseline-search -notes \"Pre-optimization baseline\" -output baseline.json")
	fmt.Println()
	fmt.Println("  # Use baseline for comparison")
	fmt.Println("  grid-search-hybrid-search -baseline baseline.json")
}

func printSummary(report *BaselineReport) {
	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("  Baseline Metrics Captured")
	fmt.Println("========================================")
	fmt.Println()

	fmt.Println("Configuration:")
	fmt.Println("-----------------------------------------")
	fmt.Printf("  BM25 Weight:      %.2f\n", report.Configuration.BM25Weight)
	fmt.Printf("  Vector Weight:    %.2f\n", report.Configuration.VectorWeight)
	fmt.Printf("  Title Boost:      %.1f\n", report.Configuration.TitleBoost)
	fmt.Printf("  Creator Boost:    %.1f\n", report.Configuration.CreatorBoost)
	fmt.Printf("  Game Boost:       %.1f\n", report.Configuration.GameBoost)
	fmt.Printf("  Engagement Boost: %.2f\n", report.Configuration.EngagementBoost)
	fmt.Printf("  Recency Boost:    %.2f\n", report.Configuration.RecencyBoost)
	fmt.Println()

	fmt.Println("Baseline Metrics:")
	fmt.Println("-----------------------------------------")
	fmt.Printf("  nDCG@5:       %.4f\n", report.Metrics.MeanNDCG5)
	fmt.Printf("  nDCG@10:      %.4f (primary metric)\n", report.Metrics.MeanNDCG10)
	fmt.Printf("  MRR:          %.4f\n", report.Metrics.MeanMRR)
	fmt.Printf("  Precision@5:  %.4f\n", report.Metrics.MeanPrecision5)
	fmt.Printf("  Precision@10: %.4f\n", report.Metrics.MeanPrecision10)
	fmt.Printf("  Precision@20: %.4f\n", report.Metrics.MeanPrecision20)
	fmt.Printf("  Recall@5:     %.4f\n", report.Metrics.MeanRecall5)
	fmt.Printf("  Recall@10:    %.4f\n", report.Metrics.MeanRecall10)
	fmt.Printf("  Recall@20:    %.4f\n", report.Metrics.MeanRecall20)
	fmt.Printf("  Query Count:  %d\n", report.QueryCount)
	fmt.Println()

	// Show target comparison
	if len(report.Status) > 0 {
		fmt.Println("Status vs Targets:")
		fmt.Println("-----------------------------------------")
		metrics := []struct {
			name   string
			key    string
			status string
		}{
			{"nDCG@5", "ndcg_at_5", report.Status["ndcg_at_5"]},
			{"nDCG@10", "ndcg_at_10", report.Status["ndcg_at_10"]},
			{"MRR", "mrr", report.Status["mrr"]},
			{"Precision@10", "precision_at_10", report.Status["precision_at_10"]},
		}

		for _, m := range metrics {
			if m.status != "" {
				icon := getStatusIcon(m.status)
				fmt.Printf("  %s %s: %s\n", icon, m.name, m.status)
			}
		}
		fmt.Println()
	}

	fmt.Println("Notes:")
	fmt.Println("-----------------------------------------")
	fmt.Println("  " + report.Notes)
	fmt.Println()

	// Improvement target reminder
	fmt.Println("Target for Optimization:")
	fmt.Println("-----------------------------------------")
	fmt.Println("  nDCG@10 improvement: ≥10%")
	fmt.Printf("  Target nDCG@10:      ≥%.4f\n", report.Metrics.MeanNDCG10*1.1)
	fmt.Println()
}

func getStatusIcon(status string) string {
	switch status {
	case "pass":
		return "✅"
	case "warning":
		return "⚠️"
	case "critical":
		return "❌"
	default:
		return "❓"
	}
}

func writeOutputFile(path string, report *BaselineReport) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
