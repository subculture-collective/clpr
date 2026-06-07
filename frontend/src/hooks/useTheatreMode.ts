import { useState, useEffect, useCallback, useRef } from 'react';

// Theatre mode preference key for localStorage
const THEATRE_MODE_PREF_KEY = 'clpr_theatre_mode';

/**
 * Custom hook to manage theatre mode state and keyboard shortcuts
 * Persists theatre mode preference in localStorage
 * 
 * @returns Theatre mode state and control functions
 */
export function useTheatreMode() {
  const [isTheatreMode, setIsTheatreMode] = useState(() => {
    // Check localStorage for user's theatre mode preference
    // Default to false for first-time users
    if (typeof window !== 'undefined') {
      try {
        const stored = localStorage.getItem(THEATRE_MODE_PREF_KEY);
        return stored === 'true';
      } catch (error) {
        // localStorage may fail in private browsing mode or when disabled
        console.error('Failed to read theatre mode preference from localStorage:', error);
        return false;
      }
    }
    return false;
  });
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [isPictureInPicture, setIsPictureInPicture] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const videoRef = useRef<HTMLVideoElement>(null);

  // Persist theatre mode preference to localStorage
  useEffect(() => {
    if (typeof window !== 'undefined') {
      try {
        const newValue = String(isTheatreMode);
        const currentValue = localStorage.getItem(THEATRE_MODE_PREF_KEY);
        // Only write if value actually changed
        if (currentValue !== newValue) {
          localStorage.setItem(THEATRE_MODE_PREF_KEY, newValue);
        }
      } catch (error) {
        // localStorage may fail when quota is exceeded or in private browsing
        console.error('Failed to save theatre mode preference to localStorage:', error);
      }
    }
  }, [isTheatreMode]);

  // Handle fullscreen changes
  useEffect(() => {
    const handleFullscreenChange = () => {
      setIsFullscreen(!!document.fullscreenElement);
    };

    document.addEventListener('fullscreenchange', handleFullscreenChange);
    return () => {
      document.removeEventListener('fullscreenchange', handleFullscreenChange);
    };
  }, []);

  // Handle picture-in-picture changes
  useEffect(() => {
    const handlePiPChange = () => {
      setIsPictureInPicture(document.pictureInPictureElement !== null);
    };

    document.addEventListener('enterpictureinpicture', handlePiPChange);
    document.addEventListener('leavepictureinpicture', handlePiPChange);

    return () => {
      document.removeEventListener('enterpictureinpicture', handlePiPChange);
      document.removeEventListener('leavepictureinpicture', handlePiPChange);
    };
  }, []);

  // Toggle theatre mode
  const toggleTheatreMode = useCallback(() => {
    setIsTheatreMode(prev => !prev);
  }, []);

  // Toggle fullscreen
  const toggleFullscreen = useCallback(() => {
    const container = containerRef.current;
    if (!container) return;

    if (!document.fullscreenElement) {
      container.requestFullscreen().catch(err => {
        console.error('Error attempting to enable fullscreen:', err);
      });
    } else {
      document.exitFullscreen();
    }
  }, []);

  // Toggle picture-in-picture
  const togglePictureInPicture = useCallback(async () => {
    const video = videoRef.current;
    if (!video) return;

    try {
      if (document.pictureInPictureElement) {
        await document.exitPictureInPicture();
      } else {
        await video.requestPictureInPicture();
      }
    } catch (err) {
      console.error('Error toggling picture-in-picture:', err);
    }
  }, []);

  // Exit theatre mode
  const exitTheatreMode = useCallback(() => {
    setIsTheatreMode(false);
  }, []);

  return {
    isTheatreMode,
    isFullscreen,
    isPictureInPicture,
    containerRef,
    videoRef,
    toggleTheatreMode,
    toggleFullscreen,
    togglePictureInPicture,
    exitTheatreMode,
  };
}
