package repository

import (
	"context"
	"fmt"
	"sort"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
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
		WITH months AS (
			SELECT generate_series(
				DATE_TRUNC('month', NOW()) - INTERVAL '1 month' * ($1 - 1),
				DATE_TRUNC('month', NOW()), INTERVAL '1 month'
			) AS month
		)
		SELECT m.month,
			COUNT(s.id) FILTER (WHERE s.created_at >= m.month AND s.created_at < m.month + INTERVAL '1 month') AS new_subscribers,
			COUNT(s.id) FILTER (WHERE s.canceled_at >= m.month AND s.canceled_at < m.month + INTERVAL '1 month') AS churned_subscribers,
			COUNT(s.id) FILTER (WHERE s.created_at < m.month + INTERVAL '1 month'
				AND (s.canceled_at IS NULL OR s.canceled_at >= m.month + INTERVAL '1 month')) AS total_subscribers
		FROM months m
		LEFT JOIN subscriptions s ON s.tier = 'pro' AND s.created_at < m.month + INTERVAL '1 month'
		GROUP BY m.month
		ORDER BY m.month
	`

	rows, err := r.db.Query(ctx, query, months)
	if err != nil {
		return nil, fmt.Errorf("failed to query subscriber growth: %w", err)
	}
	defer rows.Close()

	var growth []models.SubscriberGrowthMetric
	for rows.Next() {
		var month time.Time
		var newSubs, churnedSubs, totalSubs int

		if err := rows.Scan(&month, &newSubs, &churnedSubs, &totalSubs); err != nil {
			return nil, fmt.Errorf("failed to scan subscriber growth row: %w", err)
		}

		netChange := newSubs - churnedSubs
		growth = append(growth, models.SubscriberGrowthMetric{
			Month:     month.Format("2006-01"),
			Total:     totalSubs,
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
		SELECT DATE_TRUNC('month', created_at) AS month,
		       SUM(COALESCE(NULLIF(payload->>'amount_paid', '')::numeric, 0)) AS revenue
		FROM subscription_events
		WHERE event_type = 'invoice_paid'
		  AND created_at >= DATE_TRUNC('month', NOW()) - INTERVAL '1 month' * ($1 - 1)
		GROUP BY DATE_TRUNC('month', created_at)
		ORDER BY month
	`

	rows, err := r.db.Query(ctx, query, months)
	if err != nil {
		return nil, fmt.Errorf("failed to query revenue by month: %w", err)
	}
	defer rows.Close()

	monthly := make(map[string]*models.RevenueByMonthMetric)
	for rows.Next() {
		var month time.Time
		var revenue float64
		if err := rows.Scan(&month, &revenue); err != nil {
			return nil, fmt.Errorf("failed to scan revenue by month row: %w", err)
		}
		monthKey := month.Format("2006-01")
		monthly[monthKey] = &models.RevenueByMonthMetric{Month: monthKey, Revenue: revenue}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	mrrQuery := `
		WITH months AS (
			SELECT generate_series(DATE_TRUNC('month', NOW()) - INTERVAL '1 month' * ($1 - 1), DATE_TRUNC('month', NOW()), INTERVAL '1 month') AS month
		)
		SELECT m.month, s.stripe_price_id, COUNT(s.id)
		FROM months m
		LEFT JOIN subscriptions s ON s.tier = 'pro' AND s.created_at < m.month + INTERVAL '1 month'
			AND (s.canceled_at IS NULL OR s.canceled_at >= m.month + INTERVAL '1 month')
		GROUP BY m.month, s.stripe_price_id
		ORDER BY m.month
	`
	mrrRows, err := r.db.Query(ctx, mrrQuery, months)
	if err != nil {
		return nil, fmt.Errorf("failed to query monthly MRR: %w", err)
	}
	defer mrrRows.Close()
	for mrrRows.Next() {
		var month time.Time
		var priceID *string
		var count int
		if err := mrrRows.Scan(&month, &priceID, &count); err != nil {
			return nil, fmt.Errorf("failed to scan monthly MRR row: %w", err)
		}
		key := month.Format("2006-01")
		if monthly[key] == nil {
			monthly[key] = &models.RevenueByMonthMetric{Month: key}
		}
		if priceID != nil {
			monthly[key].MRR += priceMapping[*priceID] * float64(count)
		}
	}
	if err := mrrRows.Err(); err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(monthly))
	for key := range monthly {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]models.RevenueByMonthMetric, 0, len(keys))
	for _, key := range keys {
		result = append(result, *monthly[key])
	}

	return result, nil
}

// GetTotalRevenue returns total revenue from all paid invoices
func (r *RevenueRepository) GetTotalRevenue(ctx context.Context) (float64, error) {
	query := `
		SELECT COALESCE(SUM(COALESCE(NULLIF(payload->>'amount_paid', '')::numeric, 0)), 0)
		FROM subscription_events
		WHERE event_type = 'invoice_paid'
	`
	var totalRevenue float64
	if err := r.db.QueryRow(ctx, query).Scan(&totalRevenue); err != nil {
		return 0, fmt.Errorf("failed to query total revenue: %w", err)
	}
	return totalRevenue, nil
}
