package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"coding-challange/pkg/database"
	"coding-challange/pkg/middleware"
	"coding-challange/pkg/redis"
)

// Service configuration.
const (
	ServiceName = "problem-service"
	ServicePort = "9102"
)

// Problem models.
type Problem struct {
	ID                 int       `json:"id" db:"id"`
	Title              string    `json:"title" db:"title"`
	Slug               string    `json:"slug" db:"slug"`
	Description        string    `json:"description" db:"description"`
	Difficulty         string    `json:"difficulty" db:"difficulty"`
	CategoryID         int       `json:"category_id" db:"category_id"`
	CategoryName       string    `json:"category_name,omitempty" db:"category_name"`
	TimeLimitSeconds   int       `json:"time_limit_seconds" db:"time_limit_seconds"`
	MemoryLimitMB      int       `json:"memory_limit_mb" db:"memory_limit_mb"`
	MaxScore           int       `json:"max_score" db:"max_score"`
	FunctionName       string    `json:"function_name,omitempty" db:"function_name"`
	TemplateCode       string    `json:"template_code,omitempty" db:"template_code"`
	IsPublished       bool      `json:"is_published" db:"is_published"`
	Tags               []string  `json:"tags,omitempty"`
	TotalSubmissions   int       `json:"total_submissions,omitempty"`
	AcceptedSubmissions int     `json:"accepted_submissions,omitempty"`
	SuccessRate        float64   `json:"success_rate,omitempty"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

type TestCase struct {
	ID          int    `json:"id"`
	ProblemID   int    `json:"problem_id"`
	Input       string `json:"input"`
	Expected    string `json:"expected_output"`
	Description string `json:"description,omitempty"`
	IsSample    bool   `json:"is_sample"`
}

type ProblemFilters struct {
	Difficulty string `form:"difficulty"`
	CategoryID int    `form:"category_id"`
	Limit      int    `form:"limit"`
	Cursor     string `form:"cursor"`
	Search     string `form:"search"`
}

type CursorResponse struct {
	Data       interface{} `json:"data"`
	NextCursor string      `json:"next_cursor,omitempty"`
	HasMore    bool        `json:"has_more"`
	Total      int         `json:"total"`
}

// Server holds all dependencies.
type Server struct {
	router    *gin.Engine
	db        *database.Pool
	redis     *redis.Client
	jwtConfig middleware.JWTConfig
}

func main() {
	log.Printf("Starting %s...", ServiceName)

	// Initialize dependencies
	dbPool, err := database.NewFromEnv()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbPool.Close()

	redisClient, err := redis.NewFromEnv()
	if err != nil {
		log.Printf("Warning: Redis not available: %v", err)
	}

	srv := &Server{
		router:    gin.New(),
		db:        dbPool,
		redis:     redisClient,
		jwtConfig: middleware.NewJWTConfig(),
	}

	// Setup middleware
	srv.router.Use(gin.Recovery())
	srv.router.Use(middleware.RequestIDMiddleware())
	srv.router.Use(middleware.LoggerMiddleware())
	srv.router.Use(middleware.CORSMiddleware())
	srv.router.Use(middleware.GzipMiddleware())
	srv.router.Use(middleware.RequestValidationMiddleware(middleware.DefaultRequestValidationConfig()))

	// Register routes
	srv.setupRoutes()

	// Cache warming: pre-populate first page of problem list
	warmCtx, warmCancel := context.WithCancel(context.Background())
	if redisClient != nil {
		go srv.warmCache(warmCtx)
	}

	// Start server
	port := getEnv("PORT", ServicePort)
	runServer(srv, port)

	// Stop cache warming goroutine on shutdown
	warmCancel()
}

func (s *Server) setupRoutes() {
	s.router.GET("/health", s.healthCheck)

	api := s.router.Group("/api")
	{
		// Public routes
		api.GET("/problems", s.listProblems)
		api.GET("/problems/:id", s.getProblem)
		api.GET("/problems/:id/test-cases", s.getTestCases)

		// Admin routes
		admin := api.Group("/problems")
		admin.Use(middleware.AuthMiddleware(s.jwtConfig, nil, nil))
		admin.Use(middleware.RoleMiddleware("admin"))
		{
			admin.POST("", s.createProblem)
			admin.PUT("/:id", s.updateProblem)
		}
	}
}

func (s *Server) healthCheck(c *gin.Context) {
	ctx := c.Request.Context()
	status := "ok"
	dbStatus := "ok"

	if err := s.db.HealthCheck(ctx); err != nil {
		dbStatus = "unavailable"
		status = "degraded"
	}

	c.JSON(http.StatusOK, gin.H{
		"service": ServiceName,
		"status":  status,
		"database": dbStatus,
		"time":    time.Now().UTC(),
	})
}

// warmCache pre-populates the Redis cache with common problem list queries at startup.
func (s *Server) warmCache(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[RECOVER] warmCache panic: %v", r)
		}
	}()
	log.Println("[CACHE WARM] Starting cache warming for problem lists...")

	select {
	case <-ctx.Done():
		log.Println("[CACHE WARM] Context cancelled, skipping")
		return
	default:
	}

	// Warm the default first page (no filters)
	defaultCacheKey := fmt.Sprintf("problems:list::0::20:")
	if err := s.redis.Get(ctx, defaultCacheKey, &CursorResponse{}); err != nil {
		// Manually execute query to populate cache
		query := `
			SELECT 
				p.id, p.title, p.slug, p.difficulty, p.max_score,
				pc.name AS category_name,
				COALESCE(sub_stats.total_submissions, 0) AS total_submissions,
				COALESCE(sub_stats.accepted_submissions, 0) AS accepted_submissions,
				CASE WHEN COALESCE(sub_stats.total_submissions, 0) > 0 
					THEN ROUND(sub_stats.accepted_submissions * 100.0 / sub_stats.total_submissions, 2)
					ELSE 0 
				END AS success_rate
			FROM problems p
			JOIN problem_categories pc ON p.category_id = pc.id
			LEFT JOIN (
				SELECT problem_id,
					COUNT(*) AS total_submissions,
					COUNT(*) FILTER (WHERE status = 'accepted') AS accepted_submissions
				FROM submissions
				GROUP BY problem_id
			) sub_stats ON p.id = sub_stats.problem_id
			WHERE p.is_published = TRUE
			ORDER BY p.difficulty, p.title
			LIMIT 21
		`
		rows, err := s.db.Query(ctx, query)
		if err != nil {
			log.Printf("[CACHE WARM] Query failed: %v", err)
			return
		}
		defer rows.Close()

		var problems []Problem
		for rows.Next() {
			var p Problem
			if err := rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Difficulty, &p.MaxScore,
				&p.CategoryName, &p.TotalSubmissions, &p.AcceptedSubmissions, &p.SuccessRate); err != nil {
				continue
			}
			problems = append(problems, p)
		}

		hasMore := len(problems) > 20
		nextCursor := ""
		if hasMore {
			nextCursor = fmt.Sprintf("%d", problems[19].ID)
			problems = problems[:20]
		}

		var totalCount int
		_ = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM problems WHERE is_published = TRUE`).Scan(&totalCount)

		response := CursorResponse{
			Data:       problems,
			NextCursor: nextCursor,
			HasMore:    hasMore,
			Total:      totalCount,
		}

		if err := s.redis.Set(ctx, defaultCacheKey, response, 5*time.Minute); err != nil {
			log.Printf("[CACHE WARM] Failed to cache default list: %v", err)
		} else {
			log.Printf("[CACHE WARM] Cached default problem list (%d problems)", len(problems))
		}
	}

	// Warm by difficulty
	difficulties := []string{"easy", "medium", "hard"}
	for _, diff := range difficulties {
		cacheKey := fmt.Sprintf("problems:list:%s:0::20:", diff)
		if err := s.redis.Get(ctx, cacheKey, &CursorResponse{}); err != nil {
			_ = s.redis.Set(ctx, cacheKey, CursorResponse{Data: []Problem{}, Total: 0}, 5*time.Minute)
			log.Printf("[CACHE WARM] Seeded cache key: %s", cacheKey)
		}
	}

	log.Println("[CACHE WARM] Cache warming completed")
}

