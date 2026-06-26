import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { BrowserRouter } from "react-router-dom";
import { ProblemList } from "../components/ProblemList";
import type { Problem } from "../types";

// ── Mock Data ───────────────────────────────────────────────────────────────

const mockProblems: Problem[] = [
  {
    id: "1",
    title: "Two Sum",
    slug: "two-sum",
    description: "Find two numbers that add up to target",
    difficulty: "easy",
    category: "arrays",
    tags: ["array", "hash-map"],
    time_limit_ms: 1000,
    memory_limit_mb: 256,
    points: 100,
    solved_count: 42,
    attempt_count: 60,
    examples: [],
    constraints: [],
    function_template: "",
    test_cases: [],
    hints: [],
    created_at: "2024-01-01",
    updated_at: "2024-01-01",
  },
  {
    id: "2",
    title: "Reverse Linked List",
    slug: "reverse-linked-list",
    description: "Reverse a singly linked list",
    difficulty: "medium",
    category: "linked-lists",
    tags: ["linked-list", "recursion"],
    time_limit_ms: 1000,
    memory_limit_mb: 256,
    points: 200,
    solved_count: 30,
    attempt_count: 50,
    examples: [],
    constraints: [],
    function_template: "",
    test_cases: [],
    hints: [],
    created_at: "2024-01-02",
    updated_at: "2024-01-02",
  },
  {
    id: "3",
    title: "Merge K Sorted Lists",
    slug: "merge-k-sorted-lists",
    description: "Merge k sorted linked lists",
    difficulty: "hard",
    category: "linked-lists",
    tags: ["linked-list", "divide-and-conquer", "heap"],
    time_limit_ms: 2000,
    memory_limit_mb: 256,
    points: 300,
    solved_count: 10,
    attempt_count: 40,
    examples: [],
    constraints: [],
    function_template: "",
    test_cases: [],
    hints: [],
    created_at: "2024-01-03",
    updated_at: "2024-01-03",
  },
];

function renderWithRouter(ui: React.ReactElement) {
  return render(<BrowserRouter>{ui}</BrowserRouter>);
}

// ── Tests ───────────────────────────────────────────────────────────────────

