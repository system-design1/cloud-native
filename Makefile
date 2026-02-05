# Variables
PROJECT_NAME := go-backend-service
DOCKER_IMAGE := $(PROJECT_NAME)
DOCKER_TAG := latest
GO := go
DOCKER := docker
DOCKER_COMPOSE := $(shell docker compose version >/dev/null 2>&1 && echo "docker compose" || echo "docker-compose")
BIN_DIR := bin
BINARY := $(BIN_DIR)/$(PROJECT_NAME)

# BuildKit settings - enable for faster builds
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1

# Default target
.PHONY: help
help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Build the Go project
.PHONY: build
build: ## Build the Go project and output the binary to bin/
	@echo "Building $(PROJECT_NAME)..."
	@mkdir -p $(BIN_DIR)
	@$(GO) build -o $(BINARY) ./cmd/server
	@echo "Build completed: $(BINARY)"

# Local Development Setup
.PHONY: dev-setup
dev-setup: ## Setup local development environment (create .env from env.example)
	@if [ ! -f .env ]; then \
		echo "Creating .env file from env.example..."; \
		cp env.example .env; \
		echo ".env file created. Updating DB_HOST to 'localhost' for local development..."; \
		sed -i 's/^DB_HOST=postgres/DB_HOST=localhost/' .env 2>/dev/null || \
		sed -i '' 's/^DB_HOST=postgres/DB_HOST=localhost/' .env 2>/dev/null || true; \
		echo "Setup completed! DB_HOST set to: $$(grep '^DB_HOST' .env | cut -d'=' -f2)"; \
	else \
		echo ".env file already exists."; \
		CURRENT_DB_HOST=$$(grep '^DB_HOST' .env | cut -d'=' -f2); \
		if [ "$$CURRENT_DB_HOST" = "postgres" ]; then \
			echo "Updating DB_HOST from 'postgres' to 'localhost' for local development..."; \
			sed -i 's/^DB_HOST=postgres/DB_HOST=localhost/' .env 2>/dev/null || \
			sed -i '' 's/^DB_HOST=postgres/DB_HOST=localhost/' .env 2>/dev/null || true; \
			echo "DB_HOST updated to: $$(grep '^DB_HOST' .env | cut -d'=' -f2)"; \
		else \
			echo "DB_HOST is already set to: $$CURRENT_DB_HOST"; \
		fi \
	fi

# Start local development database
.PHONY: dev-db-up
dev-db-up: ## Start PostgreSQL database for local development
	@echo "Starting local development database..."
	@$(DOCKER_COMPOSE) -f docker-compose.dev.yml up -d
	@echo "Waiting for database to be ready..."
	@timeout=30; \
	while [ $$timeout -gt 0 ]; do \
		if docker exec go-backend-postgres-dev pg_isready -U postgres >/dev/null 2>&1; then \
			echo "Database is ready!"; \
			break; \
		fi; \
		sleep 1; \
		timeout=$$((timeout-1)); \
	done; \
	if [ $$timeout -eq 0 ]; then \
		echo "Warning: Database might not be ready yet."; \
	fi

# Stop local development database
.PHONY: dev-db-down
dev-db-down: ## Stop local development database
	@echo "Stopping local development database..."
	@$(DOCKER_COMPOSE) -f docker-compose.dev.yml down
	@echo "Database stopped"

# Run the application locally (requires .env file and database)
.PHONY: run
run: ## Run the Go project directly without Docker using go run
	@if [ ! -f .env ]; then \
		echo "Error: .env file not found. Run 'make dev-setup' first."; \
		exit 1; \
	fi
	@echo "Running $(PROJECT_NAME) locally..."
	@export $$(grep -v '^[[:space:]]*#' .env | grep -v '^[[:space:]]*$$' | xargs) && \
	if [ -f scripts/wait-for-db.sh ]; then \
		echo "Checking if database is ready..."; \
		./scripts/wait-for-db.sh || (echo "Database is not ready. Starting database..." && make dev-db-up && ./scripts/wait-for-db.sh); \
	fi
	@echo "Starting server..."
	@export $$(grep -v '^[[:space:]]*#' .env | grep -v '^[[:space:]]*$$' | xargs) && $(GO) run ./cmd/server

