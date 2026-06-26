package repository

import (
	"fmt"
	"testing"
)

func BenchmarkRepository_GetByID(b *testing.B) {
	repo, err := NewProblemRepository("../../problems")
	if err != nil {
		b.Fatalf("failed to create repo: %v", err)
	}

	ids := []string{"two-sum", "fizz-buzz", "valid-parentheses", "reverse-string", "merge-sorted-arrays"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := ids[i%len(ids)]
		_ = repo.GetByID(id)
	}
}

func BenchmarkRepository_GetAll_NoFilter(b *testing.B) {
	repo, err := NewProblemRepository("../../problems")
	if err != nil {
		b.Fatalf("failed to create repo: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.GetAll("", "", nil)
	}
}

func BenchmarkRepository_GetAll_FilterByDifficulty(b *testing.B) {
	repo, err := NewProblemRepository("../../problems")
	if err != nil {
		b.Fatalf("failed to create repo: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.GetAll("easy", "", nil)
	}
}

func BenchmarkRepository_GetAll_FilterByCategory(b *testing.B) {
	repo, err := NewProblemRepository("../../problems")
	if err != nil {
		b.Fatalf("failed to create repo: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.GetAll("", "array", nil)
	}
}

func BenchmarkRepository_GetAll_FilterByTags(b *testing.B) {
	repo, err := NewProblemRepository("../../problems")
	if err != nil {
		b.Fatalf("failed to create repo: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.GetAll("", "", []string{"dp", "string"})
	}
}

func BenchmarkRepository_GetAll_CombinedFilters(b *testing.B) {
	repo, err := NewProblemRepository("../../problems")
	if err != nil {
		b.Fatalf("failed to create repo: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.GetAll("hard", "", []string{"backtracking"})
	}
}

func BenchmarkRepository_Count(b *testing.B) {
	repo, err := NewProblemRepository("../../problems")
	if err != nil {
		b.Fatalf("failed to create repo: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.Count()
	}
}

func BenchmarkRepository_Load_FromMemory(b *testing.B) {
	repo, err := NewProblemRepository("../../problems")
	if err != nil {
		b.Fatalf("failed to create repo: %v", err)
	}

	// Benchmark re-reading from memory (concurrent reads)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = repo.GetAll("", "", nil)
			_ = repo.GetByID("two-sum")
			_ = repo.Count()
		}
	})
}

func BenchmarkRepository_GetAll_Concurrent(b *testing.B) {
	repo, err := NewProblemRepository("../../problems")
	if err != nil {
		b.Fatalf("failed to create repo: %v", err)
	}

	filters := []struct {
		difficulty string
		category   string
		tags       []string
	}{
		{"", "", nil},
		{"easy", "", nil},
		{"", "array", nil},
		{"hard", "", []string{"backtracking"}},
		{"medium", "string", nil},
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		idx := 0
		for pb.Next() {
			f := filters[idx%len(filters)]
			_ = repo.GetAll(f.difficulty, f.category, f.tags)
			idx++
		}
	})
}

func BenchmarkRepository_GetAll_NonExistentFilter(b *testing.B) {
	repo, err := NewProblemRepository("../../problems")
	if err != nil {
		b.Fatalf("failed to create repo: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.GetAll("impossible", "nonexistent", []string{"no-match"})
	}
}

func BenchmarkRepository_GetByID_NonExistent(b *testing.B) {
	repo, err := NewProblemRepository("../../problems")
	if err != nil {
		b.Fatalf("failed to create repo: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.GetByID(fmt.Sprintf("non-existent-%d", i))
	}
}
