package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"git.subcult.tv/subculture-collective/clpr/config"
	"git.subcult.tv/subculture-collective/clpr/internal/models"
	redispkg "git.subcult.tv/subculture-collective/clpr/pkg/redis"
)

func TestClipExtractionJobService_Initialization(t *testing.T) {
	service := NewClipExtractionJobService(nil)

	assert.NotNil(t, service, "Expected service to be created")
	assert.Nil(t, service.redis, "Expected redis to be nil in test setup")
}

func TestClipExtractionJobService_EnqueueJob_NilRedis(t *testing.T) {
	service := NewClipExtractionJobService(nil)
	ctx := context.Background()

	job := &models.ClipExtractionJob{
		ClipID:    "test-clip-id",
		VODURL:    "placeholder://vod/test/test",
		StartTime: 10.0,
		EndTime:   20.0,
		Quality:   "720p",
	}

	err := service.EnqueueJob(ctx, job)

	require.Error(t, err, "Expected error when redis client is nil")
	assert.Equal(t, "redis client not available", err.Error())
}

func TestClipExtractionJobService_GetJobStatus_NilRedis(t *testing.T) {
	service := NewClipExtractionJobService(nil)
	ctx := context.Background()

	_, err := service.GetJobStatus(ctx, "test-clip-id")

	require.Error(t, err, "Expected error when redis client is nil")
	assert.Equal(t, "redis client not available", err.Error())
}

func TestClipExtractionJobService_GetPendingJobsCount_NilRedis(t *testing.T) {
	service := NewClipExtractionJobService(nil)
	ctx := context.Background()

	count, err := service.GetPendingJobsCount(ctx)

	require.Error(t, err, "Expected error when redis client is nil")
	assert.Equal(t, "redis client not available", err.Error())
	assert.Equal(t, int64(0), count)
}

