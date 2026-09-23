import React from 'react';
import { cn } from '@/lib/utils';

export interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
  /**
   * Card variant
   * @default 'default'
   */
  variant?: 'default' | 'elevated' | 'outlined';
  /**
   * Whether the card is clickable
   */
  clickable?: boolean;
  /**
   * Hover effect
   */
  hover?: boolean;
  children: React.ReactNode;
}

export interface CardHeaderProps extends React.HTMLAttributes<HTMLDivElement> {
  children: React.ReactNode;
}

export interface CardBodyProps extends React.HTMLAttributes<HTMLDivElement> {
  children: React.ReactNode;
}

export interface CardFooterProps extends React.HTMLAttributes<HTMLDivElement> {
  children: React.ReactNode;
}

const variantClasses = {
  default: 'bg-card border border-border',
  elevated: 'bg-surface-raised border border-border',
  outlined: 'bg-transparent border border-line-strong',
};

/**
 * Card component with header, body, and footer sections
 */
export const Card = React.forwardRef<HTMLDivElement, CardProps>(
  ({ className, variant = 'default', clickable = false, hover = false, children, ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={cn(
          'rounded-none transition-colors duration-150',
          variantClasses[variant],
          clickable && 'cursor-pointer',
          hover && 'hover:border-line-strong hover:bg-surface-raised',
          className
        )}
        {...props}
      >
        {children}
      </div>
    );
  }
);

Card.displayName = 'Card';

/**
 * CardHeader component for card header section
 */
export const CardHeader = React.forwardRef<HTMLDivElement, CardHeaderProps>(
  ({ className, children, ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={cn('px-6 py-4 border-b border-border', className)}
        {...props}
      >
        {children}
      </div>
    );
  }
);

CardHeader.displayName = 'CardHeader';

/**
 * CardBody component for card body section
 */
export const CardBody = React.forwardRef<HTMLDivElement, CardBodyProps>(
  ({ className, children, ...props }, ref) => {
    return (
      <div ref={ref} className={cn('px-6 py-4', className)} {...props}>
        {children}
      </div>
    );
  }
);

CardBody.displayName = 'CardBody';

/**
 * CardFooter component for card footer section
 */
export const CardFooter = React.forwardRef<HTMLDivElement, CardFooterProps>(
  ({ className, children, ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={cn('px-6 py-4 border-t border-border', className)}
        {...props}
      >
        {children}
      </div>
    );
  }
);

CardFooter.displayName = 'CardFooter';
