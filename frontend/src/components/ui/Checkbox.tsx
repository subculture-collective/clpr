import React from 'react';
import { cn } from '@/lib/utils';

export interface CheckboxProps
    extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'type'> {
    /**
     * Checkbox label
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
}

/**
 * Checkbox component with custom styling
 */
export const Checkbox = React.forwardRef<HTMLInputElement, CheckboxProps>(
    ({ className, label, error, helperText, id, disabled, ...props }, ref) => {
        const generatedId = React.useId();
        const checkboxId = id || `checkbox-${generatedId}`;

        return (
            <div className='flex flex-col gap-1'>
                <div className='flex items-center gap-2'>
                    <input
                        ref={ref}
                        type='checkbox'
                        id={checkboxId}
                        className={cn(
                            'h-4 w-4 rounded border-2 transition-colors',
                            'text-link focus:ring-2 focus:ring-primary-500 focus:ring-offset-2',
                            'disabled:opacity-50 disabled:cursor-not-allowed',
                            error
                                ? 'border-error-500 focus:ring-error-500'
                                : 'border-border hover:border-primary-400',
                            className
                        )}
                        disabled={disabled}
                        aria-invalid={error ? 'true' : 'false'}
                        aria-describedby={
                            error
                                ? `${checkboxId}-error`
                                : helperText
                                ? `${checkboxId}-helper`
                                : undefined
                        }
                        {...props}
                    />
                    {label && (
                        <label
                            htmlFor={checkboxId}
                            className={cn(
                                'text-sm font-medium text-foreground cursor-pointer select-none',
                                disabled && 'opacity-50 cursor-not-allowed'
                            )}
                        >
                            {label}
                        </label>
                    )}
                </div>
                {error && (
                    <p
                        id={`${checkboxId}-error`}
                        className='text-sm text-error-500 ml-6'
                    >
                        {error}
                    </p>
                )}
                {helperText && !error && (
                    <p
                        id={`${checkboxId}-helper`}
                        className='text-sm text-muted-foreground ml-6'
                    >
                        {helperText}
                    </p>
                )}
            </div>
        );
    }
);

Checkbox.displayName = 'Checkbox';
