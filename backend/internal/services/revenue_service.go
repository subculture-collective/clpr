package services

import (
	"context"
	"time"

	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
)

// RevenueService handles revenue metrics business logic
type RevenueService struct {
	repo         *repository.RevenueRepository
	cfg          *config.Config
	priceMapping map[string]float64 // Maps price IDs to their monthly value in cents
}

// NewRevenueService creates a new revenue service
func NewRevenueService(
	repo *repository.RevenueRepository,
	cfg *config.Config,
) *RevenueService {
	// Build price mapping from config
	// Prices are in cents, sourced from configuration
	// Monthly prices used directly, yearly prices converted to monthly equivalent
	priceMapping := make(map[string]float64)
	if cfg.Stripe.ProMonthlyPriceID != "" {
		priceMapping[cfg.Stripe.ProMonthlyPriceID] = float64(cfg.Stripe.ProMonthlyPriceCents)
	}
	if cfg.Stripe.ProYearlyPriceID != "" {
		// Convert yearly price to monthly equivalent for MRR calculations
		priceMapping[cfg.Stripe.ProYearlyPriceID] = float64(cfg.Stripe.ProYearlyPriceCents) / 12
	}

	return &RevenueService{
		repo:         repo,
		cfg:          cfg,
		priceMapping: priceMapping,
	}
}

// SetPriceMapping allows updating the price mapping (for testing or dynamic pricing)
func (s *RevenueService) SetPriceMapping(mapping map[string]float64) {
	s.priceMapping = mapping
}

// GetRevenueMetrics returns comprehensive revenue metrics for the admin dashboard
func (s *RevenueService) GetRevenueMetrics(ctx context.Context) (*models.RevenueMetrics, error) {
	// Get the start of this month for period calculations
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	startOfLastMonth := startOfMonth.AddDate(0, -1, 0)

	// Calculate MRR
	mrr, err := s.repo.GetMRR(ctx, s.priceMapping)
	if err != nil {
		return nil, err
	}

	// Get active subscriber count
	activeSubscribers, err := s.repo.GetActiveSubscriberCount(ctx)
	if err != nil {
		return nil, err
	}

	// Get churned subscriber count for this month
	churnedSubscribers, err := s.repo.GetChurnedSubscriberCount(ctx, startOfMonth)
	if err != nil {
		return nil, err
	}

	// Get new subscriber count for this month
	newSubscribers, err := s.repo.GetNewSubscriberCount(ctx, startOfMonth)
	if err != nil {
		return nil, err
	}

	// Get subscribers at start of month (approximation)
	subscribersAtMonthStart := activeSubscribers - newSubscribers + churnedSubscribers

	// Calculate churn rate
	var churnRate float64
	if subscribersAtMonthStart > 0 {
		churnRate = float64(churnedSubscribers) / float64(subscribersAtMonthStart) * 100
	}

	// Calculate ARPU
	var arpu float64
	if activeSubscribers > 0 {
		arpu = mrr / float64(activeSubscribers)
	}

	// Get plan distribution
	planDistribution, err := s.repo.GetPlanDistribution(ctx)
	if err != nil {
		return nil, err
	}

	// Add friendly names and monthly values to plan distribution
	for i := range planDistribution {
		if planDistribution[i].PlanID == s.cfg.Stripe.ProMonthlyPriceID {
			planDistribution[i].PlanName = "Pro Monthly"
			planDistribution[i].MonthlyValue = float64(s.cfg.Stripe.ProMonthlyPriceCents)
		} else if planDistribution[i].PlanID == s.cfg.Stripe.ProYearlyPriceID {
			planDistribution[i].PlanName = "Pro Yearly"
			// Convert yearly price to monthly equivalent
			planDistribution[i].MonthlyValue = float64(s.cfg.Stripe.ProYearlyPriceCents) / 12
		}
	}

	// Get cohort retention (last 6 months)
	cohortRetention, err := s.repo.GetCohortRetention(ctx, 6)
	if err != nil {
		// Non-critical, continue with empty data
		cohortRetention = []models.CohortRetentionMetric{}
	}

	// Get trial conversion rate (last 30 days)
	trialConversionRate, err := s.repo.GetTrialConversionRate(ctx, startOfMonth)
	if err != nil {
		trialConversionRate = 0
	}

	// Get grace period recovery rate
	gracePeriodRecovery, err := s.repo.GetGracePeriodRecoveryRate(ctx, startOfLastMonth)
	if err != nil {
		gracePeriodRecovery = 0
	}

	// Get total revenue
	totalRevenue, err := s.repo.GetTotalRevenue(ctx, s.priceMapping)
	if err != nil {
		totalRevenue = 0
	}

	// Get revenue by month (last 12 months)
	revenueByMonth, err := s.repo.GetRevenueByMonth(ctx, 12, s.priceMapping)
	if err != nil {
		revenueByMonth = []models.RevenueByMonthMetric{}
	}

	// Get subscriber growth trend (last 12 months)
	subscriberGrowth, err := s.repo.GetSubscriberGrowthTrend(ctx, 12)
	if err != nil {
		subscriberGrowth = []models.SubscriberGrowthMetric{}
	}

	// Calculate average lifetime value (simplified)
	// LTV = ARPU / Churn Rate (when churn is monthly)
	var averageLTV float64
	if churnRate > 0 {
		averageLTV = arpu / (churnRate / 100)
	} else if arpu > 0 {
		// If no churn, estimate LTV as 12 months of ARPU
		averageLTV = arpu * 12
	}

	return &models.RevenueMetrics{
		MRR:                  mrr,
		Churn:                churnRate,
		ARPU:                 arpu,
		ActiveSubscribers:    activeSubscribers,
		TotalRevenue:         totalRevenue,
		PlanDistribution:     planDistribution,
		CohortRetention:      cohortRetention,
		ChurnedSubscribers:   churnedSubscribers,
		NewSubscribers:       newSubscribers,
		TrialConversionRate:  trialConversionRate,
		GracePeriodRecovery:  gracePeriodRecovery,
		AverageLifetimeValue: averageLTV,
		RevenueByMonth:       revenueByMonth,
		SubscriberGrowth:     subscriberGrowth,
		UpdatedAt:            now,
	}, nil
}
