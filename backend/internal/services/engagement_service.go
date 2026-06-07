package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
)

// Content engagement normalization benchmarks (90th percentile)
// These values represent high-performing content and should be periodically
// reviewed and updated based on actual platform metrics
const (
	// BenchmarkViews is the view count at which content receives 100 engagement points for views
	BenchmarkViews = 10000

	// BenchmarkComments is the comment count at which content receives 100 engagement points for comments
	BenchmarkComments = 100

	// BenchmarkShares is the share count at which content receives 100 engagement points for shares
	BenchmarkShares = 50
)

// EngagementService handles engagement metrics calculations
type EngagementService struct {
	analyticsRepo *repository.AnalyticsRepository
	userRepo      *repository.UserRepository
	clipRepo      *repository.ClipRepository
}

// NewEngagementService creates a new engagement service
func NewEngagementService(
	analyticsRepo *repository.AnalyticsRepository,
	userRepo *repository.UserRepository,
	clipRepo *repository.ClipRepository,
) *EngagementService {
	return &EngagementService{
		analyticsRepo: analyticsRepo,
		userRepo:      userRepo,
		clipRepo:      clipRepo,
	}
}

// GetUserEngagementScore calculates and returns a user's engagement score
func (s *EngagementService) GetUserEngagementScore(ctx context.Context, userID uuid.UUID) (*models.UserEngagementScore, error) {
	now := time.Now()

	// Get activity counts for the user
	postsCount, err := s.analyticsRepo.GetUserPostsCount(ctx, userID, 7) // Last 7 days
	if err != nil {
		return nil, err
	}

	commentsCount, err := s.analyticsRepo.GetUserCommentsCount(ctx, userID, 7)
	if err != nil {
		return nil, err
	}

	votesCount, err := s.analyticsRepo.GetUserVotesCount(ctx, userID, 7)
	if err != nil {
		return nil, err
	}

	loginDays, err := s.analyticsRepo.GetUserLoginDays(ctx, userID, 30)
	if err != nil {
		return nil, err
	}

	avgDailyMinutes, err := s.analyticsRepo.GetUserAvgDailyMinutes(ctx, userID, 30)
	if err != nil {
		return nil, err
	}

	// Calculate component scores
	postsScore := calculateComponentScore(float64(postsCount), 10.0)
	commentsScore := calculateComponentScore(float64(commentsCount), 20.0)
	votesScore := calculateComponentScore(float64(votesCount), 50.0)
	loginScore := calculateComponentScore(float64(loginDays), 30.0)
	timeScore := calculateComponentScore(avgDailyMinutes, 60.0)

	// Calculate weighted total score
	totalScore := int(
		float64(postsScore)*0.20 +
			float64(commentsScore)*0.25 +
			float64(votesScore)*0.20 +
			float64(loginScore)*0.20 +
			float64(timeScore)*0.15,
	)

	// Determine tier
	tier := determineEngagementTier(totalScore)

	score := &models.UserEngagementScore{
		UserID: userID,
		Score:  totalScore,
		Tier:   tier,
		Components: models.UserEngagementComponents{
			Posts: models.EngagementComponent{
				Score:  postsScore,
				Count:  postsCount,
				Weight: 0.20,
			},
			Comments: models.EngagementComponent{
				Score:  commentsScore,
				Count:  commentsCount,
				Weight: 0.25,
			},
			Votes: models.EngagementComponent{
				Score:  votesScore,
				Count:  votesCount,
				Weight: 0.20,
			},
			LoginFrequency: models.EngagementComponent{
				Score:  loginScore,
				Count:  loginDays,
				Weight: 0.20,
			},
			TimeSpent: models.EngagementComponent{
				Score:  timeScore,
				Count:  int(avgDailyMinutes),
				Weight: 0.15,
			},
		},
		CalculatedAt: now,
		UpdatedAt:    now,
	}

	return score, nil
}

// GetPlatformHealthMetrics returns platform-wide health metrics
func (s *EngagementService) GetPlatformHealthMetrics(ctx context.Context) (*models.PlatformHealthMetrics, error) {
	now := time.Now()

	// Get active user counts
	dau, err := s.analyticsRepo.GetDAU(ctx)
	if err != nil {
		return nil, err
	}

	wau, err := s.analyticsRepo.GetWAU(ctx)
	if err != nil {
		return nil, err
	}

	mau, err := s.analyticsRepo.GetMAU(ctx)
	if err != nil {
		return nil, err
	}

	// Calculate stickiness
	stickiness := 0.0
	if mau > 0 {
		stickiness = float64(dau) / float64(mau)
	}

	// Get retention rates
	day1Retention, err := s.analyticsRepo.GetRetentionRate(ctx, 1)
	if err != nil {
		return nil, err
	}

	day7Retention, err := s.analyticsRepo.GetRetentionRate(ctx, 7)
	if err != nil {
		return nil, err
	}

	day30Retention, err := s.analyticsRepo.GetRetentionRate(ctx, 30)
	if err != nil {
		return nil, err
	}

	// Get churn rate
	churnRate, err := s.analyticsRepo.GetMonthlyChurnRate(ctx)
	if err != nil {
		return nil, err
	}

	// Get trends
	dauChangeWoW, err := s.analyticsRepo.GetDAUChangeWoW(ctx)
	if err != nil {
		return nil, err
	}

	mauChangeMoM, err := s.analyticsRepo.GetMAUChangeMoM(ctx)
	if err != nil {
		return nil, err
	}

	metrics := &models.PlatformHealthMetrics{
		DAU:        dau,
		WAU:        wau,
		MAU:        mau,
		Stickiness: stickiness,
		RetentionRates: models.RetentionRates{
			Day1:  day1Retention,
			Day7:  day7Retention,
			Day30: day30Retention,
		},
		ChurnRateMonthly: churnRate,
		Trends: models.PlatformTrends{
			DAUChangeWoW: dauChangeWoW,
			MAUChangeMoM: mauChangeMoM,
		},
		CalculatedAt: now,
	}

	return metrics, nil
}

