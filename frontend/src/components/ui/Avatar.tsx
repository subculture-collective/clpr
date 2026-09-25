import React from 'react';
import { cn } from '@/lib/utils';

export interface AvatarProps extends React.HTMLAttributes<HTMLDivElement> {
  /**
   * Image source URL
   */
  src?: string | null;
  /**
   * Alt text for image. Pass an empty string when the name is shown beside
   * the avatar, so screen readers do not announce it twice.
   */
  alt?: string;
  /**
   * Fallback text shown when there is no image or it fails to load. Longer
   * names are reduced to their first letter.
   */
  fallback?: string;
  /**
   * Avatar size
   * @default 'md'
   */
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
  /**
   * Classes for the circular frame; use for sizes or borders outside the
   * preset scale (they override the `size` classes).
   */
  frameClassName?: string;
  /**
   * Status indicator
   */
  status?: 'online' | 'offline' | 'away' | 'busy';
}

const sizeClasses = {
  xs: 'h-6 w-6 text-xs',
  sm: 'h-8 w-8 text-sm',
  md: 'h-10 w-10 text-base',
  lg: 'h-12 w-12 text-lg',
  xl: 'h-16 w-16 text-xl',
};

const statusClasses = {
  online: 'bg-success-500',
  offline: 'bg-neutral-400',
  away: 'bg-warning-500',
  busy: 'bg-error-500',
};

function initialOf(...candidates: Array<string | undefined>): string {
  for (const candidate of candidates) {
    const letter = candidate?.trim().replace(/^@/, '').charAt(0);
    if (letter) return letter.toUpperCase();
  }
  return '?';
}

/**
 * Avatar component with image fallback and status indicator
 */
export const Avatar = React.forwardRef<HTMLDivElement, AvatarProps>(
  ({ className, src, alt, fallback, size = 'md', frameClassName, status, ...props }, ref) => {
    // Track the URL that failed rather than a boolean, so a new src gets a fresh attempt.
    const [failedSrc, setFailedSrc] = React.useState<string | null>(null);

    const showImage = !!src && failedSrc !== src;

    return (
      <div
        ref={ref}
        className={cn('relative inline-flex', className)}
        {...props}
      >
        <div
          role={!showImage && alt ? 'img' : undefined}
          aria-label={!showImage && alt ? alt : undefined}
          className={cn(
            'rounded-full overflow-hidden flex items-center justify-center',
            'bg-surface-raised text-link font-heading font-bold uppercase',
            'font-medium',
            sizeClasses[size],
            frameClassName
          )}
        >
          {showImage ? (
            <img
              src={src ?? undefined}
              alt={alt ?? 'Avatar'}
              loading="lazy"
              decoding="async"
              className="h-full w-full object-cover"
              onError={() => setFailedSrc(src ?? null)}
            />
          ) : (
            <span aria-hidden="true" data-testid="avatar-fallback">
              {initialOf(fallback, alt)}
            </span>
          )}
        </div>
        {status && (
          <span
            className={cn(
              'absolute bottom-0 right-0 block rounded-full ring-2 ring-background',
              statusClasses[status],
              size === 'xs' && 'h-1.5 w-1.5',
              size === 'sm' && 'h-2 w-2',
              size === 'md' && 'h-2.5 w-2.5',
              size === 'lg' && 'h-3 w-3',
              size === 'xl' && 'h-4 w-4'
            )}
            aria-label={`Status: ${status}`}
          />
        )}
      </div>
    );
  }
);

Avatar.displayName = 'Avatar';
