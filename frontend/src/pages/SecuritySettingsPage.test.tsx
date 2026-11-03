import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, expect, it, vi } from 'vitest';
import { SecuritySettingsPage } from './SecuritySettingsPage';
import * as mfa from '@/lib/mfa-api';

vi.mock('@/lib/mfa-api', async importOriginal => {
    const actual = await importOriginal<typeof import('@/lib/mfa-api')>();
    return {
        ...actual,
        getMFAStatus: vi.fn(),
        startMFAEnrollment: vi.fn(),
        verifyMFAEnrollment: vi.fn(),
        regenerateBackupCodes: vi.fn(),
        disableMFA: vi.fn(),
        getTrustedDevices: vi.fn(),
        revokeTrustedDevice: vi.fn(),
    };
});

const disabledStatus: mfa.MFAStatus = { enabled: false, backup_codes_remaining: 0, trusted_devices_count: 0, required: false, in_grace_period: false };
const enabledStatus: mfa.MFAStatus = { ...disabledStatus, enabled: true, enrolled_at: '2026-10-01T00:00:00Z', backup_codes_remaining: 9, trusted_devices_count: 1 };

function setup() {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(<QueryClientProvider client={client}><MemoryRouter><SecuritySettingsPage /></MemoryRouter></QueryClientProvider>);
    return userEvent.setup();
}

beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(mfa.getTrustedDevices).mockResolvedValue([]);
});

it('walks through enrollment: QR code, backup codes, then the verification code', async () => {
    vi.mocked(mfa.getMFAStatus).mockResolvedValueOnce(disabledStatus).mockResolvedValue(enabledStatus);
    vi.mocked(mfa.startMFAEnrollment).mockResolvedValue({
        secret: 'JBSWY3DPEHPK3PXP',
        qr_code_url: 'data:image/png;base64,AAAA',
        backup_codes: ['AAAA1111', 'BBBB2222'],
    });
    vi.mocked(mfa.verifyMFAEnrollment).mockResolvedValue(undefined);
    const user = setup();

    await user.click(await screen.findByRole('button', { name: 'Set up two-factor authentication' }));
    expect(await screen.findByAltText('QR code for your authenticator app')).toHaveAttribute('src', 'data:image/png;base64,AAAA');
    expect(screen.getByText('JBSWY3DPEHPK3PXP')).toBeInTheDocument();
    expect(screen.getByText('AAAA1111')).toBeInTheDocument();

    const turnOn = screen.getByRole('button', { name: 'Turn on' });
    expect(turnOn).toBeDisabled();
    await user.type(screen.getByLabelText('Authentication code'), '123456');
    await user.click(turnOn);

    await waitFor(() => expect(mfa.verifyMFAEnrollment).toHaveBeenCalledWith('123456'));
    expect(await screen.findByText(/signing in on a new browser asks for a code/)).toBeInTheDocument();
    expect(screen.getByText(/Two-factor authentication is on/)).toBeInTheDocument();
});

it('shows the server error when the enrollment code is wrong', async () => {
    vi.mocked(mfa.getMFAStatus).mockResolvedValue(disabledStatus);
    vi.mocked(mfa.startMFAEnrollment).mockResolvedValue({ secret: 'S', qr_code_url: 'data:image/png;base64,AAAA', backup_codes: [] });
    vi.mocked(mfa.verifyMFAEnrollment).mockRejectedValue(new Error('bad code'));
    const user = setup();

    await user.click(await screen.findByRole('button', { name: 'Set up two-factor authentication' }));
    await user.type(await screen.findByLabelText('Authentication code'), '000000');
    await user.click(screen.getByRole('button', { name: 'Turn on' }));
    expect(await screen.findByText('That code is incorrect. Try again.')).toBeInTheDocument();
});

it('warns users whose role requires MFA', async () => {
    vi.mocked(mfa.getMFAStatus).mockResolvedValue({ ...disabledStatus, required: true, in_grace_period: true, grace_period_end: '2026-11-09T00:00:00Z' });
    setup();
    expect(await screen.findByText(/Your role requires two-factor authentication/)).toBeInTheDocument();
});

it('manages backup codes, trusted browsers and disabling when MFA is on', async () => {
    vi.mocked(mfa.getMFAStatus).mockResolvedValue(enabledStatus);
    vi.mocked(mfa.getTrustedDevices).mockResolvedValue([
        { id: 7, device_name: 'Firefox on Linux', trusted_at: '2026-10-20T00:00:00Z', expires_at: '2026-11-19T00:00:00Z', last_used_at: '2026-10-21T00:00:00Z' },
    ]);
    vi.mocked(mfa.regenerateBackupCodes).mockResolvedValue(['NEWC0DE1']);
    vi.mocked(mfa.revokeTrustedDevice).mockResolvedValue(undefined);
    vi.mocked(mfa.disableMFA).mockResolvedValue(undefined);
    const user = setup();

    expect(await screen.findByText('9 unused backup codes left. Generating new codes replaces all old ones.')).toBeInTheDocument();
    const [regenerateInput, disableInput] = screen.getAllByLabelText('Authentication code');
    await user.type(regenerateInput, '111111');
    await user.click(screen.getByRole('button', { name: 'Generate new codes' }));
    expect(await screen.findByText('NEWC0DE1')).toBeInTheDocument();
    expect(mfa.regenerateBackupCodes).toHaveBeenCalledWith('111111');

    expect(await screen.findByText('Firefox on Linux')).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Revoke' }));
    await waitFor(() => expect(mfa.revokeTrustedDevice).toHaveBeenCalledWith(7));

    await user.type(disableInput, '222222');
    await user.click(screen.getByRole('button', { name: 'Turn off' }));
    await waitFor(() => expect(mfa.disableMFA).toHaveBeenCalledWith('222222'));
});
