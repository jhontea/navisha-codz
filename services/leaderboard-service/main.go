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
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"coding-challange/pkg/database"
	"coding-challange/pkg/middleware"
	"coding-challange/pkg/redis"
)

// Service configuration.
const (
	ServiceName = "leaderboard-service"
	ServicePort = "9104"

	// ELO defaults
	ELOInitialRating    = 1200
	ELOKFactor          = 32
	ELOKFactorHard      = 48
	ELOKFactorMedium    = 32
	ELOKFactorEasy      = 24
)

// --- Models ---

// LeaderboardEntry represents a row in the leaderboard.
type LeaderboardEntry struct {
	Rank             int     `json:"rank"`
	UserID           string  `json:"user_id"`
	Username         string  `json:"username"`
	TotalScore       int     `json:"total_score"`
	Rating           int     `json:"rating"`
	ProblemsSolved   int     `json:"problems_solved"`
	TotalSubmissions int     `json:"total_submissions"`
	Accuracy         float64 `json:"accuracy"`
}

type LeaderboardResponse struct {
	Period    string             `json:"period"`
	Entries   []LeaderboardEntry `json:"entries"`
	Total     int                `json:"total"`
	UpdatedAt time.Time          `json:"updated_at"`
}

type UserStats struct {
	UserID           string        `json:"user_id"`
	Username         string        `json:"username"`
	Rating           int           `json:"rating"`
	RatingHistory    []int         `json:"rating_history,omitempty"`
	TotalSolved      int           `json:"total_solved"`
	TotalSubmissions int           `json:"total_submissions"`
	WeeklyRank       *int          `json:"weekly_rank,omitempty"`
	MonthlyRank      *int          `json:"monthly_rank,omitempty"`
	AllTimeRank      *int          `json:"all_time_rank,omitempty"`
	WeeklyScore      int           `json:"weekly_score"`
	MonthlyScore     int           `json:"monthly_score"`
	AllTimeScore     int           `json:"all_time_score"`
	StreakDays       int           `json:"streak_days"`
	MaxStreakDays    int           `json:"max_streak_days"`
	LastSolveDate    *string       `json:"last_solve_date,omitempty"`
	Badges           []Badge       `json:"badges,omitempty"`
	Achievements     []Achievement `json:"achievements,omitempty"`
}

// Badge represents a user badge.
type Badge struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Emoji       string `json:"emoji"`
	Description string `json:"description"`
	UnlockedAt  string `json:"unlocked_at,omitempty"`
}

// Achievement represents a user achievement.
type Achievement struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	UnlockedAt  string `json:"unlocked_at,omitempty"`
	Progress    int    `json:"progress,omitempty"`
	MaxProgress int    `json:"max_progress,omitempty"`
}

// RatingChangeResult holds ELO calculation results.
type RatingChangeResult struct {
	NewRating     int     `json:"new_rating"`
	RatingChange  int     `json:"rating_change"`
	ExpectedScore float64 `json:"expected_score"`
}

// StreakInfo holds daily streak data.
type StreakInfo struct {
	CurrentStreak int    `json:"current_streak"`
	MaxStreak     int    `json:"max_streak"`
	LastSolveDate string `json:"last_solve_date"`
}

// --- Server ---

