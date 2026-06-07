package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/subculture-collective/clipper/config"
	"github.com/subculture-collective/clipper/pkg/database"
	"github.com/subculture-collective/clipper/pkg/redis"
	"github.com/subculture-collective/clipper/pkg/twitch"
)

// ScraperStats tracks scraping statistics
type ScraperStats struct {
	StartTime           time.Time // Time when scraping started
	EndTime             time.Time // Time when scraping ended
	BroadcastersFetched int       // Number of broadcasters found in database
	BroadcastersScraped int       // Number of broadcasters successfully processed
	ClipsChecked        int       // Total number of clips examined
	ClipsInserted       int       // Number of new clips inserted into database
	ClipsSkipped        int       // Number of clips skipped (duplicates or filtered out)
	Errors              int       // Number of errors encountered
	LastError           error     // Most recent error (if any)
	APICallsMade        int       // Number of Twitch API calls made
}

// ScraperConfig holds scraping configuration
type ScraperConfig struct {
	DryRun       bool
	BatchSize    int
	MaxAgeDays   int
	MinViews     int
	Broadcasters []string
	LookbackDays int
}

func main() {
	// Parse command-line flags
	dryRun := flag.Bool("dry-run", false, "Dry run mode - don't insert clips")
	batchSize := flag.Int("batch-size", 50, "Number of clips to fetch per broadcaster")
	maxAgeDays := flag.Int("max-age-days", 30, "Maximum age of clips to scrape (in days)")
	minViews := flag.Int("min-views", 100, "Minimum view count for clips")
	broadcastersFlag := flag.String("broadcasters", "", "Comma-separated list of broadcaster names to scrape (overrides database query)")
	lookbackDays := flag.Int("lookback-days", 30, "Number of days to look back for submissions")
	flag.Parse()

	log.Println("=== clpr Targeted Scraper ===")
	log.Printf("Configuration:")
	log.Printf("  Dry Run: %t", *dryRun)
	log.Printf("  Batch Size: %d clips per broadcaster", *batchSize)
	log.Printf("  Max Age: %d days", *maxAgeDays)
	log.Printf("  Min Views: %d", *minViews)
	log.Printf("  Lookback: %d days", *lookbackDays)
	if *broadcastersFlag != "" {
		log.Printf("  Manual Broadcasters: %s", *broadcastersFlag)
	}
	log.Println()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate Twitch credentials
	if cfg.Twitch.ClientID == "" || cfg.Twitch.ClientSecret == "" {
		log.Fatalf("TWITCH_CLIENT_ID and TWITCH_CLIENT_SECRET must be set")
	}

	// Initialize database connection
	db, err := database.NewDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("✓ Database connection established")

	// Initialize Redis client
	redisClient, err := redis.NewClient(&cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()
	log.Println("✓ Redis connection established")

	// Initialize Twitch client
	twitchClient, err := twitch.NewClient(&cfg.Twitch, redisClient)
	if err != nil {
		log.Fatalf("Failed to initialize Twitch client: %v", err)
	}
	log.Println("✓ Twitch API client initialized")
	log.Println()

	// Build scraper config
	scraperConfig := ScraperConfig{
		DryRun:       *dryRun,
		BatchSize:    *batchSize,
		MaxAgeDays:   *maxAgeDays,
		MinViews:     *minViews,
		LookbackDays: *lookbackDays,
	}

	// Parse manual broadcaster list if provided
	if *broadcastersFlag != "" {
		scraperConfig.Broadcasters = strings.Split(*broadcastersFlag, ",")
		for i := range scraperConfig.Broadcasters {
			scraperConfig.Broadcasters[i] = strings.TrimSpace(scraperConfig.Broadcasters[i])
		}
	}

	// Run scraper
	ctx := context.Background()
	stats, err := runScraper(ctx, db.Pool, twitchClient, scraperConfig)
	if err != nil {
		log.Fatalf("Scraper failed: %v", err)
	}

	// Print summary
	printSummary(stats)
}

// runScraper executes the main scraping logic
func runScraper(ctx context.Context, db *pgxpool.Pool, twitchClient *twitch.Client, cfg ScraperConfig) (*ScraperStats, error) {
	stats := &ScraperStats{
		StartTime: time.Now(),
	}

	// Step 1: Get list of broadcasters to scrape
	var broadcasters []string
	var err error

	if len(cfg.Broadcasters) > 0 {
		// Use manually specified broadcasters
		broadcasters = cfg.Broadcasters
		log.Printf("Using %d manually specified broadcasters", len(broadcasters))
	} else {
		// Query database for broadcasters with submissions
		broadcasters, err = getBroadcastersFromSubmissions(ctx, db, cfg.LookbackDays)
		if err != nil {
			return stats, fmt.Errorf("failed to get broadcasters: %w", err)
		}
		log.Printf("Found %d broadcasters with submissions in the last %d days", len(broadcasters), cfg.LookbackDays)
	}

	stats.BroadcastersFetched = len(broadcasters)

	if len(broadcasters) == 0 {
		log.Println("No broadcasters to scrape")
		stats.EndTime = time.Now()
		return stats, nil
	}

	log.Println()
	log.Println("Starting clip scraping...")

	// Step 2: Scrape clips for each broadcaster
	for i, broadcasterName := range broadcasters {
		log.Printf("[%d/%d] Processing broadcaster: %s", i+1, len(broadcasters), broadcasterName)

		// Get broadcaster ID from Twitch API
		usersResp, err := twitchClient.GetUsers(ctx, nil, []string{broadcasterName})
		stats.APICallsMade++
		if err != nil {
			log.Printf("  ✗ Failed to get broadcaster ID: %v", err)
			stats.Errors++
			stats.LastError = err
			continue
		}

		if len(usersResp.Data) == 0 {
			log.Printf("  ✗ Broadcaster not found on Twitch")
			stats.Errors++
			continue
		}

		broadcasterID := usersResp.Data[0].ID
		log.Printf("  Broadcaster ID: %s", broadcasterID)

		// Fetch clips for this broadcaster
		clipsAdded, clipsSkipped, apiCalls, err := scrapeClipsForBroadcaster(
			ctx,
			db,
			twitchClient,
			broadcasterID,
			broadcasterName,
			cfg,
			stats,
		)

		stats.APICallsMade += apiCalls
		stats.BroadcastersScraped++

		if err != nil {
			log.Printf("  ✗ Error scraping clips: %v", err)
			stats.Errors++
			stats.LastError = err
			continue
		}

		stats.ClipsInserted += clipsAdded
		stats.ClipsSkipped += clipsSkipped
		stats.ClipsChecked += clipsAdded + clipsSkipped

		log.Printf("  ✓ Processed: %d clips added, %d skipped", clipsAdded, clipsSkipped)
	}

	stats.EndTime = time.Now()
	return stats, nil
}

// getBroadcastersFromSubmissions queries the database for broadcasters with recent submissions
func getBroadcastersFromSubmissions(ctx context.Context, db *pgxpool.Pool, lookbackDays int) ([]string, error) {
	query := `
		SELECT broadcaster_name
		FROM clip_submissions
		WHERE created_at > NOW() - $1 * INTERVAL '1 day'
			AND broadcaster_name IS NOT NULL
			AND broadcaster_name != ''
		GROUP BY broadcaster_name
		ORDER BY COUNT(*) DESC
	`

	rows, err := db.Query(ctx, query, lookbackDays)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var broadcasters []string
	for rows.Next() {
		var broadcasterName string
		if err := rows.Scan(&broadcasterName); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		broadcasters = append(broadcasters, broadcasterName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return broadcasters, nil
}

// scrapeClipsForBroadcaster fetches and inserts clips for a single broadcaster
func scrapeClipsForBroadcaster(
	ctx context.Context,
	db *pgxpool.Pool,
	twitchClient *twitch.Client,
	broadcasterID string,
	broadcasterName string,
	cfg ScraperConfig,
	stats *ScraperStats,
) (int, int, int, error) {
	clipsAdded := 0
	clipsSkipped := 0
	apiCalls := 0

	// Calculate time window
	endedAt := time.Now()
	startedAt := endedAt.AddDate(0, 0, -cfg.MaxAgeDays)

	// Fetch clips from Twitch API
	params := &twitch.ClipParams{
		BroadcasterID: broadcasterID,
		StartedAt:     startedAt,
		EndedAt:       endedAt,
		First:         cfg.BatchSize,
	}

	clipsResp, err := twitchClient.GetClips(ctx, params)
	apiCalls++
	if err != nil {
		return clipsAdded, clipsSkipped, apiCalls, fmt.Errorf("failed to fetch clips: %w", err)
	}

	log.Printf("  Retrieved %d clips from Twitch API", len(clipsResp.Data))

	// Process each clip
	for _, twitchClip := range clipsResp.Data {
		// Apply filters
		if twitchClip.ViewCount < cfg.MinViews {
			clipsSkipped++
			continue
		}

		// Check if clip already exists
		exists, err := clipExists(ctx, db, twitchClip.ID)
		if err != nil {
			log.Printf("  Warning: Failed to check if clip exists (id=%s): %v", twitchClip.ID, err)
			stats.Errors++
			stats.LastError = err
			continue
		}

		if exists {
			clipsSkipped++
			continue
		}

		// Insert clip if not in dry-run mode
		if !cfg.DryRun {
			if err := insertClip(ctx, db, &twitchClip); err != nil {
				log.Printf("  Warning: Failed to insert clip (id=%s): %v", twitchClip.ID, err)
				stats.Errors++
				stats.LastError = err
				continue
			}
		}

		clipsAdded++
	}

	return clipsAdded, clipsSkipped, apiCalls, nil
}

// clipExists checks if a clip already exists in the database (both clips and discovery_clips)
func clipExists(ctx context.Context, db *pgxpool.Pool, twitchClipID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(
		SELECT 1 FROM clips WHERE twitch_clip_id = $1
		UNION ALL
		SELECT 1 FROM discovery_clips WHERE twitch_clip_id = $1
	)`
	err := db.QueryRow(ctx, query, twitchClipID).Scan(&exists)
	return exists, err
}

// insertClip inserts a new clip into the discovery_clips staging table
func insertClip(ctx context.Context, db *pgxpool.Pool, twitchClip *twitch.Clip) error {
	id := uuid.New()
	now := time.Now()

	query := `
		INSERT INTO discovery_clips (
			id, twitch_clip_id, twitch_clip_url, embed_url, title,
			creator_name, creator_id, broadcaster_name, broadcaster_id,
			game_id, game_name, language, thumbnail_url, duration, view_count,
			created_at, imported_at, is_nsfw, is_removed, is_hidden
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19, $20
		)
	`

	_, err := db.Exec(ctx, query,
		id,
		twitchClip.ID,
		twitchClip.URL,
		twitchClip.EmbedURL,
		twitchClip.Title,
		twitchClip.CreatorName,
		&twitchClip.CreatorID,
		twitchClip.BroadcasterName,
		&twitchClip.BroadcasterID,
		&twitchClip.GameID,
		nil, // game_name - would require additional API call
		&twitchClip.Language,
		&twitchClip.ThumbnailURL,
		&twitchClip.Duration,
		twitchClip.ViewCount,
		twitchClip.CreatedAt,
		now,
		false, // is_nsfw
		false, // is_removed
		false, // is_hidden
	)

	return err
}

// printSummary prints the scraping summary
func printSummary(stats *ScraperStats) {
	duration := stats.EndTime.Sub(stats.StartTime)

	log.Println()
	log.Println("=== Scraping Summary ===")
	log.Printf("Duration: %v", duration)
	log.Printf("Broadcasters fetched: %d", stats.BroadcastersFetched)
	log.Printf("Broadcasters scraped: %d", stats.BroadcastersScraped)
	log.Printf("Clips checked: %d", stats.ClipsChecked)
	log.Printf("Clips inserted: %d", stats.ClipsInserted)
	log.Printf("Clips skipped: %d", stats.ClipsSkipped)
	log.Printf("Errors: %d", stats.Errors)
	log.Printf("API calls made: %d", stats.APICallsMade)

	if stats.ClipsInserted > 0 {
		avgTime := duration / time.Duration(stats.ClipsInserted)
		log.Printf("Average time per clip inserted: %v", avgTime)
	}

	if stats.LastError != nil {
		log.Printf("Last error: %v", stats.LastError)
	}

	// Success message
	if stats.Errors == 0 {
		log.Println("✓ Scraping completed successfully!")
	} else {
		log.Printf("⚠ Scraping completed with %d errors", stats.Errors)
	}
}
