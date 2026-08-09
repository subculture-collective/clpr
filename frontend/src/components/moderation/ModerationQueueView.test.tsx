import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ModerationQueueView } from './ModerationQueueView';
import * as moderationApi from '../../lib/moderation-api';
import { BrowserRouter } from 'react-router-dom';

// Mock the dependencies
vi.mock('../../lib/moderation-api');
vi.mock('../../context/AuthContext', async () => {
    const actual = await vi.importActual('../../context/AuthContext');
    return {
        ...actual,
        useAuth: () => ({
            isAuthenticated: true,
            isAdmin: true,
            user: { id: 'user-1', username: 'admin' },
        }),
    };
});

const renderWithRouter = (component: React.ReactElement) => {
    return render(<BrowserRouter>{component}</BrowserRouter>);
};

describe('ModerationQueueView', () => {
    const mockItems: moderationApi.ModerationQueueItem[] = [
        {
            id: 'item-1',
            content_type: 'clip',
            content_id: 'clip-1',
            status: 'pending',
            created_at: '2024-01-01T00:00:00Z',
            reported_by: ['user-123'],
            report_count: 1,
            auto_flagged: false,
            priority: 50,
            reason: 'Inappropriate content',
        },
        {
            id: 'item-2',
            content_type: 'clip',
            content_id: 'clip-2',
            status: 'pending',
            created_at: '2024-01-02T00:00:00Z',
            reported_by: ['user-456'],
            report_count: 1,
            auto_flagged: false,
            priority: 30,
            reason: 'Spam',
        },
    ];

    const mockStats: moderationApi.ModerationQueueStats = {
        total_pending: 10,
        total_approved: 50,
        total_rejected: 20,
        total_escalated: 2,
        by_content_type: { clip: 5, comment: 5 },
        by_reason: {},
        auto_flagged_count: 3,
        user_reported_count: 7,
        high_priority_count: 2,
    };

    beforeEach(() => {
        vi.clearAllMocks();
        vi.mocked(moderationApi.getModerationQueue).mockResolvedValue({
            success: true,
            data: mockItems,
            meta: { count: 2, limit: 50, status: 'pending' },
        });
        vi.mocked(moderationApi.getModerationStats).mockResolvedValue({
            success: true,
            data: mockStats,
        });
    });

    describe('Rendering', () => {
        it('renders the component with title and description', async () => {
            renderWithRouter(
                <ModerationQueueView
                    contentType='clip'
                    title='Clip Moderation'
                    description='Review flagged clips'
                />,
            );

            expect(screen.getByText('Clip Moderation')).toBeInTheDocument();
            expect(
                screen.getByText('Review flagged clips'),
            ).toBeInTheDocument();
            await screen.findByText('clip-1');
        });

        it('displays loading state initially', () => {
            vi.mocked(moderationApi.getModerationQueue).mockImplementation(
                () => new Promise(() => {}),
            );
            vi.mocked(moderationApi.getModerationStats).mockImplementation(
                () => new Promise(() => {}),
            );
            renderWithRouter(
                <ModerationQueueView
                    contentType='clip'
                    title='Clip Moderation'
                    description='Review flagged clips'
                />,
            );

            expect(
                screen.getByRole('status', { hidden: true }),
            ).toBeInTheDocument();
        });

        it('loads and displays queue items', async () => {
            renderWithRouter(
                <ModerationQueueView
                    contentType='clip'
                    title='Clip Moderation'
                    description='Review flagged clips'
                />,
            );

            await waitFor(() => {
                expect(screen.getByText('clip-1')).toBeInTheDocument();
            });

            expect(screen.getByText('clip-2')).toBeInTheDocument();
        });

        it('loads stats on mount', async () => {
            renderWithRouter(
                <ModerationQueueView
                    contentType='clip'
                    title='Clip Moderation'
                    description='Review flagged clips'
                />,
            );

            await waitFor(() => {
                expect(moderationApi.getModerationStats).toHaveBeenCalled();
            });
        });
    });

    describe('Status Filtering', () => {
        it('loads pending items by default', async () => {
            renderWithRouter(
                <ModerationQueueView
                    contentType='clip'
                    title='Clip Moderation'
                    description='Review flagged clips'
                />,
            );

            await waitFor(() => {
                expect(moderationApi.getModerationQueue).toHaveBeenCalledWith(
                    'pending',
                    'clip',
                    50,
                );
            });
        });

        it('filters by approved status when selected', async () => {
            const user = userEvent.setup();
            renderWithRouter(
                <ModerationQueueView
                    contentType='clip'
                    title='Clip Moderation'
                    description='Review flagged clips'
                />,
            );

            await waitFor(() => {
                expect(screen.getByText('clip-1')).toBeInTheDocument();
            });

            // The component uses a <select> dropdown for status filtering
            const statusSelect = screen.getByRole('combobox');
            await user.selectOptions(statusSelect, 'approved');

            await waitFor(() => {
                expect(moderationApi.getModerationQueue).toHaveBeenCalledWith(
                    'approved',
                    'clip',
                    50,
                );
            });
        });
    });

    describe('Individual Item Actions', () => {
        it('approves a single item', async () => {
            const user = userEvent.setup();
            vi.mocked(moderationApi.approveQueueItem).mockResolvedValue({
                success: true,
                message: 'Item approved',
            });

            renderWithRouter(
                <ModerationQueueView
                    contentType='clip'
                    title='Clip Moderation'
                    description='Review flagged clips'
                />,
            );

            await waitFor(() => {
                expect(screen.getByText('clip-1')).toBeInTheDocument();
            });

            const approveButtons = screen.getAllByRole('button', {
                name: /approve/i,
            });
            await user.click(approveButtons[0]);

            await waitFor(() => {
                expect(moderationApi.approveQueueItem).toHaveBeenCalledWith(
                    'item-1',
                );
            });
        });

        it('rejects a single item with reason', async () => {
            const user = userEvent.setup();
            vi.mocked(moderationApi.rejectQueueItem).mockResolvedValue({
                success: true,
                message: 'Item rejected',
            });

            renderWithRouter(
                <ModerationQueueView
                    contentType='clip'
                    title='Clip Moderation'
                    description='Review flagged clips'
                />,
            );

            await waitFor(() => {
                expect(screen.getByText('clip-1')).toBeInTheDocument();
            });

            const rejectButtons = screen.getAllByRole('button', {
                name: /^reject$/i,
            });
            await user.click(rejectButtons[0]);

            // Modal should open
            await waitFor(() => {
                expect(screen.getByRole('dialog')).toBeInTheDocument();
            });

            // Enter rejection reason
            const reasonInput = screen.getByPlaceholderText(
                /reason for rejection/i,
            );
            await user.type(reasonInput, 'This violates our policy');

            // Submit rejection - the button says "Reject Clip"
            const confirmButton = screen.getByRole('button', {
                name: /reject clip/i,
            });
            await user.click(confirmButton);

            await waitFor(() => {
                expect(moderationApi.rejectQueueItem).toHaveBeenCalledWith(
                    'item-1',
                    'This violates our policy',
                );
            });
        });
    });

    describe('Bulk Actions', () => {
        it('selects multiple items', async () => {
            const user = userEvent.setup();
            renderWithRouter(
                <ModerationQueueView
                    contentType='clip'
                    title='Clip Moderation'
                    description='Review flagged clips'
                />,
            );

            await waitFor(() => {
                expect(screen.getByText('clip-1')).toBeInTheDocument();
            });

            const checkboxes = screen.getAllByRole('checkbox');
            // checkboxes[0] is "Select All", checkboxes[1] and [2] are the individual items
            await user.click(checkboxes[1]);
            await user.click(checkboxes[2]);

            expect(checkboxes[1]).toBeChecked();
            expect(checkboxes[2]).toBeChecked();
        });

        it('bulk approves selected items', async () => {
            const user = userEvent.setup();
            vi.mocked(moderationApi.bulkModerate).mockResolvedValue({
                success: true,
                processed: 2,
                total: 2,
            });

            renderWithRouter(
                <ModerationQueueView
                    contentType='clip'
                    title='Clip Moderation'
                    description='Review flagged clips'
                />,
            );

            await waitFor(() => {
                expect(screen.getByText('clip-1')).toBeInTheDocument();
            });

            // Select items (skip "Select All" at index 0)
            const checkboxes = screen.getAllByRole('checkbox');
            await user.click(checkboxes[1]);
            await user.click(checkboxes[2]);

            // Click bulk approve button - the component uses "Approve All (A)"
            const bulkApproveButton = screen.getByRole('button', {
                name: /approve all/i,
            });
            await user.click(bulkApproveButton);

            await waitFor(() => {
                expect(moderationApi.bulkModerate).toHaveBeenCalledWith({
                    item_ids: ['item-1', 'item-2'],
                    action: 'approve',
                });
            });
        });

        it('bulk rejects selected items', async () => {
            const user = userEvent.setup();
            vi.mocked(moderationApi.bulkModerate).mockResolvedValue({
                success: true,
                processed: 2,
                total: 2,
            });

            renderWithRouter(
                <ModerationQueueView
                    contentType='clip'
                    title='Clip Moderation'
                    description='Review flagged clips'
                />,
            );

            await waitFor(() => {
                expect(screen.getByText('clip-1')).toBeInTheDocument();
            });

            // Select items (skip "Select All" at index 0)
            const checkboxes = screen.getAllByRole('checkbox');
            await user.click(checkboxes[1]);
            await user.click(checkboxes[2]);

            // Click bulk reject button - the component uses "Reject All (R)"
            const bulkRejectButton = screen.getByRole('button', {
                name: /reject all/i,
            });
            await user.click(bulkRejectButton);

            await waitFor(() => {
                expect(moderationApi.bulkModerate).toHaveBeenCalledWith({
                    item_ids: ['item-1', 'item-2'],
                    action: 'reject',
                });
            });
        });
    });

    describe('Error Handling', () => {
        it('displays error message when loading fails', async () => {
            vi.mocked(moderationApi.getModerationQueue).mockRejectedValue(
                new Error('Failed to load queue'),
            );

            renderWithRouter(
                <ModerationQueueView
                    contentType='clip'
                    title='Clip Moderation'
                    description='Review flagged clips'
                />,
            );

            await waitFor(() => {
                expect(screen.getByText(/Failed to load/)).toBeInTheDocument();
            });
        });

        it('displays error when approve fails', async () => {
            const user = userEvent.setup();
            vi.mocked(moderationApi.approveQueueItem).mockRejectedValue(
                new Error('Failed to approve'),
            );

            renderWithRouter(
                <ModerationQueueView
                    contentType='clip'
                    title='Clip Moderation'
                    description='Review flagged clips'
                />,
            );

            await waitFor(() => {
                expect(screen.getByText('clip-1')).toBeInTheDocument();
            });

            const approveButtons = screen.getAllByRole('button', {
                name: /approve/i,
            });
            await user.click(approveButtons[0]);

            await waitFor(() => {
                expect(
                    screen.getByText(/Failed to approve/),
                ).toBeInTheDocument();
            });
        });
    });

    describe('Success Messages', () => {
        it('displays success message after approving items', async () => {
            const user = userEvent.setup();
            vi.mocked(moderationApi.approveQueueItem).mockResolvedValue({
                success: true,
                message: 'Item approved',
            });

            renderWithRouter(
                <ModerationQueueView
                    contentType='clip'
                    title='Clip Moderation'
                    description='Review flagged clips'
                />,
            );

            await waitFor(() => {
                expect(screen.getByText('clip-1')).toBeInTheDocument();
            });

            const approveButtons = screen.getAllByRole('button', {
                name: /approve/i,
            });
            await user.click(approveButtons[0]);

            await waitFor(() => {
                expect(
                    screen.getByText(/approved successfully/i),
                ).toBeInTheDocument();
            });
        });
    });

    describe('Empty State', () => {
        it('displays empty state when no items', async () => {
            vi.mocked(moderationApi.getModerationQueue).mockResolvedValue({
                success: true,
                data: [],
                meta: { count: 0, limit: 50, status: 'pending' },
            });

            renderWithRouter(
                <ModerationQueueView
                    contentType='clip'
                    title='Clip Moderation'
                    description='Review flagged clips'
                />,
            );

            await waitFor(() => {
                expect(
                    screen.getByText(/No.*in the moderation queue/i),
                ).toBeInTheDocument();
            });
        });
    });

    describe('Accessibility', () => {
        it('has proper heading structure', async () => {
            renderWithRouter(
                <ModerationQueueView
                    contentType='clip'
                    title='Clip Moderation'
                    description='Review flagged clips'
                />,
            );

            expect(
                screen.getByRole('heading', { name: 'Clip Moderation' }),
            ).toBeInTheDocument();
            await screen.findByText('clip-1');
        });

        it('uses checkboxes for item selection', async () => {
            renderWithRouter(
                <ModerationQueueView
                    contentType='clip'
                    title='Clip Moderation'
                    description='Review flagged clips'
                />,
            );

            await waitFor(() => {
                const checkboxes = screen.getAllByRole('checkbox');
                expect(checkboxes.length).toBeGreaterThan(0);
            });
        });
    });
});
