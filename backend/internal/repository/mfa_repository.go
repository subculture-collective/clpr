package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// MFARepository handles database operations for MFA
type MFARepository struct {
	db *pgxpool.Pool
}

// NewMFARepository creates a new MFA repository
func NewMFARepository(db *pgxpool.Pool) *MFARepository {
	return &MFARepository{db: db}
}

// GetMFAByUserID retrieves MFA configuration for a user
func (r *MFARepository) GetMFAByUserID(ctx context.Context, userID uuid.UUID) (*models.UserMFA, error) {
	query := `
		SELECT id, user_id, secret, enabled, enrolled_at, backup_codes, 
		       backup_codes_generated_at, mfa_required, mfa_required_at, 
		       grace_period_end, created_at, updated_at
		FROM user_mfa
		WHERE user_id = $1
	`

	var mfa models.UserMFA

	err := r.db.QueryRow(ctx, query, userID).Scan(
		&mfa.ID,
		&mfa.UserID,
		&mfa.Secret,
		&mfa.Enabled,
		&mfa.EnrolledAt,
		&mfa.BackupCodes,
		&mfa.BackupCodesGeneratedAt,
		&mfa.MFARequired,
		&mfa.MFARequiredAt,
		&mfa.GracePeriodEnd,
		&mfa.CreatedAt,
		&mfa.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get MFA config: %w", err)
	}

	return &mfa, nil
}

// CreateMFA creates a new MFA configuration for a user
func (r *MFARepository) CreateMFA(ctx context.Context, mfa *models.UserMFA) error {
	query := `
		INSERT INTO user_mfa (user_id, secret, enabled, enrolled_at, backup_codes, 
		                      backup_codes_generated_at, mfa_required, mfa_required_at, 
		                      grace_period_end)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		mfa.UserID,
		mfa.Secret,
		mfa.Enabled,
		mfa.EnrolledAt,
		mfa.BackupCodes,
		mfa.BackupCodesGeneratedAt,
		mfa.MFARequired,
		mfa.MFARequiredAt,
		mfa.GracePeriodEnd,
	).Scan(&mfa.ID, &mfa.CreatedAt, &mfa.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create MFA config: %w", err)
	}

	return nil
}

// UpdateMFA updates an existing MFA configuration
func (r *MFARepository) UpdateMFA(ctx context.Context, mfa *models.UserMFA) error {
	query := `
		UPDATE user_mfa
		SET secret = $1, enabled = $2, enrolled_at = $3, backup_codes = $4, 
		    backup_codes_generated_at = $5, mfa_required = $6, mfa_required_at = $7,
		    grace_period_end = $8, updated_at = NOW()
		WHERE user_id = $9
		RETURNING updated_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		mfa.Secret,
		mfa.Enabled,
		mfa.EnrolledAt,
		mfa.BackupCodes,
		mfa.BackupCodesGeneratedAt,
		mfa.MFARequired,
		mfa.MFARequiredAt,
		mfa.GracePeriodEnd,
		mfa.UserID,
	).Scan(&mfa.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update MFA config: %w", err)
	}

	return nil
}