# Run with hot reload using air (if installed)
.PHONY: dev-run
dev-run: ## Run the app locally with hot reload using air (requires air: go install github.com/air-verse/air@latest)
	@if [ ! -f .env ]; then \
		echo "Error: .env file not found. Run 'make dev-setup' first."; \
		exit 1; \
	fi
	@AIR_BIN=$$(command -v air 2>/dev/null || echo ""); \
	if [ -z "$$AIR_BIN" ]; then \
		AIR_PATH=$$($(GO) env GOPATH)/bin/air; \
		if [ -f "$$AIR_PATH" ]; then \
			AIR_BIN="$$AIR_PATH"; \
		fi; \
	fi; \
	if [ -n "$$AIR_BIN" ] && [ -f "$$AIR_BIN" ]; then \
		echo "Running $(PROJECT_NAME) with hot reload (air)..."; \
		echo "Make sure database is running (use 'make dev-db-up' if needed)"; \
		export $$(grep -v '^[[:space:]]*#' .env | grep -v '^[[:space:]]*$$' | xargs) && $$AIR_BIN; \
	else \
		echo "air is not installed. Installing..."; \
		$(GO) install github.com/air-verse/air@latest; \
		AIR_PATH=$$($(GO) env GOPATH)/bin/air; \
		if [ -f "$$AIR_PATH" ]; then \
			echo "Running $(PROJECT_NAME) with hot reload (air)..."; \
			echo "Make sure database is running (use 'make dev-db-up' if needed)"; \
			export $$(grep -v '^[[:space:]]*#' .env | grep -v '^[[:space:]]*$$' | xargs) && $$AIR_PATH; \
		else \
			echo "Error: Failed to install air. Please install manually:"; \
			echo "  go install github.com/air-verse/air@latest"; \
			echo "  Then add $$($(GO) env GOPATH)/bin to your PATH"; \
			exit 1; \
		fi \
	fi

# Run health checker for local development (calls /health, /ready, /live periodically)
.PHONY: dev-health-checker
dev-health-checker: ## Run health checker to automatically call health endpoints (for local dev)
	@if [ ! -f scripts/health-checker.sh ]; then \
		echo "Error: scripts/health-checker.sh not found."; \
		exit 1; \
	fi
	@bash scripts/health-checker.sh

# Complete local development setup: database + app
.PHONY: dev
dev: dev-setup dev-db-up ## Complete local dev setup: create .env, start database, and show instructions
	@echo ""
	@echo "=========================================="
	@echo "Local development environment is ready!"
	@echo "=========================================="
	@echo ""
	@echo "Database is running on: localhost:5432"
	@echo ""
	@echo "To run the application:"
	@echo "  make run          - Run once"
	@echo "  make dev-run      - Run with hot reload (air)"
	@echo ""
	@echo "To stop database:"
	@echo "  make dev-db-down"
	@echo ""

# Build Docker image (only rebuilds if needed)
.PHONY: docker-build
docker-build: ## Build the Docker image (only rebuilds if Dockerfile or code changed)
	@echo "Building Docker image $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	@echo "Note: Docker will use cache if nothing changed. Use 'make docker-build-rebuild' to force rebuild."
	@$(DOCKER) build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker image built successfully"

# Force rebuild Docker image
.PHONY: docker-build-rebuild
docker-build-rebuild: ## Force rebuild Docker image without cache
	@echo "WARNING: --no-cache build (VERY SLOW in Iran). Use only if cache is broken. Force rebuilding Docker image $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	@$(DOCKER) build --no-cache -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker image rebuilt successfully"

# Build Docker image without cache
.PHONY: docker-build-no-cache
docker-build-no-cache: ## Build the Docker image without using cache
	@echo "Building Docker image $(DOCKER_IMAGE):$(DOCKER_TAG) without cache..."
	@$(DOCKER) build --no-cache --progress=plain -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker image built successfully"

# Start Docker Compose services (skips rebuild if image exists)
.PHONY: docker-up
docker-up: ## Start the Docker containers (skips build if image exists, use docker-up-rebuild to force rebuild)
	@if [ ! -f .env ]; then \
		echo "Warning: .env file not found. Creating from env.example..."; \
		cp env.example .env; \
		echo ".env file created. You may need to adjust values."; \
	fi
	@echo "Starting Docker containers..."
	@echo "Note: If code changed, use 'make docker-up-api-build' (cache-friendly). Use 'make docker-up-no-cache' only if cache is broken."
	@$(DOCKER_COMPOSE) up -d
	@echo ""
	@echo "=========================================="
	@echo "Docker containers started!"
	@echo "=========================================="
	@echo "API: http://localhost:8080"
	@echo "Health: http://localhost:8080/health"
	@echo ""
	@echo "To view logs: make docker-logs"
	@echo "To stop: make docker-down"