// GetTrendingMetrics returns trending data with week-over-week changes
func (s *EngagementService) GetTrendingMetrics(ctx context.Context, metric string, days int) (*models.TrendingMetrics, error) {
	if days <= 0 || days > 365 {
		days = 7
	}

	// Get trend data points
	dataPoints, err := s.analyticsRepo.GetTrendingData(ctx, metric, days)
	if err != nil {
		return nil, err
	}

	if len(dataPoints) == 0 {
		return &models.TrendingMetrics{
			Metric:     metric,
			PeriodDays: days,
			Data:       []models.TrendingDataPoint{},
			Summary:    models.TrendSummary{},
		}, nil
	}

	// Calculate changes from previous
	trendingPoints := make([]models.TrendingDataPoint, len(dataPoints))
	for i, point := range dataPoints {
		prevValue := int64(0)
		if i > 0 {
			prevValue = dataPoints[i-1].Value
		}
		trendingPoints[i].FromTrendDataPoint(point, prevValue)
	}

	// Calculate week-over-week change
	weekOverWeekChange := 0.0
	if len(trendingPoints) >= 14 {
		lastWeekAvg := calculateAverage(trendingPoints[len(trendingPoints)-7:])
		prevWeekAvg := calculateAverage(trendingPoints[len(trendingPoints)-14 : len(trendingPoints)-7])
		if prevWeekAvg > 0 {
			weekOverWeekChange = ((lastWeekAvg - prevWeekAvg) / prevWeekAvg) * 100
		}
	}

	// Calculate summary
	summary := calculateTrendSummary(trendingPoints)

	return &models.TrendingMetrics{
		Metric:             metric,
		PeriodDays:         days,
		Data:               trendingPoints,
		WeekOverWeekChange: weekOverWeekChange,
		Summary:            summary,
	}, nil
}

// GetContentEngagementScore calculates engagement score for a clip
func (s *EngagementService) GetContentEngagementScore(ctx context.Context, clipID uuid.UUID) (*models.ContentEngagementScore, error) {
	clip, err := s.clipRepo.GetByID(ctx, clipID)
	if err != nil {
		return nil, err
	}

	// Get view count from analytics
	viewCount, err := s.analyticsRepo.GetClipViewCount(ctx, clipID)
	if err != nil {
		return nil, err
	}

	// Get share count from analytics
	shareCount, err := s.analyticsRepo.GetClipShareCount(ctx, clipID)
	if err != nil {
		return nil, err
	}

	// Calculate vote ratio as defined in documentation: upvotes / (upvotes + downvotes)
	upvotes, downvotes, err := s.analyticsRepo.GetClipVoteCounts(ctx, clipID)
	if err != nil {
		return nil, err
	}

	voteRatio := 0.5 // Default neutral if no votes
	totalVotes := upvotes + downvotes
	if totalVotes > 0 {
		voteRatio = float64(upvotes) / float64(totalVotes)
	}

	// Calculate favorite rate
	favoriteRate := 0.0
	if viewCount > 0 {
		favoriteRate = (float64(clip.FavoriteCount) / float64(viewCount)) * 100
	}

	// Normalize values using benchmark constants (90th percentile values)
	// These benchmarks should be periodically reviewed and updated based on platform growth
	normalizedViews := normalizeMetric(viewCount, BenchmarkViews)
	normalizedComments := normalizeMetric(int64(clip.CommentCount), BenchmarkComments)
	normalizedShares := normalizeMetric(shareCount, BenchmarkShares)

	// Calculate composite score
	score := int(
		float64(normalizedViews)*0.25 +
			voteRatio*100*0.30 +
			float64(normalizedComments)*0.20 +
			float64(normalizedShares)*0.15 +
			favoriteRate*0.10,
	)

	return &models.ContentEngagementScore{
		ClipID:             clipID,
		Score:              score,
		NormalizedViews:    normalizedViews,
		VoteRatio:          voteRatio,
		NormalizedComments: normalizedComments,
		NormalizedShares:   normalizedShares,
		FavoriteRate:       favoriteRate,
		CalculatedAt:       time.Now(),
	}, nil
}

