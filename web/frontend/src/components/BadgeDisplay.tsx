import { useState } from "react";
import type { Badge, Achievement } from "../types";
import { Share2, Lock, CheckCircle, TrendingUp } from "lucide-react";

interface BadgeDisplayProps {
  badges: Badge[];
  achievements: Achievement[];
  showProgress?: boolean;
  compact?: boolean;
}

export function BadgeDisplay({ badges, achievements, showProgress = true, compact = false }: BadgeDisplayProps) {
  const [activeTab, setActiveTab] = useState<"badges" | "achievements">("badges");
  const [animatingBadge, setAnimatingBadge] = useState<string | null>(null);
  const [shareMessage, setShareMessage] = useState<string | null>(null);

  const handleShare = async (badge: Badge) => {
    const shareText = `I earned the ${badge.emoji} ${badge.name} badge on Coding Challenge! ${badge.description}`;

    if (navigator.share) {
      try {
        await navigator.share({
          title: `Badge: ${badge.name}`,
          text: shareText,
          url: window.location.href,
        });
      } catch {
        // User cancelled
      }
    } else {
      // Fallback: copy to clipboard
      try {
        await navigator.clipboard.writeText(shareText);
        setShareMessage("Copied to clipboard!");
        setTimeout(() => setShareMessage(null), 2000);
      } catch {
        setShareMessage("Share not available");
        setTimeout(() => setShareMessage(null), 2000);
      }
    }
  };

  const handleAchievementShare = async (achievement: Achievement) => {
    const shareText = `I unlocked "${achievement.name}" ${achievement.icon} on Coding Challenge! ${achievement.description}`;

    if (navigator.share) {
      try {
        await navigator.share({
          title: `Achievement: ${achievement.name}`,
          text: shareText,
          url: window.location.href,
        });
      } catch {
        // User cancelled
      }
    } else {
      try {
        await navigator.clipboard.writeText(shareText);
        setShareMessage("Copied to clipboard!");
        setTimeout(() => setShareMessage(null), 2000);
      } catch {
        setShareMessage("Share not available");
        setTimeout(() => setShareMessage(null), 2000);
      }
    }
  };

  const triggerAnimation = (badgeId: string) => {
    setAnimatingBadge(badgeId);
    setTimeout(() => setAnimatingBadge(null), 1000);
  };

  const unlockedBadges = badges.filter((b) => b.unlocked_at);
  const lockedBadges = badges.filter((b) => !b.unlocked_at);
  const unlockedAchievements = achievements.filter((a) => a.unlocked_at);
  const lockedAchievements = achievements.filter((a) => !a.unlocked_at);

  if (compact) {
    return (
      <div className="space-y-3">
        {/* Mini badge display */}
        <div className="flex flex-wrap gap-2">
          {unlockedBadges.slice(0, 6).map((badge) => (
            <button
              key={badge.id}
              onClick={() => handleShare(badge)}
              className="group relative flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg bg-gradient-to-r from-amber-50 to-yellow-50 dark:from-amber-900/20 dark:to-yellow-900/20 border border-amber-200 dark:border-amber-700 hover:shadow-md transition-all"
              title={badge.description}
            >
              <span className="text-lg">{badge.emoji}</span>
              <span className="text-xs font-medium text-amber-700 dark:text-amber-300 hidden group-hover:inline">
                {badge.name}
              </span>
            </button>
          ))}
          {lockedBadges.length > 0 && (
            <span className="flex items-center gap-1 px-2 py-1 text-xs text-slate-400">
              <Lock className="w-3 h-3" />
              {lockedBadges.length} locked
            </span>
          )}
        </div>

        {showProgress && unlockedAchievements.length > 0 && (
          <div className="flex flex-wrap gap-1.5">
            {unlockedAchievements.slice(0, 4).map((ach) => (
              <span
                key={ach.id}
                className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs bg-green-50 dark:bg-green-900/20 text-green-700 dark:text-green-300 border border-green-200 dark:border-green-700"
              >
                <CheckCircle className="w-3 h-3" />
                {ach.icon} {ach.name}
              </span>
            ))}
          </div>
        )}
      </div>
    );
  }

  return (
    <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden">
      {/* Tabs */}
      <div className="flex border-b border-slate-200 dark:border-slate-700">
        <button
          onClick={() => setActiveTab("badges")}
          className={`flex-1 px-4 py-3 text-sm font-medium transition-colors ${
            activeTab === "badges"
              ? "text-indigo-600 dark:text-indigo-400 border-b-2 border-indigo-500"
              : "text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-300"
          }`}
        >
          Badges ({badges.length})
        </button>
        <button
          onClick={() => setActiveTab("achievements")}
          className={`flex-1 px-4 py-3 text-sm font-medium transition-colors ${
            activeTab === "achievements"
              ? "text-indigo-600 dark:text-indigo-400 border-b-2 border-indigo-500"
              : "text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-300"
          }`}
        >
          Achievements ({achievements.length})
        </button>
      </div>

      <div className="p-4">
        {shareMessage && (
          <div className="mb-3 px-3 py-2 rounded-lg bg-indigo-50 dark:bg-indigo-900/20 text-indigo-700 dark:text-indigo-300 text-sm text-center animate-pulse">
            {shareMessage}
          </div>
        )}

        {/* Badges Tab */}
        {activeTab === "badges" && (
          <div className="space-y-4">
            {/* Unlocked Badges */}
            {unlockedBadges.length > 0 && (
              <div>
                <h4 className="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider mb-3">
                  Unlocked
                </h4>
                <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
                  {unlockedBadges.map((badge) => (
                    <div
                      key={badge.id}
                      className={`relative group rounded-xl border-2 p-3 text-center transition-all duration-300 hover:shadow-lg ${
                        animatingBadge === badge.id
                          ? "animate-bounce border-amber-400 shadow-amber-200"
                          : "border-amber-200 dark:border-amber-700 bg-gradient-to-br from-amber-50 to-yellow-50 dark:from-amber-900/20 dark:to-yellow-900/20"
                      }`}
                      onClick={() => triggerAnimation(badge.id)}
                      onKeyDown={(e) => {
                        if (e.key === "Enter") triggerAnimation(badge.id);
                      }}
                      role="button"
                      tabIndex={0}
                      aria-label={`Badge: ${badge.name}`}
                    >
                      <div className="text-3xl mb-2">{badge.emoji}</div>
                      <div className="font-semibold text-sm text-slate-800 dark:text-slate-200">
                        {badge.name}
                      </div>
                      <div className="text-xs text-slate-500 dark:text-slate-400 mt-1 line-clamp-2">
                        {badge.description}
                      </div>
                      {badge.unlocked_at && (
                        <div className="mt-2 text-[10px] text-slate-400 dark:text-slate-500">
                          {new Date(badge.unlocked_at).toLocaleDateString()}
                        </div>
                      )}

                      {/* Share button */}
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          handleShare(badge);
                        }}
                        className="absolute top-2 right-2 p-1 rounded-full bg-white/80 dark:bg-slate-800/80 opacity-0 group-hover:opacity-100 transition-opacity hover:bg-indigo-50 dark:hover:bg-indigo-900/30"
                        aria-label={`Share ${badge.name} badge`}
                      >
                        <Share2 className="w-3 h-3 text-slate-500" />
                      </button>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Locked Badges */}
            {lockedBadges.length > 0 && (
              <div>
                <h4 className="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider mb-3">
                  Locked
                </h4>
                <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
                  {lockedBadges.map((badge) => (
                    <div
                      key={badge.id}
                      className="rounded-xl border border-slate-200 dark:border-slate-700 p-3 text-center bg-slate-50 dark:bg-slate-800/50 opacity-60"
                    >
                      <div className="text-3xl mb-2 flex items-center justify-center">
                        <span className="relative">
                          {badge.emoji}
                          <Lock className="absolute -top-1 -right-2 w-3 h-3 text-slate-400" />
                        </span>
                      </div>
                      <div className="font-semibold text-sm text-slate-500 dark:text-slate-400">
                        {badge.name}
                      </div>
                      <div className="text-xs text-slate-400 dark:text-slate-500 mt-1 line-clamp-2">
                        {badge.description}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {badges.length === 0 && (
              <div className="text-center py-8 text-slate-500 dark:text-slate-400">
                <p>No badges yet. Solve problems and climb the ranks!</p>
              </div>
            )}
          </div>
        )}

        {/* Achievements Tab */}
        {activeTab === "achievements" && (
          <div className="space-y-4">
            {/* Progress bars for locked achievements */}
            {showProgress && lockedAchievements.length > 0 && (
              <div className="space-y-3">
                <h4 className="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                  In Progress
                </h4>
                {lockedAchievements.map((achievement) => {
                  const progress = achievement.progress ?? 0;
                  const maxProgress = achievement.max_progress ?? 1;
                  const percentage = Math.min((progress / maxProgress) * 100, 100);

                  return (
                    <div key={achievement.id} className="space-y-1">
                      <div className="flex items-center justify-between text-sm">
                        <div className="flex items-center gap-2">
                          <span className="text-lg">{achievement.icon}</span>
                          <span className="font-medium text-slate-700 dark:text-slate-300">
                            {achievement.name}
                          </span>
                        </div>
                        <span className="text-xs text-slate-500 dark:text-slate-400">
                          {progress}/{maxProgress}
                        </span>
                      </div>
                      <div className="w-full bg-slate-200 dark:bg-slate-700 rounded-full h-2 overflow-hidden">
                        <div
                          className="h-full bg-gradient-to-r from-indigo-400 to-indigo-500 rounded-full transition-all duration-500"
                          style={{ width: `${percentage}%` }}
                        />
                      </div>
                      <p className="text-xs text-slate-500 dark:text-slate-400">{achievement.description}</p>
                    </div>
                  );
                })}
              </div>
            )}

            {/* Unlocked Achievements */}
            {unlockedAchievements.length > 0 && (
              <div>
                <h4 className="text-xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider mb-3">
                  Completed
                </h4>
                <div className="space-y-2">
                  {unlockedAchievements.map((achievement) => (
                    <div
                      key={achievement.id}
                      className="flex items-center justify-between p-3 rounded-lg bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-700 group"
                    >
                      <div className="flex items-center gap-3">
                        <span className="text-2xl">{achievement.icon}</span>
                        <div>
                          <div className="font-semibold text-sm text-slate-800 dark:text-slate-200">
                            {achievement.name}
                          </div>
                          <div className="text-xs text-slate-500 dark:text-slate-400">
                            {achievement.description}
                          </div>
                          {achievement.unlocked_at && (
                            <div className="text-[10px] text-green-600 dark:text-green-400 mt-0.5">
                              Unlocked {new Date(achievement.unlocked_at).toLocaleDateString()}
                            </div>
                          )}
                        </div>
                      </div>
                      <div className="flex items-center gap-2">
                        <CheckCircle className="w-5 h-5 text-green-500" />
                        <button
                          onClick={() => handleAchievementShare(achievement)}
                          className="p-1.5 rounded-lg opacity-0 group-hover:opacity-100 transition-opacity hover:bg-green-100 dark:hover:bg-green-800/30"
                          aria-label={`Share ${achievement.name} achievement`}
                        >
                          <Share2 className="w-3.5 h-3.5 text-green-600 dark:text-green-400" />
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {achievements.length === 0 && (
              <div className="text-center py-8 text-slate-500 dark:text-slate-400">
                <TrendingUp className="w-8 h-8 mx-auto mb-2 text-slate-300 dark:text-slate-600" />
                <p>No achievements yet. Keep solving problems!</p>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
