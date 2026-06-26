package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"coding-challange/pkg/database"
	"coding-challange/pkg/middleware"
	"coding-challange/pkg/redis"
)

// Service configuration.
const (
	ServiceName = "auth-service"
	ServicePort = "9101"
)

// User model.
type User struct {
	ID           string    `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Role         string    `json:"role" db:"role"`
	TotalScore   int       `json:"total_score" db:"total_score"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// RefreshToken model.
type RefreshToken struct {
	ID         string    `db:"id"`
	UserID     string    `db:"user_id"`
	TokenHash  string    `db:"token_hash"`
	DeviceInfo string    `db:"device_info"`
	ExpiresAt  time.Time `db:"expires_at"`
	IsRevoked  bool      `db:"is_revoked"`
	CreatedAt  time.Time `db:"created_at"`
}

// Request/Response models.
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int64     `json:"expires_in"`
	User         *UserInfo `json:"user"`
}

type UserInfo struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	TotalScore int    `json:"total_score"`
}

// Server holds all dependencies.
type Server struct {
	router     *gin.Engine
	db         *database.Pool
	redis      *redis.Client
	jwtConfig  middleware.JWTConfig
	bcryptCost int
}

func main() {
	log.Printf("Starting %s...", ServiceName)

	// Initialize dependencies
	jwtConfig := middleware.NewJWTConfig()
	dbPool, err := database.NewFromEnv()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbPool.Close()

	redisClient, err := redis.NewFromEnv()
	if err != nil {
		log.Printf("Warning: Redis not available: %v", err)
	}

	// Create server
	srv := &Server{
		router:     gin.New(),
		db:         dbPool,
		redis:      redisClient,
		jwtConfig:  jwtConfig,
		bcryptCost: bcrypt.DefaultCost,
	}

	// Setup middleware
	srv.router.Use(gin.Recovery())
	srv.router.Use(middleware.RequestIDMiddleware())
	srv.router.Use(middleware.LoggerMiddleware())
	srv.router.Use(middleware.CORSMiddleware())
	srv.router.Use(middleware.RequestValidationMiddleware(middleware.DefaultRequestValidationConfig()))

	// Register routes
	srv.setupRoutes()

	// Start server with graceful shutdown
	port := getEnv("PORT", ServicePort)
	runServer(srv, port)
}

func (s *Server) setupRoutes() {
	// Health check
	s.router.GET("/health", s.healthCheck)

	// Public routes
	auth := s.router.Group("/auth")
	{
		auth.POST("/register", s.register)
		auth.POST("/login", s.login)
		auth.POST("/refresh", s.refreshToken)
	}

	// Protected routes
	protected := s.router.Group("/auth")
	protected.Use(middleware.AuthMiddleware(s.jwtConfig, nil, nil))
	{
		protected.POST("/logout", s.logout)
		protected.GET("/me", s.getCurrentUser)
	}
}

// healthCheck handles GET /health.
func (s *Server) healthCheck(c *gin.Context) {
	ctx := c.Request.Context()
	status := "ok"
	dbStatus := "ok"

	if err := s.db.HealthCheck(ctx); err != nil {
		dbStatus = "unavailable"
		status = "degraded"
	}

	serviceName := ServiceName
	_ = serviceName

	c.JSON(http.StatusOK, gin.H{
		"service": serviceName,
		"status":  status,
		"database": dbStatus,
		"time":    time.Now().UTC(),
	})
}

// register handles POST /auth/register.
func (s *Server) register(c *gin.Context) {
	ctx := c.Request.Context()

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	// Check if username exists
	exists, err := s.usernameExists(ctx, req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "username already taken"})
		return
	}

	// Check if email exists
	exists, err = s.emailExists(ctx, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.bcryptCost)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Insert user
	var userID string
	err = s.db.QueryRow(ctx, `
		INSERT INTO users (username, email, password_hash, role)
		VALUES ($1, $2, $3, 'user')
		RETURNING id
	`, req.Username, req.Email, string(hash)).Scan(&userID)
	if err != nil {
		log.Printf("Failed to insert user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	// Generate tokens
	accessToken, refreshToken, _, _, err := middleware.GenerateTokenPair(
		s.jwtConfig, userID, req.Username, "user",
	)
	if err != nil {
		log.Printf("Failed to generate tokens: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	// Store refresh token hash
	refreshHash, _ := bcrypt.GenerateFromPassword([]byte(refreshToken), s.bcryptCost)
	_, err = s.db.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, string(refreshHash), time.Now().Add(s.jwtConfig.RefreshTokenTTL))
	if err != nil {
		log.Printf("Failed to store refresh token: %v", err)
	}

	c.JSON(http.StatusCreated, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtConfig.AccessTokenTTL.Seconds()),
		User: &UserInfo{
			ID:       userID,
			Username: req.Username,
			Email:    req.Email,
			Role:     "user",
		},
	})
}

