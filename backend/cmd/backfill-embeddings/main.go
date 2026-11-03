package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/pkg/database"
	"git.subcult.tv/subculture-collective/clpr/pkg/redis"
	"github.com/google/uuid"
	pgvector "github.com/pgvector/pgvector-go"
)

type BackfillStats struct {
	TotalClips     int
	ProcessedClips int
	SkippedClips   int
	FailedClips    int
	StartTime      time.Time
	EndTime        time.Time
	LastError      error
}

func main() {
	batchSize := flag.Int("batch", 50, "Number of clips to process in each batch")
	forceUpdate := flag.Bool("force", false, "Force update existing embeddings")
	dryRun := flag.Bool("dry-run", false, "Dry run mode - don't save embeddings")
	flag.Parse()

	log.Println("Starting embedding backfill job...")
	log.Printf("Configuration: batch_size=%d, force_update=%t, dry_run=%t", *batchSize, *forceUpdate, *dryRun)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if !cfg.Embedding.Enabled {
		log.Println("WARNING: Embedding service is disabled in configuration (EMBEDDING_ENABLED=false)")
		log.Println("Set EMBEDDING_ENABLED=true to enable embedding generation")
	}

	if cfg.Embedding.OpenAIAPIKey == "" {
		log.Fatalf("OPENAI_API_KEY is not set. Please set it in your environment or .env file")
	}

	// Initialize database connection
	db, err := database.NewDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Database connection established")

	// Initialize Redis client
	redisClient, err := redis.NewClient(&cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()
	log.Println("Redis connection established")

	// Initialize embedding service
	embeddingService := services.NewEmbeddingService(&services.EmbeddingConfig{
		APIKey:            cfg.Embedding.OpenAIAPIKey,
		APIBaseURL:        cfg.Embedding.APIBaseURL,
		Model:             cfg.Embedding.Model,
		RedisClient:       redisClient.GetClient(),
		RequestsPerMinute: cfg.Embedding.RequestsPerMinute,
	})
	defer embeddingService.Close()
	log.Printf("Embedding service initialized (model: %s)", cfg.Embedding.Model)

	// Run backfill
	ctx := context.Background()
	stats, err := backfillEmbeddings(ctx, db, embeddingService, *batchSize, *forceUpdate, *dryRun)
	if err != nil {
		log.Fatalf("Backfill failed: %v", err)
	}

	// Print summary
	duration := stats.EndTime.Sub(stats.StartTime)
	log.Println("\n=== Backfill Summary ===")
	log.Printf("Total clips: %d", stats.TotalClips)
	log.Printf("Processed: %d", stats.ProcessedClips)
	log.Printf("Skipped: %d", stats.SkippedClips)
	log.Printf("Failed: %d", stats.FailedClips)
	log.Printf("Duration: %v", duration)

	if stats.ProcessedClips > 0 {
		avgTime := duration / time.Duration(stats.ProcessedClips)
		log.Printf("Average time per clip: %v", avgTime)
	}

	if stats.LastError != nil {
		log.Printf("Last error: %v", stats.LastError)
	}

	if stats.FailedClips > 0 {
		log.Printf("Backfill finished with %d failed clip(s)", stats.FailedClips)
		os.Exit(1)
	}

	log.Println("\n✓ Backfill completed successfully!")
}

// embeddingGenerator is the subset of EmbeddingService used by the backfill.
type embeddingGenerator interface {
	GenerateClipEmbedding(ctx context.Context, clip *models.Clip) ([]float32, error)
	GetModel() string
}

// saveEmbeddingFunc persists a generated embedding for a clip.
type saveEmbeddingFunc func(ctx context.Context, clip *models.Clip, embedding []float32, model string) error

// backfillBatchQuery returns the batch query. Pages are selected by keyset on
// id rather than OFFSET: clips leave the "embedding IS NULL" filter as they are
// updated, so OFFSET paging would skip rows. Keyset paging is stable whether or
// not rows leave the filter.
func backfillBatchQuery(forceUpdate, hasCursor bool) string {
	query := `
			SELECT id, twitch_clip_id, title, creator_name, broadcaster_name,
			       game_id, game_name, embedding, embedding_generated_at, embedding_model
			FROM clips
			WHERE is_removed = false
		`
	if !forceUpdate {
		query += ` AND embedding IS NULL`
	}
	if hasCursor {
		query += ` AND id > $2`
	}
	query += `
			ORDER BY id ASC
			LIMIT $1
		`
	return query
}

func backfillEmbeddings(
	ctx context.Context,
	db *database.DB,
	embeddingService *services.EmbeddingService,
	batchSize int,
	forceUpdate bool,
	dryRun bool,
) (*BackfillStats, error) {
	stats := &BackfillStats{
		StartTime: time.Now(),
	}

	// Count total clips
	countQuery := `SELECT COUNT(*) FROM clips WHERE is_removed = false`
	if !forceUpdate {
		countQuery += ` AND embedding IS NULL`
	}

	if err := db.Pool.QueryRow(ctx, countQuery).Scan(&stats.TotalClips); err != nil {
		return nil, fmt.Errorf("failed to count clips: %w", err)
	}

	log.Printf("Found %d clips to process", stats.TotalClips)

	if stats.TotalClips == 0 {
		stats.EndTime = time.Now()
		return stats, nil
	}

	save := func(ctx context.Context, clip *models.Clip, embedding []float32, model string) error {
		updateQuery := `
				UPDATE clips
				SET embedding = $1,
				    embedding_generated_at = $2,
				    embedding_model = $3
				WHERE id = $4
			`
		_, err := db.Pool.Exec(ctx, updateQuery, pgvector.NewVector(embedding), time.Now(), model, clip.ID)
		return err
	}

	var cursor *uuid.UUID
	for {
		// Fetch batch of clips
		args := []interface{}{batchSize}
		if cursor != nil {
			args = append(args, *cursor)
		}
		rows, err := db.Pool.Query(ctx, backfillBatchQuery(forceUpdate, cursor != nil), args...)
		if err != nil {
			return stats, fmt.Errorf("failed to fetch clips: %w", err)
		}

		var clips []models.Clip
		for rows.Next() {
			var clip models.Clip
			err := rows.Scan(
				&clip.ID,
				&clip.TwitchClipID,
				&clip.Title,
				&clip.CreatorName,
				&clip.BroadcasterName,
				&clip.GameID,
				&clip.GameName,
				&clip.Embedding,
				&clip.EmbeddingGeneratedAt,
				&clip.EmbeddingModel,
			)
			if err != nil {
				rows.Close()
				return stats, fmt.Errorf("failed to scan clip: %w", err)
			}
			clips = append(clips, clip)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return stats, fmt.Errorf("failed to iterate clips: %w", err)
		}

		if len(clips) == 0 {
			break
		}

		// Process batch
		log.Printf("Processing batch of %d clips (progress: %.1f%%)",
			len(clips), float64(stats.ProcessedClips+stats.SkippedClips+stats.FailedClips)/float64(stats.TotalClips)*100)

		processBatch(ctx, clips, embeddingService, save, forceUpdate, dryRun, stats)

		lastID := clips[len(clips)-1].ID
		cursor = &lastID

		// Small delay between batches to avoid overwhelming the API
		time.Sleep(100 * time.Millisecond)
	}

	stats.EndTime = time.Now()
	return stats, nil
}

// processBatch generates and saves embeddings for one batch, updating stats.
// Dry runs report what would be processed without calling the embedding API.
func processBatch(
	ctx context.Context,
	clips []models.Clip,
	generator embeddingGenerator,
	save saveEmbeddingFunc,
	forceUpdate bool,
	dryRun bool,
	stats *BackfillStats,
) {
	for i := range clips {
		clip := &clips[i]

		// Skip if already has embedding and not forcing update
		if !forceUpdate && len(clip.Embedding) > 0 {
			stats.SkippedClips++
			continue
		}

		if dryRun {
			log.Printf("DRY RUN: Would generate and save embedding for clip %s", clip.ID)
			stats.ProcessedClips++
			continue
		}

		// Generate embedding
		embedding, err := generator.GenerateClipEmbedding(ctx, clip)
		if err != nil {
			log.Printf("WARNING: Failed to generate embedding for clip %s: %v", clip.ID, err)
			stats.FailedClips++
			stats.LastError = err
			continue
		}

		// Save embedding to database
		if err := save(ctx, clip, embedding, generator.GetModel()); err != nil {
			log.Printf("WARNING: Failed to save embedding for clip %s: %v", clip.ID, err)
			stats.FailedClips++
			stats.LastError = err
			continue
		}

		stats.ProcessedClips++

		// Log progress every 10 clips
		if stats.ProcessedClips%10 == 0 && stats.TotalClips > 0 {
			progress := float64(stats.ProcessedClips+stats.SkippedClips+stats.FailedClips) / float64(stats.TotalClips) * 100
			log.Printf("Progress: %d/%d (%.1f%%) - processed: %d, skipped: %d, failed: %d",
				stats.ProcessedClips+stats.SkippedClips+stats.FailedClips,
				stats.TotalClips,
				progress,
				stats.ProcessedClips,
				stats.SkippedClips,
				stats.FailedClips,
			)
		}
	}
}
