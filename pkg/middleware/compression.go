package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// GzipCompressLevel is the compression level for gzip responses.
const GzipCompressLevel = gzip.DefaultCompression

// GzipMiddleware returns a Gin middleware that compresses HTTP responses using Gzip
// when the client supports it (Accept-Encoding: gzip).
func GzipMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only compress responses, not upgrade connections (WebSocket)
		if strings.Contains(c.GetHeader("Connection"), "Upgrade") {
			c.Next()
			return
		}

		// Check if client accepts gzip encoding
		acceptEncoding := c.GetHeader("Accept-Encoding")
		if !strings.Contains(acceptEncoding, "gzip") {
			c.Next()
			return
		}

		// Skip if response is already written (e.g., from a previous middleware)
		if c.Writer.Header().Get("Content-Encoding") != "" {
			c.Next()
			return
		}

		// Wrap the response writer with gzip
		gzipWriter, err := gzip.NewWriterLevel(c.Writer, GzipCompressLevel)
		if err != nil {
			c.Next()
			return
		}
		defer gzipWriter.Close()

		c.Writer = &gzipResponseWriter{
			ResponseWriter: c.Writer,
			writer:         gzipWriter,
		}
		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")

		c.Next()
	}
}

// gzipResponseWriter wraps gin.ResponseWriter with gzip compression.
type gzipResponseWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

// Write compresses and writes data to the underlying response writer.
func (g *gzipResponseWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}

// WriteString compresses and writes a string to the underlying response writer.
func (g *gzipResponseWriter) WriteString(s string) (int, error) {
	return g.writer.Write([]byte(s))
}

// Flush ensures all compressed data is written to the underlying writer.
func (g *gzipResponseWriter) Flush() {
	g.writer.Flush()
	if flusher, ok := g.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