// listProblems handles GET /api/problems with filters and cursor-based pagination.
func (s *Server) listProblems(c *gin.Context) {
	ctx := c.Request.Context()

	var filters ProblemFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filters"})
		return
	}

	// Set defaults
	if filters.Limit < 1 || filters.Limit > 100 {
		filters.Limit = 20
	}

	// Try cache first
	cacheKey := fmt.Sprintf("problems:list:%s:%d:%s:%d:%s",
		filters.Difficulty, filters.CategoryID, filters.Search, filters.Limit, filters.Cursor)
	if s.redis != nil {
		var cached CursorResponse
		if err := s.redis.Get(ctx, cacheKey, &cached); err == nil {
			c.JSON(http.StatusOK, cached)
			return
		}
	}

	// Build cursor-based query (keyset pagination instead of OFFSET)
	countQuery := `
		SELECT COUNT(*) FROM (
			SELECT p.id
			FROM problems p
			JOIN problem_categories pc ON p.category_id = pc.id
			WHERE p.is_published = TRUE
	`
	query := `
		SELECT 
			p.id, p.title, p.slug, p.difficulty, p.max_score,
			pc.name AS category_name,
			COALESCE(sub_stats.total_submissions, 0) AS total_submissions,
			COALESCE(sub_stats.accepted_submissions, 0) AS accepted_submissions,
			CASE WHEN COALESCE(sub_stats.total_submissions, 0) > 0 
				THEN ROUND(sub_stats.accepted_submissions * 100.0 / sub_stats.total_submissions, 2)
				ELSE 0 
			END AS success_rate
		FROM problems p
		JOIN problem_categories pc ON p.category_id = pc.id
		LEFT JOIN (
			SELECT problem_id,
				COUNT(*) AS total_submissions,
				COUNT(*) FILTER (WHERE status = 'accepted') AS accepted_submissions
			FROM submissions
			GROUP BY problem_id
		) sub_stats ON p.id = sub_stats.problem_id
		WHERE p.is_published = TRUE
	`
	args := []interface{}{}
	argIdx := 1

	if filters.Difficulty != "" {
		clause := fmt.Sprintf(" AND p.difficulty = $%d", argIdx)
		query += clause
		countQuery += clause
		args = append(args, filters.Difficulty)
		argIdx++
	}
	if filters.CategoryID > 0 {
		clause := fmt.Sprintf(" AND p.category_id = $%d", argIdx)
		query += clause
		countQuery += clause
		args = append(args, filters.CategoryID)
		argIdx++
	}
	if filters.Search != "" {
		clause := fmt.Sprintf(" AND (p.title ILIKE $%d OR p.slug ILIKE $%d)", argIdx, argIdx)
		query += clause
		countQuery += clause
		args = append(args, "%"+filters.Search+"%")
		argIdx++
	}

	countQuery += " ) cnt"

	// Add cursor condition (keyset pagination) for the query only
	if filters.Cursor != "" {
		query += fmt.Sprintf(" AND (p.difficulty, p.title) > (SELECT p2.difficulty, p2.title FROM problems p2 WHERE p2.id = $%d)", argIdx)
		args = append(args, filters.Cursor)
		argIdx++
	}

	query += " ORDER BY p.difficulty, p.title"
	query += fmt.Sprintf(" LIMIT $%d", argIdx)
	args = append(args, filters.Limit+1) // Fetch one extra to detect HasMore

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		log.Printf("Failed to query problems: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	defer rows.Close()

	var problems []Problem
	nextCursor := ""
	for rows.Next() {
		var p Problem
		if err := rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Difficulty, &p.MaxScore,
			&p.CategoryName, &p.TotalSubmissions, &p.AcceptedSubmissions, &p.SuccessRate); err != nil {
			log.Printf("Failed to scan problem: %v", err)
			continue
		}
		problems = append(problems, p)
	}

	hasMore := len(problems) > filters.Limit
	if hasMore {
		// Remove the extra item
		lastItem := problems[len(problems)-2]
		nextCursor = fmt.Sprintf("%d", lastItem.ID)
		problems = problems[:filters.Limit]
	}

	// Get total count
	var totalCount int
	_ = s.db.QueryRow(ctx, countQuery, args[:len(args)-1]...).Scan(&totalCount)

	response := CursorResponse{
		Data:       problems,
		NextCursor: nextCursor,
		HasMore:    hasMore,
		Total:      totalCount,
	}

	// Cache response with TTL 5 minutes
	if s.redis != nil {
		_ = s.redis.Set(ctx, cacheKey, response, 5*time.Minute)
	}

	c.JSON(http.StatusOK, response)
}

