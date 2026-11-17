import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useAuth } from '../../context/AuthContext';
import * as moderationApi from '../../lib/moderation-api';
import userEvent from '@testing-library/user-event';
import { fireEvent, render, screen, waitFor } from '../../test/test-utils';
import { AdminModerationQueuePage } from './AdminModerationQueuePage';

// Mock the API calls
vi.mock('../../lib/moderation-api', () => ({
    getModerationQueue: vi.fn(),
    getModerationStats: vi.fn(),
    approveQueueItem: vi.fn(),
    rejectQueueItem: vi.fn(),
    bulkModerate: vi.fn(),
}));

// Mock the AuthContext
vi.mock('../../context/AuthContext', async () => {
    const actual = await vi.importActual('../../context/AuthContext');
    return {
        ...actual,
        useAuth: vi.fn(),
    };
});

// Mock react-router-dom
const mockNavigate = vi.fn();
vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual('react-router-dom');
    return {
        ...actual,
        useNavigate: () => mockNavigate,
    };
});

describe('AdminModerationQueuePage', () => {
    const mockGetModerationQueue = vi.mocked(moderationApi.getModerationQueue);
    const mockGetModerationStats = vi.mocked(moderationApi.getModerationStats);
    const mockUseAuth = vi.mocked(useAuth);

    beforeEach(() => {
        vi.clearAllMocks();
        mockUseAuth.mockReturnValue({
            user: {
                id: 'admin-123',
                twitch_id: 'twitch-123',
                username: 'adminuser',
                display_name: 'Admin User',
                karma_points: 1000,
                role: 'admin' as const,
                is_banned: false,
                created_at: '2024-01-01T00:00:00Z',
            },
            isAuthenticated: true,
            isAdmin: true,
            isModerator: false,
            isModeratorOrAdmin: true,
            login: vi.fn(),
            logout: vi.fn(),
            isLoading: false,
            refreshUser: vi.fn(),
        });
    });

    describe('Queue Loading', () => {
        it('should load and display moderation queue items', async () => {
            mockGetModerationQueue.mockResolvedValue({
                success: true,
                data: [
                    {
                        id: 'item-1',
                        content_type: 'comment',
                        content_id: 'comment-1',
                        reason: 'spam',
                        priority: 75,
                        status: 'pending',
                        reported_by: ['user-1'],
                        report_count: 1,
                        auto_flagged: false,
                        created_at: '2024-01-01T00:00:00Z',
                    },
                ],
                meta: {
                    count: 1,
                    limit: 50,
                    status: 'pending',
                },
            });

            mockGetModerationStats.mockResolvedValue({
                success: true,
                data: {
                    total_pending: 1,
                    total_approved: 0,
                    total_rejected: 0,
                    total_escalated: 0,
                    by_content_type: { comment: 1 },
                    by_reason: { spam: 1 },
                    auto_flagged_count: 0,
                    user_reported_count: 1,
                    high_priority_count: 1,
                },
            });

            render(<AdminModerationQueuePage />);

            // Wait for the page title to appear
            await waitFor(() => {
                expect(
                    screen.getByText('Moderation Queue')
                ).toBeInTheDocument();
            });

            // Page should render without errors
            expect(screen.getByText('Moderation Queue')).toBeInTheDocument();
        });

        it('allows a moderator to load an empty review queue', async () => {
            mockUseAuth.mockReturnValue({ ...mockUseAuth(), isAdmin: false, isModerator: true, isModeratorOrAdmin: true });
            mockGetModerationQueue.mockResolvedValue({
                success: true,
                data: [],
                meta: {
                    count: 0,
                    limit: 50,
                    status: 'pending',
                },
            });

            mockGetModerationStats.mockResolvedValue({
                success: true,
                data: {
                    total_pending: 0,
                    total_approved: 0,
                    total_rejected: 0,
                    total_escalated: 0,
                    by_content_type: {},
                    by_reason: {},
                    auto_flagged_count: 0,
                    user_reported_count: 0,
                    high_priority_count: 0,
                },
            });

            render(<AdminModerationQueuePage />);

            await waitFor(() => {
                expect(
                    screen.getByText('No items in the queue.')
                ).toBeInTheDocument();
            });
        });

        it('should display error message when API fails', async () => {
            mockGetModerationQueue.mockRejectedValue({
                response: {
                    data: {
                        error: 'Failed to load queue',
                    },
                },
            });

            mockGetModerationStats.mockResolvedValue({
                success: true,
                data: {
                    total_pending: 0,
                    total_approved: 0,
                    total_rejected: 0,
                    total_escalated: 0,
                    by_content_type: {},
                    by_reason: {},
                    auto_flagged_count: 0,
                    user_reported_count: 0,
                    high_priority_count: 0,
                },
            });

            render(<AdminModerationQueuePage />);

            await waitFor(() => {
                expect(
                    screen.getByText('Failed to load queue')
                ).toBeInTheDocument();
            });
        });
    });

    describe('Bulk Selection', () => {
        it('should allow selecting items', async () => {
            mockGetModerationQueue.mockResolvedValue({
                success: true,
                data: [
                    {
                        id: 'item-1',
                        content_type: 'comment',
                        content_id: 'comment-1',
                        reason: 'spam',
                        priority: 75,
                        status: 'pending',
                        reported_by: ['user-1'],
                        report_count: 1,
                        auto_flagged: false,
                        created_at: '2024-01-01T00:00:00Z',
                    },
                ],
                meta: {
                    count: 1,
                    limit: 50,
                    status: 'pending',
                },
            });

            mockGetModerationStats.mockResolvedValue({
                success: true,
                data: {
                    total_pending: 1,
                    total_approved: 0,
                    total_rejected: 0,
                    total_escalated: 0,
                    by_content_type: { comment: 1 },
                    by_reason: { spam: 1 },
                    auto_flagged_count: 0,
                    user_reported_count: 1,
                    high_priority_count: 1,
                },
            });

            render(<AdminModerationQueuePage />);

            await waitFor(() => {
                expect(
                    screen.getByText('Moderation Queue')
                ).toBeInTheDocument();
            });

            // Should not show bulk actions bar initially
            expect(screen.queryByText(/selected/)).not.toBeInTheDocument();
        });
    });

    describe('Review actions', () => {
        const mockBulkModerate = vi.mocked(moderationApi.bulkModerate);
        const mockApprove = vi.mocked(moderationApi.approveQueueItem);
        const pendingItem = {
            id: 'item-1',
            content_type: 'comment',
            content_id: 'comment-1',
            reason: 'toxicity',
            priority: 75,
            status: 'pending',
            reported_by: [],
            report_count: 0,
            auto_flagged: true,
            created_at: '2024-01-01T00:00:00Z',
        };
        const queueOf = (items: moderationApi.ModerationQueueItem[]) => ({
            success: true,
            data: items,
            meta: { count: items.length, limit: 50, status: 'pending' },
        });

        beforeEach(() => {
            mockGetModerationQueue.mockResolvedValue(queueOf([pendingItem]));
            mockGetModerationStats.mockResolvedValue({
                success: true,
                data: {
                    total_pending: 1,
                    total_approved: 0,
                    total_rejected: 0,
                    total_escalated: 0,
                    by_content_type: { comment: 1 },
                    by_reason: { toxicity: 1 },
                    auto_flagged_count: 1,
                    user_reported_count: 0,
                    high_priority_count: 1,
                },
            });
        });

        const selectFirstItem = async () => {
            const user = userEvent.setup();
            await screen.findByText('comment-1');
            await user.click(screen.getByRole('checkbox', { name: 'Select comment comment-1' }));
            return user;
        };

        it('renders a structured API error as text', async () => {
            mockGetModerationQueue.mockRejectedValue({
                response: {
                    status: 403,
                    data: { error: { code: 'MFA_REQUIRED', message: 'Set up two-factor authentication to continue.' } },
                },
            });

            render(<AdminModerationQueuePage />);

            expect(
                await screen.findByText('Set up two-factor authentication to continue.')
            ).toBeInTheDocument();
        });

        it('shows the excerpt, author and link when the API sends them', async () => {
            mockGetModerationQueue.mockResolvedValue(
                queueOf([
                    {
                        ...pendingItem,
                        content_excerpt: 'you are all idiots',
                        content_author: 'angry_viewer',
                        content_url: '/clip/abc#comment-1',
                    },
                ])
            );

            render(<AdminModerationQueuePage />);

            expect(await screen.findByText('you are all idiots')).toBeInTheDocument();
            expect(screen.getByText('angry_viewer')).toBeInTheDocument();
            expect(screen.getByRole('link', { name: 'Open comment' })).toHaveAttribute(
                'href',
                '/clip/abc#comment-1'
            );
            expect(screen.queryByText('comment-1')).not.toBeInTheDocument();
        });

        it('ignores shortcuts pressed with Ctrl or Cmd, or inside a dropdown', async () => {
            render(<AdminModerationQueuePage />);
            await selectFirstItem();

            fireEvent.keyDown(document.body, { key: 'r', ctrlKey: true });
            fireEvent.keyDown(document.body, { key: 'r', metaKey: true });
            fireEvent.keyDown(document.body, { key: 'a', ctrlKey: true });
            fireEvent.keyDown(screen.getAllByRole('combobox')[0], { key: 'r' });
            fireEvent.keyDown(screen.getAllByRole('combobox')[0], { key: 'a' });

            expect(mockBulkModerate).not.toHaveBeenCalled();
            expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
        });

        it('asks for confirmation before a bulk reject', async () => {
            mockBulkModerate.mockResolvedValue({ success: true, processed: 1, total: 1 });
            render(<AdminModerationQueuePage />);
            const user = await selectFirstItem();

            fireEvent.keyDown(document.body, { key: 'r' });

            expect(await screen.findByRole('dialog')).toBeInTheDocument();
            expect(mockBulkModerate).not.toHaveBeenCalled();

            await user.click(screen.getByRole('button', { name: 'Reject and remove' }));

            await waitFor(() =>
                expect(mockBulkModerate).toHaveBeenCalledWith({ item_ids: ['item-1'], action: 'reject' })
            );
        });

        it('shows how many items a bulk action skipped', async () => {
            mockBulkModerate.mockResolvedValue({
                success: true,
                processed: 0,
                total: 1,
                skipped: ['item-1'],
            });
            render(<AdminModerationQueuePage />);
            await selectFirstItem();

            fireEvent.keyDown(document.body, { key: 'a' });

            expect(
                await screen.findByText(/1 skipped because it was no longer pending/)
            ).toBeInTheDocument();
        });

        it('reloads the queue when an item was already reviewed', async () => {
            mockApprove.mockRejectedValue({
                response: { status: 409, data: { error: 'Item is no longer pending' } },
            });
            render(<AdminModerationQueuePage />);
            const user = userEvent.setup();
            await screen.findByText('comment-1');
            expect(mockGetModerationQueue).toHaveBeenCalledTimes(1);

            await user.click(screen.getByRole('button', { name: 'Approve' }));

            await waitFor(() => expect(mockGetModerationQueue).toHaveBeenCalledTimes(2));
            expect(await screen.findByText('Item is no longer pending')).toBeInTheDocument();
        });
    });
});
