import React from 'react';
import { cn } from '@/lib/utils';

export interface SkeletonProps extends React.HTMLAttributes<HTMLDivElement> {
  /**
   * Skeleton variant
   * @default 'rectangular'
   */
  variant?: 'text' | 'circular' | 'rectangular';
  /**
   * Width of skeleton
   */
  width?: string | number;
  /**
   * Height of skeleton
   */
  height?: string | number;
  /**
   * Whether to animate
   * @default true
   */
  animate?: boolean;
}

/**
 * Skeleton component for loading placeholders with animated shimmer
 */
export const Skeleton = React.forwardRef<HTMLDivElement, SkeletonProps>(
  (
    {
      className,
      variant = 'rectangular',
      width,
      height,
      animate = true,
      style,
      ...props
    },
    ref
  ) => {
    const variantClasses = {
      text: 'h-4 rounded-none',
      circular: 'rounded-full',
      rectangular: 'rounded-none',
    };

    return (
      <div
        ref={ref}
        className={cn(
          'bg-surface-raised relative overflow-hidden',
          variantClasses[variant],
          className
        )}
        style={{
          width: width || (variant === 'text' ? '100%' : undefined),
          height: height || (variant === 'text' ? undefined : '100%'),
          ...style,
        }}
        aria-busy="true"
        aria-live="polite"
        {...props}
      >
        {animate && (
          <div
            className="absolute inset-0 -translate-x-full animate-shimmer"
            style={{
              background:
                'linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.2), transparent)',
            }}
          />
        )}
      </div>
    );
  }
);

Skeleton.displayName = 'Skeleton';
