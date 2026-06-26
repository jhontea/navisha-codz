import React, { useState } from 'react';

export interface TestCase {
  id: number;
  name?: string;
  passed: boolean;
  input?: string;
  expectedOutput?: string;
  actualOutput?: string;
  errorMessage?: string;
  executionTime?: string;
  memoryUsage?: string;
}

export interface TestCaseResultProps {
  testCases: TestCase[];
  showDetails?: boolean;
  className?: string;
}

export const TestCaseResult: React.FC<TestCaseResultProps> = ({
  testCases,
  showDetails = true,
  className = '',
}) => {
  const [expandedCase, setExpandedCase] = useState<number | null>(null);

  const passedCount = testCases.filter((tc) => tc.passed).length;
  const allPassed = passedCount === testCases.length;

  const toggleExpand = (id: number) => {
    setExpandedCase(expandedCase === id ? null : id);
  };

  return (
    <div className={`space-y-3 ${className}`}>
      {/* Summary */}
      <div className={`flex items-center gap-3 p-3 rounded-lg ${allPassed ? 'bg-success-50 dark:bg-success-900/20' : 'bg-error-50 dark:bg-error-900/20'}`}>
        <svg
          className={`w-5 h-5 flex-shrink-0 ${allPassed ? 'text-success-500' : 'text-error-500'}`}
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          {allPassed ? (
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
          ) : (
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
          )}
        </svg>
        <span className={`text-sm font-medium ${allPassed ? 'text-success-700 dark:text-success-300' : 'text-error-700 dark:text-error-300'}`}>
          {passedCount}/{testCases.length} test cases passed
        </span>
      </div>

      {/* Test Case List */}
      <div className="space-y-2">
        {testCases.map((testCase) => {
          const isExpanded = expandedCase === testCase.id;
          return (
            <div
              key={testCase.id}
              className={`border rounded-lg overflow-hidden transition-colors ${
                testCase.passed
                  ? 'border-success-200 dark:border-success-800'
                  : 'border-error-200 dark:border-error-800'
              }`}
            >
              {/* Header */}
              <button
                onClick={() => toggleExpand(testCase.id)}
                className={`w-full flex items-center gap-3 p-3 text-left transition-colors ${
                  testCase.passed
                    ? 'bg-success-50/50 dark:bg-success-900/10 hover:bg-success-50 dark:hover:bg-success-900/20'
                    : 'bg-error-50/50 dark:bg-error-900/10 hover:bg-error-50 dark:hover:bg-error-900/20'
                }`}
                aria-expanded={isExpanded}
              >
                {/* Pass/Fail Icon */}
                <span className={`flex-shrink-0 w-6 h-6 rounded-full flex items-center justify-center ${
                  testCase.passed ? 'bg-success-500' : 'bg-error-500'
                }`}>
                  {testCase.passed ? (
                    <svg className="w-4 h-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
                    </svg>
                  ) : (
                    <svg className="w-4 h-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  )}
                </span>

                {/* Test Case Name */}
                <span className="flex-1 text-sm font-medium text-neutral-900 dark:text-neutral-100">
                  {testCase.name || `Test Case #${testCase.id}`}
                </span>

                {/* Execution Time */}
                {testCase.executionTime && (
                  <span className="text-xs text-neutral-500 dark:text-neutral-400">
                    {testCase.executionTime}
                  </span>
                )}

                {/* Expand Arrow */}
                <svg
                  className={`w-4 h-4 text-neutral-400 transition-transform ${isExpanded ? 'rotate-180' : ''}`}
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                </svg>
              </button>

              {/* Expanded Details */}
              {isExpanded && showDetails && (
                <div className="p-4 border-t border-neutral-200 dark:border-neutral-700 bg-white dark:bg-neutral-900 space-y-4">
                  {/* Input */}
                  {testCase.input && (
                    <div>
                      <h4 className="text-xs font-semibold text-neutral-500 dark:text-neutral-400 uppercase tracking-wider mb-1">
                        Input
                      </h4>
                      <pre className="p-2 text-xs font-mono bg-neutral-50 dark:bg-neutral-800 rounded border border-neutral-200 dark:border-neutral-700 overflow-x-auto">
                        {testCase.input}
                      </pre>
                    </div>
                  )}

                  {/* Expected vs Actual */}
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                    <div>
                      <h4 className="text-xs font-semibold text-neutral-500 dark:text-neutral-400 uppercase tracking-wider mb-1">
                        Expected Output
                      </h4>
                      <pre className="p-2 text-xs font-mono bg-success-50 dark:bg-success-900/20 border border-success-200 dark:border-success-800 rounded overflow-x-auto">
                        {testCase.expectedOutput}
                      </pre>
                    </div>
                    <div>
                      <h4 className="text-xs font-semibold text-neutral-500 dark:text-neutral-400 uppercase tracking-wider mb-1">
                        Your Output
                      </h4>
                      <pre className={`p-2 text-xs font-mono rounded overflow-x-auto ${
                        testCase.passed
                          ? 'bg-success-50 dark:bg-success-900/20 border border-success-200 dark:border-success-800'
                          : 'bg-error-50 dark:bg-error-900/20 border border-error-200 dark:border-error-800'
                      }`}>
                        {testCase.actualOutput}
                      </pre>
                    </div>
                  </div>

                  {/* Diff View (when outputs differ) */}
                  {!testCase.passed && testCase.expectedOutput && testCase.actualOutput && testCase.expectedOutput !== testCase.actualOutput && (
                    <div>
                      <h4 className="text-xs font-semibold text-neutral-500 dark:text-neutral-400 uppercase tracking-wider mb-1">
                        Difference
                      </h4>
                      <div className="p-2 text-xs font-mono bg-error-50 dark:bg-error-900/20 border border-error-200 dark:border-error-800 rounded overflow-x-auto">
                        <div className="text-error-600 dark:text-error-400">- Expected: {testCase.expectedOutput}</div>
                        <div className="text-success-600 dark:text-success-400">+ Got:      {testCase.actualOutput}</div>
                      </div>
                    </div>
                  )}

                  {/* Error Message */}
                  {testCase.errorMessage && (
                    <div>
                      <h4 className="text-xs font-semibold text-neutral-500 dark:text-neutral-400 uppercase tracking-wider mb-1">
                        Error
                      </h4>
                      <pre className="p-2 text-xs font-mono bg-error-50 dark:bg-error-900/20 border border-error-200 dark:border-error-800 rounded text-error-700 dark:text-error-300 overflow-x-auto whitespace-pre-wrap">
                        {testCase.errorMessage}
                      </pre>
                    </div>
                  )}

                  {/* Memory Usage */}
                  {testCase.memoryUsage && (
                    <div className="flex items-center gap-2 text-xs text-neutral-500 dark:text-neutral-400">
                      <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m16-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z" />
                      </svg>
                      <span>Memory: {testCase.memoryUsage}</span>
                    </div>
                  )}
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
};

TestCaseResult.displayName = 'TestCaseResult';

export default TestCaseResult;
