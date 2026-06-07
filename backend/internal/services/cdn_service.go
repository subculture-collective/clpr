package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// CDNProvider is the interface that all CDN providers must implement
type CDNProvider interface {
	// GenerateURL generates a CDN URL for a clip
	GenerateURL(clip *models.Clip) (string, error)

	// PurgeCache purges the cache for a specific clip
	PurgeCache(clipURL string) error

	// GetCacheHeaders returns appropriate cache headers for video content
	GetCacheHeaders() map[string]string

	// GetMetrics retrieves current metrics from the CDN
	GetMetrics(ctx context.Context) (*CDNProviderMetrics, error)
}

// CDNProviderMetrics represents metrics from a CDN provider
type CDNProviderMetrics struct {
	Bandwidth    float64 // GB transferred
	Requests     int64   // Total requests
	CacheHitRate float64 // Percentage (0-100)
	AvgLatencyMs float64 // Average latency in milliseconds
	CostUSD      float64 // Estimated cost in USD
}

// CDNRepositoryContract captures repository calls CDNService relies on.
type CDNRepositoryContract interface {
	CreateMetric(ctx context.Context, metric *models.CDNMetrics) error
	GetTotalCost(ctx context.Context, startTime time.Time, endTime time.Time) (float64, error)
	GetCacheHitRate(ctx context.Context, provider string, startTime time.Time) (float64, error)
	GetMetricsSummary(ctx context.Context, provider string, metricType string, startTime time.Time) (float64, error)
}

// CDNService manages CDN operations
type CDNService struct {
	cdnRepo   CDNRepositoryContract
	config    *config.CDNConfig
	providers map[string]CDNProvider
}

// NewCDNService creates a new CDNService
func NewCDNService(
	cdnRepo CDNRepositoryContract,
	config *config.CDNConfig,
) *CDNService {
	service := &CDNService{
		cdnRepo:   cdnRepo,
		config:    config,
		providers: make(map[string]CDNProvider),
	}

	// Initialize providers based on configuration
	if config.Enabled {
		service.initializeProviders()
	}

	return service
}

// initializeProviders initializes CDN providers based on configuration
func (s *CDNService) initializeProviders() {
	switch s.config.Provider {
	case models.CDNProviderCloudflare:
		if s.config.CloudflareZoneID != "" && s.config.CloudflareAPIKey != "" {
			s.providers[models.CDNProviderCloudflare] = NewCloudflareProvider(
				s.config.CloudflareZoneID,
				s.config.CloudflareAPIKey,
				s.config.CacheTTL,
			)
			log.Println("Initialized Cloudflare CDN provider")
		}
	case models.CDNProviderBunny:
		if s.config.BunnyAPIKey != "" && s.config.BunnyStorageZone != "" {
			s.providers[models.CDNProviderBunny] = NewBunnyProvider(
				s.config.BunnyAPIKey,
				s.config.BunnyStorageZone,
				s.config.CacheTTL,
			)
			log.Println("Initialized Bunny CDN provider")
		}
	case models.CDNProviderAWSCloudFront:
		if s.config.AWSAccessKeyID != "" && s.config.AWSSecretKey != "" {
			s.providers[models.CDNProviderAWSCloudFront] = NewAWSCloudFrontProvider(
				s.config.AWSAccessKeyID,
				s.config.AWSSecretKey,
				s.config.AWSRegion,
				s.config.CacheTTL,
			)
			log.Println("Initialized AWS CloudFront CDN provider")
		}
	}
}

// GetCDNURL generates a CDN URL for a clip
func (s *CDNService) GetCDNURL(ctx context.Context, clip *models.Clip) (string, error) {
	if !s.config.Enabled {
		return "", nil
	}

	// Get the active provider
	provider, ok := s.providers[s.config.Provider]
	if !ok {
		return "", fmt.Errorf("CDN provider %s not initialized", s.config.Provider)
	}

	// Generate URL
	cdnURL, err := provider.GenerateURL(clip)
	if err != nil {
		return "", fmt.Errorf("failed to generate CDN URL: %w", err)
	}

	// Record metric
	if err := s.recordMetric(ctx, s.config.Provider, models.CDNMetricTypeRequests, 1); err != nil {
		log.Printf("Failed to record CDN metric: %v", err)
	}

	return cdnURL, nil
}

// PurgeCache purges the cache for a specific clip
func (s *CDNService) PurgeCache(ctx context.Context, clip *models.Clip) error {
	if !s.config.Enabled {
		return nil
	}

	provider, ok := s.providers[s.config.Provider]
	if !ok {
		return fmt.Errorf("CDN provider %s not initialized", s.config.Provider)
	}

	// Get the CDN URL
	cdnURL, err := provider.GenerateURL(clip)
	if err != nil {
		return fmt.Errorf("failed to generate CDN URL: %w", err)
	}

	// Purge cache
	if err := provider.PurgeCache(cdnURL); err != nil {
		return fmt.Errorf("failed to purge cache: %w", err)
	}

	log.Printf("Successfully purged cache for clip %s", clip.ID)
	return nil
}

