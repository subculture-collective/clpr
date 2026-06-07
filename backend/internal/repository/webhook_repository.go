package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// WebhookRepository handles database operations for webhook retry and dead-letter queue
type WebhookRepository struct {
	db *pgxpool.Pool
}

// NewWebhookRepository creates a new webhook repository
func NewWebhookRepository(db *pgxpool.Pool) *WebhookRepository {
	return &WebhookRepository{db: db}
}

// AddToRetryQueue adds a failed webhook to the retry queue
func (r *WebhookRepository) AddToRetryQueue(ctx context.Context, stripeEventID, eventType string, payload interface{}, maxRetries int) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	query := `
		INSERT INTO webhook_retry_queue (stripe_event_id, event_type, payload, retry_count, max_retries, next_retry_at)
		VALUES ($1, $2, $3, 0, $4, $5)
		ON CONFLICT (stripe_event_id) DO NOTHING
	`

	nextRetry := time.Now().Add(30 * time.Second) // Initial retry after 30 seconds
	_, err = r.db.Exec(ctx, query, stripeEventID, eventType, string(payloadJSON), maxRetries, nextRetry)
	return err
}

// GetRetryQueueItem retrieves a webhook event from the retry queue by event ID
func (r *WebhookRepository) GetRetryQueueItem(ctx context.Context, stripeEventID string) (*models.WebhookRetryQueue, error) {
	query := `
		SELECT id, stripe_event_id, event_type, payload, retry_count, max_retries,
		       next_retry_at, last_error, created_at, updated_at
		FROM webhook_retry_queue
		WHERE stripe_event_id = $1
	`

	var item models.WebhookRetryQueue
	err := r.db.QueryRow(ctx, query, stripeEventID).Scan(
		&item.ID, &item.StripeEventID, &item.EventType, &item.Payload, &item.RetryCount,
		&item.MaxRetries, &item.NextRetryAt, &item.LastError, &item.CreatedAt, &item.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &item, nil
}

// GetPendingRetries retrieves all webhook events ready for retry
// Uses FOR UPDATE SKIP LOCKED to prevent concurrent processing by multiple instances
func (r *WebhookRepository) GetPendingRetries(ctx context.Context, limit int) ([]*models.WebhookRetryQueue, error) {
	query := `
		SELECT id, stripe_event_id, event_type, payload, retry_count, max_retries,
		       next_retry_at, last_error, created_at, updated_at
		FROM webhook_retry_queue
		WHERE next_retry_at <= $1 AND retry_count < max_retries
		ORDER BY next_retry_at ASC
		LIMIT $2
		FOR UPDATE SKIP LOCKED
	`

	rows, err := r.db.Query(ctx, query, time.Now(), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*models.WebhookRetryQueue
	for rows.Next() {
		var item models.WebhookRetryQueue
		err := rows.Scan(
			&item.ID, &item.StripeEventID, &item.EventType, &item.Payload, &item.RetryCount,
			&item.MaxRetries, &item.NextRetryAt, &item.LastError, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, &item)
	}

	return items, rows.Err()
}

// UpdateRetryQueueItem updates a webhook retry queue item after a retry attempt
func (r *WebhookRepository) UpdateRetryQueueItem(ctx context.Context, id uuid.UUID, retryCount int, nextRetryAt *time.Time, lastError string) error {
	query := `
		UPDATE webhook_retry_queue
		SET retry_count = $2, next_retry_at = $3, last_error = $4
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id, retryCount, nextRetryAt, lastError)
	return err
}

// RemoveFromRetryQueue removes a webhook event from the retry queue (after successful processing)
func (r *WebhookRepository) RemoveFromRetryQueue(ctx context.Context, stripeEventID string) error {
	query := `DELETE FROM webhook_retry_queue WHERE stripe_event_id = $1`
	_, err := r.db.Exec(ctx, query, stripeEventID)
	return err
}

// MoveToDeadLetterQueue moves a failed webhook event to the dead-letter queue
func (r *WebhookRepository) MoveToDeadLetterQueue(ctx context.Context, item *models.WebhookRetryQueue, finalError string) error {
	// Start a transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Insert into dead-letter queue
	dlqQuery := `
		INSERT INTO webhook_dead_letter_queue (stripe_event_id, event_type, payload, retry_count, error, original_timestamp)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (stripe_event_id) DO UPDATE
		SET error = EXCLUDED.error, retry_count = EXCLUDED.retry_count
	`

	_, err = tx.Exec(ctx, dlqQuery,
		item.StripeEventID, item.EventType, item.Payload, item.RetryCount, finalError, item.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert into dead-letter queue: %w", err)
	}

	// Remove from retry queue
	removeQuery := `DELETE FROM webhook_retry_queue WHERE stripe_event_id = $1`
	_, err = tx.Exec(ctx, removeQuery, item.StripeEventID)
	if err != nil {
		return fmt.Errorf("failed to remove from retry queue: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetDeadLetterQueueItems retrieves items from the dead-letter queue
func (r *WebhookRepository) GetDeadLetterQueueItems(ctx context.Context, limit int, offset int) ([]*models.WebhookDeadLetterQueue, error) {
	query := `
		SELECT id, stripe_event_id, event_type, payload, retry_count, error, original_timestamp, created_at
		FROM webhook_dead_letter_queue
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*models.WebhookDeadLetterQueue
	for rows.Next() {
		var item models.WebhookDeadLetterQueue
		err := rows.Scan(
			&item.ID, &item.StripeEventID, &item.EventType, &item.Payload,
			&item.RetryCount, &item.Error, &item.OriginalTimestamp, &item.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, &item)
	}

	return items, rows.Err()
}

// CountDeadLetterQueueItems returns the total count of items in the dead-letter queue
func (r *WebhookRepository) CountDeadLetterQueueItems(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM webhook_dead_letter_queue`
	err := r.db.QueryRow(ctx, query).Scan(&count)
	return count, err
}

// CountPendingRetries returns the count of pending retry items
func (r *WebhookRepository) CountPendingRetries(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM webhook_retry_queue WHERE next_retry_at <= $1 AND retry_count < max_retries`
	err := r.db.QueryRow(ctx, query, time.Now()).Scan(&count)
	return count, err
}
