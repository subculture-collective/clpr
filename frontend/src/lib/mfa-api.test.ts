import { AxiosError, AxiosHeaders } from 'axios';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({ get: vi.fn(), post: vi.fn(), delete: vi.fn() }));
vi.mock('./api', () => ({ apiClient: mocks }));
import { MFALoginError, getTrustedDevices, verifyMFALogin } from './mfa-api';

function axiosError(status: number, data: unknown) {
    const config = { headers: new AxiosHeaders() };
    return new AxiosError('Request failed', 'ERR_BAD_REQUEST', config, undefined, {
        status, statusText: '', headers: {}, config, data,
    });
}

beforeEach(() => vi.resetAllMocks());

describe('verifyMFALogin', () => {
    it('sends only the code and trust choice, without triggering a session refresh', async () => {
        mocks.post.mockResolvedValue({ data: { user: { id: 'user-1' } } });
        await expect(verifyMFALogin(' 123456 ', true)).resolves.toEqual({ id: 'user-1' });
        expect(mocks.post).toHaveBeenCalledWith(
            '/auth/mfa/verify-login',
            { code: '123456', trust_device: true },
            { skipAuthRefresh: true },
        );
    });

    it.each([
        ['MFA_CODE_INVALID', false],
        ['MFA_LOCKED', false],
        ['MFA_CHALLENGE_EXPIRED', true],
        ['MFA_CHALLENGE_EXHAUSTED', true],
        ['MFA_CHALLENGE_MISSING', true],
    ] as const)('maps %s from the server', async (code, requiresNewSignIn) => {
        mocks.post.mockRejectedValue(axiosError(401, { error: 'server message', code }));
        const error = await verifyMFALogin('123456', false).catch(e => e);
        expect(error).toBeInstanceOf(MFALoginError);
        expect(error.code).toBe(code);
        expect(error.message).toBe('server message');
        expect(error.requiresNewSignIn).toBe(requiresNewSignIn);
    });

    it('reports network failures without a code', async () => {
        mocks.post.mockRejectedValue(new Error('Network Error'));
        const error = await verifyMFALogin('123456', false).catch(e => e);
        expect(error).toBeInstanceOf(MFALoginError);
        expect(error.code).toBeUndefined();
        expect(error.requiresNewSignIn).toBe(false);
    });
});

it('treats a null trusted-device list as empty', async () => {
    mocks.get.mockResolvedValue({ data: { devices: null } });
    await expect(getTrustedDevices()).resolves.toEqual([]);
});
