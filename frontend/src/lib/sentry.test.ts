import { describe, expect, it } from 'vitest';
import type { ErrorEvent } from '@sentry/react';
import { isThirdPartyTwitchFailure } from './sentry';

describe('isThirdPartyTwitchFailure', () => {
    it('classifies Twitch player failures', () => {
        const event = {
            request: { url: 'https://clips.twitch.tv/embed?clip=example' },
        } as ErrorEvent;

        expect(isThirdPartyTwitchFailure(event)).toBe(true);
    });

    it('keeps first-party API failures in the first-party error domain', () => {
        const event = {
            request: { url: 'https://clpr.tv/api/v1/clips' },
            exception: { values: [{ value: 'request failed' }] },
        } as ErrorEvent;

        expect(isThirdPartyTwitchFailure(event)).toBe(false);
    });
});
