package main

import (
	"bufio"
	"bytes"
	"container/list"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"coding-challange/pkg/logger"
	"coding-challange/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ============================================================================
// Service Discovery
// ============================================================================

// ServiceInstance represents a single instance of a backend service.
type ServiceInstance struct {
	Name    string
	URL     *url.URL
	Healthy uint32 // atomic boolean
	Weight  int
}

// HealthChecker performs health checks on service instances.
type HealthChecker struct {
	mu         sync.Mutex
	instances  []*ServiceInstance
	interval   time.Duration
	httpClient *http.Client
	done       chan struct{}
}

// NewHealthChecker creates a new health checker.
func NewHealthChecker(instances []*ServiceInstance, interval time.Duration) *HealthChecker {
	return &HealthChecker{
		instances:  instances,
		interval:   interval,
		httpClient: &http.Client{Timeout: 3 * time.Second},
		done:       make(chan struct{}),
	}
}

// Start begins periodic health checks.
func (hc *HealthChecker) Start() {
	go func() {
		ticker := time.NewTicker(hc.interval)
		defer ticker.Stop()

		// Run initial check
		hc.checkAll()

		for {
			select {
			case <-ticker.C:
				hc.checkAll()
			case <-hc.done:
				return
			}
		}
	}()
}

// Stop stops the health checker.
func (hc *HealthChecker) Stop() {
	close(hc.done)
}

func (hc *HealthChecker) checkAll() {
	for _, inst := range hc.instances {
		go hc.check(inst)
	}
}

func (hc *HealthChecker) check(inst *ServiceInstance) {
	healthURL := inst.URL.String() + "/health"
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
	if err != nil {
		atomic.StoreUint32(&inst.Healthy, 0)
		return
	}

	resp, err := hc.httpClient.Do(req)
	if err != nil {
		atomic.StoreUint32(&inst.Healthy, 0)
		return
	}
	resp.Body.Close()

	if resp.StatusCode >= 500 {
		atomic.StoreUint32(&inst.Healthy, 0)
		return
	}

	atomic.StoreUint32(&inst.Healthy, 1)
}

// ============================================================================
// Circuit Breaker per Service
// ============================================================================

// CircuitState represents circuit breaker state.
type CircuitState int32

const (
	CircuitClosed   CircuitState = 0
	CircuitOpen     CircuitState = 1
	CircuitHalfOpen CircuitState = 2
)

// ServiceCircuitBreaker implements circuit breaker per service.
type ServiceCircuitBreaker struct {
	state            int32 // atomic CircuitState
	failureCount     int32
	successCount     int32
	failureThreshold int32
	successThreshold int32
	timeout          time.Duration
	lastFailureTime  int64 // atomic, unix nano
	halfOpenMax      int32
	halfOpenCount    int32
}

// NewServiceCircuitBreaker creates a service circuit breaker.
func NewServiceCircuitBreaker(failureThreshold, successThreshold, halfOpenMax int, timeout time.Duration) *ServiceCircuitBreaker {
	return &ServiceCircuitBreaker{
		state:            int32(CircuitClosed),
		failureThreshold: int32(failureThreshold),
		successThreshold: int32(successThreshold),
		halfOpenMax:      int32(halfOpenMax),
		timeout:          timeout,
	}
}

// Allow checks if request is allowed.
func (cb *ServiceCircuitBreaker) Allow() bool {
	for {
		state := CircuitState(atomic.LoadInt32(&cb.state))
		switch state {
		case CircuitClosed:
			return true
		case CircuitOpen:
			lastFail := atomic.LoadInt64(&cb.lastFailureTime)
			if time.Since(time.Unix(0, lastFail)) > cb.timeout {
				// Attempt to transition to half-open
				if atomic.CompareAndSwapInt32(&cb.state, int32(CircuitOpen), int32(CircuitHalfOpen)) {
					atomic.StoreInt32(&cb.halfOpenCount, 0)
					return true
				}
				// Another goroutine won the CAS, re-evaluate
				continue
			}
			return false
		case CircuitHalfOpen:
			count := atomic.AddInt32(&cb.halfOpenCount, 1)
			if count <= cb.halfOpenMax {
				return true
			}
			return false
		}
	}
}

// RecordSuccess records a successful call.
func (cb *ServiceCircuitBreaker) RecordSuccess() {
	state := CircuitState(atomic.LoadInt32(&cb.state))
	if state == CircuitHalfOpen {
		atomic.AddInt32(&cb.successCount, 1)
		if atomic.LoadInt32(&cb.successCount) >= cb.successThreshold {
			atomic.StoreInt32(&cb.state, int32(CircuitClosed))
			atomic.StoreInt32(&cb.failureCount, 0)
			atomic.StoreInt32(&cb.successCount, 0)
		}
	} else if state == CircuitClosed {
		atomic.StoreInt32(&cb.failureCount, 0)
	}
}

// RecordFailure records a failed call.
func (cb *ServiceCircuitBreaker) RecordFailure() {
	state := CircuitState(atomic.LoadInt32(&cb.state))
	if state == CircuitHalfOpen {
		atomic.StoreInt32(&cb.state, int32(CircuitOpen))
		atomic.StoreInt64(&cb.lastFailureTime, time.Now().UnixNano())
		atomic.StoreInt32(&cb.successCount, 0)
	} else if state == CircuitClosed {
		count := atomic.AddInt32(&cb.failureCount, 1)
		if count >= cb.failureThreshold {
			atomic.StoreInt32(&cb.state, int32(CircuitOpen))
			atomic.StoreInt64(&cb.lastFailureTime, time.Now().UnixNano())
		}
	}
}

// State returns the current circuit breaker state.
func (cb *ServiceCircuitBreaker) State() CircuitState {
	return CircuitState(atomic.LoadInt32(&cb.state))
}

// ============================================================================
// Sliding Window Rate Limiter
// ============================================================================

// SlidingWindowEntry tracks request timestamps.
type SlidingWindowEntry struct {
	timestamps []int64 // unix nano timestamps
	mu         sync.Mutex
}

// SlidingWindowRateLimiter implements per-user and per-IP rate limiting.
type SlidingWindowRateLimiter struct {
	mu      sync.RWMutex
	entries map[string]*SlidingWindowEntry
	limit   int
	window  time.Duration
	cleanup time.Duration
}

// NewSlidingWindowRateLimiter creates a new sliding window rate limiter.
func NewSlidingWindowRateLimiter(limit int, window, cleanup time.Duration) *SlidingWindowRateLimiter {
	rl := &SlidingWindowRateLimiter{
		entries: make(map[string]*SlidingWindowEntry),
		limit:   limit,
		window:  window,
		cleanup: cleanup,
	}
	go rl.backgroundCleanup()
	return rl
}

// Allow checks if request is allowed for given key.
func (rl *SlidingWindowRateLimiter) Allow(key string) bool {
	rl.mu.RLock()
	entry, exists := rl.entries[key]
	rl.mu.RUnlock()

	if !exists {
		entry = &SlidingWindowEntry{
			timestamps: make([]int64, 0, rl.limit),
		}
		rl.mu.Lock()
		rl.entries[key] = entry
		rl.mu.Unlock()
	}

	entry.mu.Lock()
	defer entry.mu.Unlock()

	now := time.Now().UnixNano()
	windowNano := rl.window.Nanoseconds()
	cutoff := now - windowNano

	// Remove expired timestamps
	i := 0
	for i < len(entry.timestamps) && entry.timestamps[i] < cutoff {
		i++
	}
	entry.timestamps = entry.timestamps[i:]

	if len(entry.timestamps) >= rl.limit {
		return false
	}

	entry.timestamps = append(entry.timestamps, now)
	return true
}

func (rl *SlidingWindowRateLimiter) backgroundCleanup() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now().UnixNano()
		cutoff := now - (rl.window * 2).Nanoseconds()

		rl.mu.Lock()
		for key, entry := range rl.entries {
			entry.mu.Lock()
			if len(entry.timestamps) == 0 || entry.timestamps[len(entry.timestamps)-1] < cutoff {
				delete(rl.entries, key)
			}
			entry.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

// ============================================================================
// Response Cache
// ============================================================================

// CacheEntry holds a cached response.
type CacheEntry struct {
	Data       []byte
	Headers    map[string]string
	StatusCode int
	ExpiresAt  time.Time
	ETag       string
}

// ResponseCache caches idempotent GET responses.
type ResponseCache struct {
	mu         sync.RWMutex
	entries    map[string]*CacheEntry
	maxSize    int
	defaultTTL time.Duration
	lru        *list.List
	lruMap     map[string]*list.Element
}

type cacheLRUEntry struct {
	key   string
	entry *CacheEntry
}

// NewResponseCache creates a new response cache.
func NewResponseCache(maxSize int, defaultTTL time.Duration) *ResponseCache {
	return &ResponseCache{
		entries:    make(map[string]*CacheEntry),
		maxSize:    maxSize,
		defaultTTL: defaultTTL,
		lru:        list.New(),
		lruMap:     make(map[string]*list.Element),
	}
}

// Get retrieves a cached response.
func (rc *ResponseCache) Get(key string) (*CacheEntry, bool) {
	rc.mu.RLock()
	entry, exists := rc.entries[key]
	rc.mu.RUnlock()

	if !exists {
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		rc.mu.Lock()
		delete(rc.entries, key)
		if el, ok := rc.lruMap[key]; ok {
			rc.lru.Remove(el)
			delete(rc.lruMap, key)
		}
		rc.mu.Unlock()
		return nil, false
	}

	// Move to front (LRU)
	rc.mu.Lock()
	if el, ok := rc.lruMap[key]; ok {
		rc.lru.MoveToFront(el)
	}
	rc.mu.Unlock()

	return entry, true
}

// Set stores a response in cache.
func (rc *ResponseCache) Set(key string, entry *CacheEntry) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	// Evict if full
	for rc.lru.Len() >= rc.maxSize {
		back := rc.lru.Back()
		if back == nil {
			break
		}
		lruEntry := back.Value.(*cacheLRUEntry)
		delete(rc.entries, lruEntry.key)
		delete(rc.lruMap, lruEntry.key)
		rc.lru.Remove(back)
	}

	rc.entries[key] = entry
	el := rc.lru.PushFront(&cacheLRUEntry{key: key, entry: entry})
	rc.lruMap[key] = el
}

// Invalidate removes a cached response.
func (rc *ResponseCache) Invalidate(key string) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	delete(rc.entries, key)
	if el, ok := rc.lruMap[key]; ok {
		rc.lru.Remove(el)
		delete(rc.lruMap, key)
	}
}

