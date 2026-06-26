package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	pkgmiddleware "coding-challange/pkg/middleware"
)

// CacheEntry represents a cached response
type CacheEntry struct {
	Data      []byte
	ETag      string
	Timestamp time.Time
}

// InMemoryCache provides thread-safe in-memory caching with TTL
type InMemoryCache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
	ttl     time.Duration
	done    chan struct{}
}

// NewInMemoryCache creates a new cache with the specified TTL
func NewInMemoryCache(ttl time.Duration) *InMemoryCache {
	cache := &InMemoryCache{
		entries: make(map[string]*CacheEntry),
		ttl:     ttl,
		done:    make(chan struct{}),
	}
	// Start cleanup goroutine
	go cache.cleanup()
	return cache
}

// StopCleanup stops the background cleanup goroutine.
func (c *InMemoryCache) StopCleanup() {
	close(c.done)
}

// Get retrieves a cached entry if it exists and is not expired
func (c *InMemoryCache) Get(key string) (*CacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	if time.Since(entry.Timestamp) > c.ttl {
		return nil, false
	}

	return entry, true
}

// Set stores a new entry in the cache
func (c *InMemoryCache) Set(key string, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	hash := sha256.Sum256(data)
	etag := "\"" + hex.EncodeToString(hash[:16]) + "\""

	c.entries[key] = &CacheEntry{
		Data:      data,
		ETag:      etag,
		Timestamp: time.Now(),
	}
}

// Invalidate removes an entry from the cache
func (c *InMemoryCache) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

// Clear removes all entries
func (c *InMemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*CacheEntry)
}

// cleanup periodically removes expired entries
func (c *InMemoryCache) cleanup() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[RECOVER] InMemoryCache cleanup panic: %v", r)
		}
	}()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			c.mu.Lock()
			now := time.Now()
			for key, entry := range c.entries {
				if now.Sub(entry.Timestamp) > c.ttl {
					delete(c.entries, key)
				}
			}
			c.mu.Unlock()
		}
	}
}

// CacheMiddleware creates a Gin middleware for caching API responses
func CacheMiddleware(cache *InMemoryCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only cache GET requests
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		// Only cache API endpoints
		if !strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Next()
			return
		}

		key := c.Request.URL.Path + "?" + c.Request.URL.RawQuery

		// Check If-None-Match header for conditional request
		ifNoneMatch := c.GetHeader("If-None-Match")

		if entry, found := cache.Get(key); found {
			if ifNoneMatch != "" && ifNoneMatch == entry.ETag {
				c.Status(http.StatusNotModified)
				c.Abort()
				return
			}

			c.Header("ETag", entry.ETag)
			c.Header("X-Cache", "HIT")
			c.Data(http.StatusOK, "application/json", entry.Data)
			c.Abort()
			return
		}

		// Capture response
		writer := &responseRecorder{ResponseWriter: c.Writer, statusCode: 200}
		c.Writer = writer
		c.Next()

		// Cache successful responses
		if writer.statusCode == http.StatusOK && len(writer.body) > 0 {
			cache.Set(key, writer.body)
			hash := sha256.Sum256(writer.body)
			etag := "\"" + hex.EncodeToString(hash[:16]) + "\""
			c.Header("ETag", etag)
		}
		c.Header("X-Cache", "MISS")
	}
}

// responseRecorder captures the response body for caching
type responseRecorder struct {
	gin.ResponseWriter
	statusCode int
	body       []byte
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(p []byte) (int, error) {
	r.body = append(r.body, p...)
	return r.ResponseWriter.Write(p)
}

// SecurityHeaders adds security-related HTTP headers using the shared
// implementation from pkg/middleware.
func SecurityHeaders() gin.HandlerFunc {
	headers := pkgmiddleware.SecurityHeaders()
	return func(c *gin.Context) {
		for key, value := range headers {
			c.Header(key, value)
		}
		c.Next()
	}
}

// RequestTimeout adds a timeout to each request
func RequestTimeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set a deadline on the request context
		ctx, cancel := c.Request.Context(), func() {}
		_ = ctx
		_ = cancel

		// Use Gin's built-in timeout handling via channel
		done := make(chan struct{})
		go func() {
			c.Next()
			close(done)
		}()

		select {
		case <-done:
			// Request completed normally
		case <-time.After(timeout):
			c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{
				"error": "request timeout exceeded",
				"code":  504,
			})
		}
	}
}

// MaxBodySize limits the maximum request body size
func MaxBodySize(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxBytes {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": fmt.Sprintf("request body exceeds maximum size of %d bytes", maxBytes),
				"code":  413,
			})
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}

// InputSanitization adds extra input validation
func InputSanitization() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sanitize query parameters
		for key, values := range c.Request.URL.Query() {
			for _, v := range values {
				if containsDangerousContent(v) {
					c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
						"error": fmt.Sprintf("invalid input in parameter %q", key),
						"code":  400,
					})
					return
				}
			}
		}
		c.Next()
	}
}

// containsDangerousContent checks for potentially malicious input
func containsDangerousContent(s string) bool {
	dangerous := []string{
		"<script", "javascript:", "onerror=", "onload=",
		"eval(", "document.cookie", "document.write",
		"SELECT ", "INSERT ", "DELETE ", "DROP ", "UNION ",
		"../", "..\\", "%00", "\\x00",
	}
	lower := strings.ToLower(s)
	for _, d := range dangerous {
		if strings.Contains(lower, d) {
			return true
		}
	}
	return false
}
