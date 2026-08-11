import { render, screen } from '@/test/test-utils';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { CreatorsPage } from './CreatorsPage';

const fetchCreatorDiscovery = vi.fn();
const fetchLiveBroadcasters = vi.fn();

vi.mock('@/lib/broadcaster-api', () => ({
    fetchCreatorDiscovery: (...args: unknown[]) => fetchCreatorDiscovery(...args),
    fetchLiveBroadcasters: (...args: unknown[]) => fetchLiveBroadcasters(...args),
}));

vi.mock('@/components/SEO', () => ({ SEO: () => null }));

const creator = (id: string, name: string) => ({
    broadcaster_id: id,
    broadcaster_name: name,
    total_clips: 4,
    recent_clips: 2,
    total_views: 12_000,
    recent_views: 8_000,
    view_velocity: 120,
    follower_count: 3,
    first_discovered_at: '2026-08-11T12:00:00Z',
    latest_clip_at: '2026-08-11T13:00:00Z',
    latest_clip_thumbnail: `https://example.com/${id}.jpg`,
    latest_clip_title: `${name}'s latest moment`,
    twitch_category_name: 'Just Chatting',
    score: 100,
});

describe('CreatorsPage', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        fetchCreatorDiscovery.mockResolvedValue({
            trending: [creator('trending-id', 'Trending Creator')],
            rising: [creator('rising-id', 'Rising Creator')],
            new: [creator('new-id', 'New Creator')],
        });
        fetchLiveBroadcasters.mockResolvedValue({
            success: true,
            data: [
                {
                    broadcaster_id: 'live-id',
                    user_name: 'Live Creator',
                    is_live: true,
                    viewer_count: 1200,
                    stream_title: 'Talking through the news',
                },
            ],
            meta: { page: 1, limit: 12, total_items: 1, total_pages: 1 },
        });
    });

    it('presents distinct creator discovery rails and profile links', async () => {
        render(<CreatorsPage />);

        expect(
            screen.getByRole('heading', { name: /follow the people/i }),
        ).toBeInTheDocument();
        expect(
            await screen.findByRole('heading', { name: 'Creators moving now' }),
        ).toBeInTheDocument();
        expect(screen.getByRole('heading', { name: 'On the rise' })).toBeInTheDocument();
        expect(screen.getByRole('heading', { name: 'Live right now' })).toBeInTheDocument();
        expect(screen.getByRole('heading', { name: 'Fresh faces' })).toBeInTheDocument();

        expect(screen.getByRole('link', { name: /Trending Creator/ })).toHaveAttribute(
            'href',
            '/broadcaster/trending-id',
        );
        expect(screen.getByRole('link', { name: /Live Creator/ })).toHaveAttribute(
            'href',
            '/broadcaster/live-id',
        );
    });
});
