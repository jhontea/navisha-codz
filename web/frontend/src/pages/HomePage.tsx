import { useQuery } from "react-query";
import { Link } from "react-router-dom";
import { ArrowRight, Code2, Trophy, Users, Zap, BookOpen, Target } from "lucide-react";
import { problemApi } from "../services/api";
import { useAuthStore } from "../store/authStore";

const stats = [
  { label: "Problems", value: "500+", icon: BookOpen },
  { label: "Active Users", value: "10K+", icon: Users },
  { label: "Submissions", value: "1M+", icon: Code2 },
  { label: "Daily Contests", value: "3", icon: Trophy },
];

const features = [
  {
    title: "Algorithm Mastery",
    description: "Practice sorting, searching, graph algorithms, and dynamic programming.",
    icon: Target,
  },
  {
    title: "Real-time Feedback",
    description: "Get instant test results with detailed execution metrics.",
    icon: Zap,
  },
  {
    title: "Compete & Climb",
    description: "Join weekly contests and climb the global leaderboard.",
    icon: Trophy,
  },
];

export function HomePage() {
  const { isAuthenticated } = useAuthStore();

  const { data: featuredData } = useQuery("featured-problems", () =>
    problemApi.list({ page: 1, page_size: 6 })
  );

  const featuredProblems = featuredData?.data?.items ?? [];

  return (
    <div className="space-y-12 sm:space-y-16">
      {/* Hero Section */}
      <section className="text-center py-12 sm:py-16 md:py-24">
        <div className="max-w-3xl mx-auto px-4">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-indigo-50 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-400 text-sm font-medium mb-4 sm:mb-6">
            <Zap className="w-3.5 h-3.5" />
            Master Algorithms Through Practice
          </div>
          <h1 className="text-3xl sm:text-4xl md:text-6xl font-bold text-slate-900 dark:text-white leading-tight">
            Code. Compete.{" "}
            <span className="text-indigo-600 dark:text-indigo-400">Conquer.</span>
          </h1>
          <p className="mt-4 sm:mt-6 text-base sm:text-lg text-slate-600 dark:text-slate-400 max-w-2xl mx-auto px-4">
            Sharpen your algorithmic thinking with curated coding challenges.
            Track your progress, compete with peers, and ace your technical interviews.
          </p>
          <div className="mt-6 sm:mt-8 flex flex-col sm:flex-row items-center justify-center gap-3 sm:gap-4">
            {isAuthenticated ? (
              <Link
                to="/problems"
                className="inline-flex items-center gap-2 px-6 py-3 bg-indigo-600 text-white font-semibold rounded-xl hover:bg-indigo-700 transition-colors shadow-lg shadow-indigo-200 dark:shadow-indigo-900/30 min-h-[44px]"
              >
                Start Solving
                <ArrowRight className="w-4 h-4" />
              </Link>
            ) : (
              <>
                <Link
                  to="/register"
                  className="inline-flex items-center gap-2 px-6 py-3 bg-indigo-600 text-white font-semibold rounded-xl hover:bg-indigo-700 transition-colors shadow-lg shadow-indigo-200 dark:shadow-indigo-900/30 min-h-[44px]"
                >
                  Get Started Free
                  <ArrowRight className="w-4 h-4" />
                </Link>
                <Link
                  to="/problems"
                  className="inline-flex items-center gap-2 px-6 py-3 bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-300 font-semibold rounded-xl border border-slate-200 dark:border-slate-700 hover:border-slate-300 dark:hover:border-slate-600 transition-colors min-h-[44px]"
                >
                  Browse Problems
                </Link>
              </>
            )}
          </div>
        </div>
      </section>

      {/* Stats */}
      <section className="bg-white dark:bg-slate-900 rounded-2xl border border-slate-200 dark:border-slate-700 p-6 sm:p-8">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-6 sm:gap-8">
          {stats.map((stat) => {
            const Icon = stat.icon;
            return (
              <div key={stat.label} className="text-center">
                <div className="inline-flex items-center justify-center w-10 h-10 sm:w-12 sm:h-12 rounded-xl bg-indigo-50 dark:bg-indigo-900/30 mb-2 sm:mb-3">
                  <Icon className="w-5 h-5 sm:w-6 sm:h-6 text-indigo-600 dark:text-indigo-400" />
                </div>
                <p className="text-xl sm:text-2xl font-bold text-slate-900 dark:text-white">{stat.value}</p>
                <p className="text-xs sm:text-sm text-slate-500 dark:text-slate-400 mt-0.5">{stat.label}</p>
              </div>
            );
          })}
        </div>
      </section>

      {/* Featured Problems */}
      {featuredProblems.length > 0 && (
        <section>
          <div className="flex items-center justify-between mb-4 sm:mb-6">
            <h2 className="text-xl sm:text-2xl font-bold text-slate-900 dark:text-white">Featured Problems</h2>
            <Link
              to="/problems"
              className="text-sm font-medium text-indigo-600 dark:text-indigo-400 hover:text-indigo-700 dark:hover:text-indigo-300 flex items-center gap-1"
            >
              View all
              <ArrowRight className="w-3.5 h-3.5" />
            </Link>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 sm:gap-4">
            {featuredProblems.map((problem) => (
              <Link
                key={problem.id}
                to={`/problems/${problem.slug}`}
                className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-4 sm:p-5 hover:border-indigo-300 dark:hover:border-indigo-700 hover:shadow-sm transition-all"
              >
                <div className="flex items-start justify-between">
                  <h3 className="font-semibold text-slate-900 dark:text-white text-sm sm:text-base">{problem.title}</h3>
                  <span
                    className={`px-2 py-0.5 rounded-full text-xs font-medium shrink-0 ml-2 ${
                      problem.difficulty === "easy"
                        ? "bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400"
                        : problem.difficulty === "medium"
                        ? "bg-yellow-100 dark:bg-yellow-900/30 text-yellow-700 dark:text-yellow-400"
                        : "bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400"
                    }`}
                  >
                    {problem.difficulty}
                  </span>
                </div>
                <p className="text-xs sm:text-sm text-slate-500 dark:text-slate-400 mt-2 line-clamp-2">
                  {problem.description.substring(0, 100)}...
                </p>
                <div className="flex items-center gap-3 mt-3 text-xs text-slate-400 dark:text-slate-500">
                  <span>{problem.points} pts</span>
                  <span>{problem.solved_count} solved</span>
                </div>
              </Link>
            ))}
          </div>
        </section>
      )}

      {/* Features */}
      <section>
        <h2 className="text-xl sm:text-2xl font-bold text-slate-900 dark:text-white text-center mb-6 sm:mb-8">
          Why CodeChallenge?
        </h2>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 sm:gap-6">
          {features.map((feature) => {
            const Icon = feature.icon;
            return (
              <div
                key={feature.title}
                className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-5 sm:p-6 text-center"
              >
                <div className="inline-flex items-center justify-center w-10 h-10 sm:w-12 sm:h-12 rounded-xl bg-indigo-50 dark:bg-indigo-900/30 mb-3 sm:mb-4">
                  <Icon className="w-5 h-5 sm:w-6 sm:h-6 text-indigo-600 dark:text-indigo-400" />
                </div>
                <h3 className="font-semibold text-slate-900 dark:text-white">{feature.title}</h3>
                <p className="text-sm text-slate-600 dark:text-slate-400 mt-2">{feature.description}</p>
              </div>
            );
          })}
        </div>
      </section>

      {/* Quick Start */}
      <section className="bg-gradient-to-br from-indigo-600 to-purple-700 rounded-2xl p-6 sm:p-8 md:p-12 text-white">
        <div className="max-w-2xl mx-auto text-center">
          <h2 className="text-2xl sm:text-3xl font-bold">Ready to Start?</h2>
          <p className="mt-3 text-indigo-100 text-sm sm:text-base">
            Join thousands of developers improving their coding skills every day.
          </p>
          <div className="mt-6 sm:mt-8 grid grid-cols-1 sm:grid-cols-3 gap-3 sm:gap-4 text-left">
            <div className="bg-white/10 rounded-lg p-3 sm:p-4">
              <p className="text-xl sm:text-2xl font-bold">1</p>
              <p className="text-xs sm:text-sm text-indigo-100 mt-1">Create your free account</p>
            </div>
            <div className="bg-white/10 rounded-lg p-3 sm:p-4">
              <p className="text-xl sm:text-2xl font-bold">2</p>
              <p className="text-xs sm:text-sm text-indigo-100 mt-1">Pick a problem to solve</p>
            </div>
            <div className="bg-white/10 rounded-lg p-3 sm:p-4">
              <p className="text-xl sm:text-2xl font-bold">3</p>
              <p className="text-xs sm:text-sm text-indigo-100 mt-1">Submit and climb the ranks</p>
            </div>
          </div>
        </div>
      </section>
    </div>
  );
}
