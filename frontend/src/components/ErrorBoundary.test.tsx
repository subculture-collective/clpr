import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';
import i18n from '@/i18n/config';
import ErrorBoundary from './ErrorBoundary';
import { allowTestConsole } from '../test/setup';

// Component that throws an error for testing
function ThrowError({ shouldThrow }: { shouldThrow: boolean }) {
  if (shouldThrow) {
    throw new Error('Test error');
  }
  return <div>No error</div>;
}

// Helper to wrap ErrorBoundary with i18n provider (ErrorBoundary must be outermost,
// so we can't use AllTheProviders which wraps children in ErrorBoundary's position)
function renderWithI18n(ui: React.ReactElement) {
  return render(
    <I18nextProvider i18n={i18n}>
      {ui}
    </I18nextProvider>
  );
}

describe('ErrorBoundary', () => {
  const allowBoundaryError = () => {
    allowTestConsole('error', /^%o/);
    allowTestConsole('error', /Error caught by boundary:.*Test error/);
  };

  it('renders children when there is no error', () => {
    renderWithI18n(
      <ErrorBoundary>
        <div>Test content</div>
      </ErrorBoundary>
    );
    
    expect(screen.getByText('Test content')).toBeInTheDocument();
  });

  it('renders error UI when an error is thrown', () => {
    allowBoundaryError();
    renderWithI18n(
      <ErrorBoundary>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    );
    
    expect(screen.getByText('Something went wrong')).toBeInTheDocument();
    expect(screen.getByText(/We're sorry, but something unexpected happened/)).toBeInTheDocument();
  });

  it('displays reload and go home buttons', () => {
    allowBoundaryError();
    renderWithI18n(
      <ErrorBoundary>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    );
    
    expect(screen.getByText('Reload Page')).toBeInTheDocument();
    expect(screen.getByText('Go Home')).toBeInTheDocument();
  });

  it('displays contact support link', () => {
    allowBoundaryError();
    renderWithI18n(
      <ErrorBoundary>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    );
    
    expect(screen.getByText('Contact Support')).toBeInTheDocument();
    expect(screen.getByText(/If this problem persists, please contact support/)).toBeInTheDocument();
  });

  it('renders custom fallback when provided', () => {
    allowBoundaryError();
    const customFallback = <div>Custom error message</div>;
    
    renderWithI18n(
      <ErrorBoundary fallback={customFallback}>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    );
    
    expect(screen.getByText('Custom error message')).toBeInTheDocument();
    expect(screen.queryByText('Something went wrong')).not.toBeInTheDocument();
  });
});
