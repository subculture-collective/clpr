package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// OutboundWebhookRepository handles database operations for outbound webhooks
type OutboundWebhookRepository struct {
	db *pgxpool.Pool
}

// NewOutboundWebhookRepository creates a new outbound webhook repository
func NewOutboundWebhookRepository(db *pgxpool.Pool) *OutboundWebhookRepository {
	return &OutboundWebhookRepository{db: db}
}

// CreateSubscription creates a new webhook subscription
func (r *OutboundWebhookRepository) CreateSubscription(ctx context.Context, subscription *models.WebhookSubscription) error {
	query := `
		INSERT INTO webhook_subscriptions (id, user_id, url, secret, events, is_active, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(ctx, query,
		subscription.ID,
		subscription.UserID,
		subscription.URL,
		subscription.Secret,
		subscription.Events,
		subscription.IsActive,
		subscription.Description,
	)

	return err
}

// GetSubscriptionByID retrieves a webhook subscription by ID
func (r *OutboundWebhookRepository) GetSubscriptionByID(ctx context.Context, id uuid.UUID) (*models.WebhookSubscription, error) {
	query := `
		SELECT id, user_id, url, secret, events, is_active, description, created_at, updated_at, last_delivery_at
		FROM webhook_subscriptions
		WHERE id = $1
	`

	var subscription models.WebhookSubscription
	err := r.db.QueryRow(ctx, query, id).Scan(
		&subscription.ID,
		&subscription.UserID,
		&subscription.URL,
		&subscription.Secret,
		&subscription.Events,
		&subscription.IsActive,
		&subscription.Description,
		&subscription.CreatedAt,
		&subscription.UpdatedAt,
		&subscription.LastDeliveryAt,
	)

	if err != nil {
		return nil, err
	}

	return &subscription, nil
}

// GetSubscriptionsByUserID retrieves all webhook subscriptions for a user
func (r *OutboundWebhookRepository) GetSubscriptionsByUserID(ctx context.Context, userID uuid.UUID) ([]*models.WebhookSubscription, error) {
	query := `
		SELECT id, user_id, url, secret, events, is_active, description, created_at, updated_at, last_delivery_at
		FROM webhook_subscriptions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []*models.WebhookSubscription
	for rows.Next() {
		var subscription models.WebhookSubscription
		err := rows.Scan(
			&subscription.ID,
			&subscription.UserID,
			&subscription.URL,
			&subscription.Secret,
			&subscription.Events,
			&subscription.IsActive,
			&subscription.Description,
			&subscription.CreatedAt,
			&subscription.UpdatedAt,
			&subscription.LastDeliveryAt,
		)
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, &subscription)
	}

	return subscriptions, rows.Err()
}

// GetActiveSubscriptionsByEvent retrieves all active webhook subscriptions for a specific event
func (r *OutboundWebhookRepository) GetActiveSubscriptionsByEvent(ctx context.Context, eventType string) ([]*models.WebhookSubscription, error) {
	query := `
		SELECT id, user_id, url, secret, events, is_active, description, created_at, updated_at, last_delivery_at
		FROM webhook_subscriptions
		WHERE is_active = true AND $1 = ANY(events)
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, eventType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []*models.WebhookSubscription
	for rows.Next() {
		var subscription models.WebhookSubscription
		err := rows.Scan(
			&subscription.ID,
			&subscription.UserID,
			&subscription.URL,
			&subscription.Secret,
			&subscription.Events,
			&subscription.IsActive,
			&subscription.Description,
			&subscription.CreatedAt,
			&subscription.UpdatedAt,
			&subscription.LastDeliveryAt,
		)
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, &subscription)
	}

	return subscriptions, rows.Err()
}

// UpdateSubscription updates a webhook subscription
func (r *OutboundWebhookRepository) UpdateSubscription(ctx context.Context, id uuid.UUID, url *string, events []string, isActive *bool, description *string) error {
	query := `
		UPDATE webhook_subscriptions
		SET url = COALESCE($2, url),
		    events = COALESCE($3, events),
		    is_active = COALESCE($4, is_active),
		    description = COALESCE($5, description)
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id, url, events, isActive, description)
	return err
}

