package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"coding-challange/internal/repository"
	"coding-challange/pkg/middleware"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// setupTestServer creates a test server with all routes configured.
func setupTestServer(t *testing.T) (*gin.Engine, *repository.ProblemRepository) {
	t.Helper()

	repo, err := repository.NewProblemRepository("../../problems")
	if err != nil {
		t.Fatalf("failed to create repository: %v", err)
	}

	problemSvc := service.NewProblemService(repo)
	runnerSvc := service.NewRunnerService(10, 256)
	hintSvc := service.NewHintService()

	h := handler.NewProblemHandler(problemSvc, runnerSvc, hintSvc)

	router := gin.New()
	router.Use(middleware.RequestIDMiddleware())

	api := router.Group("/api")
	{
		api.GET("/problems", h.ListProblems)
		api.GET("/problems/:id", h.GetProblem)
		api.GET("/problems/:id/template", h.GetTemplate)
		api.POST("/problems/:id/run", h.RunCode)
		api.POST("/api/validate", h.ValidateCode)
		api.GET("/problems/:id/hints", h.GetHints)
	}

	router.GET("/health", h.HealthCheck)

	return router, repo
}

// parseAPIResponse parses a standard API response.
func parseAPIResponse(t *testing.T, body []byte) map[string]interface{} {
	t.Helper()
	var resp map[string]interface{}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	return resp
}

// ─────────────────────────────────────────────────────────────
// Auth Integration Tests
// ─────────────────────────────────────────────────────────────

func TestAuth_RegisterFlow(t *testing.T) {
	router := gin.New()
	router.POST("/api/auth/register", func(c *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required"`
			Email    string `json:"email" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{
			"user_id":  "user-" + req.Username,
			"username": req.Username,
			"message":  "registration successful",
		})
	})

	body := `{"username":"testuser","email":"test@example.com","password":"secure123"}`
	req := httptest.NewRequest("POST", "/api/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["username"] != "testuser" {
		t.Errorf("expected username 'testuser', got %v", resp["username"])
	}
}

func TestAuth_LoginFlow(t *testing.T) {
	cfg := middleware.NewJWTConfig()
	cfg.AccessTokenSecret = "login-test-secret"
	cfg.RefreshTokenSecret = "login-refresh-secret"

	router := gin.New()
	router.POST("/api/auth/login", func(c *gin.Context) {
		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		accessToken, refreshToken, err := middleware.GenerateTokenPair(cfg, "user-1", req.Username, "user")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"token_type":    "Bearer",
		})
	})

	body := `{"username":"alice","password":"password123"}`
	req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse login response: %v", err)
	}

	if resp.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if resp.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}

	claims, err := middleware.ValidateToken(resp.AccessToken, cfg.AccessTokenSecret)
	if err != nil {
		t.Fatalf("access token validation failed: %v", err)
	}
	if claims.Username != "alice" {
		t.Errorf("expected username 'alice', got %q", claims.Username)
	}
}

func TestAuth_TokenRefreshFlow(t *testing.T) {
	cfg := middleware.NewJWTConfig()
	cfg.AccessTokenSecret = "refresh-access-secret"
	cfg.RefreshTokenSecret = "refresh-refresh-secret"

	router := gin.New()
	router.POST("/api/auth/login", func(c *gin.Context) {
		accessToken, refreshToken, _ := middleware.GenerateTokenPair(cfg, "user-1", "bob", "user")
		c.JSON(http.StatusOK, gin.H{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		})
	})
	router.POST("/api/auth/refresh", func(c *gin.Context) {
		var req struct {
			RefreshToken string `json:"refresh_token" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		claims, err := middleware.ValidateToken(req.RefreshToken, cfg.RefreshTokenSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
			return
		}
		if claims.TokenType != "refresh" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not a refresh token"})
			return
		}
		newAccessToken, newRefreshToken, _ := middleware.GenerateTokenPair(cfg, claims.UserID, claims.Username, claims.Role)
		c.JSON(http.StatusOK, gin.H{
			"access_token":  newAccessToken,
			"refresh_token": newRefreshToken,
		})
	})

	loginBody := `{"username":"bob","password":"pass"}`
	req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var loginResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	json.Unmarshal(w.Body.Bytes(), &loginResp)

	refreshBody := `{"refresh_token":"` + loginResp.RefreshToken + `"}`
	req = httptest.NewRequest("POST", "/api/auth/refresh", strings.NewReader(refreshBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for token refresh, got %d: %s", w.Code, w.Body.String())
	}

	var refreshResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	json.Unmarshal(w.Body.Bytes(), &refreshResp)

	if refreshResp.AccessToken == "" {
		t.Error("expected new access token after refresh")
	}
	if refreshResp.AccessToken == loginResp.AccessToken {
		t.Error("new access token should differ from old one")
	}
}

func TestAuth_ProtectedEndpoint(t *testing.T) {
	cfg := middleware.NewJWTConfig()
	cfg.AccessTokenSecret = "protected-test-secret"

	router := gin.New()
	router.GET("/api/protected", middleware.AuthMiddleware(cfg), func(c *gin.Context) {
		userID, _ := c.Get(middleware.ContextKeyUserID)
		username, _ := c.Get(middleware.ContextKeyUsername)
		c.JSON(http.StatusOK, gin.H{
			"user_id":  userID,
			"username": username,
		})
	})

	req := httptest.NewRequest("GET", "/api/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without token, got %d", w.Code)
	}

	token, _, _ := middleware.GenerateTokenPair(cfg, "user-42", "charlie", "user")
	req = httptest.NewRequest("GET", "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 with valid token, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuth_AdminEndpoint(t *testing.T) {
	cfg := middleware.NewJWTConfig()
	cfg.AccessTokenSecret = "admin-test-secret"

	router := gin.New()
	router.GET("/api/admin", middleware.AuthMiddleware(cfg), middleware.RoleMiddleware("admin"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
	})

	// Regular user should be denied
	userToken, _, _ := middleware.GenerateTokenPair(cfg, "user-1", "alice", "user")
	req := httptest.NewRequest("GET", "/api/admin", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for non-admin, got %d", w.Code)
	}

	// Admin should be granted access
	adminToken, _, _ := middleware.GenerateTokenPair(cfg, "admin-1", "admin", "admin")
	req = httptest.NewRequest("GET", "/api/admin", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for admin, got %d", w.Code)
	}
}
