package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/stripe/stripe-go/v81"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

const webhookRetryComponent = "webhook_retry"

// WebhookRetryService handles processing of webhook retries from the queue
type WebhookRetryService struct {
	webhookRepo         *repository.WebhookRepository
	subscriptionService *SubscriptionService
}

// NewWebhookRetryService creates a new webhook retry service
func NewWebhookRetryService(
	webhookRepo *repository.WebhookRepository,
	subscriptionService *SubscriptionService,
) *WebhookRetryService {
	return &WebhookRetryService{
		webhookRepo:         webhookRepo,
		subscriptionService: subscriptionService,
	}
}

// ProcessPendingRetries processes webhook events that are ready for retry
func (s *WebhookRetryService) ProcessPendingRetries(ctx context.Context, batchSize int) error {
	utils.Info("Processing pending webhook retries", map[string]interface{}{
		"component":  webhookRetryComponent,
		"batch_size": batchSize,
	})

	// Get pending retries
	items, err := s.webhookRepo.GetPendingRetries(ctx, batchSize)
	if err != nil {
		return fmt.Errorf("failed to get pending retries: %w", err)
	}

	if len(items) == 0 {
		utils.Info("No pending webhook retries found", map[string]interface{}{
			"component": webhookRetryComponent,
		})
		return nil
	}

	utils.Info("Found pending webhook retries", map[string]interface{}{
		"component": webhookRetryComponent,
		"count":     len(items),
	})

	for _, item := range items {
		if err := s.processRetry(ctx, item); err != nil {
			utils.Error("Failed to process webhook retry", err, map[string]interface{}{
				"component":  webhookRetryComponent,
				"event_id":   item.StripeEventID,
				"event_type": item.EventType,
			})
		}
	}

	return nil
}

// processRetry processes a single retry attempt
func (s *WebhookRetryService) processRetry(ctx context.Context, item *models.WebhookRetryQueue) error {
	utils.Info("Processing webhook retry", map[string]interface{}{
		"component":   webhookRetryComponent,
		"event_id":    item.StripeEventID,
		"event_type":  item.EventType,
		"retry_count": item.RetryCount + 1,
		"max_retries": item.MaxRetries,
	})

	// Parse the payload into a Stripe event
	var event stripe.Event
	if err := json.Unmarshal([]byte(item.Payload), &event); err != nil {
		errMsg := fmt.Sprintf("failed to unmarshal event payload: %v", err)
		utils.Error("Failed to unmarshal webhook payload", err, map[string]interface{}{
			"component":  webhookRetryComponent,
			"event_id":   item.StripeEventID,
			"event_type": item.EventType,
		})

		// This is a permanent error, move to DLQ
		if err := s.webhookRepo.MoveToDeadLetterQueue(ctx, item, errMsg); err != nil {
			utils.Error("Failed to move webhook event to DLQ", err, map[string]interface{}{
				"component":  webhookRetryComponent,
				"event_id":   item.StripeEventID,
				"event_type": item.EventType,
			})
		}
		return fmt.Errorf("%s", errMsg)
	}

	// Process the event using the subscription service
	err := s.subscriptionService.processWebhookWithRetry(ctx, event)

	if err != nil {
		// Increment retry count
		newRetryCount := item.RetryCount + 1

		// Check if we've exhausted retries
		if newRetryCount >= item.MaxRetries {
			utils.Warn("Max retries reached for webhook event, moving to DLQ", map[string]interface{}{
				"component":   webhookRetryComponent,
				"event_id":    item.StripeEventID,
				"event_type":  item.EventType,
				"max_retries": item.MaxRetries,
			})
			if dlqErr := s.webhookRepo.MoveToDeadLetterQueue(ctx, item, err.Error()); dlqErr != nil {
				utils.Error("Failed to move webhook event to DLQ", dlqErr, map[string]interface{}{
					"component":  webhookRetryComponent,
					"event_id":   item.StripeEventID,
					"event_type": item.EventType,
				})
			}
			return err
		}

		// Calculate next retry time with exponential backoff
		nextRetry := s.calculateNextRetry(newRetryCount)
		utils.Warn("Webhook retry failed, scheduling next retry", map[string]interface{}{
			"component":   webhookRetryComponent,
			"event_id":    item.StripeEventID,
			"event_type":  item.EventType,
			"retry_count": newRetryCount,
			"next_retry":  nextRetry,
		})

		// Update retry queue item
		if updateErr := s.webhookRepo.UpdateRetryQueueItem(ctx, item.ID, newRetryCount, &nextRetry, err.Error()); updateErr != nil {
			utils.Error("Failed to update webhook retry queue item", updateErr, map[string]interface{}{
				"component":   webhookRetryComponent,
				"event_id":    item.StripeEventID,
				"event_type":  item.EventType,
				"retry_count": newRetryCount,
			})
			// Attempt to move to DLQ to prevent rapid retry loops
			dlqErr := s.webhookRepo.MoveToDeadLetterQueue(ctx, item, fmt.Sprintf("Failed to update retry queue item: %v; original error: %v", updateErr, err))
			if dlqErr != nil {
				utils.Error("Failed to move webhook event to DLQ after update failure", dlqErr, map[string]interface{}{
					"component":  webhookRetryComponent,
					"event_id":   item.StripeEventID,
					"event_type": item.EventType,
				})
			}
			return fmt.Errorf("failed to update retry queue item for event %s: %w (original error: %v)", item.StripeEventID, updateErr, err)
		}

		return err
	}

	// Success! Remove from retry queue
	utils.Info("Successfully processed webhook event, removing from queue", map[string]interface{}{
		"component":  webhookRetryComponent,
		"event_id":   item.StripeEventID,
		"event_type": item.EventType,
	})
	if err := s.webhookRepo.RemoveFromRetryQueue(ctx, item.StripeEventID); err != nil {
		utils.Error("Failed to remove webhook event from retry queue", err, map[string]interface{}{
			"component":  webhookRetryComponent,
			"event_id":   item.StripeEventID,
			"event_type": item.EventType,
		})
		return fmt.Errorf("successfully processed event %s but failed to remove from retry queue: %w", item.StripeEventID, err)
	}

	return nil
}

// calculateNextRetry calculates the next retry time using exponential backoff
// Base delay: 30 seconds
// Formula: base * 2^(retryCount) with max of 1 hour
func (s *WebhookRetryService) calculateNextRetry(retryCount int) time.Time {
	baseDelay := 30 * time.Second
	maxDelay := 1 * time.Hour

	// Calculate exponential backoff: 30s, 1m, 2m, 4m, etc.
	delay := time.Duration(float64(baseDelay) * math.Pow(2, float64(retryCount)))

	// Cap at max delay
	if delay > maxDelay {
		delay = maxDelay
	}

	return time.Now().Add(delay)
}

// GetRetryQueueStats returns statistics about the retry queue
func (s *WebhookRetryService) GetRetryQueueStats(ctx context.Context) (map[string]interface{}, error) {
	// Get pending retries count efficiently
	pendingCount, err := s.webhookRepo.CountPendingRetries(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count pending retries: %w", err)
	}

	// Get DLQ count
	dlqCount, err := s.webhookRepo.CountDeadLetterQueueItems(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count DLQ items: %w", err)
	}

	// Update Prometheus metrics
	webhookRetryQueueSize.Set(float64(pendingCount))
	webhookDeadLetterQueueSize.Set(float64(dlqCount))

	stats := map[string]interface{}{
		"pending_retries": pendingCount,
		"dlq_items":       dlqCount,
		"timestamp":       time.Now(),
	}

	return stats, nil
}
