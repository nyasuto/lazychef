.PHONY: build run test clean setup dev deps lint fmt quality help frontend-lint frontend-build

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=lazychef
BINARY_PATH=bin/$(BINARY_NAME)
MAIN_PATH=cmd/api/main.go

## help: Display this help message
help:
	@echo "Available commands:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## setup: Initialize project dependencies and database
setup:
	@echo "Setting up LazyChef development environment..."
	cd backend && $(GOMOD) init lazychef 2>/dev/null || true
	cd backend && $(GOMOD) tidy
	$(MAKE) deps
	$(MAKE) init-db
	@echo "Setup complete! Run 'make dev' to start development."

## deps: Install Go dependencies
deps:
	@echo "Installing Go dependencies..."
	cd backend && $(GOGET) -u github.com/gin-gonic/gin
	cd backend && $(GOGET) -u github.com/mattn/go-sqlite3
	cd backend && $(GOGET) -u github.com/joho/godotenv
	cd backend && $(GOGET) -u github.com/sashabaranov/go-openai
	cd backend && $(GOMOD) tidy

## build: Build the application binary
build:
	@echo "Building $(BINARY_NAME)..."
	cd backend && $(GOBUILD) -o ../$(BINARY_PATH) $(MAIN_PATH)

## run: Run the application (development mode)
run:
	@echo "Starting LazyChef server..."
	cd backend && $(GOCMD) run $(MAIN_PATH)

## dev: Start development environment (with hot reload if available)
dev:
	@echo "Starting development server..."
	@if command -v air >/dev/null 2>&1; then \
		cd backend && air; \
	else \
		echo "Hot reload not available. Install with: go install github.com/air-verse/air@latest"; \
		$(MAKE) run; \
	fi

## test: Run all tests
test:
	@echo "Running tests..."
	cd backend && $(GOTEST) -v ./...

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	cd backend && $(GOTEST) -v -cover -coverprofile=coverage.out ./...
	cd backend && $(GOCMD) tool cover -html=coverage.out -o coverage.html

## lint: Run linter (if available)
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		cd backend && golangci-lint run; \
	else \
		echo "golangci-lint not found. Install from: https://golangci-lint.run/usage/install/"; \
		cd backend && $(GOCMD) vet ./...; \
	fi

## fmt: Format Go code
fmt:
	@echo "Formatting code..."
	cd backend && $(GOCMD) fmt ./...

## frontend-lint: Run frontend linting
frontend-lint:
	@echo "Running frontend linting..."
	@if [ -d "frontend" ]; then \
		cd frontend && npm run lint; \
	else \
		echo "Frontend directory not found, skipping frontend lint"; \
	fi

## frontend-build: Build frontend for production
frontend-build:
	@echo "Building frontend..."
	@if [ -d "frontend" ]; then \
		cd frontend && npm run build; \
	else \
		echo "Frontend directory not found, skipping frontend build"; \
	fi

## quality: Run all quality checks (format, lint, test, frontend)
quality: fmt lint test frontend-lint frontend-build
	@echo "All quality checks completed successfully!"

## init-db: Initialize SQLite database
init-db:
	@echo "Initializing database..."
	@if [ -f scripts/init_db.go ]; then \
		cd scripts && $(GOCMD) run init_db.go; \
	else \
		echo "Database initialization script not found. Creating basic structure..."; \
		mkdir -p backend/data; \
		touch backend/data/.gitkeep; \
	fi

## clean: Remove build artifacts and cache
clean:
	@echo "Cleaning up..."
	rm -rf $(BINARY_PATH)
	rm -rf backend/coverage.out
	rm -rf backend/coverage.html
	cd backend && $(GOCMD) clean -cache -modcache -testcache

## frontend-setup: Setup React frontend (to be run by Claude Code)
frontend-setup:
	@echo "Setting up React frontend..."
	@echo "This should be done by Claude Code. Please run:"
	@echo "  cd frontend && npx create-react-app . --template minimal"
	@echo "  cd frontend && npm install axios tailwindcss"

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t lazychef:latest .

## docker-run: Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env lazychef:latest

# Development shortcuts
.PHONY: start stop restart logs

## start: Quick start (alias for run)
start: run

## restart: Stop and start again  
restart: clean build run

## logs: Show application logs (placeholder)
logs:
	@echo "Logs would be shown here in production environment"