import type { AxiosError, AxiosResponse } from 'axios';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => {
    const client = Object.assign(vi.fn(), {
        post: vi.fn(),
        interceptors: {
            request: { use: vi.fn() },
            response: { use: vi.fn() },
        },
    });
    return { client };
});

vi.mock('axios', () => ({
    default: { create: vi.fn(() => mocks.client) },
    AxiosError: class MockAxiosError extends Error {},
}));

await import('./api');

type RejectionInterceptor = (error: AxiosError) => Promise<AxiosResponse>;

const getRejectionInterceptor = (): RejectionInterceptor => {
    const registration = mocks.client.interceptors.response.use.mock.calls[0];
    return registration[1] as RejectionInterceptor;
};

describe('API authentication refresh policy', () => {
    beforeEach(() => {
        mocks.client.mockReset();
        mocks.client.post.mockReset();
    });

    it('does not refresh an intentional anonymous probe', async () => {
        const error = {
            config: { url: '/auth/me', skipAuthRefresh: true },
            response: { status: 401 },
        } as AxiosError;

        await expect(getRejectionInterceptor()(error)).rejects.toBe(error);
        expect(mocks.client.post).not.toHaveBeenCalled();
    });

    it('restores an expired authenticated session after a 401', async () => {
        const refreshResponse = { status: 200 } as AxiosResponse;
        const retriedResponse = { status: 200, data: { ok: true } } as AxiosResponse;
        mocks.client.post.mockResolvedValue(refreshResponse);
        mocks.client.mockResolvedValue(retriedResponse);
        const config = { url: '/auth/me' };
        const error = {
            config,
            response: { status: 401 },
        } as AxiosError;

        await expect(getRejectionInterceptor()(error)).resolves.toBe(retriedResponse);
        expect(mocks.client.post).toHaveBeenCalledWith('/auth/refresh');
        expect(mocks.client).toHaveBeenCalledWith(
            expect.objectContaining({ url: '/auth/me', _retry: true }),
        );
    });

    it('fails closed when the refresh token is invalid', async () => {
        const refreshError = {
            config: { url: '/auth/refresh' },
            response: { status: 401 },
        } as AxiosError;
        mocks.client.post.mockRejectedValue(refreshError);
        const error = {
            config: { url: '/auth/me' },
            response: { status: 401 },
        } as AxiosError;

        await expect(getRejectionInterceptor()(error)).rejects.toBe(refreshError);
        expect(mocks.client.post).toHaveBeenCalledTimes(1);
        expect(mocks.client).not.toHaveBeenCalled();
    });
});
