import { useCallback, useState } from 'react';

export type FeedAutoplayPreference = 'manual' | 'muted';

const FEED_AUTOPLAY_KEY = 'clpr_feed_autoplay';

export function useFeedAutoplayPreference() {
  const [preference, setPreferenceState] = useState<FeedAutoplayPreference>(() => {
    if (typeof window === 'undefined') return 'manual';
    return localStorage.getItem(FEED_AUTOPLAY_KEY) === 'muted' ? 'muted' : 'manual';
  });

  const setPreference = useCallback((value: FeedAutoplayPreference) => {
    setPreferenceState(value);
    try {
      localStorage.setItem(FEED_AUTOPLAY_KEY, value);
    } catch {
      // The in-memory preference still works when storage is unavailable.
    }
  }, []);

  return { preference, setPreference };
}
