import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { beforeEach, expect, it, vi } from 'vitest';
import { MFAChallengePage } from './MFAChallengePage';
import { MFALoginError, verifyMFALogin } from '@/lib/mfa-api';
import { completeSignIn } from '@/lib/complete-sign-in';

const auth = vi.hoisted(() => ({ refreshUser: vi.fn() }));
vi.mock('@/context/AuthContext', () => ({ useAuth: () => auth }));
vi.mock('@/lib/complete-sign-in', () => ({ completeSignIn: vi.fn() }));
vi.mock('@/lib/mfa-api', async importOriginal => {
    const actual = await importOriginal<typeof import('@/lib/mfa-api')>();
    return { ...actual, verifyMFALogin: vi.fn() };
});

function setup() {
    render(
        <MemoryRouter initialEntries={['/auth/mfa']}>
            <Routes>
                <Route path='/auth/mfa' element={<MFAChallengePage />} />
                <Route path='/login' element={<p>Login page</p>} />
            </Routes>
        </MemoryRouter>,
    );
    return userEvent.setup();
}

beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(completeSignIn).mockResolvedValue(undefined);
});

it('submits the authenticator code and completes the sign-in', async () => {
    vi.mocked(verifyMFALogin).mockResolvedValue({ id: 'user-1' } as never);
    const user = setup();
    const submit = screen.getByRole('button', { name: 'Verify' });
    expect(submit).toBeDisabled();

    await user.type(screen.getByLabelText('Authentication code'), '123456');
    await user.click(screen.getByLabelText('Trust this browser for 30 days'));
    await user.click(submit);

    await waitFor(() => expect(completeSignIn).toHaveBeenCalledWith(expect.any(Function), auth.refreshUser));
    expect(verifyMFALogin).toHaveBeenCalledWith('123456', true);
});

it('accepts backup codes', async () => {
    vi.mocked(verifyMFALogin).mockResolvedValue({ id: 'user-1' } as never);
    const user = setup();
    await user.click(screen.getByRole('button', { name: 'Use a backup code instead' }));
    await user.type(screen.getByLabelText('Backup code'), 'abcd2345');
    await user.click(screen.getByRole('button', { name: 'Verify' }));
    await waitFor(() => expect(verifyMFALogin).toHaveBeenCalledWith('abcd2345', false));
});

it('lets the user retry after a wrong code', async () => {
    vi.mocked(verifyMFALogin).mockRejectedValueOnce(new MFALoginError('Invalid MFA code', 'MFA_CODE_INVALID'));
    const user = setup();
    await user.type(screen.getByLabelText('Authentication code'), '000000');
    await user.click(screen.getByRole('button', { name: 'Verify' }));

    expect(await screen.findByText('That code is incorrect. Try again.')).toBeInTheDocument();
    expect(screen.getByLabelText('Authentication code')).toHaveValue('');
    expect(completeSignIn).not.toHaveBeenCalled();
});

it.each(['MFA_CHALLENGE_EXPIRED', 'MFA_CHALLENGE_EXHAUSTED', 'MFA_CHALLENGE_MISSING'] as const)(
    'sends the user back to sign in when the pending sign-in ended (%s)',
    async code => {
        vi.mocked(verifyMFALogin).mockRejectedValueOnce(new MFALoginError('Your sign-in expired. Please sign in again.', code));
        const user = setup();
        await user.type(screen.getByLabelText('Authentication code'), '123456');
        await user.click(screen.getByRole('button', { name: 'Verify' }));

        expect(await screen.findByText('Your sign-in expired. Please sign in again.')).toBeInTheDocument();
        expect(screen.queryByLabelText('Authentication code')).not.toBeInTheDocument();
        await user.click(screen.getByRole('link', { name: 'Sign in again' }));
        expect(await screen.findByText('Login page')).toBeInTheDocument();
    },
);
