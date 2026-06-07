package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

const (
	webhookDLQComponent      = "webhook_dlq"
	webhookOutboundComponent = "webhook_outbound"
)

// OutboundWebhookService handles webhook delivery to third-party endpoints
type OutboundWebhookService struct {
	webhookRepo *repository.OutboundWebhookRepository
	httpClient  *http.Client
}

// NewOutboundWebhookService creates a new outbound webhook service
func NewOutboundWebhookService(webhookRepo *repository.OutboundWebhookRepository) *OutboundWebhookService {
	return &OutboundWebhookService{
		webhookRepo: webhookRepo,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CreateSubscription creates a new webhook subscription
func (s *OutboundWebhookService) CreateSubscription(ctx context.Context, userID uuid.UUID, req *models.CreateWebhookSubscriptionRequest) (*models.WebhookSubscription, error) {
	// Validate URL for SSRF protection
	if err := s.validateURL(req.URL); err != nil {
		return nil, err
	}

	// Validate events
	if err := s.validateEvents(req.Events); err != nil {
		return nil, err
	}

	// Generate a secure random secret for HMAC signing
	secret, err := s.generateSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to generate secret: %w", err)
	}

	subscription := &models.WebhookSubscription{
		ID:          uuid.New(),
		UserID:      userID,
		URL:         req.URL,
		Secret:      secret,
		Events:      req.Events,
		IsActive:    true,
		Description: req.Description,
	}

	if err := s.webhookRepo.CreateSubscription(ctx, subscription); err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	// Update metrics
	s.updateActiveSubscriptionsMetric(ctx)

	return subscription, nil
}

// GetSubscriptionByID retrieves a webhook subscription by ID
func (s *OutboundWebhookService) GetSubscriptionByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.WebhookSubscription, error) {
	subscription, err := s.webhookRepo.GetSubscriptionByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Ensure the subscription belongs to the user
	if subscription.UserID != userID {
		return nil, fmt.Errorf("subscription not found")
	}

	return subscription, nil
}

// GetSubscriptionsByUserID retrieves all webhook subscriptions for a user
func (s *OutboundWebhookService) GetSubscriptionsByUserID(ctx context.Context, userID uuid.UUID) ([]*models.WebhookSubscription, error) {
	return s.webhookRepo.GetSubscriptionsByUserID(ctx, userID)
}

// UpdateSubscription updates a webhook subscription
func (s *OutboundWebhookService) UpdateSubscription(ctx context.Context, id uuid.UUID, userID uuid.UUID, req *models.UpdateWebhookSubscriptionRequest) error {
	// Verify ownership
	subscription, err := s.webhookRepo.GetSubscriptionByID(ctx, id)
	if err != nil {
		return err
	}

	if subscription.UserID != userID {
		return fmt.Errorf("subscription not found")
	}

	// Validate URL if provided
	if req.URL != nil {
		if err := s.validateURL(*req.URL); err != nil {
			return err
		}
	}

	// Validate events if provided
	var eventsToUpdate []string
	if req.Events != nil {
		// Check if empty array was explicitly provided (invalid)
		if len(req.Events) == 0 {
			return fmt.Errorf("events array cannot be empty; omit the field to keep current events")
		}
		if err := s.validateEvents(req.Events); err != nil {
			return err
		}
		eventsToUpdate = req.Events
	} else {
		// nil means don't update events
		eventsToUpdate = nil
	}

	return s.webhookRepo.UpdateSubscription(ctx, id, req.URL, eventsToUpdate, req.IsActive, req.Description)
}

// DeleteSubscription deletes a webhook subscription
func (s *OutboundWebhookService) DeleteSubscription(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// Verify ownership
	subscription, err := s.webhookRepo.GetSubscriptionByID(ctx, id)
	if err != nil {
		return err
	}

	if subscription.UserID != userID {
		return fmt.Errorf("subscription not found")
	}

	if err := s.webhookRepo.DeleteSubscription(ctx, id); err != nil {
		return err
	}

	// Update metrics
	s.updateActiveSubscriptionsMetric(ctx)

	return nil
}

