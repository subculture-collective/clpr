import { expect, test } from '@playwright/test';

test('account access is open and support is optional', async ({ page }) => {
    const checkoutRequests: string[] = [];
    await page.route('**/api/v1/subscriptions/checkout', async route => {
        checkoutRequests.push(new URL(route.request().url()).pathname);
        return route.fulfill({ status: 410, contentType: 'application/json', body: '{"error":"disabled"}' });
    });

    await page.goto('/support');

    await expect(page.getByRole('heading', { name: /clpr is for the culture/i })).toBeVisible();
    await expect(page.getByText('No account tiers')).toBeVisible();
    await expect(page.getByText('No feature paywalls')).toBeVisible();
    await expect(page.getByRole('link', { name: /support subcult on patreon/i })).toHaveAttribute(
        'href',
        'https://patreon.com/subcult',
    );
    expect(checkoutRequests).toEqual([]);
});
