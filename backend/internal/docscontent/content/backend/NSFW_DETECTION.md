---
title: "NSFW DETECTION"
summary: "The NSFW Image Detection system provides automated detection and moderation of NSFW (Not Safe For Work) content in images and thumbnails. It integrates with external AI services for accurate detection"
tags: ["docs"]
area: "docs"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# NSFW Image Detection API

## Overview

The NSFW Image Detection system provides automated detection and moderation of NSFW (Not Safe For Work) content in images and thumbnails. It integrates with external AI services for accurate detection and automatically flags inappropriate content to the moderation queue.

## Features

- **Real-time Detection**: Scan images for NSFW content with <200ms p95 latency
- **Batch Processing**: Process multiple images in a single request
- **Auto-flagging**: Automatically add detected NSFW content to moderation queue
- **Configurable Thresholds**: Adjust sensitivity for different use cases
- **Metrics & Monitoring**: Comprehensive Prometheus metrics and health checks
- **Background Scanning**: Scan existing content library in batches

## Configuration

Add the following environment variables to your `.env` file:

```bash
# Enable/disable NSFW detection
NSFW_ENABLED=true

# API credentials for NSFW detection service
# Supported services: Sightengine, AWS Rekognition, Google Cloud Vision AI, Azure Content Moderator
NSFW_API_KEY=your_api_key_here
NSFW_API_URL=https://api.sightengine.com/1.0/check.json

# Detection threshold (0.0 to 1.0)
NSFW_THRESHOLD=0.80

# Scan thumbnails on upload
NSFW_SCAN_THUMBNAILS=true

# Auto-flag detected content to moderation queue
NSFW_AUTO_FLAG=true

# Performance settings
NSFW_MAX_LATENCY_MS=200
NSFW_TIMEOUT_SECONDS=5
```

## API Endpoints

All endpoints require admin authentication.

### 1. Detect Single Image

**POST** `/admin/nsfw/detect`

Performs NSFW detection on a single image URL.

#### Request Body

```json
{
  "image_url": "https://example.com/image.jpg",
  "content_type": "thumbnail",
  "content_id": "123e4567-e89b-12d3-a456-426614174000",
  "auto_flag": true
}
```

#### Parameters

- `image_url` (required, string): URL of the image to scan
- `content_type` (required, enum): Type of content - `clip`, `thumbnail`, `submission`, or `user`
- `content_id` (optional, UUID): ID of the content being scanned
- `auto_flag` (optional, boolean): Whether to auto-flag if NSFW (default: true)

#### Response

```json
{
  "success": true,
  "data": {
    "nsfw": false,
    "confidence_score": 0.05,
    "categories": {
      "nudity_raw": 0.05,
      "nudity_safe": 0.95,
      "nudity_partial": 0.02,
      "nudity_sexual": 0.01,
      "offensive": 0.03
    },
    "reason_codes": [],
    "latency_ms": 145
  }
}
```

### 2. Batch Detect

**POST** `/admin/nsfw/batch-detect`

Performs NSFW detection on multiple images in one request.

#### Request Body

```json
{
  "images": [
    {
      "image_url": "https://example.com/image1.jpg",
      "content_type": "thumbnail",
      "content_id": "123e4567-e89b-12d3-a456-426614174000"
    },
    {
      "image_url": "https://example.com/image2.jpg",
      "content_type": "clip"
    }
  ],
  "auto_flag": true
}
```

#### Parameters

- `images` (required, array): Array of image objects (max 50)
  - `image_url` (required, string): URL of the image
  - `content_type` (required, enum): Type of content
  - `content_id` (optional, UUID): ID of the content
- `auto_flag` (optional, boolean): Whether to auto-flag NSFW content

#### Response

```json
{
  "success": true,
  "data": [
    {
      "image_url": "https://example.com/image1.jpg",
      "nsfw": false,
      "confidence_score": 0.05,
      "categories": {...},
      "reason_codes": [],
      "latency_ms": 145,
      "flagged": false
    },
    {
      "image_url": "https://example.com/image2.jpg",
      "nsfw": true,
      "confidence_score": 0.92,
      "categories": {...},
      "reason_codes": ["nudity_explicit"],
      "latency_ms": 158,
      "flagged": true
    }
  ],
  "meta": {
    "total_processed": 2,
    "success_count": 2,
    "nsfw_count": 1,
    "avg_latency_ms": 151
  }
}
```

### 3. Get Metrics

**GET** `/admin/nsfw/metrics`

Retrieves NSFW detection metrics for a date range.

#### Query Parameters

- `start_date` (optional, string): Start date in YYYY-MM-DD format (default: 30 days ago)
- `end_date` (optional, string): End date in YYYY-MM-DD format (default: today)

#### Response

```json
{
  "success": true,
  "data": {
    "total_detections": 1250,
    "detections_by_reason": {
      "nsfw_nudity_explicit": {
        "count": 45,
        "avg_confidence": 0.91
      },
      "nsfw_sexual_content": {
        "count": 23,
        "avg_confidence": 0.87
      },
      "nsfw_offensive_content": {
        "count": 12,
        "avg_confidence": 0.84
      }
    },
    "avg_review_time_minutes": 12.5,
    "start_date": "2024-12-01",
    "end_date": "2024-12-30"
  }
}
```

