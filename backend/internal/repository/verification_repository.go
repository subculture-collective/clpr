package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// VerificationRepository handles database operations for creator verification applications
type VerificationRepository struct {
	db *pgxpool.Pool
}

// NewVerificationRepository creates a new verification repository
func NewVerificationRepository(db *pgxpool.Pool) *VerificationRepository {
	return &VerificationRepository{db: db}
}

// ==============================================================================
// Verification Application Operations
// ==============================================================================

// CreateApplication creates a new verification application
func (r *VerificationRepository) CreateApplication(ctx context.Context, app *models.CreatorVerificationApplication) error {
	query := `
		INSERT INTO creator_verification_applications (
			user_id, twitch_channel_url, follower_count, subscriber_count,
			avg_viewers, content_description, social_media_links,
			status, priority
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(ctx, query,
		app.UserID,
		app.TwitchChannelURL,
		app.FollowerCount,
		app.SubscriberCount,
		app.AvgViewers,
		app.ContentDescription,
		app.SocialMediaLinks,
		app.Status,
		app.Priority,
	).Scan(&app.ID, &app.CreatedAt, &app.UpdatedAt)

	return err
}

// GetApplicationByID retrieves a verification application by ID
func (r *VerificationRepository) GetApplicationByID(ctx context.Context, id uuid.UUID) (*models.CreatorVerificationApplication, error) {
	query := `
		SELECT id, user_id, twitch_channel_url, follower_count, subscriber_count,
			avg_viewers, content_description, social_media_links,
			status, priority, reviewed_by, reviewed_at, reviewer_notes,
			created_at, updated_at
		FROM creator_verification_applications
		WHERE id = $1`

	app := &models.CreatorVerificationApplication{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&app.ID,
		&app.UserID,
		&app.TwitchChannelURL,
		&app.FollowerCount,
		&app.SubscriberCount,
		&app.AvgViewers,
		&app.ContentDescription,
		&app.SocialMediaLinks,
		&app.Status,
		&app.Priority,
		&app.ReviewedBy,
		&app.ReviewedAt,
		&app.ReviewerNotes,
		&app.CreatedAt,
		&app.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("verification application not found")
	}
	return app, err
}

// GetApplicationByUserID retrieves a user's verification application by status
// If status is empty, retrieves the most recent application regardless of status
func (r *VerificationRepository) GetApplicationByUserID(ctx context.Context, userID uuid.UUID, status string) (*models.CreatorVerificationApplication, error) {
	var query string
	var args []interface{}

	if status == "" {
		// Get latest application regardless of status
		query = `
			SELECT id, user_id, twitch_channel_url, follower_count, subscriber_count,
				avg_viewers, content_description, social_media_links,
				status, priority, reviewed_by, reviewed_at, reviewer_notes,
				created_at, updated_at
			FROM creator_verification_applications
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT 1`
		args = []interface{}{userID}
	} else {
		// Get latest application with specific status
		query = `
			SELECT id, user_id, twitch_channel_url, follower_count, subscriber_count,
				avg_viewers, content_description, social_media_links,
				status, priority, reviewed_by, reviewed_at, reviewer_notes,
				created_at, updated_at
			FROM creator_verification_applications
			WHERE user_id = $1 AND status = $2
			ORDER BY created_at DESC
			LIMIT 1`
		args = []interface{}{userID, status}
	}

	app := &models.CreatorVerificationApplication{}
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&app.ID,
		&app.UserID,
		&app.TwitchChannelURL,
		&app.FollowerCount,
		&app.SubscriberCount,
		&app.AvgViewers,
		&app.ContentDescription,
		&app.SocialMediaLinks,
		&app.Status,
		&app.Priority,
		&app.ReviewedBy,
		&app.ReviewedAt,
		&app.ReviewerNotes,
		&app.CreatedAt,
		&app.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil // No application is not an error
	}
	return app, err
}

// ListApplications lists verification applications with pagination and filtering
func (r *VerificationRepository) ListApplications(ctx context.Context, status string, limit, offset int) ([]*models.CreatorVerificationApplication, error) {
	query := `
		SELECT id, user_id, twitch_channel_url, follower_count, subscriber_count,
			avg_viewers, content_description, social_media_links,
			status, priority, reviewed_by, reviewed_at, reviewer_notes,
			created_at, updated_at
		FROM creator_verification_applications
		WHERE ($1 = '' OR status = $1)
		ORDER BY priority DESC, created_at ASC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, status, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []*models.CreatorVerificationApplication
	for rows.Next() {
		app := &models.CreatorVerificationApplication{}
		err := rows.Scan(
			&app.ID,
			&app.UserID,
			&app.TwitchChannelURL,
			&app.FollowerCount,
			&app.SubscriberCount,
			&app.AvgViewers,
			&app.ContentDescription,
			&app.SocialMediaLinks,
			&app.Status,
			&app.Priority,
			&app.ReviewedBy,
			&app.ReviewedAt,
			&app.ReviewerNotes,
			&app.CreatedAt,
			&app.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}

	return apps, rows.Err()
}

// UpdateApplicationStatus updates an application's review status
func (r *VerificationRepository) UpdateApplicationStatus(ctx context.Context, id, reviewerID uuid.UUID, status, notes string) error {
	query := `
		UPDATE creator_verification_applications
		SET status = $1,
			reviewed_by = $2,
			reviewer_notes = $3,
			updated_at = NOW()
		WHERE id = $4`

	result, err := r.db.Exec(ctx, query, status, reviewerID, notes, id)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("application not found")
	}

	return nil
}

