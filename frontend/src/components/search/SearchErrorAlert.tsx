import { useState, useEffect } from 'react';
import { Alert } from '../ui/Alert';
import { Zap } from 'lucide-react';

export interface SearchErrorAlertProps {
    /**
     * Type of search error/failover state
     */
    type: 'failover' | 'error' | 'none';
    /**
     * Custom error message (optional)
     */
    message?: string;
    /**
     * Current retry count
     */
    retryCount?: number;
    /**
     * Maximum retry attempts
     */
    maxRetries?: number;
    /**
     * Callback when retry button is clicked
     */
    onRetry?: () => void;
    /**
     * Whether retry is in progress
     */
    isRetrying?: boolean;
    /**
     * Callback when cancel retry button is clicked
     */
    onCancelRetry?: () => void;
    /**
     * Callback when alert is dismissed
     */
    onDismiss?: () => void;
    /**
     * Auto-dismiss duration in milliseconds (for failover warnings)
     * @default 10000 for failover, never for errors
     */
    autoDismissMs?: number;
    /**
     * Whether circuit breaker is open
     */
    isCircuitOpen?: boolean;
}

const ERROR_MESSAGES = {
    failover: {
        title: 'Using Backup Search',
        description:
            "We're experiencing issues with our primary search service. We've automatically switched to backup search. Results may be limited.",
    },
    error: {
        title: 'Search Temporarily Unavailable',
        description:
            'Search is currently unavailable. Please try again in a moment.',
    },
    circuit_breaker: {
        title: 'Search Service Unavailable',
        description:
            "The search service is experiencing persistent issues. We've paused automatic retries to prevent overload. We'll try again automatically in 30 seconds.",
    },
};

/**
 * SearchErrorAlert - Displays search failover and error states with retry functionality
 *
 * Handles different error states:
 * - failover: Warning when backup search is being used (auto-dismisses)
 * - error: Error when search is completely unavailable (manual/automatic retry)
 * - none: No error (component hidden)
 * 
 * Features:
 * - Retry count indicator (e.g., "Retry 1/3")
 * - Visual progress indicator during retry
 * - Cancel retry button for automatic retries
 * - Circuit breaker status indication
 * - Clear recovery guidance after max retries
 */
