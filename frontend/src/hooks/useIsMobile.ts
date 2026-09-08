import { useState, useEffect, useRef, useCallback } from 'react';

/**
 * Hook to detect if viewport is mobile size
 * @param breakpoint - Breakpoint in pixels (default: 768)
 */
export function useIsMobile(breakpoint = 768): boolean {
  const [isMobile, setIsMobile] = useState(
    () => typeof window !== 'undefined' && window.innerWidth < breakpoint
  );
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

  const checkMobile = useCallback(() => {
    clearTimeout(timeoutRef.current);
    timeoutRef.current = setTimeout(() => {
      setIsMobile(window.innerWidth < breakpoint);
    }, 150);
  }, [breakpoint]);

  useEffect(() => {
    window.addEventListener('resize', checkMobile);
    return () => {
      window.removeEventListener('resize', checkMobile);
      clearTimeout(timeoutRef.current);
    };
  }, [checkMobile]);

  return isMobile;
}