// DeleteSubscription deletes a webhook subscription
func (r *OutboundWebhookRepository) DeleteSubscription(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM webhook_subscriptions WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// UpdateLastDeliveryTime updates the last delivery time for a subscription
func (r *OutboundWebhookRepository) UpdateLastDeliveryTime(ctx context.Context, id uuid.UUID, deliveryTime time.Time) error {
	query := `UPDATE webhook_subscriptions SET last_delivery_at = $2 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, deliveryTime)
	return err
}

// UpdateSubscriptionSecret updates the secret for a webhook subscription
func (r *OutboundWebhookRepository) UpdateSubscriptionSecret(ctx context.Context, id uuid.UUID, secret string) error {
	query := `UPDATE webhook_subscriptions SET secret = $2 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, secret)
	return err
}

// CreateDelivery creates a new webhook delivery record
func (r *OutboundWebhookRepository) CreateDelivery(ctx context.Context, delivery *models.WebhookDelivery) error {
	query := `
		INSERT INTO webhook_deliveries (id, subscription_id, event_type, event_id, payload, status, attempt_count, max_attempts, next_attempt_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Exec(ctx, query,
		delivery.ID,
		delivery.SubscriptionID,
		delivery.EventType,
		delivery.EventID,
		delivery.Payload,
		delivery.Status,
		delivery.AttemptCount,
		delivery.MaxAttempts,
		delivery.NextAttemptAt,
	)

	return err
}

// GetDeliveryByID retrieves a webhook delivery by ID
func (r *OutboundWebhookRepository) GetDeliveryByID(ctx context.Context, id uuid.UUID) (*models.WebhookDelivery, error) {
	query := `
		SELECT id, subscription_id, event_type, event_id, payload, status, http_status_code,
		       response_body, error_message, attempt_count, max_attempts, next_attempt_at,
		       delivered_at, created_at, updated_at
		FROM webhook_deliveries
		WHERE id = $1
	`

	var delivery models.WebhookDelivery
	err := r.db.QueryRow(ctx, query, id).Scan(
		&delivery.ID,
		&delivery.SubscriptionID,
		&delivery.EventType,
		&delivery.EventID,
		&delivery.Payload,
		&delivery.Status,
		&delivery.HTTPStatusCode,
		&delivery.ResponseBody,
		&delivery.ErrorMessage,
		&delivery.AttemptCount,
		&delivery.MaxAttempts,
		&delivery.NextAttemptAt,
		&delivery.DeliveredAt,
		&delivery.CreatedAt,
		&delivery.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &delivery, nil
}

// GetPendingDeliveries retrieves all webhook deliveries ready for retry
func (r *OutboundWebhookRepository) GetPendingDeliveries(ctx context.Context, limit int) ([]*models.WebhookDelivery, error) {
	query := `
		SELECT id, subscription_id, event_type, event_id, payload, status, http_status_code,
		       response_body, error_message, attempt_count, max_attempts, next_attempt_at,
		       delivered_at, created_at, updated_at
		FROM webhook_deliveries
		WHERE status = 'pending' AND next_attempt_at <= $1 AND attempt_count < max_attempts
		ORDER BY next_attempt_at ASC
		LIMIT $2
		FOR UPDATE SKIP LOCKED
	`

	rows, err := r.db.Query(ctx, query, time.Now(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deliveries []*models.WebhookDelivery
	for rows.Next() {
		var delivery models.WebhookDelivery
		err := rows.Scan(
			&delivery.ID,
			&delivery.SubscriptionID,
			&delivery.EventType,
			&delivery.EventID,
			&delivery.Payload,
			&delivery.Status,
			&delivery.HTTPStatusCode,
			&delivery.ResponseBody,
			&delivery.ErrorMessage,
			&delivery.AttemptCount,
			&delivery.MaxAttempts,
			&delivery.NextAttemptAt,
			&delivery.DeliveredAt,
			&delivery.CreatedAt,
			&delivery.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		deliveries = append(deliveries, &delivery)
	}

	return deliveries, rows.Err()
}

// UpdateDeliverySuccess updates a webhook delivery after successful delivery
func (r *OutboundWebhookRepository) UpdateDeliverySuccess(ctx context.Context, id uuid.UUID, statusCode int, responseBody string) error {
	query := `
		UPDATE webhook_deliveries
		SET status = 'delivered',
		    http_status_code = $2,
		    response_body = $3,
		    delivered_at = $4,
		    attempt_count = attempt_count + 1
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id, statusCode, responseBody, time.Now())
	return err
}

