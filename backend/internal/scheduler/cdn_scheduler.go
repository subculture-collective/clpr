package scheduler

import (
	"context"
	"sync"
	"time"

	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

const cdnSchedulerName = "cdn"

// CDNServiceInterface defines the interface required by the CDN scheduler
type CDNServiceInterface interface {
	CollectMetrics(ctx context.Context) error
	CheckCostThreshold(ctx context.Context) (bool, float64, error)
}

// CDNScheduler manages periodic CDN metrics collection
type CDNScheduler struct {
	cdnService      CDNServiceInterface
	metricsInterval time.Duration
	stopChan        chan struct{}
	stopOnce        sync.Once
	running         bool
	mu              sync.Mutex
}

// NewCDNScheduler creates a new CDN scheduler
func NewCDNScheduler(
	cdnService CDNServiceInterface,
	metricsIntervalMinutes int,
) *CDNScheduler {
	return &CDNScheduler{
		cdnService:      cdnService,
		metricsInterval: time.Duration(metricsIntervalMinutes) * time.Minute,
		stopChan:        make(chan struct{}),
	}
}

// Start begins the periodic CDN metrics collection
func (s *CDNScheduler) Start(ctx context.Context) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		utils.Warn("CDN scheduler is already running", map[string]interface{}{
			"scheduler": cdnSchedulerName,
		})
		return
	}
	s.running = true
	s.mu.Unlock()

	utils.Info("Starting CDN scheduler", map[string]interface{}{
		"scheduler":        cdnSchedulerName,
		"metrics_interval": s.metricsInterval.String(),
	})

	metricsTicker := time.NewTicker(s.metricsInterval)
	defer metricsTicker.Stop()
	defer func() {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
	}()

	// Run initial metrics collection
	s.collectMetrics(ctx)

	for {
		select {
		case <-metricsTicker.C:
			s.collectMetrics(ctx)
			s.checkCostThreshold(ctx)
		case <-s.stopChan:
			utils.Info("CDN scheduler stopped", map[string]interface{}{
				"scheduler": cdnSchedulerName,
			})
			return
		case <-ctx.Done():
			utils.Info("CDN scheduler stopped due to context cancellation", map[string]interface{}{
				"scheduler": cdnSchedulerName,
			})
			return
		}
	}
}

// Stop stops the CDN scheduler
func (s *CDNScheduler) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopChan)
	})
}

// collectMetrics performs CDN metrics collection
func (s *CDNScheduler) collectMetrics(ctx context.Context) {
	utils.Info("Collecting CDN metrics", map[string]interface{}{
		"scheduler": cdnSchedulerName,
	})
	start := time.Now()

	if err := s.cdnService.CollectMetrics(ctx); err != nil {
		utils.Error("CDN metrics collection failed", err, map[string]interface{}{
			"scheduler": cdnSchedulerName,
		})
		return
	}

	duration := time.Since(start)
	utils.Info("CDN metrics collection completed", map[string]interface{}{
		"scheduler": cdnSchedulerName,
		"duration":  duration.String(),
	})
}

// checkCostThreshold checks if CDN costs exceed the configured threshold
func (s *CDNScheduler) checkCostThreshold(ctx context.Context) {
	exceeded, costPerGB, err := s.cdnService.CheckCostThreshold(ctx)
	if err != nil {
		utils.Error("CDN cost threshold check failed", err, map[string]interface{}{
			"scheduler": cdnSchedulerName,
		})
		return
	}

	if exceeded {
		utils.Warn("CDN cost per GB exceeds configured threshold", map[string]interface{}{
			"scheduler":   cdnSchedulerName,
			"cost_per_gb": costPerGB,
		})
		// In a real implementation, this would trigger an alert via email/Slack/PagerDuty
	}
}
