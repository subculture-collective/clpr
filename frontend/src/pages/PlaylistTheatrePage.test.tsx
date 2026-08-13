import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { PlaylistTheatrePage } from './PlaylistTheatrePage';

const mocks = vi.hoisted(() => ({
    fetchNextPage: vi.fn().mockResolvedValue({ data: { pages: [] } }),
    infinitePlaylist: vi.fn(),
    navigate: vi.fn(),
}));

vi.mock('@/hooks/usePlaylist', () => ({
    useInfinitePlaylist: mocks.infinitePlaylist,
    useRemoveClipFromPlaylist: () => ({ mutate: vi.fn() }),
    useReorderPlaylistClips: () => ({ mutate: vi.fn() }),
}));

vi.mock('@tanstack/react-query', () => ({
    useQueryClient: () => ({ invalidateQueries: vi.fn() }),
}));

vi.mock('react-router-dom', () => ({
    useParams: () => ({ id: '11111111-1111-4111-8111-111111111111' }),
    useNavigate: () => mocks.navigate,
}));

vi.mock('@/components/playlist/PlaylistTheatreMode', () => ({
    PlaylistTheatreMode: ({ onReachEnd }: { onReachEnd?: () => void }) => (
        <button onClick={onReachEnd}>Advance past loaded clips</button>
    ),
}));

vi.mock('@/components/ui', () => ({ Spinner: () => <div>Loading</div> }));
vi.mock('@/components/SEO', () => ({ SEO: () => null }));

describe('PlaylistTheatrePage pagination', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        mocks.infinitePlaylist.mockReturnValue({
            data: {
                pages: [
                    {
                        success: true,
                        data: {
                            id: '11111111-1111-4111-8111-111111111111',
                            title: 'Long playlist',
                            clips: [
                                {
                                    id: 'clip-1',
                                    title: 'Clip one',
                                    order: 1,
                                },
                            ],
                        },
                        meta: { has_next: true, page: 1 },
                    },
                ],
            },
            isLoading: false,
            isError: false,
            hasNextPage: true,
            isFetchingNextPage: false,
            fetchNextPage: mocks.fetchNextPage,
            refetch: vi.fn(),
        });
    });

    it('loads the next 100-item page when theatre advances beyond loaded clips', async () => {
        render(<PlaylistTheatrePage />);

        fireEvent.click(
            screen.getByRole('button', { name: 'Advance past loaded clips' }),
        );
        await waitFor(() => expect(mocks.fetchNextPage).toHaveBeenCalledOnce());
        expect(mocks.infinitePlaylist).toHaveBeenCalledWith(
            '11111111-1111-4111-8111-111111111111',
            100,
        );
    });
});
