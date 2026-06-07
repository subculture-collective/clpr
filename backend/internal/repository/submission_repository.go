package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/internal/utils"
)

// SubmissionRepository handles database operations for clip submissions
type SubmissionRepository struct {
	db *pgxpool.Pool
}

// NewSubmissionRepository creates a new SubmissionRepository
func NewSubmissionRepository(db *pgxpool.Pool) *SubmissionRepository {
	return &SubmissionRepository{db: db}
}

// Create creates a new clip submission
func (r *SubmissionRepository) Create(ctx context.Context, submission *models.ClipSubmission) error {
	query := `
		INSERT INTO clip_submissions (
			id, user_id, clip_id, twitch_clip_id, twitch_clip_url, title, custom_title,
			tags, is_nsfw, submission_reason, status,
			creator_name, creator_id, broadcaster_name, broadcaster_id, broadcaster_name_override,
			game_id, game_name, thumbnail_url, duration, view_count,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11,
			$12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23
		)`

	_, err := r.db.Exec(ctx, query,
		submission.ID,
		submission.UserID,
		submission.ClipID,
		submission.TwitchClipID,
		submission.TwitchClipURL,
		submission.Title,
		submission.CustomTitle,
		submission.Tags,
		submission.IsNSFW,
		submission.SubmissionReason,
		submission.Status,
		submission.CreatorName,
		submission.CreatorID,
		submission.BroadcasterName,
		submission.BroadcasterID,
		submission.BroadcasterNameOverride,
		submission.GameID,
		submission.GameName,
		submission.ThumbnailURL,
		submission.Duration,
		submission.ViewCount,
		submission.CreatedAt,
		submission.UpdatedAt,
	)

	return err
}

// GetByID retrieves a submission by ID
func (r *SubmissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.ClipSubmission, error) {
	query := `
		SELECT id, user_id, clip_id, twitch_clip_id, twitch_clip_url, title, custom_title,
			tags, is_nsfw, submission_reason, status, rejection_reason,
			reviewed_by, reviewed_at, created_at, updated_at,
			creator_name, creator_id, broadcaster_name, broadcaster_id, broadcaster_name_override,
			game_id, game_name, thumbnail_url, duration, view_count
		FROM clip_submissions
		WHERE id = $1`

	var submission models.ClipSubmission
	err := r.db.QueryRow(ctx, query, id).Scan(
		&submission.ID,
		&submission.UserID,
		&submission.ClipID,
		&submission.TwitchClipID,
		&submission.TwitchClipURL,
		&submission.Title,
		&submission.CustomTitle,
		&submission.Tags,
		&submission.IsNSFW,
		&submission.SubmissionReason,
		&submission.Status,
		&submission.RejectionReason,
		&submission.ReviewedBy,
		&submission.ReviewedAt,
		&submission.CreatedAt,
		&submission.UpdatedAt,
		&submission.CreatorName,
		&submission.CreatorID,
		&submission.BroadcasterName,
		&submission.BroadcasterID,
		&submission.BroadcasterNameOverride,
		&submission.GameID,
		&submission.GameName,
		&submission.ThumbnailURL,
		&submission.Duration,
		&submission.ViewCount,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("submission not found")
	}
	if err != nil {
		return nil, err
	}

	return &submission, nil
}

