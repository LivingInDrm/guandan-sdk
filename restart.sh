#!/bin/bash

# Guandan Server Restart Script
# Usage: ./restart.sh [backend|frontend|all]

set -e

PROJECT_DIR="/Users/xiaochunliu/program/guandan"
BACKEND_BINARY="$PROJECT_DIR/guandan-server"
FRONTEND_DIR="$PROJECT_DIR/frontend"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Kill processes
kill_backend() {
    print_status "Stopping backend server..."
    pkill -f guandan-server || print_warning "No backend process found"
    sleep 1
}

kill_frontend() {
    print_status "Stopping frontend server..."
    pkill -f vite || print_warning "No frontend process found"
    sleep 1
}

# Start backend
start_backend() {
    print_status "Building backend..."
    cd "$PROJECT_DIR"
    go build -o guandan-server ./cmd/guandan-server
    
    print_status "Starting backend server on port 3000..."
    ./guandan-server &
    
    # Wait for backend to be ready
    print_status "Waiting for backend to start..."
    for i in {1..10}; do
        if curl -s http://localhost:3000/api/health > /dev/null 2>&1; then
            print_status "Backend is ready!"
            break
        fi
        if [ $i -eq 10 ]; then
            print_error "Backend failed to start!"
            exit 1
        fi
        sleep 1
    done
}

# Start frontend
start_frontend() {
    print_status "Starting frontend server on port 5173..."
    cd "$FRONTEND_DIR"
    npm run dev > /dev/null 2>&1 &
    
    # Wait for frontend to be ready
    print_status "Waiting for frontend to start..."
    for i in {1..15}; do
        if curl -s http://localhost:5173/ > /dev/null 2>&1; then
            print_status "Frontend is ready!"
            break
        fi
        if [ $i -eq 15 ]; then
            print_error "Frontend failed to start!"
            exit 1
        fi
        sleep 1
    done
}

# Test services
test_services() {
    print_status "Testing services..."
    
    # Test backend
    BACKEND_STATUS=$(curl -s http://localhost:3000/api/health | grep -o '"status":"ok"' || echo "failed")
    if [ "$BACKEND_STATUS" = '"status":"ok"' ]; then
        print_status "✅ Backend: http://localhost:3000/api/health"
    else
        print_error "❌ Backend health check failed"
    fi
    
    # Test frontend
    if curl -s http://localhost:5173/ | grep -q "掼蛋游戏"; then
        print_status "✅ Frontend: http://localhost:5173/"
    else
        print_error "❌ Frontend access failed"
    fi
    
    print_status "🎮 Game ready at: http://localhost:3000/"
}

# Main script
case "${1:-all}" in
    "backend")
        kill_backend
        start_backend
        test_services
        ;;
    "frontend")
        kill_frontend
        start_frontend
        test_services
        ;;
    "all")
        kill_backend
        kill_frontend
        start_backend
        start_frontend
        test_services
        ;;
    *)
        echo "Usage: $0 [backend|frontend|all]"
        echo "  backend  - Restart only the Go backend server"
        echo "  frontend - Restart only the React frontend server"
        echo "  all      - Restart both services (default)"
        exit 1
        ;;
esac

print_status "Restart complete!"