import { defineConfig, devices } from '@playwright/test';

const apiOrigin =
    process.env.PLAYWRIGHT_API_BASE_URL || 'http://127.0.0.1:18088';

/**
 * Playwright E2E Test Configuration
 *
 * This configuration sets up comprehensive E2E testing with:
 * - Configurable base URL for local/staging/production environments
 * - Proper timeouts for global (30s) and expect (5s) operations
 * - Retry logic (2 on CI, 0 locally)
 * - Parallel workers (4 on CI)
 * - Screenshot, video, and trace capture on failures
 * - Global setup/teardown for test data management
 *
 * @see https://playwright.dev/docs/test-configuration
 */
export default defineConfig({
    testDir: './e2e',

    /* Maximum time one test can run for */
    timeout: 30 * 1000,

    /* Maximum time expect() should wait for the condition to be met */
    expect: {
        timeout: 5 * 1000,
    },

    /* Run tests in files in parallel */
    fullyParallel: true,

    /* Fail the build on CI if you accidentally left test.only in the source code */
    forbidOnly: !!process.env.CI,

    /* Retry on CI only - 2 retries as per requirements */
    retries: process.env.CI ? 2 : 0,

    /* Keep local and CI concurrency identical so Vite cold-start behavior is deterministic. */
    workers: 4,

    /* Reporter to use - HTML format with CI-friendly list reporter */
    reporter: [
        ['html', { outputFolder: 'playwright-report', open: 'never' }],
        ['list'],
    ],

    /* Shared settings for all the projects below */
    use: {
        /* Base URL - configurable via environment variable for local/staging/production */
        baseURL:
            process.env.PLAYWRIGHT_BASE_URL ||
            process.env.VITE_APP_URL ||
            'http://127.0.0.1:5173',

        /* Collect trace on first retry as per requirements */
        trace: 'on-first-retry',

        /* Capture screenshot on failure */
        screenshot: 'only-on-failure',

        /* Capture video on failure */
        video: 'retain-on-failure',

        /* Maximum time for each action */
        actionTimeout: 10 * 1000,

        /* Maximum time for navigation */
        navigationTimeout: 30 * 1000,
    },

    /* Mocked UI checks cannot satisfy the real-backend release gate. */
    projects: [
        {
            name: 'mocked-chromium',
            use: { ...devices['Desktop Chrome'] },
            testMatch: /mocked\/.*\.spec\.ts/,
        },
        {
            name: 'real-chromium',
            use: { ...devices['Desktop Chrome'] },
            testMatch: /real-backend\/.*\.spec\.ts/,
        },
        {
            name: 'real-firefox',
            use: { ...devices['Desktop Firefox'] },
            testMatch: /real-backend\/.*\.spec\.ts/,
        },
        {
            name: 'real-webkit',
            use: { ...devices['Desktop Safari'] },
            testMatch: /real-backend\/.*\.spec\.ts/,
        },
    ],

    /* Run your local dev server before starting the tests */
    webServer: {
        command:
            `VITE_AUTO_CONSENT=true VITE_ENABLE_ANALYTICS=false VITE_API_BASE_URL=${apiOrigin}/api/v1 VITE_STRIPE_PRO_MONTHLY_PRICE_ID=price_e2e_monthly VITE_STRIPE_PRO_YEARLY_PRICE_ID=price_e2e_yearly` +
            (process.env.E2E_CDN_FAILOVER_MODE === 'true' ? ' VITE_CDN_FAILOVER_MODE=true' : '') +
            ' npm run dev -- --host 127.0.0.1',
        url: 'http://127.0.0.1:5173',
        reuseExistingServer: !process.env.CI,
        timeout: 120 * 1000, // 120 seconds for CI environments
        stdout: 'pipe',
        stderr: 'pipe',
    },
});
