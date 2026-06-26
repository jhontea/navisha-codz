package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
)

// ProblemType represents the type of problem (function-based or main-based).
type ProblemType string

const (
	ProblemTypeFunction ProblemType = "function"
	ProblemTypeMain     ProblemType = "main"
)

// ProblemDefinition represents a problem with test cases for harness generation.
type ProblemDefinition struct {
	ID            int           `json:"id"`
	Title         string        `json:"title"`
	Difficulty    string        `json:"difficulty"`
	Type          string        `json:"type"` // "function" or "main"
	Category      string        `json:"category,omitempty"`
	MaxScore      int           `json:"max_score"`
	TimeLimitMs   int           `json:"time_limit_ms"`
	MemoryLimitMb int           `json:"memory_limit_mb"`
	FunctionSig   string        `json:"function_sig,omitempty"`
	FunctionName  string        `json:"function_name,omitempty"`
	TestCases     []TestCaseDef `json:"test_cases"`
}

// TestResult represents the result of a single test case execution.
type TestResult struct {
	TestCaseID      int         `json:"test_case_id"`
	Status          string      `json:"status"`
	ExecutionTimeMs int         `json:"execution_time_ms"`
	MemoryUsedKb    int         `json:"memory_used_kb"`
	CPUTimeMs       int         `json:"cpu_time_ms,omitempty"`
	DiskIOKB        int         `json:"disk_io_kb,omitempty"`
	NetworkIOKB     int         `json:"network_io_kb,omitempty"`
	OutputSize      int         `json:"output_size,omitempty"`
	ActualOutput    interface{} `json:"actual_output,omitempty"`
	ExpectedOutput  interface{} `json:"expected_output,omitempty"`
	ErrorMessage    string      `json:"error_message,omitempty"`
}

// HarnessMode represents the execution mode of the test harness.
type HarnessMode string

const (
	HarnessModeNormal    HarnessMode = "normal"
	HarnessModeBenchmark HarnessMode = "benchmark"
	HarnessModeFuzz      HarnessMode = "fuzz"
	HarnessModeCustom    HarnessMode = "custom"
)

// HarnessConfig holds configuration for the test harness generator.
type HarnessConfig struct {
	// Mode is the execution mode (normal, benchmark, fuzz, custom).
	Mode HarnessMode `json:"mode"`
	// BenchmarkIterations is the number of iterations for benchmark mode.
	BenchmarkIterations int `json:"benchmark_iterations"`
	// FuzzIterations is the number of random test cases for fuzz mode.
	FuzzIterations int `json:"fuzz_iterations"`
	// CustomRunner is a user-defined test function to use instead of the default.
	CustomRunner string `json:"custom_runner,omitempty"`
	// EnableMemoryTracking enables detailed memory usage tracking in the harness.
	EnableMemoryTracking bool `json:"enable_memory_tracking"`
	// EnableCPUProfiling enables CPU profiling data collection.
	EnableCPUProfiling bool `json:"enable_cpu_profiling"`
}

// DefaultHarnessConfig returns the default harness configuration.
func DefaultHarnessConfig() HarnessConfig {
	return HarnessConfig{
		Mode:                 HarnessModeNormal,
		BenchmarkIterations:  100,
		FuzzIterations:       50,
		EnableMemoryTracking: false,
		EnableCPUProfiling:   false,
	}
}

