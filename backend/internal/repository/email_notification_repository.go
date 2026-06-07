package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// EmailNotificationRepository handles email notification data operations
type EmailNotificationRepository struct {
	db *pgxpool.Pool
}

// NewEmailNotificationRepository creates a new EmailNotificationRepository
func NewEmailNotificationRepository(db *pgxpool.Pool) *EmailNotificationRepository {
	return &EmailNotificationRepository{db: db}
}

// CreateLog creates a new email notification log entry
func (r *EmailNotificationRepository) CreateLog(ctx context.Context, log *models.EmailNotificationLog) error {
	query := `
		INSERT INTO email_notification_logs (
			id, user_id, notification_id, notification_type, recipient_email,
			subject, status, provider_message_id, error_message, sent_at,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.Exec(ctx, query,
		log.ID, log.UserID, log.NotificationID, log.NotificationType, log.RecipientEmail,
		log.Subject, log.Status, log.ProviderMessageID, log.ErrorMessage, log.SentAt,
		log.CreatedAt, log.UpdatedAt,
	)

	return err
}

// UpdateLog updates an email notification log entry
func (r *EmailNotificationRepository) UpdateLog(ctx context.Context, log *models.EmailNotificationLog) error {
	query := `
		UPDATE email_notification_logs
		SET status = $1, provider_message_id = $2, error_message = $3,
		    sent_at = $4, updated_at = $5
		WHERE id = $6
	`

	_, err := r.db.Exec(ctx, query,
		log.Status, log.ProviderMessageID, log.ErrorMessage,
		log.SentAt, log.UpdatedAt, log.ID,
	)

	return err
}

// GetLogsByUserID retrieves email logs for a user
func (r *EmailNotificationRepository) GetLogsByUserID(
	ctx context.Context,
	userID uuid.UUID,
	limit, offset int,
) ([]models.EmailNotificationLog, error) {
	query := `
		SELECT id, user_id, notification_id, notification_type, recipient_email,
		       subject, status, provider_message_id, error_message, sent_at,
		       created_at, updated_at
		FROM email_notification_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.EmailNotificationLog
	for rows.Next() {
		var log models.EmailNotificationLog
		err := rows.Scan(
			&log.ID, &log.UserID, &log.NotificationID, &log.NotificationType, &log.RecipientEmail,
			&log.Subject, &log.Status, &log.ProviderMessageID, &log.ErrorMessage, &log.SentAt,
			&log.CreatedAt, &log.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// CreateUnsubscribeToken creates a new unsubscribe token
func (r *EmailNotificationRepository) CreateUnsubscribeToken(ctx context.Context, token *models.EmailUnsubscribeToken) error {
	query := `
		INSERT INTO email_unsubscribe_tokens (
			id, user_id, token, notification_type, created_at, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(ctx, query,
		token.ID, token.UserID, token.Token, token.NotificationType,
		token.CreatedAt, token.ExpiresAt,
	)

	return err
}

// GetUnsubscribeToken retrieves an unsubscribe token
func (r *EmailNotificationRepository) GetUnsubscribeToken(ctx context.Context, token string) (*models.EmailUnsubscribeToken, error) {
	query := `
		SELECT id, user_id, token, notification_type, created_at, expires_at, used_at
		FROM email_unsubscribe_tokens
		WHERE token = $1
	`

	var t models.EmailUnsubscribeToken
	err := r.db.QueryRow(ctx, query, token).Scan(
		&t.ID, &t.UserID, &t.Token, &t.NotificationType,
		&t.CreatedAt, &t.ExpiresAt, &t.UsedAt,
	)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

// MarkTokenUsed marks an unsubscribe token as used
func (r *EmailNotificationRepository) MarkTokenUsed(ctx context.Context, token string) error {
	query := `
		UPDATE email_unsubscribe_tokens
		SET used_at = $1
		WHERE token = $2 AND used_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, time.Now(), token)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("token not found or already used")
	}

	return nil
}

// GetRateLimit retrieves the rate limit for a user in a specific window
func (r *EmailNotificationRepository) GetRateLimit(
	ctx context.Context,
	userID uuid.UUID,
	windowStart time.Time,
) (*models.EmailRateLimit, error) {
	query := `
		SELECT id, user_id, window_start, email_count, created_at, updated_at
		FROM email_rate_limits
		WHERE user_id = $1 AND window_start = $2
	`

	var rateLimit models.EmailRateLimit
	err := r.db.QueryRow(ctx, query, userID, windowStart).Scan(
		&rateLimit.ID, &rateLimit.UserID, &rateLimit.WindowStart,
		&rateLimit.EmailCount, &rateLimit.CreatedAt, &rateLimit.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &rateLimit, nil
}

// IncrementRateLimit increments the email count for rate limiting
func (r *EmailNotificationRepository) IncrementRateLimit(
	ctx context.Context,
	userID uuid.UUID,
	windowStart time.Time,
) error {
	query := `
		INSERT INTO email_rate_limits (id, user_id, window_start, email_count, created_at, updated_at)
		VALUES ($1, $2, $3, 1, $4, $5)
		ON CONFLICT (user_id, window_start)
		DO UPDATE SET
			email_count = email_rate_limits.email_count + 1,
			updated_at = $6
	`

	now := time.Now()
	_, err := r.db.Exec(ctx, query,
		uuid.New(), userID, windowStart, now, now, now,
	)

	return err
}

// CleanupExpiredTokens removes expired unsubscribe tokens (for maintenance)
func (r *EmailNotificationRepository) CleanupExpiredTokens(ctx context.Context) error {
	query := `
		DELETE FROM email_unsubscribe_tokens
		WHERE expires_at < NOW() AND used_at IS NULL
	`

	_, err := r.db.Exec(ctx, query)
	return err
}

// CleanupOldRateLimits removes old rate limit records (for maintenance)
func (r *EmailNotificationRepository) CleanupOldRateLimits(ctx context.Context, olderThan time.Duration) error {
	query := `
		DELETE FROM email_rate_limits
		WHERE window_start < $1
	`

	cutoff := time.Now().Add(-olderThan)
	_, err := r.db.Exec(ctx, query, cutoff)
	return err
}
