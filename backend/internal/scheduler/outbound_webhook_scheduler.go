package scheduler

import (
	"context"
	"sync"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

// OutboundWebhookScheduler handles periodic processing of webhook deliveries
type OutboundWebhookScheduler struct {
	webhookService *services.OutboundWebhookService
	interval       time.Duration
	batchSize      int
	stopChan       chan struct{}
	stopOnce       sync.Once
}

// NewOutboundWebhookScheduler creates a new outbound webhook scheduler
func NewOutboundWebhookScheduler(webhookService *services.OutboundWebhookService, interval time.Duration, batchSize int) *OutboundWebhookScheduler {
	return &OutboundWebhookScheduler{
		webhookService: webhookService,
		interval:       interval,
		batchSize:      batchSize,
		stopChan:       make(chan struct{}),
	}
}

// Start starts the webhook delivery scheduler
func (s *OutboundWebhookScheduler) Start(ctx context.Context) {
	utils.Info("Starting outbound webhook scheduler", map[string]interface{}{
		"interval":   s.interval.String(),
		"batch_size": s.batchSize,
		"scheduler":  "outbound_webhook",
	})

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Process immediately on start
	s.processDeliveries(ctx)

	for {
		select {
		case <-ticker.C:
			s.processDeliveries(ctx)
		case <-s.stopChan:
			utils.Info("Stopping outbound webhook scheduler", map[string]interface{}{
				"scheduler": "outbound_webhook",
			})
			return
		case <-ctx.Done():
			utils.Info("Outbound webhook scheduler stopped due to context cancellation", map[string]interface{}{
				"scheduler": "outbound_webhook",
			})
			return
		}
	}
}

// Stop stops the webhook delivery scheduler in a thread-safe manner
func (s *OutboundWebhookScheduler) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopChan)
	})
}

// processDeliveries processes pending webhook deliveries
func (s *OutboundWebhookScheduler) processDeliveries(ctx context.Context) {
	if err := s.webhookService.ProcessPendingDeliveries(ctx, s.batchSize); err != nil {
		utils.Error("Error processing outbound webhook deliveries", err, map[string]interface{}{
			"batch_size": s.batchSize,
			"scheduler":  "outbound_webhook",
		})
	}
}
