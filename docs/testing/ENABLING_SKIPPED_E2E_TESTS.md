---
title: "ENABLING SKIPPED E2E TESTS"
summary: "This guide shows exactly how to convert each category of skipped tests to use the new fixture system."
tags: ["docs","testing"]
area: "docs"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

# Quick Guide: Enabling Skipped E2E Tests

This guide shows exactly how to convert each category of skipped tests to use the new fixture system.

## Test Fixture Usage Summary

All skipped tests can be enabled by:
1. Importing the correct fixtures from `../fixtures`
2. Removing `test.skip()` calls
3. Using the fixture in the test

## Category 1: Authenticated Session Tests

**Current Status**: Skipped with `test.skip(true, 'Requires authenticated session')`

**Examples**: integration.spec.ts lines 324, 334, 474, 553

### Before

```typescript
import { test, expect } from '@playwright/test';

test('should display submission form for authenticated users', async ({ page }) => {
  test.skip(true, 'Requires authenticated session');

  await page.goto('/submit');
  const urlInput = page.locator('input[type="url"]');
  await expect(urlInput).toBeVisible();
});
```

### After

```typescript
import { test, expect } from '../fixtures';  // ← Use enhanced fixtures

test('should display submission form for authenticated users', async ({ authenticatedPage }) => {
  // Remove: test.skip(true, ...)

  await authenticatedPage.goto('/submit');
  const urlInput = authenticatedPage.locator('input[type="url"]');
  await expect(urlInput).toBeVisible();
});
```

**Key Changes**:
* Import from `../fixtures` instead of `@playwright/test`
* Use `authenticatedPage` parameter instead of `page`
* Remove `test.skip()` call

## Category 2: Multi-User Permission Tests

**Current Status**: Skipped with `test.skip()` in function body

**Examples**: channel-management.spec.ts lines 310, 323

### Before

```typescript
import { test, expect } from '@playwright/test';

test('non-owner should not be able to delete channel', async ({ page }) => {
  // Skipping for now as it requires more complex setup
  test.skip();

  // ... test code
});
```

### After

```typescript
import { test, expect } from '../fixtures';

test('non-owner should not be able to delete channel', async ({ multiUserContexts }) => {
  const { admin, regular } = multiUserContexts;

  // Admin creates channel
  await admin.page.goto('/dashboard');
  // ... create channel as admin

  // Regular user tries to delete (should fail)
  await regular.page.goto('/channels/123');
  const deleteButton = regular.page.locator('[data-testid="delete"]');
  await expect(deleteButton).not.toBeVisible();

  // Admin can delete
  await admin.page.goto('/channels/123');
  const adminDeleteBtn = admin.page.locator('[data-testid="delete"]');
  await expect(adminDeleteBtn).toBeVisible();
});
```

**Key Changes**:
* Import from `../fixtures`
* Use `multiUserContexts` parameter
* Access `.page` property on each user context
* Each context has separate authentication
* Remove test.skip() call

## Category 3: HLS Video Tests

**Current Status**: Skipped with `test.skip(true, 'No video player found')`

**Examples**: cdn-failover.spec.ts lines 115, 152

### Before

```typescript
import { test, expect } from '@playwright/test';

test('HLS video plays', async ({ page }) => {
  test.skip(true, 'No video player found for this clip');

  await page.goto('/clips/123');
  const videoPlayer = page.locator('[data-testid="video-player"]');
  await expect(videoPlayer).toBeVisible();
});
```

### After

```typescript
import { test, expect } from '../fixtures';

test('HLS video plays', async ({ page, testClips }) => {
  // testClips.withVideo includes HLS URL in videoUrl property
  await page.goto(`/clips/${testClips.withVideo.id}`);

  const videoPlayer = page.locator('[data-testid="video-player"]');
  await expect(videoPlayer).toBeVisible();

  // Can interact with video player
  const playButton = page.locator('[data-testid="play-button"]');
  await expect(playButton).toBeVisible();
});
```

**Key Changes**:
* Import from `../fixtures`
* Use `testClips` fixture (or access it from existing test data)
* Use `testClips.withVideo.id` which has HLS streaming URL
* Remove test.skip() call
* The test clip already has videoUrl configured

## Category 4: Rate Limiting Tests

**Current Status**: Skipped with conditional logic

**Examples**: clip-submission-flow.spec.ts lines 446, 471

### Before

```typescript
import { test, expect } from '@playwright/test';

test('rate limiting prevents rapid submissions', async ({ page }) => {
  // Note: Requires backend rate limiting configuration
  test.skip(!process.env.ENABLE_RATE_LIMITING);

  // ... test code
});
```

### After

```typescript
import { test, expect } from '../fixtures';

test('rate limiting prevents rapid submissions', async ({ authenticatedPage }) => {
  const submitUrl = 'http://localhost:5173/submit';

  // Rapidly submit clips to trigger rate limit
  for (let i = 0; i < 10; i++) {
    await authenticatedPage.goto(submitUrl);

    await authenticatedPage.fill('[data-testid="clip-url"]',
      `https://twitch.tv/test/clip/${i}`
    );

    const submitBtn = authenticatedPage.locator('[data-testid="submit"]');
    await submitBtn.click();

    // After threshold (e.g., 5), should trigger rate limit
    if (i >= 5) {
      const errorMsg = authenticatedPage.locator('[data-testid="error-message"]');
      await expect(errorMsg).toContainText(/rate|limit|too many/i);
      break;
    }
  }
});
```

**Key Changes**:
* Import from `../fixtures`
* Use `authenticatedPage` for authenticated requests
* Rapidly make requests to trigger rate limit
* Check for error message or 429 response
* Remove test.skip() call

## Category 5: API-Level Tests

**Current Status**: Comment skips in test code

**Examples**: Various test files with API mocking

### Before

```typescript
import { test, expect } from '@playwright/test';

