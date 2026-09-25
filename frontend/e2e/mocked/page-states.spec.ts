import { expect, test, type Page } from '@playwright/test';

// Response shapes mirror production responses captured on 2026-09-24.
const broadcaster = {
    broadcaster_id: '121059319',
    broadcaster_name: 'MOONMOON',
    display_name: 'MOONMOON',
    avatar_url: 'https://static-cdn.jtvnw.net/jtv_user_pictures/missing-profile_image-300x300.png',
    bio: 'BUSINESS INQUIRIES ONLY: MOONMOON@loaded.gg',
    twitch_url: 'https://twitch.tv/MOONMOON',
    total_clips: 120,
    follower_count: 0,
    total_views: 52597,
    avg_vote_score: 0,
    is_following: false,
};

const broadcasterClip = (index: number, title: string, viewCount: number) => ({
    id: `79334c21-2421-4422-b3d7-e83eaad7011${index}`,
    twitch_clip_id: `SpoopyPoisedStar-${index}`,
    twitch_clip_url: `https://www.twitch.tv/moonmoon/clip/SpoopyPoisedStar-${index}`,
    embed_url: `https://clips.twitch.tv/embed?clip=SpoopyPoisedStar-${index}`,
    title,
    creator_name: 'MOOnMOOn_hAS_TIny_TeeTh',
    creator_id: '23234194',
    broadcaster_name: 'MOONMOON',
    broadcaster_id: '121059319',
    game_id: '508455',
    language: 'en',
    thumbnail_url: 'https://placehold.co/480x272',
    duration: 28,
    view_count: viewCount,
    created_at: '2026-09-17T23:30:00Z',
    imported_at: '2026-09-18T00:00:00Z',
    vote_score: 0,
    comment_count: 0,
    favorite_count: 0,
    is_featured: false,
    is_nsfw: false,
    is_removed: false,
});

const zeroScoreEntries = Array.from({ length: 50 }, (_, index) => ({
    rank: index + 1,
    user_id: `00000000-0000-4000-8000-${String(index).padStart(12, '0')}`,
    username: `member_${index}`,
    display_name: `Member ${index}`,
    score: 0,
    user_rank: 'Newcomer',
}));

async function mockApi(page: Page, authRequests: string[] = []) {
    await page.route('**/api/v1/**', async route => {
        const url = new URL(route.request().url());
        const path = url.pathname.replace(/^.*\/api\/v1/, '');
        let status = 200;
        let body: unknown = {};

        if (path === '/auth/me' || path === '/auth/refresh') {
            authRequests.push(path);
            status = 401;
            body = { error: { code: 'UNAUTHORIZED', message: 'Missing authentication token' }, success: false };
        } else if (path === '/clips/00000000-0000-0000-0000-000000000000') {
            status = 404;
            body = { success: false, error: { code: 'CLIP_NOT_FOUND', message: 'Clip not found or has been removed' } };
        } else if (path === '/clips/11111111-1111-4111-8111-111111111111') {
            status = 503;
            body = { error: 'service unavailable' };
        } else if (path.startsWith('/leaderboards/')) {
            body = { entries: zeroScoreEntries, limit: 50, page: Number(url.searchParams.get('page') || 1), type: path.split('/')[2] };
        } else if (path === '/broadcasters/rankings') {
            body = { data: null, meta: { limit: 100, offset: 0, total: 0 }, success: true };
        } else if (path === '/broadcasters/121059319') {
            body = broadcaster;
        } else if (path === '/broadcasters/121059319/live-status') {
            body = { broadcaster_id: '121059319', is_live: false, is_stale: false, viewer_count: 0 };
        } else if (path === '/broadcasters/121059319/clips') {
            body = {
                data: [
                    broadcasterClip(0, 'death2', 12765),
                    broadcasterClip(1, 'Moonstradamus', 13212),
                    broadcasterClip(2, 'Big Chill', 2527),
                    broadcasterClip(3, 'Zomboid mods', 4409),
                ],
                meta: { limit: 20, page: 1, total_items: 4, total_pages: 1 },
                success: true,
            };
        } else if (/categories|tags|playlists|broadcasters/.test(path)) {
            body = { data: [], categories: [], tags: [], playlists: [] };
        }

        await route.fulfill({ status, contentType: 'application/json', body: JSON.stringify(body) });
    });
    // Keep third-party avatars offline and failing so the placeholder path is exercised.
    await page.route('https://static-cdn.jtvnw.net/**', route => route.fulfill({ status: 404, body: '' }));
    await page.route('https://placehold.co/**', route => route.fulfill({ status: 404, body: '' }));
}

async function horizontalOverflow(page: Page) {
    return page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth);
}

test('a missing clip shows a friendly not-found state', async ({ page }) => {
    await mockApi(page);
    await page.goto('/clip/00000000-0000-0000-0000-000000000000');

    await expect(page.getByRole('heading', { level: 1, name: "This clip isn't here" })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Search clips' })).toHaveAttribute('href', '/search');
    await expect(page.getByText(/status code/i)).toHaveCount(0);
});

test('a failing clip request offers a retry without the transport message', async ({ page }) => {
    await mockApi(page);
    await page.goto('/clip/11111111-1111-4111-8111-111111111111');

    await expect(page.getByRole('alert').filter({ hasText: "We couldn't load this clip" })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Try again' })).toBeVisible();
    await expect(page.getByText(/status code/i)).toHaveCount(0);
});

test('an all-zero leaderboard explains itself and fits a phone', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    await mockApi(page);
    await page.goto('/leaderboards');

    await expect(page.getByRole('heading', { name: 'No rankings yet' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Next' })).toHaveCount(0);
    await expect(page.getByRole('button', { name: /Creators/ })).toBeInViewport();
    expect(await horizontalOverflow(page)).toBeLessThanOrEqual(0);
});

test('broadcaster clip cards keep their metadata inside the card', async ({ page }) => {
    await mockApi(page);
    for (const width of [1280, 1024, 768, 390]) {
        await page.setViewportSize({ width, height: 800 });
        await page.goto('/broadcaster/121059319');
        await expect(page.getByTestId('clip-grid-card')).toHaveCount(4);
        expect(await horizontalOverflow(page), `overflow at ${width}px`).toBeLessThanOrEqual(0);
    }
    // The broadcaster avatar 404s here, so the initial placeholder is shown instead of alt text.
    await expect(page.getByTestId('avatar-fallback').first()).toHaveText('M');
});

test('logged-out visitors do not probe the session endpoint', async ({ page }) => {
    const authRequests: string[] = [];
    const consoleErrors: string[] = [];
    page.on('console', message => {
        if (message.type() === 'error' && /401/.test(message.text())) consoleErrors.push(message.text());
    });
    await mockApi(page, authRequests);
    await page.goto('/leaderboards');

    await expect(page.getByRole('heading', { level: 1, name: 'Leaderboards' })).toBeVisible();
    expect(authRequests).toEqual([]);
    expect(consoleErrors).toEqual([]);
});
