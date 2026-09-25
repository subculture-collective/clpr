---
title: "Mirror Hosting"
summary: "This guide explains how to configure and use the mirror hosting feature for global content distribution."
tags: ["backend"]
area: "backend"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Mirror Hosting Guide

This guide explains how to configure and use the mirror hosting feature for global content distribution.

## Overview

Mirror hosting replicates popular clips to multiple geographic regions to improve:
- **Global Performance**: Reduced latency for users worldwide
- **Reliability**: Automatic failover if primary content is unavailable
- **Scalability**: Distributed load across regions

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Primary Storage                          │
│                    (Twitch/Origin)                           │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│              Mirror Replication Service                      │
│  - Identifies popular clips (view_count/vote_score)         │
│  - Replicates to 2-3 regions                                │
│  - Manages TTL and cleanup                                   │
└─────────────────────────────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        ▼                   ▼                   ▼
  ┌──────────┐        ┌──────────┐        ┌──────────┐
  │ US East  │        │ EU West  │        │ AP SE    │
  │ Mirror   │        │ Mirror   │        │ Mirror   │
  └──────────┘        └──────────┘        └──────────┘
```

## Configuration

### Environment Variables

```bash
# Enable mirror hosting
MIRROR_ENABLED=true

# Regions to mirror clips to (comma-separated)
MIRROR_REGIONS=us-east-1,eu-west-1,ap-southeast-1

# Minimum view count to trigger mirroring
MIRROR_REPLICATION_THRESHOLD=1000

# Mirror TTL in days
MIRROR_TTL_DAYS=7

# Maximum mirrors per clip
MIRROR_MAX_PER_CLIP=3

# Sync interval (how often to check for popular clips)
MIRROR_SYNC_INTERVAL_MINUTES=60

# Cleanup interval (how often to remove expired mirrors)
MIRROR_CLEANUP_INTERVAL_MINUTES=1440

# Minimum mirror hit rate percentage (alert threshold)
MIRROR_MIN_HIT_RATE=60.0
```

### Storage Providers

The system supports multiple storage providers per region:

- **AWS S3**: US East, US West
- **Cloudflare R2**: EU regions
- **Bunny.net**: Asia-Pacific regions

## How It Works

### 1. Clip Selection

The background job identifies clips for mirroring based on:
- View count >= threshold
- Vote score >= threshold
- Not removed/DMCA'd
- Fewer than max mirrors already exist

### 2. Replication Process

For each selected clip:
1. Create mirror record in database (status: `pending`)
2. Upload clip to regional storage
3. Update mirror status to `active`
4. Set expiration date based on TTL

### 3. Failover Logic

When serving a clip:
1. Try to find mirror in user's region
2. If unavailable, try any active mirror
3. Record access for metrics
4. Fallback to primary if no mirrors available

### 4. Cleanup

Expired mirrors are automatically deleted:
- Daily cleanup job runs
- Removes mirrors past their TTL
- Updates clip mirror counts

## Monitoring

### Metrics

The system tracks:
- **Mirror Hit Rate**: Percentage of requests served from mirrors
- **Access Count**: Number of times each mirror was accessed
- **Failover Count**: Number of times primary was unavailable
- **Storage Usage**: Total size of mirrored content

### Alerts

Alerts trigger when:
- Mirror hit rate falls below threshold (default: 60%)
- Region becomes unhealthy
- Replication fails consistently

### Prometheus Metrics

```promql
# Mirror hit rate
mirror_hit_rate{region="us-east-1"}

# Total mirrors
total_active_mirrors

# Failover events
mirror_failover_total{region="us-east-1"}
```

## API Usage

### Get Mirror URL

```go
mirrorURL, found, err := mirrorService.GetMirrorURL(ctx, clipID, userRegion)
if err != nil {
    // Handle error
}

if found {
    // Use mirror URL
} else {
    // Use primary URL
}
```

### Check Mirror Status

```go
mirrors, err := mirrorRepo.ListByClip(ctx, clipID)
for _, mirror := range mirrors {
    log.Printf("Region: %s, Status: %s, Accessed: %d times",
        mirror.Region, mirror.Status, mirror.AccessCount)
}
```

### Get Hit Rate

```go
hitRate, err := mirrorService.GetMirrorHitRate(ctx)
log.Printf("Current mirror hit rate: %.2f%%", hitRate)
```

## Database Schema

### clip_mirrors Table

```sql
CREATE TABLE clip_mirrors (
    id UUID PRIMARY KEY,
    clip_id UUID REFERENCES clips(id),
    region VARCHAR(50) NOT NULL,
    mirror_url TEXT NOT NULL,
    status VARCHAR(20) NOT NULL, -- pending, active, failed, expired
    storage_provider VARCHAR(50) NOT NULL,
    size_bytes BIGINT,
    created_at TIMESTAMP NOT NULL,
    last_accessed_at TIMESTAMP,
    access_count INT DEFAULT 0,
    expires_at TIMESTAMP,
    failure_reason TEXT,
    UNIQUE(clip_id, region)
);
```

### mirror_metrics Table

```sql
CREATE TABLE mirror_metrics (
    id UUID PRIMARY KEY,
    clip_id UUID REFERENCES clips(id),
    region VARCHAR(50) NOT NULL,
    metric_type VARCHAR(50) NOT NULL, -- access, failover, bandwidth, cost
    metric_value FLOAT NOT NULL,
    recorded_at TIMESTAMP NOT NULL,
    metadata JSONB
);
```

## Troubleshooting

### Low Hit Rate

If mirror hit rate is below target:
1. Check if mirrors are being created
2. Verify region health checks
3. Review replication threshold (may be too high)
4. Check geo-routing logic

### High Costs

If storage costs are high:
1. Reduce mirror TTL
2. Increase replication threshold
3. Reduce max mirrors per clip
4. Review most accessed clips

### Replication Failures

If mirrors fail to replicate:
1. Check storage provider credentials
2. Verify network connectivity to regions
3. Review storage quota limits
4. Check clip file accessibility

## Best Practices

1. **Start Conservative**: Begin with higher thresholds and adjust based on metrics
2. **Monitor Costs**: Track storage and bandwidth costs closely
3. **Regional Strategy**: Prioritize regions with most users
4. **TTL Optimization**: Balance performance vs. cost with appropriate TTL
5. **Health Checks**: Ensure health monitoring is active for failover

## Security Considerations

- Mirror URLs should use HTTPS
- Implement signed URLs for premium content
- Regular security audits of storage providers
- Monitor for unauthorized access

## Performance Optimization

- Use CDN in front of mirrors
- Implement edge caching
- Optimize video encoding settings
- Consider adaptive bitrate streaming

## Future Enhancements

- [ ] Predictive mirroring based on trending content
- [ ] Adaptive TTL based on popularity decay
- [ ] Multi-tier storage (hot/warm/cold)
- [ ] Real-time replication for live clips
- [ ] Intelligent geo-routing based on user patterns