// login handles POST /auth/login.
func (s *Server) login(c *gin.Context) {
	ctx := c.Request.Context()

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	// Get user by username
	var user struct {
		ID           string
		Username     string
		Email        string
		PasswordHash string
		Role         string
		TotalScore   int
	}
	err := s.db.QueryRow(ctx, `
		SELECT id, username, email, password_hash, role, total_score
		FROM users WHERE username = $1
	`, req.Username).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.TotalScore)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Generate tokens
	accessToken, refreshToken, _, _, err := middleware.GenerateTokenPair(
		s.jwtConfig, user.ID, user.Username, user.Role,
	)
	if err != nil {
		log.Printf("Failed to generate tokens: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	// Store refresh token
	refreshHash, _ := bcrypt.GenerateFromPassword([]byte(refreshToken), s.bcryptCost)
	_, err = s.db.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, user.ID, string(refreshHash), time.Now().Add(s.jwtConfig.RefreshTokenTTL))
	if err != nil {
		log.Printf("Failed to store refresh token: %v", err)
	}

	// Update last active
	_, _ = s.db.Exec(ctx, "UPDATE users SET last_active_at = NOW() WHERE id = $1", user.ID)

	c.JSON(http.StatusOK, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.jwtConfig.AccessTokenTTL.Seconds()),
		User: &UserInfo{
			ID:         user.ID,
			Username:   user.Username,
			Email:      user.Email,
			Role:       user.Role,
			TotalScore: user.TotalScore,
		},
	})
}

// refreshToken handles POST /auth/refresh.
func (s *Server) refreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Validate refresh token
	claims, err := middleware.ValidateToken(req.RefreshToken, s.jwtConfig.RefreshTokenSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	if claims.TokenType != "refresh" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token type"})
		return
	}

	// Check if refresh token is revoked in DB
	var isRevoked bool
	err = s.db.QueryRow(c.Request.Context(),
		"SELECT is_revoked FROM refresh_tokens WHERE user_id = $1 AND expires_at > NOW() ORDER BY created_at DESC LIMIT 1",
		claims.UserID,
	).Scan(&isRevoked)
	if err != nil || isRevoked {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token revoked"})
		return
	}

	// Get user info
	var username, role string
	err = s.db.QueryRow(c.Request.Context(),
		"SELECT username, role FROM users WHERE id = $1", claims.UserID,
	).Scan(&username, &role)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	// Generate new token pair
	accessToken, newRefreshToken, _, _, err := middleware.GenerateTokenPair(
		s.jwtConfig, claims.UserID, username, role,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
		"expires_in":    int64(s.jwtConfig.AccessTokenTTL.Seconds()),
	})
}

// logout handles POST /auth/logout.
func (s *Server) logout(c *gin.Context) {
	userIDRaw, _ := c.Get(middleware.ContextKeyUserID)
	userID, ok := userIDRaw.(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Revoke all refresh tokens for user
	_, err := s.db.Exec(c.Request.Context(),
		"UPDATE refresh_tokens SET is_revoked = TRUE WHERE user_id = $1", userID,
	)
	if err != nil {
		log.Printf("Failed to revoke tokens: %v", err)
	}

	// Blacklist current access token in Redis
	if s.redis != nil {
		tokenRaw, _ := c.Get(middleware.ContextKeyToken)
		token, ok := tokenRaw.(string)
		if ok && token != "" {
			_ = s.redis.Set(c.Request.Context(),
				fmt.Sprintf("blacklist:%s", token),
				"revoked",
				s.jwtConfig.AccessTokenTTL,
			)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

// getCurrentUser handles GET /auth/me.
func (s *Server) getCurrentUser(c *gin.Context) {
	userIDRaw, _ := c.Get(middleware.ContextKeyUserID)
	userID, ok := userIDRaw.(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var user UserInfo
	err := s.db.QueryRow(c.Request.Context(),
		"SELECT id, username, email, role, total_score FROM users WHERE id = $1", userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.TotalScore)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Helper functions.
func (s *Server) usernameExists(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := s.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", username).Scan(&exists)
	return exists, err
}

func (s *Server) emailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := s.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
	return exists, err
}

func runServer(srv *Server, port string) {
	addr := ":" + port
	httpSrv := &http.Server{
		Addr:         addr,
		Handler:      srv.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("%s starting on%s", ServiceName, addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Printf("Shutting down %s...", ServiceName)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Printf("%s exited gracefully", ServiceName)
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}


