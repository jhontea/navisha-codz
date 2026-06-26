package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"coding-challange/pkg/database"
	"coding-challange/pkg/middleware"
	"coding-challange/pkg/redis"
)

// Service configuration.
const (
	ServiceName = "hint-service"
	ServicePort = "9105"
)

// Hint models.
type Hint struct {
	ID           int    `json:"id" db:"id"`
	ProblemID    int    `json:"problem_id" db:"problem_id"`
	Level        int    `json:"level" db:"level"`
	Title        string `json:"title" db:"title"`
	Content      string `json:"content,omitempty" db:"content"`
	ScorePenalty int    `json:"score_penalty" db:"score_penalty"`
	HintType     string `json:"hint_type,omitempty" db:"hint_type"` // "general", "solution_preview"
}

type HintResponse struct {
	Hint
	Unlocked   bool   `json:"unlocked"`
	UnlockInfo string `json:"unlock_info,omitempty"` // what's needed to unlock
}

type UseHintRequest struct {
	// Empty body - user identification from JWT
}

type UseHintResponse struct {
	HintID         int    `json:"hint_id"`
	Title          string `json:"title"`
	Content        string `json:"content"`
	ScorePenalty   int    `json:"score_penalty"`
	CurrentPenalty int    `json:"current_penalty_pct"`
	IsSolutionPrev bool   `json:"is_solution_preview,omitempty"`
}

type UserHintUsage struct {
	HintID   int       `json:"hint_id"`
	UsedAt   time.Time `json:"used_at"`
}

// HintAnalytics holds analytics data for a hint.
type HintAnalytics struct {
	HintID      int     `json:"hint_id"`
	Title       string  `json:"title"`
	Level       int     `json:"level"`
	UsageCount  int     `json:"usage_count"`
	HelpfulRate float64 `json:"helpful_rate"` // % of users who solved after viewing this hint
	AvgTimeToSolve int  `json:"avg_time_to_solve_ms"`
}

type AdaptiveHintLevel string

const (
	AdaptiveBeginner AdaptiveHintLevel = "beginner"
	AdaptiveIntermediate AdaptiveHintLevel = "intermediate"
	AdaptiveAdvanced AdaptiveHintLevel = "advanced"
)

// Server holds all dependencies.
type Server struct {
	router          *gin.Engine
	db              *database.Pool
	redis           *redis.Client
	jwtConfig       middleware.JWTConfig
	tokenBlacklist  *middleware.TokenBlacklist
	sessionManager  *middleware.SessionManager
}

func main() {
	log.Printf("Starting %s...", ServiceName)

	dbPool, err := database.NewFromEnv()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbPool.Close()

	redisClient, err := redis.NewFromEnv()
	if err != nil {
		log.Printf("Warning: Redis not available: %v", err)
	}

	jwtCfg := middleware.NewJWTConfig()
	srv := &Server{
		router:          gin.New(),
		db:              dbPool,
		redis:           redisClient,
		jwtConfig:       jwtCfg,
		tokenBlacklist:  middleware.NewTokenBlacklist(jwtCfg.EnableTokenBlacklist),
		sessionManager:  middleware.NewSessionManager(3),
	}

	srv.router.Use(gin.Recovery())
	srv.router.Use(middleware.RequestIDMiddleware())
	srv.router.Use(middleware.LoggerMiddleware())
	srv.router.Use(middleware.CORSMiddleware())
	srv.router.Use(middleware.RequestValidationMiddleware(middleware.DefaultRequestValidationConfig()))

	srv.setupRoutes()

	port := getEnv("PORT", ServicePort)
	runServer(srv, port)
}

func (s *Server) setupRoutes() {
	s.router.GET("/health", s.healthCheck)

	api := s.router.Group("/api")
	{
		hints := api.Group("/problems/:id/hints")
		hints.Use(middleware.AuthMiddleware(s.jwtConfig, s.tokenBlacklist, s.sessionManager))
		{
			hints.GET("", s.getHints)
			hints.POST("/:hintId/use", s.useHint)
			hints.GET("/analytics", s.getHintAnalytics)
			hints.GET("/recommended", s.getRecommendedHint)
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
		"service":  ServiceName,
		"status":   status,
		"database": dbStatus,
		"time":     time.Now().UTC(),
	})
}

