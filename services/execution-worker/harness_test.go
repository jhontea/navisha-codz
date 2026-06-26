package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// ============================================================================
// Harness Generation Tests
// ============================================================================

func TestDefaultHarnessConfig(t *testing.T) {
	cfg := DefaultHarnessConfig()
	if cfg.Mode != HarnessModeNormal {
		t.Errorf("expected normal mode, got %s", cfg.Mode)
	}
	if cfg.BenchmarkIterations != 100 {
		t.Errorf("expected 100 benchmark iterations, got %d", cfg.BenchmarkIterations)
	}
	if cfg.FuzzIterations != 50 {
		t.Errorf("expected 50 fuzz iterations, got %d", cfg.FuzzIterations)
	}
}

func TestGenerateFunctionHarness_Basic(t *testing.T) {
	code := `func solution(input string) string { return input }`
	problem := ProblemDefinition{
		ID:           1,
		Title:        "Test Problem",
		Type:         "function",
		FunctionName: "solution",
		TestCases: []TestCaseDef{
			{ID: 1, Input: "hello", Expected: "hello", Weight: 1},
		},
	}
	cfg := DefaultHarnessConfig()

	harness := generateFunctionHarness(code, problem, cfg)

	// Check essential elements
	checkContains(t, harness, "package main")
	checkContains(t, harness, "func main()")
	checkContains(t, harness, code)
	checkContains(t, harness, `"reflect"`)
	checkContains(t, harness, `"encoding/json"`)
	checkContains(t, harness, `"fmt"`)

	// Should NOT have benchmark, fuzz, or custom imports
	checkNotContains(t, harness, `"time"`)
	checkNotContains(t, harness, `"math/rand"`)
	checkNotContains(t, harness, `"runtime"`)
}

func TestGenerateFunctionHarness_WithMemoryTracking(t *testing.T) {
	code := `func solution(input string) string { return input }`
	problem := ProblemDefinition{
		ID:           1,
		Title:        "Test",
		Type:         "function",
		FunctionName: "solution",
		TestCases: []TestCaseDef{
			{ID: 1, Input: "test", Expected: "test"},
		},
	}
	cfg := HarnessConfig{
		Mode:                 HarnessModeNormal,
		EnableMemoryTracking: true,
	}

	harness := generateFunctionHarness(code, problem, cfg)
	checkContains(t, harness, `"runtime"`)
}

func TestGenerateFunctionHarness_WithCPUProfiling(t *testing.T) {
	code := `func solution(input string) string { return input }`
	problem := ProblemDefinition{
		ID:           1,
		Title:        "Test",
		Type:         "function",
		FunctionName: "solution",
		TestCases: []TestCaseDef{
			{ID: 1, Input: "test", Expected: "test"},
		},
	}
	cfg := HarnessConfig{
		Mode:              HarnessModeNormal,
		EnableCPUProfiling: true,
	}

	harness := generateFunctionHarness(code, problem, cfg)
	checkContains(t, harness, `"runtime/pprof"`)
}

func TestGenerateFunctionHarness_BenchmarkMode(t *testing.T) {
	code := `func solution(input string) string { return input }`
	problem := ProblemDefinition{
		ID:           1,
		Title:        "Benchmark Test",
		Type:         "function",
		FunctionName: "solution",
		TestCases: []TestCaseDef{
			{ID: 1, Input: "test", Expected: "test"},
		},
	}
	cfg := HarnessConfig{
		Mode:                HarnessModeBenchmark,
		BenchmarkIterations: 50,
	}

	harness := generateFunctionHarness(code, problem, cfg)
	checkContains(t, harness, `"time"`)
	checkContains(t, harness, "Benchmark:")
	checkContains(t, harness, "Warmup")
	checkContains(t, harness, "Benchmark mode doesn't check correctness")
}

