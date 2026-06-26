package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"

	"coding-challange/pkg/database"
	"coding-challange/pkg/middleware"
	"coding-challange/pkg/rabbitmq"
	"coding-challange/pkg/redis"
	"coding-challange/pkg/websocket"
)

// Service configuration.
const (
	ServiceName = "execution-service"
	ServicePort = "9103"
)

// Submission models.
type Submission struct {
	ID                string     `json:"id" db:"id"`
	UserID            string     `json:"user_id" db:"user_id"`
	ProblemID         int        `json:"problem_id" db:"problem_id"`
	Code              string     `json:"code" db:"code"`
	Language          string     `json:"language" db:"language"`
	Status            string     `json:"status" db:"status"`
	Score             int        `json:"score" db:"score"`
	ExecutionTimeMs   *int       `json:"execution_time_ms,omitempty" db:"execution_time_ms"`
	MemoryUsedKb      *int       `json:"memory_used_kb,omitempty" db:"memory_used_kb"`
	TestCasesPassed   int        `json:"test_cases_passed" db:"test_cases_passed"`
	TestCasesTotal    int        `json:"test_cases_total" db:"test_cases_total"`
	ErrorMessage      *string    `json:"error_message,omitempty" db:"error_message"`
	SubmittedAt       time.Time  `json:"submitted_at" db:"submitted_at"`
	CompletedAt       *time.Time `json:"completed_at,omitempty" db:"completed_at"`
}

type CreateSubmissionRequest struct {
	ProblemID int    `json:"problem_id" binding:"required"`
	Code      string `json:"code" binding:"required"`
	Language  string `json:"language" binding:"required,oneof=go python javascript java cpp"`
}

type SubmissionResult struct {
	SubmissionID     string           `json:"submission_id"`
	Status           string           `json:"status"`
	Score            int              `json:"score"`
	ExecutionTimeMs  int              `json:"execution_time_ms"`
	MemoryUsedKb     int              `json:"memory_used_kb"`
	TestCasesPassed  int              `json:"test_cases_passed"`
	TestCasesTotal   int              `json:"test_cases_total"`
	Results          []TestCaseResult `json:"results,omitempty"`
	ErrorMessage     string           `json:"error_message,omitempty"`
}

type TestCaseResult struct {
	TestCaseID      int    `json:"test_case_id"`
	Status          string `json:"status"`
	ExecutionTimeMs int    `json:"execution_time_ms"`
	MemoryUsedKb    int    `json:"memory_used_kb"`
}

// Message for RabbitMQ.
type ExecutionMessage struct {
	SubmissionID string `json:"submission_id"`
	ProblemID    int    `json:"problem_id"`
	UserID       string `json:"user_id"`
	Code         string `json:"code"`
	Language     string `json:"language"`
	EnqueuedAt   string `json:"enqueued_at"`
	Priority     int    `json:"priority,omitempty"` // 0=free, 1=premium
	RetryCount   int    `json:"retry_count,omitempty"`
}

// QueueStats holds queue depth monitoring statistics.
type QueueStats struct {
	mu              sync.RWMutex
	SubmissionsQueued   int           `json:"submissions_queued"`
	SubmissionsProcessed int         `json:"submissions_processed"`
	SubmissionsFailed    int         `json:"submissions_failed"`
	SubmissionsRetried   int         `json:"submissions_retried"`
	DeadLetterCount      int         `json:"dead_letter_count"`
	PrioritySubmissions  int         `json:"priority_submissions"`
	NormalSubmissions    int         `json:"normal_submissions"`
	QueueDepth           int         `json:"queue_depth"`
	AvgQueueTime         time.Duration `json:"avg_queue_time"`
	LastDepthCheck       time.Time     `json:"last_depth_check"`
}

func (qs *QueueStats) recordQueued(priority int) {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	qs.SubmissionsQueued++
	if priority >= 1 {
		qs.PrioritySubmissions++
	} else {
		qs.NormalSubmissions++
	}
}

func (qs *QueueStats) recordProcessed() {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	qs.SubmissionsProcessed++
}

