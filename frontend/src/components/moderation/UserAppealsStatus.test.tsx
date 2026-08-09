import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { UserAppealsStatus } from './UserAppealsStatus';
import * as moderationApi from '../../lib/moderation-api';

// Mock the API module
vi.mock('../../lib/moderation-api');

describe('UserAppealsStatus', () => {
    const mockAppeals: moderationApi.ModerationAppeal[] = [
        {
            id: 'appeal-1',
            moderation_action_id: 'action-1',
            user_id: 'user-1',
            username: 'testuser',
            reason: 'I believe this ban was unfair',
            status: 'pending',
            created_at: '2024-01-01T00:00:00Z',
            content_type: 'clip',
            content_id: 'clip-1',
            decision_action: 'ban',
            decision_reason: 'Spam violations',
        },
        {
            id: 'appeal-2',
            moderation_action_id: 'action-2',
            user_id: 'user-1',
            username: 'testuser',
            reason: 'This was a mistake',
            status: 'approved',
            created_at: '2024-01-02T00:00:00Z',
            resolved_at: '2024-01-03T00:00:00Z',
            resolution: 'You were correct, ban has been lifted',
            content_type: 'comment',
            content_id: 'comment-1',
            decision_action: 'remove',
            decision_reason: 'Inappropriate content',
        },
        {
            id: 'appeal-3',
            moderation_action_id: 'action-3',
            user_id: 'user-1',
            username: 'testuser',
            reason: 'Please reconsider',
            status: 'rejected',
            created_at: '2024-01-04T00:00:00Z',
            resolved_at: '2024-01-05T00:00:00Z',
            resolution: 'The original decision was correct',
            content_type: 'clip',
            content_id: 'clip-2',
            decision_action: 'timeout',
            decision_reason: 'Harassment',
        },
    ];

    beforeEach(() => {
        vi.clearAllMocks();
        vi.mocked(moderationApi.getUserAppeals).mockResolvedValue({
            success: true,
            data: mockAppeals,
        });
    });

    describe('Rendering', () => {
        it('renders the component with title', async () => {
            render(<UserAppealsStatus />);

            await waitFor(() => {
                expect(screen.getByText('My Appeals')).toBeInTheDocument();
            });
        });

        it('displays loading state initially', () => {
            vi.mocked(moderationApi.getUserAppeals).mockImplementation(
                () => new Promise(() => {}),
            );
            render(<UserAppealsStatus />);

            // The spinner div exists with the animate-spin class
            const spinners = screen.getAllByRole('generic', { hidden: true });
            const spinner = spinners.find(el =>
                el.className.includes('animate-spin'),
            );
            expect(spinner).toBeTruthy();
        });

        it('displays appeals after loading', async () => {
            render(<UserAppealsStatus />);

            await waitFor(() => {
                expect(
                    screen.getByText(/I believe this ban was unfair/),
                ).toBeInTheDocument();
            });

            expect(screen.getByText(/This was a mistake/)).toBeInTheDocument();
            expect(screen.getByText(/Please reconsider/)).toBeInTheDocument();
        });

        it('displays empty state when no appeals', async () => {
            vi.mocked(moderationApi.getUserAppeals).mockResolvedValue({
                success: true,
                data: [],
            });

            render(<UserAppealsStatus />);

            await waitFor(() => {
                expect(
                    screen.getByText(/You haven't submitted any appeals yet/),
                ).toBeInTheDocument();
            });
        });
    });

    describe('Appeal Details', () => {
        it('displays appeal information', async () => {
            render(<UserAppealsStatus />);

            await waitFor(() => {
                expect(
                    screen.getByText(/Appeal for ban Action/),
                ).toBeInTheDocument();
            });

            expect(screen.getByText('Spam violations')).toBeInTheDocument();
            expect(
                screen.getByText(/Content:\s*clip\s*\(ID:\s*clip-1\)/i),
            ).toBeInTheDocument();
        });

        it('formats dates correctly', async () => {
            render(<UserAppealsStatus />);

            await waitFor(() => {
                expect(screen.getAllByText(/Submitted:/)).toHaveLength(3);
            });
        });

        it('displays status badges with correct styling', async () => {
            render(<UserAppealsStatus />);

            await waitFor(() => {
                expect(screen.getByText('Pending')).toBeInTheDocument();
            });

            const pendingBadge = screen.getByText('Pending');
            expect(pendingBadge.className).toContain('bg-yellow-100');

            const approvedBadge = screen.getByText('Approved');
            expect(approvedBadge.className).toContain('bg-green-100');

            const rejectedBadge = screen.getByText('Rejected');
            expect(rejectedBadge.className).toContain('bg-red-100');
        });
    });

    describe('Pending Appeals', () => {
        it('shows pending status message for pending appeals', async () => {
            render(<UserAppealsStatus />);

            await waitFor(() => {
                expect(
                    screen.getByText(
                        /Your appeal is being reviewed by our moderation team/,
                    ),
                ).toBeInTheDocument();
            });
        });

        it('does not show resolution for pending appeals', async () => {
            render(<UserAppealsStatus />);

            await waitFor(() => {
                expect(screen.getByText('Pending')).toBeInTheDocument();
            });

            // The pending appeal should not have a "Moderator's Note" section
            const moderatorNotes = screen.queryAllByText("Moderator's Note:");
            expect(moderatorNotes.length).toBe(2); // Only for approved and rejected
        });
    });

    describe('Approved Appeals', () => {
        it('shows approved status message', async () => {
            render(<UserAppealsStatus />);

            await waitFor(() => {
                expect(
                    screen.getByText(/Your appeal has been approved/),
                ).toBeInTheDocument();
            });
        });

        it('displays resolution for approved appeals', async () => {
            render(<UserAppealsStatus />);

            await waitFor(() => {
                expect(
                    screen.getByText('You were correct, ban has been lifted'),
                ).toBeInTheDocument();
            });
        });

        it('displays resolved date for approved appeals', async () => {
            render(<UserAppealsStatus />);

            await waitFor(() => {
                const resolvedTexts = screen.getAllByText(/Resolved:/);
                expect(resolvedTexts.length).toBe(2); // For approved and rejected
            });
        });
    });

    describe('Rejected Appeals', () => {
        it('shows rejected status message', async () => {
            render(<UserAppealsStatus />);

            await waitFor(() => {
                expect(
                    screen.getByText(/Your appeal has been rejected/),
                ).toBeInTheDocument();
            });
        });

        it('displays resolution for rejected appeals', async () => {
            render(<UserAppealsStatus />);

            await waitFor(() => {
                expect(
                    screen.getByText('The original decision was correct'),
                ).toBeInTheDocument();
            });
        });

        it('displays resolved date for rejected appeals', async () => {
            render(<UserAppealsStatus />);

            await waitFor(() => {
                const resolvedTexts = screen.getAllByText(/Resolved:/);
                expect(resolvedTexts.length).toBe(2); // For approved and rejected
            });
        });
    });

    describe('Error Handling', () => {
        it('displays error message when loading fails', async () => {
            vi.mocked(moderationApi.getUserAppeals).mockRejectedValue(
                new Error('Failed to load appeals'),
            );

            render(<UserAppealsStatus />);

            await waitFor(() => {
                expect(
                    screen.getByText(/Failed to load appeals/),
                ).toBeInTheDocument();
            });
        });

        it('displays error in alert component', async () => {
            vi.mocked(moderationApi.getUserAppeals).mockRejectedValue(
                new Error('Network error'),
            );

            render(<UserAppealsStatus />);

            await waitFor(() => {
                const alert = screen.getByRole('alert');
                expect(alert).toBeInTheDocument();
                expect(alert).toHaveTextContent(/network error/i);
            });
        });
    });

    describe('Missing Data Handling', () => {
        it('handles missing decision_action gracefully', async () => {
            const appealsWithMissingData = [
                {
                    ...mockAppeals[0],
                    decision_action: undefined,
                },
            ];

            vi.mocked(moderationApi.getUserAppeals).mockResolvedValue({
                success: true,
                data: appealsWithMissingData,
            });

            render(<UserAppealsStatus />);

            await waitFor(() => {
                expect(
                    screen.getByText(/Appeal for Unknown Action/),
                ).toBeInTheDocument();
            });
        });

        it('handles missing decision_reason gracefully', async () => {
            const appealsWithMissingData = [
                {
                    ...mockAppeals[0],
                    decision_reason: undefined,
                },
            ];

            vi.mocked(moderationApi.getUserAppeals).mockResolvedValue({
                success: true,
                data: appealsWithMissingData,
            });

            render(<UserAppealsStatus />);

            await waitFor(() => {
                expect(
                    screen.getByText('No reason provided'),
                ).toBeInTheDocument();
            });
        });

        it('handles missing content_type and content_id gracefully', async () => {
            const appealsWithMissingData = [
                {
                    ...mockAppeals[0],
                    content_type: undefined,
                    content_id: undefined,
                },
            ];

            vi.mocked(moderationApi.getUserAppeals).mockResolvedValue({
                success: true,
                data: appealsWithMissingData,
            });

            render(<UserAppealsStatus />);

            await waitFor(() => {
                expect(
                    screen.getByText(/Content: Unknown/),
                ).toBeInTheDocument();
                expect(screen.getByText(/N\/A/)).toBeInTheDocument();
            });
        });

        it('does not display resolution section if not present', async () => {
            const appealsWithoutResolution = [
                {
                    ...mockAppeals[1],
                    resolution: undefined,
                },
            ];

            vi.mocked(moderationApi.getUserAppeals).mockResolvedValue({
                success: true,
                data: appealsWithoutResolution,
            });

            render(<UserAppealsStatus />);

            await waitFor(() => {
                expect(screen.getByText('Approved')).toBeInTheDocument();
            });

            // Should not show "Moderator's Note" when no resolution
            expect(
                screen.queryByText("Moderator's Note:"),
            ).not.toBeInTheDocument();
        });
    });

    describe('Accessibility', () => {
        it('has proper heading structure', async () => {
            render(<UserAppealsStatus />);

            await waitFor(() => {
                expect(
                    screen.getByRole('heading', { name: 'My Appeals' }),
                ).toBeInTheDocument();
            });
        });

        it('has proper heading hierarchy for appeals', async () => {
            render(<UserAppealsStatus />);

            await waitFor(() => {
                const headings = screen.getAllByRole('heading', { level: 3 });
                expect(headings.length).toBe(3); // One for each appeal
            });
        });

        it('uses semantic HTML for status messages', async () => {
            render(<UserAppealsStatus />);

            await waitFor(() => {
                const alerts = screen.getAllByRole('alert');
                expect(alerts.length).toBeGreaterThan(0);
            });
        });

        it('displays spinner with proper role during loading', () => {
            vi.mocked(moderationApi.getUserAppeals).mockImplementation(
                () => new Promise(() => {}),
            );
            render(<UserAppealsStatus />);

            const spinners = screen.getAllByRole('generic', { hidden: true });
            expect(
                spinners.some(el => el.className.includes('animate-spin')),
            ).toBe(true);
        });
    });
});
