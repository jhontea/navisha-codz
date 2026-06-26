package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "coding-challange/docs" // Generated docs package
	"coding-challange/internal/handler"
	"coding-challange/internal/repository"
	"coding-challange/internal/service"
)

// @title           Coding Challenge API
// @version         1.1.0
// @description     API for the Coding Challenge platform — list problems, get details, validate code, and retrieve hints.
// @termsOfService  https://example.com/terms

// @contact.name   API Support
// @contact.url    https://example.com/support
// @contact.email  support@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:9100
// @BasePath  /

// @schemes   http https
func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Initialize repository, services, and handler
	repo, err := repository.NewProblemRepository("problems")
	if err != nil {
		log.Fatalf("failed to load problems: %v", err)
	}

	problemSvc := service.NewProblemService(repo)
	hintSvc := service.NewHintService()
	h := handler.NewProblemHandler(problemSvc, hintSvc)

	// Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Routes
	router.GET("/health", h.HealthCheck)

	api := router.Group("/api")
	{
		api.GET("/problems", h.ListProblems)
		api.GET("/problems/:id", h.GetProblem)
		api.GET("/problems/:id/template", h.GetTemplate)
		api.POST("/validate", h.ValidateCode)
		api.GET("/problems/:id/hints", h.GetHints)
	}

	srv := &http.Server{
		Addr:         ":9100",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Server starting on :9100")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}
