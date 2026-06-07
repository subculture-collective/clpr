package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	pgvector "github.com/pgvector/pgvector-go"
	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/pkg/database"
	"git.subcult.tv/subculture-collective/clpr/pkg/redis"
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

	log.Println("\n✓ Backfill completed successfully!")
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

	offset := 0
	for {
		// Fetch batch of clips
		query := `
			SELECT id, twitch_clip_id, title, creator_name, broadcaster_name,
			       game_id, game_name, embedding, embedding_generated_at, embedding_model
			FROM clips
			WHERE is_removed = false
		`
		if !forceUpdate {
			query += ` AND embedding IS NULL`
		}
		query += `
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`

		rows, err := db.Pool.Query(ctx, query, batchSize, offset)
		if err != nil {
			return stats, fmt.Errorf("failed to fetch clips: %w", err)
		}
		defer rows.Close()

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
				return stats, fmt.Errorf("failed to scan clip: %w", err)
			}
			clips = append(clips, clip)
		}

		if len(clips) == 0 {
			break
		}

		// Process batch
		log.Printf("Processing batch of %d clips (offset: %d, progress: %.1f%%)",
			len(clips), offset, float64(stats.ProcessedClips+stats.SkippedClips+stats.FailedClips)/float64(stats.TotalClips)*100)

		for i := range clips {
			clip := &clips[i]

			// Skip if already has embedding and not forcing update
			if !forceUpdate && len(clip.Embedding) > 0 {
				stats.SkippedClips++
				continue
			}

			// Generate embedding
			embedding, err := embeddingService.GenerateClipEmbedding(ctx, clip)
			if err != nil {
				log.Printf("WARNING: Failed to generate embedding for clip %s: %v", clip.ID, err)
				stats.FailedClips++
				stats.LastError = err
				continue
			}

			if dryRun {
				log.Printf("DRY RUN: Would update clip %s with embedding (length: %d)", clip.ID, len(embedding))
				stats.ProcessedClips++
				continue
			}

			// Save embedding to database
			now := time.Now()
			model := embeddingService.GetModel()
			updateQuery := `
				UPDATE clips
				SET embedding = $1,
				    embedding_generated_at = $2,
				    embedding_model = $3
				WHERE id = $4
			`

			_, err = db.Pool.Exec(ctx, updateQuery, pgvector.NewVector(embedding), now, model, clip.ID)
			if err != nil {
				log.Printf("WARNING: Failed to save embedding for clip %s: %v", clip.ID, err)
				stats.FailedClips++
				stats.LastError = err
				continue
			}

			stats.ProcessedClips++

			// Log progress every 10 clips
			if stats.ProcessedClips%10 == 0 {
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

		offset += batchSize

		// Small delay between batches to avoid overwhelming the API
		time.Sleep(100 * time.Millisecond)
	}

	stats.EndTime = time.Now()
	return stats, nil
}
