import { useState, useCallback, useRef, useEffect } from 'react';
import { AxiosError } from 'axios';
import { trackEvent } from '@/lib/telemetry';

export type SearchErrorType = 'failover' | 'error' | 'none';

export interface SearchErrorState {
  type: SearchErrorType;
  message?: string;
  retryCount: number;
  maxRetries: number;
  isRetrying: boolean;
  isCircuitOpen: boolean;
}

export interface UseSearchErrorStateReturn {
  errorState: SearchErrorState;
  handleSearchError: (error: unknown, options?: { autoRetry?: boolean }) => void;
  handleSearchSuccess: () => void;
  retry: (searchFn: () => Promise<void>) => Promise<void>;
  cancelRetry: () => void;
  dismissError: () => void;
}

const MAX_RETRY_ATTEMPTS = 3;
const RETRY_DELAYS = [1000, 2000, 4000]; // Exponential backoff: 1s, 2s, 4s
const CIRCUIT_BREAKER_THRESHOLD = 5; // Open circuit after 5 consecutive failures
const CIRCUIT_BREAKER_TIMEOUT = 30000; // Try to close circuit after 30 seconds

/**
 * Custom hook to manage search error states and retry logic
 *
 * Features:
 * - Detects failover vs complete failure from API responses
 * - Implements exponential backoff for retries
 * - Tracks retry attempts with visible count
 * - Automatic retry on errors (configurable)
 * - Circuit breaker pattern for service health
 * - Cancel retry functionality
 * - Provides error state management with recovery guidance
 *
 * @example
 * ```tsx
 * const { errorState, handleSearchError, cancelRetry, dismissError } = useSearchErrorState();
 *
 * const performSearch = async () => {
 *   try {
 *     const result = await searchApi.search(params);
 *     handleSearchSuccess();
 *   } catch (error) {
 *     handleSearchError(error, { autoRetry: true }); // Enable automatic retry
 *   }
 * };
 * ```
 */