test('create clip via API', async ({ page }) => {
  // No fixture support for API calls
  const response = await page.request.post('/api/v1/clips', {
    // ... hardcoded data
  });

  // Manual assertion
  expect(response.ok()).toBe(true);
});
```

### After

```typescript
import { test, expect } from '../fixtures';

test('create clip via API', async ({ apiUtils }) => {
  // Use apiUtils for clean API calls
  const clip = await apiUtils.requestJson('/api/v1/clips', {
    method: 'POST',
    data: {
      title: 'Test Clip',
      url: 'https://twitch.tv/test',
      hasVideo: true,
    },
  });

  expect(clip.id).toBeDefined();
  expect(clip.title).toBe('Test Clip');

  // Can also create channels and users
  const channel = await apiUtils.createChannel({
    name: 'Test Channel',
    ownerId: clip.creatorId,
  });

  expect(channel.id).toBeDefined();
});
```

**Key Changes**:
* Import from `../fixtures`
* Use `apiUtils` instead of raw `page.request`
* apiUtils handles authentication automatically
* Use `requestJson()` for automatic parsing
* Cleaner assertions with typed responses

## Template: Converting Your Tests

Use this template to convert any skipped test:

```typescript
// 1. Change import
import { test, expect } from '../fixtures';  // Changed from '@playwright/test'

test('your test name', async ({
  page,                    // ← For normal navigation
  authenticatedPage,       // ← For authenticated user
  multiUserContexts,       // ← For multi-user scenarios
  apiUtils,               // ← For API testing
  testClips,              // ← For test data
  testUser,               // ← For test user
  // ... any other fixtures you need
}) => {
  // 2. Remove test.skip() calls

  // 3. Use the fixture in your test
  if (needsAuth) {
    await authenticatedPage.goto('/some-page');
  }

  if (needsMultiUser) {
    const { admin, regular } = multiUserContexts;
    await admin.page.goto('/admin');
  }

  if (needsApi) {
    const result = await apiUtils.requestJson('/api/v1/clips');
  }
});
```

## Quick Fixture Reference

| Fixture | Purpose | Example |
| --- | --- | --- |
| `authenticatedPage` | Pre-authenticated regular user | `await authenticatedPage.goto('/submit')` |
| `adminUser` | Admin user object | `adminUser.email`, `adminUser.role` |
| `multiUserContexts` | Multiple users with roles | `multiUserContexts.admin.page` |
| `apiUtils` | API helper functions | `apiUtils.requestJson('/api/v1/clips')` |
| `apiUrl` | Base API URL | `'http://localhost:8080/api/v1'` |
| `testUser` | Auto-created test user | `testUser.id`, `testUser.email` |
| `testClip` | Auto-created test clip | `testClip.id`, `testClip.title` |
| `testSubmission` | Auto-created submission | `testSubmission.id` |

## Verification Checklist

Before committing your changes:

* [ ] Removed all `test.skip()` calls from the test
* [ ] Imported from `../fixtures` not `@playwright/test`
* [ ] Used appropriate fixture for the test scenario
* [ ] Test runs without errors: `npm run test:e2e -- --project=chromium`
* [ ] Test passes consistently (run 2-3 times): `npm run test:e2e`
* [ ] No new skip conditions added

## Running Modified Tests

```bash
# Run all E2E tests
npm run test:e2e

# Run specific test file
npm run test:e2e cdn-failover.spec.ts

# Run with UI mode to watch
npm run test:e2e -- --ui

# Show HTML report
npx playwright show-report
```

## Common Patterns

### Testing Admin vs User Access

```typescript
test('admin can delete, user cannot', async ({ multiUserContexts }) => {
  const { admin, regular } = multiUserContexts;

  // Test as regular user
  await regular.page.goto('/channel/123');
  await expect(regular.page.locator('[data-testid="delete"]')).not.toBeVisible();

  // Test as admin
  await admin.page.goto('/channel/123');
  await expect(admin.page.locator('[data-testid="delete"]')).toBeVisible();
});
```

### Testing with Test Data

```typescript
test('comment on clip', async ({ testClip, apiUtils }) => {
  // testClip already created
  const comment = await apiUtils.requestJson(
    `/api/v1/clips/${testClip.id}/comments`,
    {
      method: 'POST',
      data: { text: 'Great clip!' },
    }
  );

  expect(comment.text).toBe('Great clip!');
  // Both clip and comment auto-deleted after test
});
```

### Testing Video Playback

```typescript
test('video plays in player', async ({ page, testClips }) => {
  // testClips.withVideo has HLS URL
  await page.goto(`/clips/${testClips.withVideo.id}`);

  const player = page.locator('[data-testid="video-player"]');
  await expect(player).toBeVisible();
});
```

## Support

For issues when converting tests:

1. Check [E2E_TEST_FIXTURES_GUIDE.md](./E2E_TEST_FIXTURES_GUIDE.md) for detailed examples
2. Review fixtures/index.ts to see all available fixtures
3. Run with verbose output: `DEBUG=pw:api npm run test:e2e`
4. Check HTML report: `npx playwright show-report`

## Progress Tracking

As you enable tests:

* [ ] Category 1: Authenticated tests (4 tests)
* [ ] Category 2: Admin/Moderator tests (10+ tests)
* [ ] Category 3: Video playback tests (5 tests)
* [ ] Category 4: Rate limiting tests (2 tests)
* [ ] Category 5: Multi-user tests (10+ tests)
* [ ] Category 6: Other skipped tests (15+ tests)

Target: **353 tests passing, 0 tests skipped**
