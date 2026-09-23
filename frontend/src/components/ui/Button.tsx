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
    /** Render styling and behavior onto the single child element. */
    asChild?: boolean;
    children?: React.ReactNode;
}

// Keep class strings separate so Tailwind can scan them properly
const primaryVariant = 'bg-primary-400 text-background hover:bg-primary-300 active:bg-primary-500 cursor-pointer';
const secondaryVariant = 'border border-line-strong bg-transparent text-text-primary hover:border-text-tertiary hover:bg-surface-hover active:bg-surface cursor-pointer';
const ghostVariant = 'bg-transparent hover:bg-surface-hover active:bg-surface-raised text-foreground cursor-pointer';
const dangerVariant = 'bg-error-400 text-background hover:bg-error-300 active:bg-error-500 cursor-pointer';
const outlineVariant = 'border border-primary-400 text-link hover:bg-primary-950 cursor-pointer';

const variantClasses = {
    primary: primaryVariant,
    secondary: secondaryVariant,
    ghost: ghostVariant,
    danger: dangerVariant,
    outline: outlineVariant,
};

const sizeClasses = {
    sm: 'px-3 py-2 text-[13px] min-h-[44px]',
    md: 'px-4 py-2.5 text-[15px] min-h-[44px]',
    lg: 'px-6 py-3 text-[17px] min-h-[48px]',
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
            asChild = false,
            disabled,
            type = 'button',
            children,
            ...props
        },
        ref
    ) => {
        const isDisabled = disabled || loading;
        const buttonClassName = cn(
            'inline-flex items-center justify-center gap-2 rounded-none font-heading font-bold uppercase tracking-[0.06em] transition-[color,background-color,border-color,box-shadow,transform,opacity] duration-150 motion-reduce:transition-none',
            'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-focus-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background',
            'disabled:opacity-50 disabled:cursor-not-allowed disabled:pointer-events-none',
            variantClasses[variant],
            sizeClasses[size],
            fullWidth && 'w-full',
            className
        );

        if (asChild) {
            if (isDisabled) {
                throw new Error('Button asChild does not support disabled or loading state');
            }
            const child = React.Children.only(children) as React.ReactElement<{
                className?: string;
            }>;
            return React.cloneElement(child, {
                className: cn(buttonClassName, child.props.className),
            });
        }

        return (
            <button
                ref={ref}
                type={type}
                aria-busy={loading || undefined}
                className={buttonClassName}
                disabled={isDisabled}
                {...props}
            >
                {loading && (
                    <svg
                        className='animate-spin motion-reduce:animate-none h-4 w-4'
                        aria-hidden='true'
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