# Rebuild and start Docker Compose services 
.PHONY: docker-up-rebuild
docker-up-rebuild: ## Rebuild and start Docker containers (cache-friendly)
	@echo "Rebuilding Docker containers (using cache)..."
	@$(DOCKER_COMPOSE) build
	@$(DOCKER_COMPOSE) up -d
	@echo "Docker containers rebuilt and started"

# Rebuild and start Docker Compose services no-cache
.PHONY: docker-up-no-cache
docker-up-no-cache: ## Rebuild and start Docker containers WITHOUT cache (very slow; last resort)
	@echo "WARNING: rebuilding WITHOUT cache (VERY SLOW). Use only if cache is broken."
	@$(DOCKER_COMPOSE) build --no-cache
	@$(DOCKER_COMPOSE) up -d
	@echo "Docker containers rebuilt (no-cache) and started"

# Use it when you change go code
.PHONY: docker-up-api-build
docker-up-api-build: ## Rebuild only api service (cache-friendly) and start it
	@echo "Rebuilding only api service (using cache)..."
	@$(DOCKER_COMPOSE) up -d --build api


# Use it when you change .env and need to recreate
.PHONY: docker-up-api-recreate
docker-up-api-recreate: ## Recreate only api container to apply runtime changes (no build)
	@echo "Recreating api container (no build) to apply env/config changes..."
	@$(DOCKER_COMPOSE) up -d --force-recreate api


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
	@rm -f $(BINARY) $(PROJECT_NAME)
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

# ============================================
# Observability (OpenTelemetry, Tempo, Prometheus)
# ============================================

# Start observability stack (Tempo, Jaeger, Prometheus, Grafana, Loki, Promtail)
.PHONY: observability-up
observability-up: ## Start observability stack (Tempo, Jaeger, Prometheus, Grafana, Loki, Promtail)
	@echo "Starting observability stack..."
	@$(DOCKER_COMPOSE) -f docker-compose.observability.yml up -d
	@echo ""
	@echo "=========================================="
	@echo "Observability stack started!"
	@echo "=========================================="
	@echo "Grafana:        http://localhost:3000 (admin/admin)"
	@echo "  - Traces:     Use Tempo datasource in Explore"
	@echo "  - Logs:       Use Loki datasource in Explore"
	@echo "  - Metrics:    Use Prometheus datasource"
	@echo "Jaeger UI:      http://localhost:16686 (memory storage only)"
	@echo "Prometheus:     http://localhost:9090"
	@echo "Tempo API:      http://localhost:3200"
	@echo "Loki API:       http://localhost:3100"
	@echo ""
	@echo "Note: Use Grafana to view traces, logs, and metrics"
	@echo "To stop: make observability-down"

# Rebuild observability stack (pull latest images and restart)
.PHONY: observability-up-rebuild
observability-up-rebuild: ## Pull latest images and restart observability stack
	@echo "Rebuilding observability stack..."
	@echo "Pulling latest images..."
	@$(DOCKER_COMPOSE) -f docker-compose.observability.yml pull
	@echo "Restarting containers..."
	@$(DOCKER_COMPOSE) -f docker-compose.observability.yml up -d --force-recreate
	@echo ""
	@echo "=========================================="
	@echo "Observability stack rebuilt and started!"
	@echo "=========================================="
	@echo "Grafana:        http://localhost:3000 (admin/admin)"
	@echo "Jaeger UI:      http://localhost:16686"
	@echo "Prometheus:     http://localhost:9090"
	@echo "Tempo API:      http://localhost:3200"
	@echo "Loki API:       http://localhost:3100"

.PHONY: observability-up-recreate
observability-up-recreate: ## Force-recreate observability containers WITHOUT pulling images (fast)
	@echo "Recreating observability stack (no pull)..."
	@$(DOCKER_COMPOSE) -f docker-compose.observability.yml up -d --force-recreate
	@echo "Observability stack recreated"


# Stop observability stack
.PHONY: observability-down
observability-down: ## Stop observability stack
	@echo "Stopping observability stack..."
	@$(DOCKER_COMPOSE) -f docker-compose.observability.yml down
	@echo "Observability stack stopped"

# Stop observability stack and remove volumes (clears all data)
.PHONY: observability-down-clean
observability-down-clean: ## Stop observability stack and remove volumes (WARNING: clears all data)
	@echo "WARNING: This will remove all observability data (traces, metrics, dashboards)"
	@echo "Stopping observability stack and removing volumes..."
	@$(DOCKER_COMPOSE) -f docker-compose.observability.yml down -v
	@echo "Observability stack stopped and volumes removed"

