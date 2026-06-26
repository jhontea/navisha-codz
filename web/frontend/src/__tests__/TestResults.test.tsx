import { render, screen } from "@testing-library/react";
import { describe, it, expect, beforeEach } from "vitest";
import { TestResults } from "../components/TestResults";
import { useSubmissionStore } from "../store/submissionStore";
import type { Submission, SubmissionStatus, TestResult } from "../types";

// ── Helper: set store state ─────────────────────────────────────────────────

function setStoreState(overrides: Record<string, any>) {
  const state = useSubmissionStore.getState();
  Object.assign(state, overrides);
  useSubmissionStore.setState({ ...state });
}

// ── Mock Data ───────────────────────────────────────────────────────────────

const mockTestResult = (
  status: SubmissionStatus,
  overrides: Partial<TestResult> = {}
): TestResult => ({
  test_case_id: `tc-${Math.random()}`,
  status,
  expected_output: "42",
  actual_output: status === "accepted" ? "42" : "99",
  execution_time_ms: 15,
  memory_used_kb: 2048,
  error_message: status === "runtime_error" ? "index out of range" : undefined,
  ...overrides,
});

const mockSubmission = (
  results: TestResult[],
  status: SubmissionStatus = "accepted"
): Submission => ({
  id: "sub-1",
  problem_id: "prob-1",
  user_id: "user-1",
  code: "package main",
  language: "go",
  status,
  score: 100,
  execution_time_ms: 50,
  memory_used_kb: 4096,
  test_results: results,
  created_at: "2024-01-01",
});

// ── Tests ───────────────────────────────────────────────────────────────────

