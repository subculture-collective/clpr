package services

import (
	"context"
	"fmt"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/repository"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

// EmailMetricsService handles email metrics calculation and monitoring
type EmailMetricsService struct {
	emailLogRepo *repository.EmailLogRepository
	logger       *utils.StructuredLogger
}

// NewEmailMetricsService creates a new EmailMetricsService
func NewEmailMetricsService(emailLogRepo *repository.EmailLogRepository) *EmailMetricsService {
	return &EmailMetricsService{
		emailLogRepo: emailLogRepo,
		logger:       utils.GetLogger(),
	}
}

// AlertThresholds defines the thresholds for triggering alerts
type AlertThresholds struct {
	BounceRateWarning     float64 // 5% - Warning threshold for bounce rate
	BounceRateCritical    float64 // 10% - Critical threshold for bounce rate
	ComplaintRateWarning  float64 // 0.5% - Warning threshold for complaint/spam rate
	ComplaintRateCritical float64 // 1% - Critical threshold for complaint/spam rate
	OpenRateDropWarning   float64 // 20% - Warning threshold for open rate drop
	OpenRateDropCritical  float64 // 40% - Critical threshold for open rate drop
	SendErrorsHourly      int     // 10 - Maximum send errors per hour
	UnsubscribeSpike      float64 // 200% - Spike percentage for unsubscribes
}

// DefaultAlertThresholds returns the default alert thresholds from the issue requirements
func DefaultAlertThresholds() *AlertThresholds {
	return &AlertThresholds{
		BounceRateWarning:     5.0,
		BounceRateCritical:    10.0,
		ComplaintRateWarning:  0.5,
		ComplaintRateCritical: 1.0,
		OpenRateDropWarning:   20.0,
		OpenRateDropCritical:  40.0,
		SendErrorsHourly:      10,
		UnsubscribeSpike:      200.0,
	}
}

// GetMetricsForPeriod retrieves metrics for a specific time period
func (s *EmailMetricsService) GetMetricsForPeriod(ctx context.Context, startTime, endTime time.Time, template *string) (*models.EmailMetricsSummary, error) {
	metrics, err := s.emailLogRepo.GetMetricsForPeriod(ctx, startTime, endTime, template)
	if err != nil {
		s.logger.Error("Failed to get metrics for period", err)
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	return metrics, nil
}

// GetDailyMetrics retrieves metrics for the last N days
func (s *EmailMetricsService) GetDailyMetrics(ctx context.Context, days int) ([]models.EmailMetricsSummary, error) {
	var allMetrics []models.EmailMetricsSummary

	now := time.Now()
	for i := 0; i < days; i++ {
		dayStart := now.AddDate(0, 0, -i).Truncate(24 * time.Hour)
		dayEnd := dayStart.Add(24 * time.Hour)

		metrics, err := s.emailLogRepo.GetMetricsForPeriod(ctx, dayStart, dayEnd, nil)
		if err != nil {
			s.logger.Error("Failed to get daily metrics", err, map[string]interface{}{"day": i})
			continue
		}

		metrics.Granularity = "daily"
		allMetrics = append(allMetrics, *metrics)
	}

	return allMetrics, nil
}

// GetMetricsByTemplate retrieves metrics grouped by template
func (s *EmailMetricsService) GetMetricsByTemplate(ctx context.Context, startTime, endTime time.Time, limit int) ([]models.EmailMetricsSummary, error) {
	metrics, err := s.emailLogRepo.GetMetricsByTemplate(ctx, startTime, endTime, limit)
	if err != nil {
		s.logger.Error("Failed to get metrics by template", err)
		return nil, fmt.Errorf("failed to get metrics by template: %w", err)
	}

	return metrics, nil
}

// CheckAlerts checks current metrics against thresholds and creates alerts if needed
func (s *EmailMetricsService) CheckAlerts(ctx context.Context, thresholds *AlertThresholds) ([]models.EmailAlert, error) {
	var alerts []models.EmailAlert

	// Check hourly metrics
	hourStart := time.Now().Add(-1 * time.Hour)
	hourEnd := time.Now()
	hourlyMetrics, err := s.emailLogRepo.GetMetricsForPeriod(ctx, hourStart, hourEnd, nil)
	if err != nil {
		s.logger.Error("Failed to get hourly metrics for alert check", err)
		return nil, err
	}

	// Check bounce rate
	if hourlyMetrics.BounceRate != nil {
		if *hourlyMetrics.BounceRate >= thresholds.BounceRateCritical {
			alert := models.EmailAlert{
				AlertType:      models.EmailAlertTypeHighBounceRate,
				Severity:       models.EmailAlertSeverityCritical,
				MetricName:     "bounce_rate",
				CurrentValue:   hourlyMetrics.BounceRate,
				ThresholdValue: &thresholds.BounceRateCritical,
				PeriodStart:    hourStart,
				PeriodEnd:      hourEnd,
				Message:        fmt.Sprintf("CRITICAL: Bounce rate is %.2f%% (threshold: %.2f%%)", *hourlyMetrics.BounceRate, thresholds.BounceRateCritical),
			}
			alerts = append(alerts, alert)
			if err := s.emailLogRepo.CreateAlert(ctx, &alert); err != nil {
				s.logger.Error("Failed to create bounce rate critical alert", err)
			}
		} else if *hourlyMetrics.BounceRate >= thresholds.BounceRateWarning {
			alert := models.EmailAlert{
				AlertType:      models.EmailAlertTypeHighBounceRate,
				Severity:       models.EmailAlertSeverityWarning,
				MetricName:     "bounce_rate",
				CurrentValue:   hourlyMetrics.BounceRate,
				ThresholdValue: &thresholds.BounceRateWarning,
				PeriodStart:    hourStart,
				PeriodEnd:      hourEnd,
				Message:        fmt.Sprintf("WARNING: Bounce rate is %.2f%% (threshold: %.2f%%)", *hourlyMetrics.BounceRate, thresholds.BounceRateWarning),
			}
			alerts = append(alerts, alert)
			if err := s.emailLogRepo.CreateAlert(ctx, &alert); err != nil {
				s.logger.Error("Failed to create bounce rate warning alert", err)
			}
		}
	}

	// Check complaint/spam rate
	if hourlyMetrics.SpamRate != nil {
		if *hourlyMetrics.SpamRate >= thresholds.ComplaintRateCritical {
			alert := models.EmailAlert{
				AlertType:      models.EmailAlertTypeHighComplaintRate,
				Severity:       models.EmailAlertSeverityCritical,
				MetricName:     "spam_rate",
				CurrentValue:   hourlyMetrics.SpamRate,
				ThresholdValue: &thresholds.ComplaintRateCritical,
				PeriodStart:    hourStart,
				PeriodEnd:      hourEnd,
				Message:        fmt.Sprintf("CRITICAL: Spam complaint rate is %.2f%% (threshold: %.2f%%)", *hourlyMetrics.SpamRate, thresholds.ComplaintRateCritical),
			}
			alerts = append(alerts, alert)
			if err := s.emailLogRepo.CreateAlert(ctx, &alert); err != nil {
				s.logger.Error("Failed to create complaint rate critical alert", err)
			}
		} else if *hourlyMetrics.SpamRate >= thresholds.ComplaintRateWarning {
			alert := models.EmailAlert{
				AlertType:      models.EmailAlertTypeHighComplaintRate,
				Severity:       models.EmailAlertSeverityWarning,
				MetricName:     "spam_rate",
				CurrentValue:   hourlyMetrics.SpamRate,
				ThresholdValue: &thresholds.ComplaintRateWarning,
				PeriodStart:    hourStart,
				PeriodEnd:      hourEnd,
				Message:        fmt.Sprintf("WARNING: Spam complaint rate is %.2f%% (threshold: %.2f%%)", *hourlyMetrics.SpamRate, thresholds.ComplaintRateWarning),
			}
			alerts = append(alerts, alert)
			if err := s.emailLogRepo.CreateAlert(ctx, &alert); err != nil {
				s.logger.Error("Failed to create complaint rate warning alert", err)
			}
		}
	}

	// Check send errors (dropped emails)
	if hourlyMetrics.TotalDropped >= thresholds.SendErrorsHourly {
		droppedFloat := float64(hourlyMetrics.TotalDropped)
		thresholdFloat := float64(thresholds.SendErrorsHourly)
		alert := models.EmailAlert{
			AlertType:      models.EmailAlertTypeSendErrors,
			Severity:       models.EmailAlertSeverityWarning,
			MetricName:     "dropped_emails",
			CurrentValue:   &droppedFloat,
			ThresholdValue: &thresholdFloat,
			PeriodStart:    hourStart,
			PeriodEnd:      hourEnd,
			Message:        fmt.Sprintf("WARNING: %d emails dropped in the last hour (threshold: %d)", hourlyMetrics.TotalDropped, thresholds.SendErrorsHourly),
		}
		alerts = append(alerts, alert)
		if err := s.emailLogRepo.CreateAlert(ctx, &alert); err != nil {
			s.logger.Error("Failed to create send errors alert", err)
		}
	}

	// Check open rate drop (compare with previous day)
	dayStart := time.Now().AddDate(0, 0, -1).Truncate(24 * time.Hour)
	dayEnd := dayStart.Add(24 * time.Hour)
	yesterdayMetrics, err := s.emailLogRepo.GetMetricsForPeriod(ctx, dayStart, dayEnd, nil)
	if err == nil && yesterdayMetrics.OpenRate != nil && hourlyMetrics.OpenRate != nil {
		if *yesterdayMetrics.OpenRate > 0 {
			dropPercent := (*yesterdayMetrics.OpenRate - *hourlyMetrics.OpenRate) / *yesterdayMetrics.OpenRate * 100
			if dropPercent >= thresholds.OpenRateDropCritical {
				alert := models.EmailAlert{
					AlertType:      models.EmailAlertTypeOpenRateDrop,
					Severity:       models.EmailAlertSeverityCritical,
					MetricName:     "open_rate_drop",
					CurrentValue:   &dropPercent,
					ThresholdValue: &thresholds.OpenRateDropCritical,
					PeriodStart:    hourStart,
					PeriodEnd:      hourEnd,
					Message:        fmt.Sprintf("CRITICAL: Open rate dropped %.2f%% compared to yesterday (threshold: %.2f%%)", dropPercent, thresholds.OpenRateDropCritical),
				}
				alerts = append(alerts, alert)
				if err := s.emailLogRepo.CreateAlert(ctx, &alert); err != nil {
					s.logger.Error("Failed to create open rate drop critical alert", err)
				}
			} else if dropPercent >= thresholds.OpenRateDropWarning {
				alert := models.EmailAlert{
					AlertType:      models.EmailAlertTypeOpenRateDrop,
					Severity:       models.EmailAlertSeverityWarning,
					MetricName:     "open_rate_drop",
					CurrentValue:   &dropPercent,
					ThresholdValue: &thresholds.OpenRateDropWarning,
					PeriodStart:    hourStart,
					PeriodEnd:      hourEnd,
					Message:        fmt.Sprintf("WARNING: Open rate dropped %.2f%% compared to yesterday (threshold: %.2f%%)", dropPercent, thresholds.OpenRateDropWarning),
				}
				alerts = append(alerts, alert)
				if err := s.emailLogRepo.CreateAlert(ctx, &alert); err != nil {
					s.logger.Error("Failed to create open rate drop warning alert", err)
				}
			}
		}
	}

	// Check unsubscribe spike (compare current hour with same hour yesterday)
	// Get metrics for the same hour window yesterday
	prevHourStart := hourStart.AddDate(0, 0, -1)
	prevHourEnd := hourEnd.AddDate(0, 0, -1)
	prevHourMetrics, err := s.emailLogRepo.GetMetricsForPeriod(ctx, prevHourStart, prevHourEnd, nil)
	if err == nil && prevHourMetrics.TotalUnsubscribes > 0 {
		// Calculate increase percentage, safe from division by zero due to check above
		increasePercent := float64(hourlyMetrics.TotalUnsubscribes-prevHourMetrics.TotalUnsubscribes) / float64(prevHourMetrics.TotalUnsubscribes) * 100
		if increasePercent >= thresholds.UnsubscribeSpike {
			alert := models.EmailAlert{
				AlertType:      models.EmailAlertTypeUnsubscribeSpike,
				Severity:       models.EmailAlertSeverityWarning,
				MetricName:     "unsubscribe_spike",
				CurrentValue:   &increasePercent,
				ThresholdValue: &thresholds.UnsubscribeSpike,
				PeriodStart:    hourStart,
				PeriodEnd:      hourEnd,
				Message:        fmt.Sprintf("WARNING: Unsubscribe rate spiked %.2f%% compared to same hour yesterday (threshold: %.2f%%)", increasePercent, thresholds.UnsubscribeSpike),
			}
			alerts = append(alerts, alert)
			if err := s.emailLogRepo.CreateAlert(ctx, &alert); err != nil {
				s.logger.Error("Failed to create unsubscribe spike alert", err)
			}
		}
	}

	s.logger.Info("Alert check completed", map[string]interface{}{"alert_count": len(alerts)})
	return alerts, nil
}

// GetUnresolvedAlerts retrieves all unresolved alerts
func (s *EmailMetricsService) GetUnresolvedAlerts(ctx context.Context, limit int) ([]models.EmailAlert, error) {
	alerts, err := s.emailLogRepo.GetUnresolvedAlerts(ctx, limit)
	if err != nil {
		s.logger.Error("Failed to get unresolved alerts", err)
		return nil, fmt.Errorf("failed to get unresolved alerts: %w", err)
	}

	return alerts, nil
}

// CalculateAndStoreDailyMetrics calculates and stores daily metrics (for scheduled jobs)
func (s *EmailMetricsService) CalculateAndStoreDailyMetrics(ctx context.Context) error {
	// Calculate metrics for yesterday
	yesterday := time.Now().AddDate(0, 0, -1).Truncate(24 * time.Hour)
	dayEnd := yesterday.Add(24 * time.Hour)

	metrics, err := s.emailLogRepo.GetMetricsForPeriod(ctx, yesterday, dayEnd, nil)
	if err != nil {
		s.logger.Error("Failed to calculate daily metrics", err)
		return err
	}

	metrics.Granularity = "daily"
	if err := s.emailLogRepo.CreateMetricsSummary(ctx, metrics); err != nil {
		s.logger.Error("Failed to store daily metrics", err)
		return err
	}

	s.logger.Info("Daily metrics calculated and stored", map[string]interface{}{"date": yesterday.Format("2006-01-02")})
	return nil
}

// CleanupOldLogs deletes email logs older than the retention period (90 days per requirements)
func (s *EmailMetricsService) CleanupOldLogs(ctx context.Context) error {
	retentionPeriod := 90 * 24 * time.Hour // 90 days as per requirements

	deleted, err := s.emailLogRepo.DeleteOldLogs(ctx, retentionPeriod)
	if err != nil {
		s.logger.Error("Failed to cleanup old email logs", err)
		return err
	}

	s.logger.Info("Old email logs cleaned up", map[string]interface{}{"deleted_count": deleted})
	return nil
}