// GetCacheHeaders returns appropriate cache headers for video content
func (s *CDNService) GetCacheHeaders() map[string]string {
	if !s.config.Enabled {
		return map[string]string{
			"Cache-Control": "public, max-age=3600",
		}
	}

	provider, ok := s.providers[s.config.Provider]
	if !ok {
		return map[string]string{
			"Cache-Control": fmt.Sprintf("public, max-age=%d", s.config.CacheTTL),
		}
	}

	return provider.GetCacheHeaders()
}

// CollectMetrics collects metrics from all active CDN providers
func (s *CDNService) CollectMetrics(ctx context.Context) error {
	if !s.config.Enabled {
		return nil
	}

	for providerName, provider := range s.providers {
		metrics, err := provider.GetMetrics(ctx)
		if err != nil {
			log.Printf("Failed to collect metrics from %s: %v", providerName, err)
			continue
		}

		// Store metrics
		if err := s.recordMetric(ctx, providerName, models.CDNMetricTypeBandwidth, metrics.Bandwidth); err != nil {
			log.Printf("Failed to record bandwidth metric: %v", err)
		}

		if err := s.recordMetric(ctx, providerName, models.CDNMetricTypeRequests, float64(metrics.Requests)); err != nil {
			log.Printf("Failed to record requests metric: %v", err)
		}

		if err := s.recordMetric(ctx, providerName, models.CDNMetricTypeCacheHitRate, metrics.CacheHitRate); err != nil {
			log.Printf("Failed to record cache hit rate metric: %v", err)
		}

		if err := s.recordMetric(ctx, providerName, models.CDNMetricTypeLatency, metrics.AvgLatencyMs); err != nil {
			log.Printf("Failed to record latency metric: %v", err)
		}

		if err := s.recordMetric(ctx, providerName, models.CDNMetricTypeCost, metrics.CostUSD); err != nil {
			log.Printf("Failed to record cost metric: %v", err)
		}

		log.Printf("Collected metrics from %s: Bandwidth=%.2fGB, Requests=%d, HitRate=%.2f%%, Latency=%.2fms, Cost=$%.4f",
			providerName, metrics.Bandwidth, metrics.Requests, metrics.CacheHitRate, metrics.AvgLatencyMs, metrics.CostUSD)
	}

	return nil
}

// GetCostMetrics returns CDN costs for a time period
func (s *CDNService) GetCostMetrics(ctx context.Context, startTime, endTime time.Time) (float64, error) {
	if !s.config.Enabled {
		return 0, nil
	}

	totalCost, err := s.cdnRepo.GetTotalCost(ctx, startTime, endTime)
	if err != nil {
		return 0, fmt.Errorf("failed to get CDN costs: %w", err)
	}

	return totalCost, nil
}

// GetCacheHitRate returns the average cache hit rate for a time period
func (s *CDNService) GetCacheHitRate(ctx context.Context, startTime time.Time) (float64, error) {
	if !s.config.Enabled {
		return 0, nil
	}

	hitRate, err := s.cdnRepo.GetCacheHitRate(ctx, s.config.Provider, startTime)
	if err != nil {
		return 0, fmt.Errorf("failed to get cache hit rate: %w", err)
	}

	return hitRate, nil
}

// CheckCostThreshold checks if CDN costs exceed the configured threshold
func (s *CDNService) CheckCostThreshold(ctx context.Context) (bool, float64, error) {
	if !s.config.Enabled {
		return false, 0, nil
	}

	// Get costs for the last 30 days
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -30)

	totalCost, err := s.GetCostMetrics(ctx, startTime, endTime)
	if err != nil {
		return false, 0, err
	}

	// Get total bandwidth for the period
	bandwidth, err := s.cdnRepo.GetMetricsSummary(ctx, s.config.Provider, models.CDNMetricTypeBandwidth, startTime)
	if err != nil {
		return false, 0, fmt.Errorf("failed to get bandwidth metrics: %w", err)
	}

	if bandwidth == 0 {
		return false, 0, nil
	}

	// Calculate cost per GB
	costPerGB := totalCost / bandwidth
	exceeded := costPerGB > s.config.MaxCostPerGB

	if exceeded {
		log.Printf("WARNING: CDN cost per GB ($%.4f) exceeds threshold ($%.4f)", costPerGB, s.config.MaxCostPerGB)
	}

	return exceeded, costPerGB, nil
}

// recordMetric records a CDN metric
func (s *CDNService) recordMetric(ctx context.Context, provider string, metricType string, value float64) error {
	metric := &models.CDNMetrics{
		ID:          uuid.New(),
		Provider:    provider,
		MetricType:  metricType,
		MetricValue: value,
		RecordedAt:  time.Now(),
	}

	return s.cdnRepo.CreateMetric(ctx, metric)
}
