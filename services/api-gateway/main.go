package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"coding-challange/pkg/logger"
)

// ServiceEndpoint represents a backend service.
type ServiceEndpoint struct {
	Name    string
	URL    *url.URL
	healthy uint32 // atomic boolean
}

// CircuitState represents the state of a circuit breaker.
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern.
type CircuitBreaker struct {
	mu               sync.RWMutex
	state            CircuitState
	failureCount     int
	successCount     int
	failureThreshold int
	successThreshold int
	timeout          time.Duration
	lastFailureTime  time.Time
}

// NewCircuitBreaker creates a new circuit breaker.
func NewCircuitBreaker(failureThreshold, successThreshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            StateClosed,
		failureThreshold: failureThreshold,
		successThreshold: successThreshold,
		timeout:          timeout,
	}
}

// Allow checks if a request is allowed through the circuit breaker.
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		if time.Since(cb.lastFailureTime) > cb.timeout {
			return true // Will transition to half-open
		}
		return false
	case StateHalfOpen:
		return true
	}
	return false
}

// RecordSuccess records a successful request.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.successCount++
	if cb.state == StateHalfOpen && cb.successCount >= cb.successThreshold {
		cb.state = StateClosed
		cb.failureCount = 0
		cb.successCount = 0
	} else if cb.state == StateClosed {
		cb.failureCount = 0
	}
}

// RecordFailure records a failed request.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.state == StateHalfOpen {
		cb.state = StateOpen
		cb.successCount = 0
	} else if cb.state == StateClosed && cb.failureCount >= cb.failureThreshold {
		cb.state = StateOpen
	}
}

// State returns the current state.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if cb.state == StateOpen && time.Since(cb.lastFailureTime) > cb.timeout {
		return StateHalfOpen
	}
	return cb.state
}

// RateLimiter implements token bucket rate limiting per key.
type RateLimiter struct {
	mu       sync.RWMutex
	visitors map[string]*rateLimitEntry
	limit    int
	window   time.Duration
}

type rateLimitEntry struct {
	tokens    int
	lastCheck time.Time
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*rateLimitEntry),
		limit:    limit,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

// Allow checks if a request is allowed for the given key.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.visitors[key]

	if !exists {
		rl.visitors[key] = &rateLimitEntry{
			tokens:    rl.limit - 1,
			lastCheck: now,
		}
		return true
	}

	// Refill tokens based on time elapsed
	elapsed := now.Sub(entry.lastCheck)
	tokensToAdd := int(int64(elapsed) * int64(rl.limit) / int64(rl.window))
	if tokensToAdd > 0 {
		entry.tokens = min(entry.tokens+tokensToAdd, rl.limit)
		entry.lastCheck = now
	}

	if entry.tokens > 0 {
		entry.tokens--
		return true
	}

	return false
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, entry := range rl.visitors {
			if now.Sub(entry.lastCheck) > rl.window*2 {
				delete(rl.visitors, key)
			}
		}
		rl.mu.Unlock()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// LoadBalancer implements round-robin load balancing.
type LoadBalancer struct {
	mu        sync.Mutex
	endpoints []*ServiceEndpoint
	current   int
}

// NewLoadBalancer creates a new load balancer.
func NewLoadBalancer(endpoints []*ServiceEndpoint) *LoadBalancer {
	return &LoadBalancer{
		endpoints: endpoints,
	}
}

// Next returns the next healthy endpoint.
func (lb *LoadBalancer) Next() *ServiceEndpoint {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	for i := 0; i < len(lb.endpoints); i++ {
		idx := lb.current % len(lb.endpoints)
		lb.current++

		if atomic.LoadUint32(&lb.endpoints[idx].healthy) == 1 {
			return lb.endpoints[idx]
		}
	}

	// If no healthy endpoints, try the current one anyway
	if len(lb.endpoints) > 0 {
		return lb.endpoints[lb.current%len(lb.endpoints)]
	}
	return nil
}

// UpdateHealth updates the health status of an endpoint.
func (lb *LoadBalancer) UpdateHealth(name string, healthy bool) {
	for _, ep := range lb.endpoints {
		if ep.Name == name {
			var val uint32
			if healthy {
				val = 1
			}
			atomic.StoreUint32(&ep.healthy, val)
			return
		}
	}
}

// GetHealthyEndpoints returns all healthy endpoints.
func (lb *LoadBalancer) GetHealthyEndpoints() []*ServiceEndpoint {
	var healthy []*ServiceEndpoint
	for _, ep := range lb.endpoints {
		if atomic.LoadUint32(&ep.healthy) == 1 {
			healthy = append(healthy, ep)
		}
	}
	return healthy
}

// API Gateway

// Gateway represents the API gateway.
type Gateway struct {
	router         *gin.Engine
	log            *logger.Logger
	loadBalancers  map[string]*LoadBalancer
	circuitBreakers map[string]*CircuitBreaker
	rateLimiter    *RateLimiter
	jwtSecret      string
}

// ServiceConfig holds configuration for a single service.
type ServiceConfig struct {
	Name    string   `json:"name"`
	URLs    []string `json:"urls"`
	Timeout int      `json:"timeout"`
}

