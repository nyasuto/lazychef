#!/bin/bash

# LazyChef Development Environment Startup Script
# This script starts the development environment with all necessary services

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Project root directory
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_DIR="$PROJECT_ROOT/backend"

echo -e "${BLUE}🚀 LazyChef Development Environment${NC}"
echo "================================================"

# Check if backend directory exists
if [ ! -d "$BACKEND_DIR" ]; then
  echo -e "${RED}❌ Backend directory not found: $BACKEND_DIR${NC}"
  exit 1
fi

# Change to backend directory
cd "$BACKEND_DIR"

# Check if .env file exists
if [ ! -f ".env" ]; then
  echo -e "${YELLOW}⚠️  .env file not found. Creating from .env.example...${NC}"
  if [ -f "../.env.example" ]; then
    cp "../.env.example" ".env"
    echo -e "${GREEN}✅ Created .env file${NC}"
    echo -e "${YELLOW}⚠️  Please edit .env file and set your OPENAI_API_KEY${NC}"
  else
    echo -e "${RED}❌ .env.example file not found${NC}"
    exit 1
  fi
fi

# Check if database is initialized
if [ ! -f "data/lazychef.db" ]; then
  echo -e "${YELLOW}📊 Database not found. Initializing...${NC}"
  cd "$PROJECT_ROOT/scripts"
  go run init_db.go
  cd "$BACKEND_DIR"
  echo -e "${GREEN}✅ Database initialized${NC}"
fi

# Install/update dependencies
echo -e "${BLUE}📦 Updating Go dependencies...${NC}"
go mod tidy

# Run tests to ensure everything is working
echo -e "${BLUE}🧪 Running tests...${NC}"
if go test ./... -v; then
  echo -e "${GREEN}✅ All tests passed${NC}"
else
  echo -e "${RED}❌ Some tests failed${NC}"
  echo -e "${YELLOW}⚠️  Continuing anyway... Fix tests when convenient${NC}"
fi

# Start the development server
echo -e "${BLUE}🏁 Starting LazyChef API server...${NC}"
echo "================================================"
echo -e "${GREEN}Server will start at: http://localhost:8080${NC}"
echo -e "${GREEN}Health check: http://localhost:8080/api/health${NC}"
echo -e "${GREEN}API docs: http://localhost:8080/api/recipes/search${NC}"
echo ""
echo -e "${YELLOW}Press Ctrl+C to stop the server${NC}"
echo "================================================"

# Export environment variables
export GIN_MODE=debug
export PORT=8080

# Start the server with live reload (if air is available)
if command -v air &> /dev/null; then
  echo -e "${BLUE}🔄 Starting with live reload (air)...${NC}"
  air
else
  echo -e "${YELLOW}💡 Install 'air' for live reload: go install github.com/air-verse/air@latest${NC}"
  echo -e "${BLUE}🔄 Starting server...${NC}"
  go run cmd/api/main.go
fi