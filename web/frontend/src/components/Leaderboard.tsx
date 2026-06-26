import { useState, useMemo } from "react";
import {
  Trophy,
  Medal,
  RefreshCw,
  Crown,
  Search,
  TrendingUp,
  Flame,
  Star,
  Gift,
  Users,
} from "lucide-react";
import type { LeaderboardEntry, LeaderboardPeriod } from "../types";
import { useAuthStore } from "../store/authStore";

interface LeaderboardProps {
  entries: LeaderboardEntry[];
  period: LeaderboardPeriod;
  onPeriodChange: (period: LeaderboardPeriod) => void;
  onRefresh: () => void;
  totalEntries?: number;
  onSearch?: (query: string) => void;
  onCompare?: () => void;
  ratingHistory?: { rating: number; change_amount: number; reason: string; created_at: string }[];
  rewards?: {
    weekly: RewardInfo[];
    monthly: RewardInfo[];
  };
}

interface RewardInfo {
  rank: number;
  reward: string;
  points: number;
  emoji: string;
}

interface RatingPoint {
  rating: number;
  change_amount: number;
  reason: string;
  created_at: string;
}

const periods: { value: LeaderboardPeriod; label: string; icon: React.ReactNode }[] = [
  { value: "weekly", label: "Weekly", icon: <Flame className="w-3.5 h-3.5" /> },
  { value: "monthly", label: "Monthly", icon: <Star className="w-3.5 h-3.5" /> },
  { value: "all-time", label: "All Time", icon: <Trophy className="w-3.5 h-3.5" /> },
];

function MedalAnimation({ rank }: { rank: number }) {
  if (rank === 1) {
    return (
      <div className="flex items-center justify-center w-10 h-10 rounded-full bg-gradient-to-br from-yellow-200 to-amber-300 shadow-lg animate-pulse">
        <Crown className="w-5 h-5 text-yellow-700" />
      </div>
    );
  }
  if (rank === 2) {
    return (
      <div className="flex items-center justify-center w-10 h-10 rounded-full bg-gradient-to-br from-slate-200 to-slate-300 shadow-md">
        <Medal className="w-5 h-5 text-slate-600" />
      </div>
    );
  }
  if (rank === 3) {
    return (
      <div className="flex items-center justify-center w-10 h-10 rounded-full bg-gradient-to-br from-amber-100 to-amber-200 shadow-md">
        <Medal className="w-5 h-5 text-amber-700" />
      </div>
    );
  }
  return (
    <div className="flex items-center justify-center w-10 h-10">
      <span className="text-sm font-bold text-slate-400 dark:text-slate-500">#{rank}</span>
    </div>
  );
}

function RatingChart({ history }: { history: RatingPoint[] }) {
  if (!history || history.length < 2) {
    return (
      <div className="text-center py-6 text-sm text-slate-400">
        Not enough data to show chart
      </div>
    );
  }

  const maxRating = Math.max(...history.map((h) => h.rating)) + 50;
  const minRating = Math.max(0, Math.min(...history.map((h) => h.rating)) - 50);
  const range = maxRating - minRating || 1;
  const width = 280;
  const height = 100;
  const padding = { top: 10, right: 10, bottom: 20, left: 40 };
  const chartW = width - padding.left - padding.right;
  const chartH = height - padding.top - padding.bottom;

  const points = history.map((h, i) => {
    const x = padding.left + (i / Math.max(history.length - 1, 1)) * chartW;
    const y = padding.top + (1 - (h.rating - minRating) / range) * chartH;
    return { x, y, ...h };
  });

  const pathD = points.map((p, i) => `${i === 0 ? "M" : "L"}${p.x.toFixed(1)},${p.y.toFixed(1)}`).join(" ");
  const areaD = `${pathD} L${points[points.length - 1].x},${padding.top + chartH} L${points[0].x},${padding.top + chartH} Z`;

  return (
    <svg width={width} height={height} viewBox={`0 0 ${width} ${height}`} className="w-full max-w-xs">
      {/* Gradient */}
      <defs>
        <linearGradient id="ratingGradient" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="#6366f1" stopOpacity="0.3" />
          <stop offset="100%" stopColor="#6366f1" stopOpacity="0.05" />
        </linearGradient>
      </defs>

      {/* Area fill */}
      <path d={areaD} fill="url(#ratingGradient)" />

      {/* Line */}
      <path d={pathD} fill="none" stroke="#6366f1" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />

      {/* Points */}
      {points.map((p, i) => (
        <circle key={i} cx={p.x} cy={p.y} r="3" fill="#6366f1" className="hover:r-5 transition-all">
          <title>{`Rating: ${p.rating} (${p.change_amount >= 0 ? "+" : ""}${p.change_amount}) - ${p.reason}`}</title>
        </circle>
      ))}

      {/* Y-axis labels */}
      <text x="0" y={padding.top + 4} fontSize="8" fill="#94a3b8">{maxRating}</text>
      <text x="0" y={padding.top + chartH} fontSize="8" fill="#94a3b8">{minRating}</text>

      {/* X-axis label */}
      <text x={width / 2} y={height - 2} fontSize="8" fill="#94a3b8" textAnchor="middle">
        Time →
      </text>
    </svg>
  );
}

