import { useEffect, useCallback } from "react";

/**
 * Hook to listen for keyboard shortcuts.
 * Returns a function to check if a shortcut is active.
 */
export function useKeyboardShortcut(
  key: string,
  ctrlKey: boolean,
  shiftKey: boolean,
  callback: () => void,
  options?: { enabled?: boolean }
) {
  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if (options?.enabled === false) return;

      // Don't trigger when typing in input/textarea/select
      const target = e.target as HTMLElement;
      if (
        target.tagName === "INPUT" ||
        target.tagName === "TEXTAREA" ||
        target.tagName === "SELECT" ||
        target.isContentEditable
      ) {
        return;
      }

      if (
        e.key.toLowerCase() === key.toLowerCase() &&
        e.ctrlKey === ctrlKey &&
        e.shiftKey === shiftKey
      ) {
        e.preventDefault();
        callback();
      }
    },
    [key, ctrlKey, shiftKey, callback, options?.enabled]
  );

  useEffect(() => {
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [handleKeyDown]);
}