func TestGenerateFunctionHarness_FuzzMode(t *testing.T) {
	code := `func solution(input string) string { return input }`
	problem := ProblemDefinition{
		ID:           1,
		Title:        "Fuzz Test",
		Type:         "function",
		FunctionName: "solution",
		TestCases: []TestCaseDef{
			{ID: 1, Input: "test", Expected: "test"},
		},
	}
	cfg := HarnessConfig{
		Mode:           HarnessModeFuzz,
		FuzzIterations: 10,
	}

	harness := generateFunctionHarness(code, problem, cfg)
	checkContains(t, harness, `"math/rand"`)
	checkContains(t, harness, "Fuzz:")
	checkContains(t, harness, "rand.Intn")
}

func TestGenerateFunctionHarness_CustomMode(t *testing.T) {
	code := `func solution(input string) string { return input }`
	problem := ProblemDefinition{
		ID:           1,
		Title:        "Custom Test",
		Type:         "function",
		FunctionName: "solution",
		TestCases: []TestCaseDef{
			{ID: 1, Input: "test", Expected: "test"},
		},
	}
	cfg := HarnessConfig{
		Mode:         HarnessModeCustom,
		CustomRunner: "func customTestRunner(input, expected string) (bool, string) { return true, \"\" }",
	}

	harness := generateFunctionHarness(code, problem, cfg)
	checkContains(t, harness, "customTestRunner")
	checkContains(t, harness, cfg.CustomRunner)
}

func TestGenerateFunctionHarness_EmptyFunctionName(t *testing.T) {
	code := `func solution(input string) string { return input }`
	problem := ProblemDefinition{
		ID:      1,
		Title:   "No Func Name",
		Type:    "function",
		TestCases: []TestCaseDef{
			{ID: 1, Input: "test", Expected: "test"},
		},
	}
	cfg := DefaultHarnessConfig()

	harness := generateFunctionHarness(code, problem, cfg)
	// Should default to "solution"
	checkContains(t, harness, `resultVal := solution(inputStr)`)
}

// ============================================================================
// Main Harness Tests
// ============================================================================

func TestGenerateMainHarness_Basic(t *testing.T) {
	code := `package main
import "fmt"
func main() { fmt.Println("test") }`
	problem := ProblemDefinition{
		ID:    1,
		Title: "Main Test",
		Type:  "main",
		TestCases: []TestCaseDef{
			{ID: 1, Input: "input", Expected: "test", Weight: 1},
		},
	}
	cfg := DefaultHarnessConfig()

	harness := generateMainHarness(code, problem, cfg)
	checkContains(t, harness, "package main")
	checkContains(t, harness, "func main()")
	checkContains(t, harness, `"os/exec"`)
	checkContains(t, harness, "exec.Command")
}

// ============================================================================
// parseTestResults Tests
// ============================================================================

func TestParseTestResults_JSONArray(t *testing.T) {
	output := `[{"id":1,"passed":true,"execution_time_ms":10,"memory_used_kb":100}]`
	testCases := []TestCaseDef{{ID: 1, Weight: 1}}

	results := parseTestResults(output, testCases)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].TestCaseID != 1 {
		t.Errorf("expected test case ID 1, got %d", results[0].TestCaseID)
	}
	if results[0].Status != "passed" {
		t.Errorf("expected status 'passed', got %q", results[0].Status)
	}
	if results[0].ExecutionTimeMs != 10 {
		t.Errorf("expected 10ms, got %d", results[0].ExecutionTimeMs)
	}
	if results[0].MemoryUsedKb != 100 {
		t.Errorf("expected 100KB, got %d", results[0].MemoryUsedKb)
	}
}

func TestParseTestResults_JSONArray_Failed(t *testing.T) {
	output := `[{"id":2,"passed":false,"error":"wrong answer","execution_time_ms":5}]`
	testCases := []TestCaseDef{{ID: 2, Weight: 1}}

	results := parseTestResults(output, testCases)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Status != "failed" {
		t.Errorf("expected status 'failed', got %q", results[0].Status)
	}
	if results[0].ErrorMessage != "wrong answer" {
		t.Errorf("expected error 'wrong answer', got %q", results[0].ErrorMessage)
	}
}