// GatewayConfig holds the gateway configuration.
type GatewayConfig struct {
	Port             int              `json:"port"`
	JWTSecret        string           `json:"jwt_secret"`
	RateLimitPerMin  int              `json:"rate_limit_per_min"`
	CircuitThreshold int              `json:"circuit_threshold"`
	CircuitTimeout   int              `json:"circuit_timeout_sec"`
	Services         []ServiceConfig  `json:"services"`
}

// NewGateway creates a new API gateway.
func NewGateway(config GatewayConfig) *Gateway {
	gw := &Gateway{
		router:          gin.New(),
		log:             logger.NewDefault("api-gateway"),
		loadBalancers:   make(map[string]*LoadBalancer),
		circuitBreakers: make(map[string]*CircuitBreaker),
		rateLimiter:     NewRateLimiter(config.RateLimitPerMin, time.Minute),
		jwtSecret:       config.JWTSecret,
	}

	// Initialize load balancers for each service
	for _, svc := range config.Services {
		var endpoints []*ServiceEndpoint
		for _, urlStr := range svc.URLs {
			u, err := url.Parse(urlStr)
			if err != nil {
				log.Printf("Warning: invalid URL for service %s: %s", svc.Name, urlStr)
				continue
			}
			endpoints = append(endpoints, &ServiceEndpoint{
				Name:    svc.Name,
				URL:    u,
				healthy: 1, // Start as healthy
			})
		}
		gw.loadBalancers[svc.Name] = NewLoadBalancer(endpoints)
		gw.circuitBreakers[svc.Name] = NewCircuitBreaker(
			config.CircuitThreshold,
			2,
			time.Duration(config.CircuitTimeout)*time.Second,
		)
	}

	return gw
}

// Setup sets up the gateway routes and middleware.
func (gw *Gateway) Setup() {
	gw.router.Use(gin.Recovery())
	gw.router.Use(gw.requestIDMiddleware())
	gw.router.Use(gw.loggerMiddleware())
	gw.router.Use(gw.corsMiddleware())
	gw.router.Use(gw.rateLimitMiddleware())

	// Health endpoints (no auth required)
	gw.router.GET("/health", gw.healthHandler())
	gw.router.GET("/health/liveness", gw.livenessHandler())
	gw.router.GET("/health/readiness", gw.readinessHandler())

	// Status endpoint
	gw.router.GET("/status", gw.statusHandler())

	// Setup service routes
	api := gw.router.Group("/api")
	{
		// Public routes (no auth)
		api.POST("/auth/login", gw.proxyTo("auth-service"))
		api.POST("/auth/register", gw.proxyTo("auth-service"))
		api.POST("/auth/refresh", gw.proxyTo("auth-service"))

		// Protected routes
		protected := api.Group("")
		protected.Use(gw.jwtAuthMiddleware())
		{
			protected.GET("/problems", gw.proxyTo("problem-service"))
			protected.GET("/problems/:id", gw.proxyTo("problem-service"))
			protected.POST("/submissions", gw.proxyTo("execution-service"))
			protected.GET("/submissions/:id", gw.proxyTo("execution-service"))
			protected.GET("/submissions/user/:userId", gw.proxyTo("execution-service"))
			protected.GET("/leaderboard", gw.proxyTo("leaderboard-service"))
			protected.GET("/hints/:problemId", gw.proxyTo("hint-service"))
			protected.GET("/users/me", gw.proxyTo("auth-service"))
		}

		// Admin routes
		admin := api.Group("/admin")
		admin.Use(gw.jwtAuthMiddleware())
		admin.Use(gw.adminRoleMiddleware())
		{
			admin.POST("/problems", gw.proxyTo("problem-service"))
			admin.PUT("/problems/:id", gw.proxyTo("problem-service"))
			admin.DELETE("/problems/:id", gw.proxyTo("problem-service"))
		}
	}
}

// proxyTo creates a handler that proxies requests to the specified service.
func (gw *Gateway) proxyTo(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		lb, ok := gw.loadBalancers[serviceName]
		if !ok {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "service not configured",
				"code":  "SERVICE_UNAVAILABLE",
			})
			return
		}

		cb, ok := gw.circuitBreakers[serviceName]
		if !ok || !cb.Allow() {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "service temporarily unavailable",
				"code":  "CIRCUIT_OPEN",
			})
			return
		}

		endpoint := lb.Next()
		if endpoint == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "no healthy instances",
				"code":  "NO_INSTANCES",
			})
			return
		}

		// Create reverse proxy
		proxy := httputil.NewSingleHostReverseProxy(endpoint.URL)
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			cb.RecordFailure()
			lb.UpdateHealth(serviceName, false)
			gw.log.WithRequestID(c.GetString("requestID")).
				WithField("service", serviceName).
				Error(fmt.Sprintf("proxy error: %v", err))
			c.JSON(http.StatusBadGateway, gin.H{
				"error": "service error",
				"code":  "PROXY_ERROR",
			})
		}

		proxy.ModifyResponse = func(resp *http.Response) error {
			if resp.StatusCode >= 500 {
				cb.RecordFailure()
			} else {
				cb.RecordSuccess()
				lb.UpdateHealth(serviceName, true)
			}
			return nil
		}

		// Forward request
		c.Request.URL.Path = strings.TrimPrefix(c.Request.URL.Path, "/api")
		if c.Request.URL.Path == "" {
			c.Request.URL.Path = "/"
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

// Middleware

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
			WithField("client_ip", c.ClientIP())

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
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Request-ID")
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
				"error": "rate limit exceeded",
				"code":  "RATE_LIMITED",
			})
			return
		}

		// Rate limit by user if authenticated
		if userID, exists := c.Get("userID"); exists {
			if !gw.rateLimiter.Allow(fmt.Sprintf("user:%v", userID)) {
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
					"error": "rate limit exceeded",
					"code":  "RATE_LIMITED",
				})
				return
			}
		}

		c.Next()
	}
}

