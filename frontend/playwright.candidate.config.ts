import { defineConfig, devices } from '@playwright/test';

const baseURL = process.env.PLAYWRIGHT_CANDIDATE_BASE_URL;
if (!baseURL || !baseURL.startsWith('https://')) {
    throw new Error('PLAYWRIGHT_CANDIDATE_BASE_URL must be an explicit HTTPS candidate URL');
}

for (const fixture of [
    'PLAYWRIGHT_CANDIDATE_CLIP_ID',
    'PLAYWRIGHT_CANDIDATE_CLIP_TITLE',
    'PLAYWRIGHT_CANDIDATE_SEARCH_QUERY',
]) {
    if (!process.env[fixture]) {
        throw new Error(`${fixture} is required for candidate accessibility execution`);
    }
}

export default defineConfig({
    testDir: './e2e/candidate',
    timeout: 45_000,
    expect: { timeout: 10_000 },
    forbidOnly: true,
    retries: process.env.CI ? 2 : 0,
    reporter: [['html', { outputFolder: 'playwright-candidate-report', open: 'never' }], ['list']],
    use: {
        baseURL,
        trace: 'on-first-retry',
        screenshot: 'only-on-failure',
        video: 'retain-on-failure',
    },
    projects: [
        { name: 'candidate-chromium', use: { ...devices['Desktop Chrome'] } },
        { name: 'candidate-firefox', use: { ...devices['Desktop Firefox'] } },
        { name: 'candidate-webkit', use: { ...devices['Desktop Safari'] } },
    ],
});
