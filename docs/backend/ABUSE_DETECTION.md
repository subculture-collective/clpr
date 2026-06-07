---
title: "ABUSE DETECTION"
summary: "The Abuse Pattern Detection system implements real-time anomaly detection for spam, vote manipulation, and coordinated abuse as specified in Roadmap 5.0 Phase 3.3."
tags: ["docs"]
area: "docs"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Abuse Pattern Detection System

## Overview

The Abuse Pattern Detection system implements real-time anomaly detection for spam, vote manipulation, and coordinated abuse as specified in Roadmap 5.0 Phase 3.3.

## Architecture

### Components

1. **AbuseFeatureExtractor** (`internal/services/abuse_feature_extractor.go`)
   - Extracts features for anomaly detection from user actions
   - Tracks velocity patterns (votes, follows, submissions per time window)
   - Analyzes IP/User-Agent signals for account coordination
   - Detects graph patterns (circular follows, coordinated voting)
   - Calculates behavioral metrics (timing entropy, pattern diversity)

2. **AnomalyScorer** (`internal/services/anomaly_scorer.go`)
   - Scores actions using extracted features
   - Implements statistical anomaly detection
   - Calculates overall anomaly scores (0.0-1.0)
   - Determines confidence levels and severity
   - Tracks metrics for monitoring

3. **AbuseAutoFlagger** (`internal/services/abuse_auto_flagger.go`)
   - Automatically flags suspicious content to moderation queue
   - Creates entries with reason codes and confidence scores
   - Calculates priority based on anomaly score
   - Links flagged items to moderation system

4. **AbuseIntegrationService** (`internal/services/abuse_integration_service.go`)
   - Integrates anomaly detection into vote, follow, and submission handlers
   - Performs real-time scoring on key actions
   - Auto-flags items that exceed thresholds

5. **AbuseAnalyticsHandler** (`internal/handlers/abuse_analytics_handler.go`)
   - Provides admin endpoints for metrics and analytics
   - Tracks false positive rate (FPR)
   - Monitors system health

## Features

### Velocity Tracking

Monitors action rates across multiple time windows:
- **5-minute bursts**: Detects rapid, automated behavior
- **1-hour velocity**: Identifies sustained abusive patterns
- **24-hour activity**: Provides broader context

**Thresholds:**
- Votes: 2 per 5 minutes, 10 per hour
- Follows: 3 per 5 minutes, 15 per hour
- Submissions: 3 per hour, 8 per day

### IP/User-Agent Analysis

Detects account coordination patterns:
- **Shared IP detection**: Flags multiple accounts from same IP (threshold: 5+)
- **User-Agent correlation**: Identifies bot farms using identical browsers
- **IP hopping**: Detects frequent IP address changes (proxy/VPN abuse)

### Graph Pattern Detection

Identifies coordinated abuse networks:
- **Coordinated voting**: Detects multiple accounts voting on same content from shared IPs
- **Circular follows**: Identifies artificial follow networks (A→B→C→A)
- **Burst patterns**: Exponential scoring for activity spikes

### Behavioral Features

Analyzes user behavior patterns:
- **Vote pattern diversity**: Measures balance between upvotes/downvotes (bots often vote monotonously)
- **Timing entropy**: Detects machine-like regular intervals between actions
- **Account age**: New accounts receive higher scrutiny

### Trust Score Integration

Leverages existing trust scores:
- High trust (80+): 0.0 anomaly contribution
- Medium trust (50-79): 0.3 anomaly contribution
- Low trust (30-49): 0.6 anomaly contribution
- Very low trust (<30): 0.9 anomaly contribution

## Scoring Algorithm

### Overall Score Calculation

```
overall_score = (velocity * 0.25) + 
                (ip_ua * 0.20) + 
                (graph_patterns * 0.25) + 
                (behavioral * 0.15) + 
                (trust_score * 0.15)
```

### Thresholds