// generateFunctionHarness creates a complete Go program that wraps user function code
// with test cases for function-based problems.
//
// The generated harness:
// 1. Defines the user function
// 2. Iterates over test cases, calling the function with parameters
// 3. Compares results using reflect.DeepEqual
// 4. Outputs JSON results per test case with per-test-case metrics
func generateFunctionHarness(userCode string, problem ProblemDefinition, harnessCfg HarnessConfig) string {
	var b strings.Builder

	b.WriteString("package main\n\n")
	b.WriteString("import (\n")
	b.WriteString("\t\"encoding/json\"\n")
	b.WriteString("\t\"fmt\"\n")
	b.WriteString("\t\"reflect\"\n")

	if harnessCfg.EnableMemoryTracking {
		b.WriteString("\t\"runtime\"\n")
	}
	if harnessCfg.EnableCPUProfiling {
		b.WriteString("\t\"runtime/pprof\"\n")
	}
	if harnessCfg.Mode == HarnessModeBenchmark {
		b.WriteString("\t\"time\"\n")
	}
	if harnessCfg.Mode == HarnessModeFuzz {
		b.WriteString("\t\"math/rand\"\n")
	}
	b.WriteString(")\n\n")

	b.WriteString("// === USER CODE ===\n")
	b.WriteString(userCode)
	b.WriteString("\n\n")
	b.WriteString("// === TEST HARNESS ===\n\n")

	// Test case structure
	b.WriteString("type TestCase struct {\n")
	b.WriteString("\tID       int             `json:\"id\"`\n")
	b.WriteString("\tInput    json.RawMessage `json:\"input\"`\n")
	b.WriteString("\tExpected json.RawMessage `json:\"expected\"`\n")
	b.WriteString("}\n\n")

	b.WriteString("type TestResult struct {\n")
	b.WriteString("\tID             int    `json:\"id\"`\n")
	b.WriteString("\tPassed         bool   `json:\"passed\"`\n")
	b.WriteString("\tError          string `json:\"error,omitempty\"`\n")
	b.WriteString("\tExecutionTimeMs int   `json:\"execution_time_ms\"`\n")
	b.WriteString("\tMemoryUsedKb   int    `json:\"memory_used_kb\"`\n")
	if harnessCfg.EnableCPUProfiling {
		b.WriteString("\tCPUTimeMs      int    `json:\"cpu_time_ms\"`\n")
	}
	b.WriteString("}\n\n")

	// Test cases data
	b.WriteString("var testCases = []TestCase{\n")
	for _, tc := range problem.TestCases {
		inputJSON, _ := json.Marshal(tc.Input)
		expectedJSON, _ := json.Marshal(tc.Expected)
		b.WriteString(fmt.Sprintf("\t{ID: %d, Input: %s, Expected: %s},\n",
			tc.ID, string(inputJSON), string(expectedJSON)))
	}
	b.WriteString("}\n\n")

	// Main function based on mode
	switch harnessCfg.Mode {
	case HarnessModeBenchmark:
		generateBenchmarkHarness(&b, problem, harnessCfg)
	case HarnessModeFuzz:
		generateFuzzHarness(&b, problem, harnessCfg)
	case HarnessModeCustom:
		if harnessCfg.CustomRunner != "" {
			b.WriteString("// === USER-DEFINED TEST RUNNER ===\n")
			b.WriteString(harnessCfg.CustomRunner)
			b.WriteString("\n\n")
		}
		generateCustomHarness(&b, problem, harnessCfg)
	default:
		generateNormalHarness(&b, problem, harnessCfg)
	}

	return b.String()
}

