package integration

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"coding-challange/internal/handler"
	"coding-challange/internal/repository"
	"coding-challange/internal/service"
)

// setupProblemHandler creates a handler wired with real services for testing.
func setupProblemHandler(t *testing.T) *handler.ProblemHandler {
	t.Helper()
	repo, err := repository.NewProblemRepository("../../problems")
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}
	problemSvc := service.NewProblemService(repo)
	runnerSvc := service.NewRunnerService(10, 256)
	hintSvc := service.NewHintService()
	return handler.NewProblemHandler(problemSvc, runnerSvc, hintSvc)
}

func TestProblemList_Flow(t *testing.T) {
	router, _ := setupTestServer(t)

	// Step 1: List all problems
	req := httptest.NewRequest("GET", "/api/problems", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	resp := parseAPIResponse(t, w.Body.Bytes())
	data, _ := resp["data"].([]interface{})
	if len(data) == 0 {
		t.Fatal("expected problems in list")
	}

	// Step 2: Get first problem's ID and fetch its detail
	first := data[0].(map[string]interface{})
	problemID := first["id"].(string)

	req = httptest.NewRequest("GET", "/api/problems/"+problemID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("detail for %s: expected 200, got %d", problemID, w.Code)
	}

	// Step 3: Get template for the problem
	req = httptest.NewRequest("GET", "/api/problems/"+problemID+"/template", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("template for %s: expected 200, got %d", problemID, w.Code)
	}
}

func TestProblemList_FilterByDifficulty(t *testing.T) {
	router, _ := setupTestServer(t)

	for _, d := range []string{"easy", "medium", "hard"} {
		req := httptest.NewRequest("GET", "/api/problems?difficulty="+d, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("difficulty=%s: expected 200, got %d", d, w.Code)
			continue
		}

		resp := parseAPIResponse(t, w.Body.Bytes())
		problems, _ := resp["data"].([]interface{})
		for _, p := range problems {
			prob := p.(map[string]interface{})
			if prob["difficulty"] != d {
				t.Errorf("filter=%s: got difficulty %v", d, prob["difficulty"])
			}
		}
	}
}

func TestProblemList_FilterByType(t *testing.T) {
	router, _ := setupTestServer(t)

	for _, tp := range []string{"function", "main"} {
		req := httptest.NewRequest("GET", "/api/problems?type="+tp, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("type=%s: expected 200, got %d", tp, w.Code)
			continue
		}

		resp := parseAPIResponse(t, w.Body.Bytes())
		problems, _ := resp["data"].([]interface{})
		for _, p := range problems {
			prob := p.(map[string]interface{})
			if prob["type"] != tp {
				t.Errorf("filter=%s: got type %v", tp, prob["type"])
			}
		}
	}
}

func TestProblemList_FilterInvalidDifficulty(t *testing.T) {
	router, _ := setupTestServer(t)

	req := httptest.NewRequest("GET", "/api/problems?difficulty=superhard", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid difficulty, got %d", w.Code)
	}
}

func TestProblemList_FilterInvalidType(t *testing.T) {
	router, _ := setupTestServer(t)

	req := httptest.NewRequest("GET", "/api/problems?type=invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid type, got %d", w.Code)
	}
}

