import React, { useEffect, useId } from 'react';
import { cn } from '@/lib/utils';
import { useFocusTrap } from '@/hooks/useFocusTrap';

export interface ModalProps {
  /**
   * Whether the modal is open
   */
  open: boolean;
  /**
   * Callback when modal should close
   */
  onClose: () => void;
  /**
   * Modal title
   */
  title?: string;
  /**
   * Modal size
   * @default 'md'
   */
  size?: 'sm' | 'md' | 'lg' | 'xl' | 'full';
  /**
   * Whether clicking backdrop closes the modal
   * @default true
   */
  closeOnBackdrop?: boolean;
  /**
   * Whether to show close button
   * @default true
   */
  showCloseButton?: boolean;
  /**
   * Modal content
   */
  children: React.ReactNode;
  /**
   * Additional class name for modal content
   */
  className?: string;
}

const sizeClasses = {
  sm: 'max-w-sm',
  md: 'max-w-md',
  lg: 'max-w-lg',
  xl: 'max-w-xl',
  full: 'max-w-full mx-4',
};

/**
 * Modal component with backdrop overlay and focus trap
 */
export const Modal: React.FC<ModalProps> = ({
  open,
  onClose,
  title,
  size = 'md',
  closeOnBackdrop = true,
  showCloseButton = true,
  children,
  className,
}) => {
  // Unique ID for aria-labelledby
  const titleId = useId();
  // Focus trap for modal
  const modalRef = useFocusTrap<HTMLDivElement>(open);

  // Handle escape key
  useEffect(() => {
    if (!open) return;

    const previousOverflow = document.body.style.overflow;
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose();
      }
    };

    document.addEventListener('keydown', handleEscape);
    document.body.style.overflow = 'hidden';

    return () => {
      document.removeEventListener('keydown', handleEscape);
      document.body.style.overflow = previousOverflow;
    };
  }, [open, onClose]);

  if (!open) return null;

  const handleBackdropClick = (e: React.MouseEvent<HTMLDivElement>) => {
    if (closeOnBackdrop && e.target === e.currentTarget) {
      onClose();
    }
  };

  return (
    <div
      className="fixed inset-0 z-modal flex items-center justify-center p-4"
      onClick={handleBackdropClick}
    >
      {/* Backdrop */}
      <div className="absolute inset-0 bg-black/50 backdrop-blur-sm animate-fade-in" />

      {/* Modal content */}
      <div
        ref={modalRef}
        className={cn(
          'relative w-full bg-card rounded-xl shadow-2xl animate-slide-in-down',
          'flex flex-col max-h-[90vh] overscroll-contain',
          sizeClasses[size],
          className
        )}
        role="dialog"
        aria-modal="true"
        aria-labelledby={title ? titleId : undefined}
      >
        {/* Header */}
        {(title || showCloseButton) && (
          <div className="flex items-center justify-between px-6 py-4 border-b border-border">
            {title && (
              <h2 id={titleId} className="text-xl font-semibold text-foreground">
                {title}
              </h2>
            )}
            {showCloseButton && (
              <button
                onClick={onClose}
                className="ml-auto p-1 rounded-lg hover:bg-muted transition-colors cursor-pointer"
                aria-label="Close modal"
              >
                <svg className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                  <path d="M6.28 5.22a.75.75 0 00-1.06 1.06L8.94 10l-3.72 3.72a.75.75 0 101.06 1.06L10 11.06l3.72 3.72a.75.75 0 101.06-1.06L11.06 10l3.72-3.72a.75.75 0 00-1.06-1.06L10 8.94 6.28 5.22z" />
                </svg>
              </button>
            )}
          </div>
        )}

        {/* Body */}
        <div className="flex-1 overflow-y-auto px-6 py-4 text-foreground">
          {children}
        </div>
      </div>
    </div>
  );
};

Modal.displayName = 'Modal';

export interface ModalFooterProps extends React.HTMLAttributes<HTMLDivElement> {
  children: React.ReactNode;
}

/**
 * ModalFooter component for modal footer actions
 */
export const ModalFooter: React.FC<ModalFooterProps> = ({ className, children, ...props }) => {
  return (
    <div
      className={cn(
        'flex items-center justify-end gap-3 px-6 py-4 border-t border-border',
        className
      )}
      {...props}
    >
      {children}
    </div>
  );
};

ModalFooter.displayName = 'ModalFooter';
