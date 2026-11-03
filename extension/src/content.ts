/**
 * Content script injected into Twitch clip pages.
 *
 * Detects the clip URL from the current page and notifies the background
 * service worker so it can show the action badge and store clip info.
 * Sends `clipInfo: null` when the current page is not a clip, so the
 * background can clear stale badge/state on SPA navigation away.
 */

import { extractTwitchClipInfo } from './lib/twitch';

function notifyBackground(): void {
  const clipInfo = extractTwitchClipInfo(window.location.href) ?? null;

  chrome.runtime.sendMessage({
    type: 'TWITCH_CLIP_DETECTED',
    clipInfo,
  });
}

// Notify on initial load.
notifyBackground();

// Twitch is a SPA. Patching history.pushState here would only affect this
// isolated world, not the page's own calls, so pushState navigations are
// tracked by the background via chrome.tabs.onUpdated (changeInfo.url).
// popstate is dispatched to all worlds and still gives an immediate update
// for back/forward navigation.
window.addEventListener('popstate', notifyBackground);
