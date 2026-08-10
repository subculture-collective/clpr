package scheduler

import (
	"context"
	"sync"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/pkg/metrics"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

const (
	clipSyncSchedulerName = "clip_sync"
	clipSyncJobName       = "clip_sync"
)

// ClipSyncServiceInterface defines the interface required by the scheduler
type ClipSyncServiceInterface interface {
	SyncTrendingClips(ctx context.Context, hours int, opts *services.TrendingSyncOptions) (*services.SyncStats, error)
	SyncFollowedBroadcasterClips(ctx context.Context, opts *services.FollowedBroadcasterSyncOptions) (*services.SyncStats, error)
	SyncGlobalTrending(ctx context.Context, limit int) (*services.SyncStats, error)
}

// ClipSyncScheduler manages periodic clip synchronization
type ClipSyncScheduler struct {
	syncService ClipSyncServiceInterface
	interval    time.Duration
	stopChan    chan struct{}
	stopOnce    sync.Once
}

// NewClipSyncScheduler creates a new scheduler
func NewClipSyncScheduler(syncService ClipSyncServiceInterface, intervalMinutes int) *ClipSyncScheduler {
	return &ClipSyncScheduler{
		syncService: syncService,
		interval:    time.Duration(intervalMinutes) * time.Minute,
		stopChan:    make(chan struct{}),
	}
}

// Start begins the periodic sync process
func (s *ClipSyncScheduler) Start(ctx context.Context) {
	utils.Info("Starting clip sync scheduler", map[string]interface{}{
		"scheduler": clipSyncSchedulerName,
		"interval":  s.interval.String(),
	})

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Run initial sync
	s.runSync(ctx)

	for {
		select {
		case <-ticker.C:
			s.runSync(ctx)
		case <-s.stopChan:
			utils.Info("Clip sync scheduler stopped", map[string]interface{}{
				"scheduler": clipSyncSchedulerName,
			})
			return
		case <-ctx.Done():
			utils.Info("Clip sync scheduler stopped due to context cancellation", map[string]interface{}{
				"scheduler": clipSyncSchedulerName,
			})
			return
		}
	}
}

// Stop stops the scheduler in a thread-safe manner
func (s *ClipSyncScheduler) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopChan)
	})
}