func (qs *QueueStats) recordFailed() {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	qs.SubmissionsFailed++
}

func (qs *QueueStats) recordRetried() {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	qs.SubmissionsRetried++
}

func (qs *QueueStats) recordDeadLetter() {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	qs.DeadLetterCount++
}

func (qs *QueueStats) updateDepth(depth int) {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	qs.QueueDepth = depth
	qs.LastDepthCheck = time.Now()
}

// Server holds all dependencies.
type Server struct {
	router         *gin.Engine
	db             *database.Pool
	redis          *redis.Client
	rabbitmq       *rabbitmq.Client
	wsHub          *websocket.Hub
	jwtConfig      middleware.JWTConfig
	blacklist      *middleware.TokenBlacklist
	sessionManager *middleware.SessionManager
	queueStats     *QueueStats
	dlqHandler     *DLQHandler
}

// DLQHandler handles dead letter queue processing with retry.
type DLQHandler struct {
	client      *rabbitmq.Client
	db          *database.Pool
	maxRetries  int
	stats       *QueueStats
}

// NewDLQHandler creates a new dead letter queue handler.
func NewDLQHandler(client *rabbitmq.Client, db *database.Pool, stats *QueueStats) *DLQHandler {
	return &DLQHandler{
		client:     client,
		db:         db,
		maxRetries: 3,
		stats:      stats,
	}
}

// Start begins monitoring and processing the dead letter queue.
func (h *DLQHandler) Start(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.processDLQ(ctx)
		}
	}
}

// processDLQ checks the dead letter queue and retries messages.
func (h *DLQHandler) processDLQ(ctx context.Context) {
	// Inspect DLX queue
	q, err := h.client.QueueInfo(rabbitmq.QueueDLX)
	if err != nil {
		log.Printf("Failed to inspect DLX queue: %v", err)
		return
	}

	if q.Messages == 0 {
		return
	}

	log.Printf("Dead letter queue has %d messages, processing...", q.Messages)

	// Consume a batch from DLQ and requeue with retry
	err = h.client.Consume(ctx, rabbitmq.QueueDLX, func(msg amqp.Delivery) error {
		var execMsg ExecutionMessage
		if err := json.Unmarshal(msg.Body, &execMsg); err != nil {
			log.Printf("Invalid message in DLQ: %v", err)
			msg.Ack(false)
			return nil
		}

		execMsg.RetryCount++
		h.stats.recordRetried()

		if execMsg.RetryCount > h.maxRetries {
			// Max retries exceeded, mark as permanently failed
			log.Printf("Submission %s exceeded max retries (%d), marking as failed",
				execMsg.SubmissionID, h.maxRetries)
			h.stats.recordDeadLetter()

			// Update status in database
			_, _ = h.db.Exec(ctx,
				"UPDATE submissions SET status = 'failed', error_message = $2 WHERE id = $1",
				execMsg.SubmissionID, "max retries exceeded")

			msg.Ack(false)
			return nil
		}

		// Re-publish with exponential backoff
		backoff := time.Duration(1<<uint(execMsg.RetryCount-1)) * time.Second
		log.Printf("Retrying submission %s (attempt %d/%d) after %v backoff",
			execMsg.SubmissionID, execMsg.RetryCount, h.maxRetries, backoff)

		time.Sleep(backoff)

		if err := h.client.PublishToQueue(ctx, rabbitmq.QueueCodeExecution, execMsg); err != nil {
			log.Printf("Failed to requeue submission %s: %v", execMsg.SubmissionID, err)
			msg.Nack(false, true) // Requeue to DLX
			return err
		}

		msg.Ack(false)
		log.Printf("Successfully requeued submission %s (attempt %d)",
			execMsg.SubmissionID, execMsg.RetryCount)
		return nil
	})
	if err != nil {
		log.Printf("DLQ consumer error: %v", err)
	}
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

	rmqClient, err := rabbitmq.NewFromEnv()
	if err != nil {
		log.Printf("Warning: RabbitMQ not available: %v", err)
	}

	wsHub := websocket.NewHub()
	go wsHub.Run()

	queueStats := &QueueStats{}
	var dlqHandler *DLQHandler
	if rmqClient != nil {
		dlqHandler = NewDLQHandler(rmqClient, dbPool, queueStats)
	}

	srv := &Server{
		router:         gin.New(),
		db:             dbPool,
		redis:          redisClient,
		rabbitmq:       rmqClient,
		wsHub:          wsHub,
		jwtConfig:      middleware.NewJWTConfig(),
		blacklist:      middleware.NewTokenBlacklist(true),
		sessionManager: middleware.NewSessionManager(5),
		queueStats:     queueStats,
		dlqHandler:     dlqHandler,
	}

	// Setup middleware
	srv.router.Use(gin.Recovery())
	srv.router.Use(middleware.RequestIDMiddleware())
	srv.router.Use(middleware.LoggerMiddleware())
	srv.router.Use(middleware.CORSMiddleware())

	// Register routes
	srv.setupRoutes()

	// Start RabbitMQ consumer and DLQ handler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if rmqClient != nil {
		go srv.startConsumer()
		if dlqHandler != nil {
			go dlqHandler.Start(ctx)
		}
		// Start queue depth monitoring
		go srv.monitorQueueDepth(ctx)
	}

	// Start server
	port := getEnv("PORT", ServicePort)
	runServer(srv, port)
}

