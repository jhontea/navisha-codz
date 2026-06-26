import React, { useState, useCallback, useRef, useEffect } from "react";
import type { DPTableResult, DPStep, DPCell, MemoFrame } from "../types";
import { Play, Pause, SkipBack, SkipForward, ChevronDown, ChevronUp, Download } from "lucide-react";

interface DPVisualizerEnhancedProps {
  data: DPTableResult;
  onStepChange?: (step: number) => void;
}

const defaultStateColors: Record<string, string> = {
  empty: "#f0f0f0",
  base_case: "#dbeafe",
  filling: "#fef3c7",
  filled: "#d1fae5",
  highlight: "#fce7f3",
  result: "#ede9fe",
  backtrack: "#f97316",
  optimal_base: "#06b6d4",
  comparing: "#f59e0b",
  dependency: "#a7f3d0",
  cached: "#93c5fd",
};

function getColorForState(state: string, customColors?: Record<string, string>): string {
  const colors = customColors ?? defaultStateColors;
  return colors[state] ?? defaultStateColors[state] ?? "#f0f0f0";
}

export function DPVisualizerEnhanced({ data, onStepChange }: DPVisualizerEnhancedProps) {
  const [currentStep, setCurrentStep] = useState(0);
  const [isPlaying, setIsPlaying] = useState(false);
  const [speed, setSpeed] = useState(1000);
  const [showMemoStack, setShowMemoStack] = useState(false);
  const [showDualView, setShowDualView] = useState(false);
  const [hoveredCell, setHoveredCell] = useState<DPCell | null>(null);
  const [viewMode, setViewMode] = useState<"tabulation" | "memoization">("tabulation");

  const containerRef = useRef<HTMLDivElement>(null);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const totalSteps = data.steps.length;

  const currentStepData = data.steps[Math.min(currentStep, totalSteps - 1)];

  // Update the table to show current step's state
  const getDisplayTable = useCallback((): DPCell[][] => {
    if (totalSteps === 0) return data.table;

    const display: DPCell[][] = data.table.map((row) =>
      row.map((cell) => ({ ...cell }))
    );

    for (let s = 0; s <= currentStep && s < totalSteps; s++) {
      const step = data.steps[s];
      for (const modified of step.cells_modified) {
        if (display[modified.row] && display[modified.row][modified.col]) {
          display[modified.row][modified.col] = {
            ...display[modified.row][modified.col],
            ...modified,
            is_base_case: modified.is_base_case ?? display[modified.row][modified.col].is_base_case,
            is_result: modified.is_result ?? display[modified.row][modified.col].is_result,
          };
        }
      }
    }

    return display;
  }, [data.table, data.steps, currentStep, totalSteps]);

  const displayTable = getDisplayTable();

  // Play/Pause logic
  useEffect(() => {
    if (isPlaying) {
      intervalRef.current = setInterval(() => {
        setCurrentStep((prev) => {
          if (prev >= totalSteps - 1) {
            setIsPlaying(false);
            return prev;
          }
          return prev + 1;
        });
      }, speed);
    } else {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    }
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [isPlaying, speed, totalSteps]);

  useEffect(() => {
    onStepChange?.(currentStep);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentStep]);

  const handlePlayPause = () => {
    if (currentStep >= totalSteps - 1) {
      setCurrentStep(0);
    }
    setIsPlaying(!isPlaying);
  };

  const handleReset = () => {
    setIsPlaying(false);
    setCurrentStep(0);
  };

  const handleStepForward = () => {
    if (currentStep < totalSteps - 1) {
      setCurrentStep(currentStep + 1);
    }
  };

  const handleStepBackward = () => {
    if (currentStep > 0) {
      setCurrentStep(currentStep - 1);
    }
  };

  const handleExportImage = async () => {
    if (!containerRef.current) return;
    try {
      const { default: html2canvas } = await import("html2canvas");
      const canvas = await html2canvas(containerRef.current, {
        backgroundColor: "#ffffff",
        scale: 2,
      });
      const link = document.createElement("a");
      link.download = `dp-visualization-${data.problem_id}-step-${currentStep}.png`;
      link.href = canvas.toDataURL("image/png");
      link.click();
    } catch {
      // html2canvas not available, try fallback
      const svgContent = createSVGSnapshot(displayTable, data.dimensions, data.title);
      const blob = new Blob([svgContent], { type: "image/svg+xml" });
      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.download = `dp-visualization-${data.problem_id}-step-${currentStep}.svg`;
      link.href = url;
      link.click();
      URL.revokeObjectURL(url);
    }
  };

  // Check if a cell should be highlighted
  const isCellActive = (row: number, col: number): boolean => {
    if (!currentStepData) return false;
    return currentStepData.active_cells.some((c) => c.row === row && c.col === col);
  };

  const isCellHighlighted = (row: number, col: number): boolean => {
    if (!currentStepData) return false;
    return currentStepData.highlight_cells.some((c) => c.row === row && c.col === col);
  };

  const isOneDimensional = data.dimensions.rows <= 1;

  return (
    <div className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-700 overflow-hidden">
      {/* Header */}
      <div className="px-4 py-3 border-b border-slate-200 dark:border-slate-700">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-lg font-bold text-slate-900 dark:text-white">{data.title}</h3>
            <p className="text-sm text-slate-500 dark:text-slate-400 mt-1">{data.description}</p>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={handleExportImage}
              className="p-2 rounded-lg text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
              title="Export as image"
              aria-label="Export visualization"
            >
              <Download className="w-4 h-4" />
            </button>
          </div>
        </div>

        {/* Approach tabs */}
        <div className="flex gap-1 mt-3 bg-slate-100 dark:bg-slate-800 rounded-lg p-1">
          <button
            onClick={() => setViewMode("tabulation")}
            className={`flex-1 px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
              viewMode === "tabulation"
                ? "bg-white dark:bg-slate-700 text-slate-900 dark:text-white shadow-sm"
                : "text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white"
            }`}
          >
            Tabulation
          </button>
          <button
            onClick={() => setViewMode("memoization")}
            className={`flex-1 px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
              viewMode === "memoization"
                ? "bg-white dark:bg-slate-700 text-slate-900 dark:text-white shadow-sm"
                : "text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white"
            }`}
          >
            Memoization
          </button>
        </div>
      </div>

      <div ref={containerRef} className="p-4 space-y-4">
        {/* Controls */}
        <div className="flex items-center justify-between gap-4">
          <div className="flex items-center gap-1">
            <button
              onClick={handleReset}
              className="p-2 rounded-lg text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors disabled:opacity-50"
              disabled={currentStep === 0}
              aria-label="Reset to beginning"
            >
              <SkipBack className="w-4 h-4" />
            </button>
            <button
              onClick={handleStepBackward}
              className="p-2 rounded-lg text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors disabled:opacity-50"
              disabled={currentStep === 0}
              aria-label="Previous step"
            >
              <ChevronDown className="w-4 h-4 rotate-90" />
            </button>
            <button
              onClick={handlePlayPause}
              className="p-2 rounded-lg bg-indigo-500 hover:bg-indigo-600 text-white transition-colors"
              aria-label={isPlaying ? "Pause" : "Play"}
            >
              {isPlaying ? <Pause className="w-4 h-4" /> : <Play className="w-4 h-4" />}
            </button>
            <button
              onClick={handleStepForward}
              className="p-2 rounded-lg text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors disabled:opacity-50"
              disabled={currentStep >= totalSteps - 1}
              aria-label="Next step"
            >
              <ChevronUp className="w-4 h-4 rotate-90" />
            </button>
            <button
              onClick={() => setCurrentStep(totalSteps - 1)}
              className="p-2 rounded-lg text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors disabled:opacity-50"
              disabled={currentStep >= totalSteps - 1}
              aria-label="Skip to end"
            >
              <SkipForward className="w-4 h-4" />
            </button>
          </div>

          <div className="flex items-center gap-3">
            <span className="text-xs text-slate-500 dark:text-slate-400">
              Step {currentStep + 1}/{totalSteps}
            </span>
            <select
              value={speed}
              onChange={(e) => setSpeed(Number(e.target.value))}
              className="text-xs border border-slate-200 dark:border-slate-600 rounded-md px-2 py-1 bg-transparent text-slate-700 dark:text-slate-300"
              aria-label="Animation speed"
            >
              <option value={2000}>0.5x</option>
              <option value={1000}>1x</option>
              <option value={500}>2x</option>
              <option value={200}>5x</option>
              <option value={100}>10x</option>
            </select>
          </div>
        </div>

        {/* Step slider */}
        <input
          type="range"
          min={0}
          max={totalSteps - 1}
          value={currentStep}
          onChange={(e) => {
            setIsPlaying(false);
            setCurrentStep(Number(e.target.value));
          }}
          className="w-full h-2 bg-slate-200 dark:bg-slate-700 rounded-lg appearance-none cursor-pointer accent-indigo-500"
          aria-label="Step slider"
        />

        {/* Step description */}
        {currentStepData && (
          <div className="bg-slate-50 dark:bg-slate-800 rounded-lg p-3">
            <div className="flex items-start gap-2">
              <span className="text-xs font-bold text-indigo-500 dark:text-indigo-400 shrink-0 mt-0.5">
                #{currentStepData.step_number}
              </span>
              <div>
                <p className="text-sm text-slate-700 dark:text-slate-300">{currentStepData.description}</p>
                {currentStepData.formula && (
                  <p className="text-xs font-mono text-slate-500 dark:text-slate-400 mt-1">
                    {currentStepData.formula}
                  </p>
                )}
                {currentStepData.complexity_now && (
                  <p className="text-xs text-amber-600 dark:text-amber-400 mt-1">
                    ⏱ {currentStepData.complexity_now}
                  </p>
                )}
              </div>
            </div>
          </div>
        )}

        {/* DP Table */}
        <div className="overflow-x-auto">
          <table className="w-full border-collapse text-xs" role="table">
            <thead>
              <tr>
                <th className="px-2 py-1 border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-slate-500 dark:text-slate-400 font-medium">
                  {isOneDimensional ? "Index" : ""}
                </th>
                {data.dimensions.col_labels.map((label, col) => (
                  <th
                    key={col}
                    className="px-2 py-1 border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-slate-500 dark:text-slate-400 font-medium text-center"
                  >
                    {label}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {displayTable.map((row, rowIdx) => (
                <tr key={rowIdx}>
                  <td className="px-2 py-1 border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-slate-500 dark:text-slate-400 font-medium text-right">
                    {data.dimensions.row_labels[rowIdx] ?? rowIdx}
                  </td>
                  {row.map((cell, colIdx) => {
                    const active = isCellActive(rowIdx, colIdx);
                    const highlighted = isCellHighlighted(rowIdx, colIdx);
                    const bgColor = getColorForState(cell.color, data.state_colors);

                    return (
                      <td
                        key={colIdx}
                        className={`px-2 py-1 border text-center font-mono text-sm transition-all duration-300 cursor-pointer relative ${
                          active ? "ring-2 ring-indigo-400 z-10" : ""
                        } ${highlighted ? "ring-2 ring-amber-400 z-10" : ""}`}
                        style={{ backgroundColor: bgColor }}
                        onMouseEnter={() => setHoveredCell(cell)}
                        onMouseLeave={() => setHoveredCell(null)}
                        onFocus={() => setHoveredCell(cell)}
                        onBlur={() => setHoveredCell(null)}
                        tabIndex={0}
                        role="cell"
                        aria-label={`Row ${rowIdx}, Col ${colIdx}, value ${String(cell.value)}`}
                      >
                        <span
                          className={`font-semibold ${
                            cell.is_result
                              ? "text-purple-700 dark:text-purple-300"
                              : cell.is_base_case
                              ? "text-blue-700 dark:text-blue-300"
                              : cell.is_backtrack
                              ? "text-orange-700 dark:text-orange-300"
                              : "text-slate-700 dark:text-slate-300"
                          }`}
                        >
                          {String(cell.value)}
                        </span>
                      </td>
                    );
                  })}
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {/* Hovered cell details */}
        {hoveredCell && (
          <div className="bg-indigo-50 dark:bg-indigo-900/20 rounded-lg p-3 border border-indigo-100 dark:border-indigo-800">
            <div className="grid grid-cols-2 gap-2 text-xs">
              <div>
                <span className="text-slate-500 dark:text-slate-400">Position:</span>
                <span className="ml-1 font-medium text-slate-700 dark:text-slate-300">
                  ({hoveredCell.row}, {hoveredCell.col})
                </span>
              </div>
              <div>
                <span className="text-slate-500 dark:text-slate-400">Value:</span>
                <span className="ml-1 font-medium text-slate-700 dark:text-slate-300">
                  {String(hoveredCell.value)}
                </span>
              </div>
              {hoveredCell.formula && (
                <div className="col-span-2">
                  <span className="text-slate-500 dark:text-slate-400">Formula:</span>
                  <span className="ml-1 font-mono text-slate-700 dark:text-slate-300">{hoveredCell.formula}</span>
                </div>
              )}
              {hoveredCell.explanation && (
                <div className="col-span-2">
                  <span className="text-slate-500 dark:text-slate-400">Info:</span>
                  <span className="ml-1 text-slate-700 dark:text-slate-300">{hoveredCell.explanation}</span>
                </div>
              )}
              {hoveredCell.dependencies && hoveredCell.dependencies.length > 0 && (
                <div className="col-span-2">
                  <span className="text-slate-500 dark:text-slate-400">Dependencies:</span>
                  <span className="ml-1 text-slate-700 dark:text-slate-300">
                    {hoveredCell.dependencies.map((d) => `(${d.row}, ${d.col})`).join(", ")}
                  </span>
                </div>
              )}
            </div>
          </div>
        )}

        {/* Complexity Analysis */}
        <div className="grid grid-cols-2 gap-3">
          <div className="bg-slate-50 dark:bg-slate-800 rounded-lg p-3">
            <span className="text-xs text-slate-500 dark:text-slate-400 block">Time Complexity</span>
            <span className="text-sm font-mono font-medium text-slate-700 dark:text-slate-300">
              {data.time_complexity}
            </span>
          </div>
          <div className="bg-slate-50 dark:bg-slate-800 rounded-lg p-3">
            <span className="text-xs text-slate-500 dark:text-slate-400 block">Space Complexity</span>
            <span className="text-sm font-mono font-medium text-slate-700 dark:text-slate-300">
              {data.space_complexity}
            </span>
          </div>
        </div>

        {/* Memoization Stack Trace */}
        {data.memo_stack && data.memo_stack.length > 0 && (
          <div>
            <button
              onClick={() => setShowMemoStack(!showMemoStack)}
              className="flex items-center gap-1 text-sm font-medium text-slate-700 dark:text-slate-300 hover:text-slate-900 dark:hover:text-white"
            >
              {showMemoStack ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
              Memoization Stack Trace ({data.memo_stack.length} frames)
            </button>
            {showMemoStack && (
              <div className="mt-2 bg-slate-50 dark:bg-slate-800 rounded-lg p-3 max-h-48 overflow-y-auto">
                {data.memo_stack.map((frame, idx) => (
                  <div
                    key={idx}
                    className={`flex items-center gap-2 text-xs py-1 ${
                      frame.cached ? "text-blue-500" : "text-slate-600 dark:text-slate-400"
                    }`}
                    style={{ paddingLeft: `${frame.depth * 12}px` }}
                  >
                    <span className="font-mono">
                      {frame.function}({String(frame.args)})
                    </span>
                    <span className="text-slate-400">→</span>
                    <span className="font-semibold">{String(frame.result)}</span>
                    {frame.cached && (
                      <span className="px-1.5 py-0.5 rounded text-xs bg-blue-100 dark:bg-blue-900 text-blue-600 dark:text-blue-400">
                        cached
                      </span>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {/* Backtrack Path */}
        {data.backtrack_path && data.backtrack_path.length > 0 && (
          <div>
            <button
              onClick={() => setShowDualView(!showDualView)}
              className="flex items-center gap-1 text-sm font-medium text-slate-700 dark:text-slate-300 hover:text-slate-900 dark:hover:text-white"
            >
              {showDualView ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
              Backtrack Path ({data.backtrack_path.length} steps)
            </button>
            {showDualView && (
              <div className="mt-2 bg-slate-50 dark:bg-slate-800 rounded-lg p-3">
                <div className="flex flex-wrap gap-1">
                  {data.backtrack_path.map((coord, idx) => (
                    <span
                      key={idx}
                      className="px-2 py-1 rounded text-xs font-mono bg-orange-100 dark:bg-orange-900/30 text-orange-700 dark:text-orange-300"
                    >
                      ({coord.row}, {coord.col})
                    </span>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

function createSVGSnapshot(
  table: DPCell[][],
  dimensions: { rows: number; cols: number; row_labels: string[]; col_labels: string[] },
  title: string
): string {
  const cellW = 60;
  const cellH = 30;
  const headerH = 30;
  const labelW = 60;
  const totalW = labelW + dimensions.cols * cellW + 20;
  const totalH = headerH + dimensions.rows * cellH + 40;

  let svg = `<svg xmlns="http://www.w3.org/2000/svg" width="${totalW}" height="${totalH}" viewBox="0 0 ${totalW} ${totalH}">`;
  svg += `<rect width="100%" height="100%" fill="white"/>`;
  svg += `<text x="10" y="15" font-family="monospace" font-size="12" font-weight="bold" fill="#334155">${title}</text>`;

  // Column headers
  for (let c = 0; c < dimensions.col_labels.length; c++) {
    const x = labelW + c * cellW;
    svg += `<rect x="${x}" y="${headerH}" width="${cellW}" height="${cellH}" fill="#f8fafc" stroke="#e2e8f0"/>`;
    svg += `<text x="${x + cellW / 2}" y="${headerH + cellH / 2 + 4}" text-anchor="middle" font-family="monospace" font-size="11" fill="#64748b">${dimensions.col_labels[c]}</text>`;
  }

  // Rows
  for (let r = 0; r < table.length; r++) {
    const y = headerH + (r + 1) * cellH;
    // Row label
    svg += `<rect x="0" y="${y}" width="${labelW}" height="${cellH}" fill="#f8fafc" stroke="#e2e8f0"/>`;
    svg += `<text x="${labelW - 5}" y="${y + cellH / 2 + 4}" text-anchor="end" font-family="monospace" font-size="11" fill="#64748b">${dimensions.row_labels[r] ?? r}</text>`;
    // Cells
    for (let c = 0; c < table[r].length; c++) {
      const cell = table[r][c];
      const bgColor = getColorForState(cell.color);
      const x = labelW + c * cellW;
      svg += `<rect x="${x}" y="${y}" width="${cellW}" height="${cellH}" fill="${bgColor}" stroke="#e2e8f0"/>`;
      const textColor = cell.is_result ? "#6d28d9" : cell.is_base_case ? "#1d4ed8" : "#334155";
      svg += `<text x="${x + cellW / 2}" y="${y + cellH / 2 + 4}" text-anchor="middle" font-family="monospace" font-size="12" font-weight="bold" fill="${textColor}">${String(cell.value)}</text>`;
    }
  }

  svg += `</svg>`;
  return svg;
}
