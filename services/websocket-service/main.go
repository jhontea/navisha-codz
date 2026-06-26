package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"coding-challange/pkg/database"
	"coding-challange/pkg/redis"
	"coding-challange/pkg/websocket"
)

// Service configuration.
const (
	ServiceName = "websocket-service"
	ServicePort = "9107"
)

// Server holds all dependencies for the WebSocket service.
type Server struct {
	router    *gin.Engine
	db        *database.Pool
	redis     *redis.Client
	wsHub     *websocket.Hub
	httpSrv   *http.Server
	mu        sync.RWMutex
	stats     ServerStats
}

// ServerStats holds runtime statistics.
type ServerStats struct {
	StartedAt       time.Time `json:"started_at"`
	TotalConnections int      `json:"total_connections"`
	MessagesSent    int64     `json:"messages_sent"`
}

// WebSocketMessage represents an incoming WebSocket message.
type WSClientMessage struct {
	Type    string      `json:"type"`
	Room    string      `json:"room,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
}

// NewServer creates a new WebSocket service server.
func NewServer() *Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(requestIDMiddleware())
	router.Use(loggerMiddleware())
	router.Use(corsMiddleware())

	return &Server{
		router: router,
		wsHub:  websocket.NewHub(),
		stats: ServerStats{
			StartedAt: time.Now().UTC(),
		},
	}
}

func main() {
	log.Printf("Starting %s...", ServiceName)

	// Initialize database (optional - for user validation)
	var dbPool *database.Pool
	if os.Getenv("DB_HOST") != "" {
		var err error
		dbPool, err = database.NewFromEnv()
		if err != nil {
			log.Printf("Warning: Database not available: %v", err)
		} else {
			defer dbPool.Close()
		}
	}

	// Initialize Redis (optional - for pub/sub)
	var redisClient *redis.Client
	if os.Getenv("REDIS_ADDR") != "" {
		var err error
		redisClient, err = redis.NewFromEnv()
		if err != nil {
			log.Printf("Warning: Redis not available: %v", err)
		} else {
			defer redisClient.Close()
		}
	}

	// Create server
	srv := NewServer()
	srv.db = dbPool
	srv.redis = redisClient

	// Setup routes
	srv.setupRoutes()

	// Start WebSocket hub in background
	go srv.wsHub.Run()

	// Start Redis subscriber for cross-service messaging
	var redisSubCtx context.Context
	var redisSubCancel context.CancelFunc
	if redisClient != nil {
		redisSubCtx, redisSubCancel = context.WithCancel(context.Background())
		go srv.startRedisSubscriber(redisSubCtx)
	}

	// Start HTTP server
	port := getEnv("PORT", ServicePort)
	srv.start(port)

	// Stop background goroutines on shutdown
	srv.wsHub.Stop()
	if redisSubCancel != nil {
		redisSubCancel()
	}
}

// setupRoutes configures all HTTP and WebSocket routes.
func (s *Server) setupRoutes() {
	// Health check
	s.router.GET("/health", s.healthCheck)

	// Stats endpoint
	s.router.GET("/stats", s.getStats)

	// WebSocket endpoints
	ws := s.router.Group("/ws")
	{
		// Main WebSocket endpoint for client connections
		ws.GET("/connect", s.handleWebSocketConnect)

		// Room-specific WebSocket endpoint
		ws.GET("/room/:roomName", s.handleRoomWebSocket)

		// Submission-specific WebSocket
		ws.GET("/submission/:id", s.handleSubmissionWebSocket)
	}

	// Internal API for service-to-service communication
	internal := s.router.Group("/internal")
	{
		internal.POST("/broadcast", s.handleInternalBroadcast)
		internal.POST("/notify/user/:userId", s.handleInternalUserNotify)
		internal.POST("/notify/room/:roomName", s.handleInternalRoomNotify)
	}
}

// healthCheck handles health check requests.
func (s *Server) healthCheck(c *gin.Context) {
	ctx := c.Request.Context()
	status := "ok"
	dbStatus := "ok"
	redisStatus := "ok"

	if s.db != nil {
		if err := s.db.HealthCheck(ctx); err != nil {
			dbStatus = "unavailable"
			status = "degraded"
		}
	}

	if s.redis != nil {
		if err := s.redis.HealthCheck(ctx); err != nil {
			redisStatus = "unavailable"
			status = "degraded"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"service":  ServiceName,
		"status":   status,
		"database":  dbStatus,
		"redis":    redisStatus,
		"websocket": s.wsHub.Stats(),
		"stats":     s.getStatsData(),
		"time":      time.Now().UTC(),
	})
}

// getStats returns WebSocket service statistics.
func (s *Server) getStats(c *gin.Context) {
	c.JSON(http.StatusOK, s.getStatsData())
}

func (s *Server) getStatsData() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"service":           ServiceName,
		"started_at":        s.stats.StartedAt,
		"uptime_seconds":    time.Since(s.stats.StartedAt).Seconds(),
		"total_connections": s.stats.TotalConnections,
		"messages_sent":     s.stats.MessagesSent,
		"websocket":         s.wsHub.Stats(),
		"time":              time.Now().UTC(),
	}
}

// handleWebSocketConnect handles new WebSocket connections.
// User ID can be provided via query parameter or token.
func (s *Server) handleWebSocketConnect(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		// Try to get from token
		userID = c.GetHeader("X-User-ID")
	}
	if userID == "" {
		// Generate anonymous user ID
		userID = "anon-" + uuid.New().String()[:8]
	}

	s.wsHub.HandleWebSocket(c, userID)
	s.incrementConnectionCount()
}

// handleRoomWebSocket handles room-specific WebSocket connections.
func (s *Server) handleRoomWebSocket(c *gin.Context) {
	roomName := c.Param("roomName")
	userID := c.Query("user_id")
	if userID == "" {
		userID = "anon-" + uuid.New().String()[:8]
	}

	// Handle WebSocket and auto-join room
	s.wsHub.HandleWebSocket(c, userID)
	s.wsHub.JoinRoom(getLastClientID(c), roomName)
	s.incrementConnectionCount()
}

// handleSubmissionWebSocket handles submission-specific WebSocket connections.
// Clients automatically join the submission-{id} room.
func (s *Server) handleSubmissionWebSocket(c *gin.Context) {
	submissionID := c.Param("id")
	userID := c.Query("user_id")
	if userID == "" {
		userID = "anon-" + uuid.New().String()[:8]
	}

	roomName := "submission-" + submissionID
	s.wsHub.HandleWebSocket(c, userID)
	s.wsHub.JoinRoom(getLastClientID(c), roomName)
	s.incrementConnectionCount()
}

// handleInternalBroadcast handles internal broadcast requests.
func (s *Server) handleInternalBroadcast(c *gin.Context) {
	var msg websocket.Message
	if err := c.ShouldBindJSON(&msg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message"})
		return
	}

	s.wsHub.Broadcast(&msg)
	c.JSON(http.StatusOK, gin.H{"status": "broadcast sent"})
}

// handleInternalUserNotify sends a notification to a specific user.
func (s *Server) handleInternalUserNotify(c *gin.Context) {
	userID := c.Param("userId")

	var msg websocket.Message
	if err := c.ShouldBindJSON(&msg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message"})
		return
	}

	s.wsHub.SendToUser(userID, &msg)
	c.JSON(http.StatusOK, gin.H{"status": "notification sent", "user_id": userID})
}

// handleInternalRoomNotify sends a notification to all clients in a room.
func (s *Server) handleInternalRoomNotify(c *gin.Context) {
	roomName := c.Param("roomName")

	var msg websocket.Message
	if err := c.ShouldBindJSON(&msg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid message"})
		return
	}

	s.wsHub.SendToRoom(roomName, &msg)
	c.JSON(http.StatusOK, gin.H{"status": "notification sent", "room": roomName})
}

// startRedisSubscriber subscribes to Redis pub/sub channels for cross-service messaging.
func (s *Server) startRedisSubscriber(ctx context.Context) {
	ch := s.redis.Subscribe(ctx, "submission:result", "leaderboard:update", "notification:broadcast")

	for msg := range ch {
		var wsMsg websocket.Message
		if err := json.Unmarshal([]byte(msg.Payload), &wsMsg); err != nil {
			log.Printf("Invalid Redis message: %v", err)
			continue
		}

		// Route message based on channel
		switch msg.Channel {
		case "submission:result":
			if wsMsg.Room != "" {
				s.wsHub.SendToRoom(wsMsg.Room, &wsMsg)
			}
		case "leaderboard:update":
			s.wsHub.Broadcast(&wsMsg)
		case "notification:broadcast":
			if wsMsg.Room != "" {
				s.wsHub.SendToRoom(wsMsg.Room, &wsMsg)
			}
		}
	}
}

// incrementConnectionCount increments the total connections counter.
func (s *Server) incrementConnectionCount() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stats.TotalConnections++
}

// start starts the HTTP server with graceful shutdown.
func (s *Server) start(port string) {
	addr := ":" + port
	s.httpSrv = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("%s starting on %s", ServiceName, addr)
		if err := s.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Printf("Shutting down %s...", ServiceName)

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpSrv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Printf("%s exited gracefully", ServiceName)
}

// Middleware

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		log.Printf("[WS] %s %s %d %v", c.Request.Method, path+"?"+query, status, latency)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID, X-Request-ID")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// getLastClientID is a helper to get the most recently connected client.
// In a real implementation, this would track client IDs more carefully.
func getLastClientID(c *gin.Context) string {
	// This is a simplified approach - in production, you'd track this properly
	return c.Query("client_id")
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
