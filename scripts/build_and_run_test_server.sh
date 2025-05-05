#!/bin/bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
WORKSPACE_ROOT="$SCRIPT_DIR/.."

echo "Building test server..."
cd "$WORKSPACE_ROOT/test-server"
go build -o server main.go

echo "Running test server... (Press Ctrl+C to stop)"
./server 
