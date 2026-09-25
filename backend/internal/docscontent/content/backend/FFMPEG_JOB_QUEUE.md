---
title: "FFMPEG JOB QUEUE"
summary: "This implementation provides a Redis-based job queue system for processing FFmpeg video extraction jobs from live Twitch streams. When users create clips from live streams, a background job is enqueue"
tags: ["docs"]
area: "docs"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# FFmpeg Video Extraction Job Queue

## Overview

This implementation provides a Redis-based job queue system for processing FFmpeg video extraction jobs from live Twitch streams. When users create clips from live streams, a background job is enqueued to extract the video segment using FFmpeg.

## Architecture

### Components

1. **ClipExtractionJobService** (`internal/services/clip_extraction_job_service.go`)
   - Manages job enqueueing and status tracking
   - Uses Redis for queue storage and job metadata

2. **StreamHandler** (`internal/handlers/stream_handler.go`)
   - Enqueues jobs when clips are created from streams
   - Passes clip metadata and timing information to the job service

3. **Redis Queue**
   - Jobs stored in Redis list `clip_extraction_jobs` (FIFO)
   - Job metadata stored with 7-day TTL at `clip_extraction_job:{clip_id}`

## Job Structure

```go
type ClipExtractionJob struct {
    ClipID    string  // UUID of the clip
    VODURL    string  // URL to the stream VOD
    StartTime float64 // Start time in seconds
    EndTime   float64 // End time in seconds
    Quality   string  // Video quality (e.g., "720p", "1080p")
}
```

## Usage

### Enqueueing a Job

When a user creates a clip from a live stream, the `CreateClipFromStream` handler automatically enqueues a job:

```go
job := &models.ClipExtractionJob{
    ClipID:    clipID.String(),
    VODURL:    vodURL,
    StartTime: req.StartTime,
    EndTime:   req.EndTime,
    Quality:   req.Quality,
}

err := jobService.EnqueueJob(ctx, job)
```

### Checking Job Status

```go
status, err := jobService.GetJobStatus(ctx, clipID)
// Returns map with: status, queued_at, clip_id, vod_url, start_time, end_time, quality
```

### Getting Queue Count

```go
count, err := jobService.GetPendingJobsCount(ctx)
```

## Worker Implementation (TODO)

The actual FFmpeg video processing is **not yet implemented**. A worker service needs to be created to:

1. **Poll the Redis queue** for pending jobs
   ```go
   // Dequeue job from Redis
   jobJSON, err := redisClient.ListPop(ctx, "clip_extraction_jobs")
   if err == redis.Nil {
       // No jobs in queue
       return
   }
   if err != nil {
       // Handle error
       return
   }
   
   var job models.ClipExtractionJob
   if err := json.Unmarshal([]byte(jobJSON), &job); err != nil {
       // Handle unmarshaling error
       return
   }
   ```

2. **Download the VOD** from Twitch
   - Use the VOD URL from the job
   - May require Twitch API authentication

3. **Extract video segment** using FFmpeg
   ```bash
   ffmpeg -ss {start_time} -to {end_time} -i {vod_url} \
          -c copy -avoid_negative_ts make_zero output.mp4
   ```

4. **Upload to storage** (S3/CDN)
   - Configure CDN provider in `config.CDNConfig`
   - Upload extracted clip

5. **Update clip record** in database
   - Update `status` to "ready" or "failed"
   - Set `twitch_clip_url` to final storage URL
   - Update `thumbnail_url`

6. **Update job metadata** in Redis
   - Set status to "completed" or "failed"
   - Add completion timestamp
   - Add error message if failed

7. **Error handling and retries**
   - Implement exponential backoff for failures
   - Maximum retry attempts (e.g., 3)
   - Dead letter queue for persistent failures

## Configuration

The job queue uses the existing Redis configuration from `config.RedisConfig`:

```env
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
```

## Error Handling

- If Redis is unavailable, jobs are not enqueued but clip creation still succeeds
- Clips remain in "processing" state until a worker completes them
- Error logging includes clip_id and context for debugging

## Monitoring (TODO)

Future enhancements should include:

- Job processing metrics (Prometheus)
- Queue depth monitoring
- Failed job alerting
- Processing time tracking
- Worker health checks

## Testing

Tests are provided for:
- Service initialization
- Error handling with nil Redis client
- Handler integration with job service

Run tests with:
```bash
go test ./internal/services/clip_extraction_job_service_test.go
go test ./internal/handlers/stream_handler_test.go
```

## Future Improvements

1. **Worker Daemon**
   - Create `cmd/worker/main.go` for background processing
   - Deploy as separate service/container
   - Scale horizontally for high throughput

2. **Job Priority**
   - Add priority field to jobs
   - Use Redis sorted sets for priority queue
   - Process premium user clips first

3. **Progress Tracking**
   - Store FFmpeg progress in Redis
   - Expose progress via API endpoint
   - Real-time updates via WebSocket

4. **Monitoring & Alerting**
   - Prometheus metrics for job throughput
   - Alert on queue depth thresholds
   - Sentry error tracking for failures

5. **Advanced Features**
   - Job deduplication
   - Batch processing
   - Automatic quality selection
   - Thumbnail extraction
   - Preview generation
