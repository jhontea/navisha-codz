import React, { memo } from "react";
import type { Category } from "../../types";

// ── Category Icon Map ───────────────────────────────────────────────────────

interface CategoryVisual {
  label: string;
  color: string;         // Tailwind bg class
  badgeColor: string;    // Tailwind text class
  icon: React.ReactNode;
  gradient: string;      // Tailwind gradient class for hover
}

const categoryIcons: Record<Category, CategoryVisual> = {
  "arrays": {
    label: "Arrays",
    color: "bg-blue-500",
    badgeColor: "text-blue-600 dark:text-blue-400",
    gradient: "from-blue-500 to-blue-600",
    icon: (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} className="w-5 h-5">
        <rect x="3" y="3" width="4" height="18" rx="1" />
        <rect x="10" y="3" width="4" height="18" rx="1" />
        <rect x="17" y="3" width="4" height="18" rx="1" />
      </svg>
    ),
  },
  "strings": {
    label: "Strings",
    color: "bg-emerald-500",
    badgeColor: "text-emerald-600 dark:text-emerald-400",
    gradient: "from-emerald-500 to-emerald-600",
    icon: (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} className="w-5 h-5">
        <path strokeLinecap="round" strokeLinejoin="round" d="M4 6h16M4 12h16M4 18h7" />
        <path strokeLinecap="round" strokeLinejoin="round" d="M15 18l3-3 3 3" />
      </svg>
    ),
  },
  "linked-lists": {
    label: "Linked Lists",
    color: "bg-amber-500",
    badgeColor: "text-amber-600 dark:text-amber-400",
    gradient: "from-amber-500 to-amber-600",
    icon: (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} className="w-5 h-5">
        <circle cx="5" cy="12" r="2.5" />
        <circle cx="12" cy="6" r="2.5" />
        <circle cx="12" cy="18" r="2.5" />
        <circle cx="19" cy="12" r="2.5" />
        <line x1="7" y1="12" x2="9.5" y2="6" />
        <line x1="14.5" y1="6" x2="16.5" y2="12" />
        <line x1="7" y1="12" x2="9.5" y2="18" />
        <line x1="14.5" y1="18" x2="16.5" y2="12" />
      </svg>
    ),
  },
  "trees": {
    label: "Trees",
    color: "bg-green-600",
    badgeColor: "text-green-600 dark:text-green-400",
    gradient: "from-green-600 to-green-700",
    icon: (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} className="w-5 h-5">
        <circle cx="12" cy="4" r="2" />
        <circle cx="6" cy="12" r="2" />
        <circle cx="18" cy="12" r="2" />
        <circle cx="12" cy="20" r="2" />
        <line x1="10.5" y1="5.5" x2="7.5" y2="10.5" />
        <line x1="13.5" y1="5.5" x2="16.5" y2="10.5" />
        <line x1="7.5" y1="13.5" x2="10.5" y2="18.5" />
        <line x1="16.5" y1="13.5" x2="13.5" y2="18.5" />
      </svg>
    ),
  },
  "graphs": {
    label: "Graphs",
    color: "bg-violet-500",
    badgeColor: "text-violet-600 dark:text-violet-400",
    gradient: "from-violet-500 to-violet-600",
    icon: (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} className="w-5 h-5">
        <circle cx="12" cy="4" r="2.5" />
        <circle cx="4" cy="18" r="2.5" />
        <circle cx="20" cy="18" r="2.5" />
        <line x1="10" y1="6" x2="6" y2="16" />
        <line x1="14" y1="6" x2="18" y2="16" />
        <line x1="6.5" y1="18" x2="17.5" y2="18" />
      </svg>
    ),
  },
  "dynamic-programming": {
    label: "DP",
    color: "bg-rose-500",
    badgeColor: "text-rose-600 dark:text-rose-400",
    gradient: "from-rose-500 to-rose-600",
    icon: (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} className="w-5 h-5">
        <rect x="3" y="13" width="4" height="8" rx="1" />
        <rect x="10" y="8" width="4" height="13" rx="1" />
        <rect x="17" y="3" width="4" height="18" rx="1" />
        <path strokeLinecap="round" strokeLinejoin="round" d="M3 21h18" />
      </svg>
    ),
  },
  "sorting": {
    label: "Sorting",
    color: "bg-cyan-500",
    badgeColor: "text-cyan-600 dark:text-cyan-400",
    gradient: "from-cyan-500 to-cyan-600",
    icon: (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} className="w-5 h-5">
        <rect x="4" y="6" width="16" height="2" rx="1" />
        <rect x="6" y="11" width="12" height="2" rx="1" />
        <rect x="8" y="16" width="8" height="2" rx="1" />
        <path strokeLinecap="round" strokeLinejoin="round" d="M19 10l2-2-2-2M5 14l-2 2 2 2" />
      </svg>
    ),
  },
  "math": {
    label: "Math",
    color: "bg-orange-500",
    badgeColor: "text-orange-600 dark:text-orange-400",
    gradient: "from-orange-500 to-orange-600",
    icon: (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} className="w-5 h-5">
        <path strokeLinecap="round" strokeLinejoin="round" d="M9 3v18M3 9h18" />
        <circle cx="17" cy="17" r="4" />
      </svg>
    ),
  },
  "greedy": {
    label: "Greedy",
    color: "bg-teal-500",
    badgeColor: "text-teal-600 dark:text-teal-400",
    gradient: "from-teal-500 to-teal-600",
    icon: (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} className="w-5 h-5">
        <path strokeLinecap="round" strokeLinejoin="round" d="M12 2v4M12 18v4M4 12H2M22 12h-2M5.64 5.64l2.83 2.83M15.53 15.53l2.83 2.83M5.64 18.36l2.83-2.83M15.53 8.47l2.83-2.83" />
        <circle cx="12" cy="12" r="3" />
      </svg>
    ),
  },
  "backtracking": {
    label: "Backtracking",
    color: "bg-pink-500",
    badgeColor: "text-pink-600 dark:text-pink-400",
    gradient: "from-pink-500 to-pink-600",
    icon: (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} className="w-5 h-5">
        <path strokeLinecap="round" strokeLinejoin="round" d="M3 12h18M12 3v18" />
        <path strokeLinecap="round" strokeLinejoin="round" d="M8 8l4-4 4 4M8 16l4 4 4-4" />
      </svg>
    ),
  },
};