### 4. Health Check

**GET** `/admin/nsfw/health`

Checks the health status of the NSFW detection service.

#### Response

```json
{
  "success": true,
  "status": "healthy",
  "latency_ms": 142
}
```

### 5. Get Configuration

**GET** `/admin/nsfw/config`

Returns non-sensitive configuration details.

#### Response

```json
{
  "success": true,
  "data": {
    "enabled": true
  }
}
```

### 6. Scan Existing Clips

**POST** `/admin/nsfw/scan-clips`

Starts a background job to scan existing clip thumbnails.

#### Request Body

```json
{
  "limit": 100,
  "auto_flag": true
}
```

#### Parameters

- `limit` (optional, integer): Number of clips to scan (1-1000, default: 100)
- `auto_flag` (optional, boolean): Whether to auto-flag detected NSFW content

#### Response

```json
{
  "success": true,
  "job_id": "987e6543-e21b-98f7-b654-321098765432",
  "message": "Scan job started",
  "limit": 100
}
```

## Database Schema

### nsfw_detection_metrics

Tracks all NSFW detection results:

```sql
CREATE TABLE nsfw_detection_metrics (
    id UUID PRIMARY KEY,
    content_type VARCHAR(20) NOT NULL,
    content_id UUID NOT NULL,
    image_url TEXT NOT NULL,
    nsfw BOOLEAN NOT NULL,
    confidence_score DECIMAL(3,2) NOT NULL,
    categories JSONB,
    reason_codes TEXT[],
    latency_ms INT NOT NULL,
    flagged_to_queue BOOLEAN DEFAULT FALSE,
    detected_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW()
);
```

### nsfw_scan_jobs

Tracks background scanning jobs:

```sql
CREATE TABLE nsfw_scan_jobs (
    id UUID PRIMARY KEY,
    job_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    total_items INT DEFAULT 0,
    processed_items INT DEFAULT 0,
    nsfw_found INT DEFAULT 0,
    auto_flag BOOLEAN DEFAULT TRUE,
    error_message TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);
```

## Prometheus Metrics

The NSFW detector exposes the following Prometheus metrics:

- `nsfw_detection_total{result}`: Total number of detections (labels: safe, nsfw, error)
- `nsfw_detection_latency_ms`: Histogram of detection latency in milliseconds
- `nsfw_content_flagged_total{content_type}`: Total number of flagged content by type
- `nsfw_detection_errors_total{error_type}`: Total number of errors by type

## Integration with Moderation Queue

When NSFW content is detected and auto-flagging is enabled:

1. An entry is created in the `moderation_queue` table
2. The entry includes:
   - `content_type` and `content_id` for the flagged content
   - `reason`: One of `nsfw_nudity_explicit`, `nsfw_sexual_content`, `nsfw_offensive_content`
   - `priority`: Based on confidence score (0-100)
   - `auto_flagged`: Set to `true`
   - `confidence_score`: Detection confidence (0.0-1.0)
   - `nsfw_categories`: Detailed category scores as JSONB
3. Moderators can review flagged content in the moderation queue

## Error Handling

Common error responses:

### 400 Bad Request
```json
{
  "error": "Invalid request: image_url is required"
}
```

### 500 Internal Server Error
```json
{
  "error": "Failed to detect NSFW content: API request timeout"
}
```

### 503 Service Unavailable
```json
{
  "success": false,
  "status": "unhealthy",
  "error": "NSFW detection service unavailable",
  "latency_ms": 5000
}
```

## Best Practices

1. **Batch Processing**: Use batch detection for multiple images to reduce overhead
2. **Error Handling**: Implement retry logic for transient API failures
3. **Monitoring**: Set up alerts on `nsfw_detection_errors_total` and latency metrics
4. **Threshold Tuning**: Adjust `NSFW_THRESHOLD` based on your content policy
5. **Rate Limiting**: External API calls may be rate limited; plan accordingly

## Supported NSFW Detection Services

### Sightengine
- API URL: `https://api.sightengine.com/1.0/check.json`
- Pricing: Pay per API call
- Features: Nudity, violence, offensive content detection

### AWS Rekognition
- API URL: AWS SDK endpoint
- Pricing: Pay per image analyzed
- Features: Moderation labels, explicit content detection

### Google Cloud Vision AI
- API URL: `https://vision.googleapis.com/v1/images:annotate`
- Pricing: Pay per image
- Features: Safe search detection, explicit content

### Azure Content Moderator
- API URL: Azure Content Moderator endpoint
- Pricing: Pay per transaction
- Features: Adult content, racy content detection

## Performance Targets

- **Latency**: <200ms p95 for single image detection
- **Throughput**: 50 images per batch request
- **Accuracy**: >95% true positive rate at threshold 0.80
- **Availability**: 99.9% uptime

## Related Documentation

- Moderation Queue API
- Content Moderation Guidelines
- Metrics & Monitoring
