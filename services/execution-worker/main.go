package main

import (
	"bytes"
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

	amqp "github.com/rabbitmq/amqp091-go"
	"coding-challange/pkg/database"
	"coding-challange/pkg/rabbitmq"
	"coding-challange/pkg/redis"
	"coding-challange/pkg/websocket"
)

// Service configuration constants.
const (
	ServiceName = "execution-worker"
	ServicePort = "9106"
)

// ExecutionMessage represents a message consumed from the RabbitMQ queue.
type ExecutionMessage struct {
	SubmissionID string `json:"submission_id"`
	ProblemID    int    `json:"problem_id"`
	UserID       string `json:"user_id"`
	Code         string `json:"code"`
	Language     string `json:"language"`
	EnqueuedAt   string `json:"enqueued_at"`
	Priority     int    `json:"priority,omitempty"` // 0=free, 1=premium
}

// ExecutionResult represents the complete result of a submission execution.
type ExecutionResult struct {
	SubmissionID     string           `json:"submission_id"`
	Status           string           `json:"status"` // "accepted", "wrong_answer", "time_limit", "runtime_error", "compilation_error"
	Score            int              `json:"score"`
	ExecutionTimeMs  int              `json:"execution_time_ms"`
	MemoryUsedKb     int              `json:"memory_used_kb"`
	TestCasesPassed  int              `json:"test_cases_passed"`
	TestCasesTotal   int              `json:"test_cases_total"`
	Results          []TestResult     `json:"results,omitempty"`
	ErrorMessage     string           `json:"error_message,omitempty"`
	CompletedAt      time.Time        `json:"completed_at"`
	PerTestCaseMetrics []TestCaseMetrics `json:"per_test_case_metrics,omitempty"`
	SandboxUsed      string           `json:"sandbox_used,omitempty"`
}

// TestCaseMetrics holds detailed metrics per test case.
type TestCaseMetrics struct {
	TestCaseID      int    `json:"test_case_id"`
	ExecutionTimeMs int    `json:"execution_time_ms"`
	MemoryUsedKb    int    `json:"memory_used_kb"`
	CPUTimeMs       int    `json:"cpu_time_ms"`
	DiskIOKB        int    `json:"disk_io_kb"`
	NetworkIOKB     int    `json:"network_io_kb"`
	OutputSize      int    `json:"output_size"`
	TimedOut        bool   `json:"timed_out,omitempty"`
	ErrorMessage    string `json:"error_message,omitempty"`
}

// executionMetrics tracks aggregate execution statistics.
type executionMetrics struct {
	mu                sync.Mutex
	TotalExecutions   int           `json:"total_executions"`
	SuccessfulExecs   int           `json:"successful_execs"`
	FailedExecs       int           `json:"failed_execs"`
	TimedOutExecs     int           `json:"timed_out_execs"`
	AvgExecutionTime  time.Duration `json:"avg_execution_time"`
	TotalExecutionTime time.Duration `json:"total_execution_time"`
	PeakMemoryUsedKb  int           `json:"peak_memory_used_kb"`
	ExecutionsByLang  map[string]int `json:"executions_by_lang"`
}

func newExecutionMetrics() *executionMetrics {
	return &executionMetrics{
		ExecutionsByLang: make(map[string]int),
	}
}

func (m *executionMetrics) record(result *ExecutionResult, lang string, execTime time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalExecutions++
	m.TotalExecutionTime += execTime
	m.AvgExecutionTime = m.TotalExecutionTime / time.Duration(m.TotalExecutions)

	if result.Status == "accepted" {
		m.SuccessfulExecs++
	} else {
		m.FailedExecs++
	}
	if result.Status == "time_limit" {
		m.TimedOutExecs++
	}

	if result.MemoryUsedKb > m.PeakMemoryUsedKb {
		m.PeakMemoryUsedKb = result.MemoryUsedKb
	}

	m.ExecutionsByLang[lang]++
}

// Server holds all dependencies for the execution worker.
type Server struct {
	config    WorkerConfig
	db        *database.Pool
	redis     *redis.Client
	rabbitmq  *rabbitmq.Client
	wsHub     *websocket.Hub
	consumer  *Consumer
	sandbox   *SandboxExecutor
	metrics   *executionMetrics
	execSem   chan struct{} // Semaphore for limiting concurrent executions
}