function RewardSection({ rewards, period }: { rewards: RewardInfo[]; period: string }) {
  if (!rewards || rewards.length === 0) return null;

  return (
    <div className="bg-gradient-to-r from-amber-50 to-yellow-50 dark:from-amber-900/10 dark:to-yellow-900/10 rounded-lg p-3 border border-amber-200 dark:border-amber-700">
      <div className="flex items-center gap-1.5 mb-2">
        <Gift className="w-4 h-4 text-amber-500" />
        <span className="text-xs font-semibold text-amber-700 dark:text-amber-300 uppercase tracking-wider">
          {period} Rewards
        </span>
      </div>
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-1.5">
        {rewards.slice(0, 8).map((reward, idx) => (
          <div
            key={idx}
            className="flex items-center gap-1.5 px-2 py-1.5 rounded bg-white/60 dark:bg-slate-800/60 text-xs"
          >
            <span>{reward.emoji}</span>
            <span className="font-medium text-slate-700 dark:text-slate-300">#{reward.rank}</span>
            <span className="text-slate-500 dark:text-slate-400 truncate">{reward.reward}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

export function Leaderboard({
  entries,
  period,
  onPeriodChange,
  onRefresh,
  totalEntries = 0,
  onSearch,
  onCompare,
  ratingHistory,
  rewards,
}: LeaderboardProps) {
  const { user } = useAuthStore();
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [showSearch, setShowSearch] = useState(false);
  const [compareMode, setCompareMode] = useState(false);
  const [selectedUsers, setSelectedUsers] = useState<string[]>([]);
  const [showRewards, setShowRewards] = useState(false);
  const [showRatingChart, setShowRatingChart] = useState(false);

  const handleRefresh = async () => {
    setIsRefreshing(true);
    onRefresh();
    setTimeout(() => setIsRefreshing(false), 1000);
  };

  const handleSearchChange = (value: string) => {
    setSearchQuery(value);
    onSearch?.(value);
  };

  const toggleUserSelection = (userId: string) => {
    setSelectedUsers((prev) =>
      prev.includes(userId) ? prev.filter((id) => id !== userId) : [...prev, userId]
    );
  };

  // Filter entries based on search
  const filteredEntries = useMemo(() => {
    if (!searchQuery) return entries;
    const q = searchQuery.toLowerCase();
    return entries.filter((e) => e.username.toLowerCase().includes(q));
  }, [entries, searchQuery]);

  // Split entries for current user highlighting
  const currentUserEntry = entries.find((e) => user && e.user_id === user.id);

  return (
    <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden">
      {/* Header */}
      <div className="px-6 py-4 border-b border-slate-200 dark:border-slate-700">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Trophy className="w-5 h-5 text-amber-500" />
            <h2 className="text-lg font-bold text-slate-900 dark:text-white">Leaderboard</h2>
            {totalEntries > 0 && (
              <span className="text-xs text-slate-400 dark:text-slate-500 bg-slate-100 dark:bg-slate-800 px-2 py-0.5 rounded-full">
                {totalEntries} players
              </span>
            )}
          </div>
          <div className="flex items-center gap-1">
            {/* Search toggle */}
            <button
              onClick={() => setShowSearch(!showSearch)}
              className={`p-2 rounded-lg transition-colors ${
                showSearch
                  ? "bg-indigo-100 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400"
                  : "text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800"
              }`}
              aria-label="Search users"
            >
              <Search className="w-4 h-4" />
            </button>

            {/* Compare toggle */}
            <button
              onClick={() => setCompareMode(!compareMode)}
              className={`p-2 rounded-lg transition-colors ${
                compareMode
                  ? "bg-indigo-100 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400"
                  : "text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800"
              }`}
              aria-label="Compare users"
            >
              <Users className="w-4 h-4" />
            </button>

            {/* Rewards toggle */}
            <button
              onClick={() => setShowRewards(!showRewards)}
              className={`p-2 rounded-lg transition-colors ${
                showRewards
                  ? "bg-amber-100 dark:bg-amber-900/30 text-amber-600 dark:text-amber-400"
                  : "text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800"
              }`}
              aria-label="Show rewards"
            >
              <Gift className="w-4 h-4" />
            </button>

            {/* Rating chart toggle */}
            {ratingHistory && ratingHistory.length > 0 && (
              <button
                onClick={() => setShowRatingChart(!showRatingChart)}
                className={`p-2 rounded-lg transition-colors ${
                  showRatingChart
                    ? "bg-green-100 dark:bg-green-900/30 text-green-600 dark:text-green-400"
                    : "text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800"
                }`}
                aria-label="Show rating chart"
              >
                <TrendingUp className="w-4 h-4" />
              </button>
            )}

            {/* Refresh */}
            <button
              onClick={handleRefresh}
              disabled={isRefreshing}
              className="p-2 rounded-lg text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors disabled:opacity-50"
              aria-label="Refresh leaderboard"
            >
              <RefreshCw className={`w-4 h-4 ${isRefreshing ? "animate-spin" : ""}`} />
            </button>
          </div>
        </div>

        {/* Search bar */}
        {showSearch && (
          <div className="mt-3">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-400" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => handleSearchChange(e.target.value)}
                placeholder="Search by username..."
                className="w-full pl-9 pr-4 py-2 text-sm border border-slate-200 dark:border-slate-600 rounded-lg bg-transparent text-slate-900 dark:text-white placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                aria-label="Search leaderboard by username"
              />
            </div>
          </div>
        )}

        {/* Compare mode info */}
        {compareMode && (
          <div className="mt-3 p-2 bg-indigo-50 dark:bg-indigo-900/20 rounded-lg">
            <p className="text-xs text-indigo-600 dark:text-indigo-400">
              Click on users to select them for comparison. Selected: {selectedUsers.length}
            </p>
            {selectedUsers.length === 2 && (
              <button
                onClick={onCompare}
                className="mt-1 px-3 py-1 text-xs font-medium bg-indigo-500 text-white rounded-md hover:bg-indigo-600 transition-colors"
              >
                Compare Selected Users
              </button>
            )}
          </div>
        )}

        {/* Rewards section */}
        {showRewards && rewards && (
          <div className="mt-3 space-y-2">
            {period === "weekly" && <RewardSection rewards={rewards.weekly} period="Weekly" />}
            {period === "monthly" && <RewardSection rewards={rewards.monthly} period="Monthly" />}
            {period === "all-time" && (
              <>
                <RewardSection rewards={rewards.weekly} period="Weekly" />
                <RewardSection rewards={rewards.monthly} period="Monthly" />
              </>
            )}
          </div>
        )}

        {/* Rating chart */}
        {showRatingChart && ratingHistory && ratingHistory.length > 0 && (
          <div className="mt-3 p-3 bg-slate-50 dark:bg-slate-800 rounded-lg">
            <div className="flex items-center justify-between mb-2">
              <span className="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                Rating History
              </span>
              <span className="text-xs text-slate-400">
                Current: {ratingHistory[ratingHistory.length - 1]?.rating ?? "N/A"}
              </span>
            </div>
            <RatingChart history={ratingHistory} />
          </div>
        )}

        {/* Period tabs */}
        <div className="flex gap-1 mt-4 bg-slate-100 dark:bg-slate-800 rounded-lg p-1" role="tablist">
          {periods.map((p) => (
            <button
              key={p.value}
              onClick={() => onPeriodChange(p.value)}
              className={`flex-1 flex items-center justify-center gap-1.5 px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
                period === p.value
                  ? "bg-white dark:bg-slate-700 text-slate-900 dark:text-white shadow-sm"
                  : "text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white"
              }`}
              role="tab"
              aria-selected={period === p.value}
            >
              {p.icon}
              {p.label}
            </button>
          ))}
        </div>
      </div>

      {/* Current user stats (if on leaderboard) */}
      {currentUserEntry && (
        <div className="px-6 py-3 bg-indigo-50 dark:bg-indigo-900/20 border-b border-indigo-100 dark:border-indigo-800">
          <div className="flex items-center justify-between text-sm">
            <div className="flex items-center gap-2">
              <MedalAnimation rank={currentUserEntry.rank} />
              <div>
                <span className="font-semibold text-indigo-700 dark:text-indigo-300">
                  {currentUserEntry.username}
                </span>
                <span className="ml-1.5 text-xs text-indigo-500">(you)</span>
              </div>
            </div>
            <div className="text-right">
              <span className="font-bold text-indigo-700 dark:text-indigo-300">
                #{currentUserEntry.rank}
              </span>
              <span className="ml-2 text-indigo-500">
                {currentUserEntry.score.toLocaleString()} pts
              </span>
            </div>
          </div>
        </div>
      )}

      {/* Table */}
      <div className="overflow-x-auto">
        <table className="w-full" role="table">
          <thead className="bg-slate-50 dark:bg-slate-800 border-b border-slate-200 dark:border-slate-700">
            <tr>
              <th className="px-4 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider w-16">
                Rank
              </th>
              <th className="px-4 py-3 text-left text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                User
              </th>
              <th className="px-4 py-3 text-right text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                Score
              </th>
              <th className="px-4 py-3 text-right text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider hidden sm:table-cell">
                Rating
              </th>
              <th className="px-4 py-3 text-right text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider hidden md:table-cell">
                Solved
              </th>
              <th className="px-4 py-3 text-right text-xs font-medium text-slate-500 dark:text-slate-400 uppercase tracking-wider hidden lg:table-cell">
                Accuracy
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100 dark:divide-slate-800">
            {filteredEntries.length === 0 ? (
              <tr>
                <td colSpan={6} className="px-4 py-8 text-center text-sm text-slate-500 dark:text-slate-400">
                  {searchQuery ? "No users found matching your search" : "No entries yet"}
                </td>
              </tr>
            ) : (
              filteredEntries.map((entry) => {
                const isCurrentUser = user && user.id === entry.user_id;
                const isSelected = selectedUsers.includes(entry.user_id);

                return (
                  <tr
                    key={entry.user_id}
                    className={`transition-colors ${
                      isCurrentUser
                        ? "bg-indigo-50 dark:bg-indigo-900/10"
                        : "hover:bg-slate-50 dark:hover:bg-slate-800/50"
                    } ${compareMode ? "cursor-pointer" : ""}`}
                    onClick={() => {
                      if (compareMode) {
                        toggleUserSelection(entry.user_id);
                      }
                    }}
                    onKeyDown={(e) => {
                      if (compareMode && e.key === "Enter") {
                        toggleUserSelection(entry.user_id);
                      }
                    }}
                    role={compareMode ? "button" : undefined}
                    tabIndex={compareMode ? 0 : undefined}
                  >
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-1">
                        {isSelected && (
                          <div className="w-2 h-2 rounded-full bg-indigo-500" />
                        )}
                        <MedalAnimation rank={entry.rank} />
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        <div className="w-8 h-8 rounded-full bg-slate-200 dark:bg-slate-700 flex items-center justify-center shrink-0">
                          {entry.avatar_url ? (
                            <img
                              src={entry.avatar_url}
                              alt=""
                              className="w-8 h-8 rounded-full object-cover"
                            />
                          ) : (
                            <span className="text-xs font-semibold text-slate-600 dark:text-slate-400">
                              {entry.username.charAt(0).toUpperCase()}
                            </span>
                          )}
                        </div>
                        <div className="min-w-0">
                          <span
                            className={`text-sm font-medium ${
                              isCurrentUser ? "text-indigo-700 dark:text-indigo-300" : "text-slate-900 dark:text-white"
                            }`}
                          >
                            {entry.username}
                            {isCurrentUser && (
                              <span className="ml-1 text-xs text-indigo-500">(you)</span>
                            )}
                          </span>
                        </div>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <span className="text-sm font-semibold text-slate-900 dark:text-white">
                        {entry.score.toLocaleString()}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right hidden sm:table-cell">
                      <span className="text-sm text-slate-600 dark:text-slate-400 font-mono">
                        {entry.rating ?? "—"}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right hidden md:table-cell">
                      <span className="text-sm text-slate-600 dark:text-slate-400">
                        {entry.problems_solved}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right hidden lg:table-cell">
                      <span className="text-sm text-slate-600 dark:text-slate-400">
                        {Math.round(entry.accuracy * 100)}%
                      </span>
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination info */}
      {totalEntries > 0 && (
        <div className="px-6 py-3 border-t border-slate-200 dark:border-slate-700 flex items-center justify-between">
          <span className="text-xs text-slate-500 dark:text-slate-400">
            Showing {filteredEntries.length} of {totalEntries} entries
          </span>
        </div>
      )}
    </div>
  );
}
