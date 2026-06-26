import React, { useState, useMemo, useCallback, useRef, useEffect, memo } from "react";
import { Link } from "react-router-dom";
import { Search, Filter, CheckCircle, Circle, Clock, ChevronLeft, ChevronRight } from "lucide-react";
import type { Difficulty, Category, Problem } from "../types";

interface ProblemListProps {
  problems: Problem[];
  total: number;
  page: number;
  pageSize: number;
  onPageChange: (page: number) => void;
}

const difficultyConfig: Record<Difficulty, { label: string; color: string }> = {
  easy: { label: "Easy", color: "bg-green-100 text-green-700" },
  medium: { label: "Medium", color: "bg-yellow-100 text-yellow-700" },
  hard: { label: "Hard", color: "bg-red-100 text-red-700" },
};

const categories: Category[] = [
  "arrays",
  "strings",
  "linked-lists",
  "trees",
  "graphs",
  "dynamic-programming",
  "sorting",
  "math",
  "greedy",
  "backtracking",
];

// Memoized status icon component
const StatusIcon = memo(function StatusIcon({ problem }: { problem: Problem }) {
  if (problem.solved_count > 0 && problem.attempt_count > 0) {
    return <CheckCircle className="w-4 h-4 text-green-500" aria-label="Solved" />;
  }
  if (problem.attempt_count > 0) {
    return <Clock className="w-4 h-4 text-yellow-500" aria-label="Attempted" />;
  }
  return <Circle className="w-4 h-4 text-slate-300" aria-label="Unattempted" />;
});

