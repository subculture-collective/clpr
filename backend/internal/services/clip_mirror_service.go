package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

// MirrorRepositoryContract captures repository calls ClipMirrorService relies on.
type MirrorRepositoryContract interface {
	Create(ctx context.Context, mirror *models.ClipMirror) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.ClipMirror, error)
	GetByClipAndRegion(ctx context.Context, clipID uuid.UUID, region string) (*models.ClipMirror, error)
	ListByClip(ctx context.Context, clipID uuid.UUID) ([]*models.ClipMirror, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, failureReason *string) error
	RecordAccess(ctx context.Context, id uuid.UUID) error
	DeleteExpired(ctx context.Context) (int64, error)
	CreateMetric(ctx context.Context, metric *models.MirrorMetrics) error
	GetMirrorHitRate(ctx context.Context, startTime time.Time) (float64, error)
	GetPopularClipsForMirroring(ctx context.Context, threshold int, maxMirrors int, limit int) ([]uuid.UUID, error)
}

// ClipRepositoryContract captures clip repository calls ClipMirrorService uses.
type ClipRepositoryContract interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.Clip, error)
}

// ClipMirrorService handles clip mirroring across regions
type ClipMirrorService struct {
	mirrorRepo MirrorRepositoryContract
	clipRepo   ClipRepositoryContract
	config     *config.MirrorConfig
}

// NewClipMirrorService creates a new ClipMirrorService
func NewClipMirrorService(
	mirrorRepo MirrorRepositoryContract,
	clipRepo ClipRepositoryContract,
	config *config.MirrorConfig,
) *ClipMirrorService {
	return &ClipMirrorService{
		mirrorRepo: mirrorRepo,
		clipRepo:   clipRepo,
		config:     config,
	}
}

// IdentifyPopularClips identifies clips that should be mirrored based on popularity
func (s *ClipMirrorService) IdentifyPopularClips(ctx context.Context) ([]uuid.UUID, error) {
	if !s.config.Enabled {
		return nil, nil
	}

	// Get clips that meet the replication threshold
	clipIDs, err := s.mirrorRepo.GetPopularClipsForMirroring(
		ctx,
		s.config.ReplicationThreshold,
		s.config.MaxMirrorsPerClip,
		100, // Process up to 100 clips per run
	)
	if err != nil {
		return nil, fmt.Errorf("failed to identify popular clips: %w", err)
	}

	return clipIDs, nil
}

// ReplicateClip replicates a clip to configured regions
func (s *ClipMirrorService) ReplicateClip(ctx context.Context, clipID uuid.UUID) error {
	if !s.config.Enabled {
		return nil
	}

	// Get the clip details
	clip, err := s.clipRepo.GetByID(ctx, clipID)
	if err != nil {
		return fmt.Errorf("failed to get clip: %w", err)
	}

	// Get existing mirrors
	existingMirrors, err := s.mirrorRepo.ListByClip(ctx, clipID)
	if err != nil {
		return fmt.Errorf("failed to list existing mirrors: %w", err)
	}

	// Check if we've reached the max mirrors per clip
	activeMirrors := 0
	existingRegions := make(map[string]bool)
	for _, mirror := range existingMirrors {
		existingRegions[mirror.Region] = true
		if mirror.Status == models.MirrorStatusActive {
			activeMirrors++
		}
	}

	if activeMirrors >= s.config.MaxMirrorsPerClip {
		utils.Debug("Clip already has max active mirrors, skipping", map[string]interface{}{"clip_id": clipID, "active_mirrors": activeMirrors})
		return nil
	}

	// Replicate to regions that don't have mirrors yet
	for _, region := range s.config.Regions {
		if existingRegions[region] {
			continue
		}

		if activeMirrors >= s.config.MaxMirrorsPerClip {
			break
		}

		// Create mirror record
		mirror := &models.ClipMirror{
			ID:              uuid.New(),
			ClipID:          clipID,
			Region:          region,
			Status:          models.MirrorStatusPending,
			StorageProvider: s.determineStorageProvider(region),
			CreatedAt:       time.Now(),
			AccessCount:     0,
		}

		// Set expiration time based on TTL
		expiresAt := time.Now().AddDate(0, 0, s.config.TTLDays)
		mirror.ExpiresAt = &expiresAt

		// Generate mirror URL (this would be replaced with actual storage URL)
		mirror.MirrorURL = s.generateMirrorURL(clip, region)

		// Create the mirror record
		if err := s.mirrorRepo.Create(ctx, mirror); err != nil {
			utils.Warn("Failed to create mirror", map[string]interface{}{"clip_id": clipID, "region": region, "error": err})
			continue
		}

		// In a real implementation, this would trigger the actual file replication
		// For now, we just mark it as active
		if err := s.mirrorRepo.UpdateStatus(ctx, mirror.ID, models.MirrorStatusActive, nil); err != nil {
			utils.Warn("Failed to update mirror status", map[string]interface{}{"clip_id": clipID, "region": region, "error": err})
			continue
		}

		activeMirrors++
		utils.Info("Successfully created mirror", map[string]interface{}{"clip_id": clipID, "region": region})
	}

	return nil
}

