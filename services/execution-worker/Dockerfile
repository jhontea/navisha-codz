# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the execution worker
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o /execution-worker \
    ./services/execution-worker/

# Runtime stage
FROM alpine:3.18

RUN apk add --no-cache ca-certificates tzdata bash curl && \
    addgroup -S appgroup && \
    adduser -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder
COPY --from=builder /execution-worker /app/execution-worker

# Create directories for sandbox
RUN mkdir -p /tmp/sandbox && chown appuser:appgroup /tmp/sandbox

# Switch to non-root user
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8086/health || exit 1

EXPOSE 9106

ENTRYPOINT ["/app/execution-worker"]
