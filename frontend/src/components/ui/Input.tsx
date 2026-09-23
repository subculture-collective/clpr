import React from 'react';
import { cn } from '@/lib/utils';

export interface InputProps
    extends React.InputHTMLAttributes<HTMLInputElement> {
    /**
     * Input label
     */
    label?: string;
    /**
     * Error message
     */
    error?: string;
    /**
     * Helper text
     */
    helperText?: string;
    /**
     * Icon to display before the input
     */
    leftIcon?: React.ReactNode;
    /**
     * Icon to display after the input
     */
    rightIcon?: React.ReactNode;
    /**
     * Full width input
     */
    fullWidth?: boolean;
}

/**
 * Input component for text, email, password, and other input types
 */
export const Input = React.forwardRef<HTMLInputElement, InputProps>(
    (
        {
            className,
            label,
            error,
            helperText,
            leftIcon,
            rightIcon,
            fullWidth = false,
            id,
            disabled,
            ...props
        },
        ref
    ) => {
        const generatedId = React.useId();
        const inputId = id || `input-${generatedId}`;

        return (
            <div className={cn('flex flex-col gap-1.5', fullWidth && 'w-full')}>
                {label && (
                    <label
                        htmlFor={inputId}
                        className='kicker'
                    >
                        {label}
                    </label>
                )}
                <div className='relative'>
                    {leftIcon && (
                        <div className='absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground'>
                            {leftIcon}
                        </div>
                    )}
                    <input
                        ref={ref}
                        id={inputId}
                        className={cn(
                            'w-full px-3 py-2.5 rounded-none border transition-colors min-h-[44px]',
                            'bg-surface text-foreground',
                            'placeholder:text-muted-foreground',
                            'focus:outline-none focus:ring-1 focus:ring-primary-400 focus:border-primary-400',
                            'disabled:opacity-50 disabled:cursor-not-allowed',
                            error
                                ? 'border-error-500 focus:ring-error-500'
                                : 'border-line-strong hover:border-text-tertiary',
                            leftIcon && 'pl-10',
                            rightIcon && 'pr-10',
                            className
                        )}
                        disabled={disabled}
                        aria-invalid={error ? 'true' : 'false'}
                        aria-describedby={
                            error
                                ? `${inputId}-error`
                                : helperText
                                ? `${inputId}-helper`
                                : undefined
                        }
                        {...props}
                    />
                    {rightIcon && (
                        <div className='absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground'>
                            {rightIcon}
                        </div>
                    )}
                </div>
                {error && (
                    <p
                        id={`${inputId}-error`}
                        className='text-sm text-error-400'
                    >
                        {error}
                    </p>
                )}
                {helperText && !error && (
                    <p
                        id={`${inputId}-helper`}
                        className='text-sm text-muted-foreground'
                    >
                        {helperText}
                    </p>
                )}
            </div>
        );
    }
);

Input.displayName = 'Input';