// ── Props ────────────────────────────────────────────────────────────────────

interface CategoryIconProps {
  category: Category;
  size?: "sm" | "md" | "lg";
  showLabel?: boolean;
  interactive?: boolean;
  className?: string;
}

// ── Size Maps ────────────────────────────────────────────────────────────────

const sizeClasses = {
  sm: {
    container: "w-8 h-8",
    icon: "w-4 h-4",
    text: "text-[10px]",
    gap: "gap-1",
  },
  md: {
    container: "w-10 h-10",
    icon: "w-5 h-5",
    text: "text-xs",
    gap: "gap-1.5",
  },
  lg: {
    container: "w-12 h-12",
    icon: "w-6 h-6",
    text: "text-sm",
    gap: "gap-2",
  },
};

// ── Component ────────────────────────────────────────────────────────────────

export const CategoryIcon = memo(function CategoryIcon({
  category,
  size = "md",
  showLabel = false,
  interactive = true,
  className = "",
}: CategoryIconProps) {
  const visual = categoryIcons[category];

  if (!visual) {
    return (
      <span className={`inline-flex items-center justify-center ${sizeClasses[size].container} rounded-lg bg-slate-100 dark:bg-slate-800 text-slate-400 ${className}`}>
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} className={sizeClasses[size].icon}>
          <circle cx="12" cy="12" r="10" />
          <path strokeLinecap="round" d="M12 8v4M12 16h.01" />
        </svg>
      </span>
    );
  }

  const s = sizeClasses[size];

  return (
    <span
      className={`
        inline-flex items-center ${s.gap}
        ${className}
      `}
    >
      <span
        className={`
          inline-flex items-center justify-center ${s.container} rounded-lg
          ${visual.color} text-white
          ${interactive ? "transition-all duration-200 hover:scale-110 hover:shadow-lg hover:shadow-current/30 cursor-pointer" : ""}
        `}
        title={visual.label}
        role="img"
        aria-label={`${visual.label} category`}
      >
        {React.cloneElement(visual.icon as React.ReactElement, {
          className: s.icon,
        })}
      </span>
      {showLabel && (
        <span className={`${s.text} font-medium ${visual.badgeColor}`}>
          {visual.label}
        </span>
      )}
    </span>
  );
});

// ── Helper: Get category color by category name ─────────────────────────────

export function getCategoryColor(category: Category): string {
  return categoryIcons[category]?.color ?? "bg-slate-400";
}

export function getCategoryLabel(category: Category): string {
  return categoryIcons[category]?.label ?? category;
}

// ── Export all category data for use in selectors etc ────────────────────────

export const CATEGORY_LIST: Array<{ value: Category; label: string; color: string }> = Object.entries(
  categoryIcons
).map(([value, visual]) => ({
  value: value as Category,
  label: visual.label,
  color: visual.color,
}));

export default CategoryIcon;
