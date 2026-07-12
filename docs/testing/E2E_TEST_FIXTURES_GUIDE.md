---
title: "E2E TEST FIXTURES GUIDE"
summary: "The fixture system provides everything needed to write comprehensive E2E tests:"
tags: ["docs","testing","guide"]
area: "docs"
status: "stable"
owner: "team-core"
version: "1.0"
last_reviewed: 2026-01-29
---

/**
 * E2E Test Fixtures Guide
 * ========================
 *
 * This guide explains how to use the enhanced Playwright test fixtures
 * to create comprehensive E2E tests with multiple users, API interactions,
 * and role-based permission testing.
 */

# E2E Test Fixtures Complete Guide

## Overview

The fixture system provides everything needed to write comprehensive E2E tests:

- **Page Objects**: Reusable page models for navigation and assertions
- **Authenticated Contexts**: Pre-authenticated pages for different user roles
- **Multi-User Contexts**: Isolated browser contexts for concurrent user testing
- **API Utilities**: Helper functions for making API requests from tests
- **Test Data**: Factories for creating consistent test data

## Standard Fixtures

### Page Objects

Page objects encapsulate UI selectors and interactions:

```typescript
import { test, expect } from '../fixtures';

test('navigate to home', async ({ homePage }) => {
  await homePage.goto();
  await homePage.verifyClipsVisible();
});

test('submit a clip', async ({ submitClipPage }) => {
  await submitClipPage.goto();
  await submitClipPage.fillUrl('https://clips.twitch.tv/test');
  await submitClipPage.submit();
});
```

Available page objects:
- `loginPage` - Authentication pages
- `homePage` - Main feed
- `clipPage` - Individual clip details
- `submitClipPage` - Clip submission form
- `adminModerationPage` - Admin moderation tools
- `searchPage` - Search interface

### Authenticated Pages

Pre-authenticated pages for different user roles:

```typescript
// Login as regular user
test('regular user can like clips', async ({ authenticatedPage }) => {
  await authenticatedPage.goto('/clips/123');
  await authenticatedPage.click('[data-testid="like-button"]');
  await expect(authenticatedPage.locator('[data-testid="like-count"]')).toContainText('1');
});

// Login as admin
test('admin can moderate content', async ({ adminUser, authenticatedPage }) => {
  // adminUser is created and authenticated
  // authenticatedPage is logged in as this user
});
```

### Test Data Fixtures

Automatically created and cleaned up after each test:

```typescript
test('clips work with comments', async ({ testClip, page }) => {
  // testClip is automatically created in the database
  await page.goto(`/clips/${testClip.id}`);

  // Test commenting on the clip
  await page.fill('[data-testid="comment-input"]', 'Great clip!');
  await page.click('[data-testid="comment-submit"]');

  // testClip is automatically deleted after the test
});
```

## Advanced Fixtures

### Multi-User Contexts

Test scenarios involving multiple users with different permissions:

```typescript
import { test, expect } from '../fixtures';

test('channel permissions work correctly', async ({ multiUserContexts }) => {
  const { admin, moderator, member, regular } = multiUserContexts;

  // Admin creates a channel as admin.page
  await admin.page.goto('/dashboard/channels');
  await admin.page.click('[data-testid="create-channel"]');
  await admin.page.fill('[data-testid="channel-name"]', 'Test Channel');
  await admin.page.click('[data-testid="create-button"]');

  // Regular user can view but cannot delete
  await regular.page.goto('/channels/test-channel');

  // Delete button should not be visible for regular user
  const deleteButton = regular.page.locator('[data-testid="delete-channel"]');
  await expect(deleteButton).not.toBeVisible();

  // But admin can delete
  await admin.page.goto('/channels/test-channel');
  const adminDeleteButton = admin.page.locator('[data-testid="delete-channel"]');
  await expect(adminDeleteButton).toBeVisible();
});
```

Multi-user context object:
```typescript
{
  user: TestUser,           // User object with email, password, role
  context: BrowserContext,  // Isolated browser context
  page: Page,               // Page object for navigation
  authToken?: string,       // Auth token if authentication succeeded
  logout: () => Promise<void> // Manual logout function
}
```

### API Utilities

Make API requests directly in tests:

