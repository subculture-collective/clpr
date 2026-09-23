import React from 'react';
import { cn } from '@/lib/utils';

export interface AlertProps extends React.HTMLAttributes<HTMLDivElement> {
    /**
     * Alert variant
     * @default 'info'
     */
    variant?: 'success' | 'warning' | 'error' | 'info';
    /**
     * Alert title
     */
    title?: string;
    /**
     * Whether the alert is dismissible
     */
    dismissible?: boolean;
    /**
     * Callback when alert is dismissed
     */
    onDismiss?: () => void;
    /**
     * Icon to display
     */
    icon?: React.ReactNode;
    children: React.ReactNode;
}

const variantClasses = {
    success: 'bg-surface border-success-800 text-success-300 border-l-[3px] border-l-success-400',
    warning: 'bg-surface border-warning-800 text-warning-300 border-l-[3px] border-l-warning-400',
    error: 'bg-surface border-error-800 text-error-300 border-l-[3px] border-l-error-400',
    info: 'bg-surface border-info-800 text-info-300 border-l-[3px] border-l-info-400',
};

const defaultIcons = {
    success: (
        <svg className='h-5 w-5' viewBox='0 0 20 20' fill='currentColor'>
            <path
                fillRule='evenodd'
                d='M10 18a8 8 0 100-16 8 8 0 000 16zm3.857-9.809a.75.75 0 00-1.214-.882l-3.483 4.79-1.88-1.88a.75.75 0 10-1.06 1.061l2.5 2.5a.75.75 0 001.137-.089l4-5.5z'
                clipRule='evenodd'
            />
        </svg>
    ),
    warning: (
        <svg className='h-5 w-5' viewBox='0 0 20 20' fill='currentColor'>
            <path
                fillRule='evenodd'
                d='M8.485 2.495c.673-1.167 2.357-1.167 3.03 0l6.28 10.875c.673 1.167-.17 2.625-1.516 2.625H3.72c-1.347 0-2.189-1.458-1.515-2.625L8.485 2.495zM10 5a.75.75 0 01.75.75v3.5a.75.75 0 01-1.5 0v-3.5A.75.75 0 0110 5zm0 9a1 1 0 100-2 1 1 0 000 2z'
                clipRule='evenodd'
            />
        </svg>
    ),
    error: (
        <svg className='h-5 w-5' viewBox='0 0 20 20' fill='currentColor'>
            <path
                fillRule='evenodd'
                d='M10 18a8 8 0 100-16 8 8 0 000 16zM8.28 7.22a.75.75 0 00-1.06 1.06L8.94 10l-1.72 1.72a.75.75 0 101.06 1.06L10 11.06l1.72 1.72a.75.75 0 101.06-1.06L11.06 10l1.72-1.72a.75.75 0 00-1.06-1.06L10 8.94 8.28 7.22z'
                clipRule='evenodd'
            />
        </svg>
    ),
    info: (
        <svg className='h-5 w-5' viewBox='0 0 20 20' fill='currentColor'>
            <path
                fillRule='evenodd'
                d='M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a.75.75 0 000 1.5h.253a.25.25 0 01.244.304l-.459 2.066A1.75 1.75 0 0010.747 15H11a.75.75 0 000-1.5h-.253a.25.25 0 01-.244-.304l.459-2.066A1.75 1.75 0 009.253 9H9z'
                clipRule='evenodd'
            />
        </svg>
    ),
};

/**
 * Alert component for displaying feedback messages
 */
export const Alert = React.forwardRef<HTMLDivElement, AlertProps>(
    (
        {
            className,
            variant = 'info',
            title,
            dismissible = false,
            onDismiss,
            icon,
            children,
            ...props
        },
        ref
    ) => {
        return (
            <div
                ref={ref}
                className={cn(
                    'rounded-none border p-4 transition-colors duration-150',
                    variantClasses[variant],
                    className
                )}
                role='alert'
                {...props}
            >
                <div className='flex gap-3'>
                    <div className='shrink-0'>
                        {icon || defaultIcons[variant]}
                    </div>
                    <div className='flex-1'>
                        {title && (
                            <h3 className='font-semibold mb-1'>{title}</h3>
                        )}
                        <div className='text-sm'>{children}</div>
                    </div>
                    {dismissible && (
                        <button
                            onClick={onDismiss}
                            className='shrink-0 ml-auto hover:opacity-70 transition-opacity'
                            aria-label='Dismiss'
                        >
                            <svg
                                className='h-5 w-5'
                                viewBox='0 0 20 20'
                                fill='currentColor'
                            >
                                <path d='M6.28 5.22a.75.75 0 00-1.06 1.06L8.94 10l-3.72 3.72a.75.75 0 101.06 1.06L10 11.06l3.72 3.72a.75.75 0 101.06-1.06L11.06 10l3.72-3.72a.75.75 0 00-1.06-1.06L10 8.94 6.28 5.22z' />
                            </svg>
                        </button>
                    )}
                </div>
            </div>
        );
    }
);

Alert.displayName = 'Alert';
