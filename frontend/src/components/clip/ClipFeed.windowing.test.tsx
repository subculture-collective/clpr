import { act, render, screen, waitFor } from '@/test/test-utils';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, vi } from 'vitest';
import * as clipApi from '@/lib/clip-api';
import { ClipFeed } from './ClipFeed';
import type { Clip, ClipFeedResponse } from '@/types/clip';

vi.mock('@/lib/clip-api');
vi.mock('react-intersection-observer', () => ({
    useInView: () => ({ ref: vi.fn(), inView: true }),
}));
vi.mock('./ClipCard', () => ({
    ClipCard: ({ clip }: { clip: Clip }) => <article data-testid='clip-card'>{clip.id}<button aria-label={`Play ${clip.id}`}>Play</button></article>,
}));
vi.mock('./DiscoverClipCard', () => ({
    DiscoverClipCard: ({ clip }: { clip: Clip }) => <article>{clip.id}</article>,
}));

const clips = Array.from({ length: 40 }, (_, index) => ({ id: `clip-${index}`, title: `Clip ${index}` })) as Clip[];
const response: ClipFeedResponse = { clips, total: clips.length, page: 1, limit: clips.length, has_more: false };
const scrollTo = (offset: number) => act(() => {
    Object.defineProperty(window, 'scrollY', { configurable: true, value: offset });
    window.dispatchEvent(new Event('scroll'));
});

beforeEach(() => {
    vi.mocked(clipApi.fetchClips).mockReset();
    Object.defineProperty(window, 'scrollY', { configurable: true, value: 0 });
    vi.spyOn(HTMLElement.prototype, 'offsetHeight', 'get').mockReturnValue(720);
    vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockImplementation(function (this: HTMLElement) {
        const top = Number(this.dataset.index || 0) * 720 - window.scrollY;
        return { x: 0, y: top, top, bottom: top + 720, left: 0, right: 800, width: 800, height: 720, toJSON: () => ({}) };
    });
});
afterEach(() => { vi.restoreAllMocks(); Object.defineProperty(window, 'scrollY', { configurable: true, value: 0 }); });

describe('ClipFeed continuity', () => {
    it('restores earlier clips when scrolling back while bounding mounted cards', async () => {
        vi.mocked(clipApi.fetchClips).mockResolvedValue(response);
        render(<ClipFeed />);
        await screen.findByText('clip-0');
        expect(screen.getAllByTestId('clip-card').length).toBeLessThanOrEqual(12);
        scrollTo(14 * 720);
        await screen.findByText('clip-14');
        expect(screen.queryByText('clip-0')).not.toBeInTheDocument();
        expect(screen.getAllByTestId('clip-card').length).toBeLessThanOrEqual(12);
        scrollTo(0);
        await screen.findByText('clip-0');
        expect(screen.queryByText('clip-14')).not.toBeInTheDocument();
    });

    it('keeps the focused card mounted while the viewport moves away', async () => {
        vi.mocked(clipApi.fetchClips).mockResolvedValue(response);
        render(<ClipFeed />);
        const button = await screen.findByRole('button', { name: 'Play clip-0' });
        act(() => button.focus());
        scrollTo(14 * 720);
        await screen.findByText('clip-14');
        expect(button).toHaveFocus();
        expect(screen.getAllByTestId('clip-card').length).toBeLessThanOrEqual(12);
    });

    it('retains loaded clips after a later page fails and retries that page', async () => {
        vi.mocked(clipApi.fetchClips)
            .mockResolvedValueOnce({ ...response, clips: [clips[0]], has_more: true, cursor: 'next-page' })
            .mockRejectedValueOnce(new Error('Connection lost'))
            .mockResolvedValueOnce({ ...response, clips: [clips[1]], page: 2 });
        render(<ClipFeed />);
        await screen.findByRole('alert');
        expect(screen.getByText('clip-0')).toBeInTheDocument();
        await userEvent.click(screen.getByRole('button', { name: 'Try again' }));
        await screen.findByText('clip-1');
        await waitFor(() => expect(screen.queryByRole('alert')).not.toBeInTheDocument());
        expect(vi.mocked(clipApi.fetchClips).mock.calls.slice(1).map(([args]) => args?.cursor)).toEqual(['next-page', 'next-page']);
    });
});
