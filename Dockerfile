# Multi-stage build for Coding Challenge Website
# Build: docker build -t coding-challenge:latest .
# Run: docker run -d -p 9100:9100 -v $(pwd)/problems:/app/problems:ro coding-challenge:latest

# ---- Build Stage ----
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first (for layer caching)
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o server cmd/server/main.go

# ---- Runtime Stage ----
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder --chown=appuser:appgroup /app/server .

# Copy static assets
COPY --from=builder --chown=appuser:appgroup /app/web ./web

# Copy problems directory
COPY --from=builder --chown=appuser:appgroup /app/problems ./problems

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 9100

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9100/health || exit 1

# Run the server
CMD ["./server"]
