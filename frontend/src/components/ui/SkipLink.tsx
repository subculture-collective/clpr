import React from 'react';
import { cn } from '@/lib/utils';

export interface SkipLinkProps {
  /**
   * The ID of the element to skip to
   */
  targetId: string;
  /**
   * Label for the skip link
   */
  label: string;
  /**
   * Additional class name
   */
  className?: string;
}

/**
 * SkipLink component for keyboard navigation accessibility
 * Allows users to skip navigation and go directly to main content
 */
export const SkipLink: React.FC<SkipLinkProps> = ({ targetId, label, className }) => {
  const focusTarget = () => {
    const target = document.getElementById(targetId);
    if (target) {
      target.focus();
      // scrollIntoView may not be available in test environment
      if (typeof target.scrollIntoView === 'function') {
        target.scrollIntoView({ behavior: 'smooth', block: 'start' });
      }
    }
  };

  const handleClick = (e: React.MouseEvent<HTMLAnchorElement>) => {
    e.preventDefault();
    focusTarget();
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLAnchorElement>) => {
    if (e.key !== 'Enter') return;
    e.preventDefault();
    focusTarget();
  };

  return (
    <a
      href={`#${targetId}`}
      onClick={handleClick}
      onKeyDown={handleKeyDown}
      className={cn(
        'sr-only focus:not-sr-only',
        'fixed top-4 left-4 z-[9999]',
        'bg-primary-500 text-white',
        'px-4 py-2 rounded-md',
        'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary-500 focus-visible:ring-offset-2',
        'transition-all duration-200',
        className
      )}
    >
      {label}
    </a>
  );
};

SkipLink.displayName = 'SkipLink';
