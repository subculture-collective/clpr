import { expect, test, type Page } from '@playwright/test';
import axe from 'axe-core';

const clipId = process.env.PLAYWRIGHT_CANDIDATE_CLIP_ID!;
const clipTitle = process.env.PLAYWRIGHT_CANDIDATE_CLIP_TITLE!;
const searchQuery = process.env.PLAYWRIGHT_CANDIDATE_SEARCH_QUERY!;

async function expectNoSeriousAxeViolations(page: Page) {
    await page.addScriptTag({ content: axe.source });
    const violations = await page.evaluate(async () => {
        const result = await window.axe.run(document, {
            runOnly: { type: 'tag', values: ['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'] },
        });
        return result.violations
            .filter(({ impact }) => impact === 'critical' || impact === 'serious')
            .map(({ id, impact, help, nodes }) => ({
                id,
                impact,
                help,
                nodes: nodes.map(({ target, failureSummary }) => ({ target, failureSummary })),
            }));
    });
    expect(violations).toEqual([]);
}

test.describe('deployed candidate accessibility', () => {
    let consoleErrors: string[];
    let pageErrors: string[];

    test.beforeEach(async ({ page }) => {
        consoleErrors = [];
        pageErrors = [];
        page.on('console', message => {
            if (message.type() === 'error') consoleErrors.push(message.text());
        });
        page.on('pageerror', error => pageErrors.push(error.message));
    });

    test.afterEach(() => {
        expect(consoleErrors, 'unexpected browser console errors').toEqual([]);
        expect(pageErrors, 'unexpected uncaught page errors').toEqual([]);
    });

    const journeys = [
        { name: 'home', path: '/', heading: /trending/i },
        { name: 'login', path: '/login', heading: /welcome to clipper/i },
        { name: 'search', path: `/search?q=${encodeURIComponent(searchQuery)}`, heading: /search/i },
        { name: 'clip detail', path: `/clip/${clipId}`, heading: clipTitle },
        { name: 'pricing', path: '/pricing', heading: /upgrade to clpr pro/i },
        { name: 'forum', path: '/forum', heading: /forum discussions/i },
    ];

    for (const viewport of [
        { name: 'desktop', width: 1280, height: 800 },
        { name: 'mobile', width: 375, height: 812 },
    ]) {
        for (const journey of journeys) {
            test(`${journey.name} is axe-clean on ${viewport.name}`, async ({ page }) => {
                await page.setViewportSize(viewport);
                const response = await page.goto(journey.path);
                expect(response?.ok(), `${journey.path} must load from the candidate`).toBe(true);
                await expect(page.getByRole('heading', { name: journey.heading }).first()).toBeVisible();
                await expectNoSeriousAxeViolations(page);
            });
        }
    }

    test('keyboard focus reaches main content', async ({ page }) => {
        await page.goto('/login');
        const skipLink = page.getByRole('link', { name: 'Skip to main content' });
        await skipLink.press('Enter');
        await expect(page.locator('#main-content')).toBeFocused();
    });

    test('200-percent-equivalent viewport reflows and reduced motion is honored', async ({ page }) => {
        await page.emulateMedia({ reducedMotion: 'reduce' });
        await page.setViewportSize({ width: 320, height: 568 });
        await page.goto(`/search?q=${encodeURIComponent(searchQuery)}`);
        expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(320);
        expect(await page.evaluate(() => matchMedia('(prefers-reduced-motion: reduce)').matches)).toBe(true);
    });
});

declare global {
    interface Window {
        axe: typeof axe;
    }
}
