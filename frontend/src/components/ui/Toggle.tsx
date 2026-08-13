import React from 'react';
import { cn } from '@/lib/utils';

export interface ToggleProps
    extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'type'> {
    /**
     * Toggle label
     */
    label?: string;
    /**
     * Helper text
     */
    helperText?: string;
}

/**
 * Toggle/Switch component with smooth animation
 */
export const Toggle = React.forwardRef<HTMLInputElement, ToggleProps>(
    (
        { className, label, helperText, id, disabled, checked, ...props },
        ref
    ) => {
        const generatedId = React.useId();
        const toggleId = id || `toggle-${generatedId}`;

        return (
            <div className='flex flex-col gap-1'>
                <div className='flex items-center gap-3'>
                    <label
                        htmlFor={toggleId}
                        className={cn(
                            'relative inline-flex min-h-11 min-w-11 items-center justify-center cursor-pointer',
                            disabled && 'opacity-50 cursor-not-allowed'
                        )}
                    >
                        <input
                            ref={ref}
                            type='checkbox'
                            role='switch'
                            id={toggleId}
                            className='sr-only peer'
                            disabled={disabled}
                            checked={checked}
                            aria-checked={checked ? 'true' : 'false'}
                            aria-describedby={
                                helperText ? `${toggleId}-helper` : undefined
                            }
                            {...props}
                        />
                        <div
                            className={cn(
                                'w-11 h-6 rounded-full transition-colors duration-200',
                                'bg-neutral-300 dark:bg-neutral-600',
                                'peer-checked:bg-primary-500',
                                'peer-focus:ring-2 peer-focus:ring-primary-500 peer-focus:ring-offset-2',
                                'peer-disabled:cursor-not-allowed',
                                className
                            )}
                        >
                            <div
                                className={cn(
                                    'absolute top-0.5 left-0.5 w-5 h-5 rounded-full transition-transform duration-200',
                                    'bg-white shadow-md',
                                    checked && 'translate-x-5'
                                )}
                            />
                        </div>
                    </label>
                    {label && (
                        <label
                            htmlFor={toggleId}
                            className={cn(
                                'text-sm font-medium text-foreground cursor-pointer select-none',
                                disabled && 'opacity-50 cursor-not-allowed'
                            )}
                        >
                            {label}
                        </label>
                    )}
                </div>
                {helperText && (
                    <p
                        id={`${toggleId}-helper`}
                        className='text-sm text-muted-foreground ml-14'
                    >
                        {helperText}
                    </p>
                )}
            </div>
        );
    }
);

Toggle.displayName = 'Toggle';
