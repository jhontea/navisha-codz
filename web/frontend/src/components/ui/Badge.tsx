import React from 'react';

export type DifficultyLevel = 'easy' | 'medium' | 'hard' | 'expert';
export type SubmissionStatus = 'pending' | 'running' | 'accepted' | 'wrong_answer' | 'timeout' | 'compilation_error';
export type BadgeSize = 'sm' | 'md' | 'lg';

export interface BadgeProps {
  variant?: 'difficulty' | 'status' | 'custom';
  difficulty?: DifficultyLevel;
  status?: SubmissionStatus;
  customColor?: string;
  customLabel?: string;
  size?: BadgeSize;
  rounded?: boolean;
  className?: string;
  children?: React.ReactNode;
}

const difficultyConfig: Record<DifficultyLevel, { bg: string; text: string; label: string }> = {
  easy: { bg: 'bg-success-100 dark:bg-success-900/30', text: 'text-success-700 dark:text-success-400', label: 'Easy' },
  medium: { bg: 'bg-warning-100 dark:bg-warning-900/30', text: 'text-warning-700 dark:text-warning-400', label: 'Medium' },
  hard: { bg: 'bg-error-100 dark:bg-error-900/30', text: 'text-error-700 dark:text-error-400', label: 'Hard' },
  expert: { bg: 'bg-secondary-100 dark:bg-secondary-900/30', text: 'text-secondary-700 dark:text-secondary-400', label: 'Expert' },
};

const statusConfig: Record<SubmissionStatus, { bg: string; text: string; label: string }> = {
  pending: { bg: 'bg-neutral-100 dark:bg-neutral-800', text: 'text-neutral-600 dark:text-neutral-400', label: 'Pending' },
  running: { bg: 'bg-primary-100 dark:bg-primary-900/30', text: 'text-primary-600 dark:text-primary-400', label: 'Running' },
  accepted: { bg: 'bg-success-100 dark:bg-success-900/30', text: 'text-success-700 dark:text-success-400', label: 'Accepted' },
  wrong_answer: { bg: 'bg-error-100 dark:bg-error-900/30', text: 'text-error-700 dark:text-error-400', label: 'Wrong Answer' },
  timeout: { bg: 'bg-warning-100 dark:bg-warning-900/30', text: 'text-warning-700 dark:text-warning-400', label: 'Time Limit' },
  compilation_error: { bg: 'bg-error-100 dark:bg-error-900/30', text: 'text-error-700 dark:text-error-400', label: 'Compile Error' },
};

const sizeClasses: Record<BadgeSize, string> = {
  sm: 'px-1.5 py-0.5 text-xs',
  md: 'px-2.5 py-0.5 text-xs',
  lg: 'px-3 py-1 text-sm',
};

export const Badge: React.FC<BadgeProps> = ({
  variant = 'custom',
  difficulty,
  status,
  customColor,
  customLabel,
  size = 'md',
  rounded = false,
  className = '',
  children,
}) => {
  let bgClass = 'bg-neutral-100 dark:bg-neutral-800';
  let textClass = 'text-neutral-700 dark:text-neutral-300';
  let label = '';

  if (variant === 'difficulty' && difficulty) {
    const config = difficultyConfig[difficulty];
    bgClass = config.bg;
    textClass = config.text;
    label = config.label;
  } else if (variant === 'status' && status) {
    const config = statusConfig[status];
    bgClass = config.bg;
    textClass = config.text;
    label = config.label;
  } else if (variant === 'custom' && customColor) {
    bgClass = customColor;
  }

  return (
    <span
      className={`
        inline-flex items-center font-medium
        ${sizeClasses[size]}
        ${rounded ? 'rounded-full' : 'rounded-md'}
        ${bgClass}
        ${textClass}
        ${className}
      `}
    >
      {children || label || customLabel}
    </span>
  );
};

Badge.displayName = 'Badge';

export default Badge;
