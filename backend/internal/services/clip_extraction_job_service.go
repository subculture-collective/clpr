package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"git.subcult.tv/subculture-collective/clpr/internal/models"
	"git.subcult.tv/subculture-collective/clpr/pkg/redis"
	"git.subcult.tv/subculture-collective/clpr/pkg/utils"
)

// Cache key constants for clip extraction jobs
const (
	KeyJobQueue    = "clip_extraction_jobs"
	KeyJobMetadata = "clip_extraction_job:%s" // clipId
)

// Cache TTL constants for clip extraction jobs
const (
	TTLJobMetadata = 7 * 24 * time.Hour // 7 days
)

// ClipExtractionJobService handles enqueueing and managing clip extraction jobs
type ClipExtractionJobService struct {
	redis *redis.Client
}

// NewClipExtractionJobService creates a new clip extraction job service
func NewClipExtractionJobService(redis *redis.Client) *ClipExtractionJobService {
	return &ClipExtractionJobService{
		redis: redis,
	}
}

// EnqueueJob adds a clip extraction job to the Redis queue
func (s *ClipExtractionJobService) EnqueueJob(ctx context.Context, job *models.ClipExtractionJob) error {
	if s.redis == nil {
		return fmt.Errorf("redis client not available")
	}

	// Serialize job to JSON
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	// Store job metadata first to avoid race condition
	// If metadata storage fails, we don't enqueue the job
	jobKey := fmt.Sprintf(KeyJobMetadata, job.ClipID)
	jobMetadata := map[string]interface{}{
		"status":     "queued",
		"queued_at":  time.Now().Unix(),
		"clip_id":    job.ClipID,
		"vod_url":    job.VODURL,
		"start_time": job.StartTime,
		"end_time":   job.EndTime,
		"quality":    job.Quality,
	}

	metadataJSON, err := json.Marshal(jobMetadata)
	if err != nil {
		return fmt.Errorf("failed to marshal job metadata: %w", err)
	}

	if err := s.redis.Set(ctx, jobKey, string(metadataJSON), TTLJobMetadata); err != nil {
		return fmt.Errorf("failed to store job metadata: %w", err)
	}

	// Now push job to Redis list (queue)
	if err := s.redis.ListPush(ctx, KeyJobQueue, jobData); err != nil {
		// Cleanup metadata if queue push fails
		_ = s.redis.Delete(ctx, jobKey)
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	utils.GetLogger().Info("Clip extraction job enqueued", map[string]interface{}{
		"clip_id":    job.ClipID,
		"vod_url":    job.VODURL,
		"start_time": job.StartTime,
		"end_time":   job.EndTime,
		"quality":    job.Quality,
	})

	return nil
}

// GetJobStatus retrieves the status of a clip extraction job
func (s *ClipExtractionJobService) GetJobStatus(ctx context.Context, clipID string) (map[string]interface{}, error) {
	if s.redis == nil {
		return nil, fmt.Errorf("redis client not available")
	}

	jobKey := fmt.Sprintf(KeyJobMetadata, clipID)
	data, err := s.redis.Get(ctx, jobKey)
	if err != nil {
		return nil, fmt.Errorf("job not found: %w", err)
	}

	var jobMetadata map[string]interface{}
	if err := json.Unmarshal([]byte(data), &jobMetadata); err != nil {
		return nil, fmt.Errorf("failed to parse job metadata: %w", err)
	}

	return jobMetadata, nil
}

// GetPendingJobsCount returns the number of jobs in the queue
func (s *ClipExtractionJobService) GetPendingJobsCount(ctx context.Context) (int64, error) {
	if s.redis == nil {
		return 0, fmt.Errorf("redis client not available")
	}

	return s.redis.ListLen(ctx, KeyJobQueue)
}
