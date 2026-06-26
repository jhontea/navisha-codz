import React, { useState } from 'react';

export interface SidebarCategory {
  id: string;
  label: string;
  icon?: React.ReactNode;
  count?: number;
}

export interface SidebarProps {
  categories?: SidebarCategory[];
  selectedCategory?: string;
  onCategorySelect?: (id: string) => void;
  difficulties?: Array<{ id: string; label: string; color: string }>;
  selectedDifficulty?: string;
  onDifficultySelect?: (id: string) => void;
  progress?: {
    solved: number;
    total: number;
    percentage: number;
  };
  collapsible?: boolean;
  className?: string;
}

export const Sidebar: React.FC<SidebarProps> = ({
  categories = [],
  selectedCategory,
  onCategorySelect,
  difficulties = [],
  selectedDifficulty,
  onDifficultySelect,
  progress,
  collapsible = true,
  className = '',
}) => {
  const [isCollapsed, setIsCollapsed] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);

  return (
    <>
      {/* Mobile Toggle */}
      {collapsible && (
        <button
          onClick={() => setMobileOpen(!mobileOpen)}
          className="lg:hidden fixed bottom-4 right-4 z-docked p-3 bg-primary-600 text-white rounded-full shadow-lg hover:bg-primary-700 transition-colors"
          aria-label="Toggle sidebar"
        >
          <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h7" />
          </svg>
        </button>
      )}

      {/* Overlay for mobile */}
      {mobileOpen && (
        <div
          className="lg:hidden fixed inset-0 bg-neutral-900/50 z-overlay"
          onClick={() => setMobileOpen(false)}
        />
      )}

      {/* Sidebar */}
      <aside
        className={`
          ${isCollapsed ? 'w-16' : 'w-64'}
          ${mobileOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'}
          fixed lg:sticky top-0 left-0 z-docked
          h-screen bg-white dark:bg-neutral-900
          border-r border-neutral-200 dark:border-neutral-800
          transition-all duration-300 ease-in-out
          overflow-y-auto
          ${className}
        `}
      >
        <div className="p-4">
          {/* Collapse Toggle (Desktop) */}
          {collapsible && (
            <button
              onClick={() => setIsCollapsed(!isCollapsed)}
              className="hidden lg:flex items-center justify-center w-full p-2 rounded-lg text-neutral-500 hover:text-neutral-700 hover:bg-neutral-100 dark:hover:bg-neutral-800 mb-4 transition-colors"
              aria-label={isCollapsed ? 'Expand sidebar' : 'Collapse sidebar'}
            >
              <svg className={`w-5 h-5 transition-transform ${isCollapsed ? 'rotate-180' : ''}`} fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 19l-7-7 7-7m8 14l-7-7 7-7" />
              </svg>
            </button>
          )}

          {/* Progress Section */}
          {progress && !isCollapsed && (
            <div className="mb-6 p-3 rounded-lg bg-neutral-50 dark:bg-neutral-800/50">
              <div className="flex items-center justify-between mb-2">
                <span className="text-xs font-medium text-neutral-600 dark:text-neutral-400">Progress</span>
                <span className="text-xs font-bold text-primary-600 dark:text-primary-400">
                  {progress.solved}/{progress.total}
                </span>
              </div>
              <div className="w-full h-2 bg-neutral-200 dark:bg-neutral-700 rounded-full overflow-hidden">
                <div
                  className="h-full bg-gradient-to-r from-primary-500 to-secondary-500 rounded-full transition-all duration-500"
                  style={{ width: `${progress.percentage}%` }}
                />
              </div>
              <p className="text-xs text-neutral-500 mt-1">{progress.percentage}% completed</p>
            </div>
          )}

          {/* Difficulty Filter */}
          {difficulties.length > 0 && (
            <div className="mb-6">
              {!isCollapsed && (
                <h3 className="text-xs font-semibold text-neutral-500 dark:text-neutral-400 uppercase tracking-wider mb-2 px-2">
                  Difficulty
                </h3>
              )}
              <div className="flex flex-col gap-1">
                {difficulties.map((diff) => (
                  <button
                    key={diff.id}
                    onClick={() => onDifficultySelect?.(diff.id)}
                    className={`
                      flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-colors
                      ${selectedDifficulty === diff.id
                        ? 'bg-neutral-100 dark:bg-neutral-800 font-medium'
                        : 'text-neutral-600 dark:text-neutral-400 hover:bg-neutral-50 dark:hover:bg-neutral-800/50'
                      }
                      ${isCollapsed ? 'justify-center' : ''}
                    `}
                    title={isCollapsed ? diff.label : undefined}
                  >
                    <span
                      className="w-3 h-3 rounded-full flex-shrink-0"
                      style={{ backgroundColor: diff.color }}
                    />
                    {!isCollapsed && <span>{diff.label}</span>}
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* Categories */}
          {categories.length > 0 && (
            <div>
              {!isCollapsed && (
                <h3 className="text-xs font-semibold text-neutral-500 dark:text-neutral-400 uppercase tracking-wider mb-2 px-2">
                  Categories
                </h3>
              )}
              <div className="flex flex-col gap-1">
                {categories.map((cat) => (
                  <button
                    key={cat.id}
                    onClick={() => onCategorySelect?.(cat.id)}
                    className={`
                      flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-colors
                      ${selectedCategory === cat.id
                        ? 'bg-primary-50 dark:bg-primary-900/20 text-primary-600 dark:text-primary-400 font-medium'
                        : 'text-neutral-600 dark:text-neutral-400 hover:bg-neutral-50 dark:hover:bg-neutral-800/50'
                      }
                      ${isCollapsed ? 'justify-center' : ''}
                    `}
                    title={isCollapsed ? cat.label : undefined}
                  >
                    {cat.icon && <span className="flex-shrink-0">{cat.icon}</span>}
                    {!isCollapsed && (
                      <>
                        <span className="flex-1 text-left">{cat.label}</span>
                        {cat.count !== undefined && (
                          <span className="text-xs text-neutral-400">{cat.count}</span>
                        )}
                      </>
                    )}
                  </button>
                ))}
              </div>
            </div>
          )}
        </div>
      </aside>
    </>
  );
};

Sidebar.displayName = 'Sidebar';

export default Sidebar;