// RegenerateSecret regenerates the webhook secret for a subscription
func (s *OutboundWebhookService) RegenerateSecret(ctx context.Context, id uuid.UUID, userID uuid.UUID) (string, error) {
	// Verify ownership
	subscription, err := s.webhookRepo.GetSubscriptionByID(ctx, id)
	if err != nil {
		return "", err
	}

	if subscription.UserID != userID {
		return "", fmt.Errorf("subscription not found")
	}

	// Generate new secret
	newSecret, err := s.generateSecret()
	if err != nil {
		return "", fmt.Errorf("failed to generate secret: %w", err)
	}

	// Update subscription with new secret
	if err := s.webhookRepo.UpdateSubscriptionSecret(ctx, id, newSecret); err != nil {
		return "", fmt.Errorf("failed to update subscription: %w", err)
	}

	return newSecret, nil
}

// TriggerEvent triggers a webhook event for all subscribed endpoints
func (s *OutboundWebhookService) TriggerEvent(ctx context.Context, eventType string, eventID uuid.UUID, data map[string]interface{}) error {
	// Get all active subscriptions for this event
	subscriptions, err := s.webhookRepo.GetActiveSubscriptionsByEvent(ctx, eventType)
	if err != nil {
		return fmt.Errorf("failed to get subscriptions: %w", err)
	}

	if len(subscriptions) == 0 {
		utils.Info("No active webhook subscriptions for event", map[string]interface{}{
			"component":  webhookOutboundComponent,
			"event_type": eventType,
		})
		return nil
	}

	utils.Info("Triggering webhook event", map[string]interface{}{
		"component":  webhookOutboundComponent,
		"event_type": eventType,
		"count":      len(subscriptions),
	})

	// Create payload
	payload := models.WebhookEventPayload{
		Event:     eventType,
		Timestamp: time.Now(),
		Data:      data,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Queue delivery for each subscription
	for _, subscription := range subscriptions {
		delivery := &models.WebhookDelivery{
			ID:             uuid.New(),
			SubscriptionID: subscription.ID,
			EventType:      eventType,
			EventID:        eventID,
			Payload:        string(payloadJSON),
			Status:         "pending",
			AttemptCount:   0,
			MaxAttempts:    5,
			NextAttemptAt:  ptrTime(time.Now()),
		}

		if err := s.webhookRepo.CreateDelivery(ctx, delivery); err != nil {
			utils.Error("Failed to create webhook delivery", err, map[string]interface{}{
				"component":       webhookOutboundComponent,
				"subscription_id": subscription.ID,
				"event_type":      eventType,
			})
			continue
		}

		utils.Info("Queued webhook delivery", map[string]interface{}{
			"component":       webhookOutboundComponent,
			"delivery_id":     delivery.ID,
			"subscription_id": subscription.ID,
			"event_type":      eventType,
		})
	}

	return nil
}

// ProcessPendingDeliveries processes pending webhook deliveries
func (s *OutboundWebhookService) ProcessPendingDeliveries(ctx context.Context, batchSize int) error {
	deliveries, err := s.webhookRepo.GetPendingDeliveries(ctx, batchSize)
	if err != nil {
		return fmt.Errorf("failed to get pending deliveries: %w", err)
	}

	if len(deliveries) == 0 {
		return nil
	}

	utils.Info("Processing pending webhook deliveries", map[string]interface{}{
		"component": webhookOutboundComponent,
		"count":     len(deliveries),
	})

	for _, delivery := range deliveries {
		if err := s.processDelivery(ctx, delivery); err != nil {
			utils.Error("Failed to process webhook delivery", err, map[string]interface{}{
				"component":   webhookOutboundComponent,
				"delivery_id": delivery.ID,
				"event_type":  delivery.EventType,
			})
		}
	}

	return nil
}

// processDelivery processes a single webhook delivery
func (s *OutboundWebhookService) processDelivery(ctx context.Context, delivery *models.WebhookDelivery) error {
	// Track delivery start time for metrics
	startTime := time.Now()

	// Get subscription details
	subscription, err := s.webhookRepo.GetSubscriptionByID(ctx, delivery.SubscriptionID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	subscriptionIDStr := subscription.ID.String()

	if !subscription.IsActive {
		utils.Warn("Webhook subscription is inactive, skipping delivery", map[string]interface{}{
			"component":       webhookOutboundComponent,
			"subscription_id": subscription.ID,
			"delivery_id":     delivery.ID,
			"event_type":      delivery.EventType,
		})
		// Mark as failed since subscription is inactive
		webhookDeliveryTotal.WithLabelValues(delivery.EventType, "failed").Inc()
		webhookDeliveryDuration.WithLabelValues(delivery.EventType, "failed").Observe(time.Since(startTime).Seconds())
		webhookSubscriptionHealth.WithLabelValues(subscriptionIDStr, "failed").Inc()
		webhookConsecutiveFailures.WithLabelValues(subscriptionIDStr, delivery.EventType).Inc()
		return s.webhookRepo.UpdateDeliveryFailure(ctx, delivery.ID, nil, "subscription is inactive", nil)
	}

	utils.Info("Delivering webhook", map[string]interface{}{
		"component":       webhookOutboundComponent,
		"subscription_id": subscription.ID,
		"delivery_id":     delivery.ID,
		"event_type":      delivery.EventType,
		"attempt":         delivery.AttemptCount + 1,
		"max_attempts":    delivery.MaxAttempts,
	})

	// Generate signature
	signature := s.generateSignature(delivery.Payload, subscription.Secret)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", subscription.URL, bytes.NewBufferString(delivery.Payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", signature)
	req.Header.Set("X-Webhook-Event", delivery.EventType)
	req.Header.Set("X-Webhook-Delivery-ID", delivery.ID.String())
	req.Header.Set("User-Agent", "Clipper-Webhooks/1.0")

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		// Network error - schedule retry or move to DLQ
		errMsg := fmt.Sprintf("network error: %v", err)

		// Track retry rate metric (only for actual retries, not the initial attempt)
		if delivery.AttemptCount > 0 {
			webhookRetryRate.WithLabelValues(delivery.EventType, strconv.Itoa(delivery.AttemptCount+1)).Inc()
		}

		// Check if this is the final attempt
		if delivery.AttemptCount+1 >= delivery.MaxAttempts {
			utils.Warn("Max retries reached for webhook delivery (network error), moving to DLQ", map[string]interface{}{
				"component":    webhookOutboundComponent,
				"delivery_id":  delivery.ID,
				"event_type":   delivery.EventType,
				"attempt":      delivery.AttemptCount + 1,
				"max_attempts": delivery.MaxAttempts,
			})

			// Update delivery with final failure status
			if updateErr := s.webhookRepo.UpdateDeliveryFailure(ctx, delivery.ID, nil, errMsg, nil); updateErr != nil {
				utils.Error("Failed to update webhook delivery failure", updateErr, map[string]interface{}{
					"component":   webhookOutboundComponent,
					"delivery_id": delivery.ID,
					"event_type":  delivery.EventType,
				})
				return fmt.Errorf("failed to update delivery failure before DLQ: %w", updateErr)
			}

			// Get updated delivery to move to DLQ
			updatedDelivery, getErr := s.webhookRepo.GetDeliveryByID(ctx, delivery.ID)
			if getErr != nil {
				utils.Error("Failed to get webhook delivery for DLQ", getErr, map[string]interface{}{
					"component":   webhookOutboundComponent,
					"delivery_id": delivery.ID,
					"event_type":  delivery.EventType,
				})
				return fmt.Errorf("failed to get delivery for DLQ: %w", getErr)
			}

			// Move to dead-letter queue
			if dlqErr := s.webhookRepo.MoveDeliveryToDeadLetterQueue(ctx, updatedDelivery); dlqErr != nil {
				utils.Error("Failed to move webhook delivery to DLQ", dlqErr, map[string]interface{}{
					"component":   webhookOutboundComponent,
					"delivery_id": delivery.ID,
					"event_type":  delivery.EventType,
				})
				return fmt.Errorf("failed to move delivery to DLQ: %w", dlqErr)
			}

			utils.Info("Webhook delivery moved to DLQ", map[string]interface{}{
				"component":   webhookOutboundComponent,
				"delivery_id": delivery.ID,
				"event_type":  delivery.EventType,
			})

			// Record metrics
			webhookDeliveryTotal.WithLabelValues(delivery.EventType, "failed").Inc()
			webhookDeliveryDuration.WithLabelValues(delivery.EventType, "failed").Observe(time.Since(startTime).Seconds())
			webhookSubscriptionHealth.WithLabelValues(subscriptionIDStr, "failed").Inc()
			webhookConsecutiveFailures.WithLabelValues(subscriptionIDStr, delivery.EventType).Inc()
			webhookDLQMovements.WithLabelValues(delivery.EventType, "max_retries_network_error").Inc()

			return fmt.Errorf("max retries exceeded: %s", errMsg)
		}

		// Schedule retry
		nextRetry := s.calculateNextRetry(delivery.AttemptCount + 1)
		utils.Warn("Webhook delivery failed, scheduling retry", map[string]interface{}{
			"component":   webhookOutboundComponent,
			"delivery_id": delivery.ID,
			"event_type":  delivery.EventType,
			"reason":      errMsg,
			"next_retry":  nextRetry,
		})

		// Record metrics
		webhookDeliveryTotal.WithLabelValues(delivery.EventType, "retry").Inc()
		webhookDeliveryDuration.WithLabelValues(delivery.EventType, "retry").Observe(time.Since(startTime).Seconds())
		webhookConsecutiveFailures.WithLabelValues(subscriptionIDStr, delivery.EventType).Inc()

		return s.webhookRepo.UpdateDeliveryFailure(ctx, delivery.ID, nil, errMsg, &nextRetry)
	}
	defer resp.Body.Close()

	// Read response body (limit to 10KB)
	responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 10*1024))

	// Record HTTP status code metric
	webhookHTTPStatusCode.WithLabelValues(delivery.EventType, strconv.Itoa(resp.StatusCode)).Inc()

	// Check status code
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Success
		utils.Info("Webhook delivery successful", map[string]interface{}{
			"component":   webhookOutboundComponent,
			"delivery_id": delivery.ID,
			"event_type":  delivery.EventType,
			"status_code": resp.StatusCode,
		})

		// Calculate time to success (time from creation/first attempt to success)
		timeToSuccess := time.Since(delivery.CreatedAt).Seconds()
		webhookTimeToSuccess.WithLabelValues(delivery.EventType).Observe(timeToSuccess)

		// Record success metrics
		webhookDeliveryTotal.WithLabelValues(delivery.EventType, "success").Inc()
		webhookDeliveryDuration.WithLabelValues(delivery.EventType, "success").Observe(time.Since(startTime).Seconds())
		webhookRetryAttempts.WithLabelValues(delivery.EventType, "success").Observe(float64(delivery.AttemptCount))
		webhookSubscriptionHealth.WithLabelValues(subscriptionIDStr, "success").Inc()

		// Reset consecutive failures for this subscription
		webhookConsecutiveFailures.WithLabelValues(subscriptionIDStr, delivery.EventType).Set(0)

		if err := s.webhookRepo.UpdateDeliverySuccess(ctx, delivery.ID, resp.StatusCode, string(responseBody)); err != nil {
			return fmt.Errorf("failed to update delivery success: %w", err)
		}

		// Update subscription's last delivery time
		if err := s.webhookRepo.UpdateLastDeliveryTime(ctx, subscription.ID, time.Now()); err != nil {
			utils.Error("Failed to update webhook subscription last delivery time", err, map[string]interface{}{
				"component":       webhookOutboundComponent,
				"subscription_id": subscription.ID,
				"event_type":      delivery.EventType,
			})
		}

		return nil
	}

	// Failed delivery - schedule retry or move to DLQ
	nextRetry := s.calculateNextRetry(delivery.AttemptCount + 1)
	errMsg := fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(responseBody))

	// Track retry rate metric (only for actual retries, not the initial attempt)
	if delivery.AttemptCount > 0 {
		webhookRetryRate.WithLabelValues(delivery.EventType, strconv.Itoa(delivery.AttemptCount+1)).Inc()
	}

	// Check if this is the final attempt
	if delivery.AttemptCount+1 >= delivery.MaxAttempts {
		utils.Warn("Max retries reached for webhook delivery, moving to DLQ", map[string]interface{}{
			"component":    webhookOutboundComponent,
			"delivery_id":  delivery.ID,
			"event_type":   delivery.EventType,
			"attempt":      delivery.AttemptCount + 1,
			"max_attempts": delivery.MaxAttempts,
		})

		// Update delivery with final failure status
		if err := s.webhookRepo.UpdateDeliveryFailure(ctx, delivery.ID, &resp.StatusCode, errMsg, nil); err != nil {
			utils.Error("Failed to update webhook delivery failure", err, map[string]interface{}{
				"component":   webhookOutboundComponent,
				"delivery_id": delivery.ID,
				"event_type":  delivery.EventType,
			})
			return fmt.Errorf("failed to update delivery failure before DLQ: %w", err)
		}

		// Get updated delivery to move to DLQ
		updatedDelivery, err := s.webhookRepo.GetDeliveryByID(ctx, delivery.ID)
		if err != nil {
			utils.Error("Failed to get webhook delivery for DLQ", err, map[string]interface{}{
				"component":   webhookOutboundComponent,
				"delivery_id": delivery.ID,
				"event_type":  delivery.EventType,
			})
			return fmt.Errorf("failed to get delivery for DLQ: %w", err)
		}

		// Move to dead-letter queue
		if dlqErr := s.webhookRepo.MoveDeliveryToDeadLetterQueue(ctx, updatedDelivery); dlqErr != nil {
			utils.Error("Failed to move webhook delivery to DLQ", dlqErr, map[string]interface{}{
				"component":   webhookOutboundComponent,
				"delivery_id": delivery.ID,
				"event_type":  delivery.EventType,
			})
			return fmt.Errorf("failed to move delivery to DLQ: %w", dlqErr)
		}

		utils.Info("Webhook delivery moved to DLQ", map[string]interface{}{
			"component":   webhookOutboundComponent,
			"delivery_id": delivery.ID,
			"event_type":  delivery.EventType,
		})

		// Record metrics
		webhookDeliveryTotal.WithLabelValues(delivery.EventType, "failed").Inc()
		webhookDeliveryDuration.WithLabelValues(delivery.EventType, "failed").Observe(time.Since(startTime).Seconds())
		webhookRetryAttempts.WithLabelValues(delivery.EventType, "failed").Observe(float64(delivery.AttemptCount + 1))
		webhookSubscriptionHealth.WithLabelValues(subscriptionIDStr, "failed").Inc()
		webhookConsecutiveFailures.WithLabelValues(subscriptionIDStr, delivery.EventType).Inc()

		// Determine DLQ movement reason
		dlqReason := "max_retries_http_error"
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			dlqReason = "max_retries_client_error"
		} else if resp.StatusCode >= 500 {
			dlqReason = "max_retries_server_error"
		}
		webhookDLQMovements.WithLabelValues(delivery.EventType, dlqReason).Inc()

		return fmt.Errorf("max retries exceeded: %s", errMsg)
	}

	// Schedule retry
	utils.Warn("Webhook delivery failed, scheduling retry", map[string]interface{}{
		"component":   webhookOutboundComponent,
		"delivery_id": delivery.ID,
		"event_type":  delivery.EventType,
		"reason":      errMsg,
		"next_retry":  nextRetry,
	})

	// Record metrics for retry
	webhookDeliveryTotal.WithLabelValues(delivery.EventType, "retry").Inc()
	webhookDeliveryDuration.WithLabelValues(delivery.EventType, "retry").Observe(time.Since(startTime).Seconds())
	webhookRetryAttempts.WithLabelValues(delivery.EventType, "retry").Observe(float64(delivery.AttemptCount + 1))
	webhookConsecutiveFailures.WithLabelValues(subscriptionIDStr, delivery.EventType).Inc()

	return s.webhookRepo.UpdateDeliveryFailure(ctx, delivery.ID, &resp.StatusCode, errMsg, &nextRetry)
}

