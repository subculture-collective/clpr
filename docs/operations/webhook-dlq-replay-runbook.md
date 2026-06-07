---
title: "Webhook DLQ Replay Runbook"
summary: "This runbook provides operational procedures for managing and replaying failed webhook deliveries from the Dead-Letter Queue (DLQ). Use these procedures to ensure safe and effective recovery from webh"
tags: ["operations","runbook"]
area: "operations"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Webhook Dead-Letter Queue (DLQ) Replay Operations Runbook

## Overview

This runbook provides operational procedures for managing and replaying failed webhook deliveries from the Dead-Letter Queue (DLQ). Use these procedures to ensure safe and effective recovery from webhook delivery failures.

## Table of Contents

- [Prerequisites](#prerequisites)
- [DLQ Monitoring](#dlq-monitoring)
- [Investigation Procedures](#investigation-procedures)
- [Replay Procedures](#replay-procedures)
- [Safety Guidelines](#safety-guidelines)
- [Troubleshooting](#troubleshooting)
- [Performance Considerations](#performance-considerations)

## Prerequisites

### Access Requirements

- Admin or Moderator role in the system
- Valid authentication token
- Access to admin panel at `/admin/webhooks/dlq`
- API access for bulk operations (optional)

### Tools Required

- Web browser for admin panel access
- `curl` or similar HTTP client for API operations
- Database access for advanced troubleshooting (optional)

## DLQ Monitoring

### Check DLQ Size

**Via Admin Panel:**
1. Navigate to `/admin/webhooks/dlq`
2. View total count in pagination info

**Via API:**
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "https://api.clpr.example/api/v1/admin/webhooks/dlq?page=1&limit=1"
```

Response includes `pagination.total` with total DLQ item count.

### Alert Thresholds

Set up monitoring alerts for:
- **Warning:** DLQ size > 100 items
- **Critical:** DLQ size > 500 items
- **Emergency:** DLQ size > 1000 items

### Key Metrics to Monitor

- `webhook_dlq_movements_total` - Items moved to DLQ
- `webhook_dlq_replay_success_total` - Successful replays
- `webhook_dlq_replay_failure_total` - Failed replays
- DLQ item age (time since moved to DLQ)

## Investigation Procedures

### Step 1: Identify Patterns

Review DLQ items for common patterns:

```bash
# Fetch recent DLQ items
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "https://api.clpr.example/api/v1/admin/webhooks/dlq?page=1&limit=20"
```

**Look for:**
- Common subscription IDs (specific webhook failing)
- Common event types (specific events failing)
- Common error messages (systemic issue)
- Time patterns (specific time windows)

### Step 2: Analyze Error Messages

Common error patterns:

| Error Pattern | Likely Cause | Action |
|--------------|--------------|--------|
| `HTTP 500: Internal Server Error` | Subscriber endpoint issue | Contact subscriber to fix endpoint |
| `network error: connection refused` | Endpoint unreachable | Verify endpoint URL and accessibility |
| `network error: timeout` | Slow subscriber endpoint | Ask subscriber to optimize or increase timeout |
| `HTTP 401: Unauthorized` | Invalid signature or auth | Verify webhook secret is correct |
| `HTTP 404: Not Found` | Incorrect webhook URL | Update subscription URL |

### Step 3: Verify Subscriber Endpoint

Before replay, verify the endpoint is healthy:

```bash
# Test endpoint availability
curl -I https://subscriber-webhook-endpoint.example/webhook

# Check if endpoint returns 200 OK
```

### Step 4: Review Subscriber Status

Check if webhook subscription is still active:

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "https://api.clpr.example/api/v1/webhooks/{subscription_id}"
```

Verify `is_active: true` before replay.

## Replay Procedures

### Single Item Replay

**Use Case:** Testing or replaying a specific failed webhook

**Via Admin Panel:**
1. Navigate to `/admin/webhooks/dlq`
2. Find the item to replay
3. Click "Replay" button
4. Confirm the action
5. Wait for confirmation message

**Via API:**
```bash
curl -X POST \
  -H "Authorization: Bearer YOUR_TOKEN" \
  "https://api.clpr.example/api/v1/admin/webhooks/dlq/{dlq_item_id}/replay"
```

**Expected Response:**
```json
{
  "message": "Webhook replayed successfully"
}
```

**Monitor:**
- Check webhook delivery logs
- Verify subscriber received the event
- Confirm DLQ item marked as replayed

### Bulk Replay

**Use Case:** Replaying multiple failed webhooks after fixing systemic issue

**Safety Check:**
1. ✅ Verify subscriber endpoint is healthy
2. ✅ Confirm subscription is active
3. ✅ Identify specific DLQ items to replay (by subscription, event type, or time range)
4. ✅ Plan for rate limiting (normal max 20 replays/sec; emergency rate of 50 replays/sec requires active monitoring)

**Procedure:**

```bash
#!/bin/bash
# Bulk replay script with rate limiting

AUTH_TOKEN="your_token_here"
API_BASE="https://api.clpr.example"
RATE_LIMIT=20  # replays per second
DELAY=$(echo "scale=3; 1/$RATE_LIMIT" | bc)

# Fetch DLQ items
DLQ_ITEMS=$(curl -s -H "Authorization: Bearer $AUTH_TOKEN" \
  "$API_BASE/api/v1/admin/webhooks/dlq?page=1&limit=100" | \
  jq -r '.items[].id')

echo "Found $(echo "$DLQ_ITEMS" | wc -l) items to replay"

# Replay each item with rate limiting
for item_id in $DLQ_ITEMS; do
  echo "Replaying $item_id..."
  
  response=$(curl -s -w "\n%{http_code}" -X POST \
    -H "Authorization: Bearer $AUTH_TOKEN" \
    "$API_BASE/api/v1/admin/webhooks/dlq/$item_id/replay")
  
  http_code=$(echo "$response" | tail -n1)
  
  if [ "$http_code" -eq 200 ]; then
    echo "  ✓ Success"
  elif [ "$http_code" -eq 429 ]; then
    echo "  ⚠ Rate limited, backing off..."
    sleep 5
  else
    echo "  ✗ Failed (HTTP $http_code)"
  fi
  
  # Rate limiting delay
  sleep $DELAY
done

echo "Bulk replay completed"
```

**Monitoring During Bulk Replay:**
```bash
# Monitor replay progress
watch -n 5 'curl -s -H "Authorization: Bearer $AUTH_TOKEN" \
  "https://api.clpr.example/api/v1/admin/webhooks/dlq" | \
  jq ".pagination.total"'
```

### Selective Replay (By Filter)

Replay only specific DLQ items:

```bash
# Example: Replay only clip.submitted events
curl -s -H "Authorization: Bearer $AUTH_TOKEN" \
  "$API_BASE/api/v1/admin/webhooks/dlq?page=1&limit=100" | \
  jq -r '.items[] | select(.event_type == "clip.submitted") | .id' | \
  while read item_id; do
    curl -X POST -H "Authorization: Bearer $AUTH_TOKEN" \
      "$API_BASE/api/v1/admin/webhooks/dlq/$item_id/replay"
    sleep 0.1
  done
```

## Safety Guidelines

### Pre-Replay Checklist

- [ ] Identify root cause of failures
- [ ] Verify root cause has been resolved
- [ ] Confirm subscriber endpoint is operational
- [ ] Check webhook subscription is active
- [ ] Estimate replay volume and duration
- [ ] Plan for rate limiting
- [ ] Notify subscriber of incoming replays (if large volume)
- [ ] Have rollback plan ready

### Rate Limiting Guidelines

**Recommended Rates:**
- **Single endpoint:** 5-10 replays/sec
- **Multiple endpoints:** 20 replays/sec total (normal maximum)
- **Emergency recovery:** 50 replays/sec (requires active monitoring, use only in critical situations)

**Backoff Strategy:**
If rate limited (HTTP 429):
1. First retry: Wait 1 second
2. Second retry: Wait 2 seconds
3. Third retry: Wait 4 seconds
4. Continue exponential backoff up to 60 seconds

### Monitoring During Replay

**Watch for:**
- Increased error rates
- Subscriber endpoint becoming overwhelmed
- Database connection pool exhaustion
- Network saturation
- HTTP 429 (rate limit) responses

**Stop replay if:**
- Error rate > 20%
- Subscriber reports issues
- System load becomes critical
- Database performance degrades

## Troubleshooting

### Issue: Replay Returns 404

**Cause:** DLQ item no longer exists or already replayed

**Solution:**
1. Verify item ID is correct
2. Check if item was already replayed successfully
3. Query DLQ to confirm item exists

### Issue: Replay Returns 500

**Cause:** Internal server error during replay

**Solution:**
1. Check server logs for detailed error
2. Verify subscription still exists
3. Check database connectivity
4. Retry after a few minutes

### Issue: High Failure Rate During Bulk Replay

**Cause:** Subscriber endpoint cannot handle load or still has issues

**Solution:**
1. Pause bulk replay immediately
2. Test single replay to verify endpoint health
3. Reduce replay rate
4. Contact subscriber to optimize endpoint
5. Consider batch size reduction

### Issue: Replays Timing Out

**Cause:** Subscriber endpoint too slow or experiencing issues

**Solution:**
1. Verify subscriber endpoint response time
2. Check subscriber logs for processing delays
3. Consider increasing webhook timeout (not recommended)
4. Ask subscriber to optimize processing

### Issue: Duplicate Events at Subscriber

**Cause:** Replay of events that were actually delivered

**Solution:**
1. Verify subscriber implements idempotency checking
2. Check if delivery was marked as failed incorrectly
3. Use `X-Webhook-Replay: true` header to distinguish replays
4. Subscriber should use `X-Webhook-Delivery-ID` for deduplication

## Performance Considerations

### Database Impact

Replaying large volumes can impact database:
- Each replay requires multiple DB queries
- Connection pool may be exhausted
- Consider off-peak hours for bulk replays

### Network Impact

- Each replay generates HTTP request
- Consider network bandwidth
- Monitor egress traffic

### Subscriber Impact

Before bulk replay:
1. Estimate peak request rate to subscriber
2. Verify subscriber can handle load
3. Consider subscriber's rate limits
4. Notify subscriber of incoming replay

### Metrics to Track

During replay operations:
- `webhook_dlq_replay_success_total`
- `webhook_dlq_replay_failure_total`
- `http_request_duration_seconds{endpoint="/api/v1/admin/webhooks/dlq/*/replay"}`
- Database query duration
- HTTP client timeout rate

## Cleanup Procedures

### Archive Old DLQ Items

For items that cannot be replayed (permanent failures):

```bash
# Delete items older than 30 days
curl -X DELETE -H "Authorization: Bearer $AUTH_TOKEN" \
  "https://api.clpr.example/api/v1/admin/webhooks/dlq/{item_id}"
```

**Consider archiving if:**
- Item is >30 days old
- Subscription no longer exists
- Endpoint permanently unavailable
- Data no longer relevant

### Verify Cleanup

```bash
# Check DLQ size after cleanup
curl -H "Authorization: Bearer $AUTH_TOKEN" \
  "https://api.clpr.example/api/v1/admin/webhooks/dlq" | \
  jq ".pagination.total"
```

## Emergency Procedures

### Critical DLQ Overflow (>1000 items)

1. **Immediate Actions:**
   - Stop investigating individual items
   - Identify if single subscription causing issue
   - Disable problematic subscription if necessary
   - Alert on-call engineer

2. **Triage:**
   - Group items by subscription
   - Identify most critical events
   - Prioritize by business impact

3. **Recovery:**
   - Fix root cause first
   - Replay in small batches (100 items)
   - Monitor system health between batches
   - Consider dropping very old items (>7 days)

### System Overload During Replay

If replay causes system issues:

```bash
# DANGER: Emergency procedure to disable all webhook deliveries
# This bypasses application logic and should be used ONLY in critical system overload situations
# IMPORTANT: Document which subscriptions are disabled for later re-enablement
# IMPORTANT: Ensure you have a database backup before running this command
# Example (requires database access):
# UPDATE outbound_webhook_subscriptions SET is_active = false WHERE is_active = true;
```

**Recovery steps:**
1. Stop all replay operations immediately
2. Create database backup
3. Document currently active subscriptions before disabling
4. Investigate system bottleneck (CPU, memory, database connections)
5. Scale resources if needed
6. Re-enable subscriptions gradually, starting with critical ones
7. Resume replay at reduced rate

## Best Practices

1. **Regular Monitoring:** Check DLQ daily
2. **Proactive Investigation:** Investigate items within 24 hours
3. **Root Cause Analysis:** Fix underlying issues before replay
4. **Rate Limiting:** Never exceed subscriber's capacity
5. **Idempotency:** Ensure subscribers handle duplicate events
6. **Communication:** Notify subscribers of bulk replays
7. **Documentation:** Log all replay operations
8. **Automation:** Use scripts for bulk operations with safety checks

## Related Documentation

- [Webhook Signature Verification](../backend/webhook-signature-verification.md)
- [Webhook Subscription Management](../backend/webhook-subscription-management.md)
- [Webhook Retry Policy](../backend/webhook-retry.md)
- [Webhook Monitoring](./webhook-monitoring.md)

## Support

For issues or questions:
1. Check this runbook first
2. Review webhook logs
3. Check DLQ admin panel
4. Contact platform team
5. Escalate to on-call if critical
