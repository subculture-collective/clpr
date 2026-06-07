---
title: Backend HLS Video Streaming Implementation Guide
summary: This document outlines the backend requirements for supporting HLS (HTTP Live Streaming) video streaming for the Theatre Mode player.
tags: ["archive", "implementation", "summary"]
area: docs
status: archived
owner: team-core
version: "1.0"
last_reviewed: 2026-01-29
---

# Backend HLS Video Streaming Implementation Guide

## Overview

This document outlines the backend requirements for supporting HLS (HTTP Live Streaming) video streaming for the Theatre Mode player.

## Architecture

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Original   │ --> │   Encoding   │ --> │  HLS Master  │
│   Video      │     │   Pipeline   │     │  Playlist    │
│   Upload     │     │   (FFmpeg)   │     │  Generation  │
└──────────────┘     └──────────────┘     └──────────────┘
                                                  │
                            ┌─────────────────────┼─────────────────────┐
                            │                     │                     │
                     ┌──────▼─────┐      ┌───────▼──────┐      ┌──────▼─────┐
                     │  480p      │      │  720p        │      │  1080p     │
                     │  Variant   │      │  Variant     │      │  Variant   │
                     │  Playlist  │      │  Playlist    │      │  Playlist  │
                     └────────────┘      └──────────────┘      └────────────┘
```

## Required Endpoints

### 1. HLS Master Playlist

**Endpoint:** `GET /api/video/:clipId/master.m3u8`

**Description:** Serves the master playlist that lists all available quality variants.

**Response:** Content-Type: `application/vnd.apple.mpegurl`

```m3u8
#EXTM3U
#EXT-X-VERSION:3
#EXT-X-STREAM-INF:BANDWIDTH=25000000,RESOLUTION=3840x2160,CODECS="hvc1.1.6.L153.B0"
/api/video/:clipId/4k.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=15000000,RESOLUTION=2560x1440,CODECS="avc1.640028"
/api/video/:clipId/2k.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=10000000,RESOLUTION=1920x1080,CODECS="avc1.640028"
/api/video/:clipId/1080p.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=5000000,RESOLUTION=1280x720,CODECS="avc1.4d401f"
/api/video/:clipId/720p.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=2000000,RESOLUTION=854x480,CODECS="avc1.42001e"
/api/video/:clipId/480p.m3u8
```

### 2. Quality Variant Playlists

**Endpoint:** `GET /api/video/:clipId/:quality.m3u8`

**Description:** Serves the playlist for a specific quality variant.

**Response:** Content-Type: `application/vnd.apple.mpegurl`

```m3u8
#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:10
#EXT-X-MEDIA-SEQUENCE:0
#EXTINF:10.0,
/api/video/:clipId/:quality/segment-0.ts
#EXTINF:10.0,
/api/video/:clipId/:quality/segment-1.ts
#EXTINF:10.0,
/api/video/:clipId/:quality/segment-2.ts
#EXT-X-ENDLIST
```

### 3. Video Segments

**Endpoint:** `GET /api/video/:clipId/:quality/segment-:number.ts`

**Description:** Serves individual video segments.

**Response:** Content-Type: `video/MP2T`

Binary video segment data.

### 4. Clip Metadata (Extended)

**Endpoint:** `GET /api/clips/:clipId`

**Description:** Existing endpoint extended to include HLS availability.

**Response:** Content-Type: `application/json`

```json
{
  "id": "clip-123",
  "title": "Amazing Gaming Moment",
  "embed_url": "https://clips.twitch.tv/embed?clip=...",
  "twitch_clip_url": "https://clips.twitch.tv/...",
  "hls_available": true,
  "hls_url": "/api/video/clip-123/master.m3u8",
  "hls_qualities": ["480p", "720p", "1080p", "2K", "4K"],
  "hls_processing_status": "completed",
  ...
}
```

## Database Schema Extensions

### clips table

Add new columns to track HLS processing:

```sql
ALTER TABLE clips ADD COLUMN hls_available BOOLEAN DEFAULT FALSE;
ALTER TABLE clips ADD COLUMN hls_master_url TEXT;
ALTER TABLE clips ADD COLUMN hls_processing_status VARCHAR(20) DEFAULT 'pending';
ALTER TABLE clips ADD COLUMN hls_qualities JSONB;
ALTER TABLE clips ADD COLUMN hls_processed_at TIMESTAMP;
```

**Processing statuses:**
- `pending` - Not yet processed
- `processing` - Currently encoding
- `completed` - All variants ready
- `failed` - Encoding failed
- `partial` - Some variants available

### video_variants table (new)

Track individual quality variants:

```sql
CREATE TABLE video_variants (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  clip_id UUID NOT NULL REFERENCES clips(id) ON DELETE CASCADE,
  quality VARCHAR(10) NOT NULL, -- '480p', '720p', '1080p', '2K', '4K'
  width INT NOT NULL,
  height INT NOT NULL,
  bitrate INT NOT NULL, -- in bps
  codec VARCHAR(50) NOT NULL,
  file_size BIGINT, -- in bytes
  duration FLOAT, -- in seconds
  segment_count INT,
  playlist_url TEXT NOT NULL,
  storage_path TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  UNIQUE(clip_id, quality)
);

