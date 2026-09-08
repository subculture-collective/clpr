import { beforeEach, describe, expect, it, vi } from 'vitest';

const mocks = vi.hoisted(() => ({
    get: vi.fn(), post: vi.fn(),
    getSecureItem: vi.fn(), setSecureItem: vi.fn(), removeSecureItem: vi.fn(), clearSecureStorage: vi.fn(),
    generatePKCEParams: vi.fn(),
}));
vi.mock('./api', () => ({ apiClient: { get: mocks.get, post: mocks.post } }));
vi.mock('./secure-storage', () => mocks);
vi.mock('./pkce', () => ({ generatePKCEParams: mocks.generatePKCEParams }));
import { getCurrentUser, handleOAuthCallback, logout, initiateOAuth, testLogin } from './auth-api';

beforeEach(() => {
    vi.resetAllMocks();
    mocks.getSecureItem.mockImplementation(async key => key === 'oauth_state' ? 'expected-state' : 'secret-verifier');
});

describe('authentication boundary', () => {
    it.each([true, false])('refresh eligibility follows anonymousProbe=%s', async anonymousProbe => {
        const user = { id: 'user-1' };
        mocks.get.mockResolvedValue({ data: user });
        await expect(getCurrentUser({ anonymousProbe })).resolves.toEqual(user);
        expect(mocks.get).toHaveBeenCalledWith('/auth/me', anonymousProbe ? { skipAuthRefresh: true } : undefined);
    });

    it('uses authenticated restoration by default', async () => {
        mocks.get.mockResolvedValue({ data: { id: 'user-1' } });
        await getCurrentUser();
        expect(mocks.get).toHaveBeenCalledWith('/auth/me', undefined);
    });

    it.each([
        { state: null, verifier: 'secret-verifier', error: 'Invalid state parameter' },
        { state: 'different-state', verifier: 'secret-verifier', error: 'Invalid state parameter' },
        { state: 'expected-state', verifier: null, error: 'Code verifier not found' },
    ])('rejects $error before exchanging credentials', async ({ state, verifier, error }) => {
        mocks.getSecureItem.mockImplementation(async key => key === 'oauth_state' ? state : verifier);
        await expect(handleOAuthCallback('code', 'expected-state')).resolves.toEqual({ success: false, error });
        expect(mocks.post).not.toHaveBeenCalled();
    });

    it('exchanges only matching state and consumes its verifier after success', async () => {
        await expect(handleOAuthCallback('code', 'expected-state')).resolves.toEqual({ success: true });
        expect(mocks.post).toHaveBeenCalledWith('/auth/twitch/callback', {
            code: 'code', state: 'expected-state', code_verifier: 'secret-verifier',
        });
        expect(mocks.removeSecureItem.mock.calls).toEqual([['oauth_state'], ['oauth_code_verifier']]);
    });

    it.each([new Error('Rejected'), 'Rejected'])('cleans up credentials after exchange failure', async error => {
        mocks.post.mockRejectedValue(error);
        await expect(handleOAuthCallback('code', 'expected-state')).resolves.toEqual({
            success: false, error: error instanceof Error ? error.message : 'Authentication failed',
        });
        expect(mocks.removeSecureItem.mock.calls).toEqual([['oauth_state'], ['oauth_code_verifier']]);
    });

    it.each([false, true])('clears local credentials even when revocation fails=%s', async fails => {
        if (fails) mocks.post.mockRejectedValue(new Error('Offline'));
        const result = logout();
        if (fails) await expect(result).rejects.toThrow('Offline');
        else await expect(result).resolves.toBeUndefined();
        expect(mocks.post).toHaveBeenCalledWith('/auth/logout');
        expect(mocks.clearSecureStorage).toHaveBeenCalledOnce();
    });

    it('persists PKCE credentials before handing off to the OAuth endpoint', async () => {
        mocks.generatePKCEParams.mockResolvedValue({ codeVerifier: 'verifier', codeChallenge: 'challenge', state: 'state' });
        const redirect = vi.fn();
        await initiateOAuth(redirect);
        expect(mocks.setSecureItem.mock.calls).toEqual([['oauth_code_verifier', 'verifier'], ['oauth_state', 'state']]);
        const target = new URL(redirect.mock.calls[0][0]);
        expect(target.pathname).toMatch(/\/auth\/twitch$/);
        expect(Object.fromEntries(target.searchParams)).toEqual({ code_challenge: 'challenge', code_challenge_method: 'S256', state: 'state' });
        expect(mocks.setSecureItem.mock.invocationCallOrder.at(-1)).toBeLessThan(redirect.mock.invocationCallOrder[0]);
    });

    it('does not redirect when PKCE persistence fails', async () => {
        mocks.generatePKCEParams.mockResolvedValue({ codeVerifier: 'verifier', codeChallenge: 'challenge', state: 'state' });
        mocks.setSecureItem.mockRejectedValue(new Error('Storage unavailable'));
        const redirect = vi.fn();
        await expect(initiateOAuth(redirect)).rejects.toThrow('Storage unavailable');
        expect(redirect).not.toHaveBeenCalled();
    });

    it('unwraps the local test-login response', async () => {
        mocks.post.mockResolvedValue({ data: { user: { id: 'fixture-user' } } });
        await expect(testLogin({ username: 'fixture' })).resolves.toEqual({ id: 'fixture-user' });
    });
});
