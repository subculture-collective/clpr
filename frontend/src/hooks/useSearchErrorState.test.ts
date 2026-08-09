import { renderHook, waitFor, act } from '@testing-library/react';
import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest';
import { useSearchErrorState } from './useSearchErrorState';
import { AxiosError, InternalAxiosRequestConfig } from 'axios';
import * as telemetry from '@/lib/telemetry';

// Mock telemetry
vi.mock('@/lib/telemetry', () => ({
  trackEvent: vi.fn(),
}));

function invokeHook<T>(callback: () => T): T {
  let value!: T;
  act(() => {
    value = callback();
  });
  return value;
}

async function advanceRetryTimer(milliseconds: number): Promise<void> {
  await act(async () => {
    vi.advanceTimersByTime(milliseconds);
    await vi.runOnlyPendingTimersAsync();
  });
}

describe('useSearchErrorState', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers({ shouldAdvanceTime: true });
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.useRealTimers();
  });

  it('should initialize with no error state', () => {
    const { result } = renderHook(() => useSearchErrorState());
    
    expect(result.current.errorState).toEqual({
      type: 'none',
      message: undefined,
      retryCount: 0,
      maxRetries: 3,
      isRetrying: false,
      isCircuitOpen: false,
    });
  });

  describe('error detection', () => {
    it('should detect failover via x-search-failover header', async () => {
      const { result } = renderHook(() => useSearchErrorState());
      
      const error: Partial<AxiosError> = {
        response: {
          headers: { 'x-search-failover': 'true' },
          status: 200,
          statusText: 'OK',
          data: {},
          config: {} as unknown as InternalAxiosRequestConfig,
        },
      };
      
      invokeHook(() => result.current.handleSearchError(error));
      
      await waitFor(() => {
        expect(result.current.errorState.type).toBe('failover');
      });
      expect(result.current.errorState.message).toContain('backup search');
    });

    it('should detect failover via x-search-status header', async () => {
      const { result } = renderHook(() => useSearchErrorState());
      
      const error: Partial<AxiosError> = {
        response: {
          headers: { 'x-search-status': 'degraded' },
          status: 200,
          statusText: 'OK',
          data: {},
          config: {} as unknown as InternalAxiosRequestConfig,
        },
      };
      
      invokeHook(() => result.current.handleSearchError(error));
      
      await waitFor(() => {
        expect(result.current.errorState.type).toBe('failover');
      });
    });

    it('should detect complete failure on 503 status', async () => {
      const { result } = renderHook(() => useSearchErrorState());
      
      const error: Partial<AxiosError> = {
        response: {
          headers: {},
          status: 503,
          statusText: 'Service Unavailable',
          data: {},
          config: {} as unknown as InternalAxiosRequestConfig,
        },
      };
      
      invokeHook(() => result.current.handleSearchError(error));
      
      await waitFor(() => {
        expect(result.current.errorState.type).toBe('error');
      });
      expect(result.current.errorState.message).toContain('temporarily unavailable');
    });

    it('should detect complete failure on 504 status', async () => {
      const { result } = renderHook(() => useSearchErrorState());
      
      const error: Partial<AxiosError> = {
        response: {
          headers: {},
          status: 504,
          statusText: 'Gateway Timeout',
          data: {},
          config: {} as unknown as InternalAxiosRequestConfig,
        },
      };
      
      invokeHook(() => result.current.handleSearchError(error));
      
      await waitFor(() => {
        expect(result.current.errorState.type).toBe('error');
      });
    });

    it('should detect network errors', async () => {
      const { result } = renderHook(() => useSearchErrorState());

      const error: Partial<AxiosError> = {
        response: undefined,
        code: 'ERR_NETWORK',
        message: 'Network Error',
      };
      
      invokeHook(() => result.current.handleSearchError(error));
      
      await waitFor(() => {
        expect(result.current.errorState.type).toBe('error');
      });
      expect(result.current.errorState.message).toContain('connection');
    });

    it('should handle unknown errors gracefully', async () => {
      const { result } = renderHook(() => useSearchErrorState());
      
      invokeHook(() => result.current.handleSearchError(new Error('Unknown error')));
      
      await waitFor(() => {
        expect(result.current.errorState.type).toBe('error');
      });
    });

    it('should track analytics event on error', async () => {
      const { result } = renderHook(() => useSearchErrorState());
      
      const error: Partial<AxiosError> = {
        response: {
          headers: {},
          status: 503,
          statusText: 'Service Unavailable',
          data: {},
          config: {} as unknown as InternalAxiosRequestConfig,
        },
      };
      
      invokeHook(() => result.current.handleSearchError(error));
      
      await waitFor(() => {
        expect(telemetry.trackEvent).toHaveBeenCalledWith('search_error', expect.objectContaining({
          error_type: 'error',
          retry_count: 0,
        }));
      });
    });
  });

  describe('retry logic', () => {
    it('should implement exponential backoff', async () => {
      const { result } = renderHook(() => useSearchErrorState());
      const searchFn = vi.fn().mockResolvedValue(undefined);
      
      // Start retry
      const retryPromise = invokeHook(() => result.current.retry(searchFn));
      
      // Wait for retrying state
      await waitFor(() => {
        expect(result.current.errorState.isRetrying).toBe(true);
        expect(result.current.errorState.retryCount).toBe(1);
      });
      
      // Fast-forward 1 second (first delay)
      await advanceRetryTimer(1000);
      
      // Wait for retry to complete
      await retryPromise;
      
      expect(searchFn).toHaveBeenCalledTimes(1);
    });

    it('should stop after max retry attempts', async () => {
      const { result } = renderHook(() => useSearchErrorState());
      const searchFn = vi.fn().mockRejectedValue(new Error('Search failed'));
      
      // First retry
      const firstRetry = invokeHook(() => result.current.retry(searchFn));
      await advanceRetryTimer(1000);
      await act(async () => firstRetry);
      await vi.waitFor(() => expect(result.current.errorState.retryCount).toBe(1));
      
      // Second retry
      const secondRetry = invokeHook(() => result.current.retry(searchFn));
      await advanceRetryTimer(2000);
      await act(async () => secondRetry);
      await vi.waitFor(() => expect(result.current.errorState.retryCount).toBe(2));
      
      // Third retry
      const thirdRetry = invokeHook(() => result.current.retry(searchFn));
      await advanceRetryTimer(4000);
      await act(async () => thirdRetry);
      await vi.waitFor(() => expect(result.current.errorState.retryCount).toBe(3));
      
      // Fourth retry should be blocked
      await act(async () => {
        await result.current.retry(searchFn);
      });
      
      await waitFor(() => {
        expect(result.current.errorState.message).toContain('Maximum retry');
      });
    });

    it('should track retry analytics event', async () => {
      const { result } = renderHook(() => useSearchErrorState());
      const searchFn = vi.fn().mockResolvedValue(undefined);
      
      const retryPromise = invokeHook(() => result.current.retry(searchFn));
      
      expect(telemetry.trackEvent).toHaveBeenCalledWith('search_retry', expect.objectContaining({
        retry_count: 1,
      }));
      
      await advanceRetryTimer(1000);
      await retryPromise;
    });

    it('should clear error on successful retry', async () => {
      const { result } = renderHook(() => useSearchErrorState());
      const searchFn = vi.fn().mockResolvedValue(undefined);
      
      // Set error state first
      const error: Partial<AxiosError> = {
        response: {
          headers: {},
          status: 503,
          statusText: 'Service Unavailable',
          data: {},
          config: {} as unknown as InternalAxiosRequestConfig,
        },
      };
      invokeHook(() => result.current.handleSearchError(error));
      
      // Retry
      const retryPromise = invokeHook(() => result.current.retry(searchFn));
      await advanceRetryTimer(1000);
      await retryPromise;
      
      expect(result.current.errorState.type).toBe('none');
      expect(result.current.errorState.retryCount).toBe(0);
    });

    it('should cleanup timeout on unmount', () => {
      const clearTimeoutSpy = vi.spyOn(globalThis, 'clearTimeout');
      const { result, unmount } = renderHook(() => useSearchErrorState());
      const searchFn = vi.fn().mockResolvedValue(undefined);
      
      invokeHook(() => result.current.retry(searchFn));
      
      act(() => unmount());
      
      expect(clearTimeoutSpy).toHaveBeenCalled();
      clearTimeoutSpy.mockRestore();
    });
  });

  describe('state transitions', () => {
    it('should transition from none to error', async () => {
      const { result } = renderHook(() => useSearchErrorState());
      
      const error: Partial<AxiosError> = {
        response: {
          headers: {},
          status: 503,
          statusText: 'Service Unavailable',
          data: {},
          config: {} as unknown as InternalAxiosRequestConfig,
        },
      };
      
      expect(result.current.errorState.type).toBe('none');
      invokeHook(() => result.current.handleSearchError(error));
      
      await waitFor(() => {
        expect(result.current.errorState.type).toBe('error');
      });
    });

    it('should transition from error to none on success', async () => {
      const { result } = renderHook(() => useSearchErrorState());
      
      const error: Partial<AxiosError> = {
        response: {
          headers: {},
          status: 503,
          statusText: 'Service Unavailable',
          data: {},
          config: {} as unknown as InternalAxiosRequestConfig,
        },
      };
      
      invokeHook(() => result.current.handleSearchError(error));
      
      await waitFor(() => {
        expect(result.current.errorState.type).toBe('error');
      });
      
      act(() => {
        result.current.handleSearchSuccess();
      });
      await waitFor(() => {
        expect(result.current.errorState.type).toBe('none');
      });
    });

    it('should dismiss error', async () => {
      const { result } = renderHook(() => useSearchErrorState());
      
      const error: Partial<AxiosError> = {
        response: {
          headers: {},
          status: 503,
          statusText: 'Service Unavailable',
          data: {},
          config: {} as unknown as InternalAxiosRequestConfig,
        },
      };
      
      invokeHook(() => result.current.handleSearchError(error));
      
      await waitFor(() => {
        expect(result.current.errorState.type).toBe('error');
      });
      
      act(() => {
        result.current.dismissError();
      });
      await waitFor(() => {
        expect(result.current.errorState.type).toBe('none');
      });
    });

    it('should track dismiss analytics event', async () => {
      const { result } = renderHook(() => useSearchErrorState());
      
      const error: Partial<AxiosError> = {
        response: {
          headers: {},
          status: 503,
          statusText: 'Service Unavailable',
          data: {},
          config: {} as unknown as InternalAxiosRequestConfig,
        },
      };
      
      invokeHook(() => result.current.handleSearchError(error));
      await waitFor(() => {
        expect(result.current.errorState.type).toBe('error');
      });
      
      vi.clearAllMocks();

      act(() => {
        result.current.dismissError();
      });

      await waitFor(() => {
        expect(telemetry.trackEvent).toHaveBeenCalledWith('search_error_dismissed', expect.objectContaining({
          error_type: 'error',
        }));
      });
    });
  });

  describe('handleSearchSuccess', () => {
    it('should clear error state', async () => {
      const { result } = renderHook(() => useSearchErrorState());
      
      const error: Partial<AxiosError> = {
        response: {
          headers: {},
          status: 503,
          statusText: 'Service Unavailable',
          data: {},
          config: {} as unknown as InternalAxiosRequestConfig,
        },
      };
      
      invokeHook(() => result.current.handleSearchError(error));
      await waitFor(() => {
        expect(result.current.errorState.type).toBe('error');
      });
      
      act(() => {
        result.current.handleSearchSuccess();
      });

      await waitFor(() => {
        expect(result.current.errorState).toEqual({
          type: 'none',
          message: undefined,
          retryCount: 0,
          maxRetries: 3,
          isRetrying: false,
          isCircuitOpen: false,
        });
      });
    });

    it('should clear pending retry timeout', async () => {
      const clearTimeoutSpy = vi.spyOn(globalThis, 'clearTimeout');
      const { result } = renderHook(() => useSearchErrorState());
      const searchFn = vi.fn().mockResolvedValue(undefined);
      
      invokeHook(() => result.current.retry(searchFn));
      await waitFor(() => {
        expect(result.current.errorState.retryCount).toBe(1);
      });
      
      invokeHook(() => result.current.handleSearchSuccess());
      
      expect(clearTimeoutSpy).toHaveBeenCalled();
      clearTimeoutSpy.mockRestore();
    });
  });

  describe('cancelRetry', () => {
    it('should cancel ongoing retry', async () => {
      const { result } = renderHook(() => useSearchErrorState());
      const searchFn = vi.fn().mockResolvedValue(undefined);
      
      // Start retry
      invokeHook(() => result.current.retry(searchFn));
      await waitFor(() => {
        expect(result.current.errorState.isRetrying).toBe(true);
      });
      
      // Cancel retry
      act(() => {
        result.current.cancelRetry();
      });

      await waitFor(() => {
        expect(result.current.errorState.isRetrying).toBe(false);
        expect(result.current.errorState.message).toContain('cancelled');
      });
    });

    it('should track cancel analytics event', async () => {
      const { result } = renderHook(() => useSearchErrorState());
      const searchFn = vi.fn().mockResolvedValue(undefined);
      
      invokeHook(() => result.current.retry(searchFn));
      vi.clearAllMocks();

      act(() => {
        result.current.cancelRetry();
      });

      await waitFor(() => {
        expect(telemetry.trackEvent).toHaveBeenCalledWith('search_retry_cancelled', expect.objectContaining({
          retry_count: expect.any(Number),
        }));
      });
    });
  });

  describe('circuit breaker', () => {
    it('should open circuit breaker after consecutive failures', async () => {
      const { result } = renderHook(() => useSearchErrorState());
      
      const error: Partial<AxiosError> = {
        response: {
          headers: {},
          status: 503,
          statusText: 'Service Unavailable',
          data: {},
          config: {} as unknown as InternalAxiosRequestConfig,
        },
      };
      
      // Trigger 5 consecutive failures
      for (let i = 0; i < 5; i++) {
        invokeHook(() => result.current.handleSearchError(error));
        await waitFor(() => {
          expect(result.current.errorState.type).toBe('error');
        });
      }
      
      // Circuit breaker should be open
      expect(result.current.errorState.isCircuitOpen).toBe(true);
      expect(result.current.errorState.message).toContain('paused');
    });

    it('should track circuit breaker open event', async () => {
      const { result } = renderHook(() => useSearchErrorState());
      
      const error: Partial<AxiosError> = {
        response: {
          headers: {},
          status: 503,
          statusText: 'Service Unavailable',
          data: {},
          config: {} as unknown as InternalAxiosRequestConfig,
        },
      };
      
      vi.clearAllMocks();
      
      // Trigger 5 consecutive failures
      for (let i = 0; i < 5; i++) {
        invokeHook(() => result.current.handleSearchError(error));
      }
      
      await waitFor(() => {
        expect(telemetry.trackEvent).toHaveBeenCalledWith('search_circuit_breaker_opened', expect.any(Object));
      });
    });

    it('should close circuit breaker after timeout', async () => {
      const { result } = renderHook(() => useSearchErrorState());
      
      const error: Partial<AxiosError> = {
        response: {
          headers: {},
          status: 503,
          statusText: 'Service Unavailable',
          data: {},
          config: {} as unknown as InternalAxiosRequestConfig,
        },
      };
      
      // Trigger 5 consecutive failures to open circuit
      for (let i = 0; i < 5; i++) {
        invokeHook(() => result.current.handleSearchError(error));
      }
      
      await waitFor(() => {
        expect(result.current.errorState.isCircuitOpen).toBe(true);
      });
      
      // Fast-forward 30 seconds
      await act(async () => {
        vi.advanceTimersByTime(30000);
        await vi.runOnlyPendingTimersAsync();
      });

      await waitFor(() => {
        expect(result.current.errorState.isCircuitOpen).toBe(false);
      });
    });

    it('should reset consecutive failures on success', () => {
      const { result } = renderHook(() => useSearchErrorState());
      
      const error: Partial<AxiosError> = {
        response: {
          headers: {},
          status: 503,
          statusText: 'Service Unavailable',
          data: {},
          config: {} as unknown as InternalAxiosRequestConfig,
        },
      };
      
      // Trigger 3 failures
      act(() => {
        for (let i = 0; i < 3; i++) {
          result.current.handleSearchError(error);
        }

        // Success should reset counter
        result.current.handleSearchSuccess();

        // Another 3 failures should not open circuit (would need 5)
        for (let i = 0; i < 3; i++) {
          result.current.handleSearchError(error);
        }
      });
      
      expect(result.current.errorState.isCircuitOpen).toBe(false);
    });

    it('should count failover errors toward circuit breaker', async () => {
      const { result } = renderHook(() => useSearchErrorState());
      
      const failoverError: Partial<AxiosError> = {
        response: {
          headers: { 'x-search-failover': 'true' },
          status: 200,
          statusText: 'OK',
          data: {},
          config: {} as unknown as InternalAxiosRequestConfig,
        },
      };
      
      const error: Partial<AxiosError> = {
        response: {
          headers: {},
          status: 503,
          statusText: 'Service Unavailable',
          data: {},
          config: {} as unknown as InternalAxiosRequestConfig,
        },
      };
      
      // Mix of failover (3) and error (2) = 5 total
      act(() => {
        result.current.handleSearchError(failoverError);
        result.current.handleSearchError(error);
        result.current.handleSearchError(failoverError);
        result.current.handleSearchError(failoverError);
        result.current.handleSearchError(error);
      });
      
      await waitFor(() => {
        expect(result.current.errorState.isCircuitOpen).toBe(true);
      });
    });
  });
});
