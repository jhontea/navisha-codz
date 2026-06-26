import React from 'react';
import { Badge } from '../ui/Badge';
import { Card } from '../ui/Card';

export type ProblemDifficulty = 'easy' | 'medium' | 'hard' | 'expert';
export type ProblemStatus = 'solved' | 'attempted' | 'unsolved';

export interface ProblemCardProps {
  id: string;
  title: string;
  difficulty: ProblemDifficulty;
  category: string;
  tags?: string[];
  status?: ProblemStatus;
  successRate?: number;
  points?: number;
  onClick?: () => void;
  className?: string;
}

const difficultyColors: Record<ProblemDifficulty, string> = {
  easy: '#22c55e',
  medium: '#f59e0b',
  hard: '#ef4444',
  expert: '#8b5cf6',
};

const statusIcons: Record<ProblemStatus, { icon: string; color: string; label: string }> = {
  solved: {
    icon: 'M5 13l4 4L19 7',
    color: 'text-success-500',
    label: 'Solved',
  },
  attempted: {
    icon: 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z',
    color: 'text-warning-500',
    label: 'Attempted',
  },
  unsolved: {
    icon: 'M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z',
    color: 'text-neutral-400',
    label: 'Unsolved',
  },
};

export const ProblemCard: React.FC<ProblemCardProps> = ({
  title,
  difficulty,
  category,
  tags = [],
  status = 'unsolved',
  successRate,
  points,
  onClick,
  className = '',
}) => {
  const statusConfig = statusIcons[status];

  return (
    <Card
      variant="problem"
      padding="none"
      hoverable
      className={`cursor-pointer overflow-hidden ${className}`}
      onClick={onClick}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          onClick?.();
        }
      }}
      aria-label={`Problem: ${title}, ${difficulty}, ${status}`}
    >
      <div className="p-5">
        {/* Header: Status + Difficulty */}
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center gap-2">
            {/* Status Icon */}
            <svg
              className={`w-5 h-5 ${statusConfig.color}`}
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              aria-label={statusConfig.label}
            >
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d={statusConfig.icon} />
            </svg>
            <span className="sr-only">{statusConfig.label}</span>
          </div>
          <Badge variant="difficulty" difficulty={difficulty} size="sm" />
        </div>

        {/* Title */}
        <h3 className="text-lg font-semibold text-neutral-900 dark:text-neutral-100 mb-2 line-clamp-2 group-hover:text-primary-600 dark:group-hover:text-primary-400">
          {title}
        </h3>

        {/* Category */}
        <p className="text-sm text-neutral-500 dark:text-neutral-400 mb-3">
          {category}
        </p>

        {/* Tags */}
        {tags.length > 0 && (
          <div className="flex flex-wrap gap-1.5 mb-3">
            {tags.slice(0, 3).map((tag) => (
              <span
                key={tag}
                className="px-2 py-0.5 text-xs font-medium rounded bg-neutral-100 dark:bg-neutral-800 text-neutral-600 dark:text-neutral-400"
              >
                {tag}
              </span>
            ))}
            {tags.length > 3 && (
              <span className="px-2 py-0.5 text-xs text-neutral-400">
                +{tags.length - 3}
              </span>
            )}
          </div>
        )}

        {/* Footer: Success Rate + Points */}
        <div className="flex items-center justify-between pt-3 border-t border-neutral-100 dark:border-neutral-800">
          {successRate !== undefined && (
            <div className="flex items-center gap-1.5">
              <span className="text-xs text-neutral-500 dark:text-neutral-400">Success:</span>
              <div className="flex items-center gap-1">
                <div className="w-16 h-1.5 bg-neutral-200 dark:bg-neutral-700 rounded-full overflow-hidden">
                  <div
                    className="h-full rounded-full bg-gradient-to-r from-success-400 to-success-500"
                    style={{ width: `${successRate}%` }}
                  />
                </div>
                <span className="text-xs font-medium text-neutral-700 dark:text-neutral-300">
                  {successRate}%
                </span>
              </div>
            </div>
          )}
          {points !== undefined && (
            <span className="text-xs font-semibold text-primary-600 dark:text-primary-400">
              {points} pts
            </span>
          )}
        </div>
      </div>

      {/* Bottom Accent Bar */}
      <div
        className="h-1 transition-all duration-300"
        style={{ backgroundColor: difficultyColors[difficulty] }}
      />
    </Card>
  );
};

ProblemCard.displayName = 'ProblemCard';

export default ProblemCard;
