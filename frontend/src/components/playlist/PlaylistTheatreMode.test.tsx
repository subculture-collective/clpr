import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { PlaylistTheatreMode } from './PlaylistTheatreMode';

vi.mock('@/components/comment/CommentSection', () => ({ CommentSection: () => null }));
vi.mock('@/components/video', () => ({
  TheatreMode: () => null,
  VideoPlayer: () => <div data-testid='player' />,
}));

const items = [
  {
    id: 'item-1',
    clip_id: 'clip-1',
    clip: { id: 'clip-1', title: 'First clip', embed_url: 'https://clips.twitch.tv/embed?clip=One' },
  },
] as Parameters<typeof PlaylistTheatreMode>[0]['items'];

describe('PlaylistTheatreMode around the Twitch player', () => {
  it('puts the show-sidebar control in its own column, not over the player', () => {
    render(
      <PlaylistTheatreMode title='Mix' items={items} currentItemId='item-1' onItemClick={vi.fn()} contained />,
    );
    fireEvent.click(screen.getByRole('button', { name: 'Hide sidebar' }));
    const toggle = screen.getByRole('button', { name: 'Show playlist' });
    expect(toggle).not.toHaveClass('absolute');
    expect(toggle.parentElement).not.toHaveClass('absolute');
    // The column sits beside the player area rather than inside it.
    expect(toggle.parentElement?.parentElement).toContainElement(screen.getByTestId('player'));
    expect(toggle.parentElement).not.toContainElement(screen.getByTestId('player'));
  });

  it('stops the full-screen theatre above the consent banner', () => {
    const { container } = render(
      <PlaylistTheatreMode title='Mix' items={items} currentItemId='item-1' onItemClick={vi.fn()} />,
    );
    expect(container.firstElementChild).toHaveClass('bottom-[var(--consent-banner-height,0px)]');
    expect(container.firstElementChild).not.toHaveClass('inset-0');
  });
});