func cacheKey(method, path, query, userID string) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%s:%s", method, path, query, userID)))
	return hex.EncodeToString(h[:])
}

// ============================================================================
// Request Deduplication
// ============================================================================

type inflightRequest struct {
	cancelFn context.CancelFunc
	done     chan resultData
}

type resultData struct {
	data []byte
	err  error
}

// RequestDeduplicator prevents duplicate requests from being executed concurrently.
type RequestDeduplicator struct {
	mu               sync.Mutex
	inflightRequests map[string]*inflightRequest
}

func NewRequestDeduplicator() *RequestDeduplicator {
	return &RequestDeduplicator{
		inflightRequests: make(map[string]*inflightRequest),
	}
}

// Execute executes a function, deduplicating by key.
func (rd *RequestDeduplicator) Execute(ctx context.Context, key string, fn func(context.Context) ([]byte, error)) ([]byte, error) {
	rd.mu.Lock()
	if existing, ok := rd.inflightRequests[key]; ok {
		rd.mu.Unlock()
		// Wait for the in-flight request
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case result := <-existing.done:
			return result.data, result.err
		}
	}

	ctx, cancel := context.WithCancel(ctx)
	req := &inflightRequest{
		cancelFn: cancel,
		done:     make(chan resultData, 1),
	}
	rd.inflightRequests[key] = req
	rd.mu.Unlock()

	go func() {
		data, err := fn(ctx)
		req.done <- resultData{data: data, err: err}

		rd.mu.Lock()
		delete(rd.inflightRequests, key)
		rd.mu.Unlock()
		close(req.done)
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-req.done:
		return result.data, result.err
	}
}

