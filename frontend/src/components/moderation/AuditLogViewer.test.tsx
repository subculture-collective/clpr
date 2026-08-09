import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { AuditLogViewer } from './AuditLogViewer';
import * as moderationApi from '../../lib/moderation-api';
import { allowTestConsole } from '../../test/setup';

// Mock the API module
vi.mock('../../lib/moderation-api');

describe('AuditLogViewer', () => {
    const mockLogs: moderationApi.AuditLogEntry[] = [
        {
            id: 'log-1',
            action: 'ban',
            entityType: 'user',
            actor: {
                id: 'actor-1',
                username: 'mod_alice',
            },
            target: {
                id: 'target-1',
                username: 'user_bob',
            },
            reason: 'Spam violations',
            createdAt: '2024-01-15T10:30:00Z',
            metadata: {},
        },
        {
            id: 'log-2',
            action: 'approve',
            entityType: 'clip',
            actor: {
                id: 'actor-2',
                username: 'mod_charlie',
            },
            target: {
                id: 'target-2',
                username: 'creator_dana',
            },
            reason: 'Content approved',
            createdAt: '2024-01-15T11:00:00Z',
            metadata: {},
        },
        {
            id: 'log-3',
            action: 'timeout',
            entityType: 'user',
            actor: {
                id: 'actor-1',
                username: 'mod_alice',
            },
            target: {
                id: 'target-3',
                username: 'user_eve',
            },
            reason: 'Temporary mute for inappropriate language',
            createdAt: '2024-01-15T12:00:00Z',
            metadata: {},
        },
    ];

    beforeEach(() => {
        vi.clearAllMocks();
        // Default mock implementations
        vi.mocked(moderationApi.getAuditLogs).mockResolvedValue({
            logs: mockLogs,
            total: 3,
            limit: 50,
            offset: 0,
        });
        vi.mocked(moderationApi.exportAuditLogs).mockResolvedValue(
            new Blob(['test csv data'], { type: 'text/csv' })
        );
    });

    describe('Rendering', () => {
        it('renders the component with filters section', async () => {
            render(<AuditLogViewer />);

            await waitFor(() => {
                expect(screen.getByText('Filters')).toBeInTheDocument();
            });
        });

        it('displays loading state initially', () => {
            vi.mocked(moderationApi.getAuditLogs).mockImplementation(
                () => new Promise(() => {}),
            );
            render(<AuditLogViewer />);

            expect(screen.getByText('Loading audit logs...')).toBeInTheDocument();
        });

        it('displays audit logs after loading', async () => {
            render(<AuditLogViewer />);

            await waitFor(() => {
                expect(screen.getAllByText('mod_alice')).toHaveLength(2); // Actor appears twice
                expect(screen.getByText('user_bob')).toBeInTheDocument();
                expect(screen.getByText('Spam violations')).toBeInTheDocument();
            });
        });

        it('displays empty state when no logs', async () => {
            vi.mocked(moderationApi.getAuditLogs).mockResolvedValue({
                logs: [],
                total: 0,
                limit: 50,
                offset: 0,
            });

            render(<AuditLogViewer />);

            await waitFor(() => {
                expect(screen.getByText('No audit logs found')).toBeInTheDocument();
            });
        });

        it('displays error state on API failure', async () => {
            allowTestConsole('error', /Audit logs error:.*API Error/);
            vi.mocked(moderationApi.getAuditLogs).mockRejectedValue(
                new Error('API Error')
            );

            render(<AuditLogViewer />);

            await waitFor(() => {
                expect(screen.getByText('Failed to load audit logs')).toBeInTheDocument();
            });
        });
    });

    describe('Filters', () => {
        it('renders all filter inputs', async () => {
            render(<AuditLogViewer />);

            await waitFor(() => {
                expect(screen.getByLabelText('Action')).toBeInTheDocument();
                expect(screen.getByLabelText('Actor (User ID)')).toBeInTheDocument();
                expect(screen.getByLabelText('Target (User ID)')).toBeInTheDocument();
                expect(screen.getByLabelText('Start Date')).toBeInTheDocument();
                expect(screen.getByLabelText('End Date')).toBeInTheDocument();
                expect(screen.getByLabelText('Search (Reason)')).toBeInTheDocument();
                expect(screen.getByLabelText('Per Page')).toBeInTheDocument();
            });
        });

        it('applies action filter', async () => {
            const user = userEvent.setup();
            render(<AuditLogViewer />);

            await waitFor(() => {
                expect(screen.getByLabelText('Action')).toBeInTheDocument();
            });

            const actionSelect = screen.getByLabelText('Action');
            await user.selectOptions(actionSelect, 'ban');

            await waitFor(() => {
                expect(moderationApi.getAuditLogs).toHaveBeenCalledWith(
                    expect.objectContaining({ action: 'ban', offset: 0 })
                );
            });
        });

        it('applies target filter', async () => {
            const user = userEvent.setup();
            render(<AuditLogViewer />);

            await waitFor(() => {
                expect(screen.getByLabelText('Target (User ID)')).toBeInTheDocument();
            });

            const targetInput = screen.getByLabelText('Target (User ID)');
            await user.type(targetInput, 'target-123');

            await waitFor(() => {
                expect(moderationApi.getAuditLogs).toHaveBeenCalledWith(
                    expect.objectContaining({ target: 'target-123', offset: 0 })
                );
            });
        });

        it('applies search filter', async () => {
            const user = userEvent.setup();
            render(<AuditLogViewer />);

            await waitFor(() => {
                expect(screen.getByLabelText('Search (Reason)')).toBeInTheDocument();
            });

            const searchInput = screen.getByLabelText('Search (Reason)');
            await user.type(searchInput, 'spam');

            await waitFor(() => {
                expect(moderationApi.getAuditLogs).toHaveBeenCalledWith(
                    expect.objectContaining({ search: 'spam', offset: 0 })
                );
            });
        });

        it('resets offset when filter changes', async () => {
            const user = userEvent.setup();
            render(<AuditLogViewer />);

            await waitFor(() => {
                expect(screen.getByLabelText('Action')).toBeInTheDocument();
            });

            const actionSelect = screen.getByLabelText('Action');
            await user.selectOptions(actionSelect, 'ban');

            await waitFor(() => {
                expect(moderationApi.getAuditLogs).toHaveBeenCalledWith(
                    expect.objectContaining({ offset: 0 })
                );
            });
        });
    });

    describe('Pagination', () => {
        beforeEach(() => {
            vi.mocked(moderationApi.getAuditLogs).mockResolvedValue({
                logs: mockLogs,
                total: 100,
                limit: 50,
                offset: 0,
            });
        });

        it('displays pagination controls when there are multiple pages', async () => {
            render(<AuditLogViewer />);

            await waitFor(() => {
                expect(screen.getByText(/Page 1 of 2/)).toBeInTheDocument();
            });
        });

        it('navigates to next page', async () => {
            const user = userEvent.setup();
            render(<AuditLogViewer />);

            await waitFor(() => {
                expect(screen.getByText(/Next/)).toBeInTheDocument();
            });

            const nextButton = screen.getByText(/Next/);
            await user.click(nextButton);

            await waitFor(() => {
                expect(moderationApi.getAuditLogs).toHaveBeenCalledWith(
                    expect.objectContaining({ offset: 50 })
                );
            });
        });

        it('navigates to previous page', async () => {
            const user = userEvent.setup();
            vi.mocked(moderationApi.getAuditLogs).mockResolvedValue({
                logs: mockLogs,
                total: 100,
                limit: 50,
                offset: 50,
            });

            render(<AuditLogViewer />);

            await waitFor(() => {
                expect(screen.getByText(/Previous/)).toBeInTheDocument();
            });

            const prevButton = screen.getByText(/Previous/);
            await user.click(prevButton);

            await waitFor(() => {
                expect(moderationApi.getAuditLogs).toHaveBeenCalledWith(
                    expect.objectContaining({ offset: 0 })
                );
            });
        });

        it('disables previous button on first page', async () => {
            vi.mocked(moderationApi.getAuditLogs).mockResolvedValue({
                logs: mockLogs,
                total: 100, // Ensure multiple pages
                limit: 50,
                offset: 0,
            });

            render(<AuditLogViewer />);

            await waitFor(() => {
                const buttons = screen.getAllByRole('button');
                const prevButton = buttons.find(b => b.textContent?.includes('Previous'));
                expect(prevButton).toBeDisabled();
            });
        });

        it('does not show pagination when only one page', async () => {
            vi.mocked(moderationApi.getAuditLogs).mockResolvedValue({
                logs: mockLogs,
                total: 3, // Only 3 logs total, fits in one page
                limit: 50,
                offset: 0,
            });

            render(<AuditLogViewer />);

            await waitFor(() => {
                // Pagination controls should not be visible when totalPages <= 1
                const buttons = screen.queryAllByRole('button');
                const hasPrevButton = buttons.some(b => b.textContent?.includes('Previous'));
                expect(hasPrevButton).toBe(false);
            });
        });
    });

    describe('Export', () => {
        it('renders export button', async () => {
            render(<AuditLogViewer />);

            await waitFor(() => {
                expect(screen.getByText(/Export CSV/)).toBeInTheDocument();
            });
        });

        it('disables export button when no logs', async () => {
            vi.mocked(moderationApi.getAuditLogs).mockResolvedValue({
                logs: [],
                total: 0,
                limit: 50,
                offset: 0,
            });

            render(<AuditLogViewer />);

            await waitFor(() => {
                const exportButton = screen.getByTitle('Export to CSV').closest('button');
                expect(exportButton).toBeDisabled();
            });
        });

        it('calls export API when export button clicked', async () => {
            const user = userEvent.setup();
            render(<AuditLogViewer />);

            await waitFor(() => {
                expect(screen.getByText(/Export CSV/)).toBeInTheDocument();
            });

            const exportButton = screen.getByText(/Export CSV/);
            await user.click(exportButton);

            await waitFor(() => {
                expect(moderationApi.exportAuditLogs).toHaveBeenCalled();
            });
        });
    });

    describe('Refresh', () => {
        it('renders refresh button', async () => {
            render(<AuditLogViewer />);

            await waitFor(() => {
                expect(screen.getByTitle('Refresh logs')).toBeInTheDocument();
            });
        });

        it('reloads logs when refresh button clicked', async () => {
            const user = userEvent.setup();
            render(<AuditLogViewer />);

            await waitFor(() => {
                expect(screen.getByTitle('Refresh logs')).toBeInTheDocument();
            });

            vi.clearAllMocks();
            const refreshButton = screen.getByTitle('Refresh logs');
            await user.click(refreshButton);

            await waitFor(() => {
                expect(moderationApi.getAuditLogs).toHaveBeenCalled();
            });
        });
    });

    describe('Action Icons and Colors', () => {
        it('displays ban action with red color', async () => {
            render(<AuditLogViewer />);

            await waitFor(() => {
                const banBadge = screen.getByText('ban');
                expect(banBadge).toHaveClass('bg-red-100', 'text-red-800');
            });
        });

        it('displays approve action with green color', async () => {
            render(<AuditLogViewer />);

            await waitFor(() => {
                const approveBadge = screen.getByText('approve');
                expect(approveBadge).toHaveClass('bg-green-100', 'text-green-800');
            });
        });

        it('displays timeout action with yellow color', async () => {
            render(<AuditLogViewer />);

            await waitFor(() => {
                const timeoutBadge = screen.getByText('timeout');
                expect(timeoutBadge).toHaveClass('bg-yellow-100', 'text-yellow-800');
            });
        });
    });

    describe('Date Formatting', () => {
        it('formats timestamps correctly', async () => {
            render(<AuditLogViewer />);

            await waitFor(() => {
                const dates = screen.getAllByText('Jan 15, 2024');
                expect(dates.length).toBeGreaterThan(0); // Should appear for all 3 logs
            });
        });
    });
});
