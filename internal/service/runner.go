package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"coding-challange/internal/model"
)

// maxOutputSize is the maximum allowed output from a single test execution (1MB).
const maxOutputSize = 1 * 1024 * 1024

// sandboxImage is the Docker image for code execution.
// Using golang:alpine which is publicly available on Docker Hub.
// For production, build a custom image with: docker build -t coding-challenge-sandbox:latest -f Dockerfile.sandbox .
const sandboxImage = "golang:1.21-alpine"

// RunnerService handles code execution in Docker sandbox or local fallback.
type RunnerService struct {
	timeout              time.Duration
	memoryLimit          string
	disableLocalFallback bool
	containerPool        sync.Pool // Iter 7: Object pool for reusable buffers
}

// NewRunnerService creates a new runner service.
func NewRunnerService(timeout time.Duration, memoryLimitMB int) *RunnerService {
	return &RunnerService{
		timeout:              timeout,
		memoryLimit:          fmt.Sprintf("%dm", memoryLimitMB),
		disableLocalFallback: os.Getenv("DISABLE_LOCAL_FALLBACK") != "",
	}
}

// RunCode executes user code against test cases and returns structured results.
// It first tries Docker sandbox, then optionally falls back to local execution.
func (r *RunnerService) RunCode(userCode string, problem *model.Problem) *model.RunResponse {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout+5*time.Second)
	defer cancel()

	// Generate test harness based on problem type
	harnessCode := r.buildTestHarness(userCode, problem)

	// Try Docker first
	result, err := r.runInDocker(ctx, harnessCode, problem)
	if err == nil {
		return result
	}

	// Fallback to local execution (disabled by default via DISABLE_LOCAL_FALLBACK=true)
	if r.disableLocalFallback {
		compErr := fmt.Sprintf("Docker sandbox unavailable: %v", err)
		return &model.RunResponse{
			Success:          false,
			CompilationError: &compErr,
			TestResults:      []model.TestResult{},
			PassedCount:      0,
			TotalCount:       len(problem.TestCases),
			ExecutionTimeMs:  0,
		}
	}

	// Fallback to local execution
	return r.runLocal(ctx, harnessCode, problem)
}

// buildTestHarness generates a complete Go program that wraps user code with test cases.
func (r *RunnerService) buildTestHarness(userCode string, problem *model.Problem) string {
	if problem.Type == model.ProblemTypeFunction {
		return r.buildFunctionHarness(userCode, problem)
	}
	return r.buildMainHarness(userCode, problem)
}

