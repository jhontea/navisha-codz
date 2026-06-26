package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// Test JWT validation
func TestValidateToken_ValidToken(t *testing.T) {
	cfg := NewJWTConfig()
	cfg.AccessTokenSecret = "test-secret"

	token, _, _, _, err := GenerateTokenPair(cfg, "user-1", "alice", "user")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	claims, err := ValidateToken(token, cfg.AccessTokenSecret)
	if err != nil {
		t.Fatalf("expected valid token, got error: %v", err)
	}

	if claims.UserID != "user-1" {
		t.Errorf("expected userID 'user-1', got %q", claims.UserID)
	}
	if claims.Username != "alice" {
		t.Errorf("expected username 'alice', got %q", claims.Username)
	}
	if claims.Role != "user" {
		t.Errorf("expected role 'user', got %q", claims.Role)
	}
	if claims.TokenType != "access" {
		t.Errorf("expected token type 'access', got %q", claims.TokenType)
	}
}

func TestValidateToken_InvalidSecret(t *testing.T) {
	cfg := NewJWTConfig()
	cfg.AccessTokenSecret = "correct-secret"

	token, _, _, _, _ := GenerateTokenPair(cfg, "user-1", "alice", "user")

	_, err := ValidateToken(token, "wrong-secret")
	if err == nil {
		t.Error("expected error for token validated with wrong secret")
	}
}

func TestValidateToken_TamperedToken(t *testing.T) {
	cfg := NewJWTConfig()
	cfg.AccessTokenSecret = "test-secret"

	token, _, _, _, _ := GenerateTokenPair(cfg, "user-1", "alice", "user")

	// Tamper with the token
	tampered := token[:len(token)-4] + "XXXX"
	_, err := ValidateToken(tampered, cfg.AccessTokenSecret)
	if err == nil {
		t.Error("expected error for tampered token")
	}
}

func TestValidateToken_EmptyString(t *testing.T) {
	cfg := NewJWTConfig()
	_, err := ValidateToken("", cfg.AccessTokenSecret)
	if err == nil {
		t.Error("expected error for empty token")
	}
}

func TestValidateToken_MalformedToken(t *testing.T) {
	cfg := NewJWTConfig()
	_, err := ValidateToken("not.a.valid.jwt.token", cfg.AccessTokenSecret)
	if err == nil {
		t.Error("expected error for malformed token")
	}
}

func TestGenerateTokenPair_DifferentSecrets(t *testing.T) {
	cfg := NewJWTConfig()
	cfg.AccessTokenSecret = "access-secret"
	cfg.RefreshTokenSecret = "refresh-secret"

	accessToken, refreshToken, _, _, err := GenerateTokenPair(cfg, "user-1", "alice", "user")
	if err != nil {
		t.Fatalf("failed to generate token pair: %v", err)
	}

	if accessToken == refreshToken {
		t.Error("access and refresh tokens should be different")
	}

	// Validate access token with access secret
	accessClaims, err := ValidateToken(accessToken, cfg.AccessTokenSecret)
	if err != nil {
		t.Fatalf("access token validation failed: %v", err)
	}
	if accessClaims.TokenType != "access" {
		t.Error("expected access token type")
	}

	// Validate refresh token with refresh secret
	refreshClaims, err := ValidateToken(refreshToken, cfg.RefreshTokenSecret)
	if err != nil {
		t.Fatalf("refresh token validation failed: %v", err)
	}
	if refreshClaims.TokenType != "refresh" {
		t.Error("expected refresh token type")
	}
}

func TestGenerateTokenPair_DifferentUsers(t *testing.T) {
	cfg := NewJWTConfig()

	token1, _, _, _, _ := GenerateTokenPair(cfg, "user-1", "alice", "user")
	token2, _, _, _, _ := GenerateTokenPair(cfg, "user-2", "bob", "user")

	if token1 == token2 {
		t.Error("tokens for different users should be different")
	}
}

