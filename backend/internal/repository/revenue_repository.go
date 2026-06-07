package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

const (
	// avgSecondsPerMonth is the average number of seconds in a month (30.44 days * 24 hours * 60 min * 60 sec)
	// Used for calculating months retained in cohort analysis across year boundaries
	avgSecondsPerMonth = 30.44 * 24 * 60 * 60
)

// RevenueRepository handles database operations for revenue metrics
type RevenueRepository struct {
	db *pgxpool.Pool
}

// NewRevenueRepository creates a new revenue repository
func NewRevenueRepository(db *pgxpool.Pool) *RevenueRepository {
	return &RevenueRepository{db: db}
}

// GetMRR calculates Monthly Recurring Revenue from active subscriptions
func (r *RevenueRepository) GetMRR(ctx context.Context, priceMapping map[string]float64) (float64, error) {
	query := `
		SELECT stripe_price_id, COUNT(*) as count
		FROM subscriptions
		WHERE status IN ('active', 'trialing')
		AND tier = 'pro'
		GROUP BY stripe_price_id
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to query MRR: %w", err)
	}
	defer rows.Close()

	var totalMRR float64
	for rows.Next() {
		var priceID *string
		var count int
		if err := rows.Scan(&priceID, &count); err != nil {
			return 0, fmt.Errorf("failed to scan MRR row: %w", err)
		}
		if priceID != nil {
			if monthlyValue, ok := priceMapping[*priceID]; ok {
				totalMRR += monthlyValue * float64(count)
			}
		}
	}

	return totalMRR, rows.Err()
}

// GetActiveSubscriberCount returns the count of active subscribers
func (r *RevenueRepository) GetActiveSubscriberCount(ctx context.Context) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM subscriptions
		WHERE status IN ('active', 'trialing')
		AND tier = 'pro'
	`

	var count int
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get active subscriber count: %w", err)
	}

	return count, nil
}

// GetChurnedSubscriberCount returns the count of subscribers who churned in the given period
func (r *RevenueRepository) GetChurnedSubscriberCount(ctx context.Context, since time.Time) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM subscriptions
		WHERE status = 'canceled'
		AND canceled_at >= $1
		AND tier = 'pro'
	`

	var count int
	err := r.db.QueryRow(ctx, query, since).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get churned subscriber count: %w", err)
	}

	return count, nil
}

// GetNewSubscriberCount returns the count of new subscribers in the given period
func (r *RevenueRepository) GetNewSubscriberCount(ctx context.Context, since time.Time) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM subscriptions
		WHERE created_at >= $1
		AND tier = 'pro'
		AND status IN ('active', 'trialing', 'canceled')
	`

	var count int
	err := r.db.QueryRow(ctx, query, since).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get new subscriber count: %w", err)
	}

	return count, nil
}

