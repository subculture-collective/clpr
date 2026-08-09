import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ModerationAnalyticsDashboard } from './ModerationAnalyticsDashboard';
import * as moderationApi from '../../lib/moderation-api';
import { allowTestConsole } from '../../test/setup';

// Mock the API module
vi.mock('../../lib/moderation-api');

// Mock chart components to simplify testing
vi.mock('../analytics/MetricCard', () => ({
    default: ({ title, value }: { title: string; value: string | number }) => (
        <div data-testid='metric-card'>
            <div>{title}</div>
            <div>{value}</div>
        </div>
    ),
}));

vi.mock('../analytics/PieChartComponent', () => ({
    default: ({ data }: { data?: Array<{ name: string; value: number }> }) => (
        <div data-testid='pie-chart'>
            {(data || []).map((d: { name: string; value: number }) => (
                <div key={d.name}>
                    {d.name}: {d.value}
                </div>
            ))}
        </div>
    ),
}));

vi.mock('../analytics/BarChartComponent', () => ({
    default: ({ data }: { data?: Array<{ name: string; value: number }> }) => (
        <div data-testid='bar-chart'>
            {(data || []).map((d: { name: string; value: number }) => (
                <div key={d.name}>
                    {d.name}: {d.value}
                </div>
            ))}
        </div>
    ),
}));

vi.mock('../analytics/LineChartComponent', () => ({
    default: ({ data }: { data?: unknown[] }) => (
        <div data-testid='line-chart'>{(data || []).length} data points</div>
    ),
}));

vi.mock('../analytics/DateRangeSelector', () => ({
    default: ({
        value,
        onChange,
    }: {
        value: number;
        onChange: (days: number) => void;
    }) => (
        <select
            data-testid='date-range-selector'
            value={value}
            onChange={e => onChange(Number(e.target.value))}
        >
            <option value={7}>Last 7 days</option>
            <option value={30}>Last 30 days</option>
            <option value={90}>Last 90 days</option>
        </select>
    ),
}));