// getProblem handles GET /api/problems/:id.
func (s *Server) getProblem(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		// Try by slug
		s.getProblemBySlug(c, idStr)
		return
	}

	// Try cache
	cacheKey := fmt.Sprintf("problem:%d", id)
	if s.redis != nil {
		var cached Problem
		if err := s.redis.Get(ctx, cacheKey, &cached); err == nil {
			c.JSON(http.StatusOK, cached)
			return
		}
	}

	var p Problem
	var tagsArray *string
	err = s.db.QueryRow(ctx, `
		SELECT 
			p.id, p.title, p.slug, p.description, p.difficulty,
			p.category_id, pc.name AS category_name,
			p.time_limit_seconds, p.memory_limit_mb, p.max_score,
			p.function_name, p.template_code, p.is_published,
			p.created_at, p.updated_at,
			ARRAY_AGG(pt.tag_name) FILTER (WHERE pt.tag_name IS NOT NULL)
		FROM problems p
		JOIN problem_categories pc ON p.category_id = pc.id
		LEFT JOIN problem_tags pt ON p.id = pt.problem_id
		WHERE p.id = $1 AND p.is_published = TRUE
		GROUP BY p.id, pc.name
	`, id).Scan(&p.ID, &p.Title, &p.Slug, &p.Description, &p.Difficulty,
		&p.CategoryID, &p.CategoryName, &p.TimeLimitSeconds, &p.MemoryLimitMB,
		&p.MaxScore, &p.FunctionName, &p.TemplateCode, &p.IsPublished,
		&p.CreatedAt, &p.UpdatedAt, &tagsArray)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "problem not found"})
		return
	}

	if tagsArray != nil {
		// Parse array (simplified)
	}

	// Cache
	if s.redis != nil {
		_ = s.redis.Set(ctx, cacheKey, p, redis.TTLProblem)
	}

	c.JSON(http.StatusOK, p)
}