// UpdateDeliveryFailure updates a webhook delivery after failed delivery
func (r *OutboundWebhookRepository) UpdateDeliveryFailure(ctx context.Context, id uuid.UUID, statusCode *int, errorMessage string, nextAttemptAt *time.Time) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Get current attempt count
	var attemptCount, maxAttempts int
	err = tx.QueryRow(ctx, "SELECT attempt_count, max_attempts FROM webhook_deliveries WHERE id = $1", id).Scan(&attemptCount, &maxAttempts)
	if err != nil {
		return fmt.Errorf("failed to get delivery info: %w", err)
	}

	newAttemptCount := attemptCount + 1
	status := "pending"

	// If we've exhausted retries, mark as failed
	if newAttemptCount >= maxAttempts {
		status = "failed"
	}

	query := `
		UPDATE webhook_deliveries
		SET status = $2,
		    http_status_code = $3,
		    error_message = $4,
		    attempt_count = $5,
		    next_attempt_at = $6
		WHERE id = $1
	`

	_, err = tx.Exec(ctx, query, id, status, statusCode, errorMessage, newAttemptCount, nextAttemptAt)
	if err != nil {
		return fmt.Errorf("failed to update delivery: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetDeliveriesBySubscriptionID retrieves all deliveries for a subscription
func (r *OutboundWebhookRepository) GetDeliveriesBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID, limit, offset int) ([]*models.WebhookDelivery, error) {
	query := `
		SELECT id, subscription_id, event_type, event_id, payload, status, http_status_code,
		       response_body, error_message, attempt_count, max_attempts, next_attempt_at,
		       delivered_at, created_at, updated_at
		FROM webhook_deliveries
		WHERE subscription_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, subscriptionID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deliveries []*models.WebhookDelivery
	for rows.Next() {
		var delivery models.WebhookDelivery
		err := rows.Scan(
			&delivery.ID,
			&delivery.SubscriptionID,
			&delivery.EventType,
			&delivery.EventID,
			&delivery.Payload,
			&delivery.Status,
			&delivery.HTTPStatusCode,
			&delivery.ResponseBody,
			&delivery.ErrorMessage,
			&delivery.AttemptCount,
			&delivery.MaxAttempts,
			&delivery.NextAttemptAt,
			&delivery.DeliveredAt,
			&delivery.CreatedAt,
			&delivery.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		deliveries = append(deliveries, &delivery)
	}

	return deliveries, rows.Err()
}

// CountDeliveriesBySubscriptionID counts deliveries for a subscription
func (r *OutboundWebhookRepository) CountDeliveriesBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM webhook_deliveries WHERE subscription_id = $1`
	err := r.db.QueryRow(ctx, query, subscriptionID).Scan(&count)
	return count, err
}

// CountActiveSubscriptions counts the number of active webhook subscriptions
func (r *OutboundWebhookRepository) CountActiveSubscriptions(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM webhook_subscriptions WHERE is_active = true`
	err := r.db.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// CountPendingDeliveries counts the number of pending webhook deliveries
func (r *OutboundWebhookRepository) CountPendingDeliveries(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM webhook_deliveries WHERE status = 'pending'`
	err := r.db.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// GetRecentDeliveryStats returns delivery statistics for the last hour
func (r *OutboundWebhookRepository) GetRecentDeliveryStats(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT
			status,
			COUNT(*) as count
		FROM webhook_deliveries
		WHERE created_at > NOW() - INTERVAL '1 hour'
		GROUP BY status
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := map[string]int{
		"delivered": 0,
		"failed":    0,
		"pending":   0,
	}

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		stats[status] = count
	}

	return stats, rows.Err()
}

// MoveDeliveryToDeadLetterQueue moves a failed delivery to the dead-letter queue
func (r *OutboundWebhookRepository) MoveDeliveryToDeadLetterQueue(ctx context.Context, delivery *models.WebhookDelivery) error {
	query := `
		INSERT INTO outbound_webhook_dead_letter_queue (
			id, subscription_id, delivery_id, event_type, event_id, payload,
			error_message, http_status_code, response_body, attempt_count, original_created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	// Ensure error_message is not nil by providing a default value
	errorMessage := "Unknown error"
	if delivery.ErrorMessage != nil && *delivery.ErrorMessage != "" {
		errorMessage = *delivery.ErrorMessage
	}

	_, err := r.db.Exec(ctx, query,
		uuid.New(),
		delivery.SubscriptionID,
		delivery.ID,
		delivery.EventType,
		delivery.EventID,
		delivery.Payload,
		errorMessage,
		delivery.HTTPStatusCode,
		delivery.ResponseBody,
		delivery.AttemptCount,
		delivery.CreatedAt,
	)

	return err
}

// GetDeadLetterQueueItems retrieves items from the outbound webhook dead-letter queue
func (r *OutboundWebhookRepository) GetDeadLetterQueueItems(ctx context.Context, limit, offset int) ([]*models.OutboundWebhookDeadLetterQueue, error) {
	query := `
		SELECT id, subscription_id, delivery_id, event_type, event_id, payload,
		       error_message, http_status_code, response_body, attempt_count,
		       original_created_at, moved_to_dlq_at, replayed_at, replay_successful, created_at
		FROM outbound_webhook_dead_letter_queue
		ORDER BY moved_to_dlq_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*models.OutboundWebhookDeadLetterQueue = make([]*models.OutboundWebhookDeadLetterQueue, 0)
	for rows.Next() {
		var item models.OutboundWebhookDeadLetterQueue
		err := rows.Scan(
			&item.ID,
			&item.SubscriptionID,
			&item.DeliveryID,
			&item.EventType,
			&item.EventID,
			&item.Payload,
			&item.ErrorMessage,
			&item.HTTPStatusCode,
			&item.ResponseBody,
			&item.AttemptCount,
			&item.OriginalCreatedAt,
			&item.MovedToDLQAt,
			&item.ReplayedAt,
			&item.ReplaySuccessful,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, &item)
	}

	return items, rows.Err()
}

// CountDeadLetterQueueItems returns the total count of items in the outbound webhook DLQ
func (r *OutboundWebhookRepository) CountDeadLetterQueueItems(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM outbound_webhook_dead_letter_queue`
	err := r.db.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// GetDeadLetterQueueItemByID retrieves a single DLQ item by ID
func (r *OutboundWebhookRepository) GetDeadLetterQueueItemByID(ctx context.Context, id uuid.UUID) (*models.OutboundWebhookDeadLetterQueue, error) {
	query := `
		SELECT id, subscription_id, delivery_id, event_type, event_id, payload,
		       error_message, http_status_code, response_body, attempt_count,
		       original_created_at, moved_to_dlq_at, replayed_at, replay_successful, created_at
		FROM outbound_webhook_dead_letter_queue
		WHERE id = $1
	`

	var item models.OutboundWebhookDeadLetterQueue
	err := r.db.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.SubscriptionID,
		&item.DeliveryID,
		&item.EventType,
		&item.EventID,
		&item.Payload,
		&item.ErrorMessage,
		&item.HTTPStatusCode,
		&item.ResponseBody,
		&item.AttemptCount,
		&item.OriginalCreatedAt,
		&item.MovedToDLQAt,
		&item.ReplayedAt,
		&item.ReplaySuccessful,
		&item.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &item, nil
}

// UpdateDLQItemReplayStatus updates the replay status of a DLQ item
func (r *OutboundWebhookRepository) UpdateDLQItemReplayStatus(ctx context.Context, id uuid.UUID, successful bool) error {
	query := `
		UPDATE outbound_webhook_dead_letter_queue
		SET replayed_at = $2, replay_successful = $3
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id, time.Now(), successful)
	return err
}

// DeleteDeadLetterQueueItem deletes a DLQ item
func (r *OutboundWebhookRepository) DeleteDeadLetterQueueItem(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM outbound_webhook_dead_letter_queue WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