// generateNormalHarness generates the standard test harness with per-test-case metrics.
func generateNormalHarness(b *strings.Builder, problem ProblemDefinition, cfg HarnessConfig) {
	b.WriteString("func main() {\n")
	b.WriteString("\tresults := make([]TestResult, 0, len(testCases))\n\n")

	funcName := problem.FunctionName
	if funcName == "" {
		funcName = "solution"
	}

	b.WriteString("\tfor _, tc := range testCases {\n")
	b.WriteString("\t\tvar result TestResult\n")
	b.WriteString("\t\tresult.ID = tc.ID\n")
	b.WriteString("\t\tresult.Passed = false\n\n")

	// Memory tracking before
	if cfg.EnableMemoryTracking {
		b.WriteString("\t\tvar memStatsBefore, memStatsAfter runtime.MemStats\n")
		b.WriteString("\t\truntime.ReadMemStats(&memStatsBefore)\n")
	}

	// CPU profiling start
	if cfg.EnableCPUProfiling {
		b.WriteString("\t\t// CPU profiling start\n")
		b.WriteString("\t\tcpuProfFile := fmt.Sprintf(\"/tmp/cpu_profile_%d.pprof\", tc.ID)\n")
		b.WriteString("\t\tf, _ := pprof.StartCPUProfile(nil)\n")
		b.WriteString("\t\tif f != nil { f.Close() }\n")
	}

	b.WriteString("\t\tstart := time.Now()\n\n")

	// Parse input based on function signature
	b.WriteString("\t\tinputStr := string(tc.Input)\n")
	b.WriteString("\t\texpectedStr := string(tc.Expected)\n\n")

	b.WriteString(fmt.Sprintf("\t\t// Call user function: %s\n", funcName))
	b.WriteString("\t\tvar passed bool\n")
	b.WriteString("\t\tvar errMsg string\n\n")

	b.WriteString("\t\tfunc() {\n")
	b.WriteString("\t\t\tdefer func() {\n")
	b.WriteString("\t\t\t\tif r := recover(); r != nil {\n")
	b.WriteString("\t\t\t\t\terrMsg = fmt.Sprintf(\"runtime panic: %v\", r)\n")
	b.WriteString("\t\t\t\t\tpassed = false\n")
	b.WriteString("\t\t\t\t}\n")
	b.WriteString("\t\t\t}()\n\n")

	b.WriteString(fmt.Sprintf("\t\t\tresultVal := %s(inputStr)\n", funcName))
	b.WriteString("\t\t\tresultStr := fmt.Sprintf(\"%v\", resultVal)\n")
	b.WriteString("\t\t\tpassed = reflect.DeepEqual(resultStr, expectedStr)\n")
	b.WriteString("\t\t\tif !passed {\n")
	b.WriteString("\t\t\t\terrMsg = fmt.Sprintf(\"expected %s, got %s\", expectedStr, resultStr)\n")
	b.WriteString("\t\t\t}\n")
	b.WriteString("\t\t}()\n\n")

	b.WriteString("\t\telapsed := time.Since(start)\n")
	b.WriteString("\t\tresult.ExecutionTimeMs = int(elapsed.Milliseconds())\n\n")

	// Memory tracking after
	if cfg.EnableMemoryTracking {
		b.WriteString("\t\truntime.ReadMemStats(&memStatsAfter)\n")
		b.WriteString("\t\tresult.MemoryUsedKb = int((memStatsAfter.Alloc - memStatsBefore.Alloc) / 1024)\n")
	} else {
		b.WriteString("\t\tresult.MemoryUsedKb = 0\n")
	}

	if cfg.EnableCPUProfiling {
		b.WriteString("\t\t// CPU profiling end placeholder\n")
	}

	b.WriteString("\t\tresult.Passed = passed\n")
	b.WriteString("\t\tresult.Error = errMsg\n")
	b.WriteString("\t\tresults = append(results, result)\n")
	b.WriteString("\t}\n\n")

	b.WriteString("\toutput, _ := json.Marshal(results)\n")
	b.WriteString("\tfmt.Println(string(output))\n")
	b.WriteString("}\n")
}

// generateBenchmarkHarness generates a benchmark harness that measures performance.
func generateBenchmarkHarness(b *strings.Builder, problem ProblemDefinition, cfg HarnessConfig) {
	funcName := problem.FunctionName
	if funcName == "" {
		funcName = "solution"
	}

	// Guard against zero iterations
	iterations := cfg.BenchmarkIterations
	if iterations < 1 {
		iterations = 1
	}

	b.WriteString("func main() {\n")
	b.WriteString("	results := make([]TestResult, 0, len(testCases))\n\n")

	b.WriteString(fmt.Sprintf("	iterations := %d\n", iterations))
	b.WriteString("\tfmt.Printf(\"Benchmark: %d iterations per test case\\n\", iterations)\n\n")

	b.WriteString("\tfor _, tc := range testCases {\n")
	b.WriteString("\t\tvar result TestResult\n")
	b.WriteString("\t\tresult.ID = tc.ID\n\n")

	b.WriteString("\t\tinputStr := string(tc.Input)\n\n")

	b.WriteString("\t\t// Warmup\n")
	b.WriteString("\t\tfor i := 0; i < 10; i++ {\n")
	b.WriteString(fmt.Sprintf("\t\t\t%s(inputStr)\n", funcName))
	b.WriteString("\t\t}\n\n")

	b.WriteString("\t\t// Benchmark\n")
	b.WriteString("\t\tstart := time.Now()\n")
	b.WriteString("\t\tfor i := 0; i < iterations; i++ {\n")
	b.WriteString(fmt.Sprintf("\t\t\t%s(inputStr)\n", funcName))
	b.WriteString("\t\t}\n")
	b.WriteString("\t\telapsed := time.Since(start)\n\n")

	b.WriteString("\t\tresult.ExecutionTimeMs = int(elapsed.Milliseconds() / int64(iterations))\n")
	b.WriteString("\t\tresult.Passed = true // Benchmark mode doesn't check correctness\n")
	b.WriteString("\t\tresults = append(results, result)\n")
	b.WriteString("\t}\n\n")

	b.WriteString("\toutput, _ := json.Marshal(results)\n")
	b.WriteString("\tfmt.Println(string(output))\n")
	b.WriteString("}\n")
}

