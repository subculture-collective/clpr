import { expect, test, type Page } from '@playwright/test';

// Response shapes mirror production responses captured on 2026-09-25.
const clip = (index: number) => ({
    id: `ee5a41bc-a1da-4221-b359-73d32d08545${index}`,
    twitch_clip_id: `FrailPluckyCod-${index}`,
    twitch_clip_url: `https://www.twitch.tv/twitch/clip/FrailPluckyCod-${index}`,
    embed_url: `https://clips.twitch.tv/embed?clip=FrailPluckyCod-${index}`,
    title: `Patch Notes clip ${index}`,
    creator_name: 'Twitch',
    creator_id: '12826',
    broadcaster_name: 'Twitch',
    broadcaster_id: '12826',
    game_id: '509658',
    language: 'en',
    thumbnail_url: 'https://placehold.co/480x272',
    duration: 23.3,
    view_count: 18713,
    created_at: '2026-09-23T20:25:49Z',
    imported_at: '2026-09-24T01:13:52Z',
    vote_score: 0,
    comment_count: 0,
    favorite_count: 0,
    is_featured: false,
    is_nsfw: false,
    is_removed: false,
});

const topic = (slug: string, name: string) => ({
    id: `topic-${slug}`,
    name,
    slug,
    position: 1,
    category_type: 'topic',
    is_featured: true,
    is_public: true,
    created_at: '2026-08-11T22:42:00Z',
    updated_at: '2026-08-14T17:42:40Z',
});

const welcomeThread = {
    id: 'bb68d38d-d6e1-4d6f-928d-0fc5e1a67f0a',
    user_id: 'eb60a942-3541-4471-b906-251c0663c13c',
    username: 'migration-system-admin',
    title: 'Welcome to the clpr Forum!',
    content: 'This is your space to discuss clips.',
    tags: ['meta'],
    view_count: 2,
    reply_count: 0,
    locked: false,
    pinned: true,
    created_at: '2026-08-10T19:45:00Z',
    updated_at: '2026-08-10T19:45:00Z',
};

const error404 = (message: string) => ({ status: 404, body: { error: message } });

/** Serves the API from fixtures and records every API request path+query. */
async function mockApi(page: Page) {
    const requests: string[] = [];
    await page.route('**/api/v1/**', async route => {
        const url = new URL(route.request().url());
        const path = url.pathname.replace(/^.*\/api\/v1/, '');
        requests.push(path + url.search);
        let status = 200;
        let body: unknown = {};

        if (path === '/auth/me' || path === '/auth/refresh') {
            status = 401;
            body = { success: false, error: { code: 'UNAUTHORIZED', message: 'Missing authentication token' } };
        } else if (path === '/feeds/clips') {
            body = {
                clips: [clip(1), clip(2), clip(3)],
                pagination: { cursor: '', has_more: false, limit: 20, offset: 0, total: 3, total_pages: 1 },
                success: true,
            };
        } else if (/^\/clips\/[^/]+\/tags$/.test(path)) {
            body = { tags: [] };
        } else if (path === '/categories') {
            body = { categories: [topic('news-politics', 'News & Politics'), topic('gaming', 'Gaming')] };
        } else if (path === '/tags') {
            body = { tags: [] };
        } else if (path === '/broadcasters/popular') {
            body = { broadcasters: [{ broadcaster_id: '12826', broadcaster_name: 'Twitch', clip_count: 6 }] };
        } else if (path === '/broadcasters/rankings') {
            body = { data: null, meta: { limit: 5, offset: 0, total: 0 }, success: true };
        } else if (path === '/playlists/featured') {
            body = { data: [], meta: { page: 1, limit: Number(url.searchParams.get('limit')), total: 0 }, success: true };
        } else if (path === '/forum/threads') {
            // The released backend rejects every sort name except these three.
            const sort = url.searchParams.get('sort');
            if (sort && !['recent', 'popular', 'replies'].includes(sort)) {
                status = 400;
                body = { error: 'sort must be recent, popular, or replies' };
            } else if (url.searchParams.has('game_id')) {
                status = 400;
                body = { error: 'unknown parameter game_id' };
            } else {
                body = { success: true, data: [welcomeThread], meta: { page: 1, limit: 20, count: 1 } };
            }
        } else if (path.startsWith('/twitch-categories/999999999')) {
            ({ status, body } = error404('Game not found'));
        } else if (path === '/creators/Twitch/analytics/overview') {
            ({ status, body } = error404('creator analytics not found'));
        } else if (path === '/creators/Twitch/analytics/clips') {
            body = { clips: [{ ...clip(1), views: 0, engagement_rate: 0 }], count: 1 };
        } else if (path === '/creators/Twitch/analytics/trends') {
            body = { data: null, days: 30, metric: url.searchParams.get('metric') };
        } else if (path === '/creators/Twitch/analytics/audience') {
            body = { top_countries: [], device_types: [], total_views: 0 };
        }

        await route.fulfill({ status, contentType: 'application/json', body: JSON.stringify(body) });
    });
    await page.route('https://placehold.co/**', route => route.fulfill({ status: 404, body: '' }));
    await page.route('https://static-cdn.jtvnw.net/**', route => route.fulfill({ status: 404, body: '' }));
    return requests;
}

