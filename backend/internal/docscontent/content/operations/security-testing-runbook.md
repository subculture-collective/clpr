---
title: "Security Testing Runbook"
summary: "This runbook provides quick reference procedures for IDOR and authorization security testing."
tags: ["operations","testing","runbook"]
area: "operations"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Security Testing Runbook

## Quick Reference for IDOR and Authorization Testing

### Daily Operations

```bash
# Run all security tests
make test-security

# Or directly
cd backend
go test -v ./tests/security/
```

### Pre-Deployment Checklist

- [ ] All IDOR tests passing (31/31)
- [ ] Authorization middleware tests passing
- [ ] **Validation middleware tests passing (all SQLi/XSS tests)**
- [ ] **Fuzzer smoke test passing (1000+ payloads)**
- [ ] No new endpoints without authorization
- [ ] **No new endpoints without input validation**
- [ ] Security audit log reviewed
- [ ] No high-severity vulnerabilities

## Test Suites

### 1. Input Validation Middleware Tests (`internal/middleware/validation_middleware_test.go` & `validation_middleware_security_test.go`)

**Purpose:** Prevent SQL injection, XSS, path traversal, and other injection attacks

**Coverage:**
- SQL injection patterns (UNION, SELECT, INSERT, UPDATE, DELETE, etc.)
- XSS attempts (script tags, event handlers, javascript: protocol)
- Path traversal attacks (../, ..\\)
- Header injection and validation
- Request body size limits
- UTF-8 validation
- Cross-field validation
- Fuzzer testing (1000+ random payloads)

**Run Tests:**
```bash
cd backend

# Run all validation tests
go test -v ./internal/middleware/ -run "TestInputValidation|TestSanitizeInput"

# Run specific test suites
go test -v ./internal/middleware/ -run TestInputValidationMiddleware_SQLInjectionEdgeCases
go test -v ./internal/middleware/ -run TestInputValidationMiddleware_XSSEdgeCases
go test -v ./internal/middleware/ -run TestInputValidationMiddleware_FuzzerSmoke

# Run integration tests (requires test database)
go test -v -tags=integration ./tests/integration/validation/
```

**Key Features:**
- ✅ Detects common SQLi patterns (UNION SELECT, subqueries)
- ✅ Blocks XSS vectors (script tags, event handlers, javascript: URLs)
- ✅ Prevents path traversal (../ patterns)
- ✅ Validates UTF-8 encoding
- ✅ Limits URL and header lengths
- ✅ Fuzzer smoke test with 1000+ random payloads
- ⚠️ Some edge cases logged as limitations (see test output)

**Known Limitations:**
- Simple semicolon-based SQLi (e.g., `; UPDATE users SET`) not caught
- Requires parameterized queries at database layer for complete protection
- Boolean/time-based blind SQLi harder to detect via pattern matching

### 2. IDOR Tests (`tests/security/idor_test.go`)

**Purpose:** Prevent unauthorized access to user resources

**Coverage:**
- Comment operations (create, read, update, delete)
- User settings access
- Clip metadata operations
- Favorite operations
- Subscription management

**Run Specific Suite:**
```bash
# Comment IDOR tests
go test -v ./tests/security/ -run TestIDORComment

# User settings IDOR tests
go test -v ./tests/security/ -run TestIDORUser

# Clip operations IDOR tests
go test -v ./tests/security/ -run TestIDORClip

# Subscription IDOR tests
go test -v ./tests/security/ -run TestIDORSubscription
```

### 3. Authorization Middleware Tests

**Purpose:** Verify authorization logic correctness

**Run:**
```bash
cd backend
go test -v ./internal/middleware/ -run TestCanAccessResource
go test -v ./internal/middleware/ -run TestPermissionMatrix
```

## Adding New Endpoints

### Checklist for New Protected Endpoints

1. **Add Permission Rule**
   ```go
   // In internal/middleware/authorization.go
   {
       Resource: ResourceTypeNewResource,
       Action: ActionUpdate,
       RequiresOwner: true,
       AllowedRoles: []string{models.RoleAdmin},
   }
   ```

2. **Create Ownership Checker** (if needed)
   ```go
   type NewResourceOwnershipChecker struct {
       repo *repository.NewResourceRepository
   }
   
   func (c *NewResourceOwnershipChecker) IsOwner(ctx context.Context, resourceID, userID uuid.UUID) (bool, error) {
       resource, err := c.repo.GetByID(ctx, resourceID)
       if err != nil {
           return false, err
       }
       return resource.OwnerID == userID, nil
   }
   ```

3. **Add IDOR Tests**
   ```go
   // In tests/security/idor_test.go
   func TestIDORNewResourceOperations(t *testing.T) {
       // Test owner access
       // Test non-owner denial
       // Test admin override
       // Test moderator access (if applicable)
   }
   ```

4. **Apply Middleware**
   ```go
   // In route setup
   checker := middleware.NewNewResourceOwnershipChecker(repo)
   router.PUT("/resources/:id",
       middleware.AuthMiddleware(authService),
       middleware.RequireResourceOwnership(
           middleware.ResourceTypeNewResource,
           middleware.ActionUpdate,
           checker,
       ),
       handler.UpdateResource,
   )
   ```

5. **Verify Tests Pass**
   ```bash
   go test -v ./tests/security/ -run TestIDORNewResource
   ```

## Manual Testing

### IDOR Test Scenarios

#### 1. Comment Modification by Non-Owner

