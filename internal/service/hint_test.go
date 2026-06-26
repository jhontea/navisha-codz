package service

import (
	"os"
	"strings"
	"testing"

	"coding-challange/internal/model"
)

func TestNewHintService(t *testing.T) {
	svc := NewHintService()
	if svc == nil {
		t.Fatal("expected HintService to be non-nil")
	}
}

func TestGetHints_NilProblem(t *testing.T) {
	svc := NewHintService()
	hints := svc.GetHints("test-user", nil)
	if hints != nil {
		t.Errorf("expected nil for nil problem, got %v", hints)
	}
}

func TestGetHints_ProgressiveReveal(t *testing.T) {
	svc := NewHintService()
	problem := &model.Problem{
		ID: "test-problem",
		Hints: []model.Hint{
			{Level: 1, Title: "Hint 1", Content: "Content 1"},
			{Level: 2, Title: "Hint 2", Content: "Content 2"},
			{Level: 3, Title: "Hint 3", Content: "Content 3"},
		},
	}

	// First request: should get first 2 hints
	hints := svc.GetHints("test-user", problem)
	if len(hints) != 2 {
		t.Errorf("expected 2 hints on first request, got %d", len(hints))
	}
	if hints[0].Level != 1 || hints[1].Level != 2 {
		t.Error("expected hints level 1 and 2")
	}

	// Second request: should get remaining 1 hint
	hints = svc.GetHints("test-user", problem)
	if len(hints) != 1 {
		t.Errorf("expected 1 hint on second request, got %d", len(hints))
	}
	if hints[0].Level != 3 {
		t.Error("expected hint level 3")
	}

	// Third request: all hints already revealed, should return all
	hints = svc.GetHints("test-user", problem)
	if len(hints) != 3 {
		t.Errorf("expected all 3 hints on third request, got %d", len(hints))
	}
}

func TestGetHints_SingleHint(t *testing.T) {
	svc := NewHintService()
	problem := &model.Problem{
		ID: "single-hint",
		Hints: []model.Hint{
			{Level: 1, Title: "Only Hint", Content: "Content"},
		},
	}

	hints := svc.GetHints("test-user", problem)
	if len(hints) != 1 {
		t.Errorf("expected 1 hint, got %d", len(hints))
	}

	// Second request: all revealed
	hints = svc.GetHints("test-user", problem)
	if len(hints) != 1 {
		t.Errorf("expected 1 hint (all revealed), got %d", len(hints))
	}
}

func TestGetHints_FourHints(t *testing.T) {
	svc := NewHintService()
	problem := &model.Problem{
		ID: "four-hints",
		Hints: []model.Hint{
			{Level: 1, Title: "H1", Content: "C1"},
			{Level: 2, Title: "H2", Content: "C2"},
			{Level: 3, Title: "H3", Content: "C3"},
			{Level: 4, Title: "H4", Content: "C4"},
		},
	}

	// First: 2 hints
	hints := svc.GetHints("test-user", problem)
	if len(hints) != 2 {
		t.Errorf("expected 2, got %d", len(hints))
	}

	// Second: remaining 2
	hints = svc.GetHints("test-user", problem)
	if len(hints) != 2 {
		t.Errorf("expected 2, got %d", len(hints))
	}

	// Third: all 4
	hints = svc.GetHints("test-user", problem)
	if len(hints) != 4 {
		t.Errorf("expected 4, got %d", len(hints))
	}
}

func TestGetHints_SortedByLevel(t *testing.T) {
	svc := NewHintService()
	problem := &model.Problem{
		ID: "sorted-hints",
		Hints: []model.Hint{
			{Level: 3, Title: "H3", Content: "C3"},
			{Level: 1, Title: "H1", Content: "C1"},
			{Level: 2, Title: "H2", Content: "C2"},
		},
	}

	hints := svc.GetHints("test-user", problem)
	if len(hints) < 2 {
		t.Fatal("expected at least 2 hints")
	}
	// Should be sorted by level
	if hints[0].Level >= hints[1].Level {
		t.Error("hints should be sorted by level ascending")
	}
}

func TestGetFullHints(t *testing.T) {
	svc := NewHintService()
	problem := &model.Problem{
		ID: "full-hints",
		Hints: []model.Hint{
			{Level: 1, Title: "H1", Content: "C1"},
			{Level: 2, Title: "H2", Content: "C2"},
		},
	}

	hints := svc.GetFullHints(problem)
	if len(hints) != 2 {
		t.Errorf("expected 2 full hints, got %d", len(hints))
	}
}

func TestGetFullHints_NilProblem(t *testing.T) {
	svc := NewHintService()
	hints := svc.GetFullHints(nil)
	if hints != nil {
		t.Error("expected nil for nil problem")
	}
}

func TestReset(t *testing.T) {
	svc := NewHintService()
	problem := &model.Problem{
		ID: "reset-test",
		Hints: []model.Hint{
			{Level: 1, Title: "H1", Content: "C1"},
			{Level: 2, Title: "H2", Content: "C2"},
			{Level: 3, Title: "H3", Content: "C3"},
		},
	}

	// Reveal some hints
	svc.GetHints("test-user", problem)

	// Reset
	svc.Reset("test-user", "reset-test")

	// Should reveal from beginning again
	hints := svc.GetHints("test-user", problem)
	if len(hints) != 2 {
		t.Errorf("expected 2 hints after reset, got %d", len(hints))
	}
}

func TestReset_NonExistent(t *testing.T) {
	svc := NewHintService()
	// Should not panic
	svc.Reset("test-user", "nonexistent-id")
}