- **Low anomaly**: 0.30-0.49
- **Medium anomaly**: 0.50-0.69
- **High anomaly**: 0.70-0.84
- **Critical anomaly**: 0.85-1.00

### Auto-Flag Threshold

Items are auto-flagged when:
- `overall_score >= 0.75` AND
- `confidence >= 0.60`

This conservative threshold helps maintain FPR < 2%.

### Confidence Calculation

Confidence is based on data availability:
- Up to five tracked features are considered; each present feature adds 0.2 to confidence (maximum 1.0 from features)
- Older accounts (>30 days) add a 0.1 bonus before capping
- The final confidence score is capped at 1.0

## Integration Points

### Vote Handler

File: `internal/handlers/clip_handler.go:VoteOnClip`

```go
// After successful vote, perform anomaly check
if abuseIntegrationService != nil {
    ip := c.ClientIP()
    userAgent := c.GetHeader("User-Agent")
    abuseIntegrationService.CheckVoteAction(ctx, userID, clipID, voteType, ip, userAgent)
}
```

### Follow Handler

File: `internal/handlers/user_handler.go:FollowUser`

```go
// After successful follow, perform anomaly check
if abuseIntegrationService != nil {
    ip := c.ClientIP()
    userAgent := c.GetHeader("User-Agent")
    abuseIntegrationService.CheckFollowAction(ctx, followerID, followingID, ip, userAgent)
}
```

### Submission Handler

File: `internal/handlers/submission_handler.go:SubmitClip`

Integration already exists via `SubmissionAbuseDetector`. Enhanced with:

```go
// After initial abuse check, perform anomaly scoring
if abuseIntegrationService != nil {
    abuseIntegrationService.CheckSubmissionAction(ctx, userID, submissionID, ip, userAgent)
}
```

## Moderation Queue Integration

Auto-flagged items create entries in the `moderation_queue` table:

```sql
INSERT INTO moderation_queue (
    content_type,      -- 'user', 'submission', 'comment'
    content_id,        -- UUID of flagged content
    reason,            -- Human-readable reason from codes
    priority,          -- 0-100, calculated from score
    status,            -- 'pending'
    auto_flagged,      -- true
    confidence_score,  -- 0.00-1.00
    report_count,      -- 1
    created_at
) VALUES (...);
```

### Reason Code Mapping

| Code | Description |
|------|-------------|
| `VOTE_VELOCITY_HIGH` | High voting velocity |
| `FOLLOW_VELOCITY_HIGH` | High follow velocity |
| `SUBMISSION_VELOCITY_HIGH` | High submission velocity |
| `IP_SHARED_MULTIPLE_ACCOUNTS` | Multiple accounts from same IP |
| `IP_HOPPING_DETECTED` | Frequent IP address changes |
| `COORDINATED_VOTING_DETECTED` | Coordinated voting pattern |
| `CIRCULAR_FOLLOW_PATTERN` | Circular follow pattern |
| `BURST_ACTIVITY_DETECTED` | Burst activity pattern |
| `VOTE_PATTERN_MONOTONOUS` | Monotonous voting pattern |
| `TIMING_PATTERN_SUSPICIOUS` | Suspicious timing pattern |
| `LOW_TRUST_SCORE` | Low trust score |
| `NEW_ACCOUNT` | New account activity |

## Monitoring & Metrics

### Dashboard

Location: `monitoring/dashboards/abuse-detection.json`

**Key Panels:**
- Anomaly detection overview (total anomalies, auto-flagged items, FPR)
- Anomalies by action type (pie chart)
- Severity distribution
- Detection rate over time
- False positive rate monitoring (with 2% threshold alert)
- Moderation queue status
- Feature score distributions
- IP/UA signals
- Coordinated activity detection
- System health

### API Endpoints

#### GET /api/v1/admin/abuse/metrics

Returns real-time abuse detection metrics.