// generateSignature generates HMAC-SHA256 signature for webhook payload
func (s *OutboundWebhookService) generateSignature(payload, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

// generateSecret generates a cryptographically secure random secret
func (s *OutboundWebhookService) generateSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// validateEvents validates that all events are supported
func (s *OutboundWebhookService) validateEvents(events []string) error {
	supportedEvents := make(map[string]bool)
	for _, event := range models.GetSupportedWebhookEvents() {
		supportedEvents[event] = true
	}

	for _, event := range events {
		if !supportedEvents[event] {
			return fmt.Errorf("unsupported event: %s", event)
		}
	}

	return nil
}

// validateURL validates webhook URL and protects against SSRF attacks
func (s *OutboundWebhookService) validateURL(urlStr string) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Only allow HTTP/HTTPS
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("only http and https schemes are allowed")
	}

	// Resolve hostname to IP
	ips, err := net.LookupIP(u.Hostname())
	if err != nil {
		return fmt.Errorf("cannot resolve hostname: %w", err)
	}

	// Check if any resolved IP is private/internal
	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() {
			return fmt.Errorf("webhook URLs cannot point to private/internal addresses")
		}
	}

	return nil
}

// calculateNextRetry calculates the next retry time using exponential backoff
func (s *OutboundWebhookService) calculateNextRetry(attemptCount int) time.Time {
	baseDelay := 30 * time.Second
	maxDelay := 1 * time.Hour

	// Calculate exponential backoff: 30s, 1m, 2m, 4m, etc.
	delay := time.Duration(float64(baseDelay) * math.Pow(2, float64(attemptCount)))

	// Cap at max delay
	if delay > maxDelay {
		delay = maxDelay
	}

	return time.Now().Add(delay)
}