type Server struct {
	router    *gin.Engine
	db        *database.Pool
	redis     *redis.Client
	jwtConfig middleware.JWTConfig
	mu        sync.Mutex
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

	srv := &Server{
		router:    gin.New(),
		db:        dbPool,
		redis:     redisClient,
		jwtConfig: middleware.NewJWTConfig(),
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
		// Leaderboard endpoints
		api.GET("/leaderboard/weekly", s.getWeeklyLeaderboard)
		api.GET("/leaderboard/monthly", s.getMonthlyLeaderboard)
		api.GET("/leaderboard/all-time", s.getAllTimeLeaderboard)

		// Enhanced endpoints
		api.GET("/leaderboard/search", s.searchLeaderboard)
		api.GET("/leaderboard/compare", s.compareUsers)
		api.GET("/leaderboard/rating-history/:userId", s.getRatingHistory)
		api.GET("/leaderboard/rewards/weekly", s.getWeeklyRewards)
		api.GET("/leaderboard/rewards/monthly", s.getMonthlyRewards)

		// User stats (protected)
		stats := api.Group("/users")
		{
			stats.GET("/:id/stats", s.getUserStats)
			stats.GET("/:id/badges", s.getUserBadges)
			stats.GET("/:id/achievements", s.getUserAchievements)
			stats.POST("/:id/daily-login", s.recordDailyLogin)
		}

		// Internal endpoints
		internal := api.Group("/internal")
		{
			internal.POST("/update-score", s.internalUpdateScore)
			internal.POST("/recalculate", s.internalRecalculate)
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
// ELO Rating System
// ============================================================================

// calculateExpectedScore computes the expected score for player with ratingA vs ratingB.
func calculateExpectedScore(ratingA, ratingB int) float64 {
	return 1.0 / (1.0 + math.Pow(10, float64(ratingB-ratingA)/400.0))
}

// calculateRatingChange computes rating change for a win/loss.
func calculateRatingChange(currentRating, opponentRating int, won bool, difficulty string) RatingChangeResult {
	expected := calculateExpectedScore(currentRating, opponentRating)

	kFactor := getKFactor(difficulty)

	var actual float64
	if won {
		actual = 1.0
	} else {
		actual = 0.0
	}

	change := int(math.Round(kFactor * (actual - expected)))
	newRating := currentRating + change

	if newRating < 100 {
		newRating = 100
	}

	return RatingChangeResult{
		NewRating:     newRating,
		RatingChange:  change,
		ExpectedScore: expected,
	}
}

// getKFactor returns the K-factor based on difficulty.
func getKFactor(difficulty string) float64 {
	switch strings.ToLower(difficulty) {
	case "hard":
		return ELOKFactorHard
	case "medium":
		return ELOKFactorMedium
	case "easy":
		return ELOKFactorEasy
	default:
		return ELOKFactor
	}
}

// updateELORating updates the user's ELO rating based on submission result.
func (s *Server) updateELORating(ctx context.Context, userID string, problemDifficulty string, passed bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var currentRating int
	err := s.db.QueryRow(ctx, "SELECT rating FROM users WHERE id = $1", userID).Scan(&currentRating)
	if err != nil {
		currentRating = ELOInitialRating
		_, err = s.db.Exec(ctx, "UPDATE users SET rating = $1 WHERE id = $2", currentRating, userID)
		if err != nil {
			return fmt.Errorf("failed to initialize rating: %w", err)
		}
	}

	var opponentRating int
	switch strings.ToLower(problemDifficulty) {
	case "easy":
		opponentRating = 1000
	case "medium":
		opponentRating = 1400
	case "hard":
		opponentRating = 1800
	default:
		opponentRating = 1200
	}

	result := calculateRatingChange(currentRating, opponentRating, passed, problemDifficulty)

	_, err = s.db.Exec(ctx, "UPDATE users SET rating = $1 WHERE id = $2", result.NewRating, userID)
	if err != nil {
		return fmt.Errorf("failed to update rating: %w", err)
	}

	_, err = s.db.Exec(ctx, `
		INSERT INTO rating_history (user_id, rating, change_amount, reason, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`, userID, result.NewRating, result.RatingChange, getRatingReason(passed, problemDifficulty))
	if err != nil {
		log.Printf("Failed to record rating history: %v", err)
	}

	return nil
}

func getRatingReason(passed bool, difficulty string) string {
	if passed {
		return fmt.Sprintf("Solved %s problem", difficulty)
	}
	return fmt.Sprintf("Failed %s problem", difficulty)
}

// ============================================================================
// Streak Tracking
// ============================================================================

// recordSolveStreak updates the user's consecutive solve streak.
func (s *Server) recordSolveStreak(ctx context.Context, userID string) error {
	today := time.Now().UTC().Format("2006-01-02")

	var lastSolveDate *string
	var currentStreak, maxStreak int

	err := s.db.QueryRow(ctx, `
		SELECT current_streak, max_streak, last_solve_date
		FROM user_streaks WHERE user_id = $1
	`, userID).Scan(&currentStreak, &maxStreak, &lastSolveDate)

	if err != nil {
		_, err = s.db.Exec(ctx, `
			INSERT INTO user_streaks (user_id, current_streak, max_streak, last_solve_date)
			VALUES ($1, 1, 1, $2)
		`, userID, today)
		return err
	}

	if lastSolveDate == nil || *lastSolveDate == "" {
		_, err = s.db.Exec(ctx, `
			UPDATE user_streaks SET current_streak = 1, max_streak = GREATEST(max_streak, 1), last_solve_date = $1
			WHERE user_id = $2
		`, today, userID)
		return err
	}

	lastDate, err := time.Parse("2006-01-02", *lastSolveDate)
	if err != nil {
		return err
	}

	todayTime, _ := time.Parse("2006-01-02", today)
	diff := todayTime.Sub(lastDate).Hours() / 24

	if diff == 0 {
		return nil
	} else if diff == 1 {
		currentStreak++
		if currentStreak > maxStreak {
			maxStreak = currentStreak
		}
	} else {
		currentStreak = 1
	}

	_, err = s.db.Exec(ctx, `
		UPDATE user_streaks SET current_streak = $1, max_streak = $2, last_solve_date = $3
		WHERE user_id = $4
	`, currentStreak, maxStreak, today, userID)

	return err
}

// recordDailyLogin records a daily login for streak tracking.
func (s *Server) recordDailyLogin(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user id required"})
		return
	}

	today := time.Now().UTC().Format("2006-01-02")
	ctx := c.Request.Context()

	var lastLogin *string
	err := s.db.QueryRow(ctx, "SELECT last_login_date FROM user_streaks WHERE user_id = $1", userID).Scan(&lastLogin)
	if err != nil {
		_, err = s.db.Exec(ctx, `
			INSERT INTO user_streaks (user_id, current_streak, max_streak, last_login_date, last_solve_date)
			VALUES ($1, 1, 1, $2, NULL)
		`, userID, today)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"streak": 1, "message": "Daily login recorded"})
		return
	}

	if lastLogin != nil && *lastLogin == today {
		c.JSON(http.StatusOK, gin.H{"streak": 0, "message": "Already logged in today"})
		return
	}

	var currentStreak int
	_ = s.db.QueryRow(ctx, "SELECT current_streak FROM user_streaks WHERE user_id = $1", userID).Scan(&currentStreak)

	if lastLogin != nil {
		lastDate, err := time.Parse("2006-01-02", *lastLogin)
		if err == nil {
			todayTime, _ := time.Parse("2006-01-02", today)
			diff := todayTime.Sub(lastDate).Hours() / 24
			if diff == 1 {
				currentStreak++
			} else {
				currentStreak = 1
			}
		}
	} else {
		currentStreak = 1
	}

	_, err = s.db.Exec(ctx, `
		UPDATE user_streaks SET current_streak = $1, max_streak = GREATEST(max_streak, $1), last_login_date = $2
		WHERE user_id = $3
	`, currentStreak, today, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"streak": currentStreak, "message": "Daily login recorded"})
}