// generateFuzzHarness generates a fuzz testing harness with random test case generation.
func generateFuzzHarness(b *strings.Builder, problem ProblemDefinition, cfg HarnessConfig) {
	funcName := problem.FunctionName
	if funcName == "" {
		funcName = "solution"
	}

	b.WriteString("func main() {\n")
	b.WriteString("\trand.Seed(time.Now().UnixNano())\n")
	b.WriteString("\tresults := make([]TestResult, 0)\n\n")

	b.WriteString(fmt.Sprintf("\tfuzzCount := %d\n", cfg.FuzzIterations))
	b.WriteString("\tfmt.Printf(\"Fuzz: generating %d random test cases\\n\", fuzzCount)\n\n")

	b.WriteString("\tfor i := 0; i < fuzzCount; i++ {\n")
	b.WriteString("\t\tvar result TestResult\n")
	b.WriteString("\t\tresult.ID = i + 1000 // Offset from original test cases\n\n")

	b.WriteString("\t\t// Generate random input based on function signature\n")
	if problem.FunctionSig != "" {
		b.WriteString(fmt.Sprintf("\t\t// Function signature: %s\n", problem.FunctionSig))
	}

	// Generate random numeric inputs as the default fuzz strategy
	b.WriteString("\t\trandInput := fmt.Sprintf(\"%d\", rand.Intn(10000))\n\n")

	b.WriteString("\t\tstart := time.Now()\n\n")
	b.WriteString("\t\tvar errMsg string\n")
	b.WriteString("\t\tfunc() {\n")
	b.WriteString("\t\t\tdefer func() {\n")
	b.WriteString("\t\t\t\tif r := recover(); r != nil {\n")
	b.WriteString("\t\t\t\t\terrMsg = fmt.Sprintf(\"fuzz panic: %v\", r)\n")
	b.WriteString("\t\t\t\t}\n")
	b.WriteString("\t\t\t}()\n")
	b.WriteString(fmt.Sprintf("\t\t\t%s(randInput)\n", funcName))
	b.WriteString("\t\t}()\n\n")

	b.WriteString("\t\telapsed := time.Since(start)\n")
	b.WriteString("\t\tresult.ExecutionTimeMs = int(elapsed.Milliseconds())\n")
	b.WriteString("\t\tresult.Passed = errMsg == \"\"\n")
	b.WriteString("\t\tresult.Error = errMsg\n")
	b.WriteString("\t\tresults = append(results, result)\n")
	b.WriteString("\t}\n\n")

	// Also run existing test cases for correctness check
	b.WriteString("\t// Also run original test cases\n")
	b.WriteString("\tfor _, tc := range testCases {\n")
	b.WriteString("\t\tvar result TestResult\n")
	b.WriteString("\t\tresult.ID = tc.ID\n")
	b.WriteString("\t\tinputStr := string(tc.Input)\n")
	b.WriteString("\t\texpectedStr := string(tc.Expected)\n\n")
	b.WriteString("\t\tstart := time.Now()\n")
	b.WriteString("\t\tvar passed bool\n")
	b.WriteString("\t\tvar errMsg string\n\n")
	b.WriteString("\t\tfunc() {\n")
	b.WriteString("\t\t\tdefer func() {\n")
	b.WriteString("\t\t\t\tif r := recover(); r != nil {\n")
	b.WriteString("\t\t\t\t\terrMsg = fmt.Sprintf(\"runtime panic: %v\", r)\n")
	b.WriteString("\t\t\t\t}\n")
	b.WriteString("\t\t\t}()\n")
	b.WriteString(fmt.Sprintf("\t\t\tresultVal := %s(inputStr)\n", funcName))
	b.WriteString("\t\t\tresultStr := fmt.Sprintf(\"%v\", resultVal)\n")
	b.WriteString("\t\t\tpassed = reflect.DeepEqual(resultStr, expectedStr)\n")
	b.WriteString("\t\t\tif !passed {\n")
	b.WriteString("\t\t\t\terrMsg = fmt.Sprintf(\"expected %s, got %s\", expectedStr, resultStr)\n")
	b.WriteString("\t\t\t}\n")
	b.WriteString("\t\t}()\n\n")
	b.WriteString("\t\tresult.ExecutionTimeMs = int(time.Since(start).Milliseconds())\n")
	b.WriteString("\t\tresult.Passed = passed\n")
	b.WriteString("\t\tresult.Error = errMsg\n")
	b.WriteString("\t\tresults = append(results, result)\n")
	b.WriteString("\t}\n\n")

	b.WriteString("\toutput, _ := json.Marshal(results)\n")
	b.WriteString("\tfmt.Println(string(output))\n")
	b.WriteString("}\n")
}