func (s *Server) setupRoutes() {
	s.router.GET("/health", s.healthCheck)

	api := s.router.Group("/api")
	{
		// Protected submission routes
		submissions := api.Group("/submissions")
		submissions.Use(middleware.AuthMiddleware(s.jwtConfig, s.blacklist, s.sessionManager))
		{
			submissions.POST("", s.createSubmission)
			submissions.GET("/:id", s.getSubmission)
			submissions.GET("/user/:userId", s.getUserSubmissions)
		}

		// Queue statistics endpoint (admin only)
		api.GET("/queue/stats", middleware.AuthMiddleware(s.jwtConfig, s.blacklist, s.sessionManager), s.queueStatsHandler)

		// WebSocket endpoint
		api.GET("/ws/submissions/:id", s.handleWebSocket)
	}
}

func (s *Server) healthCheck(c *gin.Context) {
	ctx := c.Request.Context()
	status := "ok"
	dbStatus := "ok"
	rmqStatus := "ok"

	if err := s.db.HealthCheck(ctx); err != nil {
		dbStatus = "error: " + err.Error()
		status = "degraded"
	}
	if s.rabbitmq != nil {
		if err := s.rabbitmq.HealthCheck(ctx); err != nil {
			rmqStatus = "error: " + err.Error()
			status = "degraded"
		}
	}

	// Include queue depth in health
	depthInfo := map[string]interface{}{
		"queue_depth": s.queueStats.QueueDepth,
	}

	c.JSON(http.StatusOK, gin.H{
		"service":  ServiceName,
		"status":   status,
		"database":  dbStatus,
		"rabbitmq": rmqStatus,
		"websocket": s.wsHub.Stats(),
		"queue":    depthInfo,
		"time":     time.Now().UTC(),
	})
}

// queueStatsHandler returns queue depth and processing statistics.
func (s *Server) queueStatsHandler(c *gin.Context) {
	// Check admin role
	role, _ := c.Get(middleware.ContextKeyRole)
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	s.queueStats.mu.RLock()
	stats := map[string]interface{}{
		"submissions_queued":    s.queueStats.SubmissionsQueued,
		"submissions_processed": s.queueStats.SubmissionsProcessed,
		"submissions_failed":    s.queueStats.SubmissionsFailed,
		"submissions_retried":   s.queueStats.SubmissionsRetried,
		"dead_letter_count":     s.queueStats.DeadLetterCount,
		"priority_submissions":  s.queueStats.PrioritySubmissions,
		"normal_submissions":    s.queueStats.NormalSubmissions,
		"queue_depth":           s.queueStats.QueueDepth,
		"avg_queue_time_ns":     s.queueStats.AvgQueueTime.Nanoseconds(),
		"last_depth_check":      s.queueStats.LastDepthCheck,
	}
	s.queueStats.mu.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"service": ServiceName,
		"stats":   stats,
	})
}

