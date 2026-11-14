import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { afterEach, describe, expect, it, vi } from 'vitest';
import * as broadcasterApi from '../lib/broadcaster-api';
import { BroadcasterPage } from './BroadcasterPage';

vi.mock('../lib/broadcaster-api');
vi.mock('../lib/api', () => ({ apiClient: { get: vi.fn().mockResolvedValue({ data: { data: [] } }) } }));
vi.mock('../context/AuthContext', () => ({ useAuth: () => ({ isAuthenticated: false }) }));
vi.mock('../context/ToastContext', () => ({ useToast: () => ({ showToast: vi.fn() }) }));
vi.mock('../components/SEO', () => ({ SEO: () => null }));
vi.mock('../components/countdown/CountdownFeatures', () => ({ CountdownFeatures: () => null }));
vi.mock('../components/broadcaster', () => ({
    LiveBadge: ({ onClick }: { onClick?: () => void }) => <button onClick={onClick}>LIVE</button>,
}));

describe('BroadcasterPage', () => {
    afterEach(() => vi.restoreAllMocks());

    it('opens the Twitch channel URL from the LIVE badge, not the numeric ID', async () => {
        vi.mocked(broadcasterApi.fetchBroadcasterProfile).mockResolvedValue({
            broadcaster_id: '123456789',
            broadcaster_name: 'somestreamer',
            display_name: 'SomeStreamer',
            twitch_url: 'https://twitch.tv/somestreamer',
            total_clips: 0,
            total_views: 0,
            avg_vote_score: 0,
            follower_count: 0,
            is_following: false,
        } as unknown as Awaited<ReturnType<typeof broadcasterApi.fetchBroadcasterProfile>>);
        vi.mocked(broadcasterApi.fetchBroadcasterClips).mockResolvedValue({
            data: [],
            meta: { page: 1, limit: 20, total: 0, total_pages: 0 },
        } as unknown as Awaited<ReturnType<typeof broadcasterApi.fetchBroadcasterClips>>);
        vi.mocked(broadcasterApi.fetchBroadcasterLiveStatus).mockResolvedValue({
            is_live: true,
        } as unknown as Awaited<ReturnType<typeof broadcasterApi.fetchBroadcasterLiveStatus>>);
        const open = vi.spyOn(window, 'open').mockImplementation(() => null);

        render(
            <QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}>
                <MemoryRouter initialEntries={['/broadcaster/123456789']}>
                    <Routes>
                        <Route path='/broadcaster/:broadcasterId' element={<BroadcasterPage />} />
                    </Routes>
                </MemoryRouter>
            </QueryClientProvider>,
        );

        fireEvent.click(await screen.findByRole('button', { name: 'LIVE' }));

        expect(open).toHaveBeenCalledWith('https://twitch.tv/somestreamer', '_blank', 'noopener,noreferrer');
    });
});
