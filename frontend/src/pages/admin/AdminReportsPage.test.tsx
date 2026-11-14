import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useAuth } from '../../context/AuthContext';
import * as reportApi from '../../lib/report-api';
import { render, screen, within } from '../../test/test-utils';
import type { Report } from '../../types/report';
import { AdminReportsPage } from './AdminReportsPage';

vi.mock('../../lib/report-api', () => ({
    listReports: vi.fn(),
    updateReport: vi.fn(),
}));

vi.mock('../../context/AuthContext', async () => {
    const actual = await vi.importActual('../../context/AuthContext');
    return { ...actual, useAuth: vi.fn() };
});

const report = (id: string, type: Report['reportable_type']): Report => ({
    id,
    reporter_id: 'reporter-1',
    reportable_type: type,
    reportable_id: `${type}-1`,
    reason: 'harassment',
    description: `${type} report description`,
    status: 'pending',
    created_at: '2024-01-01T00:00:00Z',
});

describe('AdminReportsPage', () => {
    const mockListReports = vi.mocked(reportApi.listReports);

    beforeEach(() => {
        vi.clearAllMocks();
        vi.mocked(useAuth).mockReturnValue({
            isAuthenticated: true,
            isAdmin: true,
        } as unknown as ReturnType<typeof useAuth>);
    });

    it('renders a structured API error as text', async () => {
        mockListReports.mockRejectedValue({
            response: {
                status: 403,
                data: { error: { code: 'MFA_REQUIRED', message: 'Set up two-factor authentication to continue.' } },
            },
        });

        render(<AdminReportsPage />);

        expect(
            await screen.findByText('Set up two-factor authentication to continue.')
        ).toBeInTheDocument();
    });

    it('offers only the actions the API accepts for each report type', async () => {
        mockListReports.mockResolvedValue({
            data: [report('r1', 'clip'), report('r2', 'comment'), report('r3', 'user')],
            meta: { page: 1, limit: 20, total: 3, total_pages: 1 },
        });

        render(<AdminReportsPage />);

        const actionsFor = async (type: string) => {
            const description = await screen.findByText(`${type} report description`);
            const row = within(description.parentElement as HTMLElement);
            return {
                remove: row.queryByRole('button', { name: 'Remove Content' }),
                ban: row.queryByRole('button', { name: 'Ban User' }),
                dismiss: row.queryByRole('button', { name: 'Dismiss' }),
            };
        };

        const clip = await actionsFor('clip');
        expect(clip.remove).toBeInTheDocument();
        expect(clip.ban).not.toBeInTheDocument();
        expect(clip.dismiss).toBeInTheDocument();

        const comment = await actionsFor('comment');
        expect(comment.remove).toBeInTheDocument();
        expect(comment.ban).toBeInTheDocument();

        const user = await actionsFor('user');
        expect(user.remove).not.toBeInTheDocument();
        expect(user.ban).toBeInTheDocument();
        expect(user.dismiss).toBeInTheDocument();
    });
});
