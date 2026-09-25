import { useLayoutEffect, useState, type RefObject } from 'react';

// Twitch requires clip embeds to be at least 400×300.
// https://dev.twitch.tv/docs/embed/video-and-clips/
export const TWITCH_EMBED_MIN_WIDTH = 400;
export const TWITCH_EMBED_MIN_HEIGHT = 300;

/**
 * Whether the element is large enough to hold a Twitch clip embed.
 * Returns null until the element has been measured.
 */
export function useTwitchEmbedFits(ref: RefObject<HTMLElement | null>): boolean | null {
  const [fits, setFits] = useState<boolean | null>(null);

  useLayoutEffect(() => {
    const element = ref.current;
    if (!element) return;
    const measure = () => {
      const { width, height } = element.getBoundingClientRect();
      setFits(
        Math.round(width) >= TWITCH_EMBED_MIN_WIDTH &&
          Math.round(height) >= TWITCH_EMBED_MIN_HEIGHT,
      );
    };
    measure();
    const observer = new ResizeObserver(measure);
    observer.observe(element);
    return () => observer.disconnect();
  }, [ref]);

  return fits;
}
