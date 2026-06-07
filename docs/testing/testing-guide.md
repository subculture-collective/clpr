---
title: "Testing Guide"
summary: "This document provides comprehensive information about testing Clipper's clip submission flow and related features."
tags: ["testing","guide"]
area: "testing"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Testing Documentation

This document provides comprehensive information about testing Clipper's clip submission flow and related features.

## Table of Contents

- [Overview](#overview)
- [Test Strategy](#test-strategy)
- [Prerequisites](#prerequisites)
- [Running Tests](#running-tests)
- [Test Coverage](#test-coverage)
- [CI/CD Integration](#cicd-integration)
- [Writing New Tests](#writing-new-tests)
- [Troubleshooting](#troubleshooting)

## Overview

Clipper employs a multi-layered testing strategy to ensure reliability and quality:

1. **Unit Tests**: Test individual components in isolation
2. **Integration Tests**: Test interactions between components with real databases
3. **E2E Tests**: Test complete user workflows through the application
4. **Load Tests**: Verify system performance under stress
5. **Security Tests**: Validate authentication, authorization, and input sanitization

This documentation focuses on E2E testing for the clip submission workflow as specified in production readiness requirements.

## Test Strategy

### Clip Submission Flow Coverage

The clip submission workflow includes the following stages:

```
User → Submit URL → Fetch Metadata → Validation → 
Rate Limiting → Duplicate Check → Moderation Queue →
Approval/Rejection → Published to Feed
```

Each stage is tested at multiple levels:

- **Backend Integration**: API endpoints, business logic, database operations
- **Frontend E2E**: User interface, form validation, API integration
- **Mobile E2E**: Native app workflows, platform-specific features
- **Load**: Performance under concurrent user load
- **Security**: Authentication, authorization, input validation

## Prerequisites

### Backend Integration Tests

**Required**:
- Docker & Docker Compose (for PostgreSQL and Redis)
- Go 1.22 or later
- golang-migrate for database migrations

**Installation**:
```bash
# Install golang-migrate
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/
migrate -version
```

**Services**:
```bash
# Start test database and Redis
docker-compose -f docker-compose.test.yml up -d

# Run migrations
cd backend
migrate -path migrations -database "postgresql://clpr:clpr_password@localhost:5437/clpr_test?sslmode=disable" up
```

### Frontend E2E Tests

**Required**:
- Node.js 20 or later
- Playwright for browser automation

**Installation**:
```bash
cd frontend
npm install
npx playwright install --with-deps chromium
```

### Mobile E2E Tests

**Required** (for full mobile testing):
- React Native development environment
- iOS Simulator (macOS) or Android Emulator
- Detox testing framework

**Installation**:
```bash
cd mobile
npm install
npx detox build --configuration ios.sim.debug
```

### Load Testing

**Required**:
- k6 load testing tool

**Installation**:
```bash
# macOS
brew install k6

# Linux
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6
```

## Running Tests

### Quick Start

```bash
# Run all backend tests (unit + integration)
make test

# Run integration tests only
make test-integration

# Run frontend E2E tests
make test-e2e

# Run load tests (smoke test with 10 users)
make test-load
```

### Backend Integration Tests

#### Submission Workflow Tests

```bash
# Test complete submission workflow
cd backend
go test -v -tags=integration ./tests/integration/submissions -run TestSubmissionWorkflowE2E

# Test metadata endpoint
go test -v -tags=integration ./tests/integration/submissions -run TestSubmissionMetadataEndpoint

# Test rate limiting
go test -v -tags=integration ./tests/integration/submissions -run TestRateLimiting

# Test NSFW flag handling
go test -v -tags=integration ./tests/integration/submissions -run TestNSFWFlag

# Run all submission tests
go test -v -tags=integration ./tests/integration/submissions
```

#### Other Integration Test Suites

```bash
# Authentication tests
go test -v -tags=integration ./tests/integration/auth

# Engagement tests (comments, votes, favorites)
go test -v -tags=integration ./tests/integration/engagement

# Search tests
go test -v -tags=integration ./tests/integration/search

# Premium subscription tests
go test -v -tags=integration ./tests/integration/premium

# General API tests
go test -v -tags=integration ./tests/integration/api
```

### Frontend E2E Tests

```bash
cd frontend

# Run all E2E tests
npm run test:e2e

# Run specific test file
npx playwright test e2e/integration.spec.ts

# Run tests in headed mode (see browser)
npx playwright test --headed

# Run tests in debug mode
npx playwright test --debug

# Generate test report
npx playwright show-report
```

### Mobile Tests

```bash
cd mobile

# Run unit/component tests
npm test

# Run E2E tests (iOS)
npx detox test --configuration ios.sim.debug

# Run E2E tests (Android)
npx detox test --configuration android.emu.debug

# Run specific test file
npx detox test e2e/submit-flow.e2e.ts --configuration ios.sim.debug
```

### Load Tests

```bash
cd backend/tests/load

# Submission endpoint load test (10 VUs, 30 seconds)
k6 run --vus 10 --duration 30s scenarios/submit.js

# Concurrent submissions (100 users, 2 minutes)
k6 run --vus 100 --duration 2m scenarios/submit.js

# With custom thresholds
k6 run --vus 50 --duration 1m scenarios/submit.js \
  --threshold 'http_req_duration{endpoint:submit_clip}<300' \
  --threshold 'errors<0.05'

# Full mixed workload scenario
k6 run scenarios/mixed_behavior.js
```

### Security Tests

```bash
cd backend/tests/security

# Run IDOR (Insecure Direct Object Reference) tests
go test -v -tags=integration ./idor_test.go

# Run all security tests
go test -v -tags=integration ./...
```

## Test Coverage

### Backend Integration Test Coverage

The following E2E scenarios are tested for clip submissions:

#### Submit Endpoint (POST /api/v1/submissions)

- [x] **Valid submission with URL only**: Basic clip submission
- [x] **Custom title and tags**: Submission with user-provided metadata
- [x] **NSFW flag**: Marking content as not-safe-for-work
- [x] **Submission reason**: Optional note explaining why clip is submitted
- [x] **Duplicate detection**: Preventing re-submission of existing clips
- [x] **Rate limiting**: Enforcing submission quotas (5/hour for new users, 10/hour authenticated)
- [x] **Invalid URL format**: Proper error handling
- [x] **Missing required fields**: Validation error responses
- [x] **Unauthenticated request**: 401 Unauthorized
- [x] **Banned user**: Rejection with appropriate message

#### Metadata Endpoint (GET /api/v1/submissions/metadata)

- [x] **Valid Twitch clip URL**: Fetching clip metadata from Twitch API
- [x] **Invalid URL format**: Validation error
- [x] **Missing URL parameter**: Bad request response
- [x] **Unauthenticated request**: 401 Unauthorized
- [x] **Rate limiting**: 100 requests/hour per user

#### Moderation Workflow

- [x] **Auto-approval logic**: Users with karma ≥ 500 bypass moderation
- [x] **Manual moderation**: Low-karma users enter pending queue
- [x] **Admin list pending**: GET /api/v1/admin/submissions
- [x] **Admin approve**: POST /api/v1/admin/submissions/:id/approve
- [x] **Admin reject**: POST /api/v1/admin/submissions/:id/reject
- [x] **Rejection reasons**: Pre-defined templates and custom messages
- [x] **Bulk operations**: Approve/reject multiple submissions
- [x] **User cannot approve own submission**: Authorization check
- [x] **Regular user cannot access admin endpoints**: 403 Forbidden

#### User Submission Management

- [x] **List user submissions**: GET /api/v1/submissions
- [x] **Filter by status**: pending, approved, rejected
- [x] **Pagination**: page and limit parameters
- [x] **Submission stats**: GET /api/v1/submissions/stats
- [x] **Check clip status**: GET /api/v1/submissions/check/:clip_id

### Frontend E2E Test Coverage

Frontend tests cover user-facing workflows:

- [ ] **Submit form validation**: URL format, title length, tag limits
- [ ] **Metadata auto-fetch**: Clicking "Fetch Info" populates fields
- [ ] **Submission success**: Confirmation message and redirect
- [ ] **User submissions page**: Displays all user submissions with status
- [ ] **Status filtering**: Filter by pending/approved/rejected
- [ ] **Admin moderation queue**: Admin-only page for pending submissions
- [ ] **Approve/reject actions**: Admin can moderate submissions
- [ ] **Approved clips in feed**: Verified clips appear in main feed
- [ ] **Rejected clip reason**: User sees rejection reason on their submissions page

**Note**: Frontend E2E tests are partially implemented. The test file `frontend/e2e/integration.spec.ts` contains basic structure but requires completion.

### Mobile E2E Test Coverage

Mobile tests validate the 4-step submission wizard:

- [ ] **Step 1: URL Input**: Validate URL format, show errors
- [ ] **Step 2: Metadata Review**: Display fetched clip info, loading states
- [ ] **Step 3: Customization**: Edit title, add tags (max 5), toggle NSFW
- [ ] **Step 4: Confirmation**: Review and submit
- [ ] **Back navigation**: Preserve state when going back
- [ ] **Tag limit enforcement**: Cannot add more than 5 tags
- [ ] **Success screen**: Shows next actions after submission
- [ ] **Error handling**: Network errors, validation failures

**Note**: Mobile E2E tests with Detox are not fully implemented. Basic Jest tests exist in `mobile/__tests__/submit-flow.test.ts`.

### Load Test Coverage

Load tests validate system performance:

- [x] **Concurrent submissions**: 100 users submitting simultaneously
- [x] **Metadata endpoint performance**: p95 < 500ms, p99 < 1000ms
- [x] **Submit endpoint performance**: p95 < 300ms, p99 < 600ms
- [x] **Rate limiting under load**: No false positives
- [x] **Database connection pool**: Stable under sustained load
- [x] **Redis cache effectiveness**: Cache hit rate > 80%

Load test scenarios exist in `backend/tests/load/scenarios/`:
- `submit.js`: Submission-focused load test
- `mixed_behavior.js`: Realistic user behavior mix
- Additional scenarios for feed, search, comments

### Security Test Coverage

Security tests validate protection mechanisms:

- [x] **Authentication required**: Unauthenticated requests rejected
- [x] **Authorization checks**: Users cannot perform admin actions
- [x] **CSRF protection**: POST requests validate tokens
- [x] **SQL injection**: Parameterized queries used throughout
- [x] **XSS prevention**: Input sanitization on titles/tags
- [x] **Rate limit bypass**: Cannot circumvent rate limits
- [x] **IDOR protection**: Users cannot access others' pending submissions

## CI/CD Integration

### GitHub Actions Workflow

The CI pipeline runs tests automatically on every push and pull request:

```yaml
# .github/workflows/ci.yml
jobs:
  backend-test:
    - Runs unit tests with coverage
    - Enforces minimum coverage threshold (currently 8%, target 15%+)

  backend-integration-test:
    - Starts PostgreSQL and Redis services
    - Runs database migrations
    - Executes integration tests
    - Reports failures

  frontend-test:
    - Runs Jest unit tests
    - Generates coverage report

  frontend-e2e:
    - Builds backend server
    - Starts services (PostgreSQL, Redis, backend API)
    - Installs Playwright browsers
    - Runs E2E tests
    - Uploads Playwright report as artifact
```

### Running Tests Locally Before Push

To avoid CI failures, run tests locally:

```bash
# Backend
cd backend
go test -v ./...
go test -v -tags=integration ./tests/integration/...

# Frontend
cd frontend
npm test
npm run lint
npm run test:e2e

# Format check
cd backend
gofmt -l .
```

### Troubleshooting CI Failures

**Backend test failures**:
1. Check if database migrations are up to date
2. Verify test database connection (port 5437, not 5432)
3. Ensure Redis is running (port 6380, not 6379)
4. Check for data leakage between tests (use unique IDs)

**Frontend E2E failures**:
1. View Playwright report artifact in GitHub Actions
2. Check if backend is reachable on localhost:8080
3. Verify API_URL environment variable is set correctly
4. Look for race conditions in async operations

**Load test failures** (if running in CI):
1. Reduce VUs (virtual users) for CI environment
2. Increase duration for slower CI machines
3. Adjust performance thresholds (p95, p99)

## Writing New Tests

### Backend Integration Tests

Integration tests follow this pattern:

```go
//go:build integration

package submissions

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestNewFeature(t *testing.T) {
    // Setup: Create test environment
    router, authService, db, redisClient := setupTestRouter(t)
    defer db.Close()
    defer redisClient.Close()

    // Create test data
    ctx := context.Background()
    // ... create users, clips, etc.

    t.Run("FeatureScenario", func(t *testing.T) {
        // Arrange: Prepare request
        req := httptest.NewRequest(...)
        w := httptest.NewRecorder()

        // Act: Execute request
        router.ServeHTTP(w, req)

        // Assert: Verify response
        assert.Equal(t, http.StatusOK, w.Code)
        // ... more assertions
    })
}
```

**Best Practices**:
- Use `//go:build integration` tag
- Clean up resources with `defer`
- Use `t.Run()` for subtests
- Generate unique IDs to avoid conflicts
- Test both success and failure cases

### Frontend E2E Tests

E2E tests use Playwright:

```typescript
import { test, expect } from '@playwright/test';

test.describe('Feature Workflow', () => {
    test('should complete user journey', async ({ page }) => {
        // Navigate to page
        await page.goto('/submit');

        // Interact with elements
        await page.fill('input[name="url"]', 'https://clips.twitch.tv/test');
        await page.click('button[type="submit"]');

        // Assert results
        await expect(page.locator('.success-message')).toBeVisible();
    });
});
```

**Best Practices**:
- Use data-testid attributes for stable selectors
- Wait for elements with `waitForSelector()`
- Test accessibility (keyboard navigation, ARIA labels)
- Take screenshots on failure
- Test mobile viewports

### Mobile E2E Tests

Mobile tests use Detox:

```typescript
describe('Submit Flow', () => {
    beforeEach(async () => {
        await device.reloadReactNative();
    });

    it('should submit clip successfully', async () => {
        // Navigate to submit screen
        await element(by.id('submit-button')).tap();

        // Fill form
        await element(by.id('url-input')).typeText('https://clips.twitch.tv/test');
        await element(by.id('submit-form-button')).tap();

        // Verify success
        await expect(element(by.id('success-message'))).toBeVisible();
    });
});
```

**Best Practices**:
- Use testID props on React Native components
- Test both iOS and Android if possible
- Handle async operations with waitFor()
- Test offline scenarios
- Verify push notifications

## Troubleshooting

### Common Issues

#### Database Connection Failed

**Symptom**: `failed to connect to database: connection refused`

**Solution**:
```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Start services if not running
docker-compose -f docker-compose.test.yml up -d

# Verify connection
psql -h localhost -p 5437 -U clpr -d clpr_test
```

#### Redis Connection Failed

**Symptom**: `failed to connect to Redis: connection refused`

**Solution**:
```bash
# Check if Redis is running
docker ps | grep redis

# Test connection
redis-cli -h localhost -p 6380 ping
```

#### Tests Fail with "userRepo.CreateUser undefined"

**Symptom**: Compilation error about undefined method

**Solution**: The user repository uses `Create()` method, not `CreateUser()`. Example:

```go
user := &models.User{
    ID:          uuid.New(),
    TwitchID:    "test123",
    Username:    "testuser",
    DisplayName: "Test User",
    Email:       "test@example.com",
    Role:        "user",
}
err := userRepo.Create(ctx, user)
```

#### Playwright Browsers Not Installed

**Symptom**: `Executable doesn't exist at...`

**Solution**:
```bash
cd frontend
npx playwright install --with-deps chromium
```

#### Load Tests Require AUTH_TOKEN

**Symptom**: `Skipping test iteration - no AUTH_TOKEN provided`

**Solution**:
```bash
# Generate token from backend
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"test"}'

# Run with token
AUTH_TOKEN="your-token-here" k6 run scenarios/submit.js
```

### Test Data Cleanup

Tests should clean up after themselves, but if data persists:

```bash
# Reset test database
docker-compose -f docker-compose.test.yml down -v
docker-compose -f docker-compose.test.yml up -d
cd backend
migrate -path migrations -database "postgresql://clpr:clpr_password@localhost:5437/clpr_test?sslmode=disable" up
```

### Debug Mode

Run tests with verbose output:

```bash
# Backend
go test -v -tags=integration ./tests/integration/... -run TestSubmission

# Frontend
npx playwright test --debug

# Mobile
npx detox test --loglevel trace
```

## Performance Targets

| Metric | Target | Notes |
|--------|--------|-------|
| Metadata fetch (p95) | < 500ms | GET /api/v1/submissions/metadata |
| Submit clip (p95) | < 300ms | POST /api/v1/submissions |
| Moderation list (p95) | < 200ms | GET /api/v1/admin/submissions |
| Cache hit rate | > 80% | Redis cache for metadata |
| Concurrent users | 100+ | Sustained load without errors |
| Error rate | < 1% | Under normal load |
| Rate limit accuracy | 100% | No false positives |

## References

- [Backend Testing Guide](backend/testing.md)
- [Load Testing Guide](backend/tests/load/README.md)
- [Security Testing Runbook](SECURITY_TESTING_RUNBOOK.md)
- [Integration & E2E Guide](testing/integration-e2e-guide.md)
- [Playwright Documentation](https://playwright.dev/)
- [Detox Documentation](https://wix.github.io/Detox/)
- [k6 Documentation](https://k6.io/docs/)

## Contributing

When adding new features:

1. **Write tests first** (TDD approach recommended)
2. **Cover happy path and edge cases**
3. **Add integration tests** for API endpoints
4. **Add E2E tests** for user workflows
5. **Update this documentation** with new test coverage
6. **Verify tests pass in CI** before merging

For questions or issues with testing:
- Open an issue in GitHub with the `testing` label
- Check existing issues for similar problems
- Consult the team in Discord #engineering channel

---

Last Updated: December 2025
