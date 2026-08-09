import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { User } from '../lib/auth-api';

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
    const { isLoading, user: currentUser, refreshUser } = useAuth();
    return (
        <div>
            <span>{isLoading ? 'loading' : currentUser?.username ?? 'anonymous'}</span>
            <button onClick={() => void refreshUser()}>Refresh user</button>
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
});
