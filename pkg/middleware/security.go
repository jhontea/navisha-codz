package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ============================================================================
// Security Headers (net/http)
// ============================================================================

// XSSHeaders returns security headers to prevent XSS attacks.
func XSSHeaders() map[string]string {
	return map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"X-XSS-Protection":          "1; mode=block",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
		"Content-Security-Policy":   "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'",
		"Permissions-Policy":        "camera=(), microphone=(), geolocation=(), payment=()",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
	}
}

// SecureHeadersMiddleware returns an HTTP middleware that adds security headers
// for use with standard net/http handlers.
func SecureHeadersMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			headers := XSSHeaders()
			for key, value := range headers {
				w.Header().Set(key, value)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ============================================================================
// Security Headers (Gin)
// ============================================================================

// SecurityHeaders returns a map of recommended security HTTP headers for Gin.
func SecurityHeaders() map[string]string {
	return map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"X-XSS-Protection":          "1; mode=block",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
		"Content-Security-Policy":   "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'",
		"Permissions-Policy":        "camera=(), microphone=(), geolocation=(), payment=()",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
	}
}

// SecurityHeadersMiddleware creates a Gin middleware that adds security
// headers to every response.
func SecurityHeadersMiddleware() gin.HandlerFunc {
	headers := SecurityHeaders()
	return func(c *gin.Context) {
		for key, value := range headers {
			c.Header(key, value)
		}
		c.Next()
	}
}
