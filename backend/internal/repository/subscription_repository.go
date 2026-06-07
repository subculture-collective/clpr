package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// SubscriptionRepository handles database operations for subscriptions
type SubscriptionRepository struct {
	db *pgxpool.Pool
}

// NewSubscriptionRepository creates a new subscription repository
func NewSubscriptionRepository(db *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// GetByUserID retrieves a subscription by user ID
func (r *SubscriptionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Subscription, error) {
	query := `
		SELECT id, user_id, stripe_customer_id, stripe_subscription_id, stripe_price_id,
		       status, tier, current_period_start, current_period_end, cancel_at_period_end,
		       canceled_at, trial_start, trial_end, grace_period_end, created_at, updated_at
		FROM subscriptions
		WHERE user_id = $1
	`

	var sub models.Subscription
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&sub.ID, &sub.UserID, &sub.StripeCustomerID, &sub.StripeSubscriptionID, &sub.StripePriceID,
		&sub.Status, &sub.Tier, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd, &sub.CancelAtPeriodEnd,
		&sub.CanceledAt, &sub.TrialStart, &sub.TrialEnd, &sub.GracePeriodEnd, &sub.CreatedAt, &sub.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &sub, nil
}

// GetByStripeCustomerID retrieves a subscription by Stripe customer ID
func (r *SubscriptionRepository) GetByStripeCustomerID(ctx context.Context, customerID string) (*models.Subscription, error) {
	query := `
		SELECT id, user_id, stripe_customer_id, stripe_subscription_id, stripe_price_id,
		       status, tier, current_period_start, current_period_end, cancel_at_period_end,
		       canceled_at, trial_start, trial_end, grace_period_end, created_at, updated_at
		FROM subscriptions
		WHERE stripe_customer_id = $1
	`

	var sub models.Subscription
	err := r.db.QueryRow(ctx, query, customerID).Scan(
		&sub.ID, &sub.UserID, &sub.StripeCustomerID, &sub.StripeSubscriptionID, &sub.StripePriceID,
		&sub.Status, &sub.Tier, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd, &sub.CancelAtPeriodEnd,
		&sub.CanceledAt, &sub.TrialStart, &sub.TrialEnd, &sub.GracePeriodEnd, &sub.CreatedAt, &sub.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &sub, nil
}

// GetByStripeSubscriptionID retrieves a subscription by Stripe subscription ID
func (r *SubscriptionRepository) GetByStripeSubscriptionID(ctx context.Context, subscriptionID string) (*models.Subscription, error) {
	query := `
		SELECT id, user_id, stripe_customer_id, stripe_subscription_id, stripe_price_id,
		       status, tier, current_period_start, current_period_end, cancel_at_period_end,
		       canceled_at, trial_start, trial_end, grace_period_end, created_at, updated_at
		FROM subscriptions
		WHERE stripe_subscription_id = $1
	`

	var sub models.Subscription
	err := r.db.QueryRow(ctx, query, subscriptionID).Scan(
		&sub.ID, &sub.UserID, &sub.StripeCustomerID, &sub.StripeSubscriptionID, &sub.StripePriceID,
		&sub.Status, &sub.Tier, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd, &sub.CancelAtPeriodEnd,
		&sub.CanceledAt, &sub.TrialStart, &sub.TrialEnd, &sub.GracePeriodEnd, &sub.CreatedAt, &sub.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &sub, nil
}

// Create creates a new subscription
func (r *SubscriptionRepository) Create(ctx context.Context, sub *models.Subscription) error {
	query := `
		INSERT INTO subscriptions (
			user_id, stripe_customer_id, stripe_subscription_id, stripe_price_id,
			status, tier, current_period_start, current_period_end, cancel_at_period_end,
			canceled_at, trial_start, trial_end
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		sub.UserID, sub.StripeCustomerID, sub.StripeSubscriptionID, sub.StripePriceID,
		sub.Status, sub.Tier, sub.CurrentPeriodStart, sub.CurrentPeriodEnd, sub.CancelAtPeriodEnd,
		sub.CanceledAt, sub.TrialStart, sub.TrialEnd,
	).Scan(&sub.ID, &sub.CreatedAt, &sub.UpdatedAt)

	return err
}

// Update updates an existing subscription
func (r *SubscriptionRepository) Update(ctx context.Context, sub *models.Subscription) error {
	query := `
		UPDATE subscriptions
		SET stripe_subscription_id = $2, stripe_price_id = $3, status = $4, tier = $5,
		    current_period_start = $6, current_period_end = $7, cancel_at_period_end = $8,
		    canceled_at = $9, trial_start = $10, trial_end = $11
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.db.QueryRow(ctx, query,
		sub.ID, sub.StripeSubscriptionID, sub.StripePriceID, sub.Status, sub.Tier,
		sub.CurrentPeriodStart, sub.CurrentPeriodEnd, sub.CancelAtPeriodEnd,
		sub.CanceledAt, sub.TrialStart, sub.TrialEnd,
	).Scan(&sub.UpdatedAt)

	return err
}

// UpdateStatus updates the subscription status
func (r *SubscriptionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `
		UPDATE subscriptions
		SET status = $2
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id, status)
	return err
}

// Delete deletes a subscription
func (r *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM subscriptions WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// CreateEvent creates a new subscription event for audit logging
func (r *SubscriptionRepository) CreateEvent(ctx context.Context, event *models.SubscriptionEvent) error {
	query := `
		INSERT INTO subscription_events (subscription_id, event_type, stripe_event_id, payload)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(ctx, query,
		event.SubscriptionID, event.EventType, event.StripeEventID, event.Payload,
	).Scan(&event.ID, &event.CreatedAt)

	return err
}

// GetEventByStripeEventID retrieves an event by Stripe event ID (for idempotency)
func (r *SubscriptionRepository) GetEventByStripeEventID(ctx context.Context, stripeEventID string) (*models.SubscriptionEvent, error) {
	query := `
		SELECT id, subscription_id, event_type, stripe_event_id, payload, created_at
		FROM subscription_events
		WHERE stripe_event_id = $1
	`

	var event models.SubscriptionEvent
	err := r.db.QueryRow(ctx, query, stripeEventID).Scan(
		&event.ID, &event.SubscriptionID, &event.EventType, &event.StripeEventID, &event.Payload, &event.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &event, nil
}

// GetEventsBySubscriptionID retrieves all events for a subscription
func (r *SubscriptionRepository) GetEventsBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) ([]*models.SubscriptionEvent, error) {
	query := `
		SELECT id, subscription_id, event_type, stripe_event_id, payload, created_at
		FROM subscription_events
		WHERE subscription_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, subscriptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*models.SubscriptionEvent
	for rows.Next() {
		var event models.SubscriptionEvent
		err := rows.Scan(
			&event.ID, &event.SubscriptionID, &event.EventType, &event.StripeEventID, &event.Payload, &event.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, &event)
	}

	return events, rows.Err()
}

// LogSubscriptionEvent is a helper to log subscription events with payload
func (r *SubscriptionRepository) LogSubscriptionEvent(ctx context.Context, subscriptionID *uuid.UUID, eventType string, stripeEventID *string, payload interface{}) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	event := &models.SubscriptionEvent{
		SubscriptionID: subscriptionID,
		EventType:      eventType,
		StripeEventID:  stripeEventID,
		Payload:        string(payloadJSON),
	}

	return r.CreateEvent(ctx, event)
}
