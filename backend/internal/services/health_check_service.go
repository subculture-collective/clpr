package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

// HealthStatus represents the health status of a service
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// RegionHealth represents the health status of a region
type RegionHealth struct {
	Region    string       `json:"region"`
	Status    HealthStatus `json:"status"`
	Latency   float64      `json:"latency_ms"`
	ErrorRate float64      `json:"error_rate"`
	LastCheck time.Time    `json:"last_check"`
	Message   string       `json:"message,omitempty"`
}

// HealthCheckService monitors the health of mirrors and CDN endpoints
type HealthCheckService struct {
	mirrorService *ClipMirrorService
	cdnService    *CDNService
	regions       []string
	healthCache   map[string]*RegionHealth
	mu            sync.RWMutex
}

// NewHealthCheckService creates a new HealthCheckService
func NewHealthCheckService(
	mirrorService *ClipMirrorService,
	cdnService *CDNService,
	regions []string,
) *HealthCheckService {
	return &HealthCheckService{
		mirrorService: mirrorService,
		cdnService:    cdnService,
		regions:       regions,
		healthCache:   make(map[string]*RegionHealth),
	}
}

// CheckRegionHealth checks the health of a specific region
func (s *HealthCheckService) CheckRegionHealth(ctx context.Context, region string) (*RegionHealth, error) {
	start := time.Now()

	health := &RegionHealth{
		Region:    region,
		Status:    HealthStatusHealthy,
		LastCheck: time.Now(),
	}

	// In a real implementation, this would:
	// 1. Ping region-specific endpoints
	// 2. Check database connectivity
	// 3. Verify mirror availability
	// 4. Test CDN performance

	// For now, we'll simulate a health check
	latency := time.Since(start).Milliseconds()
	health.Latency = float64(latency)

	// Determine status based on latency
	if latency > 1000 {
		health.Status = HealthStatusUnhealthy
		health.Message = fmt.Sprintf("High latency: %dms", latency)
	} else if latency > 500 {
		health.Status = HealthStatusDegraded
		health.Message = fmt.Sprintf("Elevated latency: %dms", latency)
	}

	// Cache the result
	s.mu.Lock()
	s.healthCache[region] = health
	s.mu.Unlock()

	return health, nil
}

// CheckAllRegions checks the health of all configured regions
func (s *HealthCheckService) CheckAllRegions(ctx context.Context) ([]*RegionHealth, error) {
	var wg sync.WaitGroup
	results := make([]*RegionHealth, len(s.regions))
	errors := make([]error, len(s.regions))

	for i, region := range s.regions {
		wg.Add(1)
		go func(idx int, r string) {
			defer wg.Done()
			health, err := s.CheckRegionHealth(ctx, r)
			results[idx] = health
			errors[idx] = err
		}(i, region)
	}

	wg.Wait()

	// Check if any errors occurred
	var firstError error
	for _, err := range errors {
		if err != nil {
			firstError = err
			break
		}
	}

	return results, firstError
}

// GetRegionHealth returns cached health status for a region
func (s *HealthCheckService) GetRegionHealth(region string) (*RegionHealth, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	health, ok := s.healthCache[region]
	return health, ok
}

// GetHealthiestRegion returns the healthiest region for failover
func (s *HealthCheckService) GetHealthiestRegion(excludeRegions []string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var bestRegion string
	var bestLatency float64 = -1

	excludeMap := make(map[string]bool)
	for _, r := range excludeRegions {
		excludeMap[r] = true
	}

	for region, health := range s.healthCache {
		if excludeMap[region] {
			continue
		}

		if health.Status == HealthStatusHealthy {
			if bestLatency < 0 || health.Latency < bestLatency {
				bestRegion = region
				bestLatency = health.Latency
			}
		}
	}

	if bestRegion == "" {
		return "", fmt.Errorf("no healthy regions available")
	}

	return bestRegion, nil
}

// StartHealthCheckLoop starts a background loop to continuously check health
func (s *HealthCheckService) StartHealthCheckLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	utils.Info("Starting health check loop", map[string]interface{}{"interval": interval})

	for {
		select {
		case <-ctx.Done():
			utils.Info("Health check loop stopped", nil)
			return
		case <-ticker.C:
			if _, err := s.CheckAllRegions(ctx); err != nil {
				utils.Warn("Health check error", map[string]interface{}{"error": err})
			}
		}
	}
}

// GetOverallHealth returns the overall health status across all regions
func (s *HealthCheckService) GetOverallHealth() HealthStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.healthCache) == 0 {
		return HealthStatusUnhealthy
	}

	healthyCount := 0
	degradedCount := 0
	unhealthyCount := 0

	for _, health := range s.healthCache {
		switch health.Status {
		case HealthStatusHealthy:
			healthyCount++
		case HealthStatusDegraded:
			degradedCount++
		case HealthStatusUnhealthy:
			unhealthyCount++
		}
	}

	total := len(s.healthCache)

	// If more than 50% are unhealthy, overall is unhealthy
	if float64(unhealthyCount)/float64(total) > 0.5 {
		return HealthStatusUnhealthy
	}

	// If more than 30% are degraded or unhealthy, overall is degraded
	if float64(degradedCount+unhealthyCount)/float64(total) > 0.3 {
		return HealthStatusDegraded
	}

	return HealthStatusHealthy
}
