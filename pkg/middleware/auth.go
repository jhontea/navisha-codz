package middleware

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"coding-challange/pkg/security"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Context keys.
const (
	ContextKeyUserID    = "userID"
	ContextKeyUsername  = "username"
	ContextKeyRole      = "role"
	ContextKeyToken     = "token"
	ContextKeySessionID = "sessionID"
	ContextKeyTokenID   = "tokenID"
)

// JWTConfig holds JWT configuration.
type JWTConfig struct {
	AccessTokenSecret    string
	RefreshTokenSecret   string
	AccessTokenTTL       time.Duration
	RefreshTokenTTL      time.Duration
	Issuer               string
	EnableTokenBlacklist bool
}

// JWTClaims represents the claims in our JWT tokens.
type JWTClaims struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"`
	SessionID string `json:"session_id,omitempty"`
	TokenID   string `json:"token_id,omitempty"`
	jwt.RegisteredClaims
}

// NewJWTConfig creates JWT config from environment variables.
// Panics if JWT_ACCESS_SECRET or JWT_REFRESH_SECRET are not set.
func NewJWTConfig() JWTConfig {
	accessSecret := getEnvOrPanic("JWT_ACCESS_SECRET")
	refreshSecret := getEnvOrPanic("JWT_REFRESH_SECRET")
	return JWTConfig{
		AccessTokenSecret:    accessSecret,
		RefreshTokenSecret:   refreshSecret,
		AccessTokenTTL:       time.Duration(getEnvInt("JWT_ACCESS_TTL_MIN", 15)) * time.Minute,
		RefreshTokenTTL:      time.Duration(getEnvInt("JWT_REFRESH_TTL_HOURS", 168)) * time.Hour,
		Issuer:               getEnv("JWT_ISSUER", "coding-challange"),
		EnableTokenBlacklist: getEnv("JWT_ENABLE_BLACKLIST", "true") == "true",
	}
}

// GenerateTokenPair generates access and refresh tokens with token IDs.
func GenerateTokenPair(cfg JWTConfig, userID, username, role string) (accessToken, refreshToken string, tokenID string, sessionID string, err error) {
	tokenID, _ = generateRandomID()
	sessionID, _ = generateRandomID()

	// Access token
	accessClaims := JWTClaims{
		UserID:    userID,
		Username:  username,
		Role:      role,
		TokenType: "access",
		SessionID: sessionID,
		TokenID:   tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    cfg.Issuer,
			ID:        tokenID,
		},
	}
	access := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = access.SignedString([]byte(cfg.AccessTokenSecret))
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	// Refresh token
	refreshClaims := JWTClaims{
		UserID:    userID,
		TokenType: "refresh",
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.RefreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    cfg.Issuer,
			ID:        tokenID,
		},
	}
	refresh := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refresh.SignedString([]byte(cfg.RefreshTokenSecret))
	if err != nil {
		return "", "", "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return accessToken, refreshToken, tokenID, sessionID, nil
}

// ValidateToken validates a JWT token and returns the claims.
func ValidateToken(tokenString string, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// ============================================================================
// Token Blacklist
// ============================================================================

// TokenBlacklist manages blacklisted (revoked) tokens.
type TokenBlacklist struct {
	mu      sync.RWMutex
	tokens  map[string]time.Time // tokenID -> expiry
	enabled bool
}

// NewTokenBlacklist creates a new token blacklist.
func NewTokenBlacklist(enabled bool) *TokenBlacklist {
	bl := &TokenBlacklist{
		tokens:  make(map[string]time.Time),
			}
	go bl.cleanup()
	return bl
}

// Blacklist adds a token to the blacklist.
func (bl *TokenBlacklist) Blacklist(tokenID string, expiry time.Time) {
	bl.mu.Lock()
	defer bl.mu.Unlock()
	bl.tokens[tokenID] = expiry
}

// IsBlacklisted checks if a token is blacklisted.
func (bl *TokenBlacklist) IsBlacklisted(tokenID string) bool {
	if !bl.enabled {
		return false
	}
	bl.mu.RLock()
	defer bl.mu.RUnlock()

	expiry, exists := bl.tokens[tokenID]
	if !exists {
		return false
	}

	// If token has expired, it's not blacklisted anymore
	if time.Now().After(expiry) {
		return false
	}
	return true
}

// cleanup removes expired entries periodically.
func (bl *TokenBlacklist) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		bl.mu.Lock()
		now := time.Now()
		for id, expiry := range bl.tokens {
			if now.After(expiry) {
				delete(bl.tokens, id)
			}
		}
		bl.mu.Unlock()
	}
}

