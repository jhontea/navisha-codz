package service

import (
	"fmt"
	"strings"
	"testing"

	"coding-challange/internal/model"
	"coding-challange/internal/repository"
)

func BenchmarkProblemService_ListProblems_NoFilter(b *testing.B) {
	repo, err := repository.NewProblemRepository("../../problems")
	if err != nil {
		b.Fatalf("failed to create repo: %v", err)
	}
	svc := NewProblemService(repo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = svc.ListProblems("", "", nil)
	}
}

func BenchmarkProblemService_ListProblems_WithDifficulty(b *testing.B) {
	repo, err := repository.NewProblemRepository("../../problems")
	if err != nil {
		b.Fatalf("failed to create repo: %v", err)
	}
	svc := NewProblemService(repo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = svc.ListProblems("medium", "", nil)
	}
}

func BenchmarkProblemService_GetProblem_Existing(b *testing.B) {
	repo, err := repository.NewProblemRepository("../../problems")
	if err != nil {
		b.Fatalf("failed to create repo: %v", err)
	}
	svc := NewProblemService(repo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := svc.GetProblem("two-sum")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkProblemService_GetProblemForAPI(b *testing.B) {
	repo, err := repository.NewProblemRepository("../../problems")
	if err != nil {
		b.Fatalf("failed to create repo: %v", err)
	}
	svc := NewProblemService(repo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := svc.GetProblemForAPI("two-sum")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkProblemService_GetTemplate(b *testing.B) {
	repo, err := repository.NewProblemRepository("../../problems")
	if err != nil {
		b.Fatalf("failed to create repo: %v", err)
	}
	svc := NewProblemService(repo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := svc.GetTemplate("two-sum")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkProblemService_ValidateCode_Valid(b *testing.B) {
	repo, err := repository.NewProblemRepository("../../problems")
	if err != nil {
		b.Fatalf("failed to create repo: %v", err)
	}
	svc := NewProblemService(repo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = svc.ValidateCode("func twoSum(nums []int, target int) []int { return nil }", "two-sum")
	}
}

func BenchmarkProblemService_ValidateCode_Empty(b *testing.B) {
	repo, err := repository.NewProblemRepository("../../problems")
	if err != nil {
		b.Fatalf("failed to create repo: %v", err)
	}
	svc := NewProblemService(repo)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = svc.ValidateCode("", "")
	}
}

func BenchmarkHintService_GetHints(b *testing.B) {
	svc := NewHintService()
	problem := &model.Problem{
		ID: "bench-problem",
		Hints: []model.Hint{
			{Level: 1, Title: "Hint 1", Content: strings.Repeat("A", 100)},
			{Level: 2, Title: "Hint 2", Content: strings.Repeat("B", 100)},
			{Level: 3, Title: "Hint 3", Content: strings.Repeat("C", 100)},
			{Level: 4, Title: "Hint 4", Content: strings.Repeat("D", 100)},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = svc.GetHints("bench-user", problem)
	}
}

func BenchmarkHintService_GetFullHints(b *testing.B) {
	svc := NewHintService()
	problem := &model.Problem{
		ID: "bench-problem",
		Hints: []model.Hint{
			{Level: 1, Title: "Hint 1", Content: "Content 1"},
			{Level: 2, Title: "Hint 2", Content: "Content 2"},
			{Level: 3, Title: "Hint 3", Content: "Content 3"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = svc.GetFullHints(problem)
	}
}

func BenchmarkSanitizeProblemID_Valid(b *testing.B) {
	ids := []string{"two-sum", "fizz-buzz", "valid-parentheses", "longest-palindromic-substring", "test-123_abc"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SanitizeProblemID(ids[i%len(ids)])
	}
}

func BenchmarkSanitizeProblemID_Invalid(b *testing.B) {
	ids := []string{"../../etc/passwd", "two sum", "bad$id", "../traverse", "test/path"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SanitizeProblemID(ids[i%len(ids)])
	}
}

func BenchmarkValidateCodeSize_Small(b *testing.B) {
	code := "func main() {}"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateCodeSize(code)
	}
}

func BenchmarkValidateCodeSize_CloseToLimit(b *testing.B) {
	code := strings.Repeat("a", 60*1024) // ~60KB
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateCodeSize(code)
	}
}

func BenchmarkBuildTestHarness_FunctionBased(b *testing.B) {
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildTestHarness(code, problem)
	}
}

func BenchmarkHintService_ConcurrentAccess(b *testing.B) {
	svc := NewHintService()
	problems := make([]*model.Problem, 10)
	for i := 0; i < 10; i++ {
		problems[i] = &model.Problem{
			ID: fmt.Sprintf("bench-problem-%d", i),
			Hints: []model.Hint{
				{Level: 1, Title: "H1", Content: "C1"},
				{Level: 2, Title: "H2", Content: "C2"},
				{Level: 3, Title: "H3", Content: "C3"},
			},
		}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		idx := 0
		for pb.Next() {
			_ = svc.GetHints("bench-user", problems[idx%len(problems)])
			idx++
		}
	})
}