// ============================================================================
// Badge System
// ============================================================================

var allBadges = []Badge{
	{ID: "gold", Name: "Gold", Emoji: "🥇", Description: "Top 1% of all users"},
	{ID: "silver", Name: "Silver", Emoji: "🥈", Description: "Top 5% of all users"},
	{ID: "bronze", Name: "Bronze", Emoji: "🥉", Description: "Top 10% of all users"},
	{ID: "streak_master", Name: "Streak Master", Emoji: "🔥", Description: "Maintained a 7+ day streak"},
	{ID: "grinder", Name: "Grinder", Emoji: "💪", Description: "Solved 50+ problems"},
	{ID: "genius", Name: "Genius", Emoji: "🧠", Description: "Solved 5 hard problems"},
}

var allAchievements = []Achievement{
	{ID: "first_solve", Name: "First Solve", Description: "Solved your first problem", Icon: "🎯"},
	{ID: "speed_demon", Name: "Speed Demon", Description: "Solved a problem in under 2 minutes", Icon: "⚡"},
	{ID: "marathon", Name: "Marathon", Description: "Solved 10 problems in a single day", Icon: "🏃"},
	{ID: "perfectionist", Name: "Perfectionist", Description: "Passed all test cases on first try", Icon: "💎"},
}

// checkAndAwardBadges checks all badge conditions and awards any new badges.
func (s *Server) checkAndAwardBadges(ctx context.Context, userID string) ([]Badge, error) {
	stats, err := s.getUserStatsData(ctx, userID)
	if err != nil {
		return nil, err
	}

	var newBadges []Badge

	existingBadges, err := s.getUserBadgeIDs(ctx, userID)
	if err != nil {
		return nil, err
	}

	existingMap := make(map[string]bool)
	for _, b := range existingBadges {
		existingMap[b] = true
	}

	var totalUsers int
	_ = s.db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&totalUsers)
	if totalUsers == 0 {
		totalUsers = 1
	}

	var allTimeRank int
	_ = s.db.QueryRow(ctx, "SELECT all_time_rank FROM leaderboard WHERE user_id = $1", userID).Scan(&allTimeRank)

	percentile := 0.0
	if allTimeRank > 0 && totalUsers > 0 {
		percentile = float64(allTimeRank) / float64(totalUsers) * 100
	}

	for _, badge := range allBadges {
		if existingMap[badge.ID] {
			continue
		}

		shouldAward := false
		switch badge.ID {
		case "gold":
			if allTimeRank > 0 && percentile <= 1.0 {
				shouldAward = true
			}
		case "silver":
			if allTimeRank > 0 && percentile <= 5.0 {
				shouldAward = true
			}
		case "bronze":
			if allTimeRank > 0 && percentile <= 10.0 {
				shouldAward = true
			}
		case "streak_master":
			if stats.MaxStreakDays >= 7 {
				shouldAward = true
			}
		case "grinder":
			if stats.TotalSolved >= 50 {
				shouldAward = true
			}
		case "genius":
			var hardSolved int
			_ = s.db.QueryRow(ctx, `
				SELECT COUNT(*) FROM submissions s
				JOIN problems p ON s.problem_id = p.id
				WHERE s.user_id = $1 AND s.status = 'accepted' AND p.difficulty = 'hard'
			`, userID).Scan(&hardSolved)
			if hardSolved >= 5 {
				shouldAward = true
			}
		}

		if shouldAward {
			badge.UnlockedAt = time.Now().UTC().Format(time.RFC3339)
			_, err := s.db.Exec(ctx, `
				INSERT INTO user_badges (user_id, badge_id, unlocked_at)
				VALUES ($1, $2, NOW())
				ON CONFLICT DO NOTHING
			`, userID, badge.ID)
			if err == nil {
				newBadges = append(newBadges, badge)
			}
		}
	}

	return newBadges, nil
}

// checkAndAwardAchievements checks all achievement conditions and awards any new achievements.
func (s *Server) checkAndAwardAchievements(ctx context.Context, userID string, submissionStatus string, executionTime int) ([]Achievement, error) {
	var newAchievements []Achievement

	existingAchievements, err := s.getUserAchievementIDs(ctx, userID)
	if err != nil {
		return nil, err
	}

	existingMap := make(map[string]bool)
	for _, a := range existingAchievements {
		existingMap[a] = true
	}

	for _, achievement := range allAchievements {
		if existingMap[achievement.ID] {
			continue
		}

		shouldAward := false

		switch achievement.ID {
		case "first_solve":
			if submissionStatus == "accepted" {
				var count int
				_ = s.db.QueryRow(ctx, "SELECT COUNT(*) FROM submissions WHERE user_id = $1 AND status = 'accepted'", userID).Scan(&count)
				if count == 1 {
					shouldAward = true
				}
			}
		case "speed_demon":
			if submissionStatus == "accepted" && executionTime > 0 && executionTime < 120000 {
				shouldAward = true
			}
		case "marathon":
			if submissionStatus == "accepted" {
				today := time.Now().UTC().Format("2006-01-02")
				var count int
				_ = s.db.QueryRow(ctx, `
					SELECT COUNT(*) FROM submissions
					WHERE user_id = $1 AND status = 'accepted' AND DATE(created_at) = $2
				`, userID, today).Scan(&count)
				if count >= 10 {
					shouldAward = true
				}
			}
		case "perfectionist":
			if submissionStatus == "accepted" {
				var problemID string
				_ = s.db.QueryRow(ctx, `
					SELECT problem_id FROM submissions
					WHERE user_id = $1 AND status = 'accepted'
					ORDER BY created_at DESC LIMIT 1
				`, userID).Scan(&problemID)

				var prevAttempts int
				_ = s.db.QueryRow(ctx, `
					SELECT COUNT(*) FROM submissions
					WHERE user_id = $1 AND problem_id = $2 AND status != 'accepted'
				`, userID, problemID).Scan(&prevAttempts)
				if prevAttempts == 0 {
					shouldAward = true
				}
			}
		}

		if shouldAward {
			achievement.UnlockedAt = time.Now().UTC().Format(time.RFC3339)
			_, err := s.db.Exec(ctx, `
				INSERT INTO user_achievements (user_id, achievement_id, unlocked_at)
				VALUES ($1, $2, NOW())
				ON CONFLICT DO NOTHING
			`, userID, achievement.ID)
			if err == nil {
				newAchievements = append(newAchievements, achievement)
			}
		}
	}

	return newAchievements, nil
}

