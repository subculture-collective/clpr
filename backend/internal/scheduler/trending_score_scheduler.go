package scheduler

import (
	"context"
	"errors"
	"sync"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/pkg/metrics"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

const (
	trendingScoreSchedulerName = "trending_score"
	trendingScoreJobName       = "trending_score_refresh"
)

// TrendingScoreRepositoryInterface defines the interface required by the trending score scheduler
type TrendingScoreRepositoryInterface interface {
	UpdateTrendingScores(ctx context.Context) (int64, error)
}

// TrendingScoreScheduler manages periodic trending score computation/updates
type TrendingScoreScheduler struct {
	clipRepo  TrendingScoreRepositoryInterface
	interval  time.Duration
	batchSize int
	stopChan  chan struct{}
	stopOnce  sync.Once
}

type batchedTrendingScoreRepository interface {
	UpdateTrendingScoresBatched(context.Context, int) (int64, error)
}

func (s *TrendingScoreScheduler) SetBatchSize(size int) *TrendingScoreScheduler {
	if size > 0 {
		s.batchSize = size
	}
	return s
}

// NewTrendingScoreScheduler creates a new trending score scheduler
func NewTrendingScoreScheduler(clipRepo TrendingScoreRepositoryInterface, intervalMinutes int) *TrendingScoreScheduler {
	if intervalMinutes <= 0 {
		intervalMinutes = 60
	}
	return &TrendingScoreScheduler{
		clipRepo: clipRepo,
		interval: time.Duration(intervalMinutes) * time.Minute,
		stopChan: make(chan struct{}),
	}
}

// Start begins the periodic trending score refresh process
func (s *TrendingScoreScheduler) Start(ctx context.Context) {
	utils.Info("Starting trending score scheduler", map[string]interface{}{
		"scheduler": trendingScoreSchedulerName,
		"interval":  s.interval.String(),
	})

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Run initial refresh
	s.refreshTrendingScores(ctx)

	for {
		select {
		case <-ticker.C:
			s.refreshTrendingScores(ctx)
		case <-s.stopChan:
			utils.Info("Trending score scheduler stopped", map[string]interface{}{
				"scheduler": trendingScoreSchedulerName,
			})
			return
		case <-ctx.Done():
			utils.Info("Trending score scheduler stopped due to context cancellation", map[string]interface{}{
				"scheduler": trendingScoreSchedulerName,
			})
			return
		}
	}
}

// Stop stops the scheduler in a thread-safe manner
func (s *TrendingScoreScheduler) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopChan)
	})
}

// refreshTrendingScores executes a trending score refresh operation
func (s *TrendingScoreScheduler) refreshTrendingScores(ctx context.Context) {
	utils.Info("Starting scheduled trending score refresh", map[string]interface{}{
		"scheduler": trendingScoreSchedulerName,
		"job":       trendingScoreJobName,
	})
	startTime := time.Now()

	var rowsUpdated int64
	var err error
	if batched, ok := s.clipRepo.(batchedTrendingScoreRepository); ok {
		rowsUpdated, err = batched.UpdateTrendingScoresBatched(ctx, s.batchSize)
	} else {
		rowsUpdated, err = s.clipRepo.UpdateTrendingScores(ctx)
	}
	duration := time.Since(startTime)

	// Record metrics
	metrics.JobExecutionDuration.WithLabelValues(trendingScoreJobName).Observe(duration.Seconds())

	if err != nil {
		if errors.Is(err, repository.ErrSchedulerLockUnavailable) {
			metrics.JobExecutionTotal.WithLabelValues(trendingScoreJobName, "skipped").Inc()
			return
		}
		utils.Error("Trending score refresh failed", err, map[string]interface{}{
			"scheduler": trendingScoreSchedulerName,
			"job":       trendingScoreJobName,
		})
		metrics.JobExecutionTotal.WithLabelValues(trendingScoreJobName, "failed").Inc()
		return
	}

	metrics.JobExecutionTotal.WithLabelValues(trendingScoreJobName, "success").Inc()
	metrics.JobLastSuccessTimestamp.WithLabelValues(trendingScoreJobName).Set(float64(time.Now().Unix()))
	metrics.JobItemsProcessed.WithLabelValues(trendingScoreJobName, "success").Add(float64(rowsUpdated))
	utils.Info("Trending score refresh completed", map[string]interface{}{
		"scheduler":    trendingScoreSchedulerName,
		"job":          trendingScoreJobName,
		"duration":     duration.String(),
		"rows_updated": rowsUpdated,
	})
}