// Memoized problem card for virtualization
const ProblemCard = memo(function ProblemCard({ problem }: { problem: Problem }) {
  const diff = difficultyConfig[problem.difficulty];
  return (
    <Link
      to={`/problems/${problem.slug}`}
      className="block bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-4 hover:border-indigo-300 dark:hover:border-indigo-700 hover:shadow-sm dark:hover:shadow-indigo-900/20 transition-all"
    >
      <div className="flex items-start justify-between gap-4">
        <div className="flex items-start gap-3">
          <div className="mt-0.5">
            <StatusIcon problem={problem} />
          </div>
          <div>
            <h3 className="font-semibold text-slate-900 dark:text-white">{problem.title}</h3>
            <div className="flex flex-wrap items-center gap-2 mt-1.5">
              <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${diff.color}`}>
                {diff.label}
              </span>
              <span className="text-xs text-slate-500 dark:text-slate-400">
                {problem.category.replace("-", " ")}
              </span>
              <span className="text-xs text-slate-400 dark:text-slate-500">{problem.points} pts</span>
            </div>
          </div>
        </div>
        <div className="text-right shrink-0">
          <p className="text-xs text-slate-500 dark:text-slate-400">{problem.solved_count} solved</p>
          <p className="text-xs text-slate-400 dark:text-slate-500">
            {Math.round((problem.solved_count / Math.max(problem.attempt_count, 1)) * 100)}% success
          </p>
        </div>
      </div>
    </Link>
  );
});

// Virtualized list for large datasets
function VirtualizedProblemList({
  problems,
  itemHeight = 100,
  overscan = 5,
}: {
  problems: Problem[];
  itemHeight?: number;
  overscan?: number;
}) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [scrollTop, setScrollTop] = useState(0);
  const [containerHeight, setContainerHeight] = useState(600);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    const observer = new ResizeObserver((entries) => {
      for (const entry of entries) {
        setContainerHeight(entry.contentRect.height);
      }
    });
    observer.observe(container);
    return () => observer.disconnect();
  }, []);

  const handleScroll = useCallback((e: React.UIEvent<HTMLDivElement>) => {
    setScrollTop(e.currentTarget.scrollTop);
  }, []);

  const { visibleProblems, totalHeight, offsetY } = useMemo(() => {
    const totalHeight = problems.length * itemHeight;
    const startIndex = Math.max(0, Math.floor(scrollTop / itemHeight) - overscan);
    const endIndex = Math.min(
      problems.length,
      Math.ceil((scrollTop + containerHeight) / itemHeight) + overscan
    );
    const visibleProblems = problems.slice(startIndex, endIndex).map((problem, i) => ({
      problem,
      index: startIndex + i,
    }));
    const offsetY = startIndex * itemHeight;

    return { visibleProblems, totalHeight, offsetY };
  }, [problems, scrollTop, containerHeight, itemHeight, overscan]);

  if (problems.length === 0) {
    return (
      <div className="text-center py-12 text-slate-500">
        <p className="text-lg font-medium">No problems found</p>
        <p className="text-sm mt-1">Try adjusting your filters</p>
      </div>
    );
  }

  return (
    <div
      ref={containerRef}
      className="space-y-3 overflow-y-auto max-h-[70vh]"
      onScroll={handleScroll}
      style={{ position: "relative" }}
    >
      <div style={{ height: totalHeight, position: "relative" }}>
        <div style={{ position: "absolute", top: offsetY, left: 0, right: 0 }}>
          {visibleProblems.map(({ problem, index }) => (
            <div key={problem.id} style={{ height: itemHeight }}>
              <ProblemCard problem={problem} />
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

export function ProblemList({ problems, total, page, pageSize, onPageChange }: ProblemListProps) {
  const [search, setSearch] = useState("");
  const [difficulty, setDifficulty] = useState<Difficulty | "">("");
  const [category, setCategory] = useState<Category | "">("");

  // Memoized filtered problems
  const filteredProblems = useMemo(() => {
    return problems.filter((p) => {
      const matchesSearch =
        !search ||
        p.title.toLowerCase().includes(search.toLowerCase()) ||
        p.tags.some((t) => t.toLowerCase().includes(search.toLowerCase()));
      const matchesDifficulty = !difficulty || p.difficulty === difficulty;
      const matchesCategory = !category || p.category === category;
      return matchesSearch && matchesDifficulty && matchesCategory;
    });
  }, [problems, search, difficulty, category]);

  const totalPages = Math.ceil(total / pageSize);

  // Memoized callbacks
  const handleSearchChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setSearch(e.target.value);
  }, []);

  const handleDifficultyChange = useCallback((e: React.ChangeEvent<HTMLSelectElement>) => {
    setDifficulty(e.target.value as Difficulty | "");
  }, []);

  const handleCategoryChange = useCallback((e: React.ChangeEvent<HTMLSelectElement>) => {
    setCategory(e.target.value as Category | "");
  }, []);

  const handlePrevPage = useCallback(() => {
    onPageChange(page - 1);
  }, [page, onPageChange]);

  const handleNextPage = useCallback(() => {
    onPageChange(page + 1);
  }, [page, onPageChange]);

  // Use virtualized list for large datasets (>50 items)
  const useVirtualization = filteredProblems.length > 50;

  return (
    <div className="space-y-6">
      {/* Filters */}
      <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-4">
        <div className="flex flex-col sm:flex-row gap-3">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
            <input
              type="text"
              placeholder="Search problems..."
              value={search}
              onChange={handleSearchChange}
              className="w-full pl-10 pr-4 py-2 border border-slate-200 dark:border-slate-700 rounded-lg text-sm bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
              aria-label="Search problems"
              style={{ minHeight: "44px" }}
            />
          </div>

          <div className="relative">
            <Filter className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
            <select
              value={difficulty}
              onChange={handleDifficultyChange}
              className="pl-10 pr-8 py-2 border border-slate-200 dark:border-slate-700 rounded-lg text-sm appearance-none bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500"
              aria-label="Filter by difficulty"
              style={{ minHeight: "44px" }}
            >
              <option value="">All Difficulties</option>
              <option value="easy">Easy</option>
              <option value="medium">Medium</option>
              <option value="hard">Hard</option>
            </select>
          </div>

          <select
            value={category}
            onChange={handleCategoryChange}
            className="px-4 py-2 border border-slate-200 dark:border-slate-700 rounded-lg text-sm appearance-none bg-white dark:bg-slate-800 text-slate-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-indigo-500"
            aria-label="Filter by category"
            style={{ minHeight: "44px" }}
          >
            <option value="">All Categories</option>
            {categories.map((cat) => (
              <option key={cat} value={cat}>
                {cat.replace("-", " ").replace(/\b\w/g, (c) => c.toUpperCase())}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Problem Cards - Virtualized or Standard */}
      {useVirtualization ? (
        <VirtualizedProblemList problems={filteredProblems} />
      ) : (
        <div className="space-y-3">
          {filteredProblems.length === 0 ? (
            <div className="text-center py-12 text-slate-500">
              <p className="text-lg font-medium">No problems found</p>
              <p className="text-sm mt-1">Try adjusting your filters</p>
            </div>
          ) : (
            filteredProblems.map((problem) => (
              <ProblemCard key={problem.id} problem={problem} />
            ))
          )}
        </div>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <nav className="flex items-center justify-between" aria-label="Pagination">
          <p className="text-sm text-slate-600 dark:text-slate-400">
            Page {page} of {totalPages}
          </p>
          <div className="flex items-center gap-2">
            <button
              onClick={handlePrevPage}
              disabled={page <= 1}
              className="p-2 rounded-lg border border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              aria-label="Previous page"
              style={{ minWidth: "44px", minHeight: "44px" }}
            >
              <ChevronLeft className="w-4 h-4" />
            </button>
            <button
              onClick={handleNextPage}
              disabled={page >= totalPages}
              className="p-2 rounded-lg border border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-slate-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
              aria-label="Next page"
              style={{ minWidth: "44px", minHeight: "44px" }}
            >
              <ChevronRight className="w-4 h-4" />
            </button>
          </div>
        </nav>
      )}
    </div>
  );
}