// generateCustomHarness generates a harness that uses a user-defined test runner.
func generateCustomHarness(b *strings.Builder, problem ProblemDefinition, cfg HarnessConfig) {
	funcName := problem.FunctionName
	if funcName == "" {
		funcName = "solution"
	}

	b.WriteString("func main() {\n")
	b.WriteString("\tresults := make([]TestResult, 0, len(testCases))\n\n")

	b.WriteString("\tfor _, tc := range testCases {\n")
	b.WriteString("\t\tvar result TestResult\n")
	b.WriteString("\t\tresult.ID = tc.ID\n")
	b.WriteString("\t\tresult.Passed = false\n\n")

	b.WriteString("\t\tinputStr := string(tc.Input)\n")
	b.WriteString("\t\texpectedStr := string(tc.Expected)\n\n")

	b.WriteString("\t\tstart := time.Now()\n\n")

	b.WriteString("\t\t// Call user-defined test runner\n")
	b.WriteString("\t\tpassed, errMsg := customTestRunner(inputStr, expectedStr)\n\n")

	b.WriteString("\t\telapsed := time.Since(start)\n")
	b.WriteString("\t\tresult.ExecutionTimeMs = int(elapsed.Milliseconds())\n")
	b.WriteString("\t\tresult.Passed = passed\n")
	b.WriteString("\t\tresult.Error = errMsg\n")
	b.WriteString("\t\tresults = append(results, result)\n")
	b.WriteString("\t}\n\n")

	b.WriteString("\toutput, _ := json.Marshal(results)\n")
	b.WriteString("\tfmt.Println(string(output))\n")
	b.WriteString("}\n\n")

	// Add default custom test runner if none provided
	if cfg.CustomRunner == "" {
		b.WriteString("// customTestRunner is a user-defined test function.\n")
		b.WriteString("// Override this to implement custom test logic.\n")
		b.WriteString("func customTestRunner(input, expected string) (bool, string) {\n")
		b.WriteString(fmt.Sprintf("\tresult := %s(input)\n", funcName))
		b.WriteString("\tresultStr := fmt.Sprintf(\"%v\", result)\n")
		b.WriteString("\tpassed := reflect.DeepEqual(resultStr, expected)\n")
		b.WriteString("\tif !passed {\n")
		b.WriteString("\t\treturn false, fmt.Sprintf(\"expected %s, got %s\", expected, resultStr)\n")
		b.WriteString("\t}\n")
		b.WriteString("\treturn true, \"\"\n")
		b.WriteString("}\n")
	}
}

