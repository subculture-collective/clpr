import { useState, useEffect } from 'react';
import type { VideoQuality } from '@/lib/adaptive-bitrate';

// Quality preference key for localStorage
const QUALITY_PREF_KEY = 'clpr_video_quality';

/**
 * Custom hook to manage video quality preference across the application.
 * Persists user's preferred quality setting in localStorage.
 * 
 * @returns {Object} Quality preference state and handlers
 */
export function useQualityPreference() {
  const [quality, setQuality] = useState<VideoQuality>(() => {
    // Check localStorage for user's quality preference
    // Default to 'auto' for first-time users
    if (typeof window !== 'undefined') {
      const stored = localStorage.getItem(QUALITY_PREF_KEY);
      if (stored && isValidQuality(stored)) {
        return stored as VideoQuality;
      }
    }
    return 'auto';
  });

  // Store quality preference when user changes it
  useEffect(() => {
    if (typeof window !== 'undefined') {
      localStorage.setItem(QUALITY_PREF_KEY, quality);
    }
  }, [quality]);

  return {
    quality,
    setQuality,
  };
}

/**
 * Validate that a string is a valid VideoQuality
 */
function isValidQuality(value: string): boolean {
  return ['480p', '720p', '1080p', '2K', '4K', 'auto'].includes(value);
}
