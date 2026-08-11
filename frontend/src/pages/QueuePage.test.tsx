import { fireEvent, render, screen } from '@/test/test-utils';
import type { PlaylistItem } from '@/components/playlist/PlaylistTheatreMode';
import type { Queue } from '@/types/queue';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { QueuePage } from './QueuePage';

let mockQueueData: Queue;

const mockRemoveFromQueue = vi.fn();

vi.mock('@/hooks/useQueue', () => ({
    useQueue: () => ({
        data: mockQueueData,
        isLoading: false,
        isError: false,
    }),
    useRemoveFromQueue: () => ({ mutate: mockRemoveFromQueue }),
    useClearQueue: () => ({ mutate: vi.fn() }),
    useReorderQueue: () => ({ mutate: vi.fn() }),
}));

vi.mock('@/components/playlist/PlaylistTheatreMode', () => ({
    PlaylistTheatreMode: ({
        items,
        currentItemId,
        onItemClick,
    }: {
        items: PlaylistItem[];
        currentItemId: string | null;
        onItemClick: (item: PlaylistItem) => void;
    }) => (
        <div>
            <div data-testid='current-item'>{currentItemId ?? 'none'}</div>
            {items.map(item => (
                <button key={item.id} onClick={() => onItemClick(item)}>
                    Select {item.id}
                </button>
            ))}
        </div>
    ),
}));

vi.mock('@/components/queue/ConvertToPlaylistDialog', () => ({
    ConvertToPlaylistDialog: () => null,
}));

vi.mock('@/components/SEO', () => ({
    SEO: () => null,
}));

const createQueue = (ids: string[]): Queue => ({
    items: ids.map((id, index) => ({
        id,
        clip_id: `clip-${index}`,
        user_id: 'user-1',
        position: index,
        added_at: '2026-08-11T00:00:00Z',
        created_at: '2026-08-11T00:00:00Z',
        updated_at: '2026-08-11T00:00:00Z',
        clip: {
            id: `clip-${index}`,
            twitch_clip_id: `twitch-${index}`,
            twitch_clip_url: `https://clips.twitch.tv/${index}`,
            embed_url: `https://clips.twitch.tv/embed?clip=${index}`,
            title: `Clip ${index}`,
            creator_name: 'Creator',
            broadcaster_name: 'Broadcaster',
            game_name: 'Category',
            thumbnail_url: `https://example.com/${index}.jpg`,
            duration: 30,
            view_count: 100,
            vote_score: 0,
            comment_count: 0,
            favorite_count: 0,
            created_at: '2026-08-11T00:00:00Z',
            imported_at: '2026-08-11T00:00:00Z',
            is_featured: false,
            is_nsfw: false,
            is_removed: false,
            user_vote: null,
            is_favorited: false,
        },
    })),
    total: ids.length,
    next_clip: undefined,
});

describe('QueuePage', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        mockQueueData = createQueue(['item-1', 'item-2']);
    });

    it('falls back to the first remaining item when the selection disappears', async () => {
        const { rerender } = render(<QueuePage />);

        expect(await screen.findByTestId('current-item')).toHaveTextContent(
            'item-1',
        );

        fireEvent.click(screen.getByRole('button', { name: 'Select item-2' }));
        expect(screen.getByTestId('current-item')).toHaveTextContent('item-2');

        mockQueueData = createQueue(['item-1']);
        rerender(<QueuePage />);

        expect(screen.getByTestId('current-item')).toHaveTextContent('item-1');
    });
});
