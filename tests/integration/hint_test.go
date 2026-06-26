package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"coding-challange/internal/model"
	"coding-challange/internal/service"
)

// ─────────────────────────────────────────────────────────────
// Hint Integration Tests
// ─────────────────────────────────────────────────────────────

func TestHint_GetHints_Integration(t *testing.T) {
	router, _ := setupTestServer(t)

	req := httptest.NewRequest("GET", "/api/problems/two-sum/hints", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := parseAPIResponse(t, w.Body.Bytes())
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp["data"])
	}

	hints, ok := data["hints"].([]interface{})
	if !ok {
		t.Fatal("expected hints array in response")
	}
	if len(hints) == 0 {
		t.Error("expected at least one hint")
	}
}

func TestHint_GetHints_InvalidProblem(t *testing.T) {
	router, _ := setupTestServer(t)

	req := httptest.NewRequest("GET", "/api/problems/nonexistent/hints", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestHint_ProgressiveReveal_Integration(t *testing.T) {
	hintSvc := service.NewHintService()
	problem := &model.Problem{
		ID: "hint-test",
		Hints: []model.Hint{
			{Level: 1, Title: "Hint 1", Content: "First hint"},
			{Level: 2, Title: "Hint 2", Content: "Second hint"},
			{Level: 3, Title: "Hint 3", Content: "Third hint"},
			{Level: 4, Title: "Hint 4", Content: "Fourth hint"},
		},
	}

	// First request: should get 2 hints
	hints := hintSvc.GetHints("test-user", problem)
	if len(hints) != 2 {
		t.Errorf("first request: expected 2 hints, got %d", len(hints))
	}
	if hints[0].Level != 1 || hints[1].Level != 2 {
		t.Error("first request should return hints level 1 and 2")
	}

	// Second request: should get remaining 2 hints
	hints = hintSvc.GetHints("test-user", problem)
	if len(hints) != 2 {
		t.Errorf("second request: expected 2 hints, got %d", len(hints))
	}
	if hints[0].Level != 3 || hints[1].Level != 4 {
		t.Error("second request should return hints level 3 and 4")
	}

	// Third request: all hints revealed, should return all 4
	hints = hintSvc.GetHints("test-user", problem)
	if len(hints) != 4 {
		t.Errorf("third request: expected all 4 hints, got %d", len(hints))
	}
}

func TestHint_TrackUsage_Integration(t *testing.T) {
	hintSvc := service.NewHintService()
	problem := &model.Problem{
		ID: "usage-track-test",
		Hints: []model.Hint{
			{Level: 1, Title: "H1", Content: "C1"},
			{Level: 2, Title: "H2", Content: "C2"},
		},
	}

	// Track usage by requesting hints multiple times
	hintSvc.GetHints("test-user", problem)
	hintSvc.GetHints("test-user", problem)

	// After 2 requests, all hints should be revealed
	hints := hintSvc.GetHints("test-user", problem)
	if len(hints) != 2 {
		t.Errorf("expected all 2 hints after full reveal, got %d", len(hints))
	}

	// Reset and verify tracking is reset
	hintSvc.Reset("test-user", "usage-track-test")
	hints = hintSvc.GetHints("test-user", problem)
	if len(hints) != 2 {
		t.Errorf("expected 2 hints after reset, got %d", len(hints))
	}
}

func TestHint_SortedByLevel(t *testing.T) {
	hintSvc := service.NewHintService()
	problem := &model.Problem{
		ID: "sort-test",
		Hints: []model.Hint{
			{Level: 3, Title: "H3", Content: "C3"},
			{Level: 1, Title: "H1", Content: "C1"},
			{Level: 2, Title: "H2", Content: "C2"},
		},
	}

	hints := hintSvc.GetHints("test-user", problem)
	if len(hints) < 2 {
		t.Fatal("expected at least 2 hints")
	}

	// Should be sorted by level ascending
	for i := 1; i < len(hints); i++ {
		if hints[i].Level < hints[i-1].Level {
			t.Errorf("hints not sorted: level %d before %d", hints[i-1].Level, hints[i].Level)
		}
	}
}

func TestHint_SingleHint(t *testing.T) {
	hintSvc := service.NewHintService()
	problem := &model.Problem{
		ID: "single-hint",
		Hints: []model.Hint{
			{Level: 1, Title: "Only Hint", Content: "Content"},
		},
	}

	hints := hintSvc.GetHints("test-user", problem)
	if len(hints) != 1 {
		t.Errorf("expected 1 hint, got %d", len(hints))
	}

	// Second request: all revealed
	hints = hintSvc.GetHints("test-user", problem)
	if len(hints) != 1 {
		t.Errorf("expected 1 hint (all revealed), got %d", len(hints))
	}
}

func TestHint_EmptyHints(t *testing.T) {
	hintSvc := service.NewHintService()
	problem := &model.Problem{
		ID:    "no-hints",
		Hints: []model.Hint{},
	}

	hints := hintSvc.GetHints("test-user", problem)
	if len(hints) != 0 {
		t.Errorf("expected 0 hints, got %d", len(hints))
	}
}

func TestHint_NilProblem(t *testing.T) {
	hintSvc := service.NewHintService()
	hints := hintSvc.GetHints("test-user", nil)
	if hints != nil {
		t.Errorf("expected nil for nil problem, got %v", hints)
	}
}

func TestHint_GetFullHints(t *testing.T) {
	hintSvc := service.NewHintService()
	problem := &model.Problem{
		ID: "full-hints",
		Hints: []model.Hint{
			{Level: 1, Title: "H1", Content: "C1"},
			{Level: 2, Title: "H2", Content: "C2"},
			{Level: 3, Title: "H3", Content: "C3"},
		},
	}

	hints := hintSvc.GetFullHints(problem)
	if len(hints) != 3 {
		t.Errorf("expected 3 full hints, got %d", len(hints))
	}
}

func TestHint_ConcurrentAccess(t *testing.T) {
	hintSvc := service.NewHintService()
	problem := &model.Problem{
		ID: "concurrent-hints",
		Hints: []model.Hint{
			{Level: 1, Title: "H1", Content: "C1"},
			{Level: 2, Title: "H2", Content: "C2"},
			{Level: 3, Title: "H3", Content: "C3"},
		},
	}

	// Run concurrent hint requests
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			hintSvc.GetHints("test-user", problem)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
	// Should not panic or deadlock
}
