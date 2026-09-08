import { act, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
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

describe('AuthProvider session restoration', () => {
    beforeEach(() => {
        localStorage.clear();
        vi.clearAllMocks();
    });

    it('uses a quiet probe when no authenticated session is known', async () => {
        authApi.getCurrentUser.mockRejectedValueOnce(new Error('401'));

        render(<AuthProvider><AuthConsumer /></AuthProvider>);

        expect(await screen.findByText('anonymous')).toBeVisible();
        expect(authApi.getCurrentUser).toHaveBeenCalledWith({ anonymousProbe: true });
    });

    it('allows refresh while restoring a known authenticated session', async () => {
        localStorage.setItem('auth_session_hint', '1');
        authApi.getCurrentUser.mockResolvedValueOnce(user('restored'));

        render(<AuthProvider><AuthConsumer /></AuthProvider>);

        expect(await screen.findByText('restored')).toBeVisible();
        expect(authApi.getCurrentUser).toHaveBeenCalledWith({ anonymousProbe: false });
    });

    it('keeps explicit refreshUser refresh-enabled', async () => {
        const browserUser = userEvent.setup();
        localStorage.setItem('auth_session_hint', '1');
        authApi.getCurrentUser
            .mockResolvedValueOnce(user('initial'))
            .mockResolvedValueOnce(user('updated'));
        render(<AuthProvider><AuthConsumer /></AuthProvider>);
        await screen.findByText('initial');

        await browserUser.click(screen.getByRole('button', { name: 'Refresh user' }));

        await waitFor(() => expect(screen.getByText('updated')).toBeVisible());
        expect(authApi.getCurrentUser).toHaveBeenNthCalledWith(2);
    });

    it('clears a revoked session once when concurrent requests become unauthorized', async () => {
        authApi.getCurrentUser.mockResolvedValueOnce({ ...user('revoked'), is_verified: true });
        render(<AuthProvider><AuthConsumer /></AuthProvider>);
        await screen.findByText('revoked');
        const unauthorized = vi.mocked(setUnauthorizedHandler).mock.calls.at(-1)?.[0];
        expect(unauthorized).toBeTypeOf('function');
        await act(async () => { unauthorized?.(); unauthorized?.(); });
        expect(await screen.findByText('anonymous')).toBeVisible();
        expect(localStorage.getItem('auth_session_hint')).toBeNull();
        expect(resetUser).toHaveBeenCalledTimes(1);
    });

    it('removes local access even when the logout endpoint fails, preserving preferences', async () => {
        allowTestConsole('error', /Logout error: Error: network unavailable/);
        authApi.getCurrentUser.mockResolvedValueOnce(user('leaving'));
        authApi.logout.mockRejectedValueOnce(new Error('network unavailable'));
        localStorage.setItem('theme', 'dark');
        const browserUser = userEvent.setup();
        render(<AuthProvider><AuthConsumer /></AuthProvider>);
        await screen.findByText('leaving');
        await browserUser.click(screen.getByRole('button', { name: 'Logout' }));
        expect(await screen.findByText('anonymous')).toBeVisible();
        expect(localStorage.getItem('auth_session_hint')).toBeNull();
        expect(localStorage.getItem('theme')).toBe('dark');
    });
});
