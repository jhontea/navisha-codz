import React, { memo } from "react";
import { Link } from "react-router-dom";
import {
  Users,
  FileCode,
  CheckCircle2,
  Activity,
  ArrowUpRight,
  ArrowDownRight,
  BarChart3,
  PlusCircle,
  Settings,
  RefreshCw,
  Server,
  Wifi,
  Database,
  Zap,
  Code2,
  UserCheck,
  BarChart,
  Shield,
} from "lucide-react";
import { Badge } from "../../components/ui/Badge";
import { Button } from "../../components/ui/Button";

// ── Types ──────────────────────────────────────────────────────────────────

interface StatCard {
  title: string;
  value: string | number;
  change: number;
  changeLabel: string;
  icon: React.ReactNode;
  color: string;
  gradient: string;
}

interface ActivityItem {
  id: string;
  type: "submission" | "user" | "problem" | "system";
  message: string;
  timestamp: string;
  user?: string;
  status?: "accepted" | "wrong_answer" | "pending" | "info";
}

interface ServerHealth {
  service: string;
  status: "healthy" | "degraded" | "down";
  latency: string;
  uptime: string;
  icon: React.ReactNode;
}

interface ChartDataPoint {
  label: string;
  submissions: number;
  users: number;
}

// ── Mock Data ───────────────────────────────────────────────────────────────

const STATS: StatCard[] = [
  {
    title: "Total Users",
    value: "2,847",
    change: 12.5,
    changeLabel: "vs last week",
    icon: <Users className="w-6 h-6" />,
    color: "from-blue-500 to-blue-600",
    gradient: "from-blue-50 to-blue-100 dark:from-blue-950/30 dark:to-blue-950/10",
  },
  {
    title: "Submissions",
    value: "18,432",
    change: 8.2,
    changeLabel: "vs last week",
    icon: <Code2 className="w-6 h-6" />,
    color: "from-emerald-500 to-emerald-600",
    gradient: "from-emerald-50 to-emerald-100 dark:from-emerald-950/30 dark:to-emerald-950/10",
  },
  {
    title: "Problems",
    value: "156",
    change: 3,
    changeLabel: "new this month",
    icon: <FileCode className="w-6 h-6" />,
    color: "from-violet-500 to-violet-600",
    gradient: "from-violet-50 to-violet-100 dark:from-violet-950/30 dark:to-violet-950/10",
  },
  {
    title: "Acceptance Rate",
    value: "67.3%",
    change: -2.1,
    changeLabel: "vs last week",
    icon: <CheckCircle2 className="w-6 h-6" />,
    color: "from-amber-500 to-amber-600",
    gradient: "from-amber-50 to-amber-100 dark:from-amber-950/30 dark:to-amber-950/10",
  },
];

const CHART_DATA: ChartDataPoint[] = [
  { label: "Mon", submissions: 420, users: 45 },
  { label: "Tue", submissions: 580, users: 52 },
  { label: "Wed", submissions: 490, users: 48 },
  { label: "Thu", submissions: 720, users: 63 },
  { label: "Fri", submissions: 610, users: 55 },
  { label: "Sat", submissions: 340, users: 28 },
  { label: "Sun", submissions: 280, users: 22 },
];

const RECENT_ACTIVITIES: ActivityItem[] = [
  {
    id: "1",
    type: "submission",
    message: 'Solved "Two Sum"',
    timestamp: "2 min ago",
    user: "alice_dev",
    status: "accepted",
  },
  {
    id: "2",
    type: "submission",
    message: 'Failed "Binary Tree Level Order"',
    timestamp: "5 min ago",
    user: "bob_coder",
    status: "wrong_answer",
  },
  {
    id: "3",
    type: "user",
    message: "New user registered",
    timestamp: "8 min ago",
    user: "charlie_new",
    status: "info",
  },
  {
    id: "4",
    type: "problem",
    message: 'Problem "Graph BFS" updated',
    timestamp: "15 min ago",
    user: "admin",
    status: "info",
  },
  {
    id: "5",
    type: "submission",
    message: 'Solved "LRU Cache"',
    timestamp: "20 min ago",
    user: "diana_eng",
    status: "accepted",
  },
  {
    id: "6",
    type: "system",
    message: "Server maintenance completed",
    timestamp: "1 hour ago",
    user: "system",
    status: "info",
  },
  {
    id: "7",
    type: "submission",
    message: 'Compilation error on "Valid Parentheses"',
    timestamp: "1 hour ago",
    user: "eve_dev",
    status: "wrong_answer",
  },
  {
    id: "8",
    type: "user",
    message: "Account upgraded to premium",
    timestamp: "2 hours ago",
    user: "frank_big",
    status: "info",
  },
];

