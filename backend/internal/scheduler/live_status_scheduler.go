package scheduler

import (
	"context"
	"sync"
	"time"

	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

const liveStatusSchedulerName = "live_status"

// LiveStatusServiceInterface defines the interface required by the live status scheduler
type LiveStatusServiceInterface interface {
	UpdateLiveStatusForBroadcasters(ctx context.Context, broadcasterIDs []string) error
}

// BroadcasterRepositoryInterface defines the interface for broadcaster data access
type BroadcasterRepositoryInterface interface {
	GetAllFollowedBroadcasterIDs(ctx context.Context) ([]string, error)
}

// LiveStatusScheduler manages periodic live status updates
type LiveStatusScheduler struct {
	liveStatusService LiveStatusServiceInterface
	broadcasterRepo   BroadcasterRepositoryInterface
	interval          time.Duration
	stopChan          chan struct{}
	stopOnce          sync.Once
}

// NewLiveStatusScheduler creates a new live status scheduler
func NewLiveStatusScheduler(
	liveStatusService LiveStatusServiceInterface,
	broadcasterRepo BroadcasterRepositoryInterface,
	intervalSeconds int,
) *LiveStatusScheduler {
	return &LiveStatusScheduler{
		liveStatusService: liveStatusService,
		broadcasterRepo:   broadcasterRepo,
		interval:          time.Duration(intervalSeconds) * time.Second,
		stopChan:          make(chan struct{}),
	}
}

// Start begins the periodic live status update process
func (s *LiveStatusScheduler) Start(ctx context.Context) {
	utils.Info("Starting live status scheduler", map[string]interface{}{
		"scheduler": liveStatusSchedulerName,
		"interval":  s.interval.String(),
	})

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Run initial check
	s.updateLiveStatuses(ctx)

	for {
		select {
		case <-ticker.C:
			s.updateLiveStatuses(ctx)
		case <-s.stopChan:
			utils.Info("Live status scheduler stopped", map[string]interface{}{
				"scheduler": liveStatusSchedulerName,
			})
			return
		case <-ctx.Done():
			utils.Info("Live status scheduler stopped due to context cancellation", map[string]interface{}{
				"scheduler": liveStatusSchedulerName,
			})
			return
		}
	}
}

// Stop stops the scheduler in a thread-safe manner
func (s *LiveStatusScheduler) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopChan)
	})
}

// updateLiveStatuses executes a live status update operation
func (s *LiveStatusScheduler) updateLiveStatuses(ctx context.Context) {
	utils.Info("Starting scheduled live status update", map[string]interface{}{
		"scheduler": liveStatusSchedulerName,
	})
	startTime := time.Now()

	// Get all unique broadcaster IDs from follows using repository
	broadcasterIDs, err := s.broadcasterRepo.GetAllFollowedBroadcasterIDs(ctx)
	if err != nil {
		utils.Error("Failed to get followed broadcasters", err, map[string]interface{}{
			"scheduler": liveStatusSchedulerName,
		})
		return
	}

	if len(broadcasterIDs) == 0 {
		utils.Info("No broadcasters to check", map[string]interface{}{
			"scheduler": liveStatusSchedulerName,
		})
		return
	}

	utils.Info("Checking live status for broadcasters", map[string]interface{}{
		"scheduler": liveStatusSchedulerName,
		"count":     len(broadcasterIDs),
	})

	err = s.liveStatusService.UpdateLiveStatusForBroadcasters(ctx, broadcasterIDs)
	if err != nil {
		utils.Error("Live status update failed", err, map[string]interface{}{
			"scheduler": liveStatusSchedulerName,
			"count":     len(broadcasterIDs),
		})
		return
	}

	duration := time.Since(startTime)
	utils.Info("Live status update completed", map[string]interface{}{
		"scheduler": liveStatusSchedulerName,
		"duration":  duration.String(),
		"count":     len(broadcasterIDs),
	})
}
