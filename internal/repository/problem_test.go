package repository

import (
	"os"
	"path/filepath"
	"testing"

	"coding-challange/internal/model"
)

func TestNewProblemRepository_ValidDir(t *testing.T) {
	repo, err := NewProblemRepository("../../problems")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if repo == nil {
		t.Fatal("expected repo to be non-nil")
	}
	if repo.Count() != 20 {
		t.Errorf("expected 20 problems, got %d", repo.Count())
	}
}

func TestNewProblemRepository_NonExistentDir(t *testing.T) {
	_, err := NewProblemRepository("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Fatal("expected error for non-existent directory")
	}
}

func TestNewProblemRepository_EmptyDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-empty-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	repo, err := NewProblemRepository(tmpDir)
	if err != nil {
		t.Fatalf("expected no error for empty dir, got: %v", err)
	}
	if repo.Count() != 0 {
		t.Errorf("expected 0 problems, got %d", repo.Count())
	}
}

func TestNewProblemRepository_InvalidYAML(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-invalid-yaml-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	invalidYAML := `this is not: valid: yaml: [}`
	os.WriteFile(filepath.Join(tmpDir, "broken.yaml"), []byte(invalidYAML), 0644)

	_, err = NewProblemRepository(tmpDir)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestNewProblemRepository_MissingID(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-missing-id-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	yamlContent := `title: "No ID Problem"
difficulty: "easy"
category: "test"
tags: ["test"]
description: "test"
`
	os.WriteFile(filepath.Join(tmpDir, "no-id.yaml"), []byte(yamlContent), 0644)

	_, err = NewProblemRepository(tmpDir)
	if err == nil {
		t.Fatal("expected error for missing ID field")
	}
}

func TestNewProblemRepository_DuplicateID(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-dup-id-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	yaml1 := `id: "dup"
title: "First"
difficulty: "easy"
category: "test"
tags: ["test"]
description: "test"
`
	yaml2 := `id: "dup"
title: "Second"
difficulty: "easy"
category: "test"
tags: ["test"]
description: "test"
`
	os.WriteFile(filepath.Join(tmpDir, "first.yaml"), []byte(yaml1), 0644)
	os.WriteFile(filepath.Join(tmpDir, "second.yaml"), []byte(yaml2), 0644)

	_, err = NewProblemRepository(tmpDir)
	if err == nil {
		t.Fatal("expected error for duplicate IDs")
	}
}

func TestGetAll_NoFilter(t *testing.T) {
	repo, err := NewProblemRepository("../../problems")
	if err != nil {
		t.Fatal(err)
	}

	problems := repo.GetAll("", "", nil)
	if len(problems) != 20 {
			t.Errorf("expected 20 problems, got %d", len(problems))
	}

	// Check sorting: easy first, then medium
	if problems[0].Difficulty != "easy" {
		t.Errorf("expected first problem to be easy, got %s", problems[0].Difficulty)
	}
}

func TestGetAll_FilterByDifficulty(t *testing.T) {
	repo, err := NewProblemRepository("../../problems")
	if err != nil {
		t.Fatal(err)
	}

	easy := repo.GetAll("easy", "", nil)
	if len(easy) != 7 {
		t.Errorf("expected 7 easy problems, got %d", len(easy))
	}
	for _, p := range easy {
		if p.Difficulty != "easy" {
			t.Errorf("expected easy, got %s", p.Difficulty)
		}
	}

	medium := repo.GetAll("medium", "", nil)
	if len(medium) != 7 {
		t.Errorf("expected 7 medium problems, got %d", len(medium))
	}

	hard := repo.GetAll("hard", "", nil)
	if len(hard) != 6 {
		t.Errorf("expected 6 hard problems, got %d", len(hard))
	}
}

func TestGetAll_FilterByCategory(t *testing.T) {
	repo, err := NewProblemRepository("../../problems")
	if err != nil {
		t.Fatal(err)
	}

	arrayProbs := repo.GetAll("", "array", nil)
	if len(arrayProbs) != 9 {
		t.Errorf("expected 9 array problems, got %d", len(arrayProbs))
	}
}

func TestGetAll_FilterByTags(t *testing.T) {
	repo, err := NewProblemRepository("../../problems")
	if err != nil {
		t.Fatal(err)
	}

	// Filter by single tag
	backtracking := repo.GetAll("", "", []string{"backtracking"})
	if len(backtracking) != 3 {
		t.Errorf("expected 3 backtracking problems, got %d", len(backtracking))
	}

	// Filter by multiple tags (AND logic)
	dpString := repo.GetAll("", "", []string{"dp", "string"})
	if len(dpString) != 2 {
		t.Errorf("expected 2 problems with dp+string tags, got %d", len(dpString))
	}

	// Filter by tags + difficulty
	hardBacktracking := repo.GetAll("hard", "", []string{"backtracking"})
	if len(hardBacktracking) != 2 {
		t.Errorf("expected 2 hard backtracking problems, got %d", len(hardBacktracking))
	}

	// Case-insensitive tags
	caseInsensitive := repo.GetAll("", "", []string{"BACKTRACKING"})
	if len(caseInsensitive) != 3 {
		t.Errorf("expected 3 problems with case-insensitive tag, got %d", len(caseInsensitive))
	}

	// Non-existent tag
	noMatch := repo.GetAll("", "", []string{"nonexistent-tag"})
	if len(noMatch) != 0 {
		t.Errorf("expected 0 problems for non-existent tag, got %d", len(noMatch))
	}

	// Empty tag list should return all
	all := repo.GetAll("", "", []string{})
	if len(all) != 20 {
		t.Errorf("expected 20 problems with empty tag filter, got %d", len(all))
	}
}

func TestGetAll_CaseInsensitive(t *testing.T) {
	repo, err := NewProblemRepository("../../problems")
	if err != nil {
		t.Fatal(err)
	}

	upper := repo.GetAll("EASY", "", nil)
	lower := repo.GetAll("easy", "", nil)
	if len(upper) != len(lower) {
		t.Errorf("case insensitive filter failed: upper=%d, lower=%d", len(upper), len(lower))
	}
}

func TestGetAll_NonExistentFilter(t *testing.T) {
	repo, err := NewProblemRepository("../../problems")
	if err != nil {
		t.Fatal(err)
	}

	result := repo.GetAll("impossible", "", nil)
	if len(result) != 0 {
		t.Errorf("expected 0 problems, got %d", len(result))
	}
}

func TestGetByID_Existing(t *testing.T) {
	repo, err := NewProblemRepository("../../problems")
	if err != nil {
		t.Fatal(err)
	}

	problem := repo.GetByID("two-sum")
	if problem == nil {
		t.Fatal("expected to find two-sum problem")
	}
	if problem.ID != "two-sum" {
		t.Errorf("expected ID 'two-sum', got %q", problem.ID)
	}
	if problem.Title != "Two Sum" {
		t.Errorf("expected title 'Two Sum', got %q", problem.Title)
	}
}

func TestGetByID_NonExistent(t *testing.T) {
	repo, err := NewProblemRepository("../../problems")
	if err != nil {
		t.Fatal(err)
	}

	problem := repo.GetByID("nonexistent")
	if problem != nil {
		t.Error("expected nil for non-existent ID")
	}
}

func TestGetByID_EmptyID(t *testing.T) {
	repo, err := NewProblemRepository("../../problems")
	if err != nil {
		t.Fatal(err)
	}

	problem := repo.GetByID("")
	if problem != nil {
		t.Error("expected nil for empty ID")
	}
}

func TestCount(t *testing.T) {
	repo, err := NewProblemRepository("../../problems")
	if err != nil {
		t.Fatal(err)
	}

	if repo.Count() != 20 {
		t.Errorf("expected count 20, got %d", repo.Count())
	}
}

func TestLoad_Subdirectory(t *testing.T) {
	// Test that subdirectories are traversed
	tmpDir, err := os.MkdirTemp("", "test-subdir-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	subDir := filepath.Join(tmpDir, "easy")
	os.MkdirAll(subDir, 0755)

	yaml := `id: "sub-problem"
title: "Sub Problem"
difficulty: "easy"
category: "test"
tags: ["test"]
description: "test"
`
	os.WriteFile(filepath.Join(subDir, "sub-problem.yaml"), []byte(yaml), 0644)

	repo, err := NewProblemRepository(tmpDir)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if repo.Count() != 1 {
		t.Errorf("expected 1 problem from subdirectory, got %d", repo.Count())
	}
}

func TestLoad_SkipsNonYAMLFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-nonyaml-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a non-YAML file
	os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("not yaml"), 0644)
	// Create a valid YAML file
	yaml := `id: "valid"
title: "Valid"
difficulty: "easy"
category: "test"
tags: ["test"]
description: "test"
`
	os.WriteFile(filepath.Join(tmpDir, "valid.yaml"), []byte(yaml), 0644)

	repo, err := NewProblemRepository(tmpDir)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if repo.Count() != 1 {
		t.Errorf("expected 1 problem (non-YAML should be skipped), got %d", repo.Count())
	}
}