describe('ModerationAnalyticsDashboard', () => {
    const mockAnalytics: moderationApi.ModerationAnalytics = {
        total_actions: 150,
        actions_by_type: {
            approve: 80,
            reject: 50,
            ban: 15,
            timeout: 5,
        },
        actions_by_moderator: {
            mod_alice: 60,
            mod_bob: 50,
            mod_charlie: 40,
        },
        content_type_breakdown: {
            clips: 90,
            comments: 60,
        },
        average_response_time_minutes: 25,
        actions_over_time: [],
        ban_reasons: {
            spam: 30,
            harassment: 20,
            inappropriate: 10,
        },
        most_banned_users: [
            {
                user_id: 'user-1',
                username: 'banned_user_1',
                ban_count: 5,
                last_ban_at: '2024-01-15T10:00:00Z',
            },
            {
                user_id: 'user-2',
                username: 'banned_user_2',
                ban_count: 3,
                last_ban_at: '2024-01-14T10:00:00Z',
            },
        ],
        appeals: {
            total_appeals: 20,
            pending_appeals: 5,
            approved_appeals: 8,
            rejected_appeals: 7,
            false_positive_rate: 40.0,
        },
    };

    beforeEach(() => {
        vi.clearAllMocks();
        vi.mocked(moderationApi.getModerationAnalytics).mockResolvedValue({
            success: true,
            data: mockAnalytics,
        });
    });

    describe('Rendering', () => {
        it('renders the component with title', async () => {
            render(<ModerationAnalyticsDashboard />);

            await waitFor(() => {
                expect(
                    screen.getByText('Moderation Analytics'),
                ).toBeInTheDocument();
            });
        });

        it('displays loading state initially', () => {
            vi.mocked(moderationApi.getModerationAnalytics).mockImplementation(
                () => new Promise(() => {}),
            );
            render(<ModerationAnalyticsDashboard />);

            expect(
                screen.getByText('Loading analytics...'),
            ).toBeInTheDocument();
        });

        it('loads analytics data on mount', async () => {
            render(<ModerationAnalyticsDashboard />);

            await waitFor(() => {
                expect(moderationApi.getModerationAnalytics).toHaveBeenCalled();
            });
        });

        it('renders date range selector', async () => {
            render(<ModerationAnalyticsDashboard />);

            await waitFor(() => {
                expect(
                    screen.getByTestId('date-range-selector'),
                ).toBeInTheDocument();
            });
        });
    });

    describe('Metric Cards', () => {
        it('displays metric cards with analytics data', async () => {
            render(<ModerationAnalyticsDashboard />);

            await waitFor(() => {
                const metricCards = screen.getAllByTestId('metric-card');
                expect(metricCards.length).toBeGreaterThan(0);
            });
        });

        it('displays total actions count', async () => {
            render(<ModerationAnalyticsDashboard />);

            await waitFor(() => {
                expect(screen.getByText('150')).toBeInTheDocument();
            });
        });

        it('displays average response time', async () => {
            render(<ModerationAnalyticsDashboard />);

            await waitFor(() => {
                expect(screen.getByText('25 min')).toBeInTheDocument();
            });
        });

        it('displays active moderators count', async () => {
            render(<ModerationAnalyticsDashboard />);

            await waitFor(() => {
                // 3 moderators in actions_by_moderator - may appear multiple times
                const threes = screen.getAllByText('3');
                expect(threes.length).toBeGreaterThan(0);
            });
        });

        it('handles missing response time gracefully', async () => {
            vi.mocked(moderationApi.getModerationAnalytics).mockResolvedValue({
                success: true,
                data: {
                    ...mockAnalytics,
                    average_response_time_minutes: undefined,
                },
            });

            render(<ModerationAnalyticsDashboard />);

            await waitFor(() => {
                expect(screen.getByText('N/A')).toBeInTheDocument();
            });
        });
    });

    describe('Charts', () => {
        it('renders pie chart for action types', async () => {
            render(<ModerationAnalyticsDashboard />);

            await waitFor(() => {
                const pieCharts = screen.getAllByTestId('pie-chart');
                expect(pieCharts.length).toBeGreaterThan(0);
            });
        });

        it('displays action type data in chart', async () => {
            render(<ModerationAnalyticsDashboard />);

            await waitFor(() => {
                expect(screen.getByText('approve: 80')).toBeInTheDocument();
                expect(screen.getByText('reject: 50')).toBeInTheDocument();
                expect(screen.getByText('ban: 15')).toBeInTheDocument();
            });
        });

        it('renders bar chart for moderator activity', async () => {
            render(<ModerationAnalyticsDashboard />);

            await waitFor(() => {
                const barCharts = screen.getAllByTestId('bar-chart');
                expect(barCharts.length).toBeGreaterThan(0);
            });
        });

        it('displays top moderators in chart', async () => {
            render(<ModerationAnalyticsDashboard />);

            await waitFor(() => {
                expect(screen.getByText('mod_alice: 60')).toBeInTheDocument();
                expect(screen.getByText('mod_bob: 50')).toBeInTheDocument();
                expect(screen.getByText('mod_charlie: 40')).toBeInTheDocument();
            });
        });

        it('limits moderator chart to top 10', async () => {
            const manyModerators: Record<string, number> = {};
            for (let i = 0; i < 15; i++) {
                manyModerators[`mod_${i}`] = i * 10;
            }

            vi.mocked(moderationApi.getModerationAnalytics).mockResolvedValue({
                success: true,
                data: {
                    ...mockAnalytics,
                    actions_by_moderator: manyModerators,
                },
            });

            render(<ModerationAnalyticsDashboard />);

            await waitFor(() => {
                const barCharts = screen.getAllByTestId('bar-chart');
                // Should only show 10 items in the first bar chart (moderator activity)
                const items = barCharts[0].querySelectorAll('div');
                expect(items.length).toBeLessThanOrEqual(10);
            });
        });

        it('renders chart for content type breakdown', async () => {
            render(<ModerationAnalyticsDashboard />);

            await waitFor(() => {
                expect(screen.getByText('clips: 90')).toBeInTheDocument();
                expect(screen.getByText('comments: 60')).toBeInTheDocument();
            });
        });
    });

    describe('Date Range Selection', () => {
        it('defaults to 30 days', async () => {
            render(<ModerationAnalyticsDashboard />);

            await waitFor(() => {
                const selector = screen.getByTestId(
                    'date-range-selector',
                ) as HTMLSelectElement;
                expect(selector.value).toBe('30');
            });
        });

        it('updates analytics when date range changes', async () => {
            const user = userEvent.setup();
            render(<ModerationAnalyticsDashboard />);

            await waitFor(() => {
                expect(
                    screen.getByTestId('date-range-selector'),
                ).toBeInTheDocument();
            });

            vi.clearAllMocks();
            const selector = screen.getByTestId('date-range-selector');
            await user.selectOptions(selector, '7');

            await waitFor(() => {
                expect(moderationApi.getModerationAnalytics).toHaveBeenCalled();
            });
        });

        it('passes correct date range to API', async () => {
            const user = userEvent.setup();
            render(<ModerationAnalyticsDashboard />);

            await waitFor(() => {
                expect(
                    screen.getByTestId('date-range-selector'),
                ).toBeInTheDocument();
            });

            vi.clearAllMocks();
            const selector = screen.getByTestId('date-range-selector');
            await user.selectOptions(selector, '90');

            await waitFor(() => {
                const call = vi.mocked(moderationApi.getModerationAnalytics)
                    .mock.calls[0][0];
                expect(call).toHaveProperty('start_date');
                expect(call).toHaveProperty('end_date');
            });
        });
    });

    describe('Error Handling', () => {
        it('displays error message when loading fails', async () => {
            allowTestConsole('error', /Analytics error:.*Failed to load analytics/);
            vi.mocked(moderationApi.getModerationAnalytics).mockRejectedValue(
                new Error('Failed to load analytics'),
            );

            render(<ModerationAnalyticsDashboard />);

            await waitFor(() => {
                expect(
                    screen.getByText('Failed to load analytics data'),
                ).toBeInTheDocument();
            });
        });

        it('shows retry button on error', async () => {
            allowTestConsole('error', /Analytics error:.*Failed to load analytics/);
            vi.mocked(moderationApi.getModerationAnalytics).mockRejectedValue(
                new Error('Failed to load analytics'),
            );

            render(<ModerationAnalyticsDashboard />);

            await waitFor(() => {
                expect(
                    screen.getByRole('button', { name: /retry/i }),
                ).toBeInTheDocument();
            });
        });

        it('retries loading when retry button is clicked', async () => {
            allowTestConsole('error', /Analytics error:.*Failed/);
            const user = userEvent.setup();
            vi.mocked(moderationApi.getModerationAnalytics)
                .mockRejectedValueOnce(new Error('Failed'))
                .mockResolvedValueOnce({ success: true, data: mockAnalytics });

            render(<ModerationAnalyticsDashboard />);

            await waitFor(() => {
                expect(
                    screen.getByText('Failed to load analytics data'),
                ).toBeInTheDocument();
            });

            const retryButton = screen.getByRole('button', { name: /retry/i });
            await user.click(retryButton);

            await waitFor(() => {
                expect(
                    screen.getByText('Moderation Analytics'),
                ).toBeInTheDocument();
                expect(
                    screen.queryByText('Failed to load analytics data'),
                ).not.toBeInTheDocument();
            });
        });
    });

    describe('Data Transformation', () => {
        it('transforms actions_by_type into chart data', async () => {
            render(<ModerationAnalyticsDashboard />);

            await waitFor(() => {
                const pieCharts = screen.getAllByTestId('pie-chart');
                expect(pieCharts.length).toBeGreaterThan(0);
            });

            // Verify all action types are present
            expect(screen.getByText('approve: 80')).toBeInTheDocument();
            expect(screen.getByText('reject: 50')).toBeInTheDocument();
            expect(screen.getByText('ban: 15')).toBeInTheDocument();
            expect(screen.getByText('timeout: 5')).toBeInTheDocument();
        });

        it('transforms content_type_breakdown into chart data', async () => {
            render(<ModerationAnalyticsDashboard />);

            await waitFor(() => {
                expect(screen.getByText('clips: 90')).toBeInTheDocument();
                expect(screen.getByText('comments: 60')).toBeInTheDocument();
            });
        });
    });

    describe('Accessibility', () => {
        it('has proper heading structure', async () => {
            render(<ModerationAnalyticsDashboard />);

            await waitFor(() => {
                expect(
                    screen.getByRole('heading', {
                        name: 'Moderation Analytics',
                    }),
                ).toBeInTheDocument();
            });
        });

        it('displays error in accessible alert', async () => {
            allowTestConsole('error', /Analytics error:.*Failed/);
            vi.mocked(moderationApi.getModerationAnalytics).mockRejectedValue(
                new Error('Failed'),
            );

            render(<ModerationAnalyticsDashboard />);

            await waitFor(() => {
                const errorContainer = screen
                    .getByText('Failed to load analytics data')
                    .closest('div');
                expect(errorContainer).toHaveClass('border-red-200');
            });
        });
    });

    describe('New Analytics Features', () => {
        it('displays false positive rate metric', async () => {
            render(<ModerationAnalyticsDashboard />);

            await waitFor(() => {
                expect(
                    screen.getByText('False Positive Rate'),
                ).toBeInTheDocument();
                expect(screen.getByText('40.0%')).toBeInTheDocument();
            });
        });

        it('renders ban reasons distribution chart when data available', async () => {
            render(<ModerationAnalyticsDashboard />);

            await waitFor(() => {
                expect(screen.getByText('spam: 30')).toBeInTheDocument();
                expect(screen.getByText('harassment: 20')).toBeInTheDocument();
                expect(
                    screen.getByText('inappropriate: 10'),
                ).toBeInTheDocument();
            });
        });

        it('does not render ban reasons chart when no data', async () => {
            vi.mocked(moderationApi.getModerationAnalytics).mockResolvedValue({
                success: true,
                data: {
                    ...mockAnalytics,
                    ban_reasons: {},
                },
            });

            render(<ModerationAnalyticsDashboard />);

            await waitFor(() => {
                expect(
                    screen.queryByText('Ban Reasons Distribution'),
                ).not.toBeInTheDocument();
            });
        });

        it('displays most banned users table', async () => {
            render(<ModerationAnalyticsDashboard />);

            await waitFor(() => {
                expect(
                    screen.getByText('Most Banned Users'),
                ).toBeInTheDocument();
                expect(screen.getByText('banned_user_1')).toBeInTheDocument();
                expect(screen.getByText('banned_user_2')).toBeInTheDocument();
                // Numbers 5 and 3 may appear in multiple metric cards
                expect(screen.getAllByText('5').length).toBeGreaterThan(0);
                expect(screen.getAllByText('3').length).toBeGreaterThan(0);
            });
        });

        it('does not render banned users table when no data', async () => {
            vi.mocked(moderationApi.getModerationAnalytics).mockResolvedValue({
                success: true,
                data: {
                    ...mockAnalytics,
                    most_banned_users: [],
                },
            });

            render(<ModerationAnalyticsDashboard />);

            await waitFor(() => {
                expect(
                    screen.queryByText('Most Banned Users'),
                ).not.toBeInTheDocument();
            });
        });

        it('displays appeals statistics section', async () => {
            render(<ModerationAnalyticsDashboard />);

            await waitFor(() => {
                expect(
                    screen.getByText('Appeals & Reversals'),
                ).toBeInTheDocument();
                expect(screen.getByText('Total Appeals')).toBeInTheDocument();
                // Numbers may appear in multiple places (metric cards, charts, etc.)
                expect(screen.getAllByText('20').length).toBeGreaterThan(0);
                expect(screen.getByText('Pending')).toBeInTheDocument();
                expect(screen.getAllByText('5').length).toBeGreaterThan(0);
                expect(screen.getByText('Approved')).toBeInTheDocument();
                expect(screen.getAllByText('8').length).toBeGreaterThan(0);
                expect(screen.getByText('Rejected')).toBeInTheDocument();
                expect(screen.getAllByText('7').length).toBeGreaterThan(0);
            });
        });

        it('does not render appeals section when no appeals data', async () => {
            vi.mocked(moderationApi.getModerationAnalytics).mockResolvedValue({
                success: true,
                data: {
                    ...mockAnalytics,
                    appeals: undefined,
                },
            });

            render(<ModerationAnalyticsDashboard />);

            await waitFor(() => {
                expect(
                    screen.queryByText('Appeals & Reversals'),
                ).not.toBeInTheDocument();
            });
        });

        it('handles missing false positive rate gracefully', async () => {
            vi.mocked(moderationApi.getModerationAnalytics).mockResolvedValue({
                success: true,
                data: {
                    ...mockAnalytics,
                    appeals: undefined,
                },
            });

            render(<ModerationAnalyticsDashboard />);

            await waitFor(() => {
                expect(
                    screen.getByText('False Positive Rate'),
                ).toBeInTheDocument();
                expect(screen.getByText('N/A')).toBeInTheDocument();
            });
        });
    });
});
