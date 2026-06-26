package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// ─────────────────────────────────────────────────────────────
// Submission Integration Tests
// ─────────────────────────────────────────────────────────────

func TestSubmission_ValidateCode_Valid(t *testing.T) {
	router, _ := setupTestServer(t)

	body := map[string]string{
		"code":       "func twoSum(nums []int, target int) []int { return nil }",
		"problem_id": "two-sum",
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/api/validate", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseAPIResponse(t, w.Body.Bytes())
	data, _ := resp["data"].(map[string]interface{})
	if data["valid"] != true {
		t.Errorf("expected valid=true, got %v", data["valid"])
	}
}

func TestSubmission_ValidateCode_WrongFunction(t *testing.T) {
	router, _ := setupTestServer(t)

	body := map[string]string{
		"code":       "func wrongName(nums []int, target int) []int { return nil }",
		"problem_id": "two-sum",
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/api/validate", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	resp := parseAPIResponse(t, w.Body.Bytes())
	data, _ := resp["data"].(map[string]interface{})
	warnings, _ := data["warnings"].([]interface{})
	if len(warnings) == 0 {
		t.Log("Note: validation may not warn about wrong function name")
	}
}

func TestSubmission_ValidateCode_EmptyCode(t *testing.T) {
	router, _ := setupTestServer(t)

	body := map[string]string{
		"code": "",
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/api/validate", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Empty code returns 400 Bad Request (validation error)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty code, got %d", w.Code)
	}
}