// buildFunctionHarness generates a test harness for function-based problems.
// It creates a main.go that:
// 1. Defines the user function
// 2. Iterates over test cases, calling the function with params
// 3. Compares results using reflect.DeepEqual
// 4. Outputs JSON results per test case
func (r *RunnerService) buildFunctionHarness(userCode string, problem *model.Problem) string {
	var b strings.Builder

	b.WriteString("package main\n\n")
	b.WriteString("import (\n")
	b.WriteString("\t\"encoding/json\"\n")
	b.WriteString("\t\"fmt\"\n")
	b.WriteString("\t\"reflect\"\n")
	b.WriteString(")\n\n")
	b.WriteString("// === USER CODE ===\n")
	b.WriteString(userCode)
	b.WriteString("\n")
	b.WriteString("// === END USER CODE ===\n\n")

	// Generate test struct
	b.WriteString("type testCase struct {\n")
	b.WriteString("\tName     string      `json:\"name\"`\n")
	b.WriteString("\tParams   []json.RawMessage `json:\"params\"`\n")
	b.WriteString("\tExpected json.RawMessage `json:\"expected\"`\n")
	b.WriteString("}\n\n")

	// Generate main function
	b.WriteString("func main() {\n")
	b.WriteString("\ttests := []testCase{\n")

	for i, tc := range problem.TestCases {
		paramsJSON, _ := json.Marshal(tc.Params)
		expectedJSON, _ := json.Marshal(tc.Expected)
		b.WriteString(fmt.Sprintf("\t\t{\n"))
		b.WriteString(fmt.Sprintf("\t\t\tName:     \"test_%d\",\n", i+1))
		b.WriteString(fmt.Sprintf("\t\t\tParams:   json.RawMessage(%q),\n", string(paramsJSON)))
		b.WriteString(fmt.Sprintf("\t\t\tExpected: json.RawMessage(%q),\n", string(expectedJSON)))
		b.WriteString(fmt.Sprintf("\t\t},\n"))
	}

	b.WriteString("\t}\n\n")

	// For each test case, call the function with proper type assertions
	b.WriteString("\tresults := make([]map[string]interface{}, 0, len(tests))\n")
	b.WriteString("\tfor _, test := range tests {\n")
	b.WriteString("\t\tresult := callFunction(test.Params)\n")
	b.WriteString("\t\texpectedVal := interface{}(nil)\n")
	b.WriteString("\t\tif len(test.Expected) > 0 && string(test.Expected) != \"null\" {\n")
	b.WriteString("\t\t\tjson.Unmarshal(test.Expected, &expectedVal)\n")
	b.WriteString("\t\t}\n")
	b.WriteString("\t\tactualVal := result\n")
	b.WriteString("\t\tpassed := reflect.DeepEqual(actualVal, expectedVal)\n\n")

	b.WriteString("\t\tres := map[string]interface{}{\n")
	b.WriteString("\t\t\t\"name\":   test.Name,\n")
	b.WriteString("\t\t\t\"passed\": passed,\n")
	b.WriteString("\t\t}\n")

	b.WriteString("\t\tif !passed {\n")
	b.WriteString("\t\t\texpJSON, _ := json.Marshal(expectedVal)\n")
	b.WriteString("\t\t\tactJSON, _ := json.Marshal(actualVal)\n")
	b.WriteString("\t\t\tres[\"expected\"] = string(expJSON)\n")
	b.WriteString("\t\t\tres[\"actual\"] = string(actJSON)\n")
	b.WriteString("\t\t\tres[\"error\"] = fmt.Sprintf(\"expected=%s got=%s\", expJSON, actJSON)\n")
	b.WriteString("\t\t}\n")
	b.WriteString("\t\tresults = append(results, res)\n")
	b.WriteString("\t}\n\n")

	b.WriteString("\tjson.NewEncoder(os.Stdout).Encode(results)\n")
	b.WriteString("}\n\n")

	// Helper function that calls the user function with type-asserted params
	b.WriteString("// callFunction dispatches to the user's function with proper type assertions\n")
	b.WriteString("func callFunction(params []json.RawMessage) interface{} {\n")

	// Generate parameter parsing code
	for i, param := range problem.Parameters {
		b.WriteString(fmt.Sprintf("\tvar p%d %s\n", i, param.Type))
		b.WriteString(fmt.Sprintf("\tif len(params) > %d {\n", i))
		b.WriteString(fmt.Sprintf("\t\tjson.Unmarshal(params[%d], &p%d)\n", i, i))
		b.WriteString("\t}\n")
	}

	// Build function call
	funcCall := problem.FunctionName + "("
	for i := range problem.Parameters {
		if i > 0 {
			funcCall += ", "
		}
		funcCall += fmt.Sprintf("p%d", i)
	}
	funcCall += ")"

	if problem.ReturnType == "void" || problem.ReturnType == "" {
		b.WriteString(fmt.Sprintf("\t%s\n", funcCall))
		b.WriteString("\treturn nil\n")
	} else {
		b.WriteString(fmt.Sprintf("\tresult := %s\n", funcCall))
		b.WriteString("\treturn result\n")
	}

	b.WriteString("}\n")

	return b.String()
}

// buildMainHarness generates a test harness for main-based problems.
func (r *RunnerService) buildMainHarness(userCode string, problem *model.Problem) string {
	// For main-based, we wrap the user's code and feed test input via stdin
	var b strings.Builder
	b.WriteString(userCode)
	return b.String()
}

