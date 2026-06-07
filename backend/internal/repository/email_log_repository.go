package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// EmailLogRepository handles email log data operations
type EmailLogRepository struct {
	db *pgxpool.Pool
}

// NewEmailLogRepository creates a new EmailLogRepository
func NewEmailLogRepository(db *pgxpool.Pool) *EmailLogRepository {
	return &EmailLogRepository{db: db}
}

// CreateEmailLog creates a new email log entry
func (r *EmailLogRepository) CreateEmailLog(ctx context.Context, log *models.EmailLog) error {
	query := `
		INSERT INTO email_logs (
			id, user_id, template, recipient, status, event_type,
			sendgrid_message_id, sendgrid_event_id, bounce_type, bounce_reason,
			spam_report_reason, link_url, ip_address, user_agent, metadata,
			sent_at, delivered_at, opened_at, clicked_at, bounced_at,
			spam_reported_at, unsubscribed_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
			$21, $22, $23, $24
		)
	`

	_, err := r.db.Exec(ctx, query,
		log.ID, log.UserID, log.Template, log.Recipient, log.Status, log.EventType,
		log.SendGridMessageID, log.SendGridEventID, log.BounceType, log.BounceReason,
		log.SpamReportReason, log.LinkURL, log.IPAddress, log.UserAgent, log.Metadata,
		log.SentAt, log.DeliveredAt, log.OpenedAt, log.ClickedAt, log.BouncedAt,
		log.SpamReportedAt, log.UnsubscribedAt, log.CreatedAt, log.UpdatedAt,
	)

	return err
}

// UpdateEmailLog updates an existing email log entry
func (r *EmailLogRepository) UpdateEmailLog(ctx context.Context, log *models.EmailLog) error {
	query := `
		UPDATE email_logs
		SET status = $1, delivered_at = $2, opened_at = $3, clicked_at = $4,
		    bounced_at = $5, spam_reported_at = $6, unsubscribed_at = $7,
		    bounce_type = $8, bounce_reason = $9, spam_report_reason = $10,
		    link_url = $11, metadata = $12, updated_at = $13
		WHERE id = $14
	`

	_, err := r.db.Exec(ctx, query,
		log.Status, log.DeliveredAt, log.OpenedAt, log.ClickedAt,
		log.BouncedAt, log.SpamReportedAt, log.UnsubscribedAt,
		log.BounceType, log.BounceReason, log.SpamReportReason,
		log.LinkURL, log.Metadata, log.UpdatedAt, log.ID,
	)

	return err
}

// GetEmailLogByMessageID retrieves an email log by SendGrid message ID
func (r *EmailLogRepository) GetEmailLogByMessageID(ctx context.Context, messageID string) (*models.EmailLog, error) {
	query := `
		SELECT id, user_id, template, recipient, status, event_type,
		       sendgrid_message_id, sendgrid_event_id, bounce_type, bounce_reason,
		       spam_report_reason, link_url, ip_address, user_agent, metadata,
		       sent_at, delivered_at, opened_at, clicked_at, bounced_at,
		       spam_reported_at, unsubscribed_at, created_at, updated_at
		FROM email_logs
		WHERE sendgrid_message_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var log models.EmailLog
	err := r.db.QueryRow(ctx, query, messageID).Scan(
		&log.ID, &log.UserID, &log.Template, &log.Recipient, &log.Status, &log.EventType,
		&log.SendGridMessageID, &log.SendGridEventID, &log.BounceType, &log.BounceReason,
		&log.SpamReportReason, &log.LinkURL, &log.IPAddress, &log.UserAgent, &log.Metadata,
		&log.SentAt, &log.DeliveredAt, &log.OpenedAt, &log.ClickedAt, &log.BouncedAt,
		&log.SpamReportedAt, &log.UnsubscribedAt, &log.CreatedAt, &log.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &log, nil
}

// GetEmailLogsByUserID retrieves email logs for a specific user
func (r *EmailLogRepository) GetEmailLogsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.EmailLog, error) {
	query := `
		SELECT id, user_id, template, recipient, status, event_type,
		       sendgrid_message_id, sendgrid_event_id, bounce_type, bounce_reason,
		       spam_report_reason, link_url, ip_address, user_agent, metadata,
		       sent_at, delivered_at, opened_at, clicked_at, bounced_at,
		       spam_reported_at, unsubscribed_at, created_at, updated_at
		FROM email_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanEmailLogs(rows)
}

