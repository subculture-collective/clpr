package repository

import (
	"context"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SubscriptionRepository reads historical subscription records for personal-data exports
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
		       canceled_at, trial_start, trial_end, grace_period_end, last_stripe_event_created, created_at, updated_at
		FROM subscriptions
		WHERE user_id = $1
	`

	var sub models.Subscription
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&sub.ID, &sub.UserID, &sub.StripeCustomerID, &sub.StripeSubscriptionID, &sub.StripePriceID,
		&sub.Status, &sub.Tier, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd, &sub.CancelAtPeriodEnd,
		&sub.CanceledAt, &sub.TrialStart, &sub.TrialEnd, &sub.GracePeriodEnd, &sub.LastStripeEventCreated, &sub.CreatedAt, &sub.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &sub, nil
}
