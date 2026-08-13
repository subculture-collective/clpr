import { render, screen, waitFor } from '@/test/test-utils';
import { vi } from 'vitest';
import * as clipApi from '@/lib/clip-api';
import { ClipFeed } from './ClipFeed';
import type { Clip, ClipFeedResponse } from '@/types/clip';

vi.mock('@/lib/clip-api');
vi.mock('react-intersection-observer', () => ({
    useInView: () => ({ ref: vi.fn(), inView: true }),
}));
vi.mock('./ClipCard', () => ({
    ClipCard: ({ clip }: { clip: Clip }) => <article data-testid='clip-card'>{clip.id}</article>,
}));
vi.mock('./DiscoverClipCard', () => ({
    DiscoverClipCard: ({ clip }: { clip: Clip }) => <article data-testid='discover-clip-card'>{clip.id}</article>,
}));

describe('ClipFeed windowing', () => {
    it('keeps only the latest 12 revealed cards mounted', async () => {
        const clips = Array.from({ length: 24 }, (_, index) => ({
            id: `clip-${index}`,
            title: `Clip ${index}`,
        })) as Clip[];
        const response: ClipFeedResponse = {
            clips,
            total: clips.length,
            page: 1,
            limit: clips.length,
            has_more: false,
        };
        vi.mocked(clipApi.fetchClips).mockResolvedValue(response);

        render(<ClipFeed />);

        await waitFor(() => {
            expect(screen.getAllByTestId('clip-card')).toHaveLength(12);
            expect(screen.getByText('clip-23')).toBeInTheDocument();
        });
        expect(screen.queryByText('clip-0')).not.toBeInTheDocument();
    });
});