// Integration tests with actual Redis client
func TestClipExtractionJobService_Integration(t *testing.T) {
	// Skip if Redis is not available
	cfg := &config.RedisConfig{
		Host:     getTestEnv("TEST_REDIS_HOST", "localhost"),
		Port:     getTestEnv("TEST_REDIS_PORT", "6380"),
		Password: "",
		DB:       1, // Use test DB
	}

	redisClient, err := redispkg.NewClient(cfg)
	if err != nil {
		t.Skip("Redis not available for testing:", err)
		return
	}
	defer redisClient.Close()

	ctx := context.Background()
	service := NewClipExtractionJobService(redisClient)

	// Clean up before and after tests
	testQueueKey := KeyJobQueue
	defer func() {
		// Clean up test data
		_ = redisClient.Delete(ctx, testQueueKey)
		_ = redisClient.Delete(ctx, "clip_extraction_job:test-clip-1")
		_ = redisClient.Delete(ctx, "clip_extraction_job:test-clip-2")
	}()

	t.Run("enqueue and retrieve job", func(t *testing.T) {
		job := &models.ClipExtractionJob{
			ClipID:    "test-clip-1",
			VODURL:    "placeholder://vod/user123/stream456",
			StartTime: 10.5,
			EndTime:   25.3,
			Quality:   "1080p",
		}

		// Enqueue job
		err := service.EnqueueJob(ctx, job)
		require.NoError(t, err, "Failed to enqueue job")

		// Check job status
		status, err := service.GetJobStatus(ctx, job.ClipID)
		require.NoError(t, err, "Failed to get job status")

		assert.Equal(t, "queued", status["status"])
		assert.Equal(t, job.ClipID, status["clip_id"])
		assert.Equal(t, job.VODURL, status["vod_url"])
		assert.Equal(t, job.StartTime, status["start_time"])
		assert.Equal(t, job.EndTime, status["end_time"])
		assert.Equal(t, job.Quality, status["quality"])

		// Check queue count
		count, err := service.GetPendingJobsCount(ctx)
		require.NoError(t, err, "Failed to get queue count")
		assert.Equal(t, int64(1), count)
	})

	t.Run("enqueue multiple jobs", func(t *testing.T) {
		// Clean up first
		_ = redisClient.Delete(ctx, testQueueKey)
		_ = redisClient.Delete(ctx, "clip_extraction_job:test-clip-1")
		_ = redisClient.Delete(ctx, "clip_extraction_job:test-clip-2")

		job1 := &models.ClipExtractionJob{
			ClipID:    "test-clip-1",
			VODURL:    "placeholder://vod/user1/stream1",
			StartTime: 5.0,
			EndTime:   15.0,
			Quality:   "720p",
		}

		job2 := &models.ClipExtractionJob{
			ClipID:    "test-clip-2",
			VODURL:    "placeholder://vod/user2/stream2",
			StartTime: 20.0,
			EndTime:   40.0,
			Quality:   "1080p",
		}

		// Enqueue both jobs
		err := service.EnqueueJob(ctx, job1)
		require.NoError(t, err)

		err = service.EnqueueJob(ctx, job2)
		require.NoError(t, err)

		// Check queue count
		count, err := service.GetPendingJobsCount(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)

		// Verify both jobs have metadata
		status1, err := service.GetJobStatus(ctx, job1.ClipID)
		require.NoError(t, err)
		assert.Equal(t, "queued", status1["status"])

		status2, err := service.GetJobStatus(ctx, job2.ClipID)
		require.NoError(t, err)
		assert.Equal(t, "queued", status2["status"])
	})

	t.Run("job not found", func(t *testing.T) {
		_, err := service.GetJobStatus(ctx, "nonexistent-clip-id")
		require.Error(t, err, "Expected error for nonexistent job")
		assert.Contains(t, err.Error(), "job not found")
	})

	t.Run("metadata TTL", func(t *testing.T) {
		job := &models.ClipExtractionJob{
			ClipID:    "test-clip-ttl",
			VODURL:    "placeholder://vod/test/test",
			StartTime: 1.0,
			EndTime:   5.0,
			Quality:   "720p",
		}

		err := service.EnqueueJob(ctx, job)
		require.NoError(t, err)

		// Check that TTL is set (should be 7 days in seconds)
		ttl, err := redisClient.TTL(ctx, "clip_extraction_job:test-clip-ttl")
		require.NoError(t, err)

		// TTL should be approximately 7 days (604800 seconds), allow some margin
		assert.Greater(t, ttl, int64(604700), "TTL should be around 7 days")
		assert.Less(t, ttl, int64(604900), "TTL should be around 7 days")

		// Clean up
		_ = redisClient.Delete(ctx, "clip_extraction_job:test-clip-ttl")
	})
}

// Test concurrent job enqueueing
func TestClipExtractionJobService_ConcurrentEnqueue(t *testing.T) {
	// Skip if Redis is not available
	cfg := &config.RedisConfig{
		Host:     getTestEnv("TEST_REDIS_HOST", "localhost"),
		Port:     getTestEnv("TEST_REDIS_PORT", "6380"),
		Password: "",
		DB:       1, // Use test DB
	}

	redisClient, err := redispkg.NewClient(cfg)
	if err != nil {
		t.Skip("Redis not available for testing:", err)
		return
	}
	defer redisClient.Close()

	ctx := context.Background()
	service := NewClipExtractionJobService(redisClient)

	// Clean up
	testQueueKey := KeyJobQueue
	_ = redisClient.Delete(ctx, testQueueKey)
	defer func() {
		_ = redisClient.Delete(ctx, testQueueKey)
		for i := 0; i < 10; i++ {
			_ = redisClient.Delete(ctx, "clip_extraction_job:concurrent-test-"+string(rune(i)))
		}
	}()

	// Enqueue 10 jobs concurrently
	numJobs := 10
	done := make(chan bool, numJobs)

	for i := 0; i < numJobs; i++ {
		go func(idx int) {
			job := &models.ClipExtractionJob{
				ClipID:    "concurrent-test-" + string(rune(idx)),
				VODURL:    "placeholder://vod/test/concurrent",
				StartTime: float64(idx),
				EndTime:   float64(idx + 10),
				Quality:   "720p",
			}
			err := service.EnqueueJob(ctx, job)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numJobs; i++ {
		<-done
	}

	// Verify queue count
	count, err := service.GetPendingJobsCount(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(numJobs), count, "All jobs should be enqueued")
}