func TestParseTestResults_LineByLine(t *testing.T) {
	output := `{"id":1,"passed":true}
{"id":2,"passed":false,"error":"failed"}`
	testCases := []TestCaseDef{{ID: 1}, {ID: 2}}

	results := parseTestResults(output, testCases)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].TestCaseID != 1 || results[0].Status != "passed" {
		t.Errorf("expected result 1 passed, got %+v", results[0])
	}
	if results[1].TestCaseID != 2 || results[1].Status != "failed" {
		t.Errorf("expected result 2 failed, got %+v", results[1])
	}
}

func TestParseTestResults_EmptyOutput(t *testing.T) {
	output := ""
	testCases := []TestCaseDef{{ID: 1, Weight: 1}, {ID: 2, Weight: 1}}

	results := parseTestResults(output, testCases)
	if len(results) != 2 {
		t.Fatalf("expected 2 results (fallback for empty), got %d", len(results))
	}
	for _, r := range results {
		if r.Status != "error" {
			t.Errorf("expected status 'error', got %q", r.Status)
		}
	}
}

func TestParseTestResults_InvalidJSON(t *testing.T) {
	output := "not json at all"
	testCases := []TestCaseDef{{ID: 1}}

	results := parseTestResults(output, testCases)
	if len(results) != 1 {
		t.Fatalf("expected 1 result (fallback), got %d", len(results))
	}
	if results[0].Status != "error" {
		t.Errorf("expected status 'error', got %q", results[0].Status)
	}
}

func TestParseTestResults_WithMetrics(t *testing.T) {
	output := `[{"id":1,"passed":true,"execution_time_ms":100,"memory_used_kb":512,"cpu_time_ms":50,"disk_io_kb":10,"network_io_kb":5,"output_size":1000}]`
	testCases := []TestCaseDef{{ID: 1}}

	results := parseTestResults(output, testCases)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.CPUTimeMs != 50 {
		t.Errorf("expected CPU time 50, got %d", r.CPUTimeMs)
	}
	if r.DiskIOKB != 10 {
		t.Errorf("expected DiskIO 10, got %d", r.DiskIOKB)
	}
	if r.NetworkIOKB != 5 {
		t.Errorf("expected NetworkIO 5, got %d", r.NetworkIOKB)
	}
	if r.OutputSize != 1000 {
		t.Errorf("expected OutputSize 1000, got %d", r.OutputSize)
	}
}

// ============================================================================
// compareResults Tests
// ============================================================================

func TestCompareResults_DirectEqual(t *testing.T) {
	if !compareResults(42, 42) {
		t.Error("expected 42 == 42")
	}
	if !compareResults("hello", "hello") {
		t.Error("expected 'hello' == 'hello'")
	}
	if !compareResults([]int{1, 2, 3}, []int{1, 2, 3}) {
		t.Error("expected slices to be equal")
	}
}

func TestCompareResults_StringNormalized(t *testing.T) {
	if !compareResults("hello", "  hello  ") {
		t.Error("expected strings to match after normalization")
	}
	if !compareResults(42, "42") {
		t.Error("expected int 42 to match string '42'")
	}
}

func TestCompareResults_Numeric(t *testing.T) {
	if !compareResults(42, 42.0) {
		t.Error("expected int 42 == float64 42")
	}
	if !compareResults(3.14, 3.14) {
		t.Error("expected float64 == float64")
	}
}

func TestCompareResults_NotEqual(t *testing.T) {
	if compareResults(42, 43) {
		t.Error("expected 42 != 43")
	}
	if compareResults("hello", "world") {
		t.Error("expected 'hello' != 'world'")
	}
}

// ============================================================================
// toFloat64 Tests
// ============================================================================

