#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
PROJECT_NAME="lazychef"
DOCKER_COMPOSE_FILE="docker-compose.yml"
ENV_FILE=".env"

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_requirements() {
    log_info "Checking requirements..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed"
        exit 1
    fi
    
    log_info "Requirements check passed"
}

check_env_file() {
    if [ ! -f "$ENV_FILE" ]; then
        log_warn "Environment file $ENV_FILE not found"
        log_info "Creating sample environment file..."
        cat > "$ENV_FILE" << 'EOF'
# OpenAI API Configuration
OPENAI_API_KEY=your_openai_api_key_here

# Application Configuration
NODE_ENV=production
PORT=8080
FRONTEND_URL=http://localhost:3000
VITE_API_URL=http://localhost:8080/api

# Database Configuration (for future use)
# DATABASE_URL=sqlite://./data/lazychef.db
EOF
        log_warn "Please edit $ENV_FILE and set your API keys before running again"
        exit 1
    fi
    
    # Check if OPENAI_API_KEY is set
    if grep -q "your_openai_api_key_here" "$ENV_FILE"; then
        log_error "Please set your OPENAI_API_KEY in $ENV_FILE"
        exit 1
    fi
    
    log_info "Environment configuration check passed"
}

build_images() {
    log_info "Building Docker images..."
    docker-compose build --no-cache
    log_info "Docker images built successfully"
}

deploy_production() {
    log_info "Deploying to production..."
    
    # Stop existing containers
    log_info "Stopping existing containers..."
    docker-compose down --remove-orphans
    
    # Start production services
    log_info "Starting production services..."
    docker-compose up -d backend frontend
    
    # Wait for services to be healthy
    log_info "Waiting for services to be healthy..."
    timeout 120 docker-compose exec backend wget --quiet --tries=1 --spider http://localhost:8080/api/health
    
    log_info "Production deployment completed successfully!"
}

deploy_development() {
    log_info "Deploying development environment..."
    
    # Stop existing containers
    docker-compose down --remove-orphans
    
    # Start development services
    docker-compose --profile dev up -d
    
    log_info "Development environment started successfully!"
}

show_status() {
    log_info "Service status:"
    docker-compose ps
    
    echo ""
    log_info "Service logs (last 10 lines):"
    docker-compose logs --tail=10
    
    echo ""
    log_info "Health check:"
    if curl -s http://localhost:8080/api/health > /dev/null; then
        echo -e "${GREEN}✓ Backend API is healthy${NC}"
    else
        echo -e "${RED}✗ Backend API is not responding${NC}"
    fi
    
    if curl -s http://localhost:3000 > /dev/null; then
        echo -e "${GREEN}✓ Frontend is accessible${NC}"
    else
        echo -e "${RED}✗ Frontend is not responding${NC}"
    fi
}

cleanup() {
    log_info "Cleaning up Docker resources..."
    docker-compose down --remove-orphans --volumes
    docker system prune -f
    log_info "Cleanup completed"
}

backup_data() {
    log_info "Creating data backup..."
    timestamp=$(date +%Y%m%d_%H%M%S)
    backup_dir="backups/backup_${timestamp}"
    mkdir -p "$backup_dir"
    
    # Backup database files
    if [ -d "./backend/data" ]; then
        cp -r ./backend/data "$backup_dir/"
        log_info "Database backup created at $backup_dir"
    else
        log_warn "No data directory found to backup"
    fi
}

restore_data() {
    if [ -z "$1" ]; then
        log_error "Usage: $0 restore <backup_timestamp>"
        log_info "Available backups:"
        ls -la backups/ 2>/dev/null || echo "No backups found"
        exit 1
    fi
    
    backup_dir="backups/backup_$1"
    if [ ! -d "$backup_dir" ]; then
        log_error "Backup directory $backup_dir not found"
        exit 1
    fi
    
    log_info "Restoring data from $backup_dir..."
    
    # Stop services
    docker-compose down
    
    # Restore data
    rm -rf ./backend/data
    cp -r "$backup_dir/data" ./backend/
    
    log_info "Data restored successfully"
}

show_usage() {
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  production    Deploy to production environment"
    echo "  development   Deploy to development environment"
    echo "  build        Build Docker images"
    echo "  status       Show service status and health"
    echo "  logs         Show service logs"
    echo "  cleanup      Clean up Docker resources"
    echo "  backup       Create data backup"
    echo "  restore      Restore data from backup"
    echo "  stop         Stop all services"
    echo "  restart      Restart all services"
    echo "  help         Show this help message"
}

# Main script
case "${1:-production}" in
    "production")
        check_requirements
        check_env_file
        build_images
        deploy_production
        show_status
        ;;
    "development"|"dev")
        check_requirements
        check_env_file
        build_images
        deploy_development
        show_status
        ;;
    "build")
        check_requirements
        build_images
        ;;
    "status")
        show_status
        ;;
    "logs")
        docker-compose logs -f
        ;;
    "cleanup")
        cleanup
        ;;
    "backup")
        backup_data
        ;;
    "restore")
        restore_data "$2"
        ;;
    "stop")
        log_info "Stopping all services..."
        docker-compose down
        ;;
    "restart")
        log_info "Restarting all services..."
        docker-compose restart
        ;;
    "help"|"-h"|"--help")
        show_usage
        ;;
    *)
        log_error "Unknown command: $1"
        show_usage
        exit 1
        ;;
esac