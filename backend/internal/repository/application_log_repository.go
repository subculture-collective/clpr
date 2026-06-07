package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// ApplicationLogRepository handles database operations for application logs
type ApplicationLogRepository struct {
	db *pgxpool.Pool
}

// NewApplicationLogRepository creates a new ApplicationLogRepository
func NewApplicationLogRepository(db *pgxpool.Pool) *ApplicationLogRepository {
	return &ApplicationLogRepository{db: db}
}

// Create creates a new application log entry
func (r *ApplicationLogRepository) Create(ctx context.Context, log *models.ApplicationLog) error {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}

	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}

	// Convert context to JSONB if present
	var contextJSON []byte
	if len(log.Context) > 0 {
		contextJSON = log.Context
	}

	query := `
		INSERT INTO application_logs (
			id, level, message, timestamp, service, platform,
			user_id, session_id, trace_id, url, user_agent,
			device_id, app_version, error, stack, context,
			ip_address, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18
		)`

	// Ensure timestamps are in UTC to avoid timezone offset issues
	timestampUTC := log.Timestamp.UTC()
	createdAtUTC := log.CreatedAt.UTC()

	_, err := r.db.Exec(ctx, query,
		log.ID,
		log.Level,
		log.Message,
		timestampUTC,
		log.Service,
		log.Platform,
		log.UserID,
		log.SessionID,
		log.TraceID,
		log.URL,
		log.UserAgent,
		log.DeviceID,
		log.AppVersion,
		log.Error,
		log.Stack,
		contextJSON,
		log.IPAddress,
		createdAtUTC,
	)

	return err
}

// DeleteOldLogs deletes logs older than the specified retention period
// This should be called periodically by a cleanup job
func (r *ApplicationLogRepository) DeleteOldLogs(ctx context.Context, retentionDays int) (int64, error) {
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	query := `
		DELETE FROM application_logs
		WHERE created_at < $1`

	result, err := r.db.Exec(ctx, query, cutoffDate)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old logs: %w", err)
	}

	return result.RowsAffected(), nil
}

// GetLogStats returns statistics about stored logs
func (r *ApplicationLogRepository) GetLogStats(ctx context.Context) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_logs,
			COUNT(DISTINCT user_id) FILTER (WHERE user_id IS NOT NULL) as unique_users,
			COUNT(*) FILTER (WHERE level = 'error') as error_count,
			COUNT(*) FILTER (WHERE level = 'warn') as warn_count,
			COUNT(*) FILTER (WHERE level = 'info') as info_count,
			COUNT(*) FILTER (WHERE level = 'debug') as debug_count,
			COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '1 hour') as logs_last_hour,
			COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '24 hours') as logs_last_24h
		FROM application_logs`

	var stats struct {
		TotalLogs    int `db:"total_logs"`
		UniqueUsers  int `db:"unique_users"`
		ErrorCount   int `db:"error_count"`
		WarnCount    int `db:"warn_count"`
		InfoCount    int `db:"info_count"`
		DebugCount   int `db:"debug_count"`
		LogsLastHour int `db:"logs_last_hour"`
		LogsLast24h  int `db:"logs_last_24h"`
	}

	err := r.db.QueryRow(ctx, query).Scan(
		&stats.TotalLogs,
		&stats.UniqueUsers,
		&stats.ErrorCount,
		&stats.WarnCount,
		&stats.InfoCount,
		&stats.DebugCount,
		&stats.LogsLastHour,
		&stats.LogsLast24h,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get log stats: %w", err)
	}

	// Convert to map for JSON response
	statsMap := map[string]interface{}{
		"total_logs":     stats.TotalLogs,
		"unique_users":   stats.UniqueUsers,
		"error_count":    stats.ErrorCount,
		"warn_count":     stats.WarnCount,
		"info_count":     stats.InfoCount,
		"debug_count":    stats.DebugCount,
		"logs_last_hour": stats.LogsLastHour,
		"logs_last_24h":  stats.LogsLast24h,
	}

	return statsMap, nil
}
