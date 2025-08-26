#!/bin/bash

# LazyChef Development Environment Setup Script

set -e

echo "🍳 LazyChef Development Environment Setup"
echo "========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed. Please install Go 1.21+ first.${NC}"
    echo "Visit: https://golang.org/doc/install"
    exit 1
fi

echo -e "${BLUE}✓ Go is installed: $(go version)${NC}"

# Check if Node.js is installed (for future frontend work)
if command -v node &> /dev/null; then
    echo -e "${BLUE}✓ Node.js is installed: $(node --version)${NC}"
else
    echo -e "${YELLOW}⚠ Node.js not found. Install it for frontend development.${NC}"
fi

# Navigate to project root
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

echo -e "${BLUE}📁 Working in: $PROJECT_ROOT${NC}"

# Copy environment file if it doesn't exist
if [ ! -f ".env" ]; then
    cp .env.example .env
    echo -e "${GREEN}✓ Created .env from .env.example${NC}"
    echo -e "${YELLOW}⚠ Please edit .env and add your OpenAI API key${NC}"
else
    echo -e "${BLUE}✓ .env file already exists${NC}"
fi

# Initialize Go module and install dependencies
echo -e "${BLUE}📦 Installing Go dependencies...${NC}"
cd backend

if [ ! -f "go.mod" ]; then
    go mod init lazychef
    echo -e "${GREEN}✓ Initialized Go module${NC}"
fi

# Install main dependencies
go get -u github.com/gin-gonic/gin
go get -u github.com/mattn/go-sqlite3
go get -u github.com/joho/godotenv
go get -u github.com/sashabaranov/go-openai

# Install development tools (optional)
echo -e "${BLUE}🔧 Installing development tools...${NC}"
go install github.com/air-verse/air@latest 2>/dev/null || echo -e "${YELLOW}⚠ Could not install air (hot reload)${NC}"

go mod tidy
echo -e "${GREEN}✓ Go dependencies installed${NC}"

# Create data directory
mkdir -p data
echo -e "${GREEN}✓ Created data directory${NC}"

cd "$PROJECT_ROOT"

# Create basic file structure if not exists
mkdir -p backend/{cmd/api,internal/{database,handlers,services,models,middleware},data}
mkdir -p scripts
mkdir -p frontend

echo -e "${GREEN}✓ Project structure created${NC}"

# Check for Git
if command -v git &> /dev/null && [ -d ".git" ]; then
    echo -e "${BLUE}✓ Git repository detected${NC}"
else
    echo -e "${YELLOW}⚠ No Git repository found. Initialize with: git init${NC}"
fi

# Final instructions
echo ""
echo -e "${GREEN}🎉 Setup completed successfully!${NC}"
echo ""
echo -e "${BLUE}Next steps:${NC}"
echo "1. Edit .env and add your OpenAI API key"
echo "2. Run: make run (to start the server)"
echo "3. For frontend: cd frontend && npx create-react-app ."
echo "4. Visit: http://localhost:8080/api/health"
echo ""
echo -e "${BLUE}Available commands:${NC}"
echo "  make help    - Show all available commands"
echo "  make run     - Start the development server"
echo "  make test    - Run tests"
echo "  make build   - Build the binary"
echo ""
echo -e "${YELLOW}Don't forget to set your OPENAI_API_KEY in .env!${NC}"