describe("TestResults", () => {
  beforeEach(() => {
    // Reset store to initial state before each test
    useSubmissionStore.setState({
      currentSubmission: null,
      liveStatus: null,
      progress: 0,
      completedTests: 0,
      totalTests: 0,
    });
  });

  it("shows empty state when no submission and no live status", () => {
    render(<TestResults />);

    expect(
      screen.getByText("Submit your solution to see test results")
    ).toBeInTheDocument();
  });

  it("shows running state when liveStatus is running", () => {
    setStoreState({
      liveStatus: "running",
      progress: 30,
      completedTests: 3,
      totalTests: 10,
    });

    render(<TestResults />);

    expect(screen.getByText("Running")).toBeInTheDocument();
    expect(screen.getByText("3/10 tests")).toBeInTheDocument();
  });

  it("shows pending state when liveStatus is pending", () => {
    setStoreState({
      liveStatus: "pending",
      progress: 0,
      completedTests: 0,
      totalTests: 5,
    });

    render(<TestResults />);

    expect(screen.getByText("Pending")).toBeInTheDocument();
    expect(screen.getByText("0/5 tests")).toBeInTheDocument();
  });

  it("renders progress bar with correct width", () => {
    setStoreState({
      liveStatus: "running",
      progress: 60,
      completedTests: 6,
      totalTests: 10,
    });

    render(<TestResults />);

    const progressBar = screen.getByRole("progressbar");
    expect(progressBar).toHaveAttribute("aria-valuenow", "60");
    expect(progressBar).toHaveAttribute("aria-valuemin", "0");
    expect(progressBar).toHaveAttribute("aria-valuemax", "100");
  });

  it("shows accepted status when all tests pass", () => {
    const results = [
      mockTestResult("accepted"),
      mockTestResult("accepted"),
      mockTestResult("accepted"),
    ];

    setStoreState({
      currentSubmission: mockSubmission(results, "accepted"),
      liveStatus: "accepted",
      progress: 100,
      completedTests: 3,
      totalTests: 3,
    });

    render(<TestResults />);

    // Status appears in both header and each test result label
    const acceptedLabels = screen.getAllByText("Accepted");
    expect(acceptedLabels.length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText("3/3 tests")).toBeInTheDocument();

    // Each test case should show "Test Case 1", "Test Case 2", etc.
    expect(screen.getByText("Test Case 1")).toBeInTheDocument();
    expect(screen.getByText("Test Case 2")).toBeInTheDocument();
    expect(screen.getByText("Test Case 3")).toBeInTheDocument();
  });

  it("shows wrong answer with expected vs actual output", () => {
    const results = [
      mockTestResult("wrong_answer", {
        expected_output: "42",
        actual_output: "99",
      }),
    ];

    setStoreState({
      currentSubmission: mockSubmission(results, "wrong_answer"),
      liveStatus: "wrong_answer",
      progress: 100,
      completedTests: 1,
      totalTests: 1,
    });

    render(<TestResults />);

    const labels = screen.getAllByText("Wrong Answer");
    expect(labels.length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText("Test Case 1")).toBeInTheDocument();

    // Expected output
    expect(screen.getByText("Expected:")).toBeInTheDocument();
    expect(screen.getByText("42")).toBeInTheDocument();

    // Actual output (got)
    expect(screen.getByText("Got:")).toBeInTheDocument();
    expect(screen.getByText("99")).toBeInTheDocument();
  });

  it("shows compilation error", () => {
    const results = [
      mockTestResult("compilation_error", {
        error_message: "cannot use string as int",
      }),
    ];

    setStoreState({
      currentSubmission: mockSubmission(results, "compilation_error"),
      liveStatus: "compilation_error",
      progress: 100,
      completedTests: 1,
      totalTests: 1,
    });

    render(<TestResults />);

    const labels = screen.getAllByText("Compilation Error");
    expect(labels.length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText("cannot use string as int")).toBeInTheDocument();
  });

  it("shows runtime error with error message", () => {
    const results = [
      mockTestResult("runtime_error", {
        error_message: "index out of range",
      }),
    ];

    setStoreState({
      currentSubmission: mockSubmission(results, "runtime_error"),
      liveStatus: "runtime_error",
      progress: 100,
      completedTests: 1,
      totalTests: 1,
    });

    render(<TestResults />);

    const labels = screen.getAllByText("Runtime Error");
    expect(labels.length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText("index out of range")).toBeInTheDocument();
  });

  it("shows time limit exceeded", () => {
    const results = [mockTestResult("time_limit_exceeded")];

    setStoreState({
      currentSubmission: mockSubmission(results, "time_limit_exceeded"),
      liveStatus: "time_limit_exceeded",
      progress: 100,
      completedTests: 1,
      totalTests: 1,
    });

    render(<TestResults />);

    const labels = screen.getAllByText("Time Limit Exceeded");
    expect(labels.length).toBeGreaterThanOrEqual(1);
  });

  it("shows memory limit exceeded", () => {
    const results = [mockTestResult("memory_limit_exceeded")];

    setStoreState({
      currentSubmission: mockSubmission(results, "memory_limit_exceeded"),
      liveStatus: "memory_limit_exceeded",
      progress: 100,
      completedTests: 1,
      totalTests: 1,
    });

    render(<TestResults />);

    const labels = screen.getAllByText("Memory Limit Exceeded");
    expect(labels.length).toBeGreaterThanOrEqual(1);
  });

  it("shows execution time and memory for each test", () => {
    const results = [
      mockTestResult("accepted", {
        execution_time_ms: 42,
        memory_used_kb: 1024,
      }),
    ];

    setStoreState({
      currentSubmission: mockSubmission(results, "accepted"),
      liveStatus: "accepted",
      progress: 100,
      completedTests: 1,
      totalTests: 1,
    });

    render(<TestResults />);

    expect(screen.getByText("42ms")).toBeInTheDocument();
    expect(screen.getByText("1024KB")).toBeInTheDocument();
  });

  it("renders running tests indicator when results array is empty but liveStatus is set", () => {
    setStoreState({
      currentSubmission: mockSubmission([], "running"),
      liveStatus: "running",
      progress: 0,
      completedTests: 0,
      totalTests: 5,
    });

    render(<TestResults />);

    expect(screen.getByText("Running tests...")).toBeInTheDocument();
  });

  it("shows green progress bar when accepted", () => {
    setStoreState({
      currentSubmission: mockSubmission([mockTestResult("accepted")], "accepted"),
      liveStatus: "accepted",
      progress: 100,
      completedTests: 1,
      totalTests: 1,
    });

    const { container } = render(<TestResults />);

    // Progress bar fill should have green-500 styles
    const progressFill = container.querySelector(".bg-green-500");
    expect(progressFill).toBeInTheDocument();
  });

  it("shows indigo progress bar when not accepted", () => {
    setStoreState({
      currentSubmission: mockSubmission([mockTestResult("wrong_answer")], "wrong_answer"),
      liveStatus: "wrong_answer",
      progress: 100,
      completedTests: 1,
      totalTests: 1,
    });

    const { container } = render(<TestResults />);

    // Progress bar fill should have indigo-500 styles
    const progressFill = container.querySelector(".bg-indigo-500");
    expect(progressFill).toBeInTheDocument();
  });
});
