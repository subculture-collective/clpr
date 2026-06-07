package scheduler

import (
	"context"
	"sync"
	"time"

	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

// WebhookRetryServiceInterface defines the interface required by the webhook retry scheduler
type WebhookRetryServiceInterface interface {
	ProcessPendingRetries(ctx context.Context, batchSize int) error
}

// WebhookRetryScheduler manages periodic webhook retry processing
type WebhookRetryScheduler struct {
	webhookRetryService WebhookRetryServiceInterface
	interval            time.Duration
	batchSize           int
	stopChan            chan struct{}
	stopOnce            sync.Once
}

// NewWebhookRetryScheduler creates a new webhook retry scheduler
func NewWebhookRetryScheduler(
	webhookRetryService WebhookRetryServiceInterface,
	intervalMinutes int,
	batchSize int,
) *WebhookRetryScheduler {
	return &WebhookRetryScheduler{
		webhookRetryService: webhookRetryService,
		interval:            time.Duration(intervalMinutes) * time.Minute,
		batchSize:           batchSize,
		stopChan:            make(chan struct{}),
	}
}

// Start begins the periodic webhook retry processing
func (s *WebhookRetryScheduler) Start(ctx context.Context) {
	utils.Info("Starting webhook retry scheduler", map[string]interface{}{
		"interval":   s.interval.String(),
		"batch_size": s.batchSize,
		"scheduler":  "webhook_retry",
	})

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Run initial processing
	s.processRetries(ctx)

	for {
		select {
		case <-ticker.C:
			s.processRetries(ctx)
		case <-s.stopChan:
			utils.Info("Webhook retry scheduler stopped", map[string]interface{}{
				"scheduler": "webhook_retry",
			})
			return
		case <-ctx.Done():
			utils.Info("Webhook retry scheduler stopped due to context cancellation", map[string]interface{}{
				"scheduler": "webhook_retry",
			})
			return
		}
	}
}

// Stop gracefully stops the scheduler
func (s *WebhookRetryScheduler) Stop() {
	s.stopOnce.Do(func() {
		utils.Info("Stopping webhook retry scheduler", map[string]interface{}{
			"scheduler": "webhook_retry",
		})
		close(s.stopChan)
	})
}

// processRetries processes pending webhook retries
func (s *WebhookRetryScheduler) processRetries(ctx context.Context) {
	utils.Info("Processing webhook retries", map[string]interface{}{
		"batch_size": s.batchSize,
		"scheduler":  "webhook_retry",
	})

	if err := s.webhookRetryService.ProcessPendingRetries(ctx, s.batchSize); err != nil {
		utils.Error("Error processing webhook retries", err, map[string]interface{}{
			"batch_size": s.batchSize,
			"scheduler":  "webhook_retry",
		})
	}
}