```typescript
import { test, expect } from '../fixtures';

test('create and manage clips via API', async ({ apiUtils, apiUrl }) => {
  // Create a clip using API
  const clipResponse = await apiUtils.requestJson('/api/v1/clips', {
    method: 'POST',
    data: {
      title: 'Test Clip',
      url: 'https://twitch.tv/test',
      hasVideo: true,
    },
  });

  const clipId = clipResponse.id;

  // Verify clip was created
  const getResponse = await apiUtils.requestJson(`/api/v1/clips/${clipId}`);
  expect(getResponse.title).toBe('Test Clip');

  // Test fails for non-existent clip
  const notFoundResponse = await apiUtils.request('/api/v1/clips/non-existent', {
    method: 'GET',
  });
  expect(notFoundResponse.status()).toBe(404);
});
```

### API Utils Methods

```typescript
// Make a request and get raw Response
apiUtils.request(endpoint, options): Promise<Response>

// Make a request and parse JSON response
apiUtils.requestJson(endpoint, options): Promise<any>

// Create a test user
apiUtils.createUser(userData): Promise<any>

// Create a test clip
apiUtils.createClip(clipData): Promise<any>

// Create a test channel
apiUtils.createChannel(channelData): Promise<any>

// Get the API base URL
apiUrl: string
```

## Example: Complete Multi-User Test

```typescript
import { test, expect } from '../fixtures';

test.describe('Channel Moderation', () => {
  test('moderator can remove members but not owner', async ({
    multiUserContexts,
    apiUtils,
  }) => {
    const { admin, moderator, regular } = multiUserContexts;

    // Step 1: Admin creates a channel via API
    const channelResponse = await apiUtils.requestJson('/api/v1/chat/channels', {
      method: 'POST',
      data: { name: 'Test Channel', ownerId: admin.user.id },
    });

    const channelId = channelResponse.id;

    // Step 2: Admin adds moderator and regular user as members
    await apiUtils.requestJson(`/api/v1/chat/channels/${channelId}/members`, {
      method: 'POST',
      data: { userId: moderator.user.id, role: 'moderator' },
    });

    await apiUtils.requestJson(`/api/v1/chat/channels/${channelId}/members`, {
      method: 'POST',
      data: { userId: regular.user.id, role: 'member' },
    });

    // Step 3: Moderator logs in and tries to remove regular user (should succeed)
    const removeMemberResponse = await moderator.page.request.post(
      `${apiUtils.apiUrl}/api/v1/chat/channels/${channelId}/members/${regular.user.id}`,
      { method: 'DELETE' }
    );
    expect(removeMemberResponse.status()).toBe(200);

    // Step 4: Moderator tries to remove owner (should fail)
    const removeOwnerResponse = await moderator.page.request.post(
      `${apiUtils.apiUrl}/api/v1/chat/channels/${channelId}/members/${admin.user.id}`,
      { method: 'DELETE' }
    );
    expect(removeOwnerResponse.status()).toBeGreaterThanOrEqual(400);

    // Step 5: Verify owner is still in channel
    const channelDetails = await apiUtils.requestJson(`/api/v1/chat/channels/${channelId}`);
    const ownerExists = channelDetails.members.some((m: any) => m.userId === admin.user.id);
    expect(ownerExists).toBe(true);
  });
});
```

## Example: Video Playback Test

```typescript
import { test, expect } from '../fixtures';

test('HLS video playback works', async ({ page, testClips }) => {
  // testClips.withVideo has a video URL
  await page.goto(`/clips/${testClips.withVideo.id}`);

  // Wait for video player to load
  const videoPlayer = page.locator('[data-testid="video-player"]');
  await expect(videoPlayer).toBeVisible({ timeout: 5000 });

  // Verify video is playable
  const playButton = page.locator('[data-testid="play-button"]');
  await expect(playButton).toBeVisible();

  // Click play and verify video starts
  await playButton.click();

  // Check for video playing indicator
  await expect(page.locator('[data-testid="video-playing"]')).toBeVisible({ timeout: 3000 });
});
```

## Example: Rate Limiting Test

