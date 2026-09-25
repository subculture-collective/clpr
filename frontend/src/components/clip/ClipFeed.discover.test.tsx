import { render, screen } from '@/test/test-utils';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, vi } from 'vitest';
import * as clipApi from '@/lib/clip-api';
import { ClipFeed } from './ClipFeed';
import type { Clip, ClipFeedResponse } from '@/types/clip';

vi.mock('@/lib/clip-api');
vi.mock('react-intersection-observer', () => ({
    useInView: () => ({ ref: vi.fn(), inView: true }),
}));

const clips = [0, 1].map(index => ({
    id: `clip-${index}`,
    twitch_clip_id: `Slug${index}`,
    twitch_clip_url: `https://clips.twitch.tv/Slug${index}`,
    title: `Clip ${index}`,
    creator_name: 'creator',
    broadcaster_name: 'broadcaster',
    created_at: '2026-09-24T00:00:00Z',
    duration: 30,
    is_nsfw: index === 0,
})) as Clip[];
const response: ClipFeedResponse = { clips, total: clips.length, page: 1, limit: clips.length, has_more: false };

beforeEach(() => {
    vi.mocked(clipApi.fetchClips).mockReset().mockResolvedValue(response);
    vi.spyOn(HTMLElement.prototype, 'offsetHeight', 'get').mockReturnValue(720);
    vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockImplementation(function (this: HTMLElement) {
        const top = Number(this.dataset.index || 0) * 720;
        return { x: 0, y: top, top, bottom: top + 720, left: 0, right: 800, width: 800, height: 720, toJSON: () => ({}) };
    });
});
afterEach(() => vi.restoreAllMocks());

describe('ClipFeed discover mode', () => {
    it('plays a discover clip when its play button is pressed, one clip at a time', async () => {
        render(<ClipFeed discoverMode />);
        await userEvent.click(await screen.findByRole('button', { name: 'Play Clip 0' }));
        expect(screen.getByTitle('Clip 0')).toHaveAttribute('src', expect.stringContaining('clip=Slug0'));
        expect(screen.queryByText('NSFW')).not.toBeInTheDocument();

        await userEvent.click(screen.getByRole('button', { name: 'Play Clip 1' }));
        expect(screen.getByTitle('Clip 1')).toBeInTheDocument();
        expect(screen.queryByTitle('Clip 0')).not.toBeInTheDocument();
        expect(screen.getByText('NSFW')).toBeInTheDocument();
    });
});
