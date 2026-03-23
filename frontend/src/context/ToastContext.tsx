import React, { createContext, useContext, useState, useCallback } from 'react';
import { Toast, ToastContainer } from '@/components/ui/Toast';
import type { ToastAction } from '@/components/ui/Toast';

export interface ToastMessage {
  id: string;
  variant: 'success' | 'warning' | 'error' | 'info';
  message: string;
  action?: ToastAction;
  duration?: number;
}

interface ToastOptions {
  duration?: number;
  action?: ToastAction;
}

interface ToastContextType {
  showToast: (message: string, variant?: 'success' | 'warning' | 'error' | 'info', duration?: number) => void;
  success: (message: string, options?: ToastOptions | number) => void;
  error: (message: string, options?: ToastOptions | number) => void;
  warning: (message: string, options?: ToastOptions | number) => void;
  info: (message: string, options?: ToastOptions | number) => void;
}

const ToastContext = createContext<ToastContextType | undefined>(undefined);

function parseOptions(options?: ToastOptions | number): { duration?: number; action?: ToastAction } {
  if (typeof options === 'number') return { duration: options };
  return options || {};
}

export const ToastProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [toasts, setToasts] = useState<ToastMessage[]>([]);

  const removeToast = useCallback((id: string) => {
    setToasts((prevToasts) => prevToasts.filter((toast) => toast.id !== id));
  }, []);

  const showToast = useCallback(
    (message: string, variant: 'success' | 'warning' | 'error' | 'info' = 'info', duration = 3000) => {
      const id = `toast-${Date.now()}-${Math.random()}`;
      const newToast: ToastMessage = { id, variant, message, duration };
      setToasts((prevToasts) => [...prevToasts, newToast]);
    },
    []
  );

  const success = useCallback(
    (message: string, options?: ToastOptions | number) => {
      const { duration, action } = parseOptions(options);
      const id = `toast-${Date.now()}-${Math.random()}`;
      setToasts((prev) => [...prev, { id, variant: 'success', message, duration, action }]);
    },
    []
  );

  const error = useCallback(
    (message: string, options?: ToastOptions | number) => {
      const { duration, action } = parseOptions(options);
      const id = `toast-${Date.now()}-${Math.random()}`;
      setToasts((prev) => [...prev, { id, variant: 'error', message, duration, action }]);
    },
    []
  );

  const warning = useCallback(
    (message: string, options?: ToastOptions | number) => {
      const { duration, action } = parseOptions(options);
      const id = `toast-${Date.now()}-${Math.random()}`;
      setToasts((prev) => [...prev, { id, variant: 'warning', message, duration, action }]);
    },
    []
  );

  const info = useCallback(
    (message: string, options?: ToastOptions | number) => {
      const { duration, action } = parseOptions(options);
      const id = `toast-${Date.now()}-${Math.random()}`;
      setToasts((prev) => [...prev, { id, variant: 'info', message, duration, action }]);
    },
    []
  );

  return (
    <ToastContext.Provider value={{ showToast, success, error, warning, info }}>
      {children}
      <ToastContainer>
        {toasts.map((toast) => (
          <Toast
            key={toast.id}
            id={toast.id}
            variant={toast.variant}
            message={toast.message}
            action={toast.action}
            duration={toast.duration}
            onDismiss={removeToast}
          />
        ))}
      </ToastContainer>
    </ToastContext.Provider>
  );
};

// Export the hook in a separate export to satisfy react-refresh
// eslint-disable-next-line react-refresh/only-export-components
export const useToast = (): ToastContextType => {
  const context = useContext(ToastContext);
  if (context === undefined) {
    throw new Error('useToast must be used within a ToastProvider');
  }
  return context;
};
