/**
 * Twitch URL detection utilities.
 * Identifies Twitch clip URLs and extracts clip identifiers.
 *
 * Supported forms:
 *   1. https://www.twitch.tv/{channel}/clip/{clipId}
 *   2. https://clips.twitch.tv/{clipId}
 *   3. https://clips.twitch.tv/embed?clip={clipId}
 */

import type { TwitchClipInfo } from '../types';

const CLIP_ID_RE = /^[A-Za-z0-9_-]+$/;

/** First path segments on clips.twitch.tv that are pages, not clip slugs. */
const RESERVED_CLIPS_SEGMENTS = new Set(['embed', 'create', 'api', 'clips']);

/** First path segments on twitch.tv that are pages, not channel names. */
const RESERVED_CHANNEL_SEGMENTS = new Set([
  'videos',
  'directory',
  'settings',
  'downloads',
  'search',
  'p',
]);

function isClipId(value: string | null | undefined): value is string {
  return !!value && CLIP_ID_RE.test(value);
}

/**
 * Extracts Twitch clip information from a URL.
 * Returns `null` if the URL is not a valid Twitch clip URL. `clipUrl` is a
 * canonical URL without query string or fragment.
 */
export function extractTwitchClipInfo(url: string): TwitchClipInfo | null {
  let parsed: URL;
  try {
    parsed = new URL(url);
  } catch {
    return null;
  }
  if (parsed.protocol !== 'https:') return null;

  const host = parsed.hostname.toLowerCase();
  const segments = parsed.pathname.split('/').filter(Boolean);

  if (host === 'clips.twitch.tv') {
    const [first] = segments;
    if (first === 'embed') {
      const clipId = parsed.searchParams.get('clip');
      if (!isClipId(clipId)) return null;
      return { clipId, clipUrl: `https://clips.twitch.tv/${clipId}` };
    }
    if (segments.length !== 1 || !isClipId(first) || RESERVED_CLIPS_SEGMENTS.has(first.toLowerCase())) {
      return null;
    }
    return { clipId: first, clipUrl: `https://clips.twitch.tv/${first}` };
  }

  if (host === 'www.twitch.tv' || host === 'twitch.tv') {
    const [channel, marker, clipId] = segments;
    if (
      segments.length !== 3 ||
      marker !== 'clip' ||
      !isClipId(channel) ||
      RESERVED_CHANNEL_SEGMENTS.has(channel.toLowerCase()) ||
      !isClipId(clipId)
    ) {
      return null;
    }
    return {
      clipId,
      clipUrl: `https://www.twitch.tv/${channel}/clip/${clipId}`,
      channel,
    };
  }

  return null;
}

/**
 * Checks whether a URL is a Twitch clip URL.
 */
export function isTwitchClipUrl(url: string): boolean {
  return extractTwitchClipInfo(url) !== null;
}
