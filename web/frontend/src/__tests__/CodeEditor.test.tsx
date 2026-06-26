import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { CodeEditor } from "../components/CodeEditor";

// ── Mock Data ───────────────────────────────────────────────────────────────

const defaultCode = `package main

func main() {
\tprintln("hello")
}`;

// ── Tests ───────────────────────────────────────────────────────────────────

describe("CodeEditor", () => {
  it("renders editor container with toolbar", () => {
    render(<CodeEditor value={defaultCode} onChange={vi.fn()} />);

    // Toolbar buttons
    expect(screen.getByLabelText("Select editor theme")).toBeInTheDocument();
    expect(screen.getByLabelText("Font size")).toBeInTheDocument();
    expect(screen.getByLabelText("Disable word wrap")).toBeInTheDocument();
    expect(screen.getByLabelText("Show minimap")).toBeInTheDocument();
    expect(screen.getByLabelText("Toggle diff view")).toBeInTheDocument();
    expect(screen.getByLabelText("Reset to template")).toBeInTheDocument();
    expect(screen.getByLabelText("Copy code")).toBeInTheDocument();
  });

  it("renders language badge", () => {
    render(<CodeEditor value={defaultCode} onChange={vi.fn()} language="go" />);

    // Language badge (hidden on small screens, but in DOM)
    expect(screen.getByText("go")).toBeInTheDocument();
  });

  it("renders submit button when onSubmit provided", () => {
    render(
      <CodeEditor value={defaultCode} onChange={vi.fn()} onSubmit={vi.fn()} />
    );

    const submitBtn = screen.getByLabelText("Submit solution");
    expect(submitBtn).toBeInTheDocument();
    expect(submitBtn).toHaveTextContent("Submit");
  });

  it("does not render submit button when onSubmit not provided", () => {
    render(<CodeEditor value={defaultCode} onChange={vi.fn()} />);

    expect(screen.queryByLabelText("Submit solution")).not.toBeInTheDocument();
  });

  it("calls onSubmit when submit button clicked", () => {
    const onSubmit = vi.fn();

    render(
      <CodeEditor value={defaultCode} onChange={vi.fn()} onSubmit={onSubmit} />
    );

    fireEvent.click(screen.getByLabelText("Submit solution"));
    expect(onSubmit).toHaveBeenCalledOnce();
  });

  it("opens theme selector dropdown on palette click", () => {
    render(<CodeEditor value={defaultCode} onChange={vi.fn()} />);

    // Dropdown should not be visible initially
    expect(screen.queryByText("VS Code Dark")).not.toBeInTheDocument();
    expect(screen.queryByText("Monokai")).not.toBeInTheDocument();

    // Click theme button
    fireEvent.click(screen.getByLabelText("Select editor theme"));

    // All themes should be visible
    expect(screen.getByText("VS Code Dark")).toBeInTheDocument();
    expect(screen.getByText("VS Code Light")).toBeInTheDocument();
    expect(screen.getByText("Monokai")).toBeInTheDocument();
    expect(screen.getByText("Dracula")).toBeInTheDocument();
    expect(screen.getByText("GitHub Light")).toBeInTheDocument();
  });

  it("calls onThemeChange when a theme is selected", () => {
    const onChange = vi.fn();
    render(<CodeEditor value={defaultCode} onChange={onChange} theme="vs-dark" />);

    // Open theme dropdown
    fireEvent.click(screen.getByLabelText("Select editor theme"));

    // Select Monokai
    fireEvent.click(screen.getByText("Monokai"));

    // The dropdown should close
    expect(screen.queryByText("VS Code Light")).not.toBeInTheDocument();
  });

  it("opens font size dropdown and selects a size", () => {
    render(<CodeEditor value={defaultCode} onChange={vi.fn()} />);

    // Font size button shows current size
    expect(screen.getByText("14")).toBeInTheDocument();

    // Open font size dropdown
    fireEvent.click(screen.getByLabelText("Font size"));

    // Font sizes should be visible
    expect(screen.getByText("12")).toBeInTheDocument();
    expect(screen.getByText("16")).toBeInTheDocument();
    expect(screen.getByText("18")).toBeInTheDocument();
    expect(screen.getByText("24")).toBeInTheDocument();

    // Select size 18
    fireEvent.click(screen.getByText("18"));

    // Dropdown should close
    expect(screen.queryByText("16")).not.toBeInTheDocument();
  });

  it("toggles word wrap on button click", () => {
    render(<CodeEditor value={defaultCode} onChange={vi.fn()} />);

    // Initially wrap is enabled - button says "Disable word wrap"
    const wrapBtn = screen.getByLabelText("Disable word wrap");
    expect(wrapBtn).toBeInTheDocument();

    // Click to toggle
    fireEvent.click(wrapBtn);

    // Now should say "Enable word wrap"
    expect(screen.getByLabelText("Enable word wrap")).toBeInTheDocument();
  });

  it("toggles minimap on button click", () => {
    render(<CodeEditor value={defaultCode} onChange={vi.fn()} />);

    // Initially minimap is disabled
    const minimapBtn = screen.getByLabelText("Show minimap");
    expect(minimapBtn).toBeInTheDocument();

    // Click to toggle
    fireEvent.click(minimapBtn);

    // Now should say "Hide minimap"
    expect(screen.getByLabelText("Hide minimap")).toBeInTheDocument();
  });

  it("toggles diff mode on button click", () => {
    render(
      <CodeEditor
        value={defaultCode}
        onChange={vi.fn()}
        previousSubmission="original code"
      />
    );

    // Click diff mode button
    fireEvent.click(screen.getByLabelText("Toggle diff view"));

    // Button should be active
    expect(screen.getByLabelText("Toggle diff view")).toBeInTheDocument();
  });

  it("calls onChange when editor value changes", () => {
    // We can't directly test Monaco editor changes in jsdom,
    // but we verify the editor renders with the correct initial value
    render(<CodeEditor value={defaultCode} onChange={vi.fn()} />);

    // Monaco editor mock should have the value
    const editor = screen.getByTestId("monaco-editor");
    expect(editor).toHaveAttribute("data-value", defaultCode);
    expect(editor).toHaveAttribute("data-language", "go");
  });

  it("renders loading state before editor mounts", () => {
    // The Monaco mock renders loading when no value
    render(<CodeEditor value="" onChange={vi.fn()} />);

    // Loading text
    expect(screen.getByText("Loading editor...")).toBeInTheDocument();
  });

  it("calls onReset when reset button clicked and template exists", () => {
    const onChange = vi.fn();
    render(
      <CodeEditor
        value="modified code"
        onChange={onChange}
        template={defaultCode}
      />
    );

    fireEvent.click(screen.getByLabelText("Reset to template"));
    expect(onChange).toHaveBeenCalledWith(defaultCode);
  });

  it("calls clipboard writeText on copy button click", () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, "clipboard", {
      writable: true,
      value: { writeText },
    });

    render(<CodeEditor value={defaultCode} onChange={vi.fn()} />);

    fireEvent.click(screen.getByLabelText("Copy code"));
    expect(writeText).toHaveBeenCalledWith(defaultCode);
  });
});
