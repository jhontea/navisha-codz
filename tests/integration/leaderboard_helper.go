package integration

import (
	"github.com/gin-gonic/gin"
)

// ginRouterWithLeaderboard creates a test router with leaderboard endpoints.
func ginRouterWithLeaderboard() *gin.Engine {
	router := gin.New()

	router.GET("/api/leaderboard", func(c *gin.Context) {
		problemID := c.Query("problem_id")
		limit := c.DefaultQuery("limit", "10")

		leaderboard := []map[string]interface{}{
			{"rank": 1, "user_id": "user-1", "username": "alice", "score": 100, "problems_solved": 10},
			{"rank": 2, "user_id": "user-2", "username": "bob", "score": 90, "problems_solved": 8},
			{"rank": 3, "user_id": "user-3", "username": "charlie", "score": 80, "problems_solved": 7},
		}

		c.JSON(200, gin.H{
			"data": leaderboard,
			"meta": gin.H{"problem_id": problemID, "limit": limit},
		})
	})

	router.POST("/api/leaderboard/score", func(c *gin.Context) {
		var req struct {
			UserID    string `json:"user_id" binding:"required"`
			ProblemID string `json:"problem_id" binding:"required"`
			Score     int    `json:"score" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{
			"user_id":    req.UserID,
			"problem_id": req.ProblemID,
			"score":      req.Score,
			"rank":       1,
			"updated":    true,
		})
	})

	router.GET("/api/leaderboard/user/:id", func(c *gin.Context) {
		userID := c.Param("id")
		c.JSON(200, gin.H{
			"user_id":  userID,
			"rank":     1,
			"score":    100,
			"solved":   5,
		})
	})

	return router
}
