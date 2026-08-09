import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { AppealsQueue } from './AppealsQueue';
import * as moderationApi from '../../lib/moderation-api';

// Mock the API module
vi.mock('../../lib/moderation-api');
// Mock the AppealResolutionModal component
vi.mock('./AppealResolutionModal', () => ({
    AppealResolutionModal: ({
        open,
        appeal,
        onClose,
        onSuccess,
    }: {
        open: boolean;
        appeal: { username: string };
        onClose: () => void;
        onSuccess: () => void;
    }) => {
        if (!open) return null;
        return (
            <div role='dialog' data-testid='resolution-modal'>
                <div>Resolve Appeal for {appeal?.username}</div>
                <button onClick={onClose}>Close Modal</button>
                <button onClick={onSuccess}>Resolve Success</button>
            </div>
        );
    },
}));

describe('AppealsQueue', () => {
    const mockAppeals: moderationApi.ModerationAppeal[] = [
        {
            id: 'appeal-1',
            moderation_action_id: 'action-1',
            user_id: 'user-1',
            username: 'user_one',
            reason: 'I did not violate any rules',
            status: 'pending',
            created_at: '2024-01-01T00:00:00Z',
            content_type: 'clip',
            content_id: 'clip-1',
            decision_action: 'ban',
            decision_reason: 'Spam',
        },
        {
            id: 'appeal-2',
            moderation_action_id: 'action-2',
            user_id: 'user-2',
            username: 'user_two',
            reason: 'This was a mistake',
            status: 'pending',
            created_at: '2024-01-02T00:00:00Z',
            content_type: 'comment',
            content_id: 'comment-1',
            decision_action: 'remove',
            decision_reason: 'Inappropriate content',
        },
    ];

    beforeEach(() => {
        vi.clearAllMocks();
        vi.mocked(moderationApi.getAdminAppeals).mockResolvedValue({
            success: true,
            data: mockAppeals,
        });
    });

    describe('Rendering', () => {
        it('renders the component with title', async () => {
            render(<AppealsQueue />);

            expect(screen.getByText('Appeals Queue')).toBeInTheDocument();
            await screen.findByText('user_one');
        });

        it('displays loading state initially', () => {
            vi.mocked(moderationApi.getAdminAppeals).mockImplementation(
                () => new Promise(() => {})
            );
            const { container } = render(<AppealsQueue />);

            const spinner = container.querySelector('.animate-spin');
            expect(spinner).toBeInTheDocument();
        });

        it('displays appeals after loading', async () => {
            render(<AppealsQueue />);

            await waitFor(() => {
                expect(screen.getByText('user_one')).toBeInTheDocument();
            });

            expect(screen.getByText('user_two')).toBeInTheDocument();
        });

        it('displays status filter buttons', async () => {
            render(<AppealsQueue />);

            expect(
                screen.getByRole('button', { name: 'Pending' })
            ).toBeInTheDocument();
            expect(
                screen.getByRole('button', { name: 'Approved' })
            ).toBeInTheDocument();
            expect(
                screen.getByRole('button', { name: 'Rejected' })
            ).toBeInTheDocument();
            await screen.findByText('user_one');
        });

        it('displays Review button for pending appeals', async () => {
            render(<AppealsQueue />);

            await waitFor(() => {
                expect(
                    screen.getAllByRole('button', { name: 'Review' })
                ).toHaveLength(2);
            });
        });

        it('displays empty state when no appeals', async () => {
            vi.mocked(moderationApi.getAdminAppeals).mockResolvedValue({
                success: true,
                data: [],
            });

            render(<AppealsQueue />);

            await waitFor(() => {
                expect(
                    screen.getByText('No pending appeals found')
                ).toBeInTheDocument();
            });
        });
    });

    describe('Status Filtering', () => {
        it('loads pending appeals by default', async () => {
            render(<AppealsQueue />);

            await waitFor(() => {
                expect(moderationApi.getAdminAppeals).toHaveBeenCalledWith(
                    'pending'
                );
            });
        });

        it('uses initialStatus prop when provided', async () => {
            render(<AppealsQueue initialStatus='approved' />);

            await waitFor(() => {
                expect(moderationApi.getAdminAppeals).toHaveBeenCalledWith(
                    'approved'
                );
            });
        });

        it('filters by approved status when button clicked', async () => {
            const user = userEvent.setup();
            render(<AppealsQueue />);

            await waitFor(() => {
                expect(screen.getByText('user_one')).toBeInTheDocument();
            });

            const approvedButton = screen.getByRole('button', {
                name: 'Approved',
            });
            await user.click(approvedButton);

            await waitFor(() => {
                expect(moderationApi.getAdminAppeals).toHaveBeenCalledWith(
                    'approved'
                );
            });
        });

        it('filters by rejected status when button clicked', async () => {
            const user = userEvent.setup();
            render(<AppealsQueue />);

            await waitFor(() => {
                expect(screen.getByText('user_one')).toBeInTheDocument();
            });

            const rejectedButton = screen.getByRole('button', {
                name: 'Rejected',
            });
            await user.click(rejectedButton);

            await waitFor(() => {
                expect(moderationApi.getAdminAppeals).toHaveBeenCalledWith(
                    'rejected'
                );
            });
        });

        it('highlights active filter button', async () => {
            render(<AppealsQueue />);

            await waitFor(() => {
                const pendingButton = screen.getByRole('button', {
                    name: 'Pending',
                });
                // Primary variant is applied to active button
                expect(pendingButton.className).toContain('primary');
            });
        });
    });

    describe('Appeal Details', () => {
        it('displays appeal information', async () => {
            render(<AppealsQueue />);

            await waitFor(() => {
                expect(screen.getByText('user_one')).toBeInTheDocument();
            });

            expect(
                screen.getByText('appeal-1', { exact: false })
            ).toBeInTheDocument();
            expect(
                screen.getByText('I did not violate any rules')
            ).toBeInTheDocument();
            expect(
                screen.getByText('ban', { exact: false })
            ).toBeInTheDocument();
            expect(
                screen.getByText('Spam', { exact: false })
            ).toBeInTheDocument();
        });

        it('formats dates correctly', async () => {
            render(<AppealsQueue />);

            await waitFor(() => {
                expect(screen.getByText('user_one')).toBeInTheDocument();
            });

            // Date format may vary by locale, just check it's displayed
            expect(screen.getAllByText(/Submitted:/)).toHaveLength(2);
        });

        it('displays status badges', async () => {
            render(<AppealsQueue />);

            await waitFor(() => {
                expect(screen.getAllByText('Pending')).toHaveLength(3); // Button + 2 badges
            });
        });

        it('displays resolution when present', async () => {
            const resolvedAppeals = [
                {
                    ...mockAppeals[0],
                    status: 'approved' as const,
                    resolution: 'Appeal was valid, ban removed',
                    resolved_at: '2024-01-03T00:00:00Z',
                },
            ];

            vi.mocked(moderationApi.getAdminAppeals).mockResolvedValue({
                success: true,
                data: resolvedAppeals,
            });

            render(<AppealsQueue />);

            await waitFor(() => {
                expect(
                    screen.getByText('Appeal was valid, ban removed')
                ).toBeInTheDocument();
                expect(screen.getByText(/Resolved:/)).toBeInTheDocument();
            });
        });
    });

    describe('Review Modal', () => {
        it('opens resolution modal when Review button is clicked', async () => {
            const user = userEvent.setup();
            render(<AppealsQueue />);

            await waitFor(() => {
                expect(
                    screen.getAllByRole('button', { name: 'Review' })
                ).toHaveLength(2);
            });

            const reviewButtons = screen.getAllByRole('button', {
                name: 'Review',
            });
            await user.click(reviewButtons[0]);

            await waitFor(() => {
                expect(
                    screen.getByTestId('resolution-modal')
                ).toBeInTheDocument();
                expect(
                    screen.getByText('Resolve Appeal for user_one')
                ).toBeInTheDocument();
            });
        });

        it('closes modal and reloads appeals on success', async () => {
            const user = userEvent.setup();
            render(<AppealsQueue />);

            await waitFor(() => {
                expect(
                    screen.getAllByRole('button', { name: 'Review' })
                ).toHaveLength(2);
            });

            const reviewButtons = screen.getAllByRole('button', {
                name: 'Review',
            });
            await user.click(reviewButtons[0]);

            await waitFor(() => {
                expect(
                    screen.getByTestId('resolution-modal')
                ).toBeInTheDocument();
            });

            vi.clearAllMocks();
            const successButton = screen.getByRole('button', {
                name: 'Resolve Success',
            });
            await user.click(successButton);

            await waitFor(() => {
                expect(moderationApi.getAdminAppeals).toHaveBeenCalled();
            });
        });

        it('closes modal when close button is clicked', async () => {
            const user = userEvent.setup();
            render(<AppealsQueue />);

            await waitFor(() => {
                expect(
                    screen.getAllByRole('button', { name: 'Review' })
                ).toHaveLength(2);
            });

            const reviewButtons = screen.getAllByRole('button', {
                name: 'Review',
            });
            await user.click(reviewButtons[0]);

            await waitFor(() => {
                expect(
                    screen.getByTestId('resolution-modal')
                ).toBeInTheDocument();
            });

            const closeButton = screen.getByRole('button', {
                name: 'Close Modal',
            });
            await user.click(closeButton);

            await waitFor(() => {
                expect(
                    screen.queryByTestId('resolution-modal')
                ).not.toBeInTheDocument();
            });
        });
    });

    describe('Error Handling', () => {
        it('displays error message when loading fails', async () => {
            vi.mocked(moderationApi.getAdminAppeals).mockRejectedValue(
                new Error('Failed to load appeals')
            );

            render(<AppealsQueue />);

            await waitFor(() => {
                expect(
                    screen.getByText(/Failed to load appeals/)
                ).toBeInTheDocument();
            });
        });

        it('allows retrying after error by changing filter', async () => {
            const user = userEvent.setup();
            vi.mocked(moderationApi.getAdminAppeals)
                .mockRejectedValueOnce(new Error('Failed to load'))
                .mockResolvedValueOnce({ success: true, data: mockAppeals });

            render(<AppealsQueue />);

            await waitFor(() => {
                expect(
                    screen.getByText(/Failed to load/)
                ).toBeInTheDocument();
            });

            const approvedButton = screen.getByRole('button', {
                name: 'Approved',
            });
            await user.click(approvedButton);

            await waitFor(() => {
                expect(screen.getByText('user_one')).toBeInTheDocument();
            });
        });
    });

    describe('Accessibility', () => {
        it('has proper heading structure', async () => {
            render(<AppealsQueue />);

            expect(
                screen.getByRole('heading', { name: 'Appeals Queue' })
            ).toBeInTheDocument();
            await screen.findByText('user_one');
        });

        it('displays appeal information in accessible format', async () => {
            render(<AppealsQueue />);

            await waitFor(() => {
                expect(screen.getByText('user_one')).toBeInTheDocument();
            });

            // Check for semantic structure
            const headings = screen.getAllByRole('heading', { level: 3 });
            expect(headings.length).toBeGreaterThan(0);
        });
    });
});