// SearchEmailLogs searches email logs with filters
func (r *EmailLogRepository) SearchEmailLogs(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]models.EmailLog, error) {
	query := `
		SELECT id, user_id, template, recipient, status, event_type,
		       sendgrid_message_id, sendgrid_event_id, bounce_type, bounce_reason,
		       spam_report_reason, link_url, ip_address, user_agent, metadata,
		       sent_at, delivered_at, opened_at, clicked_at, bounced_at,
		       spam_reported_at, unsubscribed_at, created_at, updated_at
		FROM email_logs
		WHERE 1=1
	`

	args := make([]interface{}, 0)
	argCount := 1

	if status, ok := filters["status"].(string); ok && status != "" {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
		argCount++
	}

	if template, ok := filters["template"].(string); ok && template != "" {
		query += fmt.Sprintf(" AND template = $%d", argCount)
		args = append(args, template)
		argCount++
	}

	if recipient, ok := filters["recipient"].(string); ok && recipient != "" {
		query += fmt.Sprintf(" AND recipient ILIKE $%d", argCount)
		args = append(args, "%"+recipient+"%")
		argCount++
	}

	if startDate, ok := filters["start_date"].(time.Time); ok {
		query += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, startDate)
		argCount++
	}

	if endDate, ok := filters["end_date"].(time.Time); ok {
		query += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, endDate)
		argCount++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanEmailLogs(rows)
}