// generateMainHarness creates a complete Go program for main-based problems.
// The user code is expected to read from stdin and write to stdout.
// The harness feeds input and captures output for comparison.
func generateMainHarness(userCode string, problem ProblemDefinition, cfg HarnessConfig) string {
	var b strings.Builder

	b.WriteString("package main\n\n")
	b.WriteString("import (\n")
	b.WriteString("\t\"bytes\"\n")
	b.WriteString("\t\"encoding/json\"\n")
	b.WriteString("\t\"fmt\"\n")
	b.WriteString("\t\"io\"\n")
	b.WriteString("\t\"os\"\n")
	b.WriteString("\t\"os/exec\"\n")
	b.WriteString("\t\"strings\"\n")
	b.WriteString("\t\"time\"\n")
	if cfg.EnableMemoryTracking {
		b.WriteString("\t\"runtime\"\n")
	}
	b.WriteString(")\n\n")
	b.WriteString("// === USER CODE ===\n")
	b.WriteString(userCode)
	b.WriteString("\n\n")
	b.WriteString("// === TEST HARNESS ===\n\n")

	// Test case structure
	b.WriteString("type TestCase struct {\n")
	b.WriteString("\tID       int    `json:\"id\"`\n")
	b.WriteString("\tInput    string `json:\"input\"`\n")
	b.WriteString("\tExpected string `json:\"expected\"`\n")
	b.WriteString("}\n\n")

	b.WriteString("type TestResult struct {\n")
	b.WriteString("\tID             int    `json:\"id\"`\n")
	b.WriteString("\tPassed         bool   `json:\"passed\"`\n")
	b.WriteString("\tError          string `json:\"error,omitempty\"`\n")
	b.WriteString("\tExecutionTimeMs int   `json:\"execution_time_ms\"`\n")
	b.WriteString("\tMemoryUsedKb   int    `json:\"memory_used_kb\"`\n")
	b.WriteString("}\n\n")

	// Test cases data
	b.WriteString("var testCases = []TestCase{\n")
	for _, tc := range problem.TestCases {
		inputStr := fmt.Sprintf("%v", tc.Input)
		expectedStr := fmt.Sprintf("%v", tc.Expected)
		// Escape special characters
		inputStr = strings.ReplaceAll(inputStr, "\\", "\\\\")
		inputStr = strings.ReplaceAll(inputStr, "\"", "\\\"")
		inputStr = strings.ReplaceAll(inputStr, "\n", "\\n")
		inputStr = strings.ReplaceAll(inputStr, "\t", "\\t")
		expectedStr = strings.ReplaceAll(expectedStr, "\\", "\\\\")
		expectedStr = strings.ReplaceAll(expectedStr, "\"", "\\\"")
		expectedStr = strings.ReplaceAll(expectedStr, "\n", "\\n")
		expectedStr = strings.ReplaceAll(expectedStr, "\t", "\\t")
		b.WriteString(fmt.Sprintf("\t{ID: %d, Input: \"%s\", Expected: \"%s\"},\n",
			tc.ID, inputStr, expectedStr))
	}
	b.WriteString("}\n\n")

	// Main function
	b.WriteString("func main() {\n")
	b.WriteString("\tresults := make([]TestResult, 0, len(testCases))\n\n")
	b.WriteString("\tfor _, tc := range testCases {\n")
	b.WriteString("\t\tvar result TestResult\n")
	b.WriteString("\t\tresult.ID = tc.ID\n\n")

	if cfg.EnableMemoryTracking {
		b.WriteString("\t\tvar memStatsBefore runtime.MemStats\n")
		b.WriteString("\t\truntime.ReadMemStats(&memStatsBefore)\n")
	}

	b.WriteString("\t\tstart := time.Now()\n\n")

	// Execute the main program with input
	b.WriteString("\t\tcmd := exec.Command(\"go\", \"run\", \"main.go\")\n")
	b.WriteString("\t\tcmd.Stdin = strings.NewReader(tc.Input)\n\n")

	b.WriteString("\t\tvar stdout, stderr bytes.Buffer\n")
	b.WriteString("\t\tcmd.Stdout = &stdout\n")
	b.WriteString("\t\tcmd.Stderr = &stderr\n\n")

	b.WriteString("\t\tdone := make(chan error, 1)\n")
	b.WriteString("\t\tgo func() {\n")
	b.WriteString("\t\t\tdone <- cmd.Run()\n")
	b.WriteString("\t\t}()\n\n")

	// Per-test-case timeout
	timeoutMs := 5000
	if problem.TimeLimitMs > 0 {
		timeoutMs = problem.TimeLimitMs
	}
	b.WriteString(fmt.Sprintf("\t\tselect {\n"))
	b.WriteString(fmt.Sprintf("\t\tcase <-time.After(%d * time.Millisecond):\n", timeoutMs))
	b.WriteString("\t\t\tcmd.Process.Kill()\n")
	b.WriteString("\t\t\tresult.Error = \"execution timeout\"\n")
	b.WriteString("\t\tcase err := <-done:\n")
	b.WriteString("\t\t\telapsed := time.Since(start)\n")
	b.WriteString("\t\t\tresult.ExecutionTimeMs = int(elapsed.Milliseconds())\n")

	if cfg.EnableMemoryTracking {
		b.WriteString("\t\t\tvar memStatsAfter runtime.MemStats\n")
		b.WriteString("\t\t\truntime.ReadMemStats(&memStatsAfter)\n")
		b.WriteString("\t\t\tresult.MemoryUsedKb = int((memStatsAfter.Alloc - memStatsBefore.Alloc) / 1024)\n")
	}

	b.WriteString("\t\t\tif err != nil {\n")
	b.WriteString("\t\t\t\tresult.Error = stderr.String()\n")
	b.WriteString("\t\t\t} else {\n")
	b.WriteString("\t\t\t\toutput := strings.TrimSpace(stdout.String())\n")
	b.WriteString("\t\t\t\texpected := strings.TrimSpace(tc.Expected)\n")
	b.WriteString("\t\t\t\tresult.Passed = output == expected\n")
	b.WriteString("\t\t\t\tif !result.Passed {\n")
	b.WriteString("\t\t\t\t\tresult.Error = fmt.Sprintf(\"expected '%s', got '%s'\", expected, output)\n")
	b.WriteString("\t\t\t\t}\n")
	b.WriteString("\t\t\t}\n")
	b.WriteString("\t\t}\n\n")

	b.WriteString("\t\tresults = append(results, result)\n")
	b.WriteString("\t}\n\n")

	b.WriteString("\toutput, _ := json.Marshal(results)\n")
	b.WriteString("\tfmt.Println(string(output))\n")
	b.WriteString("}\n")

	return b.String()
}