// GetByTwitchClipID checks if a submission with the given Twitch clip ID exists
func (r *SubmissionRepository) GetByTwitchClipID(ctx context.Context, twitchClipID string) (*models.ClipSubmission, error) {
	query := `
		SELECT id, user_id, clip_id, twitch_clip_id, twitch_clip_url, title, custom_title,
			tags, is_nsfw, submission_reason, status, rejection_reason,
			reviewed_by, reviewed_at, created_at, updated_at,
			creator_name, creator_id, broadcaster_name, broadcaster_id, broadcaster_name_override,
			game_id, game_name, thumbnail_url, duration, view_count
		FROM clip_submissions
		WHERE twitch_clip_id = $1
		ORDER BY created_at DESC
		LIMIT 1`

	var submission models.ClipSubmission
	err := r.db.QueryRow(ctx, query, twitchClipID).Scan(
		&submission.ID,
		&submission.UserID,
		&submission.ClipID,
		&submission.TwitchClipID,
		&submission.TwitchClipURL,
		&submission.Title,
		&submission.CustomTitle,
		&submission.Tags,
		&submission.IsNSFW,
		&submission.SubmissionReason,
		&submission.Status,
		&submission.RejectionReason,
		&submission.ReviewedBy,
		&submission.ReviewedAt,
		&submission.CreatedAt,
		&submission.UpdatedAt,
		&submission.CreatorName,
		&submission.CreatorID,
		&submission.BroadcasterName,
		&submission.BroadcasterID,
		&submission.BroadcasterNameOverride,
		&submission.GameID,
		&submission.GameName,
		&submission.ThumbnailURL,
		&submission.Duration,
		&submission.ViewCount,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &submission, nil
}

// ListByUser retrieves all submissions by a user
func (r *SubmissionRepository) ListByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]*models.ClipSubmission, int, error) {
	offset := (page - 1) * limit

	// Count total
	var total int
	countQuery := `SELECT COUNT(*) FROM clip_submissions WHERE user_id = $1`
	if err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get submissions
	query := `
		SELECT id, user_id, clip_id, twitch_clip_id, twitch_clip_url, title, custom_title,
			tags, is_nsfw, submission_reason, status, rejection_reason,
			reviewed_by, reviewed_at, created_at, updated_at,
			creator_name, creator_id, broadcaster_name, broadcaster_id, broadcaster_name_override,
			game_id, game_name, thumbnail_url, duration, view_count
		FROM clip_submissions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var submissions []*models.ClipSubmission
	for rows.Next() {
		var submission models.ClipSubmission
		err := rows.Scan(
			&submission.ID,
			&submission.UserID,
			&submission.ClipID,
			&submission.TwitchClipID,
			&submission.TwitchClipURL,
			&submission.Title,
			&submission.CustomTitle,
			&submission.Tags,
			&submission.IsNSFW,
			&submission.SubmissionReason,
			&submission.Status,
			&submission.RejectionReason,
			&submission.ReviewedBy,
			&submission.ReviewedAt,
			&submission.CreatedAt,
			&submission.UpdatedAt,
			&submission.CreatorName,
			&submission.CreatorID,
			&submission.BroadcasterName,
			&submission.BroadcasterID,
			&submission.BroadcasterNameOverride,
			&submission.GameID,
			&submission.GameName,
			&submission.ThumbnailURL,
			&submission.Duration,
			&submission.ViewCount,
		)
		if err != nil {
			return nil, 0, err
		}
		submissions = append(submissions, &submission)
	}

	return submissions, total, rows.Err()
}

// SubmissionFilters represents filters for querying submissions
type SubmissionFilters struct {
	IsNSFW          *bool
	BroadcasterName *string
	CreatorName     *string
	Tags            []string
	StartDate       *time.Time
	EndDate         *time.Time
}

// ListPending retrieves all pending submissions for moderation
func (r *SubmissionRepository) ListPending(ctx context.Context, page, limit int) ([]*models.ClipSubmissionWithUser, int, error) {
	return r.ListPendingWithFilters(ctx, SubmissionFilters{}, page, limit)
}

// ListPendingWithFilters retrieves pending submissions with optional filters
func (r *SubmissionRepository) ListPendingWithFilters(ctx context.Context, filters SubmissionFilters, page, limit int) ([]*models.ClipSubmissionWithUser, int, error) {
	offset := (page - 1) * limit

	// Build where clause with filters
	whereClause := "WHERE s.status = 'pending'"
	args := []interface{}{}
	argPos := 1

	if filters.IsNSFW != nil {
		whereClause += fmt.Sprintf(" AND s.is_nsfw = %s", utils.SQLPlaceholder(argPos))
		args = append(args, *filters.IsNSFW)
		argPos++
	}

	if filters.BroadcasterName != nil {
		whereClause += fmt.Sprintf(" AND LOWER(s.broadcaster_name) LIKE LOWER(%s)", utils.SQLPlaceholder(argPos))
		args = append(args, "%"+*filters.BroadcasterName+"%")
		argPos++
	}

	if filters.CreatorName != nil {
		whereClause += fmt.Sprintf(" AND LOWER(s.creator_name) LIKE LOWER(%s)", utils.SQLPlaceholder(argPos))
		args = append(args, "%"+*filters.CreatorName+"%")
		argPos++
	}

	if len(filters.Tags) > 0 {
		whereClause += fmt.Sprintf(" AND s.tags && %s", utils.SQLPlaceholder(argPos))
		args = append(args, filters.Tags)
		argPos++
	}

	if filters.StartDate != nil {
		whereClause += fmt.Sprintf(" AND s.created_at >= %s", utils.SQLPlaceholder(argPos))
		args = append(args, *filters.StartDate)
		argPos++
	}

	if filters.EndDate != nil {
		whereClause += fmt.Sprintf(" AND s.created_at <= %s", utils.SQLPlaceholder(argPos))
		args = append(args, *filters.EndDate)
		argPos++
	}

	// Count total
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM clip_submissions s %s", whereClause)
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get submissions with user info
	query := fmt.Sprintf(`
		SELECT
			s.id, s.user_id, s.clip_id, s.twitch_clip_id, s.twitch_clip_url, s.title, s.custom_title,
			s.tags, s.is_nsfw, s.submission_reason, s.status, s.rejection_reason,
			s.reviewed_by, s.reviewed_at, s.created_at, s.updated_at,
			s.creator_name, s.creator_id, s.broadcaster_name, s.broadcaster_id, s.broadcaster_name_override,
			s.game_id, s.game_name, s.thumbnail_url, s.duration, s.view_count,
			u.id, u.twitch_id, u.username, u.display_name, u.email, u.avatar_url,
			u.bio, u.karma_points, u.role, u.is_banned, u.created_at, u.updated_at, u.last_login_at
		FROM clip_submissions s
		JOIN users u ON s.user_id = u.id
		%s
		ORDER BY s.created_at ASC
		LIMIT %s OFFSET %s`, whereClause, utils.SQLPlaceholder(argPos), utils.SQLPlaceholder(argPos+1))

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var submissions []*models.ClipSubmissionWithUser
	for rows.Next() {
		var submission models.ClipSubmissionWithUser
		var user models.User
		err := rows.Scan(
			&submission.ID,
			&submission.UserID,
			&submission.ClipID,
			&submission.TwitchClipID,
			&submission.TwitchClipURL,
			&submission.Title,
			&submission.CustomTitle,
			&submission.Tags,
			&submission.IsNSFW,
			&submission.SubmissionReason,
			&submission.Status,
			&submission.RejectionReason,
			&submission.ReviewedBy,
			&submission.ReviewedAt,
			&submission.CreatedAt,
			&submission.UpdatedAt,
			&submission.CreatorName,
			&submission.CreatorID,
			&submission.BroadcasterName,
			&submission.BroadcasterID,
			&submission.BroadcasterNameOverride,
			&submission.GameID,
			&submission.GameName,
			&submission.ThumbnailURL,
			&submission.Duration,
			&submission.ViewCount,
			&user.ID,
			&user.TwitchID,
			&user.Username,
			&user.DisplayName,
			&user.Email,
			&user.AvatarURL,
			&user.Bio,
			&user.KarmaPoints,
			&user.Role,
			&user.IsBanned,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.LastLoginAt,
		)
		if err != nil {
			return nil, 0, err
		}
		submission.User = &user
		submissions = append(submissions, &submission)
	}

	return submissions, total, rows.Err()
}

// UpdateStatus updates the status of a submission
func (r *SubmissionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, reviewedBy uuid.UUID, rejectionReason *string) error {
	query := `
		UPDATE clip_submissions
		SET status = $1, reviewed_by = $2, reviewed_at = $3, rejection_reason = $4, updated_at = $5
		WHERE id = $6`

	_, err := r.db.Exec(ctx, query, status, reviewedBy, time.Now(), rejectionReason, time.Now(), id)
	return err
}

// UpdateClipID updates the clip_id for a submission
func (r *SubmissionRepository) UpdateClipID(ctx context.Context, submissionID uuid.UUID, clipID uuid.UUID) error {
	query := `
		UPDATE clip_submissions
		SET clip_id = $1, updated_at = $2
		WHERE id = $3`

	_, err := r.db.Exec(ctx, query, clipID, time.Now(), submissionID)
	return err
}

// CountUserSubmissions counts recent submissions by a user within a time window
func (r *SubmissionRepository) CountUserSubmissions(ctx context.Context, userID uuid.UUID, since time.Time) (int, error) {
	query := `SELECT COUNT(*) FROM clip_submissions WHERE user_id = $1 AND created_at > $2`
	var count int
	err := r.db.QueryRow(ctx, query, userID, since).Scan(&count)
	return count, err
}

// GetUserStats retrieves submission statistics for a user
func (r *SubmissionRepository) GetUserStats(ctx context.Context, userID uuid.UUID) (*models.SubmissionStats, error) {
	query := `SELECT * FROM submission_stats WHERE user_id = $1`

	var stats models.SubmissionStats
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&stats.UserID,
		&stats.TotalCount,
		&stats.ApprovedCount,
		&stats.RejectedCount,
		&stats.PendingCount,
		&stats.ApprovalRate,
	)

	if err == pgx.ErrNoRows {
		// Return empty stats if no submissions yet
		return &models.SubmissionStats{
			UserID:        userID,
			TotalCount:    0,
			ApprovedCount: 0,
			RejectedCount: 0,
			PendingCount:  0,
			ApprovalRate:  0,
		}, nil
	}
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

// GetByIDs retrieves multiple submissions by their IDs
func (r *SubmissionRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.ClipSubmission, error) {
	query := `
		SELECT id, user_id, twitch_clip_id, twitch_clip_url, title, custom_title,
			tags, is_nsfw, submission_reason, status, rejection_reason,
			reviewed_by, reviewed_at, created_at, updated_at,
			creator_name, creator_id, broadcaster_name, broadcaster_id, broadcaster_name_override,
			game_id, game_name, thumbnail_url, duration, view_count
		FROM clip_submissions
		WHERE id = ANY($1)`

	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var submissions []*models.ClipSubmission
	for rows.Next() {
		var submission models.ClipSubmission
		err := rows.Scan(
			&submission.ID,
			&submission.UserID,
			&submission.TwitchClipID,
			&submission.TwitchClipURL,
			&submission.Title,
			&submission.CustomTitle,
			&submission.Tags,
			&submission.IsNSFW,
			&submission.SubmissionReason,
			&submission.Status,
			&submission.RejectionReason,
			&submission.ReviewedBy,
			&submission.ReviewedAt,
			&submission.CreatedAt,
			&submission.UpdatedAt,
			&submission.CreatorName,
			&submission.CreatorID,
			&submission.BroadcasterName,
			&submission.BroadcasterID,
			&submission.BroadcasterNameOverride,
			&submission.GameID,
			&submission.GameName,
			&submission.ThumbnailURL,
			&submission.Duration,
			&submission.ViewCount,
		)
		if err != nil {
			return nil, err
		}
		submissions = append(submissions, &submission)
	}

	return submissions, rows.Err()
}

// BulkUpdateStatus updates the status of multiple submissions
func (r *SubmissionRepository) BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status string, reviewedBy uuid.UUID, rejectionReason *string) error {
	query := `
		UPDATE clip_submissions
		SET status = $1, reviewed_by = $2, reviewed_at = $3, rejection_reason = $4, updated_at = $5
		WHERE id = ANY($6)`

	_, err := r.db.Exec(ctx, query, status, reviewedBy, time.Now(), rejectionReason, time.Now(), ids)
	return err
}