// DeleteMFA deletes MFA configuration for a user
func (r *MFARepository) DeleteMFA(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM user_mfa WHERE user_id = $1`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete MFA config: %w", err)
	}

	return nil
}

// ConsumeBackupCode removes a backup code after it's been used
func (r *MFARepository) ConsumeBackupCode(ctx context.Context, userID uuid.UUID, codeHash string) error {
	query := `
		UPDATE user_mfa
		SET backup_codes = array_remove(backup_codes, $1),
		    updated_at = NOW()
		WHERE user_id = $2
	`

	_, err := r.db.Exec(ctx, query, codeHash, userID)
	if err != nil {
		return fmt.Errorf("failed to consume backup code: %w", err)
	}

	return nil
}

// GetTrustedDevices retrieves all trusted devices for a user
func (r *MFARepository) GetTrustedDevices(ctx context.Context, userID uuid.UUID) ([]*models.MFATrustedDevice, error) {
	query := `
		SELECT id, user_id, device_fingerprint, device_name, ip_address, user_agent,
		       trusted_at, expires_at, last_used_at, created_at
		FROM mfa_trusted_devices
		WHERE user_id = $1 AND expires_at > NOW()
		ORDER BY last_used_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trusted devices: %w", err)
	}
	defer rows.Close()

	var devices []*models.MFATrustedDevice
	for rows.Next() {
		var device models.MFATrustedDevice
		err := rows.Scan(
			&device.ID,
			&device.UserID,
			&device.DeviceFingerprint,
			&device.DeviceName,
			&device.IPAddress,
			&device.UserAgent,
			&device.TrustedAt,
			&device.ExpiresAt,
			&device.LastUsedAt,
			&device.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trusted device: %w", err)
		}
		devices = append(devices, &device)
	}

	return devices, nil
}

// GetTrustedDeviceByFingerprint retrieves a trusted device by fingerprint
func (r *MFARepository) GetTrustedDeviceByFingerprint(ctx context.Context, userID uuid.UUID, fingerprint string) (*models.MFATrustedDevice, error) {
	query := `
		SELECT id, user_id, device_fingerprint, device_name, ip_address, user_agent,
		       trusted_at, expires_at, last_used_at, created_at
		FROM mfa_trusted_devices
		WHERE user_id = $1 AND device_fingerprint = $2 AND expires_at > NOW()
	`

	var device models.MFATrustedDevice
	err := r.db.QueryRow(ctx, query, userID, fingerprint).Scan(
		&device.ID,
		&device.UserID,
		&device.DeviceFingerprint,
		&device.DeviceName,
		&device.IPAddress,
		&device.UserAgent,
		&device.TrustedAt,
		&device.ExpiresAt,
		&device.LastUsedAt,
		&device.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get trusted device: %w", err)
	}

	return &device, nil
}

// CreateTrustedDevice creates a new trusted device entry
func (r *MFARepository) CreateTrustedDevice(ctx context.Context, device *models.MFATrustedDevice) error {
	query := `
		INSERT INTO mfa_trusted_devices 
		(user_id, device_fingerprint, device_name, ip_address, user_agent, trusted_at, expires_at, last_used_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id, device_fingerprint) 
		DO UPDATE SET 
			last_used_at = EXCLUDED.last_used_at,
			expires_at = EXCLUDED.expires_at
		RETURNING id, created_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		device.UserID,
		device.DeviceFingerprint,
		device.DeviceName,
		device.IPAddress,
		device.UserAgent,
		device.TrustedAt,
		device.ExpiresAt,
		device.LastUsedAt,
	).Scan(&device.ID, &device.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create trusted device: %w", err)
	}

	return nil
}

// UpdateTrustedDeviceLastUsed updates the last used timestamp for a trusted device
func (r *MFARepository) UpdateTrustedDeviceLastUsed(ctx context.Context, deviceID int) error {
	query := `
		UPDATE mfa_trusted_devices
		SET last_used_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, deviceID)
	if err != nil {
		return fmt.Errorf("failed to update trusted device: %w", err)
	}

	return nil
}

// DeleteTrustedDevice deletes a trusted device by ID
func (r *MFARepository) DeleteTrustedDevice(ctx context.Context, userID uuid.UUID, deviceID int) error {
	query := `DELETE FROM mfa_trusted_devices WHERE id = $1 AND user_id = $2`

	result, err := r.db.Exec(ctx, query, deviceID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete trusted device: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("trusted device not found")
	}

	return nil
}