// GetApplicationStats retrieves statistics about verification applications
func (r *VerificationRepository) GetApplicationStats(ctx context.Context) (*models.VerificationApplicationStats, error) {
	query := `
		SELECT 
			COUNT(*) FILTER (WHERE status = 'pending') as total_pending,
			COUNT(*) FILTER (WHERE status = 'approved') as total_approved,
			COUNT(*) FILTER (WHERE status = 'rejected') as total_rejected,
			(SELECT COUNT(*) FROM users WHERE is_verified = true) as total_verified
		FROM creator_verification_applications`

	stats := &models.VerificationApplicationStats{}
	err := r.db.QueryRow(ctx, query).Scan(
		&stats.TotalPending,
		&stats.TotalApproved,
		&stats.TotalRejected,
		&stats.TotalVerified,
	)

	return stats, err
}

// CreateDecision records a verification decision for audit trail
func (r *VerificationRepository) CreateDecision(ctx context.Context, decision *models.CreatorVerificationDecision) error {
	query := `
		INSERT INTO creator_verification_decisions (
			application_id, reviewer_id, decision, notes, metadata
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`

	err := r.db.QueryRow(ctx, query,
		decision.ApplicationID,
		decision.ReviewerID,
		decision.Decision,
		decision.Notes,
		decision.Metadata,
	).Scan(&decision.ID, &decision.CreatedAt)

	return err
}

