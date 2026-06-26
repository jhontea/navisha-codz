import React from 'react';
import { Spinner } from '../ui/Spinner';

export type SubmissionStatusType = 'pending' | 'running' | 'accepted' | 'wrong_answer' | 'timeout' | 'compilation_error';

export interface SubmissionStatusProps {
  status: SubmissionStatusType;
  progress?: number; // 0-100 for running status
  executionTime?: string; // e.g., "120ms"
  memoryUsage?: string; // e.g., "15.2 MB"
  testCasesPassed?: number;
  totalTestCases?: number;
  className?: string;
}

const statusConfig: Record<SubmissionStatusType, { icon: string; color: string; bgColor: string; label: string }> = {
  pending: {
    icon: 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z',
    color: 'text-neutral-500',
    bgColor: 'bg-neutral-100 dark:bg-neutral-800',
    label: 'Pending',
  },
  running: {
    icon: '',
    color: 'text-primary-500',
    bgColor: 'bg-primary-50 dark:bg-primary-900/20',
    label: 'Running',
  },
  accepted: {
    icon: 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z',
    color: 'text-success-500',
    bgColor: 'bg-success-50 dark:bg-success-900/20',
    label: 'Accepted',
  },
  wrong_answer: {
    icon: 'M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z',
    color: 'text-error-500',
    bgColor: 'bg-error-50 dark:bg-error-900/20',
    label: 'Wrong Answer',
  },
  timeout: {
    icon: 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z',
    color: 'text-warning-500',
    bgColor: 'bg-warning-50 dark:bg-warning-900/20',
    label: 'Time Limit Exceeded',
  },
  compilation_error: {
    icon: 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z',
    color: 'text-error-500',
    bgColor: 'bg-error-50 dark:bg-error-900/20',
    label: 'Compilation Error',
  },
};

export const SubmissionStatus: React.FC<SubmissionStatusProps> = ({
  status,
  progress,
  executionTime,
  memoryUsage,
  testCasesPassed,
  totalTestCases,
  className = '',
}) => {
  const config = statusConfig[status];

  return (
    <div className={`space-y-4 ${className}`}>
      {/* Status Header */}
      <div className={`flex items-center gap-3 p-4 rounded-xl ${config.bgColor}`}>
        <div className={`flex-shrink-0 ${config.color}`}>
          {status === 'running' ? (
            <Spinner size="md" color="primary" />
          ) : (
            <svg className="w-8 h-8" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d={config.icon} />
            </svg>
          )}
        </div>
        <div>
          <h3 className={`text-lg font-semibold ${config.color}`}>
            {config.label}
          </h3>
          {status === 'running' && (
            <p className="text-sm text-neutral-500 dark:text-neutral-400">
              Evaluating your solution...
            </p>
          )}
        </div>
      </div>

      {/* Progress Bar (for running status) */}
      {status === 'running' && progress !== undefined && (
        <div className="space-y-2">
          <div className="flex items-center justify-between text-sm">
            <span className="text-neutral-500 dark:text-neutral-400">Progress</span>
            <span className="font-medium text-neutral-700 dark:text-neutral-300">{progress}%</span>
          </div>
          <div className="w-full h-2 bg-neutral-200 dark:bg-neutral-700 rounded-full overflow-hidden">
            <div
              className="h-full bg-gradient-to-r from-primary-500 to-secondary-500 rounded-full transition-all duration-300"
              style={{ width: `${progress}%` }}
            />
          </div>
        </div>
      )}

      {/* Test Case Progress */}
      {totalTestCases !== undefined && testCasesPassed !== undefined && status !== 'pending' && status !== 'running' && (
        <div className="space-y-2">
          <div className="flex items-center justify-between text-sm">
            <span className="text-neutral-500 dark:text-neutral-400">Test Cases</span>
            <span className={`font-medium ${testCasesPassed === totalTestCases ? 'text-success-500' : 'text-error-500'}`}>
              {testCasesPassed}/{totalTestCases} passed
            </span>
          </div>
          <div className="flex gap-0.5">
            {Array.from({ length: totalTestCases }).map((_, i) => (
              <div
                key={i}
                className={`h-2 flex-1 rounded-full transition-colors ${
                  i < testCasesPassed
                    ? 'bg-success-500'
                    : 'bg-error-300 dark:bg-error-800'
                }`}
              />
            ))}
          </div>
        </div>
      )}

      {/* Execution Stats */}
      {(executionTime || memoryUsage) && (
        <div className="grid grid-cols-2 gap-3">
          {executionTime && (
            <div className="p-3 rounded-lg bg-neutral-50 dark:bg-neutral-800/50 border border-neutral-200 dark:border-neutral-700">
              <p className="text-xs text-neutral-500 dark:text-neutral-400 mb-1">Runtime</p>
              <p className="text-lg font-semibold text-neutral-900 dark:text-neutral-100">
                {executionTime}
              </p>
            </div>
          )}
          {memoryUsage && (
            <div className="p-3 rounded-lg bg-neutral-50 dark:bg-neutral-800/50 border border-neutral-200 dark:border-neutral-700">
              <p className="text-xs text-neutral-500 dark:text-neutral-400 mb-1">Memory</p>
              <p className="text-lg font-semibold text-neutral-900 dark:text-neutral-100">
                {memoryUsage}
              </p>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

SubmissionStatus.displayName = 'SubmissionStatus';

export default SubmissionStatus;