// Consumer wraps the RabbitMQ consumer with priority queue support.
type Consumer struct {
	client      *rabbitmq.Client
	priorityQ   chan ExecutionMessage // High priority queue (premium users)
	normalQ     chan ExecutionMessage // Normal queue (free users)
	dlq         chan ExecutionMessage // Dead letter queue
	server      *Server
	workerID    string
	maxRetries  int
	pool        *WorkerPool
}

// WorkerPool manages a pool of goroutine workers for processing execution messages.
type WorkerPool struct {
	workers    int
	jobCh      chan workerJob
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

type workerJob struct {
	priority bool
	msg      ExecutionMessage
	consumer *Consumer
}

// NewWorkerPool creates a new worker pool with the specified number of workers.
func NewWorkerPool(parentCtx context.Context, numWorkers int) *WorkerPool {
	if numWorkers < 1 {
		numWorkers = 1
	}
	if numWorkers > 10 {
		numWorkers = 10
	}
	ctx, cancel := context.WithCancel(parentCtx)
	pool := &WorkerPool{
		workers: numWorkers,
		jobCh:   make(chan workerJob, 100),
		ctx:     ctx,
		cancel:  cancel,
	}
	pool.start()
	return pool
}

// start launches the worker goroutines.
func (wp *WorkerPool) start() {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// worker is a single worker goroutine that processes jobs from the channel.
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	for {
		select {
		case <-wp.ctx.Done():
			return
		case job, ok := <-wp.jobCh:
			if !ok {
				return
			}
			job.consumer.processWithRetry(wp.ctx, job.msg, 0)
		}
	}
}

// Submit sends a job to the worker pool.
func (wp *WorkerPool) Submit(job workerJob) {
	select {
	case wp.jobCh <- job:
	case <-wp.ctx.Done():
	}
}

// Stop gracefully shuts down the worker pool.
func (wp *WorkerPool) Stop() {
	wp.cancel()
	wp.wg.Wait()
}

// NewConsumer creates a new RabbitMQ consumer for the execution worker.
func NewConsumer(client *rabbitmq.Client, srv *Server) *Consumer {
	hostname, _ := os.Hostname()
	return &Consumer{
		client:     client,
		priorityQ:  make(chan ExecutionMessage, 100),
		normalQ:    make(chan ExecutionMessage, 100),
		dlq:        make(chan ExecutionMessage, 50),
		server:     srv,
		workerID:   hostname,
		maxRetries: 3,
	}
}

// Start begins consuming messages from the code execution queue with priority handling
// using a bounded worker pool.
func (c *Consumer) Start(ctx context.Context) error {
	// Create worker pool: use config's MaxConcurrentExecutions, default 4, max 10
	numWorkers := c.server.config.MaxConcurrentExecutions
	if numWorkers < 1 {
		numWorkers = 4
	}
	if numWorkers > 10 {
		numWorkers = 10
	}
	c.pool = NewWorkerPool(ctx, numWorkers)
	log.Printf("[WORKER POOL] Started %d workers (max 10)", numWorkers)

	// Start work stealing goroutine
	go c.workStealing(ctx)

	// Consume messages from the queue for priority routing
	return c.client.Consume(ctx, rabbitmq.QueueCodeExecution, c.handleMessage)
}

// handleMessage processes a single execution message with priority routing,
// submitting directly to the worker pool.
func (c *Consumer) handleMessage(msg amqp.Delivery) error {
	var execMsg ExecutionMessage
	if err := json.Unmarshal(msg.Body, &execMsg); err != nil {
		log.Printf("Invalid message format: %v", err)
		msg.Ack(false) // Don't requeue invalid messages
		return nil
	}

	log.Printf("Received submission: %s (problem: %d, user: %s, lang: %s, priority: %d)",
		execMsg.SubmissionID, execMsg.ProblemID, execMsg.UserID, execMsg.Language, execMsg.Priority)

	msg.Ack(false)
	c.pool.Submit(workerJob{
		priority: execMsg.Priority >= 1,
		msg:      execMsg,
		consumer: c,
	})

	return nil
}

// processPriorityWorker processes high-priority submissions (premium users).
func (c *Consumer) processPriorityWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-c.priorityQ:
			c.processWithRetry(ctx, msg, 0)
		}
	}
}