func TestProblemDetail_NotFound(t *testing.T) {
	router, _ := setupTestServer(t)

	req := httptest.NewRequest("GET", "/api/problems/this-does-not-exist", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestProblemDetail_InvalidID(t *testing.T) {
	router, _ := setupTestServer(t)

	// Gin normalizes paths, so we test with a space character which is invalid
	req := httptest.NewRequest("GET", "/api/problems/invalid id with spaces", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid ID with spaces, got %d", w.Code)
	}
}

func TestProblemDetail_SolutionHidden(t *testing.T) {
	router, _ := setupTestServer(t)

	req := httptest.NewRequest("GET", "/api/problems/two-sum", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	bodyStr := w.Body.String()
	if strings.Contains(bodyStr, "\"solution\"") {
		t.Error("solution should not be exposed in API response")
	}
}

func TestProblemTemplate_Flow(t *testing.T) {
	router, _ := setupTestServer(t)

	req := httptest.NewRequest("GET", "/api/problems/two-sum/template", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	resp := parseAPIResponse(t, w.Body.Bytes())
	data, _ := resp["data"].(map[string]interface{})

	if data["template"] == nil || data["template"] == "" {
		t.Error("expected non-empty template")
	}
	if data["function_name"] != "twoSum" {
		t.Errorf("expected function_name 'twoSum', got %v", data["function_name"])
	}
}

func TestProblemList_MetaPresent(t *testing.T) {
	router, _ := setupTestServer(t)

	req := httptest.NewRequest("GET", "/api/problems", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	resp := parseAPIResponse(t, w.Body.Bytes())
	meta, ok := resp["meta"].(map[string]interface{})
	if !ok {
		t.Fatal("expected meta in response")
	}
	if meta["request_id"] == nil || meta["request_id"] == "" {
		t.Error("expected request_id in meta")
	}
	if meta["timestamp"] == nil || meta["timestamp"] == "" {
		t.Error("expected timestamp in meta")
	}
}

func TestProblemService_ListProblems(t *testing.T) {
	repo, err := repository.NewProblemRepository("../../problems")
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}
	svc := service.NewProblemService(repo)

	problems := svc.ListProblems("", "")
	if len(problems) == 0 {
		t.Error("expected problems from ListProblems")
	}
}

func TestProblemService_GetProblem(t *testing.T) {
	repo, err := repository.NewProblemRepository("../../problems")
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}
	svc := service.NewProblemService(repo)

	problem, err := svc.GetProblem("two-sum")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if problem.ID != "two-sum" {
		t.Errorf("expected ID 'two-sum', got %q", problem.ID)
	}
	if problem.Solution != nil {
		t.Error("solution should be nil when returned from GetProblem")
	}
}

func TestProblemService_GetProblemForAPI(t *testing.T) {
	repo, err := repository.NewProblemRepository("../../problems")
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}
	svc := service.NewProblemService(repo)

	p, err := svc.GetProblemForAPI("two-sum")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, tc := range p.TestCases {
		if tc.Expected != nil {
			t.Error("expected values should be stripped from API response")
		}
	}
}

func TestProblemService_GetTemplate(t *testing.T) {
	repo, err := repository.NewProblemRepository("../../problems")
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}
	svc := service.NewProblemService(repo)

	tmpl, err := svc.GetTemplate("two-sum")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tmpl["template"] == nil {
		t.Error("expected template in result")
	}
	if tmpl["function_name"] != "twoSum" {
		t.Errorf("expected function_name 'twoSum', got %v", tmpl["function_name"])
	}
}

func TestProblemRepository_Load(t *testing.T) {
	repo, err := repository.NewProblemRepository("../../problems")
	if err != nil {
		t.Fatalf("failed to load problems: %v", err)
	}
	if repo.Count() == 0 {
		t.Error("expected problems to be loaded")
	}
}

func TestProblemRepository_GetByID(t *testing.T) {
	repo, err := repository.NewProblemRepository("../../problems")
	if err != nil {
		t.Fatalf("failed to load problems: %v", err)
	}

	problem := repo.GetByID("two-sum")
	if problem == nil {
		t.Fatal("expected to find 'two-sum' problem")
	}
	if problem.Title != "Two Sum" {
		t.Errorf("expected title 'Two Sum', got %q", problem.Title)
	}
}

func TestProblemRepository_GetAll_Filtered(t *testing.T) {
	repo, err := repository.NewProblemRepository("../../problems")
	if err != nil {
		t.Fatalf("failed to load problems: %v", err)
	}

	easy := repo.GetAll("easy", "")
	for _, p := range easy {
		if p.Difficulty != "easy" {
			t.Errorf("expected easy, got %s", p.Difficulty)
		}
	}

	medium := repo.GetAll("medium", "")
	for _, p := range medium {
		if p.Difficulty != "medium" {
			t.Errorf("expected medium, got %s", p.Difficulty)
		}
	}

	hard := repo.GetAll("hard", "")
	for _, p := range hard {
		if p.Difficulty != "hard" {
			t.Errorf("expected hard, got %s", p.Difficulty)
		}
	}
}
