import { useState, useEffect, useCallback } from "react";
import { Play, Pause, SkipForward, SkipBack, RotateCcw } from "lucide-react";

interface DPVisualizationProps {
  table: (number | string)[][];
  rowLabels?: string[];
  colLabels?: string[];
  highlightCells?: Array<{ row: number; col: number }>;
  memoizationOrder?: Array<{ row: number; col: number }>;
}

export function DPVisualization({
  table,
  rowLabels,
  colLabels,
  highlightCells = [],
  memoizationOrder = [],
}: DPVisualizationProps) {
  const [currentStep, setCurrentStep] = useState(0);
  const [isPlaying, setIsPlaying] = useState(false);
  const [viewMode, setViewMode] = useState<"tabulation" | "memoization">("tabulation");
  const [visitedCells, setVisitedCells] = useState<Set<string>>(new Set());

  const totalSteps = viewMode === "tabulation" ? table.length * (table[0]?.length ?? 0) : memoizationOrder.length;

  const getCellKey = (row: number, col: number) => `${row}-${col}`;

  const updateVisitedCells = useCallback(
    (step: number) => {
      if (viewMode === "tabulation") {
        const cols = table[0]?.length ?? 0;
        const newVisited = new Set<string>();
        for (let i = 0; i < step; i++) {
          const row = Math.floor(i / cols);
          const col = i % cols;
          newVisited.add(getCellKey(row, col));
        }
        setVisitedCells(newVisited);
      } else {
        const newVisited = new Set<string>();
        for (let i = 0; i < step; i++) {
          const cell = memoizationOrder[i];
          if (cell) newVisited.add(getCellKey(cell.row, cell.col));
        }
        setVisitedCells(newVisited);
      }
    },
    [viewMode, table, memoizationOrder]
  );

  useEffect(() => {
    updateVisitedCells(currentStep);
  }, [currentStep, updateVisitedCells]);

  useEffect(() => {
    setCurrentStep(0);
    setVisitedCells(new Set());
    setIsPlaying(false);
  }, [viewMode]);

  useEffect(() => {
    if (!isPlaying) return;
    if (currentStep >= totalSteps) {
      setIsPlaying(false);
      return;
    }

    const timer = setTimeout(() => {
      setCurrentStep((s) => s + 1);
    }, 500);

    return () => clearTimeout(timer);
  }, [isPlaying, currentStep, totalSteps]);

  const handlePlayPause = () => {
    if (currentStep >= totalSteps) {
      setCurrentStep(0);
    }
    setIsPlaying(!isPlaying);
  };

  const handleStepForward = () => {
    if (currentStep < totalSteps) {
      setCurrentStep((s) => s + 1);
    }
  };

  const handleStepBack = () => {
    if (currentStep > 0) {
      setCurrentStep((s) => s - 1);
    }
  };

  const handleReset = () => {
    setCurrentStep(0);
    setIsPlaying(false);
  };

  const isHighlighted = (row: number, col: number) =>
    highlightCells.some((c) => c.row === row && c.col === col);

  const isVisited = (row: number, col: number) => visitedCells.has(getCellKey(row, col));

  const getCellColor = (row: number, col: number) => {
    if (isHighlighted(row, col)) return "bg-indigo-200 dark:bg-indigo-800 border-indigo-400 dark:border-indigo-600";
    if (isVisited(row, col)) return "bg-green-100 dark:bg-green-900/30 border-green-300 dark:border-green-700";
    return "bg-white dark:bg-slate-800 border-slate-200 dark:border-slate-700";
  };

  return (
    <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 p-6 space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold text-slate-900 dark:text-white">DP Table Visualization</h3>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setViewMode("tabulation")}
            className={`px-3 py-1 text-xs font-medium rounded-md transition-colors ${
              viewMode === "tabulation"
                ? "bg-indigo-100 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-400"
                : "text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800"
            }`}
          >
            Tabulation
          </button>
          <button
            onClick={() => setViewMode("memoization")}
            className={`px-3 py-1 text-xs font-medium rounded-md transition-colors ${
              viewMode === "memoization"
                ? "bg-indigo-100 dark:bg-indigo-900/30 text-indigo-700 dark:text-indigo-400"
                : "text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800"
            }`}
          >
            Memoization
          </button>
        </div>
      </div>

      {/* Controls */}
      <div className="flex items-center justify-center gap-2">
        <button
          onClick={handleReset}
          className="p-2 rounded-lg text-slate-600 hover:bg-slate-100 transition-colors"
          aria-label="Reset"
        >
          <RotateCcw className="w-4 h-4" />
        </button>
        <button
          onClick={handleStepBack}
          disabled={currentStep <= 0}
          className="p-2 rounded-lg text-slate-600 hover:bg-slate-100 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          aria-label="Step back"
        >
          <SkipBack className="w-4 h-4" />
        </button>
        <button
          onClick={handlePlayPause}
          className="p-2.5 rounded-lg bg-indigo-600 text-white hover:bg-indigo-700 transition-colors"
          aria-label={isPlaying ? "Pause" : "Play"}
        >
          {isPlaying ? <Pause className="w-4 h-4" /> : <Play className="w-4 h-4" />}
        </button>
        <button
          onClick={handleStepForward}
          disabled={currentStep >= totalSteps}
          className="p-2 rounded-lg text-slate-600 hover:bg-slate-100 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          aria-label="Step forward"
        >
          <SkipForward className="w-4 h-4" />
        </button>
        <span className="text-xs text-slate-500 ml-2">
          Step {currentStep}/{totalSteps}
        </span>
      </div>

      {/* Table */}
      <div className="overflow-x-auto">
        <table className="border-collapse mx-auto" role="grid" aria-label="DP table">
          <thead>
            <tr>
              <th className="w-10 h-10" />
              {colLabels?.map((label, idx) => (
                <th
                  key={idx}
                  className="w-12 h-10 text-xs font-medium text-slate-600 text-center"
                >
                  {label}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {table.map((row, rowIdx) => (
              <tr key={rowIdx}>
                {rowLabels && (
                  <td className="w-10 h-10 text-xs font-medium text-slate-600 text-center">
                    {rowLabels[rowIdx]}
                  </td>
                )}
                {row.map((cell, colIdx) => (
                  <td
                    key={colIdx}
                    className={`w-12 h-12 border text-center text-sm font-mono transition-colors duration-300 ${getCellColor(
                      rowIdx,
                      colIdx
                    )}`}
                    role="gridcell"
                  >
                    {cell}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Legend */}
      <div className="flex items-center justify-center gap-4 text-xs text-slate-500">
        <div className="flex items-center gap-1.5">
          <div className="w-3 h-3 rounded border border-green-300 bg-green-100" />
          <span>Computed</span>
        </div>
        <div className="flex items-center gap-1.5">
          <div className="w-3 h-3 rounded border border-indigo-400 bg-indigo-200" />
          <span>Current</span>
        </div>
      </div>
    </div>
  );
}
