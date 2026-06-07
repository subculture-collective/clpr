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
	configAName := flag.String("config-a", "baseline", "Name of configuration A")
	configBName := flag.String("config-b", "semantic-heavy", "Name of configuration B")
	listConfigs := flag.Bool("list-configs", false, "List available configurations and exit")
	help := flag.Bool("help", false, "Show help message")

	flag.Parse()

	if *help {
		printUsage()
		os.Exit(0)
	}

	// List configurations if requested
	if *listConfigs {
		printConfigurations()
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

	// Create A/B test harness
	abHarness := services.NewABTestHarness(evalService)

	// Get configurations
	configs := services.DefaultConfigs()
	configMap := make(map[string]services.SearchWeightConfig)
	for _, c := range configs {
		configMap[c.Name] = c
	}

	configA, okA := configMap[*configAName]
	configB, okB := configMap[*configBName]

	if !okA {
		log.Fatalf("Configuration '%s' not found. Use -list-configs to see available configurations.", *configAName)
	}
	if !okB {
		log.Fatalf("Configuration '%s' not found. Use -list-configs to see available configurations.", *configBName)
	}

	log.Printf("Comparing configurations: %s vs %s", configA.Name, configB.Name)

	// Run A/B test
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	result, err := abHarness.EvaluateWithSimulated(ctx, configA, configB)
	if err != nil {
		log.Fatalf("A/B test failed: %v", err)
	}

	// Print results
	printResults(result)

	// Write output file if specified
	if *outputPath != "" {
		if err := writeOutputFile(*outputPath, result); err != nil {
			log.Fatalf("Failed to write output file: %v", err)
		}
		log.Printf("Results written to: %s", *outputPath)
	}
}

func printUsage() {
	fmt.Println("Search A/B Testing Tool")
	fmt.Println()
	fmt.Println("Compares two search configurations using the evaluation dataset.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  search-ab-test [options]")
	fmt.Println()
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # List available configurations")
	fmt.Println("  search-ab-test -list-configs")
	fmt.Println()
	fmt.Println("  # Compare baseline vs semantic-heavy")
	fmt.Println("  search-ab-test -config-a baseline -config-b semantic-heavy")
	fmt.Println()
	fmt.Println("  # Compare and save results")
	fmt.Println("  search-ab-test -config-a baseline -config-b engagement-focused -output results.json")
}

func printConfigurations() {
	fmt.Println("Available Configurations:")
	fmt.Println("========================")
	fmt.Println()

	configs := services.DefaultConfigs()
	for _, c := range configs {
		fmt.Printf("Name: %s\n", c.Name)
		fmt.Printf("  Description: %s\n", c.Description)
		fmt.Printf("  BM25 Weight: %.2f | Vector Weight: %.2f\n", c.BM25Weight, c.VectorWeight)
		fmt.Printf("  Title Boost: %.1f | Creator Boost: %.1f | Game Boost: %.1f\n",
			c.TitleBoost, c.CreatorBoost, c.GameBoost)
		fmt.Printf("  Engagement Boost: %.2f | Recency Boost: %.2f\n",
			c.EngagementBoost, c.RecencyBoost)
		fmt.Println()
	}
}

func printResults(result *services.ABTestResult) {
	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("     A/B Test Results")
	fmt.Println("========================================")
	fmt.Println()

	// Configuration details
	fmt.Println("Configuration A:", result.ConfigA.Name)
	fmt.Println("  ", result.ConfigA.Description)
	fmt.Println()
	fmt.Println("Configuration B:", result.ConfigB.Name)
	fmt.Println("  ", result.ConfigB.Description)
	fmt.Println()

	// Metrics comparison
	fmt.Println("Metrics Comparison:")
	fmt.Println("-----------------------------------------")
	fmt.Printf("%-15s | %-12s | %-12s | %-10s\n", "Metric", "Config A", "Config B", "Change")
	fmt.Println("-----------------------------------------")

	metrics := []struct {
		name string
		key  string
		valA float64
		valB float64
	}{
		{"nDCG@5", "ndcg_at_5", result.MetricsA.MeanNDCG5, result.MetricsB.MeanNDCG5},
		{"nDCG@10", "ndcg_at_10", result.MetricsA.MeanNDCG10, result.MetricsB.MeanNDCG10},
		{"MRR", "mrr", result.MetricsA.MeanMRR, result.MetricsB.MeanMRR},
		{"Precision@5", "precision_at_5", result.MetricsA.MeanPrecision5, result.MetricsB.MeanPrecision5},
		{"Precision@10", "precision_at_10", result.MetricsA.MeanPrecision10, result.MetricsB.MeanPrecision10},
		{"Precision@20", "precision_at_20", result.MetricsA.MeanPrecision20, result.MetricsB.MeanPrecision20},
		{"Recall@5", "recall_at_5", result.MetricsA.MeanRecall5, result.MetricsB.MeanRecall5},
		{"Recall@10", "recall_at_10", result.MetricsA.MeanRecall10, result.MetricsB.MeanRecall10},
		{"Recall@20", "recall_at_20", result.MetricsA.MeanRecall20, result.MetricsB.MeanRecall20},
	}

	for _, m := range metrics {
		change := result.Improvements[m.key]
		changeIcon := getChangeIcon(change)
		fmt.Printf("%-15s | %12.4f | %12.4f | %s %+7.2f%%\n",
			m.name, m.valA, m.valB, changeIcon, change)
	}
	fmt.Println()

	// Recommendation
	fmt.Println("Recommendation:")
	fmt.Println("-----------------------------------------")
	fmt.Println(result.Recommendation)
	fmt.Println()

	// Statistical summary
	fmt.Println("Statistical Summary:")
	fmt.Println("-----------------------------------------")
	fmt.Println(result.StatSummary)
	fmt.Println()
}

func getChangeIcon(change float64) string {
	if change > 5.0 {
		return "⬆️"
	} else if change > 0 {
		return "↗️"
	} else if change < -5.0 {
		return "⬇️"
	} else if change < 0 {
		return "↘️"
	}
	return "➡️"
}

func writeOutputFile(path string, result *services.ABTestResult) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
