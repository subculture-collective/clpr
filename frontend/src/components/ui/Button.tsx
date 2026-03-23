import { cn } from '@/lib/utils';
import React from 'react';

export interface ButtonProps
    extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    /**
     * Button variant
     * @default 'primary'
     */
    variant?: 'primary' | 'secondary' | 'ghost' | 'danger' | 'outline';
    /**
     * Button size
     * @default 'md'
     */
    size?: 'sm' | 'md' | 'lg';
    /**
     * Loading state
     */
    loading?: boolean;
    /**
     * Icon to display before the button text
     */
    leftIcon?: React.ReactNode;
    /**
     * Icon to display after the button text
     */
    rightIcon?: React.ReactNode;
    /**
     * Full width button
     */
    fullWidth?: boolean;
    children?: React.ReactNode;
}

// Keep class strings separate so Tailwind can scan them properly
const primaryVariant = 'bg-primary-500 text-white hover:bg-primary-600 active:bg-primary-700 dark:bg-primary-600 dark:hover:bg-primary-700 dark:text-white shadow-sm cursor-pointer';
const secondaryVariant = 'bg-secondary-500 text-white hover:bg-secondary-600 active:bg-secondary-700 cursor-pointer';
const ghostVariant = 'bg-transparent hover:bg-neutral-100 active:bg-neutral-200 dark:hover:bg-neutral-800 dark:active:bg-neutral-700 text-foreground cursor-pointer';
const dangerVariant = 'bg-error-500 text-white hover:bg-error-600 active:bg-error-700 cursor-pointer';
const outlineVariant = 'border-2 border-primary-500 text-primary-500 hover:bg-primary-50 dark:hover:bg-primary-950 cursor-pointer';

const variantClasses = {
    primary: primaryVariant,
    secondary: secondaryVariant,
    ghost: ghostVariant,
    danger: dangerVariant,
    outline: outlineVariant,
};

const sizeClasses = {
    sm: 'px-3 py-1.5 text-xs min-h-[36px]',
    md: 'px-4 py-2 text-sm min-h-[40px]',
    lg: 'px-6 py-2.5 text-base min-h-[44px]',
};

/**
 * Button component with multiple variants, sizes, and states
 */
export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
    (
        {
            className,
            variant = 'primary',
            size = 'md',
            loading = false,
            leftIcon,
            rightIcon,
            fullWidth = false,
            disabled,
            children,
            ...props
        },
        ref
    ) => {
        const isDisabled = disabled || loading;

        return (
            <button
                ref={ref}
                className={cn(
                    'inline-flex items-center justify-center gap-2 rounded-lg font-medium transition-all duration-200',
                    'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary-500 focus-visible:ring-offset-2 dark:focus-visible:ring-offset-neutral-900',
                    'disabled:opacity-50 disabled:cursor-not-allowed disabled:pointer-events-none',
                    variantClasses[variant],
                    sizeClasses[size],
                    fullWidth && 'w-full',
                    className
                )}
                disabled={isDisabled}
                {...props}
            >
                {loading && (
                    <svg
                        className='animate-spin h-4 w-4'
                        xmlns='http://www.w3.org/2000/svg'
                        fill='none'
                        viewBox='0 0 24 24'
                    >
                        <circle
                            className='opacity-25'
                            cx='12'
                            cy='12'
                            r='10'
                            stroke='currentColor'
                            strokeWidth='4'
                        />
                        <path
                            className='opacity-75'
                            fill='currentColor'
                            d='M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z'
                        />
                    </svg>
                )}
                {!loading && leftIcon && (
                    <span className='inline-flex'>{leftIcon}</span>
                )}
                {children}
                {!loading && rightIcon && (
                    <span className='inline-flex'>{rightIcon}</span>
                )}
            </button>
        );
    }
);

Button.displayName = 'Button';