// GetDecisionsByApplicationID retrieves all decisions for an application
func (r *VerificationRepository) GetDecisionsByApplicationID(ctx context.Context, applicationID uuid.UUID) ([]*models.CreatorVerificationDecision, error) {
	query := `
		SELECT id, application_id, reviewer_id, decision, notes, metadata, created_at
		FROM creator_verification_decisions
		WHERE application_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, applicationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var decisions []*models.CreatorVerificationDecision
	for rows.Next() {
		decision := &models.CreatorVerificationDecision{}
		err := rows.Scan(
			&decision.ID,
			&decision.ApplicationID,
			&decision.ReviewerID,
			&decision.Decision,
			&decision.Notes,
			&decision.Metadata,
			&decision.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		decisions = append(decisions, decision)
	}

	return decisions, rows.Err()
}

// GetApplicationWithUser retrieves an application with user information
func (r *VerificationRepository) GetApplicationWithUser(ctx context.Context, id uuid.UUID) (*models.CreatorVerificationApplicationWithUser, error) {
	query := `
		SELECT 
			a.id, a.user_id, a.twitch_channel_url, a.follower_count, a.subscriber_count,
			a.avg_viewers, a.content_description, a.social_media_links,
			a.status, a.priority, a.reviewed_by, a.reviewed_at, a.reviewer_notes,
			a.created_at, a.updated_at,
			u.id, u.twitch_id, u.username, u.display_name, u.email, u.avatar_url,
			u.bio, u.karma_points, u.trust_score, u.role, u.account_type,
			u.is_verified, u.created_at
		FROM creator_verification_applications a
		INNER JOIN users u ON a.user_id = u.id
		WHERE a.id = $1`

	appWithUser := &models.CreatorVerificationApplicationWithUser{}
	app := &appWithUser.CreatorVerificationApplication
	user := &models.User{}

	err := r.db.QueryRow(ctx, query, id).Scan(
		&app.ID,
		&app.UserID,
		&app.TwitchChannelURL,
		&app.FollowerCount,
		&app.SubscriberCount,
		&app.AvgViewers,
		&app.ContentDescription,
		&app.SocialMediaLinks,
		&app.Status,
		&app.Priority,
		&app.ReviewedBy,
		&app.ReviewedAt,
		&app.ReviewerNotes,
		&app.CreatedAt,
		&app.UpdatedAt,
		&user.ID,
		&user.TwitchID,
		&user.Username,
		&user.DisplayName,
		&user.Email,
		&user.AvatarURL,
		&user.Bio,
		&user.KarmaPoints,
		&user.TrustScore,
		&user.Role,
		&user.AccountType,
		&user.IsVerified,
		&user.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("verification application not found")
	}
	if err != nil {
		return nil, err
	}

	appWithUser.User = user
	return appWithUser, nil
}

// ==============================================================================
// Abuse Prevention Methods
// ==============================================================================

// GetRecentRejectedApplicationByUserID checks if user has a recently rejected application
// Returns the most recent rejected application within the specified days, or nil if none found
func (r *VerificationRepository) GetRecentRejectedApplicationByUserID(ctx context.Context, userID uuid.UUID, withinDays int) (*models.CreatorVerificationApplication, error) {
	query := `
		SELECT id, user_id, twitch_channel_url, follower_count, subscriber_count,
			avg_viewers, content_description, social_media_links,
			status, priority, reviewed_by, reviewed_at, reviewer_notes,
			created_at, updated_at
		FROM creator_verification_applications
		WHERE user_id = $1 
			AND status = $2
			AND reviewed_at > NOW() - INTERVAL '1 day' * $3
		ORDER BY reviewed_at DESC
		LIMIT 1`

	app := &models.CreatorVerificationApplication{}
	err := r.db.QueryRow(ctx, query, userID, models.VerificationStatusRejected, withinDays).Scan(
		&app.ID,
		&app.UserID,
		&app.TwitchChannelURL,
		&app.FollowerCount,
		&app.SubscriberCount,
		&app.AvgViewers,
		&app.ContentDescription,
		&app.SocialMediaLinks,
		&app.Status,
		&app.Priority,
		&app.ReviewedBy,
		&app.ReviewedAt,
		&app.ReviewerNotes,
		&app.CreatedAt,
		&app.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil // No recent rejection is not an error
	}
	return app, err
}

// GetApplicationCountByUserID returns the total number of applications submitted by a user
func (r *VerificationRepository) GetApplicationCountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM creator_verification_applications WHERE user_id = $1`

	var count int
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}