func TestToFloat64(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected float64
		hasErr   bool
	}{
		{float64(3.14), 3.14, false},
		{float32(2.5), 2.5, false},
		{42, 42.0, false},
		{int32(100), 100.0, false},
		{int64(999), 999.0, false},
		{"3.14", 3.14, false},
		{"not a number", 0, true},
		{true, 0, true},
		{[]int{1}, 0, true},
	}
	for _, tt := range tests {
		result, err := toFloat64(tt.input)
		if tt.hasErr {
			if err == nil {
				t.Errorf("expected error for %v (type %T)", tt.input, tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("unexpected error for %v: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("toFloat64(%v) = %f, want %f", tt.input, result, tt.expected)
			}
		}
	}
}

// ============================================================================
// calculateScore Tests
// ============================================================================

func TestCalculateScore_AllPassed(t *testing.T) {
	results := []TestResult{
		{TestCaseID: 1, Status: "passed"},
		{TestCaseID: 2, Status: "passed"},
	}
	testCases := []TestCaseDef{
		{ID: 1, Weight: 1},
		{ID: 2, Weight: 1},
	}
	score := calculateScore(results, testCases, 100)
	if score != 100 {
		t.Errorf("expected 100, got %d", score)
	}
}

func TestCalculateScore_AllFailed(t *testing.T) {
	results := []TestResult{
		{TestCaseID: 1, Status: "failed"},
		{TestCaseID: 2, Status: "failed"},
	}
	testCases := []TestCaseDef{
		{ID: 1, Weight: 1},
		{ID: 2, Weight: 1},
	}
	score := calculateScore(results, testCases, 100)
	if score != 0 {
		t.Errorf("expected 0, got %d", score)
	}
}

func TestCalculateScore_HalfPassed(t *testing.T) {
	results := []TestResult{
		{TestCaseID: 1, Status: "passed"},
		{TestCaseID: 2, Status: "failed"},
	}
	testCases := []TestCaseDef{
		{ID: 1, Weight: 1},
		{ID: 2, Weight: 1},
	}
	score := calculateScore(results, testCases, 100)
	if score != 50 {
		t.Errorf("expected 50, got %d", score)
	}
}

func TestCalculateScore_Weighted(t *testing.T) {
	results := []TestResult{
		{TestCaseID: 1, Status: "passed"},
		{TestCaseID: 2, Status: "failed"},
	}
	testCases := []TestCaseDef{
		{ID: 1, Weight: 3},
		{ID: 2, Weight: 1},
	}
	score := calculateScore(results, testCases, 100)
	if score != 75 {
		t.Errorf("expected 75, got %d", score)
	}
}

func TestCalculateScore_EmptyInputs(t *testing.T) {
	if score := calculateScore(nil, nil, 100); score != 0 {
		t.Errorf("expected 0 for nil inputs, got %d", score)
	}
	if score := calculateScore([]TestResult{}, []TestCaseDef{}, 100); score != 0 {
		t.Errorf("expected 0 for empty inputs, got %d", score)
	}
}

func TestCalculateScore_ZeroWeight(t *testing.T) {
	results := []TestResult{
		{TestCaseID: 1, Status: "passed"},
	}
	testCases := []TestCaseDef{
		{ID: 1, Weight: 0},
	}
	score := calculateScore(results, testCases, 100)
	if score != 100 {
		t.Errorf("expected 100 (weight defaults to 1), got %d", score)
	}
}

// ============================================================================
// generateFuzzTestCase Tests
// ============================================================================

func TestGenerateFuzzTestCase(t *testing.T) {
	tc := generateFuzzTestCase(42)
	if tc.ID != 42 {
		t.Errorf("expected ID 42, got %d", tc.ID)
	}
	if tc.Input == "" {
		t.Error("expected non-empty input")
	}
	if tc.Expected == "" {
		t.Error("expected non-empty expected")
	}
}

func TestGenerateFuzzTestCase_Deterministic(t *testing.T) {
	tc1 := generateFuzzTestCase(123)
	tc2 := generateFuzzTestCase(123)
	if tc1.Input != tc2.Input || tc1.Expected != tc2.Expected {
		t.Error("expected deterministic fuzz test cases")
	}
}

// ============================================================================
// SanitizeHarnessOutput Tests
// ============================================================================

func TestSanitizeHarnessOutput(t *testing.T) {
	output := "Benchmark: 100 iterations per test case\n{\"result\": true}\nFuzz: generating 10 random test cases\n"
	sanitized := SanitizeHarnessOutput(output)
	if sanitized != "{\"result\": true}" {
		t.Errorf("expected only JSON, got %q", sanitized)
	}
}

func TestSanitizeHarnessOutput_NoMetadata(t *testing.T) {
	output := "[{\"id\":1,\"passed\":true}]\n"
	sanitized := SanitizeHarnessOutput(output)
	if sanitized != "[{\"id\":1,\"passed\":true}]" {
		t.Errorf("expected unchanged output, got %q", sanitized)
	}
}

// ============================================================================
// ResourceLimits / WorkerConfig Tests
// ============================================================================

func TestDefaultResourceLimits_Easy(t *testing.T) {
	limits := DefaultResourceLimits(DifficultyEasy)
	if limits.TimeLimit != 1*time.Second {
		t.Errorf("expected 1s time limit, got %v", limits.TimeLimit)
	}
	if limits.MemoryLimitMB != 256 {
		t.Errorf("expected 256MB, got %d", limits.MemoryLimitMB)
	}
	if limits.CPUQuota != "50000" {
		t.Errorf("expected CPU quota 50000, got %s", limits.CPUQuota)
	}
}

func TestDefaultResourceLimits_Medium(t *testing.T) {
	limits := DefaultResourceLimits(DifficultyMedium)
	if limits.TimeLimit != 3*time.Second {
		t.Errorf("expected 3s time limit, got %v", limits.TimeLimit)
	}
	if limits.MemoryLimitMB != 512 {
		t.Errorf("expected 512MB, got %d", limits.MemoryLimitMB)
	}
}

func TestDefaultResourceLimits_Hard(t *testing.T) {
	limits := DefaultResourceLimits(DifficultyHard)
	if limits.TimeLimit != 5*time.Second {
		t.Errorf("expected 5s time limit, got %v", limits.TimeLimit)
	}
	if limits.MemoryLimitMB != 1024 {
		t.Errorf("expected 1024MB, got %d", limits.MemoryLimitMB)
	}
}

func TestDefaultResourceLimits_Unknown(t *testing.T) {
	limits := DefaultResourceLimits("unknown")
	// Should default to medium limits
	if limits.TimeLimit != 3*time.Second {
		t.Errorf("expected 3s (medium default), got %v", limits.TimeLimit)
	}
}

func TestDefaultWorkerConfig(t *testing.T) {
	cfg := DefaultWorkerConfig()
	if cfg.MaxConcurrentExecutions != 4 {
		t.Errorf("expected 4, got %d", cfg.MaxConcurrentExecutions)
	}
	if cfg.DefaultDifficulty != DifficultyMedium {
		t.Errorf("expected medium, got %s", cfg.DefaultDifficulty)
	}
	if cfg.WarmPoolSize != 2 {
		t.Errorf("expected pool size 2, got %d", cfg.WarmPoolSize)
	}
}

func TestGetLimitsForProblem_Basic(t *testing.T) {
	cfg := DefaultWorkerConfig()
	limits := cfg.GetLimitsForProblem(DifficultyEasy, "algorithms")
	if limits.TimeLimit != 1*time.Second {
		t.Errorf("expected 1s for easy, got %v", limits.TimeLimit)
	}
}

func TestGetLimitsForProblem_UnknownDifficulty(t *testing.T) {
	cfg := DefaultWorkerConfig()
	limits := cfg.GetLimitsForProblem("unknown", "")
	if limits.TimeLimit != 3*time.Second {
		t.Errorf("expected 3s (default), got %v", limits.TimeLimit)
	}
}

func TestGetLimitsForProblem_WithCategoryOverride(t *testing.T) {
	cfg := DefaultWorkerConfig()
	cfg.CategoryOverrides = []CategoryOverride{
		{
			Category: "dp",
			ResourceLimits: ResourceLimits{
				TimeLimit:     10 * time.Second,
				MemoryLimitMB: 2048,
				CPUQuota:      "200000",
				PIDsLimit:     300,
				MaxOutputSize: 1024 * 1024,
				DiskLimitMB:   500,
			},
		},
	}

	limits := cfg.GetLimitsForProblem(DifficultyMedium, "dp")
	if limits.TimeLimit != 10*time.Second {
		t.Errorf("expected 10s (overridden), got %v", limits.TimeLimit)
	}
	if limits.MemoryLimitMB != 2048 {
		t.Errorf("expected 2048MB (overridden), got %d", limits.MemoryLimitMB)
	}
}

func TestWorkerConfigString(t *testing.T) {
	cfg := DefaultWorkerConfig()
	s := cfg.String()
	if s == "" {
		t.Error("expected non-empty string representation")
	}
}

// ============================================================================
// Sandbox Tests
// ============================================================================

func TestDefaultSandboxConfig(t *testing.T) {
	cfg := DefaultSandboxConfig()
	if len(cfg.Environments) != 3 {
		t.Errorf("expected 3 environments, got %d", len(cfg.Environments))
	}
	if cfg.DockerImage == "" {
		t.Error("expected non-empty docker image")
	}
	if cfg.GracePeriod != 2*time.Second {
		t.Errorf("expected 2s grace period, got %v", cfg.GracePeriod)
	}
	if cfg.MaxOutputSize != 1*1024*1024 {
		t.Errorf("expected 1MB max output, got %d", cfg.MaxOutputSize)
	}
}

func TestSandboxResultDefaults(t *testing.T) {
	r := SandboxResult{}
	if r.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", r.ExitCode)
	}
	if r.TimedOut {
		t.Error("expected not timed out")
	}
}