```typescript
import { test, expect } from '../fixtures';

test('rate limiting works on clip submission', async ({
  page,
  authenticatedPage,
}) => {
  const submitUrl = 'http://localhost:5173/submit';

  // Submit clips rapidly to trigger rate limit
  for (let i = 0; i < 10; i++) {
    await authenticatedPage.goto(submitUrl);

    // Fill and submit form
    await authenticatedPage.fill('[data-testid="clip-url"]',
      `https://twitch.tv/test/clip/${i}`
    );

    const submitButton = authenticatedPage.locator('[data-testid="submit-button"]');
    const response = await Promise.race([
      submitButton.click(),
      authenticatedPage.waitForResponse(
        (resp) => resp.url().includes('/api/v1/submissions')
      ),
    ]);

    // After 5 submissions, should hit rate limit
    if (i >= 5) {
      const errorMessage = authenticatedPage.locator('[data-testid="error-message"]');
      await expect(errorMessage).toContainText(/rate|limit|too many/i);
      break;
    }
  }
});
```

## Debugging Tips

### Viewing Test Data

```typescript
test('debug test data', async ({ testClip, testUser }) => {
  console.log('Test Clip:', testClip);
  console.log('Test User:', testUser);
  // Output shows the created data for debugging
});
```

### Checking Multi-User Authentication

```typescript
test('verify multi-user auth', async ({ multiUserContexts }) => {
  const { admin, moderator, regular } = multiUserContexts;

  // Check each user is authenticated
  const adminUser = await admin.page.evaluate(() => localStorage.getItem('user'));
  const modUser = await moderator.page.evaluate(() => localStorage.getItem('user'));
  const regUser = await regular.page.evaluate(() => localStorage.getItem('user'));

  console.log('Admin:', adminUser);
  console.log('Moderator:', modUser);
  console.log('Regular:', regUser);
});
```

### Enabling Detailed Logging

```typescript
// In playwright.config.ts
{
  use: {
    // ... other options
    trace: 'on', // Always capture trace
    video: 'on', // Always record video
    screenshot: 'on', // Always capture screenshots
  }
}

// Then review in html report
// npx playwright show-report
```

## Common Patterns

### Testing with Different Roles

```typescript
test.describe('Permission Matrix', () => {
  const testCases = [
    { role: 'admin', canDelete: true },
    { role: 'moderator', canDelete: true },
    { role: 'member', canDelete: false },
  ];

  testCases.forEach(({ role, canDelete }) => {
    test(`${role} can${canDelete ? '' : 'not'} delete channel`, async ({ multiUserContexts }) => {
      const user = multiUserContexts[role === 'admin' ? 'admin' :
                                   role === 'moderator' ? 'moderator' : 'regular'];

      await user.page.goto('/channels/123');

      const deleteButton = user.page.locator('[data-testid="delete-button"]');
      if (canDelete) {
        await expect(deleteButton).toBeVisible();
      } else {
        await expect(deleteButton).not.toBeVisible();
      }
    });
  });
});
```

### Test Data Dependencies

```typescript
test('comments on clips', async ({ testClip, apiUtils }) => {
  // testClip already created in database

  // Create comment on the clip
  const commentResponse = await apiUtils.requestJson(
    `/api/v1/clips/${testClip.id}/comments`,
    {
      method: 'POST',
      data: { text: 'Great clip!' },
    }
  );

  expect(commentResponse.id).toBeDefined();

  // Both comment and clip are cleaned up after test
});
```

## Troubleshooting

### "Cannot find fixture multiUserContexts"

Make sure you import from the correct fixtures file:
```typescript
import { test, expect } from '../fixtures'; // ✓ Correct
import { test, expect } from '@playwright/test'; // ✗ Wrong - uses base test
```

### Multi-user contexts not authenticated

Check that the backend is running on port 8080:
```bash
# Backend should be accessible
curl http://localhost:8080/api/v1/health
```

### API requests returning 404

Verify the API endpoint path and HTTP method:
```typescript
// Make sure to include /api/v1 prefix
await apiUtils.requestJson('/api/v1/clips', { method: 'GET' });

// Not:
await apiUtils.requestJson('/clips', { method: 'GET' }); // ✗ Missing /api/v1
```

### Test data not being cleaned up

Verify cleanup functions are called:
```typescript
test('with explicit cleanup', async ({ testClip, apiUtils }) => {
  // testClip.id is available

  // After test completes, fixture automatically calls cleanup
  // But you can manually cleanup if needed:
  await apiUtils.requestJson(`/api/v1/clips/${testClip.id}`, {
    method: 'DELETE',
  });
});
```

## Reference

- **Test Fixture Types**: Check fixtures/index.ts
- **Test Data Factories**: Check fixtures/test-data.ts
- **Multi-User Context**: Check [fixtures/multi-user-context.ts](../../frontend/e2e/fixtures/multi-user-context.ts)
- **API Utilities**: Check [fixtures/api-utils.ts](../../frontend/e2e/fixtures/api-utils.ts)
- **Playwright Docs**: https://playwright.dev/docs/test-fixtures
