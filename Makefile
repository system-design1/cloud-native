# Variables
PROJECT_NAME := go-backend-service
DOCKER_IMAGE := $(PROJECT_NAME)
DOCKER_TAG := latest
GO := go
DOCKER := docker
DOCKER_COMPOSE := docker-compose

# Default target
.PHONY: help
help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Build the Go project
.PHONY: build
build: ## Build the Go project and output the binary
	@echo "Building $(PROJECT_NAME)..."
	@$(GO) build -o $(PROJECT_NAME) ./cmd/server
	@echo "Build completed: $(PROJECT_NAME)"

# Run the application locally
.PHONY: run
run: ## Run the Go project directly without Docker using go run
	@echo "Running $(PROJECT_NAME)..."
	@$(GO) run ./cmd/server

# Build Docker image
.PHONY: docker-build
docker-build: ## Build the Docker image for the Go project
	@echo "Building Docker image $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	@$(DOCKER) build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker image built successfully"

# Start Docker Compose services
.PHONY: docker-up
docker-up: ## Start the Docker containers with docker-compose up -d
	@echo "Starting Docker containers..."
	@$(DOCKER_COMPOSE) up -d
	@echo "Docker containers started"

# Stop Docker Compose services
.PHONY: docker-down
docker-down: ## Stop the Docker containers with docker-compose down
	@echo "Stopping Docker containers..."
	@$(DOCKER_COMPOSE) down
	@echo "Docker containers stopped"

# View Docker Compose logs
.PHONY: docker-logs
docker-logs: ## View logs from Docker containers
	@$(DOCKER_COMPOSE) logs -f

# Clean build artifacts
.PHONY: clean
clean: ## Clean the project by removing the generated binary file
	@echo "Cleaning project..."
	@rm -f $(PROJECT_NAME)
	@echo "Clean completed"

# Run Go tests
.PHONY: test
test: ## Run the Go tests using go test
	@echo "Running tests..."
	@$(GO) test -v ./...
	@echo "Tests completed"

# Run Go tests with coverage
.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@$(GO) test -v -coverprofile=coverage.out ./...
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format Go code
.PHONY: fmt
fmt: ## Format Go code using gofmt
	@echo "Formatting code..."
	@$(GO) fmt ./...
	@echo "Formatting completed"

# Run Go linter (requires golangci-lint)
.PHONY: lint
lint: ## Run golangci-lint (if installed)
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "Running linter..."; \
		golangci-lint run ./...; \
		echo "Linting completed"; \
	else \
		echo "golangci-lint not installed. Install it from https://golangci-lint.run/"; \
	fi

# Generate Swagger documentation (optional - requires swag)
.PHONY: swagger
swagger: ## Generate Swagger spec file (requires swag tool)
	@if command -v swag >/dev/null 2>&1; then \
		echo "Generating Swagger documentation..."; \
		swag init -g cmd/server/main.go -o docs/swagger; \
		echo "Swagger documentation generated"; \
	else \
		echo "swag not installed. Install it with: go install github.com/swaggo/swag/cmd/swag@latest"; \
	fi

# Download Go dependencies
.PHONY: deps
deps: ## Download Go dependencies
	@echo "Downloading dependencies..."
	@$(GO) mod download
	@echo "Dependencies downloaded"

# Tidy Go modules
.PHONY: tidy
tidy: ## Tidy Go modules
	@echo "Tidying modules..."
	@$(GO) mod tidy
	@echo "Modules tidied"

# Full setup: download dependencies, build, and test
.PHONY: setup
setup: deps fmt build test ## Full setup: download dependencies, format, build, and test