func TestNewSandboxExecutor(t *testing.T) {
	cfg := DefaultSandboxConfig()
	exec := NewSandboxExecutor(cfg)
	if exec == nil {
		t.Fatal("expected non-nil executor")
	}
}

func TestNewContainerPool(t *testing.T) {
	pool := NewContainerPool(5, "golang:1.21-alpine")
	if pool == nil {
		t.Fatal("expected non-nil pool")
	}
	if pool.size != 5 {
		t.Errorf("expected size 5, got %d", pool.size)
	}
	if pool.image != "golang:1.21-alpine" {
		t.Errorf("expected golang:1.21-alpine, got %s", pool.image)
	}
}

func TestContainerPoolAcquire_EmptyPool(t *testing.T) {
	pool := NewContainerPool(0, "test")
	id, err := pool.Acquire(nil)
	if err == nil {
		t.Error("expected error for empty pool")
	}
	if id != "" {
		t.Errorf("expected empty ID, got %q", id)
	}
}

func TestNewSandboxMemoryManager(t *testing.T) {
	sm := NewSandboxMemoryManager(10)
	if sm == nil {
		t.Fatal("expected non-nil memory manager")
	}
	if sm.size != 10 {
		t.Errorf("expected size 10, got %d", sm.size)
	}
}

func TestSandboxMemoryManager_Stats(t *testing.T) {
	sm := NewSandboxMemoryManager(20)
	stats := sm.Stats()
	if stats["cache_size"].(int) != 0 {
		t.Errorf("expected 0 cache size, got %d", stats["cache_size"])
	}
	if stats["max_size"].(int) != 20 {
		t.Errorf("expected max_size 20, got %d", stats["max_size"])
	}
}

