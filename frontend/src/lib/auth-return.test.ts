import { describe, expect, it } from 'vitest';
import { getAuthReturnTo } from './auth-return';

describe('sign-in return destination', () => {
    it('preserves a full internal location, including filters and discussion anchor', () => {
        expect(getAuthReturnTo({ pathname: '/clips/example', search: '?from=saved', hash: '#comments' }))
            .toBe('/clips/example?from=saved#comments');
        expect(getAuthReturnTo('/favorites?sort=newest')).toBe('/favorites?sort=newest');
    });

    it('rejects external destinations, invalid values and authentication loops', () => {
        for (const value of [undefined, {}, 'https://example.com', '//example.com', '/\\example.com', '/login?next=x', '/auth/callback', '/\nexample.com']) {
            expect(getAuthReturnTo(value)).toBe('/');
        }
    });
});
