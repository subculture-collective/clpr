import { Component, type ErrorInfo, type ReactNode } from 'react';
import { Translation } from 'react-i18next';
import { captureBoundaryError } from '@/lib/sentry-client';
import { ApiError, ErrorType } from '@/lib/mobile-api-client';

interface Props {
  children: ReactNode;
  fallback?: (error: Error, errorInfo: ErrorInfo) => ReactNode;
  onError?: (error: Error, errorInfo: ErrorInfo) => void;
}

interface State {
  hasError: boolean;
  error: Error | null;
  errorInfo: ErrorInfo | null;
}

class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false, error: null, errorInfo: null };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error, errorInfo: null };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo): void {
    if (import.meta.env.DEV) {
      console.error('Error caught by boundary:', error, errorInfo);
    }

    this.setState({ error, errorInfo });
    captureBoundaryError(error, errorInfo.componentStack);

    if (this.props.onError) {
      this.props.onError(error, errorInfo);
    }
  }

  resetError = () => {
    this.setState({ hasError: false, error: null, errorInfo: null });
  };

  render(): ReactNode {
    if (this.state.hasError && this.state.error) {
      if (this.props.fallback) {
        return this.props.fallback(this.state.error, this.state.errorInfo!);
      }

      return (
        <Translation>
          {(t) => {
            const error = this.state.error!;
            const errorInfo = this.state.errorInfo;
            const isApiError = error instanceof ApiError;
            const title = isApiError
              ? getErrorTitle(error.type)
              : t('error.somethingWentWrong');
            const message = isApiError ? error.userMessage : t('error.unexpectedError');

            return (
              <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-900 px-4">
                <div className="max-w-md w-full bg-white dark:bg-gray-800 rounded-lg shadow-lg p-8 text-center">
                  <div className="text-red-500 mb-4">
                    <svg className="w-16 h-16 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                    </svg>
                  </div>

                  <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-4">{title}</h1>
                  <p className="text-gray-600 dark:text-gray-400 mb-6">{message}</p>

                  {isApiError && error.type === ErrorType.OFFLINE && (
                    <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-md p-3 mb-4">
                      <p className="text-sm text-blue-800 dark:text-blue-300">
                        Your request will be automatically retried when you're back online.
                      </p>
                    </div>
                  )}

                  {import.meta.env.DEV && (
                    <details className="mb-4 text-sm">
                      <summary className="cursor-pointer text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200 mb-2">
                        Technical Details
                      </summary>
                      <pre className="text-xs bg-gray-100 dark:bg-gray-900 p-3 rounded overflow-auto max-h-40 text-red-600 dark:text-red-400">
                        {error.toString()}
                        {errorInfo?.componentStack && (
                          <>{'\n'}{errorInfo.componentStack}</>
                        )}
                      </pre>
                    </details>
                  )}

                  <div className="flex flex-col sm:flex-row gap-3 justify-center">
                    <button
                      onClick={this.resetError}
                      className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors cursor-pointer"
                    >
                      {t('common.reloadPage')}
                    </button>
                    <button
                      onClick={() => window.location.href = '/'}
                      className="px-6 py-2 bg-gray-200 dark:bg-gray-700 text-gray-900 dark:text-white rounded-lg hover:bg-gray-300 dark:hover:bg-gray-600 transition-colors cursor-pointer"
                    >
                      {t('common.goHome')}
                    </button>
                  </div>

                  <div className="mt-6 pt-6 border-t border-gray-200 dark:border-gray-700">
                    <p className="text-sm text-gray-600 dark:text-gray-400 mb-2">
                      {t('error.persistContact')}
                    </p>
                    <a
                      href="/contact"
                      className="inline-flex items-center text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 text-sm font-medium"
                    >
                      <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                      </svg>
                      {t('common.contactSupport')}
                    </a>
                  </div>
                </div>
              </div>
            );
          }}
        </Translation>
      );
    }

    return this.props.children;
  }
}

function getErrorTitle(type: ErrorType): string {
  switch (type) {
    case ErrorType.NETWORK:
      return 'Connection Problem';
    case ErrorType.TIMEOUT:
      return 'Request Timeout';
    case ErrorType.OFFLINE:
      return "You're Offline";
    case ErrorType.AUTH:
      return 'Authentication Required';
    case ErrorType.VALIDATION:
      return 'Invalid Input';
    case ErrorType.SERVER:
      return 'Server Error';
    default:
      return 'Something Went Wrong';
  }
}

export default ErrorBoundary;
