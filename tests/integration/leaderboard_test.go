package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ─────────────────────────────────────────────────────────────
// Leaderboard Integration Tests
// ─────────────────────────────────────────────────────────────

func TestLeaderboard_GetRankings(t *testing.T) {
	router := ginRouterWithLeaderboard()

	req := httptest.NewRequest("GET", "/api/leaderboard?limit=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp struct {
		Data []struct {
			Rank           int    `json:"rank"`
			UserID         string `json:"user_id"`
			Username       string `json:"username"`
			Score          int    `json:"score"`
			ProblemsSolved int    `json:"problems_solved"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse leaderboard: %v", err)
	}

	if len(resp.Data) == 0 {
		t.Error("expected non-empty leaderboard")
	}

	for i, entry := range resp.Data {
		if entry.Rank != i+1 {
			t.Errorf("expected rank %d, got %d", i+1, entry.Rank)
		}
	}
}

func TestLeaderboard_GetRankings_WithProblemFilter(t *testing.T) {
	router := ginRouterWithLeaderboard()

	req := httptest.NewRequest("GET", "/api/leaderboard?problem_id=two-sum&limit=5", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestLeaderboard_UpdateScore(t *testing.T) {
	router := ginRouterWithLeaderboard()

	body := `{"user_id":"user-1","problem_id":"two-sum","score":100}`
	req := httptest.NewRequest("POST", "/api/leaderboard/score", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["updated"] != true {
		t.Error("expected score to be updated")
	}
}

func TestLeaderboard_UpdateScore_MissingUserID(t *testing.T) {
	router := ginRouterWithLeaderboard()

	body := `{"problem_id":"two-sum","score":100}`
	req := httptest.NewRequest("POST", "/api/leaderboard/score", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing user_id, got %d", w.Code)
	}
}

func TestLeaderboard_UpdateScore_MissingProblemID(t *testing.T) {
	router := ginRouterWithLeaderboard()

	body := `{"user_id":"user-1","score":100}`
	req := httptest.NewRequest("POST", "/api/leaderboard/score", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing problem_id, got %d", w.Code)
	}
}

func TestLeaderboard_UpdateScore_MissingScore(t *testing.T) {
	router := ginRouterWithLeaderboard()

	body := `{"user_id":"user-1","problem_id":"two-sum"}`
	req := httptest.NewRequest("POST", "/api/leaderboard/score", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing score, got %d", w.Code)
	}
}

func TestLeaderboard_UpdateScore_InvalidJSON(t *testing.T) {
	router := ginRouterWithLeaderboard()

	body := `{invalid`
	req := httptest.NewRequest("POST", "/api/leaderboard/score", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", w.Code)
	}
}

func TestLeaderboard_GetUserRank(t *testing.T) {
	router := ginRouterWithLeaderboard()

	req := httptest.NewRequest("GET", "/api/leaderboard/user/user-1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestLeaderboard_GetUserRank_NotFound(t *testing.T) {
	router := ginRouterWithLeaderboard()

	req := httptest.NewRequest("GET", "/api/leaderboard/user/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// May return 200 with empty rank or 404
	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Errorf("expected 200 or 404, got %d", w.Code)
	}
}

func TestLeaderboard_Empty(t *testing.T) {
	router := ginRouterWithLeaderboard()

	req := httptest.NewRequest("GET", "/api/leaderboard?limit=0", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestLeaderboard_LargeLimit(t *testing.T) {
	router := ginRouterWithLeaderboard()

	req := httptest.NewRequest("GET", "/api/leaderboard?limit=1000", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
