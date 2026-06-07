package scheduler

import (
	"context"
	"sync"
	"time"

	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

const dunningSchedulerName = "dunning"

// DunningServiceInterface defines the interface required by the dunning scheduler
type DunningServiceInterface interface {
	ProcessExpiredGracePeriods(ctx context.Context) error
	SendGracePeriodWarnings(ctx context.Context) error
}

// DunningScheduler manages periodic dunning processing tasks
type DunningScheduler struct {
	dunningService           DunningServiceInterface
	gracePeriodCheckInterval time.Duration
	warningCheckInterval     time.Duration
	stopChan                 chan struct{}
	stopOnce                 sync.Once
	wg                       sync.WaitGroup
}

// NewDunningScheduler creates a new dunning scheduler
func NewDunningScheduler(
	dunningService DunningServiceInterface,
	gracePeriodCheckMinutes int,
	warningCheckMinutes int,
) *DunningScheduler {
	// Default intervals if not specified
	if gracePeriodCheckMinutes <= 0 {
		gracePeriodCheckMinutes = 60 // Check hourly for expired grace periods
	}
	if warningCheckMinutes <= 0 {
		warningCheckMinutes = 1440 // Check daily for warnings (24 hours)
	}

	return &DunningScheduler{
		dunningService:           dunningService,
		gracePeriodCheckInterval: time.Duration(gracePeriodCheckMinutes) * time.Minute,
		warningCheckInterval:     time.Duration(warningCheckMinutes) * time.Minute,
		stopChan:                 make(chan struct{}),
	}
}

// Start begins the periodic dunning processing
func (s *DunningScheduler) Start(ctx context.Context) {
	utils.Info("Starting dunning scheduler", map[string]interface{}{
		"scheduler":                   dunningSchedulerName,
		"grace_period_check_interval": s.gracePeriodCheckInterval.String(),
		"warning_check_interval":      s.warningCheckInterval.String(),
	})

	// Start grace period expiry checker
	s.wg.Add(1)
	go s.runGracePeriodChecker(ctx)

	// Start warning sender
	s.wg.Add(1)
	go s.runWarningChecker(ctx)
}

// runGracePeriodChecker runs the grace period expiry check loop
func (s *DunningScheduler) runGracePeriodChecker(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(s.gracePeriodCheckInterval)
	defer ticker.Stop()

	// Run initial check
	s.processExpiredGracePeriods(ctx)

	for {
		select {
		case <-ticker.C:
			s.processExpiredGracePeriods(ctx)
		case <-s.stopChan:
			utils.Info("Grace period checker stopped", map[string]interface{}{
				"scheduler": dunningSchedulerName,
			})
			return
		case <-ctx.Done():
			utils.Info("Grace period checker stopped due to context cancellation", map[string]interface{}{
				"scheduler": dunningSchedulerName,
			})
			return
		}
	}
}

// runWarningChecker runs the grace period warning loop
func (s *DunningScheduler) runWarningChecker(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(s.warningCheckInterval)
	defer ticker.Stop()

	// Run initial check
	s.sendGracePeriodWarnings(ctx)

	for {
		select {
		case <-ticker.C:
			s.sendGracePeriodWarnings(ctx)
		case <-s.stopChan:
			utils.Info("Warning checker stopped", map[string]interface{}{
				"scheduler": dunningSchedulerName,
			})
			return
		case <-ctx.Done():
			utils.Info("Warning checker stopped due to context cancellation", map[string]interface{}{
				"scheduler": dunningSchedulerName,
			})
			return
		}
	}
}

// Stop gracefully stops the scheduler
func (s *DunningScheduler) Stop() {
	s.stopOnce.Do(func() {
		utils.Info("Stopping dunning scheduler", map[string]interface{}{
			"scheduler": dunningSchedulerName,
		})
		close(s.stopChan)

		// Wait for goroutines to finish with timeout
		done := make(chan struct{})
		go func() {
			s.wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			utils.Info("Dunning scheduler stopped successfully", map[string]interface{}{
				"scheduler": dunningSchedulerName,
			})
		case <-time.After(30 * time.Second):
			utils.Warn("Dunning scheduler stop timed out", map[string]interface{}{
				"scheduler": dunningSchedulerName,
			})
		}
	})
}

// processExpiredGracePeriods processes subscriptions with expired grace periods
func (s *DunningScheduler) processExpiredGracePeriods(ctx context.Context) {
	utils.Info("Processing expired grace periods", map[string]interface{}{
		"scheduler": dunningSchedulerName,
	})

	if err := s.dunningService.ProcessExpiredGracePeriods(ctx); err != nil {
		utils.Error("Error processing expired grace periods", err, map[string]interface{}{
			"scheduler": dunningSchedulerName,
		})
	} else {
		utils.Info("Successfully processed expired grace periods", map[string]interface{}{
			"scheduler": dunningSchedulerName,
		})
	}
}

// sendGracePeriodWarnings sends warnings to users approaching grace period expiry
func (s *DunningScheduler) sendGracePeriodWarnings(ctx context.Context) {
	utils.Info("Sending grace period warnings", map[string]interface{}{
		"scheduler": dunningSchedulerName,
	})

	if err := s.dunningService.SendGracePeriodWarnings(ctx); err != nil {
		utils.Error("Error sending grace period warnings", err, map[string]interface{}{
			"scheduler": dunningSchedulerName,
		})
	} else {
		utils.Info("Successfully sent grace period warnings", map[string]interface{}{
			"scheduler": dunningSchedulerName,
		})
	}
}