CREATE INDEX idx_video_variants_clip_id ON video_variants(clip_id);
CREATE INDEX idx_video_variants_quality ON video_variants(quality);
```

## Video Encoding Pipeline

### FFmpeg Encoding Scripts

#### 1. Generate All Quality Variants

```bash
#!/bin/bash
# encode_clip.sh - Generate HLS variants for a clip

INPUT_FILE=$1
OUTPUT_DIR=$2
CLIP_ID=$3

# Create output directory
mkdir -p "$OUTPUT_DIR"

# 480p variant
ffmpeg -i "$INPUT_FILE" \
  -c:v libx264 -preset medium -profile:v main -level 3.1 \
  -b:v 2M -maxrate 2.2M -bufsize 4M \
  -s 854x480 -r 30 \
  -c:a aac -b:a 128k -ar 44100 \
  -f hls -hls_time 10 -hls_list_size 0 \
  -hls_segment_filename "$OUTPUT_DIR/480p/segment-%03d.ts" \
  "$OUTPUT_DIR/480p/playlist.m3u8"

# 720p variant
ffmpeg -i "$INPUT_FILE" \
  -c:v libx264 -preset medium -profile:v high -level 4.0 \
  -b:v 5M -maxrate 5.5M -bufsize 10M \
  -s 1280x720 -r 30 \
  -c:a aac -b:a 192k -ar 48000 \
  -f hls -hls_time 10 -hls_list_size 0 \
  -hls_segment_filename "$OUTPUT_DIR/720p/segment-%03d.ts" \
  "$OUTPUT_DIR/720p/playlist.m3u8"

# 1080p variant
ffmpeg -i "$INPUT_FILE" \
  -c:v libx264 -preset medium -profile:v high -level 4.2 \
  -b:v 10M -maxrate 11M -bufsize 20M \
  -s 1920x1080 -r 30 \
  -c:a aac -b:a 256k -ar 48000 \
  -f hls -hls_time 10 -hls_list_size 0 \
  -hls_segment_filename "$OUTPUT_DIR/1080p/segment-%03d.ts" \
  "$OUTPUT_DIR/1080p/playlist.m3u8"

# 2K variant (only if source is high enough resolution)
ffmpeg -i "$INPUT_FILE" \
  -c:v libx264 -preset slow -profile:v high -level 5.0 \
  -b:v 15M -maxrate 16.5M -bufsize 30M \
  -s 2560x1440 -r 30 \
  -c:a aac -b:a 256k -ar 48000 \
  -f hls -hls_time 10 -hls_list_size 0 \
  -hls_segment_filename "$OUTPUT_DIR/2k/segment-%03d.ts" \
  "$OUTPUT_DIR/2k/playlist.m3u8"

# 4K variant (only if source is 4K)
ffmpeg -i "$INPUT_FILE" \
  -c:v libx265 -preset slow -profile:v main \
  -b:v 25M -maxrate 27.5M -bufsize 50M \
  -s 3840x2160 -r 30 \
  -c:a aac -b:a 320k -ar 48000 \
  -f hls -hls_time 10 -hls_list_size 0 \
  -hls_segment_filename "$OUTPUT_DIR/4k/segment-%03d.ts" \
  "$OUTPUT_DIR/4k/playlist.m3u8"
```

#### 2. Generate Master Playlist

```bash
#!/bin/bash
# generate_master.sh - Create HLS master playlist

OUTPUT_DIR=$1
CLIP_ID=$2

cat > "$OUTPUT_DIR/master.m3u8" << EOF
#EXTM3U
#EXT-X-VERSION:3
EOF

# Add 480p variant
if [ -f "$OUTPUT_DIR/480p/playlist.m3u8" ]; then
  echo "#EXT-X-STREAM-INF:BANDWIDTH=2000000,RESOLUTION=854x480,CODECS=\"avc1.42001e,mp4a.40.2\"" >> "$OUTPUT_DIR/master.m3u8"
  echo "/api/video/$CLIP_ID/480p.m3u8" >> "$OUTPUT_DIR/master.m3u8"
