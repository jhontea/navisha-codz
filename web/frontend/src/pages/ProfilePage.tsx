import { useQuery } from "react-query";
import { Mail, Trophy, Target, Calendar } from "lucide-react";
import { authApi } from "../services/api";
import { useAuthStore } from "../store/authStore";
import { PageLoader } from "../components/ui/LoadingSpinner";

export function ProfilePage() {
  const { user: authUser } = useAuthStore();

  const { data, isLoading } = useQuery(
    ["profile"],
    () => authApi.getProfile().then((r) => r.data),
    { staleTime: 60000 }
  );

  const user = data?.user ?? authUser;

  if (isLoading) return <PageLoader />;
  if (!user) {
    return (
      <div className="text-center py-12">
        <p className="text-slate-500 dark:text-slate-400">Please log in to view your profile.</p>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <h1 className="text-2xl font-bold text-slate-900 dark:text-white">Profile</h1>

      <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-6">
        <div className="flex items-center gap-4">
          <div className="w-16 h-16 rounded-full bg-indigo-100 dark:bg-indigo-900/50 flex items-center justify-center shrink-0">
            {user.avatar_url ? (
              <img
                src={user.avatar_url}
                alt={user.username}
                className="w-16 h-16 rounded-full object-cover"
                loading="lazy"
              />
            ) : (
              <span className="text-2xl font-bold text-indigo-600 dark:text-indigo-400">
                {user.username.charAt(0).toUpperCase()}
              </span>
            )}
          </div>
          <div className="min-w-0">
            <h2 className="text-xl font-bold text-slate-900 dark:text-white">{user.username}</h2>
            <p className="text-slate-500 dark:text-slate-400 flex items-center gap-1 text-sm">
              <Mail className="w-4 h-4 shrink-0" />
              <span className="truncate">{user.email}</span>
            </p>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-4 text-center">
          <Trophy className="w-6 h-6 text-yellow-500 mx-auto mb-2" />
          <p className="text-2xl font-bold text-slate-900 dark:text-white">{user.score.toLocaleString()}</p>
          <p className="text-sm text-slate-500 dark:text-slate-400">Score</p>
        </div>
        <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-4 text-center">
          <Target className="w-6 h-6 text-indigo-500 mx-auto mb-2" />
          <p className="text-2xl font-bold text-slate-900 dark:text-white">#{user.rank}</p>
          <p className="text-sm text-slate-500 dark:text-slate-400">Rank</p>
        </div>
        <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-4 text-center">
          <Calendar className="w-6 h-6 text-green-500 mx-auto mb-2" />
          <p className="text-2xl font-bold text-slate-900 dark:text-white">
            {new Date(user.created_at).toLocaleDateString()}
          </p>
          <p className="text-sm text-slate-500 dark:text-slate-400">Joined</p>
        </div>
      </div>
    </div>
  );
}
