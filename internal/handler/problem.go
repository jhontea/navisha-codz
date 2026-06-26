package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"coding-challange/internal/model"
	"coding-challange/internal/service"
	"coding-challange/pkg/middleware"
)

// ProblemHandler handles HTTP requests for problems.
type ProblemHandler struct {
	problemSvc  *service.ProblemService
	hintSvc    *service.HintService
	rateLimiter *middleware.RateLimiter
}

// NewProblemHandler creates a new problem handler.
func NewProblemHandler(problemSvc *service.ProblemService, hintSvc *service.HintService) *ProblemHandler {
	return &ProblemHandler{
		problemSvc:  problemSvc,
		hintSvc:     hintSvc,
		rateLimiter: middleware.NewRateLimiter(0, time.Minute),
	}
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
// @Summary      List all problems
// @Description  Get a paginated list of coding challenge problems with optional filters
// @Tags         problems
// @Accept       json
// @Produce      json
// @Param        difficulty  query     string  false  "Filter by difficulty (easy, medium, hard)"
// @Param        category    query     string  false  "Filter by category (e.g. array, string)"
// @Param        type        query     string  false  "Filter by type (function, main)"
// @Param        tags        query     string  false  "Filter by tags (comma-separated)"
// @Success      200  {object}  model.APIResponse{data=[]model.ProblemSummary}
// @Failure      400  {object}  model.APIErrorResponse
// @Failure      429  {object}  model.APIErrorResponse
// @Failure      500  {object}  model.APIErrorResponse
// @Router       /api/problems [get]
func (h *ProblemHandler) ListProblems(c *gin.Context) {
	difficulty := c.Query("difficulty")
	category := c.Query("category")
	problemType := c.Query("type")
	tagsParam := c.Query("tags")

	// Parse tags query param (comma-separated, trimmed)
	var tags []string
	if tagsParam != "" {
		for _, t := range strings.Split(tagsParam, ",") {
			trimmed := strings.TrimSpace(t)
			if trimmed != "" {
				tags = append(tags, trimmed)
			}
		}
	}

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
	ip := c.ClientIP()
	if !h.rateLimiter.AllowWithLimit(ip, 60, time.Minute) {
		writeErrorResponse(c, http.StatusTooManyRequests, "rate_limit_error",
			"rate limit exceeded, try again later")
		return
	}

	problems := h.problemSvc.ListProblems(difficulty, category, tags)

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
// @Summary      Get a problem by ID
// @Description  Get full details of a specific coding challenge problem by its ID
// @Tags         problems
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Problem ID"
// @Success      200  {object}  model.APIResponse{data=model.Problem}
// @Failure      400  {object}  model.APIErrorResponse
// @Failure      404  {object}  model.APIErrorResponse
// @Failure      500  {object}  model.APIErrorResponse
// @Router       /api/problems/{id} [get]
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
// @Summary      Get problem template code
// @Description  Get the template/starter code for a specific problem
// @Tags         problems
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Problem ID"
// @Success      200  {object}  model.APIResponse{data=model.Problem}
// @Failure      400  {object}  model.APIErrorResponse
// @Failure      404  {object}  model.APIErrorResponse
// @Failure      500  {object}  model.APIErrorResponse
// @Router       /api/problems/{id}/template [get]
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

// ValidateCodeRequest is the request body for code validation.
type ValidateCodeRequest struct {
	Code      string `json:"code" binding:"required"`
	ProblemID string `json:"problem_id"`
}

// ValidateCode handles POST /api/validate
// @Summary      Validate code syntax
// @Description  Validate the syntax of submitted code for a problem
// @Tags         problems
// @Accept       json
// @Produce      json
// @Param        request  body  ValidateCodeRequest  true  "Code to validate"
// @Success      200  {object}  model.APIResponse{data=model.ValidationResult}
// @Failure      400  {object}  model.APIErrorResponse
// @Failure      429  {object}  model.APIErrorResponse
// @Failure      500  {object}  model.APIErrorResponse
// @Router       /api/validate [post]
func (h *ProblemHandler) ValidateCode(c *gin.Context) {
	// Rate limit: 30 req/min for validate endpoint
	ip := c.ClientIP()
	if !h.rateLimiter.AllowWithLimit(ip+":validate", 30, time.Minute) {
		writeErrorResponse(c, http.StatusTooManyRequests, "rate_limit_error",
			"rate limit exceeded, try again later")
		return
	}

	var req ValidateCodeRequest

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
// @Summary      Get hints for a problem
// @Description  Get progressive hints to help solve a specific coding challenge problem
// @Tags         problems
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Problem ID"
// @Success      200  {object}  model.APIResponse{data=map[string][]model.Hint}
// @Failure      400  {object}  model.APIErrorResponse
// @Failure      404  {object}  model.APIErrorResponse
// @Failure      500  {object}  model.APIErrorResponse
// @Router       /api/problems/{id}/hints [get]
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

	hints := h.hintSvc.GetHints("anonymous", problem)
	writeSuccessResponse(c, gin.H{"hints": hints})
}

// HealthCheck handles GET /health
// @Summary      Health check
// @Description  Check the health status of the API service
// @Tags         system
// @Accept       json
// @Produce      json
// @Success      200  {object}  model.APIResponse{data=object}
// @Failure      500  {object}  model.APIErrorResponse
// @Router       /health [get]
func (h *ProblemHandler) HealthCheck(c *gin.Context) {
	writeSuccessResponse(c, gin.H{
		"status":  "ok",
		"version": "1.1.0",
	})
}
