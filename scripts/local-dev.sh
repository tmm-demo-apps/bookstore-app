#!/bin/bash
set -e

# Local Development Environment Manager
# Use this script to manage your local Docker Compose environment

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

function show_help() {
    cat << EOF
Local Development Environment Manager

Usage: ./scripts/local-dev.sh [command]

Commands:
    start       Start all services (build if needed)
    stop        Stop all services
    restart     Restart all services
    status      Show status of all services
    logs        Show logs (add service name for specific service)
    test        Run smoke tests
    clean       Stop and remove all containers and volumes
    rebuild     Clean rebuild of all services
    shell       Open shell in app container
    db          Open PostgreSQL shell
    redis       Open Redis CLI
    migrate     Run migrations manually
    seed        Seed images from Gutenberg
    help        Show this help message

Examples:
    ./scripts/local-dev.sh start
    ./scripts/local-dev.sh logs app
    ./scripts/local-dev.sh test
    ./scripts/local-dev.sh db

EOF
}

function check_docker() {
    if ! docker info > /dev/null 2>&1; then
        echo -e "${RED}Error: Docker is not running${NC}"
        echo "Please start Docker Desktop and try again"
        exit 1
    fi
}

function wait_for_db() {
    echo "Waiting for PostgreSQL to be ready..."
    for i in {1..30}; do
        if docker compose exec -T db pg_isready -U user -d bookstore > /dev/null 2>&1; then
            echo -e "${GREEN}PostgreSQL is ready${NC}"
            return 0
        fi
        sleep 1
    done
    echo -e "${RED}PostgreSQL did not start in time${NC}"
    return 1
}

function start_services() {
    echo -e "${YELLOW}Starting local development environment...${NC}"
    check_docker
    docker compose up --build -d
    
    echo ""
    echo -e "${GREEN}Services started!${NC}"
    echo ""
    echo "Waiting for services to initialize..."
    
    # Wait for app to be ready
    for i in {1..60}; do
        if curl -s http://localhost:8080/health > /dev/null 2>&1; then
            echo -e "${GREEN}Application is ready!${NC}"
            break
        fi
        if [ $i -eq 60 ]; then
            echo -e "${YELLOW}Application is still starting...${NC}"
            echo "Check logs: ./scripts/local-dev.sh logs app"
        fi
        sleep 2
    done
    
    echo ""
    echo "Application: http://localhost:8080"
    echo "MinIO Console: http://localhost:9001 (minioadmin/minioadmin)"
    echo "Elasticsearch: http://localhost:9200"
    echo ""
    echo "Run './scripts/local-dev.sh test' to run smoke tests"
    echo "See 'docs/DEVELOPMENT-WORKFLOW.md' for complete guide"
}

function stop_services() {
    echo -e "${YELLOW}Stopping services...${NC}"
    docker compose down
    echo -e "${GREEN}Services stopped${NC}"
}

function restart_services() {
    echo -e "${YELLOW}Restarting services...${NC}"
    docker compose restart
    echo -e "${GREEN}Services restarted${NC}"
}

function show_status() {
    check_docker
    echo -e "${YELLOW}Service Status:${NC}"
    docker compose ps
}

function show_logs() {
    check_docker
    if [ -z "$1" ]; then
        docker compose logs -f
    else
        docker compose logs -f "$1"
    fi
}

function run_tests() {
    echo -e "${YELLOW}Running smoke tests...${NC}"
    check_docker
    
    # Check if services are running
    if ! docker compose ps | grep -q "Up"; then
        echo -e "${RED}Services are not running${NC}"
        echo "Start services first: ./scripts/local-dev.sh start"
        exit 1
    fi
    
    # Wait for app to be ready
    echo "Waiting for application to be ready..."
    for i in {1..30}; do
        if curl -s http://localhost:8080/health > /dev/null 2>&1; then
            echo -e "${GREEN}Application is ready${NC}"
            break
        fi
        if [ $i -eq 30 ]; then
            echo -e "${RED}Application did not start in time${NC}"
            echo "Check logs: ./scripts/local-dev.sh logs app"
            exit 1
        fi
        sleep 1
    done
    
    # Run tests
    ./tests/smoke.sh
}

function clean_all() {
    echo -e "${YELLOW}Cleaning up all containers and volumes...${NC}"
    read -p "This will delete all data. Are you sure? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker compose down -v
        echo -e "${GREEN}Cleanup complete${NC}"
    else
        echo "Cancelled"
    fi
}

function rebuild_all() {
    echo -e "${YELLOW}Rebuilding all services...${NC}"
    docker compose down
    docker compose build --no-cache
    docker compose up -d
    echo -e "${GREEN}Rebuild complete${NC}"
}

function open_shell() {
    check_docker
    echo -e "${YELLOW}Opening shell in app container...${NC}"
    docker compose exec app /bin/sh
}

function open_db() {
    check_docker
    echo -e "${YELLOW}Opening PostgreSQL shell...${NC}"
    docker compose exec db psql -U user -d bookstore
}

function open_redis() {
    check_docker
    echo -e "${YELLOW}Opening Redis CLI...${NC}"
    docker compose exec redis redis-cli
}

function run_migrations() {
    check_docker
    echo -e "${YELLOW}Running migrations...${NC}"
    
    wait_for_db || exit 1
    
    echo "Applying 001_schema.sql..."
    docker compose exec -T db psql -U user -d bookstore -f /migrations/001_schema.sql 2>/dev/null || echo "Schema may already exist"
    
    echo "Applying 002_seed_books.sql..."
    docker compose exec -T db psql -U user -d bookstore -f /migrations/002_seed_books.sql
    
    echo -e "${GREEN}Migrations complete!${NC}"
}

function seed_images() {
    check_docker
    echo -e "${YELLOW}Seeding images from Gutenberg...${NC}"
    
    # Run seed-images container
    docker compose run --rm seed-images
    
    echo -e "${GREEN}Image seeding complete!${NC}"
}

# Main command handler
case "${1:-help}" in
    start)
        start_services
        ;;
    stop)
        stop_services
        ;;
    restart)
        restart_services
        ;;
    status)
        show_status
        ;;
    logs)
        show_logs "$2"
        ;;
    test)
        run_tests
        ;;
    clean)
        clean_all
        ;;
    rebuild)
        rebuild_all
        ;;
    shell)
        open_shell
        ;;
    db)
        open_db
        ;;
    redis)
        open_redis
        ;;
    migrate)
        run_migrations
        ;;
    seed)
        seed_images
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        echo -e "${RED}Unknown command: $1${NC}"
        echo ""
        show_help
        exit 1
        ;;
esac