// ============================================================================
// Adaptive Hint System
// ============================================================================

// determineAdaptiveLevel determines the user's adaptive hint level based on submission history.
func (s *Server) determineAdaptiveLevel(ctx context.Context, userID string, problemID int) AdaptiveHintLevel {
	// Get user's submission stats for this problem
	var totalAttempts, acceptedCount int
	_ = s.db.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(CASE WHEN status = 'accepted' THEN 1 ELSE 0 END), 0)
		FROM submissions WHERE user_id = $1 AND problem_id = $2
	`, userID, problemID).Scan(&totalAttempts, &acceptedCount)

	// Get user's overall skill level (based on ELO rating)
	var rating int
	_ = s.db.QueryRow(ctx, "SELECT COALESCE(rating, 1200) FROM users WHERE id = $1", userID).Scan(&rating)

	// Get user's total solved problems
	var totalSolved int
	_ = s.db.QueryRow(ctx, "SELECT COALESCE(total_solved, 0) FROM users WHERE id = $1", userID).Scan(&totalSolved)

	// Determine adaptive level
	if totalAttempts >= 20 || (totalAttempts >= 10 && acceptedCount == 0) {
		return AdaptiveBeginner
	}
	if rating < 1000 || totalSolved < 5 {
		return AdaptiveBeginner
	}
	if rating >= 1500 && totalSolved >= 20 && totalAttempts < 5 {
		return AdaptiveAdvanced
	}
	return AdaptiveIntermediate
}

// getHintUnlockThresholds returns the number of failed attempts needed to unlock each hint level.
func getHintUnlockThresholds(level int, adaptive AdaptiveHintLevel) int {
	switch level {
	case 1:
		switch adaptive {
		case AdaptiveBeginner:
			return 1 // Available after 1 failed attempt for beginners
		case AdaptiveIntermediate:
			return 2 // Available after 2 failed attempts
		case AdaptiveAdvanced:
			return 3 // Available after 3 failed attempts
		}
	case 2:
		switch adaptive {
		case AdaptiveBeginner:
			return 3
		case AdaptiveIntermediate:
			return 5
		case AdaptiveAdvanced:
			return 7
		}
	case 3:
		switch adaptive {
		case AdaptiveBeginner:
			return 6
		case AdaptiveIntermediate:
			return 10
		case AdaptiveAdvanced:
			return 15
		}
	default:
		return level * 5
	}
	return level * 5
}

// getHintContentByAdaptiveLevel returns the appropriate content based on adaptive level.
func getHintContentByAdaptiveLevel(hint Hint, adaptive AdaptiveHintLevel) string {
	if hint.HintType == "solution_preview" {
		return hint.Content // Solution preview is always the same
	}

	// For regular hints, we could have different content per adaptive level
	// For now, just return the stored content
	return hint.Content
}

// ============================================================================
// Smart Hint Ordering
// ============================================================================

// reorderHints intelligently orders hints based on user progress.
func (s *Server) reorderHints(hints []HintResponse, adaptive AdaptiveHintLevel, failedAttempts int) []HintResponse {
	if len(hints) <= 1 {
		return hints
	}

	// Score each hint based on relevance
	type scored struct {
		hint  HintResponse
		score float64
	}

	var scoredHints []scored
	for _, h := range hints {
		score := 0.0

		// Prioritize hints that are unlockable now
		threshold := getHintUnlockThresholds(h.Level, adaptive)
		if failedAttempts >= threshold {
			score += 100.0
		}

		// Prioritize lower-level hints for beginners
		if adaptive == AdaptiveBeginner {
			score += float64(4 - h.Level) * 10
		}

		// For advanced users, prioritize higher-level hints
		if adaptive == AdaptiveAdvanced && h.Level > 1 {
			score += float64(h.Level) * 5
		}

		// Prefer hints that haven't been used
		if !h.Unlocked {
			score -= 20
		}

		// Solution preview at the end unless desperate
		if h.HintType == "solution_preview" {
			if failedAttempts < 15 {
				score -= 50
			} else {
				score += 200 // Show solution preview after many failures
			}
		}

		scoredHints = append(scoredHints, scored{h, score})
	}

	// Sort by score descending
	for i := 0; i < len(scoredHints); i++ {
		for j := i + 1; j < len(scoredHints); j++ {
			if scoredHints[j].score > scoredHints[i].score {
				scoredHints[i], scoredHints[j] = scoredHints[j], scoredHints[i]
			}
		}
	}

	result := make([]HintResponse, len(scoredHints))
	for i, sh := range scoredHints {
		result[i] = sh.hint
	}

	return result
}

// ============================================================================
// Hint Handlers
// ============================================================================

// getHints handles GET /api/problems/:id/hints - returns hints with unlock status based on attempts.
func (s *Server) getHints(c *gin.Context) {
	ctx := c.Request.Context()
	userIDRaw, exists := c.Get(middleware.ContextKeyUserID)
	if !exists || userIDRaw == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, ok := userIDRaw.(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	problemIDStr := c.Param("id")
	problemID, err := strconv.Atoi(problemIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid problem id"})
		return
	}

	// Try cache
	cacheKey := fmt.Sprintf("hint:%d:%s", problemID, userID)
	if s.redis != nil {
		var cached []HintResponse
		if err := s.redis.Get(ctx, cacheKey, &cached); err == nil {
			c.JSON(http.StatusOK, cached)
			return
		}
	}

	// Get all hints for the problem
	rows, err := s.db.Query(ctx, `
		SELECT id, problem_id, level, title, score_penalty, content, COALESCE(hint_type, 'general')
		FROM problem_hints
		WHERE problem_id = $1
		ORDER BY level ASC
	`, problemID)
	if err != nil {
		log.Printf("Failed to query hints: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	defer rows.Close()

	// Get user's used hints
	usedHints, err := s.getUserUsedHints(ctx, userID, problemID)
	if err != nil {
		log.Printf("Failed to get used hints: %v", err)
	}

	// Count failed attempts
	failedAttempts := s.countFailedAttempts(ctx, userID, problemID)

	// Determine adaptive level
	adaptive := s.determineAdaptiveLevel(ctx, userID, problemID)

	var hints []HintResponse
	for rows.Next() {
		var h Hint
		h.HintType = "general"
		if err := rows.Scan(&h.ID, &h.ProblemID, &h.Level, &h.Title, &h.ScorePenalty, &h.Content, &h.HintType); err != nil {
			continue
		}

		// Determine if hint is unlocked based on failed attempts (not just sequential usage)
		threshold := getHintUnlockThresholds(h.Level, adaptive)
		unlocked := failedAttempts >= threshold

		// Also check if hint was already used (always unlocked then)
		for _, uh := range usedHints {
			if uh.HintID == h.ID {
				unlocked = true
				break
			}
		}

		// Check if previous level hint has been used
		if !unlocked && h.Level > 1 {
			prevUsed := len(usedHints) > 0
			if !prevUsed && h.Level == 2 && failedAttempts >= 2 {
				unlocked = true
			} else if prevUsed {
				unlocked = true
			}
		}

		// Only show content if unlocked
		content := h.Content
		unlockInfo := ""
		if !unlocked {
			content = ""
			unlockInfo = fmt.Sprintf("Solve more or make %d more attempt(s) to unlock", threshold-failedAttempts)
		}

		// Apply adaptive content
		content = getHintContentByAdaptiveLevel(h, adaptive)

		hintResp := HintResponse{
			Hint: Hint{
				ID:           h.ID,
				ProblemID:    h.ProblemID,
				Level:        h.Level,
				Title:        h.Title,
				Content:      content,
				ScorePenalty: h.ScorePenalty,
				HintType:     h.HintType,
			},
			Unlocked:   unlocked,
			UnlockInfo: unlockInfo,
		}

		// Only show content if unlocked (for non-solution-preview hints)
		if !unlocked && h.HintType != "solution_preview" {
			hintResp.Content = ""
		}

		hints = append(hints, hintResp)
	}

	// Apply smart ordering
	hints = s.reorderHints(hints, adaptive, failedAttempts)

	// Cache
	if s.redis != nil {
		_ = s.redis.Set(ctx, cacheKey, hints, redis.TTLHint)
	}

	c.JSON(http.StatusOK, hints)
}

// useHint handles POST /api/problems/:id/hints/:hintId/use.
func (s *Server) useHint(c *gin.Context) {
	ctx := c.Request.Context()
	userIDRaw, exists := c.Get(middleware.ContextKeyUserID)
	if !exists || userIDRaw == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, ok := userIDRaw.(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	problemIDStr := c.Param("id")
	problemID, err := strconv.Atoi(problemIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid problem id"})
		return
	}

	hintIDStr := c.Param("hintId")
	hintID, err := strconv.Atoi(hintIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hint id"})
		return
	}

	// Get hint details
	var hint Hint
	hint.HintType = "general"
	err = s.db.QueryRow(ctx, `
		SELECT id, problem_id, level, title, content, score_penalty, COALESCE(hint_type, 'general')
		FROM problem_hints WHERE id = $1 AND problem_id = $2
	`, hintID, problemID).Scan(&hint.ID, &hint.ProblemID, &hint.Level, &hint.Title, &hint.Content, &hint.ScorePenalty, &hint.HintType)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "hint not found"})
		return
	}

	// Check if already used
	alreadyUsed, err := s.isHintUsed(ctx, userID, hintID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if alreadyUsed {
		c.JSON(http.StatusConflict, gin.H{"error": "hint already used"})
		return
	}

	// Check if hint should be unlocked based on attempts threshold
	failedAttempts := s.countFailedAttempts(ctx, userID, problemID)
	adaptive := s.determineAdaptiveLevel(ctx, userID, problemID)
	threshold := getHintUnlockThresholds(hint.Level, adaptive)

	if failedAttempts < threshold {
		// Check if previous level hint was used (allow sequential unlock)
		prevUsed := false
		if hint.Level > 1 {
			prevUsed, err = s.isPreviousLevelUsed(ctx, userID, problemID, hint.Level)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
				return
			}
		} else {
			prevUsed = true
		}

		if !prevUsed {
			c.JSON(http.StatusForbidden, gin.H{
				"error":             "hint not yet available",
				"failed_attempts":   failedAttempts,
				"attempts_needed":   threshold,
				"message":           fmt.Sprintf("You need %d more failed attempts to unlock this hint", threshold-failedAttempts),
			})
			return
		}
	}

	// Apply penalty
	penalty := hint.ScorePenalty
	if hint.HintType == "solution_preview" {
		penalty = 50 // 50% penalty for solution preview
	}

	// Record hint usage
	_, err = s.db.Exec(ctx, `
		INSERT INTO hints_used (user_id, problem_id, hint_id)
		VALUES ($1, $2, $3)
	`, userID, problemID, hintID)
	if err != nil {
		log.Printf("Failed to record hint usage: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to use hint"})
		return
	}

	// Calculate current total penalty for this problem
	totalPenalty, err := s.calculateTotalPenalty(ctx, userID, problemID)
	if err != nil {
		log.Printf("Failed to calculate penalty: %v", err)
	}

	// Update user_problem_status
	_, err = s.db.Exec(ctx, `
		INSERT INTO user_problem_status (user_id, problem_id, hints_used_count, attempts)
		VALUES ($1, $2, 1, 1)
		ON CONFLICT (user_id, problem_id) DO UPDATE SET
			hints_used_count = user_problem_status.hints_used_count + 1,
			attempts = COALESCE(user_problem_status.attempts, 0) + 1
	`, userID, problemID)
	if err != nil {
		log.Printf("Failed to update user problem status: %v", err)
	}

	// Invalidate caches
	if s.redis != nil {
		_ = s.redis.Delete(ctx, fmt.Sprintf("hint:%d:%s", problemID, userID))
		_ = s.redis.DeletePattern(ctx, fmt.Sprintf("user:stats:%s", userID))
	}

	hintContent := hint.Content
	isSolutionPrev := hint.HintType == "solution_preview"

	c.JSON(http.StatusOK, UseHintResponse{
		HintID:         hint.ID,
		Title:          hint.Title,
		Content:        hintContent,
		ScorePenalty:   penalty,
		CurrentPenalty: totalPenalty,
		IsSolutionPrev: isSolutionPrev,
	})
}

// ============================================================================
// Hint Analytics
// ============================================================================

// getHintAnalytics handles GET /api/problems/:id/hints/analytics
func (s *Server) getHintAnalytics(c *gin.Context) {
	ctx := c.Request.Context()

	problemIDStr := c.Param("id")
	problemID, err := strconv.Atoi(problemIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid problem id"})
		return
	}

	rows, err := s.db.Query(ctx, `
		SELECT
			ph.id,
			ph.title,
			ph.level,
			COALESCE(hu_counts.usage_count, 0) AS usage_count,
			COALESCE(hlep.helpful_count, 0) AS helpful_count,
			COALESCE(hlep.total_users, 0) AS total_users_with_hint,
			COALESCE(hlep.avg_time_to_solve_ms, 0) AS avg_time_to_solve_ms
		FROM problem_hints ph
		LEFT JOIN (
			SELECT hint_id, COUNT(*) AS usage_count
			FROM hints_used
			WHERE problem_id = $1
			GROUP BY hint_id
		) hu_counts ON ph.id = hu_counts.hint_id
		LEFT JOIN (
			SELECT
				hu.hint_id,
				COUNT(DISTINCT CASE WHEN s.status = 'accepted' THEN hu.user_id END) AS helpful_count,
				COUNT(DISTINCT hu.user_id) AS total_users,
				COALESCE(AVG(CASE WHEN s.status = 'accepted' THEN s.execution_time_ms END), 0) AS avg_time_to_solve_ms
			FROM hints_used hu
			LEFT JOIN submissions s ON hu.user_id = s.user_id AND hu.problem_id = s.problem_id
			WHERE hu.problem_id = $1
			GROUP BY hu.hint_id
		) hlep ON ph.id = hlep.hint_id
		WHERE ph.problem_id = $1
		ORDER BY ph.level ASC
	`, problemID)
	if err != nil {
		log.Printf("Failed to query hint analytics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	defer rows.Close()

	type HintAnalyticsEntry struct {
		HintID      int     `json:"hint_id"`
		Title       string  `json:"title"`
		Level       int     `json:"level"`
		UsageCount  int     `json:"usage_count"`
		HelpfulRate float64 `json:"helpful_rate"`
		AvgTimeToSolve int  `json:"avg_time_to_solve_ms"`
	}

	var analytics []HintAnalyticsEntry
	for rows.Next() {
		var entry HintAnalyticsEntry
		var helpfulCount, totalUsersWithHint int
		if err := rows.Scan(&entry.HintID, &entry.Title, &entry.Level, &entry.UsageCount, &helpfulCount, &totalUsersWithHint, &entry.AvgTimeToSolve); err != nil {
			continue
		}
		if totalUsersWithHint > 0 {
			entry.HelpfulRate = math.Round(float64(helpfulCount)/float64(totalUsersWithHint)*100) / 100
		}
		analytics = append(analytics, entry)
	}
	if analytics == nil {
		analytics = []HintAnalyticsEntry{}
	}

	c.JSON(http.StatusOK, gin.H{"analytics": analytics})
}

// getRecommendedHint handles GET /api/problems/:id/hints/recommended
func (s *Server) getRecommendedHint(c *gin.Context) {
	ctx := c.Request.Context()
	userIDRaw, exists := c.Get(middleware.ContextKeyUserID)
	if !exists || userIDRaw == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID, ok := userIDRaw.(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	problemIDStr := c.Param("id")
	problemID, err := strconv.Atoi(problemIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid problem id"})
		return
	}

	// Get all hints
	rows, err := s.db.Query(ctx, `
		SELECT id, problem_id, level, title, score_penalty, COALESCE(hint_type, 'general')
		FROM problem_hints WHERE problem_id = $1
		ORDER BY level ASC
	`, problemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	defer rows.Close()

	var hints []Hint
	for rows.Next() {
		var h Hint
		h.HintType = "general"
		if err := rows.Scan(&h.ID, &h.ProblemID, &h.Level, &h.Title, &h.ScorePenalty, &h.HintType); err == nil {
			hints = append(hints, h)
		}
	}

	if len(hints) == 0 {
		c.JSON(http.StatusOK, gin.H{"recommended": nil, "message": "No hints available"})
		return
	}

	// Get used hints
	usedHints, _ := s.getUserUsedHints(ctx, userID, problemID)
	usedMap := make(map[int]bool)
	for _, uh := range usedHints {
		usedMap[uh.HintID] = true
	}

	// Count failed attempts
	failedAttempts := s.countFailedAttempts(ctx, userID, problemID)
	adaptive := s.determineAdaptiveLevel(ctx, userID, problemID)

	// Find the best hint to recommend
	var recommended *Hint
	bestScore := -1000.0

	for i, h := range hints {
		if usedMap[h.ID] {
			continue
		}

		threshold := getHintUnlockThresholds(h.Level, adaptive)
		if failedAttempts < threshold {
			continue
		}

		score := float64(10 - i) // Prefer earlier hints

		// Check analytics for helpfulness
		var helpfulRate float64
		_ = s.db.QueryRow(ctx, `
			SELECT COALESCE(
				(SELECT COUNT(DISTINCT CASE WHEN s.status = 'accepted' THEN hu.user_id END)::float /
						NULLIF(COUNT(DISTINCT hu.user_id), 0)
				FROM hints_used hu
				LEFT JOIN submissions s ON hu.user_id = s.user_id AND hu.problem_id = s.problem_id
				WHERE hu.hint_id = $1), 0.5)
		`, h.ID).Scan(&helpfulRate)

		score += helpfulRate * 50

		if score > bestScore {
			bestScore = score
			recommended = &hints[i]
		}
	}

	if recommended == nil {
		c.JSON(http.StatusOK, gin.H{"recommended": nil, "message": "No hints available yet. Keep trying!"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"recommended": recommended,
		"failed_attempts": failedAttempts,
		"adaptive_level": adaptive,
	})
}

// ============================================================================
// Helper Functions
// ============================================================================

func (s *Server) getUserUsedHints(ctx context.Context, userID string, problemID int) ([]UserHintUsage, error) {
	rows, err := s.db.Query(ctx, `
		SELECT hint_id, used_at FROM hints_used
		WHERE user_id = $1 AND problem_id = $2
		ORDER BY used_at ASC
	`, userID, problemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hints []UserHintUsage
	for rows.Next() {
		var h UserHintUsage
		if err := rows.Scan(&h.HintID, &h.UsedAt); err != nil {
			continue
		}
		hints = append(hints, h)
	}
	return hints, nil
}

func (s *Server) countFailedAttempts(ctx context.Context, userID string, problemID int) int {
	var count int
	_ = s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM submissions
		WHERE user_id = $1 AND problem_id = $2 AND status != 'accepted'
	`, userID, problemID).Scan(&count)
	return count
}

