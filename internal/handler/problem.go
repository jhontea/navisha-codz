package handler

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"coding-challange/internal/model"
	"coding-challange/internal/service"
)

// ProblemHandler handles HTTP requests for problems.
type ProblemHandler struct {
	problemSvc *service.ProblemService
	runnerSvc  *service.RunnerService
	hintSvc    *service.HintService
	rateLimiter *RateLimiter
}

// NewProblemHandler creates a new problem handler.
func NewProblemHandler(problemSvc *service.ProblemService, runnerSvc *service.RunnerService, hintSvc *service.HintService) *ProblemHandler {
	return &ProblemHandler{
		problemSvc:  problemSvc,
		runnerSvc:   runnerSvc,
		hintSvc:     hintSvc,
		rateLimiter: NewRateLimiter(),
	}
}

// RateLimiter provides simple in-memory rate limiting.
type RateLimiter struct {
	mu       sync.RWMutex
	visitors map[string]*visitor
}

type visitor struct {
	lastSeen  time.Time
	count     int
	windowStart time.Time
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string]*visitor),
	}
}

// isAllowed checks if a request from the given IP is allowed under the rate limit.
// Returns true if allowed, false if rate limited.
func (rl *RateLimiter) isAllowed(ip string, limit int, window time.Duration) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[ip]

	if !exists || now.Sub(v.windowStart) > window {
		// New window
		rl.visitors[ip] = &visitor{
			lastSeen:    now,
			count:       1,
			windowStart: now,
		}
		return true
	}

	if v.count >= limit {
		return false
	}

	v.count++
	v.lastSeen = now
	return true
}

// getClientIP extracts the client IP from the request.
func getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header first
	xff := c.GetHeader("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-Ip header
	xri := c.GetHeader("X-Real-Ip")
	if xri != "" {
		return xri
	}

	// Fall back to connection IP
	ip, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
	return ip
}

// writeMeta creates a Meta struct for responses.
func writeMeta() model.Meta {
	return model.Meta{
		RequestID: fmt.Sprintf("req-%d", time.Now().UnixNano()),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// writeSuccessResponse writes a standard success response with data and meta.
func writeSuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, model.APIResponse{
		Data: data,
		Meta: writeMeta(),
	})
}

// writeErrorResponse writes a standard error response.
func writeErrorResponse(c *gin.Context, code int, errType, message string) {
	c.JSON(code, model.APIErrorResponse{
		Error: &model.APIErrorDetail{
			Message: message,
			Code:    code,
			Type:    errType,
		},
		Meta: writeMeta(),
	})
}

// ListProblems handles GET /api/problems
func (h *ProblemHandler) ListProblems(c *gin.Context) {
	difficulty := c.Query("difficulty")
	category := c.Query("category")
	problemType := c.Query("type")

	// Validate difficulty if provided
	if difficulty != "" {
		validDifficulties := map[string]bool{"easy": true, "medium": true, "hard": true}
		if !validDifficulties[difficulty] {
			writeErrorResponse(c, http.StatusBadRequest, "validation_error",
				fmt.Sprintf("invalid difficulty: %q (must be easy, medium, or hard)", difficulty))
			return
		}
	}

	// Validate type if provided
	if problemType != "" {
		validTypes := map[string]bool{"function": true, "main": true}
		if !validTypes[problemType] {
			writeErrorResponse(c, http.StatusBadRequest, "validation_error",
				fmt.Sprintf("invalid type: %q (must be function or main)", problemType))
			return
		}
	}

	// Rate limit: 60 req/min for list endpoints
	ip := getClientIP(c)
	if !h.rateLimiter.isAllowed(ip, 60, time.Minute) {
		writeErrorResponse(c, http.StatusTooManyRequests, "rate_limit_error",
			"rate limit exceeded, try again later")
		return
	}

	problems := h.problemSvc.ListProblems(difficulty, category)

	// Filter by type if specified (post-filter since repo doesn't support it)
	if problemType != "" {
		filtered := make([]model.ProblemSummary, 0, len(problems))
		for _, p := range problems {
			if p.Type == problemType {
				filtered = append(filtered, p)
			}
		}
		problems = filtered
	}

	writeSuccessResponse(c, problems)
}

