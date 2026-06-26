import { useEffect, useState } from "react";
import { CheckCircle, XCircle, Clock, Zap, HardDrive } from "lucide-react";
import { useSubmissionStore } from "../store/submissionStore";
import type { TestResult, SubmissionStatus } from "../types";

const statusConfig: Record<SubmissionStatus, { icon: typeof CheckCircle; color: string; label: string }> = {
  accepted: { icon: CheckCircle, color: "text-green-500", label: "Accepted" },
  wrong_answer: { icon: XCircle, color: "text-red-500", label: "Wrong Answer" },
  time_limit_exceeded: { icon: Clock, color: "text-orange-500", label: "Time Limit Exceeded" },
  memory_limit_exceeded: { icon: HardDrive, color: "text-orange-500", label: "Memory Limit Exceeded" },
  runtime_error: { icon: XCircle, color: "text-red-500", label: "Runtime Error" },
  compilation_error: { icon: XCircle, color: "text-red-500", label: "Compilation Error" },
  pending: { icon: Clock, color: "text-slate-400", label: "Pending" },
  running: { icon: Zap, color: "text-blue-500", label: "Running" },
};

function Confetti() {
  return (
    <div className="fixed inset-0 pointer-events-none z-50" aria-hidden="true">
      {Array.from({ length: 50 }).map((_, i) => (
        <div
          key={i}
          className="absolute w-2 h-2 rounded-full animate-confetti"
          style={{
            left: `${Math.random() * 100}%`,
            backgroundColor: ["#10b981", "#6366f1", "#f59e0b", "#ef4444", "#8b5cf6"][i % 5],
            animationDelay: `${Math.random() * 0.5}s`,
            animationDuration: `${1 + Math.random() * 1}s`,
          }}
        />
      ))}
    </div>
  );
}

export function TestResults() {
  const { currentSubmission, liveStatus, progress, completedTests, totalTests } = useSubmissionStore();
  const [showConfetti, setShowConfetti] = useState(false);

  const results = currentSubmission?.test_results ?? [];
  const isAccepted = liveStatus === "accepted";

  useEffect(() => {
    if (isAccepted && results.length > 0) {
      setShowConfetti(true);
      const timer = setTimeout(() => setShowConfetti(false), 3000);
      return () => clearTimeout(timer);
    }
  }, [isAccepted, results.length]);

  if (!currentSubmission && !liveStatus) {
    return (
      <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-6">
        <p className="text-slate-500 dark:text-slate-400 text-sm text-center">
          Submit your solution to see test results
        </p>
      </div>
    );
  }

  return (
    <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden">
      {showConfetti && <Confetti />}

      {/* Progress bar */}
      <div className="px-4 py-3 bg-slate-50 dark:bg-slate-800/50 border-b border-slate-200 dark:border-slate-700">
        <div className="flex items-center justify-between mb-2">
          <span className="text-sm font-medium text-slate-700 dark:text-slate-300">
            {liveStatus ? statusConfig[liveStatus].label : "Processing..."}
          </span>
          <span className="text-xs text-slate-500 dark:text-slate-400">
            {completedTests}/{totalTests} tests
          </span>
        </div>
        <div className="w-full bg-slate-200 dark:bg-slate-700 rounded-full h-2" role="progressbar" aria-valuenow={progress} aria-valuemin={0} aria-valuemax={100}>
          <div
            className={`h-2 rounded-full transition-all duration-300 ${
              isAccepted ? "bg-green-500" : "bg-indigo-500"
            }`}
            style={{ width: `${progress}%` }}
          />
        </div>
      </div>

      {/* Test results */}
      <div className="divide-y divide-slate-100 dark:divide-slate-800 max-h-[300px] overflow-y-auto">
        {results.length === 0 ? (
          <div className="p-4 text-center text-sm text-slate-500 dark:text-slate-400">
            <Zap className="w-5 h-5 mx-auto mb-1 animate-pulse text-indigo-500" />
            Running tests...
          </div>
        ) : (
          results.map((result, idx) => {
            const config = statusConfig[result.status];
            const Icon = config.icon;
            return (
              <div key={result.test_case_id} className="p-4">
                <div className="flex items-start gap-3">
                  <Icon className={`w-5 h-5 mt-0.5 shrink-0 ${config.color}`} />
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium text-slate-900 dark:text-white">
                        Test Case {idx + 1}
                      </span>
                      <span className="text-xs text-slate-500 dark:text-slate-400">{config.label}</span>
                    </div>
                    {result.status !== "accepted" && (
                      <div className="mt-2 space-y-1.5 font-mono text-xs">
                        <div className="flex gap-2">
                          <span className="text-slate-500 dark:text-slate-400 shrink-0">Expected:</span>
                          <span className="text-slate-800 dark:text-slate-200 break-all">{result.expected_output}</span>
                        </div>
                        {result.actual_output && (
                          <div className="flex gap-2">
                            <span className="text-slate-500 dark:text-slate-400 shrink-0">Got:</span>
                            <span className="text-red-600 dark:text-red-400 break-all">{result.actual_output}</span>
                          </div>
                        )}
                        {result.error_message && (
                          <div className="text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-950/30 rounded p-2 mt-1">
                            {result.error_message}
                          </div>
                        )}
                      </div>
                    )}
                    <div className="flex items-center gap-3 mt-2 text-xs text-slate-500 dark:text-slate-400">
                      <span className="flex items-center gap-1">
                        <Clock className="w-3 h-3" />
                        {result.execution_time_ms}ms
                      </span>
                      <span className="flex items-center gap-1">
                        <HardDrive className="w-3 h-3" />
                        {result.memory_used_kb}KB
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}
