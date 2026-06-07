package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/pkg/metrics"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

const (
	reputationSchedulerName = "reputation"
	reputationJobName       = "reputation_tasks"
)

// ReputationServiceInterface defines the interface required by the scheduler
type ReputationServiceInterface interface {
	CheckAndAwardBadges(ctx context.Context, userID uuid.UUID) ([]string, error)
	UpdateUserStats(ctx context.Context, userID uuid.UUID) error
}

// ReputationScheduler manages periodic reputation-related tasks
type ReputationScheduler struct {
	reputationService ReputationServiceInterface
	userRepo          UserRepositoryInterface
	interval          time.Duration
	stopChan          chan struct{}
	stopOnce          sync.Once
}

// UserRepositoryInterface defines the interface for getting users
type UserRepositoryInterface interface {
	GetAllActiveUserIDs(ctx context.Context) ([]uuid.UUID, error)
}

// NewReputationScheduler creates a new reputation scheduler
func NewReputationScheduler(
	reputationService ReputationServiceInterface,
	userRepo UserRepositoryInterface,
	intervalHours int,
) *ReputationScheduler {
	return &ReputationScheduler{
		reputationService: reputationService,
		userRepo:          userRepo,
		interval:          time.Duration(intervalHours) * time.Hour,
		stopChan:          make(chan struct{}),
	}
}

// Start begins the periodic reputation tasks
func (s *ReputationScheduler) Start(ctx context.Context) {
	utils.Info("Starting reputation scheduler", map[string]interface{}{
		"scheduler": reputationSchedulerName,
		"interval":  s.interval.String(),
	})

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Run initial tasks
	s.runTasks(ctx)

	for {
		select {
		case <-ticker.C:
			s.runTasks(ctx)
		case <-s.stopChan:
			utils.Info("Reputation scheduler stopped", map[string]interface{}{
				"scheduler": reputationSchedulerName,
			})
			return
		case <-ctx.Done():
			utils.Info("Reputation scheduler stopped due to context cancellation", map[string]interface{}{
				"scheduler": reputationSchedulerName,
			})
			return
		}
	}
}

// Stop stops the scheduler in a thread-safe manner
func (s *ReputationScheduler) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopChan)
	})
}

// runTasks executes reputation maintenance tasks
func (s *ReputationScheduler) runTasks(ctx context.Context) {
	utils.Info("Starting scheduled reputation tasks", map[string]interface{}{
		"scheduler": reputationSchedulerName,
		"job":       reputationJobName,
	})
	startTime := time.Now()

	// Get all active user IDs
	userIDs, err := s.userRepo.GetAllActiveUserIDs(ctx)
	if err != nil {
		utils.Error("Failed to get active users", err, map[string]interface{}{
			"scheduler": reputationSchedulerName,
			"job":       reputationJobName,
		})
		metrics.JobExecutionTotal.WithLabelValues(reputationJobName, "failed").Inc()
		metrics.JobExecutionDuration.WithLabelValues(reputationJobName).Observe(time.Since(startTime).Seconds())
		return
	}

	const workerCount = 20
	type result struct {
		badgesAwarded int
		statsUpdated  int
		errors        int
	}

	userCh := make(chan uuid.UUID)
	resultCh := make(chan result)
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for userID := range userCh {
				res := result{}

				// Check and award badges
				badges, err := s.reputationService.CheckAndAwardBadges(ctx, userID)
				if err != nil {
					utils.Error("Failed to check badges for user", err, map[string]interface{}{
						"scheduler": reputationSchedulerName,
						"user_id":   userID,
						"job":       reputationJobName,
					})
					res.errors++
				} else if len(badges) > 0 {
					res.badgesAwarded += len(badges)
					utils.Info("Awarded badges to user", map[string]interface{}{
						"scheduler":   reputationSchedulerName,
						"user_id":     userID,
						"badge_count": len(badges),
						"badges":      badges,
						"job":         reputationJobName,
					})
				}

				// Update user stats
				err = s.reputationService.UpdateUserStats(ctx, userID)
				if err != nil {
					utils.Error("Failed to update stats for user", err, map[string]interface{}{
						"scheduler": reputationSchedulerName,
						"user_id":   userID,
						"job":       reputationJobName,
					})
					res.errors++
				} else {
					res.statsUpdated++
				}

				resultCh <- res
			}
		}()
	}

	// Feed users to workers
	go func() {
		for _, userID := range userIDs {
			userCh <- userID
		}
		close(userCh)
	}()

	// Collect results
	badgesAwarded := 0
	statsUpdated := 0
	errors := 0
	for i := 0; i < len(userIDs); i++ {
		res := <-resultCh
		badgesAwarded += res.badgesAwarded
		statsUpdated += res.statsUpdated
		errors += res.errors
	}

	// Wait for all workers to finish
	wg.Wait()
	close(resultCh)
	duration := time.Since(startTime)

	// Record metrics
	metrics.JobExecutionDuration.WithLabelValues(reputationJobName).Observe(duration.Seconds())
	metrics.JobItemsProcessed.WithLabelValues(reputationJobName, "success").Add(float64(statsUpdated))

	if errors > 0 {
		metrics.JobItemsProcessed.WithLabelValues(reputationJobName, "failed").Add(float64(errors))
	}

	// Consider the job successful if majority of operations succeeded
	totalOperations := statsUpdated + errors
	if totalOperations == 0 {
		// No operations processed, treat as successful
		metrics.JobExecutionTotal.WithLabelValues(reputationJobName, "success").Inc()
		metrics.JobLastSuccessTimestamp.WithLabelValues(reputationJobName).Set(float64(time.Now().Unix()))
	} else {
		failureRatio := float64(errors) / float64(totalOperations)
		if failureRatio > 0.5 {
			// More than 50% failures - mark as failed
			metrics.JobExecutionTotal.WithLabelValues(reputationJobName, "failed").Inc()
		} else {
			// Majority succeeded - mark as success
			metrics.JobExecutionTotal.WithLabelValues(reputationJobName, "success").Inc()
			metrics.JobLastSuccessTimestamp.WithLabelValues(reputationJobName).Set(float64(time.Now().Unix()))
		}
	}

	utils.Info("Reputation tasks completed", map[string]interface{}{
		"scheduler":      reputationSchedulerName,
		"job":            reputationJobName,
		"users":          len(userIDs),
		"badges_awarded": badgesAwarded,
		"stats_updated":  statsUpdated,
		"errors":         errors,
		"duration":       duration.String(),
	})
}