// processNormalWorker processes normal submissions (free users).
func (c *Consumer) processNormalWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-c.normalQ:
			c.processWithRetry(ctx, msg, 0)
		}
	}
}

// processDeadLetterQueue handles messages that couldn't be queued normally.
func (c *Consumer) processDeadLetterQueue(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-c.dlq:
			// Retry DLQ messages with backoff
			c.processWithRetry(ctx, msg, 0)
		case <-ticker.C:
			// Log queue depth
			c.logQueueDepth()
		}
	}
}

// workStealing implements work stealing between workers.
func (c *Consumer) workStealing(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Steal from normal queue if priority workers are idle
			if len(c.priorityQ) == 0 && len(c.normalQ) > 0 {
				select {
				case msg := <-c.normalQ:
					select {
					case c.priorityQ <- msg:
						log.Printf("Work stealing: moved submission %s from normal to priority queue", msg.SubmissionID)
					default:
						// Put it back
						select {
						case c.normalQ <- msg:
						default:
						}
					}
				default:
				}
			}
		}
	}
}

// processWithRetry processes a message with exponential backoff retry.
func (c *Consumer) processWithRetry(ctx context.Context, msg ExecutionMessage, attempt int) {
	if attempt > 0 {
		log.Printf("Retry attempt %d/%d for submission %s", attempt, c.maxRetries, msg.SubmissionID)
	}

	// Acquire semaphore slot
	select {
	case c.server.execSem <- struct{}{}:
	case <-ctx.Done():
		return
	}

	// Process the submission
	result := c.server.processSubmission(ctx, msg)

	// Release semaphore
	<-c.server.execSem

	// If failed and retries remaining, retry with backoff
	if result.Status == "runtime_error" && result.ErrorMessage != "" && attempt < c.maxRetries {
		backoff := time.Duration(1<<uint(attempt)) * time.Second
		log.Printf("Submission %s failed (attempt %d), retrying in %v...", msg.SubmissionID, attempt, backoff)

		select {
		case <-time.After(backoff):
			c.processWithRetry(ctx, msg, attempt+1)
		case <-ctx.Done():
			return
		}
		return
	}

	// Publish result
	c.server.publishResult(ctx, msg, result)
}

// logQueueDepth logs the current queue depths and worker pool stats for monitoring.
func (c *Consumer) logQueueDepth() {
	poolWorkers := 0
	poolJobs := 0
	if c.pool != nil {
		poolWorkers = c.pool.workers
		poolJobs = len(c.pool.jobCh)
	}
	log.Printf("Queue depths - Priority: %d, Normal: %d, DLQ: %d, Semaphore: %d/%d, Pool: jobs=%d workers=%d",
		len(c.priorityQ), len(c.normalQ), len(c.dlq),
		len(c.server.execSem), cap(c.server.execSem),
		poolJobs, poolWorkers)
}

