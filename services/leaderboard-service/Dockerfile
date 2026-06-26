# Multi-stage build for Go services
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG SERVICE_NAME=leaderboard-service
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/bin/${SERVICE_NAME} ./services/${SERVICE_NAME}/

FROM alpine:3.18

RUN apk add --no-cache ca-certificates tzdata wget
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

WORKDIR /app

ARG SERVICE_NAME=leaderboard-service
COPY --from=builder /app/bin/${SERVICE_NAME} /app/service

USER appuser

EXPOSE 9104

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:9104/health || exit 1

ENTRYPOINT ["/app/service"]
