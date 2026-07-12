import { expect, test, type Page } from '@playwright/test';
import axe from 'axe-core';

const clip = {
    id: '11111111-2222-3333-4444-555555555555',
    twitch_clip_id: 'demo-clip',
    twitch_clip_url: 'https://clips.twitch.tv/mock-demo',
    embed_url: 'https://clips.twitch.tv/embed?clip=mock-demo&parent=localhost',
    title: 'Accessible Demo Clip',
    creator_name: 'Demo Creator',
    broadcaster_name: 'Demo Broadcaster',
    game_name: 'Demo Game',
    thumbnail_url: 'https://placehold.co/640x360',
    duration: 30,
    view_count: 100,
    vote_score: 5,
    comment_count: 0,
    favorite_count: 0,
    created_at: '2026-01-01T00:00:00Z',
    imported_at: '2026-01-01T00:00:00Z',
    is_featured: false,
    is_nsfw: false,
    is_removed: false,
};

async function mockPublicApi(page: Page) {
    await page.route('**/api/v1/**', async (route) => {
        const url = new URL(route.request().url());
        let status = 200;
        let body: unknown = {};
        if (url.pathname.endsWith('/auth/me') || url.pathname.endsWith('/auth/refresh')) {
            status = 401;
            body = { error: 'unauthorized' };
        } else if (/\/clips\/[^/]+$/.test(url.pathname)) {
            body = { success: true, data: clip };
        } else if (url.pathname.includes('/comments')) {
            body = { comments: [], total: 0, has_more: false };
        } else if (url.pathname.endsWith('/search')) {
            body = {
                query: 'accessible',
                results: { clips: [clip], creators: [], games: [], tags: [] },
                counts: { clips: 1, creators: 0, games: 0, tags: 0 },
                meta: { page: 1, limit: 20, total_items: 1, total_pages: 1 },
            };
        } else if (url.pathname.includes('/feeds/clips')) {
            body = { clips: [clip], total: 1, page: 1, total_pages: 1, has_more: false };
        } else if (url.pathname.includes('/forum/threads')) {
            body = { threads: [], total: 0 };
        } else if (url.pathname.includes('/suggestions')) {
            body = { suggestions: [] };
        } else if (/categories|tags|broadcasters|playlists|discovery-lists|games/.test(url.pathname)) {
            body = { data: [], categories: [], tags: [], broadcasters: [], playlists: [], games: [] };
        }
        await route.fulfill({ status, contentType: 'application/json', body: JSON.stringify(body) });
    });
}

async function expectNoSeriousAxeViolations(page: Page) {
    await page.addScriptTag({ content: axe.source });
    const violations = await page.evaluate(async () => {
        const result = await window.axe.run(document, {
            runOnly: { type: 'tag', values: ['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'] },
        });
        return result.violations
            .filter((violation) => violation.impact === 'critical' || violation.impact === 'serious')
            .map(({ id, impact, help, nodes }) => ({
                id,
                impact,
                help,
                nodes: nodes.map(({ target, failureSummary }) => ({ target, failureSummary })),
            }));
    });
    expect(violations).toEqual([]);
}

test.beforeEach(async ({ page }) => {
    await mockPublicApi(page);
});

for (const journey of [
    { name: 'login', path: '/login', heading: /welcome to clipper/i },
    { name: 'search', path: '/search?q=accessible', heading: /search/i },
    { name: 'clip detail', path: `/clip/${clip.id}`, heading: clip.title },
    { name: 'payment plans', path: '/pricing', heading: /upgrade to clpr pro/i },
    { name: 'forum', path: '/forum', heading: /forum discussions/i },
]) {
    test(`${journey.name} has no serious or critical axe violations`, async ({ page }) => {
        await page.goto(journey.path);
        await expect(page.getByRole('heading', { name: journey.heading }).first()).toBeVisible();
        await expectNoSeriousAxeViolations(page);
    });
}

for (const path of ['/submit', '/settings', '/admin/moderation']) {
    test(`${path} preserves the accessible authentication boundary`, async ({ page }) => {
        await page.goto(path);
        await expect(page).toHaveURL(/\/login$/);
        await expect(page.getByRole('button', { name: /continue with twitch/i })).toBeVisible();
        await expectNoSeriousAxeViolations(page);
    });
}

test('skip link transfers keyboard focus to main content', async ({ page }) => {
    await page.goto('/login');
    const skipLink = page.getByRole('link', { name: 'Skip to main content' });
    await skipLink.press('Enter');
    await expect(page.locator('#main-content')).toBeFocused();
});

test('mobile navigation has touch-sized targets and restores focus on escape', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 812 });
    await page.goto('/');
    const menuButton = page.getByRole('button', { name: 'Open menu' });
    await menuButton.click();
    const links = page.getByRole('navigation', { name: 'Mobile navigation' }).getByRole('link');
    for (let index = 0; index < await links.count(); index += 1) {
        const box = await links.nth(index).boundingBox();
        expect(box?.height).toBeGreaterThanOrEqual(44);
    }
    await page.keyboard.press('Escape');
    await expect(menuButton).toBeFocused();
    await expect(menuButton).toHaveAttribute('aria-expanded', 'false');
});

test('critical content reflows at narrow width and honors reduced motion', async ({ page }) => {
    await page.emulateMedia({ reducedMotion: 'reduce' });
    await page.setViewportSize({ width: 320, height: 568 });
    await page.goto('/search?q=accessible');
    await expect(page.getByRole('heading', { name: 'Search Results' })).toBeVisible();
    expect(await page.evaluate(() => matchMedia('(prefers-reduced-motion: reduce)').matches)).toBe(true);
    expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(320);
});

declare global {
    interface Window {
        axe: typeof axe;
    }
}