func (s *Server) isHintUsed(ctx context.Context, userID string, hintID int) (bool, error) {
	var exists bool
	err := s.db.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM hints_used WHERE user_id = $1 AND hint_id = $2)",
		userID, hintID,
	).Scan(&exists)
	return exists, err
}

func (s *Server) isPreviousLevelUsed(ctx context.Context, userID string, problemID, currentLevel int) (bool, error) {
	if currentLevel <= 1 {
		return true, nil
	}

	var exists bool
	err := s.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM hints_used hu
			JOIN problem_hints ph ON hu.hint_id = ph.id
			WHERE hu.user_id = $1 AND hu.problem_id = $2 AND ph.level = $3
		)
	`, userID, problemID, currentLevel-1).Scan(&exists)
	return exists, err
}

func (s *Server) calculateTotalPenalty(ctx context.Context, userID string, problemID int) (int, error) {
	var totalPenalty int
	err := s.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(ph.score_penalty), 0)
		FROM hints_used hu
		JOIN problem_hints ph ON hu.hint_id = ph.id
		WHERE hu.user_id = $1 AND hu.problem_id = $2
	`, userID, problemID).Scan(&totalPenalty)
	return totalPenalty, err
}

// calculateScoreWithPenalty calculates the final score after applying hint penalties.
func calculateScoreWithPenalty(baseScore int, totalPenaltyPercent int) int {
	if totalPenaltyPercent >= 100 {
		return 0
	}
	if totalPenaltyPercent <= 0 {
		return baseScore
	}
	return baseScore * (100 - totalPenaltyPercent) / 100
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

// Ensure unused imports are not flagged
var _ = strings.TrimSpace
