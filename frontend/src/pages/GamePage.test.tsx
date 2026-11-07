import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { allowTestConsole } from '../test/setup';

const { getGame, getGameClips, voteOnClip } = vi.hoisted(() => ({ getGame: vi.fn(), getGameClips: vi.fn(), voteOnClip: vi.fn() }));

vi.mock('../lib/game-api', () => ({ twitchCategoryApi: { getGame, getGameClips } }));
vi.mock('../context/AuthContext', () => ({ useAuth: () => ({ user: null }) }));
vi.mock('../context/ToastContext', () => ({ useToast: () => ({ showToast: vi.fn() }) }));
vi.mock('@/lib/clip-api', () => ({ voteOnClip }));
vi.mock('../components/clip', async () => {
    const { useClipVote } = await import('../hooks/useClips');
    return {
        ClipGridCard: ({ clip }: { clip: { id: string; title: string; vote_score?: number } }) => {
            const vote = useClipVote();
            return (
                <article>
                    <h3>{clip.title}</h3>
                    <span>score {clip.vote_score}</span>
                    <button onClick={() => vote.mutate({ clip_id: clip.id, vote_type: 1 })}>Upvote {clip.title}</button>
                </article>
            );
        },
    };
});

import { GamePage } from './GamePage';

function axiosError(status: number) {
    return Object.assign(new Error(`Request failed with status code ${status}`), {
        isAxiosError: true,
        response: { status, data: { error: 'Game not found' } },
    });
}

function renderGame(id = '999999999') {
    return render(
        <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
            <MemoryRouter initialEntries={[`/twitch-category/${id}`]}>
                <Routes>
                    <Route path='/twitch-category/:gameId' element={<GamePage />} />
                </Routes>
            </MemoryRouter>
        </QueryClientProvider>,
    );
}

describe('GamePage', () => {
    beforeEach(() => {
        getGame.mockReset();
        getGameClips.mockReset();
    });

    it('shows a not-found state with links back when the category is unknown', async () => {
        getGame.mockRejectedValue(axiosError(404));
        getGameClips.mockRejectedValue(axiosError(404));
        renderGame();

        expect(await screen.findByRole('heading', { level: 1, name: "This Twitch category isn't here" })).toBeInTheDocument();
        expect(screen.getByRole('link', { name: 'Browse topics' })).toHaveAttribute('href', '/topics');
        expect(screen.getByRole('link', { name: 'Back to the feed' })).toHaveAttribute('href', '/');
        expect(screen.queryByText(/Failed to load/)).not.toBeInTheDocument();
        expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    });

    it('still lists tracked clips when only the category details are missing', async () => {
        getGame.mockRejectedValue(axiosError(404));
        getGameClips.mockResolvedValue({ clips: [{ id: 'c1', title: 'A tracked clip' }], has_more: false });
        renderGame('509658');

        expect(await screen.findByText('A tracked clip')).toBeInTheDocument();
        expect(screen.getByRole('heading', { level: 1, name: 'Tracked clips' })).toBeInTheDocument();
        expect(screen.queryByText("This Twitch category isn't here")).not.toBeInTheDocument();
    });

    it('offers a retry for other failures without the transport message', async () => {
        allowTestConsole('error', /Failed to fetch game:.*503/);
        getGame.mockRejectedValueOnce(axiosError(503));
        getGameClips.mockResolvedValue({ clips: [], has_more: false });
        renderGame('33214');

        expect(await screen.findByRole('alert')).toHaveTextContent("We couldn't load this Twitch category");
        expect(screen.queryByText(/status code/)).not.toBeInTheDocument();

        getGame.mockResolvedValue({
            game: { id: 'g', twitch_game_id: '33214', name: 'Fortnite', clip_count: 3, follower_count: 1, is_following: false },
        });
        await userEvent.click(screen.getByRole('button', { name: 'Try again' }));
        expect(await screen.findByRole('heading', { level: 1, name: 'Fortnite' })).toBeInTheDocument();
    });

    it('shows a vote on a card without reloading the list', async () => {
        getGame.mockRejectedValue(axiosError(404));
        getGameClips.mockResolvedValue({ clips: [{ id: 'c1', title: 'A tracked clip', vote_score: 5 }], has_more: false });
        voteOnClip.mockResolvedValue({ message: 'ok', vote_score: 6, upvote_count: 6, downvote_count: 0, user_vote: 1 });
        renderGame('509658');

        expect(await screen.findByText('score 5')).toBeInTheDocument();
        await userEvent.click(screen.getByRole('button', { name: 'Upvote A tracked clip' }));
        expect(await screen.findByText('score 6')).toBeInTheDocument();
        expect(getGameClips).toHaveBeenCalledTimes(1);
    });
});