fi

# Add 720p variant
if [ -f "$OUTPUT_DIR/720p/playlist.m3u8" ]; then
  echo "#EXT-X-STREAM-INF:BANDWIDTH=5000000,RESOLUTION=1280x720,CODECS=\"avc1.4d401f,mp4a.40.2\"" >> "$OUTPUT_DIR/master.m3u8"
  echo "/api/video/$CLIP_ID/720p.m3u8" >> "$OUTPUT_DIR/master.m3u8"
fi

# Add 1080p variant
if [ -f "$OUTPUT_DIR/1080p/playlist.m3u8" ]; then
  echo "#EXT-X-STREAM-INF:BANDWIDTH=10000000,RESOLUTION=1920x1080,CODECS=\"avc1.640028,mp4a.40.2\"" >> "$OUTPUT_DIR/master.m3u8"
  echo "/api/video/$CLIP_ID/1080p.m3u8" >> "$OUTPUT_DIR/master.m3u8"
fi

# Add 2K variant
if [ -f "$OUTPUT_DIR/2k/playlist.m3u8" ]; then
  echo "#EXT-X-STREAM-INF:BANDWIDTH=15000000,RESOLUTION=2560x1440,CODECS=\"avc1.640028,mp4a.40.2\"" >> "$OUTPUT_DIR/master.m3u8"
  echo "/api/video/$CLIP_ID/2k.m3u8" >> "$OUTPUT_DIR/master.m3u8"
fi

# Add 4K variant
if [ -f "$OUTPUT_DIR/4k/playlist.m3u8" ]; then
  echo "#EXT-X-STREAM-INF:BANDWIDTH=25000000,RESOLUTION=3840x2160,CODECS=\"hvc1.1.6.L153.B0,mp4a.40.2\"" >> "$OUTPUT_DIR/master.m3u8"
  echo "/api/video/$CLIP_ID/4k.m3u8" >> "$OUTPUT_DIR/master.m3u8"
fi
```

## Go Backend Implementation (Example)

### Handler for HLS Master Playlist

```go
// internal/handlers/video_streaming_handler.go
package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "git.subcult.tv/subculture-collective/clpr/internal/repository"
)

type VideoStreamingHandler struct {
    clipRepo *repository.ClipRepository
}

func NewVideoStreamingHandler(clipRepo *repository.ClipRepository) *VideoStreamingHandler {
    return &VideoStreamingHandler{
        clipRepo: clipRepo,
    }
}

// GetHLSMasterPlaylist serves the HLS master playlist for a clip
// GET /api/video/:clipId/master.m3u8
func (h *VideoStreamingHandler) GetHLSMasterPlaylist(c *gin.Context) {
    clipID := c.Param("clipId")
    
    // Get clip from database
    clip, err := h.clipRepo.GetByID(c.Request.Context(), clipID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Clip not found"})
        return
    }
    
    // Check if HLS is available
    if !clip.HLSAvailable {
        c.JSON(http.StatusNotFound, gin.H{"error": "HLS not available for this clip"})
        return
    }
    
    // Serve the master playlist file
    c.Header("Content-Type", "application/vnd.apple.mpegurl")
    c.Header("Cache-Control", "public, max-age=3600")
    c.File(clip.HLSMasterPath)
}

// GetHLSVariantPlaylist serves a specific quality variant playlist
// GET /api/video/:clipId/:quality.m3u8
func (h *VideoStreamingHandler) GetHLSVariantPlaylist(c *gin.Context) {
    clipID := c.Param("clipId")
    quality := c.Param("quality")
    
    // Validate quality
    validQualities := map[string]bool{
        "480p": true, "720p": true, "1080p": true, "2k": true, "4k": true,
    }
    if !validQualities[quality] {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quality"})
        return
    }
    
    // Get variant from database
    variant, err := h.clipRepo.GetVariant(c.Request.Context(), clipID, quality)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Variant not found"})
        return
    }
    
    // Serve the variant playlist file
    c.Header("Content-Type", "application/vnd.apple.mpegurl")
    c.Header("Cache-Control", "public, max-age=3600")
    c.File(variant.PlaylistPath)
}

// GetHLSSegment serves a video segment
// GET /api/video/:clipId/:quality/segment-:number.ts
func (h *VideoStreamingHandler) GetHLSSegment(c *gin.Context) {
    clipID := c.Param("clipId")
    quality := c.Param("quality")
    segmentNum := c.Param("number")
    
    // Get segment path from database or construct it
    segmentPath, err := h.clipRepo.GetSegmentPath(c.Request.Context(), clipID, quality, segmentNum)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Segment not found"})
        return
    }
    
    // Serve the segment file
    c.Header("Content-Type", "video/MP2T")
    c.Header("Cache-Control", "public, max-age=31536000") // Cache segments for 1 year
    c.File(segmentPath)
}
```

### Register Routes

```go
// cmd/api/main.go