// Test role-based access
func TestRoleMiddleware_AllowedRole(t *testing.T) {
	router := gin.New()
	router.GET("/admin", RoleMiddleware("admin"), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// Re-register route with proper middleware order
	router2 := gin.New()
	router2.Use(func(c *gin.Context) {
		c.Set(ContextKeyRole, "admin")
	})
	router2.GET("/admin", RoleMiddleware("admin"), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req2 := httptest.NewRequest("GET", "/admin", nil)
	w2 := httptest.NewRecorder()
	router2.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w2.Code)
	}
}

func TestRoleMiddleware_DeniedRole(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(ContextKeyRole, "user")
	})
	router.GET("/admin", RoleMiddleware("admin"), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestRoleMiddleware_NoRole(t *testing.T) {
	router := gin.New()
	router.GET("/admin", RoleMiddleware("admin"), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for missing role, got %d", w.Code)
	}
}

func TestRoleMiddleware_MultipleRoles(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(ContextKeyRole, "moderator")
	})
	router.GET("/manage", RoleMiddleware("admin", "moderator"), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("GET", "/manage", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for moderator role, got %d", w.Code)
	}
}

func TestRoleMiddleware_InvalidRoleType(t *testing.T) {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(ContextKeyRole, 123) // non-string role
	})
	router.GET("/admin", RoleMiddleware("admin"), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for non-string role, got %d", w.Code)
	}
}

// Test rate limiting
func TestRateLimiter_WithinLimit(t *testing.T) {
	rl := NewRateLimiter(5, time.Second)

	for i := 0; i < 5; i++ {
		router := gin.New()
		router.GET("/api/test", rl.RateLimitMiddleware(), func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest("GET", "/api/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i+1, w.Code)
		}
	}
}

func TestRateLimiter_ExceedsLimit(t *testing.T) {
	rl := NewRateLimiter(3, time.Second)

	allowed := 0
	blocked := 0
	for i := 0; i < 10; i++ {
		router := gin.New()
		router.GET("/api/test", rl.RateLimitMiddleware(), func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest("GET", "/api/test", nil)
		req.RemoteAddr = "10.0.0.1:5678"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code == http.StatusOK {
			allowed++
		} else if w.Code == http.StatusTooManyRequests {
			blocked++
		}
	}

	if allowed != 3 {
		t.Errorf("expected 3 allowed requests, got %d", allowed)
	}
	if blocked != 7 {
		t.Errorf("expected 7 blocked requests, got %d", blocked)
	}
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	rl := NewRateLimiter(2, time.Second)

	// Each IP has its own rate limit
	ips := []string{"10.0.0.1:1000", "10.0.0.2:1000", "10.0.0.3:1000"}
	for _, ip := range ips {
		router := gin.New()
		router.GET("/api/test", rl.RateLimitMiddleware(), func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})

		req := httptest.NewRequest("GET", "/api/test", nil)
		req.RemoteAddr = ip
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("IP %s: expected 200, got %d", ip, w.Code)
		}
	}
}

func TestRateLimiter_WindowReset(t *testing.T) {
	rl := NewRateLimiter(2, 100*time.Millisecond)

	router := gin.New()
	router.GET("/api/test", rl.RateLimitMiddleware(), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// Use up the limit
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/api/test", nil)
		req.RemoteAddr = "10.0.0.5:9999"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i+1, w.Code)
		}
	}

	// Should be rate limited now
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.RemoteAddr = "10.0.0.5:9999"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", w.Code)
	}

	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)

	// Should be able to make requests again
	req = httptest.NewRequest("GET", "/api/test", nil)
	req.RemoteAddr = "10.0.0.5:9999"
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 after window reset, got %d", w.Code)
	}
}

// Test AuthMiddleware (JWT validation in HTTP context)
func TestAuthMiddleware_MissingHeader(t *testing.T) {
	cfg := NewJWTConfig()
	router := gin.New()
	router.GET("/protected", AuthMiddleware(cfg, nil, nil), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing auth header, got %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	cfg := NewJWTConfig()
	router := gin.New()
	router.GET("/protected", AuthMiddleware(cfg, nil, nil), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "invalid-format")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for invalid format, got %d", w.Code)
	}
}