// GetDeliveriesBySubscriptionID retrieves deliveries for a subscription with pagination
func (s *OutboundWebhookService) GetDeliveriesBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID, userID uuid.UUID, page, limit int) ([]*models.WebhookDelivery, int, error) {
	// Verify ownership
	subscription, err := s.webhookRepo.GetSubscriptionByID(ctx, subscriptionID)
	if err != nil {
		return nil, 0, err
	}

	if subscription.UserID != userID {
		return nil, 0, fmt.Errorf("subscription not found")
	}

	offset := (page - 1) * limit
	deliveries, err := s.webhookRepo.GetDeliveriesBySubscriptionID(ctx, subscriptionID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.webhookRepo.CountDeliveriesBySubscriptionID(ctx, subscriptionID)
	if err != nil {
		return nil, 0, err
	}

	return deliveries, total, nil
}

// updateActiveSubscriptionsMetric updates the Prometheus gauge for active subscriptions
func (s *OutboundWebhookService) updateActiveSubscriptionsMetric(ctx context.Context) {
	count, err := s.webhookRepo.CountActiveSubscriptions(ctx)
	if err != nil {
		utils.Error("Failed to count active webhook subscriptions for metrics", err, map[string]interface{}{
			"component": webhookOutboundComponent,
		})
		return
	}
	webhookSubscriptionsActive.Set(float64(count))
}

// GetDeliveryStats returns statistics about webhook deliveries
func (s *OutboundWebhookService) GetDeliveryStats(ctx context.Context) (map[string]interface{}, error) {
	// Get active subscriptions count
	activeCount, err := s.webhookRepo.CountActiveSubscriptions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count active subscriptions: %w", err)
	}

	// Get pending deliveries count
	pendingCount, err := s.webhookRepo.CountPendingDeliveries(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count pending deliveries: %w", err)
	}

	// Get recent delivery stats (last hour)
	recentStats, err := s.webhookRepo.GetRecentDeliveryStats(ctx)
	if err != nil {
		// Log error but don't fail
		utils.Error("Failed to get recent webhook delivery stats", err, map[string]interface{}{
			"component": webhookOutboundComponent,
		})
		recentStats = map[string]int{
			"success": 0,
			"failed":  0,
		}
	}

	stats := map[string]interface{}{
		"active_subscriptions": activeCount,
		"pending_deliveries":   pendingCount,
		"recent_deliveries":    recentStats,
	}

	return stats, nil
}