// ============================================================================
// Parse Functions Tests
// ============================================================================

func TestParseMemoryToKB(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"", 0},
		{"--", 0},
		{"100 KiB", 100},
		{"10 MiB", 10240},
		{"1 GiB", 1048576},
		{"512 KB", 512},
		{"256 MB", 262144},
		{"1024 B", 1},
		{"500", 500},
	}
	for _, tt := range tests {
		result := parseMemoryToKB(tt.input)
		if result != tt.expected {
			t.Errorf("parseMemoryToKB(%q) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

func TestParseTrafficToKB(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"", 0},
		{"--", 0},
		{"0B", 0},
		{"10 kB", 10},
		{"5 KiB", 5},
		{"1 MB", 1024},
		{"2 MiB", 2048},
		{"1 GB", 1048576},
		{"500 B", 0}, // Less than 1KB
	}
	for _, tt := range tests {
		result := parseTrafficToKB(tt.input)
		if result != tt.expected {
			t.Errorf("parseTrafficToKB(%q) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}

func TestParseDockerStats(t *testing.T) {
	output := "10MiB / 100MiB|12.5%|1kB / 0B|5kB / 2kB"
	metrics, err := parseDockerStats(output)
	if err != nil {
		t.Fatalf("parseDockerStats() returned error: %v", err)
	}
	if metrics.MemoryPeakKB != 10240 {
		t.Errorf("expected MemoryPeakKB 10240, got %d", metrics.MemoryPeakKB)
	}
	if metrics.CPUTimeMs != 125 {
		t.Errorf("expected CPUTimeMs 125, got %d", metrics.CPUTimeMs)
	}
}

func TestParseDockerStats_Invalid(t *testing.T) {
	_, err := parseDockerStats("invalid")
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

// ============================================================================
// Sanitize Output Tests
// ============================================================================

func TestSanitizeOutput(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"\x1b[31mred\x1b[0m", "red"},
		{"normal text", "normal text"},
		{"\x00null\x00", "null"},
	}
	for _, tt := range tests {
		result := sanitizeOutput(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeOutput(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestSanitizedWriter(t *testing.T) {
	capture := make([]byte, 0)
	sw := &sanitizedWriter{inner: &captureWriter{buf: &capture}}
	n, err := sw.Write([]byte("\x1b[31mhello\x1b[0m"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != 5 {
		t.Errorf("expected 5 bytes written, got %d", n)
	}
	if string(capture) != "hello" {
		t.Errorf("expected 'hello', got %q", string(capture))
	}
}

// captureWriter implements io.Writer with a byte slice.
type captureWriter struct {
	buf *[]byte
}

func (w *captureWriter) Write(p []byte) (int, error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}

func TestLimitedWriter(t *testing.T) {
	buf := make([]byte, 0, 100)
	lw := &limitedWriter{
		w:     &captureWriter{buf: &buf},
		limit: 5,
	}

	n, err := lw.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != 5 {
		t.Errorf("expected 5, got %d", n)
	}

	_, err = lw.Write([]byte("world"))
	if err == nil {
		t.Error("expected error when exceeding limit")
	}
	if string(buf) != "hello" {
		t.Errorf("expected 'hello', got %q", string(buf))
	}
}

// ============================================================================
// Sandbox Escape Detection Tests
// ============================================================================

func TestDetectSandboxEscape(t *testing.T) {
	tests := []struct {
		output  string
		escape  bool
		message string
	}{
		{"normal output", false, ""},
		{"/proc/1/cmdline", true, "attempt to read host process info"},
		{"/etc/shadow", true, "attempt to read sensitive host file"},
		{"--privileged", true, "privilege escalation attempt"},
		{"nsenter --target 1", true, "namespace escape attempt"},
		{"mount --bind /proc", true, "bind mount escape attempt"},
	}
	for _, tt := range tests {
		detected, msg := detectSandboxEscape(tt.output)
		if detected != tt.escape {
			t.Errorf("detectSandboxEscape(%q) detected=%v, want %v", tt.output, detected, tt.escape)
		}
		if tt.escape && msg != tt.message {
			t.Errorf("expected message %q, got %q", tt.message, msg)
		}
	}
}

// ============================================================================
// JSON Round-trip Tests
// ============================================================================

func TestTestResultJSON(t *testing.T) {
	tr := TestResult{
		TestCaseID:      1,
		Status:          "passed",
		ExecutionTimeMs: 10,
		MemoryUsedKb:    512,
		ErrorMessage:    "",
		ActualOutput:    "hello",
		ExpectedOutput:  "hello",
	}
	data, err := json.Marshal(tr)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	var decoded TestResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if decoded.TestCaseID != tr.TestCaseID {
		t.Errorf("expected %d, got %d", tr.TestCaseID, decoded.TestCaseID)
	}
}

// ============================================================================
// Helper function for checking harness content
// ============================================================================

func checkContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("expected output to contain %q", substr)
	}
}

func checkNotContains(t *testing.T, s, substr string) {
	t.Helper()
	if strings.Contains(s, substr) {
		t.Errorf("expected output NOT to contain %q", substr)
	}
}