func (gw *Gateway) jwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header required",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization format",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(gw.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
				"code":  "TOKEN_INVALID",
			})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token claims",
				"code":  "TOKEN_INVALID",
			})
			return
		}

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
				"error": "admin access required",
				"code":  "FORBIDDEN",
			})
			return
		}
		c.Next()
	}
}

// Health handlers

func (gw *Gateway) healthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		services := make(map[string]interface{})
		overallStatus := "ok"

		for name, lb := range gw.loadBalancers {
			healthy := lb.GetHealthyEndpoints()
			cb := gw.circuitBreakers[name]

			status := "up"
			if len(healthy) == 0 {
				status = "down"
				overallStatus = "degraded"
			} else if cb != nil && cb.State() == StateOpen {
				status = "circuit_open"
				overallStatus = "degraded"
			}

			// Try to reach the service health endpoint
			if len(healthy) > 0 {
				ep := healthy[0]
				if gw.checkServiceHealth(ctx, ep) != nil {
					status = "unreachable"
					if overallStatus == "ok" {
						overallStatus = "degraded"
					}
				}
			}

			services[name] = gin.H{
				"status":          status,
				"healthy_instances": len(healthy),
				"total_instances":  len(lb.endpoints),
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"service":  "api-gateway",
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
		// Check if at least some services are available
		ready := false
		for _, lb := range gw.loadBalancers {
			if len(lb.GetHealthyEndpoints()) > 0 {
				ready = true
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
			case StateClosed:
				circuitStates[name] = "closed"
			case StateOpen:
				circuitStates[name] = "open"
			case StateHalfOpen:
				circuitStates[name] = "half_open"
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"service":         "api-gateway",
			"status":          "running",
			"circuit_breakers": circuitStates,
			"time":            time.Now().UTC(),
		})
	}
}

func (gw *Gateway) checkServiceHealth(ctx context.Context, ep *ServiceEndpoint) error {
	client := &http.Client{Timeout: 3 * time.Second}
	url := ep.URL.String() + "/health"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return fmt.Errorf("service returned status %d", resp.StatusCode)
	}

	return nil
}

// StartHealthMonitor starts monitoring service health.
func (gw *Gateway) StartHealthMonitor(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			for name, lb := range gw.loadBalancers {
				for _, ep := range lb.endpoints {
					ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
					err := gw.checkServiceHealth(ctx, ep)
					cancel()
					lb.UpdateHealth(name, err == nil)
				}
			}
		}
	}()
}

// Run starts the gateway server.
func (gw *Gateway) Run(addr string) error {
	gw.Setup()
	gw.StartHealthMonitor(15 * time.Second)
	gw.log.Infof("API Gateway starting on %s", addr)
	return gw.router.Run(addr)
}

func main() {
	config := GatewayConfig{
		Port:             getEnvInt("PORT", 9100),
		JWTSecret:        getEnv("JWT_ACCESS_SECRET", "dev-access-secret-key"),
		RateLimitPerMin:  getEnvInt("RATE_LIMIT_PER_MIN", 100),
		CircuitThreshold: getEnvInt("CIRCUIT_THRESHOLD", 5),
		CircuitTimeout:   getEnvInt("CIRCUIT_TIMEOUT_SEC", 30),
		Services: []ServiceConfig{
			{
				Name: "auth-service",
				URLs: []string{
					getEnv("AUTH_SERVICE_URL", "http://localhost:9101"),
				},
			},
			{
				Name: "problem-service",
				URLs: []string{
					getEnv("PROBLEM_SERVICE_URL", "http://localhost:9102"),
				},
			},
			{
				Name: "execution-service",
				URLs: []string{
					getEnv("EXECUTION_SERVICE_URL", "http://localhost:9103"),
				},
			},
			{
				Name: "leaderboard-service",
				URLs: []string{
					getEnv("LEADERBOARD_SERVICE_URL", "http://localhost:9104"),
				},
			},
			{
				Name: "hint-service",
				URLs: []string{
					getEnv("HINT_SERVICE_URL", "http://localhost:9105"),
				},
			},
			{
				Name: "websocket-service",
				URLs: []string{
					getEnv("WEBSOCKET_SERVICE_URL", "http://localhost:9107"),
				},
			},
		},
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

// Ensure json is used
var _ = json.Marshal
