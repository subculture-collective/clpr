---
title: "Moderation Operations Runbook"
summary: "Operational procedures for managing the moderation system"
tags: ["operations", "runbook", "moderation", "emergency"]
area: "moderation"
status: "active"
owner: "team-ops"
version: "1.0"
last_reviewed: 2026-11-14
aliases: ["mod-ops", "moderation-runbook"]
---

# Moderation Operations Runbook

## Overview

This runbook provides operational procedures for managing the Clipper moderation system, including emergency procedures, manual operations, and routine maintenance tasks.

**Audience**: Operations team, on-call engineers, site administrators

**Prerequisites**:
- Access to production environment
- A valid JWT for an account that may moderate the channel: an admin, a platform moderator (role `moderator`), the community owner, or a community mod or admin
- The community ID (`CHANNEL_ID`). In this API a "channel" is a community; find its ID with `GET /api/v1/communities/search?q=<name>`
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

All commands use these variables:

```bash
export API_TOKEN="your_jwt_token"
export API_ROOT="https://clpr.tv/api/v1"
export API_BASE="$API_ROOT/moderation"
export CHANNEL_ID="<community uuid>"
```

Request bodies use camelCase field names: `channelId`, `userId`, `reason`. Unknown fields such as `user_id`, `expires_at` or `is_permanent` are ignored.

Bans are per community and permanent until revoked. The API has no temporary bans and no site-wide ban on these routes.

### Emergency Ban User

**When to use**: Immediate threat, spam attack, abuse, or security incident

**Time to execute**: < 2 minutes

**Prerequisites**:
- User ID or username
- Reason for ban (at most 1000 characters)
- Rights to moderate the community (see Prerequisites above)

#### Steps

1. **Get user ID (if only username is known)**
   ```bash
   USER_ID=$(curl -s -H "Authorization: Bearer $API_TOKEN" \
     "$API_ROOT/users/by-username/${USERNAME}" | jq -r '.data.id')
   ```

2. **Create the ban**
   ```bash
   curl -s -X POST "$API_BASE/ban" \
     -H "Authorization: Bearer $API_TOKEN" \
     -H "Content-Type: application/json" \
     -d "{
       \"channelId\": \"$CHANNEL_ID\",
       \"userId\": \"$USER_ID\",
       \"reason\": \"EMERGENCY: [describe threat]\"
     }" | tee /tmp/ban.json | jq
   BAN_ID=$(jq -r '.id' /tmp/ban.json)
   ```

   A `201` response returns `id`, `channelId`, `userId`, `bannedBy`, `reason` and `bannedAt`. Other responses:
   - `400`: the target is the community owner, or you tried to ban yourself
   - `403`: you may not moderate this community, or the target is a community mod or admin (only the owner, an admin or a platform moderator can ban those)
   - `404`: the community or the user does not exist

   Banning a user who is already banned updates the existing ban and returns the same ban ID.

3. **Verify ban was applied**
   ```bash
   curl -s -H "Authorization: Bearer $API_TOKEN" \
     "$API_BASE/ban/$BAN_ID" | jq
   ```

4. **Document the action**
   - Note the ban ID returned
   - Document in incident tracker

#### Rollback

To reverse an emergency ban:

```bash
curl -s -X DELETE "$API_BASE/ban/$BAN_ID" \
  -H "Authorization: Bearer $API_TOKEN" | jq
```

#### Verification

- The user is removed from the community and cannot rejoin it
- `GET $API_BASE/ban/$BAN_ID` returns the ban
- A `ban_user` entry appears in the audit log (see [Audit Log Operations](./audit-log-operations.md))

---

### Emergency Revoke Moderator

**When to use**: Moderator abuse, compromised account, security incident

**Time to execute**: < 1 minute

This removes a community moderator. To revoke a platform moderator (role `moderator`), use the Platform Moderators page in the admin panel.

#### Steps

1. **List moderators to find target**
   ```bash
   curl -s -H "Authorization: Bearer $API_TOKEN" \
     "$API_BASE/moderators?channelId=$CHANNEL_ID" | jq '.moderators[] | {id, user_id, role, joined_at}'
   ```

   The list holds membership rows. `id` is the membership ID used in the next step; `user_id` is the account.

2. **Remove moderator immediately**
   ```bash
   MODERATOR_ID="<id from the list>"
   curl -s -X DELETE "$API_BASE/moderators/$MODERATOR_ID" \
     -H "Authorization: Bearer $API_TOKEN" | jq
   ```

   The user stays in the community as an ordinary member. The community owner cannot be removed.

