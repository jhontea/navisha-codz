package service

import (
	"fmt"

	"coding-challange/internal/model"
	"coding-challange/internal/repository"
)

// ProblemService provides business logic for problems.
type ProblemService struct {
	repo *repository.ProblemRepository
}

// NewProblemService creates a new problem service.
func NewProblemService(repo *repository.ProblemRepository) *ProblemService {
	return &ProblemService{repo: repo}
}

// ListProblems returns a summary list of all problems (without solutions).
// Supports filtering by difficulty and category.
func (s *ProblemService) ListProblems(difficulty, category string) []model.ProblemSummary {
	problems := s.repo.GetAll(difficulty, category)
	summaries := make([]model.ProblemSummary, 0, len(problems))

	for _, p := range problems {
		summaries = append(summaries, model.ProblemSummary{
			ID:         p.ID,
			Title:      p.Title,
			Type:       string(p.Type),
			Difficulty: p.Difficulty,
			Category:   p.Category,
			Tags:       p.Tags,
		})
	}

	return summaries
}

// GetProblem returns the full problem details without the solution exposed.
// Test case expected values are retained for internal use (e.g., runner validation).
func (s *ProblemService) GetProblem(id string) (*model.Problem, error) {
	problem := s.repo.GetByID(id)
	if problem == nil {
		return nil, fmt.Errorf("problem %q not found", id)
	}

	// Ensure solution is never exposed via API
	problem.Solution = nil

	return problem, nil
}

// GetProblemForAPI returns problem details suitable for API response,
// stripping sensitive data like solution code and test case expected values.
func (s *ProblemService) GetProblemForAPI(id string) (*model.Problem, error) {
	problem, err := s.GetProblem(id)
	if err != nil {
		return nil, err
	}

	// Strip expected output from test cases to prevent cheating
	for i := range problem.TestCases {
		problem.TestCases[i].Expected = nil
		problem.TestCases[i].Params = nil
	}

	return problem, nil
}

// GetTemplate returns only the template and metadata for a problem (lightweight).
func (s *ProblemService) GetTemplate(id string) (map[string]interface{}, error) {
	problem, err := s.GetProblem(id)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"id":          problem.ID,
		"type":        string(problem.Type),
		"template":   problem.Template,
	}

	if problem.Type == model.ProblemTypeFunction {
		result["function_name"] = problem.FunctionName
		result["parameters"] = problem.Parameters
		result["return_type"] = problem.ReturnType
	}

	return result, nil
}

// ValidateCode validates user code syntax without execution.
func (s *ProblemService) ValidateCode(code string, problemID string) *model.ValidationResult {
	result := &model.ValidationResult{
		Valid:    true,
		Errors:   []model.ValidationError{},
		Warnings: []model.ValidationError{},
	}

	// Basic syntax check - ensure code is not empty
	if code == "" {
		result.Valid = false
		result.Errors = append(result.Errors, model.ValidationError{
			Line:     1,
			Column:   1,
			Message:  "code is empty",
			Severity: "error",
		})
		return result
	}

	// If problem_id is provided, validate against the function signature
	if problemID != "" {
		problem, err := s.GetProblem(problemID)
		if err == nil && problem.Type == model.ProblemTypeFunction {
			// Check if code contains the function declaration
			expectedFunc := "func " + problem.FunctionName + "("
			if !containsFunction(code, expectedFunc) {
				result.Warnings = append(result.Warnings, model.ValidationError{
					Line:     1,
					Column:   1,
					Message:  fmt.Sprintf("expected function signature: func %s(", problem.FunctionName),
					Severity: "warning",
				})
			}
		}
	}

	return result
}

// containsFunction checks if the code contains a function declaration.
func containsFunction(code, funcDecl string) bool {
	return len(code) > 0 && len(funcDecl) > 0 && contains(code, funcDecl)
}

// contains is a simple string containment check.
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
