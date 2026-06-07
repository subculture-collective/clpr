package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
)

// MirrorRepository handles database operations for clip mirrors
type MirrorRepository struct {
	db *pgxpool.Pool
}

// NewMirrorRepository creates a new MirrorRepository
func NewMirrorRepository(db *pgxpool.Pool) *MirrorRepository {
	return &MirrorRepository{db: db}
}

// Create creates a new clip mirror
func (r *MirrorRepository) Create(ctx context.Context, mirror *models.ClipMirror) error {
	query := `
		INSERT INTO clip_mirrors (
			id, clip_id, region, mirror_url, status, storage_provider,
			size_bytes, created_at, access_count, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(
		ctx, query,
		mirror.ID, mirror.ClipID, mirror.Region, mirror.MirrorURL, mirror.Status,
		mirror.StorageProvider, mirror.SizeBytes, time.Now(), 0, mirror.ExpiresAt,
	).Scan(&mirror.ID, &mirror.CreatedAt)

	return err
}

// GetByID retrieves a mirror by ID
func (r *MirrorRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.ClipMirror, error) {
	query := `
		SELECT id, clip_id, region, mirror_url, status, storage_provider,
			   size_bytes, created_at, last_accessed_at, access_count, expires_at, failure_reason
		FROM clip_mirrors
		WHERE id = $1
	`

	var mirror models.ClipMirror
	err := r.db.QueryRow(ctx, query, id).Scan(
		&mirror.ID, &mirror.ClipID, &mirror.Region, &mirror.MirrorURL, &mirror.Status,
		&mirror.StorageProvider, &mirror.SizeBytes, &mirror.CreatedAt, &mirror.LastAccessedAt,
		&mirror.AccessCount, &mirror.ExpiresAt, &mirror.FailureReason,
	)

	if err != nil {
		return nil, err
	}

	return &mirror, nil
}

// GetByClipAndRegion retrieves a mirror by clip ID and region
func (r *MirrorRepository) GetByClipAndRegion(ctx context.Context, clipID uuid.UUID, region string) (*models.ClipMirror, error) {
	query := `
		SELECT id, clip_id, region, mirror_url, status, storage_provider,
			   size_bytes, created_at, last_accessed_at, access_count, expires_at, failure_reason
		FROM clip_mirrors
		WHERE clip_id = $1 AND region = $2
	`

	var mirror models.ClipMirror
	err := r.db.QueryRow(ctx, query, clipID, region).Scan(
		&mirror.ID, &mirror.ClipID, &mirror.Region, &mirror.MirrorURL, &mirror.Status,
		&mirror.StorageProvider, &mirror.SizeBytes, &mirror.CreatedAt, &mirror.LastAccessedAt,
		&mirror.AccessCount, &mirror.ExpiresAt, &mirror.FailureReason,
	)

	if err != nil {
		return nil, err
	}

	return &mirror, nil
}

// ListByClip retrieves all mirrors for a clip
func (r *MirrorRepository) ListByClip(ctx context.Context, clipID uuid.UUID) ([]*models.ClipMirror, error) {
	query := `
		SELECT id, clip_id, region, mirror_url, status, storage_provider,
			   size_bytes, created_at, last_accessed_at, access_count, expires_at, failure_reason
		FROM clip_mirrors
		WHERE clip_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, clipID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mirrors []*models.ClipMirror
	for rows.Next() {
		var mirror models.ClipMirror
		err := rows.Scan(
			&mirror.ID, &mirror.ClipID, &mirror.Region, &mirror.MirrorURL, &mirror.Status,
			&mirror.StorageProvider, &mirror.SizeBytes, &mirror.CreatedAt, &mirror.LastAccessedAt,
			&mirror.AccessCount, &mirror.ExpiresAt, &mirror.FailureReason,
		)
		if err != nil {
			return nil, err
		}
		mirrors = append(mirrors, &mirror)
	}

	return mirrors, rows.Err()
}

