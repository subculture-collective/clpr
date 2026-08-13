import { describe, expect, it } from 'vitest';
import { effectiveConsentValue } from './consent-display';

describe('effectiveConsentValue', () => {
    it('shows optional consent as off while Do Not Track overrides storage', () => {
        expect(effectiveConsentValue(true, true)).toBe(false);
        expect(effectiveConsentValue(true, false)).toBe(true);
    });
});
