import { describe, it, expect, vi, afterEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { FeedFilters } from './FeedFilters';
import type { SortOption } from '@/types/clip';

describe('FeedFilters', () => {
    const mockOnSortChange = vi.fn();
    const mockOnTimeframeChange = vi.fn();

    afterEach(() => {
        vi.clearAllMocks();
    });

    describe('Sort options rendering', () => {
        it('renders all sort options', () => {
            render(
                <FeedFilters
                    sort='trending'
                    onSortChange={mockOnSortChange}
                    onTimeframeChange={mockOnTimeframeChange}
                />
            );

            const select = screen.getByRole('combobox', { name: 'Sort by' });
            expect(select).toBeInTheDocument();

            expect(
                screen.getByRole('option', { name: 'Trending' })
            ).toBeInTheDocument();
            expect(
                screen.getByRole('option', { name: 'Most Popular' })
            ).toBeInTheDocument();
            expect(
                screen.getByRole('option', { name: 'Top Rated' })
            ).toBeInTheDocument();
            expect(
                screen.getByRole('option', { name: 'Rising' })
            ).toBeInTheDocument();
            expect(
                screen.getByRole('option', { name: 'Newest' })
            ).toBeInTheDocument();
            expect(
                screen.getByRole('option', { name: 'Most Discussed' })
            ).toBeInTheDocument();
        });

        it('shows selected sort value', () => {
            render(
                <FeedFilters
                    sort='trending'
                    onSortChange={mockOnSortChange}
                    onTimeframeChange={mockOnTimeframeChange}
                />
            );

            const select = screen.getByRole('combobox', { name: 'Sort by' });
            expect(select).toHaveValue('trending');
        });

        it('shows different selected sort', () => {
            render(
                <FeedFilters
                    sort='new'
                    onSortChange={mockOnSortChange}
                    onTimeframeChange={mockOnTimeframeChange}
                />
            );

            const select = screen.getByRole('combobox', { name: 'Sort by' });
            expect(select).toHaveValue('new');
        });
    });

    describe('Sort option interaction', () => {
        it('calls onSortChange when selecting a sort option', async () => {
            const user = userEvent.setup();
            render(
                <FeedFilters
                    sort='trending'
                    onSortChange={mockOnSortChange}
                    onTimeframeChange={mockOnTimeframeChange}
                />
            );

            const select = screen.getByRole('combobox', { name: 'Sort by' });
            await user.selectOptions(select, 'new');

            expect(mockOnSortChange).toHaveBeenCalledWith('new');
            expect(mockOnSortChange).toHaveBeenCalledTimes(1);
        });

        it('calls onSortChange with correct value for each sort option', async () => {
            const user = userEvent.setup();
            render(
                <FeedFilters
                    sort='trending'
                    onSortChange={mockOnSortChange}
                    onTimeframeChange={mockOnTimeframeChange}
                />
            );

            const sortOptions: SortOption[] = [
                'trending',
                'popular',
                'top',
                'rising',
                'new',
                'discussed',
            ];
            const select = screen.getByRole('combobox', { name: 'Sort by' });

            for (const option of sortOptions) {
                await user.selectOptions(select, option);
            }

            expect(mockOnSortChange).toHaveBeenCalledTimes(6);
            sortOptions.forEach((option, index) => {
                expect(mockOnSortChange).toHaveBeenNthCalledWith(
                    index + 1,
                    option
                );
            });
        });
    });

    describe('Timeframe selector', () => {
        it('does not show timeframe selector when sort is not "top" or "trending"', () => {
            render(
                <FeedFilters
                    sort='new'
                    onSortChange={mockOnSortChange}
                    onTimeframeChange={mockOnTimeframeChange}
                />
            );

            expect(screen.queryByText('Timeframe:')).not.toBeInTheDocument();
        });

        it('shows timeframe selector when sort is "top"', () => {
            render(
                <FeedFilters
                    sort='top'
                    timeframe='day'
                    onSortChange={mockOnSortChange}
                    onTimeframeChange={mockOnTimeframeChange}
                />
            );

            expect(screen.getByText('Timeframe:')).toBeInTheDocument();
            expect(
                screen.getByRole('button', { name: 'Past Hour' })
            ).toBeInTheDocument();
            expect(
                screen.getByRole('button', { name: 'Past Day' })
            ).toBeInTheDocument();
            expect(
                screen.getByRole('button', { name: 'Past Week' })
            ).toBeInTheDocument();
            expect(
                screen.getByRole('button', { name: 'Past Month' })
            ).toBeInTheDocument();
            expect(
                screen.getByRole('button', { name: 'Past Year' })
            ).toBeInTheDocument();
            expect(
                screen.getByRole('button', { name: 'All Time' })
            ).toBeInTheDocument();
        });

        it('applies primary variant to selected timeframe', () => {
            render(
                <FeedFilters
                    sort='top'
                    timeframe='week'
                    onSortChange={mockOnSortChange}
                    onTimeframeChange={mockOnTimeframeChange}
                />
            );

            const weekButton = screen.getByRole('button', {
                name: 'Past Week',
            });
            const dayButton = screen.getByRole('button', { name: 'Past Day' });

            // Selected timeframe should have primary variant
            expect(weekButton.className).toContain('bg-primary-500');
            expect(weekButton.className).toContain('text-white');

            // Unselected timeframe should have ghost variant
            expect(dayButton.className).toContain('bg-transparent');
        });

        it('calls onTimeframeChange when clicking a timeframe option', async () => {
            const user = userEvent.setup();
            render(
                <FeedFilters
                    sort='top'
                    timeframe='day'
                    onSortChange={mockOnSortChange}
                    onTimeframeChange={mockOnTimeframeChange}
                />
            );

            const weekButton = screen.getByRole('button', {
                name: 'Past Week',
            });
            await user.click(weekButton);

            expect(mockOnTimeframeChange).toHaveBeenCalledWith('week');
            expect(mockOnTimeframeChange).toHaveBeenCalledTimes(1);
        });
    });
});
