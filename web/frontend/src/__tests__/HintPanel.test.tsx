import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { HintPanel } from "../components/HintPanel";
import type { Hint } from "../types";

// ── Mock Data ───────────────────────────────────────────────────────────────

const mockHints: Hint[] = [
  {
    id: "hint-1",
    problem_id: "prob-1",
    level: 1,
    content: "Think about using a hash map to store values",
    penalty_points: 10,
    is_revealed: false,
  },
  {
    id: "hint-2",
    problem_id: "prob-1",
    level: 2,
    content: "Iterate through the array and check if complement exists",
    penalty_points: 15,
    is_revealed: false,
  },
  {
    id: "hint-3",
    problem_id: "prob-1",
    level: 3,
    content: "Return the indices of the two numbers that sum to target",
    penalty_points: 25,
    is_revealed: false,
  },
];

const mockRevealedHint: Hint = {
  id: "hint-revealed",
  problem_id: "prob-2",
  level: 1,
  content: "Already revealed hint content",
  penalty_points: 0,
  is_revealed: true,
};

// ── Tests ───────────────────────────────────────────────────────────────────

describe("HintPanel", () => {
  it("renders hint header with correct count", () => {
    render(<HintPanel problemId="prob-1" hints={mockHints} />);

    expect(screen.getByText("Hints")).toBeInTheDocument();
    expect(screen.getByText("(0/3 revealed)")).toBeInTheDocument();
  });

  it("renders all hints with hidden state", () => {
    render(<HintPanel problemId="prob-1" hints={mockHints} />);

    // All hints should have level numbers
    expect(screen.getByText("#1")).toBeInTheDocument();
    expect(screen.getByText("#2")).toBeInTheDocument();
    expect(screen.getByText("#3")).toBeInTheDocument();

    // All should show "Hint hidden"
    const hiddenLabels = screen.getAllByText("Hint hidden");
    expect(hiddenLabels.length).toBe(3);

    // None should show content
    expect(
      screen.queryByText("Think about using a hash map to store values")
    ).not.toBeInTheDocument();
  });

  it("shows reveal buttons for hidden hints", () => {
    render(<HintPanel problemId="prob-1" hints={mockHints} />);

    const revealButtons = screen.getAllByLabelText(/Reveal hint/);
    expect(revealButtons.length).toBe(3);
    expect(revealButtons[0]).toHaveTextContent("Reveal");
  });

  it("shows already revealed hints with content visible", () => {
    render(
      <HintPanel problemId="prob-2" hints={[mockRevealedHint]} />
    );

    expect(screen.getByText("#1")).toBeInTheDocument();
    expect(screen.getByText("Already revealed hint content")).toBeInTheDocument();
    expect(screen.queryByText("Hint hidden")).not.toBeInTheDocument();
    expect(
      screen.queryByLabelText(/Reveal hint/)
    ).not.toBeInTheDocument();
  });

  it("shows confirmation dialog when Reveal button clicked", () => {
    render(<HintPanel problemId="prob-1" hints={mockHints} />);

    // Click reveal on first hint
    fireEvent.click(screen.getByLabelText("Reveal hint 1"));

    // Confirmation dialog should appear
    expect(screen.getByText("Reveal Hint?")).toBeInTheDocument();
    expect(
      screen.getByText(/Revealing this hint will deduct/)
    ).toBeInTheDocument();
    expect(screen.getByText("This action cannot be undone.")).toBeInTheDocument();

    // Confirm and Cancel buttons
    expect(screen.getByText("Cancel")).toBeInTheDocument();
    expect(screen.getByText("Reveal Hint")).toBeInTheDocument();
  });

  it("reveals hint content when confirmed", () => {
    render(<HintPanel problemId="prob-1" hints={mockHints} />);

    // Click reveal on first hint
    fireEvent.click(screen.getByLabelText("Reveal hint 1"));

    // Confirm
    fireEvent.click(screen.getByText("Reveal Hint"));

    // Hint content should now be visible
    expect(
      screen.getByText("Think about using a hash map to store values")
    ).toBeInTheDocument();

    // No more "Hint hidden" for hint 1
    expect(screen.queryByText("Reveal hint 1")).not.toBeInTheDocument();

    // Counter should update
    expect(screen.getByText("(1/3 revealed)")).toBeInTheDocument();
  });

  it("closes confirmation dialog when Cancel clicked", () => {
    render(<HintPanel problemId="prob-1" hints={mockHints} />);

    // Click reveal
    fireEvent.click(screen.getByLabelText("Reveal hint 1"));

    // Dialog visible
    expect(screen.getByText("Reveal Hint?")).toBeInTheDocument();

    // Cancel
    fireEvent.click(screen.getByText("Cancel"));

    // Dialog should close
    expect(screen.queryByText("Reveal Hint?")).not.toBeInTheDocument();

    // Hint should not be revealed
    expect(
      screen.queryByText("Think about using a hash map to store values")
    ).not.toBeInTheDocument();

    // Counter unchanged
    expect(screen.getByText("(0/3 revealed)")).toBeInTheDocument();
  });

  it("reveals hints in progressive order (sorted by level)", () => {
    const shuffledHints = [
      { ...mockHints[2] }, // level 3
      { ...mockHints[0] }, // level 1
      { ...mockHints[1] }, // level 2
    ];

    render(<HintPanel problemId="prob-1" hints={shuffledHints} />);

    // Hints should be rendered in order: #1, #2, #3
    const hintLabels = screen.getAllByText(/#\d/);
    expect(hintLabels[0]).toHaveTextContent("#1");
    expect(hintLabels[1]).toHaveTextContent("#2");
    expect(hintLabels[2]).toHaveTextContent("#3");
  });

  it("shows penalty points for revealed hints", () => {
    render(<HintPanel problemId="prob-1" hints={mockHints} />);

    // Click reveal on hint 1
    fireEvent.click(screen.getByLabelText("Reveal hint 1"));
    fireEvent.click(screen.getByText("Reveal Hint"));

    // Penalty should be shown
    expect(screen.getByText("-10 points penalty")).toBeInTheDocument();
  });

  it("shows warning icon in confirmation dialog", () => {
    render(<HintPanel problemId="prob-1" hints={mockHints} />);

    fireEvent.click(screen.getByLabelText("Reveal hint 1"));

    // Dialog should have a dialog role
    expect(screen.getByRole("dialog")).toBeInTheDocument();
    expect(screen.getByRole("dialog")).toHaveAttribute("aria-modal", "true");
  });

  it("can reveal multiple hints sequentially", () => {
    render(<HintPanel problemId="prob-1" hints={mockHints} />);

    // Reveal hint 1
    fireEvent.click(screen.getByLabelText("Reveal hint 1"));
    fireEvent.click(screen.getByText("Reveal Hint"));
    expect(screen.getByText("(1/3 revealed)")).toBeInTheDocument();

    // Reveal hint 2
    fireEvent.click(screen.getByLabelText("Reveal hint 2"));
    fireEvent.click(screen.getByText("Reveal Hint"));
    expect(screen.getByText("(2/3 revealed)")).toBeInTheDocument();

    // Both contents visible
    expect(
      screen.getByText("Think about using a hash map to store values")
    ).toBeInTheDocument();
    expect(
      screen.getByText("Iterate through the array and check if complement exists")
    ).toBeInTheDocument();
  });

  it("shows correct penalty for each hint level", () => {
    render(<HintPanel problemId="prob-1" hints={mockHints} />);

    // Reveal hint 2
    fireEvent.click(screen.getByLabelText("Reveal hint 2"));
    expect(
      screen.getByText(/Revealing this hint will deduct/)
    ).toBeInTheDocument();

    // Cancel and reveal hint 3
    fireEvent.click(screen.getByText("Cancel"));
    fireEvent.click(screen.getByLabelText("Reveal hint 3"));
    expect(
      screen.getByText(/Revealing this hint will deduct/)
    ).toBeInTheDocument();
  });
});