// GetMirrorURL returns the best mirror URL for a given clip and user region
func (s *ClipMirrorService) GetMirrorURL(ctx context.Context, clipID uuid.UUID, userRegion string) (string, bool, error) {
	if !s.config.Enabled {
		return "", false, nil
	}

	// Try to get mirror from user's region first
	mirror, err := s.mirrorRepo.GetByClipAndRegion(ctx, clipID, userRegion)
	if err == nil && mirror.Status == models.MirrorStatusActive {
		// Record access
		if err := s.mirrorRepo.RecordAccess(ctx, mirror.ID); err != nil {
			utils.Warn("Failed to record mirror access", map[string]interface{}{"error": err})
		}

		// Record metric
		if err := s.recordMetric(ctx, clipID, userRegion, models.MirrorMetricTypeAccess, 1); err != nil {
			utils.Warn("Failed to record mirror metric", map[string]interface{}{"error": err})
		}

		return mirror.MirrorURL, true, nil
	}

	// Try to get any active mirror
	mirrors, err := s.mirrorRepo.ListByClip(ctx, clipID)
	if err != nil {
		return "", false, fmt.Errorf("failed to list mirrors: %w", err)
	}

	for _, mirror := range mirrors {
		if mirror.Status == models.MirrorStatusActive {
			// Record access
			if err := s.mirrorRepo.RecordAccess(ctx, mirror.ID); err != nil {
				utils.Warn("Failed to record mirror access", map[string]interface{}{"error": err})
			}

			// Record metric
			if err := s.recordMetric(ctx, clipID, mirror.Region, models.MirrorMetricTypeAccess, 1); err != nil {
				utils.Warn("Failed to record mirror metric", map[string]interface{}{"error": err})
			}

			return mirror.MirrorURL, true, nil
		}
	}

	// Record failover metric
	if err := s.recordMetric(ctx, clipID, userRegion, models.MirrorMetricTypeFailover, 1); err != nil {
		utils.Warn("Failed to record failover metric", map[string]interface{}{"error": err})
	}

	return "", false, nil
}

// CleanupExpiredMirrors removes expired mirrors
func (s *ClipMirrorService) CleanupExpiredMirrors(ctx context.Context) (int64, error) {
	if !s.config.Enabled {
		return 0, nil
	}

	count, err := s.mirrorRepo.DeleteExpired(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired mirrors: %w", err)
	}

	utils.Info("Cleaned up expired mirrors", map[string]interface{}{"count": count})
	return count, nil
}

// GetMirrorHitRate returns the mirror hit rate for the last 24 hours
func (s *ClipMirrorService) GetMirrorHitRate(ctx context.Context) (float64, error) {
	if !s.config.Enabled {
		return 0, nil
	}

	startTime := time.Now().Add(-24 * time.Hour)
	hitRate, err := s.mirrorRepo.GetMirrorHitRate(ctx, startTime)
	if err != nil {
		return 0, fmt.Errorf("failed to get mirror hit rate: %w", err)
	}

	return hitRate, nil
}

// SyncPopularClips is the main background job that identifies and replicates popular clips
func (s *ClipMirrorService) SyncPopularClips(ctx context.Context) error {
	if !s.config.Enabled {
		return nil
	}

	utils.Info("Starting mirror sync for popular clips", nil)

	// Identify popular clips
	clipIDs, err := s.IdentifyPopularClips(ctx)
	if err != nil {
		return fmt.Errorf("failed to identify popular clips: %w", err)
	}

	utils.Info("Found clips to mirror", map[string]interface{}{"count": len(clipIDs)})

	// Replicate each clip
	successCount := 0
	for _, clipID := range clipIDs {
		if err := s.ReplicateClip(ctx, clipID); err != nil {
			utils.Warn("Failed to replicate clip", map[string]interface{}{"clip_id": clipID, "error": err})
			continue
		}
		successCount++
	}

	utils.Info("Mirror sync complete", map[string]interface{}{"success_count": successCount, "total": len(clipIDs)})
	return nil
}

// determineStorageProvider determines which storage provider to use for a region
func (s *ClipMirrorService) determineStorageProvider(region string) string {
	// In a real implementation, this would be configurable per region
	// For now, we'll use a simple mapping
	switch region {
	case "us-east-1", "us-west-2":
		return "s3"
	case "eu-west-1", "eu-central-1":
		return "cloudflare-r2"
	case "ap-southeast-1", "ap-northeast-1":
		return "bunny"
	default:
		return "s3"
	}
}

// generateMirrorURL generates a mirror URL for a clip in a specific region
func (s *ClipMirrorService) generateMirrorURL(clip *models.Clip, region string) string {
	// In a real implementation, this would generate actual storage URLs
	// For now, we'll use a placeholder format
	provider := s.determineStorageProvider(region)
	return fmt.Sprintf("https://%s.%s.clpr.cdn/%s/%s.mp4",
		provider, region, clip.TwitchClipID, clip.TwitchClipID)
}

// recordMetric records a mirror metric
func (s *ClipMirrorService) recordMetric(ctx context.Context, clipID uuid.UUID, region string, metricType string, value float64) error {
	metric := &models.MirrorMetrics{
		ClipID:      clipID,
		Region:      region,
		MetricType:  metricType,
		MetricValue: value,
		RecordedAt:  time.Now(),
	}

	return s.mirrorRepo.CreateMetric(ctx, metric)
}
