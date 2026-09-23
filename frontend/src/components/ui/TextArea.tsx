import React from 'react';
import { cn } from '@/lib/utils';

export interface TextAreaProps
    extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {
    /**
     * TextArea label
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
     * Full width textarea
     */
    fullWidth?: boolean;
    /**
     * Show character counter
     */
    showCount?: boolean;
    /**
     * Auto-resize based on content
     */
    autoResize?: boolean;
}

/**
 * TextArea component with auto-resize and character counter options
 */
export const TextArea = React.forwardRef<HTMLTextAreaElement, TextAreaProps>(
    (
        {
            className,
            label,
            error,
            helperText,
            fullWidth = false,
            showCount = false,
            autoResize = false,
            id,
            disabled,
            maxLength,
            value,
            onChange,
            ...props
        },
        ref
    ) => {
        const generatedId = React.useId();
        const textareaId = id || `textarea-${generatedId}`;
        const [internalValue, setInternalValue] = React.useState(value || '');
        const textareaRef = React.useRef<HTMLTextAreaElement | null>(null);

        const currentValue = value !== undefined ? value : internalValue;
        const currentLength = String(currentValue).length;

        React.useEffect(() => {
            if (autoResize && textareaRef.current) {
                textareaRef.current.style.height = 'auto';
                textareaRef.current.style.height = `${textareaRef.current.scrollHeight}px`;
            }
        }, [currentValue, autoResize]);

        const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
            if (value === undefined) {
                setInternalValue(e.target.value);
            }
            onChange?.(e);
        };

        const setRefs = (element: HTMLTextAreaElement | null) => {
            textareaRef.current = element;
            if (typeof ref === 'function') {
                ref(element);
            } else if (ref) {
                ref.current = element;
            }
        };

        return (
            <div className={cn('flex flex-col gap-1.5', fullWidth && 'w-full')}>
                {label && (
                    <label
                        htmlFor={textareaId}
                        className='kicker'
                    >
                        {label}
                    </label>
                )}
                <textarea
                    ref={setRefs}
                    id={textareaId}
                    className={cn(
                        'w-full px-3 py-2.5 rounded-none border transition-colors resize-none min-h-[100px]',
                        'bg-surface text-foreground',
                        'placeholder:text-muted-foreground',
                        'focus:outline-none focus:ring-1 focus:ring-primary-400 focus:border-primary-400',
                        'disabled:opacity-50 disabled:cursor-not-allowed',
                        error
                            ? 'border-error-500 focus:ring-error-500'
                            : 'border-line-strong hover:border-text-tertiary',
                        autoResize && 'overflow-hidden',
                        className
                    )}
                    disabled={disabled}
                    maxLength={maxLength}
                    value={value}
                    onChange={handleChange}
                    aria-invalid={error ? 'true' : 'false'}
                    aria-describedby={
                        error
                            ? `${textareaId}-error`
                            : helperText
                            ? `${textareaId}-helper`
                            : undefined
                    }
                    {...props}
                />
                <div className='flex items-center justify-between'>
                    <div className='flex-1'>
                        {error && (
                            <p
                                id={`${textareaId}-error`}
                                className='text-sm text-error-400'
                            >
                                {error}
                            </p>
                        )}
                        {helperText && !error && (
                            <p
                                id={`${textareaId}-helper`}
                                className='text-sm text-muted-foreground'
                            >
                                {helperText}
                            </p>
                        )}
                    </div>
                    {showCount && maxLength && (
                        <span className='text-sm text-muted-foreground'>
                            {currentLength}/{maxLength}
                        </span>
                    )}
                </div>
            </div>
        );
    }
);

TextArea.displayName = 'TextArea';
