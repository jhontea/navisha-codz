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
	if repo.Count() != 10 {
		t.Errorf("expected 10 problems, got %d", repo.Count())
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

	problems := repo.GetAll("", "")
	if len(problems) != 10 {
		t.Errorf("expected 10 problems, got %d", len(problems))
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

	easy := repo.GetAll("easy", "")
	if len(easy) != 5 {
		t.Errorf("expected 5 easy problems, got %d", len(easy))
	}
	for _, p := range easy {
		if p.Difficulty != "easy" {
			t.Errorf("expected easy, got %s", p.Difficulty)
		}
	}

	medium := repo.GetAll("medium", "")
	if len(medium) != 2 {
		t.Errorf("expected 2 medium problems, got %d", len(medium))
	}

	hard := repo.GetAll("hard", "")
	if len(hard) != 3 {
		t.Errorf("expected 3 hard problems, got %d", len(hard))
	}
}

func TestGetAll_FilterByCategory(t *testing.T) {
	repo, err := NewProblemRepository("../../problems")
	if err != nil {
		t.Fatal(err)
	}

	arrayProbs := repo.GetAll("", "array")
	if len(arrayProbs) != 6 {
		t.Errorf("expected 6 array problems, got %d", len(arrayProbs))
	}
}

func TestGetAll_CaseInsensitive(t *testing.T) {
	repo, err := NewProblemRepository("../../problems")
	if err != nil {
		t.Fatal(err)
	}

	upper := repo.GetAll("EASY", "")
	lower := repo.GetAll("easy", "")
	if len(upper) != len(lower) {
		t.Errorf("case insensitive filter failed: upper=%d, lower=%d", len(upper), len(lower))
	}
}

func TestGetAll_NonExistentFilter(t *testing.T) {
	repo, err := NewProblemRepository("../../problems")
	if err != nil {
		t.Fatal(err)
	}

	result := repo.GetAll("impossible", "")
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

	if repo.Count() != 10 {
		t.Errorf("expected count 10, got %d", repo.Count())
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
			_ = repo.GetAll("", "")
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
