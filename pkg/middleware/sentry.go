package middleware

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

// sentryEnabled tracks whether Sentry was successfully initialized.
var sentryEnabled bool

// InitSentry initializes the Sentry SDK from the SENTRY_DSN environment variable.
// If SENTRY_DSN is empty or unset, Sentry is disabled silently.
func InitSentry() error {
	dsn := os.Getenv("SENTRY_DSN")
	if dsn == "" {
		sentryEnabled = false
		return nil
	}

	environment := os.Getenv("APP_ENV")
	if environment == "" {
		environment = "development"
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Environment:      environment,
		EnableTracing:    true,
		TracesSampleRate: 0.2,
	})
	if err != nil {
		sentryEnabled = false
		return fmt.Errorf("sentry init failed: %w", err)
	}

	sentryEnabled = true
	return nil
}

// FlushSentry flushes buffered Sentry events. Call this during server shutdown.
func FlushSentry() {
	if !sentryEnabled {
		return
	}
	sentry.Flush(5 * time.Second)
}

// SentryMiddleware returns a Gin middleware that:
//   - Recovers from panics and sends them to Sentry
//   - Adds request context (method, path, IP) to Sentry events
func SentryMiddleware() gin.HandlerFunc {
	if !sentryEnabled {
		// Pass-through no-op middleware when Sentry is disabled
		return func(c *gin.Context) {
			c.Next()
		}
	}

	hub := sentry.CurrentHub().Clone()
	return func(c *gin.Context) {
		// Skip for health check and swagger to reduce noise
		path := c.Request.URL.Path
		if path == "/health" || path == "/swagger" {
			c.Next()
			return
		}

		ctx := sentry.SetHubOnContext(c.Request.Context(), hub)
		c.Request = c.Request.WithContext(ctx)

		// Set request-level Sentry tags
		scope := hub.PushScope()
		scope.SetTag("method", c.Request.Method)
		scope.SetTag("path", path)
		scope.SetTag("client_ip", c.ClientIP())
		scope.SetRequest(c.Request)

		defer func() {
			if r := recover(); r != nil {
				hub.RecoverWithContext(ctx, r)
				// Re-panic so Gin's default recovery still returns a 500
				panic(r)
			}
			hub.PopScope()
		}()

		c.Next()
	}
}

// SentryGinMiddleware returns the official sentry-go Gin middleware
// which captures panics and request context automatically.
func SentryGinMiddleware() gin.HandlerFunc {
	if !sentryEnabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}
	return sentrygin.New(sentrygin.Options{})
}

// CaptureError sends an error to Sentry with request context.
// Call this before returning an error response (5xx) to ensure the error is tracked.
func CaptureError(c *gin.Context, err error, msg string) {
	if !sentryEnabled || err == nil {
		return
	}

	hub := sentry.GetHubFromContext(c.Request.Context())
	if hub == nil {
		hub = sentry.CurrentHub()
	}

	hub.WithScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentry.LevelError)
		scope.SetTag("method", c.Request.Method)
		scope.SetTag("path", c.Request.URL.Path)
		scope.SetTag("client_ip", c.ClientIP())

		if userID, exists := c.Get(ContextKeyUserID); exists {
			scope.SetUser(sentry.User{ID: fmt.Sprintf("%v", userID)})
		}

		scope.SetContext("request", map[string]interface{}{
			"status": c.Writer.Status(),
		})
		scope.SetRequest(c.Request)

		if msg != "" {
			hub.CaptureMessage(msg)
		}
		if err != nil {
			hub.CaptureException(err)
		}
	})
}

// IsSentryEnabled returns whether Sentry is active.
func IsSentryEnabled() bool {
	return sentryEnabled
}

// InternalServerError writes a 500 response and captures the error in Sentry.
func InternalServerError(c *gin.Context, err error) {
	CaptureError(c, err, "internal server error")
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": "internal server error",
	})
}
