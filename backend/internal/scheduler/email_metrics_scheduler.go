package scheduler

import (
	"context"
	"sync"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/services"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

const emailMetricsSchedulerName = "email_metrics"

// EmailMetricsScheduler manages periodic email metrics calculation and alert checking
type EmailMetricsScheduler struct {
	emailMetricsService *services.EmailMetricsService
	metricsInterval     time.Duration // How often to calculate daily metrics
	alertsInterval      time.Duration // How often to check alerts
	cleanupInterval     time.Duration // How often to cleanup old logs
	alertThresholds     *services.AlertThresholds
	stopChan            chan struct{}
	stopOnce            sync.Once
}

// NewEmailMetricsScheduler creates a new email metrics scheduler
func NewEmailMetricsScheduler(
	emailMetricsService *services.EmailMetricsService,
	metricsIntervalHours int,
	alertsIntervalMinutes int,
	cleanupIntervalDays int,
) *EmailMetricsScheduler {
	return &EmailMetricsScheduler{
		emailMetricsService: emailMetricsService,
		metricsInterval:     time.Duration(metricsIntervalHours) * time.Hour,
		alertsInterval:      time.Duration(alertsIntervalMinutes) * time.Minute,
		cleanupInterval:     time.Duration(cleanupIntervalDays) * 24 * time.Hour,
		alertThresholds:     services.DefaultAlertThresholds(),
		stopChan:            make(chan struct{}),
	}
}

// Start begins the periodic metrics calculation, alert checking, and cleanup
func (s *EmailMetricsScheduler) Start(ctx context.Context) {
	logger := utils.GetLogger()
	logger.Info("Starting email metrics scheduler", map[string]interface{}{
		"scheduler":        emailMetricsSchedulerName,
		"metrics_interval": s.metricsInterval.String(),
		"alerts_interval":  s.alertsInterval.String(),
		"cleanup_interval": s.cleanupInterval.String(),
	})

	metricsTicker := time.NewTicker(s.metricsInterval)
	alertsTicker := time.NewTicker(s.alertsInterval)
	cleanupTicker := time.NewTicker(s.cleanupInterval)
	defer metricsTicker.Stop()
	defer alertsTicker.Stop()
	defer cleanupTicker.Stop()

	// Run initial jobs
	go s.calculateDailyMetrics(ctx)
	go s.checkAlerts(ctx)

	for {
		select {
		case <-metricsTicker.C:
			go s.calculateDailyMetrics(ctx)
		case <-alertsTicker.C:
			go s.checkAlerts(ctx)
		case <-cleanupTicker.C:
			go s.cleanupOldLogs(ctx)
		case <-s.stopChan:
			logger.Info("Email metrics scheduler stopped", map[string]interface{}{
				"scheduler": emailMetricsSchedulerName,
			})
			return
		case <-ctx.Done():
			logger.Info("Email metrics scheduler stopped due to context cancellation", map[string]interface{}{
				"scheduler": emailMetricsSchedulerName,
			})
			return
		}
	}
}

// Stop stops the scheduler in a thread-safe manner
func (s *EmailMetricsScheduler) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopChan)
	})
}

// calculateDailyMetrics calculates and stores daily metrics
func (s *EmailMetricsScheduler) calculateDailyMetrics(ctx context.Context) {
	logger := utils.GetLogger()
	logger.Debug("Calculating daily email metrics", map[string]interface{}{
		"scheduler": emailMetricsSchedulerName,
	})

	if err := s.emailMetricsService.CalculateAndStoreDailyMetrics(ctx); err != nil {
		logger.Error("Failed to calculate daily metrics", err, map[string]interface{}{
			"scheduler": emailMetricsSchedulerName,
		})
		return
	}

	logger.Info("Successfully calculated and stored daily email metrics", map[string]interface{}{
		"scheduler": emailMetricsSchedulerName,
	})
}

// checkAlerts checks current metrics against thresholds and triggers alerts
func (s *EmailMetricsScheduler) checkAlerts(ctx context.Context) {
	logger := utils.GetLogger()
	logger.Debug("Checking email metrics for alerts", map[string]interface{}{
		"scheduler": emailMetricsSchedulerName,
	})

	alerts, err := s.emailMetricsService.CheckAlerts(ctx, s.alertThresholds)
	if err != nil {
		logger.Error("Failed to check alerts", err, map[string]interface{}{
			"scheduler": emailMetricsSchedulerName,
		})
		return
	}

	if len(alerts) > 0 {
		logger.Info("Email alerts triggered", map[string]interface{}{
			"scheduler":   emailMetricsSchedulerName,
			"alert_count": len(alerts),
		})
		// Log each alert
		for _, alert := range alerts {
			logger.Warn("Email alert triggered", map[string]interface{}{
				"scheduler":  emailMetricsSchedulerName,
				"alert_type": alert.AlertType,
				"severity":   alert.Severity,
				"message":    alert.Message,
			})
		}
	} else {
		logger.Debug("No email alerts triggered", map[string]interface{}{
			"scheduler": emailMetricsSchedulerName,
		})
	}
}

// cleanupOldLogs removes email logs older than the retention period
func (s *EmailMetricsScheduler) cleanupOldLogs(ctx context.Context) {
	logger := utils.GetLogger()
	logger.Info("Cleaning up old email logs", map[string]interface{}{
		"scheduler": emailMetricsSchedulerName,
	})

	if err := s.emailMetricsService.CleanupOldLogs(ctx); err != nil {
		logger.Error("Failed to cleanup old email logs", err, map[string]interface{}{
			"scheduler": emailMetricsSchedulerName,
		})
		return
	}

	logger.Info("Successfully cleaned up old email logs", map[string]interface{}{
		"scheduler": emailMetricsSchedulerName,
	})
}