// ListActiveByRegion retrieves all active mirrors in a region
func (r *MirrorRepository) ListActiveByRegion(ctx context.Context, region string) ([]*models.ClipMirror, error) {
	query := `
		SELECT id, clip_id, region, mirror_url, status, storage_provider,
			   size_bytes, created_at, last_accessed_at, access_count, expires_at, failure_reason
		FROM clip_mirrors
		WHERE region = $1 AND status = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, region, models.MirrorStatusActive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mirrors []*models.ClipMirror
	for rows.Next() {
		var mirror models.ClipMirror
		err := rows.Scan(
			&mirror.ID, &mirror.ClipID, &mirror.Region, &mirror.MirrorURL, &mirror.Status,
			&mirror.StorageProvider, &mirror.SizeBytes, &mirror.CreatedAt, &mirror.LastAccessedAt,
			&mirror.AccessCount, &mirror.ExpiresAt, &mirror.FailureReason,
		)
		if err != nil {
			return nil, err
		}
		mirrors = append(mirrors, &mirror)
	}

	return mirrors, rows.Err()
}

// UpdateStatus updates the status of a mirror
func (r *MirrorRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, failureReason *string) error {
	query := `
		UPDATE clip_mirrors
		SET status = $1, failure_reason = $2
		WHERE id = $3
	`

	_, err := r.db.Exec(ctx, query, status, failureReason, id)
	return err
}

// RecordAccess records an access to a mirror
func (r *MirrorRepository) RecordAccess(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE clip_mirrors
		SET access_count = access_count + 1,
			last_accessed_at = $1
		WHERE id = $2
	`

	_, err := r.db.Exec(ctx, query, time.Now(), id)
	return err
}

// DeleteExpired deletes expired mirrors
func (r *MirrorRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `
		DELETE FROM clip_mirrors
		WHERE expires_at IS NOT NULL AND expires_at < $1
	`

	result, err := r.db.Exec(ctx, query, time.Now())
	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}

// Delete deletes a mirror by ID
func (r *MirrorRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM clip_mirrors WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// CreateMetric creates a mirror metric
func (r *MirrorRepository) CreateMetric(ctx context.Context, metric *models.MirrorMetrics) error {
	metadataJSON, err := json.Marshal(metric.Metadata)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO mirror_metrics (
			id, clip_id, region, metric_type, metric_value, recorded_at, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = r.db.Exec(
		ctx, query,
		uuid.New(), metric.ClipID, metric.Region, metric.MetricType,
		metric.MetricValue, time.Now(), metadataJSON,
	)

	return err
}

// GetMirrorHitRate calculates the mirror hit rate for a time period
// Hit rate = (successful mirror accesses / total access attempts) * 100
// where total attempts = successful accesses + failovers to primary
func (r *MirrorRepository) GetMirrorHitRate(ctx context.Context, startTime time.Time) (float64, error) {
	query := `
		SELECT 
			COALESCE(
				SUM(CASE WHEN metric_type = 'access' THEN metric_value ELSE 0 END) * 100.0 /
				NULLIF(SUM(metric_value), 0),
				0
			) as hit_rate
		FROM mirror_metrics
		WHERE recorded_at >= $1
			AND metric_type IN ('access', 'failover')
	`

	var hitRate float64
	err := r.db.QueryRow(ctx, query, startTime).Scan(&hitRate)
	if err != nil {
		return 0, err
	}

	return hitRate, nil
}

// GetPopularClipsForMirroring returns clips that meet the replication threshold
// maxMirrors parameter should match the MIRROR_MAX_PER_CLIP configuration value
func (r *MirrorRepository) GetPopularClipsForMirroring(ctx context.Context, threshold int, maxMirrors int, limit int) ([]uuid.UUID, error) {
	query := `
		SELECT c.id
		FROM clips c
		LEFT JOIN clip_mirrors cm ON c.id = cm.clip_id AND cm.status = 'active'
		WHERE (c.view_count >= $1 OR c.vote_score >= $1)
			AND c.is_removed = false
			AND c.dmca_removed = false
		GROUP BY c.id
		HAVING COUNT(cm.id) < $2
		ORDER BY c.view_count DESC, c.vote_score DESC
		LIMIT $3
	`

	rows, err := r.db.Query(ctx, query, threshold, maxMirrors, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clipIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		clipIDs = append(clipIDs, id)
	}

	return clipIDs, rows.Err()
}