// createSubmission handles POST /api/submissions with priority support.
func (s *Server) createSubmission(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := c.Get(middleware.ContextKeyUserID)
	role, _ := c.Get(middleware.ContextKeyRole)

	var req CreateSubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	// Verify problem exists and is published
	var problemExists bool
	err := s.db.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM problems WHERE id = $1 AND is_published = TRUE)",
		req.ProblemID,
	).Scan(&problemExists)
	if err != nil || !problemExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "problem not found or not published"})
		return
	}

	// Determine priority: premium users get priority 1, free users get 0
	priority := 0
	if role == "premium" || role == "admin" {
		priority = 1
	}

	// Create submission record
	submissionID := uuid.New().String()
	_, err = s.db.Exec(ctx, `
		INSERT INTO submissions (id, user_id, problem_id, code, language, status)
		VALUES ($1, $2, $3, $4, $5, 'pending')
	`, submissionID, userID, req.ProblemID, req.Code, req.Language)
	if err != nil {
		log.Printf("Failed to create submission: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create submission"})
		return
	}

	// Publish to RabbitMQ with priority
	msg := ExecutionMessage{
		SubmissionID: submissionID,
		ProblemID:    req.ProblemID,
		UserID:       userID.(string),
		Code:         req.Code,
		Language:     req.Language,
		EnqueuedAt:   time.Now().UTC().Format(time.RFC3339),
		Priority:     priority,
		RetryCount:   0,
	}

	if s.rabbitmq != nil {
		// Set priority header for RabbitMQ
		body, _ := json.Marshal(msg)
		err = s.rabbitmq.Publish(ctx, rabbitmq.ExchangeCodeExec, "execution.pending",
			map[string]interface{}{
				"body":         body,
				"submission_id": submissionID,
				"priority":     priority,
			})
		if err != nil {
			// Fallback to direct queue publish if exchange publish fails
			if pubErr := s.rabbitmq.PublishToQueue(ctx, rabbitmq.QueueCodeExecution, msg); pubErr != nil {
				log.Printf("Failed to publish to queue: %v", pubErr)
			}
		}
	}

	s.queueStats.recordQueued(priority)

	// Update status to queued
	_, _ = s.db.Exec(ctx, "UPDATE submissions SET status = 'queued' WHERE id = $1", submissionID)

	// Notify user via WebSocket
	s.wsHub.SendToUser(userID.(string), &websocket.Message{
		Type: "submission.queued",
		Room: fmt.Sprintf("submission-%s", submissionID),
		Payload: map[string]interface{}{
			"submission_id": submissionID,
			"status":        "queued",
			"priority":      priority,
		},
	})

	c.JSON(http.StatusAccepted, gin.H{
		"submission_id": submissionID,
		"status":        "queued",
		"priority":      priority,
		"message":       "Submission queued for execution",
	})
}

// getSubmission handles GET /api/submissions/:id.
func (s *Server) getSubmission(c *gin.Context) {
	ctx := c.Request.Context()
	userID, _ := c.Get(middleware.ContextKeyUserID)
	submissionID := c.Param("id")

	// Try cache first
	cacheKey := fmt.Sprintf("submission:%s", submissionID)
	if s.redis != nil {
		var cached Submission
		if err := s.redis.Get(ctx, cacheKey, &cached); err == nil {
			// Only return if user owns this submission or is admin
			if cached.UserID == userID || s.isAdmin(c) {
				c.JSON(http.StatusOK, cached)
				return
			}
		}
	}

	var sub Submission
	var completedAt *time.Time
	var execTimeMs, memoryKb *int
	var errMsg *string

	err := s.db.QueryRow(ctx, `
		SELECT id, user_id, problem_id, code, language, status, score,
			execution_time_ms, memory_used_kb, test_cases_passed, test_cases_total,
			error_message, submitted_at, completed_at
		FROM submissions WHERE id = $1
	`, submissionID).Scan(&sub.ID, &sub.UserID, &sub.ProblemID, &sub.Code, &sub.Language,
		&sub.Status, &sub.Score, &execTimeMs, &memoryKb,
		&sub.TestCasesPassed, &sub.TestCasesTotal, &errMsg,
		&sub.SubmittedAt, &completedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "submission not found"})
		return
	}

	// Authorization check
	if sub.UserID != userID && !s.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	sub.ExecutionTimeMs = execTimeMs
	sub.MemoryUsedKb = memoryKb
	sub.ErrorMessage = errMsg
	sub.CompletedAt = completedAt

	// Cache if completed
	if sub.Status == "completed" && s.redis != nil {
		_ = s.redis.Set(ctx, cacheKey, sub, redis.TTLSubmission)
	}

	c.JSON(http.StatusOK, sub)
}