// processSubmission executes the submission with parallel test case processing.
func (s *Server) processSubmission(ctx context.Context, msg ExecutionMessage) *ExecutionResult {
	log.Printf("Processing submission: %s (problem: %d, user: %s, lang: %s)",
		msg.SubmissionID, msg.ProblemID, msg.UserID, msg.Language)

	startTime := time.Now()

	result := &ExecutionResult{
		SubmissionID: msg.SubmissionID,
		Status:       "accepted",
		CompletedAt:  time.Now().UTC(),
	}

	// Get problem definition with test cases
	problem, err := s.getProblem(ctx, msg.ProblemID)
	if err != nil {
		result.Status = "runtime_error"
		result.ErrorMessage = fmt.Sprintf("Failed to load problem: %v", err)
		s.metrics.record(result, msg.Language, time.Since(startTime))
		return result
	}

	// Get resource limits based on problem difficulty and category
	limits := s.config.GetLimitsForProblem(problem.Difficulty, problem.Category)

	// Create harness config
	harnessCfg := DefaultHarnessConfig()

	// Generate test harness
	var harnessCode string
	if problem.Type == "function" {
		harnessCode = generateFunctionHarness(msg.Code, *problem, harnessCfg)
	} else {
		harnessCode = generateMainHarness(msg.Code, *problem, harnessCfg)
	}

	// Execute test cases in parallel using a worker pool
	type testCaseJob struct {
		tc      TestCaseDef
		timeout time.Duration
	}

	type testCaseResult struct {
		tc     TestCaseDef
		result *SandboxResult
	}

	numTestCases := len(problem.TestCases)
	if numTestCases == 0 {
		result.Status = "runtime_error"
		result.ErrorMessage = "No test cases defined"
		s.metrics.record(result, msg.Language, time.Since(startTime))
		return result
	}

	jobs := make(chan testCaseJob, numTestCases)
	results := make(chan testCaseResult, numTestCases)
	var wg sync.WaitGroup

	// Calculate per-test-case timeout
	perTestCaseTimeout := limits.TimeLimit / time.Duration(numTestCases)
	if perTestCaseTimeout < 100*time.Millisecond {
		perTestCaseTimeout = 100 * time.Millisecond
	}
	if perTestCaseTimeout > limits.TimeLimit {
		perTestCaseTimeout = limits.TimeLimit
	}

	// Determine max parallel workers based on config
	maxWorkers := s.config.MaxConcurrentExecutions
	if maxWorkers > numTestCases {
		maxWorkers = numTestCases
	}
	if maxWorkers < 1 {
		maxWorkers = 1
	}

	// Start parallel workers
	for w := 0; w < maxWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				tcCtx, tcCancel := context.WithTimeout(ctx, job.timeout)
				sandboxResult, err := s.sandbox.Execute(tcCtx, harnessCode, msg.Language, job.timeout)
				tcCancel()

				if err != nil && sandboxResult == nil {
					log.Printf("Sandbox execution failed for test case %d: %v", job.tc.ID, err)
					sandboxResult = &SandboxResult{
						ExitCode:     1,
						ErrorMessage: fmt.Sprintf("sandbox error: %v", err),
						TimedOut:     false,
					}
				}

				results <- testCaseResult{tc: job.tc, result: sandboxResult}
			}
		}()
	}

	// Send jobs
	for _, tc := range problem.TestCases {
		jobs <- testCaseJob{tc: tc, timeout: perTestCaseTimeout}
	}
	close(jobs)

	// Wait for all workers to finish
	wg.Wait()
	close(results)

	// Collect and process results
	var passed, total int
	var testResults []TestResult
	var perTestCaseMetrics []TestCaseMetrics

	for tcResult := range results {
		total++

		sandboxResult := tcResult.result
		tcMetric := TestCaseMetrics{
			TestCaseID:   tcResult.tc.ID,
			TimedOut:     sandboxResult.TimedOut,
			ErrorMessage: sanitizeOutput(sandboxResult.ErrorMessage),
		}

		// Parse harness output
		harnessResults := parseTestResults(sandboxResult.Stdout, []TestCaseDef{tcResult.tc})

		if len(harnessResults) > 0 {
			hr := harnessResults[0]
			tcMetric.ExecutionTimeMs = hr.ExecutionTimeMs
			tcMetric.MemoryUsedKb = hr.MemoryUsedKb
			tcMetric.CPUTimeMs = hr.CPUTimeMs
			tcMetric.DiskIOKB = hr.DiskIOKB
			tcMetric.NetworkIOKB = hr.NetworkIOKB

			testResults = append(testResults, hr)
			if hr.Status == "passed" {
				passed++
			}
		} else if sandboxResult.TimedOut {
			testResults = append(testResults, TestResult{
				TestCaseID:   tcResult.tc.ID,
				Status:       "time_limit",
				ErrorMessage: "execution timeout",
			})
		} else {
			testResults = append(testResults, TestResult{
				TestCaseID:   tcResult.tc.ID,
				Status:       "runtime_error",
				ErrorMessage: sandboxResult.ErrorMessage,
			})
		}

		perTestCaseMetrics = append(perTestCaseMetrics, tcMetric)
	}

	// Determine overall status
	result.TestCasesPassed = passed
	result.TestCasesTotal = total
	result.Results = testResults
	result.PerTestCaseMetrics = perTestCaseMetrics

	if passed == total {
		result.Status = "accepted"
	} else if total > 0 {
		result.Status = "wrong_answer"
	}

	// Check for timeouts
	for _, m := range perTestCaseMetrics {
		if m.TimedOut {
			if result.Status == "accepted" {
				result.Status = "time_limit"
			}
		}
	}

	// Calculate score
	result.Score = calculateScore(testResults, problem.TestCases, problem.MaxScore)

	// Calculate aggregate metrics
	var totalExecMs, totalMemKb int
	for _, m := range perTestCaseMetrics {
		totalExecMs += m.ExecutionTimeMs
		if m.MemoryUsedKb > totalMemKb {
			totalMemKb = m.MemoryUsedKb
		}
	}
	result.ExecutionTimeMs = totalExecMs
	result.MemoryUsedKb = totalMemKb

	// Record sandbox used
	result.SandboxUsed = "docker"

	// Record metrics
	s.metrics.record(result, msg.Language, time.Since(startTime))

	return result
}

