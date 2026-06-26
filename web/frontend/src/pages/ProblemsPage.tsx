import React, { useCallback } from "react";
import { useQuery } from "react-query";
import { problemApi } from "../services/api";
import { ProblemList } from "../components/ProblemList";
import { ProblemListSkeleton } from "../components/ui/LoadingSkeleton";
import { queryKeys } from "../hooks/useQueries";
import { useKeyboardShortcut } from "../hooks/useKeyboardShortcut";

export function ProblemsPage() {
  const [page, setPage] = React.useState(1);
  const [search] = React.useState("");
  const [difficulty] = React.useState("");
  const [category] = React.useState("");
  const [filterPanelOpen, setFilterPanelOpen] = React.useState(true);

  const { data, isLoading, error } = useQuery(
    queryKeys.problems.list({ page, page_size: 20, search, difficulty, category }),
    () =>
      problemApi
        .list({ page, page_size: 20, search, difficulty, category })
        .then((r) => r.data),
    { staleTime: 60000 }
  );

  // Toggle filter panel dengan Ctrl+Shift+P
  const toggleFilterPanel = useCallback(() => {
    setFilterPanelOpen((prev) => !prev);
  }, []);

  useKeyboardShortcut("p", true, true, toggleFilterPanel);

  if (isLoading) return <ProblemListSkeleton />;
  if (error) {
    return (
      <div className="text-center py-12">
        <p className="text-red-500 dark:text-red-400">Failed to load problems. Please try again.</p>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-slate-900 dark:text-white">Problems</h1>
            <p className="text-slate-500 dark:text-slate-400 mt-1">
              {data?.total ?? 0} problems to sharpen your skills
            </p>
          </div>
          <button
            onClick={toggleFilterPanel}
            className="hidden sm:inline-flex items-center gap-1 px-3 py-1.5 text-xs text-slate-500 dark:text-slate-400 bg-slate-100 dark:bg-slate-800 rounded-lg hover:bg-slate-200 dark:hover:bg-slate-700 transition-colors"
            title="Ctrl+Shift+P to toggle filters"
            aria-label="Toggle filter panel"
            style={{ minHeight: "36px" }}
          >
            <kbd className="px-1 py-0.5 bg-slate-200 dark:bg-slate-700 rounded text-[10px] font-mono">Ctrl+Shift+P</kbd>
          </button>
        </div>
      </div>
      {filterPanelOpen && (
        <ProblemList
          problems={data?.items ?? []}
          total={data?.total ?? 0}
          page={page}
          pageSize={20}
          onPageChange={setPage}
        />
      )}
      {!filterPanelOpen && (
        <div className="text-center py-12 text-slate-400 dark:text-slate-500">
          <p className="text-sm">Filter panel is hidden. Press <kbd className="px-1.5 py-0.5 bg-slate-100 dark:bg-slate-800 rounded text-xs font-mono">Ctrl+Shift+P</kbd> to show it.</p>
        </div>
      )}
    </div>
  );
}