const SERVER_HEALTH: ServerHealth[] = [
  {
    service: "API Server",
    status: "healthy",
    latency: "12ms",
    uptime: "99.97%",
    icon: <Server className="w-4 h-4" />,
  },
  {
    service: "WebSocket",
    status: "healthy",
    latency: "8ms",
    uptime: "99.99%",
    icon: <Wifi className="w-4 h-4" />,
  },
  {
    service: "Database",
    status: "healthy",
    latency: "3ms",
    uptime: "99.95%",
    icon: <Database className="w-4 h-4" />,
  },
  {
    service: "Judge Runner",
    status: "degraded",
    latency: "245ms",
    uptime: "98.2%",
    icon: <Zap className="w-4 h-4" />,
  },
];

// ── StatCard Component ──────────────────────────────────────────────────────

const DashboardStatCard = memo(function DashboardStatCard({
  stat,
  index: _index,
}: {
  stat: StatCard;
  index: number;
}) {
  const isPositive = stat.change >= 0;
  const TrendIcon = isPositive ? ArrowUpRight : ArrowDownRight;

  return (
    <div
      className={`relative overflow-hidden rounded-xl border border-slate-200 dark:border-slate-700 bg-gradient-to-br ${stat.gradient} p-6 transition-all duration-200 hover:shadow-lg hover:-translate-y-0.5`}
    >
      <div className="flex items-start justify-between">
        <div>
          <p className="text-sm font-medium text-slate-500 dark:text-slate-400">{stat.title}</p>
          <p className="text-3xl font-bold text-slate-900 dark:text-white mt-1">{stat.value}</p>
          <div className="flex items-center gap-1 mt-2">
            <TrendIcon
              className={`w-4 h-4 ${isPositive ? "text-emerald-500" : "text-red-500"}`}
            />
            <span
              className={`text-xs font-medium ${
                isPositive ? "text-emerald-600 dark:text-emerald-400" : "text-red-600 dark:text-red-400"
              }`}
            >
              {Math.abs(stat.change)}%
            </span>
            <span className="text-xs text-slate-400 dark:text-slate-500">{stat.changeLabel}</span>
          </div>
        </div>
        <div className={`p-3 rounded-lg bg-gradient-to-br ${stat.color} text-white shadow-lg`}>
          {stat.icon}
        </div>
      </div>
    </div>
  );
});

// ── Mini Chart Component ────────────────────────────────────────────────────

