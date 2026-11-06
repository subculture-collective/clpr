import { describe, expect, it } from 'vitest';
import { getErrorMessage, isRetryableError, shouldRetryQuery } from './error-utils';

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

describe('getErrorMessage', () => {
    const fallback = 'Fallback';

    it('reads the nested message from a standard error envelope', () => {
        const error = { response: { status: 403, data: { success: false, error: { code: 'FORBIDDEN', message: 'Voting is disabled' } } } };
        expect(getErrorMessage(error, fallback)).toBe('Voting is disabled');
    });

    it('reads a string error body', () => {
        expect(getErrorMessage({ response: { data: { error: 'Nope' } } }, fallback)).toBe('Nope');
    });

    it('reads a message body', () => {
        expect(getErrorMessage({ response: { data: { message: 'Slow down' } } }, fallback)).toBe('Slow down');
    });

    it('uses the fallback instead of a transport message or a non-string body', () => {
        expect(getErrorMessage({ message: 'Request failed with status code 500', response: { data: { error: { code: 'X' } } } }, fallback)).toBe(fallback);
    });

    it('handles strings, Error objects and empty values', () => {
        expect(getErrorMessage('Plain', fallback)).toBe('Plain');
        expect(getErrorMessage(new Error('Network Error'), fallback)).toBe('Network Error');
        expect(getErrorMessage(null, fallback)).toBe(fallback);
        expect(getErrorMessage({ message: { nested: true } }, fallback)).toBe(fallback);
    });
});
