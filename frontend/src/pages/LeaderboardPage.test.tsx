import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { LeaderboardEntry } from '../types/reputation';
import { allowTestConsole } from '../test/setup';

vi.mock('../context/AuthContext', () => ({ useAuth: () => ({ user: null }) }));
vi.mock('../components', () => ({ SEO: () => null }));
const apiGet = vi.hoisted(() => vi.fn());
vi.mock('../lib/api', () => ({ apiClient: { get: apiGet } }));

import LeaderboardPage from './LeaderboardPage';

function entry(rank: number, score: number): LeaderboardEntry {
    return {
        rank,
        user_id: `user-${rank}`,
        username: `member${rank}`,
        display_name: `Member ${rank}`,
        score,
        user_rank: 'Newcomer',
    };
}

function mockLeaderboard(entries: LeaderboardEntry[], page = 1) {
    const fetchMock = vi.fn(async () => new Response(
        JSON.stringify({ type: 'karma', page, limit: 50, entries }),
        { status: 200, headers: { 'content-type': 'application/json' } },
    ));
    vi.stubGlobal('fetch', fetchMock);
    return fetchMock;
}

function renderAt(path = '/leaderboards') {
    return render(
        <MemoryRouter initialEntries={[path]}>
            <LeaderboardPage />
        </MemoryRouter>,
    );
}

describe('LeaderboardPage', () => {
    afterEach(() => vi.unstubAllGlobals());

    it('explains an all-zero leaderboard instead of paginating empty pages', async () => {
        // Production shape: a full page of members who have not earned points yet.
        mockLeaderboard(Array.from({ length: 50 }, (_, index) => entry(index + 1, 0)));
        renderAt();

        expect(await screen.findByText('No rankings yet')).toBeInTheDocument();
        expect(screen.getByText(/fill in as people vote/)).toBeInTheDocument();
        expect(screen.queryByText('No leaderboard data available')).not.toBeInTheDocument();
        expect(screen.queryByRole('button', { name: 'Next' })).not.toBeInTheDocument();
    });

    it('disables Next once the ranked members run out on this page', async () => {
        mockLeaderboard(Array.from({ length: 50 }, (_, index) => entry(index + 1, index < 10 ? 100 - index : 0)));
        renderAt('/leaderboards?type=karma&page=2');

        expect(await screen.findByRole('button', { name: 'Next' })).toBeDisabled();
        expect(screen.getByRole('button', { name: 'Previous' })).toBeEnabled();
        expect(screen.getAllByRole('row')).toHaveLength(11);
    });

    it('enables Next when every member on a full page has points', async () => {
        mockLeaderboard(Array.from({ length: 50 }, (_, index) => entry(index + 1, 500 - index)));
        renderAt();

        expect(await screen.findByRole('button', { name: 'Next' })).toBeEnabled();
        expect(screen.getByRole('button', { name: 'Previous' })).toBeDisabled();
    });

    it('shows a friendly error without the server message', async () => {
        allowTestConsole('error', /Leaderboard fetch error/);
        vi.stubGlobal('fetch', vi.fn(async () => new Response('upstream exploded', { status: 502 })));
        renderAt();

        expect(await screen.findByText("We couldn't load the leaderboard")).toBeInTheDocument();
        expect(screen.queryByText(/upstream exploded|502/)).not.toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Try again' })).toBeInTheDocument();
    });

    it('lets the type tabs wrap and exposes the selected tab', async () => {
        mockLeaderboard([]);
        renderAt('/leaderboards?type=engagement');

        const group = screen.getByRole('group', { name: 'Leaderboard type' });
        expect(group.className).toContain('flex-wrap');
        expect(screen.getByRole('button', { name: /Engagement/ })).toHaveAttribute('aria-pressed', 'true');
        expect(await screen.findByText('No rankings yet')).toBeInTheDocument();
    });
});