// getProblem retrieves a problem definition with test cases from the database.
func (s *Server) getProblem(ctx context.Context, problemID int) (*ProblemDefinition, error) {
	problem := &ProblemDefinition{}

	err := s.db.QueryRow(ctx, `
		SELECT id, title, COALESCE(difficulty, 'medium') as difficulty, type, 
		       COALESCE(category, '') as category, max_score, 
		       time_limit_ms, memory_limit_mb,
		       COALESCE(function_sig, '') as function_sig, 
		       COALESCE(function_name, 'solution') as function_name
		FROM problems WHERE id = $1
	`, problemID).Scan(
		&problem.ID, &problem.Title, &problem.Difficulty, &problem.Type,
		&problem.Category, &problem.MaxScore,
		&problem.TimeLimitMs, &problem.MemoryLimitMb,
		&problem.FunctionSig, &problem.FunctionName,
	)
	if err != nil {
		return nil, fmt.Errorf("problem not found: %w", err)
	}

	// Load test cases
	rows, err := s.db.Query(ctx, `
		SELECT id, input, expected, COALESCE(weight, 1) as weight, is_hidden
		FROM test_cases WHERE problem_id = $1 ORDER BY id
	`, problemID)
	if err != nil {
		return nil, fmt.Errorf("failed to load test cases: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tc TestCaseDef
		var inputStr, expectedStr string
		if err := rows.Scan(&tc.ID, &inputStr, &expectedStr, &tc.Weight, &tc.IsHidden); err != nil {
			continue
		}
		// Parse JSON input
		json.Unmarshal([]byte(inputStr), &tc.Input)
		json.Unmarshal([]byte(expectedStr), &tc.Expected)
		problem.TestCases = append(problem.TestCases, tc)
	}

	return problem, nil
}

// publishResult sends the execution result back to the execution service.
func (s *Server) publishResult(ctx context.Context, msg ExecutionMessage, result *ExecutionResult) {
	// Update database
	_, err := s.db.Exec(ctx, `
		UPDATE submissions 
		SET status = $2, score = $3, execution_time_ms = $4, memory_used_kb = $5,
		    test_cases_passed = $6, test_cases_total = $7, 
		    error_message = $8, completed_at = NOW()
		WHERE id = $1
	`, msg.SubmissionID, result.Status, result.Score, result.ExecutionTimeMs,
		result.MemoryUsedKb, result.TestCasesPassed, result.TestCasesTotal,
		result.ErrorMessage)
	if err != nil {
		log.Printf("Failed to update submission %s: %v", msg.SubmissionID, err)
	}

	// Publish result to RabbitMQ for WebSocket notification
	if s.rabbitmq != nil {
		resultMsg := map[string]interface{}{
			"submission_id":    result.SubmissionID,
			"status":           result.Status,
			"score":            result.Score,
			"execution_time_ms": result.ExecutionTimeMs,
			"memory_used_kb":   result.MemoryUsedKb,
			"test_cases_passed": result.TestCasesPassed,
			"test_cases_total": result.TestCasesTotal,
			"user_id":          msg.UserID,
		}
		if err := s.rabbitmq.PublishToQueue(ctx, rabbitmq.QueueCodeExecution, resultMsg); err != nil {
			log.Printf("Failed to publish result: %v", err)
		}
	}

	// Send WebSocket notification
	s.wsHub.SendToUser(msg.UserID, &websocket.Message{
		Type: "submission.completed",
		Room: fmt.Sprintf("submission-%s", result.SubmissionID),
		Payload: map[string]interface{}{
			"submission_id":    result.SubmissionID,
			"status":           result.Status,
			"score":            result.Score,
			"execution_time_ms": result.ExecutionTimeMs,
			"memory_used_kb":   result.MemoryUsedKb,
			"test_cases_passed": result.TestCasesPassed,
			"test_cases_total": result.TestCasesTotal,
		},
	})

	// Cache result in Redis
	if s.redis != nil {
		cacheKey := fmt.Sprintf("submission:%s", result.SubmissionID)
		s.redis.Set(ctx, cacheKey, result, time.Hour)
	}

	// Trigger leaderboard update
	s.updateLeaderboard(ctx, msg, result)

	log.Printf("Submission %s completed: status=%s score=%d passed=%d/%d",
		result.SubmissionID, result.Status, result.Score,
		result.TestCasesPassed, result.TestCasesTotal)
}

// updateLeaderboard calls the leaderboard service to update rankings after a submission completes.
func (s *Server) updateLeaderboard(ctx context.Context, msg ExecutionMessage, result *ExecutionResult) {
	leaderboardURL := getEnv("LEADERBOARD_SERVICE_URL", "http://leaderboard-service:9104")
	endpoint := fmt.Sprintf("%s/api/internal/update-score", leaderboardURL)

	// Get problem difficulty for ELO calculation
	var difficulty string
	err := s.db.QueryRow(ctx, "SELECT COALESCE(difficulty, 'medium') FROM problems WHERE id = $1", msg.ProblemID).Scan(&difficulty)
	if err != nil {
		log.Printf("Failed to get problem difficulty for leaderboard update: %v", err)
		difficulty = "medium"
	}

	// Get problem_id as string for the leaderboard API
	problemIDStr := strconv.Itoa(msg.ProblemID)

	payload := map[string]interface{}{
		"user_id":          msg.UserID,
		"submission_id":    result.SubmissionID,
		"problem_id":       problemIDStr,
		"score":            result.Score,
		"difficulty":       difficulty,
		"status":           result.Status,
		"execution_time_ms": result.ExecutionTimeMs,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal leaderboard payload: %v", err)
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		log.Printf("Failed to create leaderboard request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Failed to call leaderboard service: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		log.Printf("Leaderboard service returned status %d for submission %s", resp.StatusCode, result.SubmissionID)
	} else {
		log.Printf("Leaderboard updated for submission %s (status: %s)", result.SubmissionID, result.Status)
	}
}

func main() {
	log.Printf("Starting %s...", ServiceName)

	// Load worker configuration
	config := DefaultWorkerConfig()

	// Initialize database
	dbPool, err := database.NewFromEnv()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbPool.Close()

	// Initialize Redis (optional)
	redisClient, err := redis.NewFromEnv()
	if err != nil {
		log.Printf("Warning: Redis not available: %v", err)
		redisClient = nil
	}

	// Initialize RabbitMQ
	rmqClient, err := rabbitmq.NewFromEnv()
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ: %v", err)
	}
	defer rmqClient.Close()

	// Initialize WebSocket hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// Initialize sandbox executor
	sandboxConfig := DefaultSandboxConfig()
	sandboxConfig.ResourceLimits = config.DefaultResourceLimits
	sandboxConfig.WarmPoolSize = config.WarmPoolSize
	sandboxConfig.EnableLocalFallback = config.EnableLocalFallback
	sandbox := NewSandboxExecutor(sandboxConfig)

	// Create server
	metrics := newExecutionMetrics()
	srv := &Server{
		config:   config,
		db:       dbPool,
		redis:    redisClient,
		rabbitmq: rmqClient,
		wsHub:    wsHub,
		sandbox:  sandbox,
		metrics:  metrics,
		execSem:  make(chan struct{}, config.MaxConcurrentExecutions),
	}

	// Start the execution consumer
	srv.consumer = NewConsumer(rmqClient, srv)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		log.Println("Starting RabbitMQ consumer for code execution...")
		if err := srv.consumer.Start(ctx); err != nil {
			log.Printf("Consumer error: %v", err)
		}
	}()

	// Start health/metrics HTTP server
	port := getEnv("PORT", ServicePort)
	runServer(srv, port)
}

