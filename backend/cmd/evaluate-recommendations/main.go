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

func main() {
	// Command line flags
	datasetPath := flag.String("dataset", "testdata/recommendation_evaluation_dataset.yaml", "Path to evaluation dataset YAML file")
	outputPath := flag.String("output", "", "Path to output JSON file (optional, defaults to stdout)")
	simulateMode := flag.Bool("simulate", true, "Use simulated results (no live recommendations)")
	verbose := flag.Bool("verbose", false, "Print detailed results for each scenario")
	help := flag.Bool("help", false, "Show help message")

	flag.Parse()

	if *help {
		printUsage()
		os.Exit(0)
	}

	// Create evaluation service
	evalService := services.NewRecommendationEvaluationService(nil)

	// Load dataset
	log.Printf("Loading evaluation dataset from: %s", *datasetPath)
	if err := evalService.LoadDataset(*datasetPath); err != nil {
		log.Fatalf("Failed to load dataset: %v", err)
	}

	dataset := evalService.GetDataset()
	log.Printf("Loaded %d evaluation scenarios", len(dataset.EvaluationScenarios))

	// Run evaluation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var report *services.RecommendationEvaluationReport
	var err error

	if *simulateMode {
		log.Println("Running evaluation with simulated (ideal) results...")
		report, err = evalService.EvaluateWithSimulatedResults(ctx)
	} else {
		// Live mode would require recommendation service connection
		// For now, fall back to simulated mode
		log.Println("Live recommendation mode not implemented, using simulated results...")
		report, err = evalService.EvaluateWithSimulatedResults(ctx)
	}

	if err != nil {
		log.Fatalf("Evaluation failed: %v", err)
	}

	// Print results
	printResults(report, *verbose)

	// Check against targets and print status
	printTargetStatus(report)

	// Write output file if specified
	if *outputPath != "" {
		if err := writeOutputFile(*outputPath, report); err != nil {
			log.Fatalf("Failed to write output file: %v", err)
		}
		log.Printf("Results written to: %s", *outputPath)
	}

	// Exit with non-zero code if critical thresholds are breached
	if hasFailures(report) {
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Recommendation Evaluation Tool")
	fmt.Println()
	fmt.Println("Evaluates recommendation quality using labeled scenarios and standard metrics:")
	fmt.Println("- Precision@k and Recall@k for accuracy")
	fmt.Println("- nDCG for ranking quality")
	fmt.Println("- Diversity for content variety")
	fmt.Println("- Serendipity for unexpected discoveries")
	fmt.Println("- Cold-start performance for new users")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  evaluate-recommendations [options]")
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Run evaluation with default settings")
	fmt.Println("  evaluate-recommendations")
	fmt.Println()
	fmt.Println("  # Run with custom dataset and save results")
	fmt.Println("  evaluate-recommendations -dataset custom_dataset.yaml -output results.json")
	fmt.Println()
	fmt.Println("  # Run with verbose output")
	fmt.Println("  evaluate-recommendations -verbose")
}

func printResults(report *services.RecommendationEvaluationReport, verbose bool) {
	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("  Recommendation Evaluation Results")
	fmt.Println("========================================")
	fmt.Println()

	// Aggregate metrics
	fmt.Println("Aggregate Metrics:")
	fmt.Println("-----------------------------------------")
	fmt.Printf("  Precision@5:     %.4f\n", report.Metrics.MeanPrecision5)
	fmt.Printf("  Precision@10:    %.4f\n", report.Metrics.MeanPrecision10)
	fmt.Printf("  Recall@5:        %.4f\n", report.Metrics.MeanRecall5)
	fmt.Printf("  Recall@10:       %.4f\n", report.Metrics.MeanRecall10)
	fmt.Printf("  nDCG@5:          %.4f\n", report.Metrics.MeanNDCG5)
	fmt.Printf("  nDCG@10:         %.4f\n", report.Metrics.MeanNDCG10)
	fmt.Printf("  Diversity@5:     %.2f games\n", report.Metrics.MeanDiversity5)
	fmt.Printf("  Diversity@10:    %.2f games\n", report.Metrics.MeanDiversity10)
	fmt.Printf("  Serendipity:     %.4f\n", report.Metrics.MeanSerendipity)
	fmt.Printf("  Coverage:        %.4f\n", report.Metrics.MeanCoverage)
	fmt.Printf("  Scenario Count:  %d\n", report.Metrics.ScenarioCount)
	fmt.Println()

	// Cold start metrics
	if report.Metrics.ColdStartCount > 0 {
		fmt.Println("Cold-Start Performance:")
		fmt.Println("-----------------------------------------")
		fmt.Printf("  Precision@5:     %.4f\n", report.Metrics.ColdStartPrecision5)
		fmt.Printf("  Recall@5:        %.4f\n", report.Metrics.ColdStartRecall5)
		fmt.Printf("  Cold-Start Count: %d\n", report.Metrics.ColdStartCount)
		fmt.Println()
	}

	if verbose {
		fmt.Println("Per-Scenario Results:")
		fmt.Println("-----------------------------------------")
		for _, sr := range report.ScenarioResults {
			coldStartMarker := ""
			if sr.IsColdStart {
				coldStartMarker = " [COLD START]"
			}
			fmt.Printf("\n[%s] %s%s\n", sr.ScenarioID, sr.Description, coldStartMarker)
			fmt.Printf("  Algorithm: %s\n", sr.Algorithm)
			fmt.Printf("  Precision@5:  %.4f | Precision@10: %.4f\n", sr.Precision5, sr.Precision10)
			fmt.Printf("  Recall@5:     %.4f | Recall@10:    %.4f\n", sr.Recall5, sr.Recall10)
			fmt.Printf("  nDCG@5:       %.4f | nDCG@10:      %.4f\n", sr.NDCG5, sr.NDCG10)
			fmt.Printf("  Diversity@5:  %.1f   | Diversity@10: %.1f\n", sr.Diversity5, sr.Diversity10)
			fmt.Printf("  Serendipity:  %.4f | Coverage:     %.4f\n", sr.SerendipityScore, sr.Coverage)
			fmt.Printf("  Retrieved: %d | Relevant Found: %d\n", sr.RetrievedCount, sr.RelevantCount)
		}
		fmt.Println()
	}
}

func printTargetStatus(report *services.RecommendationEvaluationReport) {
	if report.Status == nil || len(report.Status) == 0 {
		return
	}

	fmt.Println("Target Comparison:")
	fmt.Println("-----------------------------------------")

	metrics := []struct {
		name   string
		key    string
		value  float64
		target float64
	}{
		{"Precision@5", "precision_at_5", report.Metrics.MeanPrecision5, 0},
		{"Precision@10", "precision_at_10", report.Metrics.MeanPrecision10, 0},
		{"Recall@5", "recall_at_5", report.Metrics.MeanRecall5, 0},
		{"Recall@10", "recall_at_10", report.Metrics.MeanRecall10, 0},
		{"nDCG@5", "ndcg_at_5", report.Metrics.MeanNDCG5, 0},
		{"nDCG@10", "ndcg_at_10", report.Metrics.MeanNDCG10, 0},
		{"Diversity@5", "diversity_at_5", report.Metrics.MeanDiversity5, 0},
		{"Diversity@10", "diversity_at_10", report.Metrics.MeanDiversity10, 0},
		{"Serendipity", "serendipity", report.Metrics.MeanSerendipity, 0},
		{"Cold Start P@5", "cold_start_precision_5", report.Metrics.ColdStartPrecision5, 0},
	}

	for i, m := range metrics {
		if target, ok := report.Targets[m.key]; ok {
			metrics[i].target = target.Target
		}
	}

	for _, m := range metrics {
		status := report.Status[m.key]
		statusIcon := getStatusIcon(status)
		targetStr := ""
		if m.target > 0 {
			targetStr = fmt.Sprintf(" (target: %.2f)", m.target)
		}
		// Skip if no status (metric not evaluated)
		if status == "" {
			continue
		}
		fmt.Printf("  %s %-18s %.4f%s\n", statusIcon, m.name+":", m.value, targetStr)
	}
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
		return "  "
	}
}

func hasFailures(report *services.RecommendationEvaluationReport) bool {
	for _, status := range report.Status {
		if status == "critical" {
			return true
		}
	}
	return false
}

func writeOutputFile(path string, report *services.RecommendationEvaluationReport) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
