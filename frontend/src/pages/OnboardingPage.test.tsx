import { render, screen } from '@/test/test-utils';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { OnboardingPage } from './OnboardingPage';

const fetchCreatorDiscovery = vi.fn();
const followBroadcaster = vi.fn();
const listCategories = vi.fn();
const listTags = vi.fn();
const completeOnboarding = vi.fn();

vi.mock('@/lib/broadcaster-api', () => ({
    fetchCreatorDiscovery: (...args: unknown[]) => fetchCreatorDiscovery(...args),
    followBroadcaster: (...args: unknown[]) => followBroadcaster(...args),
}));
vi.mock('@/lib/category-api', () => ({ categoryApi: { listCategories: (...args: unknown[]) => listCategories(...args) } }));
vi.mock('@/lib/tag-api', () => ({ tagApi: { listTags: (...args: unknown[]) => listTags(...args) } }));
vi.mock('@/lib/recommendation-api', () => ({
    completeCreatorFirstOnboarding: (...args: unknown[]) => completeOnboarding(...args),
}));
vi.mock('@/components/SEO', () => ({ SEO: () => null }));

describe('OnboardingPage', () => {
    beforeEach(() => {
        vi.clearAllMocks();
        fetchCreatorDiscovery.mockResolvedValue({
            trending: [{ broadcaster_id: 'creator-1', broadcaster_name: 'Creator One', latest_clip_thumbnail: 'https://example.com/one.jpg', twitch_category_name: 'Just Chatting' }],
            rising: [],
            new: [],
        });
        listCategories.mockResolvedValue({ categories: [{ id: 'topic-1', name: 'News & Politics', slug: 'news-politics', description: 'News and public life' }] });
        listTags.mockResolvedValue({ tags: [{ id: 'tag-1', name: 'Funny', slug: 'funny' }] });
        followBroadcaster.mockResolvedValue({ message: 'ok' });
        completeOnboarding.mockResolvedValue({ onboarding_completed: true });
    });

    it('collects creators, topics, and moments in that order', async () => {
        const user = userEvent.setup();
        render(<OnboardingPage />);

        await user.click(await screen.findByRole('button', { name: /Creator One/ }));
        await user.click(screen.getByRole('button', { name: /Continue/ }));
        await user.click(await screen.findByRole('button', { name: /News & Politics/ }));
        await user.click(screen.getByRole('button', { name: /Continue/ }));
        await user.click(await screen.findByRole('button', { name: /Funny/ }));
        await user.click(screen.getByRole('button', { name: /Build my feed/ }));

        expect(followBroadcaster).toHaveBeenCalledWith('creator-1');
        expect(completeOnboarding).toHaveBeenCalledWith({
            followed_creators: ['creator-1'],
            preferred_topics: ['news-politics'],
            preferred_tags: ['tag-1'],
        });
    });
});