3. **Verify revocation**
   ```bash
   # Check they no longer appear in moderator list
   curl -s -H "Authorization: Bearer $API_TOKEN" \
     "$API_BASE/moderators?channelId=$CHANNEL_ID" | jq '.moderators[] | select(.id == "'$MODERATOR_ID'")'
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
- An admin token. `POST /moderation/ban` allows 10 requests per hour for everyone else; admins bypass the limit

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
   API_BASE="${API_BASE:-https://clpr.tv/api/v1/moderation}"
   CHANNEL_ID="${CHANNEL_ID}"
   REASON="${REASON:-Mass ban - Security incident}"
   
   if [ -z "$API_TOKEN" ] || [ -z "$CHANNEL_ID" ]; then
     echo "Error: API_TOKEN and CHANNEL_ID must be set"
     exit 1
   fi
   
   while IFS= read -r user_id; do
     echo "Banning user: $user_id"
     response=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE/ban" \
       -H "Authorization: Bearer $API_TOKEN" \
       -H "Content-Type: application/json" \
       -d "{
         \"channelId\": \"$CHANNEL_ID\",
         \"userId\": \"$user_id\",
         \"reason\": \"$REASON\"
       }")
     
     http_code=$(echo "$response" | tail -n1)
     body=$(echo "$response" | sed '$d')
     
     if [ "$http_code" = "201" ]; then
       ban_id=$(echo "$body" | jq -r '.id')
       echo "  ✓ Success - Ban ID: $ban_id"
     else
       echo "  ✗ Failed - HTTP $http_code: $body"
     fi
     
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
  curl -s -X DELETE "$API_BASE/ban/$ban_id" \
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
     "$API_ROOT/users/by-username/${USERNAME}" | jq '.data'
   
   # By user ID
   curl -s -H "Authorization: Bearer $API_TOKEN" \
     "$API_ROOT/users/${USER_ID}" | jq
   ```

2. **Create the ban and save its ID**
   ```bash
   BAN_ID=$(curl -s -X POST "$API_BASE/ban" \
     -H "Authorization: Bearer $API_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"channelId":"'$CHANNEL_ID'","userId":"'$USER_ID'","reason":"Violation of community guidelines"}' \
     | jq -r '.id')
   echo "Ban ID: $BAN_ID"
   ```

   There is no temporary ban. To suspend someone for a set time, ban them and revoke the ban by hand when the time is up.

#### Unban User

1. **Find ban ID**
   ```bash
   # The list has no user filter, so select the user from the community's bans
   curl -s -H "Authorization: Bearer $API_TOKEN" \
     "$API_BASE/bans?channelId=$CHANNEL_ID&limit=100" | \
     jq '.bans[] | select(.banned_user_id == "'$USER_ID'")'
   ```

2. **Revoke ban**
   ```bash
   BAN_ID="<ban_id>"
   curl -s -X DELETE "$API_BASE/ban/$BAN_ID" \
     -H "Authorization: Bearer $API_TOKEN" | jq
   ```

3. **Verify unbanned**
   ```bash
   curl -s -o /dev/null -w "%{http_code}\n" -H "Authorization: Bearer $API_TOKEN" \
     "$API_BASE/ban/$BAN_ID"
   # Should print 404
   ```

---

### Manual Add/Remove Moderator

#### Add Moderator

**Use case**: Granting moderation permissions to a user in one community

Admins, platform moderators, the community owner and community admins can do this.

1. **Verify user exists**
   ```bash
   curl -s -H "Authorization: Bearer $API_TOKEN" \
     "$API_ROOT/users/by-username/${USERNAME}" | jq '.data | {id, username}'
   ```

2. **Add as moderator**
   ```bash
   curl -s -X POST "$API_BASE/moderators" \
     -H "Authorization: Bearer $API_TOKEN" \
     -H "Content-Type: application/json" \
     -d "{
       \"userId\": \"$USER_ID\",
       \"channelId\": \"$CHANNEL_ID\",
       \"reason\": \"Why this user is being added\"
     }" | jq
   ```

   The user gets the community role `mod`. A user who is banned from the community is refused with `409`; revoke the ban first.

3. **Verify moderator was added**
   ```bash
   curl -s -H "Authorization: Bearer $API_TOKEN" \
     "$API_BASE/moderators?channelId=$CHANNEL_ID" | jq '.moderators[] | select(.user_id == "'$USER_ID'")'
   ```

