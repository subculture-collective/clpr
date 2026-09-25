---
title: "Permission Escalation Runbook"
summary: "Procedures for granting, revoking, and troubleshooting permissions"
tags: ["operations", "runbook", "permissions", "security", "emergency"]
area: "moderation"
status: "active"
owner: "team-ops"
version: "1.0"
last_reviewed: 2026-02-03
aliases: ["permissions", "access-control"]
---

# Permission Escalation Runbook

## Overview

This runbook provides procedures for managing user permissions in the Clipper moderation system, including emergency access grants, permission troubleshooting, and security best practices.

**Audience**: Operations team, security team, senior engineers

**Prerequisites**:
- Super-admin access
- Valid JWT authentication token
- Approval from engineering manager (for emergency escalation)
- Understanding of role-based access control (RBAC)

## Table of Contents

- [Permission Model](#permission-model)
- [Emergency Access Procedures](#emergency-access-procedures)
  - [Grant Emergency Admin Access](#grant-emergency-admin-access)
  - [Grant Temporary Moderator Access](#grant-temporary-moderator-access)
  - [Revoke Emergency Access](#revoke-emergency-access)
- [Permission Management](#permission-management)
  - [Grant Moderator Permissions](#grant-moderator-permissions)
  - [Update Moderator Permissions](#update-moderator-permissions)
  - [Revoke Moderator Permissions](#revoke-moderator-permissions)
  - [Grant Admin Access](#grant-admin-access)
- [Troubleshooting Permission Issues](#troubleshooting-permission-issues)
  - [Permission Denied Errors](#permission-denied-errors)
  - [Missing Permissions](#missing-permissions)
  - [Permission Conflicts](#permission-conflicts)
- [Audit and Compliance](#audit-and-compliance)
- [Related Runbooks](#related-runbooks)

---

## Permission Model

### User Roles

| Role | Description | Capabilities |
|------|-------------|--------------|
| **User** | Standard user | View content, submit clips, comment |
| **Moderator** | Content moderator | Ban users, moderate content, view audit logs |
| **Admin** | Administrator | All moderator permissions + manage moderators, configure settings |
| **Super Admin** | System administrator | All permissions + manage admins, system configuration |

### Permission Scopes

| Scope | Description | Example |
|-------|-------------|---------|
| **Global** | Access across all channels | Site-wide moderator |
| **Community** | Access to specific channel/community | Channel-specific moderator |

### Permission Actions

```json
{
  "permissions": [
    "view_content",           // View content (all users)
    "submit_content",         // Submit clips (all users)
    "comment",                // Comment on content (all users)
    "ban_users",              // Ban/unban users (moderator+)
    "moderate_content",       // Remove/approve content (moderator+)
    "view_audit_logs",        // View moderation logs (moderator+)
    "manage_moderators",      // Add/remove moderators (admin+)
    "manage_settings",        // Configure system settings (admin+)
    "manage_permissions",     // Grant/revoke permissions (super admin only)
    "sync_twitch_bans",       // Sync bans from Twitch (moderator+)
    "twitch_ban_user",        // Ban on Twitch (broadcaster/mod only)
    "export_audit_logs"       // Export compliance reports (admin+)
  ]
}
```

---

## Emergency Access Procedures

### Grant Emergency Admin Access

**When to use**: Security incident, critical issue requiring immediate admin access, on-call escalation

**Authorization required**: Engineering Manager or VP Engineering approval

**Time to execute**: < 2 minutes

#### Steps

1. **Obtain approval**
   ```bash
   # Document approval in ticket
   # Required: Incident ID, approver name, reason
   INCIDENT_ID="INC-2026-001"
   APPROVER="eng_manager_name"
   REASON="Security incident - unauthorized content"
   ```

2. **Verify user identity**
   ```bash
   API_TOKEN="${API_TOKEN}"
   API_BASE="https://api.clpr.tv/api/v1"
   
   # Get user details
   USER_EMAIL="oncall@clpr.tv"
   curl -s -H "Authorization: Bearer $API_TOKEN" \
     "$API_BASE/users?email=$USER_EMAIL" | jq '{id, email, username, current_role}'
   ```

3. **Grant admin role**
   ```bash
   USER_ID="user-abc123"
   
   # Promote to admin
   curl -X POST -H "Authorization: Bearer $API_TOKEN" \
     -H "Content-Type: application/json" \
     "$API_BASE/users/$USER_ID/role" \
     -d '{
       "role": "admin",
       "reason": "Emergency escalation - '$INCIDENT_ID'",
       "approver": "'$APPROVER'",
       "expires_at": "'$(date -u -d '+24 hours' '+%Y-%m-%dT%H:%M:%SZ')'"
     }'
   ```

4. **Verify access granted**
   ```bash
   curl -s -H "Authorization: Bearer $API_TOKEN" \
     "$API_BASE/users/$USER_ID" | jq '{role, permissions}'
   ```

5. **Document in audit log**
   ```bash
   # Audit log automatically created
   # Verify it exists
   curl -s -H "Authorization: Bearer $API_TOKEN" \
     "$API_BASE/moderation/audit-logs?actor_id=$USER_ID&limit=1" | jq
   ```

6. **Set reminder to revoke**
   ```bash
   # Schedule revocation after 24 hours
   echo "curl -X POST '$API_BASE/users/$USER_ID/role' -H 'Authorization: Bearer \$API_TOKEN' -d '{\"role\":\"user\"}'" | \
     at now + 24 hours
   ```

#### Post-Grant Actions

- [ ] Notify security team
- [ ] Update incident ticket
- [ ] Schedule access review
- [ ] Set expiration reminder

---

### Grant Temporary Moderator Access

**When to use**: Short-term moderation needs, event coverage, temporary staffing

**Authorization required**: Admin or channel owner

**Time to execute**: < 1 minute

#### Steps

1. **Identify user**
   ```bash
   USERNAME="temp_moderator"
   USER_ID=$(curl -s -H "Authorization: Bearer $API_TOKEN" \
     "$API_BASE/users/by-username/$USERNAME" | jq -r '.id')
   ```

2. **Grant moderator permissions**
   ```bash
   CHANNEL_ID="channel-xyz789"
   EXPIRES_AT=$(date -u -d '+7 days' '+%Y-%m-%dT%H:%M:%SZ')
   
   curl -X POST -H "Authorization: Bearer $API_TOKEN" \
     -H "Content-Type: application/json" \
     "$API_BASE/moderation/moderators" \
     -d '{
       "user_id": "'$USER_ID'",
       "channel_id": "'$CHANNEL_ID'",
       "permissions": ["ban_users", "moderate_content", "view_audit_logs"],
       "scope": "community",
       "expires_at": "'$EXPIRES_AT'",
       "reason": "Temporary event coverage"
     }'
   ```

3. **Notify user**
   ```bash
   # Send notification email (if configured)
   curl -X POST -H "Authorization: Bearer $API_TOKEN" \
     "$API_BASE/notifications/send" \
     -d '{
       "user_id": "'$USER_ID'",
       "type": "moderator_added",
       "message": "You have been granted temporary moderator access until '$EXPIRES_AT'"
     }'
   ```

---

### Revoke Emergency Access

**When to use**: After incident resolved, access no longer needed, security concern

**Time to execute**: < 1 minute

#### Steps

1. **Revoke admin role**
   ```bash
   USER_ID="user-abc123"
   
   # Demote to user
   curl -X POST -H "Authorization: Bearer $API_TOKEN" \
     -H "Content-Type: application/json" \
     "$API_BASE/users/$USER_ID/role" \
     -d '{
       "role": "user",
       "reason": "Emergency access revoked - incident resolved"
     }'
   ```

2. **Or revoke moderator permissions**
   ```bash
   # Get moderator record ID
   MODERATOR_ID=$(curl -s -H "Authorization: Bearer $API_TOKEN" \
     "$API_BASE/moderation/moderators?user_id=$USER_ID" | jq -r '.moderators[0].id')
   
   # Remove moderator
   curl -X DELETE -H "Authorization: Bearer $API_TOKEN" \
     "$API_BASE/moderation/moderators/$MODERATOR_ID"
   ```

3. **Verify revocation**
   ```bash
   curl -s -H "Authorization: Bearer $API_TOKEN" \
     "$API_BASE/users/$USER_ID" | jq '{role, permissions}'
   ```

4. **Document revocation**
   - Update incident ticket
   - Note in security log
   - Confirm with approver

---

## Permission Management

### Grant Moderator Permissions

**Use case**: Add new moderator to team

#### Standard Moderator (Common Permissions)

```bash
USER_ID="user-def456"
CHANNEL_ID="channel-xyz789"

curl -X POST -H "Authorization: Bearer $API_TOKEN" \
  -H "Content-Type: application/json" \
  "$API_BASE/moderation/moderators" \
  -d '{
    "user_id": "'$USER_ID'",
    "channel_id": "'$CHANNEL_ID'",
    "permissions": [
      "ban_users",
      "moderate_content",
      "view_audit_logs",
      "sync_twitch_bans"
    ],
    "scope": "community"
  }'
```

#### Senior Moderator (Additional Permissions)

```bash
curl -X POST -H "Authorization: Bearer $API_TOKEN" \
  -H "Content-Type: application/json" \
  "$API_BASE/moderation/moderators" \
  -d '{
    "user_id": "'$USER_ID'",
    "channel_id": "'$CHANNEL_ID'",
    "permissions": [
      "ban_users",
      "moderate_content",
      "view_audit_logs",
      "sync_twitch_bans",
      "manage_moderators",
      "export_audit_logs"
    ],
    "scope": "community"
  }'
```

#### Global Moderator

```bash
curl -X POST -H "Authorization: Bearer $API_TOKEN" \
  -H "Content-Type: application/json" \
  "$API_BASE/moderation/moderators" \
  -d '{
    "user_id": "'$USER_ID'",
    "channel_id": null,
    "permissions": [
      "ban_users",
      "moderate_content",
      "view_audit_logs"
    ],
    "scope": "global"
  }'
```

---

### Update Moderator Permissions

**Use case**: Add or remove specific permissions from existing moderator

```bash
# Get moderator ID
MODERATOR_ID=$(curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/moderation/moderators?user_id=$USER_ID&channel_id=$CHANNEL_ID" | \
  jq -r '.moderators[0].id')

# Update permissions
curl -X PATCH -H "Authorization: Bearer $API_TOKEN" \
  -H "Content-Type: application/json" \
  "$API_BASE/moderation/moderators/$MODERATOR_ID" \
  -d '{
    "permissions": [
      "ban_users",
      "moderate_content",
      "view_audit_logs",
      "manage_moderators"
    ],
    "reason": "Promoted to senior moderator"
  }'
```

#### Add Single Permission

```bash
# Get current permissions
CURRENT_PERMS=$(curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/moderation/moderators/$MODERATOR_ID" | jq -r '.permissions')

# Add new permission
NEW_PERMS=$(echo "$CURRENT_PERMS" | jq '. + ["export_audit_logs"] | unique')

# Update
curl -X PATCH -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/moderation/moderators/$MODERATOR_ID" \
  -d "{\"permissions\": $NEW_PERMS}"
```

---

### Revoke Moderator Permissions

**Use case**: Remove moderator access

```bash
# Complete removal
MODERATOR_ID="mod-abc123"
curl -X DELETE -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/moderation/moderators/$MODERATOR_ID"

# Or downgrade to read-only
curl -X PATCH -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/moderation/moderators/$MODERATOR_ID" \
  -d '{"permissions": ["view_audit_logs"]}'
```

---

### Grant Admin Access

**Use case**: Promote trusted moderator to admin

**Authorization required**: Super admin or VP Engineering

```bash
USER_ID="user-ghi789"

# Promote to admin
curl -X POST -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/users/$USER_ID/role" \
  -d '{
    "role": "admin",
    "reason": "Promotion to admin role",
    "approver": "super_admin_username"
  }'
```

---

## Troubleshooting Permission Issues

### Permission Denied Errors

#### Symptom

```json
{
  "error": "PERMISSION_DENIED",
  "message": "You do not have permission to perform this action",
  "required_permission": "ban_users",
  "user_permissions": ["view_content", "submit_content"]
}
```

#### Diagnosis

```bash
#!/bin/bash
# diagnose-permissions.sh

USER_ID="${1:-}"
REQUIRED_PERMISSION="${2:-ban_users}"

if [ -z "$USER_ID" ]; then
  echo "Usage: $0 <user_id> [required_permission]"
  exit 1
fi

echo "Permission Diagnosis for User: $USER_ID"
echo "Required Permission: $REQUIRED_PERMISSION"
echo "========================================"
echo

# Get user details
USER_INFO=$(curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/users/$USER_ID")

echo "User Role:"
echo "$USER_INFO" | jq '{role, is_admin}'
echo

echo "User Permissions:"
echo "$USER_INFO" | jq '.permissions[]' 
echo

# Check if user is moderator
MODERATOR_INFO=$(curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/moderation/moderators?user_id=$USER_ID")

echo "Moderator Status:"
echo "$MODERATOR_INFO" | jq '.moderators[] | {channel_id, scope, permissions}'
echo

# Check if has required permission
HAS_PERMISSION=$(echo "$USER_INFO" | jq --arg perm "$REQUIRED_PERMISSION" \
  '.permissions | contains([$perm])')

if [ "$HAS_PERMISSION" = "true" ]; then
  echo "✓ User HAS required permission: $REQUIRED_PERMISSION"
else
  echo "✗ User MISSING required permission: $REQUIRED_PERMISSION"
  echo
  echo "To grant permission, run:"
  echo "  # If user should be moderator:"
  echo "  curl -X POST '$API_BASE/moderation/moderators' -d '{\"user_id\":\"$USER_ID\",\"permissions\":[\"$REQUIRED_PERMISSION\"]}'"
  echo
  echo "  # Or promote to admin:"
  echo "  curl -X POST '$API_BASE/users/$USER_ID/role' -d '{\"role\":\"admin\"}'"
fi
```

---

### Missing Permissions

**Scenario**: User can't perform action despite having moderator role

**Common causes**:
1. Permission not included in moderator grant
2. Scope limitation (community vs global)
3. Channel ownership issue
4. Expired temporary access

**Solution**:

```bash
# 1. Check current permissions
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/moderation/moderators?user_id=$USER_ID" | \
  jq '.moderators[] | {channel_id, scope, permissions, expires_at}'

# 2. Check if access expired
EXPIRES_AT=$(curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/moderation/moderators?user_id=$USER_ID" | \
  jq -r '.moderators[0].expires_at')

if [ "$EXPIRES_AT" != "null" ]; then
  EXPIRES_EPOCH=$(date -d "$EXPIRES_AT" +%s)
  NOW_EPOCH=$(date +%s)
  
  if [ $NOW_EPOCH -gt $EXPIRES_EPOCH ]; then
    echo "Access expired on $EXPIRES_AT"
  fi
fi

# 3. Grant missing permission
./diagnose-permissions.sh $USER_ID ban_users
```

---

### Permission Conflicts

**Scenario**: User has conflicting permission grants

**Example**: User is moderator of Channel A, trying to moderate Channel B

**Diagnosis**:

```bash
# List all moderator records for user
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/moderation/moderators?user_id=$USER_ID" | \
  jq '.moderators[] | {id, channel_id, scope, permissions}'

# Check for global scope
curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/moderation/moderators?user_id=$USER_ID" | \
  jq '.moderators[] | select(.scope == "global")'
```

**Solution**:

```bash
# If user needs global access, create global moderator record
curl -X POST -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/moderation/moderators" \
  -d '{
    "user_id": "'$USER_ID'",
    "channel_id": null,
    "scope": "global",
    "permissions": ["ban_users", "moderate_content"]
  }'
```

---

## Audit and Compliance

### Review Permission Changes

**Use case**: Security audit, compliance check, incident investigation

```bash
# All permission grant/revoke actions in last 30 days
START_TIME=$(date -u -d '30 days ago' '+%Y-%m-%dT%H:%M:%SZ')

curl -s -H "Authorization: Bearer $API_TOKEN" \
  "$API_BASE/moderation/audit-logs?action=add_moderator,remove_moderator,update_permissions&start_time=$START_TIME&limit=500" | \
  jq '.logs[] | {
    time: .created_at,
    actor: .actor_username,
    action: .action,
    target_user: .details.username,
    permissions: .details.permissions
  }'
```

### Permission Audit Report

```bash
#!/bin/bash
# permission-audit-report.sh

OUTPUT_FILE="permission-audit-$(date +%Y%m%d).txt"

{
  echo "=== Permission Audit Report ==="
  echo "Generated: $(date -u)"
  echo "==============================="
  echo
  
  echo "=== ADMIN USERS ==="
  curl -s -H "Authorization: Bearer $API_TOKEN" \
    "$API_BASE/users?role=admin" | jq -r '.[] | "  \(.username) (\(.email))"'
  echo
  
  echo "=== GLOBAL MODERATORS ==="
  curl -s -H "Authorization: Bearer $API_TOKEN" \
    "$API_BASE/moderation/moderators?scope=global" | \
    jq -r '.moderators[] | "  \(.username) - Permissions: \(.permissions | join(", "))"'
  echo
  
  echo "=== RECENT PERMISSION CHANGES (Last 7 Days) ==="
  START_TIME=$(date -u -d '7 days ago' '+%Y-%m-%dT%H:%M:%SZ')
  curl -s -H "Authorization: Bearer $API_TOKEN" \
    "$API_BASE/moderation/audit-logs?action=add_moderator,remove_moderator,update_permissions&start_time=$START_TIME" | \
    jq -r '.logs[] | "  \(.created_at) - \(.action) - \(.actor_username) -> \(.details.username)"'
  echo
  
  echo "=== TEMPORARY ACCESS (Expires Soon) ==="
  WEEK_FROM_NOW=$(date -u -d '+7 days' '+%Y-%m-%dT%H:%M:%SZ')
  curl -s -H "Authorization: Bearer $API_TOKEN" \
    "$API_BASE/moderation/moderators" | \
    jq --arg week "$WEEK_FROM_NOW" -r '.moderators[] | select(.expires_at != null and .expires_at < $week) | "  \(.username) - Expires: \(.expires_at)"'
  
} > "$OUTPUT_FILE"

echo "Audit report generated: $OUTPUT_FILE"
```

### Compliance Checks

**SOC 2 Requirements**:
- [ ] All permission changes logged in audit trail
- [ ] Emergency access reviewed within 24 hours
- [ ] Quarterly access reviews completed
- [ ] Approvals documented for admin promotions

**GDPR Requirements**:
- [ ] User data access limited to authorized personnel
- [ ] Permission grants time-limited where appropriate
- [ ] Access logs retained for 7 years

---

## Related Runbooks

- [Moderation Operations](./moderation-operations.md) - Emergency ban/unban procedures
- [Audit Log Operations](./audit-log-operations.md) - Review permission changes
- [Moderation Incidents](./moderation-incidents.md) - Incident response
- [Moderation Rollback](./moderation-rollback.md) - Rollback procedures

---

## Emergency Contacts

### Permission Escalation Approvers

- **Engineering Manager**: eng-manager@clpr.tv
- **VP Engineering**: vp-eng@clpr.tv
- **Security Team**: security@clpr.tv

### On-Call

- **PagerDuty**: Operations escalation
- **Slack**: #ops-incidents

---

**Last Updated**: 2026-02-03  
**Document Owner**: Operations Team  
**Review Frequency**: Quarterly
