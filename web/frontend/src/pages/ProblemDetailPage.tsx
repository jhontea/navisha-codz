import React, { useState } from "react";
import { useParams, Link } from "react-router-dom";
import { useQuery } from "react-query";
import { ArrowLeft, Clock, HardDrive, Star, BookOpen, CheckCircle, Smartphone, Monitor } from "lucide-react";
import { problemApi } from "../services/api";
import { CodeEditor } from "../components/CodeEditor";
import { PageLoader } from "../components/ui/LoadingSpinner";
import { CategoryIcon } from "../components/problem/CategoryIcon";
import { queryKeys } from "../hooks/useQueries";
import { useSubmissionStore } from "../store/submissionStore";
import { TestResults } from "../components/TestResults";
import { HintPanel } from "../components/HintPanel";
import type { Problem } from "../types";

export function ProblemDetailPage() {
  const { slug } = useParams<{ slug: string }>();
  const [code, setCode] = useState("");
  const [activeTab, setActiveTab] = useState<"description" | "submissions">("description");
  const [splitView, setSplitView] = useState<"split" | "full">("split");
  const { isSubmitting } = useSubmissionStore();

  const { data: problem, isLoading, error } = useQuery(
    queryKeys.problems.bySlug(slug ?? ""),
    () => problemApi.getBySlug(slug ?? "").then((r) => r.data),
    { enabled: !!slug, staleTime: 5 * 60 * 1000 }
  );

  const handleSubmit = () => {
    if (problem) {
      problemApi.submit({ problem_id: problem.id, code, language: "go" }).then((r) => {
        console.log("Submitted:", r);
      });
    }
  };

  if (isLoading) return <PageLoader />;
  if (error || !problem) {
    return (
      <div className="text-center py-12">
        <p className="text-red-500 dark:text-red-400">Failed to load problem. Please try again.</p>
        <Link to="/problems" className="text-indigo-600 dark:text-indigo-400 hover:underline mt-2 inline-block">
          Back to problems
        </Link>
      </div>
    );
  }

  return (
    <div className="h-[calc(100vh-8rem)] flex flex-col space-y-3">
      {/* Back link + mobile view toggle */}
      <div className="flex items-center justify-between shrink-0">
        <Link
          to="/problems"
          className="inline-flex items-center gap-1 text-sm text-slate-500 dark:text-slate-400 hover:text-slate-700 dark:hover:text-slate-200"
        >
          <ArrowLeft className="w-4 h-4" />
          Back to Problems
        </Link>
        {/* View mode toggle (mobile) */}
        <div className="flex items-center gap-1 lg:hidden">
          <button
            onClick={() => setSplitView("full")}
            className={`p-1.5 rounded text-xs ${splitView === "full" ? "bg-indigo-100 dark:bg-indigo-900/30 text-indigo-700" : "text-slate-500"}`}
            aria-label="Full view"
          >
            <Smartphone className="w-4 h-4" />
          </button>
          <button
            onClick={() => setSplitView("split")}
            className={`p-1.5 rounded text-xs ${splitView === "split" ? "bg-indigo-100 dark:bg-indigo-900/30 text-indigo-700" : "text-slate-500"}`}
            aria-label="Split view"
          >
            <Monitor className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Two-panel layout */}
      <div className={`flex-1 flex flex-col lg:flex-row gap-4 overflow-hidden ${splitView === "full" ? "lg:flex-col" : ""}`}>
        {/* Left panel — Problem info */}
        <div className={`flex flex-col overflow-y-auto space-y-3 ${splitView === "full" ? "lg:max-h-[40vh]" : "lg:w-1/2"}`}>
          {/* Tabs */}
          <div className="flex gap-1 bg-white dark:bg-slate-900 rounded-lg border border-slate-200 dark:border-slate-700 p-1">
            <button
              onClick={() => setActiveTab("description")}
              className={`flex-1 flex items-center justify-center gap-1.5 px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                activeTab === "description"
                  ? "bg-indigo-50 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-400"
                  : "text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white"
              }`}
              aria-selected={activeTab === "description"}
              role="tab"
              style={{ minHeight: "40px" }}
            >
              <BookOpen className="w-4 h-4" />
              <span className="hidden sm:inline">Description</span>
            </button>
            <button
              onClick={() => setActiveTab("submissions")}
              className={`flex-1 flex items-center justify-center gap-1.5 px-3 py-2 rounded-md text-sm font-medium transition-colors ${
              activeTab === "submissions"
                ? "bg-indigo-50 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-400"
                : "text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white"
              }`}
              aria-selected={activeTab === "submissions"}
              role="tab"
              style={{ minHeight: "40px" }}
            >
              <CheckCircle className="w-4 h-4" />
              <span className="hidden sm:inline">Submissions</span>
            </button>
          </div>

          {activeTab === "description" ? (
            <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-4 sm:p-6 space-y-4 sm:space-y-6">
              {/* Title + meta */}
              <div>
                <div className="flex flex-wrap items-start justify-between gap-2">
                  <h1 className="text-xl sm:text-2xl font-bold text-slate-900 dark:text-white">{problem.title}</h1>
                  <CategoryIcon category={problem.category} size="md" />
                </div>
                <div className="flex flex-wrap items-center gap-2 mt-2">
                  <span
                    className={`px-2.5 py-0.5 rounded-full text-xs font-medium ${
                      problem.difficulty === "easy"
                        ? "bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400"
                        : problem.difficulty === "medium"
                        ? "bg-yellow-100 dark:bg-yellow-900/30 text-yellow-700 dark:text-yellow-400"
                        : "bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400"
                    }`}
                  >
                    {problem.difficulty.charAt(0).toUpperCase() + problem.difficulty.slice(1)}
                  </span>
                  <span className="text-xs sm:text-sm text-slate-500 dark:text-slate-400">{problem.points} points</span>
                  <span className="text-xs sm:text-sm text-slate-400 dark:text-slate-500 flex items-center gap-1">
                    <Clock className="w-3 h-3" />
                    {problem.time_limit_ms}ms
                  </span>
                  <span className="text-xs sm:text-sm text-slate-400 dark:text-slate-500 flex items-center gap-1">
                    <HardDrive className="w-3 h-3" />
                    {problem.memory_limit_mb}MB
                  </span>
                  <span className="text-xs sm:text-sm text-slate-400 dark:text-slate-500 flex items-center gap-1">
                    <Star className="w-3 h-3" />
                    {problem.solved_count} solved
                  </span>
                </div>
              </div>

              {/* Description */}
              <div className="prose prose-sm dark:prose-invert max-w-none">
                <div className="whitespace-pre-wrap text-slate-700 dark:text-slate-300 leading-relaxed text-sm sm:text-base">
                  {problem.description}
                </div>
              </div>

              {/* Examples */}
              {problem.examples.map((example, idx) => (
                <div key={example.id} className="space-y-2">
                  <h3 className="text-sm font-semibold text-slate-900 dark:text-white">
                    Example {idx + 1}:
                  </h3>
                  <div className="bg-slate-50 dark:bg-slate-800/50 rounded-lg p-3 sm:p-4 space-y-2 font-mono text-xs sm:text-sm">
                    <div>
                      <span className="text-slate-500 dark:text-slate-400 font-sans text-xs font-medium">Input: </span>
                      <span className="text-slate-800 dark:text-slate-200">{example.input}</span>
                    </div>
                    <div>
                      <span className="text-slate-500 dark:text-slate-400 font-sans text-xs font-medium">Output: </span>
                      <span className="text-slate-800 dark:text-slate-200">{example.output}</span>
                    </div>
                    {example.explanation && (
                      <div className="pt-1 border-t border-slate-200 dark:border-slate-700">
                        <span className="text-slate-500 dark:text-slate-400 font-sans text-xs font-medium">Explanation: </span>
                        <span className="text-slate-700 dark:text-slate-300 font-sans">{example.explanation}</span>
                      </div>
                    )}
                  </div>
                </div>
              ))}

              {/* Constraints */}
              {problem.constraints.length > 0 && (
                <div>
                  <h3 className="text-sm font-semibold text-slate-900 dark:text-white mb-2">Constraints:</h3>
                  <ul className="list-disc list-inside space-y-1 text-sm text-slate-700 dark:text-slate-300">
                    {problem.constraints.map((constraint, idx) => (
                      <li key={idx}>{constraint}</li>
                    ))}
                  </ul>
                </div>
              )}

              {/* Hints */}
              {problem.hints.length > 0 && (
                <HintPanel problemId={problem.id} hints={problem.hints} />
              )}
            </div>
          ) : (
            <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-6">
              <p className="text-slate-500 dark:text-slate-400 text-sm">Your submission history for this problem will appear here.</p>
            </div>
          )}
        </div>

        {/* Right panel — Code editor + test results */}
        <div className={`flex flex-col space-y-3 ${splitView === "full" ? "lg:w-full" : "lg:w-1/2"}`}>
          {/* Mobile tab toggle for code vs results */}
          <div className="flex lg:hidden gap-1 bg-white dark:bg-slate-900 rounded-lg border border-slate-200 dark:border-slate-700 p-1">
            <button
              onClick={() => setActiveTab("description")}
              className={`flex-1 px-3 py-2 rounded-md text-xs font-medium transition-colors ${
                activeTab === "description"
                  ? "bg-indigo-50 dark:bg-indigo-900/30 text-indigo-700"
                  : "text-slate-500"
              }`}
            >
              Code
            </button>
            <button
              onClick={() => setActiveTab("submissions")}
              className={`flex-1 px-3 py-2 rounded-md text-xs font-medium transition-colors ${
                activeTab === "submissions"
                  ? "bg-indigo-50 dark:bg-indigo-900/30 text-indigo-700"
                  : "text-slate-500"
              }`}
            >
              Results
            </button>
          </div>

          <CodeEditor
            value={code}
            onChange={setCode}
            language="go"
            template={problem.function_template}
            onSubmit={handleSubmit}
          />
          <TestResults />
        </div>
      </div>
    </div>
  );
}
