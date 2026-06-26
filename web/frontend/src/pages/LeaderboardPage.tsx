import React from "react";
import { useQuery } from "react-query";
import { Trophy, Medal, TrendingUp } from "lucide-react";
import { leaderboardApi } from "../services/api";
import { PageLoader } from "../components/ui/LoadingSpinner";
import { queryKeys } from "../hooks/useQueries";
import type { LeaderboardPeriod } from "../types";

const periods: { value: LeaderboardPeriod; label: string }[] = [
  { value: "weekly", label: "Weekly" },
  { value: "monthly", label: "Monthly" },
  { value: "all-time", label: "All Time" },
];

export function LeaderboardPage() {
  const [period, setPeriod] = React.useState<LeaderboardPeriod>("all-time");

  const { data, isLoading } = useQuery(
    queryKeys.leaderboard.list(period),
    () => leaderboardApi.get(period, 1, 50).then((r) => r.data),
    { staleTime: 60000, refetchInterval: 30000 }
  );

  if (isLoading) return <PageLoader />;

  const entries = data?.items ?? [];

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-slate-900 dark:text-white">Leaderboard</h1>
          <p className="text-slate-500 dark:text-slate-400 mt-1">Top performers this {period}</p>
        </div>
        <div className="flex items-center gap-1 bg-slate-100 dark:bg-slate-800 rounded-lg p-1">
          {periods.map((p) => (
            <button
              key={p.value}
              onClick={() => setPeriod(p.value)}
              className={`px-3 py-1.5 text-sm font-medium rounded-md transition-colors min-h-[36px] ${
                period === p.value
                  ? "bg-white dark:bg-slate-700 text-slate-900 dark:text-white shadow-sm"
                  : "text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-300"
              }`}
            >
              {p.label}
            </button>
          ))}
        </div>
      </div>

      {entries.length === 0 ? (
        <div className="text-center py-12 text-slate-500 dark:text-slate-400">
          <Trophy className="w-12 h-12 mx-auto mb-4 text-slate-300 dark:text-slate-600" />
          <p className="text-lg font-medium">No entries yet</p>
          <p className="text-sm mt-1">Be the first to solve a problem!</p>
        </div>
      ) : (
        <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden">
          <div className="divide-y divide-slate-100 dark:divide-slate-800">
            {entries.map((entry, index) => (
              <div
                key={entry.user_id}
                className="flex items-center gap-4 p-4 hover:bg-slate-50 dark:hover:bg-slate-800/50 transition-colors"
              >
                <div className="flex items-center justify-center w-8 h-8 shrink-0">
                  {index === 0 ? (
                    <Medal className="w-6 h-6 text-yellow-500" />
                  ) : index === 1 ? (
                    <Medal className="w-6 h-6 text-slate-400" />
                  ) : index === 2 ? (
                    <Medal className="w-6 h-6 text-amber-600" />
                  ) : (
                    <span className="text-sm font-bold text-slate-400 dark:text-slate-500">#{entry.rank}</span>
                  )}
                </div>
                <div className="flex-1 min-w-0">
                  <p className="font-medium text-slate-900 dark:text-white truncate">{entry.username}</p>
                  <p className="text-xs text-slate-500 dark:text-slate-400">
                    {entry.problems_solved} solved · {entry.total_submissions} submissions
                  </p>
                </div>
                <div className="text-right shrink-0">
                  <p className="font-bold text-slate-900 dark:text-white">{entry.score.toLocaleString()}</p>
                  <p className="text-xs text-slate-500 dark:text-slate-400 flex items-center gap-1 justify-end">
                    <TrendingUp className="w-3 h-3" />
                    {(entry.accuracy * 100).toFixed(1)}%
                  </p>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
