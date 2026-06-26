package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"coding-challange/internal/repository"
	"coding-challange/internal/service"
)

func setupBenchRouter(b *testing.B, disableRateLimit bool) *gin.Engine {
	gin.SetMode(gin.TestMode)

	repo, err := repository.NewProblemRepository("../../problems")
	if err != nil {
		b.Fatalf("failed to create repo: %v", err)
	}

	problemSvc := service.NewProblemService(repo)
	hintSvc := service.NewHintService()

	handler := NewProblemHandler(problemSvc, hintSvc)

	router := gin.New()
	router.GET("/health", handler.HealthCheck)

	api := router.Group("/api")
	{
		api.GET("/problems", handler.ListProblems)
		api.GET("/problems/:id", handler.GetProblem)
		api.GET("/problems/:id/template", handler.GetTemplate)
		api.POST("/validate", handler.ValidateCode)
		api.GET("/problems/:id/hints", handler.GetHints)
	}

	return router
}

func benchIP(i int) string {
	return fmt.Sprintf("10.%d.%d.%d", (i>>16)&255, (i>>8)&255, i&255)
}

func BenchmarkListProblems_All(b *testing.B) {
	router := setupBenchRouter(b, true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/problems", nil)
		req.Header.Set("X-Forwarded-For", benchIP(i))
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("expected 200, got %d", w.Code)
		}
	}
}

func BenchmarkListProblems_FilterByDifficulty(b *testing.B) {
	router := setupBenchRouter(b, true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/problems?difficulty=easy", nil)
		req.Header.Set("X-Forwarded-For", benchIP(i+1))
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("expected 200, got %d", w.Code)
		}
	}
}

func BenchmarkGetProblem_ValidID(b *testing.B) {
	router := setupBenchRouter(b, true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/problems/two-sum", nil)
		req.Header.Set("X-Forwarded-For", benchIP(i+2))
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("expected 200, got %d", w.Code)
		}
	}
}

func BenchmarkGetProblem_NotFound(b *testing.B) {
	router := setupBenchRouter(b, true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/problems/nonexistent", nil)
		req.Header.Set("X-Forwarded-For", benchIP(i+3))
		router.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			b.Fatalf("expected 404, got %d", w.Code)
		}
	}
}

func BenchmarkGetProblem_InvalidID(b *testing.B) {
	router := setupBenchRouter(b, true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/problems/bad$id", nil)
		req.Header.Set("X-Forwarded-For", benchIP(i+4))
		router.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			b.Fatalf("expected 400, got %d", w.Code)
		}
	}
}

func BenchmarkGetTemplate_ValidID(b *testing.B) {
	router := setupBenchRouter(b, true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/problems/two-sum/template", nil)
		req.Header.Set("X-Forwarded-For", benchIP(i+5))
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("expected 200, got %d", w.Code)
		}
	}
}

func BenchmarkGetHints_ValidID(b *testing.B) {
	router := setupBenchRouter(b, true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/problems/two-sum/hints", nil)
		req.Header.Set("X-Forwarded-For", benchIP(i+6))
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("expected 200, got %d", w.Code)
		}
	}
}

func BenchmarkValidateCode_Valid(b *testing.B) {
	router := setupBenchRouter(b, true)
	body := `{"code": "func twoSum(nums []int, target int) []int { return nil }", "problem_id": "two-sum"}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/validate", strings.NewReader(body))
		req.Header.Set("X-Forwarded-For", benchIP(i+7))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("expected 200, got %d", w.Code)
		}
	}
}

func BenchmarkHealthCheck(b *testing.B) {
	router := setupBenchRouter(b, true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		req.Header.Set("X-Forwarded-For", benchIP(i+8))
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("expected 200, got %d", w.Code)
		}
	}
}