function countBy(requests: string[], pattern: RegExp) {
    return requests.filter(request => pattern.test(request.split('?')[0])).length;
}

test('the home page requests each shared resource once', async ({ page }) => {
    const requests = await mockApi(page);
    await page.goto('/');

    await expect(page.getByText('Patch Notes clip 1').first()).toBeVisible();
    await page.waitForLoadState('networkidle');

    for (const path of ['/categories', '/tags', '/playlists/featured', '/broadcasters/popular', '/feeds/clips']) {
        expect(countBy(requests, new RegExp(`^${path}$`)), path).toBe(1);
    }
    const tagRequests = requests.filter(request => /^\/clips\/[^/]+\/tags$/.test(request));
    expect(tagRequests.length).toBeGreaterThan(0);
    expect(new Set(tagRequests).size, 'one tag request per clip').toBe(tagRequests.length);
});

test('the forum lists threads using the backend sort names', async ({ page }) => {
    const requests = await mockApi(page);
    await page.goto('/forum?sort=trending');

    await expect(page.getByText('Welcome to the clpr Forum!')).toBeVisible();
    await expect(page.getByText('Failed to load threads')).toHaveCount(0);
    expect(requests.filter(request => request.startsWith('/forum/threads'))).toEqual([
        '/forum/threads?page=1&limit=20&sort=popular',
    ]);

    await page.getByRole('combobox', { name: 'Sort discussions' }).selectOption('most-replied');
    await expect.poll(() => requests.at(-1)).toBe('/forum/threads?page=1&limit=20&sort=replies');
    await expect(page.getByText('Welcome to the clpr Forum!')).toBeVisible();
});

test('an unknown Twitch category shows a not-found state', async ({ page }) => {
    await mockApi(page);
    await page.goto('/game/999999999');

    await expect(page.getByRole('heading', { level: 1, name: "This Twitch category isn't here" })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Browse topics' })).toHaveAttribute('href', '/topics');
    await expect(page.getByText(/Failed to load/)).toHaveCount(0);
});

test('creator analytics without a summary fits a phone', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    await mockApi(page);
    await page.goto('/creator/Twitch/analytics');

    await expect(page.getByText('Summary not available yet')).toBeVisible();
    const group = page.getByRole('group', { name: 'Date range selection' });
    await group.scrollIntoViewIfNeeded();
    for (const button of await group.getByRole('button').all()) {
        const box = await button.boundingBox();
        expect(box && box.x + box.width).toBeLessThanOrEqual(390);
    }
    expect(await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth)).toBeLessThanOrEqual(0);
});