# Reset observability stack (stop, remove volumes, and restart)
.PHONY: observability-reset
observability-reset: ## Reset observability stack: stop, remove all data, and restart (WARNING: deletes all traces, metrics, dashboards)
	@echo "WARNING: This will delete all observability data (traces, metrics, dashboards)"
	@echo "Resetting observability stack..."
	@$(DOCKER_COMPOSE) -f docker-compose.observability.yml down -v
	@echo "Starting observability stack from scratch..."
	@$(DOCKER_COMPOSE) -f docker-compose.observability.yml up -d
	@echo ""
	@echo "=========================================="
	@echo "Observability stack reset complete!"
	@echo "=========================================="
	@echo "Grafana:        http://localhost:3000 (admin/admin)"
	@echo "Jaeger UI:      http://localhost:16686"
	@echo "Prometheus:     http://localhost:9090"
	@echo "Tempo API:      http://localhost:3200"
	@echo "Loki API:       http://localhost:3100"
	@echo ""
	@echo "All previous data has been deleted. The stack is now clean and ready for new traces."

# View observability logs
.PHONY: observability-logs
observability-logs: ## View observability stack logs
	@$(DOCKER_COMPOSE) -f docker-compose.observability.yml logs -f

# Start Tempo only
.PHONY: tempo-up
tempo-up: ## Start Tempo tracing backend only
	@echo "Starting Tempo..."
	@$(DOCKER_COMPOSE) -f docker-compose.observability.yml up -d tempo jaeger
	@echo ""
	@echo "=========================================="
	@echo "Tempo started!"
	@echo "=========================================="
	@echo "Tempo API:      http://localhost:3200"
	@echo "OTLP HTTP:      http://localhost:4318"
	@echo "OTLP gRPC:      localhost:4317"
	@echo ""
	@echo "Note: Start Grafana (make grafana-up) to view traces from Tempo"
	@echo ""

# Stop Tempo
.PHONY: tempo-down
tempo-down: ## Stop Tempo
	@$(DOCKER_COMPOSE) -f docker-compose.observability.yml stop tempo jaeger

# Start Prometheus only
.PHONY: prometheus-up
prometheus-up: ## Start Prometheus metrics collection only
	@echo "Starting Prometheus..."
	@$(DOCKER_COMPOSE) -f docker-compose.observability.yml up -d prometheus
	@echo ""
	@echo "=========================================="
	@echo "Prometheus started!"
	@echo "=========================================="
	@echo "Prometheus UI:  http://localhost:9090"
	@echo ""

# Stop Prometheus
.PHONY: prometheus-down
prometheus-down: ## Stop Prometheus
	@$(DOCKER_COMPOSE) -f docker-compose.observability.yml stop prometheus

# Start Grafana only
.PHONY: grafana-up
grafana-up: ## Start Grafana visualization only
	@echo "Starting Grafana..."
	@$(DOCKER_COMPOSE) -f docker-compose.observability.yml up -d grafana
	@echo ""
	@echo "=========================================="
	@echo "Grafana started!"
	@echo "=========================================="
	@echo "Grafana UI:     http://localhost:3000"
	@echo "Username:       admin"
	@echo "Password:       admin"
	@echo ""

# Stop Grafana
.PHONY: grafana-down
grafana-down: ## Stop Grafana
	@$(DOCKER_COMPOSE) -f docker-compose.observability.yml stop grafana

# Start Loki only
.PHONY: loki-up
loki-up: ## Start Loki log aggregation only
	@echo "Starting Loki..."
	@$(DOCKER_COMPOSE) -f docker-compose.observability.yml up -d loki promtail
	@echo ""
	@echo "=========================================="
	@echo "Loki started!"
	@echo "=========================================="
	@echo "Loki API:       http://localhost:3100"
	@echo ""
	@echo "Note: Start Grafana (make grafana-up) to view logs from Loki"
	@echo ""

# Stop Loki
.PHONY: loki-down
loki-down: ## Stop Loki
	@$(DOCKER_COMPOSE) -f docker-compose.observability.yml stop loki promtail

# View Loki logs
.PHONY: loki-logs
loki-logs: ## View Loki and Promtail logs
	@$(DOCKER_COMPOSE) -f docker-compose.observability.yml logs -f loki promtail