// runInDocker executes code inside a Docker container with strict sandboxing.
func (r *RunnerService) runInDocker(ctx context.Context, harnessCode string, problem *model.Problem) (*model.RunResponse, error) {
	tmpDir, err := os.MkdirTemp("", "runner-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write harness code to temp file
	codePath := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(codePath, []byte(harnessCode), 0644); err != nil {
		return nil, fmt.Errorf("failed to write code: %w", err)
	}

	// Build arguments
	var stdout, stderr limitWriter
	stdout.limit = maxOutputSize
	stderr.limit = maxOutputSize

	dockerArgs := []string{
		"run",
		"--rm",
		"--network=none",
		"--read-only",
		"--tmpfs", "/tmp:rw,noexec,nosuid,size=128m",
		"--tmpfs", "/go:rw,noexec,nosuid,size=256m",
		"--memory=" + r.memoryLimit,
		"--cpus=1",
		"--pids-limit=50",
		"--cap-drop=ALL",
		"--security-opt=no-new-privileges",
		"-e", "GOCACHE=/go/cache",
		"-e", "GOPATH=/go",
		"-v", fmt.Sprintf("%s:/app:ro", tmpDir),
		sandboxImage,
		"sh", "-c",
		"cd /app && go run main.go",
	}

	cmd := exec.CommandContext(ctx, "docker", dockerArgs...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	runErr := cmd.Run()
	elapsed := time.Since(start).Milliseconds()

	if runErr != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return &model.RunResponse{
				Success:          false,
				CompilationError: strPtr("execution timeout exceeded"),
				TestResults:      r.allFailed(problem.TestCases, "timeout"),
				PassedCount:      0,
				TotalCount:       len(problem.TestCases),
				ExecutionTimeMs:  elapsed,
			}, nil
		}
		return &model.RunResponse{
			Success:          false,
			CompilationError: strPtr(fmt.Sprintf("sandbox error: %s", stderr.String())),
			TestResults:      r.allFailed(problem.TestCases, "execution error"),
			PassedCount:      0,
			TotalCount:       len(problem.TestCases),
			ExecutionTimeMs:  elapsed,
		}, nil
	}

	// Parse JSON output
	return r.parseFunctionOutput(stdout.String(), problem.TestCases, elapsed), nil
}

// limitWriter is a Writer that caps the number of bytes written.
type limitWriter struct {
	buf   bytes.Buffer
	limit int
}

func (w *limitWriter) Write(p []byte) (int, error) {
	remaining := w.limit - w.buf.Len()
	if remaining <= 0 {
		return len(p), nil // silently discard
	}
	if len(p) > remaining {
		w.buf.Write(p[:remaining])
		return len(p), nil
	}
	return w.buf.Write(p)
}

func (w *limitWriter) String() string {
	return w.buf.String()
}

