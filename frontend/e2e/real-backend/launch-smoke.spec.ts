import { expect, test } from '@playwright/test';

const apiBaseUrl =
    process.env.PLAYWRIGHT_API_BASE_URL || 'http://127.0.0.1:18088';
const seedClipId =
    process.env.PLAYWRIGHT_SEED_CLIP_ID ||
    '00000000-0000-4000-8000-000000000001';
const seedGameId = process.env.PLAYWRIGHT_SEED_GAME_ID || 'release-game';

type SearchPayload = {
    results: {
        clips: Array<Record<string, unknown>>;
        creators: Array<Record<string, unknown>>;
        games: Array<Record<string, unknown>>;
        tags: Array<Record<string, unknown>>;
    };
};

function expectStableSearchArrays(payload: SearchPayload) {
    expect(Array.isArray(payload.results.clips)).toBe(true);
    expect(Array.isArray(payload.results.creators)).toBe(true);
    expect(Array.isArray(payload.results.games)).toBe(true);
    expect(Array.isArray(payload.results.tags)).toBe(true);
}

test.describe('real-backend launch smoke', () => {
    test.beforeAll(async ({ request }) => {
        const fixture = await request.get(`${apiBaseUrl}/api/v1/clips/${seedClipId}`);
        expect(fixture.status(), 'seeded search fixture must exist').toBe(200);
        await expect(fixture.json()).resolves.toMatchObject({
            data: {
                id: seedClipId,
                title: 'CLPR release smoke clip',
                game_id: seedGameId,
            },
        });
    });

    test('backend liveness is healthy', async ({ request }) => {
        const response = await request.get(`${apiBaseUrl}/health/live`);
        expect(response.status()).toBe(200);
        await expect(response.json()).resolves.toMatchObject({
            status: 'alive',
        });
    });

    test('public web shell loads against the real API', async ({ page }) => {
        const response = await page.goto('/');
        expect(response?.ok()).toBe(true);
        await expect(page.locator('main')).toBeVisible();
        await expect(
            page.getByRole('heading', { name: /trending/i }),
        ).toBeVisible();
        await expect(
            page.getByRole('heading', { name: 'Curated Collections' }),
        ).toBeVisible();
    });

    test('public empty search returns stable arrays and renders without a page error', async ({
        page,
    }) => {
        const pageErrors: Error[] = [];
        page.on('pageerror', error => pageErrors.push(error));
        const searchResponse = page.waitForResponse(
            response =>
                response.url().includes('/api/v1/search?') &&
                response.request().method() === 'GET',
        );

        await page.goto('/search?q=zzzzqvnonexistent9f731b');

        const response = await searchResponse;
        expect(response.status()).toBe(200);
        const payload = (await response.json()) as {
            results: Record<string, unknown[]>;
        };
        expect(payload.results).toMatchObject({
            clips: [],
            creators: [],
            games: [],
            tags: [],
        });
        await expect(page.getByTestId('empty-state')).toBeVisible();
        expect(pageErrors).toEqual([]);
    });

    test('public populated search returns the seeded real-backend fixture', async ({
        page,
    }) => {
        const searchResponse = page.waitForResponse(
            response => response.url().includes('/api/v1/search?') && response.request().method() === 'GET',
        );

        await page.goto('/search?q=CLPR');

        const response = await searchResponse;
        expect(response.status()).toBe(200);
        const payload = await response.json() as SearchPayload;
        expectStableSearchArrays(payload);
        expect(payload.results.clips).toEqual(
            expect.arrayContaining([expect.objectContaining({ id: seedClipId })]),
        );
        await expect(page.getByText('CLPR release smoke clip')).toBeVisible();
    });

    test('public filtered search applies the seeded game filter', async ({ request }) => {
        const response = await request.get(`${apiBaseUrl}/api/v1/search`, {
            params: { q: 'CLPR', type: 'clips', game_id: seedGameId },
        });

        expect(response.status()).toBe(200);
        const payload = await response.json() as SearchPayload;
        expectStableSearchArrays(payload);
        expect(payload.results.clips.length).toBeGreaterThan(0);
        expect(payload.results.clips).toEqual(
            expect.arrayContaining([expect.objectContaining({ id: seedClipId, game_id: seedGameId })]),
        );
    });

    test('failed search requests return a stable client error without result data', async ({ request }) => {
        const response = await request.get(`${apiBaseUrl}/api/v1/search`);

        expect(response.status()).toBe(400);
        await expect(response.json()).resolves.toEqual({
            error: "Query parameter 'q' is required",
        });
    });

    test('public clip detail loads repository data from the real API', async ({
        page,
    }) => {
        const clipResponse = page.waitForResponse(
            (response) =>
                response.url().includes(`/api/v1/clips/${seedClipId}`) &&
                response.request().method() === 'GET',
        );

        await page.goto(`/clips/${seedClipId}`);

        expect((await clipResponse).status()).toBe(200);
        await expect(
            page.getByRole('heading', { name: 'CLPR release smoke clip' }),
        ).toBeVisible();
        await expect(
            page.getByText('Release Channel', { exact: true }),
        ).toBeVisible();
    });

    test('submission requires authentication', async ({ page }) => {
        await page.goto('/submit');
        await expect(page).toHaveURL(/\/login$/);
        await expect(
            page.getByRole('button', { name: /continue with twitch/i }),
        ).toBeVisible();
    });

    test('account settings require authentication', async ({ page }) => {
        await page.goto('/settings');
        await expect(page).toHaveURL(/\/login$/);
    });

    test('administration requires authentication', async ({ page }) => {
        await page.goto('/admin/dashboard');
        await expect(page).toHaveURL(/\/login$/);
    });
});
