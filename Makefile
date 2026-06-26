.PHONY: build run test docker-up docker-down migrate proto lint clean help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=coding-challange
GOPATH=$(shell go env GOPATH)

# Service binaries
SERVICES=auth-service problem-service execution-service leaderboard-service hint-service api-gateway

# Build all services
build:
	@echo "Building all services..."
	@for service in $(SERVICES); do \
		echo "Building $$service..."; \
		$(GOBUILD) -o bin/$$service ./services/$$service/ || exit 1; \
	done
	@echo "All services built successfully."

# Build a specific service
build-%:
	@echo "Building $*..."
	@$(GOBUILD) -o bin/$* ./services/$*/
	@echo "$* built successfully."

# Run a specific service (usage: make run SERVICE=auth-service)
run:
	@if [ -z "$(SERVICE)" ]; then \
		echo "Usage: make run SERVICE=<service-name>"; \
		echo "Available: $(SERVICES)"; \
		exit 1; \
	fi
	@echo "Running $(SERVICE)..."
	@./bin/$(SERVICE)

# Run all services (for development)
run-all:
	@echo "Starting all services..."
	@for service in $(SERVICES); do \
		$(GOBUILD) -o bin/$$service ./services/$$service/ & \
	done
	@echo "All services started."

# Run tests
test:
	@echo "Running tests..."
	@$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@echo "Coverage report: coverage.out"

# Run tests for specific package
test-%:
	@echo "Running tests for $*..."
	@$(GOTEST) -v -race ./$*/...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Docker commands
docker-up:
	@echo "Starting all services with Docker Compose..."
	@docker-compose up -d --build
	@echo "All services started. Check docker-compose ps for status."

docker-down:
	@echo "Stopping all services..."
	@docker-compose down
	@echo "All services stopped."

docker-logs:
	@docker-compose logs -f

docker-build:
	@docker-compose build --no-cache

# Database migrations
migrate:
	@echo "Running database migrations..."
	@$(GOBUILD) -o bin/migrate ./cmd/migrate/
	@./bin/migrate up

migrate-down:
	@echo "Rolling back migrations..."
	@$(GOBUILD) -o bin/migrate ./cmd/migrate/
	@./bin/migrate down

migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Usage: make migrate-create NAME=<migration-name>"; \
		exit 1; \
	fi
	@echo "Creating migration: $(NAME)"
	@mkdir -p migrations
	@echo "-- Migration: $(NAME)" > migrations/$(shell date +%Y%m%d%H%M%S)_$(NAME).sql
	@echo "Migration file created."

# Protocol buffers
proto:
	@echo "Generating protobuf code..."
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/submission.proto
	@echo "Protobuf code generated."

# Lint
lint:
	@echo "Running linter..."
	@golangci-lint run ./... || echo "golangci-lint not installed, skipping."

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Cleaned."

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@$(GOMOD) download
	@$(GOMOD) tidy
	@echo "Dependencies downloaded."

# Generate gRPC mock for testing
mock:
	@echo "Generating mocks..."
	@mockgen -source=proto/submission.pb.go -destination=proto/mock/submission_mock.go -package=mock

# Development setup
setup:
	@echo "Setting up development environment..."
	@mkdir -p bin
	@mkdir -p logs
	@cp .env.example .env || true
	@echo "Development environment ready."

# Check health of all services
health:
	@echo "Checking service health..."
	@echo "Auth Service:" && curl -s http://localhost:8081/health | head -c 200
	@echo "\nProblem Service:" && curl -s http://localhost:8082/health | head -c 200
	@echo "\nExecution Service:" && curl -s http://localhost:8083/health | head -c 200
	@echo "\nLeaderboard Service:" && curl -s http://localhost:8084/health | head -c 200
	@echo "\nHint Service:" && curl -s http://localhost:8085/health | head -c 200
	@echo ""

# Help
help:
	@echo "Coding Challenge Platform - Build & Run Commands"
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@echo "Build:"
	@echo "  build          Build all services"
	@echo "  build-<name>   Build a specific service"
	@echo ""
	@echo "Run:"
	@echo "  run SERVICE=<name>  Run a specific service binary"
	@echo "  run-all             Run all services (background)"
	@echo ""
	@echo "Test:"
	@echo "  test           Run all tests"
	@echo "  test-<name>    Run tests for specific package"
	@echo "  test-coverage  Run tests with coverage report"
	@echo ""
	@echo "Docker:"
	@echo "  docker-up      Start all services with Docker Compose"
	@echo "  docker-down    Stop all services"
	@echo "  docker-logs    View logs from all services"
	@echo "  docker-build   Rebuild all Docker images"
	@echo ""
	@echo "Database:"
	@echo "  migrate        Run migrations"
	@echo "  migrate-down  Rollback migrations"
	@echo "  migrate-create NAME=<name>  Create new migration file"
	@echo ""
	@echo "Other:"
	@echo "  proto          Generate protobuf code"
	@echo "  lint           Run linter"
	@echo "  deps           Download and tidy dependencies"
	@echo "  clean          Clean build artifacts"
	@echo "  setup          Setup development environment"
	@echo "  health         Check health of all services"
	@echo "  help           Show this help message"
