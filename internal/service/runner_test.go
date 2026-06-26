package service

import (
	"testing"

	"coding-challange/internal/model"
)

func TestNewRunnerService(t *testing.T) {
	runner := NewRunnerService(10, 256)
	if runner == nil {
		t.Fatal("expected RunnerService to be non-nil")
	}
}

func TestParseFunctionOutput_AllPassed(t *testing.T) {
	runner := NewRunnerService(10, 256)
	problem := &model.Problem{
		Type: model.ProblemTypeFunction,
		TestCases: []model.TestCase{
			{Expected: "hello"},
			{Expected: "world"},
		},
	}

	// Simulate JSON output from function harness
	output := `[{"name":"test_1","passed":true},{"name":"test_2","passed":true}]`
	result := runner.parseFunctionOutput(output, problem.TestCases, 100)

	if !result.Success {
		t.Error("expected success to be true")
	}
	if result.PassedCount != 2 {
		t.Errorf("expected 2 passed, got %d", result.PassedCount)
	}
	if result.TotalCount != 2 {
		t.Errorf("expected 2 total, got %d", result.TotalCount)
	}
	if result.CompilationError != nil {
		t.Error("expected no compilation error")
	}
	if result.ExecutionTimeMs != 100 {
		t.Errorf("expected 100ms, got %d", result.ExecutionTimeMs)
	}
}

func TestParseFunctionOutput_PartialPass(t *testing.T) {
	runner := NewRunnerService(10, 256)
	problem := &model.Problem{
		Type: model.ProblemTypeFunction,
		TestCases: []model.TestCase{
			{Expected: "hello"},
			{Expected: "world"},
		},
	}

	output := `[{"name":"test_1","passed":true},{"name":"test_2","passed":false,"expected":"world","actual":"wrong","error":"expected=world got=wrong"}]`
	result := runner.parseFunctionOutput(output, problem.TestCases, 50)

	if result.Success {
		t.Error("expected success to be false")
	}
	if result.PassedCount != 1 {
		t.Errorf("expected 1 passed, got %d", result.PassedCount)
	}
	if result.TestResults[1].Passed {
		t.Error("second test should have failed")
	}
}

func TestParseFunctionOutput_AllFailed(t *testing.T) {
	runner := NewRunnerService(10, 256)
	problem := &model.Problem{
		Type: model.ProblemTypeFunction,
		TestCases: []model.TestCase{
			{Expected: "hello"},
			{Expected: "world"},
		},
	}

	output := `[{"name":"test_1","passed":false,"error":"wrong output"},{"name":"test_2","passed":false,"error":"wrong output"}]`
	result := runner.parseFunctionOutput(output, problem.TestCases, 10)

	if result.Success {
		t.Error("expected success to be false")
	}
	if result.PassedCount != 0 {
		t.Errorf("expected 0 passed, got %d", result.PassedCount)
	}
}

func TestParseFunctionOutput_EmptyOutput(t *testing.T) {
	runner := NewRunnerService(10, 256)
	problem := &model.Problem{
		Type: model.ProblemTypeFunction,
		TestCases: []model.TestCase{
			{Expected: "hello"},
		},
	}

	output := ""
	result := runner.parseFunctionOutput(output, problem.TestCases, 0)

	if result.Success {
		t.Error("expected success to be false")
	}
	if result.CompilationError == nil {
		t.Error("expected compilation error to be set")
	}
}

func TestParseFunctionOutput_MoreResultsThanTests(t *testing.T) {
	runner := NewRunnerService(10, 256)
	problem := &model.Problem{
		Type: model.ProblemTypeFunction,
		TestCases: []model.TestCase{
			{Expected: "hello"},
		},
	}

	output := `[{"name":"test_1","passed":true},{"name":"test_2","passed":true}]`
	result := runner.parseFunctionOutput(output, problem.TestCases, 10)

	if !result.Success {
		t.Error("expected success (extra results ignored)")
	}
	if len(result.TestResults) != 1 {
		t.Errorf("expected 1 test result, got %d", len(result.TestResults))
	}
}

func TestParseFunctionOutput_FewerResultsThanTests(t *testing.T) {
	runner := NewRunnerService(10, 256)
	problem := &model.Problem{
		Type: model.ProblemTypeFunction,
		TestCases: []model.TestCase{
			{Expected: "hello"},
			{Expected: "world"},
			{Expected: "foo"},
		},
	}

	output := `[{"name":"test_1","passed":true}]`
	result := runner.parseFunctionOutput(output, problem.TestCases, 10)

	if result.Success {
		t.Error("expected success to be false")
	}
	if result.PassedCount != 1 {
		t.Errorf("expected 1 passed, got %d", result.PassedCount)
	}
}