// CheckAlertThresholds checks if any metrics breach alert thresholds
func (s *EngagementService) CheckAlertThresholds(ctx context.Context) ([]*models.EngagementAlert, error) {
	alerts := []*models.EngagementAlert{}
	now := time.Now()

	// Check DAU drop
	dauChangeWoW, err := s.analyticsRepo.GetDAUChangeWoW(ctx)
	if err == nil && dauChangeWoW < -20 {
		alerts = append(alerts, &models.EngagementAlert{
			ID:             uuid.New(),
			AlertType:      "dau_drop",
			Severity:       "P1",
			Metric:         "dau",
			CurrentValue:   dauChangeWoW,
			ThresholdValue: -20.0,
			Message:        "Critical: DAU dropped more than 20% week-over-week",
			TriggeredAt:    now,
		})
	} else if err == nil && dauChangeWoW < -10 {
		alerts = append(alerts, &models.EngagementAlert{
			ID:             uuid.New(),
			AlertType:      "dau_drop",
			Severity:       "P2",
			Metric:         "dau",
			CurrentValue:   dauChangeWoW,
			ThresholdValue: -10.0,
			Message:        "Warning: DAU dropped more than 10% week-over-week",
			TriggeredAt:    now,
		})
	}

	// Check churn rate
	churnRate, err := s.analyticsRepo.GetMonthlyChurnRate(ctx)
	if err == nil && churnRate > 7 {
		alerts = append(alerts, &models.EngagementAlert{
			ID:             uuid.New(),
			AlertType:      "churn_spike",
			Severity:       "P1",
			Metric:         "churn_rate",
			CurrentValue:   churnRate,
			ThresholdValue: 7.0,
			Message:        "Critical: Monthly churn rate exceeded 7%",
			TriggeredAt:    now,
		})
	} else if err == nil && churnRate > 5 {
		alerts = append(alerts, &models.EngagementAlert{
			ID:             uuid.New(),
			AlertType:      "churn_increase",
			Severity:       "P2",
			Metric:         "churn_rate",
			CurrentValue:   churnRate,
			ThresholdValue: 5.0,
			Message:        "Warning: Monthly churn rate exceeded 5%",
			TriggeredAt:    now,
		})
	}

	// Check stickiness
	dau, _ := s.analyticsRepo.GetDAU(ctx)
	mau, _ := s.analyticsRepo.GetMAU(ctx)
	stickiness := 0.0
	if mau > 0 {
		stickiness = float64(dau) / float64(mau)
	}

	if stickiness < 0.15 {
		alerts = append(alerts, &models.EngagementAlert{
			ID:             uuid.New(),
			AlertType:      "stickiness_drop",
			Severity:       "P1",
			Metric:         "stickiness",
			CurrentValue:   stickiness,
			ThresholdValue: 0.15,
			Message:        "Critical: Platform stickiness dropped below 15%",
			TriggeredAt:    now,
		})
	}

	return alerts, nil
}

// Helper functions

func calculateComponentScore(actual, max float64) int {
	if max == 0 {
		return 0
	}
	score := int((actual / max) * 100)
	if score > 100 {
		score = 100
	}
	return score
}

func determineEngagementTier(score int) string {
	switch {
	case score >= 91:
		return "Very High Engagement"
	case score >= 76:
		return "High Engagement"
	case score >= 51:
		return "Moderate Engagement"
	case score >= 26:
		return "Low Engagement"
	default:
		return "Inactive"
	}
}

func calculateAverage(points []models.TrendingDataPoint) float64 {
	if len(points) == 0 {
		return 0
	}
	sum := int64(0)
	for _, p := range points {
		sum += p.Value
	}
	return float64(sum) / float64(len(points))
}

func calculateTrendSummary(points []models.TrendingDataPoint) models.TrendSummary {
	if len(points) == 0 {
		return models.TrendSummary{}
	}

	min := points[0].Value
	max := points[0].Value
	sum := int64(0)

	for _, p := range points {
		if p.Value < min {
			min = p.Value
		}
		if p.Value > max {
			max = p.Value
		}
		sum += p.Value
	}

	avg := sum / int64(len(points))

	// Determine trend direction
	trend := "stable"
	if len(points) >= 2 {
		firstHalf := calculateAverage(points[:len(points)/2])
		secondHalf := calculateAverage(points[len(points)/2:])
		var change float64
		if firstHalf != 0 {
			change = ((secondHalf - firstHalf) / firstHalf) * 100
		} else {
			change = 0
		}

		if change > 5 {
			trend = "increasing"
		} else if change < -5 {
			trend = "decreasing"
		}
	}

	return models.TrendSummary{
		Min:   min,
		Max:   max,
		Avg:   avg,
		Trend: trend,
	}
}

func normalizeMetric(value, max int64) int {
	if max == 0 {
		return 0
	}
	normalized := int((float64(value) / float64(max)) * 100)
	if normalized > 100 {
		normalized = 100
	}
	return normalized
}
