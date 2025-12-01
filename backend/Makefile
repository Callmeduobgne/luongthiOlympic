.PHONY: help build run test clean docker-build docker-up docker-down docker-logs docker-restart migrate-up migrate-down status

# Default target
help:
	@echo "IBN Backend - Available Commands:"
	@echo ""
	@echo "Local Development:"
	@echo "  make build          - Build the Go binary"
	@echo "  make run            - Run the application locally"
	@echo "  make test           - Run tests"
	@echo "  make clean          - Clean build artifacts"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-up      - Start backend container"
	@echo "  make docker-down    - Stop backend container"
	@echo "  make docker-restart - Restart backend container"
	@echo "  make docker-logs    - View container logs"
	@echo "  make status         - Show container status"
	@echo ""
	@echo "Database:"
	@echo "  make migrate-up     - Run database migrations"
	@echo "  make migrate-down   - Rollback database migrations"
	@echo ""
	@echo "Security:"
	@echo "  make generate-key   - Generate encryption master key"
	@echo ""
	@echo "Documentation:"
	@echo "  make swagger        - Generate Swagger/OpenAPI documentation"
	@echo ""
	@echo "Quick Start:"
	@echo "  make start          - Build and start everything"
	@echo ""

# Build the Go binary
build:
	@echo "Building IBN Backend..."
	@go build -o bin/server ./cmd/server
	@echo "Build complete: bin/server"

# Run locally
run:
	@echo "Starting IBN Backend..."
	@go run ./cmd/server

# Run tests
test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@rm -rf docs/swagger
	@echo "Clean complete"

# Generate Swagger documentation
swagger:
	@echo "Generating Swagger docs..."
	@which swag > /dev/null || (echo "Installing swag..." && go install github.com/swaggo/swag/cmd/swag@latest)
	@mkdir -p docs/swagger
	@swag init -g cmd/server/main.go -o docs/swagger --parseDependency --parseInternal
	@echo "Swagger docs generated: docs/swagger/"
	@echo "Swagger UI available at: http://localhost:9090/swagger/index.html"

# Docker commands
docker-build:
	@echo "Building Docker image..."
	@docker-compose build

docker-up:
	@echo "Starting IBN Backend..."
	@docker-compose up -d
	@echo "Backend started on http://localhost:9090"
	@echo "Waiting for backend to be healthy..."
	@sleep 5
	@docker-compose ps
	@echo ""
	@echo "Health check:"
	@curl -s http://localhost:9090/health | jq '.' || echo "Backend not ready yet, check logs: make docker-logs"

docker-down:
	@echo "Stopping IBN Backend..."
	@docker-compose down

docker-logs:
	@docker-compose logs -f backend

docker-restart:
	@echo "Restarting IBN Backend..."
	@docker-compose restart backend
	@echo "Waiting for backend..."
	@sleep 3
	@curl -s http://localhost:9090/health | jq '.'

# Database migrations
migrate-up:
	@echo "Running database migrations..."
	@export PGPASSWORD=changeme && \
	for migration in migrations/*.up.sql; do \
		echo "Running: $$(basename $$migration)"; \
		psql -h localhost -p 5432 -U gateway -d ibn_gateway -f "$$migration" 2>&1 | grep -v "already exists\|NOTICE" || true; \
	done
	@echo "Migrations complete"

migrate-down:
	@echo "Rolling back database migrations..."
	@export PGPASSWORD=changeme && \
	for migration in $$(ls -r migrations/*.down.sql); do \
		echo "Rolling back: $$(basename $$migration)"; \
		psql -h localhost -p 5432 -U gateway -d ibn_gateway -f "$$migration" || true; \
	done

# Generate encryption master key
generate-key:
	@./scripts/generate-master-key.sh

# Quick start
start: docker-build migrate-up docker-up
	@echo ""
	@echo "========================================="
	@echo "IBN Backend started successfully!"
	@echo "========================================="
	@echo "URL: http://localhost:9090"
	@echo ""
	@echo "Useful commands:"
	@echo "  make docker-logs    - View logs"
	@echo "  make status         - Check status"
	@echo "  make docker-down    - Stop backend"
	@echo ""

# Status
status:
	@echo "=== Docker Containers ==="
	@docker-compose ps
	@echo ""
	@echo "=== Health Check ==="
	@curl -s http://localhost:9090/health | jq '.' || echo "Backend not available"
	@echo ""
	@echo "=== Container Logs (last 10 lines) ==="
	@docker-compose logs --tail=10 backend
