import "@testing-library/jest-dom/vitest";
import React from "react";

// Mock matchMedia
Object.defineProperty(window, "matchMedia", {
  writable: true,
  value: vi.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(),
    removeListener: vi.fn(),
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

// Mock IntersectionObserver
class MockIntersectionObserver {
  readonly root: Element | null = null;
  readonly rootMargin: string = "";
  readonly thresholds: ReadonlyArray<number> = [];
  observe() {}
  unobserve() {}
  disconnect() {}
  takeRecords(): IntersectionObserverEntry[] {
    return [];
  }
}
Object.defineProperty(window, "IntersectionObserver", {
  value: MockIntersectionObserver,
});

// Mock ResizeObserver
class MockResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
}
Object.defineProperty(window, "ResizeObserver", {
  value: MockResizeObserver,
});

// Mock clipboard API
Object.defineProperty(navigator, "clipboard", {
  writable: true,
  value: {
    writeText: vi.fn().mockResolvedValue(undefined),
    readText: vi.fn().mockResolvedValue(""),
  },
});

// Mock Monaco Editor
vi.mock("@monaco-editor/react", () => {
  function MockEditor(props: Record<string, unknown>) {
    const { loading, value, language, theme, height } = props;
    if (loading && !value) {
      return loading as React.ReactNode;
    }
    return React.createElement("div", {
      "data-testid": "monaco-editor",
      "data-language": language,
      "data-theme": theme,
      "data-value": value,
      style: { height: height || "100%" },
    });
  }
  MockEditor.displayName = "MockEditor";

  function MockDiffEditor(props: Record<string, unknown>) {
    const { original, modified, language, theme } = props;
    return React.createElement("div", {
      "data-testid": "monaco-diff-editor",
      "data-language": language,
      "data-theme": theme,
      "data-original": original,
      "data-modified": modified,
    });
  }
  MockDiffEditor.displayName = "MockDiffEditor";

  return {
    __esModule: true,
    default: MockEditor,
    DiffEditor: MockDiffEditor,
  };
});