// runServer starts the HTTP server with graceful shutdown.
func runServer(srv *Server, port string) {
	addr := ":" + port

	// Create HTTP server with enhanced endpoints
	mux := http.NewServeMux()
	mux.HandleFunc("/health", srv.healthCheckHandler)
	mux.HandleFunc("/stats", srv.statsHandler)
	mux.HandleFunc("/metrics", srv.metricsHandler)
	mux.HandleFunc("/config", srv.configHandler)

	httpSrv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		log.Printf("%s starting on %s", ServiceName, addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Printf("Shutting down %s...", ServiceName)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Printf("%s exited gracefully", ServiceName)
}

// healthCheckHandler handles health check requests.
func (s *Server) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status := "ok"
	dbStatus := "ok"
	rmqStatus := "ok"
	sandboxStatus := "ok"

	if err := s.db.HealthCheck(ctx); err != nil {
		dbStatus = "unavailable"
		status = "degraded"
	}
	if s.rabbitmq != nil {
		if err := s.rabbitmq.HealthCheck(ctx); err != nil {
			rmqStatus = "unavailable"
			status = "degraded"
		}
	}

	// Check sandbox availability
	if !isDockerAvailable() {
		sandboxStatus = "unavailable"
		if !s.config.EnableLocalFallback {
			status = "degraded"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service":  ServiceName,
		"status":   status,
		"database": dbStatus,
		"rabbitmq": rmqStatus,
		"sandbox":  sandboxStatus,
		"time":     time.Now().UTC(),
	})
}

