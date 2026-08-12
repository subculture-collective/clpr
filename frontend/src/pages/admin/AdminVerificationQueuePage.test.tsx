import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { AdminVerificationQueuePage } from './AdminVerificationQueuePage';
import * as verificationApi from '../../lib/verification-api';

vi.mock('../../context/AuthContext', () => ({
    useAuth: () => ({ isAuthenticated: true, isAdmin: true }),
}));

vi.mock('../../lib/verification-api', async importOriginal => {
    const original = await importOriginal<typeof import('../../lib/verification-api')>();
    return {
        ...original,
        getVerificationApplications: vi.fn(),
        getVerificationStats: vi.fn(),
        reviewVerificationApplication: vi.fn(),
    };
});

const application: verificationApi.VerificationApplication = {
    id: '11111111-1111-4111-8111-111111111111',
    user_id: '22222222-2222-4222-8222-222222222222',
    twitch_channel_url: 'https://twitch.tv/example',
    status: 'pending',
    priority: 80,
    created_at: '2026-08-11T00:00:00Z',
    updated_at: '2026-08-11T00:00:00Z',
};

describe('AdminVerificationQueuePage', () => {
    beforeEach(() => {
        vi.mocked(verificationApi.getVerificationApplications).mockResolvedValue({
            success: true,
            data: [application],
            meta: { count: 1, limit: 50, page: 1, status: 'pending' },
        });
        vi.mocked(verificationApi.getVerificationStats).mockResolvedValue({
            success: true,
            data: { total_pending: 1, total_approved: 0, total_rejected: 0, total_verified: 0 },
        });
        vi.mocked(verificationApi.reviewVerificationApplication).mockResolvedValue({
            success: true,
            message: 'approved',
            data: { id: application.id, decision: 'approved' },
        });
    });

    it('opens the approval dialog and sends the decision', async () => {
        const user = userEvent.setup();
        render(<MemoryRouter><AdminVerificationQueuePage /></MemoryRouter>);

        await user.click(await screen.findByRole('button', { name: 'Approve' }));
        const dialog = screen.getByRole('dialog', { name: 'Approve Application' });
        expect(dialog).toBeInTheDocument();

        await user.click(within(dialog).getByRole('button', { name: 'Approve' }));
        await waitFor(() => expect(verificationApi.reviewVerificationApplication).toHaveBeenCalledWith(
            application.id,
            { decision: 'approved', notes: undefined },
        ));
    });
});