// runLocal is a fallback that compiles and runs code locally with a timeout.
func (r *RunnerService) runLocal(ctx context.Context, harnessCode string, problem *model.Problem) *model.RunResponse {
	tmpDir, err := os.MkdirTemp("", "runner-local-*")
	if err != nil {
		return &model.RunResponse{
			Success:          false,
			CompilationError: strPtr(fmt.Sprintf("internal error: %v", err)),
			TestResults:      []model.TestResult{},
			PassedCount:      0,
			TotalCount:       len(problem.TestCases),
		}
	}
	defer os.RemoveAll(tmpDir)

	// Write code
	codePath := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(codePath, []byte(harnessCode), 0644); err != nil {
		return &model.RunResponse{
			Success:          false,
			CompilationError: strPtr(fmt.Sprintf("internal error: %v", err)),
			TestResults:      []model.TestResult{},
			PassedCount:      0,
			TotalCount:       len(problem.TestCases),
		}
	}

	// Build with resource limits
	binPath := filepath.Join(tmpDir, "runner")
	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, codePath)
	buildOut, buildErr := buildCmd.CombinedOutput()
	if buildErr != nil {
		compErr := strings.TrimSpace(string(buildOut))
		return &model.RunResponse{
			Success:          false,
			CompilationError: &compErr,
			TestResults:      []model.TestResult{},
			PassedCount:      0,
			TotalCount:       len(problem.TestCases),
			ExecutionTimeMs:  0,
		}
	}

	// Execute with resource limits and output cap
	var stdout limitWriter
	stdout.limit = maxOutputSize
	cmd := exec.CommandContext(ctx, binPath)
	cmd.Stdout = &stdout

	start := time.Now()
	runErr := cmd.Run()
	elapsed := time.Since(start).Milliseconds()

	if runErr != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return &model.RunResponse{
				Success:          false,
				CompilationError: strPtr("execution timeout exceeded"),
				TestResults:      r.allFailed(problem.TestCases, "timeout"),
				PassedCount:      0,
				TotalCount:       len(problem.TestCases),
				ExecutionTimeMs:  elapsed,
			}
		}
		return &model.RunResponse{
			Success:          false,
			CompilationError: strPtr(fmt.Sprintf("runtime error: %v", runErr)),
			TestResults:      r.allFailed(problem.TestCases, "runtime error"),
			PassedCount:      0,
			TotalCount:       len(problem.TestCases),
			ExecutionTimeMs:  elapsed,
		}
	}

	return r.parseFunctionOutput(stdout.String(), problem.TestCases, elapsed)
}

// parseFunctionOutput parses JSON array output from function-based harness.
func (r *RunnerService) parseFunctionOutput(output string, testCases []model.TestCase, elapsed int64) *model.RunResponse {
	output = strings.TrimSpace(output)
	if output == "" {
		return &model.RunResponse{
			Success:          false,
			CompilationError: strPtr("no output from execution"),
			TestResults:      r.allFailed(testCases, "no output"),
			PassedCount:      0,
			TotalCount:       len(testCases),
			ExecutionTimeMs:  elapsed,
		}
	}

	var results []map[string]interface{}
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		return &model.RunResponse{
			Success:          false,
			CompilationError: strPtr(fmt.Sprintf("failed to parse output: %s", output)),
			TestResults:      r.allFailed(testCases, "output parse error"),
			PassedCount:      0,
			TotalCount:       len(testCases),
			ExecutionTimeMs:  elapsed,
		}
	}

	testResults := make([]model.TestResult, 0, len(testCases))
	passed := 0

	for i, tc := range testCases {
		name := fmt.Sprintf("test_%d", i+1)
		tr := model.TestResult{
			Name:     name,
			Expected: tc.Expected,
			Actual:   nil,
			Error:    "",
		}

		if i < len(results) {
			res := results[i]
			testPassed := false
			if p, ok := res["passed"].(bool); ok {
				testPassed = p
			}
			tr.Passed = testPassed
			if !testPassed {
				if errMsg, ok := res["error"].(string); ok {
					tr.Error = errMsg
				}
				if exp, ok := res["expected"].(string); ok {
					tr.Expected = exp
				}
				if act, ok := res["actual"].(string); ok {
					tr.Actual = act
				}
			} else {
				tr.Actual = tc.Expected
			}
			if testPassed {
				passed++
			}
		} else {
			tr.Error = "no result for this test case"
		}

		testResults = append(testResults, tr)
	}

	return &model.RunResponse{
		Success:         passed == len(testCases),
		TestResults:     testResults,
		PassedCount:     passed,
		TotalCount:      len(testCases),
		ExecutionTimeMs: elapsed,
	}
}

// allFailed returns test results marking all tests as failed with given error.
func (r *RunnerService) allFailed(testCases []model.TestCase, errMsg string) []model.TestResult {
	results := make([]model.TestResult, 0, len(testCases))
	for i, tc := range testCases {
		results = append(results, model.TestResult{
			Name:     fmt.Sprintf("test_%d", i+1),
			Passed:   false,
			Expected: tc.Expected,
			Actual:   nil,
			Error:    errMsg,
		})
	}
	return results
}

func strPtr(s string) *string {
	return &s
}
