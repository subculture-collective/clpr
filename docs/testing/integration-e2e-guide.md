---
title: "Integration E2e Guide"
summary: "This document provides a comprehensive guide to running and understanding the integration and end-to-end tests for Clipper."
tags: ["testing","guide"]
area: "testing"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Integration & E2E Testing Guide

This document provides a comprehensive guide to running and understanding the integration and end-to-end tests for Clipper.

## Table of Contents

- [Overview](#overview)
- [Test Types](#test-types)
- [Prerequisites](#prerequisites)
- [Running Tests](#running-tests)
- [Test Structure](#test-structure)
- [Coverage Goals](#coverage-goals)
- [CI/CD Integration](#cicd-integration)
- [Troubleshooting](#troubleshooting)

## Overview

Clipper's testing strategy includes multiple layers:

1. **Unit Tests**: Test individual components in isolation
2. **Integration Tests**: Test interactions between components and external services
3. **E2E Tests**: Test complete user journeys through the application
4. **Load Tests**: Test system performance under stress

This guide focuses on integration and E2E tests.

## Test Types

### Backend Integration Tests

Located in `backend/tests/integration/`, these tests validate:

- **Authentication** (`auth/`): Login, logout, OAuth, MFA, token refresh
- **Submissions** (`submissions/`): Clip creation, editing, deletion, validation
- **Engagement** (`engagement/`): Comments, likes, favorites, follows
- **Premium** (`premium/`): Subscriptions, payments, webhooks
- **Search** (`search/`): Keyword search, filters, pagination, fuzzy matching
- **API** (`api/`): Health checks, public endpoints, general API functionality

### Frontend E2E Tests

Located in `frontend/e2e/`, these tests validate:

- **Authentication Flows**: Login, logout, OAuth redirects
- **Submission Workflows**: Clip submission and validation
- **Search Functionality**: Search input, results, filters
- **Engagement Features**: Comments, likes, favorites
- **Premium Features**: Subscription pages and checkout flows
- **Responsive Design**: Mobile and tablet viewports
- **Accessibility**: Semantic HTML, ARIA attributes, keyboard navigation

## Prerequisites

### For Backend Integration Tests

1. **Docker & Docker Compose**: Required for test database and Redis
2. **PostgreSQL Client**: For database verification (optional)
3. **Go 1.22+**: Backend runtime
4. **golang-migrate**: For database migrations

Install golang-migrate:
```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### For Frontend E2E Tests

1. **Node.js 20+**: Frontend runtime
2. **Playwright**: E2E testing framework (installed via npm)

Install Playwright browsers:
```bash
cd frontend
npx playwright install --with-deps
```

## Running Tests

### Quick Start

```bash
# Run all integration tests
make test-integration

# Run all E2E tests
make test-e2e

# Run all tests (unit + integration + E2E)
make test
```

### Backend Integration Tests

#### Run All Integration Tests

```bash
make test-integration
```

This command:
1. Starts PostgreSQL and Redis test containers
2. Runs database migrations
3. Executes all integration tests
4. Stops and cleans up test containers

#### Run Specific Test Suites

```bash
# Authentication tests only
make test-integration-auth

# Submission tests only
make test-integration-submissions

# Engagement tests only
make test-integration-engagement

# Premium subscription tests only
make test-integration-premium

# Search tests only
make test-integration-search

# API tests only
make test-integration-api
```

#### Run Tests Manually

```bash
# Start test infrastructure
docker compose -f docker-compose.test.yml up -d

# Run migrations
migrate -path backend/migrations -database "postgresql://clpr:clpr_password@localhost:5437/clpr_test?sslmode=disable" up

# Run specific test suite
cd backend
go test -v -tags=integration ./tests/integration/auth/...

# Run all integration tests
go test -v -tags=integration ./tests/integration/...

# Run with coverage
go test -v -tags=integration -coverprofile=coverage.out ./tests/integration/...
go tool cover -html=coverage.out

# Cleanup
docker compose -f docker-compose.test.yml down
```

### Frontend E2E Tests

#### Run All E2E Tests

```bash
make test-e2e
```

Or directly:

```bash
cd frontend
npm run test:e2e
```

#### Run Specific E2E Tests

```bash
cd frontend

# Run specific test file
npx playwright test e2e/integration.spec.ts

# Run tests with specific browser
npx playwright test --project=chromium

# Run tests in headed mode (see browser)
npx playwright test --headed

# Run tests in debug mode
npx playwright test --debug

# Run tests with UI mode
npx playwright test --ui
```

#### Run E2E Tests with Backend

For full integration testing:

```bash
# Terminal 1: Start backend
docker compose -f docker-compose.test.yml up -d
make backend-dev

# Terminal 2: Run E2E tests
cd frontend
npm run test:e2e
```

## Test Structure

### Backend Integration Test Structure

```go
// +build integration

package auth

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestAuthenticationFlow(t *testing.T) {
    router, authService, db, redisClient := setupTestRouter(t)
    defer db.Close()
    defer redisClient.Close()

    t.Run("SuccessCase", func(t *testing.T) {
        // Test implementation
    })

    t.Run("ErrorCase", func(t *testing.T) {
        // Test implementation
    })
}
```

### Frontend E2E Test Structure

```typescript
import { test, expect } from '@playwright/test';

test.describe('Feature Name', () => {
    test('should perform action', async ({ page }) => {
        await page.goto('/');
        
        const element = page.locator('[data-testid="element"]');
        await expect(element).toBeVisible();
    });
});
```

## Coverage Goals

### Backend Integration Tests

| Test Suite | Coverage Target | Current Status |
|------------|----------------|----------------|
| Authentication | 90%+ | ✅ Implemented |
| Submissions | 85%+ | ✅ Implemented |
| Engagement | 85%+ | ✅ Implemented |
| Premium | 80%+ | ✅ Implemented |
| Search | 85%+ | ✅ Implemented |
| API Endpoints | 90%+ | ✅ Implemented |

### Frontend E2E Tests

| Test Suite | Coverage Target | Current Status |
|------------|----------------|----------------|
| Authentication Flows | 85%+ | ✅ Implemented |
| Submission Workflows | 80%+ | ✅ Implemented |
| Search Functionality | 85%+ | ✅ Implemented |
| Engagement Features | 80%+ | ✅ Implemented |
| Premium Features | 75%+ | ✅ Implemented |
| Mobile Responsiveness | 80%+ | ✅ Implemented |
| Accessibility | 75%+ | ✅ Implemented |

## CI/CD Integration

### GitHub Actions Workflow

Tests run automatically on every pull request via `.github/workflows/ci.yml`:

```yaml
backend-integration-test:
  runs-on: ubuntu-latest
  services:
    postgres:
      image: pgvector/pgvector:pg17
      # ... configuration
    redis:
      image: redis:7-alpine
      # ... configuration
  steps:
    - name: Run integration tests
      run: cd backend && go test -v -tags=integration ./...

frontend-e2e:
  runs-on: ubuntu-latest
  steps:
    - name: Run E2E tests
      run: cd frontend && npm run test:e2e
```

### Test Execution Time Targets

- Individual integration test: < 5 seconds
- Full integration suite: 5–8 minutes ✅
- Individual E2E test: < 30 seconds
- Full E2E suite: < 10 minutes
- Total CI pipeline: < 20 minutes

## Troubleshooting

### Backend Integration Tests

#### Database Connection Errors

```bash
# Check if test database is running
docker compose -f docker-compose.test.yml ps

# Check PostgreSQL logs
docker compose -f docker-compose.test.yml logs postgres

# Test connection manually
psql -h localhost -p 5437 -U clpr -d clpr_test
```

#### Migration Errors

```bash
# Check migration status
migrate -path backend/migrations -database "postgresql://clpr:clpr_password@localhost:5437/clpr_test?sslmode=disable" version

# Force migration to specific version
migrate -path backend/migrations -database "postgresql://clpr:clpr_password@localhost:5437/clpr_test?sslmode=disable" force VERSION
```

#### Redis Connection Errors

```bash
# Check if Redis is running
docker compose -f docker-compose.test.yml ps redis

# Test Redis connection
redis-cli -p 6380 ping

# Check Redis logs
docker compose -f docker-compose.test.yml logs redis
```

### Frontend E2E Tests

#### Browser Installation Issues

```bash
# Reinstall Playwright browsers
cd frontend
npx playwright install --with-deps chromium
```

#### Test Timeout Issues

Increase timeout in `playwright.config.ts`:

```typescript
export default defineConfig({
  timeout: 60 * 1000, // 60 seconds
  expect: {
    timeout: 10 * 1000, // 10 seconds
  },
});
```

#### Backend Not Available

Make sure backend is running before E2E tests:

```bash
# Check backend health
curl http://localhost:8080/api/v1/health

# View backend logs
docker compose logs backend
```

### General Troubleshooting

#### Clean Everything and Start Fresh

```bash
# Stop all containers
docker compose -f docker-compose.test.yml down -v

# Remove test data
rm -rf backend/coverage.out frontend/playwright-report

# Reinstall dependencies
cd backend && go mod download
cd frontend && npm ci

# Run tests again
make test-integration
make test-e2e
```

#### View Detailed Test Output

```bash
# Backend: Verbose mode
cd backend
go test -v -tags=integration ./tests/integration/... 2>&1 | tee test-output.log

# Frontend: Playwright trace viewer
cd frontend
npx playwright test --trace on
npx playwright show-report
```

## Best Practices

### Writing Integration Tests

1. **Isolate Tests**: Each test should be independent
2. **Clean Up**: Always clean up test data
3. **Use Transactions**: Rollback changes where possible
4. **Mock External Services**: Use mocks for third-party APIs
5. **Test Realistic Scenarios**: Use production-like data

### Writing E2E Tests

1. **Use Data Test IDs**: Prefer `[data-testid]` selectors
2. **Wait for Elements**: Use Playwright's auto-waiting
3. **Test User Journeys**: Focus on complete workflows
4. **Handle Async Operations**: Wait for network and animations
5. **Take Screenshots on Failure**: Already configured in Playwright

## Resources

- [Backend Integration Tests README](../../backend/tests/integration/README.md)
- [Playwright Documentation](https://playwright.dev)
- [Testing Guide](../backend/testing.md)
- [GitHub Actions Workflow](../../.github/workflows/ci.yml)
- [Makefile](../../Makefile)
