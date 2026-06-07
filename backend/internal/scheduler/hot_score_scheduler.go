package scheduler

import (
	"context"
	"sync"
	"time"

	"git.subcult.tv/subculture-collective/clpr/pkg/metrics"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

const (
	hotScoreSchedulerName = "hot_score"
	hotScoreJobName       = "hot_score_refresh"
)

// ClipRepositoryInterface defines the interface required by the hot score scheduler
type ClipRepositoryInterface interface {
	RefreshHotScores(ctx context.Context) error
}

// HotScoreScheduler manages periodic hot score computation/updates
type HotScoreScheduler struct {
	clipRepo ClipRepositoryInterface
	interval time.Duration
	stopChan chan struct{}
	stopOnce sync.Once
}

// NewHotScoreScheduler creates a new hot score scheduler
func NewHotScoreScheduler(clipRepo ClipRepositoryInterface, intervalMinutes int) *HotScoreScheduler {
	return &HotScoreScheduler{
		clipRepo: clipRepo,
		interval: time.Duration(intervalMinutes) * time.Minute,
		stopChan: make(chan struct{}),
	}
}

// Start begins the periodic hot score refresh process
func (s *HotScoreScheduler) Start(ctx context.Context) {
	utils.Info("Starting hot score scheduler", map[string]interface{}{
		"scheduler": hotScoreSchedulerName,
		"interval":  s.interval.String(),
	})

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Run initial refresh
	s.refreshHotScores(ctx)

	for {
		select {
		case <-ticker.C:
			s.refreshHotScores(ctx)
		case <-s.stopChan:
			utils.Info("Hot score scheduler stopped", map[string]interface{}{
				"scheduler": hotScoreSchedulerName,
			})
			return
		case <-ctx.Done():
			utils.Info("Hot score scheduler stopped due to context cancellation", map[string]interface{}{
				"scheduler": hotScoreSchedulerName,
			})
			return
		}
	}
}

// Stop stops the scheduler in a thread-safe manner
func (s *HotScoreScheduler) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopChan)
	})
}

// refreshHotScores executes a hot score refresh operation
func (s *HotScoreScheduler) refreshHotScores(ctx context.Context) {
	utils.Info("Starting scheduled hot score refresh", map[string]interface{}{
		"scheduler": hotScoreSchedulerName,
		"job":       hotScoreJobName,
	})
	startTime := time.Now()

	err := s.clipRepo.RefreshHotScores(ctx)
	duration := time.Since(startTime)

	// Record metrics
	metrics.JobExecutionDuration.WithLabelValues(hotScoreJobName).Observe(duration.Seconds())

	if err != nil {
		utils.Error("Hot score refresh failed", err, map[string]interface{}{
			"scheduler": hotScoreSchedulerName,
			"job":       hotScoreJobName,
		})
		metrics.JobExecutionTotal.WithLabelValues(hotScoreJobName, "failed").Inc()
		return
	}

	metrics.JobExecutionTotal.WithLabelValues(hotScoreJobName, "success").Inc()
	metrics.JobLastSuccessTimestamp.WithLabelValues(hotScoreJobName).Set(float64(time.Now().Unix()))
	utils.Info("Hot score refresh completed", map[string]interface{}{
		"scheduler": hotScoreSchedulerName,
		"job":       hotScoreJobName,
		"duration":  duration.String(),
	})
}
