import { expect, test } from '@playwright/test';

test.describe('mocked UI smoke', () => {
    test('renders the browser extension release surface', async ({ page }) => {
        await page.goto('/extension');

        await expect(
            page.getByRole('heading', { name: 'Clipper Browser Extension' }),
        ).toBeVisible();
        await expect(
            page.getByRole('link', { name: 'Get Clipper for Chrome' }),
        ).toHaveAttribute('href', /chrome|extension/);
        await expect(
            page.getByRole('link', { name: 'Get Clipper for Firefox' }),
        ).toHaveAttribute('href', /addons\.mozilla\.org/);
    });
});