```bash
# Create comment as User A
curl -X POST http://localhost:8080/api/v1/clips/{clip_id}/comments \
  -H "Authorization: Bearer $USER_A_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"content": "Test comment"}'
# Response: {"id": "comment-123", ...}

# Try to update as User B (should fail with 403)
curl -X PUT http://localhost:8080/api/v1/comments/comment-123 \
  -H "Authorization: Bearer $USER_B_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"content": "Hacked!"}'
# Expected: 403 Forbidden
```

#### 2. User Settings Access

```bash
# Get User A's ID from their profile
curl http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer $USER_A_TOKEN"
# Response: {"id": "user-a-id", ...}

# Try to access as User B (should fail with 403)
curl http://localhost:8080/api/v1/users/user-a-id/settings \
  -H "Authorization: Bearer $USER_B_TOKEN"
# Expected: 403 Forbidden

# Access own settings (should succeed)
curl http://localhost:8080/api/v1/users/me/settings \
  -H "Authorization: Bearer $USER_B_TOKEN"
# Expected: 200 OK with settings
```

#### 3. Subscription Cancellation

```bash
# Try to cancel User A's subscription as User B (should fail)
curl -X DELETE http://localhost:8080/api/v1/subscriptions/sub-a-id \
  -H "Authorization: Bearer $USER_B_TOKEN"
# Expected: 403 Forbidden
```

## Security Monitoring

### Key Metrics

Monitor these metrics in production:

1. **Authorization Failure Rate**
   - Target: < 1% of authenticated requests
   - Alert: > 5% failure rate

2. **Failed Access Attempts per User**
   - Target: < 10 per day
   - Alert: > 50 per day (potential attack)

3. **Failed Attempts per Resource**
   - Alert: > 100 attempts on single resource

### Log Analysis

```bash
# Search for authorization failures
grep "AUTHORIZATION_FAILURE" /var/log/app.log

# Count failures by user
grep "AUTHORIZATION_FAILURE" /var/log/app.log | \
  awk '{print $3}' | sort | uniq -c | sort -rn

# Count failures by resource type
grep "AUTHORIZATION_FAILURE" /var/log/app.log | \
  grep -oP 'resource=\K[^:]+' | sort | uniq -c
```

## Incident Response

### Suspected IDOR Attack

1. **Identify the Pattern**
   ```bash
   # Check logs for repeated failures from same user
   grep "AUTHORIZATION_FAILURE.*user=$SUSPICIOUS_USER_ID" /var/log/app.log
   ```

2. **Block Suspicious User**
   ```bash
   # Temporary ban
   curl -X POST http://localhost:8080/admin/users/$USER_ID/ban \
     -H "Authorization: Bearer $ADMIN_TOKEN" \
     -d '{"reason": "Suspected IDOR attack", "duration": "24h"}'
   ```

3. **Review Accessed Resources**
   ```sql
   -- Check if attacker successfully accessed any resources
   SELECT * FROM audit_logs 
   WHERE user_id = '$SUSPICIOUS_USER_ID' 
   AND created_at > NOW() - INTERVAL '1 hour'
   AND action IN ('read', 'update', 'delete');
   ```

4. **Notify Affected Users**
   - If unauthorized access occurred
   - Follow data breach notification procedures

## Common Issues and Solutions

### Issue: Tests Failing After Code Change

**Symptoms:** IDOR tests fail after modifying handlers or services

**Solution:**
1. Check if permission matrix rules changed
2. Verify ownership checker logic
3. Ensure test mocks match new behavior
4. Update tests if behavior change is intentional

### Issue: Legitimate Requests Being Denied

**Symptoms:** Users report 403 errors for valid operations

**Solution:**
1. Check user's role and account type
2. Verify ownership checker is working correctly
3. Review permission matrix for missing rules
4. Check logs for specific denial reason

### Issue: Slow Authorization Checks

**Symptoms:** High latency on protected endpoints

**Solution:**
1. Add database indices on owner_id fields
2. Cache ownership checks within request
3. Use eager loading for related data
4. Consider denormalizing ownership data

## Performance Testing

### Load Test with Authorization

```bash
# Use k6 for load testing
cd backend/tests/load
k6 run mixed_user_behavior.js

# Monitor authorization performance
# Target: < 50ms for authorization check
# Alert: > 100ms average
```

## Code Review Checklist

When reviewing PRs that add/modify endpoints:

- [ ] Endpoint requires authentication
- [ ] Authorization check implemented
- [ ] Permission matrix rule exists
- [ ] IDOR tests added
- [ ] Error messages don't leak information
- [ ] Audit logging added
- [ ] Documentation updated

## Compliance Auditing

### Quarterly Security Audit

1. **Review all endpoints**
   ```bash
   # List all routes
   cd backend
   grep -r "router\.\(GET\|POST\|PUT\|DELETE\)" cmd/api/main.go
   ```

2. **Verify authorization on each endpoint**
   - Check middleware presence
   - Verify handler-level checks
   - Confirm service-level validation

3. **Test with different user types**
   - Regular user
   - Moderator
   - Admin
   - Suspended user
   - Anonymous user

4. **Document findings**
   - Create issues for any gaps
   - Update threat model
   - Update authorization framework

## Resources

- **Documentation:** `/docs/AUTHORIZATION_FRAMEWORK.md`
- **Threat Model:** `/docs/product/threat-model.md`
- **OWASP IDOR Guide:** <https://owasp.org/www-project-web-security-testing-guide/latest/4-Web_Application_Security_Testing/05-Authorization_Testing/04-Testing_for_Insecure_Direct_Object_References>
- **Security Issues:** Label issues with `security` tag

## Contact

- **Security Team:** <security@example.com>
- **On-call:** Use PagerDuty for urgent issues
- **Slack:** #security channel

---

**Last Updated:** 2025-12-16  
**Next Review:** Quarterly (Q1 2026)
