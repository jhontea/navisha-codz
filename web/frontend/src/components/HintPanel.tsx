import { useState } from "react";
import { Lightbulb, AlertTriangle, Eye, EyeOff } from "lucide-react";
import type { Hint } from "../types";

interface HintPanelProps {
  problemId: string;
  hints: Hint[];
}

export function HintPanel({ problemId, hints }: HintPanelProps) {
  const [revealedHints, setRevealedHints] = useState<Set<string>>(
    new Set(hints.filter((h) => h.is_revealed).map((h) => h.id))
  );
  const [confirmHint, setConfirmHint] = useState<Hint | null>(null);

  const handleReveal = (hint: Hint) => {
    setRevealedHints((prev) => new Set(prev).add(hint.id));
    setConfirmHint(null);
  };

  const sortedHints = [...hints].sort((a, b) => a.level - b.level);

  return (
    <div className="border-t border-slate-200 dark:border-slate-700 pt-4">
      <div className="flex items-center gap-2 mb-3">
        <Lightbulb className="w-4 h-4 text-amber-500" />
        <h3 className="text-sm font-semibold text-slate-900 dark:text-white">Hints</h3>
        <span className="text-xs text-slate-500 dark:text-slate-400">
          ({revealedHints.size}/{hints.length} revealed)
        </span>
      </div>

      <div className="space-y-2">
        {sortedHints.map((hint) => {
          const isRevealed = revealedHints.has(hint.id);
          return (
            <div
              key={hint.id}
              className={`rounded-lg border p-3 transition-colors ${
                isRevealed
                  ? "bg-amber-50 dark:bg-amber-950/30 border-amber-200 dark:border-amber-800"
                  : "bg-slate-50 dark:bg-slate-800/50 border-slate-200 dark:border-slate-700"
              }`}
            >
              <div className="flex items-start justify-between gap-3">
                <div className="flex items-start gap-2">
                  <span className="text-xs font-medium text-slate-500 dark:text-slate-400 mt-0.5">
                    #{hint.level}
                  </span>
                  {isRevealed ? (
                    <p className="text-sm text-slate-700 dark:text-slate-300">{hint.content}</p>
                  ) : (
                    <div className="flex items-center gap-2">
                      <EyeOff className="w-3.5 h-3.5 text-slate-400" />
                      <span className="text-sm text-slate-500 dark:text-slate-400 italic">
                        Hint hidden
                      </span>
                    </div>
                  )}
                </div>
                {!isRevealed && (
                  <button
                    onClick={() => setConfirmHint(hint)}
                    className="shrink-0 flex items-center gap-1 px-2.5 py-1 text-xs font-medium text-amber-700 dark:text-amber-400 bg-amber-100 dark:bg-amber-950/50 rounded-md hover:bg-amber-200 dark:hover:bg-amber-900/50 transition-colors"
                    aria-label={`Reveal hint ${hint.level}`}
                  >
                    <Eye className="w-3 h-3" />
                    Reveal
                  </button>
                )}
              </div>
              {isRevealed && hint.penalty_points > 0 && (
                <p className="text-xs text-amber-600 dark:text-amber-400 mt-1.5">
                  -{hint.penalty_points} points penalty
                </p>
              )}
            </div>
          );
        })}
      </div>

      {/* Confirmation Dialog */}
      {confirmHint && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
          role="dialog"
          aria-modal="true"
          aria-labelledby="hint-confirm-title"
        >
          <div className="bg-white dark:bg-slate-900 rounded-xl shadow-xl max-w-sm w-full p-6">
            <div className="flex items-center gap-3 mb-4">
              <div className="p-2 bg-amber-100 dark:bg-amber-950/50 rounded-lg">
                <AlertTriangle className="w-5 h-5 text-amber-600 dark:text-amber-400" />
              </div>
              <h2 id="hint-confirm-title" className="text-lg font-semibold text-slate-900 dark:text-white">
                Reveal Hint?
              </h2>
            </div>
            <p className="text-sm text-slate-600 dark:text-slate-400 mb-2">
              Revealing this hint will deduct{" "}
              <strong className="text-amber-600 dark:text-amber-400">{confirmHint.penalty_points} points</strong>{" "}
              from your score for this problem.
            </p>
            <p className="text-xs text-slate-500 dark:text-slate-400 mb-6">
              This action cannot be undone.
            </p>
            <div className="flex items-center justify-end gap-3">
              <button
                onClick={() => setConfirmHint(null)}
                className="px-4 py-2 text-sm font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-lg transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={() => handleReveal(confirmHint)}
                className="px-4 py-2 text-sm font-medium text-white bg-amber-600 hover:bg-amber-700 rounded-lg transition-colors"
              >
                Reveal Hint
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