// runSync executes a sync operation
func (s *ClipSyncScheduler) runSync(ctx context.Context) {
	utils.Info("Starting scheduled clip sync", map[string]interface{}{
		"scheduler": clipSyncSchedulerName,
		"job":       clipSyncJobName,
	})
	startTime := time.Now()

	// Sync trending clips from the last 24 hours
	// Rotate pagination over a fixed window to keep volume low
	stats, err := s.syncService.SyncTrendingClips(ctx, 24, &services.TrendingSyncOptions{MaxPages: services.DefaultTrendingPageWindow})
	duration := time.Since(startTime)

	// Record metrics
	metrics.JobExecutionDuration.WithLabelValues(clipSyncJobName).Observe(duration.Seconds())

	if err != nil {
		utils.Error("Scheduled sync failed", err, map[string]interface{}{
			"scheduler": clipSyncSchedulerName,
			"job":       clipSyncJobName,
		})
		metrics.JobExecutionTotal.WithLabelValues(clipSyncJobName, "failed").Inc()
		return
	}

	// Sync platform-wide global trending clips (no category filter)
	globalStartTime := time.Now()
	globalStats, globalErr := s.syncService.SyncGlobalTrending(ctx, 20)
	globalDuration := time.Since(globalStartTime)
	metrics.JobExecutionDuration.WithLabelValues("clip_sync_global").Observe(globalDuration.Seconds())

	if globalErr != nil {
		utils.Error("Global trending sync failed", globalErr, map[string]interface{}{
			"scheduler": clipSyncSchedulerName,
			"job":       "clip_sync_global",
		})
		metrics.JobExecutionTotal.WithLabelValues("clip_sync_global", "failed").Inc()
	} else {
		metrics.JobExecutionTotal.WithLabelValues("clip_sync_global", "success").Inc()
		metrics.JobLastSuccessTimestamp.WithLabelValues("clip_sync_global").Set(float64(time.Now().Unix()))
		metrics.JobItemsProcessed.WithLabelValues("clip_sync_global", "success").Add(float64(globalStats.ClipsCreated + globalStats.ClipsUpdated))
		metrics.JobItemsProcessed.WithLabelValues("clip_sync_global", "skipped").Add(float64(globalStats.ClipsSkipped))
		if len(globalStats.Errors) > 0 {
			metrics.JobItemsProcessed.WithLabelValues("clip_sync_global", "failed").Add(float64(len(globalStats.Errors)))
		}

		utils.Info("Global trending sync completed", map[string]interface{}{
			"scheduler": clipSyncSchedulerName,
			"job":       "clip_sync_global",
			"fetched":   globalStats.ClipsFetched,
			"created":   globalStats.ClipsCreated,
			"updated":   globalStats.ClipsUpdated,
			"skipped":   globalStats.ClipsSkipped,
			"errors":    len(globalStats.Errors),
			"duration":  globalDuration.String(),
		})
	}

	// Sync clips from followed broadcasters (broadcasters with ≥3 clpr followers)
	followerOpts := &services.FollowedBroadcasterSyncOptions{
		MinFollowers:        3,
		ClipsPerBroadcaster: 5,
		MaxTotalClips:       200,
	}
	followerStartTime := time.Now()
	followerStats, followerErr := s.syncService.SyncFollowedBroadcasterClips(ctx, followerOpts)
	followerDuration := time.Since(followerStartTime)
	metrics.JobExecutionDuration.WithLabelValues("clip_sync_followed").Observe(followerDuration.Seconds())

	if followerErr != nil {
		utils.Error("Followed broadcaster sync failed", followerErr, map[string]interface{}{
			"scheduler": clipSyncSchedulerName,
			"job":       "clip_sync_followed",
		})
		metrics.JobExecutionTotal.WithLabelValues("clip_sync_followed", "failed").Inc()
	} else {
		metrics.JobExecutionTotal.WithLabelValues("clip_sync_followed", "success").Inc()
		metrics.JobLastSuccessTimestamp.WithLabelValues("clip_sync_followed").Set(float64(time.Now().Unix()))
		metrics.JobItemsProcessed.WithLabelValues("clip_sync_followed", "success").Add(float64(followerStats.ClipsCreated + followerStats.ClipsUpdated))
		metrics.JobItemsProcessed.WithLabelValues("clip_sync_followed", "skipped").Add(float64(followerStats.ClipsSkipped))
		if len(followerStats.Errors) > 0 {
			metrics.JobItemsProcessed.WithLabelValues("clip_sync_followed", "failed").Add(float64(len(followerStats.Errors)))
		}

		utils.Info("Followed broadcaster sync completed", map[string]interface{}{
			"scheduler":  clipSyncSchedulerName,
			"job":        "clip_sync_followed",
			"fetched":    followerStats.ClipsFetched,
			"created":    followerStats.ClipsCreated,
			"updated":    followerStats.ClipsUpdated,
			"skipped":    followerStats.ClipsSkipped,
			"errors":     len(followerStats.Errors),
			"duration":   followerDuration.String(),
		})
	}

	metrics.JobExecutionTotal.WithLabelValues(clipSyncJobName, "success").Inc()
	metrics.JobLastSuccessTimestamp.WithLabelValues(clipSyncJobName).Set(float64(time.Now().Unix()))
	metrics.JobItemsProcessed.WithLabelValues(clipSyncJobName, "success").Add(float64(stats.ClipsCreated + stats.ClipsUpdated))
	metrics.JobItemsProcessed.WithLabelValues(clipSyncJobName, "skipped").Add(float64(stats.ClipsSkipped))
	if len(stats.Errors) > 0 {
		metrics.JobItemsProcessed.WithLabelValues(clipSyncJobName, "failed").Add(float64(len(stats.Errors)))
	}

	utils.Info("Scheduled sync completed", map[string]interface{}{
		"scheduler": clipSyncSchedulerName,
		"job":       clipSyncJobName,
		"fetched":   stats.ClipsFetched,
		"created":   stats.ClipsCreated,
		"updated":   stats.ClipsUpdated,
		"skipped":   stats.ClipsSkipped,
		"errors":    len(stats.Errors),
		"duration":  stats.EndTime.Sub(stats.StartTime).String(),
	})

	if len(stats.Errors) > 0 {
		utils.Warn("Sync had errors", map[string]interface{}{
			"scheduler": clipSyncSchedulerName,
			"job":       clipSyncJobName,
			"errors":    stats.Errors,
		})
	}
}
