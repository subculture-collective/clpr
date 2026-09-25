/**
 * Utility functions for handling API errors consistently across the application
 */

interface ApiErrorResponse {
  response?: {
    data?: {
      error?: string;
    };
  };
  message?: string;
}

/**
 * Safely extracts error message from an API error response
 * @param error - The error object from an API call
 * @param fallbackMessage - Default message if error message cannot be extracted
 * @returns A user-friendly error message
 */
export function getErrorMessage(error: unknown, fallbackMessage: string = 'An unexpected error occurred'): string {
  if (!error) {
    return fallbackMessage;
  }

  // Type guard for API error response
  if (isApiErrorResponse(error)) {
    return error.response?.data?.error || error.message || fallbackMessage;
  }

  // Handle Error objects
  if (error instanceof Error) {
    return error.message;
  }

  // Handle string errors
  if (typeof error === 'string') {
    return error;
  }

  return fallbackMessage;
}

/**
 * Type guard to check if an error is an API error response
 */
function isApiErrorResponse(error: unknown): error is ApiErrorResponse {
  return (
    typeof error === 'object' &&
    error != null &&
    ('response' in error || 'message' in error)
  );
}

/**
 * Returns the HTTP status of a failed API request, when the error carries one.
 */
export function getHttpStatus(error: unknown): number | undefined {
  if (typeof error !== 'object' || error === null || !('response' in error)) {
    return undefined;
  }
  const status = (error as { response?: { status?: unknown } }).response?.status;
  return typeof status === 'number' ? status : undefined;
}

/** True when the API reported that the requested resource does not exist. */
export function isNotFoundError(error: unknown): boolean {
  return getHttpStatus(error) === 404;
}

/**
 * Formats an error for logging purposes
 * @param error - The error to format
 * @returns A formatted error string
 */
export function formatErrorForLogging(error: unknown): string {
  if (error instanceof Error) {
    return `${error.name}: ${error.message}\n${error.stack || ''}`;
  }
  
  try {
    return JSON.stringify(error, null, 2);
  } catch {
    return String(error);
  }
}
