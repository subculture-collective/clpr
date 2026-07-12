import { render, screen } from '@/test/test-utils';
import type { KarmaBreakdown } from '@/types/reputation';
import { describe, expect, it } from 'vitest';
import { KarmaBreakdownChart, KarmaStats } from './KarmaBreakdown';

const mockBreakdown: KarmaBreakdown = {
    clip_karma: 500,
    comment_karma: 734,
    total_karma: 1234,
};

const mockZeroBreakdown: KarmaBreakdown = {
    clip_karma: 0,
    comment_karma: 0,
    total_karma: 0,
};

describe('KarmaBreakdownChart', () => {
    it('displays total karma correctly', () => {
        render(<KarmaBreakdownChart breakdown={mockBreakdown} />);

        expect(screen.getByText('1,234')).toBeInTheDocument();
        expect(screen.getByText('Total Karma')).toBeInTheDocument();
    });

    it('displays clip karma amount', () => {
        render(<KarmaBreakdownChart breakdown={mockBreakdown} />);

        expect(screen.getByText('500')).toBeInTheDocument();
        expect(screen.getByText('Clip Karma')).toBeInTheDocument();
    });

    it('displays comment karma amount', () => {
        render(<KarmaBreakdownChart breakdown={mockBreakdown} />);

        expect(screen.getByText('734')).toBeInTheDocument();
        expect(screen.getByText('Comment Karma')).toBeInTheDocument();
    });

    it('calculates percentages correctly', () => {
        render(<KarmaBreakdownChart breakdown={mockBreakdown} />);

        // 500/1234 = 40.5%
        expect(screen.getByText('40.5%')).toBeInTheDocument();
        // 734/1234 = 59.5%
        expect(screen.getByText('59.5%')).toBeInTheDocument();
    });

    it('handles zero karma gracefully', () => {
        render(<KarmaBreakdownChart breakdown={mockZeroBreakdown} />);

        // Multiple zeros are rendered (total, clip, comment); ensure at least one is present
        expect(screen.getAllByText('0').length).toBeGreaterThanOrEqual(1);
        // Should show 0.0% for both
        expect(screen.getAllByText('0.0%').length).toBe(2);
    });

    it('displays karma breakdown title', () => {
        render(<KarmaBreakdownChart breakdown={mockBreakdown} />);

        expect(screen.getByText('Karma Breakdown')).toBeInTheDocument();
    });

    it('shows percentage labels', () => {
        render(<KarmaBreakdownChart breakdown={mockBreakdown} />);

        expect(screen.getByText('from clips')).toBeInTheDocument();
        expect(screen.getByText('from comments')).toBeInTheDocument();
    });
});

describe('KarmaStats', () => {
    it('displays all karma stats', () => {
        render(<KarmaStats breakdown={mockBreakdown} />);

        expect(screen.getByText('1,234')).toBeInTheDocument();
        expect(screen.getByText('500')).toBeInTheDocument();
        expect(screen.getByText('734')).toBeInTheDocument();
    });

    it('displays stat labels', () => {
        render(<KarmaStats breakdown={mockBreakdown} />);

        expect(screen.getByText('Total')).toBeInTheDocument();
        expect(screen.getByText('Clips')).toBeInTheDocument();
        expect(screen.getByText('Comments')).toBeInTheDocument();
    });

    it('formats large numbers with commas', () => {
        const largeBreakdown: KarmaBreakdown = {
            clip_karma: 50000,
            comment_karma: 73400,
            total_karma: 123400,
        };

        render(<KarmaStats breakdown={largeBreakdown} />);

        expect(screen.getByText('123,400')).toBeInTheDocument();
        expect(screen.getByText('50,000')).toBeInTheDocument();
        expect(screen.getByText('73,400')).toBeInTheDocument();
    });

    it('renders three stat cards', () => {
        const { container } = render(<KarmaStats breakdown={mockBreakdown} />);

        const statCards = container.querySelectorAll('.bg-surface-raised');
        expect(statCards.length).toBe(3);
    });
});