// DeleteAllTrustedDevices deletes all trusted devices for a user
func (r *MFARepository) DeleteAllTrustedDevices(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM mfa_trusted_devices WHERE user_id = $1`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete all trusted devices: %w", err)
	}

	return nil
}

// CreateAuditLog creates a new MFA audit log entry
func (r *MFARepository) CreateAuditLog(ctx context.Context, log *models.MFAAuditLog) error {
	query := `
		INSERT INTO mfa_audit_logs (user_id, action, success, ip_address, user_agent, details)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		log.UserID,
		log.Action,
		log.Success,
		log.IPAddress,
		log.UserAgent,
		log.Details,
	).Scan(&log.ID, &log.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create MFA audit log: %w", err)
	}

	return nil
}

// GetAuditLogs retrieves MFA audit logs for a user
func (r *MFARepository) GetAuditLogs(ctx context.Context, userID uuid.UUID, limit int) ([]*models.MFAAuditLog, error) {
	query := `
		SELECT id, user_id, action, success, ip_address, user_agent, details, created_at
		FROM mfa_audit_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*models.MFAAuditLog
	for rows.Next() {
		var log models.MFAAuditLog
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.Action,
			&log.Success,
			&log.IPAddress,
			&log.UserAgent,
			&log.Details,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}
		logs = append(logs, &log)
	}

	return logs, nil
}

// GetFailedLoginAttempts counts failed MFA login attempts in a time window
func (r *MFARepository) GetFailedLoginAttempts(ctx context.Context, userID uuid.UUID, since time.Time) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM mfa_audit_logs
		WHERE user_id = $1 
		  AND action IN ($2, $3)
		  AND success = false
		  AND created_at > $4
	`

	var count int
	err := r.db.QueryRow(
		ctx,
		query,
		userID,
		models.MFAActionLoginFailed,
		models.MFAActionBackupCodeFailed,
		since,
	).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to count failed login attempts: %w", err)
	}

	return count, nil
}

// CleanupExpiredTrustedDevices removes expired trusted devices
func (r *MFARepository) CleanupExpiredTrustedDevices(ctx context.Context) error {
	query := `DELETE FROM mfa_trusted_devices WHERE expires_at < NOW()`

	_, err := r.db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired devices: %w", err)
	}

	return nil
}

// SetMFARequired marks MFA as required for a user and sets grace period
func (r *MFARepository) SetMFARequired(ctx context.Context, userID uuid.UUID, gracePeriodDays int) error {
	now := time.Now()
	gracePeriodEnd := now.AddDate(0, 0, gracePeriodDays)

	query := `
		INSERT INTO user_mfa (user_id, secret, enabled, mfa_required, mfa_required_at, grace_period_end, backup_codes)
		VALUES ($1, '', false, true, $2, $3, ARRAY[]::TEXT[])
		ON CONFLICT (user_id) 
		DO UPDATE SET 
			mfa_required = true,
			mfa_required_at = CASE 
				WHEN user_mfa.mfa_required = false THEN $2 
				ELSE user_mfa.mfa_required_at 
			END,
			grace_period_end = CASE 
				WHEN user_mfa.mfa_required = false OR user_mfa.grace_period_end IS NULL OR user_mfa.grace_period_end < NOW() THEN $3
				ELSE user_mfa.grace_period_end 
			END,
			updated_at = NOW()
	`

	_, err := r.db.Exec(ctx, query, userID, now, gracePeriodEnd)
	if err != nil {
		return fmt.Errorf("failed to set MFA required: %w", err)
	}

	return nil
}

// GetUsersWithExpiredGracePeriod retrieves users whose MFA grace period has expired
func (r *MFARepository) GetUsersWithExpiredGracePeriod(ctx context.Context) ([]uuid.UUID, error) {
	query := `
		SELECT user_id
		FROM user_mfa
		WHERE mfa_required = true 
		  AND enabled = false 
		  AND grace_period_end < NOW()
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get users with expired grace period: %w", err)
	}
	defer rows.Close()

	var userIDs []uuid.UUID
	for rows.Next() {
		var userID uuid.UUID
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("failed to scan user ID: %w", err)
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}