// GetPlanDistribution returns the distribution of subscribers by plan
func (r *RevenueRepository) GetPlanDistribution(ctx context.Context) ([]models.PlanDistributionMetric, error) {
	query := `
		SELECT stripe_price_id, COUNT(*) as count
		FROM subscriptions
		WHERE status IN ('active', 'trialing')
		AND tier = 'pro'
		AND stripe_price_id IS NOT NULL
		GROUP BY stripe_price_id
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query plan distribution: %w", err)
	}
	defer rows.Close()

	var distribution []models.PlanDistributionMetric
	var totalSubscribers int

	for rows.Next() {
		var priceID *string
		var count int
		if err := rows.Scan(&priceID, &count); err != nil {
			return nil, fmt.Errorf("failed to scan plan distribution row: %w", err)
		}

		planName := "Unknown"
		if priceID != nil {
			planName = *priceID
		}

		distribution = append(distribution, models.PlanDistributionMetric{
			PlanID:      planName,
			PlanName:    planName,
			Subscribers: count,
		})
		totalSubscribers += count
	}

	// Calculate percentages
	for i := range distribution {
		if totalSubscribers > 0 {
			distribution[i].Percentage = float64(distribution[i].Subscribers) / float64(totalSubscribers) * 100
		}
	}

	return distribution, rows.Err()
}

// GetCohortRetention calculates cohort retention data
func (r *RevenueRepository) GetCohortRetention(ctx context.Context, months int) ([]models.CohortRetentionMetric, error) {
	// Get cohorts from the last N months
	// The 30.44 * 24 * 60 * 60 = avgSecondsPerMonth constant converts epoch seconds to months
	query := `
		WITH cohorts AS (
			SELECT 
				DATE_TRUNC('month', created_at) as cohort_month,
				id as subscription_id,
				status,
				canceled_at
			FROM subscriptions
			WHERE tier = 'pro'
			AND created_at >= NOW() - INTERVAL '1 month' * $1
		),
		cohort_sizes AS (
			SELECT 
				cohort_month,
				COUNT(*) as initial_size
			FROM cohorts
			GROUP BY cohort_month
		),
		retained AS (
			SELECT 
				c.cohort_month,
				-- Calculate months retained using epoch seconds / avgSecondsPerMonth (30.44 days)
				FLOOR(EXTRACT(EPOCH FROM (COALESCE(c.canceled_at, NOW()) - c.cohort_month)) / 2629746)::int as months_retained
			FROM cohorts c
			WHERE c.status IN ('active', 'trialing', 'canceled')
		)
		SELECT 
			cs.cohort_month,
			cs.initial_size,
			r.months_retained,
			COUNT(*) as retained_count
		FROM cohort_sizes cs
		LEFT JOIN retained r ON cs.cohort_month = r.cohort_month
		GROUP BY cs.cohort_month, cs.initial_size, r.months_retained
		ORDER BY cs.cohort_month, r.months_retained
	`

	rows, err := r.db.Query(ctx, query, months)
	if err != nil {
		return nil, fmt.Errorf("failed to query cohort retention: %w", err)
	}
	defer rows.Close()

	cohortMap := make(map[string]*models.CohortRetentionMetric)

	for rows.Next() {
		var cohortMonth time.Time
		var initialSize int
		var monthsRetained *int
		var retainedCount int

		if err := rows.Scan(&cohortMonth, &initialSize, &monthsRetained, &retainedCount); err != nil {
			return nil, fmt.Errorf("failed to scan cohort retention row: %w", err)
		}

		cohortKey := cohortMonth.Format("2006-01")
		if _, exists := cohortMap[cohortKey]; !exists {
			cohortMap[cohortKey] = &models.CohortRetentionMetric{
				CohortMonth:    cohortKey,
				InitialSize:    initialSize,
				RetentionRates: make([]float64, 0),
			}
		}

		if monthsRetained != nil && initialSize > 0 {
			monthIdx := *monthsRetained
			// Ensure we have enough slots
			for len(cohortMap[cohortKey].RetentionRates) <= monthIdx {
				cohortMap[cohortKey].RetentionRates = append(cohortMap[cohortKey].RetentionRates, 0)
			}
			cohortMap[cohortKey].RetentionRates[monthIdx] = float64(retainedCount) / float64(initialSize) * 100
		}
	}

	var result []models.CohortRetentionMetric
	for _, cohort := range cohortMap {
		result = append(result, *cohort)
	}

	return result, rows.Err()
}

// GetSubscriberGrowthTrend returns subscriber growth data for the last N months
func (r *RevenueRepository) GetSubscriberGrowthTrend(ctx context.Context, months int) ([]models.SubscriberGrowthMetric, error) {
	query := `
		WITH monthly_data AS (
			SELECT 
				DATE_TRUNC('month', created_at) as month,
				COUNT(*) as new_subscribers
			FROM subscriptions
			WHERE tier = 'pro'
			AND created_at >= NOW() - INTERVAL '1 month' * $1
			GROUP BY DATE_TRUNC('month', created_at)
		),
		churned_data AS (
			SELECT 
				DATE_TRUNC('month', canceled_at) as month,
				COUNT(*) as churned_subscribers
			FROM subscriptions
			WHERE tier = 'pro'
			AND canceled_at IS NOT NULL
			AND canceled_at >= NOW() - INTERVAL '1 month' * $2
			GROUP BY DATE_TRUNC('month', canceled_at)
		)
		SELECT 
			COALESCE(m.month, c.month) as month,
			COALESCE(m.new_subscribers, 0) as new_subscribers,
			COALESCE(c.churned_subscribers, 0) as churned_subscribers
		FROM monthly_data m
		FULL OUTER JOIN churned_data c ON m.month = c.month
		ORDER BY month
	`

	rows, err := r.db.Query(ctx, query, months, months)
	if err != nil {
		return nil, fmt.Errorf("failed to query subscriber growth: %w", err)
	}
	defer rows.Close()

	var growth []models.SubscriberGrowthMetric
	runningTotal := 0

	for rows.Next() {
		var month time.Time
		var newSubs, churnedSubs int

		if err := rows.Scan(&month, &newSubs, &churnedSubs); err != nil {
			return nil, fmt.Errorf("failed to scan subscriber growth row: %w", err)
		}

		netChange := newSubs - churnedSubs
		runningTotal += netChange

		growth = append(growth, models.SubscriberGrowthMetric{
			Month:     month.Format("2006-01"),
			Total:     runningTotal,
			New:       newSubs,
			Churned:   churnedSubs,
			NetChange: netChange,
		})
	}

	return growth, rows.Err()
}

// GetTrialConversionRate calculates trial to paid conversion rate
func (r *RevenueRepository) GetTrialConversionRate(ctx context.Context, since time.Time) (float64, error) {
	query := `
		SELECT 
			COUNT(*) FILTER (WHERE trial_start IS NOT NULL) as total_trials,
			COUNT(*) FILTER (WHERE trial_start IS NOT NULL AND status = 'active') as converted
		FROM subscriptions
		WHERE tier = 'pro'
		AND created_at >= $1
	`

	var totalTrials, converted int
	err := r.db.QueryRow(ctx, query, since).Scan(&totalTrials, &converted)
	if err != nil {
		return 0, fmt.Errorf("failed to get trial conversion rate: %w", err)
	}

	if totalTrials == 0 {
		return 0, nil
	}

	return float64(converted) / float64(totalTrials) * 100, nil
}

// GetGracePeriodRecoveryRate calculates the percentage of users who recovered from grace period
func (r *RevenueRepository) GetGracePeriodRecoveryRate(ctx context.Context, since time.Time) (float64, error) {
	query := `
		SELECT 
			COUNT(*) FILTER (WHERE grace_period_end IS NOT NULL) as entered_grace,
			COUNT(*) FILTER (WHERE grace_period_end IS NOT NULL AND status = 'active') as recovered
		FROM subscriptions
		WHERE tier = 'pro'
		AND updated_at >= $1
	`

	var enteredGrace, recovered int
	err := r.db.QueryRow(ctx, query, since).Scan(&enteredGrace, &recovered)
	if err != nil {
		return 0, fmt.Errorf("failed to get grace period recovery rate: %w", err)
	}

	if enteredGrace == 0 {
		return 0, nil
	}

	return float64(recovered) / float64(enteredGrace) * 100, nil
}

// GetRevenueByMonth returns revenue data grouped by month
func (r *RevenueRepository) GetRevenueByMonth(ctx context.Context, months int, priceMapping map[string]float64) ([]models.RevenueByMonthMetric, error) {
	query := `
		WITH monthly_subs AS (
			SELECT 
				DATE_TRUNC('month', se.created_at) as month,
				s.stripe_price_id,
				COUNT(*) as count
			FROM subscription_events se
			JOIN subscriptions s ON se.subscription_id = s.id
			WHERE se.event_type = 'invoice_paid'
			AND se.created_at >= NOW() - INTERVAL '1 month' * $1
			GROUP BY DATE_TRUNC('month', se.created_at), s.stripe_price_id
		)
		SELECT month, stripe_price_id, count
		FROM monthly_subs
		ORDER BY month
	`

	rows, err := r.db.Query(ctx, query, months)
	if err != nil {
		return nil, fmt.Errorf("failed to query revenue by month: %w", err)
	}
	defer rows.Close()

	monthlyRevenue := make(map[string]float64)
	for rows.Next() {
		var month time.Time
		var priceID *string
		var count int

		if err := rows.Scan(&month, &priceID, &count); err != nil {
			return nil, fmt.Errorf("failed to scan revenue by month row: %w", err)
		}

		monthKey := month.Format("2006-01")
		if priceID != nil {
			if monthlyValue, ok := priceMapping[*priceID]; ok {
				monthlyRevenue[monthKey] += monthlyValue * float64(count)
			}
		}
	}

	var result []models.RevenueByMonthMetric
	for month, revenue := range monthlyRevenue {
		result = append(result, models.RevenueByMonthMetric{
			Month:   month,
			Revenue: revenue,
			MRR:     revenue, // Simplified - in reality would need end-of-month calculation
		})
	}

	return result, rows.Err()
}

// GetTotalRevenue returns total revenue from all paid invoices
func (r *RevenueRepository) GetTotalRevenue(ctx context.Context, priceMapping map[string]float64) (float64, error) {
	query := `
		SELECT s.stripe_price_id, COUNT(*) as count
		FROM subscription_events se
		JOIN subscriptions s ON se.subscription_id = s.id
		WHERE se.event_type = 'invoice_paid'
		GROUP BY s.stripe_price_id
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to query total revenue: %w", err)
	}
	defer rows.Close()

	var totalRevenue float64
	for rows.Next() {
		var priceID *string
		var count int
		if err := rows.Scan(&priceID, &count); err != nil {
			return 0, fmt.Errorf("failed to scan total revenue row: %w", err)
		}
		if priceID != nil {
			if monthlyValue, ok := priceMapping[*priceID]; ok {
				totalRevenue += monthlyValue * float64(count)
			}
		}
	}

	return totalRevenue, rows.Err()
}