// getUserSubmissions handles GET /api/submissions/user/:userId.
func (s *Server) getUserSubmissions(c *gin.Context) {
	ctx := c.Request.Context()
	requestedUserID := c.Param("userId")
	currentUserID, _ := c.Get(middleware.ContextKeyUserID)

	// Only allow users to see their own submissions (or admin)
	if requestedUserID != currentUserID && !s.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	rows, err := s.db.Query(ctx, `
		SELECT s.id, s.problem_id, p.title AS problem_title, p.difficulty,
			s.language, s.status, s.score, s.execution_time_ms, s.memory_used_kb,
			s.test_cases_passed, s.test_cases_total, s.submitted_at, s.completed_at
		FROM submissions s
		JOIN problems p ON s.problem_id = p.id
		WHERE s.user_id = $1
		ORDER BY s.submitted_at DESC
		LIMIT $2 OFFSET $3
	`, requestedUserID, pageSize, (page-1)*pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	defer rows.Close()

	type SubmissionListItem struct {
		ID               string     `json:"id"`
		ProblemID        int        `json:"problem_id"`
		ProblemTitle     string     `json:"problem_title"`
		Difficulty       string     `json:"difficulty"`
		Language         string     `json:"language"`
		Status           string     `json:"status"`
		Score            int        `json:"score"`
		ExecutionTimeMs  *int       `json:"execution_time_ms,omitempty"`
		MemoryUsedKb     *int       `json:"memory_used_kb,omitempty"`
		TestCasesPassed  int        `json:"test_cases_passed"`
		TestCasesTotal   int        `json:"test_cases_total"`
		SubmittedAt      time.Time  `json:"submitted_at"`
		CompletedAt      *time.Time `json:"completed_at,omitempty"`
	}

	var submissions []SubmissionListItem
	for rows.Next() {
		var s SubmissionListItem
		if err := rows.Scan(&s.ID, &s.ProblemID, &s.ProblemTitle, &s.Difficulty,
			&s.Language, &s.Status, &s.Score, &s.ExecutionTimeMs, &s.MemoryUsedKb,
			&s.TestCasesPassed, &s.TestCasesTotal, &s.SubmittedAt, &s.CompletedAt); err != nil {
			continue
		}
		submissions = append(submissions, s)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      submissions,
		"page":      page,
		"page_size": pageSize,
	})
}

// handleWebSocket handles real-time submission status updates.
func (s *Server) handleWebSocket(c *gin.Context) {
	userID, exists := c.Get(middleware.ContextKeyUserID)
	if !exists {
		// Try to get from query param for service-to-service
		userID = c.Query("user_id")
	}
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	s.wsHub.HandleWebSocket(c, userID.(string))
}

