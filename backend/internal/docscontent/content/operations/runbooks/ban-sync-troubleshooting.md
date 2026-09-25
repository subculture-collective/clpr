---
title: "Ban Sync Troubleshooting Runbook"
summary: "Troubleshooting guide for Twitch ban synchronization issues"
tags: ["operations", "runbook", "twitch", "troubleshooting", "bans"]
area: "moderation"
status: "active"
owner: "team-ops"
version: "1.0"
last_reviewed: 2026-02-03
aliases: ["ban-sync", "twitch-sync"]
---

# Ban Sync Troubleshooting Runbook

## Overview

This runbook provides troubleshooting procedures for Twitch ban synchronization issues in the Clipper moderation system. Ban sync allows importing ban lists from Twitch to maintain consistency across platforms.

**Audience**: Operations team, on-call engineers, support team

**Prerequisites**:
- Admin permissions
- Valid JWT authentication token
- Understanding of Twitch API integration
- Access to application logs

## Table of Contents

- [How Ban Sync Works](#how-ban-sync-works)
- [Common Sync Errors](#common-sync-errors)
  - [OAuth Scope Issues](#oauth-scope-issues)
  - [Twitch API Rate Limits](#twitch-api-rate-limits)
  - [Invalid Channel/Broadcaster ID](#invalid-channelbroadcaster-id)
  - [Authentication Failures](#authentication-failures)
  - [Network Timeouts](#network-timeouts)
- [Troubleshooting Procedures](#troubleshooting-procedures)
  - [Check Sync Status](#check-sync-status)
  - [Verify OAuth Scopes](#verify-oauth-scopes)
  - [Test Twitch API Connectivity](#test-twitch-api-connectivity)
  - [Review Sync Logs](#review-sync-logs)
  - [Manual Sync Retry](#manual-sync-retry)
- [Rate Limit Management](#rate-limit-management)
- [Known Issues and Workarounds](#known-issues-and-workarounds)
- [Escalation Procedures](#escalation-procedures)
- [Related Runbooks](#related-runbooks)

---

## How Ban Sync Works

### Process Flow

1. **User initiates sync** (via UI or API)
2. **Backend validates permissions** (must be broadcaster or admin)
3. **OAuth token verified** (checks for required scopes)
4. **Twitch API called** (`GET /moderation/banned`)
5. **Bans imported** to Clipper database
6. **Audit log created** for sync operation
7. **UI updated** with new ban list

### Technical Details

- **Endpoint**: `POST /api/v1/moderation/sync-bans`
- **Twitch API**: `GET https://api.twitch.tv/helix/moderation/banned`
- **Required Scopes**: `moderator:read:banned_users` or `channel:read:banned_users`
- **Rate Limit**: 800 requests per minute (Twitch)
- **Sync Frequency**: On-demand only (no automatic sync)

### Data Synchronized

- User ID (Twitch ID)
- Ban reason
- Moderator who banned
- Ban timestamp
- Expiration (for timeouts)

---

## Common Sync Errors

### OAuth Scope Issues

#### Error Message

```json
{
  "error": "INSUFFICIENT_SCOPES",
  "message": "Missing required OAuth scopes: moderator:read:banned_users",
  "required_scopes": ["moderator:read:banned_users", "channel:read:banned_users"],
  "current_scopes": ["user:read:email"]
}
```

#### Cause

User hasn't granted Twitch OAuth permissions to read banned users list.

#### Solution

1. **User must re-authenticate with Twitch**
   ```bash
   # Direct user to re-auth URL
   echo "https://api.clpr.tv/api/v1/auth/twitch?scope=moderator:read:banned_users"
   ```

2. **Verify scopes after re-auth**
   ```bash
   curl -s -H "Authorization: Bearer $API_TOKEN" \
     "https://api.clpr.tv/api/v1/auth/me" | jq '.twitch_scopes'
   ```

3. **Test sync again**
   ```bash
   curl -X POST -H "Authorization: Bearer $API_TOKEN" \
     "https://api.clpr.tv/api/v1/moderation/sync-bans" \
     -H "Content-Type: application/json" \
     -d '{"broadcaster_id": "'$BROADCASTER_ID'"}'
   ```

#### Prevention

- Include required scopes in initial OAuth flow
- Display clear error message in UI
- Provide "Reconnect to Twitch" button

---

### Twitch API Rate Limits

#### Error Message

```json
{
  "error": "RATE_LIMIT_EXCEEDED",
  "message": "Twitch API rate limit exceeded",
  "retry_after": 60,
  "limit": 800,
  "remaining": 0,
  "reset_at": "2026-02-03T10:05:00Z"
}
```

#### Cause

Too many requests to Twitch API in short time period.

#### Solution

1. **Check rate limit status**
   ```bash
   curl -s -H "Authorization: Bearer $API_TOKEN" \
     "https://api.clpr.tv/api/v1/moderation/sync-bans/status" | jq '.rate_limit'
   ```

2. **Wait for rate limit reset**
   ```bash
   # Extract reset time from error
   RESET_TIME="2026-02-03T10:05:00Z"
   
   # Calculate wait time
   RESET_EPOCH=$(date -d "$RESET_TIME" +%s)
   NOW_EPOCH=$(date +%s)
   WAIT_SECONDS=$((RESET_EPOCH - NOW_EPOCH))
   
   echo "Wait $WAIT_SECONDS seconds before retrying"
   sleep $WAIT_SECONDS
   ```

3. **Retry sync**
   ```bash
   curl -X POST -H "Authorization: Bearer $API_TOKEN" \
     "https://api.clpr.tv/api/v1/moderation/sync-bans" \
     -H "Content-Type: application/json" \
     -d '{"broadcaster_id": "'$BROADCASTER_ID'"}'
   ```

#### Prevention

- Implement exponential backoff
- Cache sync results (don't re-sync within 5 minutes)
- Display "last synced" timestamp to users
- Queue sync requests during high traffic

---

### Invalid Channel/Broadcaster ID

#### Error Message

```json
{
  "error": "INVALID_BROADCASTER_ID",
  "message": "Broadcaster not found or not accessible",
  "broadcaster_id": "invalid-id-123"
}
```

#### Cause

1. Broadcaster ID doesn't exist
2. User doesn't have access to that channel
3. Broadcaster ID format is incorrect

#### Solution

1. **Verify broadcaster ID format**
   ```bash
   # Twitch IDs are numeric strings
   # Valid: "123456789"
   # Invalid: "username123", "channel-abc-def"
   
   BROADCASTER_ID="123456789"
   if ! [[ "$BROADCASTER_ID" =~ ^[0-9]+$ ]]; then
     echo "Invalid format: must be numeric"
   fi
   ```

2. **Lookup broadcaster by username**
   ```bash
   # If you have username, get Twitch ID
   curl -s -H "Authorization: Bearer $API_TOKEN" \
     "https://api.clpr.tv/api/v1/twitch/users?username=$USERNAME" | jq '.id'
   ```

3. **Verify user has access**
   ```bash
   # Check if user is broadcaster or moderator
   curl -s -H "Authorization: Bearer $API_TOKEN" \
     "https://api.clpr.tv/api/v1/users/me" | jq '{
       twitch_id,
       is_broadcaster: (.twitch_id == "'$BROADCASTER_ID'"),
       is_moderator: .moderator_channels[]?
     }'
   ```

4. **Test with correct ID**
   ```bash
   curl -X POST -H "Authorization: Bearer $API_TOKEN" \
     "https://api.clpr.tv/api/v1/moderation/sync-bans" \
     -H "Content-Type: application/json" \
     -d '{"broadcaster_id": "'$CORRECT_BROADCASTER_ID'"}'
   ```

---

### Authentication Failures

#### Error Message

```json
{
  "error": "TWITCH_AUTH_FAILED",
  "message": "Failed to authenticate with Twitch API",
  "details": "Invalid or expired OAuth token"
}
```

#### Cause

1. Twitch OAuth token expired
2. Token revoked by user
3. Twitch API authentication service down

#### Solution

1. **Check token expiration**
   ```bash
   curl -s -H "Authorization: Bearer $API_TOKEN" \
     "https://api.clpr.tv/api/v1/auth/me" | jq '{
       twitch_connected: .twitch_id,
       token_expires: .twitch_token_expires_at
     }'
   ```

2. **Refresh Twitch token**
   ```bash
   curl -X POST -H "Authorization: Bearer $API_TOKEN" \
     "https://api.clpr.tv/api/v1/auth/twitch/refresh"
   ```

3. **If refresh fails, require re-authentication**
   ```bash
   # User must re-connect Twitch account
   echo "Redirect user to: https://api.clpr.tv/api/v1/auth/twitch"
   ```

4. **Verify Twitch API status**
   ```bash
   # Check Twitch API health
   curl -s "https://api.twitch.tv/helix/users" \
     -H "Client-ID: $TWITCH_CLIENT_ID" \
     -H "Authorization: Bearer $TWITCH_TOKEN" | jq
   ```

#### Prevention

- Implement automatic token refresh
- Refresh tokens 24 hours before expiration
- Handle token revocation gracefully
- Monitor Twitch API status page

---

### Network Timeouts

#### Error Message

```json
{
  "error": "SYNC_TIMEOUT",
  "message": "Ban sync operation timed out",
  "duration_ms": 30000,
  "timeout_ms": 30000
}
```

#### Cause

1. Large ban list (> 1000 bans)
2. Slow Twitch API response
3. Network connectivity issues
4. Database write bottleneck

#### Solution

1. **Check sync progress**
   ```bash
   # Get sync job status
   SYNC_JOB_ID="sync-abc123"
   curl -s -H "Authorization: Bearer $API_TOKEN" \
     "https://api.clpr.tv/api/v1/moderation/sync-bans/$SYNC_JOB_ID" | jq '{
       status,
       progress: .synced_count,
       total: .total_count,
       started_at,
       updated_at
     }'
   ```

2. **Wait for background sync to complete**
   ```bash
   # Sync may continue in background
   # Check status every 30 seconds
   while true; do
     STATUS=$(curl -s -H "Authorization: Bearer $API_TOKEN" \
       "https://api.clpr.tv/api/v1/moderation/sync-bans/$SYNC_JOB_ID" | jq -r '.status')
     
     if [ "$STATUS" = "completed" ]; then
       echo "Sync completed successfully"
       break
     elif [ "$STATUS" = "failed" ]; then
       echo "Sync failed"
       break
     fi
     
     echo "Sync in progress... ($STATUS)"
     sleep 30
   done
   ```

3. **If stuck, cancel and retry**
   ```bash
   # Cancel stalled sync
   curl -X DELETE -H "Authorization: Bearer $API_TOKEN" \
     "https://api.clpr.tv/api/v1/moderation/sync-bans/$SYNC_JOB_ID"
   
   # Wait 30 seconds
   sleep 30
   
   # Retry sync
   curl -X POST -H "Authorization: Bearer $API_TOKEN" \
     "https://api.clpr.tv/api/v1/moderation/sync-bans" \
     -H "Content-Type: application/json" \
     -d '{"broadcaster_id": "'$BROADCASTER_ID'"}'
   ```

#### Prevention

- Increase timeout for large channels
- Implement pagination for large ban lists
- Add progress indicators in UI
- Use async/background jobs for syncs

---

## Troubleshooting Procedures

### Check Sync Status

**Use case**: Verify if sync is working, find recent sync history

```bash
#!/bin/bash
# check-sync-status.sh

API_TOKEN="${API_TOKEN}"
API_BASE="https://api.clpr.tv/api/v1/moderation"
BROADCASTER_ID="${1:-}"

if [ -z "$BROADCASTER_ID" ]; then
  echo "Usage: $0 <broadcaster_id>"
  exit 1
fi

echo "Ban Sync Status for Broadcaster: $BROADCASTER_ID"
echo "================================================"
echo

# Get last sync
echo "Last Sync:"
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/audit-logs?action=sync_bans&channel_id=$BROADCASTER_ID&limit=1" | \
  jq -r '.logs[0] | "  Time: \(.created_at)\n  Status: \(.details.status // "completed")\n  Count: \(.details.ban_count // "N/A") bans\n  Actor: \(.actor_username)"'

echo
echo "Recent Sync History (last 5):"
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/audit-logs?action=sync_bans&channel_id=$BROADCASTER_ID&limit=5" | \
  jq -r '.logs[] | "  \(.created_at) - \(.details.ban_count // 0) bans - \(.details.status // "completed")"'

echo
echo "Current Ban Count:"
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/bans?channel_id=$BROADCASTER_ID&limit=1" | jq -r '  Total: \(.total_count)'
```

**Usage**:
```bash
chmod +x check-sync-status.sh
./check-sync-status.sh 123456789
```

---

### Verify OAuth Scopes

**Use case**: Check if user has required permissions for ban sync

```bash
#!/bin/bash
# verify-oauth-scopes.sh

API_TOKEN="${API_TOKEN}"

echo "OAuth Scope Verification"
echo "========================="
echo

# Get user's current scopes
SCOPES=$(curl -s -H "Authorization: Bearer $API_TOKEN" \
  "https://api.clpr.tv/api/v1/auth/me" | jq -r '.twitch_scopes[]')

# Required scopes for ban sync
REQUIRED=("moderator:read:banned_users" "channel:read:banned_users")

echo "Current Scopes:"
echo "$SCOPES" | sed 's/^/  - /'
echo

echo "Required Scopes (at least one):"
for scope in "${REQUIRED[@]}"; do
  if echo "$SCOPES" | grep -q "$scope"; then
    echo "  ✓ $scope"
  else
    echo "  ✗ $scope (MISSING)"
  fi
done

echo

# Check if user can sync
CAN_SYNC=false
for scope in "${REQUIRED[@]}"; do
  if echo "$SCOPES" | grep -q "$scope"; then
    CAN_SYNC=true
    break
  fi
done

if [ "$CAN_SYNC" = true ]; then
  echo "Result: ✓ User CAN perform ban sync"
else
  echo "Result: ✗ User CANNOT perform ban sync"
  echo "Action: User must re-authenticate with Twitch"
  echo "URL: https://api.clpr.tv/api/v1/auth/twitch?scope=moderator:read:banned_users"
fi
```

---

### Test Twitch API Connectivity

**Use case**: Verify Twitch API is accessible and responding

```bash
#!/bin/bash
# test-twitch-api.sh

API_TOKEN="${API_TOKEN}"

echo "Twitch API Connectivity Test"
echo "=============================="
echo

# Get Twitch token from Clipper
TWITCH_TOKEN=$(curl -s -H "Authorization: Bearer $API_TOKEN" \
  "https://api.clpr.tv/api/v1/auth/me" | jq -r '.twitch_access_token')

if [ -z "$TWITCH_TOKEN" ] || [ "$TWITCH_TOKEN" = "null" ]; then
  echo "✗ No Twitch token found"
  echo "  User not connected to Twitch"
  exit 1
fi

# Test Twitch API with token
echo "Testing Twitch API..."
RESPONSE=$(curl -s -w "\n%{http_code}" \
  "https://api.twitch.tv/helix/users" \
  -H "Client-ID: $TWITCH_CLIENT_ID" \
  -H "Authorization: Bearer $TWITCH_TOKEN")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "200" ]; then
  echo "✓ Twitch API is accessible"
  echo "  User: $(echo "$BODY" | jq -r '.data[0].display_name')"
else
  echo "✗ Twitch API error"
  echo "  HTTP Code: $HTTP_CODE"
  echo "  Response: $BODY"
fi

# Test banned users endpoint
echo
echo "Testing banned users endpoint..."
BROADCASTER_ID="${BROADCASTER_ID:-$(echo "$BODY" | jq -r '.data[0].id')}"

RESPONSE=$(curl -s -w "\n%{http_code}" \
  "https://api.twitch.tv/helix/moderation/banned?broadcaster_id=$BROADCASTER_ID" \
  -H "Client-ID: $TWITCH_CLIENT_ID" \
  -H "Authorization: Bearer $TWITCH_TOKEN")

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "200" ]; then
  BAN_COUNT=$(echo "$BODY" | jq '.data | length')
  echo "✓ Banned users endpoint accessible"
  echo "  Bans returned: $BAN_COUNT"
else
  echo "✗ Banned users endpoint error"
  echo "  HTTP Code: $HTTP_CODE"
  echo "  Response: $BODY"
fi
```

---

### Review Sync Logs

**Use case**: Investigate failed syncs, find error patterns

```bash
# View failed sync attempts
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/audit-logs?action=sync_bans&limit=50" | \
  jq '.logs[] | select(.details.status == "failed") | {
    time: .created_at,
    actor: .actor_username,
    broadcaster: .channel_id,
    error: .details.error
  }'

# Count sync success/failure rate
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/audit-logs?action=sync_bans&limit=100" | \
  jq '[.logs[] | .details.status // "completed"] | group_by(.) | map({status: .[0], count: length})'
```

---

### Manual Sync Retry

**Use case**: Force sync after fixing issues

```bash
#!/bin/bash
# manual-sync-retry.sh

API_TOKEN="${API_TOKEN}"
API_BASE="https://api.clpr.tv/api/v1/moderation"
BROADCASTER_ID="${1:-}"

if [ -z "$BROADCASTER_ID" ]; then
  echo "Usage: $0 <broadcaster_id>"
  exit 1
fi

echo "Initiating manual ban sync for broadcaster: $BROADCASTER_ID"
echo "=============================================================="
echo

# Start sync
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
  "$API_BASE/sync-bans" \
  -H "Authorization: Bearer $API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"broadcaster_id": "'$BROADCASTER_ID'"}')

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "201" ]; then
  echo "✓ Sync initiated successfully"
  SYNC_JOB_ID=$(echo "$BODY" | jq -r '.job_id // .id')
  echo "  Job ID: $SYNC_JOB_ID"
  
  # Monitor progress
  echo
  echo "Monitoring sync progress..."
  for i in {1..20}; do
    sleep 3
    STATUS_RESPONSE=$(curl -s "$API_BASE/sync-bans/$SYNC_JOB_ID" \
      -H "Authorization: Bearer $API_TOKEN")
    
    STATUS=$(echo "$STATUS_RESPONSE" | jq -r '.status')
    PROGRESS=$(echo "$STATUS_RESPONSE" | jq -r '.synced_count // 0')
    TOTAL=$(echo "$STATUS_RESPONSE" | jq -r '.total_count // 0')
    
    echo "  Status: $STATUS - Progress: $PROGRESS/$TOTAL"
    
    if [ "$STATUS" = "completed" ]; then
      echo "✓ Sync completed successfully"
      echo "  Synced: $PROGRESS bans"
      exit 0
    elif [ "$STATUS" = "failed" ]; then
      ERROR=$(echo "$STATUS_RESPONSE" | jq -r '.error')
      echo "✗ Sync failed: $ERROR"
      exit 1
    fi
  done
  
  echo "⚠ Sync still in progress after 60 seconds"
  echo "  Check status later with job ID: $SYNC_JOB_ID"
else
  echo "✗ Failed to initiate sync"
  echo "  HTTP Code: $HTTP_CODE"
  echo "  Error: $(echo "$BODY" | jq -r '.error // .message')"
  exit 1
fi
```

---

## Rate Limit Management

### Monitor Rate Limit Status

```bash
# Check current rate limit status
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "https://api.clpr.tv/api/v1/moderation/rate-limit" | jq '{
    limit: .twitch_rate_limit,
    remaining: .twitch_rate_remaining,
    reset_at: .twitch_rate_reset_at,
    percentage_used: ((.twitch_rate_limit - .twitch_rate_remaining) / .twitch_rate_limit * 100)
  }'
```

### Implement Backoff Strategy

```bash
#!/bin/bash
# sync-with-backoff.sh

MAX_RETRIES=5
RETRY_COUNT=0
WAIT_TIME=5

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
  echo "Attempt $((RETRY_COUNT + 1))/$MAX_RETRIES"
  
  RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
    "$API_BASE/sync-bans" \
    -H "Authorization: Bearer $API_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"broadcaster_id": "'$BROADCASTER_ID'"}')
  
  HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
  BODY=$(echo "$RESPONSE" | sed '$d')
  
  if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "201" ]; then
    echo "✓ Sync successful"
    exit 0
  elif [ "$HTTP_CODE" = "429" ]; then
    echo "Rate limited. Waiting $WAIT_TIME seconds..."
    sleep $WAIT_TIME
    WAIT_TIME=$((WAIT_TIME * 2))  # Exponential backoff
    RETRY_COUNT=$((RETRY_COUNT + 1))
  else
    echo "✗ Sync failed with HTTP $HTTP_CODE"
    echo "Error: $(echo "$BODY" | jq -r '.message')"
    exit 1
  fi
done

echo "✗ Max retries exceeded"
exit 1
```

---

## Known Issues and Workarounds

### Issue: Duplicate Bans After Sync

**Symptoms**: Same user appears multiple times in ban list

**Cause**: Race condition when multiple syncs run simultaneously

**Workaround**:
```bash
# Deduplicate bans (keep most recent)
curl -X POST -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/bans/deduplicate?channel_id=$BROADCASTER_ID"
```

**Prevention**: Prevent multiple syncs from same user within 5 minutes

---

### Issue: Expired Bans Not Removed

**Symptoms**: Temporary bans still shown after expiration

**Cause**: No automatic cleanup process

**Workaround**:
```bash
# Manual cleanup of expired bans
curl -X POST -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/bans/cleanup-expired?channel_id=$BROADCASTER_ID"
```

**Long-term fix**: Implement scheduled job to clean expired bans

---

### Issue: Sync Shows 0 Bans for Active Channel

**Symptoms**: Sync completes but reports 0 bans despite Twitch showing bans

**Cause**: Incorrect broadcaster ID or insufficient permissions

**Debug steps**:
```bash
# 1. Verify broadcaster ID
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "https://api.clpr.tv/api/v1/auth/me" | jq '.twitch_id'

# 2. Check Twitch directly
curl -s "https://api.twitch.tv/helix/moderation/banned?broadcaster_id=$BROADCASTER_ID" \
  -H "Client-ID: $TWITCH_CLIENT_ID" \
  -H "Authorization: Bearer $TWITCH_TOKEN" | jq '.data | length'

# 3. Compare results
```

---

## Escalation Procedures

### When to Escalate

Escalate to engineering team if:
- [ ] Sync fails consistently for > 1 hour
- [ ] Multiple users reporting same issue
- [ ] Twitch API returns 500 errors
- [ ] Database write failures
- [ ] Security concern (e.g., bans not being applied)

### Escalation Process

1. **Gather diagnostic information**
   ```bash
   # Create diagnostic report
   {
     echo "=== Ban Sync Diagnostic Report ==="
     echo "Date: $(date -u)"
     echo "Broadcaster ID: $BROADCASTER_ID"
     echo
     echo "=== Sync History ==="
     ./check-sync-status.sh $BROADCASTER_ID
     echo
     echo "=== OAuth Scopes ==="
     ./verify-oauth-scopes.sh
     echo
     echo "=== Twitch API Test ==="
     ./test-twitch-api.sh
     echo
     echo "=== Recent Errors ==="
     curl -s -H "Authorization: Bearer $API_TOKEN" \
       "$API_BASE/audit-logs?action=sync_bans&limit=10" | \
       jq '.logs[] | select(.details.status == "failed")'
   } > diagnostic-report-$(date +%Y%m%d-%H%M%S).txt
   ```

2. **Create incident ticket**
   - **Severity**: P1 (if affecting all users), P2 (single user)
   - **Title**: "Ban Sync Failure - [Broadcaster ID]"
   - **Description**: Include diagnostic report
   - **Impact**: Number of users affected

3. **Notify on-call engineer**
   - Slack: #ops-incidents
   - PagerDuty: Trigger alert
   - Email: ops@clpr.tv

4. **Document workaround** (if available)
   - Temporary solution for users
   - Expected resolution time

---

## Related Runbooks

- [Moderation Operations](./moderation-operations.md) - General moderation procedures
- [Audit Log Operations](./audit-log-operations.md) - Review sync logs
- [Moderation Incidents](./moderation-incidents.md) - Incident response
- [Permission Escalation](./permission-escalation.md) - Grant emergency access

---

## Additional Resources

- [Twitch API Documentation](https://dev.twitch.tv/docs/api/reference#get-banned-users)
- [Twitch API Status](https://devstatus.twitch.tv/)
- [Clipper Moderation API Docs](../../backend/moderation-api.md)

---

**Last Updated**: 2026-02-03  
**Document Owner**: Operations Team  
**Review Frequency**: Quarterly