// ============================================================================
// Session Management
// ============================================================================

// Session represents a user session.
type Session struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	Role         string    `json:"role"`
	IP           string    `json:"ip"`
	UserAgent    string    `json:"user_agent"`
	CreatedAt    time.Time `json:"created_at"`
	LastActiveAt time.Time `json:"last_active_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	RefreshToken string    `json:"-"`
	Active       bool      `json:"active"`
}

// SessionManager manages user sessions.
type SessionManager struct {
	mu          sync.RWMutex
	sessions    map[string]*Session
	userSessions map[string][]string // userID -> session IDs
	maxSessions int
}

// NewSessionManager creates a new session manager.
func NewSessionManager(maxSessionsPerUser int) *SessionManager {
	if maxSessionsPerUser <= 0 {
		maxSessionsPerUser = 5
	}
	sm := &SessionManager{
		sessions:     make(map[string]*Session),
		userSessions: make(map[string][]string),
		maxSessions:  maxSessionsPerUser,
	}
	go sm.cleanup()
	return sm
}

// CreateSession creates a new session.
func (sm *SessionManager) CreateSession(userID, username, role, ip, userAgent, refreshToken string, ttl time.Duration) *Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sessionID, _ := generateRandomID()
	now := time.Now()

	session := &Session{
		ID:           sessionID,
		UserID:       userID,
		Username:     username,
		Role:         role,
		IP:           ip,
		UserAgent:    userAgent,
		CreatedAt:    now,
		LastActiveAt: now,
		ExpiresAt:    now.Add(ttl),
		RefreshToken: refreshToken,
		Active:       true,
	}

	sm.sessions[sessionID] = session
	sm.userSessions[userID] = append(sm.userSessions[userID], sessionID)

	// Enforce max sessions per user
	if len(sm.userSessions[userID]) > sm.maxSessions {
		// Revoke oldest session
		oldestID := sm.userSessions[userID][0]
		if oldest, exists := sm.sessions[oldestID]; exists {
			oldest.Active = false
		}
		sm.userSessions[userID] = sm.userSessions[userID][1:]
	}

	return session
}

// CreateSessionWithID creates a session with a specific ID (used for testing).
func (sm *SessionManager) CreateSessionWithID(sessionID, userID, username, role, ip, userAgent, refreshToken string, ttl time.Duration) *Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	session := &Session{
		ID:           sessionID,
		UserID:       userID,
		Username:     username,
		Role:         role,
		IP:           ip,
		UserAgent:    userAgent,
		CreatedAt:    now,
		LastActiveAt: now,
		ExpiresAt:    now.Add(ttl),
		RefreshToken: refreshToken,
		Active:       true,
	}

	sm.sessions[sessionID] = session
	sm.userSessions[userID] = append(sm.userSessions[userID], sessionID)

	return session
}

// GetSession retrieves a session by ID.
func (sm *SessionManager) GetSession(sessionID string) *Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists || !session.Active || time.Now().After(session.ExpiresAt) {
		return nil
	}
	return session
}

// RevokeSession revokes a session.
func (sm *SessionManager) RevokeSession(sessionID string) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return false
	}

	session.Active = false
	return true
}

// RevokeAllUserSessions revokes all sessions for a user.
func (sm *SessionManager) RevokeAllUserSessions(userID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, sessionID := range sm.userSessions[userID] {
		if session, exists := sm.sessions[sessionID]; exists {
			session.Active = false
		}
	}
}

// GetUserSessions returns all active sessions for a user.
func (sm *SessionManager) GetUserSessions(userID string) []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessions := make([]*Session, 0)
	for _, sessionID := range sm.userSessions[userID] {
		if session, exists := sm.sessions[sessionID]; exists && session.Active {
			sessions = append(sessions, session)
		}
	}
	return sessions
}

// UpdateLastActivity updates the last activity timestamp for a session.
func (sm *SessionManager) UpdateLastActivity(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[sessionID]; exists {
		session.LastActiveAt = time.Now()
	}
}

// RevokeSessionByDeviceID revokes all sessions for a specific device (by device fingerprint).
func (sm *SessionManager) RevokeSessionByDeviceID(userID, deviceFP string) int {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	revoked := 0
	for _, sessionID := range sm.userSessions[userID] {
		session, exists := sm.sessions[sessionID]
		if exists && session.Active && session.UserAgent == deviceFP {
			session.Active = false
			revoked++
		}
	}
	return revoked
}

// GetUserSessionsByDevice returns all active sessions for a user on a specific device.
func (sm *SessionManager) GetUserSessionsByDevice(userID, deviceFP string) []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var sessions []*Session
	for _, sessionID := range sm.userSessions[userID] {
		if session, exists := sm.sessions[sessionID]; exists && session.Active && session.UserAgent == deviceFP {
			sessions = append(sessions, session)
		}
	}
	return sessions
}

// cleanup removes expired sessions.
func (sm *SessionManager) cleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		sm.mu.Lock()
		now := time.Now()
		for id, session := range sm.sessions {
			if !session.Active || now.After(session.ExpiresAt) {
				delete(sm.sessions, id)
				// Also remove from userSessions index
				if userSessions, ok := sm.userSessions[session.UserID]; ok {
					for i, sid := range userSessions {
						if sid == id {
							sm.userSessions[session.UserID] = append(userSessions[:i], userSessions[i+1:]...)
							break
						}
					}
					if len(sm.userSessions[session.UserID]) == 0 {
						delete(sm.userSessions, session.UserID)
					}
				}
			}
		}
		sm.mu.Unlock()
	}
}

// ============================================================================
// 2FA Support
// ============================================================================

// TOTPVerifier interface for TOTP verification.
type TOTPVerifier interface {
	GenerateSecret() (string, error)
	GenerateCode(secret string, t time.Time) (string, error)
	Validate(secret, code string, window int) bool
	ProvisioningURI(secret, username, issuer string) string
}

// DefaultTOTPVerifier uses the security package for TOTP.
type DefaultTOTPVerifier struct{}

// GenerateSecret generates a TOTP secret.
func (v *DefaultTOTPVerifier) GenerateSecret() (string, error) {
	return security.GenerateTOTPSecret()
}

// GenerateCode generates a TOTP code.
func (v *DefaultTOTPVerifier) GenerateCode(secret string, t time.Time) (string, error) {
	return security.GenerateTOTPCode(secret, t)
}

// Validate validates a TOTP code.
func (v *DefaultTOTPVerifier) Validate(secret, code string, window int) bool {
	return security.ValidateTOTP(secret, code, window)
}

// ProvisioningURI generates a provisioning URI.
func (v *DefaultTOTPVerifier) ProvisioningURI(secret, username, issuer string) string {
	return security.TOTPProvisioningURI(secret, username, issuer)
}

// TOTPManager manages TOTP for users.
type TOTPManager struct {
	mu       sync.RWMutex
	verifier TOTPVerifier
	secrets  map[string]string    // user -> secret (in production: store encrypted)
	enabled  map[string]bool      // user -> 2FA enabled
	backupCodes map[string][]string // user -> backup codes
}

// NewTOTPManager creates a new TOTP manager.
func NewTOTPManager(verifier TOTPVerifier) *TOTPManager {
	return &TOTPManager{
		verifier:    verifier,
		secrets:     make(map[string]string),
		enabled:     make(map[string]bool),
		backupCodes: make(map[string][]string),
	}
}

// GenerateSecret generates a new TOTP secret for a user.
func (tm *TOTPManager) GenerateSecret(userID string) (string, error) {
	secret, err := tm.verifier.GenerateSecret()
	if err != nil {
		return "", err
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.secrets[userID] = secret
	tm.enabled[userID] = false // Not enabled until verified

	return secret, nil
}

// VerifyAndEnable verifies a code and enables 2FA for a user.
func (tm *TOTPManager) VerifyAndEnable(userID, code string) bool {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	secret, exists := tm.secrets[userID]
	if !exists {
		return false
	}

	if tm.verifier.Validate(secret, code, 1) {
		tm.enabled[userID] = true
		// Generate backup codes
		tm.backupCodes[userID] = generateBackupCodes(10)
		return true
	}
	return false
}

// ValidateCode validates a TOTP code for a user.
func (tm *TOTPManager) ValidateCode(userID, code string) bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	secret, exists := tm.secrets[userID]
	if !exists || !tm.enabled[userID] {
		return false
	}

	return tm.verifier.Validate(secret, code, 1)
}

// IsEnabled checks if 2FA is enabled for a user.
func (tm *TOTPManager) IsEnabled(userID string) bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.enabled[userID]
}

// Disable disables 2FA for a user.
func (tm *TOTPManager) Disable(userID string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.enabled[userID] = false
	delete(tm.secrets, userID)
	delete(tm.backupCodes, userID)
}

// GetSecret returns the stored secret (for provisioning URI).
func (tm *TOTPManager) GetSecret(userID string) (string, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	secret, exists := tm.secrets[userID]
	return secret, exists
}

// GetBackupCodes returns remaining backup codes.
func (tm *TOTPManager) GetBackupCodes(userID string) []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	codes := tm.backupCodes[userID]
	result := make([]string, len(codes))
	copy(result, codes)
	return result
}

// UseBackupCode uses a backup code.
func (tm *TOTPManager) UseBackupCode(userID, code string) bool {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	codes := tm.backupCodes[userID]
	for i, c := range codes {
		if security.ValidateCSRFToken(c, code) {
			// Remove used code
			tm.backupCodes[userID] = append(codes[:i], codes[i+1:]...)
			return true
		}
	}
	return false
}

func generateBackupCodes(count int) []string {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		b := make([]byte, 4)
		rand.Read(b)
		codes[i] = hex.EncodeToString(b) + "-" + hex.EncodeToString(b)
	}
	return codes
}

// ============================================================================
// Auth Middleware with Token Blacklist & Session Management
// ============================================================================

// AuthMiddleware creates a Gin middleware for JWT authentication.
func AuthMiddleware(cfg JWTConfig, blacklist *TokenBlacklist, sessionManager *SessionManager) gin.HandlerFunc {
	if blacklist == nil {
		blacklist = &TokenBlacklist{}
	}
	if sessionManager == nil {
		// Session checking disabled
	}
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header required",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header format",
			})
			return
		}

		claims, err := ValidateToken(parts[1], cfg.AccessTokenSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			return
		}

		if claims.TokenType != "access" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token type",
			})
			return
		}

		// Check token blacklist
		if blacklist != nil && claims.TokenID != "" {
			if blacklist.IsBlacklisted(claims.TokenID) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "token has been revoked",
				})
				return
			}
		}

		// Check session validity
		if sessionManager != nil && claims.SessionID != "" {
			session := sessionManager.GetSession(claims.SessionID)
			if session == nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "session has been revoked or expired",
				})
				return
			}
			sessionManager.UpdateLastActivity(claims.SessionID)
		}

		// Set context values
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyUsername, claims.Username)
		c.Set(ContextKeyRole, claims.Role)
		c.Set(ContextKeyToken, parts[1])
		c.Set(ContextKeySessionID, claims.SessionID)
		c.Set(ContextKeyTokenID, claims.TokenID)

		c.Next()
	}
}

// RoleMiddleware creates a Gin middleware for role-based access control.
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(ContextKeyRole)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "role not found in context",
			})
			return
		}

		roleStr, ok := role.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "invalid role type",
			})
			return
		}

		for _, allowed := range allowedRoles {
			if roleStr == allowed {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "insufficient permissions",
		})
	}
}

// RequestIDMiddleware adds a unique request ID to each request.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := generateRequestID()
		c.Set("requestID", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// CSRFTokenMiddleware adds CSRF protection.
func CSRFTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip for safe methods
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Check CSRF token for mutating methods
		csrfHeader := c.GetHeader("X-CSRF-Token")
		csrfCookie, err := c.Cookie("csrf_token")
		if err != nil || csrfHeader == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "CSRF token missing",
			})
			return
		}

		if !security.ValidateCSRFToken(csrfHeader, csrfCookie) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "CSRF token invalid",
			})
			return
		}

		c.Next()
	}
}

// IPWhitelistMiddleware restricts access to whitelisted IPs.
func IPWhitelistMiddleware(whitelistedIPs ...string) gin.HandlerFunc {
	whitelist := make(map[string]bool)
	for _, ip := range whitelistedIPs {
		whitelist[ip] = true
	}

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		if !whitelist[clientIP] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "IP not allowed",
			})
			return
		}
		c.Next()
	}
}

// LoggerMiddleware creates a structured request logger.
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		logEntry := map[string]interface{}{
			"status":     status,
			"latency":    latency.String(),
			"client_ip":  clientIP,
			"method":     method,
			"path":       path,
			"query":      query,
			"request_id": c.GetString("requestID"),
		}

		if status >= 500 {
			fmt.Printf("[ERROR] %+v\n", logEntry)
		} else if status >= 400 {
			fmt.Printf("[WARN] %+v\n", logEntry)
		} else {
			fmt.Printf("[INFO] %+v\n", logEntry)
		}
	}
}

// CORSMiddleware handles CORS headers using ALLOWED_ORIGINS env var.
// If ALLOWED_ORIGINS is empty, requests with credentials are not allowed.
func CORSMiddleware() gin.HandlerFunc {
	allowedOrigins := getEnv("ALLOWED_ORIGINS", "")
	originMap := make(map[string]bool)
	if allowedOrigins != "" {
		for _, o := range strings.Split(allowedOrigins, ",") {
			originMap[strings.TrimSpace(o)] = true
		}
	}
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if originMap[origin] {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		} else if len(originMap) == 0 && origin == "" {
			// No specific origins configured and no origin header — allow all origins
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// Helper functions
func generateRequestID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateRandomID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ============================================================================
// Refresh Token Rotation (one-time use)
// ============================================================================

// RefreshTokenStore manages refresh token rotation for one-time use.
type RefreshTokenStore struct {
	mu       sync.RWMutex
	tokens   map[string]string // tokenID -> userID
	used     map[string]bool   // tokenID -> used
	rotated  map[string]string // old tokenID -> new tokenID
}

// NewRefreshTokenStore creates a new refresh token store.
func NewRefreshTokenStore() *RefreshTokenStore {
	return &RefreshTokenStore{
		tokens:  make(map[string]string),
		used:    make(map[string]bool),
		rotated: make(map[string]string),
	}
}

// Store saves a refresh token.
func (rts *RefreshTokenStore) Store(tokenID, userID string) {
	rts.mu.Lock()
	defer rts.mu.Unlock()
	rts.tokens[tokenID] = userID
}

// ValidateAndRotate validates a refresh token and marks it as used (one-time).
// Returns the new token ID if valid, empty string otherwise.
func (rts *RefreshTokenStore) ValidateAndRotate(oldTokenID, userID string) bool {
	rts.mu.Lock()
	defer rts.mu.Unlock()

	// Check if token exists and belongs to user
	if storedUser, exists := rts.tokens[oldTokenID]; !exists || storedUser != userID {
		return false
	}

	// Check if already used
	if rts.used[oldTokenID] {
		// Token reuse detected - revoke all tokens for this user
		delete(rts.tokens, oldTokenID)
		return false
	}

	// Mark as used
	rts.used[oldTokenID] = true

	// Clean up old mapped entry
	delete(rts.tokens, oldTokenID)

	return true
}

// RevokeAllUserTokens revokes all refresh tokens for a user.
func (rts *RefreshTokenStore) RevokeAllUserTokens(userID string) {
	rts.mu.Lock()
	defer rts.mu.Unlock()

	for tokenID, uid := range rts.tokens {
		if uid == userID {
			rts.used[tokenID] = true
			delete(rts.tokens, tokenID)
		}
	}
}

// ============================================================================
// Request Validation Middleware
// ============================================================================

// RequestValidationConfig holds configuration for request validation.
type RequestValidationConfig struct {
	MaxBodySize      int64
	AllowedContentTypes []string
}

// DefaultRequestValidationConfig returns the default validation config.
func DefaultRequestValidationConfig() RequestValidationConfig {
	return RequestValidationConfig{
		MaxBodySize: 1 * 1024 * 1024, // 1MB
		AllowedContentTypes: []string{
			"application/json",
			"application/x-www-form-urlencoded",
			"multipart/form-data",
			"application/octet-stream",
		},
	}
}

// RequestValidationMiddleware creates a Gin middleware for request validation.
func RequestValidationMiddleware(cfg RequestValidationConfig) gin.HandlerFunc {
	allowedTypes := make(map[string]bool)
	for _, t := range cfg.AllowedContentTypes {
		allowedTypes[t] = true
	}

	return func(c *gin.Context) {
		// Validate Content-Type for mutating methods
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			contentType := c.GetHeader("Content-Type")
			// Extract base content type (before ;)
			if idx := strings.Index(contentType, ";"); idx > 0 {
				contentType = contentType[:idx]
			}
			contentType = strings.TrimSpace(contentType)

			if contentType != "" && !allowedTypes[contentType] {
				c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, gin.H{
					"error":      "unsupported content type",
					"code":       "UNSUPPORTED_MEDIA_TYPE",
					"allowed":    cfg.AllowedContentTypes,
				})
				return
			}
		}

		// Validate body size
		if cfg.MaxBodySize > 0 && c.Request.ContentLength > cfg.MaxBodySize {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "request body too large",
				"code":  "BODY_TOO_LARGE",
				"limit": cfg.MaxBodySize,
			})
			return
		}

		c.Next()
	}
}

// DeviceFingerprint extracts device fingerprint from request headers.
func DeviceFingerprint(c *gin.Context) string {
	fp := c.GetHeader("X-Device-Fingerprint")
	if fp == "" {
		// Fall back to User-Agent
		ua := c.GetHeader("User-Agent")
		if ua == "" {
			fp = c.ClientIP()
		} else {
			// Hash the user agent to create a fingerprint
			h := sha256.Sum256([]byte(ua))
			fp = hex.EncodeToString(h[:16])
		}
	}
	return fp
}

// getEnvOrPanic returns the value of an environment variable or panics if unset.
func getEnvOrPanic(key string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	panic(fmt.Sprintf("required environment variable %q is not set", key))
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
