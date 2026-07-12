---
title: "E2E CONFIGURATION"
summary: "The Playwright E2E tests support optional test modes that enable additional test coverage."
tags: ["docs"]
area: "docs"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# E2E Test Configuration Guide

## Overview

The Playwright E2E tests support optional test modes that enable additional test coverage.
Some tests are skipped by default and require specific configurations to run.

## Test Modes

### 1. CDN Failover Mode

**Status**: ✅ Can be enabled
**Configuration**: `E2E_CDN_FAILOVER_MODE=true`
**Tests**: ~7 additional tests

Tests CDN failure scenarios including:

-   Static asset fallback to origin
-   HLS video playback during CDN failures
-   UI responsiveness during CDN issues

**Enable with**:

```bash
source .env.e2e
# Then run tests
npm run test:e2e
```

### 2. Search Failover Mode

**Status**: ✅ Can be enabled
**Configuration**: `E2E_FAILOVER_MODE=true`
**Tests**: ~1 additional test

Tests search service failover behavior:

-   Search continues during outages
-   Graceful degradation

**Enable with**:

```bash
source .env.e2e
npm run test:e2e
```

### 3. Stripe Test Mode

**Status**: ✅ Can be enabled
**Configuration**: `E2E_STRIPE_TEST_MODE=true`
**Tests**: ~5 additional tests

Tests payment and subscription functionality:

-   Checkout flow with test cards
-   Subscription management
-   Payment processing

**Requires**: Stripe test account setup
**Enable with**:

```bash
source .env.e2e
npm run test:e2e
```

## Conditional Skips (Cannot be enabled)

Some tests are skipped because they require specific runtime conditions that cannot be configured:

### Video Playback Tests

-   **Why skipped**: Requires actual video player on page + valid HLS clip
-   **Tests affected**: CDN failover HLS tests (5 tests)
-   **How to enable**: Seed database with actual video clips during test setup

### Admin/Moderator Tests

-   **Why skipped**: Requires logged-in user with specific role
-   **Tests affected**: Channel management, moderation workflow (10+ tests)
-   **How to enable**: Modify test fixtures to create users with admin/moderator roles

### Rate Limiting Tests

-   **Why skipped**: Rate limit not triggered in test environment
-   **Tests affected**: Clip submission rate limiting (2 tests)
-   **How to enable**: Configure test backend with aggressive rate limits

### Authenticated User Tests

-   **Why skipped**: Requires pre-authenticated session
-   **Tests affected**: Some integration tests (4-5 tests)
-   **How to enable**: Pre-create authenticated sessions in test setup

## Setup E2E Environment

### Quick Setup

```bash
cd frontend
bash setup-e2e-tests.sh
```

This script will:

1. Prompt for which test modes to enable
2. Create `.env.e2e` with your configuration
3. Show available options

### Manual Setup

Create `frontend/.env.e2e`:

```bash
# Enable CDN failover tests (~7 additional tests)
E2E_CDN_FAILOVER_MODE=true

# Enable search failover tests (~1 additional test)
E2E_FAILOVER_MODE=true

# Enable Stripe subscription tests (~5 additional tests)
E2E_STRIPE_TEST_MODE=true
```

### Load Configuration

```bash
cd frontend
source .env.e2e
npm run test:e2e
```

Or use the Makefile:

```bash
make test-frontend-e2e
```

## Current Test Results

### With All Configurations Enabled

```
Running 353 tests
  ✓ 311 passed
  - 42 skipped (conditional requirements not met)
  Duration: ~3.5 minutes
```

### Breakdown

| Category              | Status          | Count   |
| --------------------- | --------------- | ------- |
| Auth Tests            | ✓ Passing       | 52      |
| Session Management    | ✓ Passing       | 38      |
| CDN Failover          | ✓ Passing       | 5       |
| CDN Failover HLS\*    | - Skipped       | 5       |
| Channel Management    | ✓ Passing       | 10      |
| Channel Permissions\* | - Skipped       | 4       |
| Clip Submission       | ✓ Passing       | 25      |
| Rate Limiting\*       | - Skipped       | 2       |
| Integration           | ✓ Passing       | 75      |
| Integration Auth\*    | - Skipped       | 4       |
| Moderation\*          | - Skipped       | 20      |
| Premium/Stripe\*      | - Skipped       | 5       |
| Social Features       | ✓ Passing       | 58      |
| **Total**             | **311 passing** | **353** |

\*Skipped due to runtime environment requirements

## Running Tests

### With Setup

```bash
# One-time setup
bash setup-e2e-tests.sh

# Run tests with configurations
source .env.e2e
npm run test:e2e
```

### Without Setup (Default)

```bash
npm run test:e2e
# Runs 310 tests (more skipped by default)
```

### With Verbose Output

```bash
source .env.e2e
bash run-playwright-tests.sh
```

### View Test Report

```bash
npx playwright show-report
```

## Troubleshooting

### Tests timing out

-   Ensure backend API is running at `http://localhost:8080`
-   Check network connectivity
-   Try with `--headed` mode to see what's happening

### Tests skipped unexpectedly

-   Check if environment variables are loaded: `echo $E2E_CDN_FAILOVER_MODE`
-   Re-source `.env.e2e` in your terminal
-   Verify files exist: `cat .env.e2e`

### Some tests still skipped

-   This is expected - 42 tests require runtime conditions
-   See "Conditional Skips" section above
-   These tests would need code modifications to enable

## Configuration Files

-   `.env.e2e` - Test environment variables (generated by setup script)
-   `playwright.config.ts` - Playwright test configuration
-   `run-playwright-tests.sh` - Test runner that loads `.env.e2e`

## See Also

-   PLAYWRIGHT_SETUP_GUIDE.md - Complete Playwright setup
-   FRONTEND_TEST_SETUP_SUMMARY.md - Infrastructure overview
-   [Playwright Documentation](https://playwright.dev)
