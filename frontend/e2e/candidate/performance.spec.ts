import { expect, test } from '@playwright/test';

test.describe('blocking candidate performance budgets', () => {
    for (const viewport of [
        { name: 'desktop', width: 1280, height: 800 },
        { name: 'mobile', width: 390, height: 844 },
    ]) {
        test(`${viewport.name} homepage stays within launch budgets`, async ({
            page,
        }) => {

            await page.addInitScript(() => {
                (
                    window as Window & {
                        __clprCLS?: number;
                        __clprTBT?: number;
                    }
                ).__clprCLS = 0;
                (
                    window as Window & {
                        __clprCLS?: number;
                        __clprTBT?: number;
                    }
                ).__clprTBT = 0;
                new PerformanceObserver((list) => {
                    for (const entry of list.getEntries()) {
                        const shift = entry as PerformanceEntry & {
                            value: number;
                            hadRecentInput: boolean;
                        };
                        if (!shift.hadRecentInput)
                            window.__clprCLS =
                                (window.__clprCLS ?? 0) + shift.value;
                    }
                }).observe({ type: 'layout-shift', buffered: true });
                new PerformanceObserver((list) => {
                    for (const entry of list.getEntries()) {
                        window.__clprTBT =
                            (window.__clprTBT ?? 0) +
                            Math.max(0, entry.duration - 50);
                    }
                }).observe({ type: 'longtask', buffered: true });
            });
            await page.setViewportSize(viewport);
            const response = await page.goto('/', { waitUntil: 'networkidle' });
            expect(response?.ok()).toBe(true);
            await page.waitForTimeout(1_000);

            const metrics = await page.evaluate(() => ({
                cls: window.__clprCLS ?? 0,
                tbt: window.__clprTBT ?? 0,
                domElements: document.getElementsByTagName('*').length,
                resources: performance.getEntriesByType('resource').length + 1,
                twitchIframes: document.querySelectorAll(
                    'iframe[src*="twitch.tv"]',
                ).length,
            }));

            expect(metrics.cls, 'CLS').toBeLessThanOrEqual(0.1);
            expect(
                metrics.twitchIframes,
                'pre-interaction Twitch iframes',
            ).toBe(0);
            expect(
                metrics.domElements,
                'initial DOM elements',
            ).toBeLessThanOrEqual(1_500);
            expect(
                metrics.resources,
                'initial resource requests',
            ).toBeLessThanOrEqual(45);
            expect(
                metrics.tbt,
                'observed total blocking time',
            ).toBeLessThanOrEqual(200);
        });
    }
});

declare global {
    interface Window {
        axe: typeof axe;
        __clprCLS?: number;
        __clprTBT?: number;
    }
}
