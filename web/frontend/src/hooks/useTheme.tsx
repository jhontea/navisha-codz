import { useState, useEffect, useCallback, createContext, useContext, type ReactNode } from "react";

type Theme = "light" | "dark";

const STORAGE_KEY = "coding-challenge-theme";
const TRANSITION_DURATION = 300;

interface ThemeContextValue {
  theme: Theme;
  isDark: boolean;
  isLight: boolean;
  toggle: () => void;
  setTheme: (t: Theme) => void;
}

const ThemeContext = createContext<ThemeContextValue | null>(null);

function getSystemPreference(): Theme {
  if (typeof window === "undefined") return "dark";
  return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
}

function getStoredTheme(): Theme | null {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored === "light" || stored === "dark") return stored;
  } catch {
    // localStorage unavailable
  }
  return null;
}

function resolveTheme(): Theme {
  return getStoredTheme() ?? getSystemPreference();
}

function applyThemeToDOM(theme: Theme, animate: boolean) {
  const root = document.documentElement;
  if (animate) {
    root.classList.add("theme-transitioning");
    document.body.style.transition = `background-color ${TRANSITION_DURATION}ms ease, color ${TRANSITION_DURATION}ms ease`;
    setTimeout(() => {
      root.classList.remove("theme-transitioning");
      document.body.style.transition = "";
    }, TRANSITION_DURATION);
  }
  root.classList.toggle("dark", theme === "dark");
  root.classList.toggle("light", theme === "light");
  root.style.colorScheme = theme;
  // Update meta theme-color
  const meta = document.querySelector('meta[name="theme-color"]');
  if (meta) {
    meta.setAttribute("content", theme === "dark" ? "#0a0a0a" : "#ffffff");
  }
}

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [theme, setThemeState] = useState<Theme>(resolveTheme);

  // Initialize on mount
  useEffect(() => {
    applyThemeToDOM(theme, false);
  }, []);

  // Listen for system preference changes
  useEffect(() => {
    const mq = window.matchMedia("(prefers-color-scheme: dark)");
    const handler = (e: MediaQueryListEvent) => {
      if (!getStoredTheme()) {
        const newTheme: Theme = e.matches ? "dark" : "light";
        setThemeState(newTheme);
        applyThemeToDOM(newTheme, true);
      }
    };
    mq.addEventListener("change", handler);
    return () => mq.removeEventListener("change", handler);
  }, []);

  const setTheme = useCallback((t: Theme) => {
    setThemeState(t);
    try {
      localStorage.setItem(STORAGE_KEY, t);
    } catch {
      // ignore
    }
    applyThemeToDOM(t, true);
  }, []);

  const toggle = useCallback(() => {
    setThemeState((prev) => {
      const next: Theme = prev === "dark" ? "light" : "dark";
      try {
        localStorage.setItem(STORAGE_KEY, next);
      } catch {
        // ignore
      }
      applyThemeToDOM(next, true);
      return next;
    });
  }, []);

  const value: ThemeContextValue = {
    theme,
    isDark: theme === "dark",
    isLight: theme === "light",
    toggle,
    setTheme,
  };

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>;
}

export function useTheme(): ThemeContextValue {
  const ctx = useContext(ThemeContext);
  if (!ctx) {
    throw new Error("useTheme must be used within a <ThemeProvider>");
  }
  return ctx;
}

export default useTheme;
