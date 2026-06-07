package scheduler

import (
	"context"
	"sync"
	"time"

	pgvector "github.com/pgvector/pgvector-go"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/pkg/database"
	"git.subcult.tv/subculture-collective/clpr/pkg/metrics"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

const embeddingSchedulerName = "embedding_generation"

// EmbeddingServiceInterface defines the interface required by the scheduler
type EmbeddingServiceInterface interface {
	GenerateClipEmbedding(ctx context.Context, clip *models.Clip) ([]float32, error)
	Close()
}

// EmbeddingScheduler manages periodic embedding generation for new clips
type EmbeddingScheduler struct {
	db               *database.DB
	embeddingService EmbeddingServiceInterface
	interval         time.Duration
	stopChan         chan struct{}
	stopOnce         sync.Once
	model            string
}

// NewEmbeddingScheduler creates a new scheduler
func NewEmbeddingScheduler(
	db *database.DB,
	embeddingService EmbeddingServiceInterface,
	intervalMinutes int,
	model string,
) *EmbeddingScheduler {
	return &EmbeddingScheduler{
		db:               db,
		embeddingService: embeddingService,
		interval:         time.Duration(intervalMinutes) * time.Minute,
		stopChan:         make(chan struct{}),
		model:            model,
	}
}

// Start begins the periodic embedding generation process
func (s *EmbeddingScheduler) Start(ctx context.Context) {
	utils.Info("Starting embedding scheduler", map[string]interface{}{
		"scheduler": embeddingSchedulerName,
		"interval":  s.interval.String(),
		"model":     s.model,
	})

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Run initial embedding generation
	s.runEmbedding(ctx)

	for {
		select {
		case <-ticker.C:
			s.runEmbedding(ctx)
		case <-s.stopChan:
			utils.Info("Embedding scheduler stopped", map[string]interface{}{
				"scheduler": embeddingSchedulerName,
			})
			return
		case <-ctx.Done():
			utils.Info("Embedding scheduler stopped due to context cancellation", map[string]interface{}{
				"scheduler": embeddingSchedulerName,
			})
			return
		}
	}
}

// Stop stops the scheduler in a thread-safe manner
func (s *EmbeddingScheduler) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopChan)
	})
}

// runEmbedding executes embedding generation for clips without embeddings
func (s *EmbeddingScheduler) runEmbedding(ctx context.Context) {
	utils.Info("Starting scheduled embedding generation", map[string]interface{}{
		"scheduler": embeddingSchedulerName,
		"model":     s.model,
	})

	// Skip if database is not available (e.g., in tests)
	if s.db == nil {
		utils.Warn("Database not available, skipping embedding generation", map[string]interface{}{
			"scheduler": embeddingSchedulerName,
			"model":     s.model,
		})
		return
	}

	startTime := time.Now()

	// Fetch clips without embeddings (created in the last 7 days to avoid old clips)
	query := `
		SELECT id, twitch_clip_id, title, creator_name, broadcaster_name,
		       game_id, game_name
		FROM clips
		WHERE is_removed = false
		  AND embedding IS NULL
		  AND created_at > NOW() - INTERVAL '7 days'
		ORDER BY created_at DESC
		LIMIT 100
	`

	rows, err := s.db.Pool.Query(ctx, query)
	if err != nil {
		utils.Error("Failed to fetch clips for embedding", err, map[string]interface{}{
			"scheduler": embeddingSchedulerName,
			"model":     s.model,
		})
		metrics.IndexingJobsTotal.WithLabelValues("failed").Inc()
		return
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
		)
		if err != nil {
			utils.Error("Failed to scan clip", err, map[string]interface{}{
				"scheduler": embeddingSchedulerName,
			})
			continue
		}
		clips = append(clips, clip)
	}

	if len(clips) == 0 {
		utils.Info("No clips need embeddings - all up to date", map[string]interface{}{
			"scheduler": embeddingSchedulerName,
			"model":     s.model,
		})
		// Update embedding coverage metrics
		s.updateEmbeddingCoverageMetrics(ctx)
		return
	}

	utils.Info("Generating embeddings for clips", map[string]interface{}{
		"scheduler": embeddingSchedulerName,
		"count":     len(clips),
		"model":     s.model,
	})

	processed := 0
	failed := 0

	for i := range clips {
		clip := &clips[i]

		// Generate embedding
		embedding, err := s.embeddingService.GenerateClipEmbedding(ctx, clip)
		if err != nil {
			utils.Error("Failed to generate embedding for clip", err, map[string]interface{}{
				"scheduler": embeddingSchedulerName,
				"clip_id":   clip.ID,
				"model":     s.model,
			})
			failed++
			continue
		}

		// Save to database
		now := time.Now()
		updateQuery := `
			UPDATE clips
			SET embedding = $1,
			    embedding_generated_at = $2,
			    embedding_model = $3
			WHERE id = $4
		`

		_, err = s.db.Pool.Exec(ctx, updateQuery, pgvector.NewVector(embedding), now, s.model, clip.ID)
		if err != nil {
			utils.Error("Failed to save embedding for clip", err, map[string]interface{}{
				"scheduler": embeddingSchedulerName,
				"clip_id":   clip.ID,
				"model":     s.model,
			})
			failed++
			continue
		}

		processed++
	}

	duration := time.Since(startTime)
	utils.Info("Scheduled embedding generation completed", map[string]interface{}{
		"scheduler": embeddingSchedulerName,
		"processed": processed,
		"failed":    failed,
		"duration":  duration.String(),
		"model":     s.model,
	})

	// Record indexing job metrics
	if failed == 0 {
		metrics.IndexingJobsTotal.WithLabelValues("success").Inc()
	} else if processed > 0 {
		metrics.IndexingJobsTotal.WithLabelValues("partial").Inc()
	} else {
		metrics.IndexingJobsTotal.WithLabelValues("failed").Inc()
	}
	metrics.IndexingJobDuration.Observe(duration.Seconds())

	// Update embedding coverage metrics
	s.updateEmbeddingCoverageMetrics(ctx)
}

// updateEmbeddingCoverageMetrics updates the Prometheus gauges for embedding coverage
func (s *EmbeddingScheduler) updateEmbeddingCoverageMetrics(ctx context.Context) {
	if s.db == nil {
		return
	}

	var withEmbeddings, withoutEmbeddings int64

	// Count clips with embeddings
	err := s.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM clips WHERE is_removed = false AND embedding IS NOT NULL
	`).Scan(&withEmbeddings)
	if err != nil {
		utils.Error("Failed to count clips with embeddings", err, map[string]interface{}{
			"scheduler": embeddingSchedulerName,
		})
		return
	}

	// Count clips without embeddings
	err = s.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM clips WHERE is_removed = false AND embedding IS NULL
	`).Scan(&withoutEmbeddings)
	if err != nil {
		utils.Error("Failed to count clips without embeddings", err, map[string]interface{}{
			"scheduler": embeddingSchedulerName,
		})
		return
	}

	metrics.ClipsWithEmbeddings.Set(float64(withEmbeddings))
	metrics.ClipsWithoutEmbeddings.Set(float64(withoutEmbeddings))
}
