import { render, screen } from '@/test/test-utils';
import type { Clip } from '@/types/clip';
import { describe, expect, it, vi } from 'vitest';
import { ClipGridCard } from './ClipGridCard';

vi.mock('@/hooks/useClips', () => ({
  useClipVote: () => ({ mutate: vi.fn(), isPending: false }),
  useClipFavorite: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('@/hooks', () => ({
  useIsAuthenticated: () => true,
  useToast: () => ({ info: vi.fn() }),
}));

const clip: Clip = {
  id: 'clip-1',
  twitch_clip_id: 'twitch-1',
  twitch_clip_url: 'https://clips.twitch.tv/twitch-1',
  embed_url: 'https://clips.twitch.tv/embed?clip=twitch-1',
  title: 'Accessible highlight',
  creator_name: 'Creator',
  broadcaster_name: 'Broadcaster',
  thumbnail_url: 'https://example.com/thumb.jpg',
  duration: 30,
  view_count: 100,
  vote_score: 5,
  comment_count: 2,
  favorite_count: 3,
  created_at: '2024-01-01T00:00:00Z',
  imported_at: '2024-01-01T00:00:00Z',
  is_featured: false,
  is_nsfw: false,
  is_removed: false,
  user_vote: 1,
  is_favorited: true,
  watch_progress: {
    progress_seconds: 15,
    duration_seconds: 30,
    progress_percent: 50,
    completed: false,
    watched_at: '2024-01-01T00:00:00Z',
  },
};

describe('ClipGridCard', () => {
  it('keeps card actions outside the navigation link', () => {
    render(<ClipGridCard clip={clip} />);

    const link = screen.getByRole('link', { name: 'Watch Accessible highlight' });
    expect(link).toHaveAttribute('href', '/clip/clip-1');
    expect(link.querySelector('button')).toBeNull();
    expect(screen.getByRole('button', { name: 'Upvote Accessible highlight' })).not.toBe(link);
  });

  it('exposes selection and watch progress state', () => {
    render(<ClipGridCard clip={clip} />);

    expect(screen.getByRole('button', { name: 'Upvote Accessible highlight' })).toHaveAttribute(
      'aria-pressed',
      'true',
    );
    expect(screen.getByRole('button', { name: 'Remove Accessible highlight from favorites' }))
      .toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByRole('progressbar', { name: 'Watch progress' })).toHaveAttribute(
      'aria-valuenow',
      '50',
    );
  });
});