func (s *Server) getProblemBySlug(c *gin.Context, slug string) {
	ctx := c.Request.Context()

	cacheKey := fmt.Sprintf("problem:%s", slug)
	if s.redis != nil {
		var cached Problem
		if err := s.redis.Get(ctx, cacheKey, &cached); err == nil {
			c.JSON(http.StatusOK, cached)
			return
		}
	}

	var p Problem
	err := s.db.QueryRow(ctx, `
		SELECT 
			p.id, p.title, p.slug, p.description, p.difficulty,
			p.category_id, pc.name AS category_name,
			p.time_limit_seconds, p.memory_limit_mb, p.max_score,
			p.function_name, p.template_code, p.is_published,
			p.created_at, p.updated_at
		FROM problems p
		JOIN problem_categories pc ON p.category_id = pc.id
		WHERE p.slug = $1 AND p.is_published = TRUE
	`, slug).Scan(&p.ID, &p.Title, &p.Slug, &p.Description, &p.Difficulty,
		&p.CategoryID, &p.CategoryName, &p.TimeLimitSeconds, &p.MemoryLimitMB,
		&p.MaxScore, &p.FunctionName, &p.TemplateCode, &p.IsPublished,
		&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "problem not found"})
		return
	}

	if s.redis != nil {
		_ = s.redis.Set(ctx, cacheKey, p, redis.TTLProblem)
	}

	c.JSON(http.StatusOK, p)
}