// statsHandler returns worker statistics.
func (s *Server) statsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	s.metrics.mu.Lock()
	metricsMap := map[string]interface{}{
		"total_executions":    s.metrics.TotalExecutions,
		"successful_execs":    s.metrics.SuccessfulExecs,
		"failed_execs":        s.metrics.FailedExecs,
		"timed_out_execs":     s.metrics.TimedOutExecs,
		"avg_execution_time":  s.metrics.AvgExecutionTime.String(),
		"total_execution_time": s.metrics.TotalExecutionTime.String(),
		"peak_memory_used_kb": s.metrics.PeakMemoryUsedKb,
		"executions_by_lang":  s.metrics.ExecutionsByLang,
	}
	s.metrics.mu.Unlock()

	sandboxCacheStats := s.sandbox.memory.Stats()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"service":           ServiceName,
		"execution_metrics": metricsMap,
		"sandbox_cache":     sandboxCacheStats,
		"queue_depth": map[string]int{
			"priority": cap(s.consumer.priorityQ),
			"normal":   cap(s.consumer.normalQ),
			"dlq":      cap(s.consumer.dlq),
		},
		"websocket": s.wsHub.Stats(),
		"database":  s.db.Stats(),
		"time":      time.Now().UTC(),
	})
}

// metricsHandler returns detailed execution metrics.
func (s *Server) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	s.metrics.mu.Lock()
	metricsMap := map[string]interface{}{
		"total_executions":     s.metrics.TotalExecutions,
		"successful_execs":     s.metrics.SuccessfulExecs,
		"failed_execs":         s.metrics.FailedExecs,
		"timed_out_execs":      s.metrics.TimedOutExecs,
		"avg_execution_time":   s.metrics.AvgExecutionTime.String(),
		"total_execution_time":  s.metrics.TotalExecutionTime.String(),
		"peak_memory_used_kb":  s.metrics.PeakMemoryUsedKb,
		"executions_by_lang":   s.metrics.ExecutionsByLang,
	}
	s.metrics.mu.Unlock()

	json.NewEncoder(w).Encode(metricsMap)
}

// configHandler returns the current worker configuration.
func (s *Server) configHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.config)
}
