import React from 'react';
import { cn } from '@/lib/utils';

export interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  /**
   * Badge variant
   * @default 'default'
   */
  variant?: 'default' | 'primary' | 'secondary' | 'success' | 'warning' | 'error' | 'info';
  /**
   * Badge size
   * @default 'md'
   */
  size?: 'sm' | 'md' | 'lg';
  /**
   * Icon to display before the badge text
   */
  leftIcon?: React.ReactNode;
  /**
   * Icon to display after the badge text
   */
  rightIcon?: React.ReactNode;
  children: React.ReactNode;
}

const variantClasses = {
  default: 'border border-line-strong text-text-secondary',
  primary: 'border border-primary-700 text-primary-300',
  secondary: 'border border-category/40 text-category',
  success: 'border border-success-800 text-success-400',
  warning: 'border border-warning-800 text-warning-400',
  error: 'border border-error-800 text-error-400',
  info: 'border border-info-800 text-info-400',
};

const sizeClasses = {
  sm: 'px-1.5 py-0.5 text-[11px]',
  md: 'px-2 py-0.5 text-xs',
  lg: 'px-2.5 py-1 text-sm',
};

/**
 * Badge component for displaying status, labels, and tags
 */
export const Badge = React.forwardRef<HTMLSpanElement, BadgeProps>(
  (
    {
      className,
      variant = 'default',
      size = 'md',
      leftIcon,
      rightIcon,
      children,
      ...props
    },
    ref
  ) => {
    return (
      <span
        ref={ref}
        className={cn(
          'inline-flex items-center gap-1 rounded-none font-mono font-medium transition-colors',
          variantClasses[variant],
          sizeClasses[size],
          className
        )}
        {...props}
      >
        {leftIcon && <span className="inline-flex">{leftIcon}</span>}
        {children}
        {rightIcon && <span className="inline-flex">{rightIcon}</span>}
      </span>
    );
  }
);

Badge.displayName = 'Badge';
