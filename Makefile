.PHONY: build run test clean setup dev deps lint fmt quality help frontend-lint frontend-build frontend-dev frontend-install fullstack-dev stop

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

## frontend-install: Install frontend dependencies
frontend-install:
	@echo "Installing frontend dependencies..."
	@if [ -d "frontend" ]; then \
		cd frontend && npm install; \
	else \
		echo "Frontend directory not found, skipping frontend install"; \
	fi

## frontend-dev: Start frontend development server
frontend-dev:
	@echo "Starting frontend development server..."
	@if [ -d "frontend" ]; then \
		cd frontend && npm run dev; \
	else \
		echo "Frontend directory not found, skipping frontend dev"; \
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
.PHONY: start stop restart logs quickstart quickstart-gui poc-demo reset-db demo-data logs-errors logs-api status

## fullstack-dev: Start both backend and frontend in development mode
fullstack-dev:
	@echo "🚀 LazyChef Full Stack Development Mode"
	@echo "Starting backend and frontend servers..."
	@echo "📋 Backend: http://localhost:8080"
	@echo "🌐 Frontend: http://localhost:3000"
	@echo ""
	@echo "💡 Use 'make stop' in another terminal to stop all services"
	@echo "Press Ctrl+C to stop all services"
	@($(MAKE) run &) && $(MAKE) frontend-dev

## quickstart: Complete setup and start for PoC (Backend only)
quickstart:
	@echo "🚀 LazyChef QuickStart - Backend API Setup & Launch"
	@echo "1. Checking environment..."
	@if [ ! -f .env ]; then \
		echo "Creating .env from template..."; \
		cp .env.example .env; \
		echo "⚠️  Please edit .env and set your OPENAI_API_KEY"; \
		echo "⚠️  Then run 'make quickstart' again"; \
		exit 1; \
	fi
	@echo "2. Installing dependencies..."
	@$(MAKE) deps >/dev/null 2>&1 || echo "Dependencies installation completed"
	@echo "3. Initializing database..."
	@$(MAKE) init-db
	@echo "4. Starting server..."
	@echo "✅ Setup complete! Backend API server starting on http://localhost:8080"
	@echo "📋 API Health: http://localhost:8080/api/health"
	@echo "🎯 Admin Panel: http://localhost:8080/api/admin/health"
	@echo ""
	@echo "💡 For full GUI experience, run 'make quickstart-gui' instead"
	@echo "💡 Use 'make stop' to stop the server"
	@$(MAKE) run

## quickstart-gui: Complete setup and start with GUI (Frontend + Backend)
quickstart-gui:
	@echo "🚀 LazyChef QuickStart - GUI版完全セットアップ & 起動"
	@echo "1. Checking environment..."
	@if [ ! -f .env ]; then \
		echo "Creating .env from template..."; \
		cp .env.example .env; \
		echo "⚠️  Please edit .env and set your OPENAI_API_KEY"; \
		echo "⚠️  Then run 'make quickstart-gui' again"; \
		exit 1; \
	fi
	@echo "2. Installing backend dependencies..."
	@$(MAKE) deps >/dev/null 2>&1 || echo "Backend dependencies installation completed"
	@echo "3. Installing frontend dependencies..."
	@$(MAKE) frontend-install >/dev/null 2>&1 || echo "Frontend dependencies installation completed"
	@echo "4. Initializing database..."
	@$(MAKE) init-db
	@echo "5. Starting backend and frontend servers..."
	@echo "✅ Setup complete! LazyChef starting with GUI"
	@echo "🌐 Frontend GUI: http://localhost:3000"
	@echo "📋 Backend API: http://localhost:8080"
	@echo "🎯 Admin Panel: http://localhost:8080/api/admin/health"
	@echo ""
	@echo "💡 Use 'make stop' in another terminal to stop all services"
	@echo "Press Ctrl+C to stop all services"
	@$(MAKE) fullstack-dev

