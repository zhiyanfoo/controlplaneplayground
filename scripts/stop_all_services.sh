#!/bin/bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE_ROOT="$SCRIPT_DIR/.."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Stopping all test services...${NC}"

# Function to stop a service by port
stop_service_by_port() {
    local port="$1"
    local service_name="$2"
    
    echo -e "${YELLOW}Looking for processes on port $port...${NC}"
    
    # Find processes using the port
    local pids=$(lsof -ti:$port 2>/dev/null)
    
    if [ -z "$pids" ]; then
        echo -e "${YELLOW}No processes found on port $port${NC}"
        return
    fi
    
    for pid in $pids; do
        # Get and print process details
        local process_info=$(ps -p $pid -o pid,ppid,cmd --no-headers 2>/dev/null)
        echo -e "${YELLOW}Found process on port $port: $process_info${NC}"
        
        echo -e "${YELLOW}Stopping $service_name (PID: $pid) on port $port...${NC}"
        
        # Try graceful shutdown first
        kill $pid
        
        # Wait a bit for graceful shutdown
        sleep 2
        
        # Check if still running and force kill if needed
        if kill -0 $pid 2>/dev/null; then
            echo -e "${RED}Force killing $service_name (PID: $pid) on port $port...${NC}"
            kill -9 $pid
        fi
        
        echo -e "${GREEN}✓ $service_name on port $port stopped${NC}"
    done
}

# Stop services by port
stop_service_by_port "50051" "Main Test Server"
stop_service_by_port "50052" "HTTP Test Server"
stop_service_by_port "50053" "Dynamic Test Server 1"
stop_service_by_port "50054" "Dynamic Test Server 2"

echo -e "${GREEN}All test services stopped!${NC}"

# Check if ports are still in use
echo -e "${BLUE}Checking if test ports are still in use...${NC}"
for port in 50051 50052 50053 50054; do
    if lsof -i:$port >/dev/null 2>&1; then
        echo -e "${RED}Warning: Port $port is still in use:${NC}"
        lsof -i:$port
    else
        echo -e "${GREEN}✓ Port $port is free${NC}"
    fi
done 
