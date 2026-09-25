import { fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { TwitchEmbed } from './TwitchEmbed';

function stubBoxSize(width: number, height: number) {
  vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockReturnValue({
    width,
    height,
  } as DOMRect);
}

describe('TwitchEmbed', () => {
  beforeEach(() => {
    localStorage.clear();
    stubBoxSize(640, 360);
  });
  afterEach(() => vi.restoreAllMocks());

  it('stays a thumbnail until activated by the viewer', () => {
    const onActivate = vi.fn();
    const { rerender } = render(
      <TwitchEmbed
        clipId='GoodClip'
        title='A good clip'
        thumbnailUrl='https://example.com/clip.jpg'
        active={false}
        onActivate={onActivate}
      />,
    );

    expect(screen.queryByTitle('A good clip')).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: /play a good clip/i }));
    expect(onActivate).toHaveBeenCalledTimes(1);

    rerender(
      <TwitchEmbed
        clipId='GoodClip'
        title='A good clip'
        thumbnailUrl='https://example.com/clip.jpg'
        active
        onActivate={onActivate}
      />,
    );

    expect(screen.getByTitle('A good clip')).toHaveAttribute(
      'src',
      expect.stringContaining('clips.twitch.tv/embed'),
    );
    expect(screen.getByTitle('A good clip')).toHaveAttribute(
      'src',
      expect.stringContaining('autoplay=true'),
    );
  });

  it('unmounts the iframe when the feed deactivates the clip', () => {
    const { rerender } = render(
      <TwitchEmbed clipId='GoodClip' title='A good clip' active />,
    );
    expect(screen.getByTitle('A good clip')).toBeInTheDocument();

    rerender(
      <TwitchEmbed clipId='GoodClip' title='A good clip' active={false} />,
    );
    expect(screen.queryByTitle('A good clip')).not.toBeInTheDocument();
  });

  it('keeps the sound control outside the player box', () => {
    render(<TwitchEmbed clipId='GoodClip' title='A good clip' active />);
    const player = screen.getByTitle('A good clip');
    const sound = screen.getByRole('button', { name: /turn on for future clips/i });
    expect(player.parentElement).not.toContainElement(sound);
  });

  it('links to Twitch instead of embedding below the 400×300 minimum', () => {
    stubBoxSize(390, 292);
    render(<TwitchEmbed clipId='GoodClip' title='A good clip' active />);
    expect(screen.queryByTitle('A good clip')).not.toBeInTheDocument();
    expect(screen.getByRole('link', { name: /watch on twitch/i })).toHaveAttribute(
      'href',
      'https://clips.twitch.tv/GoodClip',
    );
    expect(
      screen.queryByRole('button', { name: /turn on for future clips/i }),
    ).not.toBeInTheDocument();
  });
});