## poc-demo: Run PoC demonstration commands
poc-demo:
	@echo "🎯 LazyChef PoC Demonstration"
	@echo "\n1. Health Check..."
	@curl -s http://localhost:8080/api/health | jq '.' || echo "Server not running. Start with 'make quickstart'"
	@echo "\n\n2. Basic Recipe Generation Demo..."
	@curl -s -X POST http://localhost:8080/api/recipes/generate \
		-H "Content-Type: application/json" \
		-d '{"preferences": {"cooking_time": 10, "ingredients": ["卵"]}}' | jq '.data.title' || echo "Failed"
	@echo "\n\n3. Admin System Health..."
	@curl -s http://localhost:8080/api/admin/health | jq '.data.status' || echo "Admin not available"
	@echo "\n\n4. Batch Jobs Status..."
	@curl -s http://localhost:8080/api/admin/batch-generation/jobs | jq '.data.count' || echo "No batch jobs"
	@echo "\n✅ PoC Demo Complete!"

## reset-db: Reset database completely
reset-db:
	@echo "🗄️ Resetting database..."
	@rm -f backend/data/recipes.db
	@$(MAKE) init-db
	@echo "✅ Database reset complete"

## demo-data: Insert sample data for demonstration
demo-data: reset-db
	@echo "📝 Inserting demo data..."
	@cd scripts && go run init_db.go
	@echo "✅ Demo data inserted successfully"

## start: Quick start (alias for run)
start: run

## stop: Stop all LazyChef services
stop:
	@echo "🛑 Stopping LazyChef services..."
	@echo "Checking and stopping processes on ports 8080 and 3000..."
	@-lsof -ti:8080 | xargs kill -9 2>/dev/null || echo "No process on port 8080"
	@-lsof -ti:3000 | xargs kill -9 2>/dev/null || echo "No process on port 3000"
	@echo "Stopping backend services by process name..."
	@-pkill -f "go run.*main.go" 2>/dev/null || true
	@-pkill -f "lazychef" 2>/dev/null || true
	@-pkill -f "air" 2>/dev/null || true
	@echo "Stopping frontend services by process name..."
	@-pkill -f "vite.*dev" 2>/dev/null || true
	@-pkill -f "npm.*dev" 2>/dev/null || true
	@-pkill -f "node.*vite" 2>/dev/null || true
	@-pkill -f "frontend.*dev" 2>/dev/null || true
	@echo "Cleaning up any remaining processes..."
	@sleep 1
	@echo "✅ All LazyChef services stopped"
	@echo "💡 Ports 8080 and 3000 are now available"

## status: Check LazyChef services status
status:
	@echo "📊 LazyChef Services Status"
	@echo ""
	@echo "🔍 Checking port 8080 (Backend API):"
	@if lsof -i:8080 >/dev/null 2>&1; then \
		echo "✅ Backend running on port 8080"; \
		lsof -i:8080 | grep LISTEN || true; \
	else \
		echo "❌ No process on port 8080"; \
	fi
	@echo ""
	@echo "🔍 Checking port 3000 (Frontend):"
	@if lsof -i:3000 >/dev/null 2>&1; then \
		echo "✅ Frontend running on port 3000"; \
		lsof -i:3000 | grep LISTEN || true; \
	else \
		echo "❌ No process on port 3000"; \
	fi
	@echo ""
	@echo "🔍 LazyChef processes:"
	@( \
		ps aux | grep "lazychef" | grep -v grep; \
		ps aux | grep "go run.*cmd/api/main.go" | grep -v grep; \
		ps aux | grep "go run.*main.go" | grep lazychef | grep -v grep; \
		ps aux | grep "air" | grep lazychef | grep -v grep; \
		ps aux | grep "vite.*dev" | grep frontend | grep -v grep; \
		ps aux | grep "npm.*dev" | grep frontend | grep -v grep; \
	) 2>/dev/null | head -10 || echo "❌ No LazyChef processes found"
	@echo ""
	@echo "💡 Use 'make quickstart' or 'make quickstart-gui' to start services"

## restart: Stop and start again  
restart: stop
	@sleep 2
	@$(MAKE) run

## logs: Show application logs (placeholder)
logs:
	@echo "📋 Application Logs:"
	@if [ -f backend/lazychef.log ]; then tail -f backend/lazychef.log; else echo "No log file found. Starting server with 'make run' will create logs."; fi

## logs-errors: Show only error logs
logs-errors:
	@echo "🚨 Error Logs:"
	@if [ -f backend/lazychef.log ]; then grep -i error backend/lazychef.log | tail -20; else echo "No error logs found"; fi

## logs-api: Show API access logs
logs-api:
	@echo "🌐 API Access Logs:"
	@if [ -f backend/lazychef.log ]; then grep -E "(GET|POST|PUT|DELETE)" backend/lazychef.log | tail -20; else echo "No API logs found"; fi