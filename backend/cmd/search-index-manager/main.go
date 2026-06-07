package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/pkg/database"
	opensearchpkg "git.subcult.tv/subculture-collective/clpr/pkg/opensearch"
)

const usage = `Search Index Manager - Manage versioned search indices

Usage:
    search-index-manager <command> [options]

Commands:
    status      Show status of all index versions
    rebuild     Rebuild search index with zero-downtime swap
    swap        Swap alias to a specific index version
    rollback    Rollback alias to a previous version
    cleanup     Delete old index versions

Options:
    -index string     Index name (clips, users, tags, games, or 'all')
    -version int      Target version for swap/rollback (default: latest)
    -batch int        Batch size for rebuild (default: 100)
    -keep int         Number of old versions to keep (default: 2)
    -no-swap          Skip alias swap after rebuild
    -dry-run          Show what would be done without making changes
    -json             Output results as JSON
    -help             Show this help message

Examples:
    # Show status of all indices
    search-index-manager status

    # Show status of clips index as JSON
    search-index-manager status -index clips -json

    # Rebuild clips index with zero-downtime swap
    search-index-manager rebuild -index clips

    # Rebuild all indices
    search-index-manager rebuild -index all

    # Rebuild without swapping alias (for testing)
    search-index-manager rebuild -index clips -no-swap

    # Swap to a specific version
    search-index-manager swap -index clips -version 3

    # Rollback to previous version
    search-index-manager rollback -index clips -version 2

    # Clean up old versions, keeping 2 most recent
    search-index-manager cleanup -index clips -keep 2
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}

	command := os.Args[1]
	if command == "-help" || command == "--help" || command == "help" {
		fmt.Print(usage)
		os.Exit(0)
	}

	// Set up flags
	flagSet := flag.NewFlagSet(command, flag.ExitOnError)
	indexName := flagSet.String("index", "", "Index name (clips, users, tags, games, or 'all')")
	version := flagSet.Int("version", 0, "Target version for swap/rollback")
	batchSize := flagSet.Int("batch", 100, "Batch size for rebuild")
	keepVersions := flagSet.Int("keep", 2, "Number of old versions to keep")
	noSwap := flagSet.Bool("no-swap", false, "Skip alias swap after rebuild")
	dryRun := flagSet.Bool("dry-run", false, "Show what would be done")
	jsonOutput := flagSet.Bool("json", false, "Output as JSON")

	// Parse flags
	if len(os.Args) > 2 {
		if err := flagSet.Parse(os.Args[2:]); err != nil {
			log.Fatalf("Failed to parse flags: %v", err)
		}
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	db, err := database.NewDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize OpenSearch client
	osClient, err := opensearchpkg.NewClient(&opensearchpkg.Config{
		URL:                cfg.OpenSearch.URL,
		Username:           cfg.OpenSearch.Username,
		Password:           cfg.OpenSearch.Password,
		InsecureSkipVerify: cfg.OpenSearch.InsecureSkipVerify,
	})
	if err != nil {
		log.Fatalf("Failed to initialize OpenSearch client: %v", err)
	}

	ctx := context.Background()
	if err := osClient.Ping(ctx); err != nil {
		log.Fatalf("OpenSearch ping failed: %v", err)
	}

	// Initialize services
	rebuildService := services.NewIndexRebuildService(db, osClient)
	versionService := rebuildService.GetVersionService()

	// Execute command
	switch command {
	case "status":
		executeStatus(ctx, versionService, *indexName, *jsonOutput)
	case "rebuild":
		executeRebuild(ctx, rebuildService, *indexName, *batchSize, *keepVersions, !*noSwap, *dryRun, *jsonOutput)
	case "swap":
		executeSwap(ctx, versionService, *indexName, *version, *dryRun, *jsonOutput)
	case "rollback":
		executeRollback(ctx, versionService, *indexName, *version, *dryRun, *jsonOutput)
	case "cleanup":
		executeCleanup(ctx, versionService, *indexName, *keepVersions, *dryRun, *jsonOutput)
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		fmt.Print(usage)
		os.Exit(1)
	}
}

func getAllIndices() []string {
	return []string{services.ClipsIndex, services.UsersIndex, services.TagsIndex, services.GamesIndex}
}

func getTargetIndices(indexName string) []string {
	if indexName == "" || indexName == "all" {
		return getAllIndices()
	}

	// Validate index name
	validIndices := map[string]bool{
		services.ClipsIndex: true,
		services.UsersIndex: true,
		services.TagsIndex:  true,
		services.GamesIndex: true,
	}

	if !validIndices[indexName] {
		log.Fatalf("Invalid index name: %s. Valid options: clips, users, tags, games, all", indexName)
	}

	return []string{indexName}
}

func executeStatus(ctx context.Context, versionService *services.IndexVersionService, indexName string, jsonOutput bool) {
	indices := getTargetIndices(indexName)
	allInfo := make(map[string]*services.IndexVersionInfo)

	for _, idx := range indices {
		info, err := versionService.GetIndexVersionInfo(ctx, idx)
		if err != nil {
			log.Printf("WARNING: Failed to get info for %s: %v", idx, err)
			continue
		}
		allInfo[idx] = info
	}

	if jsonOutput {
		output, err := json.MarshalIndent(allInfo, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal results to JSON: %v", err)
		}
		fmt.Println(string(output))
		return
	}

	// Print formatted status
	fmt.Println("=== Search Index Versions ===")
	fmt.Println()

	for _, idx := range indices {
		info, ok := allInfo[idx]
		if !ok {
			continue
		}

		fmt.Printf("Index: %s\n", idx)
		fmt.Printf("  Total Versions: %d\n", info.TotalVersions)
		fmt.Printf("  Latest Version: v%d\n", info.LatestVersion)

		if info.ActiveVersion != nil {
			fmt.Printf("  Active Version: v%d (%s)\n", info.ActiveVersion.Version, info.ActiveVersion.Name)
			fmt.Printf("    Documents: %d\n", info.ActiveVersion.DocCount)
		} else {
			fmt.Println("  Active Version: none (alias not set)")
		}

		if len(info.AllVersions) > 0 {
			fmt.Println("  All Versions:")
			for _, v := range info.AllVersions {
				activeMarker := ""
				if v.IsActive {
					activeMarker = " [ACTIVE]"
				}
				fmt.Printf("    - %s (v%d): %d docs%s\n", v.Name, v.Version, v.DocCount, activeMarker)
			}
		}

		fmt.Println()
	}
}

func executeRebuild(ctx context.Context, rebuildService *services.IndexRebuildService, indexName string, batchSize, keepVersions int, swapAlias, dryRun, jsonOutput bool) {
	if dryRun {
		fmt.Println("DRY RUN: Would rebuild the following indices:")
		indices := getTargetIndices(indexName)
		for _, idx := range indices {
			fmt.Printf("  - %s (batch size: %d, swap: %v, keep versions: %d)\n", idx, batchSize, swapAlias, keepVersions)
		}
		return
	}

	config := &services.RebuildConfig{
		BatchSize:       batchSize,
		KeepOldVersions: keepVersions,
		SwapAfterBuild:  swapAlias,
		Verbose:         !jsonOutput,
	}

	var results interface{}

	if indexName == "" || indexName == "all" {
		result, err := rebuildService.RebuildAllIndices(ctx, config)
		if err != nil {
			log.Fatalf("Rebuild failed: %v", err)
		}
		results = result
	} else {
		var result *services.RebuildResult
		var err error

		switch indexName {
		case services.ClipsIndex:
			result, err = rebuildService.RebuildClipsIndex(ctx, config)
		case services.UsersIndex:
			result, err = rebuildService.RebuildUsersIndex(ctx, config)
		case services.TagsIndex:
			result, err = rebuildService.RebuildTagsIndex(ctx, config)
		case services.GamesIndex:
			result, err = rebuildService.RebuildGamesIndex(ctx, config)
		default:
			log.Fatalf("Invalid index name: %s", indexName)
		}

		if err != nil {
			log.Fatalf("Rebuild failed: %v", err)
		}
		results = result
	}

	if jsonOutput {
		output, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal results to JSON: %v", err)
		}
		fmt.Println(string(output))
	} else {
		fmt.Println("Rebuild completed successfully!")
	}
}

func executeSwap(ctx context.Context, versionService *services.IndexVersionService, indexName string, version int, dryRun, jsonOutput bool) {
	if indexName == "" || indexName == "all" {
		log.Fatal("Index name is required for swap operation (cannot swap 'all')")
	}

	indices := getTargetIndices(indexName)
	idx := indices[0]

	// If version is 0, get latest version
	if version == 0 {
		info, err := versionService.GetIndexVersionInfo(ctx, idx)
		if err != nil {
			log.Fatalf("Failed to get index info: %v", err)
		}
		if info.LatestVersion == 0 {
			log.Fatal("No versioned indices exist for swap operation")
		}
		version = info.LatestVersion
	}

	if dryRun {
		fmt.Printf("DRY RUN: Would swap alias %s to version v%d\n", idx, version)
		return
	}

	if err := versionService.SwapAlias(ctx, idx, version); err != nil {
		log.Fatalf("Swap failed: %v", err)
	}

	result := map[string]interface{}{
		"index":       idx,
		"new_version": version,
		"success":     true,
	}

	if jsonOutput {
		output, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal results to JSON: %v", err)
		}
		fmt.Println(string(output))
	} else {
		fmt.Printf("Successfully swapped %s alias to version v%d\n", idx, version)
	}
}

func executeRollback(ctx context.Context, versionService *services.IndexVersionService, indexName string, version int, dryRun, jsonOutput bool) {
	if indexName == "" || indexName == "all" {
		log.Fatal("Index name is required for rollback operation")
	}

	indices := getTargetIndices(indexName)
	idx := indices[0]

	// Get current info
	info, err := versionService.GetIndexVersionInfo(ctx, idx)
	if err != nil {
		log.Fatalf("Failed to get index info: %v", err)
	}

	// If version is 0, rollback to previous version
	if version == 0 {
		// Find the second highest version (previous)
		if len(info.AllVersions) < 2 {
			log.Fatal("No previous version available for rollback")
		}

		// AllVersions is sorted descending, so [1] is the previous version
		version = info.AllVersions[1].Version
	}

	// Validate target version exists
	found := false
	for _, v := range info.AllVersions {
		if v.Version == version {
			found = true
			break
		}
	}

	if !found {
		log.Fatalf("Version v%d does not exist for index %s", version, idx)
	}

	if dryRun {
		currentVersion := 0
		if info.ActiveVersion != nil {
			currentVersion = info.ActiveVersion.Version
		}
		fmt.Printf("DRY RUN: Would rollback %s from v%d to v%d\n", idx, currentVersion, version)
		return
	}

	if err := versionService.RollbackAlias(ctx, idx, version); err != nil {
		log.Fatalf("Rollback failed: %v", err)
	}

	result := map[string]interface{}{
		"index":       idx,
		"new_version": version,
		"success":     true,
	}

	if jsonOutput {
		output, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal results to JSON: %v", err)
		}
		fmt.Println(string(output))
	} else {
		fmt.Printf("Successfully rolled back %s alias to version v%d\n", idx, version)
	}
}

func executeCleanup(ctx context.Context, versionService *services.IndexVersionService, indexName string, keepVersions int, dryRun, jsonOutput bool) {
	indices := getTargetIndices(indexName)
	allDeleted := make(map[string][]string)

	for _, idx := range indices {
		info, err := versionService.GetIndexVersionInfo(ctx, idx)
		if err != nil {
			log.Printf("WARNING: Failed to get info for %s: %v", idx, err)
			continue
		}

		// Calculate which versions would be deleted
		toDelete := []string{}
		for i, v := range info.AllVersions {
			if i < keepVersions || v.IsActive {
				continue
			}
			toDelete = append(toDelete, v.Name)
		}

		if dryRun {
			fmt.Printf("DRY RUN: Would delete %d old versions from %s:\n", len(toDelete), idx)
			for _, name := range toDelete {
				fmt.Printf("  - %s\n", name)
			}
			allDeleted[idx] = toDelete
			continue
		}

		deleted, err := versionService.DeleteOldVersions(ctx, idx, keepVersions)
		if err != nil {
			log.Printf("WARNING: Failed to cleanup %s: %v", idx, err)
			continue
		}
		allDeleted[idx] = deleted
	}

	if dryRun {
		return
	}

	if jsonOutput {
		output, err := json.MarshalIndent(allDeleted, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal results to JSON: %v", err)
		}
		fmt.Println(string(output))
	} else {
		fmt.Println("Cleanup completed:")
		for idx, deleted := range allDeleted {
			fmt.Printf("  %s: deleted %d versions: %s\n", idx, len(deleted), strings.Join(deleted, ", "))
		}
	}
}