// startConsumer starts consuming execution results from RabbitMQ.
func (s *Server) startConsumer() {
	if s.rabbitmq == nil {
		return
	}

	ctx := context.Background()
	err := s.rabbitmq.Consume(ctx, rabbitmq.QueueCodeExecution, func(msg amqp.Delivery) error {
		var result ExecutionMessage
		if err := json.Unmarshal(msg.Body, &result); err != nil {
			log.Printf("Invalid message format: %v", err)
			return nil // Don't requeue invalid messages
		}

		log.Printf("Processing submission: %s", result.SubmissionID)

		// Update status to running
		_, _ = s.db.Exec(ctx,
			"UPDATE submissions SET status = 'running' WHERE id = $1",
			result.SubmissionID,
		)

		s.queueStats.recordProcessed()

		// Notify user
		s.wsHub.SendToUser(result.UserID, &websocket.Message{
			Type: "submission.running",
			Room: fmt.Sprintf("submission-%s", result.SubmissionID),
			Payload: map[string]string{
				"submission_id": result.SubmissionID,
				"status":        "running",
			},
		})

		// Invalidate cache
		if s.redis != nil {
			_ = s.redis.Delete(ctx, fmt.Sprintf("submission:%s", result.SubmissionID))
		}

		return nil
	})

	if err != nil {
		log.Printf("Consumer error: %v", err)
	}
}

// monitorQueueDepth periodically checks the queue depth.
func (s *Server) monitorQueueDepth(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if s.rabbitmq == nil {
				continue
			}

			q, err := s.rabbitmq.QueueInfo(rabbitmq.QueueCodeExecution)
			if err != nil {
				log.Printf("Failed to check queue depth: %v", err)
				continue
			}

			s.queueStats.updateDepth(q.Messages)
			log.Printf("Queue depth: %d messages (ready=%d, unacked=%d)",
				q.Messages, q.Messages, q.Messages)

			// Also check DLQ depth
			dlq, err := s.rabbitmq.QueueInfo(rabbitmq.QueueDLX)
			if err == nil && dlq.Messages > 0 {
				log.Printf("Warning: Dead letter queue has %d messages", dlq.Messages)
			}

			// If queue depth is high, log a warning
			if q.Messages > 50 {
				log.Printf("High queue depth warning: %d messages, considering scaling workers", q.Messages)
			}
		}
	}
}

// processSubmissionResult handles a completed execution result.
func (s *Server) processSubmissionResult(ctx context.Context, submissionID string, results []TestCaseResult) error {
	// Calculate score
	var passed, total int
	for _, r := range results {
		total++
		if r.Status == "passed" {
			passed++
		}
	}

	// Get problem max score
	var maxScore, problemID int
	var userID string
	err := s.db.QueryRow(ctx,
		"SELECT problem_id, user_id FROM submissions WHERE id = $1", submissionID,
	).Scan(&problemID, &userID)
	if err != nil {
		return fmt.Errorf("submission not found: %w", err)
	}

	err = s.db.QueryRow(ctx, "SELECT max_score FROM problems WHERE id = $1", problemID).Scan(&maxScore)
	if err != nil {
		return fmt.Errorf("problem not found: %w", err)
	}

	// Calculate score based on passed ratio
	score := 0
	if total > 0 {
		score = (passed * maxScore) / total
	}

	// Update submission
	_, err = s.db.Exec(ctx, `
		UPDATE submissions 
		SET status = 'completed', score = $2, test_cases_passed = $3, test_cases_total = $4, completed_at = NOW()
		WHERE id = $1
	`, submissionID, score, passed, total)
	if err != nil {
		return fmt.Errorf("failed to update submission: %w", err)
	}

	// Notify user via WebSocket
	s.wsHub.SendToUser(userID, &websocket.Message{
		Type: "submission.completed",
		Room: fmt.Sprintf("submission-%s", submissionID),
		Payload: SubmissionResult{
			SubmissionID:     submissionID,
			Status:           "completed",
			Score:            score,
			TestCasesPassed: passed,
			TestCasesTotal:  total,
			Results:          results,
		},
	})

	// Invalidate caches
	if s.redis != nil {
		_ = s.redis.Delete(ctx, fmt.Sprintf("submission:%s", submissionID))
		_ = s.redis.DeletePattern(ctx, "leaderboard:*")
		_ = s.redis.DeletePattern(ctx, fmt.Sprintf("user:stats:%s", userID))
	}

	return nil
}

func (s *Server) isAdmin(c *gin.Context) bool {
	role, _ := c.Get(middleware.ContextKeyRole)
	return role == "admin"
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
