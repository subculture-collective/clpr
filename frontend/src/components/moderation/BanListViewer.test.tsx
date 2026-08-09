import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { BanListViewer } from './BanListViewer';
import * as chatApi from '../../lib/chat-api';
import { allowTestConsole } from '../../test/setup';

// Mock the API module
vi.mock('../../lib/chat-api');

describe('BanListViewer', () => {
    const mockChannelId = 'channel-123';
    const mockBans: chatApi.ChatBan[] = [
        {
            id: 'ban-1',
            channel_id: mockChannelId,
            user_id: 'user-1',
            banned_by: 'moderator-1',
            reason: 'Spam',
            created_at: '2024-01-01T00:00:00Z',
            expires_at: '2099-12-31T23:59:59Z',
            target_username: 'spammer123',
            banned_by_username: 'mod_john',
        },
        {
            id: 'ban-2',
            channel_id: mockChannelId,
            user_id: 'user-2',
            banned_by: 'moderator-1',
            reason: 'Harassment',
            created_at: '2024-01-02T00:00:00Z',
            expires_at: undefined,
            target_username: 'harasser456',
            banned_by_username: 'mod_john',
        },
        {
            id: 'ban-3',
            channel_id: mockChannelId,
            user_id: 'user-3',
            banned_by: 'moderator-2',
            reason: 'Toxic behavior',
            created_at: '2023-12-01T00:00:00Z',
            expires_at: '2023-12-31T23:59:59Z',
            target_username: 'toxic789',
            banned_by_username: 'mod_jane',
        },
    ];

    beforeEach(() => {
        vi.clearAllMocks();
        // Default mock implementations
        vi.mocked(chatApi.getChannelBans).mockResolvedValue({
            bans: mockBans,
            total: 3,
            page: 1,
            limit: 50,
        });
    });

    describe('Rendering', () => {
        it('renders the component with title', async () => {
            render(<BanListViewer channelId={mockChannelId} />);

            await waitFor(() => {
                expect(screen.getByText('Channel Bans')).toBeInTheDocument();
            });
        });

        it('displays loading state initially', () => {
            vi.mocked(chatApi.getChannelBans).mockImplementation(
                () => new Promise(() => {}),
            );
            render(<BanListViewer channelId={mockChannelId} />);

            expect(screen.getByText('Loading bans...')).toBeInTheDocument();
        });

        it('displays bans list after loading', async () => {
            render(<BanListViewer channelId={mockChannelId} />);

            await waitFor(() => {
                expect(screen.getByText('spammer123')).toBeInTheDocument();
            });

            expect(screen.getByText('harasser456')).toBeInTheDocument();
            expect(screen.getByText('toxic789')).toBeInTheDocument();
        });

        it('displays ban status correctly', async () => {
            render(<BanListViewer channelId={mockChannelId} />);

            await waitFor(() => {
                expect(screen.getByText('spammer123')).toBeInTheDocument();
            });

            // Check for status badges
            expect(screen.getAllByText('Active').length).toBeGreaterThan(0);
            expect(screen.getAllByText('Permanent').length).toBeGreaterThan(0);
            expect(screen.getAllByText('Expired').length).toBeGreaterThan(0);
        });

        it('shows Export to CSV button', async () => {
            render(<BanListViewer channelId={mockChannelId} />);

            await waitFor(() => {
                expect(
                    screen.getByLabelText('Export bans to CSV')
                ).toBeInTheDocument();
            });
        });

        it('does not show Revoke button when canManage is false', async () => {
            render(
                <BanListViewer channelId={mockChannelId} canManage={false} />
            );

            await waitFor(() => {
                expect(screen.getByText('spammer123')).toBeInTheDocument();
            });

            expect(screen.queryByText('Revoke')).not.toBeInTheDocument();
        });

        it('shows Revoke button when canManage is true', async () => {
            render(
                <BanListViewer channelId={mockChannelId} canManage={true} />
            );

            await waitFor(() => {
                expect(screen.getByText('spammer123')).toBeInTheDocument();
            });

            // Should have revoke buttons for active bans only
            const revokeButtons = screen.getAllByText('Revoke');
            expect(revokeButtons.length).toBeGreaterThan(0);
        });
    });

    describe('Filtering', () => {
        it('filters bans by username', async () => {
            render(<BanListViewer channelId={mockChannelId} />);

            await waitFor(() => {
                expect(screen.getByText('spammer123')).toBeInTheDocument();
            });

            const userFilter = screen.getByLabelText('User');
            await userEvent.type(userFilter, 'spammer');

            await waitFor(() => {
                expect(screen.getByText('spammer123')).toBeInTheDocument();
                expect(
                    screen.queryByText('harasser456')
                ).not.toBeInTheDocument();
            });
        });

        it('filters bans by reason', async () => {
            render(<BanListViewer channelId={mockChannelId} />);

            await waitFor(() => {
                expect(screen.getByText('spammer123')).toBeInTheDocument();
            });

            const reasonFilter = screen.getByLabelText('Reason');
            await userEvent.type(reasonFilter, 'Spam');

            await waitFor(() => {
                expect(screen.getByText('spammer123')).toBeInTheDocument();
                expect(
                    screen.queryByText('harasser456')
                ).not.toBeInTheDocument();
            });
        });

        it('filters bans by status', async () => {
            render(<BanListViewer channelId={mockChannelId} />);

            await waitFor(() => {
                expect(screen.getByText('spammer123')).toBeInTheDocument();
            });

            const statusFilter = screen.getByLabelText('Status');
            await userEvent.selectOptions(statusFilter, 'active');

            await waitFor(() => {
                // Only active bans should be visible (harasser456 is permanent which counts as active)
                expect(screen.getByText('harasser456')).toBeInTheDocument();
                // Expired ban should not be visible
                expect(screen.queryByText('toxic789')).not.toBeInTheDocument();
            });
        });

        it('shows no results message when no bans match filters', async () => {
            render(<BanListViewer channelId={mockChannelId} />);

            await waitFor(() => {
                expect(screen.getByText('spammer123')).toBeInTheDocument();
            });

            const userFilter = screen.getByLabelText('User');
            await userEvent.type(userFilter, 'nonexistent');

            await waitFor(() => {
                expect(
                    screen.getByText('No bans available')
                ).toBeInTheDocument();
            });
        });
    });

    describe('Sorting', () => {
        it('sorts by username when header is clicked', async () => {
            render(<BanListViewer channelId={mockChannelId} />);

            await waitFor(() => {
                expect(screen.getByText('spammer123')).toBeInTheDocument();
            });

            const usernameHeader = screen.getByRole('columnheader', {
                name: /User/,
            });
            await userEvent.click(usernameHeader);

            // Check that sort indicator is shown
            expect(usernameHeader.textContent).toContain('↓');
        });

        it('toggles sort direction when same header is clicked twice', async () => {
            render(<BanListViewer channelId={mockChannelId} />);

            await waitFor(() => {
                expect(screen.getByText('spammer123')).toBeInTheDocument();
            });

            const usernameHeader = screen.getByRole('columnheader', {
                name: /User/,
            });
            await userEvent.click(usernameHeader);

            expect(usernameHeader.textContent).toContain('↓');

            await userEvent.click(usernameHeader);

            expect(usernameHeader.textContent).toContain('↑');
        });
    });

    describe('Pagination', () => {
        it('does not show pagination when total pages is 1', async () => {
            render(<BanListViewer channelId={mockChannelId} />);

            await waitFor(() => {
                expect(screen.getByText('spammer123')).toBeInTheDocument();
            });

            expect(
                screen.queryByLabelText('Pagination')
            ).not.toBeInTheDocument();
        });

        it('shows pagination when there are multiple pages', async () => {
            vi.mocked(chatApi.getChannelBans).mockResolvedValue({
                bans: mockBans,
                total: 150,
                page: 1,
                limit: 50,
            });

            render(<BanListViewer channelId={mockChannelId} />);

            await waitFor(() => {
                expect(screen.getByLabelText('Pagination')).toBeInTheDocument();
            });

            expect(screen.getByText(/Showing page 1 of 3/)).toBeInTheDocument();
        });

        it('navigates to next page when Next button is clicked', async () => {
            vi.mocked(chatApi.getChannelBans).mockResolvedValue({
                bans: mockBans,
                total: 150,
                page: 1,
                limit: 50,
            });

            render(<BanListViewer channelId={mockChannelId} />);

            await waitFor(() => {
                expect(screen.getByLabelText('Pagination')).toBeInTheDocument();
            });

            const nextButton = screen.getByLabelText('Next page');
            await userEvent.click(nextButton);

            // Verify API was called with page 2
            await waitFor(() => {
                expect(chatApi.getChannelBans).toHaveBeenCalledWith(
                    mockChannelId,
                    2,
                    50
                );
            });
        });
    });

    describe('Revoke Ban', () => {
        it('opens revoke confirmation modal when revoke is clicked', async () => {
            render(
                <BanListViewer channelId={mockChannelId} canManage={true} />
            );

            await waitFor(() => {
                expect(screen.getByText('spammer123')).toBeInTheDocument();
            });

            const revokeButtons = screen.getAllByText('Revoke');
            await userEvent.click(revokeButtons[0]);

            expect(screen.getByRole('dialog')).toBeInTheDocument();
            expect(
                screen.getByText(/Are you sure you want to revoke the ban for/)
            ).toBeInTheDocument();
        });

        it('closes revoke modal when cancel is clicked', async () => {
            render(
                <BanListViewer channelId={mockChannelId} canManage={true} />
            );

            await waitFor(() => {
                expect(screen.getByText('spammer123')).toBeInTheDocument();
            });

            const revokeButtons = screen.getAllByText('Revoke');
            await userEvent.click(revokeButtons[0]);

            const dialog = screen.getByRole('dialog');
            expect(dialog).toBeInTheDocument();

            const cancelButton = screen.getByRole('button', { name: 'Cancel' });
            await userEvent.click(cancelButton);

            await waitFor(() => {
                expect(
                    screen.queryByText('Revoke Ban')
                ).not.toBeInTheDocument();
            });
        });

        it('successfully revokes a ban', async () => {
            vi.mocked(chatApi.unbanUser).mockResolvedValue({
                status: 'unbanned',
            });

            render(
                <BanListViewer channelId={mockChannelId} canManage={true} />
            );

            await waitFor(() => {
                expect(screen.getByText('spammer123')).toBeInTheDocument();
            });

            // Click revoke button - the first active ban in the rendered list (sorted by created_at desc)
            // is harasser456 (Permanent, created 2024-01-02) which is user-2
            const revokeButtons = screen.getAllByText('Revoke');
            await userEvent.click(revokeButtons[0]);

            // Confirm revocation
            const confirmButton = screen.getByRole('button', {
                name: 'Revoke Ban',
            });
            await userEvent.click(confirmButton);

            // Verify unbanUser was called with correct ban ID (harasser456 -> ban-2)
            await waitFor(() => {
                expect(chatApi.unbanUser).toHaveBeenCalledWith('ban-2');
            });

            // Verify list was reloaded
            await waitFor(() => {
                expect(chatApi.getChannelBans).toHaveBeenCalledTimes(2); // Initial load + reload after revoke
            });
        });

        it('handles error when revoking a ban fails', async () => {
            vi.mocked(chatApi.unbanUser).mockRejectedValue(
                new Error('Failed to revoke ban')
            );

            render(
                <BanListViewer channelId={mockChannelId} canManage={true} />
            );

            await waitFor(() => {
                expect(screen.getByText('spammer123')).toBeInTheDocument();
            });

            // Click revoke button
            const revokeButtons = screen.getAllByText('Revoke');
            await userEvent.click(revokeButtons[0]);

            // Confirm revocation
            const confirmButton = screen.getByRole('button', {
                name: 'Revoke Ban',
            });
            await userEvent.click(confirmButton);

            // Verify error is displayed (may appear in multiple alert elements) and modal is closed
            await waitFor(() => {
                const errorElements = screen.getAllByText(
                    'Failed to revoke ban'
                );
                expect(errorElements.length).toBeGreaterThan(0);
                expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
            });
        });
    });

    describe('View Details', () => {
        it('opens details modal when details button is clicked', async () => {
            render(<BanListViewer channelId={mockChannelId} />);

            await waitFor(() => {
                expect(screen.getByText('spammer123')).toBeInTheDocument();
            });

            const detailsButtons = screen.getAllByText('Details');
            await userEvent.click(detailsButtons[0]);

            expect(screen.getByText('Ban Details')).toBeInTheDocument();
        });

        it('displays ban details in modal', async () => {
            render(<BanListViewer channelId={mockChannelId} />);

            await waitFor(() => {
                expect(screen.getByText('spammer123')).toBeInTheDocument();
            });

            const detailsButtons = screen.getAllByText('Details');
            await userEvent.click(detailsButtons[0]);

            // Check that details are shown - use getAllByText for duplicate labels
            expect(screen.getByText('Banned User')).toBeInTheDocument();
            expect(screen.getAllByText('Banned By').length).toBeGreaterThan(0);
            expect(screen.getByText('Spam')).toBeInTheDocument();
        });
    });

    describe('Error handling', () => {
        it('displays error message when loading fails', async () => {
            allowTestConsole('error', /Bans loading error:.*Failed to load bans/);
            vi.mocked(chatApi.getChannelBans).mockRejectedValue(
                new Error('Failed to load bans')
            );

            render(<BanListViewer channelId={mockChannelId} />);

            await waitFor(() => {
                expect(
                    screen.getByText('Failed to load bans')
                ).toBeInTheDocument();
            });

            expect(screen.getByText('Retry')).toBeInTheDocument();
        });

        it('allows retrying after error', async () => {
            allowTestConsole('error', /Bans loading error:.*Failed to load bans/);
            vi.mocked(chatApi.getChannelBans).mockRejectedValueOnce(
                new Error('Failed to load bans')
            );

            render(<BanListViewer channelId={mockChannelId} />);

            await waitFor(() => {
                expect(
                    screen.getByText('Failed to load bans')
                ).toBeInTheDocument();
            });

            // Mock successful response for retry
            vi.mocked(chatApi.getChannelBans).mockResolvedValueOnce({
                bans: mockBans,
                total: 3,
                page: 1,
                limit: 50,
            });

            const retryButton = screen.getByText('Retry');
            await userEvent.click(retryButton);

            await waitFor(() => {
                expect(screen.getByText('spammer123')).toBeInTheDocument();
            });
        });
    });

    describe('Export to CSV', () => {
        it('is disabled when no bans are present', async () => {
            vi.mocked(chatApi.getChannelBans).mockResolvedValue({
                bans: [],
                total: 0,
                page: 1,
                limit: 50,
            });

            render(<BanListViewer channelId={mockChannelId} />);

            await waitFor(() => {
                expect(
                    screen.getByText('No bans available')
                ).toBeInTheDocument();
            });

            const exportButton = screen.getByLabelText('Export bans to CSV');
            expect(exportButton).toBeDisabled();
        });

        it('is enabled when bans are present', async () => {
            render(<BanListViewer channelId={mockChannelId} />);

            await waitFor(() => {
                expect(screen.getByText('spammer123')).toBeInTheDocument();
            });

            const exportButton = screen.getByLabelText('Export bans to CSV');
            expect(exportButton).not.toBeDisabled();
        });
    });

    describe('Empty state', () => {
        it('shows empty state message when no bans exist', async () => {
            vi.mocked(chatApi.getChannelBans).mockResolvedValue({
                bans: [],
                total: 0,
                page: 1,
                limit: 50,
            });

            render(<BanListViewer channelId={mockChannelId} />);

            await waitFor(() => {
                expect(
                    screen.getByText('No bans available')
                ).toBeInTheDocument();
            });
        });
    });

    describe('Accessibility', () => {
        it('has proper ARIA labels on interactive elements', async () => {
            render(
                <BanListViewer channelId={mockChannelId} canManage={true} />
            );

            await waitFor(() => {
                expect(screen.getByText('spammer123')).toBeInTheDocument();
            });

            expect(
                screen.getByLabelText('Export bans to CSV')
            ).toBeInTheDocument();
            expect(screen.getByLabelText('User')).toBeInTheDocument();
            expect(screen.getByLabelText('Reason')).toBeInTheDocument();
            expect(screen.getByLabelText('Status')).toBeInTheDocument();
        });

        it('has proper table structure with headers', async () => {
            render(<BanListViewer channelId={mockChannelId} />);

            await waitFor(() => {
                expect(screen.getByText('spammer123')).toBeInTheDocument();
            });

            const table = screen.getByRole('table');
            expect(table).toBeInTheDocument();

            const headers = screen.getAllByRole('columnheader');
            expect(headers.length).toBeGreaterThan(0);
        });
    });
});
