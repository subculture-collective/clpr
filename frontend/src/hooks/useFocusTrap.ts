import { useEffect, useRef } from 'react';

/**
 * Hook to trap focus within a container element (for modals, dialogs, etc.)
 * @param isActive - Whether the focus trap is active
 * @returns ref to attach to the container element
 */
export function useFocusTrap<T extends HTMLElement>(isActive: boolean) {
  const containerRef = useRef<T>(null);
  const previouslyFocusedElement = useRef<HTMLElement | null>(null);

  useEffect(() => {
    if (!isActive) return;

    // Store the previously focused element
    previouslyFocusedElement.current = document.activeElement as HTMLElement;

    const container = containerRef.current;
    if (!container) return;

    // Get all focusable elements within the container
    const getFocusableElements = () => {
      const focusableSelectors = [
        'a[href]',
        'button:not([disabled])',
        'textarea:not([disabled])',
        'input:not([disabled])',
        'select:not([disabled])',
        '[tabindex]:not([tabindex="-1"])',
      ];
      
      return Array.from(
        container.querySelectorAll<HTMLElement>(focusableSelectors.join(','))
      ).filter(el => {
        // Filter out elements that are not visible
        // Check multiple visibility conditions:
        // 1. offsetParent is null for hidden elements (except position: fixed)
        // 2. Handle fixed position elements by checking computed style
        // 3. Check for visibility: hidden and opacity: 0
        if (el.offsetParent === null) {
          const style = window.getComputedStyle(el);
          // Allow fixed/sticky positioned elements
          if (style.position === 'fixed' || style.position === 'sticky') {
            return style.visibility !== 'hidden' && style.display !== 'none';
          }
          return false;
        }
        const style = window.getComputedStyle(el);
        return style.visibility !== 'hidden' && style.display !== 'none' && style.opacity !== '0';
      });
    };

    // Focus the first focusable element
    const focusableElements = getFocusableElements();
    if (focusableElements.length > 0) {
      focusableElements[0].focus();
    } else {
      container.focus();
    }

    // Handle tab key to trap focus
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key !== 'Tab') return;

      const focusableElements = getFocusableElements();
      if (focusableElements.length === 0) return;

      const firstElement = focusableElements[0];
      const lastElement = focusableElements[focusableElements.length - 1];

      if (e.shiftKey) {
        // Shift + Tab: move to previous element
        if (document.activeElement === firstElement) {
          e.preventDefault();
          lastElement.focus();
        }
      } else {
        // Tab: move to next element
        if (document.activeElement === lastElement) {
          e.preventDefault();
          firstElement.focus();
        }
      }
    };

    container.addEventListener('keydown', handleKeyDown);

    // Cleanup: restore focus to previously focused element
    return () => {
      container.removeEventListener('keydown', handleKeyDown);
      if (previouslyFocusedElement.current?.isConnected) {
        previouslyFocusedElement.current.focus();
      }
    };
  }, [isActive]);

  return containerRef;
}
