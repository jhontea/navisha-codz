package model

// ProblemType represents the type of problem (function-based or main-based).
type ProblemType string

const (
	ProblemTypeFunction ProblemType = "function"
	ProblemTypeMain     ProblemType = "main"
)

// Problem represents a coding challenge problem.
type Problem struct {
	ID                  string       `yaml:"id" json:"id"`
	Title               string       `yaml:"title" json:"title"`
	Type                ProblemType  `yaml:"type" json:"type"`
	Difficulty          string       `yaml:"difficulty" json:"difficulty"`
	Category            string       `yaml:"category" json:"category"`
	Tags                []string     `yaml:"tags" json:"tags"`
	Description         string       `yaml:"description" json:"description"`
	Examples            []Example    `yaml:"examples" json:"examples"`
	Hints               []Hint       `yaml:"hints" json:"hints"`
	Template            string       `yaml:"template" json:"template"`
	TestCases           []TestCase   `yaml:"test_cases" json:"test_cases"`
	FunctionName        string       `yaml:"function_name" json:"function_name,omitempty"`
	Parameters          []Parameter  `yaml:"parameters" json:"parameters,omitempty"`
	ReturnType          string       `yaml:"return_type" json:"return_type,omitempty"`
	Constraints         []string     `yaml:"constraints" json:"constraints"`
	TimeComplexityHint  string       `yaml:"time_complexity_hint" json:"time_complexity_hint"`
	SpaceComplexityHint string       `yaml:"space_complexity_hint" json:"space_complexity_hint"`
	Solution            *Solution    `yaml:"solution" json:"-"` // Hidden from API
}

// Example represents an input/output example for a problem.
type Example struct {
	Input       string `yaml:"input" json:"input"`
	Output      string `yaml:"output" json:"output"`
	Explanation string `yaml:"explanation" json:"explanation"`
}

// Hint represents a progressive hint for solving a problem.
type Hint struct {
	Level   int    `yaml:"level" json:"level"`
	Title   string `yaml:"title" json:"title"`
	Content string `yaml:"content" json:"content"`
}

// Parameter represents a function parameter for function-based problems.
type Parameter struct {
	Name        string `yaml:"name" json:"name"`
	Type        string `yaml:"type" json:"type"`
	Description string `yaml:"description" json:"description,omitempty"`
}

// TestCase represents a single test case for validation.
// For function-based: use Params ([]interface{}) and Expected (interface{})
// For main-based: use Input (string) and Expected (string)
type TestCase struct {
	Input       string      `yaml:"input" json:"input,omitempty"`
	Params      []any       `yaml:"params" json:"params,omitempty"`
	Expected    any         `yaml:"expected" json:"expected"`
	Description string      `yaml:"description" json:"description"`
}

// Solution is the reference solution (hidden from API responses).
type Solution struct {
	Code            string `yaml:"code" json:"-"`
	Approach        string `yaml:"approach" json:"-"`
	TimeComplexity  string `yaml:"time_complexity" json:"-"`
	SpaceComplexity string `yaml:"space_complexity" json:"-"`
}

// TestResult represents the result of a single test case execution.
type TestResult struct {
	Name           string      `json:"name"`
	Passed         bool        `json:"passed"`
	Expected       interface{} `json:"expected"`
	Actual         interface{} `json:"actual"`
	Error          string      `json:"error"`
	ExecutionTimeMs int64      `json:"execution_time_ms"`
}

// RunResponse represents the full response from a code execution.
type RunResponse struct {
	Success          bool         `json:"success"`
	CompilationError *string      `json:"compilation_error"`
	TestResults      []TestResult `json:"test_results"`
	PassedCount      int          `json:"passed_count"`
	TotalCount       int          `json:"total_count"`
	ExecutionTimeMs  int64        `json:"execution_time_ms"`
}

// ProblemSummary is a lightweight view of a problem (used in list endpoints).
type ProblemSummary struct {
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	Type       string   `json:"type"`
	Difficulty string   `json:"difficulty"`
	Category   string   `json:"category"`
	Tags       []string `json:"tags"`
}

// ValidationResult represents the result of a syntax validation.
type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []ValidationError `json:"errors"`
	Warnings []ValidationError `json:"warnings"`
}

// ValidationError represents a single validation error or warning.
type ValidationError struct {
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Message  string `json:"message"`
	Severity string `json:"severity"` // "error" or "warning"
}

// Meta is the standard metadata wrapper for API responses.
type Meta struct {
	RequestID string `json:"request_id"`
	Timestamp string `json:"timestamp"`
}

// APIResponse is the standard wrapper for successful API responses.
type APIResponse struct {
	Data interface{} `json:"data"`
	Meta Meta        `json:"meta"`
}

// APIError is the standard wrapper for error API responses.
type APIErrorResponse struct {
	Error *APIErrorDetail `json:"error"`
	Meta  Meta            `json:"meta"`
}

// APIErrorDetail contains error information.
type APIErrorDetail struct {
	Message  string        `json:"message"`
	Code     int           `json:"code"`
	Type     string        `json:"type"`
	Details []ErrorDetail `json:"details,omitempty"`
}

// ErrorDetail represents a field-level error detail.
type ErrorDetail struct {
	Field  string `json:"field"`
	Issue  string `json:"issue"`
}