// getTestCases handles GET /api/problems/:id/test-cases (sample only).
func (s *Server) getTestCases(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid problem id"})
		return
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, problem_id, input, expected_output, description, is_sample
		FROM test_cases
		WHERE problem_id = $1 AND is_sample = TRUE AND is_hidden = FALSE
		ORDER BY id
	`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	defer rows.Close()

	var testCases []TestCase
	for rows.Next() {
		var tc TestCase
		if err := rows.Scan(&tc.ID, &tc.ProblemID, &tc.Input, &tc.Expected, &tc.Description, &tc.IsSample); err != nil {
			continue
		}
		testCases = append(testCases, tc)
	}

	c.JSON(http.StatusOK, testCases)
}

// createProblem handles POST /api/problems (admin).
func (s *Server) createProblem(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		Title            string   `json:"title" binding:"required"`
		Slug             string   `json:"slug" binding:"required"`
		Description      string   `json:"description" binding:"required"`
		Difficulty       string   `json:"difficulty" binding:"required,oneof=easy medium hard"`
		CategoryID       int      `json:"category_id" binding:"required"`
		TimeLimitSeconds int      `json:"time_limit_seconds"`
		MemoryLimitMB    int      `json:"memory_limit_mb"`
		MaxScore         int      `json:"max_score"`
		FunctionName     string   `json:"function_name"`
		TemplateCode     string   `json:"template_code"`
		Tags             []string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	// Set defaults
	if req.TimeLimitSeconds == 0 {
		req.TimeLimitSeconds = 1
	}
	if req.MemoryLimitMB == 0 {
		req.MemoryLimitMB = 256
	}
	if req.MaxScore == 0 {
		req.MaxScore = 100
	}

	var problemID int
	tx, err := s.db.Begin(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "transaction error"})
		return
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO problems (title, slug, description, difficulty, category_id, 
			time_limit_seconds, memory_limit_mb, max_score, function_name, template_code, is_published)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, FALSE)
		RETURNING id
	`, req.Title, req.Slug, req.Description, req.Difficulty, req.CategoryID,
		req.TimeLimitSeconds, req.MemoryLimitMB, req.MaxScore, req.FunctionName, req.TemplateCode,
	).Scan(&problemID)
	if err != nil {
		log.Printf("Failed to create problem: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create problem"})
		return
	}

	// Insert tags
	for _, tag := range req.Tags {
		_, _ = tx.Exec(ctx, "INSERT INTO problem_tags (problem_id, tag_name) VALUES ($1, $2)", problemID, tag)
	}

	if err := tx.Commit(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit"})
		return
	}

	// Invalidate cache
	if s.redis != nil {
		_ = s.redis.DeletePattern(ctx, "problems:list:*")
	}

	c.JSON(http.StatusCreated, gin.H{"id": problemID, "message": "problem created"})
}

// updateProblem handles PUT /api/problems/:id (admin).
func (s *Server) updateProblem(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid problem id"})
		return
	}

	var req struct {
		Title            *string `json:"title"`
		Description      *string `json:"description"`
		Difficulty       *string `json:"difficulty"`
		CategoryID       *int    `json:"category_id"`
		TimeLimitSeconds *int    `json:"time_limit_seconds"`
		MemoryLimitMB    *int    `json:"memory_limit_mb"`
		MaxScore         *int    `json:"max_score"`
		TemplateCode     *string `json:"template_code"`
		IsPublished      *bool   `json:"is_published"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Build dynamic update query
	updates := []interface{}{}
	setClauses := []string{}
	argIdx := 1

	if req.Title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argIdx))
		updates = append(updates, *req.Title)
		argIdx++
	}
	if req.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIdx))
		updates = append(updates, *req.Description)
		argIdx++
	}
	if req.Difficulty != nil {
		setClauses = append(setClauses, fmt.Sprintf("difficulty = $%d", argIdx))
		updates = append(updates, *req.Difficulty)
		argIdx++
	}
	if req.IsPublished != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_published = $%d", argIdx))
		updates = append(updates, *req.IsPublished)
		argIdx++
	}

	if len(setClauses) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	setClauses = append(setClauses, "updated_at = NOW()")
	updates = append(updates, id)

	query := fmt.Sprintf("UPDATE problems SET %s WHERE id = $%d",
		joinClauses(setClauses, ", "), argIdx)

	_, err = s.db.Exec(ctx, query, updates...)
	if err != nil {
		log.Printf("Failed to update problem: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}

	// Invalidate cache
	if s.redis != nil {
		_ = s.redis.Delete(ctx, fmt.Sprintf("problem:%d", id))
		_ = s.redis.DeletePattern(ctx, "problems:list:*")
	}

	c.JSON(http.StatusOK, gin.H{"message": "problem updated"})
}

func joinClauses(clauses []string, sep string) string {
	result := ""
	for i, c := range clauses {
		if i > 0 {
			result += sep
		}
		result += c
	}
	return result
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
