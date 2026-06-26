import { useState } from "react";
import { BookOpen, CheckCircle } from "lucide-react";
import type { Problem } from "../types";
import { CodeEditor } from "./CodeEditor";
import { TestResults } from "./TestResults";
import { HintPanel } from "./HintPanel";

interface ProblemDetailProps {
  problem: Problem;
  onSubmit: (code: string) => void;
  isSubmitting: boolean;
}

export function ProblemDetail({ problem, onSubmit, isSubmitting: _isSubmitting }: ProblemDetailProps) {
  const [code, setCode] = useState(problem.function_template);
  const [activeTab, setActiveTab] = useState<"description" | "submissions">("description");

  const handleSubmit = () => {
    onSubmit(code);
  };

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 sm:gap-6 h-full">
      {/* Left panel — Problem info */}
      <div className="flex flex-col overflow-y-auto space-y-4">
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
            Description
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
            Submissions
          </button>
        </div>

        {activeTab === "description" ? (
          <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-4 sm:p-6 space-y-4 sm:space-y-6">
            {/* Title */}
            <div>
              <h1 className="text-xl sm:text-2xl font-bold text-slate-900 dark:text-white">{problem.title}</h1>
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
                <span className="text-xs sm:text-sm text-slate-400 dark:text-slate-500">
                  Time: {problem.time_limit_ms}ms
                </span>
                <span className="text-xs sm:text-sm text-slate-400 dark:text-slate-500">
                  Memory: {problem.memory_limit_mb}MB
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

      {/* Right panel — Code editor */}
      <div className="flex flex-col space-y-4">
        <CodeEditor
          value={code}
          onChange={setCode}
          language="go"
          theme="vs-dark"
          template={problem.function_template}
          onSubmit={handleSubmit}
        />

        {/* Test Results */}
        <TestResults />
      </div>
    </div>
  );
}
