# Makefile for IBN Network
# Unified build and test commands for the entire project

.PHONY: all build test clean docker-up docker-down lint help

# Default target
all: build

# Help command
help:
	@echo "IBN Network Makefile"
	@echo "===================="
	@echo "Available commands:"
	@echo "  make build       - Build all components (Backend, Frontend, Chaincode)"
	@echo "  make test        - Run tests for all components"
	@echo "  make clean       - Clean up build artifacts"
	@echo "  make docker-up   - Start all services with Docker Compose"
	@echo "  make docker-down - Stop all services"
	@echo "  make lint        - Run linters"

# Build all components
build: build-backend build-frontend build-chaincode

build-backend:
	@echo "Building Backend..."
	@cd backend && go build -v -o ibn-backend ./cmd/server

build-frontend:
	@echo "Building Frontend..."
	@cd frontend && npm install && npm run build

build-chaincode:
	@echo "Building Chaincode..."
	@cd teaTraceCC && npm install && npm run build

# Run tests
test: test-backend test-frontend test-chaincode

test-backend:
	@echo "Testing Backend..."
	@cd backend && go test -v ./...

test-frontend:
	@echo "Testing Frontend..."
	@cd frontend && npm test -- --watchAll=false

test-chaincode:
	@echo "Testing Chaincode..."
	@cd teaTraceCC && npm test

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	@rm -f backend/ibn-backend
	@rm -rf frontend/dist
	@rm -rf teaTraceCC/dist
	@rm -f teaTraceCC.tar.gz

# Docker commands
docker-up:
	@echo "Starting services..."
	@docker compose up -d

docker-down:
	@echo "Stopping services..."
	@docker compose down

# Linting
lint:
	@echo "Linting Backend..."
	@cd backend && golangci-lint run || echo "golangci-lint not installed or failed"
	@echo "Linting Frontend..."
	@cd frontend && npm run lint
	@echo "Linting Chaincode..."
	@cd teaTraceCC && npm run lint
