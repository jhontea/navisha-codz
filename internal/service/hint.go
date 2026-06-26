package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"coding-challange/internal/model"
)

// HintService provides progressive hint reveal with concurrency-safe tracking.
type HintService struct {
	mu      sync.RWMutex
	revealed map[string]int // problemID -> number of hints revealed so far
}

// NewHintService creates a new hint service.
func NewHintService() *HintService {
	return &HintService{
		revealed: make(map[string]int),
	}
}

// GetHints returns the next batch of hints for a problem based on how many
// have already been revealed. At most 2 hints are revealed per request.
func (s *HintService) GetHints(problem *model.Problem) []model.Hint {
	if problem == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	revealed := s.revealed[problem.ID]

	// Return all hints up to current reveal level + 2 (progressive reveal)
	nextLevel := revealed + 2
	end := nextLevel
	if end > len(problem.Hints) {
		end = len(problem.Hints)
	}

	if revealed >= len(problem.Hints) {
		// All hints revealed
		return problem.Hints
	}

	result := make([]model.Hint, 0, end-revealed)
	for i := revealed; i < end; i++ {
		result = append(result, problem.Hints[i])
	}

	s.revealed[problem.ID] = end

	// Sort by level
	sort.Slice(result, func(i, j int) bool {
		return result[i].Level < result[j].Level
	})

	return result
}

// GetFullHints returns all hints for a problem (used internally).
func (s *HintService) GetFullHints(problem *model.Problem) []model.Hint {
	if problem == nil {
		return nil
	}
	return problem.Hints
}

// Reset resets the reveal tracking for a problem.
func (s *HintService) Reset(problemID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.revealed, problemID)
}

// BuildTestHarness generates a Go test harness source file from test cases.
// For function-based problems, it generates a harness that calls the user function.
// For main-based problems, it generates a harness that feeds input via stdin.
func BuildTestHarness(userCode string, problem *model.Problem) string {
	if problem.Type == model.ProblemTypeFunction {
		return buildFunctionTestHarness(userCode, problem)
	}
	return buildMainTestHarness(userCode, problem)
}

// buildFunctionTestHarness generates a harness for function-based problems.
func buildFunctionTestHarness(userCode string, problem *model.Problem) string {
	var b strings.Builder

	b.WriteString("package main\n\n")
	b.WriteString("import (\n")
	b.WriteString("\t\"encoding/json\"\n")
	b.WriteString("\t\"fmt\"\n")
	b.WriteString("\t\"reflect\"\n")
	b.WriteString(")\n\n")
	b.WriteString(userCode)
	b.WriteString("\n\n")

	b.WriteString("type testCase struct {\n")
	b.WriteString("\tParams   []json.RawMessage `json:\"params\"`\n")
	b.WriteString("\tExpected json.RawMessage `json:\"expected\"`\n")
	b.WriteString("}\n\n")

	b.WriteString("func main() {\n")
	b.WriteString("\ttests := []testCase{\n")

	for _, tc := range problem.TestCases {
		paramsJSON, _ := json.Marshal(tc.Params)
		expectedJSON, _ := json.Marshal(tc.Expected)
		b.WriteString(fmt.Sprintf("\t\t{Params: json.RawMessage(%q), Expected: json.RawMessage(%q)},\n",
			string(paramsJSON), string(expectedJSON)))
	}

	b.WriteString("\t}\n\n")
	b.WriteString("\tfor i, test := range tests {\n")

	// Build function call with type assertions
	funcCall := problem.FunctionName + "("
	for i, param := range problem.Parameters {
		if i > 0 {
			funcCall += ", "
		}
		// Generate type assertion from json.RawMessage
		funcCall += fmt.Sprintf("parseParam[%s](test.params[%d])", param.Type, i)
	}
	funcCall += ")"

	if problem.ReturnType == "void" || problem.ReturnType == "" {
		b.WriteString(fmt.Sprintf("\t\t%s\n", funcCall))
		b.WriteString("\t\tfmt.Printf(\"test_%d: PASS\\n\", i+1)\n")
	} else {
		b.WriteString(fmt.Sprintf("\t\tresult := %s\n", funcCall))
		b.WriteString("\t\texpectedVal := interface{}(nil)\n")
		b.WriteString("\t\tif len(test.Expected) > 0 && string(test.Expected) != \"null\" {\n")
		b.WriteString("\t\t\tjson.Unmarshal(test.Expected, &expectedVal)\n")
		b.WriteString("\t\t}\n")
		b.WriteString("\t\tif reflect.DeepEqual(result, expectedVal) {\n")
		b.WriteString("\t\t\tfmt.Printf(\"test_%d: PASS\\n\", i+1)\n")
		b.WriteString("\t\t} else {\n")
		b.WriteString("\t\t\texp, _ := json.Marshal(expectedVal)\n")
		b.WriteString("\t\t\tact, _ := json.Marshal(result)\n")
		b.WriteString("\t\t\tfmt.Printf(\"test_%d: FAIL exp=%s got=%s\\n\", i+1, exp, act)\n")
		b.WriteString("\t\t}\n")
	}

	b.WriteString("\t}\n")
	b.WriteString("}\n\n")

	// Helper functions for parsing params
	for _, param := range problem.Parameters {
		b.WriteString(fmt.Sprintf("func parseParam%s(raw json.RawMessage) %s {\n", param.Type, param.Type))
		b.WriteString(fmt.Sprintf("\tvar v %s\n", param.Type))
		b.WriteString("\tjson.Unmarshal(raw, &v)\n")
		b.WriteString("\treturn v\n")
		b.WriteString("}\n\n")
	}

	return b.String()
}

