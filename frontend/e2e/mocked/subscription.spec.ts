import { expect, test } from '@playwright/test';

test('checkout initiation and return wait for confirmed backend entitlement', async ({ page }) => {
    let subscriptionReads = 0;
    const checkoutRequests: unknown[] = [];
    await page.route('**/api/v1/**', async route => {
        const request = route.request();
        const pathname = new URL(request.url()).pathname;
        if (pathname.endsWith('/auth/me')) {
            return route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    id: '00000000-0000-0000-0000-000000000101',
                    twitch_id: '101',
                    username: 'checkout-user',
                    display_name: 'Checkout User',
                    role: 'user',
                    account_type: 'member',
                    account_status: 'active',
                    is_banned: false,
                    created_at: '2026-01-01T00:00:00Z',
                    updated_at: '2026-01-01T00:00:00Z',
                }),
            });
        }
        if (pathname.endsWith('/subscriptions/checkout')) {
            checkoutRequests.push(request.postDataJSON());
            return route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    session_id: 'cs_mock_contract',
                    session_url: 'http://127.0.0.1:5173/subscription/success?session_id=cs_mock_contract',
                }),
            });
        }
        if (pathname.endsWith('/subscriptions/me')) {
            subscriptionReads += 1;
            const active = subscriptionReads > 1;
            return route.fulfill({
                status: 200,
                contentType: 'application/json',
                body: JSON.stringify({
                    id: 'subscription-contract',
                    user_id: '00000000-0000-0000-0000-000000000101',
                    stripe_customer_id: 'cus_contract',
                    status: active ? 'active' : 'inactive',
                    tier: active ? 'pro' : 'free',
                    cancel_at_period_end: false,
                    created_at: '2026-01-01T00:00:00Z',
                    updated_at: '2026-01-01T00:00:00Z',
                }),
            });
        }
        return route.fulfill({ status: 200, contentType: 'application/json', body: '{}' });
    });

    await page.goto('/pricing');
    await expect(page.getByRole('heading', { name: /upgrade to clpr pro/i })).toBeVisible();
    await page.getByRole('button', { name: 'Subscribe Now' }).click();

    await expect.poll(() => checkoutRequests).toEqual([{ price_id: 'price_e2e_monthly' }]);
    await expect(page).toHaveURL(/subscription\/success\?session_id=cs_mock_contract/);
    await expect(page.getByRole('heading', { name: 'Confirming your subscription' })).toBeVisible();
    await expect(page.getByText('Welcome to clpr Pro!')).toBeVisible({ timeout: 5_000 });
    expect(subscriptionReads).toBeGreaterThanOrEqual(2);
});
