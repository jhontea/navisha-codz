# Multi-stage build for Go services
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the service (SERVICE_NAME is passed as build arg)
ARG SERVICE_NAME=auth-service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/bin/${SERVICE_NAME} ./services/${SERVICE_NAME}/

# Final stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder
ARG SERVICE_NAME=auth-service
COPY --from=builder /app/bin/${SERVICE_NAME} /app/service

# Copy any config files if needed
COPY --chown=appuser:appgroup .env* ./

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 9101

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9101/health || exit 1

# Run
ENTRYPOINT ["/app/service"]