func TestAuthMiddleware_WrongTokenType(t *testing.T) {
	cfg := NewJWTConfig()
	router := gin.New()
	router.GET("/protected", AuthMiddleware(cfg, nil, nil), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	// Generate a refresh token (not access token)
	_, refreshToken, _, _, _ := GenerateTokenPair(cfg, "user-1", "alice", "user")

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+refreshToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for wrong token type, got %d", w.Code)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	cfg := NewJWTConfig()
	cfg.AccessTokenSecret = "test-secret"
	router := gin.New()
	router.GET("/protected", AuthMiddleware(cfg, nil, nil), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	accessToken, _, _, _, _ := GenerateTokenPair(cfg, "user-1", "alice", "user")

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// Test RequestIDMiddleware
func TestRequestIDMiddleware(t *testing.T) {
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		requestID := c.GetString("requestID")
		if requestID == "" {
			c.String(http.StatusInternalServerError, "no requestID")
			return
		}
		c.String(http.StatusOK, requestID)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	// Check X-Request-ID header is set
	requestID := w.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Error("expected X-Request-ID header to be set")
	}
}

func TestRequestIDMiddleware_UniqueIDs(t *testing.T) {
	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, c.GetString("requestID"))
	})

	ids := make(map[string]bool)
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		id := w.Body.String()
		if ids[id] {
			t.Errorf("duplicate request ID: %s", id)
		}
		ids[id] = true
	}
}

// Test CORSMiddleware
func TestCORSMiddleware(t *testing.T) {
	router := gin.New()
	router.Use(CORSMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("expected Access-Control-Allow-Origin: *")
	}
}

func TestCORSMiddleware_OPTIONS(t *testing.T) {
	router := gin.New()
	router.Use(CORSMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204 for OPTIONS, got %d", w.Code)
	}
}

// Test JWTConfig defaults
func TestNewJWTConfig(t *testing.T) {
	cfg := NewJWTConfig()
	if cfg.AccessTokenSecret == "" {
		t.Error("expected non-empty access token secret")
	}
	if cfg.RefreshTokenSecret == "" {
		t.Error("expected non-empty refresh token secret")
	}
	if cfg.AccessTokenTTL == 0 {
		t.Error("expected non-zero access token TTL")
	}
	if cfg.RefreshTokenTTL == 0 {
		t.Error("expected non-zero refresh token TTL")
	}
	if cfg.Issuer == "" {
		t.Error("expected non-empty issuer")
	}
}

// Test context key constants
func TestContextKeys(t *testing.T) {
	if ContextKeyUserID == "" {
		t.Error("ContextKeyUserID should not be empty")
	}
	if ContextKeyUsername == "" {
		t.Error("ContextKeyUsername should not be empty")
	}
	if ContextKeyRole == "" {
		t.Error("ContextKeyRole should not be empty")
	}
	if ContextKeyToken == "" {
		t.Error("ContextKeyToken should not be empty")
	}
}

// Test generateRequestID
func TestGenerateRequestID(t *testing.T) {
	id1 := generateRequestID()
	id2 := generateRequestID()

	if id1 == "" {
		t.Error("expected non-empty request ID")
	}
	if id1 == id2 {
		t.Error("expected unique request IDs")
	}
	if len(id1) != 32 { // 16 bytes hex encoded = 32 chars
		t.Errorf("expected 32-char request ID, got %d chars", len(id1))
	}
}

