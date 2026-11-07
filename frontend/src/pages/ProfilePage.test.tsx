import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, expect, it, vi } from 'vitest';

const { fetchUserUpvotedClips, voteOnClip } = vi.hoisted(() => ({
    fetchUserUpvotedClips: vi.fn(),
    voteOnClip: vi.fn(),
}));

vi.mock('../lib/user-api', () => ({
    fetchUserComments: vi.fn(),
    fetchUserUpvotedClips,
    fetchUserDownvotedClips: vi.fn(),
    reauthorizeTwitch: vi.fn(),
}));
vi.mock('../lib/api', () => ({
    apiClient: { get: vi.fn().mockResolvedValue({ data: {} }) },
}));
vi.mock('@/lib/clip-api', () => ({ voteOnClip }));
vi.mock('../context/AuthContext', () => ({
    useAuth: () => ({
        user: { id: 'u1', username: 'viewer', display_name: 'Viewer', karma_points: 1, role: 'user' },
        isAdmin: false,
    }),
}));
vi.mock('../context/ToastContext', () => ({ useToast: () => ({ error: vi.fn(), success: vi.fn() }) }));
vi.mock('../components/reputation', () => ({
    BadgeGrid: () => null,
    KarmaBreakdownChart: () => null,
    ReputationDisplay: () => null,
}));
vi.mock('../components/clip/ClipCard', async () => {
    const { useClipVote } = await import('../hooks/useClips');
    return {
        ClipCard: ({ clip }: { clip: { id: string; title: string; vote_score: number } }) => {
            const vote = useClipVote();
            return (
                <article>
                    <h3>{clip.title}</h3>
                    <span>score {clip.vote_score}</span>
                    <button onClick={() => vote.mutate({ clip_id: clip.id, vote_type: -1 })}>Downvote {clip.title}</button>
                </article>
            );
        },
    };
});

import { ProfilePage } from './ProfilePage';

describe('ProfilePage', () => {
    it('shows a vote on an upvoted clip without reloading the list', async () => {
        fetchUserUpvotedClips.mockResolvedValue({
            clips: [{ id: 'c1', title: 'Liked clip', vote_score: 5 }],
            total: 1,
            page: 1,
            limit: 10,
            has_more: false,
        });
        voteOnClip.mockResolvedValue({ message: 'ok', vote_score: 3, upvote_count: 0, downvote_count: 1, user_vote: -1 });
        render(
            <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
                <MemoryRouter>
                    <ProfilePage />
                </MemoryRouter>
            </QueryClientProvider>,
        );

        await userEvent.click(screen.getByRole('tab', { name: 'Upvoted' }));
        expect(await screen.findByText('score 5')).toBeInTheDocument();
        await userEvent.click(screen.getByRole('button', { name: 'Downvote Liked clip' }));
        expect(await screen.findByText('score 3')).toBeInTheDocument();
        expect(fetchUserUpvotedClips).toHaveBeenCalledTimes(1);
    });

    it('appends the next page of upvoted clips', async () => {
        fetchUserUpvotedClips.mockReset();
        fetchUserUpvotedClips
            .mockResolvedValueOnce({ clips: [{ id: 'c1', title: 'First', vote_score: 1 }], total: 2, page: 1, limit: 10, has_more: true })
            .mockResolvedValueOnce({ clips: [{ id: 'c2', title: 'Second', vote_score: 1 }], total: 2, page: 2, limit: 10, has_more: false });
        render(
            <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
                <MemoryRouter>
                    <ProfilePage />
                </MemoryRouter>
            </QueryClientProvider>,
        );

        await userEvent.click(screen.getByRole('tab', { name: 'Upvoted' }));
        expect(await screen.findByText('First')).toBeInTheDocument();
        await userEvent.click(screen.getByRole('button', { name: 'Load More' }));
        expect(await screen.findByText('Second')).toBeInTheDocument();
        expect(screen.getByText('First')).toBeInTheDocument();
        expect(fetchUserUpvotedClips).toHaveBeenLastCalledWith('u1', 2);
        expect(screen.queryByRole('button', { name: 'Load More' })).not.toBeInTheDocument();
    });
});