// buildMainTestHarness generates a harness for main-based problems.
func buildMainTestHarness(userCode string, problem *model.Problem) string {
	// Rename user's main() to userMain() to avoid conflict
	userCode = strings.Replace(userCode, "func main()", "func userMain()", 1)

	var b strings.Builder
	b.WriteString("package main\n\n")
	b.WriteString("import (\n")
	b.WriteString("\t\"encoding/json\"\n")
	b.WriteString("\t\"fmt\"\n")
	b.WriteString("\t\"os\"\n")
	b.WriteString("\t\"os/exec\"\n")
	b.WriteString("\t\"strings\"\n")
	b.WriteString(")\n\n")
	b.WriteString("type testCase struct {\n")
	b.WriteString("\tInput    string `json:\"input\"`\n")
	b.WriteString("\tExpected string `json:\"expected\"`\n")
	b.WriteString("\tDesc     string `json:\"description\"`\n")
	b.WriteString("}\n\n")
	b.WriteString(userCode)
	b.WriteString("\n\n")
	b.WriteString("func main() {\n")
	b.WriteString("\tif len(os.Args) > 1 && os.Args[1] == \"--test-case\" {\n")
	b.WriteString("\t\tif len(os.Args) > 2 {\n")
	b.WriteString("\t\t\tos.Args = os.Args[2:]\n")
	b.WriteString("\t\t}\n")
	b.WriteString("\t\tuserMain()\n")
	b.WriteString("\t\treturn\n")
	b.WriteString("\t}\n\n")
	b.WriteString("\ttests := []testCase{\n")
	for _, tc := range problem.TestCases {
		b.WriteString(fmt.Sprintf("\t\t{Input: %q, Expected: %q, Desc: %q},\n", tc.Input, fmt.Sprintf("%v", tc.Expected), tc.Description))
	}
	b.WriteString("\t}\n\n")
	b.WriteString("\tresults := make([]map[string]interface{}, 0, len(tests))\n")
	b.WriteString("\tfor i, tc := range tests {\n")
	b.WriteString("\t\tactual := runTestForUser(tc.Input)\n")
	b.WriteString("\t\tresult := map[string]interface{}{\n")
	b.WriteString("\t\t\t\"name\":     fmt.Sprintf(\"test_%d\", i+1),\n")
	b.WriteString("\t\t\t\"passed\":   actual == tc.Expected,\n")
	b.WriteString("\t\t\t\"expected\": tc.Expected,\n")
	b.WriteString("\t\t\t\"actual\":   actual,\n")
	b.WriteString("\t\t}\n")
	b.WriteString("\t\tif actual != tc.Expected {\n")
	b.WriteString("\t\t\tresult[\"error\"] = \"output mismatch\"\n")
	b.WriteString("\t\t}\n")
	b.WriteString("\t\tresults = append(results, result)\n")
	b.WriteString("\t}\n\n")
	b.WriteString("\tjson.NewEncoder(os.Stdout).Encode(results)\n")
	b.WriteString("}\n\n")
	b.WriteString("func runTestForUser(input string) string {\n")
	b.WriteString("\tbinPath := os.Args[0]\n")
	b.WriteString("\tcmd := exec.Command(binPath, \"--test-case\", input)\n")
	b.WriteString("\tout, err := cmd.CombinedOutput()\n")
	b.WriteString("\tif err != nil {\n")
	b.WriteString("\t\treturn \"\"\n")
	b.WriteString("\t}\n")
	b.WriteString("\treturn strings.TrimSpace(string(out))\n")
	b.WriteString("}\n")
	return b.String()
}

// SanitizeProblemID validates that a problem ID contains only allowed characters.
func SanitizeProblemID(id string) error {
	if id == "" {
		return fmt.Errorf("problem ID is empty")
	}
	for _, r := range id {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return fmt.Errorf("invalid character %q in problem ID", string(r))
		}
	}
	return nil
}

// ValidateCodeSize checks that submitted code doesn't exceed the maximum size.
func ValidateCodeSize(code string) error {
	maxSize := 64 * 1024 // 64KB
	if len(code) > maxSize {
		return fmt.Errorf("code exceeds maximum size of %d bytes", maxSize)
	}
	return nil
}

// EnsureTempDir ensures a temporary directory exists and returns its path.
func EnsureTempDir() (string, error) {
	tmpDir, err := os.MkdirTemp("", "coding-challenge-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	return filepath.Clean(tmpDir), nil
}
