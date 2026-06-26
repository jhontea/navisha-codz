import React from 'react';
import { Badge } from '../ui/Badge';
import { Input } from '../ui/Input';
import { Select } from '../ui/Select';

export interface ProblemFiltersProps {
  difficulties?: Array<{ id: string; label: string; color: string }>;
  categories?: Array<{ value: string; label: string }>;
  statuses?: Array<{ value: string; label: string }>;
  selectedDifficulty?: string;
  selectedCategory?: string;
  selectedStatus?: string;
  searchQuery?: string;
  onDifficultyChange?: (difficulty: string) => void;
  onCategoryChange?: (category: string) => void;
  onStatusChange?: (status: string) => void;
  onSearchChange?: (query: string) => void;
  onClearAll?: () => void;
  className?: string;
}

export const ProblemFilters: React.FC<ProblemFiltersProps> = ({
  difficulties = [],
  categories = [],
  statuses = [],
  selectedDifficulty = '',
  selectedCategory = '',
  selectedStatus = '',
  searchQuery = '',
  onDifficultyChange,
  onCategoryChange,
  onStatusChange,
  onSearchChange,
  onClearAll,
  className = '',
}) => {
  const hasActiveFilters = selectedDifficulty || selectedCategory || selectedStatus || searchQuery;

  return (
    <div className={`space-y-4 ${className}`}>
      {/* Search Bar */}
      <div className="w-full">
        <Input
          type="search"
          placeholder="Search problems by title or tag..."
          value={searchQuery}
          onChange={(e) => onSearchChange?.(e.target.value)}
          icon={
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
          }
        />
      </div>

      {/* Filter Row */}
      <div className="flex flex-wrap items-center gap-3">
        {/* Difficulty Filter Buttons */}
        {difficulties.length > 0 && (
          <div className="flex items-center gap-2" role="group" aria-label="Filter by difficulty">
            <span className="text-xs font-medium text-neutral-500 dark:text-neutral-400">Difficulty:</span>
            <div className="flex gap-1">
              <button
                onClick={() => onDifficultyChange?.('')}
                className={`
                  px-3 py-1.5 text-xs font-medium rounded-full transition-colors
                  ${!selectedDifficulty
                    ? 'bg-neutral-900 text-white dark:bg-neutral-100 dark:text-neutral-900'
                    : 'bg-neutral-100 text-neutral-600 hover:bg-neutral-200 dark:bg-neutral-800 dark:text-neutral-400 dark:hover:bg-neutral-700'
                  }
                `}
              >
                All
              </button>
              {difficulties.map((diff) => (
                <button
                  key={diff.id}
                  onClick={() => onDifficultyChange?.(diff.id)}
                  className={`
                    px-3 py-1.5 text-xs font-medium rounded-full transition-colors
                    ${selectedDifficulty === diff.id
                      ? 'text-white'
                      : 'bg-neutral-100 text-neutral-600 hover:bg-neutral-200 dark:bg-neutral-800 dark:text-neutral-400 dark:hover:bg-neutral-700'
                    }
                  `}
                  style={selectedDifficulty === diff.id ? { backgroundColor: diff.color } : {}}
                >
                  {diff.label}
                </button>
              ))}
            </div>
          </div>
        )}

        {/* Status Filter */}
        {statuses.length > 0 && (
          <div className="flex items-center gap-2">
            <span className="text-xs font-medium text-neutral-500 dark:text-neutral-400">Status:</span>
            <div className="flex gap-1">
              <button
                onClick={() => onStatusChange?.('')}
                className={`
                  px-3 py-1.5 text-xs font-medium rounded-full transition-colors
                  ${!selectedStatus
                    ? 'bg-neutral-900 text-white dark:bg-neutral-100 dark:text-neutral-900'
                    : 'bg-neutral-100 text-neutral-600 hover:bg-neutral-200 dark:bg-neutral-800 dark:text-neutral-400 dark:hover:bg-neutral-700'
                  }
                `}
              >
                All
              </button>
              {statuses.map((status) => (
                <button
                  key={status.value}
                  onClick={() => onStatusChange?.(status.value)}
                  className={`
                    px-3 py-1.5 text-xs font-medium rounded-full transition-colors
                    ${selectedStatus === status.value
                      ? 'bg-primary-600 text-white'
                      : 'bg-neutral-100 text-neutral-600 hover:bg-neutral-200 dark:bg-neutral-800 dark:text-neutral-400 dark:hover:bg-neutral-700'
                    }
                  `}
                >
                  {status.label}
                </button>
              ))}
            </div>
          </div>
        )}

        {/* Category Dropdown */}
        {categories.length > 0 && (
          <Select
            options={categories}
            value={selectedCategory}
            onChange={(val) => onCategoryChange?.(val as string)}
            placeholder="All Categories"
            size="sm"
            clearable
            className="w-48"
          />
        )}

        {/* Clear All Button */}
        {hasActiveFilters && onClearAll && (
          <button
            onClick={onClearAll}
            className="ml-auto flex items-center gap-1 px-3 py-1.5 text-xs font-medium text-error-600 dark:text-error-400 hover:bg-error-50 dark:hover:bg-error-950/50 rounded-lg transition-colors"
          >
            <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
            Clear All
          </button>
        )}
      </div>

      {/* Active Filter Badges */}
      {hasActiveFilters && (
        <div className="flex flex-wrap items-center gap-2">
          <span className="text-xs text-neutral-500 dark:text-neutral-400">Active filters:</span>
          {selectedDifficulty && (
            <Badge variant="custom" size="sm" customColor="bg-primary-100 dark:bg-primary-900/30" customLabel={difficulties.find(d => d.id === selectedDifficulty)?.label || selectedDifficulty}>
              <button onClick={() => onDifficultyChange?.('')} className="ml-1 hover:text-error-500">×</button>
            </Badge>
          )}
          {selectedStatus && (
            <Badge variant="custom" size="sm" customColor="bg-secondary-100 dark:bg-secondary-900/30" customLabel={statuses.find(s => s.value === selectedStatus)?.label || selectedStatus}>
              <button onClick={() => onStatusChange?.('')} className="ml-1 hover:text-error-500">×</button>
            </Badge>
          )}
          {selectedCategory && (
            <Badge variant="custom" size="sm" customColor="bg-neutral-200 dark:bg-neutral-700" customLabel={categories.find(c => c.value === selectedCategory)?.label || selectedCategory}>
              <button onClick={() => onCategoryChange?.('')} className="ml-1 hover:text-error-500">×</button>
            </Badge>
          )}
        </div>
      )}
    </div>
  );
};

ProblemFilters.displayName = 'ProblemFilters';

export default ProblemFilters;
