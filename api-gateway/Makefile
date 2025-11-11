.PHONY: help build run test clean docker-build docker-up docker-down migrate-up migrate-down sqlc

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	@echo "Building..."
	@go build -o bin/api-gateway cmd/server/main.go

run: ## Run the application
	@echo "Running..."
	@go run cmd/server/main.go

test: ## Run tests
	@echo "Running tests..."
	@go test -v -cover ./...

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@go test -v -tags=integration ./tests/integration/...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@go clean

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -f docker/Dockerfile -t ibn-api-gateway:latest .

docker-up: ## Start all services
	@echo "Starting services..."
	@docker-compose -f docker/docker-compose.yml up -d

docker-down: ## Stop all services
	@echo "Stopping services..."
	@docker-compose -f docker/docker-compose.yml down

docker-logs: ## View logs
	@docker-compose -f docker/docker-compose.yml logs -f

migrate-up: ## Run database migrations
	@echo "Running migrations..."
	@migrate -path migrations -database "postgres://gateway:changeme@localhost:5432/ibn_gateway?sslmode=disable" up

migrate-down: ## Rollback database migrations
	@echo "Rolling back migrations..."
	@migrate -path migrations -database "postgres://gateway:changeme@localhost:5432/ibn_gateway?sslmode=disable" down 1

migrate-create: ## Create new migration (use: make migrate-create NAME=migration_name)
	@echo "Creating migration..."
	@migrate create -ext sql -dir migrations -seq $(NAME)

sqlc: ## Generate Go code from SQL
	@echo "Generating sqlc code..."
	@sqlc generate

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...

swagger: ## Generate Swagger documentation
	@echo "Generating Swagger docs..."
	@swag init -g cmd/server/main.go -o docs/swagger

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

setup-wallet: ## Setup Fabric wallet
	@echo "Setting up wallet..."
	@bash scripts/setup-wallet.sh

test-connection: ## Test Fabric connection
	@echo "Testing Fabric connection..."
	@go run scripts/test-connection.go

