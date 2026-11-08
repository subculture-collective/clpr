import { expect, test, type Page } from '@playwright/test';

const clip = (position: number, title: string, broadcaster: string) => ({
    position,
    clip_id: `00000000-0000-4000-8000-00000000000${position}`,
    title,
    broadcaster_name: broadcaster,
    broadcaster_id: `10${position}`,
    twitch_clip_id: `TwitchClip-${position}`,
    twitch_clip_url: `https://clips.twitch.tv/TwitchClip-${position}`,
    thumbnail_url: 'https://placehold.co/480x272',
    duration: 30,
});

const edition = {
    edition_date: '2026-11-07',
    series_slug: 'overall',
    series_name: 'Overall',
    size: 3,
    clips: [clip(1, 'The ace nobody saw coming', 'alpha'), clip(2, 'Chat loses it', 'bravo'), clip(3, 'A very long title that has to wrap on a phone without pushing the page sideways', 'charlie')],
};

async function mockApi(page: Page) {
    await page.route('**/api/v1/**', async route => {
        const path = new URL(route.request().url()).pathname.replace(/^.*\/api\/v1/, '');
        let status = 200;
        let body: unknown = { data: [] };
        if (path === '/auth/me' || path === '/auth/refresh') {
            status = 401;
            body = { success: false, error: { code: 'UNAUTHORIZED', message: 'Missing authentication token' } };
        } else if (path === '/countdowns') {
            body = { success: true, data: [edition], meta: { next_before: null } };
        } else if (path === '/countdowns/2026-11-07/overall') {
            body = { success: true, data: edition };
        } else if (path.startsWith('/countdowns/')) {
            status = 404;
            body = { success: false, error: { code: 'NOT_FOUND', message: 'Countdown not found' } };
        }
        await route.fulfill({ status, contentType: 'application/json', body: JSON.stringify(body) });
    });
    await page.route('https://placehold.co/**', route => route.fulfill({ status: 404, body: '' }));
}

test('a visitor opens a daily countdown and plays clips one at a time', async ({ page }) => {
    await mockApi(page);
    await page.goto('/countdowns');

    await expect(page.getByRole('heading', { level: 1, name: 'Daily countdowns' })).toBeVisible();
    await page.getByRole('link', { name: /Overall/ }).click();

    await expect(page).toHaveURL(/\/countdowns\/2026-11-07\/overall$/);
    await expect(page.getByRole('heading', { level: 1, name: 'Top 3: Overall' })).toBeVisible();
    await expect(page.getByRole('listitem', { name: /^Number \d+$/ })).toHaveText([/#3/, /#2/, /#1/]);
    await expect(page.locator('iframe')).toHaveCount(0);
    await expect(page).toHaveTitle(/Top 3 Twitch clips: Overall, November 7, 2026/);
});

test('a countdown fits a phone screen', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    await mockApi(page);
    await page.goto('/countdowns/2026-11-07/overall');

    await expect(page.getByRole('heading', { level: 1, name: 'Top 3: Overall' })).toBeVisible();
    expect(await page.evaluate(() => document.documentElement.scrollWidth - window.innerWidth)).toBeLessThanOrEqual(0);
});

test('an unknown countdown shows the not-found page', async ({ page }) => {
    await mockApi(page);
    await page.goto('/countdowns/2026-11-01/nope');
    await expect(page.getByRole('heading', { name: 'Page not found' })).toBeVisible();
});
