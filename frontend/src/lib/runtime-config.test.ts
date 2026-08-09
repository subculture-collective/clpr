import { afterEach, describe, expect, it } from 'vitest';
import { getAnalyticsRuntimeConfig } from './runtime-config';

describe('analytics runtime configuration', () => {
    afterEach(() => {
        delete window.__CLPR_ANALYTICS_CONFIG__;
    });

    it('supports deterministic runtime injection without mutating import.meta', () => {
        window.__CLPR_ANALYTICS_CONFIG__ = {
            enabled: true,
            autoConsent: false,
            googleMeasurementId: 'G-TEST',
            postHogApiKey: 'ph_test',
            postHogHost: 'https://analytics.example.test',
        };

        expect(getAnalyticsRuntimeConfig()).toEqual(
            expect.objectContaining({
                enabled: true,
                autoConsent: false,
                googleMeasurementId: 'G-TEST',
                postHogApiKey: 'ph_test',
                postHogHost: 'https://analytics.example.test',
            }),
        );
    });
});
