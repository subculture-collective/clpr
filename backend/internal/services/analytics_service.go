package services

import (
	"context"
	"encoding/json"
	"net"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
)

// AnalyticsService handles analytics business logic
type AnalyticsService struct {
	analyticsRepo *repository.AnalyticsRepository
	clipRepo      *repository.ClipRepository
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(analyticsRepo *repository.AnalyticsRepository, clipRepo *repository.ClipRepository) *AnalyticsService {
	return &AnalyticsService{
		analyticsRepo: analyticsRepo,
		clipRepo:      clipRepo,
	}
}

// TrackEvent records an analytics event
func (s *AnalyticsService) TrackEvent(ctx context.Context, eventType string, userID *uuid.UUID, clipID *uuid.UUID, metadata map[string]interface{}, ipAddress, userAgent, referrer string) error {
	// Anonymize IP address (remove last octet for privacy)
	anonymizedIP := anonymizeIP(ipAddress)

	// Convert metadata to JSON string
	var metadataJSON *string
	if metadata != nil {
		data, err := json.Marshal(metadata)
		if err != nil {
			return err
		}
		str := string(data)
		metadataJSON = &str
	}

	event := &models.AnalyticsEvent{
		EventType: eventType,
		UserID:    userID,
		ClipID:    clipID,
		Metadata:  metadataJSON,
		IPAddress: &anonymizedIP,
		UserAgent: &userAgent,
		Referrer:  &referrer,
	}

	return s.analyticsRepo.TrackEvent(ctx, event)
}

// GetCreatorAnalyticsOverview retrieves summary metrics for a creator
func (s *AnalyticsService) GetCreatorAnalyticsOverview(ctx context.Context, creatorName string) (*models.CreatorAnalyticsOverview, error) {
	analytics, err := s.analyticsRepo.GetCreatorAnalytics(ctx, creatorName)
	if err != nil {
		return nil, err
	}

	engagementRate := 0.0
	if analytics.AvgEngagementRate != nil {
		engagementRate = *analytics.AvgEngagementRate
	}

	return &models.CreatorAnalyticsOverview{
		TotalClips:        analytics.TotalClips,
		TotalViews:        analytics.TotalViews,
		TotalUpvotes:      analytics.TotalUpvotes,
		TotalComments:     analytics.TotalComments,
		AvgEngagementRate: engagementRate,
		FollowerCount:     analytics.FollowerCount,
	}, nil
}

// GetCreatorTopClips retrieves top-performing clips for a creator
func (s *AnalyticsService) GetCreatorTopClips(ctx context.Context, creatorName string, sortBy string, limit int) ([]models.CreatorTopClip, error) {
	if sortBy == "" {
		sortBy = "votes"
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	return s.analyticsRepo.GetCreatorTopClips(ctx, creatorName, sortBy, limit)
}

// GetCreatorTrends retrieves time-series data for creator metrics
func (s *AnalyticsService) GetCreatorTrends(ctx context.Context, creatorName string, metricType string, days int) ([]models.TrendDataPoint, error) {
	if days <= 0 || days > 365 {
		days = 30
	}

	return s.analyticsRepo.GetCreatorTrends(ctx, creatorName, metricType, days)
}

// GetClipAnalytics retrieves analytics for a specific clip
func (s *AnalyticsService) GetClipAnalytics(ctx context.Context, clipID uuid.UUID) (*models.ClipAnalytics, error) {
	return s.analyticsRepo.GetClipAnalytics(ctx, clipID)
}

// GetUserAnalytics retrieves personal statistics for a user
func (s *AnalyticsService) GetUserAnalytics(ctx context.Context, userID uuid.UUID) (*models.UserAnalytics, error) {
	return s.analyticsRepo.GetUserAnalytics(ctx, userID)
}

// GetPlatformOverview retrieves current platform KPIs for admin dashboard
func (s *AnalyticsService) GetPlatformOverview(ctx context.Context) (*models.PlatformOverviewMetrics, error) {
	return s.analyticsRepo.GetPlatformOverviewMetrics(ctx)
}

// GetContentMetrics retrieves content-related metrics for admin dashboard
func (s *AnalyticsService) GetContentMetrics(ctx context.Context) (*models.ContentMetrics, error) {
	// Get most popular games
	games, err := s.analyticsRepo.GetMostPopularGames(ctx, 10)
	if err != nil {
		return nil, err
	}

	// Get most popular creators
	creators, err := s.analyticsRepo.GetMostPopularCreators(ctx, 10)
	if err != nil {
		return nil, err
	}

	// Get trending tags (last 7 days)
	tags, err := s.analyticsRepo.GetTrendingTags(ctx, 7, 10)
	if err != nil {
		return nil, err
	}

	// Calculate average clip vote score
	avgVoteScore, err := s.analyticsRepo.GetAverageClipVoteScore(ctx)
	if err != nil {
		return nil, err
	}

	return &models.ContentMetrics{
		MostPopularGames:    games,
		MostPopularCreators: creators,
		TrendingTags:        tags,
		AvgClipVoteScore:    avgVoteScore,
	}, nil
}

// GetPlatformTrends retrieves time-series data for platform metrics
func (s *AnalyticsService) GetPlatformTrends(ctx context.Context, metricType string, days int) ([]models.TrendDataPoint, error) {
	if days <= 0 || days > 365 {
		days = 30
	}

	return s.analyticsRepo.GetPlatformTrends(ctx, metricType, days)
}

// GetCreatorAudienceInsights retrieves audience insights for a creator
func (s *AnalyticsService) GetCreatorAudienceInsights(ctx context.Context, creatorName string, limit int) (*models.CreatorAudienceInsights, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	return s.analyticsRepo.GetCreatorAudienceInsights(ctx, creatorName, limit)
}

// anonymizeIP anonymizes an IP address for privacy using Go's net package.
// For IPv4 addresses, the last octet (8 bits) is zeroed out.
// For IPv6 addresses, the last 80 bits are zeroed out.
func anonymizeIP(ip string) string {
	if ip == "" {
		return ""
	}

	// Parse the IP address using Go's standard library
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		// Invalid IP address
		return "invalid"
	}

	// Check if it's an IPv4 address
	if ipv4 := parsedIP.To4(); ipv4 != nil {
		// Zero out the last octet (8 bits) of IPv4
		ipv4[3] = 0
		return ipv4.String()
	}

	// It's an IPv6 address - zero out the last 80 bits (10 bytes)
	// IPv6 addresses are 128 bits (16 bytes), so we keep the first 48 bits (6 bytes)
	ipv6 := parsedIP.To16()
	if ipv6 != nil {
		for i := 6; i < 16; i++ {
			ipv6[i] = 0
		}
		return ipv6.String()
	}

	// Fallback for unexpected cases
	return "invalid"
}
