---
title: "CDN Integration"
summary: "This guide explains how to configure and use the CDN (Content Delivery Network) integration for optimized video delivery."
tags: ["backend"]
area: "backend"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# CDN Integration Guide

This guide explains how to configure and use the CDN (Content Delivery Network) integration for optimized video delivery.

## Overview

The CDN integration provides:
- **Global Performance**: Reduced latency through edge caching
- **Cost Optimization**: Efficient bandwidth usage
- **Video Streaming**: Optimized headers for video playback
- **Multi-Provider Support**: Cloudflare, Bunny.net, AWS CloudFront

## Supported CDN Providers

### Cloudflare

**Pros:**
- Global network with 200+ data centers
- Free tier available
- Built-in DDoS protection
- Easy integration

**Pricing:** ~$0.04-0.08/GB

### Bunny.net

**Pros:**
- Competitive pricing
- Storage + CDN combo
- Good performance
- Simple API

**Pricing:** ~$0.01-0.03/GB

### AWS CloudFront

**Pros:**
- Enterprise-grade reliability
- Integration with AWS services
- Advanced features
- Global coverage

**Pricing:** ~$0.085-0.20/GB

## Configuration

### Cloudflare Setup

```bash
# Enable CDN
CDN_ENABLED=true
CDN_PROVIDER=cloudflare

# Cloudflare credentials
CDN_CLOUDFLARE_ZONE_ID=your_zone_id
CDN_CLOUDFLARE_API_KEY=your_api_key

# Cache settings
CDN_CACHE_TTL=3600
CDN_MAX_COST_PER_GB=0.10
```

**Steps:**
1. Sign up for Cloudflare account
2. Add your domain
3. Get Zone ID from dashboard
4. Generate API token with cache purge permissions
5. Update environment variables

### Bunny.net Setup

```bash
# Enable CDN
CDN_ENABLED=true
CDN_PROVIDER=bunny

# Bunny credentials
CDN_BUNNY_API_KEY=your_api_key
CDN_BUNNY_STORAGE_ZONE=your_zone_name

# Cache settings
CDN_CACHE_TTL=3600
CDN_MAX_COST_PER_GB=0.10
```

**Steps:**
1. Sign up for Bunny account
2. Create a Pull Zone or Storage Zone
3. Get API key from account settings
4. Configure storage zone name
5. Update environment variables

### AWS CloudFront Setup

```bash
# Enable CDN
CDN_ENABLED=true
CDN_PROVIDER=aws-cloudfront

# AWS credentials
CDN_AWS_ACCESS_KEY_ID=your_access_key
CDN_AWS_SECRET_KEY=your_secret_key
CDN_AWS_REGION=us-east-1

# Cache settings
CDN_CACHE_TTL=3600
CDN_MAX_COST_PER_GB=0.10
```

**Steps:**
1. Create AWS account
2. Set up CloudFront distribution
3. Create IAM user with CloudFront permissions
4. Generate access keys
5. Update environment variables

## Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────┐
│        CDN Edge (Nearest POP)       │
│  - Cache video content              │
│  - Serve from edge if cached        │
│  - Fetch from origin if cache miss  │
└─────────────────────────────────────┘
       │
       ▼ (cache miss)
┌─────────────────────────────────────┐
│         Origin Server               │
│  - Mirror or Primary Storage        │
│  - Return video with cache headers  │
└─────────────────────────────────────┘
```

## Cache Strategy

### Video Content Caching

```go
// Cache headers for video content
headers := cdnService.GetCacheHeaders()

// Example headers:
// Cache-Control: public, max-age=3600
// CDN-Cache-Control: max-age=3600
// Vary: Accept-Encoding
```

### Cache TTL Recommendations

- **Popular clips (>10K views)**: 7 days
- **Recent clips (<24h old)**: 1 hour
- **Static assets**: 30 days
- **Thumbnails**: 7 days

### Cache Purging

Purge cache when:
- Clip is removed/DMCA'd
- Content is updated
- Scheduled refresh needed

```go
// Purge cache for a clip
err := cdnService.PurgeCache(ctx, clip)
```

## Monitoring

### Metrics Tracked

1. **Bandwidth**: Total GB transferred
2. **Requests**: Total number of requests
3. **Cache Hit Rate**: Percentage of cached responses
4. **Latency**: Average response time
5. **Cost**: Estimated monthly cost

### Cost Monitoring

```go
// Check if costs exceed threshold
exceeded, costPerGB, err := cdnService.CheckCostThreshold(ctx)
if exceeded {
    log.Printf("WARNING: Cost per GB ($%.4f) exceeds limit", costPerGB)
}
```

### Prometheus Metrics

```promql
# Cache hit rate
cdn_cache_hit_rate{provider="cloudflare"}

# Total bandwidth (GB)
cdn_bandwidth_total{provider="cloudflare"}

# Total requests
cdn_requests_total{provider="cloudflare"}

# Average latency
cdn_latency_avg{provider="cloudflare"}

# Total cost
cdn_cost_total{provider="cloudflare"}
```

## API Usage

### Generate CDN URL

```go
// Get CDN URL for a clip
cdnURL, err := cdnService.GetCDNURL(ctx, clip)
if err != nil {
    // Handle error
}