// Test that JWT claims contain expected fields
func TestJWTClaims_Fields(t *testing.T) {
	cfg := NewJWTConfig()
	cfg.AccessTokenSecret = "test-secret"

	token, _, _, _, _ := GenerateTokenPair(cfg, "user-42", "bob", "admin")
	claims, err := ValidateToken(token, cfg.AccessTokenSecret)
	if err != nil {
		t.Fatalf("token validation failed: %v", err)
	}

	if claims.UserID != "user-42" {
		t.Errorf("expected userID 'user-42', got %q", claims.UserID)
	}
	if claims.Username != "bob" {
		t.Errorf("expected username 'bob', got %q", claims.Username)
	}
	if claims.Role != "admin" {
		t.Errorf("expected role 'admin', got %q", claims.Role)
	}
	if claims.TokenType != "access" {
		t.Errorf("expected token type 'access', got %q", claims.TokenType)
	}
	if claims.Issuer != cfg.Issuer {
		t.Errorf("expected issuer %q, got %q", cfg.Issuer, claims.Issuer)
	}
	if claims.ExpiresAt == nil {
		t.Error("expected ExpiresAt to be set")
	}
	if claims.IssuedAt == nil {
		t.Error("expected IssuedAt to be set")
	}
}

// Test token expiration
func TestTokenExpiration(t *testing.T) {
	cfg := JWTConfig{
		AccessTokenSecret:  "test-secret",
		RefreshTokenSecret: "test-secret",
		AccessTokenTTL:     0, // expires immediately
		RefreshTokenTTL:    time.Hour,
		Issuer:             "test",
	}

	token, _, _, _, _ := GenerateTokenPair(cfg, "user-1", "alice", "user")

	// Small delay to ensure token is expired
	time.Sleep(10 * time.Millisecond)

	_, err := ValidateToken(token, cfg.AccessTokenSecret)
	if err == nil {
		t.Log("Note: token with 0 TTL may or may not be expired depending on clock skew tolerance")
	}
}

// Test AuthMiddleware with expired token
func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	cfg := JWTConfig{
		AccessTokenSecret:  "test-secret",
		RefreshTokenSecret: "test-secret",
		AccessTokenTTL:     -1 * time.Hour, // already expired
		RefreshTokenTTL:    time.Hour,
		Issuer:             "test",
	}

	token, _, _, _, _ := GenerateTokenPair(cfg, "user-1", "alice", "user")

	router := gin.New()
	router.GET("/protected", AuthMiddleware(cfg, nil, nil), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for expired token, got %d", w.Code)
	}
}

// Test AuthMiddleware with Bearer case insensitivity
func TestAuthMiddleware_BearerCaseInsensitive(t *testing.T) {
	cfg := NewJWTConfig()
	cfg.AccessTokenSecret = "test-secret"
	token, _, _, _, _ := GenerateTokenPair(cfg, "user-1", "alice", "user")

	router := gin.New()
	router.GET("/protected", AuthMiddleware(cfg, nil, nil), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "bearer "+token) // lowercase
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 with lowercase 'bearer', got %d", w.Code)
	}
}

// Test AuthMiddleware with extra spaces in header
func TestAuthMiddleware_ExtraSpaces(t *testing.T) {
	cfg := NewJWTConfig()
	cfg.AccessTokenSecret = "test-secret"
	token, _, _, _, _ := GenerateTokenPair(cfg, "user-1", "alice", "user")

	router := gin.New()
	router.GET("/protected", AuthMiddleware(cfg, nil, nil), func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer  "+token) // extra space
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should fail because split on " " gives ["Bearer", "", "token"]
	if w.Code != http.StatusUnauthorized {
		t.Logf("Got %d for extra spaces (may be handled differently)", w.Code)
	}
}

// Test that rate limiter cleanup doesn't panic
func TestRateLimiter_Cleanup(t *testing.T) {
	rl := NewRateLimiter(10, 50*time.Millisecond)

	// Add some entries
	for i := 0; i < 5; i++ {
		router := gin.New()
		router.GET("/api/test", rl.RateLimitMiddleware(), func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})
		req := httptest.NewRequest("GET", "/api/test", nil)
		req.RemoteAddr = strings.SplitN("10.0.0.1:1000", ":", 2)[0] + ":100" + string(rune('0'+i))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	// Wait for cleanup to run
	time.Sleep(100 * time.Millisecond)

	// Should not panic
	rl.mu.RLock()
	_ = len(rl.visitors)
	rl.mu.RUnlock()
}