// ============================================================================
// Request Aggregation (Batch requests)
// ============================================================================

// BatchRequest represents a request in the batch.
type BatchRequest struct {
	ID      string            `json:"id"`
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Query   string            `json:"query"`
	Body    json.RawMessage   `json:"body,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// BatchResponse represents a response in the batch.
type BatchResponse struct {
	ID         string            `json:"id"`
	StatusCode int               `json:"status_code"`
	Body       json.RawMessage   `json:"body"`
	Headers    map[string]string `json:"headers,omitempty"`
}

// BatchHandler processes batch requests concurrently.
type BatchHandler struct {
	gateway *Gateway
}

func (bh *BatchHandler) HandleBatch(c *gin.Context) {
	var batchReq struct {
		Requests []BatchRequest `json:"requests"`
	}
	if err := c.ShouldBindJSON(&batchReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid batch request"})
		return
	}

	if len(batchReq.Requests) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty batch"})
		return
	}

	responses := make([]BatchResponse, len(batchReq.Requests))
	var wg sync.WaitGroup

	for i, req := range batchReq.Requests {
		wg.Add(1)
		go func(idx int, br BatchRequest) {
			defer wg.Done()

			// Reconstruct full URL path
			_ = br.Query

			// Determine which service to route to
			serviceName := bh.gateway.routeToService(br.Method, br.Path)
			if serviceName == "" {
				responses[idx] = BatchResponse{
					ID:         br.ID,
					StatusCode: http.StatusNotFound,
					Body:       json.RawMessage(`{"error":"no route"}`),
				}
				return
			}

			// Create an in-memory recorder
			rec := &responseRecorder{
				headers: make(http.Header),
				body:    bytes.NewBuffer(nil),
			}

			subReq := c.Request.Clone(context.Background())
			subReq.Method = br.Method
			subReq.URL, _ = url.Parse(br.Path)
			if br.Query != "" {
				subReq.URL.RawQuery = br.Query
			}
			if br.Body != nil {
				subReq.Body = io.NopCloser(bytes.NewReader(br.Body))
			}

			subCtx := &gin.Context{
				Request: subReq,
				Writer:  rec,
			}

			bh.gateway.serveProxy(subCtx, serviceName, "")

			respHeaders := make(map[string]string)
			for k, vals := range rec.headers {
				if len(vals) > 0 {
					respHeaders[k] = vals[0]
				}
			}

			responses[idx] = BatchResponse{
				ID:         br.ID,
				StatusCode: rec.statusCode,
				Body:       rec.body.Bytes(),
				Headers:    respHeaders,
			}
		}(i, req)
	}

	wg.Wait()

	// Sort responses by original order
	sort.Slice(responses, func(i, j int) bool {
		return responses[i].ID < responses[j].ID
	})

	c.JSON(http.StatusOK, gin.H{
		"responses": responses,
	})
}

type responseRecorder struct {
	statusCode int
	headers    http.Header
	body       *bytes.Buffer
}

func (r *responseRecorder) Header() http.Header { return r.headers }

func (r *responseRecorder) Write(data []byte) (int, error) { return r.body.Write(data) }

func (r *responseRecorder) WriteHeader(statusCode int) { r.statusCode = statusCode }

func (r *responseRecorder) CloseNotify() <-chan bool { return make(chan bool) }

func (r *responseRecorder) Flush() {}

func (r *responseRecorder) Written() bool { return r.statusCode != 0 }

func (r *responseRecorder) WriteHeaderNow() {}

func (r *responseRecorder) Size() int { return r.body.Len() }

func (r *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, fmt.Errorf("hijack not supported")
}

func (r *responseRecorder) Pusher() http.Pusher { return nil }

func (r *responseRecorder) Status() int { return r.statusCode }

func (r *responseRecorder) WriteString(s string) (int, error) { return r.body.WriteString(s) }

// ============================================================================
// Connection Pool
// ============================================================================

// PooledTransport provides connection pooling for backend services.
type PooledTransport struct {
	*http.Transport
}

// NewPooledTransport creates a pooled HTTP transport.
func NewPooledTransport(maxIdleConns, maxConnsPerHost int, idleTimeout time.Duration) *PooledTransport {
	return &PooledTransport{
		Transport: &http.Transport{
			MaxIdleConns:        maxIdleConns,
			MaxConnsPerHost:     maxConnsPerHost,
			MaxIdleConnsPerHost: maxIdleConns / 2,
			IdleConnTimeout:     idleTimeout,
			DisableCompression:  false,
		},
	}
}

// ============================================================================
// Gateway
// ============================================================================

// Gateway is the main API gateway.
type Gateway struct {
	router              *gin.Engine
	log                 *logger.Logger
	healthChecker       *HealthChecker
	circuitBreakers     map[string]*ServiceCircuitBreaker
	rateLimiter         *SlidingWindowRateLimiter
	responseCache       *ResponseCache
	deduplicator        *RequestDeduplicator
	batchHandler        *BatchHandler
	transport           *PooledTransport
	services            map[string][]*ServiceInstance
	serviceRoutes       map[string]string // path prefix -> service name
	rrCounters          map[string]*uint64
	cfg                 GatewayConfig
	startTime           time.Time
	requestCount        uint64
	activeConnections   int64
	traceIDGenerator    func() string
	endpointRateLimits  map[string]int
	deprecatedPaths     map[string]time.Time
	jwtConfig           middleware.JWTConfig
}

// GatewayConfig holds the gateway configuration.
type GatewayConfig struct {
	Port                int
	JWTSecret           string
	RateLimitPerMin     int
	RateLimitPerUserMin int
	CircuitThreshold    int
	CircuitSuccess      int
	CircuitTimeout      time.Duration
	CacheMaxSize        int
	CacheDefaultTTL     time.Duration
	MaxIdleConns        int
	MaxConnsPerHost     int
	IdleConnTimeout     time.Duration
	HealthCheckInterval time.Duration
	Services            []ServiceConfig
	EndpointRateLimits  map[string]int
	DeprecatedPaths     map[string]string
}

// ServiceConfig holds configuration for a single service.
type ServiceConfig struct {
	Name    string
	URLs    []string
	Timeout time.Duration
	Weight  int
}

// NewGateway creates a new API gateway.
func NewGateway(cfg GatewayConfig) *Gateway {
	jwtCfg := middleware.NewJWTConfig()
	jwtCfg.AccessTokenTTL = 15 * time.Minute
	jwtCfg.RefreshTokenTTL = 7 * 24 * time.Hour

	gw := &Gateway{
		router:              gin.New(),
		log:                 logger.NewDefault("api-gateway-golang"),
		circuitBreakers:     make(map[string]*ServiceCircuitBreaker),
		rateLimiter:         NewSlidingWindowRateLimiter(cfg.RateLimitPerMin, time.Minute, 5*time.Minute),
		responseCache:       NewResponseCache(cfg.CacheMaxSize, cfg.CacheDefaultTTL),
		deduplicator:        NewRequestDeduplicator(),
		transport:           NewPooledTransport(cfg.MaxIdleConns, cfg.MaxConnsPerHost, cfg.IdleConnTimeout),
		services:            make(map[string][]*ServiceInstance),
		serviceRoutes:       make(map[string]string),
		rrCounters:          make(map[string]*uint64),
		cfg:                 cfg,
		startTime:           time.Now(),
		endpointRateLimits:  make(map[string]int),
		deprecatedPaths:     make(map[string]time.Time),
		traceIDGenerator: func() string {
			b := make([]byte, 16)
			rand.Read(b)
			return hex.EncodeToString(b)
		},
		jwtConfig: jwtCfg,
	}

	// Initialize endpoint rate limits
	for group, limit := range cfg.EndpointRateLimits {
		gw.endpointRateLimits[group] = limit
	}

	// Initialize deprecated paths
	for path, sunsetDate := range cfg.DeprecatedPaths {
		if t, err := time.Parse(time.RFC3339, sunsetDate); err == nil {
			gw.deprecatedPaths[path] = t
		}
	}

	// Initialize services and load balancers
	for _, svc := range cfg.Services {
		counter := uint64(0)
		gw.rrCounters[svc.Name] = &counter
		gw.circuitBreakers[svc.Name] = NewServiceCircuitBreaker(
			cfg.CircuitThreshold,
			cfg.CircuitSuccess,
			3,
			cfg.CircuitTimeout,
		)

		for _, urlStr := range svc.URLs {
			u, err := url.Parse(urlStr)
			if err != nil {
				log.Printf("Warning: invalid URL for service %s: %s", svc.Name, urlStr)
				continue
			}
			inst := &ServiceInstance{
				Name:    svc.Name,
				URL:     u,
				Healthy: 1,
				Weight:  svc.Weight,
			}
			gw.services[svc.Name] = append(gw.services[svc.Name], inst)
		}
	}

	// Collect all instances for health checker
	var allInstances []*ServiceInstance
	for _, instances := range gw.services {
		allInstances = append(allInstances, instances...)
	}
	gw.healthChecker = NewHealthChecker(allInstances, cfg.HealthCheckInterval)
	gw.batchHandler = &BatchHandler{gateway: gw}

	// Setup service routes
	gw.setupServiceRoutes()

	return gw
}

func (gw *Gateway) setupServiceRoutes() {
	gw.serviceRoutes["/api/v1/auth"] = "auth-service"
	gw.serviceRoutes["/api/v1/problems"] = "problem-service"
	gw.serviceRoutes["/api/v1/submissions"] = "execution-service"
	gw.serviceRoutes["/api/v1/leaderboard"] = "leaderboard-service"
	gw.serviceRoutes["/api/v1/hints"] = "hint-service"
	gw.serviceRoutes["/api/v1/users"] = "auth-service"
	gw.serviceRoutes["/api/v1/admin"] = "problem-service"
	gw.serviceRoutes["/api/ws"] = "websocket-service"

	// Also support unversioned routes
	gw.serviceRoutes["/api/auth"] = "auth-service"
	gw.serviceRoutes["/api/problems"] = "problem-service"
	gw.serviceRoutes["/api/submissions"] = "execution-service"
	gw.serviceRoutes["/api/leaderboard"] = "leaderboard-service"
	gw.serviceRoutes["/api/hints"] = "hint-service"
	gw.serviceRoutes["/api/users"] = "auth-service"
	gw.serviceRoutes["/api/admin"] = "problem-service"
}

func (gw *Gateway) routeToService(method, path string) string {
	var matchedService string
	var matchedLen int

	for prefix, service := range gw.serviceRoutes {
		if strings.HasPrefix(path, prefix) && len(prefix) > matchedLen {
			matchedService = service
			matchedLen = len(prefix)
		}
	}
	_ = method

	return matchedService
}

// Setup sets up the gateway routes and middleware.
func (gw *Gateway) Setup() {
	gw.router.Use(gin.Recovery())
	gw.router.Use(gw.traceMiddleware())
	gw.router.Use(gw.requestIDMiddleware())
	gw.router.Use(gw.loggerMiddleware())
	gw.router.Use(gw.corsMiddleware())
	gw.router.Use(gw.rateLimitMiddleware())
	gw.router.Use(gw.requestValidationMiddleware())

	// Health endpoints (no auth)
	gw.router.GET("/health", gw.healthHandler())
	gw.router.GET("/health/liveness", gw.livenessHandler())
	gw.router.GET("/health/readiness", gw.readinessHandler())

	// Status endpoint
	gw.router.GET("/status", gw.statusHandler())

	// Version info
	gw.router.GET("/version", gw.versionHandler())

	// API Routes
	api := gw.router.Group("/api")
	{
		// Versioned API routes
		v1 := api.Group("/v1")
		{
			// Public routes
			v1.POST("/auth/login", gw.proxyTo("auth-service", ""))
			v1.POST("/auth/register", gw.proxyTo("auth-service", ""))
			v1.POST("/auth/refresh", gw.proxyTo("auth-service", ""))
			v1.POST("/auth/logout", gw.proxyTo("auth-service", ""))
			v1.POST("/auth/revoke", gw.proxyTo("auth-service", ""))

			// Protected routes
			protected := v1.Group("")
			protected.Use(gw.jwtAuthMiddleware())
			{
				protected.GET("/problems", gw.cacheMiddleware(), gw.proxyTo("problem-service", "problems:list"))
				protected.GET("/problems/:id", gw.cacheMiddleware(), gw.proxyTo("problem-service", "problems:detail"))
				protected.POST("/submissions", gw.proxyTo("execution-service", ""))
				protected.GET("/submissions/:id", gw.cacheMiddleware(), gw.proxyTo("execution-service", ""))
				protected.GET("/submissions/user/:userId", gw.cacheMiddleware(), gw.proxyTo("execution-service", ""))
				protected.GET("/leaderboard", gw.cacheMiddleware(), gw.proxyTo("leaderboard-service", ""))
				protected.GET("/hints/:problemId", gw.cacheMiddleware(), gw.proxyTo("hint-service", ""))
				protected.GET("/users/me", gw.cacheMiddleware(), gw.proxyTo("auth-service", ""))
			}

			// Admin routes
			admin := v1.Group("/admin")
			admin.Use(gw.jwtAuthMiddleware())
			admin.Use(gw.adminRoleMiddleware())
			{
				admin.POST("/problems", gw.proxyTo("problem-service", ""))
				admin.PUT("/problems/:id", gw.proxyTo("problem-service", ""))
				admin.DELETE("/problems/:id", gw.proxyTo("problem-service", ""))
			}
		}

		// Legacy unversioned routes (backward compatible)
		api.POST("/auth/login", gw.proxyTo("auth-service", ""))
		api.POST("/auth/register", gw.proxyTo("auth-service", ""))
		api.POST("/auth/refresh", gw.proxyTo("auth-service", ""))
		api.POST("/auth/logout", gw.proxyTo("auth-service", ""))
		api.POST("/auth/revoke", gw.proxyTo("auth-service", ""))

		protected := api.Group("")
		protected.Use(gw.jwtAuthMiddleware())
		{
			protected.GET("/problems", gw.cacheMiddleware(), gw.proxyTo("problem-service", "problems:list"))
			protected.GET("/problems/:id", gw.cacheMiddleware(), gw.proxyTo("problem-service", "problems:detail"))
			protected.POST("/submissions", gw.proxyTo("execution-service", ""))
			protected.GET("/submissions/:id", gw.cacheMiddleware(), gw.proxyTo("execution-service", ""))
			protected.GET("/submissions/user/:userId", gw.cacheMiddleware(), gw.proxyTo("execution-service", ""))
			protected.GET("/leaderboard", gw.cacheMiddleware(), gw.proxyTo("leaderboard-service", ""))
			protected.GET("/hints/:problemId", gw.cacheMiddleware(), gw.proxyTo("hint-service", ""))
			protected.GET("/users/me", gw.cacheMiddleware(), gw.proxyTo("auth-service", ""))
		}

		admin := api.Group("/admin")
		admin.Use(gw.jwtAuthMiddleware())
		admin.Use(gw.adminRoleMiddleware())
		{
			admin.POST("/problems", gw.proxyTo("problem-service", ""))
			admin.PUT("/problems/:id", gw.proxyTo("problem-service", ""))
			admin.DELETE("/problems/:id", gw.proxyTo("problem-service", ""))
		}

		// Batch endpoint
		api.POST("/batch", gw.batchHandler.HandleBatch)
	}

	// WebSocket route
	gw.router.GET("/api/ws", gw.proxyTo("websocket-service", ""))
	gw.router.GET("/api/v1/ws", gw.proxyTo("websocket-service", ""))
}

// proxyTo creates a handler that proxies requests to the specified service.
func (gw *Gateway) proxyTo(serviceName, cacheGroup string) gin.HandlerFunc {
	return func(c *gin.Context) {
		gw.serveProxy(c, serviceName, cacheGroup)
	}
}

func (gw *Gateway) serveProxy(c *gin.Context, serviceName, cacheGroup string) {
	atomic.AddUint64(&gw.requestCount, 1)
	atomic.AddInt64(&gw.activeConnections, 1)
	defer atomic.AddInt64(&gw.activeConnections, -1)

	traceID := c.GetString("traceID")

	// Check sunset headers for deprecated paths
	if sunsetTime, ok := gw.deprecatedPaths[c.Request.URL.Path]; ok {
		c.Header("Sunset", sunsetTime.Format(http.TimeFormat))
		c.Header("Deprecation", "true")
		if time.Now().After(sunsetTime) {
			c.JSON(http.StatusGone, gin.H{
				"error": "this API version has been sunset",
				"code":  "API_DEPRECATED",
			})
			return
		}
	}

	// Backward compatibility headers
	c.Header("X-API-Version", "v1")
	c.Header("X-API-Deprecated", "")

	path := c.Request.URL.Path
	if strings.Contains(path, "/v2/") {
		c.Header("X-API-Version", "v2")
	} else if strings.Contains(path, "/v1/") {
		c.Header("X-API-Version", "v1")
	}

	// Check circuit breaker
	cb, ok := gw.circuitBreakers[serviceName]
	if !ok || !cb.Allow() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":    "service temporarily unavailable",
			"code":     "CIRCUIT_OPEN",
			"trace_id": traceID,
		})
		return
	}

	// Select instance (round-robin with health awareness)
	instances, ok := gw.services[serviceName]
	if !ok || len(instances) == 0 {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":    "service not configured",
			"code":     "SERVICE_UNAVAILABLE",
			"trace_id": traceID,
		})
		return
	}

	// Try cache for idempotent GET requests
	if c.Request.Method == http.MethodGet && cacheGroup != "" {
		userID := c.GetString("userID")
		ck := cacheKey(c.Request.Method, c.Request.URL.Path, c.Request.URL.RawQuery, userID)

		if cached, ok := gw.responseCache.Get(ck); ok {
			for k, v := range cached.Headers {
				c.Header(k, v)
			}
			c.Header("X-Cache", "HIT")
			c.Data(cached.StatusCode, cached.Headers["Content-Type"], cached.Data)
			return
		}

		// Check for in-flight duplicate
		dedupKey := fmt.Sprintf("%s:%s", ck, traceID)
		data, err := gw.deduplicator.Execute(c.Request.Context(), dedupKey, func(ctx context.Context) ([]byte, error) {
			result, err := gw.executeProxy(ctx, c, serviceName, instances, cb)
			return result, err
		})
		if err != nil {
			return // response already written
		}

		c.Header("X-Cache", "MISS")
		c.Data(http.StatusOK, "application/json", data)
		return
	}

	gw.executeProxy(c.Request.Context(), c, serviceName, instances, cb)
}

func (gw *Gateway) executeProxy(ctx context.Context, c *gin.Context, serviceName string, instances []*ServiceInstance, cb *ServiceCircuitBreaker) ([]byte, error) {
	// Pick healthy instance (round-robin)
	var selected *ServiceInstance
	counter := atomic.AddUint64(gw.rrCounters[serviceName], 1)

	for i := 0; i < len(instances); i++ {
		idx := int(counter+uint64(i)) % len(instances)
		if atomic.LoadUint32(&instances[idx].Healthy) == 1 {
			selected = instances[idx]
			break
		}
	}

	if selected == nil {
		selected = instances[int(counter)%len(instances)]
	}

	// Create reverse proxy with connection pooling
	proxy := httputil.NewSingleHostReverseProxy(selected.URL)
	proxy.Transport = gw.transport.Transport

	// Modify request
	outReq := c.Request.Clone(ctx)
	outReq.URL.Path = strings.TrimPrefix(outReq.URL.Path, "/api")
	if strings.HasPrefix(outReq.URL.Path, "/v1") || strings.HasPrefix(outReq.URL.Path, "/v2") {
		parts := strings.SplitN(outReq.URL.Path, "/", 3)
		if len(parts) >= 3 {
			outReq.URL.Path = "/" + parts[2]
		}
	}
	if outReq.URL.Path == "" {
		outReq.URL.Path = "/"
	}

	// Add tracing headers
	outReq.Header.Set("X-Trace-ID", c.GetString("traceID"))
	outReq.Header.Set("X-Request-ID", c.GetString("requestID"))
	outReq.Header.Set("X-Forwarded-For", c.ClientIP())
	outReq.Header.Set("X-Forwarded-Proto", "http")

	// Record result for circuit breaker
	proxy.ModifyResponse = func(resp *http.Response) error {
		if resp.StatusCode >= 500 {
			cb.RecordFailure()
		} else {
			cb.RecordSuccess()
		}
		return nil
	}

	// Cache successful GET responses
	if c.Request.Method == http.MethodGet {
		rec := &responseRecorder{
			headers: make(http.Header),
			body:    bytes.NewBuffer(nil),
		}

		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			cb.RecordFailure()
			gw.log.WithRequestID(c.GetString("requestID")).
				WithField("service", serviceName).
				WithField("trace_id", c.GetString("traceID")).
				Error(fmt.Sprintf("proxy error: %v", err))
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(gin.H{
				"error":    "service error",
				"code":     "PROXY_ERROR",
				"trace_id": c.GetString("traceID"),
			})
		}

		proxy.ServeHTTP(rec, outReq)

		// Cache successful responses
		if rec.statusCode >= 200 && rec.statusCode < 300 {
			userID := c.GetString("userID")
			ck := cacheKey(c.Request.Method, c.Request.URL.Path, c.Request.URL.RawQuery, userID)
			headers := make(map[string]string)
			for k, vals := range rec.headers {
				if len(vals) > 0 {
					headers[k] = vals[0]
				}
			}
			gw.responseCache.Set(ck, &CacheEntry{
				Data:       rec.body.Bytes(),
				Headers:    headers,
				StatusCode: rec.statusCode,
				ExpiresAt:  time.Now().Add(gw.cfg.CacheDefaultTTL),
			})
		}

		// Copy headers and body
		for k, vals := range rec.headers {
			for _, v := range vals {
				c.Header(k, v)
			}
		}
		c.Status(rec.statusCode)
		c.Writer.Write(rec.body.Bytes())

		return rec.body.Bytes(), nil
	}

	// Non-GET: use standard proxy
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		cb.RecordFailure()
		gw.log.WithRequestID(c.GetString("requestID")).
			WithField("service", serviceName).
			WithField("trace_id", c.GetString("traceID")).
			Error(fmt.Sprintf("proxy error: %v", err))
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(gin.H{
			"error":    "service error",
			"code":     "PROXY_ERROR",
			"trace_id": c.GetString("traceID"),
		})
	}

	proxy.ServeHTTP(c.Writer, outReq)
	return nil, nil
}

// ============================================================================
// Middleware
// ============================================================================

func (gw *Gateway) traceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = gw.traceIDGenerator()
		}
		c.Set("traceID", traceID)
		c.Header("X-Trace-ID", traceID)
		c.Next()
	}
}

func (gw *Gateway) requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("requestID", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func (gw *Gateway) loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		entry := gw.log.WithRequestID(c.GetString("requestID")).
			WithField("method", c.Request.Method).
			WithField("path", path).
			WithField("status", status).
			WithField("latency", latency.String()).
			WithField("client_ip", c.ClientIP()).
			WithField("trace_id", c.GetString("traceID"))

		if status >= 500 {
			entry.Error(fmt.Sprintf("server error: %s", path))
		} else if status >= 400 {
			entry.Warn(fmt.Sprintf("client error: %s", path))
		} else {
			entry.Info(fmt.Sprintf("completed: %s", path))
		}
	}
}

func (gw *Gateway) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Request-ID, X-Trace-ID")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func (gw *Gateway) rateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Rate limit by IP
		ip := c.ClientIP()
		if !gw.rateLimiter.Allow(fmt.Sprintf("ip:%s", ip)) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":    "rate limit exceeded",
				"code":     "RATE_LIMITED_IP",
				"trace_id": c.GetString("traceID"),
			})
			return
		}

		// Rate limit by user if authenticated
		if userID, exists := c.Get("userID"); exists {
			userLimiter := NewSlidingWindowRateLimiter(gw.cfg.RateLimitPerUserMin, time.Minute, 5*time.Minute)
			if !userLimiter.Allow(fmt.Sprintf("user:%v", userID)) {
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"error":    "rate limit exceeded",
					"code":     "RATE_LIMITED_USER",
					"trace_id": c.GetString("traceID"),
				})
				return
			}
		}

		// Rate limit by endpoint group
		path := c.Request.URL.Path
		for group, limit := range gw.endpointRateLimits {
			if strings.Contains(path, group) {
				epLimiter := NewSlidingWindowRateLimiter(limit, time.Minute, 5*time.Minute)
				if !epLimiter.Allow(fmt.Sprintf("endpoint:%s:%s", group, ip)) {
					c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
						"error":    "endpoint rate limit exceeded",
						"code":     "RATE_LIMITED_ENDPOINT",
						"trace_id": c.GetString("traceID"),
					})
					return
				}
			}
		}

		c.Next()
	}
}

func (gw *Gateway) requestValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate Content-Type for POST/PUT/PATCH
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			contentType := c.GetHeader("Content-Type")
			if contentType != "" && !strings.HasPrefix(contentType, "multipart/form-data") && !strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
				if !strings.HasPrefix(contentType, "application/json") && !strings.HasPrefix(contentType, "application/octet-stream") {
					c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, gin.H{
						"error":    "unsupported content type",
						"code":     "UNSUPPORTED_MEDIA_TYPE",
						"trace_id": c.GetString("traceID"),
					})
					return
				}
			}
		}

		// Limit body size to 10MB
		if c.Request.ContentLength > 10*1024*1024 {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"error":    "request body too large",
				"code":     "BODY_TOO_LARGE",
				"trace_id": c.GetString("traceID"),
			})
			return
		}

		c.Next()
	}
}

func (gw *Gateway) cacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}

func (gw *Gateway) jwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":    "authorization header required",
				"code":     "UNAUTHORIZED",
				"trace_id": c.GetString("traceID"),
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":    "invalid authorization format",
				"code":     "UNAUTHORIZED",
				"trace_id": c.GetString("traceID"),
			})
			return
		}

		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(gw.cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":    "invalid or expired token",
				"code":     "TOKEN_INVALID",
				"trace_id": c.GetString("traceID"),
			})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":    "invalid token claims",
				"code":     "TOKEN_INVALID",
				"trace_id": c.GetString("traceID"),
			})
			return
		}

		// Extract device fingerprint from headers
		deviceFP := c.GetHeader("X-Device-Fingerprint")
		if deviceFP == "" {
			deviceFP = c.GetHeader("User-Agent")
		}
		c.Set("deviceFP", deviceFP)

		c.Set("userID", claims["user_id"])
		c.Set("username", claims["username"])
		c.Set("role", claims["role"])
		c.Next()
	}
}

func (gw *Gateway) adminRoleMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role.(string) != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":    "admin access required",
				"code":     "FORBIDDEN",
				"trace_id": c.GetString("traceID"),
			})
			return
		}
		c.Next()
	}
}

// ============================================================================
// Handlers
// ============================================================================

func (gw *Gateway) healthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		services := make(map[string]interface{})
		overallStatus := "ok"

		for name, instances := range gw.services {
			healthy := 0
			for _, inst := range instances {
				if atomic.LoadUint32(&inst.Healthy) == 1 {
					healthy++
				}
			}

			cb := gw.circuitBreakers[name]
			status := "up"
			if healthy == 0 {
				status = "down"
				overallStatus = "degraded"
			} else if cb != nil && cb.State() != CircuitClosed {
				status = "circuit_open"
				if overallStatus == "ok" {
					overallStatus = "degraded"
				}
			}

			services[name] = gin.H{
				"status":            status,
				"healthy_instances": healthy,
				"total_instances":   len(instances),
				"circuit_state":     cb.State(),
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"service":  "api-gateway-golang",
			"status":   overallStatus,
			"services": services,
			"time":     time.Now().UTC(),
		})
	}
}

func (gw *Gateway) livenessHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "alive",
			"time":   time.Now().UTC(),
		})
	}
}

func (gw *Gateway) readinessHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ready := false
		for _, instances := range gw.services {
			for _, inst := range instances {
				if atomic.LoadUint32(&inst.Healthy) == 1 {
					ready = true
					break
				}
			}
			if ready {
				break
			}
		}

		if !ready {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not_ready",
				"time":   time.Now().UTC(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
			"time":   time.Now().UTC(),
		})
	}
}

func (gw *Gateway) statusHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		circuitStates := make(map[string]string)
		for name, cb := range gw.circuitBreakers {
			switch cb.State() {
			case CircuitClosed:
				circuitStates[name] = "closed"
			case CircuitOpen:
				circuitStates[name] = "open"
			case CircuitHalfOpen:
				circuitStates[name] = "half_open"
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"service":            "api-gateway-golang",
			"status":             "running",
			"uptime":             time.Since(gw.startTime).String(),
			"request_count":      atomic.LoadUint64(&gw.requestCount),
			"active_connections": atomic.LoadInt64(&gw.activeConnections),
			"circuit_breakers":   circuitStates,
			"time":               time.Now().UTC(),
		})
	}
}

func (gw *Gateway) versionHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"versions":   []string{"v1", "v2"},
			"current":    "v1",
			"deprecated": []string{},
			"sunset":     nil,
		})
	}
}

// ============================================================================
// Graceful Shutdown
// ============================================================================

func (gw *Gateway) Run(addr string) error {
	gw.Setup()
	gw.healthChecker.Start()

	srv := &http.Server{
		Addr:    addr,
		Handler: gw.router,
	}

	// Graceful shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		gw.log.Info("Shutting down gateway...")

		// Stop health checker
		gw.healthChecker.Stop()

		// Drain active connections with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			gw.log.Error(fmt.Sprintf("Gateway forced shutdown"), err)
		}

		gw.log.Info("Gateway stopped")
	}()

	gw.log.Infof("API Gateway (Go) starting on %s", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func main() {
	config := GatewayConfig{
		Port:                getEnvInt("PORT", 9100),
		JWTSecret:           getEnv("JWT_ACCESS_SECRET", "dev-access-secret-key"),
		RateLimitPerMin:     getEnvInt("RATE_LIMIT_PER_MIN", 100),
		RateLimitPerUserMin: getEnvInt("RATE_LIMIT_PER_USER_MIN", 60),
		CircuitThreshold:    getEnvInt("CIRCUIT_THRESHOLD", 5),
		CircuitSuccess:      getEnvInt("CIRCUIT_SUCCESS_THRESHOLD", 2),
		CircuitTimeout:      time.Duration(getEnvInt("CIRCUIT_TIMEOUT_SEC", 30)) * time.Second,
		CacheMaxSize:        getEnvInt("CACHE_MAX_SIZE", 1000),
		CacheDefaultTTL:     time.Duration(getEnvInt("CACHE_DEFAULT_TTL_SEC", 30)) * time.Second,
		MaxIdleConns:        getEnvInt("MAX_IDLE_CONNS", 100),
		MaxConnsPerHost:     getEnvInt("MAX_CONNS_PER_HOST", 20),
		IdleConnTimeout:     time.Duration(getEnvInt("IDLE_CONN_TIMEOUT_SEC", 90)) * time.Second,
		HealthCheckInterval: time.Duration(getEnvInt("HEALTH_CHECK_INTERVAL_SEC", 15)) * time.Second,
		Services: []ServiceConfig{
			{
				Name: "auth-service",
				URLs: []string{
					getEnv("AUTH_SERVICE_URL", "http://localhost:9101"),
				},
				Timeout: 10 * time.Second,
				Weight:  1,
			},
			{
				Name: "problem-service",
				URLs: []string{
					getEnv("PROBLEM_SERVICE_URL", "http://localhost:9102"),
				},
				Timeout: 10 * time.Second,
				Weight:  1,
			},
			{
				Name: "execution-service",
				URLs: []string{
					getEnv("EXECUTION_SERVICE_URL", "http://localhost:9103"),
				},
				Timeout: 30 * time.Second,
				Weight:  1,
			},
			{
				Name: "leaderboard-service",
				URLs: []string{
					getEnv("LEADERBOARD_SERVICE_URL", "http://localhost:9104"),
				},
				Timeout: 10 * time.Second,
				Weight:  1,
			},
			{
				Name: "hint-service",
				URLs: []string{
					getEnv("HINT_SERVICE_URL", "http://localhost:9105"),
				},
				Timeout: 10 * time.Second,
				Weight:  1,
			},
			{
				Name: "websocket-service",
				URLs: []string{
					getEnv("WEBSOCKET_SERVICE_URL", "http://localhost:9107"),
				},
				Timeout: 60 * time.Second,
				Weight:  1,
			},
		},
		EndpointRateLimits: map[string]int{
			"/auth/login":    10,
			"/auth/register": 5,
			"/admin":         30,
		},
		DeprecatedPaths: map[string]string{},
	}

	gateway := NewGateway(config)

	if err := gateway.Run(fmt.Sprintf(":%d", config.Port)); err != nil {
		log.Fatalf("Failed to start gateway: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil {
			return n
		}
	}
	return fallback
}

// Ensure imports are used
var _ = list.New
var _ = json.Marshal