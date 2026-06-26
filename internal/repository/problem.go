package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"

	"coding-challange/internal/model"
)

// ProblemRepository provides thread-safe access to problems loaded from YAML files.
type ProblemRepository struct {
	mu       sync.RWMutex
	problems map[string]*model.Problem
}

// NewProblemRepository creates a new repository and loads all problems from the given directory.
func NewProblemRepository(dir string) (*ProblemRepository, error) {
	repo := &ProblemRepository{
		problems: make(map[string]*model.Problem),
	}

	if err := repo.Load(dir); err != nil {
		return nil, fmt.Errorf("failed to load problems: %w", err)
	}

	return repo, nil
}

// Load reads all YAML files from the directory (recursively) and caches them in memory.
func (r *ProblemRepository) Load(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("cannot read problems directory %q: %w", dir, err)
	}

	newProblems := make(map[string]*model.Problem)

	for _, entry := range entries {
		if entry.IsDir() {
			subDir := filepath.Join(dir, entry.Name())
			if err := r.loadFromDir(subDir, newProblems); err != nil {
				return err
			}
			continue
		}

		if !isYAMLFile(entry.Name()) {
			continue
		}

		fullPath := filepath.Join(dir, entry.Name())
		problem, err := r.loadFile(fullPath)
		if err != nil {
			return fmt.Errorf("error loading %q: %w", fullPath, err)
		}

		if _, exists := newProblems[problem.ID]; exists {
			return fmt.Errorf("duplicate problem ID %q", problem.ID)
		}

		newProblems[problem.ID] = problem
	}

	r.mu.Lock()
	r.problems = newProblems
	r.mu.Unlock()

	return nil
}

// loadFromDir loads all YAML files from a subdirectory.
func (r *ProblemRepository) loadFromDir(dir string, dest map[string]*model.Problem) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("cannot read subdirectory %q: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			subDir := filepath.Join(dir, entry.Name())
			if err := r.loadFromDir(subDir, dest); err != nil {
				return err
			}
			continue
		}

		if !isYAMLFile(entry.Name()) {
			continue
		}

		fullPath := filepath.Join(dir, entry.Name())
		problem, err := r.loadFile(fullPath)
		if err != nil {
			return fmt.Errorf("error loading %q: %w", fullPath, err)
		}

		if _, exists := dest[problem.ID]; exists {
			return fmt.Errorf("duplicate problem ID %q", problem.ID)
		}

		dest[problem.ID] = problem
	}

	return nil
}

// loadFile parses a single YAML file into a Problem.
func (r *ProblemRepository) loadFile(path string) (*model.Problem, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}

	var problem model.Problem
	if err := yaml.Unmarshal(data, &problem); err != nil {
		return nil, fmt.Errorf("cannot parse YAML: %w", err)
	}

	if problem.ID == "" {
		return nil, fmt.Errorf("missing required field 'id'")
	}

	// Set default type if not specified
	if problem.Type == "" {
		problem.Type = model.ProblemTypeMain
	}

	// Validate function-based problems
	if problem.Type == model.ProblemTypeFunction {
		if problem.FunctionName == "" {
			return nil, fmt.Errorf("function-based problem %q requires 'function_name'", problem.ID)
		}
		if len(problem.Parameters) == 0 {
			return nil, fmt.Errorf("function-based problem %q requires 'parameters'", problem.ID)
		}
		if problem.ReturnType == "" {
			return nil, fmt.Errorf("function-based problem %q requires 'return_type'", problem.ID)
		}
	}

	return &problem, nil
}

// GetAll returns all problems sorted by difficulty then title, optionally filtered.
// When tags is non-empty, only problems that have ALL the specified tags are returned.
func (r *ProblemRepository) GetAll(difficulty, category string, tags []string) []model.Problem {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]model.Problem, 0, len(r.problems))

	for _, p := range r.problems {
		if difficulty != "" && !strings.EqualFold(p.Difficulty, difficulty) {
			continue
		}
		if category != "" && !strings.EqualFold(p.Category, category) {
			continue
		}
		if len(tags) > 0 && !hasAllTags(p.Tags, tags) {
			continue
		}
		result = append(result, *p)
	}

	// Sort by difficulty then title
	difficultyOrder := map[string]int{"easy": 0, "medium": 1, "hard": 2}
	sort.Slice(result, func(i, j int) bool {
		di, oki := difficultyOrder[result[i].Difficulty]
		dj, okj := difficultyOrder[result[j].Difficulty]
		if !oki {
			di = 99
		}
		if !okj {
			dj = 99
		}
		if di != dj {
			return di < dj
		}
		return result[i].Title < result[j].Title
	})

	return result
}

// GetByID returns a problem by its ID, or nil if not found.
func (r *ProblemRepository) GetByID(id string) *model.Problem {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if p, ok := r.problems[id]; ok {
		return p
	}
	return nil
}

// Count returns the total number of loaded problems.
func (r *ProblemRepository) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.problems)
}

func isYAMLFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".yaml" || ext == ".yml"
}

// hasAllTags checks that problemTags contains every required tag (case-insensitive).
func hasAllTags(problemTags, requiredTags []string) bool {
	for _, required := range requiredTags {
		found := false
		for _, pt := range problemTags {
			if strings.EqualFold(pt, required) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