func TestParseFunctionOutput_TestResultNames(t *testing.T) {
	runner := NewRunnerService(10, 256)
	problem := &model.Problem{
		Type: model.ProblemTypeFunction,
		TestCases: []model.TestCase{
			{Expected: "a"},
			{Expected: "b"},
			{Expected: "c"},
		},
	}

	output := `[{"name":"test_1","passed":true},{"name":"test_2","passed":true},{"name":"test_3","passed":true}]`
	result := runner.parseFunctionOutput(output, problem.TestCases, 10)

	expectedNames := []string{"test_1", "test_2", "test_3"}
	for i, tr := range result.TestResults {
		if tr.Name != expectedNames[i] {
			t.Errorf("expected name %q, got %q", expectedNames[i], tr.Name)
		}
	}
}

func TestAllFailed(t *testing.T) {
	runner := NewRunnerService(10, 256)
	testCases := []model.TestCase{
		{Expected: "a"},
		{Expected: "b"},
		{Expected: "c"},
	}

	results := runner.allFailed(testCases, "timeout")

	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}

	for i, r := range results {
		if r.Passed {
			t.Errorf("test %d should not have passed", i+1)
		}
		if r.Error != "timeout" {
			t.Errorf("expected error 'timeout', got %q", r.Error)
		}
	}
}

func TestAllFailed_Empty(t *testing.T) {
	runner := NewRunnerService(10, 256)
	results := runner.allFailed([]model.TestCase{}, "error")
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestRunCode_CompilationError(t *testing.T) {
	runner := NewRunnerService(10, 256)
	problem := &model.Problem{
		Type:         model.ProblemTypeFunction,
		FunctionName: "twoSum",
		Parameters: []model.Parameter{
			{Name: "nums", Type: "[]int"},
			{Name: "target", Type: "int"},
		},
		ReturnType: "[]int",
		TestCases: []model.TestCase{
			{Expected: "a"},
		},
	}

	// Invalid Go code
	badCode := "this is not valid go code !!!"
	result := runner.RunCode(badCode, problem)

	if result.Success {
		t.Error("expected success to be false for invalid code")
	}
	if result.CompilationError == nil {
		t.Error("expected compilation error to be set")
	}
}

func TestRunCode_EmptyCode(t *testing.T) {
	runner := NewRunnerService(10, 256)
	problem := &model.Problem{
		Type:         model.ProblemTypeFunction,
		FunctionName: "twoSum",
		Parameters: []model.Parameter{
			{Name: "nums", Type: "[]int"},
			{Name: "target", Type: "int"},
		},
		ReturnType: "[]int",
		TestCases: []model.TestCase{
			{Expected: "a"},
		},
	}

	result := runner.RunCode("", problem)

	if result.Success {
		t.Error("expected success to be false for empty code")
	}
}

func TestBuildTestHarness_Function(t *testing.T) {
	problem := &model.Problem{
		Type:         model.ProblemTypeFunction,
		FunctionName: "twoSum",
		Parameters: []model.Parameter{
			{Name: "nums", Type: "[]int"},
			{Name: "target", Type: "int"},
		},
		ReturnType: "[]int",
		TestCases: []model.TestCase{
			{Params: []any{[]any{2, 7, 11, 15}, 9}, Expected: []any{0, 1}},
			{Params: []any{[]any{3, 2, 4}, 6}, Expected: []any{1, 2}},
		},
	}

	code := "func twoSum(nums []int, target int) []int { return nil }"
	harness := BuildTestHarness(code, problem)

	// Verify harness contains key elements
	if len(harness) == 0 {
		t.Error("harness should not be empty")
	}
}

func TestBuildTestHarness_Main(t *testing.T) {
	problem := &model.Problem{
		Type: model.ProblemTypeMain,
		TestCases: []model.TestCase{
			{Input: "hello", Expected: "hello"},
		},
	}

	code := "package main\nimport \"fmt\"\nfunc main() { fmt.Println(\"hello\") }"
	harness := BuildTestHarness(code, problem)

	if len(harness) == 0 {
		t.Error("harness should not be empty")
	}
}

func TestStrPtr(t *testing.T) {
	s := strPtr("test")
	if s == nil {
		t.Fatal("expected non-nil pointer")
	}
	if *s != "test" {
		t.Errorf("expected 'test', got %q", *s)
	}
}
