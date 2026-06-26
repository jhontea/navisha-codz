import React from 'react';

export interface SpinnerProps {
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
  color?: 'primary' | 'secondary' | 'white' | 'success' | 'error' | 'warning';
  className?: string;
  label?: string;
}

const sizeClasses: Record<string, string> = {
  xs: 'w-3 h-3 border',
  sm: 'w-4 h-4 border-2',
  md: 'w-6 h-6 border-2',
  lg: 'w-8 h-8 border-[3px]',
  xl: 'w-12 h-12 border-4',
};

const colorClasses: Record<string, string> = {
  primary: 'border-primary-200 border-t-primary-600',
  secondary: 'border-secondary-200 border-t-secondary-600',
  white: 'border-white/30 border-t-white',
  success: 'border-success-200 border-t-success-600',
  error: 'border-error-200 border-t-error-600',
  warning: 'border-warning-200 border-t-warning-600',
};

export const Spinner: React.FC<SpinnerProps> = ({
  size = 'md',
  color = 'primary',
  className = '',
  label,
}) => {
  return (
    <span
      role="status"
      aria-label={label || 'Loading'}
      className={`inline-block ${className}`}
    >
      <span
        className={`
          ${sizeClasses[size]}
          ${colorClasses[color]}
          rounded-full animate-spin
          inline-block
        `}
        style={{ borderStyle: 'solid' }}
      />
      {label && (
        <span className="ml-2 text-sm text-neutral-600 dark:text-neutral-400">
          {label}
        </span>
      )}
    </span>
  );
};

Spinner.displayName = 'Spinner';

export default Spinner;
