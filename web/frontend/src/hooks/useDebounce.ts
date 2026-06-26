import { useState, useEffect, useRef, useCallback } from "react";

/**
 * Debounce a value by a given delay
 */
export function useDebounce<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState<T>(value);

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);

    return () => {
      clearTimeout(handler);
    };
  }, [value, delay]);

  return debouncedValue;
}

/**
 * Debounce a callback function
 */
export function useDebouncedCallback<T extends (...args: unknown[]) => unknown>(
  callback: T,
  delay: number
): (...args: Parameters<T>) => void {
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const debouncedCallback = useCallback(
    (...args: Parameters<T>) => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
      timeoutRef.current = setTimeout(() => {
        callback(...args);
      }, delay);
    },
    [callback, delay]
  );

  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, []);

  return debouncedCallback;
}

/**
 * Simple event emitter for batching messages
 */
export function createBatcher<T>(
  flush: (items: T[]) => void,
  interval: number
) {
  let batch: T[] = [];
  let timer: ReturnType<typeof setInterval> | null = null;

  const flushBatch = () => {
    if (batch.length > 0) {
      flush([...batch]);
      batch = [];
    }
  };

  return {
    add(item: T) {
      batch.push(item);
      if (!timer) {
        timer = setInterval(() => {
          flushBatch();
          if (batch.length === 0 && timer) {
            clearInterval(timer);
            timer = null;
          }
        }, interval);
      }
    },
    flush: flushBatch,
    destroy() {
      if (timer) {
        clearInterval(timer);
        timer = null;
      }
      flushBatch();
    },
  };
}
