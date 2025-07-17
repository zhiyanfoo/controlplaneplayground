#!/bin/bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE_ROOT="$SCRIPT_DIR/../.."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Starting Stateful Session Test...${NC}"

# Check if test servers are running
echo -e "${YELLOW}Checking if test servers are running...${NC}"
if ! pgrep -f "server -port 50053" > /dev/null; then
    echo -e "${RED}Dynamic Test Server 1 (port 50053) is not running${NC}"
    echo -e "${YELLOW}Please run: $WORKSPACE_ROOT/scripts/start_all_services.sh${NC}"
    exit 1
fi

if ! pgrep -f "server -port 50054" > /dev/null; then
    echo -e "${RED}Dynamic Test Server 2 (port 50054) is not running${NC}"
    echo -e "${YELLOW}Please run: $WORKSPACE_ROOT/scripts/start_all_services.sh${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Test servers are running${NC}"

# Start Envoy in the background
echo -e "${YELLOW}Starting Envoy with stateful session config...${NC}"
cd "$SCRIPT_DIR"
envoy -c envoy.yaml &
ENVOY_PID=$!

# Function to cleanup on exit
cleanup() {
    echo -e "\n${YELLOW}Stopping Envoy...${NC}"
    kill $ENVOY_PID 2>/dev/null || true
    exit 0
}

# Set up trap to catch Ctrl+C and call cleanup
trap cleanup SIGINT SIGTERM

# Wait for Envoy to start
echo -e "${YELLOW}Waiting for Envoy to start...${NC}"
sleep 3

# Check if Envoy is running
if ! kill -0 $ENVOY_PID 2>/dev/null; then
    echo -e "${RED}✗ Envoy failed to start${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Envoy is running${NC}"

# Build and run the client
echo -e "${YELLOW}Building and running client...${NC}"
cd "$WORKSPACE_ROOT"
go run scenarios/stateful_sessions/client.go

# Cleanup
cleanup