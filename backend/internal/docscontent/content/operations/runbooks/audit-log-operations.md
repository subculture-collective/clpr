---
title: "Audit Log Operations Runbook"
summary: "Procedures for reviewing, exporting, and analyzing moderation audit logs"
tags: ["operations", "runbook", "audit", "compliance", "security"]
area: "moderation"
status: "active"
owner: "team-ops"
version: "1.0"
last_reviewed: 2026-02-03
aliases: ["audit-logs", "log-review"]
---

# Audit Log Operations Runbook

## Overview

This runbook provides procedures for reviewing, exporting, and analyzing moderation audit logs in the Clipper platform. Audit logs track all moderation actions for compliance, security, and troubleshooting purposes.

**Audience**: Operations team, security team, compliance officers

**Prerequisites**:
- Admin or audit-reviewer role
- Valid JWT authentication token
- Access to log storage (S3/database)

## Table of Contents

- [Audit Log Overview](#audit-log-overview)
- [Reviewing Audit Logs](#reviewing-audit-logs)
  - [View Recent Logs](#view-recent-logs)
  - [Filter by Action Type](#filter-by-action-type)
  - [Filter by User/Actor](#filter-by-useractor)
  - [Search by Time Range](#search-by-time-range)
- [Exporting Audit Logs](#exporting-audit-logs)
  - [Export to CSV](#export-to-csv)
  - [Export to JSON](#export-to-json)
  - [Scheduled Exports](#scheduled-exports)
- [Common Patterns to Look For](#common-patterns-to-look-for)
  - [Unauthorized Access Attempts](#unauthorized-access-attempts)
  - [Suspicious Ban/Unban Activity](#suspicious-banunban-activity)
  - [Moderator Abuse](#moderator-abuse)
  - [Mass Operations](#mass-operations)
- [Investigation Procedures](#investigation-procedures)
- [Compliance and Retention](#compliance-and-retention)
- [Related Runbooks](#related-runbooks)

---

## Audit Log Overview

### What Gets Logged

All moderation actions are logged, including:

| Action Type | Description | Example |
|------------|-------------|---------|
| `ban_user` | User banned from channel | User123 banned for spam |
| `unban_user` | User ban revoked | User123 unbanned |
| `add_moderator` | Moderator role granted | User456 added as moderator |
| `remove_moderator` | Moderator role revoked | User456 removed as moderator |
| `sync_bans` | Bans synced from Twitch | 15 bans synced |
| `twitch_ban_user` | User banned on Twitch | User789 banned on Twitch |
| `twitch_unban_user` | User unbanned on Twitch | User789 unbanned on Twitch |
| `update_permissions` | Moderator permissions changed | User456 permissions updated |
| `moderate_content` | Content moderated | Clip removed |

### Log Entry Structure

```json
{
  "id": "log-abc123",
  "actor_id": "user-def456",
  "actor_username": "admin_user",
  "action": "ban_user",
  "resource_type": "user",
  "resource_id": "user-ghi789",
  "channel_id": "channel-jkl012",
  "details": {
    "reason": "Spam",
    "duration": null,
    "is_permanent": true
  },
  "ip_address": "192.168.1.1",
  "user_agent": "Mozilla/5.0...",
  "created_at": "2026-02-03T10:00:00Z"
}
```

### Retention Policy

- **Active logs**: Stored in PostgreSQL
- **Archived logs**: Moved to S3 after 90 days
- **Retention period**: 7 years (compliance requirement)
- **Deletion**: Never (except by legal requirement)

---

## Reviewing Audit Logs

### View Recent Logs

**Use case**: Quick check of recent moderation activity

```bash
# Set environment
export API_TOKEN="your_jwt_token"
export API_BASE="https://api.clpr.tv/api/v1/moderation"

# Get last 20 audit logs
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/audit-logs?limit=20" | jq '.logs[] | {
    time: .created_at,
    actor: .actor_username,
    action: .action,
    resource: .resource_type,
    details: .details.reason
  }'
```

**Example output**:
```json
{
  "time": "2026-02-03T10:15:00Z",
  "actor": "admin_jane",
  "action": "ban_user",
  "resource": "user",
  "details": "Violation of TOS"
}
{
  "time": "2026-02-03T10:10:00Z",
  "actor": "mod_john",
  "action": "moderate_content",
  "resource": "clip",
  "details": "Inappropriate content"
}
```

---

### Filter by Action Type

**Use case**: Review specific type of actions (e.g., all bans, all moderator changes)

```bash
# View all ban actions
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/audit-logs?action=ban_user&limit=50" | jq

# View all moderator additions
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/audit-logs?action=add_moderator&limit=50" | jq

# View all Twitch sync operations
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/audit-logs?action=sync_bans&limit=50" | jq
```

#### Multiple Action Types

```bash
# View bans and unbans
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/audit-logs?action=ban_user,unban_user&limit=50" | jq
```

---

### Filter by User/Actor

**Use case**: Investigate actions by specific user or moderator

```bash
# View all actions by specific actor
ACTOR_ID="user-abc123"
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/audit-logs?actor_id=$ACTOR_ID&limit=100" | jq

# View all actions affecting specific user
RESOURCE_ID="user-def456"
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/audit-logs?resource_id=$RESOURCE_ID&limit=100" | jq
```

#### By Username

```bash
# Get actor ID from username first
ACTOR_USERNAME="suspicious_mod"
ACTOR_ID=$(curl -s -H "Authorization: Bearer $API_TOKEN" \
  "https://api.clpr.tv/api/v1/users/by-username/$ACTOR_USERNAME" | jq -r '.id')

# Then get their actions
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/audit-logs?actor_id=$ACTOR_ID&limit=100" | jq
```

---

### Search by Time Range

**Use case**: Review actions during specific period (incident investigation, compliance audit)

```bash
# Last 24 hours
START_TIME=$(date -u -d '24 hours ago' '+%Y-%m-%dT%H:%M:%SZ')
END_TIME=$(date -u '+%Y-%m-%dT%H:%M:%SZ')

curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/audit-logs?start_time=$START_TIME&end_time=$END_TIME&limit=1000" | jq

# Specific date range
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/audit-logs?start_time=2026-02-01T00:00:00Z&end_time=2026-02-02T00:00:00Z&limit=1000" | jq
```

#### Pagination for Large Results

```bash
# First page
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/audit-logs?start_time=$START_TIME&limit=100&offset=0" | jq > page1.json

# Second page
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/audit-logs?start_time=$START_TIME&limit=100&offset=100" | jq > page2.json

# Continue until no more results
```

---

## Exporting Audit Logs

### Export to CSV

**Use case**: Compliance reports, spreadsheet analysis

```bash
# Export audit logs to CSV
curl -s -H "Authorization: Bearer $API_TOKEN" \
  -H "Accept: text/csv" \
  "$API_BASE/audit-logs/export?start_time=$START_TIME&end_time=$END_TIME" \
  -o "audit-logs-$(date +%Y%m%d).csv"
```

**CSV Format**:
```csv
timestamp,actor_username,actor_id,action,resource_type,resource_id,channel_id,reason,ip_address
2026-02-03T10:00:00Z,admin_jane,user-abc123,ban_user,user,user-def456,channel-ghi789,Spam,192.168.1.1
2026-02-03T10:05:00Z,mod_john,user-jkl012,unban_user,user,user-def456,channel-ghi789,Appeal approved,192.168.1.2
```

---

### Export to JSON

**Use case**: Programmatic processing, archival

```bash
# Export to JSON file
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/audit-logs?start_time=$START_TIME&end_time=$END_TIME&limit=10000" \
  | jq > "audit-logs-$(date +%Y%m%d).json"
```

#### Compressed Export for Large Datasets

```bash
# Export and compress
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/audit-logs?start_time=$START_TIME&end_time=$END_TIME&limit=50000" \
  | gzip > "audit-logs-$(date +%Y%m%d).json.gz"
```

---

### Scheduled Exports

**Use case**: Automated compliance reporting, backups

#### Daily Export Script

```bash
#!/bin/bash
# /opt/clpr/scripts/daily-audit-export.sh

set -euo pipefail

API_TOKEN="${API_TOKEN}"
API_BASE="https://api.clpr.tv/api/v1/moderation"
EXPORT_DIR="/var/log/clpr/audit-exports"

# Create export directory if it doesn't exist
mkdir -p "$EXPORT_DIR"

# Calculate date range (yesterday)
START_TIME=$(date -u -d 'yesterday 00:00:00' '+%Y-%m-%dT%H:%M:%SZ')
END_TIME=$(date -u -d 'yesterday 23:59:59' '+%Y-%m-%dT%H:%M:%SZ')
DATE_LABEL=$(date -u -d 'yesterday' '+%Y-%m-%d')

# Export to CSV
echo "Exporting audit logs for $DATE_LABEL..."
curl -s -H "Authorization: Bearer $API_TOKEN" \
  -H "Accept: text/csv" \
  "$API_BASE/audit-logs/export?start_time=$START_TIME&end_time=$END_TIME" \
  -o "$EXPORT_DIR/audit-logs-$DATE_LABEL.csv"

# Export to JSON (compressed)
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/audit-logs?start_time=$START_TIME&end_time=$END_TIME&limit=50000" \
  | gzip > "$EXPORT_DIR/audit-logs-$DATE_LABEL.json.gz"

# Calculate checksums
cd "$EXPORT_DIR"
sha256sum "audit-logs-$DATE_LABEL.csv" > "audit-logs-$DATE_LABEL.csv.sha256"
sha256sum "audit-logs-$DATE_LABEL.json.gz" > "audit-logs-$DATE_LABEL.json.gz.sha256"

# Upload to S3 (if configured)
if [ -n "${S3_BUCKET:-}" ]; then
  aws s3 cp "$EXPORT_DIR/audit-logs-$DATE_LABEL.csv" \
    "s3://$S3_BUCKET/audit-logs/$DATE_LABEL/"
  aws s3 cp "$EXPORT_DIR/audit-logs-$DATE_LABEL.json.gz" \
    "s3://$S3_BUCKET/audit-logs/$DATE_LABEL/"
fi

# Clean up local files older than 30 days
find "$EXPORT_DIR" -name "audit-logs-*.csv" -mtime +30 -delete
find "$EXPORT_DIR" -name "audit-logs-*.json.gz" -mtime +30 -delete

echo "Export complete: $EXPORT_DIR/audit-logs-$DATE_LABEL.*"
```

#### Setup Cron Job

```bash
# Add to crontab
crontab -e

# Run daily at 2 AM UTC
0 2 * * * /opt/clpr/scripts/daily-audit-export.sh >> /var/log/clpr/audit-export.log 2>&1
```

---

## Common Patterns to Look For

### Unauthorized Access Attempts

**Red flags**:
- Multiple failed permission checks from same user
- Access attempts outside normal hours
- Geographic anomalies (IP from unexpected location)

```bash
# Find failed authorization attempts
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/audit-logs?action=permission_denied&limit=100" | \
  jq '.logs[] | {actor: .actor_username, time: .created_at, ip: .ip_address}'
```

**Actions to take**:
1. Review user account for compromise
2. Check if account credentials leaked
3. Consider temporary account suspension
4. Enable MFA if not already enabled
5. Reset password

---

### Suspicious Ban/Unban Activity

**Red flags**:
- Rapid ban/unban cycles (same user)
- Unusual ban reason patterns
- Bans during off-hours
- Mass bans without documented incident

```bash
# Find ban/unban pairs for same user
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/audit-logs?action=ban_user,unban_user&limit=500" | \
  jq 'group_by(.resource_id) | .[] | select(length > 1) | {
    user_id: .[0].resource_id,
    actions: [.[] | {action: .action, time: .created_at, actor: .actor_username}]
  }'
```

**Investigation steps**:
1. Review ban reasons
2. Check if bans were legitimate
3. Interview moderators involved
4. Look for pattern of abuse
5. Review user appeals

---

### Moderator Abuse

**Red flags**:
- Moderator adds/removes other moderators frequently
- Moderator actions outside their scope
- Unusual volume of bans from single moderator
- Permission escalation attempts

```bash
# Moderators who added other moderators
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/audit-logs?action=add_moderator&limit=100" | \
  jq '.logs[] | {actor: .actor_username, added: .details.username, time: .created_at}'

# Check moderator's action volume
MODERATOR_ID="user-abc123"
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/audit-logs?actor_id=$MODERATOR_ID&limit=1000" | \
  jq 'group_by(.action) | .[] | {action: .[0].action, count: length}'
```

**Actions to take**:
1. Suspend moderator privileges immediately (see [Moderation Operations](./moderation-operations.md#emergency-revoke-moderator))
2. Review all actions by moderator
3. Reverse unauthorized actions
4. Interview moderator
5. Document incident

---

### Mass Operations

**Red flags**:
- Large number of bans in short time
- Bulk moderator additions
- Mass content removals

```bash
# Find potential mass operations
START_TIME=$(date -u -d '1 hour ago' '+%Y-%m-%dT%H:%M:%SZ')
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/audit-logs?start_time=$START_TIME&limit=1000" | \
  jq 'group_by(.actor_id) | .[] | select(length > 20) | {
    actor: .[0].actor_username,
    action_count: length,
    actions: group_by(.action) | map({action: .[0].action, count: length})
  }'
```

**Verification steps**:
1. Confirm mass operation was authorized
2. Check for incident ticket/approval
3. Review operation logs
4. Verify results are correct
5. Document in incident log

---

## Investigation Procedures

### Full User Action History

**Use case**: Investigate user complaint, compliance request, security incident

```bash
#!/bin/bash
# user-action-history.sh

USERNAME="${1:-}"
if [ -z "$USERNAME" ]; then
  echo "Usage: $0 <username>"
  exit 1
fi

API_TOKEN="${API_TOKEN}"
API_BASE="https://api.clpr.tv/api/v1/moderation"

# Get user ID
USER_ID=$(curl -s -H "Authorization: Bearer $API_TOKEN" \
  "https://api.clpr.tv/api/v1/users/by-username/$USERNAME" | jq -r '.id')

echo "User: $USERNAME (ID: $USER_ID)"
echo "==================================="
echo

# Actions BY user (as actor)
echo "Actions performed by user:"
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/audit-logs?actor_id=$USER_ID&limit=500" | \
  jq -r '.logs[] | "\(.created_at) | \(.action) | \(.details.reason // "N/A")"'

echo
echo "==================================="
echo

# Actions ON user (as resource)
echo "Actions performed on user:"
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/audit-logs?resource_id=$USER_ID&limit=500" | \
  jq -r '.logs[] | "\(.created_at) | \(.action) by \(.actor_username) | \(.details.reason // "N/A")"'
```

**Usage**:
```bash
chmod +x user-action-history.sh
./user-action-history.sh suspicious_user > investigation-report-$(date +%Y%m%d).txt
```

---

### Time-Based Analysis

**Use case**: Identify patterns, find anomalies

```bash
# Actions per hour for last 24 hours
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/audit-logs?start_time=$(date -u -d '24 hours ago' '+%Y-%m-%dT%H:%M:%SZ')&limit=5000" | \
  jq -r '.logs[] | .created_at' | \
  cut -d'T' -f2 | cut -d':' -f1 | \
  sort | uniq -c | \
  awk '{printf "%02d:00 - %s actions\n", $2, $1}'
```

---

### Geolocation Analysis

**Use case**: Detect suspicious access from unexpected locations

```bash
# Extract unique IP addresses from audit logs
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/audit-logs?limit=1000" | \
  jq -r '.logs[] | .ip_address' | \
  sort | uniq -c | sort -nr

# For detailed IP analysis, use external service
# Example: ipinfo.io (requires API key)
IP_ADDRESS="192.168.1.1"
curl -s "https://ipinfo.io/$IP_ADDRESS?token=YOUR_TOKEN" | jq
```

---

## Compliance and Retention

### Regulatory Requirements

**GDPR Compliance**:
- Right to access: Export user's audit log entries
- Right to erasure: Requires legal review (audit logs may be retained for compliance)
- Data minimization: Only essential fields logged

**SOC 2 Compliance**:
- Complete audit trail of all system changes
- Retention period: Minimum 1 year (Clipper retains 7 years)
- Regular review: Quarterly audit log reviews

### Data Subject Access Request (DSAR)

**Use case**: User requests their data under GDPR

```bash
#!/bin/bash
# dsar-audit-export.sh

USER_EMAIL="${1:-}"
if [ -z "$USER_EMAIL" ]; then
  echo "Usage: $0 <user_email>"
  exit 1
fi

# Get user ID from email
USER_ID=$(curl -s -H "Authorization: Bearer $API_TOKEN" \
  "https://api.clpr.tv/api/v1/users?email=$USER_EMAIL" | jq -r '.[0].id')

# Export all audit logs where user is actor or resource
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/audit-logs?actor_id=$USER_ID&limit=10000" \
  > "dsar-$USER_ID-actor.json"

curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/audit-logs?resource_id=$USER_ID&limit=10000" \
  > "dsar-$USER_ID-resource.json"

# Combine and anonymize IPs (GDPR requirement)
jq -s '.[0].logs + .[1].logs | unique_by(.id) | map(del(.ip_address))' \
  "dsar-$USER_ID-actor.json" "dsar-$USER_ID-resource.json" \
  > "dsar-$USER_ID-final.json"

echo "DSAR export complete: dsar-$USER_ID-final.json"
```

---

### Archival to Long-Term Storage

**Use case**: Move old logs to S3 for cost savings and compliance

```bash
#!/bin/bash
# archive-old-audit-logs.sh

# Archive logs older than 90 days to S3
CUTOFF_DATE=$(date -u -d '90 days ago' '+%Y-%m-%d')
START_TIME="2020-01-01T00:00:00Z"
END_TIME="${CUTOFF_DATE}T23:59:59Z"

# Export from database
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/audit-logs?start_time=$START_TIME&end_time=$END_TIME&limit=100000" \
  | gzip > "archive-${CUTOFF_DATE}.json.gz"

# Upload to S3
aws s3 cp "archive-${CUTOFF_DATE}.json.gz" \
  "s3://clpr-audit-archives/$(date +%Y)/" \
  --storage-class GLACIER

# Mark as archived in database (if API supports)
# curl -X POST "$API_BASE/audit-logs/archive" ...

echo "Archived logs up to $CUTOFF_DATE to S3 Glacier"
```

---

## Related Runbooks

- [Moderation Operations](./moderation-operations.md) - Emergency ban/unban procedures
- [Ban Sync Troubleshooting](./ban-sync-troubleshooting.md) - Twitch sync issues
- [Moderation Incidents](./moderation-incidents.md) - Incident response
- [Permission Escalation](./permission-escalation.md) - Grant emergency access

---

## Troubleshooting

### Empty Results

**Issue**: Query returns no logs

```bash
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/audit-logs?action=ban_user&limit=100" | jq '.logs | length'
# Returns: 0
```

**Possible causes**:
1. No actions of that type have occurred
2. Date range is incorrect
3. Insufficient permissions
4. Database connectivity issue

**Debug steps**:
```bash
# Check if ANY logs exist
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/audit-logs?limit=1" | jq

# Check API health
curl -s "$API_BASE/health" | jq
```

---

### Performance Issues

**Issue**: Queries are slow (> 5 seconds)

**Optimization strategies**:

1. **Reduce time range**
   ```bash
   # Instead of querying 30 days
   # curl "$API_BASE/audit-logs?start_time=2026-01-01T00:00:00Z&end_time=2026-01-31T23:59:59Z"
   
   # Query by day
   for day in {01..31}; do
     curl "$API_BASE/audit-logs?start_time=2026-01-${day}T00:00:00Z&end_time=2026-01-${day}T23:59:59Z&limit=1000"
   done
   ```

2. **Use pagination**
   ```bash
   # Fetch in smaller chunks
   for offset in {0..1000..100}; do
     curl "$API_BASE/audit-logs?limit=100&offset=$offset"
   done
   ```

3. **Add specific filters**
   ```bash
   # Instead of fetching all logs
   # Add action type, actor_id, or resource_id filters
   curl "$API_BASE/audit-logs?action=ban_user&actor_id=$ACTOR_ID"
   ```

---

**Last Updated**: 2026-02-03  
**Document Owner**: Operations Team  
**Review Frequency**: Quarterly