func TestSolutionField_Exists(t *testing.T) {
	repo, err := NewProblemRepository("../../problems")
	if err != nil {
		t.Fatal(err)
	}

	problem := repo.GetByID("two-sum")
	if problem == nil {
		t.Fatal("expected to find two-sum")
	}
	if problem.Solution == nil {
		t.Fatal("expected solution to be loaded from YAML")
	}
	if problem.Solution.Code == "" {
		t.Error("expected solution code to be non-empty")
	}
}

func TestProblem_ModelIntegrity(t *testing.T) {
	repo, err := NewProblemRepository("../../problems")
	if err != nil {
		t.Fatal(err)
	}

	problem := repo.GetByID("two-sum")
	if problem == nil {
		t.Fatal("expected two-sum")
	}

	// Verify all expected fields
	if problem.Title == "" {
		t.Error("title should not be empty")
	}
	if problem.Difficulty == "" {
		t.Error("difficulty should not be empty")
	}
	if problem.Category == "" {
		t.Error("category should not be empty")
	}
	if len(problem.Tags) == 0 {
		t.Error("tags should not be empty")
	}
	if problem.Description == "" {
		t.Error("description should not be empty")
	}
	if len(problem.Examples) == 0 {
		t.Error("examples should not be empty")
	}
	if len(problem.Hints) == 0 {
		t.Error("hints should not be empty")
	}
	if problem.Template == "" {
		t.Error("template should not be empty")
	}
	if len(problem.TestCases) == 0 {
		t.Error("test_cases should not be empty")
	}
}

func TestConcurrency(t *testing.T) {
	repo, err := NewProblemRepository("../../problems")
	if err != nil {
		t.Fatal(err)
	}

	// Run concurrent reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_ = repo.GetAll("", "", nil)
			_ = repo.GetByID("two-sum")
			_ = repo.Count()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// Helper to create a test problem
func createTestProblem(id string) *model.Problem {
	return &model.Problem{
		ID:          id,
		Title:       "Test " + id,
		Difficulty:  "easy",
		Category:    "test",
		Tags:        []string{"test"},
		Description: "Test description",
		Template:    "func test() {}",
		TestCases: []model.TestCase{
			{Input: "1", Expected: "1", Description: "test case"},
		},
		Hints: []model.Hint{
			{Level: 1, Title: "Hint 1", Content: "Try hard"},
		},
	}
}