export function useSearchErrorState(): UseSearchErrorStateReturn {
  const [errorState, setErrorState] = useState<SearchErrorState>({
    type: 'none',
    message: undefined,
    retryCount: 0,
    maxRetries: MAX_RETRY_ATTEMPTS,
    isRetrying: false,
    isCircuitOpen: false,
  });

  const retryTimeoutRef = useRef<NodeJS.Timeout>();
  const circuitBreakerTimeoutRef = useRef<NodeJS.Timeout>();
  const consecutiveFailuresRef = useRef(0);
  const pendingRetryRef = useRef<(() => Promise<void>) | null>(null);
  const isCancelledRef = useRef(false);

  // Cleanup timeouts on unmount
  useEffect(() => {
    return () => {
      if (retryTimeoutRef.current) {
        clearTimeout(retryTimeoutRef.current);
      }
      if (circuitBreakerTimeoutRef.current) {
        clearTimeout(circuitBreakerTimeoutRef.current);
      }
    };
  }, []);

  /**
   * Analyze error to determine if it's a failover or complete failure
   */
  const analyzeError = useCallback((error: unknown): { type: SearchErrorType; message?: string } => {
    if (!error) {
      return { type: 'none' };
    }

    // Check if it's an Axios error with response
    if (typeof error === 'object' && error !== null && 'response' in error) {
      const axiosError = error as AxiosError;
      const response = axiosError.response;

      if (response) {
        // Check for failover header from backend
        const failoverHeader = response.headers['x-search-failover'];
        const searchStatus = response.headers['x-search-status'];

        if (failoverHeader === 'true' || searchStatus === 'degraded') {
          return {
            type: 'failover',
            message: "We're using backup search. Results may be limited.",
          };
        }

        // Check status codes
        const status = response.status;

        // Service Unavailable or Gateway Timeout indicates complete failure
        if (status === 503 || status === 504) {
          return {
            type: 'error',
            message: 'Search service is temporarily unavailable.',
          };
        }

        // Other 5xx errors also indicate failure
        if (status >= 500) {
          return {
            type: 'error',
            message: 'An error occurred while searching. Please try again.',
          };
        }
      }

      // Network error (no response)
      if (!response && axiosError.code === 'ERR_NETWORK') {
        return {
          type: 'error',
          message: 'Unable to reach search service. Please check your connection.',
        };
      }
    }

    // Default to generic error for unknown errors
    return {
      type: 'error',
      message: 'An unexpected error occurred.',
    };
  }, []);

  /**
   * Open circuit breaker after too many failures
   */
  const openCircuitBreaker = useCallback(() => {
    setErrorState(prev => ({
      ...prev,
      type: 'error',
      isCircuitOpen: true,
      message: 'Search service is experiencing issues. Automatic retries have been paused. We\'ll try again in 30 seconds.',
    }));

    trackEvent('search_circuit_breaker_opened', {
      consecutive_failures: consecutiveFailuresRef.current,
    });

    // Try to close circuit after timeout
    circuitBreakerTimeoutRef.current = setTimeout(() => {
      setErrorState(prev => ({
        ...prev,
        isCircuitOpen: false,
      }));
      consecutiveFailuresRef.current = 0;

      trackEvent('search_circuit_breaker_closed', {});
    }, CIRCUIT_BREAKER_TIMEOUT);
  }, []);

  /**
   * Handle search error and update state accordingly
   */
  const handleSearchError = useCallback((error: unknown) => {
    const { type, message } = analyzeError(error);

    // Increment consecutive failures for circuit breaker
    // Count both 'error' and 'failover' types as they both indicate degraded service
    if (type === 'error' || type === 'failover') {
      consecutiveFailuresRef.current++;

      // Check if we should open circuit breaker
      if (consecutiveFailuresRef.current >= CIRCUIT_BREAKER_THRESHOLD) {
        openCircuitBreaker();
        return;
      }
    }

    setErrorState(prev => {
      // Track analytics event with current retry count
      trackEvent('search_error', {
        error_type: type,
        retry_count: prev.retryCount,
        message,
        consecutive_failures: consecutiveFailuresRef.current,
      });

      return {
        ...prev,
        type,
        message,
        isRetrying: false,
      };
    });
  }, [analyzeError, openCircuitBreaker]);

  /**
   * Handle successful search and clear error state
   */
  const handleSearchSuccess = useCallback(() => {
    // Reset consecutive failures on success
    consecutiveFailuresRef.current = 0;

    // Clear circuit breaker timeout if it exists
    if (circuitBreakerTimeoutRef.current) {
      clearTimeout(circuitBreakerTimeoutRef.current);
      circuitBreakerTimeoutRef.current = undefined;
    }

    setErrorState({
      type: 'none',
      message: undefined,
      retryCount: 0,
      maxRetries: MAX_RETRY_ATTEMPTS,
      isRetrying: false,
      isCircuitOpen: false,
    });

    // Clear any pending retry timeout
    if (retryTimeoutRef.current) {
      clearTimeout(retryTimeoutRef.current);
      retryTimeoutRef.current = undefined;
    }

    // Clear pending retry function
    pendingRetryRef.current = null;
    isCancelledRef.current = false;
  }, []);

  /**
   * Retry search with exponential backoff
   * Supports both manual and automatic retries
   */
  const retry = useCallback(async (searchFn: () => Promise<void>) => {
    // Check if circuit breaker is open by reading from current state
    if (errorState.isCircuitOpen) {
      trackEvent('search_retry_blocked_by_circuit_breaker', {});
      return;
    }

    // Reset cancellation flag
    isCancelledRef.current = false;

    // Store the search function for potential cancellation
    pendingRetryRef.current = searchFn;

    // Read the count from the current render. A value assigned inside a state
    // updater is not available synchronously under concurrent React rendering.
    const currentRetryCount = errorState.retryCount;

    setErrorState(prev => {
      // Check if max retries exceeded
      if (currentRetryCount >= MAX_RETRY_ATTEMPTS) {
        return {
          ...prev,
          type: 'error',
          message: 'Maximum retry attempts reached. The search service may be temporarily unavailable. Please try again later or check your internet connection.',
          isRetrying: false,
        };
      }

      // Track analytics event
      trackEvent('search_retry', {
        retry_count: currentRetryCount + 1,
        max_retries: MAX_RETRY_ATTEMPTS,
      });

      // Set retrying state
      return {
        ...prev,
        isRetrying: true,
        retryCount: currentRetryCount + 1,
      };
    });

    // Stop if max retries reached
    if (currentRetryCount >= MAX_RETRY_ATTEMPTS) {
      return;
    }

    // Apply exponential backoff delay
    const delay = RETRY_DELAYS[Math.min(currentRetryCount, RETRY_DELAYS.length - 1)];

    await new Promise(resolve => {
      retryTimeoutRef.current = setTimeout(resolve, delay);
    });

    // Check if retry was cancelled during the delay
    if (isCancelledRef.current) {
      setErrorState(prev => ({
        ...prev,
        isRetrying: false,
        message: 'Retry cancelled. You can try again manually.',
      }));
      trackEvent('search_retry_cancelled', {
        retry_count: currentRetryCount + 1,
      });
      return;
    }

    // Attempt the search
    try {
      await searchFn();
      handleSearchSuccess();
    } catch (error) {
      handleSearchError(error);
    } finally {
      // Clear pending retry
      pendingRetryRef.current = null;
    }
  }, [errorState.isCircuitOpen, errorState.retryCount, handleSearchError, handleSearchSuccess]);

  /**
   * Cancel ongoing retry
   */
  const cancelRetry = useCallback(() => {
    // Set cancellation flag
    isCancelledRef.current = true;

    // Clear retry timeout
    if (retryTimeoutRef.current) {
      clearTimeout(retryTimeoutRef.current);
      retryTimeoutRef.current = undefined;
    }

    // Clear pending retry function
    pendingRetryRef.current = null;

    setErrorState(prev => {
      // Track analytics event inside state update to access current retry count
      trackEvent('search_retry_cancelled', {
        retry_count: prev.retryCount,
      });

      return {
        ...prev,
        isRetrying: false,
        message: 'Retry cancelled. You can try again manually.',
      };
    });
  }, []);

  /**
   * Manually dismiss the error
   */
  const dismissError = useCallback(() => {
    // Cancel any ongoing retry when dismissing
    if (retryTimeoutRef.current) {
      clearTimeout(retryTimeoutRef.current);
      retryTimeoutRef.current = undefined;
    }

    isCancelledRef.current = true;
    pendingRetryRef.current = null;

    setErrorState(prev => {
      // Track analytics event with current state before dismissing
      trackEvent('search_error_dismissed', {
        error_type: prev.type,
        retry_count: prev.retryCount,
      });

      return {
        ...prev,
        type: 'none',
        isRetrying: false,
      };
    });
  }, []);

  return {
    errorState,
    handleSearchError,
    handleSearchSuccess,
    retry,
    cancelRetry,
    dismissError,
  };
}