#### Change a Moderator's Role

Community roles are `mod` and `admin`. There are no per-permission settings.

```bash
curl -s -X PATCH "$API_BASE/moderators/$MODERATOR_ID" \
  -H "Authorization: Bearer $API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"role": "admin"}' | jq
```

#### Remove Moderator

1. **Get moderator ID**
   ```bash
   MODERATOR_ID=$(curl -s -H "Authorization: Bearer $API_TOKEN" \
     "$API_BASE/moderators?channelId=$CHANNEL_ID" | \
     jq -r '.moderators[] | select(.user_id == "'$USER_ID'") | .id')
   ```

2. **Remove moderator**
   ```bash
   curl -s -X DELETE "$API_BASE/moderators/$MODERATOR_ID" \
     -H "Authorization: Bearer $API_TOKEN" | jq
   ```

3. **Verify removal**
   ```bash
   curl -s -H "Authorization: Bearer $API_TOKEN" \
     "$API_BASE/moderators?channelId=$CHANNEL_ID" | \
     jq '.moderators[] | select(.user_id == "'$USER_ID'")'
   # Should return empty
   ```

---

### Check User Ban Status

**Use case**: Investigate user complaints, verify ban state

The ban list is per community and has no user filter. The parameter must be spelled `channelId`. Without it the request lists bans across all communities, which only admins and platform moderators may do, so never read a result from a request that left it out as an answer about one community.

```bash
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/bans?channelId=$CHANNEL_ID&limit=100" | \
  jq '[.bans[] | select(.banned_user_id == "'$USER_ID'")] | {
    is_banned: (length > 0),
    bans: map({id, reason, banned_at, banned_by_user_id})
  }'
```

The list returns at most 100 bans per request. If `total` in the response is larger, page with `offset` (a multiple of `limit`) until you have checked them all.

#### Interpret Results

```json
{
  "is_banned": true,
  "bans": [
    {
      "id": "0b9c2f0e-6f0a-4f0e-9d53-0d2a8a1f4c11",
      "reason": "Spam",
      "banned_at": "2026-02-01T10:00:00Z",
      "banned_by_user_id": "5a1d0c52-9a1b-4d6e-8f2c-3e7b6a9d1f20"
    }
  ]
}
```

- `is_banned: true` → The user is banned from this community until the ban is revoked
- `is_banned: false` → No ban in the bans checked; make sure you paged through all of them

---

### List All Moderators

**Use case**: Audit moderator list

```bash
# List all moderators for a community (at most 100 per request; page with offset)
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/moderators?channelId=$CHANNEL_ID&limit=100" | \
  jq '.moderators[] | {id, user_id, role, joined_at}'
```

#### Filter by Role

```bash
# Find community admins
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/moderators?channelId=$CHANNEL_ID&limit=100" | \
  jq -r '.moderators[] | select(.role == "admin") | .user_id'
```

---

## Routine Operations

### Daily Health Checks

**Frequency**: Daily, automated or manual

**Time required**: 5 minutes

#### Checklist

- [ ] Check API health (the moderation routes have no health endpoint of their own)
  ```bash
  curl -s "https://clpr.tv/health/ready" | jq
  ```

- [ ] Verify recent audit logs are being created (admins and platform moderators only)
  ```bash
  curl -s -H "Authorization: Bearer $API_TOKEN" \
    "$API_BASE/audit-logs?limit=10" | jq '.logs[0]'
  ```

- [ ] Check the review queue is not backing up (admins and platform moderators only)
  ```bash
  curl -s -H "Authorization: Bearer $API_TOKEN" \
    "$API_ROOT/admin/moderation/queue/stats" | jq '.data | {total_pending, high_priority_count, oldest_pending_age_hours}'
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
1. Verify you are an admin, a platform moderator, the community owner, or a mod or admin of that community
2. Check `channelId` is the right community
3. See [Permission Escalation](./permission-escalation.md) for emergency access

#### 429 Rate Limit Exceeded

```
Error: Rate limit exceeded
```

**Cause**: Too many requests in time window

**Solution**:
1. Check `X-RateLimit-Reset` header for reset time
2. Ban, revoke and moderator changes allow 10 requests per hour per account; list requests allow 60 per minute
3. For an emergency, use an admin account, which bypasses these limits

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

**Last Updated**: 2026-11-14  
**Document Owner**: Operations Team  
**Review Frequency**: Quarterly
