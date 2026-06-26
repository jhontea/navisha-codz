package middleware

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ============================================================================
// RateLimiter — sliding-window per-key rate limiter
// ============================================================================

// RateLimiter implements a simple in-memory sliding-window rate limiter.
type RateLimiter struct {
	mu       sync.RWMutex
	visitors map[string]*visitor
	limit    int
	window   time.Duration
}

type visitor struct {
	count    int
	lastSeen time.Time
}

// NewRateLimiter creates a new rate limiter with a fixed limit and window.
func NewRateLimiter(requests int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    requests,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

// Allow checks if a request from the given key is allowed.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[key]

	if !exists || now.Sub(v.lastSeen) > rl.window {
		rl.visitors[key] = &visitor{count: 1, lastSeen: now}
		return true
	}

	v.count++
	v.lastSeen = now

	if v.count > rl.limit {
		return false
	}

	return true
}

// AllowWithLimit checks if a request from the given key is allowed using a
// per-call limit and window instead of the fixed values from construction.
func (rl *RateLimiter) AllowWithLimit(key string, limit int, window time.Duration) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[key]

	if !exists || now.Sub(v.lastSeen) > window {
		rl.visitors[key] = &visitor{count: 1, lastSeen: now}
		return true
	}

	v.count++
	v.lastSeen = now

	if v.count > limit {
		return false
	}

	return true
}

// RateLimitMiddleware creates a Gin middleware that rate-limits by client IP
// (or user ID when authenticated) + request path, and sets rate limit headers.
func (rl *RateLimiter) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		if userID, exists := c.Get(ContextKeyUserID); exists {
			key = userID.(string)
		}

		key = fmt.Sprintf("%s:%s", key, c.FullPath())

		rl.mu.Lock()
		v, exists := rl.visitors[key]
		now := time.Now()

		if !exists || now.Sub(v.lastSeen) > rl.window {
			rl.visitors[key] = &visitor{count: 1, lastSeen: now}
			// Set rate limit headers
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.limit))
			c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", rl.limit-1))
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", now.Add(rl.window).Unix()))
			rl.mu.Unlock()
			c.Next()
			return
		}

		v.count++
		v.lastSeen = now
		remaining := rl.limit - v.count
		if remaining < 0 {
			remaining = 0
		}
		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.limit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", now.Add(rl.window).Unix()))
		rl.mu.Unlock()

		if v.count > rl.limit {
			c.AbortWithStatusJSON(429, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": rl.window.Seconds(),
			})
			return
		}

		c.Next()
	}
}

// GetClientIP extracts the real client IP from request headers, checking
// X-Forwarded-For and X-Real-Ip before falling back to RemoteAddr.
func GetClientIP(c *gin.Context) string {
	xff := c.GetHeader("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	xri := c.GetHeader("X-Real-Ip")
	if xri != "" {
		return xri
	}

	ip, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
	return ip
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, v := range rl.visitors {
			if now.Sub(v.lastSeen) > rl.window {
				delete(rl.visitors, key)
			}
		}
		rl.mu.Unlock()
	}
}

// ============================================================================
// EndpointRateLimiter — per-endpoint-group rate limiting
// ============================================================================

// EndpointRateLimiter provides rate limiting per endpoint group.
type EndpointRateLimiter struct {
	mu       sync.RWMutex
	limits   map[string]int
	windows  map[string]time.Duration
	visitors map[string]*endpointVisitor
}

type endpointVisitor struct {
	count    int
	lastSeen time.Time
}

// NewEndpointRateLimiter creates a new endpoint rate limiter.
func NewEndpointRateLimiter() *EndpointRateLimiter {
	erl := &EndpointRateLimiter{
		limits:   make(map[string]int),
		windows:  make(map[string]time.Duration),
		visitors: make(map[string]*endpointVisitor),
	}
	go erl.cleanup()
	return erl
}

// SetLimit sets the rate limit for an endpoint group.
func (erl *EndpointRateLimiter) SetLimit(group string, requests int, window time.Duration) {
	erl.mu.Lock()
	defer erl.mu.Unlock()
	erl.limits[group] = requests
	erl.windows[group] = window
}

// Allow checks if a request is allowed for the given endpoint group and key.
func (erl *EndpointRateLimiter) Allow(group, key string) bool {
	erl.mu.RLock()
	limit, hasLimit := erl.limits[group]
	window, hasWindow := erl.windows[group]
	erl.mu.RUnlock()

	if !hasLimit || !hasWindow {
		return true // No limit configured
	}

	visitorKey := fmt.Sprintf("%s:%s", group, key)

	erl.mu.Lock()
	defer erl.mu.Unlock()

	v, exists := erl.visitors[visitorKey]
	now := time.Now()

	if !exists || now.Sub(v.lastSeen) > window {
		erl.visitors[visitorKey] = &endpointVisitor{count: 1, lastSeen: now}
		return true
	}

	v.count++
	v.lastSeen = now

	if v.count > limit {
		return false
	}

	return true
}

func (erl *EndpointRateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		erl.mu.Lock()
		now := time.Now()

		var maxWindow time.Duration
		for _, w := range erl.windows {
			if w > maxWindow {
				maxWindow = w
			}
		}

		for key, v := range erl.visitors {
			if now.Sub(v.lastSeen) > maxWindow*2 {
				delete(erl.visitors, key)
			}
		}
		erl.mu.Unlock()
	}
}
