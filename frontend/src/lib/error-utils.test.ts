import { describe, expect, it } from 'vitest';
import { isRetryableError, shouldRetryQuery } from './error-utils';

const httpError = (status: number) => ({ response: { status } });

describe('query retry policy', () => {
    it.each([400, 401, 403, 404, 409, 429])('does not retry a %i response', status => {
        expect(isRetryableError(httpError(status))).toBe(false);
        expect(shouldRetryQuery(0, httpError(status))).toBe(false);
    });

    it.each([408, 500, 502, 503])('retries a %i response once', status => {
        expect(shouldRetryQuery(0, httpError(status))).toBe(true);
        expect(shouldRetryQuery(1, httpError(status))).toBe(false);
    });

    it('retries a network failure without a response once', () => {
        expect(shouldRetryQuery(0, new Error('Network Error'))).toBe(true);
        expect(shouldRetryQuery(1, new Error('Network Error'))).toBe(false);
    });
});
