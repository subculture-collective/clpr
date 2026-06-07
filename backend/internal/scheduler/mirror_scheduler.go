package scheduler

import (
	"context"
	"sync"
	"time"

	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

const mirrorSchedulerName = "mirror"

// MirrorServiceInterface defines the interface required by the mirror scheduler
type MirrorServiceInterface interface {
	SyncPopularClips(ctx context.Context) error
	CleanupExpiredMirrors(ctx context.Context) (int64, error)
}

// MirrorScheduler manages periodic mirror sync and cleanup operations
type MirrorScheduler struct {
	mirrorService   MirrorServiceInterface
	syncInterval    time.Duration
	cleanupInterval time.Duration
	stopChan        chan struct{}
	stopOnce        sync.Once
	startOnce       sync.Once
	running         bool
	mu              sync.Mutex
}

// NewMirrorScheduler creates a new mirror scheduler
func NewMirrorScheduler(
	mirrorService MirrorServiceInterface,
	syncIntervalMinutes int,
	cleanupIntervalMinutes int,
) *MirrorScheduler {
	return &MirrorScheduler{
		mirrorService:   mirrorService,
		syncInterval:    time.Duration(syncIntervalMinutes) * time.Minute,
		cleanupInterval: time.Duration(cleanupIntervalMinutes) * time.Minute,
		stopChan:        make(chan struct{}),
	}
}

// Start begins the periodic mirror sync and cleanup processes
func (s *MirrorScheduler) Start(ctx context.Context) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		utils.Warn("Mirror scheduler is already running", map[string]interface{}{
			"scheduler": mirrorSchedulerName,
		})
		return
	}
	s.running = true
	s.mu.Unlock()

	utils.Info("Starting mirror scheduler", map[string]interface{}{
		"scheduler":        mirrorSchedulerName,
		"sync_interval":    s.syncInterval.String(),
		"cleanup_interval": s.cleanupInterval.String(),
	})

	syncTicker := time.NewTicker(s.syncInterval)
	cleanupTicker := time.NewTicker(s.cleanupInterval)
	defer syncTicker.Stop()
	defer cleanupTicker.Stop()
	defer func() {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
	}()

	// Run initial sync
	s.syncMirrors(ctx)

	for {
		select {
		case <-syncTicker.C:
			s.syncMirrors(ctx)
		case <-cleanupTicker.C:
			s.cleanupMirrors(ctx)
		case <-s.stopChan:
			utils.Info("Mirror scheduler stopped", map[string]interface{}{
				"scheduler": mirrorSchedulerName,
			})
			return
		case <-ctx.Done():
			utils.Info("Mirror scheduler stopped due to context cancellation", map[string]interface{}{
				"scheduler": mirrorSchedulerName,
			})
			return
		}
	}
}

// Stop stops the mirror scheduler
func (s *MirrorScheduler) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopChan)
	})
}

// syncMirrors performs the mirror sync operation
func (s *MirrorScheduler) syncMirrors(ctx context.Context) {
	utils.Info("Running mirror sync", map[string]interface{}{
		"scheduler": mirrorSchedulerName,
		"task":      "sync",
	})
	start := time.Now()

	if err := s.mirrorService.SyncPopularClips(ctx); err != nil {
		utils.Error("Mirror sync failed", err, map[string]interface{}{
			"scheduler": mirrorSchedulerName,
			"task":      "sync",
		})
		return
	}

	duration := time.Since(start)
	utils.Info("Mirror sync completed", map[string]interface{}{
		"scheduler": mirrorSchedulerName,
		"task":      "sync",
		"duration":  duration.String(),
	})
}

// cleanupMirrors performs the mirror cleanup operation
func (s *MirrorScheduler) cleanupMirrors(ctx context.Context) {
	utils.Info("Running mirror cleanup", map[string]interface{}{
		"scheduler": mirrorSchedulerName,
		"task":      "cleanup",
	})
	start := time.Now()

	count, err := s.mirrorService.CleanupExpiredMirrors(ctx)
	if err != nil {
		utils.Error("Mirror cleanup failed", err, map[string]interface{}{
			"scheduler": mirrorSchedulerName,
			"task":      "cleanup",
		})
		return
	}

	duration := time.Since(start)
	utils.Info("Mirror cleanup completed", map[string]interface{}{
		"scheduler": mirrorSchedulerName,
		"task":      "cleanup",
		"duration":  duration.String(),
		"removed":   count,
	})
}