func (s *Server) getUserBadgeIDs(ctx context.Context, userID string) ([]string, error) {
	rows, err := s.db.Query(ctx, "SELECT badge_id FROM user_badges WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func (s *Server) getUserAchievementIDs(ctx context.Context, userID string) ([]string, error) {
	rows, err := s.db.Query(ctx, "SELECT achievement_id FROM user_achievements WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

// getUserStatsData retrieves full stats including streak info.
func (s *Server) getUserStatsData(ctx context.Context, userID string) (*UserStats, error) {
	var stats UserStats
	var lastSolveDate *string
	var currentStreak, maxStreak int

	err := s.db.QueryRow(ctx, `
		SELECT u.id, u.username, u.rating, u.total_solved, u.total_submissions,
			COALESCE(lb.weekly_rank, 0), COALESCE(lb.monthly_rank, 0), COALESCE(lb.all_time_rank, 0),
			COALESCE(lb.weekly_score, 0), COALESCE(lb.monthly_score, 0), COALESCE(lb.all_time_score, 0)
		FROM users u
		LEFT JOIN leaderboard lb ON u.id = lb.user_id
		WHERE u.id = $1
	`, userID).Scan(&stats.UserID, &stats.Username, &stats.Rating,
		&stats.TotalSolved, &stats.TotalSubmissions,
		&stats.WeeklyRank, &stats.MonthlyRank, &stats.AllTimeRank,
		&stats.WeeklyScore, &stats.MonthlyScore, &stats.AllTimeScore)

	if err != nil {
		return nil, err
	}

	// Get streak info
	_ = s.db.QueryRow(ctx, `
		SELECT COALESCE(current_streak, 0), COALESCE(max_streak, 0), last_solve_date
		FROM user_streaks WHERE user_id = $1
	`, userID).Scan(&currentStreak, &maxStreak, &lastSolveDate)

	stats.StreakDays = currentStreak
	stats.MaxStreakDays = maxStreak
	if lastSolveDate != nil {
		stats.LastSolveDate = lastSolveDate
	}

	// Get badges
	badgeRows, err := s.db.Query(ctx, `
		SELECT ub.badge_id, ub.unlocked_at
		FROM user_badges ub
		WHERE ub.user_id = $1
	`, userID)
	if err == nil {
		defer badgeRows.Close()
		for badgeRows.Next() {
			var badgeID, unlockedAt string
			if err := badgeRows.Scan(&badgeID, &unlockedAt); err == nil {
				for _, b := range allBadges {
					if b.ID == badgeID {
						b.UnlockedAt = unlockedAt
						stats.Badges = append(stats.Badges, b)
						break
					}
				}
			}
		}
	}

	// Get achievements
	achRows, err := s.db.Query(ctx, `
		SELECT ua.achievement_id, ua.unlocked_at
		FROM user_achievements ua
		WHERE ua.user_id = $1
	`, userID)
	if err == nil {
		defer achRows.Close()
		for achRows.Next() {
			var achID, unlockedAt string
			if err := achRows.Scan(&achID, &unlockedAt); err == nil {
				for _, a := range allAchievements {
					if a.ID == achID {
						a.UnlockedAt = unlockedAt
						stats.Achievements = append(stats.Achievements, a)
						break
					}
				}
			}
		}
	}

	// Get rating history (last 30 entries)
	histRows, err := s.db.Query(ctx, `
		SELECT rating FROM rating_history
		WHERE user_id = $1
		ORDER BY created_at DESC LIMIT 30
	`, userID)
	if err == nil {
		defer histRows.Close()
		for histRows.Next() {
			var rating int
			if err := histRows.Scan(&rating); err == nil {
				stats.RatingHistory = append(stats.RatingHistory, rating)
			}
		}
		for i, j := 0, len(stats.RatingHistory)-1; i < j; i, j = i+1, j-1 {
			stats.RatingHistory[i], stats.RatingHistory[j] = stats.RatingHistory[j], stats.RatingHistory[i]
		}
	}

	return &stats, nil
}

// ============================================================================
// Leaderboard Handlers
// ============================================================================

func (s *Server) getWeeklyLeaderboard(c *gin.Context) {
	ctx := c.Request.Context()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "100"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 500 {
		pageSize = 100
	}

	cacheKey := "leaderboard:weekly"
	if s.redis != nil {
		var cached LeaderboardResponse
		if err := s.redis.Get(ctx, cacheKey, &cached); err == nil {
			c.JSON(http.StatusOK, cached)
			return
		}
	}

	rows, err := s.db.Query(ctx, `
		SELECT
			lb.weekly_rank AS rank,
			u.id AS user_id,
			u.username,
			lb.weekly_score AS total_score,
			u.rating,
			u.total_solved,
			u.total_submissions,
			CASE WHEN u.total_submissions > 0 THEN ROUND(u.total_solved::float / u.total_submissions * 100, 1) ELSE 0 END AS accuracy,
			COUNT(*) OVER() AS total
		FROM leaderboard lb
		JOIN users u ON lb.user_id = u.id
		WHERE lb.weekly_rank IS NOT NULL AND lb.weekly_score > 0
		ORDER BY lb.weekly_rank ASC
		LIMIT $1 OFFSET $2
	`, pageSize, (page-1)*pageSize)
	if err != nil {
		log.Printf("Failed to query weekly leaderboard: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	var total int
	for rows.Next() {
		var e LeaderboardEntry
		if err := rows.Scan(&e.Rank, &e.UserID, &e.Username, &e.TotalScore, &e.Rating,
			&e.ProblemsSolved, &e.TotalSubmissions, &e.Accuracy, &total); err != nil {
			continue
		}
		entries = append(entries, e)
	}

	if total == 0 {
		_ = s.db.QueryRow(ctx,
			"SELECT COUNT(*) FROM leaderboard WHERE weekly_rank IS NOT NULL",
		).Scan(&total)
	}

	response := LeaderboardResponse{
		Period:    "weekly",
		Entries:   entries,
		Total:     total,
		UpdatedAt: time.Now().UTC(),
	}

	if s.redis != nil {
		_ = s.redis.Set(ctx, cacheKey, response, redis.TTLLeaderboard)
	}

	c.JSON(http.StatusOK, response)
}

func (s *Server) getMonthlyLeaderboard(c *gin.Context) {
	ctx := c.Request.Context()

	cacheKey := "leaderboard:monthly"
	if s.redis != nil {
		var cached LeaderboardResponse
		if err := s.redis.Get(ctx, cacheKey, &cached); err == nil {
			c.JSON(http.StatusOK, cached)
			return
		}
	}

	rows, err := s.db.Query(ctx, `
		SELECT
			lb.monthly_rank AS rank,
			u.id AS user_id,
			u.username,
			lb.monthly_score AS total_score,
			u.rating,
			u.total_solved,
			u.total_submissions,
			CASE WHEN u.total_submissions > 0 THEN ROUND(u.total_solved::float / u.total_submissions * 100, 1) ELSE 0 END AS accuracy,
			COUNT(*) OVER() AS total
		FROM leaderboard lb
		JOIN users u ON lb.user_id = u.id
		WHERE lb.monthly_rank IS NOT NULL AND lb.monthly_score > 0
		ORDER BY lb.monthly_rank ASC
		LIMIT 100
	`)
	if err != nil {
		log.Printf("Failed to query monthly leaderboard: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	var total int
	for rows.Next() {
		var e LeaderboardEntry
		if err := rows.Scan(&e.Rank, &e.UserID, &e.Username, &e.TotalScore, &e.Rating,
			&e.ProblemsSolved, &e.TotalSubmissions, &e.Accuracy, &total); err != nil {
			continue
		}
		entries = append(entries, e)
	}

	response := LeaderboardResponse{
		Period:    "monthly",
		Entries:   entries,
		Total:     total,
		UpdatedAt: time.Now().UTC(),
	}

	if s.redis != nil {
		_ = s.redis.Set(ctx, cacheKey, response, redis.TTLLeaderboard*5)
	}

	c.JSON(http.StatusOK, response)
}

func (s *Server) getAllTimeLeaderboard(c *gin.Context) {
	ctx := c.Request.Context()

	cacheKey := "leaderboard:alltime"
	if s.redis != nil {
		var cached LeaderboardResponse
		if err := s.redis.Get(ctx, cacheKey, &cached); err == nil {
			c.JSON(http.StatusOK, cached)
			return
		}
	}

	rows, err := s.db.Query(ctx, `
		SELECT
			lb.all_time_rank AS rank,
			u.id AS user_id,
			u.username,
			lb.all_time_score AS total_score,
			u.rating,
			u.total_solved,
			u.total_submissions,
			CASE WHEN u.total_submissions > 0 THEN ROUND(u.total_solved::float / u.total_submissions * 100, 1) ELSE 0 END AS accuracy,
			COUNT(*) OVER() AS total
		FROM leaderboard lb
		JOIN users u ON lb.user_id = u.id
		WHERE lb.all_time_rank IS NOT NULL AND lb.all_time_score > 0
		ORDER BY lb.all_time_rank ASC
		LIMIT 100
	`)
	if err != nil {
		log.Printf("Failed to query all-time leaderboard: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	var total int
	for rows.Next() {
		var e LeaderboardEntry
		if err := rows.Scan(&e.Rank, &e.UserID, &e.Username, &e.TotalScore, &e.Rating,
			&e.ProblemsSolved, &e.TotalSubmissions, &e.Accuracy, &total); err != nil {
			continue
		}
		entries = append(entries, e)
	}

	response := LeaderboardResponse{
		Period:    "all-time",
		Entries:   entries,
		Total:     total,
		UpdatedAt: time.Now().UTC(),
	}

	if s.redis != nil {
		_ = s.redis.Set(ctx, cacheKey, response, redis.TTLLeaderboard*10)
	}

	c.JSON(http.StatusOK, response)
}

// searchLeaderboard handles GET /api/leaderboard/search?q=username
func (s *Server) searchLeaderboard(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search query required"})
		return
	}

	ctx := c.Request.Context()

	rows, err := s.db.Query(ctx, `
		SELECT
			lb.all_time_rank AS rank,
			u.id AS user_id,
			u.username,
			lb.all_time_score AS total_score,
			u.rating,
			u.total_solved,
			u.total_submissions,
			CASE WHEN u.total_submissions > 0 THEN ROUND(u.total_solved::float / u.total_submissions * 100, 1) ELSE 0 END AS accuracy
		FROM leaderboard lb
		JOIN users u ON lb.user_id = u.id
		WHERE u.username ILIKE '%' || $1 || '%'
		ORDER BY lb.all_time_rank ASC
		LIMIT 20
	`, query)
	if err != nil {
		log.Printf("Failed to search leaderboard: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	for rows.Next() {
		var e LeaderboardEntry
		if err := rows.Scan(&e.Rank, &e.UserID, &e.Username, &e.TotalScore, &e.Rating,
			&e.ProblemsSolved, &e.TotalSubmissions, &e.Accuracy); err != nil {
			continue
		}
		entries = append(entries, e)
	}

	if entries == nil {
		entries = []LeaderboardEntry{}
	}

	c.JSON(http.StatusOK, gin.H{"entries": entries, "total": len(entries)})
}

// compareUsers handles GET /api/leaderboard/compare?user1=id&user2=id
func (s *Server) compareUsers(c *gin.Context) {
	user1ID := c.Query("user1")
	user2ID := c.Query("user2")
	if user1ID == "" || user2ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "both user1 and user2 required"})
		return
	}

	ctx := c.Request.Context()

	stats1, err := s.getUserStatsData(ctx, user1ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user1 not found"})
		return
	}

	stats2, err := s.getUserStatsData(ctx, user2ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user2 not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user1": stats1,
		"user2": stats2,
	})
}

// getRatingHistory handles GET /api/leaderboard/rating-history/:userId
func (s *Server) getRatingHistory(c *gin.Context) {
	userID := c.Param("userId")
	ctx := c.Request.Context()

	rows, err := s.db.Query(ctx, `
		SELECT rating, change_amount, reason, created_at
		FROM rating_history
		WHERE user_id = $1
		ORDER BY created_at ASC
		LIMIT 100
	`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	defer rows.Close()

	type RatingPoint struct {
		Rating       int       `json:"rating"`
		ChangeAmount int       `json:"change_amount"`
		Reason       string    `json:"reason"`
		CreatedAt    time.Time `json:"created_at"`
	}

	var history []RatingPoint
	for rows.Next() {
		var p RatingPoint
		if err := rows.Scan(&p.Rating, &p.ChangeAmount, &p.Reason, &p.CreatedAt); err == nil {
			history = append(history, p)
		}
	}

	if history == nil {
		history = []RatingPoint{}
	}

	c.JSON(http.StatusOK, gin.H{"history": history})
}

func (s *Server) getUserStats(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.Param("id")

	cacheKey := fmt.Sprintf("user:stats:%s", userID)
	if s.redis != nil {
		var cached UserStats
		if err := s.redis.Get(ctx, cacheKey, &cached); err == nil {
			c.JSON(http.StatusOK, cached)
			return
		}
	}

	stats, err := s.getUserStatsData(ctx, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if s.redis != nil {
		_ = s.redis.Set(ctx, cacheKey, stats, redis.TTLUserStats)
	}

	c.JSON(http.StatusOK, stats)
}

func (s *Server) getUserBadges(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.Param("id")

	badgeRows, err := s.db.Query(ctx, `
		SELECT ub.badge_id, ub.unlocked_at
		FROM user_badges ub
		WHERE ub.user_id = $1
		ORDER BY ub.unlocked_at DESC
	`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	defer badgeRows.Close()

	var badges []Badge
	for badgeRows.Next() {
		var badgeID, unlockedAt string
		if err := badgeRows.Scan(&badgeID, &unlockedAt); err == nil {
			for _, b := range allBadges {
				if b.ID == badgeID {
					b.UnlockedAt = unlockedAt
					badges = append(badges, b)
					break
				}
			}
		}
	}
	if badges == nil {
		badges = []Badge{}
	}

	c.JSON(http.StatusOK, gin.H{"badges": badges})
}

func (s *Server) getUserAchievements(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.Param("id")

	achRows, err := s.db.Query(ctx, `
		SELECT ua.achievement_id, ua.unlocked_at
		FROM user_achievements ua
		WHERE ua.user_id = $1
		ORDER BY ua.unlocked_at DESC
	`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	defer achRows.Close()

	var achievements []Achievement
	for achRows.Next() {
		var achID, unlockedAt string
		if err := achRows.Scan(&achID, &unlockedAt); err == nil {
			for _, a := range allAchievements {
				if a.ID == achID {
					a.UnlockedAt = unlockedAt
					achievements = append(achievements, a)
					break
				}
			}
		}
	}

	existingMap := make(map[string]bool)
	for _, a := range achievements {
		existingMap[a.ID] = true
	}

	for _, a := range allAchievements {
		if !existingMap[a.ID] {
			a.Progress = s.getAchievementProgress(ctx, userID, a.ID)
			a.MaxProgress = s.getAchievementMaxProgress(a.ID)
			achievements = append(achievements, a)
		}
	}

	if achievements == nil {
		achievements = []Achievement{}
	}

	c.JSON(http.StatusOK, gin.H{"achievements": achievements})
}

func (s *Server) getAchievementProgress(ctx context.Context, userID, achievementID string) int {
	switch achievementID {
	case "first_solve":
		var count int
		_ = s.db.QueryRow(ctx, "SELECT COUNT(*) FROM submissions WHERE user_id = $1 AND status = 'accepted'", userID).Scan(&count)
		if count > 0 {
			return 1
		}
		return 0
	case "speed_demon":
		var count int
		_ = s.db.QueryRow(ctx, "SELECT COUNT(*) FROM submissions WHERE user_id = $1 AND status = 'accepted' AND execution_time_ms < 120000", userID).Scan(&count)
		return count
	case "marathon":
		today := time.Now().UTC().Format("2006-01-02")
		var count int
		_ = s.db.QueryRow(ctx, "SELECT COUNT(*) FROM submissions WHERE user_id = $1 AND status = 'accepted' AND DATE(created_at) = $2", userID, today).Scan(&count)
		return count
	case "perfectionist":
		var count int
		_ = s.db.QueryRow(ctx, `SELECT COUNT(*) FROM submissions s1 WHERE user_id = $1 AND status = 'accepted' AND NOT EXISTS (SELECT 1 FROM submissions s2 WHERE s2.user_id = s1.user_id AND s2.problem_id = s1.problem_id AND s2.status != 'accepted' AND s2.created_at < s1.created_at)`, userID).Scan(&count)
		return count
	}
	return 0
}

func (s *Server) getAchievementMaxProgress(achievementID string) int {
	switch achievementID {
	case "first_solve":
		return 1
	case "speed_demon":
		return 1
	case "marathon":
		return 10
	case "perfectionist":
		return 1
	}
	return 1
}

// getWeeklyRewards handles GET /api/leaderboard/rewards/weekly
func (s *Server) getWeeklyRewards(c *gin.Context) {
	ctx := c.Request.Context()

	type RewardDef struct {
		Rank   int    `json:"rank"`
		Reward string `json:"reward"`
		Points int    `json:"points"`
		Emoji  string `json:"emoji"`
	}

	rewardDefs := []RewardDef{
		{Rank: 1, Reward: "Gold Trophy + 500 bonus points", Points: 500, Emoji: "🏆"},
		{Rank: 2, Reward: "Silver Trophy + 300 bonus points", Points: 300, Emoji: "🥈"},
		{Rank: 3, Reward: "Bronze Trophy + 200 bonus points", Points: 200, Emoji: "🥉"},
		{Rank: 4, Reward: "100 bonus points", Points: 100, Emoji: "🎖️"},
		{Rank: 5, Reward: "80 bonus points", Points: 80, Emoji: "🎖️"},
		{Rank: 10, Reward: "50 bonus points", Points: 50, Emoji: "⭐"},
		{Rank: 25, Reward: "30 bonus points", Points: 30, Emoji: "⭐"},
		{Rank: 50, Reward: "10 bonus points", Points: 10, Emoji: "✨"},
	}

	rows, err := s.db.Query(ctx, `
		SELECT u.id, u.username, lb.weekly_score, lb.weekly_rank
		FROM leaderboard lb
		JOIN users u ON lb.user_id = u.id
		WHERE lb.weekly_rank IS NOT NULL AND lb.weekly_rank <= 50
		ORDER BY lb.weekly_rank ASC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	defer rows.Close()

	type RewardEntry struct {
		Rank     int    `json:"rank"`
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		Score    int    `json:"score"`
		Reward   string `json:"reward"`
		Points   int    `json:"points"`
		Emoji    string `json:"emoji"`
	}

	var entries []RewardEntry
	for rows.Next() {
		var userID, username string
		var score, rank int
		if err := rows.Scan(&userID, &username, &score, &rank); err != nil {
			continue
		}
		entry := RewardEntry{
			Rank:     rank,
			UserID:   userID,
			Username: username,
			Score:    score,
		}
		for _, r := range rewardDefs {
			if r.Rank == rank {
				entry.Reward = r.Reward
				entry.Points = r.Points
				entry.Emoji = r.Emoji
				break
			}
		}
		if entry.Reward == "" {
			if rank <= 10 {
				entry.Reward = "50 bonus points"
				entry.Points = 50
				entry.Emoji = "⭐"
			} else if rank <= 25 {
				entry.Reward = "30 bonus points"
				entry.Points = 30
				entry.Emoji = "⭐"
			} else {
				entry.Reward = "10 bonus points"
				entry.Points = 10
				entry.Emoji = "✨"
			}
		}
		entries = append(entries, entry)
	}
	if entries == nil {
		entries = []RewardEntry{}
	}

	c.JSON(http.StatusOK, gin.H{
		"rewards": rewardDefs,
		"entries": entries,
	})
}

// getMonthlyRewards handles GET /api/leaderboard/rewards/monthly
func (s *Server) getMonthlyRewards(c *gin.Context) {
	ctx := c.Request.Context()

	type RewardDef struct {
		Rank   int    `json:"rank"`
		Reward string `json:"reward"`
		Points int    `json:"points"`
		Emoji  string `json:"emoji"`
	}

	rewardDefs := []RewardDef{
		{Rank: 1, Reward: "🥇 Gold Badge + 2000 bonus points", Points: 2000, Emoji: "🥇"},
		{Rank: 2, Reward: "🥈 Silver Badge + 1500 bonus points", Points: 1500, Emoji: "🥈"},
		{Rank: 3, Reward: "🥉 Bronze Badge + 1000 bonus points", Points: 1000, Emoji: "🥉"},
		{Rank: 5, Reward: "500 bonus points", Points: 500, Emoji: "🎖️"},
		{Rank: 10, Reward: "200 bonus points", Points: 200, Emoji: "🎖️"},
		{Rank: 25, Reward: "100 bonus points", Points: 100, Emoji: "⭐"},
		{Rank: 50, Reward: "50 bonus points", Points: 50, Emoji: "⭐"},
		{Rank: 100, Reward: "25 bonus points", Points: 25, Emoji: "✨"},
	}

	rows, err := s.db.Query(ctx, `
		SELECT u.id, u.username, lb.monthly_score, lb.monthly_rank
		FROM leaderboard lb
		JOIN users u ON lb.user_id = u.id
		WHERE lb.monthly_rank IS NOT NULL AND lb.monthly_rank <= 100
		ORDER BY lb.monthly_rank ASC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	defer rows.Close()

	type RewardEntry struct {
		Rank     int    `json:"rank"`
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		Score    int    `json:"score"`
		Reward   string `json:"reward"`
		Points   int    `json:"points"`
		Emoji    string `json:"emoji"`
	}

	var entries []RewardEntry
	for rows.Next() {
		var userID, username string
		var score, rank int
		if err := rows.Scan(&userID, &username, &score, &rank); err != nil {
			continue
		}
		entry := RewardEntry{
			Rank:     rank,
			UserID:   userID,
			Username: username,
			Score:    score,
		}
		for _, r := range rewardDefs {
			if r.Rank == rank {
				entry.Reward = r.Reward
				entry.Points = r.Points
				entry.Emoji = r.Emoji
				break
			}
		}
		if entry.Reward == "" {
			if rank <= 25 {
				entry.Reward = "100 bonus points"
				entry.Points = 100
				entry.Emoji = "⭐"
			} else if rank <= 50 {
				entry.Reward = "50 bonus points"
				entry.Points = 50
				entry.Emoji = "⭐"
			} else {
				entry.Reward = "25 bonus points"
				entry.Points = 25
				entry.Emoji = "✨"
			}
		}
		entries = append(entries, entry)
	}
	if entries == nil {
		entries = []RewardEntry{}
	}

	c.JSON(http.StatusOK, gin.H{
		"rewards": rewardDefs,
		"entries": entries,
	})
}

// ============================================================================
// Internal Handlers
// ============================================================================

func (s *Server) internalUpdateScore(c *gin.Context) {
	var req struct {
		UserID          string `json:"user_id"`
		SubmissionID    string `json:"submission_id"`
		ProblemID       string `json:"problem_id"`
		Score           int    `json:"score"`
		Difficulty      string `json:"difficulty"`
		Status          string `json:"status"`
		ExecutionTimeMs int    `json:"execution_time_ms"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	ctx := c.Request.Context()
	passed := req.Status == "accepted"

	if err := s.updateELORating(ctx, req.UserID, req.Difficulty, passed); err != nil {
		log.Printf("Failed to update ELO: %v", err)
	}

	if passed {
		if err := s.recordSolveStreak(ctx, req.UserID); err != nil {
			log.Printf("Failed to record streak: %v", err)
		}
		_, _ = s.db.Exec(ctx, `UPDATE users SET total_solved = total_solved + 1, total_submissions = total_submissions + 1 WHERE id = $1`, req.UserID)
	} else {
		_, _ = s.db.Exec(ctx, `UPDATE users SET total_submissions = total_submissions + 1 WHERE id = $1`, req.UserID)
	}

	_, err := s.db.Exec(ctx, `
		INSERT INTO leaderboard (user_id, weekly_score, monthly_score, all_time_score)
		VALUES ($1, $2, $2, $2)
		ON CONFLICT (user_id) DO UPDATE SET
			weekly_score = leaderboard.weekly_score + $2,
			monthly_score = leaderboard.monthly_score + $2,
			all_time_score = leaderboard.all_time_score + $2,
			updated_at = NOW()
	`, req.UserID, req.Score)
	if err != nil {
		log.Printf("Failed to update score: %v", err)
	}

	newBadges, _ := s.checkAndAwardBadges(ctx, req.UserID)
	newAchievements, _ := s.checkAndAwardAchievements(ctx, req.UserID, req.Status, req.ExecutionTimeMs)

	if s.redis != nil {
		_ = s.redis.DeletePattern(ctx, "leaderboard:*")
		_ = s.redis.Delete(ctx, fmt.Sprintf("user:stats:%s", req.UserID))
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       "ok",
		"new_badges":   newBadges,
		"achievements": newAchievements,
	})
}

func (s *Server) internalRecalculate(c *gin.Context) {
	ctx := c.Request.Context()

	_, err := s.db.Exec(ctx, `
		WITH ranked AS (
			SELECT id, ROW_NUMBER() OVER (ORDER BY weekly_score DESC) AS new_rank
			FROM leaderboard WHERE weekly_score > 0
		)
		UPDATE leaderboard SET weekly_rank = ranked.new_rank
		FROM ranked WHERE leaderboard.id = ranked.id
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("weekly rank update failed: %v", err)})
		return
	}

	_, err = s.db.Exec(ctx, `
		WITH ranked AS (
			SELECT id, ROW_NUMBER() OVER (ORDER BY monthly_score DESC) AS new_rank
			FROM leaderboard WHERE monthly_score > 0
		)
		UPDATE leaderboard SET monthly_rank = ranked.new_rank
		FROM ranked WHERE leaderboard.id = ranked.id
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("monthly rank update failed: %v", err)})
		return
	}

	_, err = s.db.Exec(ctx, `
		WITH ranked AS (
			SELECT id, ROW_NUMBER() OVER (ORDER BY all_time_score DESC) AS new_rank
			FROM leaderboard WHERE all_time_score > 0
		)
		UPDATE leaderboard SET all_time_rank = ranked.new_rank
		FROM ranked WHERE leaderboard.id = ranked.id
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("all-time rank update failed: %v", err)})
		return
	}

	if s.redis != nil {
		_ = s.redis.DeletePattern(ctx, "leaderboard:*")
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
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
