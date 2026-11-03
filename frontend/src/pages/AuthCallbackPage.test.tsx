import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { beforeEach, expect, it, vi } from 'vitest';
import { AuthCallbackPage } from './AuthCallbackPage';
import { handleOAuthCallback } from '@/lib/auth-api';
import { completeSignIn } from '@/lib/complete-sign-in';

const auth = vi.hoisted(() => ({ refreshUser: vi.fn() }));
vi.mock('@/context/AuthContext', () => ({ useAuth: () => auth }));
vi.mock('@/lib/auth-api', () => ({ handleOAuthCallback: vi.fn() }));
vi.mock('@/lib/complete-sign-in', () => ({ completeSignIn: vi.fn() }));
vi.mock('@/lib/telemetry', () => ({ trackEvent: vi.fn(), AuthEvents: { LOGIN_FAILED: 'login_failed', LOGIN_COMPLETED: 'login_completed' } }));

function renderAt(url: string) {
    render(
        <MemoryRouter initialEntries={[url]}>
            <Routes>
                <Route path='/auth/success' element={<AuthCallbackPage />} />
                <Route path='/auth/mfa' element={<p>MFA step</p>} />
            </Routes>
        </MemoryRouter>,
    );
}

beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(completeSignIn).mockResolvedValue(undefined);
});

it('continues to the MFA step when the PKCE exchange requires a code', async () => {
    vi.mocked(handleOAuthCallback).mockResolvedValue({ success: true, mfaRequired: true });
    renderAt('/auth/success?code=abc&state=xyz');
    expect(await screen.findByText('MFA step')).toBeInTheDocument();
    expect(completeSignIn).not.toHaveBeenCalled();
});

it('continues to the MFA step after a server-side redirect that requires a code', async () => {
    renderAt('/auth/success?mfa_required=true');
    expect(await screen.findByText('MFA step')).toBeInTheDocument();
    expect(handleOAuthCallback).not.toHaveBeenCalled();
    expect(completeSignIn).not.toHaveBeenCalled();
});

it('completes the sign-in when no code is needed', async () => {
    vi.mocked(handleOAuthCallback).mockResolvedValue({ success: true, mfaRequired: false });
    renderAt('/auth/success?code=abc&state=xyz');
    await waitFor(() => expect(completeSignIn).toHaveBeenCalledWith(expect.any(Function), auth.refreshUser));
});