// GetApplicationsByTwitchURL checks if there are existing applications with the same Twitch URL
// Excludes applications from the specified user ID (for checking duplicates from other users)
// This check includes all statuses (pending, approved, rejected) to prevent channel claiming abuse
func (r *VerificationRepository) GetApplicationsByTwitchURL(ctx context.Context, twitchURL string, excludeUserID uuid.UUID) ([]*models.CreatorVerificationApplication, error) {
	query := `
		SELECT id, user_id, twitch_channel_url, follower_count, subscriber_count,
			avg_viewers, content_description, social_media_links,
			status, priority, reviewed_by, reviewed_at, reviewer_notes,
			created_at, updated_at
		FROM creator_verification_applications
		WHERE LOWER(twitch_channel_url) = LOWER($1)
			AND user_id != $2
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, twitchURL, excludeUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []*models.CreatorVerificationApplication
	for rows.Next() {
		app := &models.CreatorVerificationApplication{}
		err := rows.Scan(
			&app.ID,
			&app.UserID,
			&app.TwitchChannelURL,
			&app.FollowerCount,
			&app.SubscriberCount,
			&app.AvgViewers,
			&app.ContentDescription,
			&app.SocialMediaLinks,
			&app.Status,
			&app.Priority,
			&app.ReviewedBy,
			&app.ReviewedAt,
			&app.ReviewerNotes,
			&app.CreatedAt,
			&app.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}

	return apps, rows.Err()
}

// ==============================================================================
// Audit Log Operations
// ==============================================================================

// CreateAuditLog creates a new verification audit log entry
func (r *VerificationRepository) CreateAuditLog(ctx context.Context, log *models.VerificationAuditLog) error {
	query := `
		INSERT INTO verification_audit_logs (
			user_id, audit_type, status, findings, notes, audited_by, action_taken
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at`

	err := r.db.QueryRow(ctx, query,
		log.UserID,
		log.AuditType,
		log.Status,
		log.Findings,
		log.Notes,
		log.AuditedBy,
		log.ActionTaken,
	).Scan(&log.ID, &log.CreatedAt)

	return err
}

// GetAuditLogsByUserID retrieves all audit logs for a user
func (r *VerificationRepository) GetAuditLogsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.VerificationAuditLog, error) {
	query := `
		SELECT id, user_id, audit_type, status, findings, notes, audited_by, action_taken, created_at
		FROM verification_audit_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.VerificationAuditLog
	for rows.Next() {
		log := &models.VerificationAuditLog{}
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.AuditType,
			&log.Status,
			&log.Findings,
			&log.Notes,
			&log.AuditedBy,
			&log.ActionTaken,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// GetVerifiedUsersForAudit retrieves verified users that need periodic audit
// Returns users who haven't been audited in the last N days
// Note: Includes banned users so their verification can be automatically revoked
func (r *VerificationRepository) GetVerifiedUsersForAudit(ctx context.Context, lastAuditedDaysAgo int, limit int) ([]*models.User, error) {
	query := `
		SELECT u.id, u.twitch_id, u.username, u.display_name, u.email, u.avatar_url,
			u.bio, u.karma_points, u.trust_score, u.role, u.account_type,
			u.is_verified, u.verified_at, u.is_banned, u.dmca_terminated, u.dmca_strikes_count,
			u.created_at
		FROM users u
		WHERE u.is_verified = true
			AND NOT EXISTS (
				SELECT 1 FROM verification_audit_logs val
				WHERE val.user_id = u.id
					AND val.created_at > NOW() - INTERVAL '1 day' * $1
			)
		ORDER BY u.verified_at ASC
		LIMIT $2`

	rows, err := r.db.Query(ctx, query, lastAuditedDaysAgo, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID,
			&user.TwitchID,
			&user.Username,
			&user.DisplayName,
			&user.Email,
			&user.AvatarURL,
			&user.Bio,
			&user.KarmaPoints,
			&user.TrustScore,
			&user.Role,
			&user.AccountType,
			&user.IsVerified,
			&user.VerifiedAt,
			&user.IsBanned,
			&user.DMCATerminated,
			&user.DMCAStrikesCount,
			&user.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

// GetFlaggedAudits retrieves audit logs that require attention
func (r *VerificationRepository) GetFlaggedAudits(ctx context.Context, limit, offset int) ([]*models.VerificationAuditLog, error) {
	query := `
		SELECT id, user_id, audit_type, status, findings, notes, audited_by, action_taken, created_at
		FROM verification_audit_logs
		WHERE status IN ($1, $2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.db.Query(ctx, query, models.AuditStatusFlagged, models.AuditStatusRevoked, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.VerificationAuditLog
	for rows.Next() {
		log := &models.VerificationAuditLog{}
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.AuditType,
			&log.Status,
			&log.Findings,
			&log.Notes,
			&log.AuditedBy,
			&log.ActionTaken,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// RevokeUserVerification revokes a user's verified status
func (r *VerificationRepository) RevokeUserVerification(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users
		SET is_verified = false,
			verified_at = NULL,
			updated_at = NOW()
		WHERE id = $1 AND is_verified = true`

	result, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found or already not verified")
	}

	return nil
}

// IsUserVerified checks if a user is currently verified
func (r *VerificationRepository) IsUserVerified(ctx context.Context, userID uuid.UUID) (bool, error) {
	query := `SELECT is_verified FROM users WHERE id = $1`

	var isVerified bool
	err := r.db.QueryRow(ctx, query, userID).Scan(&isVerified)
	if err == pgx.ErrNoRows {
		return false, fmt.Errorf("user not found")
	}
	return isVerified, err
}