// parseTestResults parses the JSON output from the test harness execution.
// It takes the raw stdout output and the test case definitions, returning
// a structured slice of TestResult with metrics.
func parseTestResults(output string, testCases []TestCaseDef) []TestResult {
	// Try to parse as JSON array first
	var rawResults []struct {
		ID             int    `json:"id"`
		Passed         bool   `json:"passed"`
		Error          string `json:"error,omitempty"`
		ExecutionTimeMs int   `json:"execution_time_ms"`
		MemoryUsedKb   int    `json:"memory_used_kb"`
		CPUTimeMs      int    `json:"cpu_time_ms,omitempty"`
		DiskIOKB       int    `json:"disk_io_kb,omitempty"`
		NetworkIOKB    int    `json:"network_io_kb,omitempty"`
		OutputSize     int    `json:"output_size,omitempty"`
	}

	output = strings.TrimSpace(output)

	// Try JSON array format
	if err := json.Unmarshal([]byte(output), &rawResults); err == nil {
		results := make([]TestResult, 0, len(rawResults))
		for _, raw := range rawResults {
			status := "passed"
			if !raw.Passed {
				status = "failed"
			}
			results = append(results, TestResult{
				TestCaseID:      raw.ID,
				Status:          status,
				ExecutionTimeMs: raw.ExecutionTimeMs,
				MemoryUsedKb:    raw.MemoryUsedKb,
				CPUTimeMs:       raw.CPUTimeMs,
				DiskIOKB:        raw.DiskIOKB,
				NetworkIOKB:     raw.NetworkIOKB,
				OutputSize:      raw.OutputSize,
				ErrorMessage:    raw.Error,
			})
		}
		return results
	}

	// Fallback: try line-by-line format (one JSON object per line)
	results := make([]TestResult, 0)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var raw struct {
			ID             int    `json:"id"`
			Passed         bool   `json:"passed"`
			Error          string `json:"error,omitempty"`
			ExecutionTimeMs int   `json:"execution_time_ms,omitempty"`
			MemoryUsedKb   int    `json:"memory_used_kb,omitempty"`
		}
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue
		}

		status := "passed"
		if !raw.Passed {
			status = "failed"
		}
		results = append(results, TestResult{
			TestCaseID:      raw.ID,
			Status:          status,
			ExecutionTimeMs: raw.ExecutionTimeMs,
			MemoryUsedKb:    raw.MemoryUsedKb,
			ErrorMessage:    raw.Error,
		})
	}

	// If no results were parsed, create failure entries for all test cases
	if len(results) == 0 {
		for _, tc := range testCases {
			results = append(results, TestResult{
				TestCaseID:   tc.ID,
				Status:       "error",
				ErrorMessage: "no test results could be parsed from output",
			})
		}
	}

	return results
}