export function SearchErrorAlert({
    type,
    message,
    retryCount = 0,
    maxRetries = 3,
    onRetry,
    isRetrying = false,
    onCancelRetry,
    onDismiss,
    autoDismissMs = 10000,
    isCircuitOpen = false,
}: SearchErrorAlertProps) {
    const [isDismissed, setIsDismissed] = useState(type === 'none');

    // Auto-dismiss failover warnings
    useEffect(() => {
        if (type === 'failover' && autoDismissMs > 0) {
            const timer = setTimeout(() => {
                setIsDismissed(true);
                onDismiss?.();
            }, autoDismissMs);

            return () => clearTimeout(timer);
        }
    }, [type, autoDismissMs, onDismiss]);

    // Reset dismissed when type changes from 'none' to another state
    useEffect(() => {
        setIsDismissed(type === 'none');
    }, [type]);

    if (type === 'none' || isDismissed) {
        return null;
    }

    const handleDismiss = () => {
        setIsDismissed(true);
        onDismiss?.();
    };

    const handleRetry = () => {
        if (onRetry && !isRetrying) {
            onRetry();
        }
    };

    const handleCancelRetry = () => {
        if (onCancelRetry) {
            onCancelRetry();
        }
    };

    const errorConfig = isCircuitOpen 
        ? ERROR_MESSAGES.circuit_breaker 
        : ERROR_MESSAGES[type];
    const displayMessage = message || errorConfig.description;

    // Show retry count if retrying or if count > 0, but hide once max retries are reached
    const showRetryCount = (isRetrying || retryCount > 0) && retryCount < maxRetries;

    // Determine Alert variant - circuit breaker always shows as error
    const alertVariant = isCircuitOpen ? 'error' : (type === 'failover' ? 'warning' : 'error');

    return (
        <div
            className='mb-4 animate-in fade-in slide-in-from-top-2 duration-200'
            data-testid={
                type === 'failover'
                    ? 'search-failover-warning'
                    : 'search-error-alert'
            }
            role='alert'
            aria-live='polite'
        >
            <Alert
                variant={alertVariant}
                title={errorConfig.title}
                dismissible={type === 'failover' && !isCircuitOpen}
                onDismiss={handleDismiss}
            >
                <div className='space-y-3'>
                    <div>
                        <p>{displayMessage}</p>
                        
                        {/* Retry Count Indicator */}
                        {showRetryCount && (
                            <p className='mt-2 text-sm font-medium' data-testid='retry-count-indicator'>
                                {isRetrying ? 'Retrying' : 'Retry'} attempt {retryCount}/{maxRetries}
                            </p>
                        )}

                        {/* Circuit Breaker Status */}
                        {isCircuitOpen && (
                            <p className='mt-2 text-sm text-muted-foreground' data-testid='circuit-breaker-status'>
                                <Zap size={14} strokeWidth={1.75} className='inline mr-1' /> Service protection active - automatic retries paused
                            </p>
                        )}
                    </div>

                    {/* Retry Progress Bar */}
                    {isRetrying && (
                        <div className='w-full bg-surface-raised rounded-full h-2 overflow-hidden' data-testid='retry-progress-bar'>
                            <div 
                                className='h-full bg-primary animate-pulse'
                                style={{ width: `${((retryCount - 1) / maxRetries) * 100}%` }}
                                role='progressbar'
                                aria-valuenow={retryCount - 1}
                                aria-valuemin={0}
                                aria-valuemax={maxRetries}
                                aria-label={`Retry progress: attempt ${retryCount} of ${maxRetries}`}
                            />
                        </div>
                    )}

                    {/* Action Buttons */}
                    {type === 'error' && !isCircuitOpen && (onRetry || (isRetrying && onCancelRetry)) && (
                        <div className='flex gap-2 flex-wrap'>
                        {onRetry && (
                        <button
                            onClick={handleRetry}
                            disabled={isRetrying}
                            className='inline-flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-md bg-surface-raised border border-border hover:bg-surface-hover disabled:opacity-50 disabled:cursor-not-allowed transition-colors'
                            data-testid='retry-search'
                            aria-label={
                                isRetrying
                                    ? 'Retrying search...'
                                    : 'Retry search'
                            }
                        >
                            {isRetrying ? (
                                <>
                                    <svg
                                        className='animate-spin h-4 w-4'
                                        fill='none'
                                        viewBox='0 0 24 24'
                                        aria-hidden='true'
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
                                    <span>Retrying...</span>
                                </>
                            ) : (
                                <>
                                    <svg
                                        className='h-4 w-4'
                                        fill='none'
                                        stroke='currentColor'
                                        viewBox='0 0 24 24'
                                        aria-hidden='true'
                                    >
                                        <path
                                            strokeLinecap='round'
                                            strokeLinejoin='round'
                                            strokeWidth={2}
                                            d='M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15'
                                        />
                                    </svg>
                                    <span>Try Again</span>
                                </>
                            )}
                        </button>
                        )}

                        {/* Cancel Retry Button - shown when retrying */}
                        {isRetrying && onCancelRetry && (
                            <button
                                onClick={handleCancelRetry}
                                className='inline-flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-md bg-surface-raised border border-border hover:bg-surface-hover transition-colors'
                                data-testid='cancel-retry'
                                aria-label='Cancel retry'
                            >
                                <svg
                                    className='h-4 w-4'
                                    fill='none'
                                    stroke='currentColor'
                                    viewBox='0 0 24 24'
                                    aria-hidden='true'
                                >
                                    <path
                                        strokeLinecap='round'
                                        strokeLinejoin='round'
                                        strokeWidth={2}
                                        d='M6 18L18 6M6 6l12 12'
                                    />
                                </svg>
                                <span>Cancel</span>
                            </button>
                        )}
                        </div>
                    )}
                </div>
            </Alert>
        </div>
    );
}
