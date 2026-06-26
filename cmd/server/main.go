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

	"coding-challange/internal/config"
	"coding-challange/internal/handler"
	"coding-challange/internal/middleware"
	"coding-challange/internal/repository"
	"coding-challange/internal/service"
)

func main() {
	cfg := config.Load()

	// Set Gin mode
	if cfg.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize repository
	repo, err := repository.NewProblemRepository(cfg.ProblemsDir)
	if err != nil {
		log.Fatalf("Failed to load problems: %v", err)
	}
	log.Printf("Loaded %d problems from %s", repo.Count(), cfg.ProblemsDir)

	// Initialize services
	problemSvc := service.NewProblemService(repo)
	runnerSvc := service.NewRunnerService(cfg.DockerTimeout, cfg.MaxMemoryMB)
	hintSvc := service.NewHintService()

	// Initialize handler
	h := handler.NewProblemHandler(problemSvc, runnerSvc, hintSvc)

	// Initialize cache (5 minute TTL)
	cache := middleware.NewInMemoryCache(5 * time.Minute)

	// Setup router
	router := gin.Default()

	// Global middleware (order matters)
	router.Use(gin.Recovery())
	router.Use(middleware.SecurityHeaders())          // Iter 8: Security headers (CSP, HSTS, etc.)
	router.Use(middleware.MaxBodySize(64*1024))       // Iter 8: Max request body size
	router.Use(middleware.InputSanitization())        // Iter 8: Input sanitization
	router.Use(middleware.CacheMiddleware(cache))     // Iter 7: In-memory cache with ETag
	router.Use(corsMiddleware())

	// Static file serving for web/static/ with ETag support
	router.Static("/static", "./web/static")

	// Serve HTML templates at root
	router.StaticFile("/", "./web/templates/index.html")
	router.StaticFile("/problem", "./web/templates/problem.html")
	router.StaticFile("/index.html", "./web/templates/index.html")

	// Routes
	router.GET("/health", h.HealthCheck)

	api := router.Group("/api")
	{
		api.GET("/problems", h.ListProblems)
		api.GET("/problems/:id", h.GetProblem)
		api.GET("/problems/:id/template", h.GetTemplate)
		api.POST("/problems/:id/run", h.RunCode)
		api.POST("/validate", h.ValidateCode)
		api.GET("/problems/:id/hints", h.GetHints)
	}

	// Start server with graceful shutdown
	addr := fmt.Sprintf(":%s", cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on http://localhost%s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with 30s timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

// corsMiddleware adds CORS headers for development.
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