// GetProblem handles GET /api/problems/:id
func (h *ProblemHandler) GetProblem(c *gin.Context) {
	id := c.Param("id")

	if err := service.SanitizeProblemID(id); err != nil {
		writeErrorResponse(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	problem, err := h.problemSvc.GetProblemForAPI(id)
	if err != nil {
		writeErrorResponse(c, http.StatusNotFound, "not_found", err.Error())
		return
	}

	writeSuccessResponse(c, problem)
}

// GetTemplate handles GET /api/problems/:id/template
func (h *ProblemHandler) GetTemplate(c *gin.Context) {
	id := c.Param("id")

	if err := service.SanitizeProblemID(id); err != nil {
		writeErrorResponse(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	template, err := h.problemSvc.GetTemplate(id)
	if err != nil {
		writeErrorResponse(c, http.StatusNotFound, "not_found", err.Error())
		return
	}

	writeSuccessResponse(c, template)
}

// RunCode handles POST /api/problems/:id/run
func (h *ProblemHandler) RunCode(c *gin.Context) {
	id := c.Param("id")

	if err := service.SanitizeProblemID(id); err != nil {
		writeErrorResponse(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	// Rate limit: 10 req/min for run endpoint
	ip := getClientIP(c)
	if !h.rateLimiter.isAllowed(ip+":run", 10, time.Minute) {
		writeErrorResponse(c, http.StatusTooManyRequests, "rate_limit_error",
			"rate limit exceeded, try again in 30 seconds")
		return
	}

	var req struct {
		Code string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		writeErrorResponse(c, http.StatusBadRequest, "validation_error", "code field is required")
		return
	}

	if strings.TrimSpace(req.Code) == "" {
		writeErrorResponse(c, http.StatusBadRequest, "validation_error", "code field cannot be empty")
		return
	}

	if err := service.ValidateCodeSize(req.Code); err != nil {
		writeErrorResponse(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	problem, err := h.problemSvc.GetProblem(id)
	if err != nil {
		writeErrorResponse(c, http.StatusNotFound, "not_found", err.Error())
		return
	}

	result := h.runnerSvc.RunCode(req.Code, problem)
	writeSuccessResponse(c, result)
}

// ValidateCode handles POST /api/validate
func (h *ProblemHandler) ValidateCode(c *gin.Context) {
	// Rate limit: 30 req/min for validate endpoint
	ip := getClientIP(c)
	if !h.rateLimiter.isAllowed(ip+":validate", 30, time.Minute) {
		writeErrorResponse(c, http.StatusTooManyRequests, "rate_limit_error",
			"rate limit exceeded, try again later")
		return
	}

	var req struct {
		Code       string `json:"code" binding:"required"`
		ProblemID string `json:"problem_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		writeErrorResponse(c, http.StatusBadRequest, "validation_error", "code field is required")
		return
	}

	if strings.TrimSpace(req.Code) == "" {
		writeErrorResponse(c, http.StatusBadRequest, "validation_error", "code field cannot be empty")
		return
	}

	result := h.problemSvc.ValidateCode(req.Code, req.ProblemID)
	writeSuccessResponse(c, result)
}

// GetHints handles GET /api/problems/:id/hints
func (h *ProblemHandler) GetHints(c *gin.Context) {
	id := c.Param("id")

	if err := service.SanitizeProblemID(id); err != nil {
		writeErrorResponse(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	problem, err := h.problemSvc.GetProblem(id)
	if err != nil {
		writeErrorResponse(c, http.StatusNotFound, "not_found", err.Error())
		return
	}

	hints := h.hintSvc.GetHints(problem)
	writeSuccessResponse(c, gin.H{"hints": hints})
}

// HealthCheck handles GET /health
func (h *ProblemHandler) HealthCheck(c *gin.Context) {
	writeSuccessResponse(c, gin.H{
		"status":  "ok",
		"version": "1.1.0",
	})
}
