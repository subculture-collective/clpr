import { expect, test } from '@playwright/test';

test('analytics starts only after consent and stops after withdrawal', async ({
    page,
}) => {
    await page.addInitScript(() => {
        localStorage.removeItem('clpr_consent_preferences');
        window.__CLPR_ANALYTICS_CONFIG__ = {
            enabled: true,
            autoConsent: false,
            googleMeasurementId: 'G-E2E-CONSENT',
            postHogApiKey: 'ph_e2e_consent',
            postHogHost: 'https://analytics.example.test',
        };
    });

    const analyticsRequests: string[] = [];
    await page.route(/googletagmanager\.com|posthog\.com|analytics\.example\.test|us-assets\.i\.posthog\.com/, route => {
        analyticsRequests.push(route.request().url());
        return route.fulfill({ status: 204, body: '' });
    });

    await page.goto('/');
    await expect(
        page.getByRole('heading', { name: 'Privacy & Cookie Preferences' }),
    ).toBeVisible();
    expect(analyticsRequests).toHaveLength(0);

    await page.getByRole('button', { name: 'Accept All' }).click();
    await expect.poll(() => analyticsRequests.length).toBeGreaterThan(0);

    await page.goto('/settings/cookies');
    await page
        .getByRole('button', { name: 'Reject All Optional Cookies' })
        .click();
    const requestsAtWithdrawal = analyticsRequests.length;

    await page.reload();
    await page.waitForTimeout(250);
    expect(analyticsRequests).toHaveLength(requestsAtWithdrawal);
});
