import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { TwitchEmbed } from './TwitchEmbed';

describe('TwitchEmbed', () => {
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
});