// Video streaming routes
videoHandler := handlers.NewVideoStreamingHandler(clipRepo)
api.GET("/video/:clipId/master.m3u8", videoHandler.GetHLSMasterPlaylist)
api.GET("/video/:clipId/:quality.m3u8", videoHandler.GetHLSVariantPlaylist)
api.GET("/video/:clipId/:quality/segment-:number.ts", videoHandler.GetHLSSegment)
```

## Storage Considerations

### Directory Structure

```
/var/clpr/video/
├── [clip-id-1]/
│   ├── master.m3u8
│   ├── 480p/
│   │   ├── playlist.m3u8
│   │   ├── segment-000.ts
│   │   ├── segment-001.ts
│   │   └── ...
│   ├── 720p/
│   ├── 1080p/
│   ├── 2k/
│   └── 4k/
└── [clip-id-2]/
    └── ...
```

### CDN Integration

For production, serve HLS content through a CDN:

1. Upload encoded files to S3/CloudFront
2. Update HLS URLs to point to CDN
3. Configure appropriate caching headers
4. Enable byte-range requests

### Storage Requirements

Approximate storage per minute of video:

| Quality | Storage/min | Storage/hour |
|---------|-------------|--------------|
| 480p | ~15 MB | ~900 MB |
| 720p | ~37.5 MB | ~2.25 GB |
| 1080p | ~75 MB | ~4.5 GB |
| 2K | ~112.5 MB | ~6.75 GB |
| 4K | ~187.5 MB | ~11.25 GB |

## Processing Queue

Implement a background job queue for video encoding:

```go
type VideoEncodingJob struct {
    ClipID      string
    SourceURL   string
    Qualities   []string
    Priority    int
    CreatedAt   time.Time
}

func EnqueueVideoEncoding(clipID string, sourceURL string) error {
    job := VideoEncodingJob{
        ClipID:    clipID,
        SourceURL: sourceURL,
        Qualities: []string{"480p", "720p", "1080p"},
        Priority:  5,
        CreatedAt: time.Now(),
    }
    
    return jobQueue.Enqueue(job)
}
```

## CORS Configuration

Ensure proper CORS headers for HLS streaming:

```go
func SetupCORS(r *gin.Engine) {
    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"https://yourdomain.com"},
        AllowMethods:     []string{"GET", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Range"},
        ExposeHeaders:    []string{"Content-Length", "Content-Range"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }))
}
```

## Monitoring and Analytics

Track HLS streaming metrics:

```sql
CREATE TABLE video_streaming_analytics (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  clip_id UUID NOT NULL REFERENCES clips(id),
  user_id UUID REFERENCES users(id),
  quality VARCHAR(10),
  bandwidth_mbps FLOAT,
  buffer_health INT,
  session_duration INT, -- seconds
  quality_switches INT,
  buffering_events INT,
  created_at TIMESTAMP DEFAULT NOW()
);
```

## Security Considerations

1. **Rate Limiting**: Limit HLS requests per IP/user
2. **Token Authentication**: Optional signed URLs for premium content
3. **Bandwidth Monitoring**: Track and limit bandwidth per user
4. **DDoS Protection**: Use CDN with DDoS protection
5. **Content Validation**: Verify clip ownership before serving

## Testing

```bash
# Test master playlist
curl -v http://localhost:8080/api/video/test-clip-123/master.m3u8

# Test variant playlist
curl -v http://localhost:8080/api/video/test-clip-123/720p.m3u8

# Test segment
curl -v http://localhost:8080/api/video/test-clip-123/720p/segment-000.ts

# Test with HLS.js
# Use the frontend theatre mode player to test end-to-end
```

## Migration Strategy

1. **Phase 1**: Implement endpoints and encoding pipeline
2. **Phase 2**: Encode new clips automatically
3. **Phase 3**: Backfill existing popular clips
4. **Phase 4**: Migrate all clips gradually
5. **Phase 5**: Deprecate direct Twitch embed dependency

## Performance Targets

- **Encoding Time**: < 2x real-time (10 min video → < 20 min to encode)
- **First Byte**: < 500ms for master playlist
- **Segment Loading**: < 100ms per segment
- **Quality Switch**: < 2 seconds
- **Cache Hit Rate**: > 95% for segments
