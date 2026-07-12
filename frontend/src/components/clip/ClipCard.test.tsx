import { render, screen } from '@/test/test-utils';
import type { Clip } from '@/types/clip';
import { describe, expect, it, vi } from 'vitest';
import { ClipCard } from './ClipCard';

// Mock the hooks
vi.mock('@/hooks/useClips', () => ({
    useClipVote: () => ({
        mutate: vi.fn(),
        isPending: false,
    }),
    useClipFavorite: () => ({
        mutate: vi.fn(),
        isPending: false,
    }),
}));

const mockClip: Clip = {
    id: '123e4567-e89b-12d3-a456-426614174000',
    twitch_clip_id: 'test_clip_1',
    twitch_clip_url: 'https://clips.twitch.tv/test_clip_1',
    embed_url: 'https://clips.twitch.tv/embed?clip=test_clip_1',
    title: 'Amazing Play',
    creator_name: 'TestCreator',
    broadcaster_name: 'TestStreamer',
    game_name: 'Test Game',
    thumbnail_url: 'https://example.com/thumb.jpg',
    duration: 30.5,
    view_count: 1500,
    vote_score: 42,
    comment_count: 10,
    favorite_count: 5,
    created_at: '2024-01-01T00:00:00Z',
    imported_at: '2024-01-01T00:00:00Z',
    is_featured: false,
    is_nsfw: false,
    is_removed: false,
    user_vote: null,
    is_favorited: false,
};