**Response:**
```json
{
  "anomaly_detection": {
    "anomalies_vote": "42",
    "anomalies_follow": "18",
    "anomalies_submission": "9",
    "auto_flagged_vote": "15",
    "auto_flagged_follow": "6",
    "auto_flagged_submission": "3",
    "severity_vote": {
      "low": "20",
      "medium": "15",
      "high": "5",
      "critical": "2"
    }
  },
  "auto_flagging": {
    "by_type": {
      "user": {
        "count": 21,
        "avg_confidence": 0.78,
        "avg_priority": 75
      },
      "submission": {
        "count": 3,
        "avg_confidence": 0.82,
        "avg_priority": 80
      }
    },
    "by_status": {
      "pending": 18,
      "approved": 4,
      "rejected": 2
    }
  },
  "timestamp": "2024-12-30T16:00:00Z",
  "time_range": "24h"
}
```

#### GET /api/v1/admin/abuse/metrics/history?hours=24

Returns historical metrics with FPR calculation.

**Response:**
```json
{
  "stats": { ... },
  "false_positive_rate": 0.016,
  "time_range_hours": 24,
  "fpr_target": 0.02,
  "fpr_meets_requirement": true
}
```

#### GET /api/v1/admin/abuse/health

Returns health status of abuse detection system.

**Response:**
```json
{
  "status": "healthy",
  "processing": true,
  "false_positive_rate": 0.016,
  "fpr_target": 0.02,
  "fpr_meets_requirement": true,
  "timestamp": "2024-12-30T16:00:00Z"
}
```

## Performance Considerations

### Redis Key Expiration

All tracking keys have automatic expiration:
- Velocity windows: 1-5 minutes to 1 hour
- IP/UA tracking: 24 hours
- Graph patterns: 7 days
- Timing patterns: 24 hours

This prevents unbounded Redis memory growth.

### Async Processing

- Auto-flagging occurs asynchronously (doesn't block user actions)
- Errors in anomaly detection are logged but don't fail requests
- Graceful degradation if Redis is unavailable

### Caching

Anomaly scores are cached in Redis for dashboard performance:
- Hourly anomaly score buckets for aggregation
- 7-day retention for anomaly score buckets (underlying velocity, IP/UA, timing, and other tracking keys expire according to their own windows: typically 1 hour to 24 hours, with graph patterns up to 7 days)

## False Positive Rate (FPR)

**Target:** < 2% as specified in roadmap

**Measurement:**
```
FPR = approved_flags / total_flags

Where:
- approved_flags = items reviewed and found legitimate
- total_flags = all auto-flagged items reviewed
```

**Monitoring:**
- Dashboard alert triggers if FPR > 2% for 5 minutes
- Weekly FPR reports sent to moderation team
- Threshold adjustments based on FPR trends

**Tuning:**
- Increase `autoFlagThreshold` (0.75) to reduce FPR
- Adjust component weights in scoring algorithm
- Refine feature thresholds based on observed patterns

## Security Considerations

1. **Privacy:** IP addresses and user agents are hashed (SHA-256) for all storage to protect user privacy while maintaining detection capability
2. **Access Control:** Analytics endpoints require admin role
3. **Rate Limiting:** Abuse checks don't create additional rate limit burden
4. **Audit Trail:** All auto-flags are logged with full context
5. **False Positives:** Legitimate users can appeal via moderation system

## Future Enhancements

1. **Machine Learning:** Replace statistical models with trained ML models
2. **Cross-Platform Patterns:** Detect abuse across web/mobile/API
3. **Reputation Networks:** Share abuse signals with trusted platforms
4. **Behavioral Fingerprinting:** Browser fingerprinting for better tracking
5. **Graph Algorithms:** More sophisticated network analysis

## References

- Roadmap 5.0 Phase 3.3: Anomaly Detection
- Issue: subculture-collective/clpr#805
- Moderation Queue: `migrations/000049_add_moderation_queue_system.up.sql`
- Existing Abuse Detection: `internal/services/submission_abuse_detection.go`
