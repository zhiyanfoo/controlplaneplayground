#!/bin/bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE_ROOT="$SCRIPT_DIR/.."
LOG_FILE="$WORKSPACE_ROOT/test_services.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to cleanup on exit
cleanup() {
    echo -e "\n${YELLOW}Received interrupt signal. Stopping all services...${NC}"
    "$SCRIPT_DIR/stop_all_services.sh"
    exit 0
}

# Set up trap to catch Ctrl+C and call cleanup
trap cleanup SIGINT SIGTERM

echo -e "${BLUE}Starting all test services...${NC}"

# Clean up any existing log file
rm -f "$LOG_FILE"
touch "$LOG_FILE"

# Function to start a service in background
start_service() {
    local name="$1"
    local cmd="$2"
    local log_prefix="$3"
    
    echo -e "${YELLOW}Starting $name...${NC}"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Starting $name" >> "$LOG_FILE"
    
    # Start the service in background and redirect output to log file
    eval "$cmd" >> "$LOG_FILE" 2>&1 &
    local pid=$!
    
    echo -e "${GREEN}$name started with PID: $pid${NC}"
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $name started with PID: $pid" >> "$LOG_FILE"
    
    # Give it a moment to start
    sleep 1
    
    # Check if process is still running
    if kill -0 $pid 2>/dev/null; then
        echo -e "${GREEN}✓ $name is running${NC}"
    else
        echo -e "${RED}✗ $name failed to start${NC}"
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] $name failed to start" >> "$LOG_FILE"
    fi
}

# Build the test server first
echo -e "${BLUE}Building test server...${NC}"
cd "$WORKSPACE_ROOT/test-server"
go build -o server main.go
cd "$WORKSPACE_ROOT"

# Start the main test server (port 50051)
start_service "Main Test Server" \
    "cd $WORKSPACE_ROOT/test-server && ./server -port 50051 -message 'Main'" \
    "[MAIN]"

# Start dynamic test server 1 (port 50053)
start_service "Dynamic Test Server 1" \
    "cd $WORKSPACE_ROOT/test-server && ./server -port 50053 -message 'Dynamic1'" \
    "[DYN1]"

# Start dynamic test server 2 (port 50054)
start_service "Dynamic Test Server 2" \
    "cd $WORKSPACE_ROOT/test-server && ./server -port 50054 -message 'Dynamic2'" \
    "[DYN2]"

# Start HTTP test server (port 50052)
start_service "HTTP Test Server" \
    "cd $WORKSPACE_ROOT/test-server-http && ./http_server -port 50052 -message 'HTTP'" \
    "[HTTP]"

echo -e "${BLUE}All services started!${NC}"
echo -e "${YELLOW}Log file: $LOG_FILE${NC}"
echo ""
echo -e "${BLUE}=== Service Management Commands ===${NC}"
echo -e "${GREEN}# View live logs:${NC}"
echo "tail -f $LOG_FILE"
echo -e "${GREEN}# Test the services:${NC}"
echo "# Test main server:"
echo "grpcurl -plaintext -import-path $WORKSPACE_ROOT/test-server -proto $WORKSPACE_ROOT/test-server/test.proto -d '{\"name\": \"Test User\"}' localhost:50051 test.TestService/SayHello"
echo ""
echo "# Test dynamic server 1:"
echo "grpcurl -plaintext -import-path $WORKSPACE_ROOT/test-server -proto $WORKSPACE_ROOT/test-server/test.proto -d '{\"name\": \"Test User\"}' localhost:50053 test.TestService/SayHello"
echo ""
echo "# Test dynamic server 2:"
echo "grpcurl -plaintext -import-path $WORKSPACE_ROOT/test-server -proto $WORKSPACE_ROOT/test-server/test.proto -d '{\"name\": \"Test User\"}' localhost:50054 test.TestService/SayHello"
echo ""
echo -e "${BLUE}=== Starting live log view ===${NC}"
echo -e "${YELLOW}Press Ctrl+C to stop the log view (services will continue running)${NC}"
echo ""

# Start live log view
tail -f "$LOG_FILE" 