describe('ClipCard', () => {
    it('renders clip information correctly', () => {
        render(<ClipCard clip={mockClip} />);

        expect(screen.getByText('Amazing Play')).toBeInTheDocument();
        expect(screen.getByText('TestStreamer')).toBeInTheDocument();
        expect(screen.getByText('Test Game')).toBeInTheDocument();
        expect(screen.getByText('42')).toBeInTheDocument();
    });

    it('displays vote score with correct formatting', () => {
        render(<ClipCard clip={mockClip} />);

        const voteScore = screen.getByText('42');
        expect(voteScore).toBeInTheDocument();
        expect(voteScore).toHaveClass('text-xs', 'font-bold');
    });

    it('formats large numbers correctly', () => {
        const clipWithLargeViews = {
            ...mockClip,
            view_count: 1500000,
        };

        render(<ClipCard clip={clipWithLargeViews} />);
        expect(screen.getByText('1.5M')).toBeInTheDocument();
    });

    it('formats duration correctly', () => {
        render(<ClipCard clip={mockClip} />);
        // Duration is 30.5 seconds = 0:30
        expect(screen.getByText('0:30')).toBeInTheDocument();
    });

    it('has upvote and downvote buttons', () => {
        render(<ClipCard clip={mockClip} />);

        // When not authenticated, buttons show login prompt
        const upvoteButton = screen.getByLabelText(/upvote/i);
        const downvoteButton = screen.getByLabelText(/downvote/i);

        expect(upvoteButton).toBeInTheDocument();
        expect(downvoteButton).toBeInTheDocument();
    });

    it('displays positive vote score in green', () => {
        render(<ClipCard clip={mockClip} />);

        const voteScore = screen.getByText('42');
        expect(voteScore).toHaveClass('text-upvote');
    });

    it('displays negative vote score in red', () => {
        const clipWithNegativeScore = {
            ...mockClip,
            vote_score: -5,
        };

        render(<ClipCard clip={clipWithNegativeScore} />);

        const voteScore = screen.getByText('-5');
        expect(voteScore).toHaveClass('text-downvote');
    });

    it('displays zero vote score in muted color', () => {
        const clipWithZeroScore = {
            ...mockClip,
            vote_score: 0,
        };

        render(<ClipCard clip={clipWithZeroScore} />);

        const voteScore = screen.getByText('0');
        expect(voteScore).toHaveClass('text-muted-foreground');
    });

    it('highlights upvote button when user has upvoted', () => {
        const upvotedClip = {
            ...mockClip,
            user_vote: 1 as const,
        };

        render(<ClipCard clip={upvotedClip} />);

        const upvoteButton = screen.getByLabelText(/upvote/i);
        // When not authenticated, button is disabled and shows login prompt
        expect(upvoteButton).toBeInTheDocument();
    });

    it('highlights downvote button when user has downvoted', () => {
        const downvotedClip = {
            ...mockClip,
            user_vote: -1 as const,
        };

        render(<ClipCard clip={downvotedClip} />);

        const downvoteButton = screen.getByLabelText(/downvote/i);
        // When not authenticated, button is disabled and shows login prompt
        expect(downvoteButton).toBeInTheDocument();
    });

    it('formats clip age correctly', () => {
        const recentClip = {
            ...mockClip,
            created_at: new Date(Date.now() - 3600000).toISOString(), // 1 hour ago
        };

        render(<ClipCard clip={recentClip} />);
        expect(screen.getByText(/about 1 hour/i)).toBeInTheDocument();
    });

    it('displays NSFW badge when clip is marked NSFW', () => {
        const nsfwClip = {
            ...mockClip,
            is_nsfw: true,
        };

        render(<ClipCard clip={nsfwClip} />);
        expect(screen.getByText('NSFW')).toBeInTheDocument();
    });

    it('displays featured badge when clip is featured', () => {
        const featuredClip = {
            ...mockClip,
            is_featured: true,
        };

        render(<ClipCard clip={featuredClip} />);
        expect(screen.getByText('Featured')).toBeInTheDocument();
    });

    it('has a link to clip detail page', () => {
        render(<ClipCard clip={mockClip} />);

        const link = screen.getByRole('link', { name: /amazing play/i });
        expect(link).toHaveAttribute('href', `/clip/${mockClip.id}`);
    });

    it('displays watch progress indicator when progress data is available', () => {
        const clipWithProgress = {
            ...mockClip,
            watch_progress: {
                progress_seconds: 15,
                duration_seconds: 30,
                progress_percent: 50,
                completed: false,
                watched_at: '2024-01-01T12:00:00Z',
            },
        };

        render(<ClipCard clip={clipWithProgress} />);

        // Check for progress bar with correct aria attributes
        const progressBar = screen.getByRole('progressbar');
        expect(progressBar).toBeInTheDocument();
        expect(progressBar).toHaveAttribute('aria-valuenow', '50');
        expect(progressBar).toHaveAttribute('aria-valuemin', '0');
        expect(progressBar).toHaveAttribute('aria-valuemax', '100');
        expect(progressBar).toHaveAttribute('aria-label', '50% watched');
    });

    it('displays "Watched" badge when clip is completed', () => {
        const completedClip = {
            ...mockClip,
            watch_progress: {
                progress_seconds: 30,
                duration_seconds: 30,
                progress_percent: 100,
                completed: true,
                watched_at: '2024-01-01T12:00:00Z',
            },
        };

        render(<ClipCard clip={completedClip} />);

        // Check for "Watched" badge
        expect(screen.getByText('Watched')).toBeInTheDocument();
    });

    it('does not display progress indicator when watch progress is not available', () => {
        render(<ClipCard clip={mockClip} />);

        // Should not have progress bar
        expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
        // Should not have "Watched" badge
        expect(screen.queryByText('Watched')).not.toBeInTheDocument();
    });

    it('clamps progress percentage to 100 when value exceeds maximum', () => {
        const clipWithExcessProgress = {
            ...mockClip,
            watch_progress: {
                progress_seconds: 40,
                duration_seconds: 30,
                progress_percent: 133, // Over 100%
                completed: false,
                watched_at: '2024-01-01T12:00:00Z',
            },
        };

        render(<ClipCard clip={clipWithExcessProgress} />);

        const progressBar = screen.getByRole('progressbar');
        // Should clamp to 100 for ARIA attributes
        expect(progressBar).toHaveAttribute('aria-valuenow', '100');
        expect(progressBar).toHaveAttribute('aria-label', '100% watched');
        // Visual indicator should be clamped to 100%
        const progressFill = progressBar.querySelector('.bg-primary-600');
        expect(progressFill).toHaveStyle({ width: '100%' });
    });

    it('clamps progress percentage to 0 when value is negative', () => {
        const clipWithNegativeProgress = {
            ...mockClip,
            watch_progress: {
                progress_seconds: -5,
                duration_seconds: 30,
                progress_percent: -10,
                completed: false,
                watched_at: '2024-01-01T12:00:00Z',
            },
        };

        render(<ClipCard clip={clipWithNegativeProgress} />);

        const progressBar = screen.getByRole('progressbar');
        // Should clamp to 0 for ARIA attributes
        expect(progressBar).toHaveAttribute('aria-valuenow', '0');
        expect(progressBar).toHaveAttribute('aria-label', '0% watched');
        // Visual indicator should be clamped to 0%
        const progressFill = progressBar.querySelector('.bg-primary-600');
        expect(progressFill).toHaveStyle({ width: '0%' });
    });
});
