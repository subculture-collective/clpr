import { expect, test } from '@playwright/test';

const apiBaseUrl =
    process.env.PLAYWRIGHT_API_BASE_URL || 'http://127.0.0.1:18088';

test.describe('real-backend launch smoke', () => {
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

    test('submission requires authentication', async ({ page }) => {
        await page.goto('/submit');
        await expect(page).toHaveURL(/\/login$/);
        await expect(
            page.getByRole('button', { name: /continue with twitch/i }),
        ).toBeVisible();
    });
});
