import { act, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { User } from '../lib/auth-api';
import { allowTestConsole } from '../test/setup';

const authApi = vi.hoisted(() => ({
    getCurrentUser: vi.fn(),
    logout: vi.fn(),
    initiateOAuth: vi.fn(),
    testLogin: vi.fn(),
}));

vi.mock('../lib/auth-api', () => authApi);
vi.mock('../lib/sentry-client', () => ({
    setUser: vi.fn(),
    clearUser: vi.fn(),
}));
vi.mock('../lib/telemetry', () => ({
    resetUser: vi.fn(),
    identifyUser: vi.fn(),
    trackEvent: vi.fn(),
    AuthEvents: { LOGOUT: 'logout' },
}));
vi.mock('../lib/api', () => ({ setUnauthorizedHandler: vi.fn() }));

import { AuthProvider, useAuth } from './AuthContext';
import { setUnauthorizedHandler } from '../lib/api';
import { resetUser } from '../lib/telemetry';

const user = (username: string): User => ({
    id: `id-${username}`,
    twitch_id: `twitch-${username}`,
    username,
    display_name: username,
    role: 'user',
    karma_points: 0,
    is_banned: false,
    created_at: '2026-01-01T00:00:00Z',
});

function AuthConsumer() {
    const { isLoading, user: currentUser, refreshUser, logout } = useAuth();
    return (
        <div>
            <span>{isLoading ? 'loading' : currentUser?.username ?? 'anonymous'}</span>
            <button onClick={() => void refreshUser()}>Refresh user</button>
            <button onClick={() => void logout()}>Logout</button>
        </div>
    );
}

function renderAuth() {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(<QueryClientProvider client={client}><AuthProvider><AuthConsumer /></AuthProvider></QueryClientProvider>);
    return client;
}

describe('AuthProvider session restoration', () => {
    beforeEach(() => {
        localStorage.clear();
        vi.clearAllMocks();
    });

    it('skips the session probe when this browser has no recorded sign-in', async () => {
        const client = renderAuth();
        client.setQueryData(['publicFeed'], { clips: [] });

        expect(await screen.findByText('anonymous')).toBeVisible();
        expect(client.getQueryData(['publicFeed'])).toEqual({ clips: [] });
        expect(authApi.getCurrentUser).not.toHaveBeenCalled();
    });

    it('uses a quiet probe when storage is unavailable', async () => {
        const getItem = vi.spyOn(localStorage, 'getItem').mockImplementation(() => {
            throw new DOMException('blocked', 'SecurityError');
        });
        authApi.getCurrentUser.mockRejectedValueOnce(new Error('401'));
        try {
            renderAuth();
            expect(await screen.findByText('anonymous')).toBeVisible();
            expect(authApi.getCurrentUser).toHaveBeenCalledWith({ anonymousProbe: true });
        } finally {
            getItem.mockRestore();
        }
    });

    it('allows refresh while restoring a known authenticated session', async () => {
        localStorage.setItem('auth_session_hint', '1');
        authApi.getCurrentUser.mockResolvedValueOnce(user('restored'));

        renderAuth();

        expect(await screen.findByText('restored')).toBeVisible();
        expect(authApi.getCurrentUser).toHaveBeenCalledWith({ anonymousProbe: false });
    });

    it('keeps explicit refreshUser refresh-enabled', async () => {
        const browserUser = userEvent.setup();
        localStorage.setItem('auth_session_hint', '1');
        authApi.getCurrentUser
            .mockResolvedValueOnce(user('initial'))
            .mockResolvedValueOnce(user('updated'));
        renderAuth();
        await screen.findByText('initial');

        await browserUser.click(screen.getByRole('button', { name: 'Refresh user' }));

        await waitFor(() => expect(screen.getByText('updated')).toBeVisible());
        expect(authApi.getCurrentUser).toHaveBeenNthCalledWith(2);
    });

    it('clears a revoked session once when concurrent requests become unauthorized', async () => {
        localStorage.setItem('auth_session_hint', '1');
        authApi.getCurrentUser.mockResolvedValueOnce({ ...user('revoked'), is_verified: true });
        const client = renderAuth();
        await screen.findByText('revoked');
        client.setQueryData(['playlist', 'private'], { title: 'Private cached title' });
        const unauthorized = vi.mocked(setUnauthorizedHandler).mock.calls.at(-1)?.[0];
        expect(unauthorized).toBeTypeOf('function');
        await act(async () => { unauthorized?.(); unauthorized?.(); });
        expect(await screen.findByText('anonymous')).toBeVisible();
        expect(localStorage.getItem('auth_session_hint')).toBeNull();
        expect(resetUser).toHaveBeenCalledTimes(1);
        expect(client.getQueryCache().getAll()).toHaveLength(0);
    });

    it('removes local access even when the logout endpoint fails, preserving preferences', async () => {
        allowTestConsole('error', /Logout error: Error: network unavailable/);
        localStorage.setItem('auth_session_hint', '1');
        authApi.getCurrentUser.mockResolvedValueOnce(user('leaving'));
        authApi.logout.mockRejectedValueOnce(new Error('network unavailable'));
        localStorage.setItem('theme', 'dark');
        const browserUser = userEvent.setup();
        const client = renderAuth();
        await screen.findByText('leaving');
        client.setQueryData(['playlist', 'private'], { title: 'Private cached title' });
        let finish: (value: string) => void = () => {};
        let aborted = false;
        const pending = client.fetchQuery({ queryKey: ['private', 'inflight'], queryFn: ({ signal }) => {
            signal.addEventListener('abort', () => { aborted = true; });
            return new Promise<string>(resolve => { finish = resolve; });
        } }).catch(() => undefined);
        await browserUser.click(screen.getByRole('button', { name: 'Logout' }));
        expect(await screen.findByText('anonymous')).toBeVisible();
        expect(localStorage.getItem('auth_session_hint')).toBeNull();
        expect(localStorage.getItem('theme')).toBe('dark');
        expect(aborted).toBe(true);
        finish('Late private response');
        await pending;
        expect(client.getQueryCache().getAll()).toHaveLength(0);
    });
    it('retains cache for the same account but discards it when the principal changes', async () => {
        localStorage.setItem('auth_session_hint', '1');
        authApi.getCurrentUser.mockResolvedValueOnce(user('member'))
            .mockResolvedValueOnce({ ...user('member'), display_name: 'Updated profile' })
            .mockResolvedValueOnce(user('second'));
        const client = renderAuth();
        await screen.findByText('member');
        client.setQueryData(['userSettings'], { profile_visibility: 'private' });
        const browserUser = userEvent.setup();
        await browserUser.click(screen.getByRole('button', { name: 'Refresh user' }));
        expect(client.getQueryData(['userSettings'])).toEqual({ profile_visibility: 'private' });
        await browserUser.click(screen.getByRole('button', { name: 'Refresh user' }));
        await screen.findByText('second');
        expect(client.getQueryCache().getAll()).toHaveLength(0);
    });

});
