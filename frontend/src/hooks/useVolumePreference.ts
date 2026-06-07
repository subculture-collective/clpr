import { useState, useEffect } from 'react';

// Volume preference key for localStorage
const VOLUME_PREF_KEY = 'clpr_video_muted';

/**
 * Custom hook to manage video volume preference across the application.
 * 
 * @returns {Object} Volume preference state and handlers
 * @returns {boolean} userPrefersUnmuted - Whether user prefers videos unmuted
 * @returns {boolean} hasSetPreference - Whether user has explicitly set a preference
 * @returns {() => void} setUnmutedPreference - Function to set unmuted preference
 * @returns {boolean} embedMuted - Calculated muted state for iframe embed URL
 */
export function useVolumePreference() {
  const [hasSetPreference, setHasSetPreference] = useState(() => {
    // Check if user has ever set a preference
    if (typeof window !== 'undefined') {
      return localStorage.getItem(VOLUME_PREF_KEY) !== null;
    }
    return false;
  });

  const [userPrefersUnmuted, setUserPrefersUnmuted] = useState(() => {
    // Check localStorage for user's volume preference
    // Default to false (start muted first time for compatibility)
    if (typeof window !== 'undefined') {
      const stored = localStorage.getItem(VOLUME_PREF_KEY);
      // stored 'unmuted' means user wants sound, otherwise muted
      return stored === 'unmuted';
    }
    return false;
  });

  // Store volume preference when user changes it
  useEffect(() => {
    if (typeof window !== 'undefined' && hasSetPreference) {
      const newValue = userPrefersUnmuted ? 'unmuted' : 'muted';
      const currentValue = localStorage.getItem(VOLUME_PREF_KEY);
      // Only write if value actually changed
      if (currentValue !== newValue) {
        localStorage.setItem(VOLUME_PREF_KEY, newValue);
      }
    }
  }, [userPrefersUnmuted, hasSetPreference]);

  const setUnmutedPreference = () => {
    setUserPrefersUnmuted(true);
    setHasSetPreference(true);
  };

  // Calculate whether embed should be muted (inverse of user preference)
  const embedMuted = !userPrefersUnmuted;

  return {
    userPrefersUnmuted,
    hasSetPreference,
    setUnmutedPreference,
    embedMuted,
  };
}
