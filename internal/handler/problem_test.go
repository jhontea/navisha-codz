package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"coding-challange/internal/model"
	"coding-challange/internal/repository"
	"coding-challange/internal/service"
)

// setupTestRouter creates a gin router with the handler for testing
func setupTestRouter(t *testing.T) *gin.Engine {
	gin.SetMode(gin.TestMode)

	// Create a real repository from test problems
	repo, err := repository.NewProblemRepository("../../problems")
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
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

func TestHealthCheck(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["data"] == nil {
		t.Fatal("expected data field in response")
	}
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected data to be a map")
	}
	if data["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", data["status"])
	}
}

func TestListProblems_All(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/problems", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp model.APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	problems, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatalf("expected data to be array, got %T", resp.Data)
	}
	if len(problems) != 25 {
		t.Errorf("expected 25 problems, got %d", len(problems))
	}

	// Verify summary fields (no solution, no test_cases)
	for _, p := range problems {
		probMap, ok := p.(map[string]interface{})
		if !ok {
			continue
		}
		if probMap["id"] == nil {
			t.Error("problem ID should not be empty")
		}
		if probMap["title"] == nil {
			t.Error("problem title should not be empty")
		}
	}
}

func TestListProblems_FilterByDifficulty(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/problems?difficulty=easy", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp model.APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	problems, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatalf("expected data to be array, got %T", resp.Data)
	}
	if len(problems) != 9 {
		t.Errorf("expected 9 easy problems, got %d", len(problems))
	}
}

func TestListProblems_InvalidDifficulty(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/problems?difficulty=impossible", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if _, ok := resp["error"]; !ok {
		t.Error("expected error field in response")
	}
}

func TestGetProblem_ValidID(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/problems/two-sum", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp model.APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	problemMap, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be map, got %T", resp.Data)
	}
	if problemMap["id"] != "two-sum" {
		t.Errorf("expected ID 'two-sum', got %v", problemMap["id"])
	}
	if problemMap["title"] != "Two Sum" {
		t.Errorf("expected title 'Two Sum', got %v", problemMap["title"])
	}
	// Solution should NOT be in response
	if _, hasSolution := problemMap["solution"]; hasSolution {
		t.Error("solution should not be exposed in API response")
	}
}

func TestGetProblem_NotFound(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/problems/nonexistent", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestGetProblem_InvalidID(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/problems/bad$id", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetTemplate_ValidID(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/problems/two-sum/template", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp model.APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	dataMap, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be map, got %T", resp.Data)
	}
	if dataMap["template"] == nil {
		t.Error("expected template in response")
	}
	if dataMap["type"] != "function" {
		t.Errorf("expected type 'function', got %v", dataMap["type"])
	}
}

func TestGetTemplate_NotFound(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/problems/nonexistent/template", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestGetHints_ValidID(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/problems/two-sum/hints", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp model.APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	dataMap, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be map, got %T", resp.Data)
	}
	hints, ok := dataMap["hints"].([]interface{})
	if !ok {
		t.Fatal("expected hints array in response")
	}
	if len(hints) == 0 {
		t.Error("expected at least one hint")
	}
}

func TestGetHints_NotFound(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/problems/nonexistent/hints", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestGetHints_InvalidID(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/problems/bad$id/hints", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestListProblems_FilterByCategory(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/problems?category=array", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp model.APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	problems, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatalf("expected data to be array, got %T", resp.Data)
	}
	// two-sum, merge-sorted-arrays, contains-duplicate, max-subarray, binary-search, majority-element, coin-change, trapping-rain-water, permutations, best-time-to-buy-sell-stock, product-of-array-except-self, first-missing-positive are array category
	if len(problems) != 12 {
		t.Errorf("expected 12 array problems, got %d", len(problems))
	}
}

func TestGetProblem_ResponseHasNoSolution(t *testing.T) {
	router := setupTestRouter(t)

	// Test multiple problems to ensure solution is never exposed
	problemIDs := []string{"two-sum", "fizz-buzz", "valid-parentheses", "reverse-string", "merge-sorted-arrays", "longest-palindromic-substring"}
	for _, id := range problemIDs {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/problems/"+id, nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200 for %s, got %d", id, w.Code)
			continue
		}

		// Parse as raw map to check for solution field
		var raw map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
			t.Fatalf("failed to parse %s: %v", id, err)
		}

		// Check data field doesn't have solution
		if data, ok := raw["data"].(map[string]interface{}); ok {
			if _, hasSolution := data["solution"]; hasSolution {
				t.Errorf("problem %s should NOT have 'solution' field in response", id)
			}
		}
	}
}

func TestValidateCode_Valid(t *testing.T) {
	router := setupTestRouter(t)

	body := `{"code": "func twoSum(nums []int, target int) []int { return nil }", "problem_id": "two-sum"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp model.APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	resultMap, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be map, got %T", resp.Data)
	}
	if resultMap["valid"] != true {
		t.Errorf("expected valid=true, got %v", resultMap["valid"])
	}
}

func TestValidateCode_MissingCode(t *testing.T) {
	router := setupTestRouter(t)

	body := `{}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestValidateCode_EmptyCode(t *testing.T) {
	router := setupTestRouter(t)

	body := `{"code": "   "}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/validate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestResponseHasMeta(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/problems", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if _, ok := resp["meta"]; !ok {
		t.Error("expected meta field in response")
	}
}

func TestListProblems_FilterByType(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/problems?type=function", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp model.APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	problems, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatalf("expected data to be array, got %T", resp.Data)
	}
	// All problems are function-based now
	if len(problems) != 25 {
		t.Errorf("expected 25 function problems, got %d", len(problems))
	}
}

func TestListProblems_InvalidType(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/problems?type=invalid", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestListProblems_FilterByTags(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/problems?tags=backtracking", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp model.APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	problems, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatalf("expected data to be array, got %T", resp.Data)
	}
	if len(problems) != 4 {
		t.Errorf("expected 4 backtracking problems, got %d", len(problems))
	}
}

func TestListProblems_FilterByMultipleTags(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/problems?tags=dp,string", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp model.APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	problems, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatalf("expected data to be array, got %T", resp.Data)
	}
	if len(problems) != 2 {
		t.Errorf("expected 2 problems with dp+string tags, got %d", len(problems))
	}
}

func TestListProblems_FilterByTagsAndDifficulty(t *testing.T) {
	router := setupTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/problems?tags=backtracking&difficulty=hard", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp model.APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	problems, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatalf("expected data to be array, got %T", resp.Data)
	}
	if len(problems) != 2 {
		t.Errorf("expected 2 hard backtracking problems, got %d", len(problems))
	}
}