// Use CDN URL in response
clipResponse.CDNURL = cdnURL
```

### Collect Metrics

```go
// Collect metrics from CDN provider
err := cdnService.CollectMetrics(ctx)
```

### Get Performance Stats

```go
// Get cache hit rate for last 24h
startTime := time.Now().Add(-24 * time.Hour)
hitRate, err := cdnService.GetCacheHitRate(ctx, startTime)
log.Printf("Cache hit rate: %.2f%%", hitRate)

// Get costs for last 30 days
endTime := time.Now()
startTime = endTime.AddDate(0, 0, -30)
totalCost, err := cdnService.GetCostMetrics(ctx, startTime, endTime)
log.Printf("Total CDN cost: $%.2f", totalCost)
```

## Video Streaming Optimization

### Headers for Video Streaming

```go
// Optimized headers for video delivery
Cache-Control: public, max-age=3600
Accept-Ranges: bytes
Content-Type: video/mp4
X-Content-Type-Options: nosniff
```

### Range Request Support

CDN providers automatically support HTTP range requests for:
- Seeking in videos
- Partial content delivery
- Bandwidth optimization

### Adaptive Bitrate

For multiple quality tiers:
```
/clips/{id}/720p.mp4
/clips/{id}/1080p.mp4
/clips/{id}/source.mp4
```

## Performance Optimization

### 1. Cache Warming

Pre-populate CDN cache for popular content:
```go
// Identify popular clips
popularClips, err := clipRepo.GetPopularClips(ctx, limit)

// Request each through CDN to warm cache
for _, clip := range popularClips {
    cdnURL, _ := cdnService.GetCDNURL(ctx, clip)
    // Make HTTP request to warm cache
}
```

### 2. Geographic Optimization

Route users to nearest edge:
- Use CDN provider's geo-routing
- Implement custom geo-logic if needed
- Monitor latency by region

### 3. Compression

Enable compression at CDN level:
- Gzip for text assets
- Brotli for modern browsers
- No compression for videos (already compressed)

## Cost Optimization

### 1. Smart Caching

- Higher TTL for popular content
- Lower TTL for new content
- Purge less-viewed content

### 2. Bandwidth Management

```go
// Monitor bandwidth usage
if bandwidth > threshold {
    // Increase cache TTL
    // Reduce quality for free users
    // Implement rate limiting
}
```

### 3. Regional Strategy

- Use cheaper providers in low-traffic regions
- Premium providers in high-traffic regions
- Monitor cost per region

## Troubleshooting

### High Cache Miss Rate

**Symptoms:**
- Cache hit rate < 60%
- High origin bandwidth
- Slow performance

**Solutions:**
1. Increase cache TTL
2. Pre-warm cache for popular content
3. Review cache key strategy
4. Check purge frequency

### High Costs

**Symptoms:**
- Cost per GB exceeds threshold
- Unexpected bandwidth usage

**Solutions:**
1. Reduce cache TTL for unpopular content
2. Implement access controls
3. Review bandwidth by region
4. Consider switching providers

### Slow Performance

**Symptoms:**
- High latency
- Buffering issues
- User complaints

**Solutions:**
1. Check CDN provider status
2. Review edge location coverage
3. Optimize video encoding
4. Enable compression

## Security

### 1. Signed URLs

For premium/private content:
```go
// Generate signed URL with expiration
signedURL := cdnProvider.GenerateSignedURL(clip, expiresAt)
```

### 2. DDoS Protection

- Use Cloudflare's DDoS protection
- Implement rate limiting
- Monitor for abuse

### 3. Access Control

- Validate tokens before serving
- Log access for auditing
- Implement geo-blocking if needed

## Database Schema

### cdn_configurations Table

```sql
CREATE TABLE cdn_configurations (
    id UUID PRIMARY KEY,
    provider VARCHAR(50) NOT NULL,
    region VARCHAR(50),
    is_active BOOLEAN DEFAULT true,
    priority INT DEFAULT 0,
    config JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    UNIQUE(provider, region)
);
```

### cdn_metrics Table

```sql
CREATE TABLE cdn_metrics (
    id UUID PRIMARY KEY,
    provider VARCHAR(50) NOT NULL,
    region VARCHAR(50),
    metric_type VARCHAR(50) NOT NULL,
    metric_value FLOAT NOT NULL,
    recorded_at TIMESTAMP NOT NULL,
    metadata JSONB
);
```

## Best Practices

1. **Start Small**: Test with one provider first
2. **Monitor Closely**: Track costs and performance daily
3. **Cache Strategically**: Balance performance vs. cost
4. **Regular Reviews**: Monthly cost and performance reviews
5. **Failover Plan**: Have backup CDN or origin serving

## Migration Guide

### From Direct Serving to CDN

1. **Phase 1: Setup**
   - Configure CDN provider
   - Test with sample content
   - Monitor metrics

2. **Phase 2: Gradual Rollout**
   - Enable for 10% of users
   - Monitor performance and costs
   - Increase gradually

3. **Phase 3: Full Migration**
   - Enable for all users
   - Keep origin as fallback
   - Monitor and optimize

### Switching CDN Providers

1. Configure new provider
2. Test thoroughly
3. Update DNS/URLs gradually
4. Monitor both providers
5. Complete migration
6. Decommission old provider

## Future Enhancements

- [ ] Multi-CDN strategy (active-active)
- [ ] Automatic provider selection based on cost/performance
- [ ] Real-time cost alerts
- [ ] Advanced cache warming strategies
- [ ] CDN provider benchmarking