// GetDeadLetterQueueItems retrieves items from the dead-letter queue with pagination
func (s *OutboundWebhookService) GetDeadLetterQueueItems(ctx context.Context, page, limit int) ([]*models.OutboundWebhookDeadLetterQueue, int, error) {
	offset := (page - 1) * limit
	items, err := s.webhookRepo.GetDeadLetterQueueItems(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.webhookRepo.CountDeadLetterQueueItems(ctx)
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// ReplayDeadLetterQueueItem attempts to replay a failed webhook delivery
func (s *OutboundWebhookService) ReplayDeadLetterQueueItem(ctx context.Context, dlqID uuid.UUID) error {
	startTime := time.Now()

	// Get the DLQ item
	dlqItem, err := s.webhookRepo.GetDeadLetterQueueItemByID(ctx, dlqID)
	if err != nil {
		return fmt.Errorf("failed to get DLQ item: %w", err)
	}

	// Get subscription details
	subscription, err := s.webhookRepo.GetSubscriptionByID(ctx, dlqItem.SubscriptionID)
	if err != nil {
		webhookDLQReplayFailure.WithLabelValues(dlqItem.EventType, "subscription_not_found").Inc()
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	if !subscription.IsActive {
		webhookDLQReplayFailure.WithLabelValues(dlqItem.EventType, "subscription_inactive").Inc()
		return fmt.Errorf("subscription is inactive")
	}

	utils.Info("Replaying webhook DLQ item", map[string]interface{}{
		"component":   webhookDLQComponent,
		"dlq_id":      dlqID,
		"event_type":  dlqItem.EventType,
		"delivery_id": dlqItem.DeliveryID,
	})

	// Generate signature
	signature := s.generateSignature(dlqItem.Payload, subscription.Secret)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", subscription.URL, bytes.NewBufferString(dlqItem.Payload))
	if err != nil {
		webhookDLQReplayFailure.WithLabelValues(dlqItem.EventType, "request_creation_failed").Inc()
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", signature)
	req.Header.Set("X-Webhook-Event", dlqItem.EventType)
	req.Header.Set("X-Webhook-Delivery-ID", dlqItem.DeliveryID.String())
	req.Header.Set("X-Webhook-Replay", "true")
	req.Header.Set("User-Agent", "Clipper-Webhooks/1.0")

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		// Update DLQ item with failed replay
		_ = s.webhookRepo.UpdateDLQItemReplayStatus(ctx, dlqID, false)
		webhookDLQReplayFailure.WithLabelValues(dlqItem.EventType, "network_error").Inc()
		webhookDLQReplayDuration.WithLabelValues(dlqItem.EventType, "failed").Observe(time.Since(startTime).Seconds())
		return fmt.Errorf("network error during replay: %w", err)
	}
	defer resp.Body.Close()

	// Read response body (limit to 10KB)
	responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 10*1024))

	// Check status code
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Success
		utils.Info("Webhook DLQ replay successful", map[string]interface{}{
			"component":   webhookDLQComponent,
			"dlq_id":      dlqID,
			"event_type":  dlqItem.EventType,
			"status_code": resp.StatusCode,
		})
		if err := s.webhookRepo.UpdateDLQItemReplayStatus(ctx, dlqID, true); err != nil {
			utils.Error("Failed to update DLQ replay status", err, map[string]interface{}{
				"component":  webhookDLQComponent,
				"dlq_id":     dlqID,
				"event_type": dlqItem.EventType,
			})
			return fmt.Errorf("replay succeeded but failed to update DLQ replay status: %w", err)
		}

		// Track success metrics
		webhookDLQReplaySuccess.WithLabelValues(dlqItem.EventType).Inc()
		webhookDLQReplayDuration.WithLabelValues(dlqItem.EventType, "success").Observe(time.Since(startTime).Seconds())

		return nil
	}

	// Failed replay
	_ = s.webhookRepo.UpdateDLQItemReplayStatus(ctx, dlqID, false)

	// Track failure metrics
	webhookDLQReplayFailure.WithLabelValues(dlqItem.EventType, fmt.Sprintf("http_%d", resp.StatusCode)).Inc()
	webhookDLQReplayDuration.WithLabelValues(dlqItem.EventType, "failed").Observe(time.Since(startTime).Seconds())

	return fmt.Errorf("replay failed with HTTP %d: %s", resp.StatusCode, string(responseBody))
}

// DeleteDeadLetterQueueItem deletes a DLQ item
func (s *OutboundWebhookService) DeleteDeadLetterQueueItem(ctx context.Context, dlqID uuid.UUID) error {
	return s.webhookRepo.DeleteDeadLetterQueueItem(ctx, dlqID)
}

// Helper function to create a pointer to time.Time
func ptrTime(t time.Time) *time.Time {
	return &t
}
