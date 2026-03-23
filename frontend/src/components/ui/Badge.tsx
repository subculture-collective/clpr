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
  default: 'bg-neutral-100 text-neutral-800 dark:bg-neutral-800 dark:text-neutral-100',
  primary: 'bg-primary-100 text-primary-800 dark:bg-primary-900 dark:text-primary-100',
  secondary: 'bg-secondary-100 text-secondary-800 dark:bg-secondary-900 dark:text-secondary-100',
  success: 'bg-success-100 text-success-800 dark:bg-success-900 dark:text-success-100',
  warning: 'bg-warning-100 text-warning-800 dark:bg-warning-900 dark:text-warning-100',
  error: 'bg-error-100 text-error-800 dark:bg-error-900 dark:text-error-100',
  info: 'bg-info-100 text-info-800 dark:bg-info-900 dark:text-info-100',
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
          'inline-flex items-center gap-1 rounded-full font-medium transition-colors',
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
