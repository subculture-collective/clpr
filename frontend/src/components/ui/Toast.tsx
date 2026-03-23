import React, { useEffect } from 'react';
import { cn } from '@/lib/utils';

export interface ToastAction {
  label: string;
  onClick: () => void;
}

export interface ToastProps {
  id: string;
  /**
   * Toast variant
   * @default 'info'
   */
  variant?: 'success' | 'warning' | 'error' | 'info';
  /**
   * Toast message
   */
  message: string;
  /**
   * Optional action button (e.g. "Undo")
   */
  action?: ToastAction;
  /**
   * Duration in milliseconds before auto-dismiss
   * @default 3000
   */
  duration?: number;
  /**
   * Callback when toast is dismissed
   */
  onDismiss: (id: string) => void;
}

const variantClasses = {
  success: 'text-white',
  warning: 'text-white',
  error: 'text-white',
  info: 'text-white',
};

const variantStyles = {
  success: { backgroundColor: '#16a34a', borderColor: '#15803d' },
  warning: { backgroundColor: '#d97706', borderColor: '#b45309' },
  error: { backgroundColor: '#dc2626', borderColor: '#b91c1c' },
  info: { backgroundColor: '#2563eb', borderColor: '#1d4ed8' },
};

const defaultIcons = {
  success: (
    <svg className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
      <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.857-9.809a.75.75 0 00-1.214-.882l-3.483 4.79-1.88-1.88a.75.75 0 10-1.06 1.061l2.5 2.5a.75.75 0 001.137-.089l4-5.5z" clipRule="evenodd" />
    </svg>
  ),
  warning: (
    <svg className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
      <path fillRule="evenodd" d="M8.485 2.495c.673-1.167 2.357-1.167 3.03 0l6.28 10.875c.673 1.167-.17 2.625-1.516 2.625H3.72c-1.347 0-2.189-1.458-1.515-2.625L8.485 2.495zM10 5a.75.75 0 01.75.75v3.5a.75.75 0 01-1.5 0v-3.5A.75.75 0 0110 5zm0 9a1 1 0 100-2 1 1 0 000 2z" clipRule="evenodd" />
    </svg>
  ),
  error: (
    <svg className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
      <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.28 7.22a.75.75 0 00-1.06 1.06L8.94 10l-1.72 1.72a.75.75 0 101.06 1.06L10 11.06l1.72 1.72a.75.75 0 101.06-1.06L11.06 10l1.72-1.72a.75.75 0 00-1.06-1.06L10 8.94 8.28 7.22z" clipRule="evenodd" />
    </svg>
  ),
  info: (
    <svg className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
      <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a.75.75 0 000 1.5h.253a.25.25 0 01.244.304l-.459 2.066A1.75 1.75 0 0010.747 15H11a.75.75 0 000-1.5h-.253a.25.25 0 01-.244-.304l.459-2.066A1.75 1.75 0 009.253 9H9z" clipRule="evenodd" />
    </svg>
  ),
};

/**
 * Toast component for displaying temporary notifications
 */
export const Toast: React.FC<ToastProps> = ({
  id,
  variant = 'info',
  message,
  action,
  duration = 3000,
  onDismiss,
}) => {
  useEffect(() => {
    const timer = setTimeout(() => {
      onDismiss(id);
    }, action ? Math.max(duration, 5000) : duration);

    return () => clearTimeout(timer);
  }, [id, duration, action, onDismiss]);

  return (
    <div
      className={cn(
        'flex items-center gap-3 rounded-lg px-4 py-3 shadow-lg min-w-[300px] max-w-md border',
        'transition-all duration-200',
        variantClasses[variant]
      )}
      role="alert"
      style={{
        ...variantStyles[variant],
        opacity: 1,
      }}
    >
      <div className="flex-shrink-0">
        {defaultIcons[variant]}
      </div>
      <div className="flex-1 text-sm font-medium">
        {message}
      </div>
      {action && (
        <button
          onClick={() => {
            action.onClick();
            onDismiss(id);
          }}
          className="flex-shrink-0 text-sm font-bold underline underline-offset-2 hover:opacity-80 transition-opacity cursor-pointer"
        >
          {action.label}
        </button>
      )}
      <button
        onClick={() => onDismiss(id)}
        className="flex-shrink-0 hover:opacity-70 transition-opacity cursor-pointer"
        aria-label="Dismiss"
      >
        <svg className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
          <path d="M6.28 5.22a.75.75 0 00-1.06 1.06L8.94 10l-3.72 3.72a.75.75 0 101.06 1.06L10 11.06l3.72 3.72a.75.75 0 101.06-1.06L11.06 10l3.72-3.72a.75.75 0 00-1.06-1.06L10 8.94 6.28 5.22z" />
        </svg>
      </button>
    </div>
  );
};

/**
 * ToastContainer component to hold all active toasts
 */
export const ToastContainer: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  return (
    <div
      className="fixed top-4 right-4 z-50 flex flex-col gap-2 pointer-events-none"
      aria-live="polite"
      aria-atomic="true"
    >
      <div className="pointer-events-auto">
        {children}
      </div>
    </div>
  );
};
