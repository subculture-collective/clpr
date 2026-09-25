import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { Clip } from '../types/clip';

const useClipByIdMock = vi.fn();

vi.mock('../components', async () => {
    const { ResourceUnavailable } = await vi.importActual<typeof import('../components/ui/ResourceUnavailable')>(
        '../components/ui/ResourceUnavailable',
    );
    return {
        Container: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
        Spinner: () => <div role='progressbar' />,
        CommentSection: () => null,
        SEO: () => null,
        VideoPlayer: () => <div data-testid='player' />,
        TheatreMode: () => null,
        ResourceUnavailable,
    };
});

vi.mock('../hooks', () => ({
    useClipById: (id: string) => useClipByIdMock(id),
    useUser: () => null,
    useClipVote: () => ({ mutate: vi.fn(), isPending: false }),
    useClipFavorite: () => ({ mutate: vi.fn(), isPending: false }),
    useIsAuthenticated: () => false,
    useToast: () => ({ info: vi.fn(), error: vi.fn() }),
    useWatchHistory: () => ({
        progress: 0,
        hasProgress: false,
        isLoading: false,
        recordProgress: vi.fn(),
        recordProgressOnPause: vi.fn(),
    }),
}));

vi.mock('@/components/clip/ShareButton', () => ({ ShareButton: () => null }));
vi.mock('@/components/tag/TagList', () => ({ TagList: () => null }));
vi.mock('@/lib/topic-api', () => ({ topicApi: { getClipTopics: vi.fn(async () => ({ topics: [] })) } }));

import { ClipDetailPage } from './ClipDetailPage';

function axiosError(status: number) {
    return Object.assign(new Error(`Request failed with status code ${status}`), {
        isAxiosError: true,
        response: { status, data: {} },
    });
}

function renderClip(id = '00000000-0000-0000-0000-000000000000') {
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return render(
        <QueryClientProvider client={client}>
            <MemoryRouter initialEntries={[`/clip/${id}`]}>
                <Routes>
                    <Route path='/clip/:id' element={<ClipDetailPage />} />
                </Routes>
            </MemoryRouter>
        </QueryClientProvider>,
    );
}

const clip: Clip = {
    id: 'clip-1',
    twitch_clip_id: 'twitch-1',
    twitch_clip_url: 'https://clips.twitch.tv/twitch-1',
    embed_url: 'https://clips.twitch.tv/embed?clip=twitch-1',
    title: 'Special Olympus',
    creator_name: 'Creator',
    broadcaster_name: 'sodapoppin',
    broadcaster_id: '26301881',
    thumbnail_url: 'https://example.com/thumb.jpg',
    duration: 57,
    view_count: 499,
    vote_score: 0,
    comment_count: 0,
    favorite_count: 0,
    created_at: '2026-09-17T00:00:00Z',
    imported_at: '2026-09-17T00:00:00Z',
    is_featured: false,
    is_nsfw: false,
    is_removed: false,
};

describe('ClipDetailPage', () => {
    beforeEach(() => useClipByIdMock.mockReset());

    it('shows a friendly not-found state for a 404 without the transport message', () => {
        useClipByIdMock.mockReturnValue({ data: undefined, isLoading: false, error: axiosError(404), refetch: vi.fn() });
        renderClip();

        expect(screen.getByRole('heading', { level: 1, name: "This clip isn't here" })).toBeInTheDocument();
        expect(screen.getByRole('link', { name: 'Back to the feed' })).toHaveAttribute('href', '/');
        expect(screen.getByRole('link', { name: 'Search clips' })).toHaveAttribute('href', '/search');
        expect(screen.queryByText(/status code/i)).not.toBeInTheDocument();
        expect(screen.queryByRole('button', { name: 'Try again' })).not.toBeInTheDocument();
    });

    it('offers a retry for other failures without the transport message', async () => {
        const refetch = vi.fn();
        useClipByIdMock.mockReturnValue({ data: undefined, isLoading: false, error: axiosError(503), refetch });
        renderClip();

        expect(screen.getByRole('alert')).toHaveTextContent("We couldn't load this clip");
        expect(screen.queryByText(/status code/i)).not.toBeInTheDocument();
        await userEvent.click(screen.getByRole('button', { name: 'Try again' }));
        expect(refetch).toHaveBeenCalledTimes(1);
    });

    it('labels the stored view count as a sync snapshot, not the live Twitch count', () => {
        useClipByIdMock.mockReturnValue({ data: clip, isLoading: false, error: null, refetch: vi.fn() });
        renderClip('clip-1');

        const views = screen.getByText('499 views at last sync');
        expect(views).toHaveAttribute('title', expect.stringContaining('The player shows the live count'));
    });
});
