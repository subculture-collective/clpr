package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// ExportRepository handles database operations for data exports
type ExportRepository struct {
	pool *pgxpool.Pool
}

// NewExportRepository creates a new export repository
func NewExportRepository(pool *pgxpool.Pool) *ExportRepository {
	return &ExportRepository{
		pool: pool,
	}
}

// CreateExportRequest creates a new export request
func (r *ExportRepository) CreateExportRequest(ctx context.Context, req *models.ExportRequest) error {
	query := `
		INSERT INTO export_requests (
			id, user_id, creator_name, format, status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
	`
	_, err := r.pool.Exec(ctx, query,
		req.ID,
		req.UserID,
		req.CreatorName,
		req.Format,
		req.Status,
		req.CreatedAt,
		req.UpdatedAt,
	)
	return err
}

// GetExportRequestByID retrieves an export request by ID
func (r *ExportRepository) GetExportRequestByID(ctx context.Context, id uuid.UUID) (*models.ExportRequest, error) {
	query := `
		SELECT 
			id, user_id, creator_name, format, status, file_path, 
			file_size_bytes, error_message, expires_at, email_sent,
			created_at, updated_at, completed_at
		FROM export_requests
		WHERE id = $1
	`
	var req models.ExportRequest
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&req.ID,
		&req.UserID,
		&req.CreatorName,
		&req.Format,
		&req.Status,
		&req.FilePath,
		&req.FileSizeBytes,
		&req.ErrorMessage,
		&req.ExpiresAt,
		&req.EmailSent,
		&req.CreatedAt,
		&req.UpdatedAt,
		&req.CompletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

// GetUserExportRequests retrieves all export requests for a user
func (r *ExportRepository) GetUserExportRequests(ctx context.Context, userID uuid.UUID, limit int) ([]*models.ExportRequest, error) {
	query := `
		SELECT 
			id, user_id, creator_name, format, status, file_path,
			file_size_bytes, error_message, expires_at, email_sent,
			created_at, updated_at, completed_at
		FROM export_requests
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*models.ExportRequest
	for rows.Next() {
		var req models.ExportRequest
		err := rows.Scan(
			&req.ID,
			&req.UserID,
			&req.CreatorName,
			&req.Format,
			&req.Status,
			&req.FilePath,
			&req.FileSizeBytes,
			&req.ErrorMessage,
			&req.ExpiresAt,
			&req.EmailSent,
			&req.CreatedAt,
			&req.UpdatedAt,
			&req.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, &req)
	}
	return requests, rows.Err()
}

// UpdateExportStatus updates the status of an export request
func (r *ExportRepository) UpdateExportStatus(ctx context.Context, id uuid.UUID, status string, errorMsg *string) error {
	query := `
		UPDATE export_requests
		SET status = $2, error_message = $3, updated_at = $4
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, id, status, errorMsg, time.Now())
	return err
}

// CompleteExportRequest marks an export request as completed
func (r *ExportRepository) CompleteExportRequest(ctx context.Context, id uuid.UUID, filePath string, fileSize int64, expiresAt time.Time) error {
	query := `
		UPDATE export_requests
		SET 
			status = $2,
			file_path = $3,
			file_size_bytes = $4,
			expires_at = $5,
			completed_at = $6,
			updated_at = $7
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query,
		id,
		models.ExportStatusCompleted,
		filePath,
		fileSize,
		expiresAt,
		time.Now(),
		time.Now(),
	)
	return err
}

// MarkEmailSent marks that an email notification has been sent for an export
func (r *ExportRepository) MarkEmailSent(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE export_requests
		SET email_sent = true, updated_at = $2
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, id, time.Now())
	return err
}

// GetPendingExportRequests retrieves all pending export requests
func (r *ExportRepository) GetPendingExportRequests(ctx context.Context, limit int) ([]*models.ExportRequest, error) {
	query := `
		SELECT 
			id, user_id, creator_name, format, status, file_path,
			file_size_bytes, error_message, expires_at, email_sent,
			created_at, updated_at, completed_at
		FROM export_requests
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT $2
	`
	rows, err := r.pool.Query(ctx, query, models.ExportStatusPending, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*models.ExportRequest
	for rows.Next() {
		var req models.ExportRequest
		err := rows.Scan(
			&req.ID,
			&req.UserID,
			&req.CreatorName,
			&req.Format,
			&req.Status,
			&req.FilePath,
			&req.FileSizeBytes,
			&req.ErrorMessage,
			&req.ExpiresAt,
			&req.EmailSent,
			&req.CreatedAt,
			&req.UpdatedAt,
			&req.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, &req)
	}
	return requests, rows.Err()
}

// GetExpiredExportRequests retrieves export requests that have expired
func (r *ExportRepository) GetExpiredExportRequests(ctx context.Context) ([]*models.ExportRequest, error) {
	query := `
		SELECT 
			id, user_id, creator_name, format, status, file_path,
			file_size_bytes, error_message, expires_at, email_sent,
			created_at, updated_at, completed_at
		FROM export_requests
		WHERE status = $1 AND expires_at < $2
	`
	rows, err := r.pool.Query(ctx, query, models.ExportStatusCompleted, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*models.ExportRequest
	for rows.Next() {
		var req models.ExportRequest
		err := rows.Scan(
			&req.ID,
			&req.UserID,
			&req.CreatorName,
			&req.Format,
			&req.Status,
			&req.FilePath,
			&req.FileSizeBytes,
			&req.ErrorMessage,
			&req.ExpiresAt,
			&req.EmailSent,
			&req.CreatedAt,
			&req.UpdatedAt,
			&req.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		requests = append(requests, &req)
	}
	return requests, rows.Err()
}

// MarkExportExpired marks an export request as expired
func (r *ExportRepository) MarkExportExpired(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE export_requests
		SET status = $2, updated_at = $3
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, id, models.ExportStatusExpired, time.Now())
	return err
}

// GetCreatorClipsForExport retrieves all clips for a creator for export purposes
func (r *ExportRepository) GetCreatorClipsForExport(ctx context.Context, creatorName string) ([]*models.Clip, error) {
	query := `
		SELECT 
			id, twitch_clip_id, twitch_clip_url, embed_url, title,
			creator_name, creator_id, broadcaster_name, broadcaster_id,
			game_id, game_name, language, thumbnail_url, duration,
			view_count, created_at, imported_at, vote_score,
			comment_count, favorite_count, is_featured, is_nsfw,
			is_removed, removed_reason, is_hidden
		FROM clips
		WHERE creator_name = $1 AND is_removed = false
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, creatorName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clips []*models.Clip
	for rows.Next() {
		var clip models.Clip
		err := rows.Scan(
			&clip.ID,
			&clip.TwitchClipID,
			&clip.TwitchClipURL,
			&clip.EmbedURL,
			&clip.Title,
			&clip.CreatorName,
			&clip.CreatorID,
			&clip.BroadcasterName,
			&clip.BroadcasterID,
			&clip.GameID,
			&clip.GameName,
			&clip.Language,
			&clip.ThumbnailURL,
			&clip.Duration,
			&clip.ViewCount,
			&clip.CreatedAt,
			&clip.ImportedAt,
			&clip.VoteScore,
			&clip.CommentCount,
			&clip.FavoriteCount,
			&clip.IsFeatured,
			&clip.IsNSFW,
			&clip.IsRemoved,
			&clip.RemovedReason,
			&clip.IsHidden,
		)
		if err != nil {
			return nil, err
		}
		clips = append(clips, &clip)
	}
	return clips, rows.Err()
}