// GetRecentBounces retrieves recent bounced emails
func (r *EmailLogRepository) GetRecentBounces(ctx context.Context, limit int) ([]models.EmailLog, error) {
	query := `
		SELECT id, user_id, template, recipient, status, event_type,
		       sendgrid_message_id, sendgrid_event_id, bounce_type, bounce_reason,
		       spam_report_reason, link_url, ip_address, user_agent, metadata,
		       sent_at, delivered_at, opened_at, clicked_at, bounced_at,
		       spam_reported_at, unsubscribed_at, created_at, updated_at
		FROM email_logs
		WHERE status = $1 AND bounced_at IS NOT NULL
		ORDER BY bounced_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, models.EmailLogStatusBounce, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanEmailLogs(rows)
}

// GetMetricsForPeriod calculates email metrics for a given period
func (r *EmailLogRepository) GetMetricsForPeriod(ctx context.Context, startTime, endTime time.Time, template *string) (*models.EmailMetricsSummary, error) {
	query := `
		SELECT
			COUNT(*) FILTER (WHERE status = 'delivered' OR status = 'processed') as total_sent,
			COUNT(*) FILTER (WHERE status = 'delivered') as total_delivered,
			COUNT(*) FILTER (WHERE status = 'bounce') as total_bounced,
			COUNT(*) FILTER (WHERE status = 'bounce' AND bounce_type = 'hard') as total_hard_bounced,
			COUNT(*) FILTER (WHERE status = 'bounce' AND bounce_type = 'soft') as total_soft_bounced,
			COUNT(*) FILTER (WHERE status = 'dropped') as total_dropped,
			COUNT(*) FILTER (WHERE status = 'open' OR opened_at IS NOT NULL) as total_opened,
			COUNT(*) FILTER (WHERE status = 'click' OR clicked_at IS NOT NULL) as total_clicked,
			COUNT(*) FILTER (WHERE status = 'spam_report') as total_spam_reports,
			COUNT(*) FILTER (WHERE status = 'unsubscribe') as total_unsubscribes,
			COUNT(DISTINCT CASE WHEN opened_at IS NOT NULL THEN recipient END) as unique_opened,
			COUNT(DISTINCT CASE WHEN clicked_at IS NOT NULL THEN recipient END) as unique_clicked
		FROM email_logs
		WHERE created_at >= $1 AND created_at <= $2
	`

	args := []interface{}{startTime, endTime}
	if template != nil && *template != "" {
		query += " AND template = $3"
		args = append(args, *template)
	}

	var (
		totalSent         int
		totalDelivered    int
		totalBounced      int
		totalHardBounced  int
		totalSoftBounced  int
		totalDropped      int
		totalOpened       int
		totalClicked      int
		totalSpamReports  int
		totalUnsubscribes int
		uniqueOpened      int
		uniqueClicked     int
	)

	err := r.db.QueryRow(ctx, query, args...).Scan(
		&totalSent, &totalDelivered, &totalBounced, &totalHardBounced,
		&totalSoftBounced, &totalDropped, &totalOpened, &totalClicked,
		&totalSpamReports, &totalUnsubscribes, &uniqueOpened, &uniqueClicked,
	)
	if err != nil {
		return nil, err
	}

	// Calculate rates
	var bounceRate, openRate, clickRate, spamRate *float64
	if totalSent > 0 {
		br := float64(totalBounced) / float64(totalSent) * 100
		bounceRate = &br

		sr := float64(totalSpamReports) / float64(totalSent) * 100
		spamRate = &sr
	}
	if totalDelivered > 0 {
		or := float64(uniqueOpened) / float64(totalDelivered) * 100
		openRate = &or

		cr := float64(uniqueClicked) / float64(totalDelivered) * 100
		clickRate = &cr
	}

	return &models.EmailMetricsSummary{
		PeriodStart:       startTime,
		PeriodEnd:         endTime,
		Template:          template,
		TotalSent:         totalSent,
		TotalDelivered:    totalDelivered,
		TotalBounced:      totalBounced,
		TotalHardBounced:  totalHardBounced,
		TotalSoftBounced:  totalSoftBounced,
		TotalDropped:      totalDropped,
		TotalOpened:       totalOpened,
		TotalClicked:      totalClicked,
		TotalSpamReports:  totalSpamReports,
		TotalUnsubscribes: totalUnsubscribes,
		UniqueOpened:      uniqueOpened,
		UniqueClicked:     uniqueClicked,
		BounceRate:        bounceRate,
		OpenRate:          openRate,
		ClickRate:         clickRate,
		SpamRate:          spamRate,
	}, nil
}

// GetMetricsByTemplate retrieves metrics grouped by template
func (r *EmailLogRepository) GetMetricsByTemplate(ctx context.Context, startTime, endTime time.Time, limit int) ([]models.EmailMetricsSummary, error) {
	query := `
		SELECT
			template,
			COUNT(*) FILTER (WHERE status = 'delivered' OR status = 'processed') as total_sent,
			COUNT(*) FILTER (WHERE status = 'delivered') as total_delivered,
			COUNT(*) FILTER (WHERE status = 'bounce') as total_bounced,
			COUNT(*) FILTER (WHERE status = 'open' OR opened_at IS NOT NULL) as total_opened,
			COUNT(*) FILTER (WHERE status = 'click' OR clicked_at IS NOT NULL) as total_clicked,
			COUNT(DISTINCT CASE WHEN opened_at IS NOT NULL THEN recipient END) as unique_opened,
			COUNT(DISTINCT CASE WHEN clicked_at IS NOT NULL THEN recipient END) as unique_clicked
		FROM email_logs
		WHERE created_at >= $1 AND created_at <= $2 AND template IS NOT NULL
		GROUP BY template
		ORDER BY total_sent DESC
		LIMIT $3
	`

	rows, err := r.db.Query(ctx, query, startTime, endTime, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []models.EmailMetricsSummary
	for rows.Next() {
		var (
			template       string
			totalSent      int
			totalDelivered int
			totalBounced   int
			totalOpened    int
			totalClicked   int
			uniqueOpened   int
			uniqueClicked  int
		)

		err := rows.Scan(&template, &totalSent, &totalDelivered, &totalBounced,
			&totalOpened, &totalClicked, &uniqueOpened, &uniqueClicked)
		if err != nil {
			return nil, err
		}

		var openRate, clickRate *float64
		if totalDelivered > 0 {
			or := float64(uniqueOpened) / float64(totalDelivered) * 100
			openRate = &or

			cr := float64(uniqueClicked) / float64(totalDelivered) * 100
			clickRate = &cr
		}

		metrics = append(metrics, models.EmailMetricsSummary{
			PeriodStart:    startTime,
			PeriodEnd:      endTime,
			Template:       &template,
			TotalSent:      totalSent,
			TotalDelivered: totalDelivered,
			TotalBounced:   totalBounced,
			TotalOpened:    totalOpened,
			TotalClicked:   totalClicked,
			UniqueOpened:   uniqueOpened,
			UniqueClicked:  uniqueClicked,
			OpenRate:       openRate,
			ClickRate:      clickRate,
		})
	}

	return metrics, rows.Err()
}

// DeleteOldLogs removes email logs older than the specified duration
func (r *EmailLogRepository) DeleteOldLogs(ctx context.Context, olderThan time.Duration) (int64, error) {
	query := `
		DELETE FROM email_logs
		WHERE created_at < $1
	`

	cutoff := time.Now().Add(-olderThan)
	result, err := r.db.Exec(ctx, query, cutoff)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}

// Helper function to scan email logs from rows
func (r *EmailLogRepository) scanEmailLogs(rows pgx.Rows) ([]models.EmailLog, error) {
	var logs []models.EmailLog
	for rows.Next() {
		var log models.EmailLog
		err := rows.Scan(
			&log.ID, &log.UserID, &log.Template, &log.Recipient, &log.Status, &log.EventType,
			&log.SendGridMessageID, &log.SendGridEventID, &log.BounceType, &log.BounceReason,
			&log.SpamReportReason, &log.LinkURL, &log.IPAddress, &log.UserAgent, &log.Metadata,
			&log.SentAt, &log.DeliveredAt, &log.OpenedAt, &log.ClickedAt, &log.BouncedAt,
			&log.SpamReportedAt, &log.UnsubscribedAt, &log.CreatedAt, &log.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// CreateMetricsSummary creates a new email metrics summary entry
func (r *EmailLogRepository) CreateMetricsSummary(ctx context.Context, summary *models.EmailMetricsSummary) error {
	query := `
		INSERT INTO email_metrics_summary (
			id, period_start, period_end, granularity, template,
			total_sent, total_delivered, total_bounced, total_hard_bounced, total_soft_bounced,
			total_dropped, total_opened, total_clicked, total_spam_reports, total_unsubscribes,
			unique_opened, unique_clicked, bounce_rate, open_rate, click_rate, spam_rate,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
			$21, $22, $23
		)
		ON CONFLICT (period_start, granularity, COALESCE(template, ''))
		DO UPDATE SET
			total_sent = EXCLUDED.total_sent,
			total_delivered = EXCLUDED.total_delivered,
			total_bounced = EXCLUDED.total_bounced,
			total_hard_bounced = EXCLUDED.total_hard_bounced,
			total_soft_bounced = EXCLUDED.total_soft_bounced,
			total_dropped = EXCLUDED.total_dropped,
			total_opened = EXCLUDED.total_opened,
			total_clicked = EXCLUDED.total_clicked,
			total_spam_reports = EXCLUDED.total_spam_reports,
			total_unsubscribes = EXCLUDED.total_unsubscribes,
			unique_opened = EXCLUDED.unique_opened,
			unique_clicked = EXCLUDED.unique_clicked,
			bounce_rate = EXCLUDED.bounce_rate,
			open_rate = EXCLUDED.open_rate,
			click_rate = EXCLUDED.click_rate,
			spam_rate = EXCLUDED.spam_rate,
			updated_at = EXCLUDED.updated_at
	`

	summary.ID = uuid.New()
	summary.CreatedAt = time.Now()
	summary.UpdatedAt = time.Now()

	_, err := r.db.Exec(ctx, query,
		summary.ID, summary.PeriodStart, summary.PeriodEnd, summary.Granularity, summary.Template,
		summary.TotalSent, summary.TotalDelivered, summary.TotalBounced, summary.TotalHardBounced, summary.TotalSoftBounced,
		summary.TotalDropped, summary.TotalOpened, summary.TotalClicked, summary.TotalSpamReports, summary.TotalUnsubscribes,
		summary.UniqueOpened, summary.UniqueClicked, summary.BounceRate, summary.OpenRate, summary.ClickRate, summary.SpamRate,
		summary.CreatedAt, summary.UpdatedAt,
	)

	return err
}

// CreateAlert creates a new email alert
func (r *EmailLogRepository) CreateAlert(ctx context.Context, alert *models.EmailAlert) error {
	query := `
		INSERT INTO email_alerts (
			id, alert_type, severity, metric_name, current_value, threshold_value,
			period_start, period_end, message, metadata, triggered_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	alert.ID = uuid.New()
	alert.TriggeredAt = time.Now()
	alert.CreatedAt = time.Now()

	_, err := r.db.Exec(ctx, query,
		alert.ID, alert.AlertType, alert.Severity, alert.MetricName, alert.CurrentValue, alert.ThresholdValue,
		alert.PeriodStart, alert.PeriodEnd, alert.Message, alert.Metadata, alert.TriggeredAt, alert.CreatedAt,
	)

	return err
}

// GetUnresolvedAlerts retrieves unresolved email alerts
func (r *EmailLogRepository) GetUnresolvedAlerts(ctx context.Context, limit int) ([]models.EmailAlert, error) {
	query := `
		SELECT id, alert_type, severity, metric_name, current_value, threshold_value,
		       period_start, period_end, message, metadata, triggered_at,
		       acknowledged_at, acknowledged_by, resolved_at, created_at
		FROM email_alerts
		WHERE resolved_at IS NULL
		ORDER BY triggered_at DESC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []models.EmailAlert
	for rows.Next() {
		var alert models.EmailAlert
		err := rows.Scan(
			&alert.ID, &alert.AlertType, &alert.Severity, &alert.MetricName, &alert.CurrentValue, &alert.ThresholdValue,
			&alert.PeriodStart, &alert.PeriodEnd, &alert.Message, &alert.Metadata, &alert.TriggeredAt,
			&alert.AcknowledgedAt, &alert.AcknowledgedBy, &alert.ResolvedAt, &alert.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, alert)
	}

	return alerts, rows.Err()
}

// AcknowledgeAlert acknowledges an email alert
func (r *EmailLogRepository) AcknowledgeAlert(ctx context.Context, alertID, userID uuid.UUID) error {
	query := `
		UPDATE email_alerts
		SET acknowledged_at = $1, acknowledged_by = $2
		WHERE id = $3 AND acknowledged_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, time.Now(), userID, alertID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("alert not found or already acknowledged")
	}

	return nil
}

// ResolveAlert resolves an email alert
func (r *EmailLogRepository) ResolveAlert(ctx context.Context, alertID uuid.UUID) error {
	query := `
		UPDATE email_alerts
		SET resolved_at = $1
		WHERE id = $2 AND resolved_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, time.Now(), alertID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("alert not found or already resolved")
	}

	return nil
}