func TestSanitizeProblemID_Valid(t *testing.T) {
	validIDs := []string{"two-sum", "problem123", "my_problem", "A", "a", "1", "test-123_abc"}
	for _, id := range validIDs {
		if err := SanitizeProblemID(id); err != nil {
			t.Errorf("expected %q to be valid, got error: %v", id, err)
		}
	}
}

func TestSanitizeProblemID_Empty(t *testing.T) {
	err := SanitizeProblemID("")
	if err == nil {
		t.Error("expected error for empty ID")
	}
}

func TestSanitizeProblemID_InvalidChars(t *testing.T) {
	invalidIDs := []string{
		"two sum",
		"two$sum",
		"../../etc/passwd",
		"problem.id",
		"test/path",
		"hello world",
		"special@char",
		"back\\slash",
	}
	for _, id := range invalidIDs {
		if err := SanitizeProblemID(id); err == nil {
			t.Errorf("expected %q to be invalid", id)
		}
	}
}

func TestSanitizeProblemID_PathTraversal(t *testing.T) {
	// Path traversal attempts should be rejected
	ids := []string{"../secret", "..\\windows", "/etc/passwd", "foo/../bar"}
	for _, id := range ids {
		if err := SanitizeProblemID(id); err == nil {
			t.Errorf("path traversal %q should be rejected", id)
		}
	}
}

func TestValidateCodeSize_Valid(t *testing.T) {
	if err := ValidateCodeSize("func main() {}"); err != nil {
		t.Errorf("expected small code to be valid, got: %v", err)
	}
}

func TestValidateCodeSize_Empty(t *testing.T) {
	if err := ValidateCodeSize(""); err != nil {
		t.Errorf("expected empty code to be valid, got: %v", err)
	}
}

func TestValidateCodeSize_Exactly64KB(t *testing.T) {
	code := strings.Repeat("a", 64*1024)
	if err := ValidateCodeSize(code); err != nil {
		t.Errorf("expected 64KB code to be valid, got: %v", err)
	}
}

func TestValidateCodeSize_Over64KB(t *testing.T) {
	code := strings.Repeat("a", 64*1024+1)
	err := ValidateCodeSize(code)
	if err == nil {
		t.Error("expected error for code over 64KB")
	}
}

func TestBuildTestHarness_FunctionBased(t *testing.T) {
	problem := &model.Problem{
		Type:         model.ProblemTypeFunction,
		FunctionName: "twoSum",
		Parameters: []model.Parameter{
			{Name: "nums", Type: "[]int"},
			{Name: "target", Type: "int"},
		},
		ReturnType: "[]int",
		TestCases: []model.TestCase{
			{Params: []any{[]any{1, 2, 3}, 5}, Expected: []any{1, 2}},
		},
	}

	code := "func twoSum(nums []int, target int) []int { return nil }"
	harness := BuildTestHarness(code, problem)

	// Verify harness contains key elements
	if !strings.Contains(harness, "package main") {
		t.Error("harness should contain 'package main'")
	}
	if !strings.Contains(harness, "func main()") {
		t.Error("harness should contain 'func main()'")
	}
	if !strings.Contains(harness, code) {
		t.Error("harness should contain user code")
	}
}

func TestBuildTestHarness_MainBased(t *testing.T) {
	problem := &model.Problem{
		Type: model.ProblemTypeMain,
		TestCases: []model.TestCase{
			{Input: "test", Expected: "result"},
		},
	}

	code := "package main\nimport \"fmt\"\nfunc main() { fmt.Println(\"test\") }"
	harness := BuildTestHarness(code, problem)

	if !strings.Contains(harness, "package main") {
		t.Error("harness should contain 'package main'")
	}
	if !strings.Contains(harness, "func main()") {
		t.Error("harness should contain 'func main()'")
	}
}

func TestBuildTestHarness_EmptyCode(t *testing.T) {
	problem := &model.Problem{
		Type:         model.ProblemTypeFunction,
		FunctionName: "test",
		Parameters:   []model.Parameter{},
		ReturnType:   "void",
		TestCases: []model.TestCase{
			{Params: []any{}, Expected: nil},
		},
	}

	harness := BuildTestHarness("", problem)
	if !strings.Contains(harness, "package main") {
		t.Error("harness should contain 'package main' even with empty code")
	}
}

func TestBuildTestHarness_NoTestCases(t *testing.T) {
	problem := &model.Problem{
		Type:         model.ProblemTypeFunction,
		FunctionName: "test",
		Parameters:   []model.Parameter{},
		ReturnType:   "void",
		TestCases:    []model.TestCase{},
	}

	harness := BuildTestHarness("func test() {}", problem)
	if !strings.Contains(harness, "package main") {
		t.Error("harness should contain 'package main'")
	}
}

func TestEnsureTempDir(t *testing.T) {
	dir, err := EnsureTempDir()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if dir == "" {
		t.Error("expected non-empty directory path")
	}
	// Verify directory exists on filesystem
	if _, err := os.Stat(dir); err != nil {
		t.Errorf("temp dir should exist: %v", err)
	}
	// Cleanup and verify removal
	os.RemoveAll(dir)
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Errorf("temp dir should be removed after cleanup, but still exists")
	}
}

func TestGetHints_ConcurrentAccess(t *testing.T) {
	svc := NewHintService()
	problem := &model.Problem{
		ID: "concurrent-test",
		Hints: []model.Hint{
			{Level: 1, Title: "H1", Content: "C1"},
			{Level: 2, Title: "H2", Content: "C2"},
			{Level: 3, Title: "H3", Content: "C3"},
		},
	}

	// Run concurrent hint requests
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			svc.GetHints("test-user", problem)
			done <- true
		}()
	}

	for i := 0; i < 5; i++ {
		<-done
	}
}