describe("ProblemList", () => {
  it("renders problem cards with titles and difficulty badges", () => {
    renderWithRouter(
      <ProblemList
        problems={mockProblems}
        total={3}
        page={1}
        pageSize={10}
        onPageChange={vi.fn()}
      />
    );

    // All titles visible
    expect(screen.getByText("Two Sum")).toBeInTheDocument();
    expect(screen.getByText("Reverse Linked List")).toBeInTheDocument();
    expect(screen.getByText("Merge K Sorted Lists")).toBeInTheDocument();

    // Difficulty badges - also appear in <select> options, use getAllByText
    const easyBadges = screen.getAllByText("Easy");
    expect(easyBadges.length).toBeGreaterThanOrEqual(1);
    const mediumBadges = screen.getAllByText("Medium");
    expect(mediumBadges.length).toBeGreaterThanOrEqual(1);
    const hardBadges = screen.getAllByText("Hard");
    expect(hardBadges.length).toBeGreaterThanOrEqual(1);

    // Solved counts
    expect(screen.getByText("42 solved")).toBeInTheDocument();
    expect(screen.getByText("30 solved")).toBeInTheDocument();
    expect(screen.getByText("10 solved")).toBeInTheDocument();
  });

  it("shows status icons based on solved/attempt counts", () => {
    const problemsWithStatus = [
      { ...mockProblems[0], solved_count: 5, attempt_count: 5 }, // Solved
      { ...mockProblems[1], solved_count: 0, attempt_count: 3 }, // Attempted
      { ...mockProblems[2], solved_count: 0, attempt_count: 0 }, // Unattempted
    ];

    renderWithRouter(
      <ProblemList
        problems={problemsWithStatus}
        total={3}
        page={1}
        pageSize={10}
        onPageChange={vi.fn()}
      />
    );

    // Check for aria labels indicating status
    expect(screen.getByLabelText("Solved")).toBeInTheDocument();
    expect(screen.getByLabelText("Attempted")).toBeInTheDocument();
    expect(screen.getByLabelText("Unattempted")).toBeInTheDocument();
  });

  it("shows category and points for each problem", () => {
    renderWithRouter(
      <ProblemList
        problems={mockProblems}
        total={3}
        page={1}
        pageSize={10}
        onPageChange={vi.fn()}
      />
    );

    // Category labels (kebab-case replaced)
    expect(screen.getByText("arrays")).toBeInTheDocument();
    // linked-lists appears twice
    const linkedListLabels = screen.getAllByText("linked lists");
    expect(linkedListLabels.length).toBe(2);

    // Points
    expect(screen.getByText("100 pts")).toBeInTheDocument();
    expect(screen.getByText("200 pts")).toBeInTheDocument();
    expect(screen.getByText("300 pts")).toBeInTheDocument();
  });

  it("filters problems by difficulty", () => {
    renderWithRouter(
      <ProblemList
        problems={mockProblems}
        total={3}
        page={1}
        pageSize={10}
        onPageChange={vi.fn()}
      />
    );

    // All 3 visible initially
    expect(screen.getByText("Two Sum")).toBeInTheDocument();
    expect(screen.getByText("Reverse Linked List")).toBeInTheDocument();
    expect(screen.getByText("Merge K Sorted Lists")).toBeInTheDocument();

    // Select "Easy" difficulty
    const difficultySelect = screen.getByLabelText("Filter by difficulty");
    fireEvent.change(difficultySelect, { target: { value: "easy" } });

    // Only easy problem visible
    expect(screen.getByText("Two Sum")).toBeInTheDocument();
    expect(screen.queryByText("Reverse Linked List")).not.toBeInTheDocument();
    expect(screen.queryByText("Merge K Sorted Lists")).not.toBeInTheDocument();

    // Change to "Hard"
    fireEvent.change(difficultySelect, { target: { value: "hard" } });

    expect(screen.getByText("Merge K Sorted Lists")).toBeInTheDocument();
    expect(screen.queryByText("Two Sum")).not.toBeInTheDocument();
    expect(screen.queryByText("Reverse Linked List")).not.toBeInTheDocument();
  });

  it("filters problems by category", () => {
    renderWithRouter(
      <ProblemList
        problems={mockProblems}
        total={3}
        page={1}
        pageSize={10}
        onPageChange={vi.fn()}
      />
    );

    const categorySelect = screen.getByLabelText("Filter by category");
    fireEvent.change(categorySelect, { target: { value: "arrays" } });

    // Only arrays problem visible
    expect(screen.getByText("Two Sum")).toBeInTheDocument();
    expect(screen.queryByText("Reverse Linked List")).not.toBeInTheDocument();
  });

  it("filters problems by search query", () => {
    renderWithRouter(
      <ProblemList
        problems={mockProblems}
        total={3}
        page={1}
        pageSize={10}
        onPageChange={vi.fn()}
      />
    );

    const searchInput = screen.getByLabelText("Search problems");
    fireEvent.change(searchInput, { target: { value: "linked" } });

    // Problems with "linked" in title
    expect(screen.getByText("Reverse Linked List")).toBeInTheDocument();
    expect(screen.getByText("Merge K Sorted Lists")).toBeInTheDocument();
    expect(screen.queryByText("Two Sum")).not.toBeInTheDocument();
  });

  it("searches by tags", () => {
    renderWithRouter(
      <ProblemList
        problems={mockProblems}
        total={3}
        page={1}
        pageSize={10}
        onPageChange={vi.fn()}
      />
    );

    const searchInput = screen.getByLabelText("Search problems");
    fireEvent.change(searchInput, { target: { value: "hash-map" } });

    // Two Sum has "hash-map" tag
    expect(screen.getByText("Two Sum")).toBeInTheDocument();
    expect(screen.queryByText("Reverse Linked List")).not.toBeInTheDocument();
  });

  it("shows empty state when no problems match filters", () => {
    renderWithRouter(
      <ProblemList
        problems={mockProblems}
        total={3}
        page={1}
        pageSize={10}
        onPageChange={vi.fn()}
      />
    );

    const searchInput = screen.getByLabelText("Search problems");
    fireEvent.change(searchInput, { target: { value: "zzznonexistent" } });

    expect(screen.getByText("No problems found")).toBeInTheDocument();
    expect(screen.getByText("Try adjusting your filters")).toBeInTheDocument();
  });

  it("shows pagination when total pages > 1", () => {
    renderWithRouter(
      <ProblemList
        problems={mockProblems}
        total={25}
        page={1}
        pageSize={10}
        onPageChange={vi.fn()}
      />
    );

    // Pagination text
    expect(screen.getByText("Page 1 of 3")).toBeInTheDocument();

    // Navigation buttons
    const prevButton = screen.getByLabelText("Previous page");
    const nextButton = screen.getByLabelText("Next page");

    // First page, prev should be disabled
    expect(prevButton).toBeDisabled();
    expect(nextButton).not.toBeDisabled();
  });

  it("calls onPageChange when pagination buttons clicked", () => {
    const onPageChange = vi.fn();

    renderWithRouter(
      <ProblemList
        problems={mockProblems}
        total={25}
        page={2}
        pageSize={10}
        onPageChange={onPageChange}
      />
    );

    const nextButton = screen.getByLabelText("Next page");
    fireEvent.click(nextButton);
    expect(onPageChange).toHaveBeenCalledWith(3);

    const prevButton = screen.getByLabelText("Previous page");
    fireEvent.click(prevButton);
    expect(onPageChange).toHaveBeenCalledWith(1);
  });

  it("disables prev on first page and next on last page", () => {
    const { rerender } = render(
      <BrowserRouter>
        <ProblemList
          problems={mockProblems}
          total={25}
          page={1}
          pageSize={10}
          onPageChange={vi.fn()}
        />
      </BrowserRouter>
    );

    expect(screen.getByLabelText("Previous page")).toBeDisabled();
    expect(screen.getByLabelText("Next page")).not.toBeDisabled();

    rerender(
      <BrowserRouter>
        <ProblemList
          problems={mockProblems}
          total={25}
          page={3}
          pageSize={10}
          onPageChange={vi.fn()}
        />
      </BrowserRouter>
    );

    expect(screen.getByLabelText("Previous page")).not.toBeDisabled();
    expect(screen.getByLabelText("Next page")).toBeDisabled();
  });

  it("renders problem cards as links with correct href", () => {
    renderWithRouter(
      <ProblemList
        problems={mockProblems}
        total={3}
        page={1}
        pageSize={10}
        onPageChange={vi.fn()}
      />
    );

    const links = screen.getAllByRole("link");
    expect(links.length).toBe(3);
    expect(links[0]).toHaveAttribute("href", "/problems/two-sum");
    expect(links[1]).toHaveAttribute("href", "/problems/reverse-linked-list");
    expect(links[2]).toHaveAttribute("href", "/problems/merge-k-sorted-lists");
  });
});
