package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// DunningRepository handles database operations for payment failures and dunning
type DunningRepository struct {
	db *pgxpool.Pool
}

// NewDunningRepository creates a new dunning repository
func NewDunningRepository(db *pgxpool.Pool) *DunningRepository {
	return &DunningRepository{db: db}
}

// CreatePaymentFailure creates a new payment failure record
func (r *DunningRepository) CreatePaymentFailure(ctx context.Context, failure *models.PaymentFailure) error {
	query := `
		INSERT INTO payment_failures (
			subscription_id, stripe_invoice_id, stripe_payment_intent_id, amount_due,
			currency, attempt_count, failure_reason, next_retry_at, resolved
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		failure.SubscriptionID, failure.StripeInvoiceID, failure.StripePaymentIntentID,
		failure.AmountDue, failure.Currency, failure.AttemptCount, failure.FailureReason,
		failure.NextRetryAt, failure.Resolved,
	).Scan(&failure.ID, &failure.CreatedAt, &failure.UpdatedAt)

	return err
}

// GetPaymentFailureByInvoiceID retrieves a payment failure by Stripe invoice ID
func (r *DunningRepository) GetPaymentFailureByInvoiceID(ctx context.Context, invoiceID string) (*models.PaymentFailure, error) {
	query := `
		SELECT id, subscription_id, stripe_invoice_id, stripe_payment_intent_id,
		       amount_due, currency, attempt_count, failure_reason, next_retry_at,
		       resolved, resolved_at, created_at, updated_at
		FROM payment_failures
		WHERE stripe_invoice_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var failure models.PaymentFailure
	err := r.db.QueryRow(ctx, query, invoiceID).Scan(
		&failure.ID, &failure.SubscriptionID, &failure.StripeInvoiceID, &failure.StripePaymentIntentID,
		&failure.AmountDue, &failure.Currency, &failure.AttemptCount, &failure.FailureReason,
		&failure.NextRetryAt, &failure.Resolved, &failure.ResolvedAt, &failure.CreatedAt, &failure.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &failure, nil
}

// GetPaymentFailuresBySubscriptionID retrieves all payment failures for a subscription
func (r *DunningRepository) GetPaymentFailuresBySubscriptionID(ctx context.Context, subscriptionID uuid.UUID) ([]*models.PaymentFailure, error) {
	query := `
		SELECT id, subscription_id, stripe_invoice_id, stripe_payment_intent_id,
		       amount_due, currency, attempt_count, failure_reason, next_retry_at,
		       resolved, resolved_at, created_at, updated_at
		FROM payment_failures
		WHERE subscription_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, subscriptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var failures []*models.PaymentFailure
	for rows.Next() {
		var failure models.PaymentFailure
		err := rows.Scan(
			&failure.ID, &failure.SubscriptionID, &failure.StripeInvoiceID, &failure.StripePaymentIntentID,
			&failure.AmountDue, &failure.Currency, &failure.AttemptCount, &failure.FailureReason,
			&failure.NextRetryAt, &failure.Resolved, &failure.ResolvedAt, &failure.CreatedAt, &failure.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		failures = append(failures, &failure)
	}

	return failures, rows.Err()
}

// GetUnresolvedPaymentFailures retrieves all unresolved payment failures
func (r *DunningRepository) GetUnresolvedPaymentFailures(ctx context.Context) ([]*models.PaymentFailure, error) {
	query := `
		SELECT id, subscription_id, stripe_invoice_id, stripe_payment_intent_id,
		       amount_due, currency, attempt_count, failure_reason, next_retry_at,
		       resolved, resolved_at, created_at, updated_at
		FROM payment_failures
		WHERE resolved = false
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var failures []*models.PaymentFailure
	for rows.Next() {
		var failure models.PaymentFailure
		err := rows.Scan(
			&failure.ID, &failure.SubscriptionID, &failure.StripeInvoiceID, &failure.StripePaymentIntentID,
			&failure.AmountDue, &failure.Currency, &failure.AttemptCount, &failure.FailureReason,
			&failure.NextRetryAt, &failure.Resolved, &failure.ResolvedAt, &failure.CreatedAt, &failure.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		failures = append(failures, &failure)
	}

	return failures, rows.Err()
}

// UpdatePaymentFailure updates a payment failure record
func (r *DunningRepository) UpdatePaymentFailure(ctx context.Context, failure *models.PaymentFailure) error {
	query := `
		UPDATE payment_failures
		SET attempt_count = $2, failure_reason = $3, next_retry_at = $4,
		    resolved = $5, resolved_at = $6, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.db.QueryRow(ctx, query,
		failure.ID, failure.AttemptCount, failure.FailureReason,
		failure.NextRetryAt, failure.Resolved, failure.ResolvedAt,
	).Scan(&failure.UpdatedAt)

	return err
}

// MarkPaymentFailureResolved marks a payment failure as resolved
func (r *DunningRepository) MarkPaymentFailureResolved(ctx context.Context, failureID uuid.UUID) error {
	query := `
		UPDATE payment_failures
		SET resolved = true, resolved_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, failureID)
	return err
}

// CreateDunningAttempt creates a new dunning attempt record
func (r *DunningRepository) CreateDunningAttempt(ctx context.Context, attempt *models.DunningAttempt) error {
	query := `
		INSERT INTO dunning_attempts (
			payment_failure_id, user_id, attempt_number, notification_type,
			email_sent, email_sent_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(ctx, query,
		attempt.PaymentFailureID, attempt.UserID, attempt.AttemptNumber,
		attempt.NotificationType, attempt.EmailSent, attempt.EmailSentAt,
	).Scan(&attempt.ID, &attempt.CreatedAt)

	return err
}

// GetDunningAttemptsByFailureID retrieves all dunning attempts for a payment failure
func (r *DunningRepository) GetDunningAttemptsByFailureID(ctx context.Context, failureID uuid.UUID) ([]*models.DunningAttempt, error) {
	query := `
		SELECT id, payment_failure_id, user_id, attempt_number, notification_type,
		       email_sent, email_sent_at, created_at
		FROM dunning_attempts
		WHERE payment_failure_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, failureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attempts []*models.DunningAttempt
	for rows.Next() {
		var attempt models.DunningAttempt
		err := rows.Scan(
			&attempt.ID, &attempt.PaymentFailureID, &attempt.UserID, &attempt.AttemptNumber,
			&attempt.NotificationType, &attempt.EmailSent, &attempt.EmailSentAt, &attempt.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		attempts = append(attempts, &attempt)
	}

	return attempts, rows.Err()
}

// GetDunningAttemptsByUserID retrieves all dunning attempts for a user
func (r *DunningRepository) GetDunningAttemptsByUserID(ctx context.Context, userID uuid.UUID, limit int) ([]*models.DunningAttempt, error) {
	query := `
		SELECT id, payment_failure_id, user_id, attempt_number, notification_type,
		       email_sent, email_sent_at, created_at
		FROM dunning_attempts
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attempts []*models.DunningAttempt
	for rows.Next() {
		var attempt models.DunningAttempt
		err := rows.Scan(
			&attempt.ID, &attempt.PaymentFailureID, &attempt.UserID, &attempt.AttemptNumber,
			&attempt.NotificationType, &attempt.EmailSent, &attempt.EmailSentAt, &attempt.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		attempts = append(attempts, &attempt)
	}

	return attempts, rows.Err()
}

// GetSubscriptionsInGracePeriod retrieves all subscriptions currently in grace period
func (r *DunningRepository) GetSubscriptionsInGracePeriod(ctx context.Context) ([]*models.Subscription, error) {
	where := `grace_period_end IS NOT NULL AND grace_period_end > NOW() AND status IN ('past_due', 'unpaid')`
	return r.listSubscriptionsByGracePeriod(ctx, where)
}

// GetExpiredGracePeriodSubscriptions retrieves subscriptions whose grace period has expired
func (r *DunningRepository) GetExpiredGracePeriodSubscriptions(ctx context.Context) ([]*models.Subscription, error) {
	where := `grace_period_end IS NOT NULL AND grace_period_end <= NOW() AND status IN ('past_due', 'unpaid') AND tier != 'free'`
	return r.listSubscriptionsByGracePeriod(ctx, where)
}

// listSubscriptionsByGracePeriod fetches subscriptions matching a WHERE predicate ordered by grace period
func (r *DunningRepository) listSubscriptionsByGracePeriod(ctx context.Context, where string) ([]*models.Subscription, error) {
	query := `
		SELECT id, user_id, stripe_customer_id, stripe_subscription_id, stripe_price_id,
			   status, tier, current_period_start, current_period_end, cancel_at_period_end,
			   canceled_at, trial_start, trial_end, grace_period_end, created_at, updated_at
		FROM subscriptions
		WHERE ` + where + `
		ORDER BY grace_period_end ASC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []*models.Subscription
	for rows.Next() {
		var sub models.Subscription
		if err := rows.Scan(
			&sub.ID, &sub.UserID, &sub.StripeCustomerID, &sub.StripeSubscriptionID, &sub.StripePriceID,
			&sub.Status, &sub.Tier, &sub.CurrentPeriodStart, &sub.CurrentPeriodEnd, &sub.CancelAtPeriodEnd,
			&sub.CanceledAt, &sub.TrialStart, &sub.TrialEnd, &sub.GracePeriodEnd, &sub.CreatedAt, &sub.UpdatedAt,
		); err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, &sub)
	}

	return subscriptions, rows.Err()
}

// SetGracePeriod sets the grace period end time for a subscription
func (r *DunningRepository) SetGracePeriod(ctx context.Context, subscriptionID uuid.UUID, gracePeriodEnd time.Time) error {
	query := `
		UPDATE subscriptions
		SET grace_period_end = $2, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, subscriptionID, gracePeriodEnd)
	return err
}

// ClearGracePeriod clears the grace period for a subscription
func (r *DunningRepository) ClearGracePeriod(ctx context.Context, subscriptionID uuid.UUID) error {
	query := `
		UPDATE subscriptions
		SET grace_period_end = NULL, updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, subscriptionID)
	return err
}
