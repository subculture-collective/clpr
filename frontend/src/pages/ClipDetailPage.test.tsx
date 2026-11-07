import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Link, MemoryRouter, Route, Routes } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { Clip } from '../types/clip';
import { apiClient } from '@/lib/api';

const useClipByIdMock = vi.fn();
const isAuthenticatedMock = vi.fn(() => false);
const voteMutate = vi.fn();
const toastError = vi.fn();

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
        TheatreMode: ({ captionsUrl, captionsLanguage }: { captionsUrl?: string; captionsLanguage?: string }) => (
            <div data-testid='theatre' data-captions={captionsUrl ?? ''} data-captions-lang={captionsLanguage ?? ''} />
        ),
        ResourceUnavailable,
    };
});

vi.mock('../hooks', () => ({
    useClipById: (id: string) => useClipByIdMock(id),
    useUser: () => null,
    useClipVote: () => ({ mutate: voteMutate, isPending: false }),
    useClipFavorite: () => ({ mutate: vi.fn(), isPending: false }),
    useIsAuthenticated: () => isAuthenticatedMock(),
    useToast: () => ({ info: vi.fn(), error: toastError }),
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
    beforeEach(() => {
        useClipByIdMock.mockReset();
        isAuthenticatedMock.mockReturnValue(false);
        voteMutate.mockReset();
        toastError.mockReset();
    });

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

    it('links an archive import to its source passage and attaches its captions', () => {
        const passageURL = 'https://archive.example/api/share/videos/v1?start_ms=754000&end_ms=763000';
        const cue = { start_ms: 0, end_ms: 1840, text: 'and that is the part nobody reads,' };
        useClipByIdMock.mockReturnValue({
            data: {
                ...clip,
                video_url: '/api/v1/clips/clip-1/media',
                source_metadata: {
                    archive_provenance: {
                        export_id: '9b1c2d3e-4f50-4a61-8b72-93a4b5c6d7e8',
                        archive: { name: 'Transcript Archive', origin: 'https://archive.example' },
                        passage: { url: passageURL, player_url: passageURL, start_ms: 754000, end_ms: 763000, duration_ms: 9000 },
                        captions: { source_cues: [cue], cues: [cue], edited: false, language: 'en' },
                    },
                },
            },
            isLoading: false,
            error: null,
            refetch: vi.fn(),
        });
        renderClip('clip-1');

        const link = screen.getByRole('link', { name: /Transcript Archive passage, 12:34.0–12:43.0/ });
        expect(link).toHaveAttribute('href', passageURL);
        expect(link).toHaveAttribute('rel', expect.stringContaining('noopener'));
        expect(screen.getByText(/not verified by clpr/)).toBeInTheDocument();
        expect(screen.getByTestId('theatre')).toHaveAttribute('data-captions', '/api/v1/clips/clip-1/captions');
        expect(screen.getByTestId('theatre')).toHaveAttribute('data-captions-lang', 'en');
    });

    it('shows no archive source for an ordinary clip', () => {
        useClipByIdMock.mockReturnValue({ data: { ...clip, video_url: '/api/v1/clips/clip-1/media' }, isLoading: false, error: null, refetch: vi.fn() });
        renderClip('clip-1');

        expect(screen.queryByTestId('archive-source')).not.toBeInTheDocument();
        expect(screen.getByTestId('theatre')).toHaveAttribute('data-captions', '');
    });

    it('labels the stored view count as a sync snapshot, not the live Twitch count', () => {
        useClipByIdMock.mockReturnValue({ data: clip, isLoading: false, error: null, refetch: vi.fn() });
        renderClip('clip-1');

        const views = screen.getByText('499 Twitch views at last sync');
        expect(views).toHaveAttribute('title', expect.stringContaining('The player shows the live count'));
    });

    it('records an embed view for each clip when moving between clips', async () => {
        isAuthenticatedMock.mockReturnValue(true);
        const post = vi.spyOn(apiClient, 'post').mockResolvedValue({ data: {} });
        useClipByIdMock.mockImplementation((id: string) => ({ data: { ...clip, id }, isLoading: false, error: null, refetch: vi.fn() }));
        const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
        render(
            <QueryClientProvider client={client}>
                <MemoryRouter initialEntries={['/clip/clip-1']}>
                    <Routes>
                        <Route path='/clip/:id' element={<><ClipDetailPage /><Link to='/clip/clip-2'>Next clip</Link></>} />
                    </Routes>
                </MemoryRouter>
            </QueryClientProvider>,
        );

        await userEvent.click(screen.getByRole('link', { name: 'Next clip' }));

        const viewed = post.mock.calls.filter(([url]) => url === '/watch-history').map(([, body]) => (body as { clip_id: string }).clip_id);
        expect(viewed).toEqual(['clip-1', 'clip-2']);
        post.mockRestore();
    });

    it('shows the message from an error envelope as text when a vote fails', async () => {
        isAuthenticatedMock.mockReturnValue(true);
        vi.spyOn(apiClient, 'post').mockResolvedValue({ data: {} });
        voteMutate.mockImplementation((_payload, options: { onError: (error: unknown) => void }) => {
            options.onError({ response: { status: 403, data: { success: false, error: { code: 'FORBIDDEN', message: 'Voting is paused' } } } });
        });
        useClipByIdMock.mockReturnValue({ data: clip, isLoading: false, error: null, refetch: vi.fn() });
        renderClip('clip-1');

        await userEvent.click(screen.getByRole('button', { name: 'Upvote: 0 votes' }));

        expect(toastError).toHaveBeenCalledWith('Voting is paused');
        vi.restoreAllMocks();
    });
});