function SubmissionsChart({ data }: { data: ChartDataPoint[] }) {
  const maxValue = Math.max(...data.map((d) => d.submissions));
  const maxUsers = Math.max(...data.map((d) => d.users));

  return (
    <div className="space-y-3">
      {/* Submissions bars */}
      <div>
        <div className="flex items-center justify-between mb-2">
          <span className="text-xs font-medium text-slate-500 dark:text-slate-400">Submissions</span>
          <span className="text-xs text-slate-400">{maxValue}</span>
        </div>
        <div className="flex items-end gap-1.5 h-24">
          {data.map((point) => (
            <div key={point.label} className="flex-1 flex flex-col items-center gap-1">
              <div
                className="w-full rounded-t bg-gradient-to-t from-indigo-500 to-indigo-400 transition-all duration-300 hover:from-indigo-600 hover:to-indigo-500"
                style={{ height: `${(point.submissions / maxValue) * 100}%` }}
              />
              <span className="text-[10px] text-slate-400 dark:text-slate-500">{point.label}</span>
            </div>
          ))}
        </div>
      </div>

      {/* Users line */}
      <div>
        <div className="flex items-center justify-between mb-2">
          <span className="text-xs font-medium text-slate-500 dark:text-slate-400">Active Users</span>
          <span className="text-xs text-slate-400">{maxUsers}</span>
        </div>
        <div className="flex items-end gap-1.5 h-16">
          {data.map((point) => (
            <div key={point.label} className="flex-1 flex flex-col items-center gap-1">
              <div
                className="w-full rounded-t bg-gradient-to-t from-emerald-500 to-emerald-400 transition-all duration-300 hover:from-emerald-600 hover:to-emerald-500"
                style={{ height: `${(point.users / maxUsers) * 100}%` }}
              />
              <span className="text-[10px] text-slate-400">{point.label}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

// ── Activity Feed ───────────────────────────────────────────────────────────

const activityIcons: Record<string, React.ReactNode> = {
  submission: <Code2 className="w-4 h-4" />,
  user: <UserCheck className="w-4 h-4" />,
  problem: <FileCode className="w-4 h-4" />,
  system: <Settings className="w-4 h-4" />,
};

const activityColors: Record<string, string> = {
  submission: "bg-indigo-100 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400",
  user: "bg-emerald-100 dark:bg-emerald-900/30 text-emerald-600 dark:text-emerald-400",
  problem: "bg-amber-100 dark:bg-amber-900/30 text-amber-600 dark:text-amber-400",
  system: "bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400",
};

const ActivityFeed = memo(function ActivityFeed({
  activities,
}: {
  activities: ActivityItem[];
}) {
  return (
    <div className="space-y-1">
      {activities.map((item, idx) => (
        <div
          key={item.id}
          className={`flex items-start gap-3 p-3 rounded-lg transition-colors hover:bg-slate-50 dark:hover:bg-slate-800/50 ${
            idx < activities.length - 1 ? "border-b border-slate-100 dark:border-slate-800" : ""
          }`}
        >
          <div className={`p-1.5 rounded-lg ${activityColors[item.type] || ""}`}>
            {activityIcons[item.type]}
          </div>
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2">
              {item.user && (
                <span className="text-xs font-medium text-slate-900 dark:text-white">{item.user}</span>
              )}
              {item.status === "accepted" && (
                <Badge variant="status" status="accepted" size="sm" />
              )}
              {item.status === "wrong_answer" && (
                <Badge variant="status" status="wrong_answer" size="sm" />
              )}
            </div>
            <p className="text-sm text-slate-600 dark:text-slate-400 truncate">{item.message}</p>
            <span className="text-xs text-slate-400 dark:text-slate-500">{item.timestamp}</span>
          </div>
        </div>
      ))}
    </div>
  );
});

// ── Server Health ───────────────────────────────────────────────────────────

const statusConfig: Record<string, { label: string; color: string; dot: string }> = {
  healthy: {
    label: "Healthy",
    color: "bg-emerald-50 dark:bg-emerald-950/30 text-emerald-700 dark:text-emerald-400 border-emerald-200 dark:border-emerald-900",
    dot: "bg-emerald-500",
  },
  degraded: {
    label: "Degraded",
    color: "bg-amber-50 dark:bg-amber-950/30 text-amber-700 dark:text-amber-400 border-amber-200 dark:border-amber-900",
    dot: "bg-amber-500",
  },
  down: {
    label: "Down",
    color: "bg-red-50 dark:bg-red-950/30 text-red-700 dark:text-red-400 border-red-200 dark:border-red-900",
    dot: "bg-red-500",
  },
};

const ServerHealthCard = memo(function ServerHealthCard({
  services,
}: {
  services: ServerHealth[];
}) {
  return (
    <div className="space-y-3">
      {services.map((svc) => {
        const status = statusConfig[svc.status];
        return (
          <div
            key={svc.service}
            className={`flex items-center justify-between p-3 rounded-lg border ${status.color}`}
          >
            <div className="flex items-center gap-3">
              <span className="text-slate-500 dark:text-slate-400">{svc.icon}</span>
              <div>
                <p className="text-sm font-medium text-slate-900 dark:text-white">{svc.service}</p>
                <div className="flex items-center gap-2 mt-0.5">
                  <span className={`w-2 h-2 rounded-full ${status.dot}`} />
                  <span className="text-xs text-slate-500">{status.label}</span>
                </div>
              </div>
            </div>
            <div className="text-right">
              <p className="text-xs font-mono text-slate-600 dark:text-slate-400">{svc.latency}</p>
              <p className="text-[10px] text-slate-400">{svc.uptime} uptime</p>
            </div>
          </div>
        );
      })}
    </div>
  );
});

// ── Main Dashboard ──────────────────────────────────────────────────────────

export function AdminDashboard() {

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-slate-900 dark:text-white">Admin Dashboard</h1>
          <p className="text-sm text-slate-500 dark:text-slate-400 mt-1">
            Monitor and manage the coding challenge platform
          </p>
        </div>
        <div className="flex items-center gap-3">
          <Button
            variant="outline"
            size="sm"
            icon={<RefreshCw className="w-4 h-4" />}
            onClick={() => window.location.reload()}
          >
            Refresh
          </Button>
          <Link to="/admin/problems/new">
            <Button variant="primary" size="sm" icon={<PlusCircle className="w-4 h-4" />}>
              New Problem
            </Button>
          </Link>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {STATS.map((stat, idx) => (
          <DashboardStatCard key={stat.title} stat={stat} index={idx} />
        ))}
      </div>

      {/* Charts + Activity */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Submissions Chart */}
        <div className="lg:col-span-2 bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-6">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-lg font-semibold text-slate-900 dark:text-white flex items-center gap-2">
              <BarChart3 className="w-5 h-5 text-indigo-500" />
              Activity Overview
            </h2>
            <div className="flex items-center gap-4 text-xs text-slate-500">
              <span className="flex items-center gap-1">
                <span className="w-2.5 h-2.5 rounded-full bg-indigo-500" />
                Submissions
              </span>
              <span className="flex items-center gap-1">
                <span className="w-2.5 h-2.5 rounded-full bg-emerald-500" />
                Users
              </span>
            </div>
          </div>
          <SubmissionsChart data={CHART_DATA} />
        </div>

        {/* Recent Activity */}
        <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold text-slate-900 dark:text-white flex items-center gap-2">
              <Activity className="w-5 h-5 text-emerald-500" />
              Recent Activity
            </h2>
            <Button variant="ghost" size="sm">
              View all
            </Button>
          </div>
          <div className="max-h-[400px] overflow-y-auto -mx-2 px-2">
            <ActivityFeed activities={RECENT_ACTIVITIES} />
          </div>
        </div>
      </div>

      {/* Server Health */}
      <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-slate-900 dark:text-white flex items-center gap-2">
            <Shield className="w-5 h-5 text-green-500" />
            Server Health
          </h2>
          <Badge variant="status" status="running" size="sm" />
        </div>
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-3">
          <ServerHealthCard services={SERVER_HEALTH} />
        </div>
      </div>

      {/* Quick Actions */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <Link
          to="/admin/problems/new"
          className="flex items-center gap-4 p-4 rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 hover:border-indigo-300 dark:hover:border-indigo-700 transition-all group"
        >
          <div className="p-3 rounded-lg bg-indigo-50 dark:bg-indigo-950/30 text-indigo-600 dark:text-indigo-400 group-hover:scale-110 transition-transform">
            <PlusCircle className="w-5 h-5" />
          </div>
          <div>
            <p className="text-sm font-semibold text-slate-900 dark:text-white">Create Problem</p>
            <p className="text-xs text-slate-500">Add a new coding challenge</p>
          </div>
        </Link>
        <Link
          to="/problems"
          className="flex items-center gap-4 p-4 rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 hover:border-emerald-300 dark:hover:border-emerald-700 transition-all group"
        >
          <div className="p-3 rounded-lg bg-emerald-50 dark:bg-emerald-950/30 text-emerald-600 dark:text-emerald-400 group-hover:scale-110 transition-transform">
            <BarChart className="w-5 h-5" />
          </div>
          <div>
            <p className="text-sm font-semibold text-slate-900 dark:text-white">View Problems</p>
            <p className="text-xs text-slate-500">Browse and manage all problems</p>
          </div>
        </Link>
        <Link
          to="/leaderboard"
          className="flex items-center gap-4 p-4 rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 hover:border-violet-300 dark:hover:border-violet-700 transition-all group"
        >
          <div className="p-3 rounded-lg bg-violet-50 dark:bg-violet-950/30 text-violet-600 dark:text-violet-400 group-hover:scale-110 transition-transform">
            <Users className="w-5 h-5" />
          </div>
          <div>
            <p className="text-sm font-semibold text-slate-900 dark:text-white">Users & Ranking</p>
            <p className="text-xs text-slate-500">View leaderboard and user stats</p>
          </div>
        </Link>
      </div>
    </div>
  );
}

export default AdminDashboard;
