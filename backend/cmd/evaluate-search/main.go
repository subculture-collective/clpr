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
	datasetPath := flag.String("dataset", "testdata/search_evaluation_dataset.yaml", "Path to evaluation dataset YAML file")
	outputPath := flag.String("output", "", "Path to output JSON file (optional, defaults to stdout)")
	simulateMode := flag.Bool("simulate", true, "Use simulated results (no live search)")
	verbose := flag.Bool("verbose", false, "Print detailed results for each query")
	help := flag.Bool("help", false, "Show help message")

	flag.Parse()

	if *help {
		printUsage()
		os.Exit(0)
	}

	// Create evaluation service
	evalService := services.NewSearchEvaluationService(nil)

	// Load dataset
	log.Printf("Loading evaluation dataset from: %s", *datasetPath)
	if err := evalService.LoadDataset(*datasetPath); err != nil {
		log.Fatalf("Failed to load dataset: %v", err)
	}

	dataset := evalService.GetDataset()
	log.Printf("Loaded %d evaluation queries", len(dataset.EvaluationQueries))

	// Run evaluation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var report *services.EvaluationReport
	var err error

	if *simulateMode {
		log.Println("Running evaluation with simulated (ideal) results...")
		report, err = evalService.EvaluateWithSimulatedResults(ctx)
	} else {
		// Live mode would require database connection
		// For now, fall back to simulated mode
		log.Println("Live search mode not implemented, using simulated results...")
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
	fmt.Println("Search Evaluation Tool")
	fmt.Println()
	fmt.Println("Evaluates search quality using labeled queries and standard IR metrics (nDCG, MRR).")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  evaluate-search [options]")
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Run evaluation with default settings")
	fmt.Println("  evaluate-search")
	fmt.Println()
	fmt.Println("  # Run with custom dataset and save results")
	fmt.Println("  evaluate-search -dataset custom_dataset.yaml -output results.json")
	fmt.Println()
	fmt.Println("  # Run with verbose output")
	fmt.Println("  evaluate-search -verbose")
}

func printResults(report *services.EvaluationReport, verbose bool) {
	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("     Search Evaluation Results")
	fmt.Println("========================================")
	fmt.Println()

	// Aggregate metrics
	fmt.Println("Aggregate Metrics:")
	fmt.Println("-----------------------------------------")
	fmt.Printf("  nDCG@5:       %.4f\n", report.Metrics.MeanNDCG5)
	fmt.Printf("  nDCG@10:      %.4f\n", report.Metrics.MeanNDCG10)
	fmt.Printf("  MRR:          %.4f\n", report.Metrics.MeanMRR)
	fmt.Printf("  Precision@5:  %.4f\n", report.Metrics.MeanPrecision5)
	fmt.Printf("  Precision@10: %.4f\n", report.Metrics.MeanPrecision10)
	fmt.Printf("  Precision@20: %.4f\n", report.Metrics.MeanPrecision20)
	fmt.Printf("  Recall@5:     %.4f\n", report.Metrics.MeanRecall5)
	fmt.Printf("  Recall@10:    %.4f\n", report.Metrics.MeanRecall10)
	fmt.Printf("  Recall@20:    %.4f\n", report.Metrics.MeanRecall20)
	fmt.Printf("  Query Count:  %d\n", report.Metrics.QueryCount)
	fmt.Println()

	if verbose {
		fmt.Println("Per-Query Results:")
		fmt.Println("-----------------------------------------")
		for _, qr := range report.QueryResults {
			fmt.Printf("\n[%s] %s\n", qr.QueryID, qr.Query)
			fmt.Printf("  nDCG@5:  %.4f | nDCG@10: %.4f | MRR: %.4f\n",
				qr.NDCG5, qr.NDCG10, qr.MRR)
			fmt.Printf("  Prec@5:  %.4f | Prec@10: %.4f | Prec@20: %.4f\n",
				qr.Precision5, qr.Precision10, qr.Precision20)
			fmt.Printf("  Recall@5: %.4f | Recall@10: %.4f | Recall@20: %.4f\n",
				qr.Recall5, qr.Recall10, qr.Recall20)
			fmt.Printf("  Retrieved: %d | Relevant Found: %d\n",
				qr.RetrievedResults, qr.RelevantResults)
		}
		fmt.Println()
	}
}

func printTargetStatus(report *services.EvaluationReport) {
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
		{"nDCG@5", "ndcg_at_5", report.Metrics.MeanNDCG5, 0},
		{"nDCG@10", "ndcg_at_10", report.Metrics.MeanNDCG10, 0},
		{"MRR", "mrr", report.Metrics.MeanMRR, 0},
		{"Precision@5", "precision_at_5", report.Metrics.MeanPrecision5, 0},
		{"Precision@10", "precision_at_10", report.Metrics.MeanPrecision10, 0},
		{"Precision@20", "precision_at_20", report.Metrics.MeanPrecision20, 0},
		{"Recall@5", "recall_at_5", report.Metrics.MeanRecall5, 0},
		{"Recall@10", "recall_at_10", report.Metrics.MeanRecall10, 0},
		{"Recall@20", "recall_at_20", report.Metrics.MeanRecall20, 0},
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
		fmt.Printf("  %s %-14s %.4f%s\n", statusIcon, m.name+":", m.value, targetStr)
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
		return "❓"
	}
}

func hasFailures(report *services.EvaluationReport) bool {
	for _, status := range report.Status {
		if status == "critical" {
			return true
		}
	}
	return false
}

func writeOutputFile(path string, report *services.EvaluationReport) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
