import { useEffect, useRef, useState, useCallback } from 'react';

export type AutoSaveStatus = 'idle' | 'saving' | 'saved' | 'error';

interface UseAutoSaveOptions {
  /** Auto-save interval in milliseconds (default: 30000 = 30 seconds) */
  interval?: number;
  /** Callback to save the content */
  onSave: (content: string) => void | Promise<void>;
  /** Minimum content length to trigger auto-save (default: 1) */
  minLength?: number;
  /** Debounce delay after user stops typing (default: 2000ms) */
  debounceDelay?: number;
}

interface UseAutoSaveReturn {
  status: AutoSaveStatus;
  lastSaved: Date | null;
  forceSave: () => void;
}

/**
 * Hook to auto-save content at regular intervals
 * Saves content to localStorage and optionally to a backend
 */
export function useAutoSave(
  content: string,
  options: UseAutoSaveOptions
): UseAutoSaveReturn {
  const {
    interval = 30000,
    onSave,
    minLength = 1,
    debounceDelay = 2000,
  } = options;

  const [status, setStatus] = useState<AutoSaveStatus>('idle');
  const [lastSaved, setLastSaved] = useState<Date | null>(null);
  const lastContentRef = useRef<string>('');
  const debounceTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const intervalTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const statusTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const isMountedRef = useRef<boolean>(true);

  // Store the latest onSave in a ref to avoid dependency issues
  const onSaveRef = useRef(onSave);
  useEffect(() => {
    onSaveRef.current = onSave;
  }, [onSave]);

  const performSave = useCallback(async () => {
    // Don't save if content is too short or hasn't changed
    if (content.trim().length < minLength || content === lastContentRef.current) {
      return;
    }

    // Only update state if component is still mounted
    if (isMountedRef.current) {
      setStatus('saving');
    }

    try {
      await onSaveRef.current(content);
      lastContentRef.current = content;
      
      if (isMountedRef.current) {
        setStatus('saved');
        setLastSaved(new Date());

        // Clear any existing status timer
        if (statusTimerRef.current) {
          clearTimeout(statusTimerRef.current);
        }
        
        // Reset status after 2 seconds
        statusTimerRef.current = setTimeout(() => {
          if (isMountedRef.current) {
            setStatus('idle');
          }
        }, 2000);
      }
    } catch (error) {
      console.error('Auto-save failed:', error);
      
      if (isMountedRef.current) {
        setStatus('error');
        
        // Clear any existing status timer
        if (statusTimerRef.current) {
          clearTimeout(statusTimerRef.current);
        }
        
        // Reset error status after 3 seconds
        statusTimerRef.current = setTimeout(() => {
          if (isMountedRef.current) {
            setStatus('idle');
          }
        }, 3000);
      }
    }
  }, [content, minLength]);

  const forceSave = useCallback(() => {
    performSave();
  }, [performSave]);

  // Debounced save after user stops typing
  useEffect(() => {
    if (content === lastContentRef.current) {
      return;
    }

    // Clear existing debounce timer
    if (debounceTimerRef.current) {
      clearTimeout(debounceTimerRef.current);
    }

    // Set new debounce timer
    debounceTimerRef.current = setTimeout(() => {
      performSave();
    }, debounceDelay);

    return () => {
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current);
      }
    };
  }, [content, debounceDelay, performSave]);

  // Periodic auto-save
  useEffect(() => {
    intervalTimerRef.current = setInterval(() => {
      performSave();
    }, interval);

    return () => {
      if (intervalTimerRef.current) {
        clearInterval(intervalTimerRef.current);
      }
    };
  }, [interval, performSave]);

  // Track mount status and cleanup all timers on unmount
  useEffect(() => {
    isMountedRef.current = true;
    
    return () => {
      isMountedRef.current = false;
      
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current);
      }
      if (intervalTimerRef.current) {
        clearInterval(intervalTimerRef.current);
      }
      if (statusTimerRef.current) {
        clearTimeout(statusTimerRef.current);
      }
    };
  }, []);

  return {
    status,
    lastSaved,
    forceSave,
  };
}

/**
 * Hook to manage draft storage in localStorage
 */
export function useDraftStorage(key: string) {
  const saveDraft = useCallback((content: string) => {
    if (content.trim().length > 0) {
      try {
        localStorage.setItem(key, content);
      } catch (error) {
        console.error('Failed to save draft to localStorage:', error);
        // Gracefully handle localStorage failures (e.g., quota exceeded, private browsing)
      }
    }
  }, [key]);

  const loadDraft = useCallback((): string | null => {
    try {
      return localStorage.getItem(key);
    } catch (error) {
      console.error('Failed to load draft from localStorage:', error);
      return null;
    }
  }, [key]);

  const clearDraft = useCallback(() => {
    try {
      localStorage.removeItem(key);
    } catch (error) {
      console.error('Failed to clear draft from localStorage:', error);
    }
  }, [key]);

  return {
    saveDraft,
    loadDraft,
    clearDraft,
  };
}
