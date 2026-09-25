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
	GetLiveStatusCandidateBroadcasterIDs(ctx context.Context, limit int, window time.Duration) ([]string, error)
}

// LiveStatusCandidateConfig bounds the broadcasters checked because they
// appear on the public site (live list, live badges), in addition to
// broadcasters that users follow.
type LiveStatusCandidateConfig struct {
	// Limit is the maximum number of candidates; 0 disables them. Every 100
	// broadcasters checked cost one Helix Get Streams request per tick.
	Limit int
	// Window is how far back clips count toward a broadcaster's ranking.
	Window time.Duration
	// Refresh is how often the candidate list is recomputed.
	Refresh time.Duration
}

// LiveStatusScheduler manages periodic live status updates
type LiveStatusScheduler struct {
	liveStatusService LiveStatusServiceInterface
	broadcasterRepo   BroadcasterRepositoryInterface
	interval          time.Duration
	candidates        LiveStatusCandidateConfig
	stopChan          chan struct{}
	stopOnce          sync.Once

	candidateIDs      []string
	candidatesFetched time.Time
	now               func() time.Time
}

// NewLiveStatusScheduler creates a new live status scheduler
func NewLiveStatusScheduler(
	liveStatusService LiveStatusServiceInterface,
	broadcasterRepo BroadcasterRepositoryInterface,
	intervalSeconds int,
	candidates LiveStatusCandidateConfig,
) *LiveStatusScheduler {
	if candidates.Refresh <= 0 {
		candidates.Refresh = 10 * time.Minute
	}
	if candidates.Window <= 0 {
		candidates.Window = 30 * 24 * time.Hour
	}
	return &LiveStatusScheduler{
		liveStatusService: liveStatusService,
		broadcasterRepo:   broadcasterRepo,
		interval:          time.Duration(intervalSeconds) * time.Second,
		candidates:        candidates,
		stopChan:          make(chan struct{}),
		now:               time.Now,
	}
}

// Start begins the periodic live status update process
func (s *LiveStatusScheduler) Start(ctx context.Context) {
	utils.Info("Starting live status scheduler", map[string]interface{}{
		"scheduler":       liveStatusSchedulerName,
		"interval":        s.interval.String(),
		"candidate_limit": s.candidates.Limit,
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

// candidateBroadcasterIDs returns the cached public-site candidates,
// recomputing them every candidates.Refresh. On failure the previous list is
// kept so a slow database does not empty the live list.
func (s *LiveStatusScheduler) candidateBroadcasterIDs(ctx context.Context) []string {
	if s.candidates.Limit <= 0 {
		return nil
	}
	now := s.now()
	if !s.candidatesFetched.IsZero() && now.Sub(s.candidatesFetched) < s.candidates.Refresh {
		return s.candidateIDs
	}
	ids, err := s.broadcasterRepo.GetLiveStatusCandidateBroadcasterIDs(ctx, s.candidates.Limit, s.candidates.Window)
	if err != nil {
		utils.Error("Failed to get live status candidate broadcasters", err, map[string]interface{}{
			"scheduler": liveStatusSchedulerName,
			"cached":    len(s.candidateIDs),
		})
		return s.candidateIDs
	}
	s.candidateIDs = ids
	s.candidatesFetched = now
	return ids
}

// mergeBroadcasterIDs returns followed IDs, then candidates, without
// duplicates or empty IDs.
func mergeBroadcasterIDs(followed, candidates []string) []string {
	seen := make(map[string]struct{}, len(followed)+len(candidates))
	merged := make([]string, 0, len(followed)+len(candidates))
	for _, list := range [][]string{followed, candidates} {
		for _, id := range list {
			if id == "" {
				continue
			}
			if _, dup := seen[id]; dup {
				continue
			}
			seen[id] = struct{}{}
			merged = append(merged, id)
		}
	}
	return merged
}

// updateLiveStatuses executes a live status update operation
func (s *LiveStatusScheduler) updateLiveStatuses(ctx context.Context) {
	utils.Info("Starting scheduled live status update", map[string]interface{}{
		"scheduler": liveStatusSchedulerName,
	})
	startTime := time.Now()

	// Followed broadcasters drive notifications; candidates keep the public
	// live list and live badges accurate for popular broadcasters.
	followedIDs, err := s.broadcasterRepo.GetAllFollowedBroadcasterIDs(ctx)
	if err != nil {
		utils.Error("Failed to get followed broadcasters", err, map[string]interface{}{
			"scheduler": liveStatusSchedulerName,
		})
		followedIDs = nil
	}
	candidateIDs := s.candidateBroadcasterIDs(ctx)
	broadcasterIDs := mergeBroadcasterIDs(followedIDs, candidateIDs)

	if len(broadcasterIDs) == 0 {
		utils.Info("No broadcasters to check", map[string]interface{}{
			"scheduler": liveStatusSchedulerName,
		})
		return
	}

	utils.Info("Checking live status for broadcasters", map[string]interface{}{
		"scheduler":  liveStatusSchedulerName,
		"count":      len(broadcasterIDs),
		"followed":   len(followedIDs),
		"candidates": len(candidateIDs),
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
