---
title: "Moderation Operations Runbook"
summary: "Operational procedures for managing the moderation system"
tags: ["operations", "runbook", "moderation", "emergency"]
area: "moderation"
status: "active"
owner: "team-ops"
version: "1.0"
last_reviewed: 2026-02-03
aliases: ["mod-ops", "moderation-runbook"]
---

# Moderation Operations Runbook

## Overview

This runbook provides operational procedures for managing the Clipper moderation system, including emergency procedures, manual operations, and routine maintenance tasks.

**Audience**: Operations team, on-call engineers, site administrators

**Prerequisites**:
- Access to production environment
- Admin or super-admin role
- Valid JWT authentication token
- Familiarity with Moderation API ([docs/backend/moderation-api.md](../../backend/moderation-api.md))

## Table of Contents

- [Emergency Procedures](#emergency-procedures)
  - [Emergency Ban User](#emergency-ban-user)
  - [Emergency Revoke Moderator](#emergency-revoke-moderator)
  - [Mass Ban (Security Incident)](#mass-ban-security-incident)
- [Manual Operations](#manual-operations)
  - [Manual Ban/Unban User](#manual-banunban-user)
  - [Manual Add/Remove Moderator](#manual-addremove-moderator)
  - [Check User Ban Status](#check-user-ban-status)
  - [List All Moderators](#list-all-moderators)
- [Routine Operations](#routine-operations)
  - [Daily Health Checks](#daily-health-checks)
  - [Weekly Audit Review](#weekly-audit-review)
- [Related Runbooks](#related-runbooks)

---

## Emergency Procedures

### Emergency Ban User

**When to use**: Immediate threat, spam attack, abuse, or security incident

**Time to execute**: < 2 minutes

**Prerequisites**:
- User ID or username
- Reason for ban
- Admin or moderator permissions

#### Steps

1. **Obtain authentication token**
   ```bash
   # If not already authenticated
   export API_TOKEN="your_jwt_token"
   export API_BASE="https://api.clpr.tv/api/v1/moderation"
   ```

2. **Get user ID (if only username is known)**
   ```bash
   USER_ID=$(curl -s -H "Authorization: Bearer $API_TOKEN" \
     "https://api.clpr.tv/api/v1/users/by-username/${USERNAME}" | jq -r '.id')
   ```

3. **Create permanent ban**
   ```bash
   curl -X POST "$API_BASE/bans" \
     -H "Authorization: Bearer $API_TOKEN" \
     -H "Content-Type: application/json" \
     -d "{
       \"user_id\": \"$USER_ID\",
       \"channel_id\": \"$CHANNEL_ID\",
       \"reason\": \"EMERGENCY: [describe threat]\",
       \"expires_at\": null,
       \"is_permanent\": true
     }"
   ```

4. **Verify ban was applied**
   ```bash
   curl -s -H "Authorization: Bearer $API_TOKEN" \
     "$API_BASE/bans?channel_id=$CHANNEL_ID&user_id=$USER_ID" | jq
   ```

5. **Document the action**
   - Note the ban ID returned
   - Document in incident tracker
   - Create audit trail note

#### Rollback

To reverse an emergency ban:

```bash
BAN_ID="<ban_id_from_creation>"
curl -X DELETE "$API_BASE/bans/$BAN_ID" \
  -H "Authorization: Bearer $API_TOKEN"
```

#### Verification

- User cannot post content
- User cannot submit clips
- User sees "banned" status in profile
- Action logged in audit logs (see [Audit Log Operations](./audit-log-operations.md))

---

### Emergency Revoke Moderator

**When to use**: Moderator abuse, compromised account, security incident

**Time to execute**: < 1 minute

#### Steps

1. **List moderators to find target**
   ```bash
   curl -s -H "Authorization: Bearer $API_TOKEN" \
     "$API_BASE/moderators?channel_id=$CHANNEL_ID" | jq '.moderators[] | {id, user_id, username}'
   ```

2. **Remove moderator immediately**
   ```bash
   MODERATOR_ID="<moderator_id>"
   curl -X DELETE "$API_BASE/moderators/$MODERATOR_ID" \
     -H "Authorization: Bearer $API_TOKEN"
   ```

3. **Verify revocation**
   ```bash
   # Check they no longer appear in moderator list
   curl -s -H "Authorization: Bearer $API_TOKEN" \
     "$API_BASE/moderators?channel_id=$CHANNEL_ID" | jq '.moderators[] | select(.id == "'$MODERATOR_ID'")'
   # Should return empty
   ```

4. **Optional: Ban if necessary**
   - If account is compromised or abusive, follow [Emergency Ban User](#emergency-ban-user)

#### Post-Action

- [ ] Notify security team
- [ ] Review audit logs for suspicious activity
- [ ] Check for any unauthorized actions taken
- [ ] Document in incident report
- [ ] Consider password reset if account compromise suspected

---

### Mass Ban (Security Incident)

**When to use**: Coordinated spam attack, bot network, mass abuse

**Time to execute**: 5-15 minutes depending on volume

**Prerequisites**:
- List of user IDs to ban
- Incident ticket number
- Security team approval

#### Steps

1. **Prepare user list**
   ```bash
   # Create file with user IDs (one per line)
   cat > /tmp/users_to_ban.txt << EOF
   user-id-1
   user-id-2
   user-id-3
   EOF
   ```

2. **Execute mass ban script**
   ```bash
   #!/bin/bash
   # mass-ban.sh
   
   API_TOKEN="${API_TOKEN}"
   API_BASE="${API_BASE:-https://api.clpr.tv/api/v1/moderation}"
   CHANNEL_ID="${CHANNEL_ID}"
   REASON="${REASON:-Mass ban - Security incident}"
   
   if [ -z "$API_TOKEN" ] || [ -z "$CHANNEL_ID" ]; then
     echo "Error: API_TOKEN and CHANNEL_ID must be set"
     exit 1
   fi
   
   while IFS= read -r user_id; do
     echo "Banning user: $user_id"
     response=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE/bans" \
       -H "Authorization: Bearer $API_TOKEN" \
       -H "Content-Type: application/json" \
       -d "{
         \"user_id\": \"$user_id\",
         \"channel_id\": \"$CHANNEL_ID\",
         \"reason\": \"$REASON\",
         \"is_permanent\": true
       }")
     
     http_code=$(echo "$response" | tail -n1)
     body=$(echo "$response" | sed '$d')
     
     if [ "$http_code" = "201" ]; then
       ban_id=$(echo "$body" | jq -r '.id')
       echo "  ✓ Success - Ban ID: $ban_id"
     else
       echo "  ✗ Failed - HTTP $http_code: $body"
     fi
     
     # Rate limit protection
     sleep 0.5
   done < /tmp/users_to_ban.txt
   
   echo "Mass ban complete"
   ```

3. **Execute with logging**
   ```bash
   chmod +x mass-ban.sh
   ./mass-ban.sh 2>&1 | tee /tmp/mass-ban-$(date +%Y%m%d-%H%M%S).log
   ```

4. **Verify results**
   ```bash
   # Check ban count
   TOTAL_USERS=$(wc -l < /tmp/users_to_ban.txt)
   BANNED_COUNT=$(grep "✓ Success" /tmp/mass-ban-*.log | wc -l)
   echo "Banned $BANNED_COUNT of $TOTAL_USERS users"
   ```

5. **Document incident**
   - Save log file: `/tmp/mass-ban-YYYYMMDD-HHMMSS.log`
   - Create incident report
   - Update security team
   - Review audit logs

#### Rollback

If mass ban was issued in error:

```bash
# Extract ban IDs from log
grep "Ban ID:" /tmp/mass-ban-*.log | awk '{print $NF}' > /tmp/bans_to_revoke.txt

# Revoke all
while IFS= read -r ban_id; do
  echo "Revoking ban: $ban_id"
  curl -X DELETE "$API_BASE/bans/$ban_id" \
    -H "Authorization: Bearer $API_TOKEN"
  sleep 0.5
done < /tmp/bans_to_revoke.txt
```

---

## Manual Operations

### Manual Ban/Unban User

#### Ban User

**Use case**: Single user moderation action

1. **Get user information**
   ```bash
   # By username
   curl -s -H "Authorization: Bearer $API_TOKEN" \
     "https://api.clpr.tv/api/v1/users/by-username/${USERNAME}" | jq
   
   # By user ID
   curl -s -H "Authorization: Bearer $API_TOKEN" \
     "https://api.clpr.tv/api/v1/users/${USER_ID}" | jq
   ```

2. **Create ban with options**

   **Permanent ban:**
   ```bash
   curl -X POST "$API_BASE/bans" \
     -H "Authorization: Bearer $API_TOKEN" \
     -H "Content-Type: application/json" \
     -d "{
       \"user_id\": \"$USER_ID\",
       \"channel_id\": \"$CHANNEL_ID\",
       \"reason\": \"Violation of community guidelines\",
       \"is_permanent\": true
     }"
   ```

   **Temporary ban (expires in 7 days):**
   ```bash
   EXPIRES_AT=$(date -u -d '+7 days' '+%Y-%m-%dT%H:%M:%SZ')
   curl -X POST "$API_BASE/bans" \
     -H "Authorization: Bearer $API_TOKEN" \
     -H "Content-Type: application/json" \
     -d "{
       \"user_id\": \"$USER_ID\",
       \"channel_id\": \"$CHANNEL_ID\",
       \"reason\": \"Temporary suspension\",
       \"expires_at\": \"$EXPIRES_AT\"
     }"
   ```

3. **Save ban ID for future reference**
   ```bash
   BAN_ID=$(curl -s -X POST "$API_BASE/bans" \
     -H "Authorization: Bearer $API_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"user_id":"'$USER_ID'","channel_id":"'$CHANNEL_ID'","reason":"Test","is_permanent":true}' \
     | jq -r '.id')
   echo "Ban ID: $BAN_ID"
   ```

#### Unban User

1. **Find ban ID**
   ```bash
   # List bans for user
   curl -s -H "Authorization: Bearer $API_TOKEN" \
     "$API_BASE/bans?channel_id=$CHANNEL_ID&user_id=$USER_ID" | jq
   ```

2. **Revoke ban**
   ```bash
   BAN_ID="<ban_id>"
   curl -X DELETE "$API_BASE/bans/$BAN_ID" \
     -H "Authorization: Bearer $API_TOKEN"
   ```

3. **Verify unbanned**
   ```bash
   curl -s -H "Authorization: Bearer $API_TOKEN" \
     "$API_BASE/bans?channel_id=$CHANNEL_ID&user_id=$USER_ID" | jq '.bans | length'
   # Should return 0 if no active bans
   ```

---

### Manual Add/Remove Moderator

#### Add Moderator

**Use case**: Granting moderation permissions to a user

1. **Verify user exists**
   ```bash
   curl -s -H "Authorization: Bearer $API_TOKEN" \
     "https://api.clpr.tv/api/v1/users/by-username/${USERNAME}" | jq '{id, username, email}'
   ```

2. **Add as moderator**
   ```bash
   curl -X POST "$API_BASE/moderators" \
     -H "Authorization: Bearer $API_TOKEN" \
     -H "Content-Type: application/json" \
     -d "{
       \"user_id\": \"$USER_ID\",
       \"channel_id\": \"$CHANNEL_ID\",
       \"permissions\": [\"ban_users\", \"view_audit_logs\"],
       \"scope\": \"community\"
     }"
   ```

3. **Verify moderator was added**
   ```bash
   curl -s -H "Authorization: Bearer $API_TOKEN" \
     "$API_BASE/moderators?channel_id=$CHANNEL_ID" | jq '.moderators[] | select(.user_id == "'$USER_ID'")'
   ```

#### Moderator Permission Scopes

- `global`: Can moderate all channels
- `community`: Can moderate specific channel/community only

#### Common Permission Sets

```json
// View-only moderator
{"permissions": ["view_audit_logs"]}

// Standard moderator
{"permissions": ["ban_users", "view_audit_logs", "moderate_content"]}

// Senior moderator
{"permissions": ["ban_users", "view_audit_logs", "moderate_content", "manage_moderators"]}

// Admin-level
{"permissions": ["ban_users", "view_audit_logs", "moderate_content", "manage_moderators", "manage_settings"]}
```

#### Remove Moderator

1. **Get moderator ID**
   ```bash
   MODERATOR_ID=$(curl -s -H "Authorization: Bearer $API_TOKEN" \
     "$API_BASE/moderators?channel_id=$CHANNEL_ID" | \
     jq -r '.moderators[] | select(.user_id == "'$USER_ID'") | .id')
   ```

2. **Remove moderator**
   ```bash
   curl -X DELETE "$API_BASE/moderators/$MODERATOR_ID" \
     -H "Authorization: Bearer $API_TOKEN"
   ```

3. **Verify removal**
   ```bash
   curl -s -H "Authorization: Bearer $API_TOKEN" \
     "$API_BASE/moderators?channel_id=$CHANNEL_ID" | \
     jq '.moderators[] | select(.user_id == "'$USER_ID'")'
   # Should return empty
   ```

---

### Check User Ban Status

**Use case**: Investigate user complaints, verify ban state

```bash
# Check specific user
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/bans?channel_id=$CHANNEL_ID&user_id=$USER_ID" | \
  jq '{
    is_banned: (.bans | length > 0),
    active_bans: [.bans[] | {
      id,
      reason,
      created_at,
      expires_at,
      is_permanent
    }]
  }'
```

#### Interpret Results

```json
{
  "is_banned": true,
  "active_bans": [
    {
      "id": "ban-123",
      "reason": "Spam",
      "created_at": "2026-02-01T10:00:00Z",
      "expires_at": null,
      "is_permanent": true
    }
  ]
}
```

- `is_banned: true` → User is currently banned
- `expires_at: null` → Permanent ban
- `expires_at: "2026-02-08T10:00:00Z"` → Temporary ban, expires on date shown

---

### List All Moderators

**Use case**: Audit moderator list, verify permissions

```bash
# List all moderators for a channel
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/moderators?channel_id=$CHANNEL_ID" | \
  jq '.moderators[] | {
    username: .username,
    permissions: .permissions,
    scope: .scope,
    added_at: .created_at
  }'
```

#### Filter by Permission

```bash
# Find moderators with specific permission
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/moderators?channel_id=$CHANNEL_ID" | \
  jq '.moderators[] | select(.permissions | contains(["manage_moderators"])) | .username'
```

---

## Routine Operations

### Daily Health Checks

**Frequency**: Daily, automated or manual

**Time required**: 5 minutes

#### Checklist

- [ ] Check moderation API health
  ```bash
  curl -s "$API_BASE/health" | jq
  ```

- [ ] Verify recent audit logs are being created
  ```bash
  curl -s -H "Authorization: Bearer $API_TOKEN" \
    "$API_BASE/audit-logs?limit=10" | jq '.logs[0].created_at'
  ```

- [ ] Check for failed ban sync operations
  ```bash
  # See ban-sync-troubleshooting.md
  curl -s -H "Authorization: Bearer $API_TOKEN" \
    "$API_BASE/audit-logs?action=sync_bans&status=failed&limit=10" | jq
  ```

- [ ] Review error rate (should be < 1%)
  ```bash
  # Check metrics endpoint (if available)
  curl -s "$API_BASE/metrics" | grep moderation_error_rate
  ```

---

### Weekly Audit Review

**Frequency**: Weekly

**Time required**: 15-30 minutes

**See**: [Audit Log Operations](./audit-log-operations.md) for detailed procedures

#### Checklist

- [ ] Export last week's audit logs
- [ ] Review for unusual patterns
- [ ] Check for unauthorized moderator additions
- [ ] Verify ban/unban ratios are normal
- [ ] Document any anomalies

---

## Related Runbooks

- [Audit Log Operations](./audit-log-operations.md) - Review and export audit logs
- [Ban Sync Troubleshooting](./ban-sync-troubleshooting.md) - Twitch sync issues
- [Permission Escalation](./permission-escalation.md) - Grant emergency access
- [Moderation Rollback](./moderation-rollback.md) - Rollback procedures
- [Moderation Monitoring](./moderation-monitoring.md) - Metrics and alerts
- [Moderation Incidents](./moderation-incidents.md) - Common issues and solutions

---

## Troubleshooting

### Common Errors

#### 401 Unauthorized

```
Error: Authorization required
```

**Cause**: Missing or expired JWT token

**Solution**:
1. Obtain new token: `POST /api/v1/auth/refresh`
2. Export token: `export API_TOKEN="<new_token>"`

#### 403 Forbidden

```
Error: Permission denied
```

**Cause**: Insufficient permissions for operation

**Solution**:
1. Verify you have admin/moderator role
2. Check channel ownership
3. See [Permission Escalation](./permission-escalation.md) for emergency access

#### 429 Rate Limit Exceeded

```
Error: Rate limit exceeded
```

**Cause**: Too many requests in time window

**Solution**:
1. Check `X-RateLimit-Reset` header for reset time
2. Implement exponential backoff
3. For emergency, contact ops team to temporarily increase limits

---

## Emergency Contacts

### Security Incidents
- **Email**: security@clpr.tv
- **Slack**: #security-incidents
- **On-call**: PagerDuty escalation

### Operations Team
- **Email**: ops@clpr.tv
- **Slack**: #ops-team
- **On-call**: PagerDuty escalation

### Escalation Path
1. On-call engineer
2. Operations team lead
3. Engineering manager
4. VP of Engineering

---

**Last Updated**: 2026-02-03  
**Document Owner**: Operations Team  
**Review Frequency**: Quarterly
