import React, { useEffect, useState, useCallback } from 'react';

export type ToastVariant = 'success' | 'error' | 'warning' | 'info';
export type ToastPosition = 'top-right' | 'top-left' | 'bottom-right' | 'bottom-left' | 'top-center' | 'bottom-center';

export interface ToastProps {
  id: string;
  variant: ToastVariant;
  message: string;
  title?: string;
  duration?: number; // ms, 0 = no auto-dismiss
  position?: ToastPosition;
  dismissible?: boolean;
  onDismiss: (id: string) => void;
}

const variantConfig: Record<ToastVariant, { icon: string; bg: string; border: string; text: string; iconColor: string }> = {
  success: {
    icon: 'M5 13l4 4L19 7',
    bg: 'bg-success-50 dark:bg-success-950/50',
    border: 'border-success-200 dark:border-success-800',
    text: 'text-success-800 dark:text-success-200',
    iconColor: 'text-success-500',
  },
  error: {
    icon: 'M6 18L18 6M6 6l12 12',
    bg: 'bg-error-50 dark:bg-error-950/50',
    border: 'border-error-200 dark:border-error-800',
    text: 'text-error-800 dark:text-error-200',
    iconColor: 'text-error-500',
  },
  warning: {
    icon: 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z',
    bg: 'bg-warning-50 dark:bg-warning-950/50',
    border: 'border-warning-200 dark:border-warning-800',
    text: 'text-warning-800 dark:text-warning-200',
    iconColor: 'text-warning-500',
  },
  info: {
    icon: 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z',
    bg: 'bg-primary-50 dark:bg-primary-950/50',
    border: 'border-primary-200 dark:border-primary-800',
    text: 'text-primary-800 dark:text-primary-200',
    iconColor: 'text-primary-500',
  },
};

export const Toast: React.FC<ToastProps> = ({
  id,
  variant,
  message,
  title,
  duration = 5000,
  dismissible = true,
  onDismiss,
}) => {
  const [isExiting, setIsExiting] = useState(false);
  const config = variantConfig[variant];

  const handleDismiss = useCallback(() => {
    setIsExiting(true);
    setTimeout(() => onDismiss(id), 200);
  }, [id, onDismiss]);

  useEffect(() => {
    if (duration > 0) {
      const timer = setTimeout(handleDismiss, duration);
      return () => clearTimeout(timer);
    }
  }, [duration, handleDismiss]);

  return (
    <div
      role="alert"
      aria-live="assertive"
      className={`
        flex items-start gap-3 p-4 rounded-lg border shadow-lg
        transform transition-all duration-200 ease-out
        ${config.bg} ${config.border}
        ${isExiting ? 'opacity-0 translate-x-4' : 'opacity-100 translate-x-0'}
      `}
    >
      {/* Icon */}
      <svg className={`w-5 h-5 flex-shrink-0 mt-0.5 ${config.iconColor}`} fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d={config.icon} />
      </svg>

      {/* Content */}
      <div className={`flex-1 text-sm ${config.text}`}>
        {title && <p className="font-semibold mb-0.5">{title}</p>}
        <p>{message}</p>
      </div>

      {/* Dismiss Button */}
      {dismissible && (
        <button
          onClick={handleDismiss}
          className="flex-shrink-0 p-1 rounded hover:bg-neutral-200/50 dark:hover:bg-neutral-700/50 transition-colors"
          aria-label="Dismiss"
        >
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      )}
    </div>
  );
};

Toast.displayName = 'Toast';

// Toast Container
export interface ToastContainerProps {
  toasts: Array<{
    id: string;
    variant: ToastVariant;
    message: string;
    title?: string;
    duration?: number;
    position?: ToastPosition;
  }>;
  onDismiss: (id: string) => void;
  position?: ToastPosition;
}

const positionClasses: Record<ToastPosition, string> = {
  'top-right': 'top-4 right-4 items-end',
  'top-left': 'top-4 left-4 items-start',
  'bottom-right': 'bottom-4 right-4 items-end',
  'bottom-left': 'bottom-4 left-4 items-start',
  'top-center': 'top-4 left-1/2 -translate-x-1/2 items-center',
  'bottom-center': 'bottom-4 left-1/2 -translate-x-1/2 items-center',
};

export const ToastContainer: React.FC<ToastContainerProps> = ({
  toasts,
  onDismiss,
  position = 'top-right',
}) => {
  return (
    <div
      className={`fixed z-toast flex flex-col gap-2 w-full max-w-sm ${positionClasses[position]}`}
      aria-label="Notifications"
    >
      {toasts.map((toast) => (
        <Toast
          key={toast.id}
          {...toast}
          position={position}
          onDismiss={onDismiss}
        />
      ))}
    </div>
  );
};

ToastContainer.displayName = 'ToastContainer';

export default Toast;
