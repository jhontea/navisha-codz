package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Status represents the health status of a component.
type Status string

const (
	StatusUp   Status = "up"
	StatusDown Status = "down"
	StatusWarn Status = "warn"
)

// CheckResult represents the result of a single health check.
type CheckResult struct {
	Name      string                 `json:"name"`
	Status    Status                 `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Duration  time.Duration          `json:"duration_ms"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// Checker defines the interface for health check components.
type Checker interface {
	Name() string
	CheckLiveness(ctx context.Context) error
	CheckReadiness(ctx context.Context) error
}

// Response represents the full health check response.
type Response struct {
	Status    Status                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Uptime    time.Duration          `json:"uptime,omitempty"`
	Checks    []CheckResult          `json:"checks"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Manager manages health checks for all services.
type Manager struct {
	mu        sync.RWMutex
	checkers  map[string]Checker
	startTime time.Time
}

// NewManager creates a new health check manager.
func NewManager() *Manager {
	return &Manager{
		checkers:  make(map[string]Checker),
		startTime: time.Now(),
	}
}

// Register registers a new health checker.
func (m *Manager) Register(checker Checker) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checkers[checker.Name()] = checker
}

// Deregister removes a health checker.
func (m *Manager) Deregister(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.checkers, name)
}

// CheckLiveness performs all liveness checks.
func (m *Manager) CheckLiveness(ctx context.Context) Response {
	return m.runChecks(ctx, false)
}

// CheckReadiness performs all readiness checks.
func (m *Manager) CheckReadiness(ctx context.Context) Response {
	return m.runChecks(ctx, true)
}

func (m *Manager) runChecks(ctx context.Context, readiness bool) Response {
	m.mu.RLock()
	checkers := make([]Checker, 0, len(m.checkers))
	for _, c := range m.checkers {
		checkers = append(checkers, c)
	}
	m.mu.RUnlock()

	results := make([]CheckResult, len(checkers))
	var wg sync.WaitGroup
	var overallStatus Status = StatusUp

	for i, checker := range checkers {
		wg.Add(1)
		go func(idx int, c Checker) {
			defer wg.Done()
			start := time.Now()
			var err error
			if readiness {
				err = c.CheckReadiness(ctx)
			} else {
				err = c.CheckLiveness(ctx)
			}

			status := StatusUp
			msg := "ok"
			if err != nil {
				status = StatusDown
				msg = err.Error()
				overallStatus = StatusDown
			}

			results[idx] = CheckResult{
				Name:      c.Name(),
				Status:    status,
				Message:   msg,
				Duration:  time.Since(start),
				Timestamp: time.Now().UTC(),
			}
		}(i, checker)
	}

	wg.Wait()

	// If any critical check failed, mark as down
	for _, r := range results {
		if r.Status == StatusDown {
			overallStatus = StatusDown
			break
		}
	}

	return Response{
		Status:    overallStatus,
		Timestamp: time.Now().UTC(),
		Uptime:    time.Since(m.startTime),
		Checks:    results,
	}
}

// HTTPHandler returns an HTTP handler for K8s-compatible health endpoints.
func (m *Manager) HTTPHandler() *Handler {
	return &Handler{manager: m}
}

// Handler provides HTTP handlers for health endpoints.
type Handler struct {
	manager *Manager
}

// LivenessHandler handles K8s liveness probe requests.
// Returns 200 if the process is alive (not deadlocked).
func (h *Handler) LivenessHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	result := h.manager.CheckLiveness(ctx)
	writeResponse(w, result, result.Status == StatusUp)
}

// ReadinessHandler handles K8s readiness probe requests.
// Returns 200 if the service is ready to accept traffic.
func (h *Handler) ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	result := h.manager.CheckReadiness(ctx)
	writeResponse(w, result, result.Status == StatusUp)
}

// HealthHandler handles general health requests (aggregated).
func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	result := h.manager.CheckReadiness(ctx)
	writeResponse(w, result, result.Status == StatusUp)
}

func writeResponse(w http.ResponseWriter, result Response, healthy bool) {
	w.Header().Set("Content-Type", "application/json")
	if !healthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(result)
}

// Common checkers

// DBChecker checks database health.
type DBChecker struct {
	name string
	ping func(ctx context.Context) error
}

// NewDBChecker creates a new database health checker.
func NewDBChecker(name string, ping func(ctx context.Context) error) *DBChecker {
	return &DBChecker{name: name, ping: ping}
}

func (c *DBChecker) Name() string { return c.name }

func (c *DBChecker) CheckLiveness(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.ping(ctx)
}

func (c *DBChecker) CheckReadiness(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.ping(ctx)
}

// RedisChecker checks Redis health.
type RedisChecker struct {
	name string
	ping func(ctx context.Context) error
}

// NewRedisChecker creates a new Redis health checker.
func NewRedisChecker(name string, ping func(ctx context.Context) error) *RedisChecker {
	return &RedisChecker{name: name, ping: ping}
}

func (c *RedisChecker) Name() string { return c.name }

func (c *RedisChecker) CheckLiveness(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.ping(ctx)
}

func (c *RedisChecker) CheckReadiness(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.ping(ctx)
}

// RabbitMQChecker checks RabbitMQ health.
type RabbitMQChecker struct {
	name      string
	isHealthy func(ctx context.Context) error
}

// NewRabbitMQChecker creates a new RabbitMQ health checker.
func NewRabbitMQChecker(name string, isHealthy func(ctx context.Context) error) *RabbitMQChecker {
	return &RabbitMQChecker{name: name, isHealthy: isHealthy}
}

func (c *RabbitMQChecker) Name() string { return c.name }

func (c *RabbitMQChecker) CheckLiveness(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.isHealthy(ctx)
}

func (c *RabbitMQChecker) CheckReadiness(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.isHealthy(ctx)
}

// FunctionChecker wraps a simple function as a health checker.
type FunctionChecker struct {
	name      string
	checkFunc func(ctx context.Context) error
}

// NewFunctionChecker creates a new function-based health checker.
func NewFunctionChecker(name string, checkFunc func(ctx context.Context) error) *FunctionChecker {
	return &FunctionChecker{name: name, checkFunc: checkFunc}
}

func (c *FunctionChecker) Name() string { return c.name }

func (c *FunctionChecker) CheckLiveness(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.checkFunc(ctx)
}

func (c *FunctionChecker) CheckReadiness(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return c.checkFunc(ctx)
}

// Ensure interfaces are satisfied
var _ Checker = (*DBChecker)(nil)
var _ Checker = (*RedisChecker)(nil)
var _ Checker = (*RabbitMQChecker)(nil)
var _ Checker = (*FunctionChecker)(nil)

// Compile-time check for fmt (used in error messages)
var _ = fmt.Sprintf