// compareResults compares expected and actual values using deep equality.
// It handles type coercion for numeric types and string normalization.
func compareResults(expected, actual interface{}) bool {
	// Direct deep equality check
	if reflect.DeepEqual(expected, actual) {
		return true
	}

	// Try string comparison (handles type mismatches)
	expectedStr := fmt.Sprintf("%v", expected)
	actualStr := fmt.Sprintf("%v", actual)

	// Normalize whitespace
	expectedStr = strings.TrimSpace(expectedStr)
	actualStr = strings.TrimSpace(actualStr)

	if expectedStr == actualStr {
		return true
	}

	// Try numeric comparison
	expectedNum, expectedErr := toFloat64(expected)
	actualNum, actualErr := toFloat64(actual)
	if expectedErr == nil && actualErr == nil {
		return expectedNum == actualNum
	}

	return false
}

// toFloat64 attempts to convert a value to float64 for numeric comparison.
func toFloat64(v interface{}) (float64, error) {
	switch n := v.(type) {
	case float64:
		return n, nil
	case float32:
		return float64(n), nil
	case int:
		return float64(n), nil
	case int32:
		return float64(n), nil
	case int64:
		return float64(n), nil
	case string:
		return strconv.ParseFloat(n, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}

// calculateScore computes the final score based on test case results.
// It returns a score between 0 and the problem's max score.
func calculateScore(results []TestResult, testCases []TestCaseDef, maxScore int) int {
	if len(results) == 0 || len(testCases) == 0 {
		return 0
	}

	// Build weight map
	weightMap := make(map[int]int)
	totalWeight := 0
	for _, tc := range testCases {
		weight := tc.Weight
		if weight == 0 {
			weight = 1
		}
		weightMap[tc.ID] = weight
		totalWeight += weight
	}

	// Calculate weighted score
	passedWeight := 0
	for _, result := range results {
		if result.Status == "passed" {
			if weight, ok := weightMap[result.TestCaseID]; ok {
				passedWeight += weight
			}
		}
	}

	if totalWeight == 0 {
		return 0
	}

	score := (passedWeight * maxScore) / totalWeight
	return score
}

// generateFuzzTestCase generates a random test case based on a seed.
func generateFuzzTestCase(seed int64) TestCaseDef {
	rng := rand.New(rand.NewSource(seed))
	return TestCaseDef{
		ID:       int(seed),
		Input:    fmt.Sprintf("%d", rng.Intn(10000)),
		Expected: fmt.Sprintf("%d", rng.Intn(10000)),
		Weight:   0,
	}
}

// SanitizeHarnessOutput removes harness-specific metadata from output,
// leaving only the test results JSON.
func SanitizeHarnessOutput(output string) string {
	lines := strings.Split(output, "\n")
	var resultLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip benchmark/fuzz metadata lines
		if strings.HasPrefix(line, "Benchmark:") || strings.HasPrefix(line, "Fuzz:") {
			continue
		}
		if line != "" {
			resultLines = append(resultLines, line)
		}
	}
	return strings.Join(resultLines, "\n")
}
